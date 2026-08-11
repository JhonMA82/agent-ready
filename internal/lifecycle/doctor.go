package lifecycle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/JhonMA82/agent-ready/internal/opencode"
	"github.com/JhonMA82/agent-ready/internal/tools"
	"github.com/JhonMA82/agent-ready/internal/validation"
)

// Check is one CLI-level doctor check result.
type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | warning | fail
	Detail string `json:"detail,omitempty"`
}

// DoctorFacts is the CLI-level agent-ready.doctor/v1 schema (§34 merged
// checks: lifecycle checks plus the tools tier subset).
type DoctorFacts struct {
	SchemaVersion string  `json:"schema_version"`
	Checks        []Check `json:"checks"`
	Healthy       bool    `json:"healthy"`
}

// DoctorSchemaVersion is the CLI-level agent-ready.doctor/v1 schema.
const DoctorSchemaVersion = "agent-ready.doctor/v1"

// Doctor runs the §34 read-only checks. Required failures (repository,
// compatible opencode, initialized harness) make Healthy false (the command
// exits 1); recommended-tool warnings keep Healthy true (exit 0).
func Doctor(ctx context.Context, root string) (DoctorFacts, error) {
	facts := DoctorFacts{SchemaVersion: DoctorSchemaVersion, Healthy: true}
	add := func(name, status, detail string) {
		facts.Checks = append(facts.Checks, Check{Name: name, Status: status, Detail: detail})
		if status == "fail" {
			facts.Healthy = false
		}
	}
	if _, err := exec.LookPath("git"); err != nil {
		add("repository", "fail", "git required on PATH")
	} else if _, err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output(); err != nil {
		add("repository", "fail", "not inside a Git worktree")
	} else {
		add("repository", "ok", "")
	}
	if res, err := opencode.Preflight(ctx, []byte("{}")); err != nil {
		add("opencode", "fail", err.Error())
	} else {
		add("opencode", "ok", "version "+res.Version)
		if res.Drift != "" {
			add("opencode drift", "warning", res.Drift)
		}
	}
	if _, version, _, err := installedManifest(root); err != nil {
		add("initialized", "fail", err.Error())
	} else if version == "" {
		add("initialized", "fail", "run agent-ready init first")
	} else {
		add("initialized", "ok", "install_version "+version)
	}
	if _, err := os.Stat(filepath.Join(root, ".opencode", "commands", "agent-ready.md")); err != nil {
		add("command", "fail", "local /agent-ready command missing")
	} else {
		add("command", "ok", "")
	}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready", "skills")); err != nil {
		add("skill source", "fail", ".agent-ready/skills missing")
	} else {
		add("skill source", "ok", "")
	}
	if factsVal, err := validation.Validate(root); err != nil {
		add("skill validity", "warning", err.Error())
	} else if factsVal.Verdict != "pass" {
		add("skill validity", "warning", "installed skills fail pinned validation")
	} else {
		add("skill validity", "ok", "")
	}
	missingDirs := 0
	for _, dir := range []string{".agent-ready/state", ".agent-ready/checkpoints"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil {
			missingDirs++
		}
	}
	if missingDirs > 0 {
		add("runtime dirs", "warning", strconv.Itoa(missingDirs)+" runtime dir(s) missing (run init)")
	} else {
		add("runtime dirs", "ok", "")
	}
	if toolFacts, err := tools.Doctor(root); err == nil {
		for _, check := range toolFacts.Checks {
			if check.Name == "rg" || check.Name == "fd" || check.Name == "jq" {
				add("tool:"+check.Name, check.Status, check.Detail)
			}
		}
	}
	return facts, nil
}

// Summary renders the compact default output.
func (f DoctorFacts) Summary() string {
	fails, warnings := 0, 0
	for _, check := range f.Checks {
		switch check.Status {
		case "fail":
			fails++
		case "warning":
			warnings++
		}
	}
	switch {
	case fails > 0:
		return "Doctor: FAIL"
	case warnings > 0:
		return "Doctor: ok (warnings)"
	default:
		return "Doctor: ok"
	}
}
