package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEmbeddedRecipesUnchanged(t *testing.T) {
	recipes, err := loadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recipes) != 5 {
		t.Fatalf("expected 5 recipes, got %d: %v", len(recipes), recipeIDs())
	}
	want := []string{"ast-grep", "fd", "gh", "jq", "rg"}
	for i, recipe := range recipes {
		if recipe.ID != want[i] {
			t.Fatalf("sorted recipes: got %s at %d, want %s", recipe.ID, i, want[i])
		}
		if len(recipe.Executables) == 0 || len(recipe.Install) == 0 {
			t.Fatalf("recipe %s incomplete: %+v", recipe.ID, recipe)
		}
	}
}

func TestCatalogOrderedSupportTruth(t *testing.T) {
	catalog := Catalog()
	if len(catalog) != 17 {
		t.Fatalf("expected 17 catalog entries, got %d", len(catalog))
	}
	for i := 1; i < len(catalog); i++ {
		if catalog[i-1].ID >= catalog[i].ID {
			t.Fatalf("catalog not sorted by stable identifier: %s before %s", catalog[i-1].ID, catalog[i].ID)
		}
	}
	for _, entry := range catalog {
		hasRecipe := entry.Install != nil && len(entry.Install) > 0
		if (entry.Capabilities.Install == Supported) != hasRecipe {
			t.Fatalf("%s install truth must match recipe presence: %+v", entry.ID, entry.Capabilities)
		}
		if entry.Capabilities.Detect == Supported && len(entry.Executables) == 0 {
			t.Fatalf("%s claims detect support without executables", entry.ID)
		}
		// §20: every entry with install support declares a safety level.
		if hasRecipe && entry.SafetyLevel == "" {
			t.Fatalf("%s has a recipe but no safety level", entry.ID)
		}
		// OQ-1: recipe ops execute the PM binary deterministically.
		for pm, op := range entry.Install {
			if op.Executable != pm || len(op.Args) == 0 {
				t.Fatalf("%s %s recipe must execute %s deterministically with fixed args, got %+v", entry.ID, pm, pm, op)
			}
		}
	}
}

func TestCapabilityStatesDistinguishable(t *testing.T) {
	all := Capabilities{Detect: Supported, Version: Unsupported, Recommend: Unknown, Install: Unsupported, Configure: Unsupported, Integration: Unsupported, SideEffects: Unsupported}
	data, err := json.Marshal(all)
	if err != nil {
		t.Fatal(err)
	}
	// All seven fields serialize in fixed order with distinct state values.
	const want = `{"detect":"supported","version":"unsupported","recommend":"unknown","install":"unsupported","configure":"unsupported","integration":"unsupported","side_effects":"unsupported"}`
	if string(data) != want {
		t.Fatalf("capability JSON: got %s, want %s", data, want)
	}
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

// Threat-matrix RED: sh/bash executables and pipe args are rejected.
func TestValidateRecipeRejectsShellAndPipe(t *testing.T) {
	for _, tt := range []struct {
		exe  string
		args []string
		want string
	}{
		{"sh", []string{"-c", "curl -s https://x | sh"}, "shell interpreter"},
		{"bash", []string{"-c", "curl -s https://x | sh"}, "shell interpreter"},
		{"curl", []string{"-sSL", "https://example.com/install.sh", "|", "sh"}, "shell metacharacters"},
	} {
		recipe := Recipe{ID: "bad", Executables: []string{"x"}, Install: map[string]RecipeOp{"apt": {Executable: tt.exe, Args: tt.args}}}
		if err := ValidateRecipe(recipe); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("%s recipe must be rejected, got %v", tt.exe, err)
		}
	}
}

// D7: explain renders declared facts for a known tool and fails naming the
// id for an unknown one, without executing or installing anything.
func TestExplain(t *testing.T) {
	facts, err := Explain("uv")
	if err != nil {
		t.Fatal(err)
	}
	if facts.SchemaVersion != ExplainSchemaVersion || facts.ID != "uv" || facts.Kind != "ecosystem" {
		t.Fatalf("explain facts: %+v", facts)
	}
	if facts.SafetyLevel != SafetySafeRecipe || facts.Capabilities.Detect != Supported || facts.Capabilities.Install != Unsupported {
		t.Fatalf("uv declared facts: %+v", facts)
	}
	if s := facts.Summary(); !strings.Contains(s, "safety_level=SAFE_RECIPE") || !strings.Contains(s, "detect=supported") {
		t.Fatalf("summary must render safety and capability facts: %s", s)
	}
	facts, err = Explain("definitely-not-a-tool")
	if err == nil || !strings.Contains(err.Error(), "definitely-not-a-tool") || !strings.Contains(err.Error(), "not in the catalog") {
		t.Fatalf("unknown tool must fail naming the id, got %+v %v", facts, err)
	}
}

// D10: §21 managers detect in fixed order; zypper/apk/winget hosts detect
// their manager; AUR helpers and nix are never auto-selected.
func TestDetectPackageManagerOrderAndNewManagers(t *testing.T) {
	if got := strings.Join(pmOrder, ","); got != "apt,pacman,dnf,brew,zypper,apk,winget" {
		t.Fatalf("pm order must be pinned: %s", got)
	}
	for _, pm := range []string{"zypper", "apk", "winget"} {
		t.Setenv("PATH", fakeBin(t, pm, ""))
		if got := DetectPackageManager(); got != pm {
			t.Fatalf("%s host must detect %s, got %q", pm, pm, got)
		}
	}
	// Earlier managers win even when later ones appear earlier in PATH.
	t.Setenv("PATH", fakeBin(t, "winget", "")+string(os.PathListSeparator)+fakeBin(t, "apt", ""))
	if got := DetectPackageManager(); got != "apt" {
		t.Fatalf("with apt+winget, want apt, got %q", got)
	}
	// AUR helpers and nix are never selected, even alone.
	for _, name := range []string{"yay", "paru", "nix"} {
		t.Setenv("PATH", fakeBin(t, name, ""))
		if got := DetectPackageManager(); got != "" {
			t.Fatalf("%s must never be auto-selected, got %q", name, got)
		}
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

// fakeBins writes one fake executable per name and returns a PATH covering
// all of them.
func fakeBins(t *testing.T, versions map[string]string) string {
	t.Helper()
	paths := make([]string, 0, len(versions))
	for name, version := range versions {
		paths = append(paths, fakeBin(t, name, version))
	}
	return strings.Join(paths, string(os.PathListSeparator))
}

func entryByID(t *testing.T, id string) Entry {
	t.Helper()
	for _, entry := range Catalog() {
		if entry.ID == id {
			return entry
		}
	}
	t.Fatalf("catalog entry %q not found", id)
	return Entry{}
}

func TestDetectTool(t *testing.T) {
	rg := Entry{ID: "rg", Executables: []string{"rg"}, VersionArgs: []string{"--version"}}
	dir := fakeBin(t, "rg", "ripgrep 14.1.0")
	t.Setenv("PATH", dir)
	present, version := detect(rg)
	if !present || !strings.Contains(version, "ripgrep 14.1.0") {
		t.Fatalf("detect rg: %v %q", present, version)
	}
	absent := Entry{ID: "absent", Executables: []string{"definitely-not-a-tool"}}
	if present, _ := detect(absent); present {
		t.Fatal("absent tool reported present")
	}
}

func TestDetectEcosystemTools(t *testing.T) {
	t.Setenv("PATH", fakeBins(t, map[string]string{
		"go":   "go version go1.24.1 linux/amd64",
		"node": "v22.11.0",
	}))
	for id, want := range map[string]string{"go": "go version go1.24.1 linux/amd64", "node": "v22.11.0"} {
		entry := entryByID(t, id)
		present, version := detect(entry)
		if !present || !strings.Contains(version, want) {
			t.Fatalf("detect %s: %v %q, want %q", id, present, version, want)
		}
	}
	// Provider entries have no executable contract: never present.
	for _, id := range []string{"codegraph", "context7", "rtk", "semble"} {
		entry := entryByID(t, id)
		if present, _ := detect(entry); present {
			t.Fatalf("provider %s must never be detected", id)
		}
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

func TestStatusFamiliesOrderedAndStable(t *testing.T) {
	t.Setenv("PATH", fakeBins(t, map[string]string{
		"ast-grep": "ast-grep 0.21.0",
		"fd":       "fd 9.0.0",
		"gh":       "gh version 2.45.0",
		"go":       "go version go1.24.1 linux/amd64",
		"jq":       "jq-1.7.1",
		"node":     "v22.11.0",
		"rg":       "ripgrep 14.1.0",
	}))
	first, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	// Golden shape: fixed-order families with ID-sorted tools, presence
	// evidence, and all seven capability states.
	order := []Family{FamilyEcosystem, FamilyProductivity, FamilyProvider}
	want := [][]string{{"composer", "gh", "go", "gradle", "maven", "node", "pip", "rustup", "uv"}, {"ast-grep", "fd", "jq", "rg"}, {"codegraph", "context7", "rtk", "semble"}}
	if len(first.Families) != 3 {
		t.Fatalf("expected 3 families, got %d", len(first.Families))
	}
	for i, family := range order {
		if first.Families[i].ID != family {
			t.Fatalf("family %d: got %s, want %s", i, first.Families[i].ID, family)
		}
		for j, id := range want[i] {
			if first.Families[i].Tools[j].ID != id {
				t.Fatalf("family %d tool %d: got %s, want %s", i, j, first.Families[i].Tools[j].ID, id)
			}
		}
	}
	eco, prod, prov := first.Families[0].Tools, first.Families[1].Tools, first.Families[2].Tools
	// Detect-only ecosystem tools carry presence evidence yet stay
	// uninstallable; providers never present and carry no lifecycle support.
	if goTool := eco[2]; !goTool.Present || !strings.Contains(goTool.Version, "go version go1.24.1") || goTool.Capabilities.Install != Unsupported {
		t.Fatalf("go truth: %+v", goTool)
	}
	// §20 levels surface in status --json with recipe methods.
	if rgTool := prod[3]; rgTool.Capabilities.Install != Supported || rgTool.SafetyLevel != SafetySafeRecipe || strings.Join(rgTool.Methods, ",") != "apt,dnf,pacman" {
		t.Fatalf("rg truth: %+v", rgTool)
	}
	if !strings.Contains(string(data), `"safety_level":"SAFE_RECIPE"`) {
		t.Fatalf("status JSON must surface the safety level: %s", data)
	}
	// Declared detect-only entries carry their level too.
	if uvTool := eco[8]; uvTool.SafetyLevel != SafetySafeRecipe || uvTool.Capabilities.Install != Unsupported {
		t.Fatalf("uv safety metadata: %+v", uvTool)
	}
	// RTK splits binary-install safety from its separate global side effect.
	if rtkTool := prov[2]; rtkTool.SafetyLevel != SafetySafeRecipe || rtkTool.SideEffects != "GLOBAL_SIDE_EFFECT" || rtkTool.IntegrationMode != "opt-in" {
		t.Fatalf("rtk split metadata: %+v", rtkTool)
	}
	for _, tool := range first.Families[2].Tools {
		if tool.Present || tool.Capabilities.Install != Unsupported || tool.Capabilities.Configure != Unsupported || tool.Capabilities.Integration != Unsupported || tool.Capabilities.SideEffects != Unsupported {
			t.Fatalf("provider truth: %+v", tool)
		}
	}
	// Byte-stable across runs.
	second, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(second)
	if string(data) != string(b) {
		t.Fatal("status JSON must be byte-stable")
	}
}

func TestV1ReaderCompatibility(t *testing.T) {
	dir := fakeBin(t, "jq", "jq-1.7.1")
	t.Setenv("PATH", dir)
	facts, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	// A reader of only the original agent-ready.tools/v1 fields must keep
	// their types and meanings; additive families are never required.
	var v1 struct {
		SchemaVersion string `json:"schema_version"`
		OS            string `json:"os"`
		Tools         map[string]struct {
			RecipeID string `json:"recipe_id"`
			Present  bool   `json:"present"`
			Version  string `json:"version"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(data, &v1); err != nil {
		t.Fatalf("V1 reader must accept enriched output: %v", err)
	}
	if v1.SchemaVersion != SchemaVersion || v1.OS == "" {
		t.Fatalf("V1 header lost: %+v", v1)
	}
	if len(v1.Tools) != 5 {
		t.Fatalf("V1 tools map must stay recipe-keyed (5), got %d", len(v1.Tools))
	}
	jq := v1.Tools["jq"]
	if jq.RecipeID != "jq" || !jq.Present || !strings.Contains(jq.Version, "jq-1.7.1") {
		t.Fatalf("V1 jq facts: %+v", jq)
	}
}
