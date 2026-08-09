package main

import (
	"context"
	"errors"
	"os"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/cli"
	"github.com/JhonMA82/agent-ready/internal/inventory"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/repository"
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
	if err := cli.NewRoot(run, inspect, validate).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
