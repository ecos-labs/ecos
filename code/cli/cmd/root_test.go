package cmd

import (
	"testing"

	"github.com/ecos-labs/ecos/code/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "command is registered",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if rootCmd == nil {
					t.Error("rootCmd is nil")
				}
			},
		},
		{
			name: "command use is correct",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if rootCmd.Use != "ecos" {
					t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "ecos")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if rootCmd.Short == "" {
					t.Error("rootCmd.Short is empty")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if rootCmd.Long == "" {
					t.Error("rootCmd.Long is empty")
				}
			},
		},
		{
			name: "command has version set",
			checkFunc: func(t *testing.T) {
				t.Helper()
				expectedVersion := version.GetInfo().Version
				if rootCmd.Version != expectedVersion {
					t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, expectedVersion)
				}
			},
		},
		{
			name: "command has subcommands",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if !rootCmd.HasSubCommands() {
					t.Error("rootCmd should have subcommands")
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

func TestRootCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		isPersistent bool
		checkDefault func(t *testing.T)
	}{
		{
			name:         "config flag exists",
			flagName:     "config",
			shorthand:    "",
			isPersistent: true,
			checkDefault: func(t *testing.T) {
				t.Helper()
				flag := rootCmd.PersistentFlags().Lookup("config")
				if flag == nil {
					t.Error("config flag not found")
					return
				}
				if flag.DefValue != "" {
					t.Errorf("config flag default = %q, want empty string", flag.DefValue)
				}
			},
		},
		{
			name:         "dry-run flag exists",
			flagName:     "dry-run",
			shorthand:    "",
			isPersistent: true,
			checkDefault: func(t *testing.T) {
				t.Helper()
				flag := rootCmd.PersistentFlags().Lookup("dry-run")
				if flag == nil {
					t.Error("dry-run flag not found")
					return
				}
				if flag.DefValue != "false" {
					t.Errorf("dry-run flag default = %q, want %q", flag.DefValue, "false")
				}
			},
		},
		{
			name:         "verbose flag exists",
			flagName:     "verbose",
			shorthand:    "",
			isPersistent: true,
			checkDefault: func(t *testing.T) {
				t.Helper()
				flag := rootCmd.PersistentFlags().Lookup("verbose")
				if flag == nil {
					t.Error("verbose flag not found")
					return
				}
				if flag.DefValue != "false" {
					t.Errorf("verbose flag default = %q, want %q", flag.DefValue, "false")
				}
			},
		},
		{
			name:         "version flag exists",
			flagName:     "version",
			shorthand:    "",
			isPersistent: false,
			checkDefault: func(t *testing.T) {
				t.Helper()
				flag := rootCmd.Flags().Lookup("version")
				if flag == nil {
					t.Error("version flag not found")
					return
				}
				if flag.DefValue != "false" {
					t.Errorf("version flag default = %q, want %q", flag.DefValue, "false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flag *pflag.Flag
			if tt.isPersistent {
				flag = rootCmd.PersistentFlags().Lookup(tt.flagName)
			} else {
				flag = rootCmd.Flags().Lookup(tt.flagName)
			}

			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if tt.checkDefault != nil {
				tt.checkDefault(t)
			}
		})
	}
}

func TestIsVerbose(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     bool
	}{
		{
			name: "verbose is false by default",
			setup: func() {
				verbose = false
			},
			teardown: func() {
				verbose = false
			},
			want: false,
		},
		{
			name: "verbose is true when set",
			setup: func() {
				verbose = true
			},
			teardown: func() {
				verbose = false
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			got := IsVerbose()
			if got != tt.want {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDryRun(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     bool
	}{
		{
			name: "dry-run is false by default",
			setup: func() {
				dryRun = false
			},
			teardown: func() {
				dryRun = false
			},
			want: false,
		},
		{
			name: "dry-run is true when set",
			setup: func() {
				dryRun = true
			},
			teardown: func() {
				dryRun = false
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			got := IsDryRun()
			if got != tt.want {
				t.Errorf("IsDryRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		want  bool // true if config should be nil
	}{
		{
			name: "returns nil when config not loaded",
			setup: func() {
				cfg = nil
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got := GetConfig()
			isNil := got == nil

			if isNil != tt.want {
				t.Errorf("GetConfig() nil = %v, want nil = %v", isNil, tt.want)
			}
		})
	}
}

func TestRootCommandSubcommands(t *testing.T) {
	expectedCommands := []string{
		"init",
		"transform",
		"version",
		// Add other expected subcommands here
	}

	for _, cmdName := range expectedCommands {
		t.Run("has "+cmdName+" subcommand", func(t *testing.T) {
			found := false
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == cmdName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected subcommand %q not found", cmdName)
			}
		})
	}
}

func TestRootCommandPersistentPreRunE(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		setup       func()
		teardown    func()
		wantErr     bool
	}{
		{
			name:        "init command skips config loading",
			commandName: "init",
			setup: func() {
				verbose = false
			},
			teardown: func() {
				verbose = false
			},
			wantErr: false,
		},
		{
			name:        "plugins command skips config loading",
			commandName: "plugins",
			setup: func() {
				verbose = false
			},
			teardown: func() {
				verbose = false
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
			}

			// Create a mock command with the specified name
			mockCmd := &cobra.Command{
				Use: tt.commandName,
			}

			err := rootCmd.PersistentPreRunE(mockCmd, []string{})

			if tt.wantErr && err == nil {
				t.Error("PersistentPreRunE() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("PersistentPreRunE() unexpected error: %v", err)
			}
		})
	}
}

func TestRootCommandDescription(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		contains string
	}{
		{
			name:     "short description mentions CLI",
			field:    "short",
			contains: "CLI",
		},
		{
			name:     "short description mentions FinOps",
			field:    "short",
			contains: "FinOps",
		},
		{
			name:     "long description mentions plugin",
			field:    "long",
			contains: "plugin",
		},
		{
			name:     "long description mentions dbt",
			field:    "long",
			contains: "dbt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var text string
			switch tt.field {
			case "short":
				text = rootCmd.Short
			case "long":
				text = rootCmd.Long
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

// Helper function for case-insensitive string contains check
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

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
