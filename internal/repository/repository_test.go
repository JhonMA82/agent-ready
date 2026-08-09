package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiscoverRootAndRefusal(t *testing.T) {
	git := "git"
	root := t.TempDir()
	if out, err := exec.Command(git, "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	for _, invocation := range []string{nested, filepath.Join("a", "b")} {
		got, err := Discover(context.Background(), invocation, git)
		if err != nil || got.Root != root || !filepath.IsAbs(got.Invocation) {
			t.Fatalf("Discover(%q) = %#v, %v", invocation, got, err)
		}
	}
	if _, err := Discover(context.Background(), t.TempDir(), git); err == nil {
		t.Fatal("expected no-worktree refusal")
	}
}

func TestContained(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		path string
		ok   bool
	}{{"safe/file", true}, {"../outside", false}, {"escape/file", false}, {"/absolute", false}} {
		_, err := Contained(root, tt.path)
		if (err == nil) != tt.ok {
			t.Errorf("%s: %v", tt.path, err)
		}
	}
}
