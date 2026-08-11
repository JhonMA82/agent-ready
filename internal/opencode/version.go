package opencode

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed compatibility.json
var compatibilityJSON []byte

var compatibility = func() map[string]string {
	var c map[string]string
	if err := json.Unmarshal(compatibilityJSON, &c); err != nil || c["min_version"] == "" || c["tested_version"] == "" || c["skills_paths_shape"] != "object" {
		panic("invalid embedded OpenCode compatibility metadata")
	}
	return c
}()

// MinimumVersion is the minimum compatible OpenCode version. Host runtimes at
// or above it are accepted; older runtimes fail closed with guidance so host
// updates never block init, audit, sync, or review.
func MinimumVersion() string { return compatibility["min_version"] }

// TestedVersion is the baseline the harness was validated against. Installed
// versions that drift from it are accepted with a non-blocking warning only.
func TestedVersion() string { return compatibility["tested_version"] }

// Result reports the accepted runtime after a successful preflight: the
// resolved binary and the installed version recorded as a fact, plus a
// non-blocking drift warning when the installed version differs from the
// tested baseline.
type Result struct {
	Binary  string
	Version string
	Drift   string
}

func Preflight(ctx context.Context, config []byte) (Result, error) {
	binary, err := exec.LookPath("opencode")
	if err != nil {
		return Result{}, fmt.Errorf("OpenCode %s or newer is required on PATH", MinimumVersion())
	}
	// Isolate the real probe so a successful preflight never mutates the
	// caller's HOME/XDG trees (real OpenCode creates cache/config/data/state).
	probe, err := os.MkdirTemp("", "agent-ready-opencode-probe-")
	if err != nil {
		return Result{}, fmt.Errorf("isolate OpenCode probe: %w", err)
	}
	defer os.RemoveAll(probe)
	cmd := exec.CommandContext(ctx, binary, "--version")
	cmd.Env = append(os.Environ(),
		"HOME="+probe,
		"XDG_CONFIG_HOME="+filepath.Join(probe, "config"),
		"XDG_CACHE_HOME="+filepath.Join(probe, "cache"),
		"XDG_DATA_HOME="+filepath.Join(probe, "data"),
		"XDG_STATE_HOME="+filepath.Join(probe, "state"),
	)
	out, err := cmd.Output()
	if err != nil {
		return Result{}, fmt.Errorf("read OpenCode version: %w", err)
	}
	version, err := parseVersion(string(out))
	if err != nil {
		return Result{}, err
	}
	if compareVersion(version, MinimumVersion()) < 0 {
		return Result{}, fmt.Errorf("unsupported OpenCode version %q; install at least %s", version, MinimumVersion())
	}
	if err := validateSchema(config); err != nil {
		return Result{}, err
	}
	result := Result{Binary: binary, Version: version}
	if version != TestedVersion() {
		result.Drift = fmt.Sprintf("OpenCode %s installed; tested baseline is %s", version, TestedVersion())
	}
	return result, nil
}

// parseVersion extracts the x.y.z semver token from the --version output. It
// tolerates host format drift (version banners, v prefixes, build suffixes)
// so runtime updates never block, but fails closed when no version can be
// determined.
func parseVersion(out string) (string, error) {
	for _, field := range strings.Fields(out) {
		token := strings.TrimPrefix(field, "v")
		if token == "" {
			continue
		}
		parts := strings.Split(token, ".")
		if len(parts) < 2 {
			continue
		}
		nums := make([]int, 3)
		valid := true
		for i := 0; i < 3; i++ {
			if i >= len(parts) {
				break
			}
			part := parts[i]
			if cut := strings.IndexAny(part, "-+"); cut >= 0 {
				part = part[:cut]
			}
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 {
				valid = false
				break
			}
			nums[i] = n
		}
		if !valid {
			continue
		}
		return fmt.Sprintf("%d.%d.%d", nums[0], nums[1], nums[2]), nil
	}
	return "", fmt.Errorf("unrecognized OpenCode version output %q", strings.TrimSpace(out))
}

// compareVersion returns -1, 0, or 1 comparing x.y.z versions numerically.
func compareVersion(a, b string) int {
	na, nb := versionParts(a), versionParts(b)
	for i := 0; i < 3; i++ {
		if na[i] < nb[i] {
			return -1
		}
		if na[i] > nb[i] {
			return 1
		}
	}
	return 0
}

func versionParts(v string) [3]int {
	var nums [3]int
	for i, part := range strings.Split(v, ".") {
		if i == 3 {
			break
		}
		if n, err := strconv.Atoi(part); err == nil {
			nums[i] = n
		}
	}
	return nums
}

func validateSchema(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 || data[0] != '{' {
		return fmt.Errorf("unsupported OpenCode config: expected object")
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("unsupported OpenCode config: %w", err)
	}
	skillsJSON := root["skills"]
	if len(skillsJSON) == 0 {
		return nil
	}
	if skillsJSON[0] != '{' {
		return fmt.Errorf("unsupported OpenCode skills schema: expected object")
	}
	var skills map[string]json.RawMessage
	if err := json.Unmarshal(skillsJSON, &skills); err != nil {
		return fmt.Errorf("unsupported OpenCode skills schema: %w", err)
	}
	pathsJSON := skills["paths"]
	if len(pathsJSON) == 0 {
		return nil
	}
	if pathsJSON[0] != '[' {
		return fmt.Errorf("unsupported OpenCode skills.paths schema: expected string array")
	}
	var paths []string
	if err := json.Unmarshal(pathsJSON, &paths); err != nil {
		return fmt.Errorf("unsupported OpenCode skills.paths schema: expected string array")
	}
	return nil
}
