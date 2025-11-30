package config

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/viper"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// ConfigFilename is the default configuration filename
const ConfigFilename = ".ecos.yaml"

// LoadConfig loads configuration from the specified file
func LoadConfig(configFile string) (*EcosConfig, error) {
	var config EcosConfig

	// Initialize viper
	v := viper.New()

	if configFile != "" {
		// Use config file from the flag
		v.SetConfigFile(configFile)
	} else {
		// Find current directory
		pwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		// Search config in current directory with name ".ecos" (without extension)
		v.AddConfigPath(pwd)
		v.SetConfigType("yaml")
		v.SetConfigName(".ecos")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	// Unmarshal config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	config.SetDefaults()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// GetConfigFilePath returns the path to the configuration file in the project directory
// Returns empty string if no config file is found
func GetConfigFilePath() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(pwd, ConfigFilename)
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// FindConfigFile walks upward from CWD to root to find .ecos.yaml
func FindConfigFile() (string, error) {
	const name = ConfigFilename // ".ecos.yaml"

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to determine working directory: %w", err)
	}

	dir, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}

	for {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached root
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("%s not found", name)
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *EcosConfig {
	return &EcosConfig{
		ProjectName:  "my-cost-analysis",
		ModelVersion: "latest",
		Global: GlobalConfig{
			LogLevel: "info",
		},
		// TODO: Ingest section is intentionally minimal as it's not implemented yet
		Ingest: IngestConfig{
			// No default plugin or configuration
		},
		Transform: TransformConfig{
			Plugin: "dbt",
			DBT: DBTConfig{
				ProjectDir:  "./transform/dbt",
				ProfileDir:  "./transform/dbt",
				ProfileFile: "profiles.yml",
				Profile:     "ecos-athena",
				Target:      "prod",
			},
		},
		// TODO: Report section is intentionally minimal as it's not implemented yet
		Report: ReportConfig{
			Format:     "table",
			OutputPath: "./reports",
		},
	}
}

// SetDefaults sets default values for the configuration
func (c *EcosConfig) SetDefaults() {
	// Project defaults
	if c.ModelVersion == "" {
		c.ModelVersion = "latest"
	}

	// Global defaults
	if c.Global.LogLevel == "" {
		c.Global.LogLevel = "info"
	}

	// Transform defaults
	if c.Transform.Plugin == "" {
		c.Transform.Plugin = "dbt"
	}
	if c.Transform.DBT.ProjectDir == "" {
		c.Transform.DBT.ProjectDir = "./transform/dbt"
	}
	if c.Transform.DBT.ProfileDir == "" {
		c.Transform.DBT.ProfileDir = "./transform/dbt"
	}
	if c.Transform.DBT.ProfileFile == "" {
		c.Transform.DBT.ProfileFile = "profiles.yml"
	}
	if c.Transform.DBT.Profile == "" {
		c.Transform.DBT.Profile = "ecos-athena"
	}
	if c.Transform.DBT.Target == "" {
		c.Transform.DBT.Target = "prod"
	}

	// Report defaults
	if c.Report.Format == "" {
		c.Report.Format = "table"
	}
	if c.Report.OutputPath == "" {
		c.Report.OutputPath = "./reports"
	}
}

// Validate validates the entire EcosConfig structure
func (c *EcosConfig) Validate() error {
	if err := validateGlobalConfig(&c.Global); err != nil {
		return fmt.Errorf("global config validation failed: %w", err)
	}

	if err := validateIngestConfig(&c.Ingest); err != nil {
		return fmt.Errorf("ingest config validation failed: %w", err)
	}

	if err := validateTransformConfig(&c.Transform); err != nil {
		return fmt.Errorf("transform config validation failed: %w", err)
	}

	if err := validateReportConfig(&c.Report); err != nil {
		return fmt.Errorf("report config validation failed: %w", err)
	}

	return nil
}

// validateGlobalConfig validates GlobalConfig
func validateGlobalConfig(g *GlobalConfig) error {
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if g.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if g.LogLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level '%s', must be one of: %v", g.LogLevel, validLogLevels)
		}
	}

	return nil
}

// validateIngestConfig validates IngestConfig
func validateIngestConfig(_ *IngestConfig) error {
	// Ingest functionality is not yet implemented
	// This is a placeholder for future validation logic
	return nil
}

// validateTransformConfig validates TransformConfig
func validateTransformConfig(t *TransformConfig) error {
	// Plugin can be empty if it will be provided via command line
	// Only validate plugin-specific configurations if plugin is specified
	if t.Plugin == "" {
		return nil
	}

	// Validate plugin-specific configurations
	switch t.Plugin {
	case "dbt":
		if t.DBT.ProjectDir == "" {
			return errors.New("DBT project directory must be specified for dbt plugin")
		}
	case "sql":
		if t.SQL.ConnectionString == "" {
			return errors.New("SQL connection string must be specified for sql plugin")
		}
	}

	return nil
}

// validateReportConfig validates ReportConfig
func validateReportConfig(r *ReportConfig) error {
	if r.Format != "" {
		validFormats := []string{"json", "csv", "table", "yaml"}
		valid := false
		for _, format := range validFormats {
			if r.Format == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid report format '%s', must be one of: %v", r.Format, validFormats)
		}
	}

	return nil
}

// GenerateEcosConfig generates and writes a .ecos.yaml file to the specified directory
func GenerateEcosConfig(data EcosConfigTemplate, targetDir string) error {
	content, err := generateEcosConfigFromTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to generate ecos config: %w", err)
	}

	return WriteConfigFile(content, targetDir, ".ecos.yaml")
}

// generateEcosConfigFromTemplate generates a .ecos.yaml file using template data
func generateEcosConfigFromTemplate(data EcosConfigTemplate) (string, error) {
	tmpl, err := template.New("ecos.yaml.tmpl").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(templateFS, "templates/ecos.yaml.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse ecos config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute ecos config template: %w", err)
	}

	return buf.String(), nil
}

// WriteConfigFile writes configuration content to a file with proper path handling
func WriteConfigFile(content, targetDir, filename string) error {
	// Handle different path formats
	//nolint:gocritic // ifElseChain - if-else chain is clearer than switch for path handling
	if targetDir == "" {
		// Use current directory if targetDir is empty
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else if strings.HasPrefix(targetDir, "~/") {
		// Expand ~ to home directory if needed
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		targetDir = filepath.Join(home, targetDir[2:])
	} else if !filepath.IsAbs(targetDir) {
		// Convert relative path to absolute path
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		targetDir = absPath
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Check if directory exists and is writable
	dirInfo, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("failed to check target directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("path %s exists but is not a directory", targetDir)
	}

	// Write to file
	filePath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}
