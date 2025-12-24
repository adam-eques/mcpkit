// Package version exposes build metadata stamped in at link time via -ldflags.
package version

import "runtime/debug"

// These variables are overridden at build time, for example:
//
//	go build -ldflags "-X github.com/adam-eques/mcpkit/internal/version.Version=1.2.3"
var (
	// Version is the semantic version of the build.
	Version = "0.0.0-dev"
	// Commit is the short git SHA of the build.
	Commit = "unknown"
	// Date is the RFC 3339 build timestamp.
	Date = "unknown"
)

// String returns a human-readable build identifier.
func String() string {
	v := Version
	if Commit != "unknown" {
		v += "+" + Commit
	}
	return v
}

// FromBuildInfo fills unset fields from the embedded module build info, so
// `go install`-ed binaries still report a meaningful revision.
func FromBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if Commit == "unknown" && len(s.Value) >= 7 {
				Commit = s.Value[:7]
			}
		case "vcs.time":
			if Date == "unknown" {
				Date = s.Value
			}
		}
	}
}
