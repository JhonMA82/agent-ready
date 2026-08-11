package lifecycle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/opencode"
)

func fakeBin(t *testing.T, name, version string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake executables are Unix-only")
	}
	dir := t.TempDir()
	script := "#!/bin/sh\nexit 0\n"
	if version != "" {
		script = "#!/bin/sh\nprintf '%s\\n' '" + version + "'\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func initRepo(t *testing.T, root string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	// Install the harness via the app path would be heavy; seed the minimal
	// manifest-backed layout used by the lifecycle commands.
	dirs := []string{".agent-ready/skills/agent-ready-orchestrator", ".agent-ready/state", ".agent-ready/checkpoints", ".opencode/commands"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"schema":"agent-ready.manifest/v1","install_version":"2","compatibility_version":"1.18.15","config_file":"opencode.json","config_path":"./.agent-ready/skills","assets":[{"path":".agent-ready/skills/agent-ready-orchestrator/SKILL.md","sha256":"` + zeroHash() + `","mode":420}]}` + "\n"
	if err := os.WriteFile(filepath.Join(root, ".agent-ready", "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agent-ready/skills/agent-ready-orchestrator", "SKILL.md"), []byte("placeholder"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".opencode/commands", "agent-ready.md"), []byte("---\ndescription: run\n---\n$ARGUMENTS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func zeroHash() string {
	return strings.Repeat("0", 64)
}

func TestStatusUninitializedNotError(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	uninit := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(uninit, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", uninit, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	facts, err := Status(uninit)
	if err != nil || facts.Initialized || facts.AssetCount != 0 {
		t.Fatalf("uninitialized status: %+v %v", facts, err)
	}
	if !strings.Contains(facts.Summary(), "Not initialized") {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

func TestStatusInitializedAndMismatch(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	facts, err := Status(root)
	if err != nil || !facts.Initialized || facts.ManifestSchema != "agent-ready.manifest/v1" || facts.InstallVersion != "2" {
		t.Fatalf("status: %+v %v", facts, err)
	}
	// Drift the owned file -> mismatch listed.
	if err := os.WriteFile(filepath.Join(root, ".agent-ready/skills/agent-ready-orchestrator", "SKILL.md"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	facts, err = Status(root)
	if err != nil || len(facts.MismatchPaths) != 1 || !strings.Contains(facts.MismatchPaths[0], "agent-ready-orchestrator") {
		t.Fatalf("mismatch: %+v %v", facts, err)
	}
	if !strings.Contains(facts.Summary(), "mismatches: 1") {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

func TestDoctorHealthyAndRequiredFail(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	bin := fakeBin(t, "opencode", opencode.TestedVersion())
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	facts, err := Doctor(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !facts.Healthy {
		t.Fatalf("doctor must be healthy: %+v", facts.Checks)
	}
	// Missing opencode -> required fail.
	empty := t.TempDir()
	t.Setenv("PATH", empty)
	facts, err = Doctor(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if facts.Healthy || !strings.Contains(facts.Summary(), "FAIL") {
		t.Fatalf("doctor must fail: %+v", facts.Checks)
	}
	failed := false
	for _, check := range facts.Checks {
		if check.Status == "fail" {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("no fail checks: %+v", facts.Checks)
	}
}
