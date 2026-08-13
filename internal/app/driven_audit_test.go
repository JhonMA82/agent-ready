package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
// events and newly written state. The visible oracle requires fact-helper
// use, the mandatory categorized Tool / Capability Assessment (ecosystem,
// productivity, provider) with evidence and reasons or a reasoned
// NO_ADDITIONAL_TOOLS, and the V1 corrective output sections — all over the
// OpenCode text stream only; persisted state is asserted separately and never
// satisfies the visible contract. It never prescribes model-owned verdicts.
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
	// §57: seed a global OpenCode config; init must leave it byte-identical.
	global := filepath.Join(home, ".config", "opencode", "opencode.json")
	globalBytes := []byte("{\"model\":\"acme/small\"}\n")
	if err := os.MkdirAll(filepath.Dir(global), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(global, globalBytes, 0o600); err != nil {
		t.Fatal(err)
	}

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
	if after, err := os.ReadFile(global); err != nil || !bytes.Equal(after, globalBytes) {
		t.Fatalf("init modified the global OpenCode config: %v\n%s", err, after)
	}
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
	if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
		t.Fatalf("audit did not write state decisions.jsonl: %v", err)
	}
	// The visible output contract is asserted over the OpenCode text stream
	// only; persisted state is asserted separately and never satisfies it.
	assertAuditStructure(t, events, events.text.String())
	assertPersistedAuditState(t, repo)
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

// assertAuditStructure is the structural oracle over the visible OpenCode
// text: fact-helper use, completed executions, and the mandatory categorized
// assessment with reasons, plus observable evidence of the V1 corrective
// contracts: Repository, Context Placement, Artifact Decisions, Tool /
// Capability Assessment, and Checkpoint, with an explicit RTK evaluation
// inside Productivity. Model verdicts (NO_ACTION vs candidates, scores,
// artifact choices) stay free.
func assertAuditStructure(t *testing.T, events auditEvents, visible string) {
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
	assertVisibleAuditOutput(t, visible)
}

// visibleAuditRequirements are the tokens the visible OpenCode output MUST
// demonstrate per the V1 output contract: the five mandatory sections, the
// three tool families, and an explicit RTK evaluation. Token matching is
// case-insensitive because models capitalize section headings.
var visibleAuditRequirements = []string{
	"repository",
	"context placement",
	"artifact",
	"tool",
	"checkpoint",
	"ecosystem",
	"productivity",
	"provider",
	"rtk",
}

// missingVisibleAuditTokens returns the output-contract tokens absent from
// the visible OpenCode text. Persisted state never contributes here: a run
// whose visible output omits Repository or Context Placement fails even when
// state carries that data.
func missingVisibleAuditTokens(visible string) []string {
	lower := strings.ToLower(visible)
	var missing []string
	for _, token := range visibleAuditRequirements {
		if !strings.Contains(lower, token) {
			missing = append(missing, token)
		}
	}
	return missing
}

// assertVisibleAuditOutput is the visible-only structural oracle: the
// mandatory sections and the categorized Tool / Capability Assessment with
// reasons must appear in the text OpenCode actually sent. It never reads
// persisted files. No specific RTK verdict is required.
func assertVisibleAuditOutput(t *testing.T, visible string) {
	t.Helper()
	if missing := missingVisibleAuditTokens(visible); len(missing) > 0 {
		t.Fatalf("visible audit output missing structural tokens %v", missing)
	}
	lower := strings.ToLower(visible)
	reasoned := strings.Contains(lower, "reason") || strings.Contains(lower, "because") || strings.Contains(lower, "warrant")
	if strings.Contains(lower, "no_additional_tools") && !reasoned {
		t.Fatal("NO_ADDITIONAL_TOOLS must carry a reason")
	}
	if !strings.Contains(lower, "no_additional_tools") && !reasoned {
		t.Fatal("assessment must contain recommendations with reasons or reasoned NO_ADDITIONAL_TOOLS")
	}
}

// assertStateAfterAudit applies the repository profile assertion after a
// completed audit: the model must persist the repository profile with
// kind.primary and kind.confidence before the run completes.
func assertStateAfterAudit(t *testing.T, repo string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "repository-profile.yaml"))
	if err != nil {
		t.Fatalf("repository-profile.yaml not persisted after audit: %v", err)
	}
	lower := strings.ToLower(string(data))
	for _, token := range []string{"kind", "primary", "confidence"} {
		if !strings.Contains(lower, token) {
			t.Fatalf("repository-profile.yaml missing %q", token)
		}
	}
}

// assertPersistedAuditState applies the persisted-state assertions after a
// completed audit, each layer validated on its own: repository profile,
// decisions, checkpoint, and — when the run's evidence makes them applicable
// — the Context Placement verdict and the Boilerplate Assessment. Persisted
// files never satisfy the visible output contract, which is asserted
// separately over the OpenCode text only.
func assertPersistedAuditState(t *testing.T, repo string) {
	t.Helper()
	assertStateAfterAudit(t, repo)
	if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
		t.Fatalf("audit did not persist state decisions.jsonl: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".agent-ready", "checkpoints", "latest.json")); err != nil {
		t.Fatalf("audit did not persist checkpoint latest.json: %v", err)
	}
	if hasExistingGuidance(t, repo) {
		assertContextPlacementPersisted(t, repo)
	}
	assertBoilerplatePersisted(t, repo)
}

// hasExistingGuidance reports whether the repository carries existing
// guidance (AGENTS.md or installed local skills) that a REUSE/NO_ACTION
// conclusion must evaluate before persisting a Context Placement verdict.
func hasExistingGuidance(t *testing.T, repo string) bool {
	t.Helper()
	if _, err := os.Stat(filepath.Join(repo, "AGENTS.md")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(repo, ".opencode", "skills")); err == nil {
		return true
	}
	return false
}

// assertContextPlacementPersisted requires structural Context Placement
// evidence in decisions.jsonl: a record identifiable as the context_placement
// stage/type and carrying a subject, a decision/verdict, and a reason or
// evidence. A bare {"decision":"NO_ACTION"} never satisfies it.
func assertContextPlacementPersisted(t *testing.T, repo string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl"))
	if err != nil {
		t.Fatalf("decisions.jsonl not persisted: %v", err)
	}
	if err := contextPlacementEvidence(data); err != nil {
		t.Fatal(err)
	}
}

// contextPlacementEvidence reports whether data contains a Context Placement
// record per the project's JSONL contract: identifiable by a stage/type
// (kind/category) value of context_placement and carrying a subject (subject
// or artifact), a verdict (decision, verdict, or action), and a reason or
// evidence. It returns nil only for a complete structural record.
func contextPlacementEvidence(data []byte) error {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if !recordIdentifiesPlacement(rec) {
			continue
		}
		if !recordHasField(rec, "subject", "artifact") || !recordHasField(rec, "decision", "verdict", "action") || !recordHasField(rec, "reason", "evidence") {
			return errors.New("context_placement record found but incomplete: needs subject, decision/verdict, and reason/evidence")
		}
		return nil
	}
	return errors.New("no record identified as context_placement (stage/type == context_placement) with subject, verdict, and reason/evidence in decisions.jsonl")
}

func recordIdentifiesPlacement(rec map[string]any) bool {
	for _, key := range []string{"stage", "type", "kind", "category"} {
		if v, ok := rec[key].(string); ok && strings.EqualFold(v, "context_placement") {
			return true
		}
	}
	return false
}

func recordHasField(rec map[string]any, keys ...string) bool {
	for _, key := range keys {
		if v, ok := rec[key].(string); ok && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

// boilerplateAssessmentKeys are the evaluation dimensions the Boilerplate
// Assessment must demonstrate when the repository kind is
// starter/boilerplate/template. No positive finding is required: each
// dimension may be assessed, partial, or not_found; the evaluation itself
// must be demonstrable.
var boilerplateAssessmentKeys = []string{
	"extension_points",
	"editable_boundaries",
	"generated_files",
	"feature_addition_workflow",
	"upgrade_strategy",
}

// boilerplateAssessmentGaps returns the required assessment dimensions absent
// from a repository profile. When the profile classifies kind.primary as
// starter/boilerplate/template, the structured assessment must be present.
func boilerplateAssessmentGaps(profile string) []string {
	lower := strings.ToLower(profile)
	marker := "boilerplate_assessment"
	idx := strings.Index(lower, marker)
	if idx < 0 {
		return append([]string(nil), boilerplateAssessmentKeys...)
	}
	section := lower[idx+len(marker):]
	if nl := strings.IndexByte(section, '\n'); nl >= 0 {
		section = section[nl+1:]
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		hasKey := false
		for _, key := range boilerplateAssessmentKeys {
			if strings.Contains(line, key) {
				seen[key] = true
				hasKey = true
			}
		}
		if line == trimmed && !hasKey { // next top-level key ends the section
			break
		}
	}
	var missing []string
	for _, key := range boilerplateAssessmentKeys {
		if !seen[key] {
			missing = append(missing, key)
		}
	}
	return missing
}

// profilePrimaryKind extracts the repository profile's kind.primary value.
func profilePrimaryKind(profile string) string {
	m := primaryKindRE.FindStringSubmatch(profile)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

// primaryKindRE matches the profile's `primary: <kind>` line.
var primaryKindRE = regexp.MustCompile(`(?im)^\s*primary\s*:\s*([a-z_/\-]+)`)

// kindRequiresBoilerplate reports whether a primary kind triggers the
// Boilerplate Assessment contract (starter/boilerplate/template).
func kindRequiresBoilerplate(kind string) bool {
	for _, k := range []string{"starter", "boilerplate", "template"} {
		if strings.Contains(kind, k) {
			return true
		}
	}
	return false
}

// assertBoilerplatePersisted requires the structured Boilerplate Assessment
// in persisted state when the classified kind is starter/boilerplate/template.
// The assessment may live in the repository profile (contract location) or
// the model-owned decisions record; each dimension may report not_found.
func assertBoilerplatePersisted(t *testing.T, repo string) {
	t.Helper()
	var profile string
	if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "repository-profile.yaml")); err == nil {
		profile = string(data)
	}
	if !kindRequiresBoilerplate(profilePrimaryKind(profile)) {
		return // not applicable
	}
	if missing := boilerplateAssessmentGaps(profile); len(missing) == 0 {
		return
	}
	if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err == nil {
		if missing := boilerplateAssessmentGaps(string(data)); len(missing) == 0 {
			return
		}
	}
	t.Fatalf("starter/boilerplate/template repository lacks a structured boilerplate assessment (dimensions: %v) in persisted state", boilerplateAssessmentKeys)
}

// TestVisibleAuditContractIsVisibleOnly guards the false-positive class the
// patch removes: persisted state must never satisfy the visible output
// contract. A visible response that omits Repository and Context Placement
// must fail even when state would carry that data (§10).
func TestVisibleAuditContractIsVisibleOnly(t *testing.T) {
	visible := `Outcome: NO_ACTION

Artifact Decisions:
  REUSE existing AGENTS guidance

Tool / Capability Assessment:
  ecosystem: go, node
  productivity: rg, fd, jq, RTK NOT_JUSTIFIED
  provider: none justified; evaluated Context7, Semble, Serena, CodeGraph, Headroom

Checkpoint:
  complete
`
	missing := missingVisibleAuditTokens(visible)
	if len(missing) == 0 {
		t.Fatal("visible output omitting Repository and Context Placement must fail the visible contract")
	}
	for _, absent := range []string{"repository", "context placement"} {
		found := false
		for _, m := range missing {
			if m == absent {
				found = true
			}
		}
		if !found {
			t.Fatalf("visible contract must flag %q as missing, got %v", absent, missing)
		}
	}

	complete := `Repository
  starter/template, TanStack Start + React + shadcn

Context Placement
  screen creation → REUSE

Artifact Decisions
  NO_ACTION

Tool / Capability Assessment
  ecosystem: go, node
  productivity: rg, fd, jq, RTK NOT_JUSTIFIED
  provider: none justified

Checkpoint
  complete
`
	if missing := missingVisibleAuditTokens(complete); len(missing) > 0 {
		t.Fatalf("complete visible output must satisfy the contract, missing %v", missing)
	}
}

// TestContextPlacementEvidenceStructural pins the Context Placement persisted
// evidence contract (§11): a bare {"decision":"NO_ACTION"} record never
// demonstrates Context Placement; a record identifiable as the
// context_placement stage/type with subject, verdict, and reason/evidence
// does. The project-native artifact/action shape is accepted too.
func TestContextPlacementEvidenceStructural(t *testing.T) {
	insufficient := []byte(`{"decision":"NO_ACTION"}`)
	if err := contextPlacementEvidence(insufficient); err == nil {
		t.Fatal("bare NO_ACTION decision must not satisfy Context Placement evidence")
	}

	valid := []byte(`{"stage":"context_placement","subject":"screen-creation","decision":"REUSE","reason":"existing AGENTS guidance covers screen creation"}`)
	if err := contextPlacementEvidence(valid); err != nil {
		t.Fatalf("structural context_placement record must satisfy evidence: %v", err)
	}

	native := []byte("{\"type\":\"context_placement\",\"artifact\":\"docs/ai/index.md\",\"action\":\"REUSE\",\"evidence\":\"repeated onboarding questions\"}\n")
	if err := contextPlacementEvidence(native); err != nil {
		t.Fatalf("project-native context_placement record must satisfy evidence: %v", err)
	}

	if err := contextPlacementEvidence([]byte("")); err == nil {
		t.Fatal("empty decisions.jsonl must not satisfy Context Placement evidence")
	}
}

// TestBoilerplateAssessmentStructural pins the Boilerplate Assessment
// persisted evidence contract (§12): a starter/boilerplate/template primary
// kind without a structured assessment fails; the assessment with all
// required dimensions (each allowed to report not_found) passes.
func TestBoilerplateAssessmentStructural(t *testing.T) {
	starterNoAssessment := `kind:
  primary: starter
  confidence: 0.9
`
	if kind := profilePrimaryKind(starterNoAssessment); !kindRequiresBoilerplate(kind) {
		t.Fatalf("primary kind %q must require the boilerplate assessment", kind)
	}
	if missing := boilerplateAssessmentGaps(starterNoAssessment); len(missing) == 0 {
		t.Fatal("starter profile without boilerplate_assessment must fail")
	}

	assessed := `kind:
  primary: starter
  confidence: 0.9
boilerplate_assessment:
  extension_points:
    status: found
  editable_boundaries:
    status: found
  generated_files:
    status: found
  feature_addition_workflow:
    status: covered
  upgrade_strategy:
    status: not_found
`
	if missing := boilerplateAssessmentGaps(assessed); len(missing) > 0 {
		t.Fatalf("full boilerplate assessment must pass, missing dimensions %v", missing)
	}

	application := `kind:
  primary: application
  confidence: 0.9
`
	if kind := profilePrimaryKind(application); kindRequiresBoilerplate(kind) {
		t.Fatalf("primary kind %q must not require the boilerplate assessment", kind)
	}
}
