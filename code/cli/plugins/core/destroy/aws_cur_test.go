package destroy

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/smithy-go"
	cliConfig "github.com/ecos-labs/ecos-core/code/cli/config"
)

func TestAwsCurDestroyPlugin_Name(t *testing.T) {
	plugin := &AwsCurDestroyPlugin{}
	if plugin.Name() != "aws_cur" {
		t.Fatalf("Name() = %s, want aws_cur", plugin.Name())
	}
}

func TestAwsCurDestroyPlugin_LoadFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *cliConfig.EcosConfig
		wantErr     bool
		wantErrText string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "missing region",
			cfg: &cliConfig.EcosConfig{
				AWS: cliConfig.AWSRootConfig{
					ResultsBucket: "bucket",
					DBTWorkgroup:  "dbt",
				},
			},
			wantErr:     true,
			wantErrText: "aws.region missing",
		},
		{
			name: "no resources defined",
			cfg: &cliConfig.EcosConfig{
				AWS: cliConfig.AWSRootConfig{
					Region: "us-east-1",
				},
			},
			wantErr:     true,
			wantErrText: "does not contain any ecos-managed resource names",
		},
		{
			name: "valid config sets fields",
			cfg: &cliConfig.EcosConfig{
				AWS: cliConfig.AWSRootConfig{
					Region:         "us-east-1",
					ResultsBucket:  "bucket",
					DBTWorkgroup:   "dbt",
					AdhocWorkgroup: "adhoc",
				},
				Transform: cliConfig.TransformConfig{
					DBT: cliConfig.DBTConfig{
						AWSProfile: "profile",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &AwsCurDestroyPlugin{}
			err := plugin.LoadFromConfig(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("error %q does not contain %q", err, tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plugin.region != "us-east-1" ||
				plugin.bucket != "bucket" ||
				plugin.dbtWorkgroup != "dbt" ||
				plugin.adhocWorkgroup != "adhoc" ||
				plugin.awsProfile != "profile" {
				t.Fatalf("LoadFromConfig did not populate fields correctly: %+v", plugin)
			}
		})
	}
}

func TestHumanizePreviewError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		resource string
		want     string
	}{
		{
			name: "no such bucket",
			err:  mockAPIError{code: "NoSuchBucket", message: "bucket missing"},
			want: "resource not found",
		},
		{
			name: "resource not found workgroup",
			err:  mockAPIError{code: "ResourceNotFoundException", message: "missing"},
			want: "resource not found",
		},
		{
			name: "access denied",
			err:  mockAPIError{code: "AccessDenied", message: "denied"},
			want: "access denied",
		},
		{
			name: "unknown api error returns code",
			err:  mockAPIError{code: "SomeOtherError", message: "boom"},
			want: "SomeOtherError",
		},
		{
			name:     "generic error trimmed",
			err:      errors.New("custom failure"),
			resource: "custom",
			want:     "[resource] failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanizePreviewError(tt.err, tt.resource)
			if got != tt.want {
				t.Fatalf("humanizePreviewError(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

type mockAPIError struct {
	code    string
	message string
}

func (m mockAPIError) Error() string {
	return m.message
}

func (m mockAPIError) ErrorCode() string {
	return m.code
}

func (m mockAPIError) ErrorMessage() string {
	return m.message
}

func (m mockAPIError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultUnknown
}
