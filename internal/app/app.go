package app

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/JhonMA82/agent-ready/internal/bootstrap"
	"github.com/JhonMA82/agent-ready/internal/opencode"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/repository"
	"github.com/JhonMA82/agent-ready/internal/safeio"
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

func Result(p Plan, invocation string, dryRun bool) plan.Result {
	actions := make([]plan.Action, len(p.changes))
	allNoop := true
	for i, change := range p.changes {
		actions[i] = plan.Action{Kind: change.kind, Path: change.path}
		allNoop = allNoop && change.kind == "noop"
	}
	if dryRun {
		r := plan.NewResult(p.root, plan.DryRun, true, actions)
		setInvocation(&r, p.root, invocation)
		return r
	}
	if allNoop {
		r := plan.NewResult(p.root, plan.Noop, false, actions)
		r.NextStep = "/agent-ready"
		setInvocation(&r, p.root, invocation)
		return r
	}
	r := plan.NewResult(p.root, plan.Changed, false, actions)
	r.NextStep = "/agent-ready"
	setInvocation(&r, p.root, invocation)
	return r
}

// setInvocation reports the invocation only when it differs from the target root.
func setInvocation(r *plan.Result, root, invocation string) {
	if invocation != "" && invocation != root {
		r.Invocation = invocation
	}
}

func Init(ctx context.Context, dryRun bool) plan.Result {
	return initWithOptions(ctx, dryRun, safeio.Options{})
}

func initWithOptions(ctx context.Context, dryRun bool, options safeio.Options) plan.Result {
	cwd, err := os.Getwd()
	if err != nil {
		return refusedAt("", "", "repository", dryRun, err)
	}
	selection, err := repository.Discover(ctx, cwd, "git")
	if err != nil {
		return refusedAt("", "", "repository", dryRun, err)
	}
	journal := filepath.Join(selection.Root, ".agent-ready", "transaction.json")
	if _, err := os.Lstat(journal); err == nil {
		if dryRun {
			r := failedAt(selection.Root, selection.Invocation, plan.RecoveryRequired, "recovery", errors.New("recovery journal requires a non-dry-run init"))
			r.DryRun = true
			return r
		}
		if _, err := safeio.Recover(selection.Root, options); err != nil {
			return failedAt(selection.Root, selection.Invocation, plan.RecoveryRequired, "recovery", err)
		}
	} else if !os.IsNotExist(err) {
		return failedAt(selection.Root, selection.Invocation, plan.RecoveryRequired, "recovery", err)
	}
	if _, err := opencode.Preflight(ctx, []byte("{}")); err != nil {
		return refusedAt(selection.Root, selection.Invocation, "opencode", dryRun, err)
	}
	p, err := Build(selection.Root)
	if err != nil {
		return refusedAt(selection.Root, selection.Invocation, "plan", dryRun, err)
	}
	r := Result(p, selection.Invocation, dryRun)
	if dryRun || r.Outcome == plan.Noop {
		return r
	}
	result, err := safeio.Commit(p, options)
	if err == nil {
		return r
	}
	if result.RecoveryPath != "" {
		return commitFailed(r, plan.RecoveryRequired, "recovery", err)
	}
	return commitFailed(r, plan.CommitFailed, "commit", err)
}

func refusedAt(root, invocation, category string, dryRun bool, err error) plan.Result {
	r := plan.NewResult(root, plan.Refused, dryRun, nil)
	setInvocation(&r, root, invocation)
	r.Refusal = &plan.Refusal{Category: category, Message: err.Error(), Remediation: "resolve the reported preflight error and retry"}
	return r
}

func failedAt(root, invocation string, outcome plan.Outcome, category string, err error) plan.Result {
	r := plan.NewResult(root, outcome, false, nil)
	setInvocation(&r, root, invocation)
	r.Refusal = &plan.Refusal{Category: category, Message: err.Error(), Remediation: "resolve the repository transaction error and retry"}
	return r
}

func commitFailed(r plan.Result, outcome plan.Outcome, category string, err error) plan.Result {
	r.Outcome, r.NextStep = outcome, ""
	r.Refusal = &plan.Refusal{Category: category, Message: err.Error(), Remediation: "resolve the repository transaction error and retry"}
	return r
}
