package lifecycle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/JhonMA82/agent-ready/internal/bootstrap"
	"github.com/JhonMA82/agent-ready/internal/safeio"
)

// Change implements the safeio.Change contract for update reconciles.
type Change struct {
	path, kind    string
	before, after []byte
	mode          fs.FileMode
}

// Path returns the repository-relative path.
func (c Change) Path() string { return c.path }

// Kind returns create | update | noop.
func (c Change) Kind() string { return c.kind }

// Before returns the prior bytes (nil for new files).
func (c Change) Before() []byte { return bytes.Clone(c.before) }

// After returns the desired bytes.
func (c Change) After() []byte { return bytes.Clone(c.after) }

// Mode returns the desired mode.
func (c Change) Mode() fs.FileMode { return c.mode }

// Plan implements the safeio.Plan contract.
type Plan struct {
	root    string
	changes []Change
}

// Root returns the repository root.
func (p Plan) Root() string { return p.root }

// Changes returns the sorted change set.
func (p Plan) Changes() []Change { return p.changes }

// UpdatePlan reconciles the installed manifest-owned assets to the binary's
// embedded assets (drift-tolerant: only owned paths with byte drift change;
// state/checkpoints and generated artifacts are untouched by construction).
// Returns a noop plan when everything is already byte-identical.
func UpdatePlan(root string) (Plan, error) {
	schema, _, installed, err := installedManifest(root)
	if err != nil {
		return Plan{}, err
	}
	if schema == "" {
		return Plan{}, fmt.Errorf("not initialized: run agent-ready init first")
	}
	// The installed manifest records the config owner; re-read it fully.
	data, err := os.ReadFile(filepath.Join(root, ".agent-ready", "manifest.json"))
	if err != nil {
		return Plan{}, err
	}
	var full struct {
		ConfigFile string `json:"config_file"`
	}
	if err := json.Unmarshal(data, &full); err != nil {
		return Plan{}, err
	}
	configFile := full.ConfigFile

	desired, err := bootstrap.Desired(configFile)
	if err != nil {
		return Plan{}, err
	}
	owned := map[string]bool{}
	for _, asset := range installed {
		owned[asset.Path] = true
	}
	changes := []Change{}
	for _, file := range desired {
		if !owned[file.Path] {
			// Never touch files outside the installed manifest ownership.
			continue
		}
		before, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil && !os.IsNotExist(err) {
			return Plan{}, err
		}
		kind := "update"
		switch {
		case os.IsNotExist(err):
			kind = "create"
		case bytes.Equal(before, file.After):
			kind = "noop"
		}
		changes = append(changes, Change{path: file.Path, kind: kind, before: before, after: bytes.Clone(file.After), mode: file.Mode})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	return Plan{root: root, changes: changes}, nil
}

// ApplyUpdate commits the reconcile plan via the safeio transaction.
func ApplyUpdate(plan Plan) error {
	_, err := safeio.Commit(plan, safeio.Options{})
	return err
}
