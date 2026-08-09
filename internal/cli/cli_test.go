package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/JhonMA82/agent-ready/internal/plan"
	"strings"
	"testing"
)

func TestInitContracts(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{{[]string{"init", "--dry-run"}, "Outcome: dry_run"}, {[]string{"init", "--dry-run", "--json"}, `"schema_version":"agent-ready.result/v1"`}} {
		var out bytes.Buffer
		cmd := NewRoot(func(context.Context, Options) plan.Result { return plan.NewResult("/repo", plan.DryRun, true, nil) })
		cmd.SetArgs(tt.args)
		cmd.SetOut(&out)
		if err := cmd.Execute(); err != nil || !strings.Contains(out.String(), tt.want) {
			t.Fatalf("output %q, error %v", out.String(), err)
		}
	}
	cmd := NewRoot(func(context.Context, Options) plan.Result { return plan.Result{} })
	cmd.SetArgs([]string{"init", "extra"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected usage error")
	}
}

func TestRenderContract(t *testing.T) {
	refused := plan.Result{Outcome: plan.Refused, Root: "/repo", Invocation: "/repo/a/b", Refusal: &plan.Refusal{Category: "repository", Message: "not a worktree", Remediation: "retry inside a repository"}}
	var out bytes.Buffer
	if err := Render(&out, refused, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Root: /repo", "Invocation: /repo/a/b", "Refused (repository): not a worktree", "Remediation: retry inside a repository"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %q", want, out.String())
		}
	}
	out.Reset()
	nested := plan.NewResult("/repo", plan.DryRun, true, nil)
	nested.Invocation = "/repo/a"
	if err := Render(&out, nested, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"invocation":"/repo/a"`) {
		t.Fatalf("nested invocation missing: %s", out.String())
	}
}

type summaryFacts struct{ Total int }

func (s summaryFacts) Summary() string { return fmt.Sprintf("files: %d", s.Total) }

func TestHelperContracts(t *testing.T) {
	helper := Helper{Name: "inspect", Run: func(context.Context, Options) (any, error) { return summaryFacts{Total: 3}, nil }}
	for _, tt := range []struct {
		args []string
		want string
	}{{[]string{"inspect"}, "files: 3\n"}, {[]string{"inspect", "--json"}, `"Total":3`}} {
		var out bytes.Buffer
		cmd := NewRoot(nil, helper)
		cmd.SetArgs(tt.args)
		cmd.SetOut(&out)
		if err := cmd.Execute(); err != nil || !strings.Contains(out.String(), tt.want) {
			t.Fatalf("output %q, error %v", out.String(), err)
		}
	}
	cmd := NewRoot(nil, Helper{Name: "inspect", Run: func(context.Context, Options) (any, error) { return nil, errors.New("preflight failed") }})
	cmd.SetArgs([]string{"inspect"})
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if exit, ok := err.(ExitError); !ok || exit.Code != 1 || !strings.Contains(errOut.String(), "preflight failed") {
		t.Fatalf("expected ExitError{1} with reason on stderr, got %v / %q", err, errOut.String())
	}
}

func TestNestedHelperContracts(t *testing.T) {
	save := Helper{Name: "save", Run: func(_ context.Context, options Options) (any, error) {
		if options.Stage == "" {
			return nil, errors.New("--stage is required")
		}
		return summaryFacts{Total: 1}, nil
	}}
	status := Helper{Name: "status", Run: func(context.Context, Options) (any, error) { return summaryFacts{Total: 0}, nil }}
	ckpt := Helper{Name: "checkpoint", Subs: []Helper{save, status}}
	for _, tt := range []struct {
		args []string
		want string
	}{{[]string{"checkpoint", "save", "--stage", "plan"}, "files: 1\n"}, {[]string{"checkpoint", "save", "--json", "--stage", "plan"}, `"Total":1`}, {[]string{"checkpoint", "status"}, "files: 0\n"}} {
		var out bytes.Buffer
		cmd := NewRoot(nil, ckpt)
		cmd.SetArgs(tt.args)
		cmd.SetOut(&out)
		if err := cmd.Execute(); err != nil || !strings.Contains(out.String(), tt.want) {
			t.Fatalf("output %q, error %v", out.String(), err)
		}
	}
	var errOut bytes.Buffer
	cmd := NewRoot(nil, ckpt)
	cmd.SetArgs([]string{"checkpoint", "save"})
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if exit, ok := err.(ExitError); !ok || exit.Code != 1 || !strings.Contains(errOut.String(), "--stage is required") {
		t.Fatalf("expected ExitError{1} with reason on stderr, got %v / %q", err, errOut.String())
	}
}

func TestValidateHelperContracts(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := NewRoot(nil, Helper{Name: "validate", Run: func(context.Context, Options) (any, error) {
		return summaryFacts{Total: 2}, errors.New("validation failed: 1 of 1 skills violate the pinned OpenCode 1.18.15 rules")
	}})
	cmd.SetArgs([]string{"validate"})
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if exit, ok := err.(ExitError); !ok || exit.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(out.String(), "files: 2") || !strings.Contains(errOut.String(), "validation failed") {
		t.Fatalf("facts on stdout (%q) and reason on stderr (%q) required", out.String(), errOut.String())
	}
}

func TestDoctorHelperExitContract(t *testing.T) {
	// A Run error must yield ExitError{1} with facts still on stdout.
	doctor := Helper{Name: "doctor", Run: func(context.Context, Options) (any, error) {
		return summaryFacts{Total: 0}, errors.New("required tool checks failed")
	}}
	var out, errOut bytes.Buffer
	cmd := NewRoot(nil, Helper{Name: "tools", Subs: []Helper{doctor}})
	cmd.SetArgs([]string{"tools", "doctor"})
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if exit, ok := err.(ExitError); !ok || exit.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(out.String(), "files: 0") || !strings.Contains(errOut.String(), "required tool checks failed") {
		t.Fatalf("facts on stdout (%q) and reason on stderr (%q) required", out.String(), errOut.String())
	}
}

func TestVersionFlagContract(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRoot(func(context.Context, Options) plan.Result { return plan.NewResult("/repo", plan.Noop, false, nil) })
	cmd.SetArgs([]string{"--version"})
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.Contains(out.String(), "agent-ready") {
		t.Fatalf("--version output: %q", out.String())
	}
}
