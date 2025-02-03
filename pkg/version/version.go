// Package version provides version information for the Vector Leads Scraper service.
package version

import (
	"fmt"
	"runtime"
)

var (
	// ServiceName is the name of the service
	ServiceName = "vector-leads-scraper"

	// Version is the current version of the service
	Version = "1.0.0"

	// GitCommit is the git commit hash, injected at build time
	GitCommit = "unknown"

	// BuildTime is the build timestamp, injected at build time
	BuildTime = "unknown"

	// GoVersion is the Go runtime version
	GoVersion = runtime.Version()

	// Platform is the OS/Arch combination
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info returns a map containing all version information
func Info() map[string]string {
	return map[string]string{
		"service":   ServiceName,
		"version":   Version,
		"gitCommit": GitCommit,
		"buildTime": BuildTime,
		"goVersion": GoVersion,
		"platform":  Platform,
	}
}

// String returns a string representation of the version information
func String() string {
	return fmt.Sprintf(
		"%s version %s\nGit commit: %s\nBuild time: %s\nGo version: %s\nPlatform: %s",
		ServiceName,
		Version,
		GitCommit,
		BuildTime,
		GoVersion,
		Platform,
	)
}
