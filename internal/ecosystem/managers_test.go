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
		{"lockfile confirms cargo", []string{"Cargo.lock"}, []ManagerCandidate{{ID: "cargo", Confidence: ConfidenceConfirmed, Evidence: []string{"Cargo.lock"}}}},
		{"lockfile confirms composer", []string{"composer.lock"}, []ManagerCandidate{{ID: "composer", Confidence: ConfidenceConfirmed, Evidence: []string{"composer.lock"}}}},
		{"lockfile confirms conan", []string{"conan.lock"}, []ManagerCandidate{{ID: "conan", Confidence: ConfidenceConfirmed, Evidence: []string{"conan.lock"}}}},
		{"lockfile confirms helm", []string{"Chart.lock"}, []ManagerCandidate{{ID: "helm", Confidence: ConfidenceConfirmed, Evidence: []string{"Chart.lock"}}}},
		{"lockfile confirms mix", []string{"mix.lock"}, []ManagerCandidate{{ID: "mix", Confidence: ConfidenceConfirmed, Evidence: []string{"mix.lock"}}}},
		{"lockfile confirms nuget", []string{"packages.lock.json"}, []ManagerCandidate{{ID: "nuget", Confidence: ConfidenceConfirmed, Evidence: []string{"packages.lock.json"}}}},
		{"lockfile confirms pdm", []string{"pdm.lock"}, []ManagerCandidate{{ID: "pdm", Confidence: ConfidenceConfirmed, Evidence: []string{"pdm.lock"}}}},
		{"lockfile confirms pub", []string{"pubspec.lock"}, []ManagerCandidate{{ID: "pub", Confidence: ConfidenceConfirmed, Evidence: []string{"pubspec.lock"}}}},
		{"lockfile confirms swift", []string{"Package.resolved"}, []ManagerCandidate{{ID: "swift", Confidence: ConfidenceConfirmed, Evidence: []string{"Package.resolved"}}}},
		{"lockfile confirms terraform", []string{".terraform.lock.hcl"}, []ManagerCandidate{{ID: "terraform", Confidence: ConfidenceConfirmed, Evidence: []string{".terraform.lock.hcl"}}}},
		{"manifest infers bundler", []string{"Gemfile"}, []ManagerCandidate{{ID: "bundler", Confidence: ConfidenceInferred, Evidence: []string{"Gemfile"}}}},
		{"manifest infers cargo", []string{"Cargo.toml"}, []ManagerCandidate{{ID: "cargo", Confidence: ConfidenceInferred, Evidence: []string{"Cargo.toml"}}}},
		{"manifest infers cmake", []string{"CMakeLists.txt"}, []ManagerCandidate{{ID: "cmake", Confidence: ConfidenceInferred, Evidence: []string{"CMakeLists.txt"}}}},
		{"manifest infers composer", []string{"composer.json"}, []ManagerCandidate{{ID: "composer", Confidence: ConfidenceInferred, Evidence: []string{"composer.json"}}}},
		{"manifest infers conan", []string{"conanfile.py"}, []ManagerCandidate{{ID: "conan", Confidence: ConfidenceInferred, Evidence: []string{"conanfile.py"}}}},
		{"manifest infers gradle", []string{"build.gradle"}, []ManagerCandidate{{ID: "gradle", Confidence: ConfidenceInferred, Evidence: []string{"build.gradle"}}}},
		{"manifest infers maven", []string{"pom.xml"}, []ManagerCandidate{{ID: "maven", Confidence: ConfidenceInferred, Evidence: []string{"pom.xml"}}}},
		{"manifest infers mix", []string{"mix.exs"}, []ManagerCandidate{{ID: "mix", Confidence: ConfidenceInferred, Evidence: []string{"mix.exs"}}}},
		{"manifest infers pub", []string{"pubspec.yaml"}, []ManagerCandidate{{ID: "pub", Confidence: ConfidenceInferred, Evidence: []string{"pubspec.yaml"}}}},
		{"manifest infers rustup", []string{"rust-toolchain.toml"}, []ManagerCandidate{{ID: "rustup", Confidence: ConfidenceInferred, Evidence: []string{"rust-toolchain.toml"}}}},
		{"manifest infers swift", []string{"Package.swift"}, []ManagerCandidate{{ID: "swift", Confidence: ConfidenceInferred, Evidence: []string{"Package.swift"}}}},
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
		{"pdm and poetry lockfiles", []string{"pdm.lock", "poetry.lock"}, []string{"pdm", "poetry"}, "lockfiles evidence distinct managers: pdm and poetry"},
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
		{"pdm lockfile confirms pdm over pyproject", []string{"pyproject.toml", "pdm.lock"}, []ManagerCandidate{{ID: "pdm", Confidence: ConfidenceConfirmed, Evidence: []string{"pdm.lock"}}}},
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

func TestResolveManagersFamiliesIndependent(t *testing.T) {
	// composer.lock and Cargo.lock confirm managers within their own
	// ecosystems; distinct ecosystems never produce a cross-ecosystem conflict.
	candidates, conflicts := resolveManagers([]string{"composer.lock", "Cargo.lock", "Gemfile.lock", "mix.lock", "pubspec.lock", "conan.lock", "packages.lock.json", ".terraform.lock.hcl"})
	got := map[string]string{}
	for _, c := range candidates {
		got[c.ID] = c.Confidence
	}
	want := map[string]string{
		"bundler": ConfidenceConfirmed, "cargo": ConfidenceConfirmed, "composer": ConfidenceConfirmed,
		"conan": ConfidenceConfirmed, "mix": ConfidenceConfirmed, "nuget": ConfidenceConfirmed,
		"pub": ConfidenceConfirmed, "terraform": ConfidenceConfirmed,
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("families:\ngot  %v\nwant %v", got, want)
	}
	if len(conflicts) != 0 {
		t.Fatalf("cross-ecosystem managers must not conflict: %v", conflicts)
	}
}

func TestResolveManagersDeterministicAndDecisionFree(t *testing.T) {
	paths := []string{"gradlew", "go.mod", "go.sum", "package.json", "pyproject.toml", "pnpm-lock.yaml", "bun.lock", "uv.lock"}
	reversed := []string{"uv.lock", "bun.lock", "pnpm-lock.yaml", "pyproject.toml", "package.json", "go.sum", "go.mod", "gradlew"}
	forward, err := json.Marshal(Detect("", paths))
	if err != nil {
		t.Fatal(err)
	}
	back, err := json.Marshal(Detect("", reversed))
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
