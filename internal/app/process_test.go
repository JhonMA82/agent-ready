package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JhonMA82/agent-ready/internal/opencode"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/safeio"
)

func TestProcessAcceptance(t *testing.T) {
	if mode := os.Getenv("AGENT_READY_TEST_HELPER"); mode != "" {
		runHelper(mode)
		return
	}
	if testing.Short() {
		t.Skip("process acceptance builds and executes the command")
	}
	base, home, xdg := t.TempDir(), t.TempDir(), t.TempDir()
	canaries := map[string][]byte{filepath.Join(home, ".config/opencode/canary"): []byte("home"), filepath.Join(xdg, "opencode/canary"): []byte("xdg")}
	for path, data := range canaries {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	homeBefore, xdgBefore := snapshot(t, home), snapshot(t, xdg)
	binDir, binary := filepath.Join(base, "bin"), filepath.Join(base, "agent-ready")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' '"+opencode.RequiredVersion()+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/agent-ready")
	build.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	env := append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+xdg, "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	repo := newRepo(t, base, "success")
	nested := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, repo)
	dry := runJSON(t, binary, nested, env, 0, "init", "--dry-run", "--json")
	if dry.Outcome != plan.DryRun || dry.Root != repo || dry.Invocation != nested || !reflect.DeepEqual(before, snapshot(t, repo)) {
		t.Fatalf("dry run=%+v", dry)
	}
	out, code := runCommand(t, binary, nested, env, "init")
	if code != 0 || !strings.Contains(out, "Outcome: changed") || !strings.Contains(out, "Invocation: "+nested) || !strings.Contains(out, "Next: /agent-ready") {
		t.Fatalf("success code=%d: %s", code, out)
	}
	command, err := os.ReadFile(filepath.Join(repo, ".opencode/commands/agent-ready.md"))
	if err != nil || !bytes.Contains(command, []byte("$ARGUMENTS")) {
		t.Fatalf("command: %v %q", err, command)
	}
	for _, forbidden := range []string{"model:", "agent:", "subtask:"} {
		if bytes.Contains(command, []byte(forbidden)) {
			t.Fatalf("command contains %s", forbidden)
		}
	}
	installed := snapshot(t, repo)
	rerun := runJSON(t, binary, repo, env, 0, "init", "--json")
	if rerun.Outcome != plan.Noop || rerun.NextStep != "/agent-ready" || !reflect.DeepEqual(installed, snapshot(t, repo)) {
		t.Fatalf("rerun=%+v", rerun)
	}

	conflict := newRepo(t, base, "conflict")
	path := filepath.Join(conflict, ".opencode/commands/agent-ready.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("unowned"), 0o644); err != nil {
		t.Fatal(err)
	}
	conflictBefore := snapshot(t, conflict)
	refused := runJSON(t, binary, conflict, env, 3, "init", "--json")
	if refused.Outcome != plan.Refused || !reflect.DeepEqual(conflictBefore, snapshot(t, conflict)) {
		t.Fatalf("refusal=%+v", refused)
	}
	humanRefused, humanCode := runCommand(t, binary, conflict, env, "init")
	if humanCode != 3 || !strings.Contains(humanRefused, "Refused (plan):") || !strings.Contains(humanRefused, "Remediation:") {
		t.Fatalf("human refusal code=%d: %s", humanCode, humanRefused)
	}

	failed := newRepo(t, base, "failed")
	failedBefore := snapshot(t, failed)
	commitFailed := runJSON(t, os.Args[0], failed, append(env, "AGENT_READY_TEST_HELPER=commit"), 4, "-test.run=TestProcessAcceptance")
	if commitFailed.Outcome != plan.CommitFailed || len(commitFailed.Actions) == 0 || !reflect.DeepEqual(failedBefore, snapshot(t, failed)) {
		t.Fatalf("commit failure=%+v", commitFailed)
	}
	recovery := newRepo(t, base, "recovery")
	required := runJSON(t, os.Args[0], recovery, append(env, "AGENT_READY_TEST_HELPER=rollback"), 5, "-test.run=TestProcessAcceptance")
	if required.Outcome != plan.RecoveryRequired {
		t.Fatalf("recovery required=%+v", required)
	}
	recovered := runJSON(t, binary, recovery, env, 0, "init", "--json")
	if recovered.Outcome != plan.Changed {
		t.Fatalf("recovered=%+v", recovered)
	}

	invalid := newRepo(t, base, "invalid")
	journal := filepath.Join(invalid, ".agent-ready/transaction.json")
	if err := os.MkdirAll(filepath.Dir(journal), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(journal, []byte(`{"schema":"agent-ready.transaction/v1","entries":[{"path":"../outside","applied":true}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidBefore := snapshot(t, invalid)
	bad := runJSON(t, binary, invalid, env, 5, "init", "--json")
	if bad.Outcome != plan.RecoveryRequired || !reflect.DeepEqual(invalidBefore, snapshot(t, invalid)) {
		t.Fatalf("invalid recovery=%+v", bad)
	}
	for path, want := range canaries {
		if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, want) {
			t.Fatalf("global canary %s: %q %v", path, got, err)
		}
	}
	if got := snapshot(t, home); !reflect.DeepEqual(homeBefore, got) {
		t.Fatalf("HOME tree changed after init outcomes: %v", got)
	}
	if got := snapshot(t, xdg); !reflect.DeepEqual(xdgBefore, got) {
		t.Fatalf("XDG tree changed after init outcomes: %v", got)
	}
}

func runHelper(mode string) {
	commits := 0
	r := initWithOptions(context.Background(), false, safeio.Options{Hook: func(phase, _ string) error {
		if phase == "commit" {
			commits++
			if commits == 2 {
				return errors.New("injected commit failure")
			}
		}
		if mode == "rollback" && phase == "rollback" {
			return errors.New("injected rollback failure")
		}
		return nil
	}})
	_ = json.NewEncoder(os.Stdout).Encode(r)
	os.Exit(plan.ExitCode(r))
}

func newRepo(t *testing.T, base, name string) string {
	t.Helper()
	root := filepath.Join(base, name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return root
}

func runJSON(t *testing.T, command, dir string, env []string, wantCode int, args ...string) plan.Result {
	t.Helper()
	out, code := runCommand(t, command, dir, env, args...)
	if code != wantCode {
		t.Fatalf("code=%d want=%d: %s", code, wantCode, out)
	}
	var result plan.Result
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("JSON: %v: %s", err, out)
	}
	return result
}

func runCommand(t *testing.T, command, dir string, env []string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(command, args...)
	cmd.Dir, cmd.Env = dir, env
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	t.Fatalf("run: %v: %s", err, out)
	return "", -1
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out[rel] = string(data) + "\x00" + info.Mode().String()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
