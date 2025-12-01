package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ecos-labs/ecos/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos/code/cli/plugins/types"
	"github.com/ecos-labs/ecos/code/cli/utils"
	"github.com/spf13/cobra"

	// Import init plugins to trigger plugin self-registration
	_ "github.com/ecos-labs/ecos/code/cli/plugins/core/init"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ecos project for cloud cost analysis",
	Long: `Initialize a new ecos project for cloud cost analysis with interactive setup.

This command validates prerequisites and creates resources needed to transform
cloud billing data (like AWS CUR) using transform plugins like dbt with appropriate adapters.

Key features:
• Validates required tools
• Configures connection to existing Cost and Usage Reports
• Creates Athena workgroups for dbt transformations
• Sets up S3 bucket structure for query results
• Generates project configuration and directory structure

The setup creates:

.ecos.yaml configuration file with cloud provider settings
Directory structure for plugins, models, and outputs
Project-specific cloud resources for data transformation`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "overwrite existing files without prompting")
	initCmd.Flags().StringP("output", "o", ".", "output directory for the project")

	initCmd.Flags().StringP("source", "s", "", "data source to configure (aws_cur, aws_focus)")
	initCmd.Flags().StringP("model-version", "m", "latest", "version of ecos models to use")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Parse flags
	force, _ := cmd.Flags().GetBool("force")
	outputPath, _ := cmd.Flags().GetString("output")

	dataSource, _ := cmd.Flags().GetString("source")
	modelVersion, _ := cmd.Flags().GetString("model-version")

	utils.PrintHeader("🚀 ecos init")

	// Step 1: Check for existing .ecos.yaml
	configPath := filepath.Join(outputPath, ".ecos.yaml")
	if utils.FileExists(configPath) {
		if force {
			utils.PrintWarning(fmt.Sprintf("Overwriting existing project in '%s' (--force flag used)", outputPath))
			utils.PrintInfo("This will regenerate .ecos.yaml and dbt configuration files (dbt_project.yml, profiles.yml)")
		} else {
			utils.PrintWarning(fmt.Sprintf("An ecos project already exists in '%s'", outputPath))
			utils.PrintInfo("Overwriting will regenerate .ecos.yaml and dbt configuration files (dbt_project.yml, profiles.yml)")
			confirm := utils.ConfirmPrompt("Do you want to continue and overwrite the existing project")
			if !confirm {
				utils.PrintWarning("Project initialization cancelled.")
				return nil
			}
		}
	}
	// Data source selection and plugin instantiation
	sourceOptions := map[int]string{
		0: "aws_cur",
		1: "aws_focus",
	}

	if dataSource == "" {
		displayOptions := []string{
			"aws_cur                   (AWS Cost and Usage Report - CUR legacy and CUR 2.0)",
			"aws_focus                 (AWS FinOps Open Cost and Usage Report - FOCUS 1.2) - coming soon",
		}

		utils.PrintSubHeader("📊 Data Source Selection")
		utils.PrintInfo("Select how ecos will connect to and analyze your cloud cost data")
		fmt.Println()

		i, _, err := utils.Select("Data Source", displayOptions, 0, true, false)
		if err != nil {
			return fmt.Errorf("data source selection cancelled: %w", err)
		}

		dataSource = sourceOptions[i]
	}

	// Load plugin from registry
	initPlugin, err := registry.LoadInitPlugin(dataSource, force, outputPath)
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	// Run interactive setup - plugin fills its own config
	if err := initPlugin.RunInteractiveSetup(); err != nil {
		return fmt.Errorf("interactive setup failed: %w", err)
	}

	// Add model version to plugin config if specified via flag
	if modelVersion != "" && modelVersion != "latest" {
		if err := initPlugin.SetModelVersion(modelVersion); err != nil {
			return fmt.Errorf("failed to set model version: %w", err)
		}
	}

	// Validate prerequisites
	if err := initPlugin.ValidatePrerequisites(); err != nil {
		return fmt.Errorf("prerequisite validation failed: %w", err)
	}

	// Step 4: Create project structure, configs, resources (provider-specific)
	return runInitExecute(initPlugin)
}

func runInitExecute(plugin types.InitPlugin) error {
	// Always use full step weights to show true completion percentage
	stepWeights := []int{5, 5, 55, 30, 5}
	progress := utils.NewWeightedProgressBar(stepWeights, "Setting up project")

	// Step 1: Directory structure
	if err := plugin.CreateDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	progress.AdvanceStep(0, "Directory structure created")

	// Step 2: Base files
	if err := plugin.InitializeBaseFiles(); err != nil {
		return fmt.Errorf("failed to create base files: %w", err)
	}
	progress.AdvanceStep(1, "Project documentation created")

	// Step 3: Cloud resources
	resourceCreationFailed := false
	if err := plugin.CreateResources(); err != nil {
		resourceCreationFailed = true
		utils.PrintWarning(fmt.Sprintf("Resource creation failed: %v", err))
		utils.PrintWarning("You will need to create these resources manually or run with proper credentials")
	} else {
		progress.AdvanceStep(2, "Cloud resources processed")
	}

	// Step 4: Transform models
	transformFailed := false
	if !resourceCreationFailed {
		// Don't set a default model_version - let the plugin fetch the latest release

		version, err := plugin.DownloadTransformModels()
		if err != nil {
			transformFailed = true
			utils.PrintWarning(fmt.Sprintf("Transform models download failed: %v", err))
			utils.PrintWarning("You will need to set up transform models manually")
		} else {
			// Update the plugin's config with the downloaded version
			if err := plugin.SetModelVersion(version); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to set model version: %v", err))
			}
			progress.AdvanceStep(3, "")
		}
	}

	// Step 5: Configuration
	if !resourceCreationFailed && !transformFailed {
		if err := plugin.GenerateConfig(); err != nil {
			utils.PrintWarning("Project setup completed with warnings (configuration failed)")
			return fmt.Errorf("failed to generate config: %w", err)
		}
		progress.AdvanceStep(4, "Project configuration and customization completed")
	} else {
		utils.PrintWarning("Project setup completed with warnings")
	}

	progress.Finish()

	_ = plugin.PostInitSummary()
	return nil
}
