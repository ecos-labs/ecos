package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDBTProfilesFromTemplate(t *testing.T) {
	data := DBTProfilesTemplate{
		Profile:       "ecos-athena",
		Target:        "prod",
		AWSProfile:    "default",
		AWSRegion:     "us-east-1",
		ResultsBucket: "my-bucket",
		Database:      "my_db",
		Workgroup:     "my-wg",
	}

	out, err := generateDBTProfilesFromTemplate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ecos-athena") {
		t.Errorf("expected profile name in output")
	}
	if !strings.Contains(out, "my-bucket") {
		t.Errorf("expected bucket name in output")
	}
}

func TestGenerateDBTProjectFromTemplate(t *testing.T) {
	data := DBTProjectTemplate{
		Profile:               "ecos-athena",
		DatasourceVars:        []DatasourceVar{{Key: "cur_table", Value: "cur"}},
		MaterializationMode:   "view",
		BronzeMaterialization: "view",
		SilverMaterialization: "table",
		GoldMaterialization:   "table",
		EnablePartitioning:    true,
	}

	out, err := generateDBTProjectFromTemplate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ecos-athena") {
		t.Errorf("expected profile field in project output")
	}
	if !strings.Contains(out, "cur_table") {
		t.Errorf("expected datasource var in output")
	}
}

func TestGenerateDBTProfiles_FileOutput(t *testing.T) {
	tmp := t.TempDir()

	data := DBTProfilesTemplate{
		Profile:       "ecos-athena",
		Target:        "prod",
		AWSProfile:    "default",
		AWSRegion:     "eu-west-1",
		ResultsBucket: "test-bucket",
		Database:      "db",
		Workgroup:     "wg",
	}

	err := GenerateDBTProfiles(data, tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outFile := filepath.Join(tmp, "profiles.yml")
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("profiles.yml not written: %v", err)
	}
}

func TestGenerateDBTProject_FileOutput(t *testing.T) {
	tmp := t.TempDir()

	data := DBTProjectTemplate{
		Profile:               "ecos-athena",
		DatasourceVars:        []DatasourceVar{{Key: "cur_schema", Value: "cur"}},
		MaterializationMode:   "view",
		BronzeMaterialization: "view",
		SilverMaterialization: "view",
		GoldMaterialization:   "table",
	}

	err := GenerateDBTProject(data, tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outFile := filepath.Join(tmp, "dbt_project.yml")
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("dbt_project.yml not written: %v", err)
	}
}

func TestExtractDBTDataFromEcosConfig(t *testing.T) {
	cfg := &EcosConfig{
		Transform: TransformConfig{
			DBT: DBTConfig{
				Profile: "ecos-athena",
				Target:  "prod",
				Vars: map[string]string{
					"cur_table": "cur_data",
					"cur_db":    "cur",
				},
				Materialization: &MaterializationConfig{
					Mode: "table",
					LayerOverrides: map[string]string{
						"bronze": "view",
						"gold":   "table",
					},
				},
			},
		},
		AWS: AWSRootConfig{
			Region:        "us-east-1",
			ResultsBucket: "bucket-x",
			DBTWorkgroup:  "wg-x",
			Database:      "db-x",
		},
	}

	project, profiles, err := ExtractDBTDataFromEcosConfig(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.MaterializationMode != "table" {
		t.Errorf("expected table mode")
	}
	if profiles.Workgroup != "wg-x" {
		t.Errorf("workgroup mismatch")
	}
}

func TestCompareFileWithExpected_NoDiff(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.yml")
	expected := "abc: 123"

	if err := os.WriteFile(path, []byte(expected), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	report, err := compareFileWithExpected(path, expected, "file.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.HasChanges {
		t.Errorf("expected no changes")
	}
}

func TestCompareFileWithExpected_WithDiff(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.yml")

	actual := "abc: 123"
	expected := "abc: 456" // differs

	if err := os.WriteFile(path, []byte(actual), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	report, err := compareFileWithExpected(path, expected, "file.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.HasChanges {
		t.Errorf("expected changes")
	}

	if !strings.Contains(report.Diff, "- abc: 456") {
		t.Errorf("missing expected diff output")
	}
}

func TestGenerateUnifiedDiff(t *testing.T) {
	diff := generateUnifiedDiff("line1\nline2", "line1\nDIFF", "expected", "actual")

	if !strings.Contains(diff, "- line2") {
		t.Errorf("expected removal")
	}
	if !strings.Contains(diff, "+ DIFF") {
		t.Errorf("expected addition")
	}
}

func TestDetectDriftFromEcosConfig_NoDrift(t *testing.T) {
	tmp := t.TempDir()

	// Create minimal .ecos.yaml
	ecos := `
project_name: test
data_source: aws_cur
aws:
  region: us-east-1
  results_bucket: test-bucket
  database: testdb
  dbt_workgroup: wg
transform:
  dbt:
    vars: {}
`
	if err := os.WriteFile(filepath.Join(tmp, ".ecos.yaml"), []byte(ecos), 0o600); err != nil {
		t.Fatalf("failed to write .ecos.yaml: %v", err)
	}

	// Render files into expected directory
	if err := os.MkdirAll(filepath.Join(tmp, "transform/dbt"), 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Generate expected files
	projectData := DBTProjectTemplate{
		Profile:               "ecos-athena",
		DatasourceVars:        []DatasourceVar{},
		MaterializationMode:   "view",
		BronzeMaterialization: "view",
		SilverMaterialization: "view",
		GoldMaterialization:   "view",
		EnablePartitioning:    true,
	}
	profilesData := DBTProfilesTemplate{
		Profile:       "ecos-athena",
		Target:        "prod",
		AWSProfile:    "default",
		AWSRegion:     "us-east-1",
		ResultsBucket: "test-bucket",
		Database:      "testdb",
		Workgroup:     "wg",
	}

	p1, _ := generateDBTProjectFromTemplate(projectData)
	p2, _ := generateDBTProfilesFromTemplate(profilesData)

	if err := os.WriteFile(filepath.Join(tmp, "transform/dbt/dbt_project.yml"), []byte(p1), 0o600); err != nil {
		t.Fatalf("failed to write dbt_project.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "transform/dbt/profiles.yml"), []byte(p2), 0o600); err != nil {
		t.Fatalf("failed to write profiles.yml: %v", err)
	}

	report, err := DetectDriftFromEcosConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.HasChanges {
		t.Errorf("expected no drift but detected drift")
	}
}

func TestDetectDriftFromEcosConfig_WithDrift(t *testing.T) {
	tmp := t.TempDir()

	// Create minimal .ecos.yaml
	ecos := `
project_name: test
data_source: aws_cur
aws:
  region: us-east-1
  results_bucket: test-bucket
  database: testdb
  dbt_workgroup: wg
transform:
  dbt:
    vars: {}
`
	if err := os.WriteFile(filepath.Join(tmp, ".ecos.yaml"), []byte(ecos), 0o600); err != nil {
		t.Fatalf("failed to write .ecos.yaml: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(tmp, "transform/dbt"), 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// dbt_project.yml is wrong on purpose
	if err := os.WriteFile(
		filepath.Join(tmp, "transform/dbt/dbt_project.yml"),
		[]byte("WRONG CONTENT"),
		0o600,
	); err != nil {
		t.Fatalf("failed to write dbt_project.yml: %v", err)
	}

	// profiles.yml missing entirely → ignored

	report, err := DetectDriftFromEcosConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	if !report.HasChanges {
		t.Fatalf("expected drift but none detected")
	}

	if _, ok := report.Files["dbt_project.yml"]; !ok {
		t.Fatalf("expected dbt_project.yml report")
	}
}
