// Package version provides build-time version information for the ISP Visual Monitor application.
package version

import (
	"fmt"
	"runtime"
)

// Build-time variables (set via -ldflags)
var (
	// Version is the semantic version of the application
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// GitBranch is the git branch name
	GitBranch = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// Info contains version and build information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a human-readable version string
func (i Info) String() string {
	return fmt.Sprintf("ISP Visual Monitor %s (%s) built on %s with %s",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion)
}

// Short returns a short version string
func Short() string {
	return Version
}
