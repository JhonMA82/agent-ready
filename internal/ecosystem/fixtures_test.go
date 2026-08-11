package ecosystem

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFixtureMatrix: deterministic §33–§42 table; exact managers/lockfiles,
func TestFixtureMatrix(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		man   string // "id=confidence,id=confidence" (exact)
		lock  string // "id:path,id:path" ("" = none)
		eco   string // ecosystem id whose manifest evidence must fire
	}{
		{"§33 laravel", []string{"composer.json", "composer.lock", "artisan", "package.json", "bun.lock"}, "bun=confirmed,composer=confirmed", "bun:bun.lock,composer:composer.lock", "javascript"},
		{"§34 uv", []string{"pyproject.toml", "uv.lock", "src/main.py"}, "uv=confirmed", "uv:uv.lock", "python"},
		{"§35 pip", []string{"pyproject.toml", "requirements.txt"}, "pip=inferred", "", "python"},
		{"§36 npm", []string{"package.json", "package-lock.json"}, "npm=confirmed", "npm:package-lock.json", "javascript"},
		{"§36 pnpm workspace", []string{"package.json", "pnpm-workspace.yaml", "pnpm-lock.yaml"}, "pnpm=confirmed", "pnpm:pnpm-lock.yaml", "javascript"},
		{"§36 bun", []string{"package.json", "bun.lock"}, "bun=confirmed", "bun:bun.lock", "javascript"},
		{"§36 deno", []string{"deno.json", "deno.lock"}, "deno=confirmed", "deno:deno.lock", "javascript"},
		{"§37 gradle wrapper", []string{"gradlew", "build.gradle.kts"}, "gradle=confirmed", "", ""},
		{"§37 maven wrapper", []string{"mvnw", "pom.xml"}, "maven=confirmed", "", ""},
		{"§38 dotnet", []string{"App.sln", "src/App.csproj", "packages.lock.json"}, "nuget=confirmed", "nuget:packages.lock.json", "dotnet"},
		{"§39 ruby", []string{"Gemfile", "Gemfile.lock", "bin/rails"}, "bundler=confirmed", "bundler:Gemfile.lock", ""},
		{"§40 elixir", []string{"mix.exs", "mix.lock"}, "mix=confirmed", "mix:mix.lock", ""},
		{"§41 flutter", []string{"pubspec.yaml", "pubspec.lock"}, "pub=confirmed", "pub:pubspec.lock", ""},
		{"§42 cmake conan", []string{"CMakeLists.txt", "CMakeUserPresets.json", "conanfile.py", "conan.lock"}, "cmake=inferred,conan=confirmed", "conan:conan.lock", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect("", tt.paths)
			raw, _ := json.Marshal(got)
			if got := managersOf(got); got != tt.man {
				t.Fatalf("managers: got %s, want %s", got, tt.man)
			}
			if got := lockfilesOf(got); got != tt.lock {
				t.Fatalf("lockfiles: got %s, want %s", got, tt.lock)
			}
			if tt.eco != "" && !strings.Contains(string(raw), `"id":"`+tt.eco+`"`) {
				t.Fatalf("ecosystem %s must fire: %s", tt.eco, raw)
			}
		})
	}
}
func managersOf(f Facts) string {
	return candidatesRender(f.PackageManagers)
}

func candidatesRender(cs []ManagerCandidate) string {
	parts := []string{}
	for _, c := range cs {
		parts = append(parts, c.ID+"="+c.Confidence)
	}
	return strings.Join(parts, ",")
}
func lockfilesOf(f Facts) string {
	parts := []string{}
	for _, s := range f.Lockfiles {
		parts = append(parts, s.ID+":"+s.Path)
	}
	return strings.Join(parts, ",")
}

// TestFixtureAcceptance: §52 lockfile recognition, §59 retained pnpm/Bun
func TestFixtureAcceptance(t *testing.T) {
	got := Detect("", []string{"Cargo.lock", "flake.lock", "composer.lock", "Gemfile.lock", "pubspec.lock", "mix.lock"})
	want := "bundler:Gemfile.lock,cargo:Cargo.lock,composer:composer.lock,mix:mix.lock,nix:flake.lock,pub:pubspec.lock"
	if got := lockfilesOf(got); got != want {
		t.Fatalf("§52 lockfiles: got %s, want %s", got, want)
	}
	candidates, conflicts := resolveManagers([]string{"pnpm-lock.yaml", "bun.lock", "package.json"})
	if got := candidatesRender(candidates); got != "bun=confirmed,pnpm=confirmed" {
		t.Fatalf("§59 candidates: got %s", got)
	}
	if len(conflicts) != 1 || strings.Join(conflicts[0].Managers, ",") != "bun,pnpm" {
		t.Fatalf("§59 must retain one bun/pnpm conflict, got %v", conflicts)
	}
	if !strings.Contains(conflicts[0].Reason, "bun") || !strings.Contains(conflicts[0].Reason, "pnpm") {
		t.Fatalf("§59 conflict reason must cite both candidates: %q", conflicts[0].Reason)
	}
	for _, forbidden := range []string{"preferred", "migration", "recommend", "selected", "primary", "choose"} {
		if strings.Contains(conflicts[0].Reason, forbidden) {
			t.Fatalf("§59 conflict reason contains decision token %q", forbidden)
		}
	}
	candidates, conflicts = resolveManagers([]string{"gradlew", "mvnw", "build.gradle", "pom.xml"})
	for _, c := range candidates {
		if (c.ID == "gradle" || c.ID == "maven") && c.Confidence != ConfidenceConfirmed {
			t.Fatalf("§60 wrapper must confirm %s: %v", c.ID, candidates)
		}
	}
	if len(conflicts) != 1 || conflicts[0].Reason != "project wrappers evidence distinct managers: gradle and maven" {
		t.Fatalf("§60 wrapper conflict: %v", conflicts)
	}
	mixed := []struct {
		name  string
		paths []string
		lock  string
	}{
		{"rust cargo nix", []string{"Cargo.toml", "Cargo.lock", "flake.nix", "flake.lock"}, "cargo:Cargo.lock,nix:flake.lock"},
		{"php composer js", []string{"composer.json", "composer.lock", "package.json", "bun.lock"}, "bun:bun.lock,composer:composer.lock"},
		{"dart native", []string{"pubspec.yaml", "pubspec.lock", "CMakeLists.txt", "conanfile.py"}, "pub:pubspec.lock"},
	}
	for _, tt := range mixed {
		got := Detect("", tt.paths)
		raw, _ := json.Marshal(got)
		if got := lockfilesOf(got); got != tt.lock {
			t.Fatalf("mixed lockfiles: got %s, want %s", got, tt.lock)
		}
		for _, forbidden := range []string{"primary", "exclusive", "single", "selected", "preferred"} {
			if strings.Contains(string(raw), forbidden) {
				t.Fatalf("mixed facts must not carry an exclusive label (%q)", forbidden)
			}
		}
	}
}
