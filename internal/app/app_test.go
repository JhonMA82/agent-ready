package app

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/opencode"
	"github.com/JhonMA82/agent-ready/internal/plan"
)

func TestDryRunPlanAndOwnedRerun(t *testing.T) {
	root := t.TempDir()
	if err := exec.Command("git", "-C", root, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte("#!/bin/sh\nprintf '%s\\n' '"+opencode.TestedVersion()+"'\n"), 0o755); err != nil {
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
	dry, changed := Result(p, "", true), Result(p, "", false)
	if dry.Outcome != plan.DryRun || changed.Outcome != plan.Changed || changed.NextStep != "/agent-ready" || !reflect.DeepEqual(dry.Actions, changed.Actions) {
		t.Fatalf("dry/changed plans differ: %#v %#v", dry, changed)
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
	if got := Result(p, root, true); got.Invocation != "" {
		t.Fatalf("equal invocation must be omitted: %+v", got)
	}
	if got := Result(p, root+"/nested", true); got.Invocation != root+"/nested" {
		t.Fatalf("nested invocation must be reported: %+v", got)
	}
}
