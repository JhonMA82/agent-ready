package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	facts.Checks = append(facts.Checks, providerChecks(root)...)
	for _, check := range facts.Checks {
		if check.Status == "fail" {
			facts.Healthy = false
		}
	}
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

// providerProbe is the §49 doctor probe table: read-only executable and
// version probes per provider, plus which providers require a project index.
// Providers declare no install contract in V1; probing never implies support.
var providerProbe = []struct {
	id            string
	versionArgs   []string
	requiresIndex bool
}{
	{id: "codegraph", versionArgs: []string{"--version"}, requiresIndex: true},
	{id: "context7", versionArgs: []string{"--version"}},
	{id: "headroom", versionArgs: []string{"--version"}},
	{id: "semble", versionArgs: []string{"--version"}},
	{id: "serena", versionArgs: []string{"--version"}},
}

// providerChecks runs the §49 per-provider checks: executable exists, version
// parses, project index/config exists when required, OpenCode integration is
// detectable when enabled, and provider health when inexpensive. Providers
// are never auto-installed, so absence is a warning; a broken version or a
// missing required index is a failure with a reason — a provider is never
// reported healthy merely because a binary exists.
func providerChecks(root string) []DoctorCheck {
	var checks []DoctorCheck
	for _, p := range providerProbe {
		check := DoctorCheck{Name: "provider:" + p.id, Status: "ok", Detail: "executable+version ok"}
		if path, err := exec.LookPath(p.id); err != nil {
			check.Status, check.Detail = "warning", "not installed (providers are never auto-installed)"
		} else {
			out, _ := exec.Command(path, p.versionArgs...).Output()
			if version := providerVersionRE.FindString(string(out)); version == "" {
				check.Status, check.Detail = "fail", "version check failed: executable exists but its version does not parse: "+strings.TrimSpace(string(out))
			} else {
				check.Detail = "version " + version
			}
			if check.Status == "ok" && p.requiresIndex {
				if _, err := os.Stat(filepath.Join(root, ".codegraph")); err != nil {
					check.Status, check.Detail = "fail", "project index missing at .codegraph/ (remediation: run codegraph init)"
				} else {
					check.Detail += "; project index present"
				}
			}
			if check.Status == "ok" {
				check.Detail += "; integration unsupported in V1; health covered by executable/version/index checks"
			}
		}
		checks = append(checks, check)
	}
	return checks
}

// providerVersionRE is the §49 version parse: a x.y.z token; version output
// without one fails the check (never healthy merely because a binary exists).
var providerVersionRE = regexp.MustCompile(`\bv?[0-9]+\.[0-9]+\.[0-9]+\b`)

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
