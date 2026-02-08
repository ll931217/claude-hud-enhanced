package version

// Build information populated at build time
var (
	// Version is the application version
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = "unknown"
)

// VersionInfo returns complete version information
func VersionInfo() string {
	return Version
}

// FullVersionInfo returns detailed version information
func FullVersionInfo() string {
	if Version == "dev" {
		return "claude-hud-enhanced version dev (development build)"
	}
	return "claude-hud-enhanced version " + Version
}

// BuildInfo returns complete build information
func BuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     GitCommit,
		"built_at":   BuildDate,
		"go_version": GoVersion,
	}
}
