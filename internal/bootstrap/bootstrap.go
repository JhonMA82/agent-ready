package bootstrap

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

//go:embed assets/manifest.json
var marker []byte

//go:embed assets/skills/agent-ready-orchestrator/SKILL.md
var skill []byte

//go:embed assets/commands/agent-ready.md
var command []byte

type File struct {
	Path          string
	Before, After []byte
	Mode          fs.FileMode
}

type manifest struct {
	Schema               string  `json:"schema"`
	InstallVersion       string  `json:"install_version"`
	CompatibilityVersion string  `json:"compatibility_version"`
	ConfigFile           string  `json:"config_file"`
	ConfigPath           string  `json:"config_path"`
	Assets               []asset `json:"assets"`
}
type asset struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Mode   uint32 `json:"mode"`
}

// Plan validates canonical ownership and returns deterministic desired files.
func Plan(root, config string) ([]File, error) {
	assets := []File{
		{Path: ".agent-ready/skills/agent-ready-orchestrator/SKILL.md", After: bytes.Clone(skill), Mode: 0o644},
		{Path: ".opencode/commands/agent-ready.md", After: bytes.Clone(command), Mode: 0o644},
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	desiredManifest, err := canonicalManifest(config, assets)
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(root, ".agent-ready", "manifest.json")
	manifestInfo, statErr := os.Lstat(manifestPath)
	owned := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, statErr
	}
	var existing []byte
	if owned {
		if manifestInfo.Mode()&os.ModeSymlink != 0 || manifestInfo.Mode().Perm() != 0o644 {
			return nil, fmt.Errorf("modified ownership manifest mode")
		}
		existing, err = os.ReadFile(manifestPath)
		if err != nil {
			return nil, err
		}
	}
	if owned && !bytes.Equal(existing, desiredManifest) {
		return nil, fmt.Errorf("modified or incompatible ownership manifest")
	}
	for i := range assets {
		full := filepath.Join(root, filepath.FromSlash(assets[i].Path))
		info, statErr := os.Lstat(full)
		if os.IsNotExist(statErr) {
			if owned {
				return nil, fmt.Errorf("owned target is missing: %s", assets[i].Path)
			}
			continue
		}
		if statErr != nil {
			return nil, statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("managed target is a symlink: %s", assets[i].Path)
		}
		data, readErr := os.ReadFile(full)
		if readErr != nil {
			return nil, readErr
		}
		if !owned || !bytes.Equal(data, assets[i].After) || info.Mode().Perm() != assets[i].Mode {
			return nil, fmt.Errorf("unowned or modified target: %s", assets[i].Path)
		}
		assets[i].Before = bytes.Clone(data)
	}
	assets = append(assets, File{Path: ".agent-ready/manifest.json", Before: bytes.Clone(existing), After: desiredManifest, Mode: 0o644})
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	return assets, nil
}

func canonicalManifest(config string, files []File) ([]byte, error) {
	var m manifest
	if err := json.Unmarshal(marker, &m); err != nil {
		return nil, fmt.Errorf("invalid embedded manifest marker: %w", err)
	}
	m.ConfigFile, m.ConfigPath = config, "./.agent-ready/skills"
	for _, file := range files {
		sum := sha256.Sum256(file.After)
		m.Assets = append(m.Assets, asset{Path: file.Path, SHA256: fmt.Sprintf("%x", sum), Mode: uint32(file.Mode)})
	}
	out, err := json.Marshal(m)
	return append(out, '\n'), err
}
