package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ecos-labs/ecos/code/cli/plugins/types/mocks"
	"github.com/spf13/cobra"
	"go.uber.org/mock/gomock"
)

func TestRunInitExecute(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*mocks.MockInitPlugin)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "successful execution",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(nil),
					m.EXPECT().CreateResources().Return(nil),
					m.EXPECT().DownloadTransformModels().Return("v1.0.0", nil),
					m.EXPECT().SetModelVersion("v1.0.0").Return(nil),
					m.EXPECT().GenerateConfig().Return(nil),
					m.EXPECT().PostInitSummary().Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name: "directory structure creation fails",
			setupMock: func(m *mocks.MockInitPlugin) {
				m.EXPECT().CreateDirectoryStructure().Return(errors.New("failed to create directory"))
			},
			wantErr:         true,
			wantErrContains: "failed to create directory structure",
		},
		{
			name: "base files initialization fails",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(errors.New("failed to create files")),
				)
			},
			wantErr:         true,
			wantErrContains: "failed to create base files",
		},
		{
			name: "resource creation fails but continues",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(nil),
					m.EXPECT().CreateResources().Return(errors.New("resource creation failed")),
					m.EXPECT().PostInitSummary().Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name: "transform models download fails but continues",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(nil),
					m.EXPECT().CreateResources().Return(nil),
					m.EXPECT().DownloadTransformModels().Return("", errors.New("download failed")),
					m.EXPECT().PostInitSummary().Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name: "config generation fails",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(nil),
					m.EXPECT().CreateResources().Return(nil),
					m.EXPECT().DownloadTransformModels().Return("v1.0.0", nil),
					m.EXPECT().SetModelVersion("v1.0.0").Return(nil),
					m.EXPECT().GenerateConfig().Return(errors.New("config generation failed")),
				)
			},
			wantErr:         true,
			wantErrContains: "failed to generate config",
		},
		{
			name: "set model version fails but continues",
			setupMock: func(m *mocks.MockInitPlugin) {
				gomock.InOrder(
					m.EXPECT().CreateDirectoryStructure().Return(nil),
					m.EXPECT().InitializeBaseFiles().Return(nil),
					m.EXPECT().CreateResources().Return(nil),
					m.EXPECT().DownloadTransformModels().Return("v1.0.0", nil),
					m.EXPECT().SetModelVersion("v1.0.0").Return(errors.New("version set failed")),
					m.EXPECT().GenerateConfig().Return(nil),
					m.EXPECT().PostInitSummary().Return(nil),
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPlugin := mocks.NewMockInitPlugin(ctrl)
			tt.setupMock(mockPlugin)

			err := runInitExecute(mockPlugin)

			if tt.wantErr && err == nil {
				t.Errorf("runInitExecute() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("runInitExecute() unexpected error: %v", err)
			}
			if tt.wantErrContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("runInitExecute() error = %q, want to contain %q", err.Error(), tt.wantErrContains)
				}
			}
		})
	}
}

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "command is registered",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if initCmd == nil {
					t.Error("initCmd is nil")
				}
			},
		},
		{
			name: "command use is correct",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if initCmd.Use != "init" {
					t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if initCmd.Short == "" {
					t.Error("initCmd.Short is empty")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if initCmd.Long == "" {
					t.Error("initCmd.Long is empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t)
		})
	}
}

func TestInitCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue interface{}
		checkDefault func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name:      "force flag exists",
			flagName:  "force",
			shorthand: "f",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetBool("force")
				if err != nil {
					t.Errorf("failed to get force flag: %v", err)
				}
				if val != false {
					t.Errorf("force flag default = %v, want false", val)
				}
			},
		},
		{
			name:      "output flag exists",
			flagName:  "output",
			shorthand: "o",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetString("output")
				if err != nil {
					t.Errorf("failed to get output flag: %v", err)
				}
				if val != "." {
					t.Errorf("output flag default = %q, want %q", val, ".")
				}
			},
		},
		{
			name:      "source flag exists",
			flagName:  "source",
			shorthand: "s",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetString("source")
				if err != nil {
					t.Errorf("failed to get source flag: %v", err)
				}
				if val != "" {
					t.Errorf("source flag default = %q, want empty string", val)
				}
			},
		},
		{
			name:      "model-version flag exists",
			flagName:  "model-version",
			shorthand: "m",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetString("model-version")
				if err != nil {
					t.Errorf("failed to get model-version flag: %v", err)
				}
				if val != "latest" {
					t.Errorf("model-version flag default = %q, want %q", val, "latest")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := initCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
				return
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
			if tt.checkDefault != nil {
				tt.checkDefault(t, initCmd)
			}
		})
	}
}

// Test helper functions
func TestCreateTempTestDir(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	// Verify directory exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("temp directory was not created")
	}

	// Verify it's a directory
	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("failed to stat temp dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("temp path is not a directory")
	}
}

func TestCreateTestEcosConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ".ecos.yaml")
	content := []byte("project_name: test\n")

	err := os.WriteFile(configPath, content, 0o600)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created")
	}

	// Verify content
	readContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("config content = %q, want %q", string(readContent), string(content))
	}
}
