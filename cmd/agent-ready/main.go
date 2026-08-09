package main

import (
	"context"
	"os"

	"github.com/gentle-ai/agent-ready/internal/cli"
	"github.com/gentle-ai/agent-ready/internal/plan"
)

func main() {
	run := func(_ context.Context, options cli.Options) plan.Result {
		r := plan.NewResult("", plan.Refused, options.DryRun, nil)
		r.Refusal = &plan.Refusal{Category: "foundation_only", Message: "repository initialization is not available yet", Remediation: "install the complete agent-ready V1 build"}
		return r
	}
	if err := cli.NewRoot(run).Execute(); err != nil {
		if exit, ok := err.(cli.ExitError); ok {
			os.Exit(exit.Code)
		}
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
}
