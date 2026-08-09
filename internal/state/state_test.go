package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadMissingAll(t *testing.T) {
	root := t.TempDir()
	facts, err := Read(root)
	if err != nil || facts.SchemaVersion != SchemaVersion || len(facts.Files) != 4 {
		t.Fatalf("facts: %v, %+v", err, facts)
	}
	for _, file := range facts.Files {
		if file.Exists || file.Path == "" {
			t.Fatalf("missing file must report exists=false: %+v", file)
		}
	}
}

func TestReadPresentFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".agent-ready", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	decisions := "{\"action\":\"noop\"}\n{\"action\":\"create\"}\n"
	provenance := "line1\nline2\nline3\n"
	if err := os.WriteFile(filepath.Join(dir, "decisions.jsonl"), []byte(decisions), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "provenance.jsonl"), []byte(provenance), 0o644); err != nil {
		t.Fatal(err)
	}
	facts, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]FileFact{}
	for _, file := range facts.Files {
		byPath[file.Path] = file
	}
	if d := byPath[".agent-ready/state/decisions.jsonl"]; !d.Exists || d.Entries != 2 || d.Bytes != int64(len(decisions)) {
		t.Fatalf("decisions facts: %+v", d)
	}
	if p := byPath[".agent-ready/state/provenance.jsonl"]; !p.Exists || p.Entries != 3 {
		t.Fatalf("provenance facts: %+v", p)
	}
	if y := byPath[".agent-ready/state/artifact-graph.yaml"]; y.Exists || y.Entries != 0 {
		t.Fatalf("yaml facts: %+v", y)
	}
	if _, err := time.Parse(time.RFC3339, byPath[".agent-ready/state/decisions.jsonl"].ModTime); err != nil {
		t.Fatalf("mod_time not RFC3339: %v", err)
	}
	if facts.Summary() != "State files: 2/4 present" {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

func TestReadDeterministicOrder(t *testing.T) {
	root := t.TempDir()
	first, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := range first.Files {
		if first.Files[i].Path != second.Files[i].Path {
			t.Fatalf("non-deterministic order: %+v vs %+v", first.Files, second.Files)
		}
	}
}
