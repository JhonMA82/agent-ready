package tools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// InstallPlan is the exact, reviewable installation plan rendered before any
// execution (spec §19). Elevation is modeled separately and is never part of
// recipe args; V1 recipes are non-elevated.
type InstallPlan struct {
	Tool       string   `json:"tool"`
	PM         string   `json:"pm"`
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
	Elevation  bool     `json:"elevation"`
	Reason     string   `json:"reason"`
}

// InstallResult reports the execution and post-install verification outcome.
type InstallResult struct {
	Tool     string `json:"tool"`
	Executed bool   `json:"executed"`
	Verified bool   `json:"verified"`
	Version  string `json:"version,omitempty"`
}

// InstallSchemaVersion is the agent-ready.install/v1 schema.
const InstallSchemaVersion = "agent-ready.install/v1"

// Plan selects the single verified recipe for the detected package manager
// and renders the exact execution plan. It fails closed with remediation
// when the tool has no recipe or the platform has no supported PM.
func Plan(toolID string) (InstallPlan, error) {
	if toolID == "" {
		return InstallPlan{}, errors.New("tool id required")
	}
	var recipe *Recipe
	for i := range Catalog() {
		if Catalog()[i].ID == toolID {
			recipe = &Catalog()[i]
			break
		}
	}
	if recipe == nil {
		return InstallPlan{}, fmt.Errorf("no verified install recipe for %q; recipes exist for: %s", toolID, recipeIDs())
	}
	pm := DetectPackageManager()
	if pm == "" {
		return InstallPlan{}, errors.New("no supported package manager detected (apt, pacman, dnf, brew)")
	}
	op, ok := recipe.Install[pm]
	if !ok {
		return InstallPlan{}, fmt.Errorf("no verified %s recipe for %q; supported managers: %s", pm, toolID, recipeManagers(*recipe))
	}
	return InstallPlan{Tool: toolID, PM: pm, Executable: op.Executable, Args: append([]string{}, op.Args...), Elevation: false, Reason: "verified embedded recipe"}, nil
}

func recipeIDs() string {
	ids := make([]string, 0, 8)
	for _, recipe := range Catalog() {
		ids = append(ids, recipe.ID)
	}
	return strings.Join(ids, ", ")
}

func recipeManagers(recipe Recipe) string {
	managers := make([]string, 0, len(recipe.Install))
	for pm := range recipe.Install {
		managers = append(managers, pm)
	}
	return strings.Join(managers, ", ")
}

// Install executes the approved plan with the recipe's fixed executable+args
// (never a shell), then verifies the tool is present on PATH afterward. A
// verification failure fails closed with remediation.
func Install(plan InstallPlan) (InstallResult, error) {
	if plan.Executable == "" {
		return InstallResult{}, errors.New("install plan is empty; run Plan first")
	}
	cmd := exec.Command(plan.Executable, plan.Args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return InstallResult{Tool: plan.Tool, Executed: true}, fmt.Errorf("%s install failed: %w (remediation: check package manager output and retry)", plan.Tool, err)
	}
	verified, version := presentAfterInstall(plan.Tool)
	if !verified {
		return InstallResult{Tool: plan.Tool, Executed: true, Verified: false}, fmt.Errorf("%s install completed but the tool is not on PATH (remediation: open a new shell or refresh PATH)", plan.Tool)
	}
	return InstallResult{Tool: plan.Tool, Executed: true, Verified: true, Version: version}, nil
}

// presentAfterInstall re-detects the tool; PATH refresh limitations are
// reported through the fail-closed remediation above.
func presentAfterInstall(toolID string) (bool, string) {
	for _, recipe := range Catalog() {
		if recipe.ID != toolID {
			continue
		}
		return detect(recipe)
	}
	return false, ""
}

// ConfirmConsent reads one line from r and returns true only for an
// explicit affirmative answer. Any read failure or empty input declines:
// consent never defaults to yes.
func ConfirmConsent(r io.Reader, plan InstallPlan) (bool, error) {
	fmt.Fprintf(os.Stdout, "Plan: install %s via %s %s\n", plan.Tool, plan.Executable, strings.Join(plan.Args, " "))
	fmt.Fprint(os.Stdout, "Proceed? [y/N] ")
	var answer string
	if _, err := fmt.Fscanln(r, &answer); err != nil {
		return false, nil // unreadable or empty input declines
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// Summary renders the compact default output.
func (p InstallPlan) Summary() string {
	return fmt.Sprintf("Install %s: %s %s (PM: %s)", p.Tool, p.Executable, strings.Join(p.Args, " "), p.PM)
}

// Summary renders the compact default output.
func (r InstallResult) Summary() string {
	return fmt.Sprintf("Install %s: executed=%t verified=%t", r.Tool, r.Executed, r.Verified)
}
