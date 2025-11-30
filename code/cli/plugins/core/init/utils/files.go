package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ecos-labs/ecos-core/code/cli/utils"
)

// SetupDirectories creates the standard ecos directory structure
func SetupDirectories(outputPath string) error {
	dirs := []string{"plugins/ingest", "plugins/transform", "transform/dbt", "logs", "output"}
	for _, dir := range dirs {
		fp := filepath.Join(outputPath, dir)
		if !utils.DirectoryExists(fp) {
			if err := os.MkdirAll(fp, 0o750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}
	return nil
}

// SetupBaseFiles creates standard .gitignore and README.md files
func SetupBaseFiles(outputPath, dataSourceName string) error {
	gitignore := `# ecos generated files
logs/
output/
.ecos/temp/
# OS & IDE
.DS_Store
.vscode/
.idea/
*.swp
*.tmp
`

	readme := fmt.Sprintf(`# ecos Project for %s
This is an ecos project for managing cloud cost and billing data.

## Getting Started
1. Review and customize .ecos.yaml configuration
2. Verify setup: Run 'ecos transform debug' to test connection and resources
3. Seed data: Run 'ecos transform seed' to load seed data
4. Transform data: Run 'ecos transform run' to execute dbt models

## Project Structure
- plugins/ - Custom plugins for ingest and transform operations
- transform/dbt/ - dbt models and transformation scripts
- logs/ - Log files and operation metadata
- output/ - Generated reports and output files

The .ecos.yaml file contains all configuration for your ecos project.
`, dataSourceName)

	gitignorePath := filepath.Join(outputPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.WriteFile(gitignorePath, []byte(gitignore)); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	readmePath := filepath.Join(outputPath, "README.md")
	if !utils.FileExists(readmePath) {
		if err := utils.WriteFile(readmePath, []byte(readme)); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
	}

	return nil
}

// PrintPostInitSummary displays the post-initialization summary with next steps
func PrintPostInitSummary() {
	fmt.Println() // Add spacing after progress bar
	utils.PrintSuccess("Project initialized successfully")
	fmt.Println() // Add extra spacing before next steps

	fmt.Println("Next steps:")
	fmt.Printf("  %s1.%s Review and customize .ecos.yaml configuration\n", utils.ColorCyan, utils.ColorReset)
	fmt.Printf("  %s2.%s Verify setup: Run 'ecos transform debug' to test connection and resources\n", utils.ColorCyan, utils.ColorReset)
	fmt.Printf("  %s3.%s Seed data: Run 'ecos transform seed' to load seed data\n", utils.ColorCyan, utils.ColorReset)
	fmt.Printf("  %s4.%s Transform data: Run 'ecos transform run' to execute dbt models\n", utils.ColorCyan, utils.ColorReset)
}
