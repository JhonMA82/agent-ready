package lifecycle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

type updateManifest struct {
	Schema               string      `json:"schema"`
	InstallVersion       string      `json:"install_version"`
	CompatibilityVersion string      `json:"compatibility_version"`
	ConfigFile           string      `json:"config_file"`
	ConfigPath           string      `json:"config_path"`
	Assets               []AssetInfo `json:"assets"`
}

// UpdatePlan advances unchanged owned assets to the binary's embedded assets.
// Modified ownership and unmanaged collisions fail planning without writes;
// state, checkpoints, generated artifacts, and obsolete ownership stay intact.
func UpdatePlan(root string) (Plan, error) {
	data, err := os.ReadFile(filepath.Join(root, ".agent-ready", "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return Plan{}, fmt.Errorf("not initialized: run agent-ready init first")
		}
		return Plan{}, err
	}
	var installed updateManifest
	if err := json.Unmarshal(data, &installed); err != nil {
		return Plan{}, err
	}
	if installed.Schema == "" {
		return Plan{}, fmt.Errorf("not initialized: run agent-ready init first")
	}
	desired, err := bootstrap.Desired(installed.ConfigFile)
	if err != nil {
		return Plan{}, err
	}
	return updatePlan(root, data, installed.Assets, desired)
}

func updatePlan(root string, manifestBefore []byte, installed []AssetInfo, desired []bootstrap.File) (Plan, error) {
	owned := make(map[string]AssetInfo, len(installed))
	for _, asset := range installed {
		owned[asset.Path] = asset
	}
	changes := []Change{}
	conflicts := []string{}
	desiredPaths := map[string]bool{}
	var manifestFile bootstrap.File
	for _, file := range desired {
		if file.Path == ".agent-ready/manifest.json" {
			manifestFile = file
			continue
		}
		desiredPaths[file.Path] = true
		asset, wasOwned := owned[file.Path]
		full := filepath.Join(root, filepath.FromSlash(file.Path))
		before, err := os.ReadFile(full)
		if err != nil && !os.IsNotExist(err) {
			return Plan{}, err
		}
		if wasOwned && !AssetMatches(root, asset) {
			conflicts = append(conflicts, file.Path+" (modified owned asset)")
			continue
		}
		if !wasOwned && err == nil {
			conflicts = append(conflicts, file.Path+" (unmanaged collision)")
			continue
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
	for _, asset := range installed {
		if !desiredPaths[asset.Path] && !AssetMatches(root, asset) {
			conflicts = append(conflicts, asset.Path+" (modified obsolete asset)")
		}
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return Plan{}, fmt.Errorf("ownership conflicts: %s", strings.Join(conflicts, ", "))
	}
	var reconciled updateManifest
	if err := json.Unmarshal(manifestFile.After, &reconciled); err != nil {
		return Plan{}, err
	}
	for _, asset := range installed {
		if !desiredPaths[asset.Path] {
			reconciled.Assets = append(reconciled.Assets, asset)
		}
	}
	sort.Slice(reconciled.Assets, func(i, j int) bool { return reconciled.Assets[i].Path < reconciled.Assets[j].Path })
	manifestAfter, err := json.Marshal(reconciled)
	if err != nil {
		return Plan{}, err
	}
	manifestAfter = append(manifestAfter, '\n')
	kind := "update"
	if bytes.Equal(manifestBefore, manifestAfter) {
		kind = "noop"
	}
	changes = append(changes, Change{path: manifestFile.Path, kind: kind, before: bytes.Clone(manifestBefore), after: manifestAfter, mode: manifestFile.Mode})
	sort.Slice(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	return Plan{root: root, changes: changes}, nil
}

// ApplyUpdate commits the reconcile plan via the safeio transaction.
func ApplyUpdate(plan Plan) error {
	_, err := safeio.Commit(plan, safeio.Options{})
	return err
}
