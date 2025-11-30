package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ecos project configuration",
	Long: `Manage ecos project configuration and sync with generated files.

The config command helps you manage your .ecos.yaml configuration and ensure
that generated files (like dbt_project.yml and profiles.yml) stay in sync.

Available subcommands:
  diff      Show differences between .ecos.yaml and generated files
  generate  Regenerate dbt files from .ecos.yaml

Examples:
  ecos config diff
  ecos config generate
  ecos config generate --project-dir ./my-project`,
}

// configDiffCmd represents the config diff command
var configDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show configuration drift between .ecos.yaml and generated files",
	Long: `Show differences between what .ecos.yaml says files should be and what they actually are.

This command compares your .ecos.yaml configuration with the generated dbt files
(dbt_project.yml and profiles.yml) and shows any differences.

Examples:
  ecos config diff
  ecos config diff --project-dir ./my-project`,
	RunE: runConfigDiff,
}

// configGenerateCmd represents the config generate command
var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Regenerate dbt configuration files from .ecos.yaml",
	Long: `Regenerate dbt_project.yml and profiles.yml from your .ecos.yaml configuration.

This command reads your .ecos.yaml file and regenerates the dbt configuration files
to match. This is useful when you've manually edited .ecos.yaml and want to sync
the generated files.

Examples:
  ecos config generate
  ecos config generate --project-dir ./my-project
  ecos config generate --force  # Skip confirmation prompt`,
	RunE: runConfigGenerate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configDiffCmd)
	configCmd.AddCommand(configGenerateCmd)

	// Add flags
	configDiffCmd.Flags().StringP("project-dir", "p", ".", "ecos project directory path")
	configGenerateCmd.Flags().StringP("project-dir", "p", ".", "ecos project directory path")
	configGenerateCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

func runConfigDiff(cmd *cobra.Command, args []string) error {
	projectDir, _ := cmd.Flags().GetString("project-dir")

	utils.PrintHeader("ecos config diff")

	// Check if .ecos.yaml exists
	configPath := filepath.Join(projectDir, ".ecos.yaml")
	if !utils.FileExists(configPath) {
		utils.PrintError("No .ecos.yaml found in project directory")
		return fmt.Errorf(".ecos.yaml not found in %s", projectDir)
	}

	// Detect drift
	report, err := config.DetectDriftFromEcosConfig(projectDir)
	if err != nil {
		utils.PrintError("Failed to check configuration drift")
		return fmt.Errorf("drift detection failed: %w", err)
	}

	// If no drift, show success message
	if !report.HasChanges {
		utils.PrintSuccess("No configuration drift detected")
		utils.PrintInfo("All files are in sync with .ecos.yaml")
		return nil
	}

	// Show drift details
	utils.PrintWarning("Configuration drift detected")
	fmt.Println()

	// Show which files have drift
	var driftedFiles []string
	for filename, fileReport := range report.Files {
		if fileReport.HasChanges {
			driftedFiles = append(driftedFiles, filename)
		}
	}
	utils.PrintInfo(fmt.Sprintf("Files out of sync: %d", len(driftedFiles)))
	fmt.Println()

	// Show diff for each changed file
	for filename, fileReport := range report.Files {
		if fileReport.HasChanges {
			utils.PrintSubHeader(fmt.Sprintf("📄 %s", filename))
			fmt.Println(fileReport.Diff)
			fmt.Println()
		}
	}

	// Show instructions
	utils.PrintInfo("To fix drift, run:")
	utils.PrintInfo("  ecos config generate")
	fmt.Println()

	return nil
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	projectDir, _ := cmd.Flags().GetString("project-dir")

	utils.PrintHeader("ecos config generate")

	// Check if .ecos.yaml exists
	configPath := filepath.Join(projectDir, ".ecos.yaml")
	if !utils.FileExists(configPath) {
		utils.PrintError("No .ecos.yaml found in project directory")
		return fmt.Errorf(".ecos.yaml not found in %s", projectDir)
	}

	// Warn user about overwriting files
	utils.PrintWarning("This will overwrite existing dbt configuration files")
	utils.PrintInfo("Files that will be regenerated: dbt_project.yml, profiles.yml")

	// Check if --force flag is set
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		confirm := utils.ConfirmPrompt("Do you want to continue")
		if !confirm {
			utils.PrintWarning("Configuration generation cancelled.")
			return nil
		}
	}

	// Load .ecos.yaml
	spinner := utils.NewSpinner("Loading configuration")
	spinner.Start()

	ecosConfig, err := config.LoadConfig(configPath)
	if err != nil {
		spinner.Error("Failed to load configuration")
		return fmt.Errorf("failed to load .ecos.yaml: %w", err)
	}

	spinner.Success("Configuration loaded")

	// Extract template data
	spinner = utils.NewSpinner("Extracting dbt configuration")
	spinner.Start()

	dbtProjectData, dbtProfilesData, err := config.ExtractDBTDataFromEcosConfig(ecosConfig, projectDir)
	if err != nil {
		spinner.Error("Failed to extract dbt data")
		return fmt.Errorf("failed to extract dbt data: %w", err)
	}

	spinner.Success("DBT configuration extracted")

	// Determine dbt directory
	dbtDir := ecosConfig.Transform.DBT.ProjectDir
	if dbtDir == "" {
		dbtDir = filepath.Join(projectDir, "transform", "dbt")
	} else if !filepath.IsAbs(dbtDir) {
		dbtDir = filepath.Join(projectDir, dbtDir)
	}

	// Regenerate dbt_project.yml
	spinner = utils.NewSpinner("Generating dbt_project.yml")
	spinner.Start()

	if err := config.GenerateDBTProject(dbtProjectData, dbtDir); err != nil {
		spinner.Error("Failed to generate dbt_project.yml")
		return fmt.Errorf("failed to generate dbt_project.yml: %w", err)
	}

	spinner.Success("dbt_project.yml generated")

	// Regenerate profiles.yml
	spinner = utils.NewSpinner("Generating profiles.yml")
	spinner.Start()

	if err := config.GenerateDBTProfiles(dbtProfilesData, dbtDir); err != nil {
		spinner.Error("Failed to generate profiles.yml")
		return fmt.Errorf("failed to generate profiles.yml: %w", err)
	}

	spinner.Success("profiles.yml generated")

	fmt.Println()
	utils.PrintSuccess("Configuration files regenerated successfully")
	utils.PrintInfo(fmt.Sprintf("Files updated in: %s", dbtDir))

	return nil
}
