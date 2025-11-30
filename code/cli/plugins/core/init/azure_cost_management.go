package init

import (
	"context"
	"errors"

	initUtils "github.com/ecos-labs/ecos-core/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
)

// AzureCostManagementInitPlugin handles initialization for Azure Cost Management data source
type AzureCostManagementInitPlugin struct {
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// Name returns the plugin name.
func (p *AzureCostManagementInitPlugin) Name() string           { return "azure-cost-management-init" }
func (p *AzureCostManagementInitPlugin) Version() string        { return "1.0.0" }
func (p *AzureCostManagementInitPlugin) Type() types.PluginType { return types.PluginTypeInit }
func (p *AzureCostManagementInitPlugin) IsCore() bool           { return true }
func (p *AzureCostManagementInitPlugin) Author() string         { return "ecos team" }
func (p *AzureCostManagementInitPlugin) CloudProvider() string  { return "azure" }
func (p *AzureCostManagementInitPlugin) Description() string {
	return "Initialize ecos project for Azure Cost Management data analysis"
}

func (p *AzureCostManagementInitPlugin) Documentation() string {
	return `
Azure Cost Management Init Plugin

This plugin sets up an ecos project for Azure cost analysis with:
 - Azure Cost Management API integration
 - Azure billing data export
 - Cost optimization insights

Prerequisites:
 - Azure CLI installed and configured
 - Azure Cost Management API access
 - dbt Core installed
 - dbt-synapse or dbt-fabric adapter
`
}

func (p *AzureCostManagementInitPlugin) SupportedEngines() []types.EngineOption {
	return []types.EngineOption{
		{Code: "synapse", DisplayName: "Azure Synapse Analytics", Supported: false, Default: true},
		{Code: "fabric", DisplayName: "Microsoft Fabric", Supported: false},
		{Code: "databricks", DisplayName: "Azure Databricks", Supported: false},
	}
}

func (p *AzureCostManagementInitPlugin) SupportedRegions(engine string) []types.RegionOption {
	return []types.RegionOption{
		{Code: "eastus", DisplayName: "East US", Default: true},
		{Code: "westus2", DisplayName: "West US 2"},
		{Code: "westeurope", DisplayName: "West Europe"},
		{Code: "northeurope", DisplayName: "North Europe"},
	}
}

func (p *AzureCostManagementInitPlugin) ValidateRegion(region string) error {
	// TODO: Implement Azure region validation
	// For now, accept any region as this plugin is not fully implemented
	return nil
}

func (p *AzureCostManagementInitPlugin) SupportedTransformTools() []types.TransformToolOption {
	return []types.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

func (p *AzureCostManagementInitPlugin) ValidatePrerequisites() error {
	config := &initUtils.PrereqConfig{
		AWS:        true,      // AWS CLI + credentials
		Python:     true,      // Python for dbt
		DBTAdapter: "postgre", // dbt-athena adapter (shows as "athena" in dbt --version)
	}
	return initUtils.RunPrerequisiteChecks(context.Background(), config)
}

func (p *AzureCostManagementInitPlugin) RunInteractiveSetup() error {
	return errors.New("azure Cost Management interactive setup is not implemented yet")
}

func (p *AzureCostManagementInitPlugin) GenerateConfig() error {
	return errors.New("azure Cost Management config generation is not implemented yet")
}

func (p *AzureCostManagementInitPlugin) CreateResources() error {
	return errors.New("azure Cost Management resource creation is not implemented yet")
}

func (p *AzureCostManagementInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

func (p *AzureCostManagementInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "Azure Cost Management")
}

func (p *AzureCostManagementInitPlugin) DownloadTransformModels() (string, error) {
	return "", errors.New("azure Cost Management transform models are not available yet")
}

func (p *AzureCostManagementInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

func (p *AzureCostManagementInitPlugin) SetModelVersion(version string) error {
	// This plugin doesn't support model versions yet
	return nil
}

func (p *AzureCostManagementInitPlugin) Validate(config map[string]interface{}) error {
	return errors.New("azure Cost Management validation is not implemented yet")
}

func (p *AzureCostManagementInitPlugin) Execute(ctx context.Context, config map[string]interface{}) (*types.PluginResult, error) {
	return &types.PluginResult{
		Success: false,
		Message: "Azure Cost Management plugin is not fully implemented yet",
	}, errors.New("not implemented")
}

// NewAzureCostManagement creates a new Azure Cost Management plugin instance.
func NewAzureCostManagement(force bool, outputPath string) (types.InitPlugin, error) {
	return &AzureCostManagementInitPlugin{
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("azure_cost_management", NewAzureCostManagement)
}
