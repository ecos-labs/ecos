package transform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ecos-labs/ecos/code/cli/plugins/types"
)

func TestDBTTransformPlugin_Metadata(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{
			name:     "Name returns correct value",
			method:   plugin.Name,
			expected: "dbt",
		},
		{
			name:     "Version returns correct value",
			method:   plugin.Version,
			expected: "1.0.0",
		},
		{
			name:     "Author returns correct value",
			method:   plugin.Author,
			expected: "ecos team",
		},
		{
			name:     "TransformEngine returns correct value",
			method:   plugin.TransformEngine,
			expected: "dbt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDBTTransformPlugin_Type(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	got := plugin.Type()
	want := types.PluginTypeTransform

	if got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
}

func TestDBTTransformPlugin_IsCore(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	if !plugin.IsCore() {
		t.Error("IsCore() should return true for core plugin")
	}
}

func TestDBTTransformPlugin_Description(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	desc := plugin.Description()

	if desc == "" {
		t.Error("Description() should not be empty")
	}

	// Check for key terms
	expectedTerms := []string{"dbt", "transform"}
	for _, term := range expectedTerms {
		if !contains(desc, term) {
			t.Errorf("Description should contain %q", term)
		}
	}
}

func TestDBTTransformPlugin_Documentation(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	docs := plugin.Documentation()

	if docs == "" {
		t.Error("Documentation() should not be empty")
	}
}

func TestDBTTransformPlugin_GetSupportedCommands(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	commands := plugin.GetSupportedCommands()

	if len(commands) == 0 {
		t.Error("GetSupportedCommands() should return at least one command")
	}

	// Check for essential commands
	essentialCommands := []string{"run", "test", "compile", "deps", "seed"}
	for _, cmd := range essentialCommands {
		found := false
		for _, supported := range commands {
			if supported == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedCommands() should include %q", cmd)
		}
	}
}

func TestDBTTransformPlugin_GetProjectPath(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		want       string
	}{
		{
			name:       "returns custom project dir when set",
			projectDir: "/custom/path",
			want:       "/custom/path",
		},
		{
			name:       "returns default path when not set",
			projectDir: "",
			want:       "transform/dbt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &DBTTransformPlugin{
				ProjectDir: tt.projectDir,
			}

			got := plugin.GetProjectPath()
			if got != tt.want {
				t.Errorf("GetProjectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDBTTransformPlugin_GetProjectDir(t *testing.T) {
	tests := []struct {
		name   string
		plugin *DBTTransformPlugin
		config map[string]any
		want   string
	}{
		{
			name:   "uses dbt_project_dir from config",
			plugin: &DBTTransformPlugin{},
			config: map[string]any{
				"dbt_project_dir": "/explicit/dbt/path",
			},
			want: "/explicit/dbt/path",
		},
		{
			name:   "derives from project_dir",
			plugin: &DBTTransformPlugin{},
			config: map[string]any{
				"project_dir": "/project",
			},
			want: "/project/transform/dbt",
		},
		{
			name: "uses plugin ProjectDir as fallback",
			plugin: &DBTTransformPlugin{
				ProjectDir: "/fallback/path",
			},
			config: map[string]any{},
			want:   "/fallback/path",
		},
		{
			name:   "empty config returns empty string",
			plugin: &DBTTransformPlugin{},
			config: map[string]any{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugin.getProjectDir(tt.config)
			if got != tt.want {
				t.Errorf("getProjectDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDBTTransformPlugin_BuildSuccessMessage(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name     string
		command  string
		args     []string
		contains []string
	}{
		{
			name:     "simple command without args",
			command:  "run",
			args:     []string{},
			contains: []string{"dbt run", "completed successfully"},
		},
		{
			name:     "command with select flag",
			command:  "run",
			args:     []string{"--select", "my_model"},
			contains: []string{"dbt run", "completed successfully", "--select my_model"},
		},
		{
			name:     "command with exclude flag",
			command:  "run",
			args:     []string{"--exclude", "staging"},
			contains: []string{"dbt run", "completed successfully", "--exclude staging"},
		},
		{
			name:     "command with full-refresh",
			command:  "run",
			args:     []string{"--full-refresh"},
			contains: []string{"dbt run", "completed successfully", "--full-refresh"},
		},
		{
			name:     "command with target",
			command:  "run",
			args:     []string{"--target", "prod"},
			contains: []string{"dbt run", "completed successfully", "--target prod"},
		},
		{
			name:     "command with multiple flags",
			command:  "run",
			args:     []string{"--select", "my_model", "--full-refresh", "--target", "prod"},
			contains: []string{"dbt run", "completed successfully", "--select my_model", "--full-refresh", "--target prod"},
		},
		{
			name:     "test command",
			command:  "test",
			args:     []string{"--models", "tag:daily"},
			contains: []string{"dbt test", "completed successfully", "--models tag:daily"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plugin.buildSuccessMessage(tt.command, tt.args)

			for _, substr := range tt.contains {
				if !contains(got, substr) {
					t.Errorf("buildSuccessMessage() should contain %q, got %q", substr, got)
				}
			}
		})
	}
}

func TestDBTTransformPlugin_NeedsDependencyInstall(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		want      bool
	}{
		{
			name: "needs install when lock file missing",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				// Return a non-existent directory
				return "/tmp/nonexistent-dbt-project"
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &DBTTransformPlugin{}
			projectDir := tt.setupFunc(t)

			got := plugin.needsDependencyInstall(projectDir)
			if got != tt.want {
				t.Errorf("needsDependencyInstall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBTTransformPlugin_ShowCommandDescription(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	commands := []string{"run", "test", "seed", "compile", "docs", "deps", "unknown"}

	for _, cmd := range commands {
		t.Run("shows description for "+cmd, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			// Actual output is printed to stdout, so we can't easily test it
			plugin.showCommandDescription(cmd)
		})
	}
}

func TestDBTTransformPlugin_ShowCommonFlags(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	commands := []string{"run", "test", "seed", "compile", "docs", "deps", "unknown"}

	for _, cmd := range commands {
		t.Run("shows flags for "+cmd, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			plugin.showCommonFlags(cmd)
		})
	}
}

func TestDBTTransformPlugin_ShowCommandExamples(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	commands := []string{"run", "test", "seed", "compile", "docs", "deps", "unknown"}

	for _, cmd := range commands {
		t.Run("shows examples for "+cmd, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			plugin.showCommandExamples(cmd)
		})
	}
}

func TestDBTTransformPlugin_BuildConfig(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name       string
		ecosConfig any
		command    string
		args       []string
		wantKeys   []string
	}{
		{
			name:       "builds basic config",
			ecosConfig: nil,
			command:    "run",
			args:       []string{"--select", "my_model"},
			wantKeys:   []string{"command", "args"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plugin.BuildConfig(tt.ecosConfig, tt.command, tt.args)

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("BuildConfig() missing key %q", key)
				}
			}

			if cmd, ok := got["command"].(string); ok {
				if cmd != tt.command {
					t.Errorf("BuildConfig() command = %q, want %q", cmd, tt.command)
				}
			}

			if args, ok := got["args"].([]string); ok {
				if len(args) != len(tt.args) {
					t.Errorf("BuildConfig() args length = %d, want %d", len(args), len(tt.args))
				}
			}
		})
	}
}

func TestDBTTransformPlugin_Validate(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "validates with non-existent project",
			config: map[string]any{
				"dbt_project_dir": "/tmp/nonexistent-dbt-project",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Validate(tt.config)

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestDBTTransformPlugin_ValidateEnvironment(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "fails when dbt_project.yml missing",
			config: map[string]any{
				"dbt_project_dir": "/tmp/nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.ValidateEnvironment(tt.config)

			if tt.wantErr && err == nil {
				t.Error("ValidateEnvironment() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateEnvironment() unexpected error: %v", err)
			}
		})
	}
}

func TestDBTTransformPlugin_ValidateProjectStructure(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "fails when project structure invalid",
			config: map[string]any{
				"dbt_project_dir": "/tmp/nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.ValidateProjectStructure(tt.config)

			if tt.wantErr && err == nil {
				t.Error("ValidateProjectStructure() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateProjectStructure() unexpected error: %v", err)
			}
		})
	}
}

func TestDBTTransformPlugin_BuildConfig_WithEcosConfig(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	// Create a mock ecos config
	type mockDBTConfig struct {
		Profile    string
		Target     string
		ProjectDir string
		Vars       map[string]any
	}

	type mockTransformConfig struct {
		Plugin string
		DBT    mockDBTConfig
		Config map[string]any
	}

	type mockEcosConfig struct {
		Transform mockTransformConfig
	}

	ecosConfig := &mockEcosConfig{
		Transform: mockTransformConfig{
			Plugin: "dbt",
			DBT: mockDBTConfig{
				Profile:    "my_profile",
				Target:     "prod",
				ProjectDir: "/custom/dbt/path",
				Vars: map[string]any{
					"key1": "value1",
				},
			},
			Config: map[string]any{
				"custom_key": "custom_value",
			},
		},
	}

	config := plugin.BuildConfig(ecosConfig, "run", []string{"--select", "my_model"})

	// Verify command and args are set
	if cmd, ok := config["command"].(string); !ok || cmd != "run" {
		t.Errorf("command = %v, want %q", config["command"], "run")
	}

	if args, ok := config["args"].([]string); !ok || len(args) != 2 {
		t.Errorf("args = %v, want 2 elements", config["args"])
	}

	// Note: The actual BuildConfig uses type assertion to *ecosconfig.EcosConfig
	// which won't match our mock, so these won't be set. This tests the fallback behavior.
}

func TestDBTTransformPlugin_PostTransformSummary(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name    string
		command string
		config  map[string]any
	}{
		{
			name:    "basic summary",
			command: "run",
			config: map[string]any{
				"dbt_project_dir": "/test/path",
			},
		},
		{
			name:    "summary with args",
			command: "test",
			config: map[string]any{
				"dbt_project_dir": "/test/path",
				"args":            []string{"--select", "my_model"},
			},
		},
		{
			name:    "summary with target",
			command: "run",
			config: map[string]any{
				"dbt_project_dir": "/test/path",
				"target":          "prod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.PostTransformSummary(tt.command, tt.config)
			if err != nil {
				t.Errorf("PostTransformSummary() unexpected error: %v", err)
			}
		})
	}
}

func TestDBTTransformPlugin_ShowCommandHelp(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	commands := []string{"run", "test", "seed", "compile", "docs", "deps", "unknown"}

	for _, cmd := range commands {
		t.Run("shows help for "+cmd, func(t *testing.T) {
			err := plugin.ShowCommandHelp(cmd)
			if err != nil {
				t.Errorf("ShowCommandHelp(%q) unexpected error: %v", cmd, err)
			}
		})
	}
}

func TestDBTTransformPlugin_LoadEnvironmentFile(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (string, func())
		wantErr   bool
	}{
		{
			name: "no .env file returns nil",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir, cleanup := tt.setupFunc(t)
			defer cleanup()

			config := map[string]any{
				"dbt_project_dir": projectDir,
			}

			err := plugin.loadEnvironmentFile(config)

			if tt.wantErr && err == nil {
				t.Error("loadEnvironmentFile() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("loadEnvironmentFile() unexpected error: %v", err)
			}
		})
	}
}

func TestDBTTransformPlugin_ValidateEnvironment_WithValidProject(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	// Create a temporary directory with dbt_project.yml
	tmpDir := t.TempDir()

	// Create a dbt_project.yml file
	dbtProjectPath := filepath.Join(tmpDir, "dbt_project.yml")
	content := []byte("name: test_project\nversion: '1.0.0'\n")
	if err := os.WriteFile(dbtProjectPath, content, 0o600); err != nil {
		t.Fatalf("failed to create dbt_project.yml: %v", err)
	}

	config := map[string]any{
		"dbt_project_dir": tmpDir,
	}

	err := plugin.ValidateEnvironment(config)
	if err != nil {
		t.Errorf("ValidateEnvironment() unexpected error with valid project: %v", err)
	}
}

func TestDBTTransformPlugin_ValidateProjectStructure_WithValidProject(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	// Create a temporary directory with required structure
	tmpDir := t.TempDir()

	// Create required files and directories
	dbtProjectPath := filepath.Join(tmpDir, "dbt_project.yml")
	if err := os.WriteFile(dbtProjectPath, []byte("name: test\n"), 0o600); err != nil {
		t.Fatalf("failed to create dbt_project.yml: %v", err)
	}

	modelsDir := filepath.Join(tmpDir, "models")
	if err := os.MkdirAll(modelsDir, 0o750); err != nil {
		t.Fatalf("failed to create models dir: %v", err)
	}

	seedsDir := filepath.Join(tmpDir, "seeds")
	if err := os.MkdirAll(seedsDir, 0o750); err != nil {
		t.Fatalf("failed to create seeds dir: %v", err)
	}

	config := map[string]any{
		"dbt_project_dir": tmpDir,
	}

	err := plugin.ValidateProjectStructure(config)
	if err != nil {
		t.Errorf("ValidateProjectStructure() unexpected error with valid structure: %v", err)
	}
}

func TestDBTTransformPlugin_BuildSuccessMessage_EdgeCases(t *testing.T) {
	plugin := &DBTTransformPlugin{}

	tests := []struct {
		name     string
		command  string
		args     []string
		contains []string
	}{
		{
			name:     "handles -s shorthand for select",
			command:  "run",
			args:     []string{"-s", "my_model"},
			contains: []string{"dbt run", "completed successfully", "--select my_model"},
		},
		{
			name:     "handles -t shorthand for target",
			command:  "run",
			args:     []string{"-t", "prod"},
			contains: []string{"dbt run", "completed successfully", "--target prod"},
		},
		{
			name:     "handles -m shorthand for models",
			command:  "test",
			args:     []string{"-m", "tag:daily"},
			contains: []string{"dbt test", "completed successfully", "--models tag:daily"},
		},
		{
			name:     "handles vars flag",
			command:  "run",
			args:     []string{"--vars", `{"key":"value"}`},
			contains: []string{"dbt run", "completed successfully", "--vars"},
		},
		{
			name:     "handles threads flag",
			command:  "run",
			args:     []string{"--threads", "4"},
			contains: []string{"dbt run", "completed successfully", "--threads 4"},
		},
		{
			name:     "ignores unknown flags",
			command:  "run",
			args:     []string{"--unknown-flag", "value"},
			contains: []string{"dbt run", "completed successfully"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plugin.buildSuccessMessage(tt.command, tt.args)

			for _, substr := range tt.contains {
				if !contains(got, substr) {
					t.Errorf("buildSuccessMessage() should contain %q, got %q", substr, got)
				}
			}
		})
	}
}

// Helper functions

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
