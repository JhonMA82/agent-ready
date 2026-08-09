package main

import (
	"context"
	"errors"
	"os"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/checkpoint"
	"github.com/JhonMA82/agent-ready/internal/cli"
	"github.com/JhonMA82/agent-ready/internal/inventory"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/repository"
	statestore "github.com/JhonMA82/agent-ready/internal/state"
	"github.com/JhonMA82/agent-ready/internal/validation"
)

func main() {
	run := func(ctx context.Context, options cli.Options) plan.Result { return app.Init(ctx, options.DryRun) }
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
	if err := cli.NewRoot(run, inspect, validate, ckpt, changes, state).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
