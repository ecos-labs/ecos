package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/core/transform"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
	"github.com/spf13/cobra"
)

const transformHelpText = `Transform cloud cost data using configured transformation tools.

This command provides a wrapper around transformation tools to process
cloud billing data with standardized commands and project-aware configuration.

The transformation tool is determined by the 'plugin' setting in your .ecos.yaml
transform configuration. Supported plugins include:
• dbt      - Data Build Tool for SQL transformations

All tool-specific arguments and flags are passed through directly. For help with
specific tool commands, use: ecos transform [command] --help

Examples:
  ecos transform run
  ecos transform run --select my_model
  ecos transform test --models tag:daily
  ecos transform seed --full-refresh

The transform command automatically detects the transformation tool configured
in your .ecos.yaml file and delegates to the appropriate plugin.`

// ParsedTransformArgs holds the parsed command line arguments
type ParsedTransformArgs struct {
	Command      string
	FilteredArgs []string
	ProjectDir   string
	IsDryRun     bool
	IgnoreDrift  bool
}

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:                "transform [command] [args...]",
	Short:              "Transform cloud cost data using configured transformation tools",
	Long:               transformHelpText,
	DisableFlagParsing: true, // Allow all flags to pass through to the tool
	SilenceUsage:       true, // Don't show usage/help on error
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTransformCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)
}

func parseTransformArgs(args []string) *ParsedTransformArgs {
	command := args[0]
	cmdArgs := args[1:]

	parsed := &ParsedTransformArgs{
		Command:     command,
		ProjectDir:  ".",
		IsDryRun:    false,
		IgnoreDrift: false,
	}

	// Simple flag parsing for ecos-specific flags
	filteredArgs := []string{command}
	for i := 0; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]
		if arg == "--project-dir" || arg == "-p" {
			if i+1 < len(cmdArgs) {
				parsed.ProjectDir = cmdArgs[i+1]
				i++ // skip the value
			}
			continue
		}
		if arg == "--dry-run" {
			parsed.IsDryRun = true
			continue
		}
		if arg == "--ignore-drift" {
			parsed.IgnoreDrift = true
			continue
		}
		// Handle verbose flag conflict: --verbose is for ecos, -v is for dbt
		if arg == "--verbose" {
			// This is ecos verbose, don't pass to dbt
			utils.SetVerbose(true)
			continue
		}
		// -v should pass through to dbt (for dbt version, etc.)
		filteredArgs = append(filteredArgs, arg)
	}

	parsed.FilteredArgs = filteredArgs
	return parsed
}

func runTransformCommand(args []string) error {
	// Handle special case: no arguments provided
	if len(args) == 0 {
		showTransformHelp()
		return nil
	}

	// Check if help is requested - if so, show help regardless of other flags
	if containsHelpFlag(args) {
		// If first arg is help flag, show general help
		if args[0] == "--help" || args[0] == "-h" {
			showTransformHelp()
			return nil
		}

		// If help flag is present but first arg is a command, show command-specific help
		// We need to extract the project dir first for command help
		projectDir := "."
		for i, arg := range args {
			if (arg == "--project-dir" || arg == "-p") && i+1 < len(args) {
				projectDir = args[i+1]
				break
			}
		}

		// Find the command (first non-flag argument)
		var command string
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				command = arg
				break
			}
		}

		if command != "" {
			return showCommandHelp(command, projectDir)
		}
		showTransformHelp()
		return nil
	}

	// Parse arguments and flags
	parsedArgs := parseTransformArgs(args)

	utils.PrintHeader(fmt.Sprintf("ecos transform %s", parsedArgs.Command))

	// Load project configuration
	configPath := filepath.Join(parsedArgs.ProjectDir, ".ecos.yaml")
	var ecosConfig *config.EcosConfig

	if utils.FileExists(configPath) {
		var err error
		ecosConfig, err = config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Check for drift between .ecos.yaml and dbt files (unless ignored)
		if !parsedArgs.IgnoreDrift {
			if err := checkConfigDrift(parsedArgs.ProjectDir); err != nil {
				return err
			}
		}
	} else {
		utils.PrintWarning("No .ecos.yaml found, using default configuration")
		ecosConfig = config.NewDefaultConfig()
	}

	// Determine the transform plugin
	pluginName := ecosConfig.Transform.Plugin
	if pluginName == "" {
		pluginName = "dbt" // Default fallback
	}

	// Instantiate the transform plugin directly
	var plugin types.TransformPlugin
	switch strings.ToLower(pluginName) {
	case "dbt":
		plugin = &transform.DBTTransformPlugin{}
	default:
		return fmt.Errorf("unsupported transform plugin: %s. Available plugins: dbt", pluginName)
	}

	// Build configuration for the plugin
	pluginConfig := plugin.BuildConfig(ecosConfig, parsedArgs.Command, parsedArgs.FilteredArgs[1:])
	pluginConfig["project_dir"] = parsedArgs.ProjectDir
	pluginConfig["verbose"] = IsVerbose()

	// Handle dry run case
	if parsedArgs.IsDryRun {
		return runTransformDryRun(plugin, parsedArgs.FilteredArgs, pluginConfig)
	}

	// Execute the transform command
	return runTransformExecute(plugin, parsedArgs.FilteredArgs, pluginConfig)
}

func showTransformHelp() {
	// Print the help text from the constant
	utils.PrintInfo(transformHelpText)

	// Add ecos-specific flags that aren't shown in the cobra command
	utils.PrintInfo("\necos-specific flags:")
	utils.PrintInfo("  --project-dir, -p    ecos project directory path (default: \".\")")
	utils.PrintInfo("  --dry-run            show what would be executed without running")
	utils.PrintInfo("  --ignore-drift       ignore configuration drift and proceed anyway")
	utils.PrintInfo("  --verbose            enable ecos verbose output")
}

func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func showCommandHelp(command string, projectDir string) error {
	// Load project configuration
	configPath := filepath.Join(projectDir, ".ecos.yaml")
	var ecosConfig *config.EcosConfig
	var err error

	if utils.FileExists(configPath) {
		ecosConfig, err = config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	} else {
		ecosConfig = config.NewDefaultConfig()
	}

	// Determine the transform plugin
	pluginName := ecosConfig.Transform.Plugin
	if pluginName == "" {
		pluginName = "dbt" // Default fallback
	}

	// Instantiate the transform plugin directly
	var plugin types.TransformPlugin
	switch strings.ToLower(pluginName) {
	case "dbt":
		plugin = &transform.DBTTransformPlugin{}
	default:
		return fmt.Errorf("unsupported transform plugin: %s. Available plugins: dbt", pluginName)
	}

	return plugin.ShowCommandHelp(command)
}

func runTransformDryRun(plugin types.TransformPlugin, args []string, config map[string]any) error {
	utils.PrintDryRun(fmt.Sprintf("Would execute plugin: %s", plugin.Name()))
	utils.PrintDryRun(fmt.Sprintf("Plugin version: %s", plugin.Version()))
	utils.PrintDryRun(fmt.Sprintf("Command: %s", strings.Join(args, " ")))

	// Show dbt project directory if available
	if dbtProjectDir, ok := config["dbt_project_dir"].(string); ok {
		utils.PrintDryRun(fmt.Sprintf("DBT project directory: %s", dbtProjectDir))
	}

	// Show configuration that would be passed to plugin
	utils.PrintDryRun("Plugin configuration:")
	for key, value := range config {
		if key != "args" { // Don't repeat args
			utils.PrintDryRun(fmt.Sprintf("  %s: %v", key, value))
		}
	}

	return nil
}

func runTransformExecute(plugin types.TransformPlugin, args []string, config map[string]any) error {
	// Step 1: Validate plugin configuration
	spinner := utils.NewSpinner("Validating plugin configuration")
	spinner.Start()

	if err := plugin.Validate(config); err != nil {
		spinner.Error("Plugin validation failed")
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	spinner.Success("Plugin configuration validated")

	// Step 2: Execute the plugin
	command := args[0]
	verbose, _ := config["verbose"].(bool)

	// Transform tools (dbt, Spark, Airflow, etc.) are typically external CLI tools
	// that benefit from streaming output in real-time, so this approach works well for all
	displayArgs := command
	if verbose && len(args) > 1 {
		displayArgs = strings.Join(args, " ")
	}
	utils.PrintProgress(fmt.Sprintf("Running %s %s", plugin.Name(), displayArgs))

	result, err := plugin.Execute(context.Background(), config)
	if err != nil {
		return fmt.Errorf("%s %s failed: %w", plugin.Name(), command, err)
	}

	if !result.Success {
		return fmt.Errorf("%s %s failed: %s", plugin.Name(), command, result.Message)
	}

	utils.PrintSuccess(fmt.Sprintf("%s %s completed successfully", plugin.Name(), command))
	return nil
}

// checkConfigDrift checks if dbt files match .ecos.yaml and exits if drift is detected
func checkConfigDrift(projectDir string) error {
	// Detect drift
	report, err := config.DetectDriftFromEcosConfig(projectDir)
	if err != nil {
		// If drift detection fails, just warn and continue
		utils.PrintDebug(fmt.Sprintf("Drift check skipped: %v", err))
		return nil
	}

	// If no drift, continue normally
	if !report.HasChanges {
		return nil
	}

	// Drift detected - show error and exit
	fmt.Println()
	utils.PrintError("Configuration drift detected")

	// Show which files have drift
	var driftedFiles []string
	for filename, fileReport := range report.Files {
		if fileReport.HasChanges {
			driftedFiles = append(driftedFiles, filename)
		}
	}
	utils.PrintWarning(fmt.Sprintf("Files out of sync with .ecos.yaml: %s", strings.Join(driftedFiles, ", ")))
	fmt.Println()

	// Show instructions
	utils.PrintInfo("To fix this issue:")
	utils.PrintInfo("  1. Run 'ecos config diff' to see what changed")
	utils.PrintInfo("  2. Run 'ecos config generate' to sync files from .ecos.yaml")
	fmt.Println()
	utils.PrintInfo("Or ignore drift and proceed:")
	utils.PrintInfo("  ecos transform run --ignore-drift")
	fmt.Println()

	return errors.New("configuration drift detected, please sync your files")
}
