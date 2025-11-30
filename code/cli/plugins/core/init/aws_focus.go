//nolint:dupl // Intentional similarity with aws_trusted_advisor.go
package init

import (
	"context"
	"errors"
	"fmt"

	initUtils "github.com/ecos-labs/ecos-core/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
)

// AWSFocusInitPlugin handles initialization for AWS FOCUS data source
type AWSFocusInitPlugin struct {
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// Name returns the plugin name.
func (p *AWSFocusInitPlugin) Name() string { return "aws-focus-init" }

// Version returns the plugin version.
func (p *AWSFocusInitPlugin) Version() string { return "1.0.0" }

// Type returns the plugin type.
func (p *AWSFocusInitPlugin) Type() types.PluginType { return types.PluginTypeInit }

// IsCore returns true for core plugins.
func (p *AWSFocusInitPlugin) IsCore() bool { return true }

// Author returns the plugin author.
func (p *AWSFocusInitPlugin) Author() string { return "ecos team" }

// CloudProvider returns the cloud provider name.
func (p *AWSFocusInitPlugin) CloudProvider() string { return "aws" }

// Description returns a brief description of the plugin.
func (p *AWSFocusInitPlugin) Description() string {
	return "Initialize ecos project for AWS FinOps Open Cost and Usage Report (FOCUS 1.0)"
}

// Documentation returns detailed documentation for the plugin.
func (p *AWSFocusInitPlugin) Documentation() string {
	return `
AWS FOCUS Init Plugin

This plugin sets up an ecos project for AWS FOCUS (FinOps Open Cost and Usage Specification) analysis with:
 - FOCUS 1.0 compliant cost data
 - Standardized FinOps metrics
 - Cross-cloud cost comparison capabilities

Prerequisites:
 - AWS CLI installed and configured
 - AWS Cost and Billing Conductor access
 - dbt Core installed
`
}

// SupportedEngines returns the list of supported data engines.
func (p *AWSFocusInitPlugin) SupportedEngines() []types.EngineOption {
	return []types.EngineOption{
		{Code: "athena", DisplayName: "Athena (serverless, pay-per-query)", Supported: true, Default: true},
		{Code: "redshift", DisplayName: "Redshift (dedicated cluster)", Supported: false},
	}
}

// ValidateRegion validates the provided AWS region.
func (p *AWSFocusInitPlugin) ValidateRegion(region string) error {
	if !initUtils.IsValidRegion(region) {
		return fmt.Errorf("invalid AWS region '%s', please enter a valid AWS region", region)
	}
	return nil
}

// SupportedTransformTools returns the list of supported transformation tools.
func (p *AWSFocusInitPlugin) SupportedTransformTools() []types.TransformToolOption {
	return []types.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

// ValidatePrerequisites checks if all required prerequisites are met.
func (p *AWSFocusInitPlugin) ValidatePrerequisites() error {
	config := &initUtils.PrereqConfig{
		AWS:        true,     // AWS CLI + credentials
		Python:     true,     // Python for dbt
		DBTAdapter: "athena", // dbt-athena adapter (shows as "athena" in dbt --version)
	}

	return initUtils.RunPrerequisiteChecks(context.Background(), config)
}

// RunInteractiveSetup runs the interactive setup wizard.
func (p *AWSFocusInitPlugin) RunInteractiveSetup() error {
	return errors.New("AWS FOCUS interactive setup is not implemented yet")
}

// GenerateConfig generates the configuration file.
func (p *AWSFocusInitPlugin) GenerateConfig() error {
	return errors.New("AWS FOCUS config generation is not implemented yet")
}

// CreateResources creates the required AWS resources.
func (p *AWSFocusInitPlugin) CreateResources() error {
	return errors.New("AWS FOCUS resource creation is not implemented yet")
}

// CreateDirectoryStructure creates the project directory structure.
func (p *AWSFocusInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

// InitializeBaseFiles initializes the base project files.
func (p *AWSFocusInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "AWS FOCUS")
}

// DownloadTransformModels downloads the transformation models.
func (p *AWSFocusInitPlugin) DownloadTransformModels() (string, error) {
	return "", errors.New("AWS FOCUS transform models are not available yet")
}

// PostInitSummary displays a summary after initialization.
func (p *AWSFocusInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

// SetModelVersion sets the model version for the plugin.
func (p *AWSFocusInitPlugin) SetModelVersion(version string) error {
	// This plugin doesn't support model versions yet
	return nil
}

// Validate validates the plugin configuration.
func (p *AWSFocusInitPlugin) Validate(config map[string]interface{}) error {
	return errors.New("AWS FOCUS validation is not implemented yet")
}

// Execute executes the plugin with the given configuration.
func (p *AWSFocusInitPlugin) Execute(ctx context.Context, config map[string]interface{}) (*types.PluginResult, error) {
	return &types.PluginResult{
		Success: false,
		Message: "AWS FOCUS plugin is not fully implemented yet",
	}, errors.New("not implemented")
}

// NewAWSFocus creates a new AWS FOCUS init plugin instance.
func NewAWSFocus(force bool, outputPath string) (types.InitPlugin, error) {
	return &AWSFocusInitPlugin{
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("aws_focus", NewAWSFocus)
}
