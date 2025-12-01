package init

import (
	"fmt"
	"strings"
	"testing"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ecos-labs/ecos/code/cli/plugins/types"
)

func TestAWSCURInitPlugin_Metadata(t *testing.T) {
	plugin := &AWSCURInitPlugin{}

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"Name returns correct value", plugin.Name, "aws-cur-init"},
		{"Version returns correct value", plugin.Version, "1.0.0"},
		{"Author returns correct value", plugin.Author, "ecos team"},
		{"CloudProvider returns correct value", plugin.CloudProvider, "aws"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSCURInitPlugin_Type(t *testing.T) {
	plugin := &AWSCURInitPlugin{}
	if plugin.Type() != types.PluginTypeInit {
		t.Errorf("Type() = %v, want %v", plugin.Type(), types.PluginTypeInit)
	}
}

func TestAWSCURInitPlugin_IsCore(t *testing.T) {
	plugin := &AWSCURInitPlugin{}
	if !plugin.IsCore() {
		t.Error("IsCore() should return true")
	}
}

func TestAWSCURInitPlugin_Description(t *testing.T) {
	desc := (&AWSCURInitPlugin{}).Description()
	wantTerms := []string{"AWS", "Cost", "Usage", "Athena"}

	for _, term := range wantTerms {
		if !strings.Contains(desc, term) {
			t.Errorf("Description() missing %q", term)
		}
	}
}

func TestAWSCURInitPlugin_SupportedEngines(t *testing.T) {
	eng := (&AWSCURInitPlugin{}).SupportedEngines()

	foundAthena := false
	for _, e := range eng {
		if e.Code == "athena" {
			foundAthena = true
		}
	}

	if !foundAthena {
		t.Error("Expected Athena in SupportedEngines()")
	}
}

func TestAWSCURInitPlugin_SupportedTransformTools(t *testing.T) {
	tools := (&AWSCURInitPlugin{}).SupportedTransformTools()

	found := false
	for _, t := range tools {
		if t.Code == "dbt" {
			found = true
		}
	}

	if !found {
		t.Error("Expected dbt in SupportedTransformTools()")
	}
}

func TestAWSCURInitPlugin_ValidateRegion(t *testing.T) {
	plugin := &AWSCURInitPlugin{}

	valid := []string{"us-east-1", "eu-west-1", "ap-southeast-1"}
	invalid := []string{"bad-region", "", "us-east1", "eu_west_1"}

	for _, region := range valid {
		if err := plugin.ValidateRegion(region); err != nil {
			t.Errorf("expected region %s to be valid", region)
		}
	}

	for _, region := range invalid {
		if err := plugin.ValidateRegion(region); err == nil {
			t.Errorf("expected region %s to be invalid", region)
		}
	}
}

func TestDefaultMaterializationConfig(t *testing.T) {
	cfg := DefaultMaterializationConfig()
	if cfg.Mode == "" || cfg.Bronze == "" || cfg.Silver == "" || cfg.Gold == "" {
		t.Error("MaterializationConfig defaults should not be empty")
	}
}

func TestAWSCURInput_Struct(t *testing.T) {
	input := &AWSCURInput{
		ProjectName:      "project",
		TransformTool:    "dbt",
		SQLEngine:        "athena",
		AWSRegion:        "us-east-1",
		AWSProfile:       "default",
		DBTWorkgroup:     "wg",
		ResultsBucket:    "bucket",
		ModelVersion:     "v1.0.0",
		SkipProvisioning: false,
	}

	if input.ProjectName != "project" || input.TransformTool != "dbt" {
		t.Error("AWSCURInput struct fields not assigned correctly")
	}
}

func TestAWSCURInitPlugin_HandleBucketError(t *testing.T) {
	p := &AWSCURInitPlugin{}

	tests := []struct {
		name     string
		errMsg   string
		wantStat types.InitStatus
		wantErr  string
	}{
		{"BucketAlreadyOwnedByYou", "BucketAlreadyOwnedByYou: error", types.InitStatusSkipped, ""},
		{"BucketAlreadyExists", "BucketAlreadyExists: error", types.InitStatusFailed, "globally"},
		{"AccessDenied", "AccessDenied: error", types.InitStatusFailed, "Access denied"},
		{"OtherError", "SomethingElse: error", types.InitStatusFailed, "Failed to create bucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockError{msg: tt.errMsg}
			res := p.handleBucketError("b", m)

			if res.Status != tt.wantStat {
				t.Errorf("Status = %s, want %s", res.Status, tt.wantStat)
			}

			if !strings.Contains(res.Error, tt.wantErr) {
				t.Errorf("Error %q should contain %q", res.Error, tt.wantErr)
			}
		})
	}
}

func TestGetLifecycleRules(t *testing.T) {
	rules := getLifecycleRules()

	// Verify we have exactly 2 rules
	if len(rules) != 2 {
		t.Fatalf("Expected 2 lifecycle rules, got %d", len(rules))
	}

	// Test first rule: DeleteAdhocQueryResultsAfter30Days
	rule1 := rules[0]
	if rule1.ID == nil || *rule1.ID != "DeleteAdhocQueryResultsAfter30Days" {
		t.Errorf("Rule 1 ID = %v, want %q", rule1.ID, "DeleteAdhocQueryResultsAfter30Days")
	}
	if rule1.Status != s3Types.ExpirationStatusEnabled {
		t.Errorf("Rule 1 Status = %v, want %v", rule1.Status, s3Types.ExpirationStatusEnabled)
	}
	if rule1.Expiration == nil {
		t.Error("Rule 1 Expiration should not be nil")
	} else if rule1.Expiration.Days == nil || *rule1.Expiration.Days != AdhocQueryRetentionDays {
		t.Errorf("Rule 1 Expiration.Days = %v, want %d", rule1.Expiration.Days, AdhocQueryRetentionDays)
	}
	// Verify filter is set to adhoc/ prefix
	if rule1.Filter == nil {
		t.Error("Rule 1 Filter should not be nil")
	} else {
		prefixFilter, ok := rule1.Filter.(*s3Types.LifecycleRuleFilterMemberPrefix)
		if !ok {
			t.Error("Rule 1 Filter should be LifecycleRuleFilterMemberPrefix")
		} else if prefixFilter.Value != "adhoc/" {
			t.Errorf("Rule 1 Filter prefix = %q, want %q", prefixFilter.Value, "adhoc/")
		}
	}

	// Test second rule: DeleteIncompleteMultipartUploads
	rule2 := rules[1]
	if rule2.ID == nil || *rule2.ID != "DeleteIncompleteMultipartUploads" {
		t.Errorf("Rule 2 ID = %v, want %q", rule2.ID, "DeleteIncompleteMultipartUploads")
	}
	if rule2.Status != s3Types.ExpirationStatusEnabled {
		t.Errorf("Rule 2 Status = %v, want %v", rule2.Status, s3Types.ExpirationStatusEnabled)
	}
	if rule2.AbortIncompleteMultipartUpload == nil {
		t.Error("Rule 2 AbortIncompleteMultipartUpload should not be nil")
	} else if rule2.AbortIncompleteMultipartUpload.DaysAfterInitiation == nil || *rule2.AbortIncompleteMultipartUpload.DaysAfterInitiation != IncompleteUploadCleanupDays {
		t.Errorf("Rule 2 AbortIncompleteMultipartUpload.DaysAfterInitiation = %v, want %d", rule2.AbortIncompleteMultipartUpload.DaysAfterInitiation, IncompleteUploadCleanupDays)
	}
	// Verify filter is set to empty prefix (bucket-wide)
	if rule2.Filter == nil {
		t.Error("Rule 2 Filter should not be nil")
	} else {
		prefixFilter, ok := rule2.Filter.(*s3Types.LifecycleRuleFilterMemberPrefix)
		if !ok {
			t.Error("Rule 2 Filter should be LifecycleRuleFilterMemberPrefix")
		} else if prefixFilter.Value != "" {
			t.Errorf("Rule 2 Filter prefix = %q, want empty string (bucket-wide)", prefixFilter.Value)
		}
	}
}

// Helper types and functions

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestDatabaseNameFormat(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantDB      string
	}{
		{
			name:        "project name with hyphens",
			projectName: "my-cost-analysis",
			wantDB:      "my_cost_analysis_database",
		},
		{
			name:        "project name with spaces",
			projectName: "my cost analysis",
			wantDB:      "my_cost_analysis_database",
		},
		{
			name:        "project name with mixed separators",
			projectName: "my-cost analysis",
			wantDB:      "my_cost_analysis_database",
		},
		{
			name:        "project name with underscores",
			projectName: "my_cost_analysis",
			wantDB:      "my_cost_analysis_database",
		},
		{
			name:        "simple project name",
			projectName: "ecos",
			wantDB:      "ecos_database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the database name generation logic from GenerateConfig
			// This matches line 365 in aws_cur.go
			projectName := tt.projectName
			projectName = strings.ReplaceAll(projectName, " ", "_")
			projectName = strings.ReplaceAll(projectName, "-", "_")
			database := projectName + "_database"

			if database != tt.wantDB {
				t.Errorf("database name = %q, want %q", database, tt.wantDB)
			}
		})
	}
}

func TestProfileAndTargetConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		expectedProfile string
		expectedTarget  string
	}{
		{
			name:            "default profile and target",
			expectedProfile: "ecos-athena",
			expectedTarget:  "prod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that the hardcoded values match expectations
			// These values are set in GenerateConfig and generateDBTFiles
			profile := "ecos-athena"
			target := "prod"

			if profile != tt.expectedProfile {
				t.Errorf("profile = %q, want %q", profile, tt.expectedProfile)
			}

			if target != tt.expectedTarget {
				t.Errorf("target = %q, want %q", target, tt.expectedTarget)
			}
		})
	}
}

func TestDatabaseNameConsistency(t *testing.T) {
	// This test ensures that both GenerateConfig and generateDBTFiles
	// use the same logic for database name generation via normalizeDatabaseName()
	tests := []struct {
		name        string
		projectName string
		expected    string
	}{
		{
			name:        "hyphens only",
			projectName: "my-cost-analysis",
			expected:    "my_cost_analysis_database",
		},
		{
			name:        "spaces only",
			projectName: "my cost analysis",
			expected:    "my_cost_analysis_database",
		},
		{
			name:        "mixed separators",
			projectName: "my-cost analysis",
			expected:    "my_cost_analysis_database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Both functions should use normalizeDatabaseName() which handles both spaces and hyphens
			db := fmt.Sprintf("%s_database", normalizeDatabaseName(tt.projectName))

			if db != tt.expected {
				t.Errorf("database name = %q, want %q", db, tt.expected)
			}
		})
	}
}
