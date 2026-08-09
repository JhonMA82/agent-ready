package plan

import "sort"

const SchemaVersion = "agent-ready.result/v1"

type Outcome string

const (
	Changed          Outcome = "changed"
	Noop             Outcome = "noop"
	DryRun           Outcome = "dry_run"
	Refused          Outcome = "refused"
	CommitFailed     Outcome = "commit_failed"
	RecoveryRequired Outcome = "recovery_required"
)

type Action struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}
type Refusal struct {
	Category    string `json:"category"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
}
type Result struct {
	SchemaVersion string   `json:"schema_version"`
	Outcome       Outcome  `json:"outcome"`
	Root          string   `json:"root"`
	Invocation    string   `json:"invocation,omitempty"`
	DryRun        bool     `json:"dry_run"`
	Actions       []Action `json:"actions"`
	Refusal       *Refusal `json:"refusal,omitempty"`
	NextStep      string   `json:"next_step,omitempty"`
}

func NewResult(root string, outcome Outcome, dryRun bool, actions []Action) Result {
	actions = append([]Action{}, actions...)
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Path < actions[j].Path || actions[i].Path == actions[j].Path && actions[i].Kind < actions[j].Kind
	})
	return Result{SchemaVersion: SchemaVersion, Outcome: outcome, Root: root, DryRun: dryRun, Actions: actions}
}

var exitCodes = map[Outcome]int{Refused: 3, CommitFailed: 4, RecoveryRequired: 5}

func ExitCode(r Result) int { return exitCodes[r.Outcome] }
