package lifecycle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

type assetUpgrade struct {
	path string
	old  []byte
	new  bool
}

func TestEmbeddedAssetUpgradeGate(t *testing.T) {
	runEmbeddedAssetUpgradeGate(t, "slice-1-ownership-safe-upgrades", []assetUpgrade{
		{path: ".opencode/commands/agent-ready.md", old: []byte("old command\n")},
		{path: ".agent-ready/skills/agent-ready-orchestrator/SKILL.md", old: []byte("old skill\n")},
		{path: ".agent-ready/references/skill-system/anti-patterns.md", new: true},
	})
}

// TestEmbeddedAssetUpgradeGateSlice7 reruns the Slice 1 parameterized gate
// over Slice 7's declared changed assets (task 4.2): the four initial-audit
// contract files must advance unchanged, preserve modifications, install when
// absent, and keep protected state intact. As in Slice 1, the new-marked
// entry models an older install that lacks that path so the gate can prove
// absent installation and unmanaged collisions for this slice's asset set.
func TestEmbeddedAssetUpgradeGateSlice7(t *testing.T) {
	runEmbeddedAssetUpgradeGate(t, "slice-7-driven-audit", []assetUpgrade{
		{path: ".agent-ready/skills/agent-ready-orchestrator/SKILL.md", old: []byte("old orchestrator skill\n")},
		{path: ".agent-ready/skills/repository-analysis/SKILL.md", old: []byte("old repository analysis\n")},
		{path: ".agent-ready/skills/agent-ready-orchestrator/references/audit-flow.md", new: true},
		{path: ".agent-ready/skills/repository-analysis/references/inventory-facts.md", old: []byte("old inventory facts\n")},
	})
}

// TestEmbeddedAssetUpgradeGateSlice8 reruns the Slice 1 parameterized gate
// over Slice 8's declared changed assets (task 4.3): the two relevant-sync
// contract files must advance unchanged, preserve modifications, install
// when absent, and keep protected state intact. As in Slices 1 and 7, the
// new-marked entry (sync-flow.md) models an older install lacking that path
// so the gate can prove absent installation and unmanaged collisions for
// this slice's asset set; the command entry follows the Slice 1 pattern as
// an owned-path control.
func TestEmbeddedAssetUpgradeGateSlice8(t *testing.T) {
	runEmbeddedAssetUpgradeGate(t, "slice-8-driven-sync", []assetUpgrade{
		{path: ".agent-ready/skills/incremental-evolution/SKILL.md", old: []byte("old evolution skill\n")},
		{path: ".opencode/commands/agent-ready.md", old: []byte("old sync command\n")},
		{path: ".agent-ready/skills/incremental-evolution/references/sync-flow.md", new: true},
	})
}

func runEmbeddedAssetUpgradeGate(t *testing.T, slice string, changedAssets []assetUpgrade) {
	t.Helper()
	t.Run(slice+"/advances-and-installs", func(t *testing.T) {
		root, desired := installOlderVersion(t, changedAssets)
		protected := writeProtectedState(t, root)
		plan, err := UpdatePlan(root)
		if err != nil {
			t.Fatal(err)
		}
		assertPathOrdered(t, plan.Changes())
		if err := ApplyUpdate(plan); err != nil {
			t.Fatal(err)
		}
		for _, upgrade := range changedAssets {
			got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(upgrade.path)))
			if err != nil || !bytes.Equal(got, desired[upgrade.path]) {
				t.Fatalf("%s did not advance/install: %v", upgrade.path, err)
			}
		}
		manifest, err := os.ReadFile(filepath.Join(root, ".agent-ready", "manifest.json"))
		if err != nil || !bytes.Contains(manifest, []byte(".agent-ready/skills/obsolete/SKILL.md")) {
			t.Fatalf("unchanged obsolete asset lost ownership: %v", err)
		}
		assertProtectedState(t, root, protected)
		again, err := UpdatePlan(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, change := range again.Changes() {
			if change.Kind() != "noop" {
				t.Fatalf("second plan contains %s for %s", change.Kind(), change.Path())
			}
		}
	})

	t.Run(slice+"/preserves-and-orders-conflicts", func(t *testing.T) {
		root, _ := installOlderVersion(t, changedAssets)
		protected := writeProtectedState(t, root)
		modified := changedAssets[1]
		collision := changedAssets[2]
		writeFile(t, root, modified.path, []byte("user modification\n"), 0o644)
		writeFile(t, root, collision.path, []byte("user collision\n"), 0o644)
		writeFile(t, root, ".agent-ready/skills/obsolete/SKILL.md", []byte("modified obsolete\n"), 0o644)
		_, err := UpdatePlan(root)
		if err == nil {
			t.Fatal("expected ownership conflicts")
		}
		message := err.Error()
		if !strings.Contains(message, "modified owned asset") || !strings.Contains(message, "modified obsolete asset") || !strings.Contains(message, "unmanaged collision") {
			t.Fatalf("conflict reasons missing: %s", message)
		}
		if strings.Index(message, collision.path) > strings.Index(message, modified.path) {
			t.Fatalf("conflicts are not path ordered: %s", message)
		}
		for path, want := range map[string][]byte{
			modified.path: []byte("user modification\n"), collision.path: []byte("user collision\n"),
			changedAssets[0].path: changedAssets[0].old, ".agent-ready/skills/obsolete/SKILL.md": []byte("modified obsolete\n"),
		} {
			got, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
			if readErr != nil || !bytes.Equal(got, want) {
				t.Fatalf("conflict planning changed %s: %q, %v", path, got, readErr)
			}
		}
		assertProtectedState(t, root, protected)
	})
}

func installOlderVersion(t *testing.T, upgrades []assetUpgrade) (string, map[string][]byte) {
	t.Helper()
	root := t.TempDir()
	initRepo(t, root)
	desired, err := bootstrap.Desired("opencode.json")
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string][]byte{}
	old := make([]bootstrap.File, 0, len(desired))
	byPath := map[string]assetUpgrade{}
	for _, upgrade := range upgrades {
		byPath[upgrade.path] = upgrade
	}
	for _, file := range desired {
		if file.Path == ".agent-ready/manifest.json" {
			continue
		}
		wanted[file.Path] = file.After
		if upgrade, ok := byPath[file.Path]; ok {
			if upgrade.new {
				continue
			}
			file.After = upgrade.old
		}
		old = append(old, file)
		writeFile(t, root, file.Path, file.After, file.Mode)
	}
	obsolete := bootstrap.File{Path: ".agent-ready/skills/obsolete/SKILL.md", After: []byte("owned obsolete\n"), Mode: 0o644}
	old = append(old, obsolete)
	writeFile(t, root, obsolete.Path, obsolete.After, obsolete.Mode)
	manifest, err := bootstrap.Manifest("opencode.json", old)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, ".agent-ready/manifest.json", manifest, 0o644)
	return root, wanted
}

func writeFile(t *testing.T, root, path string, data []byte, mode os.FileMode) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, mode); err != nil {
		t.Fatal(err)
	}
}

func writeProtectedState(t *testing.T, root string) map[string][]byte {
	t.Helper()
	protected := map[string][]byte{
		".agent-ready/state/model.json":         []byte("model"),
		".agent-ready/checkpoints/current.json": []byte("checkpoint"),
		".agent-ready/generated/report.md":      []byte("generated"),
	}
	for path, data := range protected {
		writeFile(t, root, path, data, 0o644)
	}
	return protected
}

func assertProtectedState(t *testing.T, root string, protected map[string][]byte) {
	t.Helper()
	for path, want := range protected {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("protected state changed at %s: %v", path, err)
		}
	}
}

func assertPathOrdered(t *testing.T, changes []Change) {
	t.Helper()
	for i := 1; i < len(changes); i++ {
		if changes[i-1].Path() > changes[i].Path() {
			t.Fatalf("changes are not path ordered: %s before %s", changes[i-1].Path(), changes[i].Path())
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
