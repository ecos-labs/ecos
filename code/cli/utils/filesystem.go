package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateDirectory creates a directory if it doesn't exist
func CreateDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirectoryExists checks if a directory exists
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// WriteFile writes content to a file, creating directories if necessary
func WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := CreateDirectory(dir); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o600)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Clean paths to prevent directory traversal
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	sourceFile, err := os.ReadFile(cleanSrc) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}

	return WriteFile(cleanDst, sourceFile)
}

// RemoveFile removes a file if it exists
func RemoveFile(path string) error {
	if FileExists(path) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", path, err)
		}
	}
	return nil
}

// GetWorkingDirectory returns the current working directory
func GetWorkingDirectory() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return pwd, nil
}

// GetHomeDirectory returns the user's home directory
func GetHomeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// EnsureDirectoryStructure creates the standard ecos directory structure
func EnsureDirectoryStructure(basePath string) error {
	directories := []string{
		filepath.Join(basePath, "plugins"),
		filepath.Join(basePath, "plugins", "ingest"),
		filepath.Join(basePath, "plugins", "transform"),
		filepath.Join(basePath, "models"),
		filepath.Join(basePath, "logs"),
		filepath.Join(basePath, "output"),
	}

	for _, dir := range directories {
		if err := CreateDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory structure: %w", err)
		}
	}

	return nil
}
