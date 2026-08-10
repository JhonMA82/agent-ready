package app

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// gitRoot resolves the containing Git worktree root for dir. It accepts
// relative and absolute paths and fails closed when Git is missing or the
// path is not inside a worktree.
func gitRoot(gitBin, dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	out, err := exec.Command(gitBin, "-C", abs, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", err
	}
	root := strings.TrimSuffix(string(out), "\n")
	if !filepath.IsAbs(root) {
		return "", os.ErrInvalid
	}
	return filepath.Clean(root), nil
}

// TestDrivenAuditGitSelectors proves the driven harness accepts relative and
// absolute temporary Git roots and fails closed outside a worktree or when
// the Git binary is missing (design threat matrix: Git repository selection).
func TestDrivenAuditGitSelectors(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	cases := []struct {
		name   string
		git    string
		dir    string
		want   string
		broken bool
	}{
		{name: "absolute temp root", git: "git", dir: root, want: root},
		{name: "relative temp root", git: "git", dir: rel, want: root},
		{name: "nested inside root", git: "git", dir: nested, want: root},
		{name: "outside root fails closed", git: "git", dir: outside, broken: true},
		{name: "missing Git fails closed", git: "git-definitely-missing", dir: root, broken: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gitRoot(tc.git, tc.dir)
			if tc.broken {
				if err == nil {
					t.Fatalf("expected failure, got root %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("root = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestDrivenAudit drives the installed command through real OpenCode in an
// isolated temporary Git repository and observes the audit through JSONL
// events and newly written state. The structural oracle requires fact-helper
// use and the mandatory categorized Tool / Capability Assessment (ecosystem,
// productivity, provider) with evidence and reasons, or a reasoned
// NO_ADDITIONAL_TOOLS. It never prescribes model-owned verdicts.
func TestDrivenAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("driven audit executes the real OpenCode runtime")
	}
	model := os.Getenv("AGENT_READY_DRIVEN_MODEL")
	if model == "" {
		t.Skip("AGENT_READY_DRIVEN_MODEL unset; driven audit proof requires a model")
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

	repo := filepath.Join(base, "repo")
	copyFixture(t, filepath.Join("testdata", "acceptance", "driven", "audit"), repo)
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if got, err := gitRoot("git", repo); err != nil || got != repo {
		t.Fatalf("gitRoot(%q) = %q, %v", repo, got, err)
	}
	if out, code := runCommand(t, binary, repo, env, "init", "--json"); code != 0 {
		t.Fatalf("init: code=%d %s", code, out)
	}
	baseline := snapshot(t, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "opencode", "run", "--dir", repo, "--model", model,
		"--format", "json", "--command", "agent-ready", "audit")
	cmd.Dir, cmd.Env = repo, env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("opencode run: %v\n%s", err, out)
	}
	events := observeAuditJSONL(t, out)
	after := snapshot(t, repo)
	if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
		t.Fatalf("audit did not write state decisions.jsonl: %v", err)
	}
	assertAuditStructure(t, events, auditDocument(events, baseline, after))
}

// auditEvents is the structural view of the OpenCode JSONL event stream.
type auditEvents struct {
	text     strings.Builder
	helpers  map[string]bool
	executed int
	failed   string
}

var auditHelperRE = regexp.MustCompile(`agent-ready\s+(tools\s+)?(inspect|state|changes|checkpoint|doctor|recommend|status|init)\b`)

func observeAuditJSONL(t *testing.T, data []byte) auditEvents {
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
			t.Fatalf("invalid JSONL event: %v: %s", err, line)
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

// auditDocument joins assistant text with every newly written or changed
// repository file so the oracle observes the audit's produced output without
// prescribing its verdicts. OpenCode's own runtime byproduct under
// .opencode/node_modules is excluded: it is not audit output.
func auditDocument(events auditEvents, baseline, after map[string]string) string {
	doc := events.text.String()
	var paths []string
	for path := range after {
		if strings.HasPrefix(path, ".git/") || strings.HasPrefix(path, ".opencode/node_modules/") {
			continue
		}
		if before, ok := baseline[path]; !ok || before != after[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		doc += "\n" + after[path]
	}
	return doc
}

// assertAuditStructure is the structural oracle: fact-helper use, completed
// executions, and the mandatory categorized assessment with reasons. Model
// verdicts (NO_ACTION vs candidates, scores, artifact choices) stay free.
// Token matching is case-insensitive: models capitalize section headings.
func assertAuditStructure(t *testing.T, events auditEvents, doc string) {
	t.Helper()
	if events.failed != "" {
		t.Fatalf("audit failed: %s", events.failed)
	}
	if events.executed == 0 {
		t.Fatal("no completed tool executions observed")
	}
	if len(events.helpers) == 0 {
		t.Fatal("no agent-ready fact-helper use observed in JSONL events")
	}
	lower := strings.ToLower(doc)
	for _, family := range []string{"ecosystem", "productivity", "provider"} {
		if !strings.Contains(lower, family) {
			t.Fatalf("Tool / Capability Assessment missing family %q", family)
		}
	}
	reasoned := strings.Contains(lower, "reason") || strings.Contains(lower, "because") || strings.Contains(lower, "warrant")
	if strings.Contains(lower, "no_additional_tools") && !reasoned {
		t.Fatal("NO_ADDITIONAL_TOOLS must carry a reason")
	}
	if !strings.Contains(lower, "no_additional_tools") && !reasoned {
		t.Fatal("assessment must contain recommendations with reasons or reasoned NO_ADDITIONAL_TOOLS")
	}
}
