package utils

import (
	"net/http"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "adds v prefix when missing",
			version: "1.0.0",
			want:    "v1.0.0",
		},
		{
			name:    "keeps v prefix when present",
			version: "v1.0.0",
			want:    "v1.0.0",
		},
		{
			name:    "handles version with patch",
			version: "1.2.3",
			want:    "v1.2.3",
		},
		{
			name:    "handles version with v prefix and patch",
			version: "v2.5.1",
			want:    "v2.5.1",
		},
		{
			name:    "rejects 'latest' keyword returns empty",
			version: "latest",
			want:    "",
		},
		{
			name:    "rejects 'main' branch returns empty",
			version: "main",
			want:    "",
		},
		{
			name:    "rejects 'master' branch returns empty",
			version: "master",
			want:    "",
		},
		{
			name:    "handles empty string adds v prefix",
			version: "",
			want:    "v",
		},
		{
			name:    "handles version with leading zeros",
			version: "01.02.03",
			want:    "v01.02.03",
		},
		{
			name:    "handles version with beta suffix",
			version: "1.0.0-beta",
			want:    "v1.0.0-beta",
		},
		{
			name:    "handles version with rc suffix",
			version: "v2.0.0-rc1",
			want:    "v2.0.0-rc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVersion(tt.version)
			if got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestNewGitHubClient(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(*testing.T)
		wantErr     bool
		wantToken   string
		description string
	}{
		{
			name: "creates client without token",
			setupEnv: func(t *testing.T) {
				t.Helper()
				// Don't set GITHUB_TOKEN - it should be unset
				// t.Setenv doesn't support unsetting, so we rely on it not being set
			},
			wantErr:     false,
			wantToken:   "",
			description: "should succeed when GITHUB_TOKEN is not set",
		},
		{
			name: "creates client with token",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("GITHUB_TOKEN", "test-token-123")
			},
			wantErr:     false,
			wantToken:   "test-token-123",
			description: "should succeed when GITHUB_TOKEN is set",
		},
		{
			name: "creates client with empty token",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("GITHUB_TOKEN", "")
			},
			wantErr:     false,
			wantToken:   "",
			description: "should succeed when GITHUB_TOKEN is set to empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv(t)

			// Create client
			client, err := NewGitHubClient()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGitHubClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected no error, verify client properties
			if !tt.wantErr {
				if client == nil {
					t.Error("NewGitHubClient() returned nil client")
					return
				}

				if client.client == nil {
					t.Error("NewGitHubClient() client.http.Client is nil")
				}

				if client.token != tt.wantToken {
					t.Errorf("NewGitHubClient() token = %q, want %q", client.token, tt.wantToken)
				}
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedResult bool
		description    string
	}{
		{
			name: "detects rate limit when remaining is 0",
			headers: map[string]string{
				"X-RateLimit-Remaining": "0",
			},
			expectedResult: true,
			description:    "should return true when X-RateLimit-Remaining is 0",
		},
		{
			name: "does not detect rate limit when remaining is greater than 0",
			headers: map[string]string{
				"X-RateLimit-Remaining": "10",
			},
			expectedResult: false,
			description:    "should return false when X-RateLimit-Remaining is greater than 0",
		},
		{
			name: "does not detect rate limit when header is missing",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expectedResult: false,
			description:    "should return false when X-RateLimit-Remaining header is missing",
		},
		{
			name:           "does not detect rate limit when no headers",
			headers:        map[string]string{},
			expectedResult: false,
			description:    "should return false when no headers are present",
		},
		{
			name: "does not detect rate limit when remaining is empty string",
			headers: map[string]string{
				"X-RateLimit-Remaining": "",
			},
			expectedResult: false,
			description:    "should return false when X-RateLimit-Remaining is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP response
			resp := &http.Response{
				Header: make(http.Header),
			}

			// Set headers
			for key, value := range tt.headers {
				resp.Header.Set(key, value)
			}

			// Test the function
			result := isRateLimitError(resp)

			if result != tt.expectedResult {
				t.Errorf("isRateLimitError() = %v, want %v. %s", result, tt.expectedResult, tt.description)
			}
		})
	}
}
