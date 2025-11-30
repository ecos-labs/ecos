package utils

import (
	"testing"
)

func TestIsValidRegion(t *testing.T) {
	tests := []struct {
		name   string
		region string
		want   bool
	}{
		// Valid AWS regions
		{
			name:   "us-east-1 is valid",
			region: "us-east-1",
			want:   true,
		},
		{
			name:   "us-west-2 is valid",
			region: "us-west-2",
			want:   true,
		},
		{
			name:   "eu-west-1 is valid",
			region: "eu-west-1",
			want:   true,
		},
		{
			name:   "eu-central-1 is valid",
			region: "eu-central-1",
			want:   true,
		},
		{
			name:   "ap-southeast-1 is valid",
			region: "ap-southeast-1",
			want:   true,
		},
		{
			name:   "ap-northeast-1 is valid",
			region: "ap-northeast-1",
			want:   true,
		},
		{
			name:   "sa-east-1 is valid",
			region: "sa-east-1",
			want:   true,
		},
		{
			name:   "ca-central-1 is valid",
			region: "ca-central-1",
			want:   true,
		},
		{
			name:   "me-south-1 is valid",
			region: "me-south-1",
			want:   true,
		},
		{
			name:   "af-south-1 is valid",
			region: "af-south-1",
			want:   true,
		},
		// Invalid regions
		{
			name:   "empty string is invalid",
			region: "",
			want:   false,
		},
		{
			name:   "invalid-region is invalid",
			region: "invalid-region",
			want:   false,
		},
		{
			name:   "us-east-99 is invalid",
			region: "us-east-99",
			want:   false,
		},
		{
			name:   "random-string is invalid",
			region: "random-string-123",
			want:   false,
		},
		{
			name:   "localhost is invalid",
			region: "localhost",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidRegion(tt.region)
			if got != tt.want {
				t.Errorf("IsValidRegion(%q) = %v, want %v", tt.region, got, tt.want)
			}
		})
	}
}
