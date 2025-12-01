package types

import (
	"context"
	"time"

	"github.com/ecos-labs/ecos/code/cli/config"
)

// PluginType represents the type of plugin
type PluginType string

const (
	// PluginTypeIngest represents an ingest plugin type.
	PluginTypeIngest PluginType = "ingest"
	// PluginTypeTransform represents a transform plugin type.
	PluginTypeTransform PluginType = "transform"
	// PluginTypeVerify represents a verify plugin type.
	PluginTypeVerify PluginType = "verify"
	// PluginTypeReport represents a report plugin type.
	PluginTypeReport PluginType = "report"
	// PluginTypeInit represents an init plugin type.
	PluginTypeInit PluginType = "init"
)

// PluginResult represents the result from any plugin execution
type PluginResult struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	Duration time.Duration  `json:"duration"`
	Metadata map[string]any `json:"metadata"`
	Error    string         `json:"error,omitempty"`
	ExitCode int            `json:"exit_code"`
}

// Plugin represents the base interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name (e.g., "aws-cur", "gcp-billing")
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a brief description of what the plugin does
	Description() string

	// Type returns the plugin type (ingest, transform, verify, report)
	Type() PluginType

	// Validate validates the plugin configuration
	Validate(config map[string]any) error

	// Execute runs the plugin with the given configuration
	Execute(ctx context.Context, config map[string]any) (*PluginResult, error)
}

// CorePlugin represents built-in plugins compiled into the ecos binary
type CorePlugin interface {
	Plugin

	// IsCore always returns true for core plugins
	IsCore() bool

	// Author returns the plugin author (typically "ecos team")
	Author() string

	// Documentation returns help documentation for the plugin
	Documentation() string
}

// ValidationReport represents the validation status of generated files
type ValidationReport struct {
	HasChanges bool
	Files      map[string]*FileDiffReport
}

// FileDiffReport represents the diff status of a single file
type FileDiffReport struct {
	FilePath   string
	HasChanges bool
	Diff       string
}

// IngestPlugin represents plugins that handle data ingestion from various sources
type IngestPlugin interface {
	CorePlugin

	// DataSource returns the data source type (e.g., "aws-cur", "gcp-billing")
	DataSource() string

	// ValidateConnection validates the connection to the data source
	ValidateConnection(config map[string]any) error

	// FetchData fetches data from the source
	FetchData(config map[string]any) error

	// ProcessData processes the fetched data
	ProcessData(config map[string]any) error

	// StoreData stores the processed data
	StoreData(config map[string]any) error
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Type        PluginType `json:"type"`
	Author      string     `json:"author"`
	IsCore      bool       `json:"is_core"`
	Location    string     `json:"location"`
}

// EngineOption describes a SQL engine supported by a cloud provider.
type EngineOption struct {
	Code        string // Short code/id, e.g. "athena", "redshift"
	DisplayName string // User-friendly label for CLI dropdown
	Description string // Optional: more detail for display/help
	Supported   bool   // Whether this engine is available for use
	Default     bool   // Should this be the default selection
}

// RegionOption represents a supported region for a cloud/engine.
type RegionOption struct {
	Code        string // Region code, e.g. "us-east-1"
	DisplayName string // User display name, e.g. "us-east-1 (N. Virginia)"
	Default     bool   // Default selection for region list
}

// PluginFactory creates a plugin instance with empty config
type PluginFactory func(force bool, outputPath string) (InitPlugin, error)

// InitPlugin represents plugins that handle project initialization
// and project scaffolding for specific cloud providers.
type InitPlugin interface {
	CorePlugin

	// CloudProvider returns the cloud provider name (e.g., "aws", "gcp", "azure").
	CloudProvider() string

	// ValidatePrerequisites checks if all necessary tools/dependencies are installed.
	// Returns an error with details for anything missing.
	ValidatePrerequisites() error

	// SupportedEngines lists SQL engines supported by this provider (e.g., Athena, Redshift).
	SupportedEngines() []EngineOption

	// ValidateRegions validates that the specified region is valid for the cloud provider.
	ValidateRegion(region string) error

	// SupportedTransformTools lists available transform options (e.g., dbt).
	SupportedTransformTools() []TransformToolOption

	// RunInteractiveSetup runs question-based (survey/dropdown) interactive setup.
	// Fills the plugin's internal config struct directly.
	RunInteractiveSetup() error

	// GenerateConfig generates and writes the yaml config file based on the collected input.
	GenerateConfig() error

	// CreateResources provisions cloud resources if needed (e.g., S3, workgroups).
	CreateResources() error

	// CreateDirectoryStructure makes core folders (e.g., transform, ingest, output).
	CreateDirectoryStructure() error

	// InitializeBaseFiles writes starter files (e.g., .gitignore, README.md, etc).
	InitializeBaseFiles() error

	// DownloadTransformModels fetches/extracts transform model packages (e.g., dbt models).
	// Returns the version of the downloaded models.
	DownloadTransformModels() (string, error)

	// PostInitSummary prints or logs a summary of what was done/created.
	PostInitSummary() error

	// SetModelVersion sets the model version for the plugin if supported
	SetModelVersion(version string) error
}

// InitStatus represents the status of an initialization operation.
type InitStatus string

const (
	// InitStatusCreated indicates the resource was successfully created.
	InitStatusCreated InitStatus = "created"
	// InitStatusSkipped indicates the resource creation was skipped.
	InitStatusSkipped InitStatus = "skipped"
	// InitStatusFailed indicates the resource creation failed.
	InitStatusFailed InitStatus = "failed"
	// InitStatusPartiallyCreated indicates the resource was partially created.
	InitStatusPartiallyCreated InitStatus = "partially_created"
)

// InitResourceResult represents the result of creating an AWS resource
type InitResourceResult struct {
	Kind    string
	Name    string
	Status  InitStatus
	Error   string
	Warning string // Optional warning message for partially created resources
}

// TransformPlugin represents plugins that handle data transformation
type TransformPlugin interface {
	CorePlugin

	// TransformEngine returns the transformation engine name (e.g., "dbt", "sqlmesh")
	TransformEngine() string

	// GetSupportedCommands returns all native commands this tool supports
	GetSupportedCommands() []string

	// ExecuteCommand runs native tool commands with full argument pass-through
	// This enables: ecos transform dbt run --select +my_model
	ExecuteCommand(ctx context.Context, command string, args []string, config map[string]any) error

	// GetProjectPath returns the path to the transform project directory
	GetProjectPath() string

	// PrepareEnvironment sets up the execution environment
	PrepareEnvironment(ctx context.Context, config map[string]any) error

	// ValidateEnvironment checks if the tool and project are properly configured
	ValidateEnvironment(config map[string]any) error

	// ShowCommandHelp displays help information for a specific command
	// This allows each plugin to provide its own command-specific help
	ShowCommandHelp(command string) error

	// BuildConfig builds plugin-specific configuration from ecos config
	// This allows each plugin to handle its own configuration logic
	BuildConfig(ecosConfig any, command string, args []string) map[string]any
}

// TransformStatus represents the status of a transform setup
type TransformStatus struct {
	Tool          string            `json:"tool"`
	Version       string            `json:"version"`
	ProjectValid  bool              `json:"project_valid"`
	ProfilesValid bool              `json:"profiles_valid"`
	Dependencies  map[string]string `json:"dependencies"`
	LastRun       *time.Time        `json:"last_run,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
}

// TransformToolOption represents a supported transformation engine/tool.
type TransformToolOption struct {
	Code        string // "dbt", "sql", etc.
	DisplayName string // User display name (for dropdowns)
	Supported   bool   // Is this tool supported for this provider/engine
	Default     bool   // Should this be pre-selected
}

// DestroyStatus represents the status of a resource destruction operation.
type DestroyStatus string

const (
	// DestroyStatusDeleted indicates the resource was successfully deleted.
	DestroyStatusDeleted DestroyStatus = "deleted"
	// DestroyStatusFailed indicates the resource deletion failed.
	DestroyStatusFailed DestroyStatus = "failed"
	// DestroyStatusSkipped indicates the resource deletion was skipped.
	DestroyStatusSkipped DestroyStatus = "skipped"
	// DestroyStatusCreated indicates the resource was created (unexpected state for destroy).
	DestroyStatusCreated DestroyStatus = "created"
)

// DestroyResourceResult represents the result of destroying a resource.
type DestroyResourceResult struct {
	Kind   string
	Name   string
	Status DestroyStatus
	Error  string
}

// DestroyResourcePreview represents a preview of a resource that will be destroyed.
type DestroyResourcePreview struct {
	Kind    string
	Name    string
	Managed bool
	Error   string
}

// DestroyPlugin is the base interface all destroy plugins implement.
type DestroyPlugin interface {
	// Identify the provider (e.g. aws_cur, azure, gcp)
	Name() string

	// Validate environment (credentials, region, etc.)
	ValidatePrerequisites() error

	// Destroy cloud resources
	DestroyResources() ([]DestroyResourceResult, error)
}

// ConfigurableDestroyPlugin extends DestroyPlugin by supporting config loading
// and describing the resources to be destroyed.
type ConfigurableDestroyPlugin interface {
	DestroyPlugin

	// Load plugin state from .ecos.yaml (typed)
	LoadFromConfig(cfg *config.EcosConfig) error

	// Describe all cloud resources that will be destroyed.
	// The destroy command will print this before asking for confirmation.
	DescribeDestruction() []DestroyResourcePreview
}

// DestroyConfigLoader supports config loading for destroy plugins.
type DestroyConfigLoader interface {
	LoadFromConfig(cfg *config.EcosConfig) error
}

// DestroyPreviewer supports resource preview for destroy plugins.
type DestroyPreviewer interface {
	DescribeDestruction() []DestroyResourcePreview
}

// DestroyExecutor supports actual destruction of resources.
type DestroyExecutor interface {
	DestroyResources() ([]DestroyResourceResult, error)
}
