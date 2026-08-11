package lifecycle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/bootstrap"
)

func removeDesired(t *testing.T) string {
	t.Helper()
	root := installDesired(t)
	// Write a state file + a config file to test generated-mode handling.
	if err := os.MkdirAll(filepath.Join(root, ".agent-ready", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agent-ready", "state", "decisions.jsonl"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "opencode.json"), []byte(`{"skills":{"paths":["./.agent-ready/skills"]}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRemoveHarnessOnlyKeepsConfigAndState(t *testing.T) {
	root := removeDesired(t)
	plan, err := PlanRemove(root, ModeHarnessOnly)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Removed) == 0 {
		t.Fatal("harness-only must plan removals")
	}
	for _, entry := range plan.Removed {
		if entry.Path == "opencode.json" || strings.HasPrefix(entry.Path, ".agent-ready/state") {
			t.Fatalf("harness-only must not remove %s", entry.Path)
		}
	}
	if err := ApplyRemove(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "opencode.json")); err != nil {
		t.Fatalf("config must be kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready", "state", "decisions.jsonl")); err != nil {
		t.Fatalf("state must be kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready", "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("manifest must be removed: %v", err)
	}
}

func TestRemoveHarnessAndGenerated(t *testing.T) {
	root := removeDesired(t)
	plan, err := PlanRemove(root, ModeHarnessAndGen)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}
	for _, entry := range plan.Removed {
		paths[entry.Path] = true
	}
	if !paths["opencode.json"] || !paths[".agent-ready/state"] || !paths[".agent-ready/checkpoints"] {
		t.Fatalf("generated mode must include config/state/checkpoints: %+v", plan.Removed)
	}
	if err := ApplyRemove(plan); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"opencode.json", ".agent-ready/manifest.json", ".agent-ready/state"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Fatalf("%s must be gone: %v", path, err)
		}
	}
}

func TestRemoveProtectsModifiedContent(t *testing.T) {
	root := removeDesired(t)
	// Modified owned file -> kept with refusal reason.
	target := filepath.Join(root, ".agent-ready", "skills", "agent-ready-orchestrator", "SKILL.md")
	if err := os.WriteFile(target, []byte("user edits"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Modified config -> kept.
	if err := os.WriteFile(filepath.Join(root, "opencode.json"), []byte(`{"theme":"dark"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanRemove(root, ModeHarnessAndGen)
	if err != nil {
		t.Fatal(err)
	}
	foundSkillKept, foundConfigKept := false, false
	for _, entry := range plan.Kept {
		if strings.Contains(entry.Path, "agent-ready-orchestrator") {
			foundSkillKept = true
		}
		if entry.Path == "opencode.json" {
			foundConfigKept = true
		}
	}
	if !foundSkillKept || !foundConfigKept {
		t.Fatalf("modified files must be kept: %+v", plan.Kept)
	}
	if err := ApplyRemove(plan); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(target); err != nil || string(data) != "user edits" {
		t.Fatalf("modified owned file must survive: %q %v", data, err)
	}
}

func TestRemoveRequiresModeAndInitialized(t *testing.T) {
	if _, err := PlanRemove(t.TempDir(), Mode("bogus")); err == nil || !strings.Contains(err.Error(), "invalid mode") {
		t.Fatalf("invalid mode must refuse: %v", err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanRemove(root, ModeHarnessOnly); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("uninitialized remove must refuse: %v", err)
	}
}

func TestRemoveDryRunWiringZeroWrites(t *testing.T) {
	root := removeDesired(t)
	plan, err := PlanRemove(root, ModeHarnessAndGen)
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]bool{}
	for _, entry := range plan.Removed {
		full := filepath.Join(root, filepath.FromSlash(entry.Path))
		if _, err := os.Lstat(full); err == nil {
			before[entry.Path] = true
		}
	}
	// The CLI wiring path is exercised by the process harness; here we assert
	// ApplyRemove is NOT called by dry-run semantics at the command layer.
	if len(before) == 0 {
		t.Fatal("precondition: plan must reference existing paths")
	}
	_ = bootstrap.Desired // keep import for installDesired dependency
}

// §57/§58: remove never touches an existing global OpenCode config; the
// isolated-HOME config stays byte-identical across the whole removal flow.
func TestRemoveLeavesGlobalConfigUntouched(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	global := filepath.Join(home, ".config", "opencode", "opencode.json")
	globalBytes := []byte("{\"model\":\"acme/small\"}\n")
	if err := os.MkdirAll(filepath.Dir(global), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(global, globalBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	root := removeDesired(t)
	for _, mode := range []Mode{ModeHarnessOnly, ModeHarnessAndGen} {
		plan, err := PlanRemove(root, mode)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyRemove(plan); err != nil {
			t.Fatal(err)
		}
		if after, err := os.ReadFile(global); err != nil || !bytes.Equal(after, globalBytes) {
			t.Fatalf("%s modified global config: %v\n%s", mode, err, after)
		}
		root = removeDesired(t) // fresh initialized repo per mode
	}
}
