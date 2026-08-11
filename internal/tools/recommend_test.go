package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// writeFile creates rel with content under root.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// addGitHubRemote declares an origin remote on the already-initialized root.
func addGitHubRemote(t *testing.T, root string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", root, "remote", "add", "origin", "https://github.com/acme/app.git").CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %v: %s", err, out)
	}
}

func TestRecommendEmptyRepo(t *testing.T) {
	root := t.TempDir()
	fakeGit(t, root)
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts.Candidates) != 0 {
		t.Fatalf("empty repo must have no candidates: %+v", facts.Candidates)
	}
	data, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	// Empty ordered collection, never null.
	if !strings.Contains(string(data), `"candidates":[]`) {
		t.Fatalf("candidates must serialize as an empty array: %s", data)
	}
	if facts.Summary() != "No capability candidates" {
		t.Fatalf("summary: %s", facts.Summary())
	}
}

func TestRecommendNoSynthesizedDefault(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 501; i++ {
		writeFile(t, root, "f"+strconv.Itoa(i)+".txt", "x")
	}
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	// A large signal-free repo used to synthesize a Semble verdict; Slice 6
	// grounds only documented signals, so no default may appear.
	if len(facts.Candidates) != 0 {
		t.Fatalf("signal-free repo must stay empty: %+v", facts.Candidates)
	}
}

func TestRecommendGroundedOrderAndStability(t *testing.T) {
	root := t.TempDir()
	fakeGit(t, root)
	addGitHubRemote(t, root)
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.24\n")
	writeFile(t, root, "go.sum", "checksum\n")
	writeFile(t, root, "package.json", "{}\n")
	writeFile(t, root, "pnpm-workspace.yaml", "packages:\n  - packages/*\n")
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: '9'\n")
	writeFile(t, root, "bun.lock", "lockfileVersion: 1\n")
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	first, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	// D3: lockfile-only Context7 candidates are gone; D4 keeps workspace
	// topology (CodeGraph), LSP surface + module boundaries (Serena), and
	// output pressure (RTK + Headroom) as the provider signals.
	wantCaps := []string{"token_optimized_shell", "github_context", "dependency_graph", "go_toolchain", "javascript_toolchain", "symbol_intelligence", "general_context_compression"}
	wantIDs := []string{"rtk", "gh", "codegraph", "go", "node", "serena", "headroom"}
	if len(first.Candidates) != len(wantCaps) {
		t.Fatalf("expected %d candidates, got %d: %+v", len(wantCaps), len(first.Candidates), first.Candidates)
	}
	for i, candidate := range first.Candidates {
		if candidate.Capability != wantCaps[i] || candidate.CatalogID != wantIDs[i] {
			t.Fatalf("candidate %d: capability=%q catalog=%q, want %q/%q", i, candidate.Capability, candidate.CatalogID, wantCaps[i], wantIDs[i])
		}
		if candidate.Reason == "" || candidate.Observed == "" || candidate.Signal == "" || candidate.Capabilities == nil {
			t.Fatalf("candidate %d incomplete: %+v", i, candidate)
		}
	}
	if !strings.Contains(first.Candidates[2].Observed, "pnpm-workspace.yaml") {
		t.Fatalf("CodeGraph observed must cite the workspace signal: %+v", first.Candidates[2])
	}
	if !strings.Contains(first.Candidates[5].Observed, "go") || !strings.Contains(first.Candidates[5].Observed, "workspace_signals") {
		t.Fatalf("Serena observed must cite LSP surface and module boundaries: %+v", first.Candidates[5])
	}
	data, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(again) {
		t.Fatal("recommend JSON must be byte-stable across runs")
	}
}

// D3: lockfile presence and manager conflicts alone never generate Context7
// candidates — only central framework facts with a parsed version do.
func TestRecommendLockfileOnlyNoContext7(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: '9'\n")
	writeFile(t, root, "bun.lock", "lockfileVersion: 1\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range facts.Candidates {
		if candidate.CatalogID == "context7" {
			t.Fatalf("lockfile-only must not fire Context7 (D3): %+v", facts.Candidates)
		}
	}
	if len(facts.Candidates) != 0 {
		t.Fatalf("lockfile-only repository must have no candidates: %+v", facts.Candidates)
	}
}

func TestRecommendDetectOnlyTools(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files map[string]string
		want  string
	}{
		{name: "go manifest", files: map[string]string{"go.mod": "module example.com/app\n"}, want: "go"},
		{name: "javascript manifest", files: map[string]string{"package.json": "{}\n"}, want: "node"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for rel, content := range tc.files {
				writeFile(t, root, rel, content)
			}
			facts, err := Recommend(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(facts.Candidates) != 1 || facts.Candidates[0].CatalogID != tc.want {
				t.Fatalf("expected exactly one %s candidate: %+v", tc.want, facts.Candidates)
			}
			caps := *facts.Candidates[0].Capabilities
			if caps.Detect != Supported || caps.Install != Unsupported || caps.Configure != Unsupported {
				t.Fatalf("%s must be detect-only: %+v", tc.want, caps)
			}
		})
	}
	// §45: every detect-only toolchain entry (npm…tofu, plus the §20 entries
	// without recipes) carries executables/versionArgs and honest capability
	// truth: detect and version supported, install never claimed.
	detectOnly := []string{"bun", "bundle", "cargo", "cmake", "conan", "dart", "deno", "dotnet",
		"flutter", "go", "gradle", "maven", "mix", "nix", "node", "npm", "pdm", "pip", "pipenv",
		"pnpm", "poetry", "rustup", "terraform", "tofu", "yarn"}
	for _, id := range detectOnly {
		entry := entryByID(t, id)
		if entry.Install != nil || len(entry.Executables) == 0 || len(entry.VersionArgs) == 0 {
			t.Fatalf("%s must be detect-only with executables/versionArgs: %+v", id, entry)
		}
		if entry.Capabilities.Detect != Supported || entry.Capabilities.Version != Supported || entry.Capabilities.Install != Unsupported {
			t.Fatalf("%s must claim detect/version only: %+v", id, entry.Capabilities)
		}
	}
}

// D3 scenario: a central version-sensitive framework (ratatui 0.29.0) with a
// proposed-artifact-relevant API produces a Context7 candidate carrying
// catalog-truth capabilities and no lifecycle support.
func TestRecommendContext7FrameworkGated(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Cargo.toml", "[package]\nname = \"demo\"\n\n[dependencies]\nratatui = \"0.29.0\"\n")
	writeFile(t, root, "src/main.rs", "use ratatui::Frame;\nfn main() {}\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts.Candidates) != 1 {
		t.Fatalf("expected one context7 candidate: %+v", facts.Candidates)
	}
	got := facts.Candidates[0]
	caps := *got.Capabilities
	if got.CatalogID != "context7" || got.Capability != "versioned_documentation" || !strings.Contains(got.Observed, "ratatui 0.29.0") ||
		caps != entryByID(t, "context7").Capabilities || caps.Install != Unsupported || caps.Configure != Unsupported || caps.Integration != Unsupported {
		t.Fatalf("context7 candidate must carry catalog truth without lifecycle: %+v", got)
	}
}

func TestRecommendV1ReaderCompatibility(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n")
	writeFile(t, root, "go.sum", "checksum\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	var v1 struct {
		SchemaVersion string `json:"schema_version"`
		Candidates    []struct {
			Capability string `json:"capability"`
			Candidate  string `json:"candidate"`
			Signal     string `json:"signal"`
			Observed   string `json:"observed"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &v1); err != nil {
		t.Fatalf("V1 reader must accept enriched output: %v", err)
	}
	if v1.SchemaVersion != RecommendSchemaVersion {
		t.Fatalf("schema: %s", v1.SchemaVersion)
	}
	// D3: go.sum lockfile alone no longer generates a Context7 candidate.
	if len(v1.Candidates) != 1 || v1.Candidates[0].Candidate != "go" {
		t.Fatalf("V1 fields lost: %+v", v1.Candidates)
	}
}

// D4: provider signals stay conditional — CodeGraph needs workspace topology
// (never a deps>N verdict), Semble the documented §14.3 scale band, and a
// small single-crate repository stays free of provider candidates.
func TestRecommendD4ConditionalProviderSignals(t *testing.T) {
	depsJSON := "{\"dependencies\":{"
	for i := 0; i < 60; i++ {
		depsJSON += "\"dep" + strconv.Itoa(i) + "\":\"1.0.0\","
	}
	depsJSON = strings.TrimSuffix(depsJSON, ",") + "}}\n"
	root := t.TempDir()
	writeFile(t, root, "package.json", depsJSON)
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	semble := false
	for _, candidate := range facts.Candidates {
		switch candidate.CatalogID {
		case "semble":
			semble = true
			if !strings.Contains(candidate.Observed, "deps=60") || candidate.Capability != "semantic_retrieval" {
				t.Fatalf("semble scale evidence: %+v", candidate)
			}
		case "codegraph":
			t.Fatalf("deps>N must not fire CodeGraph without workspace topology: %+v", facts.Candidates)
		}
	}
	if !semble {
		t.Fatalf("60-dependency repository must meet the scale band: %+v", facts.Candidates)
	}
	writeFile(t, root, "pnpm-workspace.yaml", "packages:\n  - packages/*\n")
	facts, err = Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range facts.Candidates {
		if candidate.CatalogID == "codegraph" && strings.Contains(candidate.Observed, "pnpm-workspace.yaml") {
			found = true
		}
	}
	if !found {
		t.Fatalf("workspace topology must fire CodeGraph: %+v", facts.Candidates)
	}
	small := t.TempDir()
	writeFile(t, small, "Cargo.toml", "[package]\nname = \"small\"\n")
	writeFile(t, small, "src/main.rs", "fn main() {}\n")
	facts, err = Recommend(small)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range facts.Candidates {
		switch candidate.CatalogID {
		case "codegraph", "context7", "headroom", "semble", "serena":
			t.Fatalf("small repository must not fire provider candidates: %+v", facts.Candidates)
		}
	}
}
