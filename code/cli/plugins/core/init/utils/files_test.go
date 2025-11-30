package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Run SetupDirectories
	err := SetupDirectories(tmpDir)
	if err != nil {
		t.Fatalf("SetupDirectories() error = %v", err)
	}

	// Verify expected directories were created
	expectedDirs := []string{
		"plugins/ingest",
		"plugins/transform",
		"transform/dbt",
		"logs",
		"output",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("directory %q was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

func TestSetupDirectories_ExistingDirs(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Pre-create one of the directories
	preExistingDir := filepath.Join(tmpDir, "logs")
	if err := os.MkdirAll(preExistingDir, 0o750); err != nil {
		t.Fatalf("failed to create pre-existing dir: %v", err)
	}

	// Run SetupDirectories - should not fail on existing directories
	err := SetupDirectories(tmpDir)
	if err != nil {
		t.Fatalf("SetupDirectories() should not fail on existing directories: %v", err)
	}

	// Verify the pre-existing directory still exists
	if _, err := os.Stat(preExistingDir); err != nil {
		t.Errorf("pre-existing directory was removed: %v", err)
	}
}

func TestSetupBaseFiles(t *testing.T) {
	tests := []struct {
		name           string
		dataSourceName string
		expectedFiles  []string
	}{
		{
			name:           "creates files for AWS CUR",
			dataSourceName: "AWS CUR",
			expectedFiles:  []string{".gitignore", "README.md"},
		},
		{
			name:           "creates files for GCP",
			dataSourceName: "GCP Billing",
			expectedFiles:  []string{".gitignore", "README.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()

			// Run SetupBaseFiles
			err := SetupBaseFiles(tmpDir, tt.dataSourceName)
			if err != nil {
				t.Fatalf("SetupBaseFiles() error = %v", err)
			}

			// Verify expected files were created
			for _, file := range tt.expectedFiles {
				fullPath := filepath.Join(tmpDir, file)
				info, err := os.Stat(fullPath)
				if err != nil {
					t.Errorf("file %q was not created: %v", file, err)
					continue
				}
				if info.IsDir() {
					t.Errorf("%q should be a file, not a directory", file)
				}
				if info.Size() == 0 {
					t.Errorf("file %q is empty", file)
				}
			}

			// Verify .gitignore contains expected patterns
			gitignorePath := filepath.Join(tmpDir, ".gitignore")
			content, err := os.ReadFile(gitignorePath)
			if err != nil {
				t.Fatalf("failed to read .gitignore: %v", err)
			}

			expectedPatterns := []string{"logs/", "output/", ".DS_Store"}
			for _, pattern := range expectedPatterns {
				if !contains(string(content), pattern) {
					t.Errorf(".gitignore should contain pattern %q", pattern)
				}
			}

			// Verify README.md contains data source name
			readmePath := filepath.Join(tmpDir, "README.md")
			readmeContent, err := os.ReadFile(readmePath)
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}

			if !contains(string(readmeContent), tt.dataSourceName) {
				t.Errorf("README.md should contain data source name %q", tt.dataSourceName)
			}
		})
	}
}

func TestSetupBaseFiles_SkipsExisting(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create an existing .gitignore with different content
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	existingContent := []byte("# Old content\n")
	if err := os.WriteFile(gitignorePath, existingContent, 0o600); err != nil {
		t.Fatalf("failed to create existing .gitignore: %v", err)
	}

	// Run SetupBaseFiles
	err := SetupBaseFiles(tmpDir, "Test Source")
	if err != nil {
		t.Fatalf("SetupBaseFiles() error = %v", err)
	}

	// Verify the file was NOT overwritten (SetupBaseFiles skips existing files)
	newContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if string(newContent) != string(existingContent) {
		t.Error(".gitignore should not be overwritten when it already exists")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
