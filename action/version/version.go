package version

import (
	"fmt"
	"runtime"
)

// Version information - will be set during build
var (
	// Version is the semantic version of the application
	Version = "v1.0.0"

	// Commit is the git commit hash
	Commit = "unknown"

	// Date is the build date
	Date = "unknown"

	// GoVersion is the Go version used to build the application
	GoVersion = runtime.Version()
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

// GetInfo returns the version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("nobl9-action version %s (commit: %s, date: %s, go: %s)",
		Version, Commit, Date, GoVersion)
}

// Short returns a short version string
func Short() string {
	return fmt.Sprintf("v%s", Version)
}
