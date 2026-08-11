package tools

import (
	"bytes"
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
	if _, err := Plan("rg"); err == nil || !strings.Contains(err.Error(), "no supported package manager") || !strings.Contains(err.Error(), "remediation") {
		t.Fatalf("no-PM must fail closed with remediation, got %v", err)
	}
	// §21/OQ-1 fail-closed: AUR opt-in only; nix environment-only; nothing
	// executes.
	for bin, want := range map[string]string{"yay": "opt-in", "paru": "opt-in", "nix": "environment only"} {
		t.Setenv("PATH", fakeExecutable(t, bin, "#!/bin/sh\nexit 0\n"))
		if _, err := Plan("rg"); err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "remediation") {
			t.Fatalf("%s-only host must fail closed with %q + remediation, got %v", bin, want, err)
		}
	}
	// 2.7: rtk has a brew-only deterministic recipe; other PM hosts fail
	// closed with remediation and nothing executes.
	t.Setenv("PATH", fakeExecutable(t, "apt", "#!/bin/sh\nexit 0\n"))
	if _, err := Plan("rtk"); err == nil || !strings.Contains(err.Error(), `no verified apt recipe for "rtk"; supported managers: brew`) {
		t.Fatalf("rtk on apt must fail closed naming brew as the only manager, got %v", err)
	}
}

// §56 honesty: every recipe-backed entry added in 2b (rtk, uv, composer) is
// proven by the five behaviors — detect, version, plan, execute, verify —
// with PATH-faking and an isolated HOME so no global configuration can be
// touched by the flow.
func TestRecipeFiveBehaviors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	for _, tc := range []struct {
		id, pm, version string
		args            []string
	}{
		{id: "uv", pm: "apt", version: "uv 0.7.2", args: []string{"install", "-y", "uv"}},
		{id: "rtk", pm: "brew", version: "rtk 0.28.2", args: []string{"install", "rtk"}},
		{id: "composer", pm: "apt", version: "Composer version 2.8.0", args: []string{"install", "-y", "composer"}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			entry := entryByID(t, tc.id)
			if entry.SafetyLevel == "" {
				t.Fatalf("%s recipe must declare its safety level", tc.id)
			}
			// 1+2. detect and version via a fake executable on PATH.
			t.Setenv("PATH", fakeBin(t, tc.id, tc.version))
			present, version := detect(entry)
			if !present || version != tc.version {
				t.Fatalf("%s detect/version: %v %q, want %q", tc.id, present, version, tc.version)
			}
			// 3. plan selects the deterministic recipe op for the host PM.
			log := filepath.Join(t.TempDir(), "log")
			pmDir := fakeExecutable(t, tc.pm, "#!/bin/sh\nprintf '%s\\n' \"$@\" >> "+log+"\n")
			t.Setenv("PATH", pmDir)
			plan, err := Plan(tc.id)
			if err != nil {
				t.Fatal(err)
			}
			if plan.PM != tc.pm || plan.Executable != tc.pm || plan.Level != entry.SafetyLevel || strings.Join(plan.Args, " ") != strings.Join(tc.args, " ") {
				t.Fatalf("%s plan: %+v", tc.id, plan)
			}
			// 4+5. execute runs the fixed argv; verify finds the tool after.
			t.Setenv("PATH", pmDir+string(os.PathListSeparator)+fakeBin(t, tc.id, tc.version))
			result, err := Install(plan)
			if err != nil {
				t.Fatalf("%s install: %v", tc.id, err)
			}
			if !result.Executed || !result.Verified || !strings.Contains(result.Version, tc.version) {
				t.Fatalf("%s result: %+v", tc.id, result)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(string(data)); got != strings.Join(tc.args, "\n") {
				t.Fatalf("%s executed args %q, want %q", tc.id, got, strings.Join(tc.args, "\n"))
			}
		})
	}
}

// §47/§57/§58: the rtk binary install never touches global OpenCode
// configuration and never runs the OpenCode integration (opt-in only, never
// during install); the isolated HOME stays free of any opencode config.
func TestRtkInstallIsolatedFromGlobalOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	log := filepath.Join(t.TempDir(), "log")
	brewDir := fakeExecutable(t, "brew", "#!/bin/sh\nprintf '%s\\n' \"$@\" >> "+log+"\n")
	rtkDir := fakeBin(t, "rtk", "rtk 0.28.2")
	t.Setenv("PATH", brewDir)
	plan, err := Plan("rtk")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", brewDir+string(os.PathListSeparator)+rtkDir)
	result, err := Install(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Executed || !result.Verified {
		t.Fatalf("result: %+v", result)
	}
	// Only the brew install op ran: no rtk init/-g/--opencode integration.
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "install\nrtk" {
		t.Fatalf("install must run only `brew install rtk`, got %q", got)
	}
	// The isolated HOME contains no global OpenCode configuration.
	err = filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.Contains(strings.ToLower(d.Name()), "opencode") {
			t.Fatalf("install created global config at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// 2.8: composer detect/version/explain/plan across the V1 platforms;
// execution only where a deterministic recipe exists — other hosts fail
// closed and nothing executes.
func TestComposerLifecycleAcrossPlatforms(t *testing.T) {
	// detect + version.
	t.Setenv("PATH", fakeBin(t, "composer", "Composer version 2.8.0"))
	entry := entryByID(t, "composer")
	present, version := detect(entry)
	if !present || !strings.Contains(version, "Composer version 2.8.0") {
		t.Fatalf("composer detect/version: %v %q", present, version)
	}
	// explain renders the declared lifecycle facts.
	facts, err := Explain("composer")
	if err != nil {
		t.Fatal(err)
	}
	if facts.SafetyLevel != SafetyVersionSensitive || facts.Capabilities.Detect != Supported || facts.Capabilities.Install != Supported {
		t.Fatalf("composer explain: %+v", facts)
	}
	// plan succeeds on every PM with a deterministic recipe.
	for _, pm := range []string{"apt", "pacman", "dnf", "brew"} {
		t.Setenv("PATH", fakeExecutable(t, pm, "#!/bin/sh\nexit 0\n"))
		plan, err := Plan("composer")
		if err != nil {
			t.Fatalf("composer plan on %s: %v", pm, err)
		}
		if plan.PM != pm || plan.Executable != pm || plan.Level != SafetyVersionSensitive {
			t.Fatalf("composer plan on %s: %+v", pm, plan)
		}
	}
	// Fail closed where no deterministic recipe exists; nothing executes.
	for _, pm := range []string{"zypper", "apk", "winget"} {
		t.Setenv("PATH", fakeExecutable(t, pm, "#!/bin/sh\nexit 0\n"))
		if _, err := Plan("composer"); err == nil || !strings.Contains(err.Error(), "no verified "+pm+" recipe") {
			t.Fatalf("composer on %s must fail closed, got %v", pm, err)
		}
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
	for _, id := range []string{"go", "node", "npm", "terraform", "tofu", "cargo", "codegraph", "context7", "semble"} {
		if _, err := Plan(id); err == nil || !strings.Contains(err.Error(), "no verified install recipe") {
			t.Fatalf("install: unsupported %s must fail closed without a recipe, got %v", id, err)
		}
	}
}

func TestConfirmConsentNeverDefaultsYes(t *testing.T) {
	plan := InstallPlan{Tool: "rg", PM: "apt", Executable: "apt", Args: []string{"install", "-y", "ripgrep"}}
	// §46: empty or unreadable input declines via the read-error path.
	for input, want := range map[string]bool{"n": false, "N": false, "": false, "no": false, "y": true, "Y": true, "yes": true} {
		got, err := ConfirmConsent(strings.NewReader(input+"\n"), plan)
		if err != nil || got != want {
			t.Fatalf("input %q: got %v %v, want %v", input, got, err, want)
		}
	}
}

// §46 UX golden: the complete plan renders before the consent prompt.
func TestRenderPlanUXGolden(t *testing.T) {
	plan := InstallPlan{Tool: "uv", Kind: "ecosystem", Evidence: "verified embedded recipe", PM: "apt",
		Method: "verified recipe", Level: SafetySafeRecipe, Executable: "apt", Args: []string{"install", "-y", "uv"}}
	var out bytes.Buffer
	RenderPlan(&out, plan)
	want := strings.Join([]string{
		"Tool: uv", "Kind: ecosystem", "Evidence: verified embedded recipe", "Safety level: SAFE_RECIPE",
		"", "Plan", "  platform: " + runtime.GOOS, "  method: verified recipe", "  executable: apt", "  args: install -y uv",
		"", "Changes", "  installs user-level/global executable", "  does NOT modify OpenCode",
		"  does NOT modify project dependencies", "", "Proceed? [y/N] ",
	}, "\n")
	if out.String() != want {
		t.Fatalf("plan render mismatch:\ngot:\n%s\nwant:\n%s", out.String(), want)
	}
}
