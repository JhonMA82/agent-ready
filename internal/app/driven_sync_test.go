package app

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDrivenSync drives the installed command through real OpenCode in two
// isolated temporary Git repositories and observes the sync contract
// (audit-evidence-gates: relevant sync reassessment). The lockfile cohort
// adds a lockfile after the ChangeSet baseline and must reassess tool
// capabilities with reasons; the prose cohort edits only a prose file and
// must record why reassessment was unnecessary. The structural oracle
// observes JSONL events and newly written state; it never prescribes
// model-owned verdicts.
func TestDrivenSync(t *testing.T) {
	if testing.Short() {
		t.Skip("driven sync executes the real OpenCode runtime")
	}
	model := os.Getenv("AGENT_READY_DRIVEN_MODEL")
	if model == "" {
		t.Skip("AGENT_READY_DRIVEN_MODEL unset; driven sync proof requires a model")
	}
	if _, err := exec.LookPath("opencode"); err != nil {
		t.Skipf("opencode binary required on PATH: %v", err)
	}
	authSrc := filepath.Join(os.Getenv("HOME"), ".local", "share", "opencode", "auth.json")
	if _, err := os.Stat(authSrc); err != nil {
		t.Skipf("opencode auth %s required for the driven run: %v", authSrc, err)
	}

	base := t.TempDir()
	binDir, binary := filepath.Join(base, "bin"), filepath.Join(base, "agent-ready")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/agent-ready")
	build.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	home := filepath.Join(base, "home")
	authDir := filepath.Join(home, ".local", "share", "opencode")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatal(err)
	}
	authData, err := os.ReadFile(authSrc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth.json"), authData, 0o600); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME="+filepath.Join(home, ".local", "state"),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	for _, cohort := range []string{"lockfile", "prose"} {
		t.Run(cohort, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+cohort)
			copyFixture(t, filepath.Join("testdata", "acceptance", "driven", "sync", cohort), repo)
			if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			if got, err := gitRoot("git", repo); err != nil || got != repo {
				t.Fatalf("gitRoot(%q) = %q, %v", repo, got, err)
			}
			if out, code := runCommand(t, binary, repo, env, "init", "--json"); code != 0 {
				t.Fatalf("init: code=%d %s", code, out)
			}
			// Warm up the OpenCode runtime before recording the ChangeSet
			// baseline: OpenCode rewrites opencode.json (adding $schema)
			// the first time it loads the repository config. A baseline
			// recorded before that rewrite would report opencode.json as a
			// tool-fact change and force reassessment even in the prose
			// cohort. One invocation with an unknown command fails fast
			// without calling the model; its non-zero exit is expected and
			// the rewritten config is stable across later runs.
			warmCtx, warmCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			warmup := exec.CommandContext(warmCtx, "opencode", "run", "--dir", repo, "--model", model,
				"--format", "json", "--command", "agent-ready-config-warmup")
			warmup.Dir, warmup.Env = repo, env
			_ = warmup.Run()
			warmCancel()
			// The ChangeSet baseline is deterministic harness state (a
			// checkpoint of the inventory) created before the mutation; no
			// model conclusion is seeded.
			if out, code := runCommand(t, binary, repo, env, "checkpoint", "save", "--stage", "baseline", "--complete"); code != 0 {
				t.Fatalf("checkpoint save: code=%d %s", code, out)
			}
			baseline := snapshot(t, repo)
			mutateSyncRepo(t, repo, cohort)

			ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "opencode", "run", "--dir", repo, "--model", model,
				"--format", "json", "--command", "agent-ready", "sync")
			cmd.Dir, cmd.Env = repo, env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("opencode run: %v\n%s", err, out)
			}
			events := observeSyncJSONL(t, out)
			after := snapshot(t, repo)
			if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
				t.Fatalf("sync did not write state decisions.jsonl: %v", err)
			}
			assertSyncStructure(t, events, auditDocument(events, baseline, after), cohort)
		})
	}
}

// observeSyncJSONL is the tolerant variant of observeAuditJSONL for sync
// runs: the OpenCode runtime may interleave non-JSON noise lines (e.g.
// ANSI-colored permission notices) with the event stream, so malformed
// lines are skipped instead of failing the run. Every parseable event is
// still observed exactly as in the audit observer.
func observeSyncJSONL(t *testing.T, data []byte) auditEvents {
	t.Helper()
	var events auditEvents
	events.helpers = map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e struct {
			Type string `json:"type"`
			Part struct {
				Tool  string `json:"tool"`
				State *struct {
					Status string          `json:"status"`
					Input  json.RawMessage `json:"input"`
				} `json:"state"`
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"part"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // runtime noise line (permission notice, banner); not an event
		}
		switch {
		case e.Type == "error" && e.Error != nil:
			if events.failed == "" {
				events.failed = e.Error.Message
			}
		case e.Type == "tool_use":
			if e.Part.State == nil {
				continue
			}
			if e.Part.State.Status == "completed" {
				events.executed++
			}
			var input struct {
				Command string `json:"command"`
			}
			if json.Unmarshal(e.Part.State.Input, &input) == nil {
				if m := auditHelperRE.FindStringSubmatch(input.Command); m != nil {
					events.helpers[m[2]] = true
				}
			}
		case e.Part.Type == "text" && e.Part.Text != "":
			events.text.WriteString(e.Part.Text)
			events.text.WriteString("\n")
		}
	}
	return events
}

// mutateSyncRepo applies the post-baseline repository mutation for one sync
// cohort: the lockfile cohort gains a lockfile (relevant evidence), the
// prose cohort gains a prose-only edit (irrelevant evidence).
func mutateSyncRepo(t *testing.T, repo, cohort string) {
	t.Helper()
	switch cohort {
	case "lockfile":
		lock := `{
  "name": "driven-sync",
  "lockfileVersion": 3,
  "requires": true,
  "packages": {
    "": {"name": "driven-sync", "dependencies": {"chalk": "^5.3.0"}},
    "node_modules/chalk": {"version": "5.3.0", "resolved": "https://registry.npmjs.org/chalk/-/chalk-5.3.0.tgz", "license": "MIT"}
  }
}
`
		if err := os.WriteFile(filepath.Join(repo, "package-lock.json"), []byte(lock), 0o644); err != nil {
			t.Fatal(err)
		}
	case "prose":
		f, err := os.OpenFile(filepath.Join(repo, "README.md"), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if _, err := f.WriteString("More prose after the baseline.\n"); err != nil {
			t.Fatal(err)
		}
	}
}

// assertSyncStructure is the structural oracle for the sync contract: the
// lockfile cohort must reassess tool capabilities with reasons and complete
// with categorized recommendations or reasoned NO_ADDITIONAL_TOOLS; the
// prose cohort must record why reassessment was unnecessary. Model verdicts
// (NO_ACTION vs updates, artifact choices) stay free. Token matching is
// case-insensitive: models capitalize headings.
func assertSyncStructure(t *testing.T, events auditEvents, doc, cohort string) {
	t.Helper()
	if events.failed != "" {
		t.Fatalf("sync failed: %s", events.failed)
	}
	if events.executed == 0 {
		t.Fatal("no completed tool executions observed")
	}
	if len(events.helpers) == 0 {
		t.Fatal("no agent-ready fact-helper use observed in JSONL events")
	}
	lower := strings.ToLower(doc)
	reasoned := strings.Contains(lower, "reason") || strings.Contains(lower, "because") || strings.Contains(lower, "warrant") || strings.Contains(lower, "justification")
	switch cohort {
	case "lockfile":
		if !strings.Contains(lower, "lockfile") && !strings.Contains(lower, "package-lock") {
			t.Fatal("lockfile sync must cite the changed lockfile evidence")
		}
		if !strings.Contains(lower, "reassess") && !strings.Contains(lower, "re-evaluat") && !strings.Contains(lower, "re-assess") {
			t.Fatal("lockfile sync must reassess tool capabilities")
		}
		if !reasoned {
			t.Fatal("reassessment must carry reasons")
		}
		categorized := strings.Contains(lower, "ecosystem") || strings.Contains(lower, "productivity") || strings.Contains(lower, "provider")
		if !strings.Contains(lower, "no_additional_tools") && !categorized {
			t.Fatal("completed reassessment must include categorized recommendations or NO_ADDITIONAL_TOOLS")
		}
	case "prose":
		skipped := strings.Contains(lower, "skip") || strings.Contains(lower, "unnecessar") ||
			strings.Contains(lower, "not affect") || strings.Contains(lower, "does not affect") ||
			strings.Contains(lower, "no reassess") || strings.Contains(lower, "irrelevant") || strings.Contains(lower, "no tool")
		if !skipped {
			t.Fatal("prose sync must record why reassessment was unnecessary")
		}
		if !reasoned {
			t.Fatal("recorded skip must carry a reason")
		}
	}
}
