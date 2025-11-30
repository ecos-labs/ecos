package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// canonical returns a fully-resolved, absolute, symlink-free version of the path.
func canonical(p string) string {
	resolved, err := filepath.EvalSymlinks(p)
	if err == nil {
		p = resolved
	}
	abs, err := filepath.Abs(p)
	if err == nil {
		return abs
	}
	return p // last resort
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	badFile := filepath.Join(tmp, ".ecos.yaml")

	if err := os.WriteFile(badFile, []byte("{invalid_yaml:"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadConfig(badFile)
	if err == nil {
		t.Fatalf("expected YAML unmarshal error but got none")
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, ".ecos.yaml")

	content := `
project_name: my-test
transform:
  dbt:
    project_dir: "./x"
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectName != "my-test" {
		t.Errorf("ProjectName = %q, want %q", cfg.ProjectName, "my-test")
	}
}

func TestFindConfigFile_Found(t *testing.T) {
	root := t.TempDir()

	// Create a nested path
	nested := filepath.Join(root, "a/b/c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Config file lives at root/.ecos.yaml
	cfgPath := filepath.Join(root, ".ecos.yaml")
	if err := os.WriteFile(cfgPath, []byte("project_name: test"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Change working directory to nested folder
	t.Chdir(nested)

	found, err := FindConfigFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if canonical(found) != canonical(cfgPath) {
		t.Errorf("FindConfigFile = %q, want %q", found, cfgPath)
	}
}

func TestFindConfigFile_NotFound(t *testing.T) {
	tmp := t.TempDir()

	t.Chdir(tmp)

	_, err := FindConfigFile()
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}

	if !strings.Contains(err.Error(), ".ecos.yaml not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, ".ecos.yaml")
	if err := os.WriteFile(cfg, []byte("x: y"), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Chdir(tmp)

	got := GetConfigFilePath()

	if canonical(got) != canonical(cfg) {
		t.Errorf("GetConfigFilePath = %q, want %q", got, cfg)
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.ProjectName == "" ||
		cfg.ModelVersion == "" ||
		cfg.Transform.DBT.ProjectDir == "" ||
		cfg.Report.OutputPath == "" {
		t.Error("default config should populate required fields")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &EcosConfig{}
	cfg.SetDefaults()

	if cfg.ModelVersion != "latest" {
		t.Errorf("default model_version = %q, want 'latest'", cfg.ModelVersion)
	}
	if cfg.Transform.Plugin != "dbt" {
		t.Errorf("default transform.plugin = %q, want 'dbt'", cfg.Transform.Plugin)
	}
}

func TestValidate_GlobalConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	// Valid levels
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg.Global.LogLevel = level
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected valid log level %q, got error %v", level, err)
		}
	}

	// Invalid level
	cfg.Global.LogLevel = "badlevel"
	err := cfg.Validate()
	if err == nil {
		t.Errorf("expected error for invalid log level, got none")
	}
}

func TestValidate_TransformConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	// Valid
	cfg.Transform.Plugin = "dbt"
	cfg.Transform.DBT.ProjectDir = "x"
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid dbt config, got %v", err)
	}

	// Invalid: SQL plugin without connection string
	cfg.Transform.Plugin = "sql"
	cfg.Transform.SQL.ConnectionString = ""
	err := cfg.Validate()
	if err == nil {
		t.Errorf("expected error for missing SQL connection string")
	}
}
