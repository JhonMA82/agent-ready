package bootstrap

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanOwnershipAndIdempotency(t *testing.T) {
	root := t.TempDir()
	first, err := Plan(root, "opencode.json")
	if err != nil {
		t.Fatalf("first plan = %#v, %v", first, err)
	}
	var manifestSeen bool
	for _, file := range first {
		if file.Path == ".agent-ready/manifest.json" {
			manifestSeen = true
		}
	}
	if !manifestSeen {
		t.Fatal("ownership manifest missing from plan")
	}
	for _, file := range first {
		full := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, file.After, file.Mode); err != nil {
			t.Fatal(err)
		}
	}
	second, err := Plan(root, "opencode.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatalf("plan size changed between runs: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Path != second[i].Path || !bytes.Equal(first[i].After, second[i].After) || !bytes.Equal(second[i].Before, second[i].After) {
			t.Fatalf("non-idempotent file %q", first[i].Path)
		}
	}
}

func TestPlanRefusesConflicts(t *testing.T) {
	for _, name := range []string{"unowned target", "modified owned target", "modified owned mode"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			files, err := Plan(root, "opencode.json")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if name == "unowned target" && file.Path != ".opencode/commands/agent-ready.md" {
					continue
				}
				full := filepath.Join(root, filepath.FromSlash(file.Path))
				_ = os.MkdirAll(filepath.Dir(full), 0o755)
				_ = os.WriteFile(full, file.After, file.Mode)
			}
			if name == "modified owned target" {
				_ = os.WriteFile(filepath.Join(root, ".opencode/commands/agent-ready.md"), []byte("changed"), 0o644)
			}
			if name == "modified owned mode" {
				_ = os.Chmod(filepath.Join(root, ".opencode/commands/agent-ready.md"), 0o600)
			}
			if _, err := Plan(root, "opencode.json"); err == nil || !strings.Contains(err.Error(), "target") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

// TestEmbeddedWalkMatchesPlannedAssets locks the data-driven golden: every
// embedded asset (other than the manifest marker) is routed to exactly one
// planned file whose bytes are identical to the embedded walk, and the plan
// contains no other assets.
func TestEmbeddedWalkMatchesPlannedAssets(t *testing.T) {
	root := t.TempDir()
	files, err := Plan(root, "opencode.json")
	if err != nil {
		t.Fatal(err)
	}
	planned := map[string]File{}
	for _, file := range files {
		planned[file.Path] = file
	}
	var embedded int
	err = fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "assets/manifest.json" {
			return nil
		}
		embedded++
		rel := strings.TrimPrefix(path, "assets/")
		target, err := route(rel)
		if err != nil {
			t.Fatalf("unroutable embedded asset %q: %v", path, err)
		}
		file, ok := planned[target]
		if !ok {
			t.Fatalf("embedded asset %q is not planned at %q", path, target)
		}
		data, err := assetsFS.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(file.After, data) {
			t.Fatalf("planned bytes for %q differ from the embedded walk", target)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(planned) != embedded+1 {
		t.Fatalf("plan has %d files, embedded walk has %d assets", len(planned), embedded)
	}
}

// TestCanonicalManifestHashesCoverAllAssets locks the ownership golden: the
// desired manifest lists every planned asset with its exact sha256 and mode,
// and keeps the schema and the bumped install_version.
func TestCanonicalManifestHashesCoverAllAssets(t *testing.T) {
	root := t.TempDir()
	files, err := Plan(root, "opencode.json")
	if err != nil {
		t.Fatal(err)
	}
	var desired []byte
	assets := map[string][]byte{}
	for _, file := range files {
		if file.Path == ".agent-ready/manifest.json" {
			desired = file.After
			continue
		}
		assets[file.Path] = file.After
	}
	if desired == nil {
		t.Fatal("desired manifest missing from plan")
	}
	var m manifest
	if err := json.Unmarshal(desired, &m); err != nil {
		t.Fatal(err)
	}
	if m.Schema != "agent-ready.manifest/v1" || m.InstallVersion != "2" || m.CompatibilityVersion != "1.18.15" {
		t.Fatalf("manifest marker = %+v", m)
	}
	if m.ConfigFile != "opencode.json" || m.ConfigPath != "./.agent-ready/skills" {
		t.Fatalf("config fields = %q %q", m.ConfigFile, m.ConfigPath)
	}
	if len(m.Assets) != len(assets) {
		t.Fatalf("manifest lists %d assets, plan has %d", len(m.Assets), len(assets))
	}
	for _, a := range m.Assets {
		after, ok := assets[a.Path]
		if !ok {
			t.Fatalf("manifest lists unplanned asset %q", a.Path)
		}
		sum := sha256.Sum256(after)
		if a.SHA256 != fmt.Sprintf("%x", sum) {
			t.Fatalf("manifest hash mismatch for %q", a.Path)
		}
		if a.Mode != uint32(0o644) {
			t.Fatalf("manifest mode for %q = %o", a.Path, a.Mode)
		}
	}
}

func TestManifestCanonicalizesCallerOrder(t *testing.T) {
	files := []File{
		{Path: "z", After: []byte("z"), Mode: 0o644},
		{Path: "a", After: []byte("a"), Mode: 0o600},
	}
	first, err := Manifest("opencode.json", files)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Manifest("opencode.json", []File{files[1], files[0]})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || files[0].Path != "z" {
		t.Fatal("manifest must be deterministic without mutating caller order")
	}
}
