package version

import (
	"fmt"
	"runtime"
	"time"
)

// These variables are set during build time using -ldflags
var (
	// Version is the current version of the CLI
	Version = "dev"

	// Date is the date when the binary was built (matches GoReleaser's .Date)
	Date = "unknown"

	// Commit is the git commit hash (matches GoReleaser's .Commit)
	Commit = "unknown"
)

// Info contains version information
type Info struct {
	Version   string    `json:"version"`
	BuildDate time.Time `json:"build_date"`
	GitCommit string    `json:"git_commit"`
	GoVersion string    `json:"go_version"`
	Platform  string    `json:"platform"`
}

// GetInfo returns version information
func GetInfo() Info {
	var buildTime time.Time
	if Date != "unknown" {
		var err error
		buildTime, err = time.Parse(time.RFC3339, Date)
		if err != nil {
			buildTime = time.Time{}
		}
	}

	return Info{
		Version:   Version,
		BuildDate: buildTime,
		GitCommit: Commit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a string representation of version information
func (i Info) String() string {
	return fmt.Sprintf("ecos version %s", i.Version)
}

// VerboseString returns a detailed string representation of version information
func (i Info) VerboseString() string {
	buildDate := "unknown"
	if !i.BuildDate.IsZero() {
		buildDate = i.BuildDate.Format(time.RFC3339)
	}

	return fmt.Sprintf(`ecos version %s
Git commit: %s
Build date: %s
Go version: %s
Platform: %s`,
		i.Version,
		i.GitCommit,
		buildDate,
		i.GoVersion,
		i.Platform)
}
