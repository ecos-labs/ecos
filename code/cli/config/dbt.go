package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// GenerateDBTProfiles generates and writes a dbt profiles.yml file to the specified directory
func GenerateDBTProfiles(data DBTProfilesTemplate, targetDir string) error {
	content, err := generateDBTProfilesFromTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to generate profiles: %w", err)
	}

	return writeDBTFile(content, targetDir, "profiles.yml")
}

// GenerateDBTProject generates and writes a dbt_project.yml file to the specified directory
func GenerateDBTProject(data DBTProjectTemplate, targetDir string) error {
	content, err := generateDBTProjectFromTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	return writeDBTFile(content, targetDir, "dbt_project.yml")
}

// generateDBTProfilesFromTemplate generates a dbt profiles.yml file using template data
func generateDBTProfilesFromTemplate(data DBTProfilesTemplate) (string, error) {
	tmpl, err := template.New("dbt_profiles.yml.tmpl").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(templateFS, "templates/dbt_profiles.yml.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse profiles template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute profiles template: %w", err)
	}

	return buf.String(), nil
}

// generateDBTProjectFromTemplate generates a dbt_project.yml file using template data
func generateDBTProjectFromTemplate(data DBTProjectTemplate) (string, error) {
	tmpl, err := template.New("dbt_project.yml.tmpl").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(templateFS, "templates/dbt_project.yml.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse project template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute project template: %w", err)
	}

	return buf.String(), nil
}

// writeDBTFile is a wrapper around WriteConfigFile for backward compatibility
func writeDBTFile(content, targetDir, filename string) error {
	return WriteConfigFile(content, targetDir, filename)
}

// FileDiffReport represents the diff status of a single generated file
type FileDiffReport struct {
	FilePath   string
	HasChanges bool
	Diff       string
}

// ValidationReport represents the validation status of all generated files
type ValidationReport struct {
	HasChanges bool
	Files      map[string]*FileDiffReport
}

// DetectDriftFromEcosConfig reads .ecos.yaml and checks if dbt files match what should be generated
// This is the main drift detection function that uses .ecos.yaml as the source of truth
func DetectDriftFromEcosConfig(outputPath string) (*ValidationReport, error) {
	// Load .ecos.yaml configuration
	ecosConfigPath := filepath.Join(outputPath, ConfigFilename)
	if _, err := os.Stat(ecosConfigPath); err != nil {
		return nil, fmt.Errorf(".ecos.yaml not found in %s: %w", outputPath, err)
	}

	ecosConfig, err := LoadConfig(ecosConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load .ecos.yaml: %w", err)
	}

	// Extract data needed to generate dbt files from .ecos.yaml
	dbtProjectData, dbtProfilesData, err := ExtractDBTDataFromEcosConfig(ecosConfig, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract dbt data from .ecos.yaml: %w", err)
	}

	// Generate expected content in memory
	expectedProjectContent, err := generateDBTProjectFromTemplate(dbtProjectData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected dbt_project.yml: %w", err)
	}

	expectedProfilesContent, err := generateDBTProfilesFromTemplate(dbtProfilesData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected profiles.yml: %w", err)
	}

	report := &ValidationReport{
		HasChanges: false,
		Files:      make(map[string]*FileDiffReport),
	}

	// Check dbt_project.yml
	dbtProjectPath := filepath.Join(outputPath, "transform", "dbt", "dbt_project.yml")
	if _, err := os.Stat(dbtProjectPath); err == nil {
		projectReport, err := compareFileWithExpected(dbtProjectPath, expectedProjectContent, "dbt_project.yml")
		if err != nil {
			return nil, fmt.Errorf("failed to check dbt_project.yml: %w", err)
		}
		if projectReport.HasChanges {
			report.HasChanges = true
		}
		report.Files["dbt_project.yml"] = projectReport
	}

	// Check profiles.yml
	profilesPath := filepath.Join(outputPath, "transform", "dbt", "profiles.yml")
	if _, err := os.Stat(profilesPath); err == nil {
		profilesReport, err := compareFileWithExpected(profilesPath, expectedProfilesContent, "profiles.yml")
		if err != nil {
			return nil, fmt.Errorf("failed to check profiles.yml: %w", err)
		}
		if profilesReport.HasChanges {
			report.HasChanges = true
		}
		report.Files["profiles.yml"] = profilesReport
	}

	return report, nil
}

// ExtractDBTDataFromEcosConfig extracts the data needed to generate dbt files from EcosConfig
// This is a public function that can be used by commands to regenerate dbt files from .ecos.yaml
func ExtractDBTDataFromEcosConfig(ecosConfig *EcosConfig, outputPath string) (DBTProjectTemplate, DBTProfilesTemplate, error) {
	// Extract datasource vars from transform.dbt.vars
	var datasourceVars []DatasourceVar
	for key, value := range ecosConfig.Transform.DBT.Vars {
		datasourceVars = append(datasourceVars, DatasourceVar{
			Key:   key,
			Value: value,
		})
	}

	// Sort datasource vars by key for consistent ordering
	// This prevents false drift detection due to different ordering
	sort.Slice(datasourceVars, func(i, j int) bool {
		return datasourceVars[i].Key < datasourceVars[j].Key
	})

	// Extract materialization settings
	matMode := "view"
	bronzeMat := "view"
	silverMat := "view"
	goldMat := "view"

	if ecosConfig.Transform.DBT.Materialization != nil {
		if ecosConfig.Transform.DBT.Materialization.Mode != "" {
			matMode = ecosConfig.Transform.DBT.Materialization.Mode
		}
		if ecosConfig.Transform.DBT.Materialization.LayerOverrides != nil {
			if val, ok := ecosConfig.Transform.DBT.Materialization.LayerOverrides["bronze"]; ok {
				bronzeMat = val
			}
			if val, ok := ecosConfig.Transform.DBT.Materialization.LayerOverrides["silver"]; ok {
				silverMat = val
			}
			if val, ok := ecosConfig.Transform.DBT.Materialization.LayerOverrides["gold"]; ok {
				goldMat = val
			}
		}
	}

	// Build DBTProjectTemplate
	dbtProjectData := DBTProjectTemplate{
		Profile:               ecosConfig.Transform.DBT.Profile,
		DatasourceVars:        datasourceVars,
		IcebergEnabled:        false, // Default value
		BillingPeriodStart:    nil,
		BillingPeriodEnd:      nil,
		MaterializationMode:   matMode,
		BronzeMaterialization: bronzeMat,
		SilverMaterialization: silverMat,
		GoldMaterialization:   goldMat,
		UseIceberg:            false,
		EnablePartitioning:    true,
	}

	// Build DBTProfilesTemplate
	// Extract target name from transform.dbt.target or use default
	target := ecosConfig.Transform.DBT.Target
	if target == "" {
		target = "prod"
	}

	// Extract AWS profile name from transform.dbt.aws_profile or use default
	awsProfile := ecosConfig.Transform.DBT.AWSProfile
	if awsProfile == "" {
		awsProfile = "default"
	}

	dbtProfilesData := DBTProfilesTemplate{
		Profile:       ecosConfig.Transform.DBT.Profile,
		Target:        target,
		AWSProfile:    awsProfile,
		AWSRegion:     ecosConfig.AWS.Region,
		ResultsBucket: ecosConfig.AWS.ResultsBucket,
		Database:      ecosConfig.AWS.Database,
		Workgroup:     ecosConfig.AWS.DBTWorkgroup,
	}

	return dbtProjectData, dbtProfilesData, nil
}

// compareFileWithExpected compares an existing file with expected content and returns a diff report
func compareFileWithExpected(existingPath, expectedContent, filename string) (*FileDiffReport, error) {
	report := &FileDiffReport{
		FilePath:   existingPath,
		HasChanges: false,
	}

	// Read existing file (safe: path is validated by caller)
	actualContent, err := os.ReadFile(filepath.Clean(existingPath)) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read existing file: %w", err)
	}

	// Compare
	if string(actualContent) == expectedContent {
		return report, nil
	}

	// Files differ - generate diff
	report.HasChanges = true
	diff := generateUnifiedDiff(
		expectedContent,
		string(actualContent),
		fmt.Sprintf("%s (expected)", filename),
		fmt.Sprintf("%s (current)", filename),
	)
	report.Diff = diff

	return report, nil
}

// generateUnifiedDiff generates a unified diff between expected and actual content
func generateUnifiedDiff(expected, actual, fromFile, toFile string) string {
	// Simple line-by-line diff
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s\n", fromFile))
	diff.WriteString(fmt.Sprintf("+++ %s\n", toFile))

	maxLen := len(expectedLines)
	if len(actualLines) > maxLen {
		maxLen = len(actualLines)
	}

	for i := 0; i < maxLen; i++ {
		var expectedLine, actualLine string
		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			if expectedLine != "" {
				diff.WriteString(fmt.Sprintf("- %s\n", expectedLine))
			}
			if actualLine != "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", actualLine))
			}
		}
	}

	return diff.String()
}
