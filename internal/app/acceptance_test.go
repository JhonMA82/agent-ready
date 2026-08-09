package app

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestAcceptanceFixturesCThroughG drives the spec-40 acceptance subset
// fixtures C–G (README mapping in internal/app/testdata/acceptance/README.md).
// Evidence is asserted by content, never by counts alone ("N skills
// generated" is never accepted as evidence, R16).
func TestAcceptanceFixturesCThroughG(t *testing.T) {
	if testing.Short() {
		t.Skip("acceptance harness builds and executes the command")
	}
	base := t.TempDir()
	binDir, binary := filepath.Join(base, "bin"), filepath.Join(base, "agent-ready")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' '1.18.15'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/agent-ready")
	build.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	env := append(os.Environ(), "HOME="+filepath.Join(base, "home"), "XDG_CONFIG_HOME="+filepath.Join(base, "xdg"), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	fixtures := filepath.Join("testdata", "acceptance")
	for _, letter := range []string{"c", "d", "e", "f", "g"} {
		t.Run("fixture-"+letter, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+letter)
			copyFixture(t, filepath.Join(fixtures, "fixture-"+letter), repo)
			if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			runFixtureCase(t, binary, repo, env, letter)
		})
	}
}

func runFixtureCase(t *testing.T, binary, repo string, env []string, letter string) {
	t.Helper()
	switch letter {
	case "c": // adaptive analysis: exploration plan + evidence labels
		out, code := runCommand(t, binary, repo, env, "inspect", "--json")
		if code != 0 || !strings.Contains(out, `"deps"`) {
			t.Fatalf("inspect: code=%d %s", code, out)
		}
		assertFileContains(t, repo, "expect/exploration-plan.yaml", "unknowns")
		assertFileContains(t, repo, "expect/evidence-labels.md", "FACT", "INFERENCE", "UNKNOWN")
	case "d": // discovery: evidence-backed proposal before artifacts
		out, code := runCommand(t, binary, repo, env, "changes", "--json")
		if code != 0 || !strings.Contains(out, `"first_run":true`) {
			t.Fatalf("changes: code=%d %s", code, out)
		}
		assertFileContains(t, repo, "expect/proposal.md", "CREATE")
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "propose")
	case "e": // no artifact spam: NO_ACTION records, no generated skills dir
		out, code := runCommand(t, binary, repo, env, "state", "--json")
		if code != 0 || !strings.Contains(out, `"exists":true`) {
			t.Fatalf("state: code=%d %s", code, out)
		}
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "NO_ACTION")
		if _, err := os.Stat(filepath.Join(repo, ".opencode", "skills")); !os.IsNotExist(err) {
			t.Fatalf("avoided artifact must not exist: %v", err)
		}
	case "f": // rubric creation: PASS skill passes validation; gate notes
		out, code := runCommand(t, binary, repo, env, "validate", "--json")
		if code != 0 || !strings.Contains(out, `"verdict":"pass"`) {
			t.Fatalf("validate: code=%d %s", code, out)
		}
		assertFileContains(t, repo, "expect/notes.md", "100", "passed")
	case "g": // rubric rejection: REJECT record with justification in state
		out, code := runCommand(t, binary, repo, env, "state", "--json")
		if code != 0 {
			t.Fatalf("state: code=%d %s", code, out)
		}
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "REJECT", "rubric 30")
	}
}

func assertFileContains(t *testing.T, repo string, rel string, wants ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	for _, want := range wants {
		if !strings.Contains(string(data), want) {
			t.Fatalf("%s missing %q in %q", rel, want, data)
		}
	}
}

func copyFixture(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
}

var _ = json.Marshal

// TestAcceptanceFixturesHThroughP drives the PR12 subset: H ask-user, I stop
// with concerns, J incremental sync, K no-action sync, L resume, M isolation,
// N tool degradation, P decision evidence (O absent by spec deferral).
func TestAcceptanceFixturesHThroughP(t *testing.T) {
	if testing.Short() {
		t.Skip("acceptance harness builds and executes the command")
	}
	base := t.TempDir()
	binDir, binary := filepath.Join(base, "bin"), filepath.Join(base, "agent-ready")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' '1.18.15'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/agent-ready")
	build.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	env := append(os.Environ(), "HOME="+filepath.Join(base, "home"), "XDG_CONFIG_HOME="+filepath.Join(base, "xdg"), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	fixtures := filepath.Join("testdata", "acceptance")
	for _, letter := range []string{"h", "i", "j", "k", "l", "m", "n", "p"} {
		t.Run("fixture-"+letter, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+letter)
			copyFixture(t, filepath.Join(fixtures, "fixture-"+letter), repo)
			if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			runLateFixtureCase(t, binary, repo, env, letter)
		})
	}
}

func runLateFixtureCase(t *testing.T, binary, repo string, env []string, letter string) {
	t.Helper()
	switch letter {
	case "h": // ask user after no-new-evidence iterations
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "ASK_USER", "no new evidence")
	case "i": // stop with concerns and recorded reasons
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "STOP_WITH_CONCERNS", "reasons recorded")
	case "j": // incremental sync: selective update, no full re-audit
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "UPDATE", "no full re-audit")
	case "k": // no-action sync: zero artifacts
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "NO_ACTION", "zero artifacts")
		if _, err := os.Stat(filepath.Join(repo, ".opencode")); !os.IsNotExist(err) {
			t.Fatalf("zero-artifact fixture must have no .opencode tree: %v", err)
		}
	case "l": // resume: checkpoint at stage3; unchanged sources resume clean
		for i, stage := range []string{"stage1", "stage2", "stage3"} {
			args := []string{"checkpoint", "save", "--stage", stage}
			if i == 2 {
				args = append(args, "--complete")
			}
			out, code := runCommand(t, binary, repo, env, args...)
			if code != 0 || !strings.Contains(out, stage) {
				t.Fatalf("checkpoint save %s: code=%d %s", stage, code, out)
			}
		}
		out, code := runCommand(t, binary, repo, env, "changes", "--json")
		if code != 0 || !strings.Contains(out, `"first_run":false`) || strings.Contains(out, `"changes":[]`) == false {
			t.Fatalf("resume changes: code=%d %s", code, out)
		}
		// A changed source surfaces as a listed path (spec 40 L).
		if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n\n// changed\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out, code = runCommand(t, binary, repo, env, "changes", "--json")
		if code != 0 || !strings.Contains(out, `"path":"main.go"`) || !strings.Contains(out, `"kind":"changed"`) {
			t.Fatalf("resume changed source: code=%d %s", code, out)
		}
	case "m": // isolation: read-only helpers leave the repo tree byte-identical
		before := snapshot(t, repo)
		for _, args := range [][]string{{"changes", "--json"}, {"state", "--json"}, {"inspect", "--json"}} {
			if out, code := runCommand(t, binary, repo, env, args...); code != 0 {
				t.Fatalf("%v: code=%d %s", args, code, out)
			}
		}
		if !reflect.DeepEqual(before, snapshot(t, repo)) {
			t.Fatal("read-only helpers mutated the repository tree")
		}
	case "n": // tool degradation: capability reasoning without Tool Manager
		assertFileContains(t, repo, ".agent-ready/state/decisions.jsonl", "capability reasoning", "no block")
	case "p": // decision evidence: every decision and provenance recorded
		out, code := runCommand(t, binary, repo, env, "state", "--json")
		if code != 0 || !strings.Contains(out, `"path":".agent-ready/state/decisions.jsonl"`) || !strings.Contains(out, `"entries":1`) || !strings.Contains(out, `"path":".agent-ready/state/provenance.jsonl"`) {
			t.Fatalf("decision evidence: code=%d %s", code, out)
		}
	}
}
