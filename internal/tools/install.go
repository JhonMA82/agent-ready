package tools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// InstallPlan is the exact, reviewable installation plan rendered before any
// execution (spec §46): tool, kind, evidence, safety level, side effects,
// and the plan fields. Elevation is modeled separately, never in args.
type InstallPlan struct {
	Tool        string      `json:"tool"`
	Kind        string      `json:"kind,omitempty"`
	Evidence    string      `json:"evidence,omitempty"`
	PM          string      `json:"pm"`
	Method      string      `json:"method,omitempty"`
	Level       SafetyLevel `json:"level,omitempty"`
	SideEffects string      `json:"side_effects,omitempty"`
	Executable  string      `json:"executable"`
	Args        []string    `json:"args"`
	Elevation   bool        `json:"elevation"`
	Reason      string      `json:"reason"`
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
// when the tool has no recipe or the platform has no supported PM; AUR is
// opt-in only (never auto-selected) and nix is environment-only.
func Plan(toolID string) (InstallPlan, error) {
	if toolID == "" {
		return InstallPlan{}, errors.New("tool id required")
	}
	var entry *Entry
	for i := range Catalog() {
		if Catalog()[i].ID == toolID {
			entry = &Catalog()[i]
			break
		}
	}
	if entry == nil || entry.Install == nil {
		return InstallPlan{}, fmt.Errorf("no verified install recipe for %q; recipes exist for: %s", toolID, recipeIDs())
	}
	pm := DetectPackageManager()
	if pm == "" {
		if aur := detectAUR(); aur != "" {
			return InstallPlan{}, fmt.Errorf("only the AUR helper %q is available; AUR is opt-in only and never auto-selected (remediation: install a supported package manager or use an explicitly approved AUR flow)", aur)
		}
		if nixPresent() {
			return InstallPlan{}, errors.New("nix is detected as an environment only and is never used as an automatic universal installer (remediation: install a supported package manager: apt, pacman, dnf, brew, zypper, apk, winget)")
		}
		return InstallPlan{}, errors.New("no supported package manager detected (apt, pacman, dnf, brew, zypper, apk, winget); remediation: install one of these package managers")
	}
	op, ok := entry.Install[pm]
	if !ok {
		return InstallPlan{}, fmt.Errorf("no verified %s recipe for %q; supported managers: %s", pm, toolID, recipeManagers(*entry))
	}
	return InstallPlan{
		Tool: toolID, Kind: string(entry.Family), Evidence: "verified embedded recipe",
		PM: pm, Method: "verified recipe", Level: entry.SafetyLevel, SideEffects: entry.SideEffects,
		Executable: op.Executable, Args: append([]string{}, op.Args...), Elevation: false,
		Reason: "verified embedded recipe",
	}, nil
}

// recipeIDs lists the recipe-backed catalog entries.
func recipeIDs() string {
	ids := make([]string, 0, 8)
	for _, entry := range Catalog() {
		if entry.Install != nil {
			ids = append(ids, entry.ID)
		}
	}
	return strings.Join(ids, ", ")
}

func recipeManagers(entry Entry) string {
	managers := make([]string, 0, len(entry.Install))
	for pm := range entry.Install {
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
	for _, entry := range Catalog() {
		if entry.ID != toolID {
			continue
		}
		return detect(entry)
	}
	return false, ""
}

// RenderPlan writes the §46 install plan and consent prompt to w: tool, kind,
// evidence, safety level, side effects, the plan fields, and the three
// Changes lines, followed by the explicit-consent prompt.
func RenderPlan(w io.Writer, plan InstallPlan) {
	fmt.Fprintf(w, "Tool: %s\nKind: %s\nEvidence: %s\n", plan.Tool, plan.Kind, plan.Evidence)
	if plan.Level != "" {
		fmt.Fprintf(w, "Safety level: %s\n", plan.Level)
	}
	if plan.SideEffects != "" {
		fmt.Fprintf(w, "Side effects: %s\n", plan.SideEffects)
	}
	fmt.Fprintf(w, "\nPlan\n  platform: %s\n  method: %s\n  executable: %s\n  args: %s\n", runtime.GOOS, plan.Method, plan.Executable, strings.Join(plan.Args, " "))
	fmt.Fprint(w, "\nChanges\n  installs user-level/global executable\n  does NOT modify OpenCode\n  does NOT modify project dependencies\n\nProceed? [y/N] ")
}

// ConfirmConsent renders the plan to stdout, reads one line from r, and
// returns true only for an explicit affirmative answer. Any read failure or
// empty input declines: consent never defaults to yes.
func ConfirmConsent(r io.Reader, plan InstallPlan) (bool, error) {
	RenderPlan(os.Stdout, plan)
	var answer string
	if _, err := fmt.Fscanln(r, &answer); err != nil {
		return false, nil // unreadable or empty input declines
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// GlobalIntegrationPrompt asks the separate §47 opt-in question after a
// successful binary install, but only when the entry declares an opt-in
// OpenCode integration (rtk). The default is N and declining leaves global
// configuration untouched. A Y answer without a verified integration recipe
// fails with explicit remediation; the harness never writes global config
// directly. The prompt never appears during init — init has no install path.
func GlobalIntegrationPrompt(w io.Writer, r io.Reader, toolID string) error {
	for _, entry := range Catalog() {
		if entry.ID != toolID || entry.IntegrationMode != "opt-in" {
			continue
		}
		fmt.Fprint(w, "Enable global integration? [y/N] ")
		var answer string
		if _, err := fmt.Fscanln(r, &answer); err != nil {
			return nil // unreadable or empty input declines
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			return nil
		}
		return fmt.Errorf("no verified OpenCode integration recipe for %q in V1 (remediation: %s stays usable explicitly — e.g. %s status; nothing was modified)", toolID, toolID, toolID)
	}
	return nil
}

// Summary renders the compact default output.
func (p InstallPlan) Summary() string {
	return fmt.Sprintf("Install %s: %s %s (PM: %s)", p.Tool, p.Executable, strings.Join(p.Args, " "), p.PM)
}

// Summary renders the compact default output.
func (r InstallResult) Summary() string {
	return fmt.Sprintf("Install %s: executed=%t verified=%t", r.Tool, r.Executed, r.Verified)
}
