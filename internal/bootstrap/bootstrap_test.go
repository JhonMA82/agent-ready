package bootstrap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanOwnershipAndIdempotency(t *testing.T) {
	root := t.TempDir()
	first, err := Plan(root, "opencode.json")
	if err != nil || len(first) != 3 {
		t.Fatalf("first plan = %#v, %v", first, err)
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
