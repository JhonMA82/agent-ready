package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeGit(t *testing.T, root string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

func TestDoctorTiers(t *testing.T) {
	root := t.TempDir()
	fakeGit(t, root)
	// git/opencode on PATH (real), recommended absent via restricted PATH.
	if runtime.GOOS == "windows" {
		t.Skip("fake PATH is Unix-only")
	}
	empty := t.TempDir()
	t.Setenv("PATH", empty)
	// Required tier absent -> fail.
	facts, err := Doctor(root)
	if err != nil {
		t.Fatal(err)
	}
	failed := false
	for _, check := range facts.Checks {
		if check.Name == "git" && check.Status != "fail" {
			t.Fatalf("git must fail with empty PATH: %+v", check)
		}
		if check.Status == "fail" {
			failed = true
		}
	}
	if !failed || facts.Healthy {
		t.Fatalf("doctor must be unhealthy: %+v", facts)
	}
	if !strings.Contains(facts.Summary(), "FAIL") {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

func TestDoctorHealthyWithWarnings(t *testing.T) {
	root := t.TempDir()
	fakeGit(t, root)
	facts, err := Doctor(root)
	if err != nil {
		t.Fatal(err)
	}
	// Real system usually has git; project state may warn without init.
	if !facts.Healthy {
		t.Fatalf("doctor unhealthy: %+v", facts.Checks)
	}
}

func TestRecommendSignals(t *testing.T) {
	root := t.TempDir()
	fakeGit(t, root)
	// No signals -> empty candidates.
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts.Candidates) != 0 {
		t.Fatalf("trivial repo must have no candidates: %+v", facts.Candidates)
	}
	// Lockfile + output dir -> Context7 + RTK.
	if err := os.WriteFile(filepath.Join(root, "go.sum"), []byte("checksum\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	facts, err = Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, candidate := range facts.Candidates {
		got[candidate.Candidate] = true
		if candidate.Capability == "" || candidate.Signal == "" || candidate.Observed == "" {
			t.Fatalf("candidate missing fields: %+v", candidate)
		}
	}
	if !got["Context7"] || !got["RTK"] {
		t.Fatalf("expected Context7+RTK candidates, got %+v", got)
	}
	if facts.Summary() == "No capability candidates" {
		t.Fatalf("summary: %s", facts.Summary())
	}
}
