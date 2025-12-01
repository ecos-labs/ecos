package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ecos-labs/ecos/code/cli/config"
	"github.com/ecos-labs/ecos/code/cli/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// RepoConfig stores repository configuration for version tracking
type RepoConfig struct {
	ModelVersion string `yaml:"model_version"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display version information for the ecos CLI and models.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion(IsVerbose())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion(verbose bool) {
	info := version.GetInfo()

	if verbose {
		fmt.Println(info.VerboseString())
	} else {
		fmt.Println(info.String())
	}

	// Check if we're in an initialized ecos repository
	// and display model version if available
	modelVersion := getModelVersion()
	if modelVersion != "" {
		fmt.Printf("Models version: %s\n", modelVersion)
	}
}

// getModelVersion attempts to find the model version in the current repository
func getModelVersion() string {
	// Try to find .ecos.yaml in current directory
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check for .ecos.yaml
	configPath := filepath.Clean(filepath.Join(pwd, config.ConfigFilename))
	if _, err := os.Stat(configPath); err != nil {
		return ""
	}

	// Read the config file (safe: path is constructed from known components)
	data, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		return ""
	}

	// Parse the config file
	var repoConfig RepoConfig
	if err := yaml.Unmarshal(data, &repoConfig); err != nil {
		return ""
	}

	return repoConfig.ModelVersion
}
