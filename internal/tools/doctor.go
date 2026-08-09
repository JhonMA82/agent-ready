package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DoctorCheck is one read-only doctor check result.
type DoctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | warning | fail
	Detail string `json:"detail,omitempty"`
}

// DoctorFacts is the agent-ready.doctor/v1 schema.
type DoctorFacts struct {
	SchemaVersion string        `json:"schema_version"`
	Checks        []DoctorCheck `json:"checks"`
	Healthy       bool          `json:"healthy"`
}

// DoctorSchemaVersion is the agent-ready.doctor/v1 schema.
const DoctorSchemaVersion = "agent-ready.doctor/v1"

// Doctor runs read-only integration checks (spec §18): required tier
// (git, opencode) fail hard; recommended tier (rg, fd, jq) warn; recipe
// availability and project integration report. Never mutates.
func Doctor(root string) (DoctorFacts, error) {
	facts := DoctorFacts{SchemaVersion: DoctorSchemaVersion, Healthy: true}
	for _, name := range []string{"git", "opencode"} {
		check := DoctorCheck{Name: name}
		if _, err := exec.LookPath(name); err != nil {
			check.Status, check.Detail = "fail", "required tool missing from PATH"
			facts.Healthy = false
		} else {
			check.Status = "ok"
		}
		facts.Checks = append(facts.Checks, check)
	}
	for _, name := range []string{"rg", "fd", "jq"} {
		check := DoctorCheck{Name: name}
		if _, err := exec.LookPath(name); err != nil {
			check.Status, check.Detail = "warning", "recommended tool missing"
		} else {
			check.Status = "ok"
		}
		facts.Checks = append(facts.Checks, check)
	}
	pm := DetectPackageManager()
	recipeCheck := DoctorCheck{Name: "recipes"}
	if pm == "" {
		recipeCheck.Status, recipeCheck.Detail = "warning", "no supported package manager detected"
	} else {
		recipeCheck.Status = "ok"
		recipeCheck.Detail = "recipes available for package manager " + pm
	}
	facts.Checks = append(facts.Checks, recipeCheck)
	facts.Checks = append(facts.Checks, integrationChecks(root)...)
	return facts, nil
}

// integrationChecks verifies the project harness state and OpenCode config
// presence. A present but malformed config is reported as a warning, never a
// failure (read-only semantics per the pinned doctor contract).
func integrationChecks(root string) []DoctorCheck {
	var checks []DoctorCheck
	state := DoctorCheck{Name: "project state"}
	if _, err := os.Stat(filepath.Join(root, ".agent-ready", "state")); err == nil {
		state.Status = "ok"
	} else if os.IsNotExist(err) {
		state.Status, state.Detail = "warning", "run agent-ready init first"
	} else {
		state.Status, state.Detail = "warning", err.Error()
	}
	checks = append(checks, state)
	config := DoctorCheck{Name: "opencode config"}
	configPath := ""
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			configPath = name
			break
		}
	}
	if configPath == "" {
		config.Status, config.Detail = "warning", "no opencode.json/jsonc found"
	} else {
		config.Status = "ok"
		data, err := os.ReadFile(filepath.Join(root, configPath))
		if err != nil || !strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
			config.Status, config.Detail = "warning", configPath+" present but unparseable"
		}
	}
	checks = append(checks, config)
	return checks
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
	return "Doctor: " + summaryStatus(fails, warnings)
}

func summaryStatus(fails, warnings int) string {
	switch {
	case fails > 0:
		return "FAIL"
	case warnings > 0:
		return "ok (warnings)"
	default:
		return "ok"
	}
}
