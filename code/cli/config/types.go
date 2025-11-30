package config

// EcosConfig represents the complete configuration structure for ecos CLI
type EcosConfig struct {
	ProjectName  string `yaml:"project_name,omitempty" mapstructure:"project_name"`
	ModelVersion string `yaml:"model_version,omitempty" mapstructure:"model_version"`
	DataSource   string `yaml:"data_source,omitempty" mapstructure:"data_source"`

	Global    GlobalConfig    `yaml:"global" mapstructure:"global"`
	Ingest    IngestConfig    `yaml:"ingest" mapstructure:"ingest"`
	Transform TransformConfig `yaml:"transform" mapstructure:"transform"`
	Report    ReportConfig    `yaml:"report" mapstructure:"report"`
	AWS       AWSRootConfig   `yaml:"aws,omitempty" mapstructure:"aws"`
}

// GlobalConfig contains global settings that apply across all commands
type GlobalConfig struct {
	LogLevel string `yaml:"log_level" mapstructure:"log_level"`
}

// IngestConfig contains configuration for the ingest command
type IngestConfig struct {
	Plugin      string                 `yaml:"plugin" mapstructure:"plugin"`
	OutputTable string                 `yaml:"output_table,omitempty" mapstructure:"output_table"`
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
	AWS         AWSConfig              `yaml:"aws,omitempty" mapstructure:"aws"`
	GCP         GCPConfig              `yaml:"gcp,omitempty" mapstructure:"gcp"`
	Azure       AzureConfig            `yaml:"azure,omitempty" mapstructure:"azure"`
	Athena      AthenaConfig           `yaml:"athena,omitempty" mapstructure:"athena"`
}

// AthenaConfig contains Athena-specific configuration
type AthenaConfig struct {
	Database string `yaml:"database" mapstructure:"database"`
	Table    string `yaml:"table" mapstructure:"table"`
	Region   string `yaml:"region" mapstructure:"region"`
}

// TransformConfig contains configuration for the transform command
type TransformConfig struct {
	Plugin string                 `yaml:"plugin" mapstructure:"plugin"`
	Config map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
	DBT    DBTConfig              `yaml:"dbt,omitempty" mapstructure:"dbt"`
	SQL    SQLConfig              `yaml:"sql,omitempty" mapstructure:"sql"`
}

// ReportConfig contains configuration for the report command
type ReportConfig struct {
	Plugin      string                 `yaml:"plugin,omitempty" mapstructure:"plugin"`
	Format      string                 `yaml:"format" mapstructure:"format"`
	OutputPath  string                 `yaml:"output_path" mapstructure:"output_path"`
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
	DataSources []DataSourceConfig     `yaml:"data_sources,omitempty" mapstructure:"data_sources"`
}

// AWSConfig contains AWS-specific configuration settings (used in IngestConfig).
type AWSConfig struct {
	Bucket    string `yaml:"bucket,omitempty" mapstructure:"bucket"`
	Region    string `yaml:"region,omitempty" mapstructure:"region"`
	Profile   string `yaml:"profile,omitempty" mapstructure:"profile"`
	AccessKey string `yaml:"access_key,omitempty" mapstructure:"access_key"`
	SecretKey string `yaml:"secret_key,omitempty" mapstructure:"secret_key"`
}

// AWSRootConfig contains AWS-specific configuration at root level (used in EcosConfig)
type AWSRootConfig struct {
	Region         string `yaml:"region,omitempty" mapstructure:"region"`
	Database       string `yaml:"database,omitempty" mapstructure:"database"`
	DBTWorkgroup   string `yaml:"dbt_workgroup,omitempty" mapstructure:"dbt_workgroup"`
	AdhocWorkgroup string `yaml:"adhoc_workgroup,omitempty" mapstructure:"adhoc_workgroup"`
	ResultsBucket  string `yaml:"results_bucket,omitempty" mapstructure:"results_bucket"`
}

// GCPConfig contains Google Cloud Platform-specific configuration settings
type GCPConfig struct {
	ProjectID         string `yaml:"project_id,omitempty" mapstructure:"project_id"`
	ServiceAccountKey string `yaml:"service_account_key,omitempty" mapstructure:"service_account_key"`
	BillingDatasetID  string `yaml:"billing_dataset_id,omitempty" mapstructure:"billing_dataset_id"`
	BillingTableID    string `yaml:"billing_table_id,omitempty" mapstructure:"billing_table_id"`
}

// AzureConfig contains Microsoft Azure-specific configuration settings
type AzureConfig struct {
	SubscriptionID string `yaml:"subscription_id,omitempty" mapstructure:"subscription_id"`
	TenantID       string `yaml:"tenant_id,omitempty" mapstructure:"tenant_id"`
	ClientID       string `yaml:"client_id,omitempty" mapstructure:"client_id"`
	ClientSecret   string `yaml:"client_secret,omitempty" mapstructure:"client_secret"`
}

// DBTConfig contains dbt (Data Build Tool) specific configuration settings.
type DBTConfig struct {
	ProjectDir      string                 `yaml:"project_dir" mapstructure:"project_dir"`
	ProfileDir      string                 `yaml:"profile_dir" mapstructure:"profile_dir"`
	ProfileFile     string                 `yaml:"profile_file" mapstructure:"profile_file"`
	Profile         string                 `yaml:"profile" mapstructure:"profile"`
	Target          string                 `yaml:"target" mapstructure:"target"`
	AWSProfile      string                 `yaml:"aws_profile,omitempty" mapstructure:"aws_profile"`
	Vars            map[string]string      `yaml:"vars,omitempty" mapstructure:"vars"`
	Materialization *MaterializationConfig `yaml:"materialization,omitempty" mapstructure:"materialization"`
}

// MaterializationConfig contains materialization settings for dbt models
type MaterializationConfig struct {
	Mode           string            `yaml:"mode,omitempty" mapstructure:"mode"`
	LayerOverrides map[string]string `yaml:"layer_overrides,omitempty" mapstructure:"layer_overrides"`
}

// SQLConfig contains SQL-based transformation configuration settings
type SQLConfig struct {
	ConnectionString string            `yaml:"connection_string" mapstructure:"connection_string"`
	Driver           string            `yaml:"driver" mapstructure:"driver"`
	ScriptPaths      []string          `yaml:"script_paths,omitempty" mapstructure:"script_paths"`
	Variables        map[string]string `yaml:"variables,omitempty" mapstructure:"variables"`
}

// DataSourceConfig represents a data source for reporting
type DataSourceConfig struct {
	Name           string `yaml:"name" mapstructure:"name"`
	Type           string `yaml:"type" mapstructure:"type"`
	ConnectionInfo string `yaml:"connection_info" mapstructure:"connection_info"`
}

// PluginConfig represents generic plugin configuration
type PluginConfig struct {
	Name    string                 `yaml:"name" mapstructure:"name"`
	Version string                 `yaml:"version" mapstructure:"version"`
	Config  map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
}

// DatasourceVar represents a key-value pair for datasource variables.
type DatasourceVar struct {
	Key   string
	Value string
}

// DBTProfilesTemplate represents template data for dbt profiles.yml
type DBTProfilesTemplate struct {
	Profile       string
	Target        string
	AWSProfile    string
	AWSRegion     string
	ResultsBucket string
	Database      string
	Workgroup     string
}

// DBTProjectTemplate represents template data for dbt_project.yml
type DBTProjectTemplate struct {
	Profile               string
	DatasourceVars        []DatasourceVar
	IcebergEnabled        bool
	BillingPeriodStart    any // can be string or null
	BillingPeriodEnd      any // can be string or null
	MaterializationMode   string
	BronzeMaterialization string
	SilverMaterialization string
	GoldMaterialization   string
	UseIceberg            bool
	EnablePartitioning    bool
}

// EcosConfigTemplate represents template data for .ecos.yaml
type EcosConfigTemplate struct {
	ProjectName           string
	ModelVersion          string
	DataSource            string
	ProjectDir            string
	ProfileDir            string
	Profile               string
	Target                string
	AWSProfile            string
	DatasourceVars        []DatasourceVar
	AWSRegion             string
	Database              string
	DBTWorkgroup          string
	AdhocWorkgroup        string
	ResultsBucket         string
	MaterializationMode   string
	BronzeMaterialization string
	SilverMaterialization string
	GoldMaterialization   string
}
