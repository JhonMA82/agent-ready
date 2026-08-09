package bootstrap

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed assets
var assetsFS embed.FS

// walkAssets walks the embedded assets tree, returning the manifest marker
// and every other asset routed to its installed location.
func walkAssets() ([]byte, []File, error) {
	var marker []byte
	var files []File
	err := fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := assetsFS.ReadFile(path)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, "assets/")
		if rel == "manifest.json" {
			marker = bytes.Clone(data)
			return nil
		}
		target, err := route(rel)
		if err != nil {
			return err
		}
		files = append(files, File{Path: target, After: data, Mode: 0o644})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return marker, files, nil
}

// route maps an embedded asset path to its installed target.
func route(rel string) (string, error) {
	trees := []struct{ src, dst string }{
		{"skills/", ".agent-ready/skills/"},
		{"references/", ".agent-ready/references/"},
		{"commands/", ".opencode/commands/"},
	}
	for _, tree := range trees {
		if strings.HasPrefix(rel, tree.src) {
			return filepath.FromSlash(tree.dst + strings.TrimPrefix(rel, tree.src)), nil
		}
	}
	return "", fmt.Errorf("embedded asset outside routed trees: assets/%s", rel)
}

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
	marker, assets, err := walkAssets()
	if err != nil {
		return nil, err
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	desiredManifest, err := canonicalManifest(marker, config, assets)
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

func canonicalManifest(marker []byte, config string, files []File) ([]byte, error) {
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
