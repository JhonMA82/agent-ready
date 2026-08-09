// Package state reports read-only facts about the model-owned semantic
// state files under .agent-ready/state (spec R10): decisions.jsonl,
// provenance.jsonl, artifact-graph.yaml, repository-profile.yaml. Go never
// writes these files; the model owns their content.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SchemaVersion is the agent-ready.state/v1 schema.
const SchemaVersion = "agent-ready.state/v1"

// FileFact is one semantic state file fact.
type FileFact struct {
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Bytes   int64  `json:"bytes,omitempty"`
	Entries int    `json:"entries,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

// Facts is the agent-ready.state/v1 schema; files sorted by path.
type Facts struct {
	SchemaVersion string     `json:"schema_version"`
	Files         []FileFact `json:"files"`
}

// Names are the model-owned semantic state files under .agent-ready/state.
var Names = []string{
	"decisions.jsonl",
	"provenance.jsonl",
	"artifact-graph.yaml",
	"repository-profile.yaml",
}

// Read collects facts for each semantic state file under root/.agent-ready/state.
func Read(root string) (Facts, error) {
	dir := filepath.Join(root, ".agent-ready", "state")
	facts := Facts{SchemaVersion: SchemaVersion}
	for _, name := range Names {
		path := filepath.Join(dir, name)
		info, err := os.Lstat(path)
		file := FileFact{Path: ".agent-ready/state/" + name}
		if os.IsNotExist(err) {
			facts.Files = append(facts.Files, file)
			continue
		}
		if err != nil {
			return Facts{}, err
		}
		file.Exists = true
		file.Bytes = info.Size()
		file.ModTime = info.ModTime().UTC().Format(time.RFC3339)
		if strings.HasSuffix(name, ".jsonl") {
			data, err := os.ReadFile(path)
			if err != nil {
				return Facts{}, err
			}
			file.Entries = countLines(data)
		}
		facts.Files = append(facts.Files, file)
	}
	sort.Slice(facts.Files, func(i, j int) bool { return facts.Files[i].Path < facts.Files[j].Path })
	return facts, nil
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	if data[len(data)-1] == '\n' {
		count--
	}
	return count
}

// Summary renders the compact default output.
func (f Facts) Summary() string {
	existing := 0
	for _, file := range f.Files {
		if file.Exists {
			existing++
		}
	}
	return fmt.Sprintf("State files: %d/%d present", existing, len(f.Files))
}
