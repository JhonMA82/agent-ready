package checkpoint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for path, data := range files {
		os.MkdirAll(filepath.Join(root, filepath.Dir(path)), 0o755)
		if err := os.WriteFile(filepath.Join(root, path), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestSaveEnvelopeContract(t *testing.T) {
	root := writeRepo(t, map[string]string{"go.mod": "module x\n", "src/main.go": "package main\n"})
	env, err := Save(root, "exploration_plan", false)
	if err != nil || env.ID != "0001" || env.Stage != "exploration_plan" || env.Complete || env.SchemaVersion != SchemaVersion || env.SourceHash == "" || len(env.InventoryHashes) != 2 {
		t.Fatalf("save: %v, %+v", err, env)
	}
	if _, err := time.Parse(time.RFC3339, env.CreatedAt); err != nil {
		t.Fatalf("created_at not RFC3339: %v", err)
	}
	latest, err := os.ReadFile(filepath.Join(root, dir, "latest.json"))
	if err != nil {
		t.Fatal(err)
	}
	history, err := os.ReadFile(filepath.Join(root, dir, "history", env.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(latest, history) {
		t.Fatal("latest.json must equal history/<id>.json byte-for-byte")
	}
	var decoded Envelope
	if err := json.Unmarshal(latest, &decoded); err != nil || decoded.Stage != env.Stage {
		t.Fatalf("envelope file mismatch: %v", err)
	}
	// Source hash must stay stable across saves while files are unchanged.
	again, err := Save(root, "evidence", false)
	if err != nil || again.SourceHash != env.SourceHash {
		t.Fatalf("second save: %v", err)
	}
}

func TestStageTransitionsAndStatus(t *testing.T) {
	root := writeRepo(t, map[string]string{"a.txt": "a", "b.txt": "b"})
	first, err := Save(root, "plan", false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Save(root, "complete", true)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "0001" || second.ID != "0002" {
		t.Fatalf("ids must advance: %s then %s", first.ID, second.ID)
	}
	status, err := Status(root)
	if err != nil || !status.Exists || status.Checkpoint == nil || status.Checkpoint.ID != "0002" || !status.Checkpoint.Complete || status.Checkpoint.Stage != "complete" {
		t.Fatalf("status: %v, %+v", err, status)
	}
	for _, id := range []string{"0001", "0002"} {
		if _, err := os.Stat(filepath.Join(root, dir, "history", id+".json")); err != nil {
			t.Fatalf("history %s missing: %v", id, err)
		}
	}
	if _, err := Save(root, "", false); err == nil || !strings.Contains(err.Error(), "stage is required") {
		t.Fatalf("empty stage must error, got %v", err)
	}
	none, err := Status(writeRepo(t, map[string]string{"x.txt": "x"}))
	if err != nil || none.Exists || none.Checkpoint != nil {
		t.Fatalf("no-checkpoint status: %v, %+v", err, none)
	}
}

func TestChangesFirstRun(t *testing.T) {
	root := writeRepo(t, map[string]string{"b.txt": "b", "a.txt": "a"})
	facts, err := Changes(root)
	if err != nil || !facts.FirstRun || facts.Baseline.Exists || facts.SchemaVersion != ChangesSchemaVersion {
		t.Fatalf("first run: %v, %+v", err, facts)
	}
	want := []Change{{"a.txt", "added"}, {"b.txt", "added"}}
	if len(facts.Changes) != len(want) {
		t.Fatalf("full inventory must be added, got %+v", facts.Changes)
	}
	for i := range want {
		if facts.Changes[i] != want[i] {
			t.Fatalf("sorted added changes, got %+v", facts.Changes)
		}
	}
}

func TestChangesDiffAgainstBaseline(t *testing.T) {
	root := writeRepo(t, map[string]string{"a.txt": "a", "b.txt": "b"})
	if _, err := Save(root, "plan", false); err != nil {
		t.Fatal(err)
	}
	if facts, err := Changes(root); err != nil || facts.FirstRun || !facts.Baseline.Exists || len(facts.Changes) != 0 {
		t.Fatalf("unchanged repo: %v, %+v", err, facts)
	}
	os.WriteFile(filepath.Join(root, "a.txt"), []byte("a2"), 0o644)
	os.WriteFile(filepath.Join(root, "c.txt"), []byte("c"), 0o644)
	os.Remove(filepath.Join(root, "b.txt"))
	// Model state and checkpoint files must never surface as input changes.
	os.MkdirAll(filepath.Join(root, ".agent-ready", "state"), 0o755)
	os.WriteFile(filepath.Join(root, ".agent-ready", "state", "decisions.jsonl"), []byte("{}"), 0o644)
	facts, err := Changes(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{{"a.txt", "changed"}, {"b.txt", "removed"}, {"c.txt", "added"}}
	if len(facts.Changes) != len(want) || facts.Baseline.CheckpointID != "0001" {
		t.Fatalf("diff: %v, %+v", err, facts.Changes)
	}
	for i := range want {
		if facts.Changes[i] != want[i] {
			t.Fatalf("sorted diff, got %+v", facts.Changes)
		}
	}
}
