package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JhonMA82/agent-ready/internal/inventory"
)

// Candidate is one capability-need candidate with its observed signal
// (spec §36 deterministic criteria; evidence only, never a verdict).
type Candidate struct {
	Capability string `json:"capability"`
	Candidate  string `json:"candidate"`
	Signal     string `json:"signal"`
	Observed   string `json:"observed"`
}

// RecommendFacts is the agent-ready.recommend/v1 schema; candidates sorted
// by capability. An empty array means no documented signal fired.
type RecommendFacts struct {
	SchemaVersion string      `json:"schema_version"`
	Candidates    []Candidate `json:"candidates"`
}

// RecommendSchemaVersion is the agent-ready.recommend/v1 schema.
const RecommendSchemaVersion = "agent-ready.recommend/v1"

// Recommend emits candidate evidence implementing only the documented §36
// signal table over deterministic inventory facts. The model makes the final
// tool-budget decision (spec 13/35); Go never emits install verdicts.
func Recommend(root string) (RecommendFacts, error) {
	facts, err := inventory.Inspect(root, "")
	if err != nil {
		return RecommendFacts{}, err
	}
	candidates := []Candidate{}
	githubRemote := gitHubRemote(root)
	if hasOutputDirs(root) {
		candidates = append(candidates, Candidate{Capability: "token_optimized_shell", Candidate: "RTK", Signal: "build/test outputs are large", Observed: "output directories present"})
	}
	if githubRemote != "" {
		candidates = append(candidates, Candidate{Capability: "github_context", Candidate: "gh", Signal: "GitHub remote", Observed: githubRemote})
	}
	if len(facts.Workspaces) > 1 || depCount(facts) > 40 {
		candidates = append(candidates, Candidate{Capability: "dependency_graph", Candidate: "CodeGraph", Signal: "workspaces or dense dependencies", Observed: "workspaces=" + strconv.Itoa(len(facts.Workspaces)) + " deps=" + strconv.Itoa(depCount(facts))})
	}
	if hasLockfiles(root) {
		candidates = append(candidates, Candidate{Capability: "versioned_documentation", Candidate: "Context7", Signal: "version-sensitive dependencies", Observed: "lockfiles present"})
	}
	if len(candidates) == 0 && facts.Files.Total > 500 {
		candidates = append(candidates, Candidate{Capability: "semantic_retrieval", Candidate: "Semble", Signal: "lexical search noise", Observed: "file_count=" + strconv.Itoa(facts.Files.Total)})
	}
	return RecommendFacts{SchemaVersion: RecommendSchemaVersion, Candidates: candidates}, nil
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

// hasLockfiles reports version-sensitive dependency lockfiles.
func hasLockfiles(root string) bool {
	for _, name := range []string{"package-lock.json", "bun.lock", "go.sum", "yarn.lock"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
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
