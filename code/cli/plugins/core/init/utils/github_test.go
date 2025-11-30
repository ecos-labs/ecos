package utils

import (
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "adds v prefix when missing",
			version: "1.0.0",
			want:    "v1.0.0",
		},
		{
			name:    "keeps v prefix when present",
			version: "v1.0.0",
			want:    "v1.0.0",
		},
		{
			name:    "handles version with patch",
			version: "1.2.3",
			want:    "v1.2.3",
		},
		{
			name:    "handles version with v prefix and patch",
			version: "v2.5.1",
			want:    "v2.5.1",
		},
		{
			name:    "rejects 'latest' keyword returns empty",
			version: "latest",
			want:    "",
		},
		{
			name:    "rejects 'main' branch returns empty",
			version: "main",
			want:    "",
		},
		{
			name:    "rejects 'master' branch returns empty",
			version: "master",
			want:    "",
		},
		{
			name:    "handles empty string adds v prefix",
			version: "",
			want:    "v",
		},
		{
			name:    "handles version with leading zeros",
			version: "01.02.03",
			want:    "v01.02.03",
		},
		{
			name:    "handles version with beta suffix",
			version: "1.0.0-beta",
			want:    "v1.0.0-beta",
		},
		{
			name:    "handles version with rc suffix",
			version: "v2.0.0-rc1",
			want:    "v2.0.0-rc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVersion(tt.version)
			if got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
