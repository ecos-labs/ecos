package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ecos-labs/ecos-core/code/cli/utils"
)

// PrereqResult represents the result of prerequisite checking
type PrereqResult struct {
	Found    []string
	Missing  []string
	Warnings []string
}

// PrereqCheck represents a single prerequisite check
type PrereqCheck struct {
	Name string
	Test func(ctx context.Context) (bool, string)
}

// PrereqConfig defines what prerequisites to check
type PrereqConfig struct {
	AWS        bool          // Include AWS checks (CLI, credentials)
	Python     bool          // Include Python check
	DBTAdapter string        // Check dbt Core + specific adapter (e.g., "athena", "bigquery")
	Commands   []string      // Check specific commands exist
	Custom     []PrereqCheck // Custom prerequisite checks
}

// RunPrerequisiteChecks is the unified entry point for prerequisite validation
func RunPrerequisiteChecks(ctx context.Context, config *PrereqConfig) error {
	utils.PrintDebug("Starting prerequisite checks")
	utils.PrintDebug(fmt.Sprintf("Config: AWS=%t, Python=%t, DBTAdapter=%s, Commands=%v, Custom=%d",
		config.AWS, config.Python, config.DBTAdapter, config.Commands, len(config.Custom)))

	sp := utils.NewSpinner("Checking prerequisites")
	sp.Start()

	checks := buildPrereqChecks(config)
	utils.PrintDebug(fmt.Sprintf("Built %d prerequisite checks", len(checks)))

	result := CheckPrerequisites(ctx, checks)

	sp.Stop()
	fmt.Println()
	result.LogPrerequisites()

	if len(result.Missing) > 0 {
		utils.PrintDebug(fmt.Sprintf("Prerequisites check failed with %d missing items", len(result.Missing)))
		return fmt.Errorf("missing prerequisites: %v", result.Missing)
	}

	utils.PrintDebug("All prerequisite checks passed successfully")
	return nil
}

// buildPrereqChecks constructs the prerequisite checks based on config
func buildPrereqChecks(config *PrereqConfig) []PrereqCheck {
	var checks []PrereqCheck

	if config.Python {
		checks = append(checks, PrereqCheck{"Python", CheckPython})
	}

	if config.AWS {
		checks = append(checks, checkAWS()...)
	}

	if config.DBTAdapter != "" {
		checks = append(checks, checkDBT(config.DBTAdapter)...)
	}

	// Add command checks
	for _, cmd := range config.Commands {
		cmdName := cmd
		checks = append(checks, PrereqCheck{
			Name: cmdName,
			Test: func(ctx context.Context) (bool, string) {
				return CheckCommand(ctx, cmdName, "--version")
			},
		})
	}

	// Add custom checks
	if len(config.Custom) > 0 {
		checks = append(checks, config.Custom...)
	}

	return checks
}

// checkAWS returns AWS-related prerequisite checks
func checkAWS() []PrereqCheck {
	return []PrereqCheck{
		{"AWS CLI", CheckAWSCLI},
		{"AWS credentials", CheckAWSCredentials},
	}
}

// checkDBT returns dbt-related prerequisite checks for a specific adapter
func checkDBT(adapterName string) []PrereqCheck {
	return []PrereqCheck{
		{"dbt with " + adapterName + " adapter", func(ctx context.Context) (bool, string) {
			return CheckDBTWithAdapter(ctx, adapterName)
		}},
	}
}

// CheckPrerequisites runs a set of prerequisite checks
func CheckPrerequisites(ctx context.Context, checks []PrereqCheck) *PrereqResult {
	result := &PrereqResult{}

	for i, check := range checks {
		utils.PrintDebug(fmt.Sprintf("Running prerequisite check %d/%d: %s", i+1, len(checks), check.Name))

		if success, detail := check.Test(ctx); success {
			utils.PrintDebug(fmt.Sprintf("✓ %s check passed", check.Name))
			if detail != "" {
				result.Found = append(result.Found, detail)
				utils.PrintDebug(fmt.Sprintf("  Detail: %s", detail))
			} else {
				result.Found = append(result.Found, check.Name)
			}
		} else {
			utils.PrintDebug(fmt.Sprintf("✗ %s check failed", check.Name))
			result.Missing = append(result.Missing, check.Name)
		}
	}

	utils.PrintDebug(fmt.Sprintf("Prerequisite check summary: %d found, %d missing", len(result.Found), len(result.Missing)))
	return result
}

// CheckPython checks if Python is installed (python3 preferred, then python)
func CheckPython(ctx context.Context) (bool, string) {
	// Safe: using hardcoded list of known commands
	for _, cmd := range []string{"python3", "python"} {
		if exec.CommandContext(ctx, cmd, "--version").Run() == nil { //nolint:gosec
			return true, "Python (" + cmd + ")"
		}
	}
	return false, ""
}

// CheckCommand checks if a command is available
func CheckCommand(ctx context.Context, command string, args ...string) (bool, string) {
	err := exec.CommandContext(ctx, command, args...).Run()
	return err == nil, ""
}

// RunWithTimeout executes a command with a timeout
func RunWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return exec.CommandContext(timeoutCtx, name, args...).Run()
}

// LogPrerequisites displays prerequisite check results
func (r *PrereqResult) LogPrerequisites() {
	utils.PrintDebug(fmt.Sprintf("Logging prerequisite results: %d found, %d missing, %d warnings",
		len(r.Found), len(r.Missing), len(r.Warnings)))

	if len(r.Missing) == 0 && len(r.Warnings) == 0 {
		totalExpected := len(r.Found) + len(r.Missing)
		utils.PrintDebug("All prerequisites satisfied, showing success message")
		utils.PrintSuccess(fmt.Sprintf("All required prerequisites verified (%d/%d found)", len(r.Found), totalExpected))
		return
	}

	utils.PrintDebug("Some prerequisites missing or warnings present, showing detailed results")
	r.logDetailed()
}

func (r *PrereqResult) logDetailed() {
	var output strings.Builder

	if len(r.Found) > 0 {
		utils.PrintDebug(fmt.Sprintf("Displaying %d found prerequisites", len(r.Found)))
		utils.PrintSuccess("Found:")
		for _, f := range r.Found {
			output.WriteString("  • ")
			output.WriteString(f)
			output.WriteString("\n")
		}
		fmt.Print(output.String())
		output.Reset()
	}

	if len(r.Missing) > 0 {
		utils.PrintDebug(fmt.Sprintf("Displaying %d missing prerequisites", len(r.Missing)))
		fmt.Println()
		utils.PrintError("Missing:")
		for _, m := range r.Missing {
			output.WriteString("  • ")
			output.WriteString(m)
			output.WriteString("\n")
		}
		fmt.Print(output.String())
		output.Reset()

		utils.PrintDebug("Showing installation instructions for missing prerequisites")
		r.logInstallInstructions()
	}

	if len(r.Warnings) > 0 {
		utils.PrintDebug(fmt.Sprintf("Displaying %d warnings", len(r.Warnings)))
		fmt.Println()
		utils.PrintWarning("Warnings:")
		for _, w := range r.Warnings {
			output.WriteString("  • ")
			output.WriteString(w)
			output.WriteString("\n")
		}
		fmt.Print(output.String())
	}
}

func (r *PrereqResult) logInstallInstructions() {
	fmt.Println()
	utils.PrintInfo("Install:")

	instructionsShown := 0
	for _, m := range r.Missing {
		var instruction string

		switch {
		case m == "AWS CLI":
			instruction = "  • AWS CLI: https://aws.amazon.com/cli/"
		case m == "Python":
			instruction = "  • Python: https://www.python.org/downloads/"
		case m == "AWS credentials":
			instruction = "  • AWS credentials: aws configure"
		case strings.Contains(m, "dbt with") && strings.Contains(m, "adapter"):
			// Extract adapter name from "dbt with <adapter> adapter"
			parts := strings.Split(m, " ")
			if len(parts) >= 3 {
				adapterName := parts[2] // "athena", "bigquery", etc.
				instruction = fmt.Sprintf("  • dbt-%s: pip install dbt-%s", adapterName, adapterName)
			}
		case m == "dbt Core":
			instruction = "  • dbt Core: pip install dbt-core"
		}

		if instruction != "" {
			utils.PrintDebug(fmt.Sprintf("Showing install instruction for: %s", m))
			fmt.Println(instruction)
			instructionsShown++
		} else {
			utils.PrintDebug(fmt.Sprintf("No install instruction available for: %s", m))
		}
	}

	utils.PrintDebug(fmt.Sprintf("Showed %d install instructions out of %d missing prerequisites",
		instructionsShown, len(r.Missing)))
}

// CheckDBTWithAdapter checks if dbt Core and a specific adapter are installed
func CheckDBTWithAdapter(ctx context.Context, adapterName string) (bool, string) {
	cmd := exec.CommandContext(ctx, "dbt", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	outputStr := string(output)
	var coreVersion string
	var adapterVersion string
	foundCore := false
	foundAdapter := false

	// Parse the output
	lines := strings.Split(outputStr, "\n")
	inPluginsSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for dbt Core version
		if strings.Contains(strings.ToLower(line), "core") && !foundCore {
			foundCore = true
			coreVersion = line
		}

		// Check if we're entering the Plugins section
		if strings.HasPrefix(strings.ToLower(line), "plugins:") {
			inPluginsSection = true
			continue
		}

		// If we're in plugins section, look for the adapter name directly
		if inPluginsSection && !foundAdapter {
			lineLower := strings.ToLower(line)
			// Simply check if the line contains the adapter name
			if strings.Contains(lineLower, strings.ToLower(adapterName)) {
				foundAdapter = true
				adapterVersion = line
			}
		}

		// Stop parsing if we hit another section or empty line after plugins
		if inPluginsSection && line == "" {
			break
		}
	}

	// Check results
	if !foundCore || !foundAdapter {
		return false, ""
	}

	// Both found - return combined result
	result := fmt.Sprintf("dbt Core (%s) + %s adapter (%s)", coreVersion, adapterName, adapterVersion)
	return true, result
}

// CheckAWSCLI checks if AWS CLI is installed and accessible
func CheckAWSCLI(ctx context.Context) (bool, string) {
	err := exec.CommandContext(ctx, "aws", "--version").Run()
	return err == nil, ""
}

// CheckAWSCredentials validates AWS credentials by making a simple STS call
func CheckAWSCredentials(ctx context.Context) (bool, string) {
	err := ValidateAWSCredentials(ctx, 0)
	return err == nil, ""
}
