package ecosystem

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestDetectMixedRepositoryFacts(t *testing.T) {
	paths := []string{
		"vitest.config.ts", "uv.lock", "pyproject.toml", "pnpm-workspace.yaml",
		"package.json", "next.config.js", "go.sum", "go.mod", "gradlew",
	}
	want := Facts{
		Ecosystems: []Ecosystem{
			{ID: "go", Evidence: []string{"go.mod"}},
			{ID: "javascript", Evidence: []string{"package.json"}},
			{ID: "python", Evidence: []string{"pyproject.toml"}},
		},
		Manifests:        []Signal{{ID: "go", Path: "go.mod"}, {ID: "javascript", Path: "package.json"}, {ID: "python", Path: "pyproject.toml"}},
		Lockfiles:        []Signal{{ID: "go", Path: "go.sum"}, {ID: "uv", Path: "uv.lock"}},
		WorkspaceSignals: []Signal{{ID: "pnpm", Path: "pnpm-workspace.yaml"}},
		ProjectWrappers:  []Signal{{ID: "gradle", Path: "gradlew"}},
		FrameworkSignals: []Signal{{ID: "nextjs", Path: "next.config.js"}},
		TestTools:        []Signal{{ID: "vitest", Path: "vitest.config.ts"}},
		PackageManagers: []ManagerCandidate{
			{ID: "bun", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "go", Confidence: ConfidenceConfirmed, Evidence: []string{"go.mod", "go.sum"}},
			{ID: "gradle", Confidence: ConfidenceConfirmed, Evidence: []string{"gradlew"}},
			{ID: "npm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "pnpm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "uv", Confidence: ConfidenceConfirmed, Evidence: []string{"uv.lock"}},
			{ID: "yarn", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
		},
	}
	if got := fmt.Sprint(Detect(paths)); got != fmt.Sprint(want) {
		t.Fatalf("mixed facts:\ngot  %s\nwant %s", got, fmt.Sprint(want))
	}

	forward, _ := json.Marshal(Detect(paths))
	reverse, _ := json.Marshal(Detect([]string{
		"gradlew", "go.mod", "go.sum", "next.config.js", "package.json",
		"pnpm-workspace.yaml", "pyproject.toml", "uv.lock", "vitest.config.ts",
	}))
	if string(forward) != string(reverse) {
		t.Fatalf("input order changed bytes:\n%s\n%s", forward, reverse)
	}
	for _, forbidden := range []string{"primary", "preferred", "migration", "recommend"} {
		if strings.Contains(string(forward), forbidden) {
			t.Fatalf("facts contain semantic decision %q: %s", forbidden, forward)
		}
	}
}

func TestDetectBuildTestAndFrameworkTables(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		field func(Facts) []Signal
		want  Signal
	}{
		{"build tool", "tools/Makefile", func(f Facts) []Signal { return f.BuildTools }, Signal{ID: "make", Path: "tools/Makefile"}},
		{"test tool", "api/pytest.ini", func(f Facts) []Signal { return f.TestTools }, Signal{ID: "pytest", Path: "api/pytest.ini"}},
		{"framework", "web/angular.json", func(f Facts) []Signal { return f.FrameworkSignals }, Signal{ID: "angular", Path: "web/angular.json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.field(Detect([]string{tt.path}))
			if fmt.Sprint(got) != fmt.Sprint([]Signal{tt.want}) {
				t.Fatalf("got %v, want %v", got, []Signal{tt.want})
			}
		})
	}
}
