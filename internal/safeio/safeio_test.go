package safeio_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/safeio"
)

func TestCommitOrdersManifestLastAndPreservesDesiredModes(t *testing.T) {
	root := t.TempDir()
	p := mustPlan(t, root)
	var order []string
	_, err := safeio.Commit(p, safeio.Options{Hook: func(phase, path string) error {
		if phase == "commit" {
			order = append(order, path)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".agent-ready/skills/agent-ready-orchestrator/SKILL.md", ".opencode/commands/agent-ready.md", "opencode.json", ".agent-ready/manifest.json"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order=%v", order)
	}
	for _, c := range p.Changes() {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(c.Path())))
		if err != nil || !bytes.Equal(data, c.After()) {
			t.Fatalf("%s bytes: %v", c.Path(), err)
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(c.Path())))
		if err != nil || info.Mode().Perm() != c.Mode() {
			t.Fatalf("%s mode: %v", c.Path(), err)
		}
	}
}

func TestCommitRevalidatesAllStateBeforeWriting(t *testing.T) {
	root := t.TempDir()
	p := mustPlan(t, root)
	changed := p.Changes()[0]
	path := filepath.Join(root, filepath.FromSlash(changed.Path()))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("late change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := safeio.Commit(p, safeio.Options{}); err == nil {
		t.Fatal("expected revalidation failure")
	}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready/transaction.json")); !os.IsNotExist(err) {
		t.Fatalf("journal must not be written: %v", err)
	}
	for _, c := range p.Changes()[1:] {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(c.Path()))); !os.IsNotExist(err) {
			t.Fatalf("unexpected write to %s", c.Path())
		}
	}
}

func TestCommitFailureRollsBackExactPriorState(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "opencode.json")
	before := []byte("{}\n")
	if err := os.WriteFile(config, before, 0o600); err != nil {
		t.Fatal(err)
	}
	p := mustPlan(t, root)
	commits := 0
	_, err := safeio.Commit(p, safeio.Options{Hook: func(phase, _ string) error {
		if phase == "commit" {
			commits++
			if commits == 2 {
				return errors.New("injected commit failure")
			}
		}
		return nil
	}})
	if err == nil {
		t.Fatal("expected failure")
	}
	assertPriorState(t, root, p, before, 0o600)
}

func TestIncompleteRollbackRetainsRecoverableJournal(t *testing.T) {
	root := t.TempDir()
	p := mustPlan(t, root)
	commits := 0
	result, err := safeio.Commit(p, safeio.Options{Hook: func(phase, _ string) error {
		switch phase {
		case "commit":
			commits++
			if commits == 2 {
				return errors.New("injected commit failure")
			}
		case "rollback":
			return errors.New("injected rollback failure")
		}
		return nil
	}})
	if err == nil || result.RecoveryPath == "" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := safeio.Recover(root, safeio.Options{}); err != nil {
		t.Fatal(err)
	}
	assertPriorState(t, root, p, nil, 0)
}

func mustPlan(t *testing.T, root string) app.Plan {
	t.Helper()
	p, err := app.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func assertPriorState(t *testing.T, root string, p app.Plan, config []byte, mode os.FileMode) {
	t.Helper()
	for _, c := range p.Changes() {
		path := filepath.Join(root, filepath.FromSlash(c.Path()))
		if c.Path() != "opencode.json" || config == nil {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("new file remains at %s: %v", c.Path(), err)
			}
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, config) {
			t.Fatalf("config bytes: %q %v", data, err)
		}
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != mode.Perm() {
			t.Fatalf("config mode: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready/transaction.json")); !os.IsNotExist(err) {
		t.Fatalf("journal remains: %v", err)
	}
}
