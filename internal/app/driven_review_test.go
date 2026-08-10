package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestDrivenReview drives the installed command through real OpenCode in two
// isolated temporary Git repositories and observes the review contract
// (audit-evidence-gates: External Verification Gate and reviewer rejection
// contract). The grounded cohort reviews a candidate skill whose
// version-sensitive toolchain knowledge cites current versioned evidence
// (OpenCode 1.18.15 official documentation) and must be accepted; the
// ungrounded cohort reviews a candidate skill that embeds framework claims
// without versioned evidence or exemption and must be rejected (or the
// exemption explicitly handled). The structural oracle observes JSONL events
// and newly written state; verdict wording and score identity stay free.
func TestDrivenReview(t *testing.T) {
	if testing.Short() {
		t.Skip("driven review executes the real OpenCode runtime")
	}
	model := os.Getenv("AGENT_READY_DRIVEN_MODEL")
	if model == "" {
		t.Skip("AGENT_READY_DRIVEN_MODEL unset; driven review proof requires a model")
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

	for _, cohort := range []string{"grounded", "ungrounded"} {
		t.Run(cohort, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+cohort)
			copyFixture(t, filepath.Join("testdata", "acceptance", "driven", "review", cohort), repo)
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
				"--format", "json", "--command", "agent-ready", "review")
			cmd.Dir, cmd.Env = repo, env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("opencode run: %v\n%s", err, out)
			}
			events := observeSyncJSONL(t, out)
			after := snapshot(t, repo)
			if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
				t.Fatalf("review did not write state decisions.jsonl: %v", err)
			}
			assertReviewStructure(t, events, auditDocument(events, baseline, after), cohort)
		})
	}
}

// reviewAcceptRE matches acceptance language as standalone words. It is used
// only inside the verdict window or as a whole-document fallback, never as a
// bare substring: "not accepted" must not count as acceptance.
var reviewAcceptRE = regexp.MustCompile(`\b(pass(es|ed)?|accepted|accepts?|approved?|compliant|meets the gate)\b`)

// reviewRejectRE matches rejection language. Bare "reject" is excluded from
// the whole-document variant because the loaded rejection contract quotes
// "Reject artifacts with ..."; only "reject" followed by a score marker
// (parenthesis, dash, or colon) is unambiguous rejection.
var reviewRejectRE = regexp.MustCompile(`\b(rejected|revise|revised|does not pass|not accepted|fails the gate|gate failure|ungrounded)\b|reject\s*[(:—-]`)

// reviewVerdict classifies the outcome near the word "verdict" when present.
// Both observed run shapes put the outcome on the verdict line ("Verdict:
// PASS (85/100)", "Verdict: REJECT — 26/100"), and contract quotes never
// contain the word "verdict".
func reviewVerdict(lower string) string {
	idx := strings.Index(lower, "verdict")
	if idx == -1 {
		return ""
	}
	end := idx + 120
	if end > len(lower) {
		end = len(lower)
	}
	win := lower[idx:end]
	switch {
	case reviewAcceptRE.MatchString(win):
		return "accept"
	case reviewRejectRE.MatchString(win):
		return "reject"
	}
	return ""
}

// assertReviewStructure is the structural oracle for the review contract.
// The grounded cohort must show versioned-evidence acceptance through the
// External Verification Gate; the ungrounded cohort must show gate
// failure/exemption handling. Model verdict wording and score identity stay
// free. Token matching is case-insensitive: models capitalize headings.
func assertReviewStructure(t *testing.T, events auditEvents, doc, cohort string) {
	t.Helper()
	if events.failed != "" {
		t.Fatalf("review failed: %s", events.failed)
	}
	if events.executed == 0 {
		t.Fatal("no completed tool executions observed")
	}
	if len(events.helpers) == 0 {
		t.Fatal("no agent-ready fact-helper use observed in JSONL events")
	}
	lower := strings.ToLower(doc)
	reasoned := strings.Contains(lower, "reason") || strings.Contains(lower, "because") || strings.Contains(lower, "warrant") || strings.Contains(lower, "justification")
	if !reasoned {
		t.Fatal("review verdict must carry a reason")
	}
	versioned := strings.Contains(lower, "versioned evidence") || strings.Contains(lower, "official") || strings.Contains(lower, "documentation") || strings.Contains(lower, "1.18.15") || strings.Contains(lower, "version-sensitive")
	switch cohort {
	case "grounded":
		verdict := reviewVerdict(lower)
		accepted := reviewAcceptRE.MatchString(lower)
		rejected := reviewRejectRE.MatchString(lower)
		if verdict == "reject" || (verdict == "" && rejected) || !accepted {
			t.Fatal("grounded review must accept the versioned-evidence candidate")
		}
		if !versioned {
			t.Fatal("grounded review must observe versioned-evidence acceptance")
		}
	case "ungrounded":
		verdict := reviewVerdict(lower)
		rejected := reviewRejectRE.MatchString(lower)
		exempted := strings.Contains(lower, "exemption") && strings.Contains(lower, "no external claim")
		if verdict == "accept" || (verdict == "" && !rejected && !exempted) {
			t.Fatal("ungrounded review must reject the gate failure or handle the exemption")
		}
	}
}
