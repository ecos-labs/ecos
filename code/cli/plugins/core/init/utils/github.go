package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ecos-labs/ecos-core/code/cli/utils"
)

// Constants for GitHub repository owner and name.
// TODO: Adjust these if repo or owner changes.
const (
	DefaultRepoOwner = "ecos-labs"
	DefaultRepoName  = "ecos"
)

// GitHubClient provides GitHub API operations
type GitHubClient struct {
	client *http.Client
	token  string
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"assets"`
}

// NewGitHubClient creates a GitHub client. GITHUB_TOKEN is optional for public repositories.
// If provided, it will be used for authentication which increases rate limits.
// For public repositories, unauthenticated access works but with lower rate limits.
func NewGitHubClient() (*GitHubClient, error) {
	token := os.Getenv("GITHUB_TOKEN")

	return &GitHubClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil // Allow all redirects
			},
		},
		token: token,
	}, nil
}

// GetLatestReleaseVersion gets the latest release version for a specific datasource
func (gc *GitHubClient) GetLatestReleaseVersion(ctx context.Context, repoOwner, repoName, datasource string) (string, error) {
	releases, err := gc.getAllReleases(ctx, repoOwner, repoName)
	if err != nil {
		return "", err
	}

	dsPrefix := fmt.Sprintf("ds/%s/", datasource)
	var matchingReleases []string

	for _, release := range releases {
		if strings.HasPrefix(release.TagName, dsPrefix) {
			version := strings.TrimPrefix(release.TagName, dsPrefix)
			return version, nil
		}
		matchingReleases = append(matchingReleases, release.TagName)
	}

	if len(releases) == 0 {
		return "", errors.New("no releases found in repository")
	}

	return "", fmt.Errorf("no releases found for datasource '%s' with prefix '%s'. Available releases: %v. Please ensure releases are tagged with format 'ds/%s/v1.0.0'",
		datasource, dsPrefix, matchingReleases, datasource)
}

// DownloadReleaseAsset downloads a specific asset from a GitHub release
func (gc *GitHubClient) DownloadReleaseAsset(ctx context.Context, repoOwner, repoName, releaseTag, assetName string) ([]byte, error) {
	release, err := gc.getRelease(ctx, repoOwner, repoName, releaseTag)
	if err != nil {
		return nil, err
	}

	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.URL
			break
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("asset '%s' not found in release '%s'", assetName, releaseTag)
	}

	return gc.downloadAsset(ctx, assetURL)
}

// DownloadDatasourcePackage downloads and extracts a datasource package from GitHub
func (gc *GitHubClient) DownloadDatasourcePackage(ctx context.Context, repoOwner, repoName, datasource, version, destPath string) error {
	packageName := strings.ReplaceAll(datasource, "_", "-")
	versionWithoutV := strings.TrimPrefix(version, "v")
	releaseTag := fmt.Sprintf("ds/%s/%s", datasource, version)
	filename := fmt.Sprintf("%s-%s.tar.gz", packageName, versionWithoutV)

	assetData, err := gc.DownloadReleaseAsset(ctx, repoOwner, repoName, releaseTag, filename)
	if err != nil {
		return err
	}

	return ExtractTarGz(bytes.NewReader(assetData), destPath)
}

// getAllReleases fetches all releases from a repository
func (gc *GitHubClient) getAllReleases(ctx context.Context, repoOwner, repoName string) ([]GitHubRelease, error) {
	releasesAPIURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", repoOwner, repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", releasesAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}

	if gc.token != "" {
		req.Header.Set("Authorization", "token "+gc.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Check for rate limit errors
		if resp.StatusCode == http.StatusForbidden {
			if isRateLimitError(resp) {
				if gc.token == "" {
					return nil, errors.New("GitHub API rate limit exceeded (60 requests/hour for unauthenticated requests). Set GITHUB_TOKEN environment variable to increase limit to 5000 requests/hour")
				}
				return nil, errors.New("GitHub API rate limit exceeded")
			}
		}
		return nil, fmt.Errorf("failed to get releases (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases response: %w", err)
	}

	return releases, nil
}

// getRelease fetches a specific release by tag
func (gc *GitHubClient) getRelease(ctx context.Context, repoOwner, repoName, releaseTag string) (*GitHubRelease, error) {
	releaseAPIURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", repoOwner, repoName, releaseTag)

	req, err := http.NewRequestWithContext(ctx, "GET", releaseAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}

	if gc.token != "" {
		req.Header.Set("Authorization", "token "+gc.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get release info: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("release not found: tag '%s' does not exist in repository %s/%s", releaseTag, repoOwner, repoName)
	case http.StatusUnauthorized:
		if gc.token != "" {
			return nil, errors.New("authentication failed - check GITHUB_TOKEN validity")
		}
		return nil, errors.New("authentication required - repository may be private or rate limit exceeded")
	case http.StatusForbidden:
		if isRateLimitError(resp) {
			if gc.token == "" {
				return nil, errors.New("GitHub API rate limit exceeded (60 requests/hour for unauthenticated requests). Set GITHUB_TOKEN environment variable to increase limit to 5000 requests/hour")
			}
			return nil, errors.New("GitHub API rate limit exceeded")
		}
		if gc.token != "" {
			return nil, fmt.Errorf("access forbidden - check GITHUB_TOKEN permissions for repository %s/%s", repoOwner, repoName)
		}
		return nil, fmt.Errorf("access forbidden - repository %s/%s may be private", repoOwner, repoName)
	case http.StatusOK:
		// continue
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get release info (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release response: %w", err)
	}

	return &release, nil
}

// downloadAsset downloads an asset from GitHub
func (gc *GitHubClient) downloadAsset(ctx context.Context, assetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset download request: %w", err)
	}

	if gc.token != "" {
		req.Header.Set("Authorization", "token "+gc.token)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := gc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Check for rate limit errors
		if resp.StatusCode == http.StatusForbidden {
			if isRateLimitError(resp) {
				if gc.token == "" {
					return nil, errors.New("GitHub API rate limit exceeded (60 requests/hour for unauthenticated requests). Set GITHUB_TOKEN environment variable to increase limit to 5000 requests/hour")
				}
				return nil, errors.New("GitHub API rate limit exceeded")
			}
		}
		return nil, fmt.Errorf("failed to download asset (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ExtractTarGz extracts tar.gz content from an io.Reader stream to the destination directory safely
func ExtractTarGz(reader io.Reader, destPath string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("not a valid gzip archive: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	extractedCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			continue // skip non-file/dir entries
		}

		targetPath := filepath.Join(destPath, header.Name) // #nosec G305
		targetPath = filepath.Clean(targetPath)

		// Security: ensure no path traversal outside destPath
		// Safe: validated against path traversal attacks below
		if !strings.HasPrefix(targetPath, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in archive (path traversal detected): %s", header.Name)
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, 0o750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", targetPath, err)
		}

		if err := extractTarFile(tarReader, targetPath); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", targetPath, err)
		}

		extractedCount++
	}

	if extractedCount == 0 {
		return errors.New("no files found in tar.gz archive")
	}
	return nil
}

// DownloadTransformModels handles the complete workflow of downloading transform models:
// - Resolves version (uses provided version or gets latest from GitHub)
// - Creates destination directory
// - Downloads and extracts the datasource package
func (gc *GitHubClient) DownloadTransformModels(ctx context.Context, datasource, modelVersion, destPath string) (string, error) {
	// Determine version: use provided version or get latest from GitHub
	var version string
	if modelVersion != "" {
		version = normalizeVersion(modelVersion)
		if version == "" {
			// Invalid version provided, fall back to latest
			utils.PrintDebug(fmt.Sprintf("Invalid version '%s' provided, fetching latest release from GitHub...", modelVersion))
			latestVersion, err := gc.GetLatestReleaseVersion(ctx, DefaultRepoOwner, DefaultRepoName, datasource)
			if err != nil {
				return "", fmt.Errorf("failed to get latest version for datasource '%s': %w", datasource, err)
			}
			version = latestVersion
			utils.PrintDebug(fmt.Sprintf("Latest version found: %s", version))
		} else {
			utils.PrintDebug(fmt.Sprintf("Using specified version: %s (normalized from: %s)", version, modelVersion))
		}
	} else {
		utils.PrintDebug(fmt.Sprintf("No version specified, fetching latest release from GitHub for %s", datasource))
		latestVersion, err := gc.GetLatestReleaseVersion(ctx, DefaultRepoOwner, DefaultRepoName, datasource)
		if err != nil {
			return "", fmt.Errorf("failed to get latest version for datasource '%s': %w", datasource, err)
		}
		version = latestVersion
		utils.PrintDebug(fmt.Sprintf("Latest version found: %s", version))
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Clean(destPath), 0o750); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	utils.PrintDebug(fmt.Sprintf("Downloading and extracting %s version %s...", datasource, version))

	// Download and extract the datasource package
	if err := gc.DownloadDatasourcePackage(ctx, DefaultRepoOwner, DefaultRepoName, datasource, version, destPath); err != nil {
		return "", fmt.Errorf("failed to download version %s: %w", version, err)
	}

	return version, nil
}

// normalizeVersion ensures version has 'v' prefix and validates format
func normalizeVersion(version string) string {
	// Reject invalid version strings
	if version == "main" || version == "master" || version == "latest" {
		// These are not valid version tags, should use latest release instead
		return ""
	}

	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// isRateLimitError checks if the HTTP response indicates a rate limit error
// GitHub returns rate limit errors with X-RateLimit-Remaining header set to 0
func isRateLimitError(resp *http.Response) bool {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	return remaining == "0"
}

// extractTarFile extracts a single file from the tar archive reader
// targetPath is validated by the caller to prevent path traversal
func extractTarFile(tarReader *tar.Reader, targetPath string) error {
	out, err := os.Create(filepath.Clean(targetPath)) // #nosec G304
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, tarReader)
	return err
}
