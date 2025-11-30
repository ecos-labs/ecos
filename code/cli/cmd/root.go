package cmd

import (
	"fmt"
	"os"

	"github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
	"github.com/ecos-labs/ecos-core/code/cli/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	dryRun  bool
	cfg     *config.EcosConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecos",
	Short: "CLI tool for the ecos FinOps data stack",
	Long: `The ecos CLI helps you work with ecos, an open source FinOps data stack that
transforms AWS Cost and Usage Reports (CUR) into clean, enriched, high-performance
datasets. Its analytics-ready semantic layer enables cost transparency, allocation,
and optimization with actionable insights.

The CLI provides one-command project setup, provisioning, and data model deployment.
It uses a plugin-based extensible architecture, supporting major cloud providers
and transformation tools like dbt.`,

	Version: version.GetInfo().Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set verbose mode for UI based on global flag
		utils.SetVerbose(verbose)

		// Skip config loading for commands that don't need existing config
		if cmd.Name() == "init" || cmd.Name() == "plugins" {
			return nil
		}

		return initConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .ecos.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output for debugging")

	// Local flags
	rootCmd.Flags().BoolP("version", "", false, "show version information")

	// Set custom help function to show banner
	originalHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		utils.PrintEcosBanner()
		originalHelpFunc(cmd, args)
	})
}

// initConfig reads in config file and ENV variables.
func initConfig() error {
	var err error

	// Load config file
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set global verbose flag based on config
	if cfg.Global.LogLevel == "debug" {
		verbose = true
		utils.SetVerbose(true)
	}

	if verbose {
		configPath := config.GetConfigFilePath()
		if configPath != "" {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", configPath)
		} else {
			fmt.Fprintf(os.Stderr, "No config file found, using defaults\n")
		}
	}

	return nil
}

// GetConfig returns the loaded configuration
func GetConfig() *config.EcosConfig {
	return cfg
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// IsDryRun returns whether dry-run mode is enabled
func IsDryRun() bool {
	return dryRun
}
