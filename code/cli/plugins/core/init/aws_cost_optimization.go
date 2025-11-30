package init

import (
	"context"
	"errors"
	"fmt"

	initUtils "github.com/ecos-labs/ecos-core/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
)

// AWSCostOptimizationInitPlugin handles initialization for AWS Cost Optimization Hub data source
type AWSCostOptimizationInitPlugin struct {
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// Name returns the plugin name.
func (p *AWSCostOptimizationInitPlugin) Name() string { return "aws-cost-optimization-init" }

// Version returns the plugin version.
func (p *AWSCostOptimizationInitPlugin) Version() string { return "1.0.0" }

// Type returns the plugin type.
func (p *AWSCostOptimizationInitPlugin) Type() types.PluginType { return types.PluginTypeInit }

// IsCore returns true for core plugins.
func (p *AWSCostOptimizationInitPlugin) IsCore() bool { return true }

// Author returns the plugin author.
func (p *AWSCostOptimizationInitPlugin) Author() string { return "ecos team" }

// CloudProvider returns the cloud provider name.
func (p *AWSCostOptimizationInitPlugin) CloudProvider() string { return "aws" }

// Description returns a brief description of the plugin.
func (p *AWSCostOptimizationInitPlugin) Description() string {
	return "Initialize ecos project for AWS Cost Optimization Hub recommendations"
}

// Documentation returns detailed documentation for the plugin.
func (p *AWSCostOptimizationInitPlugin) Documentation() string {
	return `
AWS Cost Optimization Hub Init Plugin

This plugin sets up an ecos project for AWS Cost Optimization Hub analysis with:
 - Cost optimization recommendations
 - Right-sizing suggestions
 - Reserved Instance and Savings Plans recommendations
 - Resource utilization insights

Prerequisites:
 - AWS CLI installed and configured
 - AWS Cost Optimization Hub access
 - dbt Core installed
`
}

// SupportedEngines returns the list of supported data engines.
func (p *AWSCostOptimizationInitPlugin) SupportedEngines() []types.EngineOption {
	return []types.EngineOption{
		{Code: "athena", DisplayName: "Athena (serverless, pay-per-query)", Supported: true, Default: true},
		{Code: "redshift", DisplayName: "Redshift (dedicated cluster)", Supported: false},
	}
}

// ValidateRegion validates the provided AWS region.
func (p *AWSCostOptimizationInitPlugin) ValidateRegion(region string) error {
	if !initUtils.IsValidRegion(region) {
		return fmt.Errorf("invalid AWS region '%s', please enter a valid AWS region", region)
	}
	return nil
}

// SupportedTransformTools returns the list of supported transformation tools.
func (p *AWSCostOptimizationInitPlugin) SupportedTransformTools() []types.TransformToolOption {
	return []types.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

// ValidatePrerequisites checks if all required prerequisites are met.
func (p *AWSCostOptimizationInitPlugin) ValidatePrerequisites() error {
	utils.PrintWarning("AWS Cost Optimization Hub plugin is not fully implemented yet")
	return errors.New("AWS Cost Optimization Hub integration is coming soon")
}

// RunInteractiveSetup runs the interactive setup wizard.
func (p *AWSCostOptimizationInitPlugin) RunInteractiveSetup() error {
	return errors.New("AWS Cost Optimization Hub interactive setup is not implemented yet")
}

// GenerateConfig generates the configuration file.
func (p *AWSCostOptimizationInitPlugin) GenerateConfig() error {
	return errors.New("AWS Cost Optimization Hub config generation is not implemented yet")
}

// CreateResources creates the required AWS resources.
func (p *AWSCostOptimizationInitPlugin) CreateResources() error {
	return errors.New("AWS Cost Optimization Hub resource creation is not implemented yet")
}

// CreateDirectoryStructure creates the project directory structure.
func (p *AWSCostOptimizationInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

// InitializeBaseFiles initializes the base project files.
func (p *AWSCostOptimizationInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "AWS Cost Optimization Hub")
}

// DownloadTransformModels downloads the transformation models.
func (p *AWSCostOptimizationInitPlugin) DownloadTransformModels() (string, error) {
	return "", errors.New("AWS Cost Optimization Hub transform models are not available yet")
}

// PostInitSummary displays a summary after initialization.
func (p *AWSCostOptimizationInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

// SetModelVersion sets the model version for the plugin.
func (p *AWSCostOptimizationInitPlugin) SetModelVersion(version string) error {
	// This plugin doesn't support model versions yet
	return nil
}

// Validate validates the plugin configuration.
func (p *AWSCostOptimizationInitPlugin) Validate(config map[string]interface{}) error {
	return errors.New("AWS Cost Optimization Hub validation is not implemented yet")
}

// Execute executes the plugin with the given configuration.
func (p *AWSCostOptimizationInitPlugin) Execute(ctx context.Context, config map[string]interface{}) (*types.PluginResult, error) {
	return &types.PluginResult{
		Success: false,
		Message: "AWS Cost Optimization Hub plugin is not fully implemented yet",
	}, errors.New("not implemented")
}

// NewAWSCostOptimization creates a new AWS Cost Optimization plugin instance.
func NewAWSCostOptimization(force bool, outputPath string) (types.InitPlugin, error) {
	return &AWSCostOptimizationInitPlugin{
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("aws_cost_optimization", NewAWSCostOptimization)
}
