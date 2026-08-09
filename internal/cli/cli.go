package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/spf13/cobra"
)

type Options struct {
	DryRun bool
	JSON   bool
}
type Runner func(context.Context, Options) plan.Result
type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

func NewRoot(run Runner) *cobra.Command {
	root := &cobra.Command{Use: "agent-ready", Short: "Prepare a repository for agent-ready workflows", SilenceErrors: true, SilenceUsage: true}
	var options Options
	init := &cobra.Command{Use: "init", Short: "Initialize the containing repository", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		r := run(cmd.Context(), options)
		if err := Render(cmd.OutOrStdout(), r, options.JSON); err != nil {
			return err
		}
		if code := plan.ExitCode(r); code != 0 {
			return ExitError{Code: code}
		}
		return nil
	}}
	init.Flags().BoolVar(&options.DryRun, "dry-run", false, "show the plan without writing")
	init.Flags().BoolVar(&options.JSON, "json", false, "render one JSON result")
	root.AddCommand(init)
	return root
}

func Render(w io.Writer, r plan.Result, jsonMode bool) error {
	if jsonMode {
		return json.NewEncoder(w).Encode(r)
	}
	fmt.Fprintf(w, "Outcome: %s\nRoot: %s\n", r.Outcome, r.Root)
	if r.Invocation != "" {
		fmt.Fprintf(w, "Invocation: %s\n", r.Invocation)
	}
	fmt.Fprintf(w, "Dry run: %t\n", r.DryRun)
	for _, a := range r.Actions {
		fmt.Fprintf(w, "- %s %s\n", a.Kind, a.Path)
	}
	if r.Refusal != nil {
		fmt.Fprintf(w, "Refused (%s): %s\nRemediation: %s\n", r.Refusal.Category, r.Refusal.Message, r.Refusal.Remediation)
	}
	if r.NextStep != "" {
		fmt.Fprintf(w, "Next: %s\n", r.NextStep)
	}
	return nil
}
