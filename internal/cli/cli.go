package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JhonMA82/agent-ready/internal/plan"
	"github.com/JhonMA82/agent-ready/internal/version"
	"github.com/spf13/cobra"
)

type Options struct {
	DryRun   bool
	JSON     bool
	Target   string // validate --target (empty: discover from cwd)
	Stage    string // checkpoint save --stage
	Complete bool   // checkpoint save --complete
	Tool     string // tools install --tool
	Mode     string // remove --mode
}
type Runner func(context.Context, Options) plan.Result

// Helper is a deterministic JSON-fact subcommand (spec R8): Run returns facts
// as JSON with --json or as the compact Summary; helpers exit 0/1 and never
// perform semantic routing. Subs nests subcommands. Use, when set, is the
// usage line for one positional arg (e.g. "explain TOOL"); it enables
// cobra.ExactArgs(1) and passes the arg as Options.Tool (D7).
type Helper struct {
	Name string
	Use  string
	Run  func(context.Context, Options) (any, error)
	Subs []Helper
}

type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

// NewRoot builds the root command: init plus one subcommand per Helper with
// the helper exit-code contract (0 success / 1 failure); helpers with Subs
// become parent commands (e.g. checkpoint save/status).
func NewRoot(run Runner, helpers ...Helper) *cobra.Command {
	return NewRootWithCommands(run, nil, nil, helpers...)
}

// NewRootWithCommands additionally wires lifecycle commands such as update
// and remove (plan.Result semantics, --dry-run) alongside init.
func NewRootWithCommands(run Runner, update Runner, remove Runner, helpers ...Helper) *cobra.Command {
	root := &cobra.Command{Use: "agent-ready", Short: "Prepare a repository for agent-ready workflows", SilenceErrors: true, SilenceUsage: true}
	root.Version = version.String()
	root.SetVersionTemplate("{{.Version}}\n")
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
	if update != nil {
		upd := &cobra.Command{Use: "update", Short: "Refresh installed harness assets", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			r := update(cmd.Context(), options)
			if err := Render(cmd.OutOrStdout(), r, options.JSON); err != nil {
				return err
			}
			if code := plan.ExitCode(r); code != 0 {
				return ExitError{Code: code}
			}
			return nil
		}}
		upd.Flags().BoolVar(&options.DryRun, "dry-run", false, "show the plan without writing")
		upd.Flags().BoolVar(&options.JSON, "json", false, "render one JSON result")
		root.AddCommand(upd)
	}
	if remove != nil {
		rm := &cobra.Command{Use: "remove", Short: "Remove the local harness", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			r := remove(cmd.Context(), options)
			if err := Render(cmd.OutOrStdout(), r, options.JSON); err != nil {
				return err
			}
			if code := plan.ExitCode(r); code != 0 {
				return ExitError{Code: code}
			}
			return nil
		}}
		rm.Flags().StringVar(&options.Mode, "mode", "", "removal mode: harness-only or harness+generated")
		rm.Flags().BoolVar(&options.DryRun, "dry-run", false, "show the plan without writing")
		rm.Flags().BoolVar(&options.JSON, "json", false, "render one JSON result")
		root.AddCommand(rm)
	}
	for _, helper := range helpers {
		if len(helper.Subs) > 0 {
			root.AddCommand(parentFor(helper, &options))
			continue
		}
		sub := &cobra.Command{Use: helper.Name, Short: "Emit deterministic " + helper.Name + " facts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				options.Tool = args[0]
			}
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
		if helper.Use != "" {
			sub.Use = helper.Use
			sub.Args = cobra.ExactArgs(1)
		}
		sub.Flags().BoolVar(&options.JSON, "json", false, "emit JSON facts")
		if helper.Name == "validate" {
			sub.Flags().StringVar(&options.Target, "target", "", "repository to validate (default: discovered from cwd)")
		}
		root.AddCommand(sub)
	}
	return root
}

// parentFor builds a parent command from helper.Subs (e.g. checkpoint
// save/status). It stays runnable so unknown subcommands fail Args validation
// (exit 2) instead of silently showing help.
func parentFor(helper Helper, options *Options) *cobra.Command {
	sub := &cobra.Command{Use: helper.Name, Short: "Emit deterministic " + helper.Name + " facts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	for _, nested := range helper.Subs {
		child := &cobra.Command{Use: nested.Name, Short: "Emit deterministic " + nested.Name + " facts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				options.Tool = args[0]
			}
			value, err := nested.Run(cmd.Context(), *options)
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
		if nested.Use != "" {
			child.Use = nested.Use
			child.Args = cobra.ExactArgs(1)
		}
		child.Flags().BoolVar(&options.JSON, "json", false, "emit JSON facts")
		if nested.Name == "save" {
			child.Flags().StringVar(&options.Stage, "stage", "", "checkpoint stage (e.g. exploration_plan)")
			child.Flags().BoolVar(&options.Complete, "complete", false, "mark the checkpoint complete")
		}
		if nested.Name == "install" {
			child.Flags().StringVar(&options.Tool, "tool", "", "tool id to install (e.g. rg)")
			child.Flags().BoolVar(&options.DryRun, "dry-run", false, "render the plan without executing")
		}
		sub.AddCommand(child)
	}
	return sub
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
