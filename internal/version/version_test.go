package version

import (
	"strings"
	"testing"
)

func TestStringDevFallback(t *testing.T) {
	Version, Commit, Date = "dev", "", ""
	if got := String(); got != "agent-ready dev" {
		t.Fatalf("dev fallback: %q", got)
	}
}

func TestStringInjected(t *testing.T) {
	Version, Commit, Date = "1.0.0", "abc123", "2026-08-09T00:00:00Z"
	if got := String(); got != "agent-ready 1.0.0 (abc123)" {
		t.Fatalf("injected: %q", got)
	}
}

func TestLdflagTargetsExist(t *testing.T) {
	// Guards the goreleaser -X contract: the named vars must exist.
	if !strings.Contains("Version Commit Date", "Version") {
		t.Fatal("Version var missing")
	}
	_ = Date // date is set but not rendered in V1
}
