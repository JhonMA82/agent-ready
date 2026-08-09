package lifecycle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/bootstrap"
)

// installDesired writes the embedded desired assets into a repo and returns
// its root, mimicking a fresh init's installed state.
func installDesired(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initRepo(t, root)
	desired, err := bootstrap.Desired("opencode.json")
	if err != nil {
		t.Fatal(err)
	}
	// Rebuild the manifest like init would (assets over placeholder).
	for _, file := range desired {
		full := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, file.After, file.Mode); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestUpdateNoopWhenIdentical(t *testing.T) {
	root := installDesired(t)
	plan, err := UpdatePlan(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range plan.Changes() {
		if change.Kind() != "noop" {
			t.Fatalf("expected all noop, got %s %s", change.Kind(), change.Path())
		}
	}
}

func TestUpdateDriftTolerantPartial(t *testing.T) {
	root := installDesired(t)
	// Drift ONE owned asset; leave another pristine.
	target := filepath.Join(root, ".agent-ready", "skills", "agent-ready-orchestrator", "SKILL.md")
	otherPath := filepath.Join(root, ".opencode", "commands", "agent-ready.md")
	otherBefore, err := os.ReadFile(otherPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("drifted content"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := UpdatePlan(root)
	if err != nil {
		t.Fatal(err)
	}
	updated, noops := 0, 0
	for _, change := range plan.Changes() {
		switch change.Kind() {
		case "update":
			updated++
		case "noop":
			noops++
		}
	}
	if updated != 1 || noops != len(plan.Changes())-1 {
		t.Fatalf("expected exactly 1 update, got %d updates %d noops", updated, noops)
	}
	// Apply and confirm the drifted asset is restored, pristine untouched.
	if err := ApplyUpdate(plan); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(target)
	if err != nil || bytes.Equal(after, []byte("drifted content")) {
		t.Fatalf("drifted asset not restored: %q %v", after, err)
	}
	other, err := os.ReadFile(otherPath)
	if err != nil || !bytes.Equal(other, otherBefore) {
		t.Fatalf("pristine asset must be byte-identical: %v", err)
	}
	// State/checkpoints preserved.
	for _, dir := range []string{".agent-ready/state", ".agent-ready/checkpoints"} {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil || !info.IsDir() {
			t.Fatalf("runtime dir %s not preserved: %v", dir, err)
		}
	}
}

func TestUpdateRequiresInitialized(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdatePlan(root); err == nil || !bytes.Contains([]byte(err.Error()), []byte("not initialized")) {
		t.Fatalf("uninitialized update must refuse: %v", err)
	}
}

func TestUpdateAfterApplyIsNoop(t *testing.T) {
	root := installDesired(t)
	target := filepath.Join(root, ".agent-ready", "skills", "agent-ready-orchestrator", "SKILL.md")
	if err := os.WriteFile(target, []byte("drift"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := UpdatePlan(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyUpdate(plan); err != nil {
		t.Fatal(err)
	}
	again, err := UpdatePlan(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range again.Changes() {
		if change.Kind() != "noop" {
			t.Fatalf("post-update reconcile must be noop, got %s %s", change.Kind(), change.Path())
		}
	}
}
