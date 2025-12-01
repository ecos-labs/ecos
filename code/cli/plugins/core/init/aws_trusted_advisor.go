//nolint:dupl // Intentional similarity with aws_focus.go
package init

import (
	"context"
	"errors"
	"fmt"

	initUtils "github.com/ecos-labs/ecos/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos/code/cli/plugins/types"
)

// AWSTrustedAdvisorInitPlugin handles initialization for AWS Trusted Advisor data source
type AWSTrustedAdvisorInitPlugin struct {
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// Name returns the plugin name.
func (p *AWSTrustedAdvisorInitPlugin) Name() string { return "aws-trusted-advisor-init" }

// Version returns the plugin version.
func (p *AWSTrustedAdvisorInitPlugin) Version() string { return "1.0.0" }

// Type returns the plugin type.
func (p *AWSTrustedAdvisorInitPlugin) Type() types.PluginType { return types.PluginTypeInit }

// IsCore returns whether this is a core plugin.
func (p *AWSTrustedAdvisorInitPlugin) IsCore() bool { return true }

// Author returns the plugin author.
func (p *AWSTrustedAdvisorInitPlugin) Author() string { return "ecos team" }

// CloudProvider returns the cloud provider name.
func (p *AWSTrustedAdvisorInitPlugin) CloudProvider() string { return "aws" }

// Description returns a short description of the plugin.
func (p *AWSTrustedAdvisorInitPlugin) Description() string {
	return "Initialize ecos project for AWS Trusted Advisor recommendations"
}

// Documentation returns detailed documentation for the plugin.
func (p *AWSTrustedAdvisorInitPlugin) Documentation() string {
	return `
AWS Trusted Advisor Init Plugin

This plugin sets up an ecos project for AWS Trusted Advisor analysis with:
 - AWS Trusted Advisor API integration
 - Cost optimization recommendations
 - Performance and security insights

Prerequisites:
 - AWS CLI installed and configured
 - AWS Support API access (Business or Enterprise support plan)
 - dbt Core installed
`
}

// SupportedEngines returns the list of supported data engines.
func (p *AWSTrustedAdvisorInitPlugin) SupportedEngines() []types.EngineOption {
	return []types.EngineOption{
		{Code: "athena", DisplayName: "Athena (serverless, pay-per-query)", Supported: true, Default: true},
		{Code: "redshift", DisplayName: "Redshift (dedicated cluster)", Supported: false},
	}
}

// ValidateRegion validates the provided AWS region.
func (p *AWSTrustedAdvisorInitPlugin) ValidateRegion(region string) error {
	if !initUtils.IsValidRegion(region) {
		return fmt.Errorf("invalid AWS region '%s', please enter a valid AWS region", region)
	}
	return nil
}

// SupportedTransformTools returns the list of supported transformation tools.
func (p *AWSTrustedAdvisorInitPlugin) SupportedTransformTools() []types.TransformToolOption {
	return []types.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

func (p *AWSTrustedAdvisorInitPlugin) ValidatePrerequisites() error {
	config := &initUtils.PrereqConfig{
		AWS:        true,     // AWS CLI + credentials
		Python:     true,     // Python for dbt
		DBTAdapter: "athena", // dbt-athena adapter (shows as "athena" in dbt --version)
	}

	return initUtils.RunPrerequisiteChecks(context.Background(), config)
}

func (p *AWSTrustedAdvisorInitPlugin) RunInteractiveSetup() error {
	return errors.New("AWS Trusted Advisor interactive setup is not implemented yet")
}

func (p *AWSTrustedAdvisorInitPlugin) GenerateConfig() error {
	return errors.New("AWS Trusted Advisor config generation is not implemented yet")
}

func (p *AWSTrustedAdvisorInitPlugin) CreateResources() error {
	return errors.New("AWS Trusted Advisor resource creation is not implemented yet")
}

func (p *AWSTrustedAdvisorInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

func (p *AWSTrustedAdvisorInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "AWS Trusted Advisor")
}

func (p *AWSTrustedAdvisorInitPlugin) DownloadTransformModels() (string, error) {
	return "", errors.New("AWS Trusted Advisor transform models are not available yet")
}

func (p *AWSTrustedAdvisorInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

func (p *AWSTrustedAdvisorInitPlugin) SetModelVersion(version string) error {
	// This plugin doesn't support model versions yet
	return nil
}

func (p *AWSTrustedAdvisorInitPlugin) Validate(config map[string]interface{}) error {
	return errors.New("AWS Trusted Advisor validation is not implemented yet")
}

func (p *AWSTrustedAdvisorInitPlugin) Execute(ctx context.Context, config map[string]interface{}) (*types.PluginResult, error) {
	return &types.PluginResult{
		Success: false,
		Message: "AWS Trusted Advisor plugin is not fully implemented yet",
	}, errors.New("not implemented")
}

// NewAWSTrustedAdvisor creates a new AWS Trusted Advisor plugin instance.
func NewAWSTrustedAdvisor(force bool, outputPath string) (types.InitPlugin, error) {
	return &AWSTrustedAdvisorInitPlugin{
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("aws_trusted_advisor", NewAWSTrustedAdvisor)
}
