package init

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	athenaTypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ecos-labs/ecos/code/cli/config"
	initUtils "github.com/ecos-labs/ecos/code/cli/plugins/core/init/utils"
	"github.com/ecos-labs/ecos/code/cli/plugins/registry"
	initTypes "github.com/ecos-labs/ecos/code/cli/plugins/types"
	"github.com/ecos-labs/ecos/code/cli/utils"
)

const (
	// AdhocQueryRetentionDays is the number of days to retain adhoc query results before deletion
	AdhocQueryRetentionDays = 30
	// IncompleteUploadCleanupDays is the number of days after which incomplete multipart uploads are aborted
	IncompleteUploadCleanupDays = 7
)

// normalizeDatabaseName converts a project name to a valid database name by replacing
// spaces and hyphens with underscores. This ensures consistent database naming across
// all code paths (GenerateConfig and generateDBTFiles).
func normalizeDatabaseName(projectName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(projectName, " ", "_"), "-", "_")
}

// ------------------- Plugin Metadata -----------------------

type AWSCURInitPlugin struct {
	Config     *AWSCURInput
	Force      bool
	OutputPath string
	SkipPrereq bool
}

// AWSCURInput represents the user input for AWS CUR initialization
type AWSCURInput struct {
	ProjectName      string `mapstructure:"project_name"`
	TransformTool    string `mapstructure:"transform_tool"`
	SQLEngine        string `mapstructure:"sql_engine"`
	CURDatabase      string `mapstructure:"cur_database"`
	CURSchema        string `mapstructure:"cur_schema"`
	CURTable         string `mapstructure:"cur_table"`
	AWSRegion        string `mapstructure:"aws_region"`
	AWSProfile       string `mapstructure:"aws_profile"`
	CreateResources  bool   `mapstructure:"create_resources"`
	SkipProvisioning bool   `mapstructure:"skip_provisioning"`
	DBTWorkgroup     string `mapstructure:"dbt_workgroup"`
	AdhocWorkgroup   string `mapstructure:"adhoc_workgroup"`
	ResultsBucket    string `mapstructure:"results_bucket"`
	S3StagingDir     string `mapstructure:"s3_staging_dir"`
	AccountID        string `mapstructure:"account_id"`
	DetectedRegion   string `mapstructure:"detected_region"`
	ModelVersion     string `mapstructure:"model_version"`
	DryRun           bool   `mapstructure:"dry_run"`
}

// MaterializationConfig represents materialization settings
type MaterializationConfig struct {
	Mode   string // Global materialization mode
	Bronze string // Bronze layer materialization
	Silver string // Silver layer materialization
	Gold   string // Gold layer materialization
}

func DefaultMaterializationConfig() MaterializationConfig {
	return MaterializationConfig{
		Mode:   "view",
		Bronze: "view",
		Silver: "view",
		Gold:   "view",
	}
}

func (p *AWSCURInitPlugin) Name() string               { return "aws-cur-init" }
func (p *AWSCURInitPlugin) Version() string            { return "1.0.0" }
func (p *AWSCURInitPlugin) Type() initTypes.PluginType { return initTypes.PluginTypeInit }
func (p *AWSCURInitPlugin) IsCore() bool               { return true }
func (p *AWSCURInitPlugin) Author() string             { return "ecos team" }
func (p *AWSCURInitPlugin) CloudProvider() string      { return "aws" }

func (p *AWSCURInitPlugin) Description() string {
	return "Initialize ecos project for AWS Cost and Usage Reports with Athena integration"
}

func (p *AWSCURInitPlugin) Documentation() string {
	return `
AWS Init Plugin

This plugin sets up an ecos project for AWS cost analysis with:
 - AWS Cost and Usage Report (CUR) integration
 - Athena workgroups for dbt transformations
 - S3 bucket structure for query results

Prerequisites:
 - AWS CLI installed and configured
 - dbt Core installed
 - dbt-athena-community adapter installed
`
}

func (p *AWSCURInitPlugin) SupportedEngines() []initTypes.EngineOption {
	return []initTypes.EngineOption{
		{Code: "athena", DisplayName: "Athena (serverless, pay-per-query)", Supported: true, Default: true},
		{Code: "redshift", DisplayName: "Redshift (dedicated cluster)", Supported: false},
	}
}

func (p *AWSCURInitPlugin) SupportedTransformTools() []initTypes.TransformToolOption {
	return []initTypes.TransformToolOption{
		{Code: "dbt", DisplayName: "dbt (data build tool)", Supported: true, Default: true},
	}
}

func (p *AWSCURInitPlugin) ValidateRegion(region string) error {
	if !initUtils.IsValidRegion(region) {
		return fmt.Errorf("invalid AWS region '%s', please enter a valid AWS region", region)
	}
	return nil
}

func (p *AWSCURInitPlugin) ValidatePrerequisites() error {
	// Check basic prerequisites first
	config := &initUtils.PrereqConfig{
		AWS:        true,     // AWS CLI + credentials
		Python:     true,     // Python for dbt + Athena adapter
		DBTAdapter: "athena", // dbt-athena adapter (shows as "athena" in dbt --version)
	}

	return initUtils.RunPrerequisiteChecks(context.Background(), config)
}

func (p *AWSCURInitPlugin) RunInteractiveSetup() error {
	ctx := context.Background()

	// Initialize config if not already done
	if p.Config == nil {
		p.Config = &AWSCURInput{}
	}

	// 2. Transform Engine (dropdown, plugin-driven)
	toolOpts := p.SupportedTransformTools()
	toolDisplay, defaultToolIdx := []string{}, 0
	for idx, t := range toolOpts {
		label := t.DisplayName
		if !t.Supported {
			label += " (coming soon)"
		}
		toolDisplay = append(toolDisplay, label)
		if t.Default {
			defaultToolIdx = idx
		}
	}
	toolIdx, _, err := utils.Select("Transform Engine", toolDisplay, defaultToolIdx, true, false)
	if err != nil {
		return err
	}
	if !toolOpts[toolIdx].Supported {
		return fmt.Errorf("%s is not supported yet", toolOpts[toolIdx].DisplayName)
	}
	p.Config.TransformTool = toolOpts[toolIdx].Code

	// 3. Data Warehouse (dropdown, plugin-driven)
	engineOpts := p.SupportedEngines()
	engineDisplay, defaultEngineIdx := []string{}, 0
	for idx, e := range engineOpts {
		label := e.DisplayName
		if !e.Supported {
			label += " (coming soon)"
		}
		engineDisplay = append(engineDisplay, label)
		if e.Default {
			defaultEngineIdx = idx
		}
	}
	engineIdx, _, err := utils.Select("Data Warehouse", engineDisplay, defaultEngineIdx, true, false)
	if err != nil {
		return err
	}
	if !engineOpts[engineIdx].Supported {
		return fmt.Errorf("%s is not supported yet", engineOpts[engineIdx].DisplayName)
	}
	p.Config.SQLEngine = engineOpts[engineIdx].Code

	// 4. CUR Datasource Details
	utils.PrintSubHeader("🗄️ CUR Datasource Details")

	// 4a. Database
	curDatabase, err := utils.Input("Database", "awsdatacatalog", true, true, nil)
	if err != nil {
		return err
	}
	p.Config.CURDatabase = curDatabase

	// 4b. Schema
	curSchema, err := utils.Input("Schema", "cur", true, true, nil)
	if err != nil {
		return err
	}
	p.Config.CURSchema = curSchema

	// 4c. Table
	curTable, err := utils.Input("Table", "cur-data", true, true, nil)
	if err != nil {
		return err
	}
	p.Config.CURTable = curTable

	fmt.Println()

	// 5. Project Name
	uiProjectName, err := utils.Input("Project Name", "my-cost-analysis", true, false, nil)
	if err != nil {
		return err
	}
	p.Config.ProjectName = uiProjectName

	// 6. AWS Configuration
	utils.PrintSubHeader("☁️ AWS Configuration")

	// 6a. AWS Region
	awsRegion, err := utils.Input("Region", "eu-west-1", true, true, p.ValidateRegion)
	if err != nil {
		return err
	}
	p.Config.AWSRegion = awsRegion

	// 6b. AWS Profile
	awsProfile, err := utils.Input("Profile", "default", true, true, nil)
	if err != nil {
		return err
	}
	p.Config.AWSProfile = awsProfile

	// 7. Resource Preview
	projectName := strings.ReplaceAll(uiProjectName, " ", "-")

	// Get account ID for resource naming using the specified profile
	accountID, detectedRegion, err := initUtils.GetAWSAccountAndRegionWithProfile(ctx, 0, awsProfile)
	if err != nil {
		return fmt.Errorf("failed to get AWS account and region for resource preview: %w", err)
	}

	// Use detected region if it matches user input, otherwise warn
	if detectedRegion != awsRegion {
		utils.PrintWarning(fmt.Sprintf("AWS config region (%s) differs from selected region (%s). Using selected region.", detectedRegion, awsRegion))
	}

	utils.PrintSubHeader("📦 Resource Preview")
	utils.PrintInfo("The following AWS resources are required for data transformation and analysis:")
	fmt.Println()
	headers := []string{"Type", "Name"}
	rows := [][]string{
		{"S3 Bucket", fmt.Sprintf("%s-bucket-%s-%s", projectName, accountID, awsRegion)},
		{"Workgroup", fmt.Sprintf("%s-dbt", projectName)},
		{"Workgroup", fmt.Sprintf("%s-adhoc", projectName)},
	}
	utils.PrintTable(headers, rows)
	fmt.Println()

	// 8. Resource Provisioning
	provisionItems := []string{
		"Have ecos provision these resources (recommended)",
		"Use existing AWS resources",
		"Skip provisioning",
	}
	provisionIdx, _, err := utils.Select("Resource Provisioning", provisionItems, 0, false, false)
	if err != nil {
		return err
	}

	switch provisionIdx {
	case 0: // Let ecos create them
		p.Config.CreateResources = true
		p.Config.DBTWorkgroup = fmt.Sprintf("%s-dbt", projectName)
		p.Config.AdhocWorkgroup = fmt.Sprintf("%s-adhoc", projectName)
		p.Config.ResultsBucket = fmt.Sprintf("%s-bucket-%s-%s", projectName, accountID, awsRegion)
		p.Config.AccountID = accountID

		// Ask for confirmation
		confirm := utils.ConfirmPrompt("Do you want to proceed with creating these resources")
		if !confirm {
			fmt.Println("Resource creation cancelled. You can create these resources manually later.")
			p.Config.CreateResources = false
			p.Config.SkipProvisioning = true
		}
	case 1: // Use my existing resources
		p.Config.CreateResources = false

		dbtWg, err := utils.Input("dbt Workgroup name", "", false, false, nil)
		if err != nil {
			return err
		}
		p.Config.DBTWorkgroup = dbtWg

		adhocWg, err := utils.Input("Adhoc Workgroup name (optional)", "", false, false, nil)
		if err != nil {
			return err
		}
		p.Config.AdhocWorkgroup = adhocWg

		resBucket, err := utils.Input("S3 Results Bucket", "", false, false, nil)
		if err != nil {
			return err
		}
		p.Config.ResultsBucket = resBucket

		stagingDir, err := utils.Input("S3 Staging Directory", "dbt/", false, false, nil)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(stagingDir, "/") {
			stagingDir += "/"
		}
		p.Config.S3StagingDir = fmt.Sprintf("s3://%s/%s", resBucket, stagingDir)
		utils.PrintInfo("Will use your provided existing resources!")
	case 2: // Skip provisioning
		p.Config.CreateResources = false
		p.Config.SkipProvisioning = true
		fmt.Println("Skipping automatic provisioning")
	}

	return nil
}

func (p *AWSCURInitPlugin) GenerateConfig() error {
	userInput := p.Config

	// Use the configured resource names
	dbtWorkgroup := userInput.DBTWorkgroup
	adhocWorkgroup := userInput.AdhocWorkgroup
	resultsBucket := userInput.ResultsBucket

	database := fmt.Sprintf("%s_database", normalizeDatabaseName(userInput.ProjectName))
	projectDir := filepath.Join(p.OutputPath, "transform/dbt")

	// Ensure the project directory exists
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Get default materialization configuration
	matConfig := DefaultMaterializationConfig()

	// Generate ecos configuration using template
	utils.PrintDebug("Generating .ecos.yaml configuration file")
	ecosConfigData := config.EcosConfigTemplate{
		ProjectName:  userInput.ProjectName,
		ModelVersion: userInput.ModelVersion,
		DataSource:   "aws_cur",
		ProjectDir:   projectDir,
		ProfileDir:   projectDir,
		Profile:      "ecos-athena",
		Target:       "prod",
		AWSProfile:   userInput.AWSProfile,
		DatasourceVars: []config.DatasourceVar{
			{Key: "cur_database", Value: userInput.CURDatabase},
			{Key: "cur_schema", Value: userInput.CURSchema},
			{Key: "cur_table", Value: userInput.CURTable},
		},
		AWSRegion:             userInput.AWSRegion,
		Database:              database,
		DBTWorkgroup:          dbtWorkgroup,
		AdhocWorkgroup:        adhocWorkgroup,
		ResultsBucket:         resultsBucket,
		MaterializationMode:   matConfig.Mode,
		BronzeMaterialization: matConfig.Bronze,
		SilverMaterialization: matConfig.Silver,
		GoldMaterialization:   matConfig.Gold,
	}

	if err := config.GenerateEcosConfig(ecosConfigData, p.OutputPath); err != nil {
		return fmt.Errorf("failed to generate ecos config: %w", err)
	}
	utils.PrintDebug("Successfully generated .ecos.yaml")

	// Generate DBT configuration files using templates
	if err := p.generateDBTFiles(projectDir, userInput, matConfig); err != nil {
		return fmt.Errorf("failed to generate dbt files: %w", err)
	}

	return nil
}

func (p *AWSCURInitPlugin) CreateResources() error {
	userInput := p.Config

	// If user chose not to create resources, just return success
	if !userInput.CreateResources {
		utils.PrintInfo("Cloud resources skipped")
		return nil
	}

	ctx := context.Background()
	if userInput.DryRun {
		p.showDryRunPreview(userInput)
		return nil
	}

	// Setup AWS clients
	if userInput.AWSRegion == "" {
		return errors.New("aws_region is required for resource creation")
	}

	// Validate credentials and get account ID using the specified profile
	accountID, detectedRegion, err := initUtils.GetAWSAccountAndRegionWithProfile(ctx, 0, userInput.AWSProfile)
	if err != nil {
		return fmt.Errorf("failed to validate AWS credentials: %w", err)
	}

	// Use the specified region and profile for AWS config
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var awsCfg aws.Config
	if userInput.AWSProfile != "" && userInput.AWSProfile != "default" {
		awsCfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(userInput.AWSRegion), awsconfig.WithSharedConfigProfile(userInput.AWSProfile))
	} else {
		awsCfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(userInput.AWSRegion))
	}
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	// Store account ID in userInput for later use
	userInput.AccountID = accountID
	userInput.DetectedRegion = detectedRegion

	athenaClient := athena.NewFromConfig(awsCfg)
	s3Client := s3.NewFromConfig(awsCfg)

	// Extract config
	bucketName := userInput.ResultsBucket
	workgroups := []string{
		userInput.DBTWorkgroup,
		userInput.AdhocWorkgroup,
	}
	folders := []string{"dbt/", "adhoc/", "temp/"}
	projectName := userInput.ProjectName

	spinner := utils.NewSpinner("Creating AWS resources...")
	spinner.Start()

	var results []initTypes.InitResourceResult
	var hasError bool
	var hasPartial bool

	// Create S3 bucket
	bucketResult := p.createS3Bucket(s3Client, bucketName, userInput.AWSRegion)
	results = append(results, bucketResult)
	bucketExists := bucketResult.Status == initTypes.InitStatusCreated || bucketResult.Status == initTypes.InitStatusSkipped || bucketResult.Status == initTypes.InitStatusPartiallyCreated
	//nolint:staticcheck // QF1003: Using if/else instead of switch - we only need to check specific error/partial states, not exhaustive matching
	if bucketResult.Status == initTypes.InitStatusFailed {
		hasError = true
	} else if bucketResult.Status == initTypes.InitStatusPartiallyCreated {
		hasPartial = true
	}

	// Create S3 folders
	for _, folder := range folders {
		folderResult := p.createS3Folder(s3Client, bucketName, folder, bucketExists)
		results = append(results, folderResult)
		//nolint:staticcheck // QF1003: Using if/else instead of switch - we only need to check specific error/partial states, not exhaustive matching
		if folderResult.Status == initTypes.InitStatusFailed {
			hasError = true
		} else if folderResult.Status == initTypes.InitStatusPartiallyCreated {
			hasPartial = true
		}
	}

	// Create Athena workgroups
	for _, wg := range workgroups {
		wgResult := p.createAthenaWorkgroup(athenaClient, wg, bucketName, projectName, bucketExists)
		results = append(results, wgResult)
		//nolint:staticcheck // QF1003: Using if/else instead of switch - we only need to check specific error/partial states, not exhaustive matching
		if wgResult.Status == initTypes.InitStatusFailed {
			hasError = true
		} else if wgResult.Status == initTypes.InitStatusPartiallyCreated {
			hasPartial = true
		}
	}

	// Show results
	//nolint:gocritic // ifElseChain - if-else chain is clearer than switch for boolean conditions
	if hasError {
		spinner.Stop()
	} else if hasPartial {
		spinner.Success("AWS resources processed with warnings")
	} else {
		spinner.Success("AWS resources processed successfully")
	}

	p.showResourceSummary(results)

	// Show warning message if there are partial resources
	if hasPartial && !hasError {
		utils.PrintWarning("Some resources were partially created. Review warnings above and consider re-running 'ecos init' to complete configuration.")
	}

	if hasError {
		return errors.New("one or more resources failed to create")
	}

	return nil
}

func (p *AWSCURInitPlugin) CreateDirectoryStructure() error {
	return initUtils.SetupDirectories(p.OutputPath)
}

func (p *AWSCURInitPlugin) InitializeBaseFiles() error {
	return initUtils.SetupBaseFiles(p.OutputPath, "AWS CUR")
}

func (p *AWSCURInitPlugin) Validate(cfg map[string]any) error {
	return nil
}

func (p *AWSCURInitPlugin) Execute(ctx context.Context, cfg map[string]any) (*initTypes.PluginResult, error) {
	return &initTypes.PluginResult{
		Success: true,
		Message: "AWS init plugin executed successfully",
	}, nil
}

func (p *AWSCURInitPlugin) DownloadTransformModels() (string, error) {
	userInput := p.Config

	spinner := utils.NewSpinner("Downloading transform models")
	spinner.Start()
	defer spinner.Stop()

	ghClient, err := initUtils.NewGitHubClient()
	if err != nil {
		spinner.Error("Failed to create GitHub client")
		return "", fmt.Errorf("failed to create GitHub client: %w", err)
	}

	ctx := context.Background()
	destPath := filepath.Join(p.OutputPath, "transform", "dbt")

	version, err := ghClient.DownloadTransformModels(ctx, "aws_cur", userInput.ModelVersion, destPath)
	if err != nil {
		spinner.Error("Failed to download transform models")
		return "", fmt.Errorf("transform models download failed: %w", err)
	}

	spinner.Success(fmt.Sprintf("Transform models for aws_cur downloaded successfully (version: %s)", version))
	return version, nil
}

func (p *AWSCURInitPlugin) PostInitSummary() error {
	initUtils.PrintPostInitSummary()
	return nil
}

func (p *AWSCURInitPlugin) SetModelVersion(version string) error {
	p.Config.ModelVersion = version
	return nil
}

func (p *AWSCURInitPlugin) createS3Bucket(s3Client *s3.Client, bucketName, region string) initTypes.InitResourceResult {
	// Check if bucket exists
	_, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		return initTypes.InitResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucketName,
			Status: initTypes.InitStatusSkipped,
			Error:  "",
		}
	}

	// Create bucket
	var createBucketInput *s3.CreateBucketInput
	if region == "us-east-1" {
		createBucketInput = &s3.CreateBucketInput{Bucket: aws.String(bucketName)}
	} else {
		createBucketInput = &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
				LocationConstraint: s3Types.BucketLocationConstraint(region),
			},
		}
	}

	_, err = s3Client.CreateBucket(context.Background(), createBucketInput)
	if err != nil {
		return p.handleBucketError(bucketName, err)
	}

	// Bucket created successfully, now configure it
	var warnings []string

	// Enable versioning
	_, err = s3Client.PutBucketVersioning(context.Background(), &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3Types.VersioningConfiguration{
			Status: s3Types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to enable versioning for bucket %q: %v", bucketName, err))
	}

	// Configure lifecycle policy: delete adhoc query results after AdhocQueryRetentionDays and incomplete MPUs after IncompleteUploadCleanupDays
	lifecycleRules := getLifecycleRules()

	_, err = s3Client.PutBucketLifecycleConfiguration(context.Background(), &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &s3Types.BucketLifecycleConfiguration{
			Rules: lifecycleRules,
		},
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to configure lifecycle policy for bucket %q: %v. Bucket created but automatic cleanup (%d-day adhoc deletion, %d-day incomplete MPU cleanup) will not work.", bucketName, err, AdhocQueryRetentionDays, IncompleteUploadCleanupDays))
	}

	// Tag bucket as ecos-managed
	_, err = s3Client.PutBucketTagging(context.Background(), &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &s3Types.Tagging{
			TagSet: []s3Types.Tag{
				{Key: aws.String("ecos:managed"), Value: aws.String("true")},
				{Key: aws.String("ecos:project"), Value: aws.String(p.Config.ProjectName)},
			},
		},
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to tag bucket %q with ecos:managed and ecos:project tags: %v", bucketName, err))
	}

	// If bucket was created but some operations failed, return partial status
	if len(warnings) > 0 {
		return initTypes.InitResourceResult{
			Kind:    "S3 Bucket",
			Name:    bucketName,
			Status:  initTypes.InitStatusPartiallyCreated,
			Error:   "",
			Warning: strings.Join(warnings, "; "),
		}
	}

	// All operations succeeded
	return initTypes.InitResourceResult{
		Kind:   "S3 Bucket",
		Name:   bucketName,
		Status: initTypes.InitStatusCreated,
		Error:  "",
	}
}

func (p *AWSCURInitPlugin) createS3Folder(s3Client *s3.Client, bucketName, folder string, bucketExists bool) initTypes.InitResourceResult {
	if !bucketExists {
		return initTypes.InitResourceResult{
			Kind:   "S3 Folder",
			Name:   folder,
			Status: initTypes.InitStatusSkipped,
			Error:  "Bucket not accessible",
		}
	}

	// Check if folder exists
	_, err := s3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(folder),
	})
	if err == nil {
		return initTypes.InitResourceResult{
			Kind:   "S3 Folder",
			Name:   folder,
			Status: initTypes.InitStatusSkipped,
			Error:  "",
		}
	}

	// Create folder
	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(folder),
	})
	if err != nil {
		return initTypes.InitResourceResult{
			Kind:   "S3 Folder",
			Name:   folder,
			Status: initTypes.InitStatusFailed,
			Error:  fmt.Sprintf("Failed to create folder: %v", err),
		}
	}

	return initTypes.InitResourceResult{
		Kind:   "S3 Folder",
		Name:   folder,
		Status: initTypes.InitStatusCreated,
		Error:  "",
	}
}

func (p *AWSCURInitPlugin) createAthenaWorkgroup(
	athenaClient *athena.Client,
	workgroupName, bucketName, projectName string,
	bucketExists bool,
) initTypes.InitResourceResult {
	if !bucketExists {
		return initTypes.InitResourceResult{
			Kind:   "Athena Workgroup",
			Name:   workgroupName,
			Status: initTypes.InitStatusSkipped,
			Error:  "S3 bucket not accessible",
		}
	}

	// Check if workgroup exists
	_, err := athenaClient.GetWorkGroup(context.Background(), &athena.GetWorkGroupInput{
		WorkGroup: aws.String(workgroupName),
	})
	if err == nil {
		return initTypes.InitResourceResult{
			Kind:   "Athena Workgroup",
			Name:   workgroupName,
			Status: initTypes.InitStatusSkipped,
			Error:  "",
		}
	}

	// Create workgroup
	_, err = athenaClient.CreateWorkGroup(context.Background(), &athena.CreateWorkGroupInput{
		Name: aws.String(workgroupName),
		Configuration: &athenaTypes.WorkGroupConfiguration{
			ResultConfiguration: &athenaTypes.ResultConfiguration{
				OutputLocation: aws.String(fmt.Sprintf("s3://%s/%s/", bucketName, workgroupName)),
			},
			EnforceWorkGroupConfiguration:   aws.Bool(true),
			PublishCloudWatchMetricsEnabled: aws.Bool(true),
			RequesterPaysEnabled:            aws.Bool(false),
		},
		Description: aws.String(fmt.Sprintf("Workgroup created by ecos cli for project %s", projectName)),
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return initTypes.InitResourceResult{
				Kind:   "Athena Workgroup",
				Name:   workgroupName,
				Status: initTypes.InitStatusSkipped,
				Error:  "",
			}
		}
		return initTypes.InitResourceResult{
			Kind:   "Athena Workgroup",
			Name:   workgroupName,
			Status: initTypes.InitStatusFailed,
			Error:  fmt.Sprintf("Failed to create workgroup: %v", err),
		}
	}

	// Workgroup created successfully, now configure it
	var warnings []string

	// Construct ARN correctly
	wgARN := fmt.Sprintf(
		"arn:aws:athena:%s:%s:workgroup/%s",
		p.Config.AWSRegion,
		p.Config.AccountID,
		workgroupName,
	)

	// Tag workgroup as ecos-managed
	_, err = athenaClient.TagResource(context.Background(), &athena.TagResourceInput{
		ResourceARN: aws.String(wgARN),
		Tags: []athenaTypes.Tag{
			{Key: aws.String("ecos:managed"), Value: aws.String("true")},
			{Key: aws.String("ecos:project"), Value: aws.String(projectName)},
		},
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to tag workgroup %q (ARN: %s) with ecos:managed and ecos:project tags: %v", workgroupName, wgARN, err))
	}

	// If workgroup was created but some operations failed, return partial status
	if len(warnings) > 0 {
		return initTypes.InitResourceResult{
			Kind:    "Athena Workgroup",
			Name:    workgroupName,
			Status:  initTypes.InitStatusPartiallyCreated,
			Error:   "",
			Warning: strings.Join(warnings, "; "),
		}
	}

	// All operations succeeded
	return initTypes.InitResourceResult{
		Kind:   "Athena Workgroup",
		Name:   workgroupName,
		Status: initTypes.InitStatusCreated,
		Error:  "",
	}
}

func (p *AWSCURInitPlugin) handleBucketError(bucketName string, err error) initTypes.InitResourceResult {
	switch {
	case strings.Contains(err.Error(), "BucketAlreadyOwnedByYou"):
		return initTypes.InitResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucketName,
			Status: initTypes.InitStatusSkipped,
			Error:  "",
		}
	case strings.Contains(err.Error(), "BucketAlreadyExists"):
		return initTypes.InitResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucketName,
			Status: initTypes.InitStatusFailed,
			Error:  fmt.Sprintf("Bucket name %s is already taken globally", bucketName),
		}
	case strings.Contains(err.Error(), "AccessDenied"):
		return initTypes.InitResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucketName,
			Status: initTypes.InitStatusFailed,
			Error:  "Access denied - check AWS permissions",
		}
	default:
		return initTypes.InitResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucketName,
			Status: initTypes.InitStatusFailed,
			Error:  fmt.Sprintf("Failed to create bucket: %v", err),
		}
	}
}

// getLifecycleRules returns the lifecycle rules for S3 bucket query results
// - DeleteAdhocQueryResultsAfter30Days: Deletes objects in adhoc/ folder after AdhocQueryRetentionDays
// - DeleteIncompleteMultipartUploads: Aborts incomplete multipart uploads after IncompleteUploadCleanupDays (bucket-wide)
func getLifecycleRules() []s3Types.LifecycleRule {
	// Filter for adhoc folder only
	adhocPrefixFilter := &s3Types.LifecycleRuleFilterMemberPrefix{Value: "adhoc/"}
	// Empty prefix filter for bucket-wide rules (incomplete multipart uploads)
	emptyPrefixFilter := &s3Types.LifecycleRuleFilterMemberPrefix{Value: ""}

	return []s3Types.LifecycleRule{
		{
			ID:     aws.String("DeleteAdhocQueryResultsAfter30Days"),
			Status: s3Types.ExpirationStatusEnabled,
			Filter: adhocPrefixFilter,
			Expiration: &s3Types.LifecycleExpiration{
				Days: aws.Int32(AdhocQueryRetentionDays),
			},
		},
		{
			ID:     aws.String("DeleteIncompleteMultipartUploads"),
			Status: s3Types.ExpirationStatusEnabled,
			Filter: emptyPrefixFilter,
			AbortIncompleteMultipartUpload: &s3Types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int32(IncompleteUploadCleanupDays),
			},
		},
	}
}

func (p *AWSCURInitPlugin) showResourceSummary(results []initTypes.InitResourceResult) {
	utils.PrintSubHeader("📦 AWS Resources Summary")

	for _, res := range results {
		color := utils.ColorGreen
		label := "created"
		switch res.Status {
		case initTypes.InitStatusSkipped:
			color = utils.ColorYellow
			label = "skipped"
		case initTypes.InitStatusFailed:
			color = utils.ColorRed
			label = "failed"
		case initTypes.InitStatusPartiallyCreated:
			color = utils.ColorYellow
			label = "partially created"
		}
		fmt.Printf("  • %s%s%s %s (%s)\n", color, res.Kind, utils.ColorReset, res.Name, label)

		// Show error message for failed resources
		if res.Status == initTypes.InitStatusFailed && res.Error != "" {
			fmt.Printf("    %s%s%s\n", utils.ColorRed, res.Error, utils.ColorReset)
		}

		// Show warning message for partially created resources
		if res.Status == initTypes.InitStatusPartiallyCreated && res.Warning != "" {
			fmt.Printf("    %s⚠ Warning: %s%s\n", utils.ColorYellow, res.Warning, utils.ColorReset)
		}
	}
}

func (p *AWSCURInitPlugin) showDryRunPreview(userInput *AWSCURInput) {
	workgroups := []string{
		userInput.DBTWorkgroup,
		userInput.AdhocWorkgroup,
	}
	folders := []string{"dbt/", "adhoc/", "temp/"}

	utils.PrintDryRun("Would create the following resources:")
	fmt.Printf("  • S3 Bucket: s3://%s\n", userInput.ResultsBucket)
	fmt.Printf("  • Athena Workgroups: %s\n", strings.Join(workgroups, ", "))
	fmt.Printf("  • S3 Folders: %s\n", strings.Join(folders, ", "))
}

func (p *AWSCURInitPlugin) generateDBTFiles(destPath string, userInput *AWSCURInput, matConfig MaterializationConfig) error {
	// Use the actual resource names that were configured/created
	dbtWorkgroup := userInput.DBTWorkgroup
	resultsBucket := userInput.ResultsBucket

	database := fmt.Sprintf("%s_database", normalizeDatabaseName(userInput.ProjectName))

	// Prepare profiles template data
	profilesData := config.DBTProfilesTemplate{
		Profile:       "ecos-athena",
		Target:        "prod",
		AWSProfile:    userInput.AWSProfile,
		AWSRegion:     userInput.AWSRegion,
		ResultsBucket: resultsBucket,
		Database:      database,
		Workgroup:     dbtWorkgroup,
	}

	// Prepare project template data - use materialization settings from ecos config
	projectData := config.DBTProjectTemplate{
		Profile: "ecos-athena",
		DatasourceVars: []config.DatasourceVar{
			{Key: "cur_database", Value: userInput.CURDatabase},
			{Key: "cur_schema", Value: userInput.CURSchema},
			{Key: "cur_table", Value: userInput.CURTable},
		},
		IcebergEnabled:        false,
		BillingPeriodStart:    nil,
		BillingPeriodEnd:      nil,
		MaterializationMode:   matConfig.Mode,
		BronzeMaterialization: matConfig.Bronze,
		SilverMaterialization: matConfig.Silver,
		GoldMaterialization:   matConfig.Gold,
		UseIceberg:            false,
		EnablePartitioning:    true,
	}

	// Generate profiles.yml
	utils.PrintDebug("Generating dbt profiles.yml configuration file")
	if err := config.GenerateDBTProfiles(profilesData, destPath); err != nil {
		return fmt.Errorf("failed to generate profiles.yml: %w", err)
	}
	utils.PrintDebug("Successfully generated profiles.yml")

	// Generate dbt_project.yml (always overwrite)
	utils.PrintDebug("Generating dbt_project.yml configuration file")
	if err := config.GenerateDBTProject(projectData, destPath); err != nil {
		return fmt.Errorf("failed to generate dbt_project.yml: %w", err)
	}
	utils.PrintDebug("Successfully generated dbt_project.yml")

	return nil
}

// NewAWSCUR creates a new AWS CUR plugin instance.
func NewAWSCUR(force bool, outputPath string) (initTypes.InitPlugin, error) {
	return &AWSCURInitPlugin{
		Config:     &AWSCURInput{},
		Force:      force,
		OutputPath: outputPath,
	}, nil
}

// Self-register the plugin
func init() {
	registry.RegisterInitPlugin("aws_cur", NewAWSCUR)
}
