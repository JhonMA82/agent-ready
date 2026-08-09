// Package version carries build-injected metadata set via ldflags
// (goreleaser: -X .../internal/version.Version={{.Version}}, Commit, Date).
// Without injection the binary reports "dev".
package version

import "fmt"

// Build metadata; overridden at link time. Mutable for tests.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

// String renders the canonical version line.
func String() string {
	if Commit == "" {
		return fmt.Sprintf("agent-ready %s", Version)
	}
	return fmt.Sprintf("agent-ready %s (%s)", Version, Commit)
}
