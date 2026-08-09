package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCatalogEmbeddedRecipes(t *testing.T) {
	recipes := Catalog()
	if len(recipes) != 5 {
		t.Fatalf("expected 5 recipes, got %d: %v", len(recipes), ids(recipes))
	}
	want := []string{"ast-grep", "fd", "gh", "jq", "rg"}
	for i, recipe := range recipes {
		if recipe.ID != want[i] {
			t.Fatalf("sorted catalog: got %s at %d, want %s", recipe.ID, i, want[i])
		}
		if len(recipe.Executables) == 0 || len(recipe.Install) == 0 {
			t.Fatalf("recipe %s incomplete: %+v", recipe.ID, recipe)
		}
	}
}

func ids(recipes []Recipe) []string {
	out := make([]string, len(recipes))
	for i, r := range recipes {
		out[i] = r.ID
	}
	return out
}

func TestValidateRecipeRejectsShellMeta(t *testing.T) {
	bad := Recipe{ID: "bad", Executables: []string{"x"}, Install: map[string]RecipeOp{"apt": {Executable: "apt", Args: []string{"install", "pkg; rm -rf /"}}}}
	if err := ValidateRecipe(bad); err == nil || !strings.Contains(err.Error(), "shell metacharacters") {
		t.Fatalf("expected shell-meta rejection, got %v", err)
	}
	ok := Recipe{ID: "ok", Executables: []string{"x"}, Install: map[string]RecipeOp{"apt": {Executable: "apt", Args: []string{"install", "pkg"}}}}
	if err := ValidateRecipe(ok); err != nil {
		t.Fatalf("valid recipe rejected: %v", err)
	}
}

func fakeBin(t *testing.T, name, version string) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Skip("fake executables are Unix-only")
	}
	script := "#!/bin/sh\nexit 0\n"
	if version != "" {
		script = "#!/bin/sh\nprintf '%s\\n' '" + version + "'\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDetectTool(t *testing.T) {
	rg := Recipe{ID: "rg", Executables: []string{"rg"}, VersionArgs: []string{"--version"}}
	dir := fakeBin(t, "rg", "ripgrep 14.1.0")
	t.Setenv("PATH", dir)
	present, version := detect(rg)
	if !present || !strings.Contains(version, "ripgrep 14.1.0") {
		t.Fatalf("detect rg: %v %q", present, version)
	}
	absent := Recipe{ID: "absent", Executables: []string{"definitely-not-a-tool"}}
	if present, _ := detect(absent); present {
		t.Fatal("absent tool reported present")
	}
}

func TestStatusFacts(t *testing.T) {
	dir := fakeBin(t, "jq", "jq-1.7.1")
	t.Setenv("PATH", dir)
	facts, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	if facts.SchemaVersion != SchemaVersion || facts.OS == "" {
		t.Fatalf("schema/os: %+v", facts)
	}
	jq, ok := facts.Tools["jq"]
	if !ok || !jq.Present || !strings.Contains(jq.Version, "jq-1.7.1") {
		t.Fatalf("jq facts: %+v %+v", ok, jq)
	}
	if rg, ok := facts.Tools["rg"]; ok && rg.Present {
		t.Fatalf("rg must be absent in fake PATH: %+v", rg)
	}
	first, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Tools) != len(facts.Tools) {
		t.Fatal("non-deterministic tool set")
	}
	for id, fact := range facts.Tools {
		if other := first.Tools[id]; other != fact {
			t.Fatalf("non-deterministic facts for %s: %+v vs %+v", id, fact, other)
		}
	}
}
