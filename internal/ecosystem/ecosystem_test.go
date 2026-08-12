package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		FrameworkFacts: []FrameworkFact{
			{Name: "nextjs", Evidence: []string{"next.config.js"}},
		},
		TestTools: []Signal{{ID: "vitest", Path: "vitest.config.ts"}},
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
	if got := fmt.Sprint(Detect("", paths)); got != fmt.Sprint(want) {
		t.Fatalf("mixed facts:\ngot  %s\nwant %s", got, fmt.Sprint(want))
	}

	forward, _ := json.Marshal(Detect("", paths))
	reverse, _ := json.Marshal(Detect("", []string{
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
			got := tt.field(Detect("", []string{tt.path}))
			if fmt.Sprint(got) != fmt.Sprint([]Signal{tt.want}) {
				t.Fatalf("got %v, want %v", got, []Signal{tt.want})
			}
		})
	}
}

func TestSuffixRules(t *testing.T) {
	tests := [][2]string{
		{"infra/main.tf", "terraform"}, {"app/App.sln", "dotnet"}, {"app/App.slnx", "dotnet"},
		{"src/Lib.csproj", "dotnet"}, {"src/App.fsproj", "dotnet"}, {"lib/demo.gemspec", "ruby"},
		{"ios/App.xcodeproj", "swift"}, {"ios/App.xcworkspace", "swift"}, {"phpunit.xml.dist", "phpunit"},
		{"rust-toolchain.toml", "rust"}, {"CMakeUserPresets.json", "cmake"}, {"CMakeLists.txt", "cmake"},
	}
	for _, tt := range tests {
		raw, _ := json.Marshal(Detect("", []string{tt[0]}))
		if !strings.Contains(string(raw), `"id":"`+tt[1]+`"`) || !strings.Contains(string(raw), `"path":"`+tt[0]+`"`) {
			t.Fatalf("suffix %q did not fire for %s: %s", tt[1], tt[0], raw)
		}
	}

	if json, _ := json.Marshal(Detect("", []string{"notes.txt", "script.sh", "archive.tar.gz", "config.yaml"})); len(json) != 2 {
		t.Fatalf("unknown suffix fabricated signals: %s", json)
	}

	mixed := []string{"CMakeLists.txt", "CMakeUserPresets.json", "infra/main.tf", "rust-toolchain.toml", "phpunit.xml.dist", "src/App.csproj"}
	forward, _ := json.Marshal(Detect("", mixed))
	reversed := []string{"src/App.csproj", "phpunit.xml.dist", "rust-toolchain.toml", "infra/main.tf", "CMakeUserPresets.json", "CMakeLists.txt"}
	back, _ := json.Marshal(Detect("", reversed))
	if string(forward) != string(back) {
		t.Fatalf("suffix interleave changed bytes:\n%s\n%s", forward, back)
	}
}

func TestFullLockfileCoverage(t *testing.T) {
	tests := [][2]string{
		{"package-lock.json", "npm"}, {"npm-shrinkwrap.json", "npm"}, {"pnpm-lock.yaml", "pnpm"}, {"yarn.lock", "yarn"},
		{"bun.lock", "bun"}, {"deno.lock", "deno"}, {"uv.lock", "uv"}, {"poetry.lock", "poetry"}, {"Pipfile.lock", "python"},
		{"pdm.lock", "pdm"}, {"composer.lock", "composer"}, {"Cargo.lock", "cargo"}, {"flake.lock", "nix"},
		{"go.sum", "go"}, {"go.work.sum", "go"}, {"Gemfile.lock", "bundler"}, {"mix.lock", "mix"}, {"pubspec.lock", "pub"},
		{"Package.resolved", "swift"}, {"packages.lock.json", "nuget"}, {"conan.lock", "conan"}, {".terraform.lock.hcl", "terraform"}, {"Chart.lock", "helm"},
	}
	for _, tt := range tests {
		got := Detect("", []string{tt[0]})
		if fmt.Sprint(got.Lockfiles) != fmt.Sprint([]Signal{{ID: tt[1], Path: tt[0]}}) {
			t.Fatalf("lockfile %s: got %v", tt[0], got.Lockfiles)
		}
	}
}

func TestManifestEcosystemMatrix(t *testing.T) {
	tests := []struct {
		path string
		id   string
	}{
		{"composer.json", "php"}, {"Cargo.toml", "rust"}, {"flake.nix", "nix"},
		{"pubspec.yaml", "dart"}, {"Package.swift", "swift"}, {"Dockerfile", "docker"},
		{"Chart.yaml", "helm"}, {"kustomization.yaml", "kustomize"}, {"ansible.cfg", "ansible"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := Detect("", []string{tt.path}); len(got.Ecosystems) != 1 || got.Ecosystems[0].ID != tt.id {
				t.Fatalf("ecosystem for %s: got %+v, want %s", tt.path, got.Ecosystems, tt.id)
			}
		})
	}
}

func TestCargoInlineTableVersion(t *testing.T) {
	// Regression: cargoDepRe must match the inline-table alternative
	// `name = { version = "..." }` (stray 0x08 byte previously killed it).
	root := t.TempDir()
	write(t, root, "Cargo.toml", "[package]\nname = \"demo\"\n\n[dependencies]\nratatui = { version = \"0.29.0\", features = [\"crossterm\"] }\nserde = { version = \"1.0.0\" }\n")
	write(t, root, "src/main.rs", "fn main() {}\n")
	got := Detect(root, []string{"Cargo.toml", "src/main.rs"})
	if len(got.FrameworkFacts) != 1 || got.FrameworkFacts[0].Name != "ratatui" || got.FrameworkFacts[0].Version != "0.29.0" {
		t.Fatalf("inline-table version not parsed: %+v", got.FrameworkFacts)
	}
	if got := fmt.Sprint(got.FrameworkFacts[0].Evidence); got != fmt.Sprint([]string{"Cargo.toml"}) {
		t.Fatalf("evidence: got %s", got)
	}
}

func TestFrameworkVersionEvidence(t *testing.T) {
	root := t.TempDir()
	write(t, root, "Cargo.toml", "[package]\nname = \"demo\"\n\n[dependencies]\nratatui = \"0.29.0\"\n")
	write(t, root, "src/ui.rs", "use ratatui::Frame;\n")
	write(t, root, "src/main.rs", "fn main() {}\n")
	write(t, root, "manage.py", "#!/usr/bin/env python\nimport django.core.management\n")
	write(t, root, "views.py", "from django.http import HttpResponse\n")
	got := Detect(root, []string{"Cargo.toml", "manage.py", "src/main.rs", "src/ui.rs", "views.py"})
	want := []FrameworkFact{{Name: "django", Evidence: []string{"manage.py"}, CentralitySignals: []Signal{{ID: "django", Path: "views.py"}}}, {Name: "ratatui", Version: "0.29.0", Evidence: []string{"Cargo.toml"}, CentralitySignals: []Signal{{ID: "ratatui", Path: "src/ui.rs"}}}}
	if fmt.Sprint(got.FrameworkFacts) != fmt.Sprint(want) {
		t.Fatalf("framework facts:\ngot  %s\nwant %s", fmt.Sprint(got.FrameworkFacts), fmt.Sprint(want))
	}
}

func TestFrameworkVersionAbsentRetainsEvidence(t *testing.T) {
	root := t.TempDir()
	write(t, root, "package.json", `{"dependencies":{"next":{}}}`)
	got := Detect(root, []string{"package.json"})
	if len(got.FrameworkFacts) != 1 || got.FrameworkFacts[0].Name != "nextjs" || got.FrameworkFacts[0].Version != "" {
		t.Fatalf("framework without version was dropped: %+v", got.FrameworkFacts)
	}
	if fmt.Sprint(got.FrameworkFacts[0].Evidence) != fmt.Sprint([]string{"package.json"}) {
		t.Fatalf("framework evidence: %v", got.FrameworkFacts[0].Evidence)
	}
}
