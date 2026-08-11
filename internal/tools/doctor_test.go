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
	// §49 provider checks probe real provider binaries, so isolate PATH to
	// the fake opencode plus git's own directory (host-sensitive otherwise).
	if runtime.GOOS == "windows" {
		t.Skip("fake PATH is Unix-only")
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := fakeBin(t, "opencode", "1.18.15")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+filepath.Dir(gitPath))
	facts, err := Doctor(root)
	if err != nil {
		t.Fatal(err)
	}
	// git/opencode present; project state may warn without init.
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
	// Output dir -> RTK + Headroom (D4: RTK evidence + output-pressure
	// signals coexist); lockfile-only Context7 stays absent (D3).
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
	if !got["RTK"] || !got["Headroom"] || got["Context7"] {
		t.Fatalf("expected RTK+Headroom without Context7, got %+v", got)
	}
	if facts.Summary() == "No capability candidates" {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

// §49: per-provider doctor checks. A provider is never reported healthy
// merely because a binary exists: a broken version and a missing project
// index are failures carrying a reason; absence is a warning (providers are
// never auto-installed).
func TestProviderDoctorChecks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake executables are Unix-only")
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDir := filepath.Dir(gitPath)
	root := t.TempDir()
	fakeGit(t, root)
	withIndex := t.TempDir()
	if err := os.MkdirAll(filepath.Join(withIndex, ".codegraph"), 0o755); err != nil {
		t.Fatal(err)
	}
	pathFor := func(bins ...string) string {
		return strings.Join(append(bins, gitDir), string(os.PathListSeparator))
	}
	healthy := fakeBin(t, "opencode", "1.18.15") + string(os.PathListSeparator) + fakeBin(t, "codegraph", "codegraph 1.5.0")
	broken := fakeBin(t, "opencode", "1.18.15") + string(os.PathListSeparator) + fakeBin(t, "codegraph", "garbage without version")
	for _, tc := range []struct {
		name, path, repo, wantStatus, wantDetail string
		wantHealthy                              bool
	}{
		{name: "healthy provider", path: pathFor(healthy), repo: withIndex, wantStatus: "ok", wantHealthy: true},
		{name: "broken version", path: pathFor(broken), repo: withIndex, wantStatus: "fail", wantDetail: "version does not parse"},
		{name: "missing project index", path: pathFor(healthy), repo: root, wantStatus: "fail", wantDetail: ".codegraph/"},
		{name: "absent provider", path: pathFor(fakeBin(t, "opencode", "1.18.15")), repo: root, wantStatus: "warning", wantDetail: "never auto-installed", wantHealthy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PATH", tc.path)
			facts, err := Doctor(tc.repo)
			if err != nil {
				t.Fatal(err)
			}
			if facts.Healthy != tc.wantHealthy {
				t.Fatalf("healthy=%v, want %v: %+v", facts.Healthy, tc.wantHealthy, facts.Checks)
			}
			for _, check := range facts.Checks {
				if check.Name == "provider:codegraph" {
					if check.Status != tc.wantStatus || tc.wantDetail != "" && !strings.Contains(check.Detail, tc.wantDetail) {
						t.Fatalf("provider:codegraph: %+v", check)
					}
				}
			}
		})
	}
}
