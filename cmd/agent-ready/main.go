package main

import (
	"context"
	"os"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/cli"
	"github.com/JhonMA82/agent-ready/internal/inventory"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/repository"
)

func main() {
	run := func(ctx context.Context, options cli.Options) plan.Result { return app.Init(ctx, options.DryRun) }
	inspect := cli.Helper{
		Name: "inspect",
		Run: func(ctx context.Context, _ cli.Options) (any, error) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			selection, err := repository.Discover(ctx, cwd, "git")
			if err != nil {
				return nil, err
			}
			return inventory.Inspect(selection.Root, selection.Invocation)
		},
	}
	if err := cli.NewRoot(run, inspect).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
