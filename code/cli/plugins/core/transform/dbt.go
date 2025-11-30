package transform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ecosconfig "github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
	"github.com/subosito/gotenv"
)

// DBTTransformPlugin implements the TransformPlugin interface for dbt
type DBTTransformPlugin struct {
	ProjectDir string
}

// Name returns the plugin name
func (p *DBTTransformPlugin) Name() string {
	return "dbt"
}

// Version returns the plugin version
func (p *DBTTransformPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *DBTTransformPlugin) Description() string {
	return "dbt (data build tool) transformation plugin for ecos"
}

// Author returns the plugin author
func (p *DBTTransformPlugin) Author() string {
	return "ecos team"
}

// IsCore returns true since this is a core plugin
func (p *DBTTransformPlugin) IsCore() bool {
	return true
}

// Documentation returns plugin documentation
func (p *DBTTransformPlugin) Documentation() string {
	return "dbt plugin for transforming cloud cost data using SQL models"
}

// Type returns the plugin type (required by CorePlugin interface)
func (p *DBTTransformPlugin) Type() types.PluginType {
	return types.PluginTypeTransform
}

// Validate validates the plugin configuration (required by CorePlugin interface)
func (p *DBTTransformPlugin) Validate(config map[string]any) error {
	return p.ValidateEnvironment(config)
}

// Execute runs the plugin with the given configuration (required by CorePlugin interface)
func (p *DBTTransformPlugin) Execute(ctx context.Context, config map[string]any) (*types.PluginResult, error) {
	// Default to running dbt run if no specific command is provided
	command := "run"
	if cmd, ok := config["command"].(string); ok && cmd != "" {
		command = cmd
	}

	var args []string
	if cmdArgs, ok := config["args"].([]string); ok {
		args = cmdArgs
	}

	// Prepare environment (install dependencies if needed) before executing command
	if err := p.PrepareEnvironment(ctx, config); err != nil {
		return &types.PluginResult{
			Success: false,
			Message: fmt.Sprintf("Failed to prepare dbt environment: %v", err),
		}, err
	}

	err := p.ExecuteCommand(ctx, command, args, config)
	if err != nil {
		return &types.PluginResult{
			Success: false,
			Message: fmt.Sprintf("dbt %s failed: %v", command, err),
		}, err
	}

	// Build success message with flags if any were provided
	successMsg := p.buildSuccessMessage(command, args)

	return &types.PluginResult{
		Success: true,
		Message: successMsg,
	}, nil
}

// TransformEngine returns the transformation engine name
func (p *DBTTransformPlugin) TransformEngine() string {
	return "dbt"
}

// ValidatePrerequisites checks if dbt is installed
func (p *DBTTransformPlugin) ValidatePrerequisites() error {
	// Check if dbt is installed
	if _, err := exec.LookPath("dbt"); err != nil {
		return errors.New("dbt is not installed or not in PATH")
	}
	return nil
}

// ValidateEnvironment checks if the dbt environment is properly configured
func (p *DBTTransformPlugin) ValidateEnvironment(config map[string]any) error {
	// Get project directory
	projectDir := p.getProjectDir(config)

	// Check if dbt_project.yml exists
	dbtProjectPath := filepath.Join(projectDir, "dbt_project.yml")
	if !utils.FileExists(dbtProjectPath) {
		return fmt.Errorf("dbt_project.yml not found in %s", projectDir)
	}

	return nil
}

// ValidateProjectStructure validates the dbt project structure
func (p *DBTTransformPlugin) ValidateProjectStructure(config map[string]any) error {
	projectDir := p.getProjectDir(config)

	// Check required files and directories
	requiredPaths := []string{
		"dbt_project.yml",
		"models",
		"seeds",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(projectDir, path)
		if !utils.FileExists(fullPath) && !utils.DirectoryExists(fullPath) {
			return fmt.Errorf("required path not found: %s", fullPath)
		}
	}

	return nil
}

// ValidateConnection validates connection to the data warehouse
func (p *DBTTransformPlugin) ValidateConnection(config map[string]any) error {
	projectDir := p.getProjectDir(config)

	// Run dbt debug to check connection
	cmd := exec.CommandContext(context.Background(), "dbt", "debug")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dbt connection validation failed: %s", string(output))
	}

	return nil
}

// GetProjectPath returns the path to the dbt project directory
func (p *DBTTransformPlugin) GetProjectPath() string {
	if p.ProjectDir != "" {
		return p.ProjectDir
	}
	return filepath.Join(".", "transform", "dbt")
}

// GetSupportedCommands returns all dbt commands supported
func (p *DBTTransformPlugin) GetSupportedCommands() []string {
	return []string{
		"run", "test", "compile", "parse", "docs", "deps", "seed", "snapshot",
		"source", "freshness", "clean", "debug", "list", "show", "run-operation",
	}
}

// ExecuteCommand runs a dbt command with the given arguments
func (p *DBTTransformPlugin) ExecuteCommand(ctx context.Context, command string, args []string, config map[string]any) error {
	projectDir := p.getProjectDir(config)

	// Convert to absolute path for dbt
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for project directory: %w", err)
	}

	// Ensure the project directory exists
	if !utils.DirectoryExists(absProjectDir) {
		return fmt.Errorf("dbt project directory does not exist: %s", absProjectDir)
	}

	// Build dbt command - just pass through all arguments as-is
	cmdArgs := []string{command}

	// Add profiles-dir flag to use the project directory for profiles
	cmdArgs = append(cmdArgs, "--profiles-dir", absProjectDir)

	cmdArgs = append(cmdArgs, args...)

	// Execute dbt command (safe: using hardcoded "dbt" command)
	cmd := exec.CommandContext(ctx, "dbt", cmdArgs...) // #nosec G204
	cmd.Dir = absProjectDir

	// Inherit the current process environment (including loaded .env variables)
	// Note: os.Environ() captures the current state including any variables loaded by gotenv
	cmd.Env = os.Environ()

	// Capture both stdout and stderr to check for dependency errors while still showing output to user
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	if err != nil {
		// Combine stdout and stderr outputs to analyze the complete error message
		// This allows us to detect dependency-related errors that might appear in either stream
		// and automatically attempt to resolve them by running 'dbt deps'
		combinedOutput := stdoutBuf.String() + stderrBuf.String()
		if strings.Contains(combinedOutput, "dbt deps") && strings.Contains(combinedOutput, "install package dependencies") {
			utils.PrintWarning("dbt detected missing dependencies, installing them...")
			if prepErr := p.prepareEnvironmentWithOptions(ctx, config, true); prepErr != nil {
				return fmt.Errorf("failed to install dependencies: %w", prepErr)
			}
			utils.PrintSuccess("Dependencies installed, retrying original command...")
			return p.ExecuteCommand(ctx, command, args, config)
		}
	}

	return err
}

// PrepareEnvironment sets up the dbt execution environment
func (p *DBTTransformPlugin) PrepareEnvironment(ctx context.Context, config map[string]any) error {
	// Load .env file if it exists
	if err := p.loadEnvironmentFile(config); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to load .env file: %v", err))
		// Don't fail - just warn and continue
	}

	return p.prepareEnvironmentWithOptions(ctx, config, false)
}

// prepareEnvironmentWithOptions sets up the dbt execution environment with optional force install
func (p *DBTTransformPlugin) prepareEnvironmentWithOptions(ctx context.Context, config map[string]any, forceInstall bool) error {
	projectDir := p.getProjectDir(config)

	// Install dependencies if packages.yml exists
	packagesPath := filepath.Join(projectDir, "packages.yml")
	if !utils.FileExists(packagesPath) {
		utils.PrintWarning("Skipping dbt deps install, missing packages.yml")
		return nil
	}

	// Check if we need to install dependencies (skip check if forcing install)
	if !forceInstall && !p.needsDependencyInstall(projectDir) {
		utils.PrintSuccess("dbt dependencies are up to date")
		return nil
	}

	spinner := utils.NewSpinner("Installing dbt dependencies")
	spinner.Start()

	cmd := exec.CommandContext(ctx, "dbt", "deps")
	cmd.Dir = projectDir

	if err := cmd.Run(); err != nil {
		spinner.Error("dbt deps failed")
		utils.PrintWarning(fmt.Sprintf("Failed to install dbt dependencies: %v", err))
		utils.PrintWarning("Continuing anyway - you may need to run 'dbt deps' manually")
		// Don't return error - just warn and continue
	} else {
		spinner.Success("dbt dependencies installed")
	}

	return nil
}

// needsDependencyInstall checks if dbt dependencies need to be installed
func (p *DBTTransformPlugin) needsDependencyInstall(projectDir string) bool {
	lockPath := filepath.Join(projectDir, "package-lock.yml")
	dbtPackagesDir := filepath.Join(projectDir, "dbt_packages")

	// Basic checks: if no lock file exists or dbt_packages directory doesn't exist, need to install
	lockFileExists := utils.FileExists(lockPath)
	dbtPackagesDirExists := utils.DirectoryExists(dbtPackagesDir)

	return !lockFileExists || !dbtPackagesDirExists
}

// PostTransformSummary prints a summary after command execution
func (p *DBTTransformPlugin) PostTransformSummary(command string, config map[string]any) error {
	// Get args from config to build enhanced message
	var args []string
	if cmdArgs, ok := config["args"].([]string); ok {
		args = cmdArgs
	}

	// Build enhanced success message
	successMsg := p.buildSuccessMessage(command, args)
	utils.PrintSuccess(successMsg)

	projectDir := p.getProjectDir(config)
	utils.PrintInfo(fmt.Sprintf("Project directory: %s", projectDir))

	if target, ok := config["target"].(string); ok && target != "" {
		utils.PrintInfo(fmt.Sprintf("Target: %s", target))
	}

	return nil
}

// ShowCommandHelp displays help information for a specific dbt command
func (p *DBTTransformPlugin) ShowCommandHelp(command string) error {
	utils.PrintSubHeader(fmt.Sprintf("🔧 dbt %s", command))

	// Show command description
	p.showCommandDescription(command)

	fmt.Printf("\n%sUsage:%s\n", utils.ColorYellow, utils.ColorReset)
	fmt.Printf("  ecos transform %s [dbt-flags...]\n\n", command)

	fmt.Printf("%secos-specific flags:%s\n", utils.ColorYellow, utils.ColorReset)
	fmt.Printf("  --project-dir, -p    ecos project directory path (default: \".\")\n")
	fmt.Printf("  --dry-run           show what would be executed without running\n\n")

	fmt.Printf("%sdbt flags:%s\n", utils.ColorYellow, utils.ColorReset)
	fmt.Printf("  All dbt %s flags are supported and passed through directly.\n\n", command)

	// Show common flags for the specific command
	p.showCommonFlags(command)

	fmt.Printf("  For complete dbt flag documentation, run: %sdbt %s --help%s\n\n", utils.ColorGreen, command, utils.ColorReset)

	fmt.Printf("%sExamples:%s\n", utils.ColorYellow, utils.ColorReset)
	p.showCommandExamples(command)

	return nil
}

func (p *DBTTransformPlugin) showCommandDescription(command string) {
	switch command {
	case "run":
		utils.PrintInfo("Execute dbt models")
	case "test":
		utils.PrintInfo("Run dbt tests")
	case "seed":
		utils.PrintInfo("Load seed data")
	case "compile":
		utils.PrintInfo("Compile dbt models without running")
	case "docs":
		utils.PrintInfo("Generate or serve dbt documentation")
		utils.PrintInfo("Subcommands: generate, serve")
	case "deps":
		utils.PrintInfo("Install dbt dependencies")
	default:
		utils.PrintInfo(fmt.Sprintf("Run dbt %s command", command))
		utils.PrintInfo(fmt.Sprintf("Run 'dbt %s --help' for detailed information", command))
	}
}

func (p *DBTTransformPlugin) showCommonFlags(command string) {
	fmt.Printf("  Common dbt %s flags:\n", command)
	switch command {
	case "run":
		fmt.Printf("    --select MODEL       Run only specified models\n")
		fmt.Printf("    --exclude MODEL      Exclude specified models\n")
		fmt.Printf("    --full-refresh       Perform full refresh of incremental models\n")
		fmt.Printf("    --target TARGET      Specify target environment (dev, prod, etc.)\n")
		fmt.Printf("    --vars '{\"key\":\"val\"}'  Pass variables to dbt\n")
		fmt.Printf("    --threads N          Number of threads to use\n")
	case "test":
		fmt.Printf("    --select MODEL       Test only specified models\n")
		fmt.Printf("    --exclude MODEL      Exclude specified models from testing\n")
		fmt.Printf("    --target TARGET      Specify target environment (dev, prod, etc.)\n")
		fmt.Printf("    --vars '{\"key\":\"val\"}'  Pass variables to dbt\n")
	case "seed":
		fmt.Printf("    --select SEED        Load only specified seeds\n")
		fmt.Printf("    --full-refresh       Drop and recreate seed tables\n")
		fmt.Printf("    --target TARGET      Specify target environment (dev, prod, etc.)\n")
	case "compile":
		fmt.Printf("    --select MODEL       Compile only specified models\n")
		fmt.Printf("    --exclude MODEL      Exclude specified models\n")
		fmt.Printf("    --target TARGET      Specify target environment (dev, prod, etc.)\n")
		fmt.Printf("    --vars '{\"key\":\"val\"}'  Pass variables to dbt\n")
	case "docs":
		fmt.Printf("    --target TARGET      Specify target environment (dev, prod, etc.)\n")
		fmt.Printf("    --port PORT          Port for docs serve (default: 8080)\n")
	case "deps":
		fmt.Printf("    No common flags for deps command\n")
	default:
		fmt.Printf("    See 'dbt %s --help' for available flags\n", command)
	}
	fmt.Println()
}

func (p *DBTTransformPlugin) showCommandExamples(command string) {
	switch command {
	case "run":
		fmt.Printf("  ecos transform run\n")
		fmt.Printf("  ecos transform run --select my_model\n")
		fmt.Printf("  ecos transform run --exclude tag:staging\n")
		fmt.Printf("  ecos transform run --full-refresh\n")
	case "test":
		fmt.Printf("  ecos transform test\n")
		fmt.Printf("  ecos transform test --select my_model\n")
		fmt.Printf("  ecos transform test --models tag:daily\n")
	case "seed":
		fmt.Printf("  ecos transform seed\n")
		fmt.Printf("  ecos transform seed --select my_seed\n")
		fmt.Printf("  ecos transform seed --full-refresh\n")
	case "compile":
		fmt.Printf("  ecos transform compile\n")
		fmt.Printf("  ecos transform compile --select my_model\n")
	case "docs":
		fmt.Printf("  ecos transform docs generate\n")
		fmt.Printf("  ecos transform docs serve\n")
	case "deps":
		fmt.Printf("  ecos transform deps\n")
	default:
		fmt.Printf("  ecos transform %s\n", command)
		fmt.Printf("  ecos transform %s --help\n", command)
	}
}

// buildSuccessMessage creates a detailed success message including flags used
func (p *DBTTransformPlugin) buildSuccessMessage(command string, args []string) string {
	baseMsg := fmt.Sprintf("dbt %s completed successfully", command)

	if len(args) == 0 {
		return baseMsg
	}

	// Filter out common flags to show in the success message
	var importantFlags []string

	for i, arg := range args {
		switch arg {
		case "--select", "-s":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--select %s", args[i+1]))
			}
		case "--exclude":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--exclude %s", args[i+1]))
			}
		case "--full-refresh":
			importantFlags = append(importantFlags, "--full-refresh")
		case "--target", "-t":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--target %s", args[i+1]))
			}
		case "--vars":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--vars %s", args[i+1]))
			}
		case "--threads":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--threads %s", args[i+1]))
			}
		case "--models", "-m":
			if i+1 < len(args) {
				importantFlags = append(importantFlags, fmt.Sprintf("--models %s", args[i+1]))
			}
		}
	}

	if len(importantFlags) > 0 {
		return fmt.Sprintf("%s with %s", baseMsg, strings.Join(importantFlags, ", "))
	}

	return baseMsg
}

// BuildConfig builds dbt-specific configuration from ecos config
func (p *DBTTransformPlugin) BuildConfig(ecosConfig any, command string, args []string) map[string]any {
	config := make(map[string]any)

	// Add basic configuration
	config["command"] = command
	config["args"] = args

	// Cast ecosConfig to the expected type
	if cfg, ok := ecosConfig.(*ecosconfig.EcosConfig); ok && cfg.Transform.Plugin == "dbt" {
		// Add other dbt-specific config
		if cfg.Transform.DBT.Profile != "" {
			config["profile"] = cfg.Transform.DBT.Profile
		}
		if cfg.Transform.DBT.Target != "" {
			config["target"] = cfg.Transform.DBT.Target
		}
		if cfg.Transform.DBT.Vars != nil {
			config["vars"] = cfg.Transform.DBT.Vars
		}

		// Set dbt project directory - prioritize explicit config from .ecos.yaml
		if cfg.Transform.DBT.ProjectDir != "" {
			config["dbt_project_dir"] = cfg.Transform.DBT.ProjectDir
		}

		// Add transform-specific configuration from .ecos.yaml
		if cfg.Transform.Config != nil {
			for k, v := range cfg.Transform.Config {
				config[k] = v
			}
		}
	}

	return config
}

// loadEnvironmentFile loads environment variables from .env file in dbt project directory
func (p *DBTTransformPlugin) loadEnvironmentFile(config map[string]any) error {
	dbtProjectDir := p.getProjectDir(config)

	// Convert to absolute path for better reliability
	absDbtProjectDir, err := filepath.Abs(dbtProjectDir)
	if err != nil {
		absDbtProjectDir = dbtProjectDir // fallback to relative path
	}

	envPath := filepath.Join(absDbtProjectDir, ".env")

	if !utils.FileExists(envPath) {
		return nil // No .env file found, skip silently
	}

	// Load the .env file
	if err := gotenv.Load(envPath); err != nil {
		return fmt.Errorf("failed to load .env file from %s: %w", envPath, err)
	}
	return nil
}

// getProjectDir returns the project directory using centralized DBT config logic
func (p *DBTTransformPlugin) getProjectDir(config map[string]any) string {
	// First check if dbt_project_dir is explicitly set in config
	if dbtProjectDir, ok := config["dbt_project_dir"].(string); ok && dbtProjectDir != "" {
		return dbtProjectDir
	}

	// If project_dir is available, derive dbt project directory from it
	if projectDir, ok := config["project_dir"].(string); ok && projectDir != "" {
		return filepath.Join(projectDir, "transform", "dbt")
	}

	// Use plugin's ProjectDir as fallback
	return p.ProjectDir
}
