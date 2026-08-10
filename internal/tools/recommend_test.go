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
	wantCaps := []string{"token_optimized_shell", "github_context", "dependency_graph", "go_toolchain", "javascript_toolchain", "versioned_documentation", "package_manager_conflict"}
	wantIDs := []string{"rtk", "gh", "codegraph", "go", "node", "context7", "context7"}
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

func TestRecommendConflictAwareEvidence(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: '9'\n")
	writeFile(t, root, "bun.lock", "lockfileVersion: 1\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts.Candidates) != 2 {
		t.Fatalf("expected 2 candidates (documentation + conflict), got %+v", facts.Candidates)
	}
	conflict := facts.Candidates[1]
	if conflict.Capability != "package_manager_conflict" || conflict.CatalogID != "context7" {
		t.Fatalf("conflict candidate: %+v", conflict)
	}
	if !strings.Contains(conflict.Observed, "bun") || !strings.Contains(conflict.Observed, "pnpm") {
		t.Fatalf("conflict observed must name both managers: %+v", conflict)
	}
	if conflict.Capabilities.Install != Unsupported || conflict.Capabilities.Configure != Unsupported {
		t.Fatalf("provider lifecycle must stay unsupported: %+v", conflict.Capabilities)
	}
	data, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"preferred", "primary", "selected"} {
		if strings.Contains(string(data), token) {
			t.Fatalf("conflict output must not choose a manager (%q): %s", token, data)
		}
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
}

func TestRecommendProviderWithoutLifecycle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "package-lock.json", "{}\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts.Candidates) != 1 || facts.Candidates[0].CatalogID != "context7" {
		t.Fatalf("expected one context7 candidate: %+v", facts.Candidates)
	}
	got := *facts.Candidates[0].Capabilities
	want := entryByID(t, "context7").Capabilities
	if got != want {
		t.Fatalf("candidate capabilities must match catalog truth: %+v vs %+v", got, want)
	}
	for _, state := range []CapabilityState{got.Install, got.Configure, got.Integration, got.SideEffects} {
		if state != Unsupported {
			t.Fatalf("provider lifecycle must remain unsupported: %+v", got)
		}
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
	if len(v1.Candidates) != 2 || v1.Candidates[0].Candidate != "go" {
		t.Fatalf("V1 fields lost: %+v", v1.Candidates)
	}
}
