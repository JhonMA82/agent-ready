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
	Target string // validate --target (empty: discover from cwd)
}
type Runner func(context.Context, Options) plan.Result

// Helper is a deterministic JSON-fact subcommand (spec R8): Run returns facts
// rendered as JSON with --json or as the value's compact Summary; helpers exit
// 0 on success, 1 on failure, and never perform semantic routing.
type Helper struct {
	Name string
	Run  func(context.Context, Options) (any, error)
}

type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

// NewRoot builds the root command: init (unchanged) plus one subcommand per
// Helper with the helper exit-code contract (0 success / 1 failure).
func NewRoot(run Runner, helpers ...Helper) *cobra.Command {
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
	for _, helper := range helpers {
		sub := &cobra.Command{Use: helper.Name, Short: "Emit deterministic " + helper.Name + " facts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			value, err := helper.Run(cmd.Context(), options)
			// Render a non-nil result even when Run failed so a failing
			// helper (validate verdict fail) still emits facts before exit 1.
			if value != nil {
				if rerr := RenderHelper(cmd.OutOrStdout(), value, options.JSON); rerr != nil && err == nil {
					err = rerr
				}
			}
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return ExitError{Code: 1}
			}
			return nil
		}}
		sub.Flags().BoolVar(&options.JSON, "json", false, "emit JSON facts")
		if helper.Name == "validate" {
			sub.Flags().StringVar(&options.Target, "target", "", "repository to validate (default: discovered from cwd)")
		}
		root.AddCommand(sub)
	}
	return root
}

// summarizer is implemented by helper facts with a compact rendering (D5).
type summarizer interface{ Summary() string }

func RenderHelper(w io.Writer, value any, jsonMode bool) error {
	if jsonMode {
		return json.NewEncoder(w).Encode(value)
	}
	summary, ok := value.(summarizer)
	if !ok {
		return fmt.Errorf("helper result has no compact summary")
	}
	fmt.Fprintln(w, summary.Summary())
	return nil
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
