// Package inventory emits deterministic repository inventory facts (spec R8).
package inventory

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/JhonMA82/agent-ready/internal/ecosystem"
)

// SchemaVersion is the agent-ready.inspect/v1 fact schema.
const SchemaVersion = "agent-ready.inspect/v1"

type Dep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
type Script struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}
type Files struct {
	Total       int            `json:"total"`
	ByExtension map[string]int `json:"by_extension"`
}
type CI struct {
	Present bool     `json:"present"`
	Files   []string `json:"files"`
}
type Presence struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

// BoilerplateFacts records explicit extension and generated/editable markers.
// Paths are repository-relative and sorted; empty collections stay omitted.
type BoilerplateFacts struct {
	ExtensionPoints []string `json:"extension_points,omitempty"`
	GeneratedFiles  []string `json:"generated_files,omitempty"`
	EditableFiles   []string `json:"editable_files,omitempty"`
}

// Facts is the agent-ready.inspect/v1 schema; slices and map keys are sorted.
type Facts struct {
	SchemaVersion string     `json:"schema_version"`
	Root          string     `json:"root"`
	Invocation    string     `json:"invocation,omitempty"`
	Deps          []Dep      `json:"deps"`
	Scripts       []Script   `json:"scripts"`
	Workspaces    []string   `json:"workspaces"`
	Files         Files      `json:"files"`
	CI            CI         `json:"ci"`
	Presence      []Presence `json:"presence,omitempty"`
	ecosystem.Facts
	OutputSignals []ecosystem.Signal `json:"output_signals,omitempty"`
	BoilerplateFacts
}

var manifests = []struct {
	name string
	kind string
	read func([]byte) ([]Dep, []Script, []string, error)
}{
	{"go.mod", "gomod", readGoMod},
	{"package.json", "npm", readPackageJSON},
}

var ciRoot = []string{".gitlab-ci.yml", ".circleci/config.yml", "azure-pipelines.yml", "Jenkinsfile"}

// heavyTrees (D11, spec §9): known dependency and output trees that are never
// recursively scanned; their existence stays path-and-kind presence evidence.
// cmake-build-* matches by prefix; storage/logs matches by nested path.
var heavyTrees = []string{
	".dart_tool", ".next", ".nuxt", ".venv", "__pycache__", "_build", "bin",
	"build", "coverage", "deps", "dist", "node_modules", "obj", "out",
	"result", "storage/logs", "target", "vendor", "venv",
}

// isHeavyTree reports whether a directory is a heavy tree, matching exact
// names at any depth (the same signal surface as outputSignalID), so a PHP
// bin/console under bin/ is pruned while the bin presence signal is retained.
func isHeavyTree(rel, name string) bool {
	if strings.HasPrefix(name, "cmake-build-") {
		return true
	}
	for _, dir := range heavyTrees {
		if name == dir || rel == dir || strings.HasSuffix(rel, "/"+dir) {
			return true
		}
	}
	return false
}

// outputSignalDirs: §9 output/build dirs as presence signals (cmake-build-*
// prefix, nested storage/logs); candidate-only, no verdict.
var outputSignalDirs = []string{
	".dart_tool", ".next", ".nuxt", ".venv", "__pycache__", "_build", "bin",
	"build", "coverage", "deps", "dist", "node_modules", "obj", "out",
	"result", "storage/logs", "target", "vendor", "venv",
}

func outputSignalID(rel, name string) string {
	if strings.HasPrefix(name, "cmake-build-") {
		return name
	}
	for _, dir := range outputSignalDirs {
		if name == dir || rel == dir || strings.HasSuffix(rel, "/"+dir) {
			return dir
		}
	}
	return ""
}

// Paths returns the sorted relative paths of regular files under root,
// excluding .git and symlinks (the same walk Inspect reports).
func Paths(root string) ([]string, error) {
	paths, _, _, _, err := collectFiles(root)
	return paths, err
}

// Inspect collects deterministic facts for root; invocation is reported only
// when it differs from root.
func Inspect(root, invocation string) (Facts, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Facts{}, err
	}
	info, err := os.Stat(root)
	if err == nil && !info.IsDir() {
		err = fmt.Errorf("not a directory: %s", root)
	}
	if err != nil {
		return Facts{}, err
	}
	paths, presence, outputSignals, files, err := collectFiles(root)
	if err != nil {
		return Facts{}, err
	}
	deps, scripts, workspaces := []Dep{}, []Script{}, []string{}
	for _, m := range manifests {
		data, err := os.ReadFile(filepath.Join(root, m.name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Facts{}, err
		}
		d, s, w, err := m.read(data)
		if err != nil {
			return Facts{}, fmt.Errorf("parse %s: %w", m.name, err)
		}
		deps, scripts, workspaces = append(deps, d...), append(scripts, s...), append(workspaces, w...)
	}
	sort.Slice(deps, func(i, j int) bool {
		a, b := deps[i], deps[j]
		return a.Kind < b.Kind || a.Kind == b.Kind && (a.Name < b.Name || a.Name == b.Name && a.Version < b.Version)
	})
	sort.Slice(scripts, func(i, j int) bool { return scripts[i].Name < scripts[j].Name })
	sort.Strings(workspaces)
	facts := Facts{SchemaVersion: SchemaVersion, Root: root, Deps: deps, Scripts: scripts, Workspaces: workspaces, Files: files, CI: findCI(paths), Presence: presence, OutputSignals: outputSignals, Facts: ecosystem.Detect(root, paths), BoilerplateFacts: collectBoilerplateFacts(root, paths)}
	if invocation != "" && invocation != root {
		facts.Invocation = invocation
	}
	return facts, nil
}

func collectBoilerplateFacts(root string, paths []string) BoilerplateFacts {
	facts := BoilerplateFacts{}
	for _, rel := range paths {
		lower := strings.ToLower(rel)
		if strings.Contains(lower, "extension-point") || strings.Contains(lower, "extension_points") {
			facts.ExtensionPoints = append(facts.ExtensionPoints, rel)
		}
		if generatedPath(rel) {
			facts.GeneratedFiles = append(facts.GeneratedFiles, rel)
		}
	}
	if len(facts.ExtensionPoints) == 0 && len(facts.GeneratedFiles) == 0 {
		return facts
	}
	for _, rel := range paths {
		if slices.Contains(facts.GeneratedFiles, rel) {
			continue
		}
		lower := strings.ToLower(rel)
		for _, dir := range []string{"app/", "apps/", "lib/", "src/", "variants/"} {
			if strings.HasPrefix(lower, dir) {
				facts.EditableFiles = append(facts.EditableFiles, rel)
				break
			}
		}
	}
	return facts
}

func generatedPath(rel string) bool {
	lower := strings.ToLower(rel)
	return strings.HasPrefix(lower, "generated/") || strings.Contains(lower, "/generated/") || strings.HasSuffix(lower, ".g.dart") || strings.Contains(lower, ".generated.")
}

func collectFiles(root string) ([]string, []Presence, []ecosystem.Signal, Files, error) {
	var paths []string
	var presence []Presence
	var outputSignals []ecosystem.Signal
	files := Files{ByExtension: map[string]int{}}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			if path != root {
				rel, _ := filepath.Rel(root, path)
				rel = filepath.ToSlash(rel)
				if id := outputSignalID(rel, d.Name()); id != "" {
					outputSignals = append(outputSignals, ecosystem.Signal{ID: id, Path: rel})
				}
				if isHeavyTree(rel, d.Name()) {
					presence = append(presence, Presence{rel, "directory"})
					return fs.SkipDir
				}
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		paths = append(paths, rel)
		files.Total++
		if ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(rel), ".")); ext != "" {
			files.ByExtension[ext]++
		}
		return nil
	})
	sort.Strings(paths)
	sort.Slice(presence, func(i, j int) bool { return presence[i].Path < presence[j].Path })
	sort.Slice(outputSignals, func(i, j int) bool {
		return outputSignals[i].ID < outputSignals[j].ID || outputSignals[i].ID == outputSignals[j].ID && outputSignals[i].Path < outputSignals[j].Path
	})
	return paths, presence, outputSignals, files, err
}

func findCI(paths []string) CI {
	ci := CI{Files: []string{}}
	for _, path := range paths {
		dir, name := filepath.Split(path)
		if slices.Contains(ciRoot, path) || dir == ".github/workflows/" && (strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")) {
			ci.Files = append(ci.Files, path)
		}
	}
	ci.Present = len(ci.Files) > 0
	return ci
}

func readGoMod(data []byte) ([]Dep, []Script, []string, error) {
	var deps []Dep
	block := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "" || strings.HasPrefix(line, "//"):
		case line == "require (":
			block = true
		case block && line == ")":
			block = false
		case block || strings.HasPrefix(line, "require "):
			if fields := strings.Fields(strings.TrimPrefix(line, "require ")); len(fields) >= 2 {
				deps = append(deps, Dep{Name: fields[0], Version: fields[1], Kind: "gomod"})
			}
		}
	}
	return deps, nil, nil, nil
}

func readPackageJSON(data []byte) ([]Dep, []Script, []string, error) {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
		Workspaces      json.RawMessage   `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil, nil, err
	}
	deps, scripts, workspaces := []Dep{}, []Script{}, []string{}
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
		for name, version := range group {
			deps = append(deps, Dep{Name: name, Version: version, Kind: "npm"})
		}
	}
	for name, command := range pkg.Scripts {
		scripts = append(scripts, Script{Name: name, Command: command})
	}
	if list := []string{}; len(pkg.Workspaces) > 0 && json.Unmarshal(pkg.Workspaces, &list) == nil {
		workspaces = append(workspaces, list...)
	}
	return deps, scripts, workspaces, nil
}

func (f Facts) Summary() string {
	ci := "absent"
	if f.CI.Present {
		ci = "present (" + strings.Join(f.CI.Files, ", ") + ")"
	}
	return fmt.Sprintf("Root: %s\nFiles: %d\nDeps: %d\nScripts: %d\nWorkspaces: %d\nCI: %s", f.Root, f.Files.Total, len(f.Deps), len(f.Scripts), len(f.Workspaces), ci)
}
