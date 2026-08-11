package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// Facts is the agent-ready.tools/v1 schema; tools keyed by recipe id.
// Families is additive support truth: optional, never required to read
// presence or version.
type Facts struct {
	SchemaVersion  string              `json:"schema_version"`
	OS             string              `json:"os"`
	PackageManager string              `json:"package_manager,omitempty"`
	Tools          map[string]ToolFact `json:"tools"`
	Families       []ToolFamily        `json:"families,omitempty"`
}

// ToolFact is one tool's presence/version fact; version is omitted (null)
// when the recipe cannot verify it.
type ToolFact struct {
	RecipeID string `json:"recipe_id"`
	Present  bool   `json:"present"`
	Version  string `json:"version,omitempty"`
}

// ToolFamily is one ordered family with per-tool capability truth; families
// serialize in fixed order and tools sort by stable identifier.
type ToolFamily struct {
	ID    Family       `json:"id"`
	Tools []FamilyTool `json:"tools"`
}

// FamilyTool is one tool's presence evidence, capability truth, and §20
// additive safety metadata (optional, never required).
type FamilyTool struct {
	ID              string       `json:"id"`
	Present         bool         `json:"present"`
	Version         string       `json:"version,omitempty"`
	Capabilities    Capabilities `json:"capabilities"`
	SafetyLevel     SafetyLevel  `json:"safety_level,omitempty"`
	Methods         []string     `json:"methods,omitempty"`
	SideEffects     string       `json:"side_effects,omitempty"`
	IntegrationMode string       `json:"integration_mode,omitempty"`
}

// SchemaVersion is the agent-ready.tools/v1 schema.
const SchemaVersion = "agent-ready.tools/v1"

// pmOrder is the §21 detection order (D10); AUR helpers and nix are absent:
// AUR is opt-in only, nix is an environment, never an automatic installer.
var pmOrder = []string{"apt", "pacman", "dnf", "brew", "zypper", "apk", "winget"}

// aurHelpers are AUR package helpers; opt-in only, never auto-selected.
var aurHelpers = []string{"yay", "paru"}

// DetectPackageManager finds the first available package manager binary in
// fixed §21 order.
func DetectPackageManager() string {
	for _, pm := range pmOrder {
		if path, err := exec.LookPath(pm); err == nil && path != "" {
			return pm
		}
	}
	return ""
}

// detectAUR reports the first available AUR helper for opt-in-only remediation.
func detectAUR() string {
	for _, name := range aurHelpers {
		if path, err := exec.LookPath(name); err == nil && path != "" {
			return name
		}
	}
	return ""
}

// nixPresent reports whether nix is available as an environment.
func nixPresent() bool {
	path, err := exec.LookPath("nix")
	return err == nil && path != ""
}

// Status collects deterministic facts for every catalog tool. The V1 tools
// map keeps its recipe-keyed presence/version meaning; families adds ordered
// capability truth for every entry. Version is best-effort; failures yield none.
func Status() (Facts, error) {
	facts := Facts{SchemaVersion: SchemaVersion, OS: runtime.GOOS, Tools: map[string]ToolFact{}}
	facts.PackageManager = DetectPackageManager()
	byFamily := map[Family][]FamilyTool{}
	for _, entry := range Catalog() {
		present, version := detect(entry)
		if entry.Install != nil {
			facts.Tools[entry.ID] = ToolFact{RecipeID: entry.ID, Present: present, Version: version}
		}
		byFamily[entry.Family] = append(byFamily[entry.Family], FamilyTool{
			ID: entry.ID, Present: present, Version: version, Capabilities: entry.Capabilities,
			SafetyLevel: entry.SafetyLevel, Methods: entry.Methods, SideEffects: entry.SideEffects,
			IntegrationMode: entry.IntegrationMode,
		})
	}
	for _, family := range []Family{FamilyEcosystem, FamilyProductivity, FamilyProvider} {
		tools := byFamily[family]
		sort.Slice(tools, func(i, j int) bool { return tools[i].ID < tools[j].ID })
		facts.Families = append(facts.Families, ToolFamily{ID: family, Tools: tools})
	}
	return facts, nil
}

// detect looks up the entry executables and best-effort parses a version.
func detect(entry Entry) (present bool, version string) {
	for _, name := range entry.Executables {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if len(entry.VersionArgs) == 0 {
			return true, ""
		}
		out, err := exec.Command(path, entry.VersionArgs...).Output()
		if err != nil {
			return true, ""
		}
		line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
		return true, line
	}
	return false, ""
}

// Summary renders the compact default output.
func (f Facts) Summary() string {
	present := 0
	for _, tool := range f.Tools {
		if tool.Present {
			present++
		}
	}
	return fmt.Sprintf("Tools: %d/%d present (PM: %s)", present, len(f.Tools), f.PackageManager)
}
