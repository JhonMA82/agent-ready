// Package lifecycle implements the remaining V1 public CLI surface
// (spec §30–§34): deterministic status facts, CLI-level doctor checks,
// ownership-preserving update, and ownership-driven removal.
package lifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/JhonMA82/agent-ready/internal/checkpoint"
	"github.com/JhonMA82/agent-ready/internal/tools"
)

// AssetInfo mirrors one manifest-owned asset entry.
type AssetInfo struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Mode   uint32 `json:"mode"`
}

// StatusFacts is the agent-ready.status/v1 schema.
type StatusFacts struct {
	SchemaVersion  string          `json:"schema_version"`
	Initialized    bool            `json:"initialized"`
	ManifestSchema string          `json:"manifest_schema,omitempty"`
	InstallVersion string          `json:"install_version,omitempty"`
	AssetCount     int             `json:"asset_count"`
	MismatchPaths  []string        `json:"mismatch_paths"`
	Checkpoint     *CheckpointFact `json:"checkpoint,omitempty"`
	ConfigFile     string          `json:"config_file,omitempty"`
	ToolsSummary   string          `json:"tools_summary,omitempty"`
}

// CheckpointFact is the latest checkpoint summary.
type CheckpointFact struct {
	ID       string `json:"id"`
	Stage    string `json:"stage"`
	Complete bool   `json:"complete"`
}

// StatusSchemaVersion is the agent-ready.status/v1 schema.
const StatusSchemaVersion = "agent-ready.status/v1"

// installedManifest reads the installed manifest when present.
func installedManifest(root string) (schema, version string, assets []AssetInfo, err error) {
	data, err := os.ReadFile(filepath.Join(root, ".agent-ready", "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil, nil
		}
		return "", "", nil, err
	}
	var m struct {
		Schema         string      `json:"schema"`
		InstallVersion string      `json:"install_version"`
		Assets         []AssetInfo `json:"assets"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return "", "", nil, err
	}
	return m.Schema, m.InstallVersion, m.Assets, nil
}

// AssetMatches reports whether the installed file matches the manifest entry
// (bytes and mode). Ownership is byte-level: any modification breaks the match.
func AssetMatches(root string, asset AssetInfo) bool {
	full := filepath.Join(root, filepath.FromSlash(asset.Path))
	info, err := os.Lstat(full)
	if err != nil || info.Mode().IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return false
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != asset.SHA256 {
		return false
	}
	return uint32(info.Mode().Perm()) == asset.Mode
}

// Status collects read-only lifecycle facts. Uninitialized is NOT an error
// (exit 0); read failures are.
func Status(root string) (StatusFacts, error) {
	facts := StatusFacts{SchemaVersion: StatusSchemaVersion}
	schema, version, assets, err := installedManifest(root)
	if err != nil {
		return StatusFacts{}, err
	}
	if schema == "" {
		return facts, nil
	}
	facts.Initialized = true
	facts.ManifestSchema = schema
	facts.InstallVersion = version
	facts.AssetCount = len(assets)
	for _, asset := range assets {
		if !AssetMatches(root, asset) {
			facts.MismatchPaths = append(facts.MismatchPaths, asset.Path)
		}
	}
	if status, err := checkpoint.Status(root); err == nil && status.Exists && status.Checkpoint != nil {
		facts.Checkpoint = &CheckpointFact{ID: status.Checkpoint.ID, Stage: status.Checkpoint.Stage, Complete: status.Checkpoint.Complete}
	}
	facts.ConfigFile = configPath(root)
	if toolFacts, err := tools.Status(); err == nil {
		facts.ToolsSummary = toolFacts.Summary()
	}
	return facts, nil
}

// Summary renders the compact default output.
func (f StatusFacts) Summary() string {
	if !f.Initialized {
		return "Not initialized (run agent-ready init)"
	}
	out := "Initialized: " + f.ManifestSchema + " v" + f.InstallVersion + " | assets: " + strconv.Itoa(f.AssetCount)
	if len(f.MismatchPaths) > 0 {
		out += " | mismatches: " + strconv.Itoa(len(f.MismatchPaths))
	}
	if f.Checkpoint != nil {
		out += " | checkpoint: " + f.Checkpoint.ID + " (" + f.Checkpoint.Stage + ")"
	}
	if f.ConfigFile != "" {
		out += " | config: " + f.ConfigFile
	}
	return out
}

// configPath returns the first present OpenCode config file.
func configPath(root string) string {
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return name
		}
	}
	return ""
}
