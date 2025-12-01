package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/ecos-labs/ecos/code/cli/plugins/types"
	"github.com/ecos-labs/ecos/code/cli/plugins/types/mocks"
	"go.uber.org/mock/gomock"
)

func TestParseTransformArgs(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantCommand      string
		wantProjectDir   string
		wantIsDryRun     bool
		wantIgnoreDrift  bool
		wantFilteredArgs []string
	}{
		{
			name:             "simple command",
			args:             []string{"run"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "command with args",
			args:             []string{"run", "--select", "my_model"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run", "--select", "my_model"},
		},
		{
			name:             "with project dir flag",
			args:             []string{"run", "--project-dir", "/path/to/project"},
			wantCommand:      "run",
			wantProjectDir:   "/path/to/project",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "with project dir shorthand",
			args:             []string{"run", "-p", "/custom/path"},
			wantCommand:      "run",
			wantProjectDir:   "/custom/path",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "with dry run flag",
			args:             []string{"run", "--dry-run"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     true,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "with ignore drift flag",
			args:             []string{"run", "--ignore-drift"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  true,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "with verbose flag",
			args:             []string{"run", "--verbose"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run"},
		},
		{
			name:             "with dbt -v flag (should pass through)",
			args:             []string{"run", "-v"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run", "-v"},
		},
		{
			name:             "complex command with multiple flags",
			args:             []string{"run", "--project-dir", "/path", "--select", "tag:daily", "--dry-run"},
			wantCommand:      "run",
			wantProjectDir:   "/path",
			wantIsDryRun:     true,
			wantIgnoreDrift:  false,
			wantFilteredArgs: []string{"run", "--select", "tag:daily"},
		},
		{
			name:             "with ignore drift and other flags",
			args:             []string{"run", "--ignore-drift", "--select", "my_model"},
			wantCommand:      "run",
			wantProjectDir:   ".",
			wantIsDryRun:     false,
			wantIgnoreDrift:  true,
			wantFilteredArgs: []string{"run", "--select", "my_model"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseTransformArgs(tt.args)

			if parsed.Command != tt.wantCommand {
				t.Errorf("Command = %q, want %q", parsed.Command, tt.wantCommand)
			}
			if parsed.ProjectDir != tt.wantProjectDir {
				t.Errorf("ProjectDir = %q, want %q", parsed.ProjectDir, tt.wantProjectDir)
			}
			if parsed.IsDryRun != tt.wantIsDryRun {
				t.Errorf("IsDryRun = %v, want %v", parsed.IsDryRun, tt.wantIsDryRun)
			}
			if parsed.IgnoreDrift != tt.wantIgnoreDrift {
				t.Errorf("IgnoreDrift = %v, want %v", parsed.IgnoreDrift, tt.wantIgnoreDrift)
			}
			if len(parsed.FilteredArgs) != len(tt.wantFilteredArgs) {
				t.Errorf("FilteredArgs length = %d, want %d", len(parsed.FilteredArgs), len(tt.wantFilteredArgs))
			} else {
				for i, arg := range parsed.FilteredArgs {
					if arg != tt.wantFilteredArgs[i] {
						t.Errorf("FilteredArgs[%d] = %q, want %q", i, arg, tt.wantFilteredArgs[i])
					}
				}
			}
		})
	}
}

func TestContainsHelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "no help flag",
			args: []string{"run", "--select", "my_model"},
			want: false,
		},
		{
			name: "has --help",
			args: []string{"run", "--help"},
			want: true,
		},
		{
			name: "has -h",
			args: []string{"-h"},
			want: true,
		},
		{
			name: "help flag in middle",
			args: []string{"run", "--help", "--select", "model"},
			want: true,
		},
		{
			name: "empty args",
			args: []string{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsHelpFlag(tt.args)
			if got != tt.want {
				t.Errorf("containsHelpFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunTransformExecute(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		config          map[string]any
		setupMock       func(*mocks.MockTransformPlugin)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "successful execution",
			args: []string{"run"},
			config: map[string]any{
				"verbose": false,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Validate(gomock.Any()).Return(nil)
				m.EXPECT().Name().Return("dbt").AnyTimes()
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&types.PluginResult{
					Success: true,
					Message: "completed",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "validation fails",
			args: []string{"run"},
			config: map[string]any{
				"verbose": false,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Validate(gomock.Any()).Return(errors.New("validation error"))
			},
			wantErr:         true,
			wantErrContains: "plugin validation failed",
		},
		{
			name: "execution fails with error",
			args: []string{"run"},
			config: map[string]any{
				"verbose": false,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Validate(gomock.Any()).Return(nil)
				m.EXPECT().Name().Return("dbt").AnyTimes()
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("execution error"))
			},
			wantErr:         true,
			wantErrContains: "failed",
		},
		{
			name: "execution returns unsuccessful result",
			args: []string{"run"},
			config: map[string]any{
				"verbose": false,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Validate(gomock.Any()).Return(nil)
				m.EXPECT().Name().Return("dbt").AnyTimes()
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&types.PluginResult{
					Success: false,
					Message: "command failed",
				}, nil)
			},
			wantErr:         true,
			wantErrContains: "failed",
		},
		{
			name: "verbose mode with multiple args",
			args: []string{"run", "--select", "my_model"},
			config: map[string]any{
				"verbose": true,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Validate(gomock.Any()).Return(nil)
				m.EXPECT().Name().Return("dbt").AnyTimes()
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&types.PluginResult{
					Success: true,
					Message: "completed",
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPlugin := mocks.NewMockTransformPlugin(ctrl)
			tt.setupMock(mockPlugin)

			err := runTransformExecute(mockPlugin, tt.args, tt.config)

			if tt.wantErr && err == nil {
				t.Errorf("runTransformExecute() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("runTransformExecute() unexpected error: %v", err)
			}
			if tt.wantErrContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("runTransformExecute() error = %q, want to contain %q", err.Error(), tt.wantErrContains)
				}
			}
		})
	}
}

func TestRunTransformDryRun(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		config    map[string]any
		setupMock func(*mocks.MockTransformPlugin)
	}{
		{
			name: "dry run with basic config",
			args: []string{"run"},
			config: map[string]any{
				"verbose": false,
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Name().Return("dbt")
				m.EXPECT().Version().Return("1.0.0")
			},
		},
		{
			name: "dry run with dbt project dir",
			args: []string{"run", "--select", "my_model"},
			config: map[string]any{
				"verbose":         true,
				"dbt_project_dir": "/path/to/project",
			},
			setupMock: func(m *mocks.MockTransformPlugin) {
				m.EXPECT().Name().Return("dbt")
				m.EXPECT().Version().Return("1.5.0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPlugin := mocks.NewMockTransformPlugin(ctrl)
			tt.setupMock(mockPlugin)

			err := runTransformDryRun(mockPlugin, tt.args, tt.config)
			if err != nil {
				t.Errorf("runTransformDryRun() unexpected error: %v", err)
			}
		})
	}
}

func TestTransformCommand(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "command is registered",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if transformCmd == nil {
					t.Error("transformCmd is nil")
				}
			},
		},
		{
			name: "command use is correct",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if transformCmd.Use != "transform [command] [args...]" {
					t.Errorf("transformCmd.Use = %q, want %q", transformCmd.Use, "transform [command] [args...]")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if transformCmd.Short == "" {
					t.Error("transformCmd.Short is empty")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if transformCmd.Long == "" {
					t.Error("transformCmd.Long is empty")
				}
			},
		},
		{
			name: "command has flag parsing disabled",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if !transformCmd.DisableFlagParsing {
					t.Error("transformCmd.DisableFlagParsing should be true")
				}
			},
		},
		{
			name: "command has silence usage enabled",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if !transformCmd.SilenceUsage {
					t.Error("transformCmd.SilenceUsage should be true")
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

func TestParsedTransformArgs(t *testing.T) {
	// Test the ParsedTransformArgs struct initialization
	parsed := &ParsedTransformArgs{
		Command:      "run",
		FilteredArgs: []string{"run", "--select", "model"},
		ProjectDir:   "/custom/path",
		IsDryRun:     true,
		IgnoreDrift:  false,
	}

	if parsed.Command != "run" {
		t.Errorf("Command = %q, want %q", parsed.Command, "run")
	}
	if parsed.ProjectDir != "/custom/path" {
		t.Errorf("ProjectDir = %q, want %q", parsed.ProjectDir, "/custom/path")
	}
	if !parsed.IsDryRun {
		t.Error("IsDryRun should be true")
	}
	if parsed.IgnoreDrift {
		t.Error("IgnoreDrift should be false")
	}
	if len(parsed.FilteredArgs) != 3 {
		t.Errorf("FilteredArgs length = %d, want 3", len(parsed.FilteredArgs))
	}
}

func TestCheckConfigDrift_NoEcosYaml(t *testing.T) {
	// Create a temporary directory without .ecos.yaml
	tmpDir := t.TempDir()

	// Should not return error when .ecos.yaml doesn't exist
	err := checkConfigDrift(tmpDir)
	if err != nil {
		t.Errorf("checkConfigDrift() should not error when .ecos.yaml missing, got: %v", err)
	}
}

func TestCheckConfigDrift_NoDrift(t *testing.T) {
	// This test would require setting up a complete .ecos.yaml and matching dbt files
	// For now, we'll test that the function exists and can be called
	tmpDir := t.TempDir()

	// Should handle missing config gracefully
	err := checkConfigDrift(tmpDir)
	if err != nil {
		t.Errorf("checkConfigDrift() unexpected error: %v", err)
	}
}

func TestTransformCommandDescription(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		contains string
	}{
		{
			name:     "short description mentions transform",
			field:    "short",
			contains: "transform",
		},
		{
			name:     "long description mentions dbt",
			field:    "long",
			contains: "dbt",
		},
		{
			name:     "long description mentions plugin",
			field:    "long",
			contains: "plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var text string
			switch tt.field {
			case "short":
				text = transformCmd.Short
			case "long":
				text = transformCmd.Long
			}

			if text == "" {
				t.Errorf("%s description is empty", tt.field)
				return
			}

			// Case-insensitive check
			if !containsIgnoreCase(text, tt.contains) {
				t.Errorf("%s description should contain %q", tt.field, tt.contains)
			}
		})
	}
}
