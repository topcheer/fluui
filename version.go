package fluui

import "fmt"

// Version is the current Fluui version. This is overridden at build time
// using -ldflags "-X github.com/topcheer/fluui.Version=v1.0.0".
// When not overridden, it defaults to "dev".
var Version = "dev"

// VersionInfo holds structured version information.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// Info returns the current version info. Commit and Date are populated
// via build-time ldflags when using goreleaser or CI builds.
var Info = VersionInfo{
	Version: Version,
}

// String returns a human-readable version string.
func (v VersionInfo) String() string {
	if v.Commit != "" && v.Date != "" {
		return fmt.Sprintf("fluui %s (commit: %s, built: %s)", v.Version, v.Commit, v.Date)
	}
	return fmt.Sprintf("fluui %s", v.Version)
}

// IsDev returns true when running from source without a version tag.
func (v VersionInfo) IsDev() bool {
	return v.Version == "dev" || v.Version == ""
}

// ComponentCount returns the total number of components in the library.
const ComponentCount = 126

// ProtocolCount returns the total number of terminal protocol functions supported.
const ProtocolCount = 64
