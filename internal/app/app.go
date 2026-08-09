package app

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/gentle-ai/agent-ready/internal/bootstrap"
	"github.com/gentle-ai/agent-ready/internal/opencode"
	"github.com/gentle-ai/agent-ready/internal/plan"
	"github.com/gentle-ai/agent-ready/internal/repository"
)

type Change struct {
	path, kind    string
	before, after []byte
	mode          fs.FileMode
}

func (c Change) Path() string      { return c.path }
func (c Change) Kind() string      { return c.kind }
func (c Change) Before() []byte    { return bytes.Clone(c.before) }
func (c Change) After() []byte     { return bytes.Clone(c.after) }
func (c Change) Mode() fs.FileMode { return c.mode }

type Plan struct {
	root    string
	changes []Change
}

func (p Plan) Root() string { return p.root }
func (p Plan) Changes() []Change {
	out := make([]Change, len(p.changes))
	for i, c := range p.changes {
		out[i] = Change{path: c.path, kind: c.kind, before: bytes.Clone(c.before), after: bytes.Clone(c.after), mode: c.mode}
	}
	return out
}

// Build validates every planned path and returns a frozen, sorted plan without writing.
func Build(root string) (Plan, error) {
	jsonFile, err := readConfig(root, "opencode.json")
	if err != nil {
		return Plan{}, err
	}
	jsoncFile, err := readConfig(root, "opencode.jsonc")
	if err != nil {
		return Plan{}, err
	}
	config, err := opencode.PlanConfig(jsonFile, jsoncFile)
	if err != nil {
		return Plan{}, err
	}
	files, err := bootstrap.Plan(root, config.Path)
	if err != nil {
		return Plan{}, err
	}
	files = append(files, bootstrap.File{Path: config.Path, Before: config.Before, After: config.After, Mode: config.Mode})
	changes := make([]Change, 0, len(files))
	for _, file := range files {
		if _, err := repository.Contained(root, filepath.FromSlash(file.Path)); err != nil {
			return Plan{}, err
		}
		kind := "update"
		switch {
		case bytes.Equal(file.Before, file.After):
			kind = "noop"
		case file.Before == nil:
			kind = "create"
		}
		changes = append(changes, Change{path: file.Path, kind: kind, before: bytes.Clone(file.Before), after: bytes.Clone(file.After), mode: file.Mode})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	return Plan{root: root, changes: changes}, nil
}

func readConfig(root, name string) (*opencode.ConfigFile, error) {
	full, err := repository.Contained(root, name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(full)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(full)
	if err != nil {
		return nil, err
	}
	return &opencode.ConfigFile{Data: data, Mode: info.Mode().Perm()}, nil
}

func Result(p Plan, dryRun bool) plan.Result {
	actions := make([]plan.Action, len(p.changes))
	allNoop := true
	for i, change := range p.changes {
		actions[i] = plan.Action{Kind: change.kind, Path: change.path}
		allNoop = allNoop && change.kind == "noop"
	}
	if dryRun {
		return plan.NewResult(p.root, plan.DryRun, true, actions)
	}
	if allNoop {
		return plan.NewResult(p.root, plan.Noop, false, actions)
	}
	r := plan.NewResult(p.root, plan.Refused, false, actions)
	r.Refusal = &plan.Refusal{Category: "commit_unavailable", Message: "repository writes are not available yet", Remediation: "use --dry-run or install the complete agent-ready V1 build"}
	return r
}

func Init(ctx context.Context, dryRun bool) plan.Result {
	cwd, err := os.Getwd()
	if err != nil {
		return refusedAt("", "repository", dryRun, err)
	}
	selection, err := repository.Discover(ctx, cwd, "git")
	if err != nil {
		return refusedAt("", "repository", dryRun, err)
	}
	if _, err := opencode.Preflight(ctx, []byte("{}")); err != nil {
		return refusedAt(selection.Root, "opencode", dryRun, err)
	}
	p, err := Build(selection.Root)
	if err != nil {
		return refusedAt(selection.Root, "plan", dryRun, err)
	}
	return Result(p, dryRun)
}

func refusedAt(root, category string, dryRun bool, err error) plan.Result {
	r := plan.NewResult(root, plan.Refused, dryRun, nil)
	r.Refusal = &plan.Refusal{Category: category, Message: err.Error(), Remediation: "resolve the reported preflight error and retry"}
	return r
}
