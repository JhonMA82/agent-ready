package cli

import (
	"bytes"
	"context"
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
