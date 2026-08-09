package main

import (
	"context"
	"os"

	"github.com/JhonMA82/agent-ready/internal/app"
	"github.com/JhonMA82/agent-ready/internal/cli"
	"github.com/JhonMA82/agent-ready/internal/plan"
)

func main() {
	run := func(ctx context.Context, options cli.Options) plan.Result { return app.Init(ctx, options.DryRun) }
	if err := cli.NewRoot(run).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
