package tools

import (
	"fmt"
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
		candidates = append(candidates, candidate(caps, "token_optimized_shell", "RTK", "rtk", "build/test outputs are large", "output directories present", "large build/test outputs dominate shell output; RTK is a productivity candidate whose binary install is consent-gated and whose OpenCode global integration is a separate opt-in"))
	}
	if githubRemote != "" {
		candidates = append(candidates, candidate(caps, "github_context", "gh", "gh", "GitHub remote", githubRemote, "the repository declares a GitHub remote; gh is detectable and provides GitHub context"))
	}
	if len(facts.WorkspaceSignals) > 0 || len(facts.Workspaces) > 1 {
		observed := "workspaces=" + strconv.Itoa(len(facts.Workspaces)) + " deps=" + strconv.Itoa(depCount(facts))
		if signals := joinPaths(facts.WorkspaceSignals, ""); signals != "" {
			observed += " workspace_signals=" + signals
		}
		candidates = append(candidates, candidate(caps, "dependency_graph", "CodeGraph", "codegraph", "workspace or cross-package topology", observed, "workspace or cross-package topology evidence suggests a dependency graph; CodeGraph is a provider candidate without install/config/integration support"))
	}
	if paths := goEvidence(facts); paths != "" {
		candidates = append(candidates, candidate(caps, "go_toolchain", "go", "go", "go manifest or lockfile evidence", paths, "Go manifests declare the Go toolchain; go is detectable but has no verified install recipe"))
	}
	if paths := joinPaths(facts.Manifests, "javascript"); paths != "" {
		candidates = append(candidates, candidate(caps, "javascript_toolchain", "node", "node", "package.json manifest", paths, "package.json declares a JavaScript project; node is detectable but has no verified install recipe"))
	}
	// D3: Context7 fires only on central version-sensitive framework evidence;
	// lockfile presence or manager conflicts alone never generate a candidate.
	if paths := frameworkEvidence(facts); paths != "" {
		candidates = append(candidates, candidate(caps, "versioned_documentation", "Context7", "context7", "central version-sensitive framework", paths, "a central framework with a parsed version makes version-sensitive documentation relevant; Context7 is a provider candidate without install/config/integration support"))
	}
	if surface, boundaries := lspSurface(facts); surface != "" && boundaries != "" {
		candidates = append(candidates, candidate(caps, "symbol_intelligence", "Serena", "serena", "LSP-language source surface with module boundaries", "lsp_surface="+surface+" "+boundaries, "supported-language source surface plus module boundaries indicates cross-file symbol navigation; Serena is a provider candidate without install/config/integration support"))
	}
	if scale, ok := scaleBand(facts); ok {
		candidates = append(candidates, candidate(caps, "semantic_retrieval", "Semble", "semble", "medium/large repository scale band", scale, "the documented §14.3 scale band indicates textual retrieval would reduce candidates; Semble is a provider candidate without install/config/integration support"))
	}
	if hasOutputDirs(root) {
		candidates = append(candidates, candidate(caps, "general_context_compression", "Headroom", "headroom", "context-compression pressure with RTK evidence", "output directories present alongside RTK evidence", "output-pressure signals coexist with RTK evidence; Headroom is a provider candidate without install/config/integration support and whether compression remains a problem is model-owned"))
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

// frameworkEvidence joins "name version (evidence)" for every framework fact
// with a parsed version. It is the only Context7 trigger (D3): lockfile-only
// or manager-conflict evidence never generates a candidate.
func frameworkEvidence(facts inventory.Facts) string {
	parts := []string{}
	for _, f := range facts.FrameworkFacts {
		if f.Version != "" {
			parts = append(parts, f.Name+" "+f.Version+" ("+strings.Join(f.Evidence, ",")+")")
		}
	}
	return strings.Join(parts, "; ")
}

// lspSurfaceIDs are manifest signal IDs of languages with mature language
// servers (the §14.4 Serena LSP-language source surface).
var lspSurfaceIDs = []string{"go", "javascript", "python", "rust"}

// lspSurface returns the comma-joined LSP-language manifest evidence and the
// module-boundary evidence (workspace signals or multiple workspaces). Serena
// fires only when both exist; a small single-package repo stays signal-free.
func lspSurface(facts inventory.Facts) (string, string) {
	paths := []string{}
	for _, id := range lspSurfaceIDs {
		if joined := joinPaths(facts.Manifests, id); joined != "" {
			paths = append(paths, joined)
		}
	}
	boundaries := ""
	if signals := joinPaths(facts.WorkspaceSignals, ""); signals != "" {
		boundaries = "workspace_signals=" + signals
	} else if len(facts.Workspaces) > 1 {
		boundaries = "workspaces=" + strconv.Itoa(len(facts.Workspaces))
	}
	return strings.Join(paths, "; "), boundaries
}

// sourceExtensions are the documented §14.3 scale-band source extensions;
// counts are observed evidence, never a verdict.
var sourceExtensions = map[string]bool{
	"c": true, "cpp": true, "cs": true, "dart": true, "ex": true, "exs": true,
	"go": true, "h": true, "hpp": true, "java": true, "js": true, "jsx": true,
	"kt": true, "php": true, "py": true, "rb": true, "rs": true, "swift": true,
	"ts": true, "tsx": true,
}

// scaleBand reports whether the repository meets the documented §14.3 scale
// band (>=200 source files or >=60 declared dependencies); the observed
// counts accompany the candidate as evidence only.
func scaleBand(facts inventory.Facts) (string, bool) {
	src := 0
	for ext, count := range facts.Files.ByExtension {
		if sourceExtensions[ext] {
			src += count
		}
	}
	if src >= 200 || len(facts.Deps) >= 60 {
		return fmt.Sprintf("source_files=%d deps=%d", src, len(facts.Deps)), true
	}
	return "", false
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
