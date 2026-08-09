package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Facts is the agent-ready.tools/v1 schema; tools keyed by recipe id.
type Facts struct {
	SchemaVersion  string              `json:"schema_version"`
	OS             string              `json:"os"`
	PackageManager string              `json:"package_manager,omitempty"`
	Tools          map[string]ToolFact `json:"tools"`
}

// ToolFact is one tool's presence/version fact; version is omitted (null)
// when the recipe cannot verify it.
type ToolFact struct {
	RecipeID string `json:"recipe_id"`
	Present  bool   `json:"present"`
	Version  string `json:"version,omitempty"`
}

// SchemaVersion is the agent-ready.tools/v1 schema.
const SchemaVersion = "agent-ready.tools/v1"

// DetectPackageManager finds the first available package manager binary.
func DetectPackageManager() string {
	for _, pm := range []string{"apt", "pacman", "dnf", "brew"} {
		if path, err := exec.LookPath(pm); err == nil && path != "" {
			return pm
		}
	}
	return ""
}

// Status collects deterministic facts for every catalog tool. Version is
// best-effort per-recipe version_args; any failure yields no version.
func Status() (Facts, error) {
	facts := Facts{SchemaVersion: SchemaVersion, OS: runtime.GOOS, Tools: map[string]ToolFact{}}
	facts.PackageManager = DetectPackageManager()
	for _, recipe := range Catalog() {
		present, version := detect(recipe)
		facts.Tools[recipe.ID] = ToolFact{RecipeID: recipe.ID, Present: present, Version: version}
	}
	return facts, nil
}

// detect looks up the recipe executables and best-effort parses a version.
func detect(recipe Recipe) (present bool, version string) {
	for _, name := range recipe.Executables {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if len(recipe.VersionArgs) == 0 {
			return true, ""
		}
		out, err := exec.Command(path, recipe.VersionArgs...).Output()
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
