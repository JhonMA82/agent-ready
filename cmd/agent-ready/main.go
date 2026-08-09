package main

import (
	"context"
	"errors"
	"os"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/checkpoint"
	"github.com/JhonMA82/agent-ready/internal/cli"
	"github.com/JhonMA82/agent-ready/internal/inventory"
	"github.com/JhonMA82/agent-ready/internal/lifecycle"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/repository"
	statestore "github.com/JhonMA82/agent-ready/internal/state"
	"github.com/JhonMA82/agent-ready/internal/tools"
	"github.com/JhonMA82/agent-ready/internal/validation"
)

// refusedLifecycle renders a refused plan.Result for lifecycle commands.
func refusedLifecycle(root, category string, dryRun bool, err error) plan.Result {
	r := plan.NewResult(root, plan.Refused, dryRun, nil)
	r.Refusal = &plan.Refusal{Category: category, Message: err.Error(), Remediation: "resolve the reported error and retry"}
	return r
}

func main() {
	run := func(ctx context.Context, options cli.Options) plan.Result { return app.Init(ctx, options.DryRun) }
	update := func(ctx context.Context, options cli.Options) plan.Result {
		selection, err := repository.Discover(ctx, "", "git")
		if err != nil {
			return refusedLifecycle("", "repository", options.DryRun, err)
		}
		p, err := lifecycle.UpdatePlan(selection.Root)
		if err != nil {
			return refusedLifecycle(selection.Root, "plan", options.DryRun, err)
		}
		actions := make([]plan.Action, 0, len(p.Changes()))
		allNoop := true
		for _, change := range p.Changes() {
			actions = append(actions, plan.Action{Kind: change.Kind(), Path: change.Path()})
			allNoop = allNoop && change.Kind() == "noop"
		}
		if options.DryRun {
			return plan.NewResult(selection.Root, plan.DryRun, true, actions)
		}
		if allNoop {
			return plan.NewResult(selection.Root, plan.Noop, false, actions)
		}
		if err := lifecycle.ApplyUpdate(p); err != nil {
			r := plan.NewResult(selection.Root, plan.CommitFailed, false, actions)
			r.Refusal = &plan.Refusal{Category: "commit", Message: err.Error(), Remediation: "resolve the transaction error and retry"}
			return r
		}
		r := plan.NewResult(selection.Root, plan.Changed, false, actions)
		r.NextStep = "/agent-ready"
		return r
	}
	inspect := cli.Helper{
		Name: "inspect",
		Run: func(ctx context.Context, _ cli.Options) (any, error) {
			selection, err := repository.Discover(ctx, "", "git")
			if err != nil {
				return nil, err
			}
			return inventory.Inspect(selection.Root, selection.Invocation)
		},
	}
	validate := cli.Helper{
		Name: "validate",
		Run: func(ctx context.Context, options cli.Options) (any, error) {
			selection, err := repository.Discover(ctx, options.Target, "git")
			if err != nil {
				return nil, err
			}
			facts, err := validation.Validate(selection.Root)
			if err != nil {
				return nil, err
			}
			if facts.Verdict != "pass" {
				return facts, errors.New("validation failed")
			}
			return facts, nil
		},
	}
	// withRoot discovers the containing repository and hands its root to f.
	withRoot := func(f func(string) (any, error)) func(context.Context, cli.Options) (any, error) {
		return func(ctx context.Context, _ cli.Options) (any, error) {
			selection, err := repository.Discover(ctx, "", "git")
			if err != nil {
				return nil, err
			}
			return f(selection.Root)
		}
	}
	ckpt := cli.Helper{Name: "checkpoint", Subs: []cli.Helper{
		{Name: "save", Run: func(ctx context.Context, options cli.Options) (any, error) {
			if options.Stage == "" {
				return nil, errors.New("--stage is required (e.g. --stage exploration_plan)")
			}
			selection, err := repository.Discover(ctx, "", "git")
			if err != nil {
				return nil, err
			}
			return checkpoint.Save(selection.Root, options.Stage, options.Complete)
		}},
		{Name: "status", Run: withRoot(func(root string) (any, error) { return checkpoint.Status(root) })},
	}}
	changes := cli.Helper{Name: "changes", Run: withRoot(func(root string) (any, error) { return checkpoint.Changes(root) })}
	state := cli.Helper{Name: "state", Run: withRoot(func(root string) (any, error) { return statestore.Read(root) })}
	toolsStatus := cli.Helper{Name: "status", Run: func(context.Context, cli.Options) (any, error) { return tools.Status() }}
	toolsDoctor := cli.Helper{Name: "doctor", Run: withRoot(func(root string) (any, error) {
		facts, err := tools.Doctor(root)
		if err != nil {
			return nil, err
		}
		if !facts.Healthy {
			return facts, errors.New("required tool checks failed")
		}
		return facts, nil
	})}
	toolsRecommend := cli.Helper{Name: "recommend", Run: withRoot(func(root string) (any, error) { return tools.Recommend(root) })}
	toolsInstall := cli.Helper{Name: "install", Run: func(_ context.Context, options cli.Options) (any, error) {
		plan, err := tools.Plan(options.Tool)
		if err != nil {
			return nil, err
		}
		if options.DryRun {
			return plan, nil
		}
		approved, err := tools.ConfirmConsent(os.Stdin, plan)
		if err != nil {
			return nil, err
		}
		if !approved {
			return plan, errors.New("installation cancelled; nothing was executed")
		}
		return tools.Install(plan)
	}}
	toolsCmd := cli.Helper{Name: "tools", Subs: []cli.Helper{toolsStatus, toolsDoctor, toolsRecommend, toolsInstall}}
	status := cli.Helper{Name: "status", Run: withRoot(func(root string) (any, error) { return lifecycle.Status(root) })}
	doctor := cli.Helper{Name: "doctor", Run: func(ctx context.Context, _ cli.Options) (any, error) {
		selection, err := repository.Discover(ctx, "", "git")
		if err != nil {
			return nil, err
		}
		facts, err := lifecycle.Doctor(ctx, selection.Root)
		if err != nil {
			return nil, err
		}
		if !facts.Healthy {
			return facts, errors.New("required checks failed")
		}
		return facts, nil
	}}
	remove := func(ctx context.Context, options cli.Options) plan.Result {
		selection, err := repository.Discover(ctx, "", "git")
		if err != nil {
			return refusedLifecycle("", "repository", options.DryRun, err)
		}
		if options.Mode == "" {
			return refusedLifecycle(selection.Root, "mode", options.DryRun, errors.New("--mode is required: harness-only or harness+generated"))
		}
		p, err := lifecycle.PlanRemove(selection.Root, lifecycle.Mode(options.Mode))
		if err != nil {
			return refusedLifecycle(selection.Root, "plan", options.DryRun, err)
		}
		actions := make([]plan.Action, 0, len(p.Removed)+len(p.Kept))
		for _, entry := range p.Removed {
			actions = append(actions, plan.Action{Kind: "remove", Path: entry.Path})
		}
		for _, entry := range p.Kept {
			actions = append(actions, plan.Action{Kind: "kept", Path: entry.Path})
		}
		if options.DryRun {
			return plan.NewResult(selection.Root, plan.DryRun, true, actions)
		}
		if len(p.Removed) == 0 {
			return plan.NewResult(selection.Root, plan.Noop, false, actions)
		}
		if err := lifecycle.ApplyRemove(p); err != nil {
			r := plan.NewResult(selection.Root, plan.CommitFailed, false, actions)
			r.Refusal = &plan.Refusal{Category: "commit", Message: err.Error(), Remediation: "resolve the removal error and retry"}
			return r
		}
		r := plan.NewResult(selection.Root, plan.Changed, false, actions)
		r.NextStep = "/agent-ready (not initialized)"
		return r
	}
	if err := cli.NewRootWithCommands(run, update, remove, inspect, validate, ckpt, changes, state, toolsCmd, status, doctor).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
