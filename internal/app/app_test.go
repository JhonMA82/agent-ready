package app

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gentle-ai/agent-ready/internal/opencode"
	"github.com/gentle-ai/agent-ready/internal/plan"
)

func TestDryRunPlanAndOwnedRerun(t *testing.T) {
	root := t.TempDir()
	if err := exec.Command("git", "-C", root, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte("#!/bin/sh\nprintf '%s\\n' '"+opencode.RequiredVersion()+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	old, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	p, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	dry, refused := Result(p, true), Result(p, false)
	if dry.Outcome != plan.DryRun || refused.Outcome != plan.Refused || !reflect.DeepEqual(dry.Actions, refused.Actions) {
		t.Fatalf("dry/refused plans differ: %#v %#v", dry, refused)
	}
	if got := Init(context.Background(), true); !reflect.DeepEqual(got.Actions, dry.Actions) {
		t.Fatalf("runtime dry-run differs: %#v", got)
	}
	for _, change := range p.Changes() {
		full := filepath.Join(root, filepath.FromSlash(change.Path()))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, change.After(), change.Mode()); err != nil {
			t.Fatal(err)
		}
	}
	rerun, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range rerun.Changes() {
		if change.Kind() != "noop" || !bytes.Equal(change.Before(), change.After()) {
			t.Fatalf("rerun change = %#v", change)
		}
	}
	if got := Init(context.Background(), false); got.Outcome != plan.Noop {
		t.Fatalf("rerun result = %#v", got)
	}
}
