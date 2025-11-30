package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfigCommand(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "command is registered",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if configCmd == nil {
					t.Error("configCmd is nil")
				}
			},
		},
		{
			name: "command use is correct",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if configCmd.Use != "config" {
					t.Errorf("configCmd.Use = %q, want %q", configCmd.Use, "config")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if configCmd.Short == "" {
					t.Error("configCmd.Short is empty")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if configCmd.Long == "" {
					t.Error("configCmd.Long is empty")
				}
			},
		},
		{
			name: "command has subcommands",
			checkFunc: func(t *testing.T) {
				t.Helper()
				if !configCmd.HasSubCommands() {
					t.Error("configCmd should have subcommands")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			tt.checkFunc(t)
		})
	}
}

// testConfigSubcommand validates a config subcommand has the expected structure
func testConfigSubcommand(t *testing.T, cmd *cobra.Command, expectedUse string) {
	t.Helper()
	if cmd == nil {
		t.Error("command is nil")
		return
	}
	if cmd.Use != expectedUse {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, expectedUse)
	}
	if cmd.Short == "" {
		t.Error("cmd.Short is empty")
	}
	if cmd.Long == "" {
		t.Error("cmd.Long is empty")
	}
	if cmd.RunE == nil {
		t.Error("cmd.RunE is nil")
	}
}

func TestConfigDiffCommand(t *testing.T) {
	testConfigSubcommand(t, configDiffCmd, "diff")
}

func TestConfigGenerateCommand(t *testing.T) {
	testConfigSubcommand(t, configGenerateCmd, "generate")
}

func TestConfigCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		command      *cobra.Command
		flagName     string
		shorthand    string
		checkDefault func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name:      "diff command has project-dir flag",
			command:   configDiffCmd,
			flagName:  "project-dir",
			shorthand: "p",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetString("project-dir")
				if err != nil {
					t.Errorf("failed to get project-dir flag: %v", err)
				}
				if val != "." {
					t.Errorf("project-dir flag default = %q, want %q", val, ".")
				}
			},
		},
		{
			name:      "generate command has project-dir flag",
			command:   configGenerateCmd,
			flagName:  "project-dir",
			shorthand: "p",
			checkDefault: func(t *testing.T, cmd *cobra.Command) {
				t.Helper()
				val, err := cmd.Flags().GetString("project-dir")
				if err != nil {
					t.Errorf("failed to get project-dir flag: %v", err)
				}
				if val != "." {
					t.Errorf("project-dir flag default = %q, want %q", val, ".")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.command.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
				return
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
			if tt.checkDefault != nil {
				tt.checkDefault(t, tt.command)
			}
		})
	}
}

func TestConfigCommandSubcommands(t *testing.T) {
	expectedCommands := []string{
		"diff",
		"generate",
	}

	for _, cmdName := range expectedCommands {
		t.Run("has "+cmdName+" subcommand", func(t *testing.T) {
			found := false
			for _, cmd := range configCmd.Commands() {
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

func TestRunConfigDiff_NoEcosYaml(t *testing.T) {
	// Create a temporary directory without .ecos.yaml
	tmpDir := t.TempDir()

	// Create a mock command with project-dir flag
	cmd := &cobra.Command{}
	cmd.Flags().StringP("project-dir", "p", tmpDir, "")

	// Run the command
	err := runConfigDiff(cmd, []string{})

	// Should return error because .ecos.yaml doesn't exist
	if err == nil {
		t.Error("runConfigDiff() expected error for missing .ecos.yaml, got nil")
	}
}

func TestRunConfigGenerate_NoEcosYaml(t *testing.T) {
	// Create a temporary directory without .ecos.yaml
	tmpDir := t.TempDir()

	// Create a mock command with project-dir flag
	cmd := &cobra.Command{}
	cmd.Flags().StringP("project-dir", "p", tmpDir, "")

	// Run the command
	err := runConfigGenerate(cmd, []string{})

	// Should return error because .ecos.yaml doesn't exist
	if err == nil {
		t.Error("runConfigGenerate() expected error for missing .ecos.yaml, got nil")
	}
}

func TestRunConfigDiff_WithValidConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a minimal .ecos.yaml
	ecosConfig := `project_name: test-project
model_version: v1.0.0

transform:
  dbt:
    project_dir: transform/dbt
    profile_dir: transform/dbt
    profile_file: profiles.yml
    profile: athena
    target: default
    vars:
      cur_database: "awsdatacatalog"
      cur_schema: "cur"
      cur_table: "cur-data"
    materialization:
      mode: view
      layer_overrides:
        bronze: view
        silver: view
        gold: view

aws:
  region: us-east-1
  database: test_database
  dbt_workgroup: test-dbt
  results_bucket: test-bucket
`
	configPath := filepath.Join(tmpDir, ".ecos.yaml")
	if err := os.WriteFile(configPath, []byte(ecosConfig), 0o600); err != nil {
		t.Fatalf("failed to write .ecos.yaml: %v", err)
	}

	// Create dbt directory
	dbtDir := filepath.Join(tmpDir, "transform", "dbt")
	if err := os.MkdirAll(dbtDir, 0o750); err != nil {
		t.Fatalf("failed to create dbt directory: %v", err)
	}

	// Create matching dbt files (no drift)
	dbtProjectContent := `name: "ecos"
version: "1.0.0"
require-dbt-version: ">=1.9.0"

profile: "athena"

model-paths: ["models"]
analysis-paths: ["analysis"]
test-paths: ["tests"]
seed-paths: ["seeds"]
macro-paths: ["macros"]
snapshot-paths: ["snapshots"]

target-path: "target"
clean-targets:
  - "target"
  - "dbt_packages"

flags:
  debug: false
  fail_fast: true
  send_anonymous_usage_stats: false

vars:
  cur_database: "awsdatacatalog"
  cur_schema: "cur"
  cur_table: "cur-data"
  iceberg_enabled: false
  billing_period_start: null
  billing_period_end: null
  ecos_materialization_mode: "view"
  ecos_layer_materializations:
    bronze: "view"
    silver: "view"
    gold: "view"
  ecos_use_iceberg: false
  ecos_enable_partitioning: true

models:
  ecos:
    1_bronze: {}
    2_silver: {}
    3_gold: {}
    4_serve:
      +materialized: view

seeds:
  ecos:
    +enabled: true
    +full_refresh: true

tests:
  +severity: error
  +store_failures: true
  +database: awsdatacatalog
  +schema: audit
`
	dbtProjectPath := filepath.Join(dbtDir, "dbt_project.yml")
	if err := os.WriteFile(dbtProjectPath, []byte(dbtProjectContent), 0o600); err != nil {
		t.Fatalf("failed to write dbt_project.yml: %v", err)
	}

	profilesContent := `athena:
  target: default
  outputs:
    default:
      type: athena
      s3_staging_dir: s3://test-bucket/dbt/staging/
      s3_data_dir: s3://test-bucket/dbt/data/
      s3_data_naming: schema_table
      s3_tmp_table_dir: s3://test-bucket/dbt/tmp/
      region_name: us-east-1
      database: awsdatacatalog
      schema: test_database
      work_group: test-dbt
      aws_profile_name: default
      threads: 8
      num_retries: 1
      num_boto3_retries: 5
`
	profilesPath := filepath.Join(dbtDir, "profiles.yml")
	if err := os.WriteFile(profilesPath, []byte(profilesContent), 0o600); err != nil {
		t.Fatalf("failed to write profiles.yml: %v", err)
	}

	// Create a mock command with project-dir flag
	cmd := &cobra.Command{}
	cmd.Flags().StringP("project-dir", "p", tmpDir, "")

	// Run the command
	err := runConfigDiff(cmd, []string{})
	// Should not return error when files are in sync
	if err != nil {
		t.Errorf("runConfigDiff() unexpected error: %v", err)
	}
}

func TestRunConfigGenerate_WithValidConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a minimal .ecos.yaml
	ecosConfig := `project_name: test-project
model_version: v1.0.0

transform:
  dbt:
    project_dir: transform/dbt
    profile_dir: transform/dbt
    profile_file: profiles.yml
    profile: athena
    target: default
    vars:
      cur_database: "awsdatacatalog"
      cur_schema: "cur"
      cur_table: "cur-data"
    materialization:
      mode: view
      layer_overrides:
        bronze: view
        silver: view
        gold: view

aws:
  region: us-east-1
  database: test_database
  dbt_workgroup: test-dbt
  results_bucket: test-bucket
`
	configPath := filepath.Join(tmpDir, ".ecos.yaml")
	if err := os.WriteFile(configPath, []byte(ecosConfig), 0o600); err != nil {
		t.Fatalf("failed to write .ecos.yaml: %v", err)
	}

	// Create a mock command with project-dir and force flags
	cmd := &cobra.Command{}
	cmd.Flags().StringP("project-dir", "p", tmpDir, "")
	cmd.Flags().BoolP("force", "f", false, "")
	_ = cmd.Flags().Set("force", "true")

	// Run the command
	err := runConfigGenerate(cmd, []string{})
	// Should not return error
	if err != nil {
		t.Errorf("runConfigGenerate() unexpected error: %v", err)
	}

	// Verify files were created
	dbtDir := filepath.Join(tmpDir, "transform", "dbt")
	dbtProjectPath := filepath.Join(dbtDir, "dbt_project.yml")
	profilesPath := filepath.Join(dbtDir, "profiles.yml")

	if _, err := os.Stat(dbtProjectPath); os.IsNotExist(err) {
		t.Error("dbt_project.yml was not created")
	}

	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		t.Error("profiles.yml was not created")
	}
}

func TestConfigCommandDescription(t *testing.T) {
	tests := []struct {
		name     string
		command  *cobra.Command
		field    string
		contains string
	}{
		{
			name:     "config short description mentions configuration",
			command:  configCmd,
			field:    "short",
			contains: "configuration",
		},
		{
			name:     "diff short description mentions drift",
			command:  configDiffCmd,
			field:    "short",
			contains: "drift",
		},
		{
			name:     "generate short description mentions regenerate",
			command:  configGenerateCmd,
			field:    "short",
			contains: "regenerate",
		},
		{
			name:     "config long description mentions .ecos.yaml",
			command:  configCmd,
			field:    "long",
			contains: ".ecos.yaml",
		},
		{
			name:     "diff long description mentions dbt",
			command:  configDiffCmd,
			field:    "long",
			contains: "dbt",
		},
		{
			name:     "generate long description mentions dbt",
			command:  configGenerateCmd,
			field:    "long",
			contains: "dbt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var text string
			switch tt.field {
			case "short":
				text = tt.command.Short
			case "long":
				text = tt.command.Long
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
