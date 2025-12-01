package init

import (
	"context"
	"errors"

	initUtils "github.com/ecos-labs/ecos/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos/code/cli/plugins/types"
	"github.com/ecos-labs/ecos/code/cli/utils"
)

// GCPBillingExportInitPlugin handles initialization for GCP Billing Export data source
type GCPBillingExportInitPlugin struct {
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// Name returns the plugin name.
func (p *GCPBillingExportInitPlugin) Name() string           { return "gcp-billing-export-init" }
func (p *GCPBillingExportInitPlugin) Version() string        { return "1.0.0" }
func (p *GCPBillingExportInitPlugin) Type() types.PluginType { return types.PluginTypeInit }
func (p *GCPBillingExportInitPlugin) IsCore() bool           { return true }
func (p *GCPBillingExportInitPlugin) Author() string         { return "ecos team" }
func (p *GCPBillingExportInitPlugin) CloudProvider() string  { return "gcp" }
func (p *GCPBillingExportInitPlugin) Description() string {
	return "Initialize ecos project for GCP Billing Export data analysis"
}

func (p *GCPBillingExportInitPlugin) Documentation() string {
	return `
GCP Billing Export Init Plugin

This plugin sets up an ecos project for GCP cost analysis with:
 - GCP Billing Export integration
 - BigQuery data warehouse setup
 - Cost optimization insights

Prerequisites:
 - Google Cloud CLI (gcloud) installed and configured
 - GCP Billing Export configured
 - dbt Core installed
 - dbt-bigquery adapter
`
}

func (p *GCPBillingExportInitPlugin) SupportedEngines() []types.EngineOption {
	return []types.EngineOption{
		{Code: "bigquery", DisplayName: "BigQuery (serverless, pay-per-query)", Supported: false, Default: true},
		{Code: "dataproc", DisplayName: "Dataproc (managed Spark)", Supported: false},
	}
}

func (p *GCPBillingExportInitPlugin) SupportedRegions(engine string) []types.RegionOption {
	return []types.RegionOption{
		{Code: "us-central1", DisplayName: "us-central1 (Iowa)", Default: true},
		{Code: "us-east1", DisplayName: "us-east1 (South Carolina)"},
		{Code: "us-west1", DisplayName: "us-west1 (Oregon)"},
		{Code: "europe-west1", DisplayName: "europe-west1 (Belgium)"},
		{Code: "asia-east1", DisplayName: "asia-east1 (Taiwan)"},
	}
}

func (p *GCPBillingExportInitPlugin) ValidateRegion(region string) error {
	// TODO: Implement GCP region validation
	// For now, accept any region as this plugin is not fully implemented
	return nil
}

func (p *GCPBillingExportInitPlugin) SupportedTransformTools() []types.TransformToolOption {
	return []types.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

func (p *GCPBillingExportInitPlugin) ValidatePrerequisites() error {
	utils.PrintWarning("GCP Billing Export plugin is not fully implemented yet")
	return errors.New("GCP Billing Export integration is coming soon")
}

func (p *GCPBillingExportInitPlugin) RunInteractiveSetup() error {
	return errors.New("GCP Billing Export interactive setup is not implemented yet")
}

func (p *GCPBillingExportInitPlugin) GenerateConfig() error {
	return errors.New("GCP Billing Export config generation is not implemented yet")
}

func (p *GCPBillingExportInitPlugin) CreateResources() error {
	return errors.New("GCP Billing Export resource creation is not implemented yet")
}

func (p *GCPBillingExportInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

func (p *GCPBillingExportInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "GCP Billing Export")
}

func (p *GCPBillingExportInitPlugin) DownloadTransformModels() (string, error) {
	return "", errors.New("GCP Billing Export transform models are not available yet")
}

func (p *GCPBillingExportInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

func (p *GCPBillingExportInitPlugin) SetModelVersion(version string) error {
	// This plugin doesn't support model versions yet
	return nil
}

func (p *GCPBillingExportInitPlugin) Validate(config map[string]interface{}) error {
	return errors.New("GCP Billing Export validation is not implemented yet")
}

func (p *GCPBillingExportInitPlugin) Execute(ctx context.Context, config map[string]interface{}) (*types.PluginResult, error) {
	return &types.PluginResult{
		Success: false,
		Message: "GCP Billing Export plugin is not fully implemented yet",
	}, errors.New("not implemented")
}

// NewGCPBillingExport creates a new GCP Billing Export plugin instance.
func NewGCPBillingExport(force bool, outputPath string) (types.InitPlugin, error) {
	return &GCPBillingExportInitPlugin{
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("gcp_billing_export", NewGCPBillingExport)
}
