package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeExecutable(t *testing.T, name, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake executables are Unix-only")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPlanSelectionAndFailClosed(t *testing.T) {
	// rg has apt/pacman/dnf; ast-grep only brew. Force PM via PATH.
	aptDir := fakeExecutable(t, "apt", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", aptDir)
	plan, err := Plan("rg")
	if err != nil || plan.PM != "apt" || plan.Executable != "apt" || plan.Elevation {
		t.Fatalf("plan: %+v %v", plan, err)
	}
	// ast-grep has no apt recipe -> fail closed.
	if _, err := Plan("ast-grep"); err == nil || !strings.Contains(err.Error(), "no verified apt recipe") {
		t.Fatalf("ast-grep apt must fail closed, got %v", err)
	}
	// Unknown tool -> no-recipe.
	if _, err := Plan("definitely-not-a-tool"); err == nil || !strings.Contains(err.Error(), "no verified install recipe") {
		t.Fatalf("unknown tool must fail closed, got %v", err)
	}
	// Empty PATH -> no PM.
	empty := t.TempDir()
	t.Setenv("PATH", empty)
	if _, err := Plan("rg"); err == nil || !strings.Contains(err.Error(), "no supported package manager") {
		t.Fatalf("no-PM must fail closed, got %v", err)
	}
}

func TestInstallExecutesRecipeAndVerifies(t *testing.T) {
	// Fake apt records args; fake rg appears only AFTER install.
	log := filepath.Join(t.TempDir(), "log")
	binDir := fakeExecutable(t, "apt", "#!/bin/sh\nprintf '%s\\n' \"$@\" >> "+log+"\n")
	rgDir := fakeExecutable(t, "rg", "#!/bin/sh\nprintf 'ripgrep 99\\n'\n")
	t.Setenv("PATH", binDir)
	plan, err := Plan("rg")
	if err != nil {
		t.Fatal(err)
	}
	// Tool absent now; install should run the recipe and then find rg.
	if present, _ := detect(entryByID(t, "rg")); present {
		t.Fatal("test precondition: rg must be absent before install")
	}
	// Add rg dir to PATH before install so verification passes.
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+rgDir)
	result, err := Install(plan)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !result.Executed || !result.Verified || !strings.Contains(result.Version, "ripgrep 99") {
		t.Fatalf("result: %+v", result)
	}
	data, _ := os.ReadFile(log)
	if !strings.Contains(string(data), "install\n-y\nripgrep\n") {
		t.Fatalf("recipe args not executed: %q", data)
	}
}

func TestInstallVerifyFailureFailsClosed(t *testing.T) {
	binDir := fakeExecutable(t, "apt", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir)
	plan, err := Plan("jq")
	if err != nil {
		t.Fatal(err)
	}
	// jq never appears on PATH -> verification fails closed.
	result, err := Install(plan)
	if err == nil || !strings.Contains(err.Error(), "not on PATH") {
		t.Fatalf("verify failure must fail closed, got %+v %v", result, err)
	}
	if !result.Executed || result.Verified {
		t.Fatalf("result must record executed-but-unverified: %+v", result)
	}
}

func TestInstallRecipeFailureFailsClosed(t *testing.T) {
	binDir := fakeExecutable(t, "apt", "#!/bin/sh\nexit 1\n")
	t.Setenv("PATH", binDir)
	plan, err := Plan("fd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Install(plan); err == nil || !strings.Contains(err.Error(), "install failed") {
		t.Fatalf("recipe failure must fail closed, got %v", err)
	}
}

func TestInstallSupportBackedByRecipe(t *testing.T) {
	// install: supported must stay backed by the embedded recipe contract;
	// install: unsupported tools must never become installable by implication.
	for _, entry := range Catalog() {
		hasRecipe := entry.Install != nil && len(entry.Install) > 0
		if entry.Capabilities.Install == Supported && !hasRecipe {
			t.Fatalf("%s claims install support without a recipe", entry.ID)
		}
		if hasRecipe && entry.Capabilities.Install != Supported {
			t.Fatalf("%s has a recipe but install is %s", entry.ID, entry.Capabilities.Install)
		}
	}
	aptDir := fakeExecutable(t, "apt", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", aptDir)
	for _, id := range []string{"go", "node", "codegraph", "context7", "rtk", "semble"} {
		if _, err := Plan(id); err == nil || !strings.Contains(err.Error(), "no verified install recipe") {
			t.Fatalf("install: unsupported %s must fail closed without a recipe, got %v", id, err)
		}
	}
}

func TestConfirmConsentNeverDefaultsYes(t *testing.T) {
	plan := InstallPlan{Tool: "rg", PM: "apt", Executable: "apt", Args: []string{"install", "-y", "ripgrep"}}
	for input, want := range map[string]bool{"n": false, "N": false, "": false, "no": false, "y": true, "Y": true, "yes": true} {
		got, err := ConfirmConsent(strings.NewReader(input+"\n"), plan)
		if err != nil || got != want {
			t.Fatalf("input %q: got %v %v, want %v", input, got, err, want)
		}
	}
}
