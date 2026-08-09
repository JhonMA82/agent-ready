package opencode

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed compatibility.json
var compatibilityJSON []byte

var compatibility = func() map[string]string {
	var c map[string]string
	if err := json.Unmarshal(compatibilityJSON, &c); err != nil || c["version"] == "" || c["skills_paths_shape"] != "object" {
		panic("invalid embedded OpenCode compatibility metadata")
	}
	return c
}()

func RequiredVersion() string { return compatibility["version"] }

func Preflight(ctx context.Context, config []byte) (string, error) {
	binary, err := exec.LookPath("opencode")
	if err != nil {
		return "", fmt.Errorf("OpenCode %s is required on PATH", RequiredVersion())
	}
	// Isolate the real probe so a successful preflight never mutates the
	// caller's HOME/XDG trees (real OpenCode 1.18.15 creates cache/config/data/state).
	probe, err := os.MkdirTemp("", "agent-ready-opencode-probe-")
	if err != nil {
		return "", fmt.Errorf("isolate OpenCode probe: %w", err)
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
		return "", fmt.Errorf("read OpenCode version: %w", err)
	}
	version := strings.TrimSuffix(string(out), "\n")
	if version != RequiredVersion() {
		return "", fmt.Errorf("unsupported OpenCode version %q; install %s", version, RequiredVersion())
	}
	if err := validateSchema(config); err != nil {
		return "", err
	}
	return binary, nil
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
