package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JhonMA82/agent-ready/internal/ecosystem"
	"github.com/JhonMA82/agent-ready/internal/inventory"
)

// Candidate is one capability-need candidate: capability need, candidate
// tool, observed evidence, and reason (spec §36 deterministic criteria;
// evidence only, never a verdict). Reason, catalog_id, and capabilities are
// additive; the V1 fields keep their types and meanings.
type Candidate struct {
	Capability   string        `json:"capability"`
	Candidate    string        `json:"candidate"`
	Signal       string        `json:"signal"`
	Observed     string        `json:"observed"`
	Reason       string        `json:"reason,omitempty"`
	CatalogID    string        `json:"catalog_id,omitempty"`
	Capabilities *Capabilities `json:"capabilities,omitempty"`
}

// RecommendFacts is the agent-ready.recommend/v1 schema; candidates are
// emitted in the documented signal-table order. An empty array means no
// grounded signal fired; no default verdict is synthesized.
type RecommendFacts struct {
	SchemaVersion string      `json:"schema_version"`
	Candidates    []Candidate `json:"candidates"`
}

// RecommendSchemaVersion is the agent-ready.recommend/v1 schema.
const RecommendSchemaVersion = "agent-ready.recommend/v1"

// Recommend emits candidate evidence from the documented §36 signal table
// over deterministic inventory facts; Slice 6 grounds signals in ecosystem
// facts (lockfiles, manifests, workspace signals, manager conflicts). Every
// candidate carries capability truth from the single catalog. Go never
// decides tool centrality, Tool Budget, installation, artifact suitability,
// or the final recommendation; those remain model-owned.
func Recommend(root string) (RecommendFacts, error) {
	facts, err := inventory.Inspect(root, "")
	if err != nil {
		return RecommendFacts{}, err
	}
	caps := make(map[string]Capabilities, len(Catalog()))
	for _, entry := range Catalog() {
		caps[entry.ID] = entry.Capabilities
	}
	candidates := []Candidate{}
	githubRemote := gitHubRemote(root)
	if hasOutputDirs(root) {
		candidates = append(candidates, candidate(caps, "token_optimized_shell", "RTK", "rtk", "build/test outputs are large", "output directories present", "large build/test outputs dominate shell output; RTK is a provider candidate without install/config/integration support"))
	}
	if githubRemote != "" {
		candidates = append(candidates, candidate(caps, "github_context", "gh", "gh", "GitHub remote", githubRemote, "the repository declares a GitHub remote; gh is detectable and provides GitHub context"))
	}
	if len(facts.WorkspaceSignals) > 0 || len(facts.Workspaces) > 1 || depCount(facts) > 40 {
		observed := "workspaces=" + strconv.Itoa(len(facts.Workspaces)) + " deps=" + strconv.Itoa(depCount(facts))
		if signals := joinPaths(facts.WorkspaceSignals, ""); signals != "" {
			observed += " workspace_signals=" + signals
		}
		candidates = append(candidates, candidate(caps, "dependency_graph", "CodeGraph", "codegraph", "workspaces or dense dependencies", observed, "workspace or dense-dependency evidence suggests a dependency graph; CodeGraph is a provider candidate without install/config/integration support"))
	}
	if paths := goEvidence(facts); paths != "" {
		candidates = append(candidates, candidate(caps, "go_toolchain", "go", "go", "go manifest or lockfile evidence", paths, "Go manifests declare the Go toolchain; go is detectable but has no verified install recipe"))
	}
	if paths := joinPaths(facts.Manifests, "javascript"); paths != "" {
		candidates = append(candidates, candidate(caps, "javascript_toolchain", "node", "node", "package.json manifest", paths, "package.json declares a JavaScript project; node is detectable but has no verified install recipe"))
	}
	if paths := joinPaths(facts.Lockfiles, ""); paths != "" {
		candidates = append(candidates, candidate(caps, "versioned_documentation", "Context7", "context7", "version-sensitive dependencies", paths, "version-sensitive dependencies are pinned by lockfiles; Context7 is a provider candidate without install/config/integration support"))
	}
	if len(facts.ManagerConflicts) > 0 {
		reasons := make([]string, 0, len(facts.ManagerConflicts))
		for _, conflict := range facts.ManagerConflicts {
			reasons = append(reasons, conflict.Reason)
		}
		candidates = append(candidates, candidate(caps, "package_manager_conflict", "Context7", "context7", "conflicting package-manager evidence", strings.Join(reasons, "; "), "conflicting package-manager lockfiles remain unresolved without choosing one"))
	}
	return RecommendFacts{SchemaVersion: RecommendSchemaVersion, Candidates: candidates}, nil
}

// candidate builds one evidence candidate with the catalog capability truth.
func candidate(caps map[string]Capabilities, capability, name, catalogID, signal, observed, reason string) Candidate {
	c := Candidate{Capability: capability, Candidate: name, Signal: signal, Observed: observed, Reason: reason, CatalogID: catalogID}
	if truth, ok := caps[catalogID]; ok {
		c.Capabilities = &truth
	}
	return c
}

// joinPaths returns the comma-joined paths of signals with the given id
// (all signals when id is empty); signals are already sorted.
func joinPaths(signals []ecosystem.Signal, id string) string {
	parts := []string{}
	for _, signal := range signals {
		if id == "" || signal.ID == id {
			parts = append(parts, signal.Path)
		}
	}
	return strings.Join(parts, ", ")
}

// goEvidence returns go ecosystem manifest and lockfile evidence paths.
func goEvidence(facts inventory.Facts) string {
	paths := joinPaths(facts.Manifests, "go")
	if lockfiles := joinPaths(facts.Lockfiles, "go"); lockfiles != "" {
		if paths != "" {
			paths += ", "
		}
		paths += lockfiles
	}
	return paths
}

func depCount(facts inventory.Facts) int {
	count := 0
	for _, dep := range facts.Deps {
		_ = dep
		count++
	}
	return count
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// hasOutputDirs reports build/test output directories commonly producing
// large shell output.
func hasOutputDirs(root string) bool {
	for _, dir := range []string{"dist", "build", "coverage"} {
		if info, err := os.Stat(filepath.Join(root, dir)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// gitHubRemote returns the GitHub remote URL when the repository has one.
func gitHubRemote(root string) string {
	out, err := exec.Command("git", "-C", root, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	remote := strings.TrimSpace(string(out))
	if strings.Contains(remote, "github.com") {
		return remote
	}
	return ""
}

// Summary renders the compact default output.
func (f RecommendFacts) Summary() string {
	if len(f.Candidates) == 0 {
		return "No capability candidates"
	}
	names := make([]string, 0, len(f.Candidates))
	for _, candidate := range f.Candidates {
		names = append(names, candidate.Candidate)
	}
	return "Candidates: " + strings.Join(names, ", ")
}
