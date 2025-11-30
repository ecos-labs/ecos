package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	// Import destroy plugins to trigger plugin self-registration
	_ "github.com/ecos-labs/ecos-core/code/cli/plugins/core/destroy"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"

	"github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos-core/code/cli/utils"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy ecos-managed cloud resources using .ecos.yaml",
	RunE:  runDestroy,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().StringP("source", "s", "", "Force provider (aws_cur)")
}

var (
	registryLoadDestroy = registry.LoadDestroyPlugin
	utilsConfirmPrompt  = utils.ConfirmPrompt
)

// createConfigBackup creates a backup of the config file before destruction.
// Returns the backup path and an error if backup creation fails.
func createConfigBackup(configPath string) (string, error) {
	backupPath := configPath + ".backup"
	if err := os.Rename(configPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup config before destruction: %w", err)
	}
	return backupPath, nil
}

// restoreConfigBackup restores the config file from backup.
// Returns an error if restore fails, nil if backup doesn't exist (already restored).
func restoreConfigBackup(backupPath, configPath string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return nil // Backup already restored or doesn't exist
	}
	if err := os.Rename(backupPath, configPath); err != nil {
		return fmt.Errorf("failed to restore config backup: %w", err)
	}
	return nil
}

func runDestroy(cmd *cobra.Command, args []string) error {
	// Load `.ecos.yaml`
	configPath, err := config.FindConfigFile()
	if err != nil {
		// Do not show Cobra usage or default error output
		cmd.SilenceUsage = true

		msg := `🚫 ecos configuration file not found

This directory does not appear to be an ecos project.
The destroy command requires a .ecos.yaml file that tracks
the cloud resources created or managed by ecos.

Run "ecos init" to create a new project, or if your config file
is located elsewhere, run:

    ecos --config /path/to/.ecos.yaml destroy

For more information, run:

    ecos help
`

		cmd.PrintErrln(msg)
		return nil
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load ecos config: %w", err)
	}

	utils.PrintHeader("💣 ecos destroy")
	fmt.Println()

	sourceFlag, _ := cmd.Flags().GetString("source")
	provider := detectProvider(cfg, sourceFlag)

	destroyPlugin, err := registryLoadDestroy(provider)
	if err != nil {
		return fmt.Errorf("failed to load destroy plugin '%s': %w", provider, err)
	}

	if loader, ok := destroyPlugin.(types.DestroyConfigLoader); ok {
		if err := loader.LoadFromConfig(cfg); err != nil {
			cmd.SilenceUsage = true
			cmd.PrintErrln(err)
			return nil
		}
	}

	utils.PrintInfo(fmt.Sprintf("Using ecos configuration file: %s", configPath))

	if err := destroyPlugin.ValidatePrerequisites(); err != nil {
		return fmt.Errorf("prerequisite validation failed: %w", err)
	}

	previewer, ok := destroyPlugin.(types.DestroyPreviewer)
	if !ok {
		return errors.New("plugin does not support resource preview")
	}

	previews := previewer.DescribeDestruction()

	utils.PrintSubHeader("📦 Resource Destruction Preview")
	fmt.Println()

	headers := []string{"Type", "Name", "Managed"}
	var rows [][]string
	hasUnmanaged := false
	hasPreviewErrors := false

	for _, p := range previews {
		managedText := "yes"

		switch {
		case strings.TrimSpace(p.Error) != "":
			hasPreviewErrors = true
			managedText = fmt.Sprintf("unknown ⚠️ (%s)", p.Error)
		case !p.Managed:
			hasUnmanaged = true
			managedText = "no ⚠️"
		}

		rows = append(rows, []string{p.Kind, p.Name, managedText})
	}

	utils.PrintTable(headers, rows)
	fmt.Println()

	if hasPreviewErrors {
		utils.PrintWarning("⚠️ WARNING: Some resources could not be verified due to errors.")
		utils.PrintWarning("Review the error messages shown in the preview before continuing.\n")
	}

	if hasUnmanaged {
		utils.PrintWarning("⚠️ WARNING: Some resources appear to be unmanaged by ecos.")
		utils.PrintWarning("Destroying them may impact production workloads.\n")
	}

	var confirm bool
	if hasUnmanaged || hasPreviewErrors {
		confirm = utilsConfirmPrompt("Do you STILL want to destroy ALL listed resources")
	} else {
		confirm = utilsConfirmPrompt("Do you want to proceed with destroying these resources")
	}

	if !confirm {
		utils.PrintWarning("Destruction cancelled by user.")
		return nil
	}

	utils.PrintSubHeader("🔎 Starting resource destruction")

	destroyer, ok := destroyPlugin.(types.DestroyExecutor)
	if !ok {
		return errors.New("plugin does not support resource destruction")
	}

	// Create backup right before destruction starts
	backupPath, err := createConfigBackup(configPath)
	if err != nil {
		return err
	}

	// Restore backup on any failure or cancellation (default behavior)
	// Only skip restore if destruction succeeds completely
	shouldRestore := true
	defer func() {
		if shouldRestore {
			if restoreErr := restoreConfigBackup(backupPath, configPath); restoreErr != nil {
				cmd.PrintErrf("Warning: %v\n", restoreErr)
			}
		}
	}()

	results, err := destroyer.DestroyResources()
	if err != nil {
		return fmt.Errorf("resource destruction failed: %w", err)
	}

	// Check if destruction was cancelled (empty results with no error indicates cancellation)
	if len(results) == 0 {
		utils.PrintWarning("Destruction cancelled by user. Config restored.")
		return nil
	}

	// Destruction succeeded - keep backup file for reference/recovery
	shouldRestore = false

	utils.PrintSubHeader("🧹 ecos Resources Destruction Summary")

	for _, r := range results {
		color := utils.ColorGreen
		switch r.Status {
		case types.DestroyStatusFailed:
			color = utils.ColorRed
		case types.DestroyStatusSkipped:
			color = utils.ColorYellow
		}

		suffix := string(r.Status)
		if strings.TrimSpace(r.Error) != "" {
			suffix = fmt.Sprintf("%s: %s", r.Status, r.Error)
		}

		fmt.Printf(
			"  • %s%s%s %s (%s)\n",
			color,
			r.Kind,
			utils.ColorReset,
			r.Name,
			suffix,
		)
	}

	return nil
}

func detectProvider(cfg *config.EcosConfig, flag string) string {
	if flag != "" {
		return flag
	}
	if cfg != nil && cfg.DataSource != "" {
		return cfg.DataSource
	}
	return ""
}
