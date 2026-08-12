package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, data string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixtureRepo exercises every fact family: gomod/npm manifests, scripts,
// workspaces, CI config, nested files, ext-less/uppercase extensions, .git.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, "go.mod", "module example.com/demo\n\ngo 1.24\n\nrequire (\n\tgithub.com/spf13/cobra v1.9.1\n\tgithub.com/tailscale/hujson v0.0.0-20260718110524-10d7940d4c87 // indirect\n)\n\nrequire github.com/sirupsen/logrus v1.9.3\n")
	write(t, root, "package.json", `{"name":"demo","dependencies":{"lodash":"4.17.21"},"devDependencies":{"jest":"29.0.0"},"scripts":{"build":"tsc","test":"jest --ci"},"workspaces":["packages/*","libs/a"]}`)
	write(t, root, ".github/workflows/ci.yml", "name: ci\n")
	write(t, root, "cmd/main.go", "package main\n")
	write(t, root, "README.md", "# demo\n")
	write(t, root, "LICENSE", "MIT\n")
	write(t, root, "assets/logo.PNG", "x")
	write(t, root, ".git/HEAD", "ref: refs/heads/main\n")
	return root
}

func TestInspectFacts(t *testing.T) {
	root := fixtureRepo(t)
	facts, err := Inspect(root, filepath.Join(root, "nested"))
	if err != nil {
		t.Fatal(err)
	}
	if facts.SchemaVersion != SchemaVersion || facts.Root != root || facts.Invocation != filepath.Join(root, "nested") {
		t.Fatalf("header mismatch: %+v", facts)
	}
	if got, want := fmt.Sprint(facts.Deps), fmt.Sprint([]Dep{{"github.com/sirupsen/logrus", "v1.9.3", "gomod"}, {"github.com/spf13/cobra", "v1.9.1", "gomod"}, {"github.com/tailscale/hujson", "v0.0.0-20260718110524-10d7940d4c87", "gomod"}, {"jest", "29.0.0", "npm"}, {"lodash", "4.17.21", "npm"}}); got != want {
		t.Fatalf("deps:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(facts.Scripts), fmt.Sprint([]Script{{"build", "tsc"}, {"test", "jest --ci"}}); got != want {
		t.Fatalf("scripts:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(facts.Workspaces), fmt.Sprint([]string{"libs/a", "packages/*"}); got != want {
		t.Fatalf("workspaces:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(facts.Files), fmt.Sprint(Files{Total: 7, ByExtension: map[string]int{"go": 1, "json": 1, "md": 1, "mod": 1, "png": 1, "yml": 1}}); got != want {
		t.Fatalf("files:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(facts.CI), fmt.Sprint(CI{Present: true, Files: []string{".github/workflows/ci.yml"}}); got != want {
		t.Fatalf("ci:\ngot  %s\nwant %s", got, want)
	}
	if got := facts.Summary(); !strings.Contains(got, "Files: 7") || !strings.Contains(got, "CI: present (.github/workflows/ci.yml)") {
		t.Fatalf("summary mismatch:\n%s", got)
	}
}

func TestInspectEdgeCases(t *testing.T) {
	root := fixtureRepo(t)
	facts, err := Inspect(root, root)
	if err != nil || facts.Invocation != "" {
		t.Fatalf("invocation must be omitted when equal to root: %q, %v", facts.Invocation, err)
	}
	empty := t.TempDir()
	write(t, empty, "hello.txt", "hi")
	write(t, empty, ".git/objects/aa/aaaa", "binary")
	if err := os.Symlink("hello.txt", filepath.Join(empty, "link.txt")); err != nil {
		t.Fatal(err)
	}
	facts, err = Inspect(empty, "")
	if err != nil || facts.Files.Total != 1 || facts.Files.ByExtension["txt"] != 1 || len(facts.Deps) != 0 || facts.CI.Present {
		t.Fatalf("unexpected edge facts (git/symlink must be excluded): %+v, %v", facts, err)
	}
	bad := t.TempDir()
	write(t, bad, "package.json", "{not json")
	if _, err := Inspect(bad, ""); err == nil {
		t.Fatal("expected error for malformed package.json")
	}
}

func TestInspectDeterministic(t *testing.T) {
	root := fixtureRepo(t)
	a, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(a)
	if jb, _ := json.Marshal(b); string(ja) != string(jb) || !strings.Contains(string(ja), `"schema_version":"agent-ready.inspect/v1"`) {
		t.Fatalf("non-deterministic or missing schema_version:\n%s\n%s", ja, jb)
	}
	if strings.Contains(string(ja), `"presence"`) {
		t.Fatalf("legacy JSON unexpectedly includes empty presence facts: %s", ja)
	}
}

func TestInspectPrunesHeavyTreesAndRetainsPresence(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	for _, dir := range []string{"vendor", "target", "node_modules", "obj", ".venv", "bin"} {
		write(t, root, dir+"/nested/ignored.js", "ignored\n")
	}

	wantPresence := []Presence{
		{Path: ".venv", Kind: "directory"},
		{Path: "bin", Kind: "directory"},
		{Path: "node_modules", Kind: "directory"},
		{Path: "obj", Kind: "directory"},
		{Path: "target", Kind: "directory"},
		{Path: "vendor", Kind: "directory"},
	}
	before, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprint(before.Presence), fmt.Sprint(wantPresence); got != want {
		t.Fatalf("presence:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(before.Files), fmt.Sprint(Files{Total: 1, ByExtension: map[string]int{"go": 1}}); got != want {
		t.Fatalf("files include heavy-tree descendants:\ngot  %s\nwant %s", got, want)
	}
	paths, err := Paths(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprint(paths), fmt.Sprint([]string{"main.go"}); got != want {
		t.Fatalf("paths include heavy-tree descendants:\ngot  %s\nwant %s", got, want)
	}

	write(t, root, "node_modules/changed/.github/workflows/ci.yml", "name: ignored\n")
	after, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatalf("heavy-tree descendant changed facts:\nbefore %s\nafter  %s", beforeJSON, afterJSON)
	}
}

func TestHeavyTreePresenceOnly(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	write(t, root, "src/app.go", "package src\n")
	dirs := []string{
		".dart_tool", ".next", ".nuxt", ".venv", "__pycache__", "_build", "bin",
		"build", "coverage", "deps", "dist", "node_modules", "obj", "out",
		"result", "storage/logs", "target", "vendor", "venv",
	}
	for _, dir := range dirs {
		write(t, root, dir+"/nested/deep/ignored.txt", "ignored\n")
	}
	write(t, root, "cmake-build-debug/CMakeCache.txt", "cache\n")
	// OQ-3: PHP bin/console stays pruned, but bin presence signal is retained.
	write(t, root, "bin/console", "#!/usr/bin/env php\n")

	before, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	wantDirs := append([]string(nil), dirs...)
	wantDirs = append(wantDirs, "cmake-build-debug")
	sort.Strings(wantDirs)
	wantPresence := make([]Presence, len(wantDirs))
	for i, dir := range wantDirs {
		wantPresence[i] = Presence{Path: dir, Kind: "directory"}
	}
	if got, want := fmt.Sprint(before.Presence), fmt.Sprint(wantPresence); got != want {
		t.Fatalf("presence:\ngot  %s\nwant %s", got, want)
	}
	if got, want := fmt.Sprint(before.Files), fmt.Sprint(Files{Total: 2, ByExtension: map[string]int{"go": 2}}); got != want {
		t.Fatalf("heavy-tree descendants counted:\ngot  %s\nwant %s", got, want)
	}
	paths, err := Paths(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprint(paths), fmt.Sprint([]string{"main.go", "src/app.go"}); got != want {
		t.Fatalf("paths include heavy-tree descendants:\ngot  %s\nwant %s", got, want)
	}

	// Descendant changes must not change any fact (spec scenario "Heavy tree is present").
	write(t, root, "dist/assets/app.js", "changed\n")
	write(t, root, "node_modules/pkg/index.js", "changed\n")
	write(t, root, ".next/server/app.js", "changed\n")
	write(t, root, "storage/logs/laravel.log", "changed\n")
	after, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	beforeJSON, _ := json.Marshal(before)
	if afterJSON, _ := json.Marshal(after); string(beforeJSON) != string(afterJSON) {
		t.Fatalf("heavy-tree descendant changed facts:\nbefore %s\nafter  %s", beforeJSON, afterJSON)
	}
}

func TestInspectAttachesEcosystemFactsFromBoundedPaths(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"go.mod", "go.sum", "package.json", "pnpm-workspace.yaml", "pyproject.toml", "uv.lock", "next.config.js", "pytest.ini", "Makefile", "gradlew"} {
		data := "\n"
		if path == "package.json" {
			data = `{}`
		}
		write(t, root, path, data)
	}
	write(t, root, "node_modules/hidden/angular.json", "{}")

	first, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) {
		t.Fatalf("inspect bytes changed:\n%s\n%s", a, b)
	}
	got := string(a)
	for _, required := range []string{
		`"id":"go"`, `"id":"javascript"`, `"id":"python"`,
		`"path":"go.mod"`, `"path":"package.json"`, `"path":"pyproject.toml"`,
		`"path":"go.sum"`, `"path":"uv.lock"`, `"path":"pnpm-workspace.yaml"`,
		`"path":"gradlew"`, `"path":"next.config.js"`, `"path":"Makefile"`, `"path":"pytest.ini"`,
	} {
		if !strings.Contains(got, required) {
			t.Fatalf("missing %s in %s", required, got)
		}
	}
	if strings.Contains(got, "angular") {
		t.Fatalf("heavy-tree descendant became ecosystem evidence: %s", got)
	}

	plain := t.TempDir()
	write(t, plain, "README.md", "# plain\n")
	legacy, err := Inspect(plain, "")
	if err != nil {
		t.Fatal(err)
	}
	legacyJSON, _ := json.Marshal(legacy)
	for _, key := range []string{`"ecosystems"`, `"manifests"`, `"lockfiles"`, `"workspace_signals"`, `"project_wrappers"`, `"framework_signals"`, `"build_tools"`, `"test_tools"`} {
		if strings.Contains(string(legacyJSON), key) {
			t.Fatalf("legacy V1 JSON gained additive key %s: %s", key, legacyJSON)
		}
	}
}

func TestAgentsMDFact(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
		want    int
	}{
		{name: "trailing newline", content: "line one\nline two\nline three\n", want: 3},
		{name: "no trailing newline", content: "line one\nline two\nline three", want: 3},
		{name: "single line", content: "# demo\n", want: 1},
		{name: "empty file", content: "", want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "AGENTS.md", tc.content)
			facts, err := Inspect(root, "")
			if err != nil {
				t.Fatal(err)
			}
			if facts.AgentsMD == nil || facts.AgentsMD.Path != "AGENTS.md" || facts.AgentsMD.Lines != tc.want {
				t.Fatalf("agents_md fact: got %+v, want AGENTS.md with %d lines", facts.AgentsMD, tc.want)
			}
		})
	}
}

func TestAgentsMDFactOmitted(t *testing.T) {
	root := t.TempDir()
	write(t, root, "README.md", "# demo\n")
	facts, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if facts.AgentsMD != nil {
		t.Fatalf("agents_md must be nil without AGENTS.md: %+v", facts.AgentsMD)
	}
	data, _ := json.Marshal(facts)
	if strings.Contains(string(data), "agents_md") {
		t.Fatalf("agents_md must be omitted from JSON without AGENTS.md: %s", data)
	}
}

func TestOutputSignals(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	dirs := []string{
		".dart_tool", ".next", ".nuxt", ".venv", "__pycache__", "_build", "bin",
		"build", "cmake-build-debug", "coverage", "deps", "dist", "node_modules",
		"obj", "out", "result", "storage/logs", "target", "vendor", "venv",
	}
	for _, dir := range dirs {
		write(t, root, dir+"/nested/ignored.txt", "ignored\n")
	}

	first, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(first.OutputSignals))
	for i, sig := range first.OutputSignals {
		got[i] = sig.Path
	}
	if fmt.Sprint(got) != fmt.Sprint(dirs) {
		t.Fatalf("output signals:\ngot  %v\nwant %v", got, dirs)
	}
	second, err := Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	if b, _ := json.Marshal(second); string(a) != string(b) {
		t.Fatalf("output signals not byte-stable:\n%s\n%s", a, b)
	}
	for _, forbidden := range []string{"primary", "preferred", "recommend", "verdict"} {
		if strings.Contains(string(a), forbidden) {
			t.Fatalf("output signals contain decision %q: %s", forbidden, a)
		}
	}
}
