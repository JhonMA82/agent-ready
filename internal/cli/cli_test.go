package cli

import (
	"bytes"
	"context"
	"github.com/gentle-ai/agent-ready/internal/plan"
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
