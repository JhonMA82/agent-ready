package plan

import (
	"encoding/json"
	"testing"
)

func TestResultAndExitContracts(t *testing.T) {
	r := NewResult("/repo", DryRun, true, []Action{{Kind: "create", Path: "z"}, {Kind: "noop", Path: "a"}})
	b, err := json.Marshal(r)
	if err != nil || string(b) != `{"schema_version":"agent-ready.result/v1","outcome":"dry_run","root":"/repo","dry_run":true,"actions":[{"kind":"noop","path":"a"},{"kind":"create","path":"z"}]}` {
		t.Fatalf("result: %s (%v)", b, err)
	}
	for outcome, code := range map[Outcome]int{Changed: 0, DryRun: 0, Refused: 3, CommitFailed: 4, RecoveryRequired: 5} {
		if got := ExitCode(Result{Outcome: outcome}); got != code {
			t.Errorf("%s: got %d", outcome, got)
		}
	}
}
