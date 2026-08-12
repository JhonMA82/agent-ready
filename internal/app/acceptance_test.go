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

// TestAcceptanceFixturesQThroughU drives the context-placement regression
// fixtures Q–U (refinement §37–41): boilerplate kind recognition, AGENTS.md
// pressure signals, deterministic-script and external-skill decisions. JSON
// facts are asserted by content, never by counts alone (R16).
func TestAcceptanceFixturesQThroughU(t *testing.T) {
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
	for _, letter := range []string{"q", "r", "s", "t", "u"} {
		t.Run("fixture-"+letter, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+letter)
			copyFixture(t, filepath.Join(fixtures, "fixture-"+letter), repo)
			if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			runPlacementFixtureCase(t, binary, repo, env, letter)
		})
	}
}

// inspectFacts is the subset of agent-ready.inspect/v1 the placement
// fixtures assert on (deps, scripts, file scan, AGENTS.md fact).
type inspectFacts struct {
	Deps []struct {
		Name string `json:"name"`
	} `json:"deps"`
	Scripts []struct {
		Name string `json:"name"`
	} `json:"scripts"`
	Files struct {
		ByExtension map[string]int `json:"by_extension"`
	} `json:"files"`
	AgentsMD *struct {
		Path  string `json:"path"`
		Lines int    `json:"lines"`
	} `json:"agents_md"`
}

type candidateFacts struct {
	Capability string `json:"capability"`
	Candidate  string `json:"candidate"`
	Signal     string `json:"signal"`
	Observed   string `json:"observed"`
	CatalogID  string `json:"catalog_id"`
}

// recommendFacts is the subset of agent-ready.recommend/v1 the placement
// fixtures assert on.
type recommendFacts struct {
	Candidates []candidateFacts `json:"candidates"`
}

func runInspectJSON(t *testing.T, binary, repo string, env []string) inspectFacts {
	t.Helper()
	out, code := runCommand(t, binary, repo, env, "inspect", "--json")
	if code != 0 {
		t.Fatalf("inspect: code=%d %s", code, out)
	}
	var facts inspectFacts
	if err := json.Unmarshal([]byte(out), &facts); err != nil {
		t.Fatalf("inspect JSON: %v: %s", err, out)
	}
	return facts
}

func runRecommendJSON(t *testing.T, binary, repo string, env []string) recommendFacts {
	t.Helper()
	out, code := runCommand(t, binary, repo, env, "tools", "recommend", "--json")
	if code != 0 {
		t.Fatalf("recommend: code=%d %s", code, out)
	}
	var facts recommendFacts
	if err := json.Unmarshal([]byte(out), &facts); err != nil {
		t.Fatalf("recommend JSON: %v: %s", err, out)
	}
	return facts
}

func findCandidate(facts recommendFacts, name string) (candidateFacts, bool) {
	for _, c := range facts.Candidates {
		if c.Candidate == name {
			return c, true
		}
	}
	return candidateFacts{}, false
}

func hasDep(facts inspectFacts, name string) bool {
	for _, d := range facts.Deps {
		if d.Name == name {
			return true
		}
	}
	return false
}

func hasScript(facts inspectFacts, name string) bool {
	for _, s := range facts.Scripts {
		if s.Name == name {
			return true
		}
	}
	return false
}

func runPlacementFixtureCase(t *testing.T, binary, repo string, env []string, letter string) {
	t.Helper()
	switch letter {
	case "q": // tanstack-shadcn boilerplate: kind, placement analysis, RTK
		ins := runInspectJSON(t, binary, repo, env)
		if !hasDep(ins, "@tanstack/react-router") || !hasDep(ins, "react") {
			t.Fatalf("boilerplate deps must include tanstack/react packages: %+v", ins.Deps)
		}
		if !hasScript(ins, "build") || !hasScript(ins, "generate-routes") {
			t.Fatalf("boilerplate scripts must include build and generate-routes: %+v", ins.Scripts)
		}
		if ins.AgentsMD == nil || ins.AgentsMD.Lines >= 300 {
			t.Fatalf("boilerplate AGENTS.md must be present and under 300 lines: %+v", ins.AgentsMD)
		}
		// components.json is a seed file, not a heavy-tree presence entry; the
		// scanned-file fact proves the audit covers it.
		if ins.Files.ByExtension["json"] < 1 {
			t.Fatalf("components.json must be scanned as a json file: %+v", ins.Files.ByExtension)
		}
		rec := runRecommendJSON(t, binary, repo, env)
		rtk, ok := findCandidate(rec, "RTK")
		if !ok || !strings.Contains(rtk.Observed, "scripts=build") {
			t.Fatalf("boilerplate must fire RTK on build scripts: %+v", rec.Candidates)
		}
		for _, name := range []string{"context-placement", "ast-grep"} {
			if _, ok := findCandidate(rec, name); ok {
				t.Fatalf("boilerplate must not fire %s: %+v", name, rec.Candidates)
			}
		}
		assertFileContains(t, repo, "expect/repository-kind.md", "boilerplate", "confidence")
		assertFileContains(t, repo, "expect/placement-screen-creation.md", "REUSE", "EXTRACT_TO_SKILL", "COMPACT")
		assertFileContains(t, repo, "expect/external-skill-reuse.md", "REUSE_EXTERNAL_SKILL")
	case "r": // long AGENTS: pressure signal + four placement outcomes
		ins := runInspectJSON(t, binary, repo, env)
		if ins.AgentsMD == nil || ins.AgentsMD.Lines < 500 {
			t.Fatalf("long AGENTS.md must report 500+ lines: %+v", ins.AgentsMD)
		}
		rec := runRecommendJSON(t, binary, repo, env)
		cp, ok := findCandidate(rec, "context-placement")
		if !ok || cp.Capability != "context_placement_pressure" || cp.CatalogID != "" || !strings.Contains(cp.Observed, "AGENTS.md:") {
			t.Fatalf("long AGENTS.md must fire context-placement without catalog id: %+v", rec.Candidates)
		}
		if _, ok := findCandidate(rec, "ast-grep"); ok {
			t.Fatalf("long AGENTS fixture must not fire ast-grep: %+v", rec.Candidates)
		}
		assertFileContains(t, repo, "expect/placement-analysis.md", "COMPACT", "EXTRACT migration workflow", "EXTRACT release workflow", "MOVE examples", "router")
	case "s": // short optimal AGENTS: REUSE + NO_ACTION, no pressure signal
		ins := runInspectJSON(t, binary, repo, env)
		if ins.AgentsMD == nil || ins.AgentsMD.Lines > 70 {
			t.Fatalf("short AGENTS.md must be <= 70 lines: %+v", ins.AgentsMD)
		}
		rec := runRecommendJSON(t, binary, repo, env)
		for _, name := range []string{"context-placement", "ast-grep"} {
			if _, ok := findCandidate(rec, name); ok {
				t.Fatalf("short AGENTS fixture must not fire %s: %+v", name, rec.Candidates)
			}
		}
		assertFileContains(t, repo, "expect/placement-reuse.md", "REUSE", "NO_ACTION")
	case "t": // deterministic procedure replaced by the existing scripts
		ins := runInspectJSON(t, binary, repo, env)
		if !hasScript(ins, "generate-routes") || !hasScript(ins, "validate-presets") {
			t.Fatalf("deterministic fixture scripts must include generate-routes and validate-presets: %+v", ins.Scripts)
		}
		rec := runRecommendJSON(t, binary, repo, env)
		for _, name := range []string{"context-placement", "RTK"} {
			if _, ok := findCandidate(rec, name); ok {
				t.Fatalf("deterministic fixture must not fire %s: %+v", name, rec.Candidates)
			}
		}
		assertFileContains(t, repo, "expect/decision-replace-with-script.md", "REPLACE_WITH_SCRIPT", "COMPACT")
	case "u": // external canonical skill reused, no wrapper
		ins := runInspectJSON(t, binary, repo, env)
		if !hasDep(ins, "class-variance-authority") || !hasScript(ins, "dev") {
			t.Fatalf("external-skill fixture deps/scripts facts missing: deps=%+v scripts=%+v", ins.Deps, ins.Scripts)
		}
		rec := runRecommendJSON(t, binary, repo, env)
		for _, name := range []string{"context-placement", "RTK"} {
			if _, ok := findCandidate(rec, name); ok {
				t.Fatalf("external-skill fixture must not fire %s: %+v", name, rec.Candidates)
			}
		}
		assertFileContains(t, repo, "expect/decision-reuse-external-skill.md", "REUSE_EXTERNAL_SKILL")
	}
}
