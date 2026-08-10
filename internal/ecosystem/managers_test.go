package ecosystem

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestResolveManagersConfidenceLevels(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  []ManagerCandidate
	}{
		{"lockfile confirms npm", []string{"package-lock.json"}, []ManagerCandidate{{ID: "npm", Confidence: ConfidenceConfirmed, Evidence: []string{"package-lock.json"}}}},
		{"lockfile confirms pnpm", []string{"pnpm-lock.yaml"}, []ManagerCandidate{{ID: "pnpm", Confidence: ConfidenceConfirmed, Evidence: []string{"pnpm-lock.yaml"}}}},
		{"lockfile confirms bun", []string{"bun.lock"}, []ManagerCandidate{{ID: "bun", Confidence: ConfidenceConfirmed, Evidence: []string{"bun.lock"}}}},
		{"lockfile confirms yarn", []string{"yarn.lock"}, []ManagerCandidate{{ID: "yarn", Confidence: ConfidenceConfirmed, Evidence: []string{"yarn.lock"}}}},
		{"lockfile confirms go", []string{"go.sum"}, []ManagerCandidate{{ID: "go", Confidence: ConfidenceConfirmed, Evidence: []string{"go.sum"}}}},
		{"lockfile confirms uv", []string{"uv.lock"}, []ManagerCandidate{{ID: "uv", Confidence: ConfidenceConfirmed, Evidence: []string{"uv.lock"}}}},
		{"lockfile confirms poetry", []string{"poetry.lock"}, []ManagerCandidate{{ID: "poetry", Confidence: ConfidenceConfirmed, Evidence: []string{"poetry.lock"}}}},
		{"lockfile confirms pipenv", []string{"Pipfile.lock"}, []ManagerCandidate{{ID: "pipenv", Confidence: ConfidenceConfirmed, Evidence: []string{"Pipfile.lock"}}}},
		{"lockfile confirms deno", []string{"deno.lock"}, []ManagerCandidate{{ID: "deno", Confidence: ConfidenceConfirmed, Evidence: []string{"deno.lock"}}}},
		{"wrapper confirms gradle", []string{"gradlew"}, []ManagerCandidate{{ID: "gradle", Confidence: ConfidenceConfirmed, Evidence: []string{"gradlew"}}}},
		{"wrapper confirms maven", []string{"mvnw"}, []ManagerCandidate{{ID: "maven", Confidence: ConfidenceConfirmed, Evidence: []string{"mvnw"}}}},
		{"manifest infers go", []string{"go.mod"}, []ManagerCandidate{{ID: "go", Confidence: ConfidenceInferred, Evidence: []string{"go.mod"}}}},
		{"manifest infers pip", []string{"requirements.txt"}, []ManagerCandidate{{ID: "pip", Confidence: ConfidenceInferred, Evidence: []string{"requirements.txt"}}}},
		{"manifest infers pipenv", []string{"Pipfile"}, []ManagerCandidate{{ID: "pipenv", Confidence: ConfidenceInferred, Evidence: []string{"Pipfile"}}}},
		{"manifest infers deno", []string{"deno.json"}, []ManagerCandidate{{ID: "deno", Confidence: ConfidenceInferred, Evidence: []string{"deno.json"}}}},
		{"pyproject alone is ambiguous", []string{"pyproject.toml"}, []ManagerCandidate{
			{ID: "pdm", Confidence: ConfidenceAmbiguous, Evidence: []string{"pyproject.toml"}},
			{ID: "pip", Confidence: ConfidenceAmbiguous, Evidence: []string{"pyproject.toml"}},
			{ID: "poetry", Confidence: ConfidenceAmbiguous, Evidence: []string{"pyproject.toml"}},
			{ID: "uv", Confidence: ConfidenceAmbiguous, Evidence: []string{"pyproject.toml"}},
		}},
		{"package.json alone is ambiguous", []string{"package.json"}, []ManagerCandidate{
			{ID: "bun", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "npm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "pnpm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "yarn", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates, _ := resolveManagers(tt.paths)
			if fmt.Sprint(candidates) != fmt.Sprint(tt.want) {
				t.Fatalf("candidates:\ngot  %s\nwant %s", fmt.Sprint(candidates), fmt.Sprint(tt.want))
			}
		})
	}
}

func TestResolveManagersWrapperPrecedence(t *testing.T) {
	// A project wrapper is execution-grade evidence: gradle is confirmed
	// while generic manifests stay ambiguous, outranking any global binary.
	candidates, conflicts := resolveManagers([]string{"gradlew", "package.json", "pyproject.toml"})
	confidence := map[string]string{}
	for _, c := range candidates {
		confidence[c.ID] = c.Confidence
	}
	if confidence["gradle"] != ConfidenceConfirmed {
		t.Fatalf("wrapper must confirm gradle: %v", candidates)
	}
	for _, id := range []string{"bun", "npm", "pnpm", "yarn", "pip", "poetry", "pdm", "uv"} {
		if confidence[id] != ConfidenceAmbiguous {
			t.Fatalf("generic manifests must stay ambiguous for %s: %v", id, candidates)
		}
	}
	if len(conflicts) != 0 {
		t.Fatalf("unexpected conflicts: %v", conflicts)
	}
}

func TestResolveManagersPnpmBunConflict(t *testing.T) {
	candidates, conflicts := resolveManagers([]string{"pnpm-lock.yaml", "bun.lock"})
	want := []ManagerCandidate{
		{ID: "bun", Confidence: ConfidenceConfirmed, Evidence: []string{"bun.lock"}},
		{ID: "pnpm", Confidence: ConfidenceConfirmed, Evidence: []string{"pnpm-lock.yaml"}},
	}
	if fmt.Sprint(candidates) != fmt.Sprint(want) {
		t.Fatalf("candidates:\ngot  %s\nwant %s", fmt.Sprint(candidates), fmt.Sprint(want))
	}
	if len(conflicts) != 1 || fmt.Sprint(conflicts[0].Managers) != fmt.Sprint([]string{"bun", "pnpm"}) {
		t.Fatalf("want one bun/pnpm conflict, got %v", conflicts)
	}
	if !strings.Contains(conflicts[0].Reason, "pnpm") || !strings.Contains(conflicts[0].Reason, "bun") {
		t.Fatalf("conflict reason must retain both managers: %q", conflicts[0].Reason)
	}
	for _, forbidden := range []string{"preferred", "migration", "recommend", "selected", "primary"} {
		if strings.Contains(conflicts[0].Reason, forbidden) {
			t.Fatalf("conflict reason contains decision token %q: %q", forbidden, conflicts[0].Reason)
		}
	}
}

func TestResolveManagersConflictVariants(t *testing.T) {
	tests := []struct {
		name       string
		paths      []string
		want       []string
		wantReason string
	}{
		{"npm and yarn lockfiles", []string{"yarn.lock", "package-lock.json"}, []string{"npm", "yarn"}, "lockfiles evidence distinct managers: npm and yarn"},
		{"poetry and uv lockfiles", []string{"poetry.lock", "uv.lock"}, []string{"poetry", "uv"}, "lockfiles evidence distinct managers: poetry and uv"},
		{"deno and pnpm lockfiles", []string{"deno.lock", "pnpm-lock.yaml"}, []string{"deno", "pnpm"}, "lockfiles evidence distinct managers: deno and pnpm"},
		{"gradle and maven wrappers", []string{"mvnw", "gradlew"}, []string{"gradle", "maven"}, "project wrappers evidence distinct managers: gradle and maven"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, conflicts := resolveManagers(tt.paths)
			if len(conflicts) != 1 {
				t.Fatalf("want one conflict, got %v", conflicts)
			}
			if fmt.Sprint(conflicts[0].Managers) != fmt.Sprint(tt.want) || conflicts[0].Reason != tt.wantReason {
				t.Fatalf("conflict:\ngot  %+v\nwant managers %v reason %q", conflicts[0], tt.want, tt.wantReason)
			}
		})
	}
	_, conflicts := resolveManagers([]string{"go.sum", "package-lock.json"})
	if len(conflicts) != 0 {
		t.Fatalf("cross-ecosystem managers must not conflict: %v", conflicts)
	}
}

func TestResolveManagersFamilyEvidencePrecedence(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  []ManagerCandidate
	}{
		{"pyproject plus uv lockfile confirms only uv", []string{"pyproject.toml", "uv.lock"}, []ManagerCandidate{{ID: "uv", Confidence: ConfidenceConfirmed, Evidence: []string{"uv.lock"}}}},
		{"package.json plus yarn lockfile confirms only yarn", []string{"package.json", "yarn.lock"}, []ManagerCandidate{{ID: "yarn", Confidence: ConfidenceConfirmed, Evidence: []string{"yarn.lock"}}}},
		{"deno manifest with package.json keeps JS ambiguous", []string{"package.json", "deno.json"}, []ManagerCandidate{
			{ID: "bun", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "deno", Confidence: ConfidenceInferred, Evidence: []string{"deno.json"}},
			{ID: "npm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "pnpm", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
			{ID: "yarn", Confidence: ConfidenceAmbiguous, Evidence: []string{"package.json"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates, _ := resolveManagers(tt.paths)
			if fmt.Sprint(candidates) != fmt.Sprint(tt.want) {
				t.Fatalf("candidates:\ngot  %s\nwant %s", fmt.Sprint(candidates), fmt.Sprint(tt.want))
			}
		})
	}
}

func TestResolveManagersDeterministicAndDecisionFree(t *testing.T) {
	paths := []string{"gradlew", "go.mod", "go.sum", "package.json", "pyproject.toml", "pnpm-lock.yaml", "bun.lock", "uv.lock"}
	reversed := []string{"uv.lock", "bun.lock", "pnpm-lock.yaml", "pyproject.toml", "package.json", "go.sum", "go.mod", "gradlew"}
	forward, err := json.Marshal(Detect(paths))
	if err != nil {
		t.Fatal(err)
	}
	back, err := json.Marshal(Detect(reversed))
	if err != nil {
		t.Fatal(err)
	}
	if string(forward) != string(back) {
		t.Fatalf("input order changed manager bytes:\n%s\n%s", forward, back)
	}
	for _, forbidden := range []string{"primary", "preferred", "migration", "recommend", "selected"} {
		if strings.Contains(string(forward), forbidden) {
			t.Fatalf("manager facts contain decision token %q: %s", forbidden, forward)
		}
	}
}
