// Package checkpoint owns Go-side checkpoint envelope files (spec R10):
// `save` snapshots stage + inventory hashes, `status` reads them, and
// `changes` diffs the current inventory against the latest snapshot (R9).
// Semantic state files are model-owned and never written here.
package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JhonMA82/agent-ready/internal/inventory"
)

const (
	SchemaVersion        = "agent-ready.checkpoint/v1"
	ChangesSchemaVersion = "agent-ready.changes/v1"
)

// dir is the Go-owned checkpoint storage under the repo root (design D4).
const dir = ".agent-ready/checkpoints"

// Envelope is the agent-ready.checkpoint/v1 envelope (D7); latest.json and
// history/<id>.json are byte-identical copies.
type Envelope struct {
	SchemaVersion   string            `json:"schema_version"`
	ID              string            `json:"id"`
	Stage           string            `json:"stage"`
	Complete        bool              `json:"complete"`
	SourceHash      string            `json:"source_hash"`
	InventoryHashes map[string]string `json:"inventory_hashes"`
	CreatedAt       string            `json:"created_at"`
}

// StatusFacts reports whether a checkpoint exists (envelope attached only
// when one is recorded).
type StatusFacts struct {
	SchemaVersion string    `json:"schema_version"`
	Exists        bool      `json:"exists"`
	Checkpoint    *Envelope `json:"checkpoint,omitempty"`
}

// Change is one path-level difference (D5).
type Change struct {
	Path string `json:"path"`
	Kind string `json:"kind"` // added | removed | changed
}

// Baseline identifies the checkpoint a change set is diffed against.
type Baseline struct {
	Exists       bool   `json:"exists"`
	CheckpointID string `json:"checkpoint_id,omitempty"`
	SourceHash   string `json:"source_hash,omitempty"`
}

// Facts is the agent-ready.changes/v1 schema (D5/D6); changes sorted by path.
type Facts struct {
	SchemaVersion string   `json:"schema_version"`
	Baseline      Baseline `json:"baseline"`
	FirstRun      bool     `json:"first_run"`
	Changes       []Change `json:"changes"`
}

// Save snapshots inventory hashes into a new envelope written as a
// byte-identical pair latest.json + history/<id>.json, creating the
// checkpoint directories when absent.
func Save(root, stage string, complete bool) (Envelope, error) {
	if stage == "" {
		return Envelope{}, errors.New("stage is required")
	}
	hashes, err := snapshot(root)
	if err != nil {
		return Envelope{}, err
	}
	for _, d := range []string{filepath.Join(root, dir), filepath.Join(root, dir, "history")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return Envelope{}, fmt.Errorf("create checkpoint directory: %w", err)
		}
	}
	id, err := nextID(root)
	if err != nil {
		return Envelope{}, err
	}
	env := Envelope{SchemaVersion: SchemaVersion, ID: id, Stage: stage, Complete: complete, SourceHash: sourceHash(hashes), InventoryHashes: hashes, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return Envelope{}, err
	}
	data = append(data, '\n')
	for _, path := range []string{filepath.Join(root, dir, "history", id+".json"), filepath.Join(root, dir, "latest.json")} {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return Envelope{}, err
		}
	}
	return env, nil
}

// Load reads the latest checkpoint envelope.
func Load(root string) (Envelope, error) {
	data, err := os.ReadFile(filepath.Join(root, dir, "latest.json"))
	if err != nil {
		return Envelope{}, err
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("parse checkpoint envelope: %w", err)
	}
	return env, nil
}

// Status reports the latest checkpoint; Exists is false when none is saved.
func Status(root string) (StatusFacts, error) {
	env, err := Load(root)
	if os.IsNotExist(err) {
		return StatusFacts{SchemaVersion: SchemaVersion}, nil
	}
	if err != nil {
		return StatusFacts{}, err
	}
	return StatusFacts{SchemaVersion: SchemaVersion, Exists: true, Checkpoint: &env}, nil
}

// Changes diffs current inventory hashes against the latest checkpoint (D6):
// no baseline -> every path added with first_run true and no error; otherwise
// paths are added/removed/changed by hash comparison only.
func Changes(root string) (Facts, error) {
	current, err := snapshot(root)
	if err != nil {
		return Facts{}, err
	}
	env, err := Load(root)
	if os.IsNotExist(err) {
		paths := make([]string, 0, len(current))
		for path := range current {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		changes := make([]Change, 0, len(paths))
		for _, path := range paths {
			changes = append(changes, Change{Path: path, Kind: "added"})
		}
		return Facts{SchemaVersion: ChangesSchemaVersion, FirstRun: true, Changes: changes}, nil
	}
	if err != nil {
		return Facts{}, err
	}
	changes := []Change{}
	for path, hash := range current {
		if base, ok := env.InventoryHashes[path]; !ok {
			changes = append(changes, Change{Path: path, Kind: "added"})
		} else if base != hash {
			changes = append(changes, Change{Path: path, Kind: "changed"})
		}
	}
	for path := range env.InventoryHashes {
		if _, ok := current[path]; !ok {
			changes = append(changes, Change{Path: path, Kind: "removed"})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return Facts{SchemaVersion: ChangesSchemaVersion, Baseline: Baseline{Exists: true, CheckpointID: env.ID, SourceHash: env.SourceHash}, Changes: changes}, nil
}

// snapshot hashes every inventory file except the harness-owned trees
// (.agent-ready, .opencode) so checkpoint and model artifacts never
// invalidate their own baseline.
func snapshot(root string) (map[string]string, error) {
	paths, err := inventory.Paths(root)
	if err != nil {
		return nil, err
	}
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, ".agent-ready/") || strings.HasPrefix(path, ".opencode/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		hashes[path] = hex.EncodeToString(sum[:])
	}
	return hashes, nil
}

// sourceHash digests the sorted path:hash lines into one envelope fingerprint.
func sourceHash(hashes map[string]string) string {
	paths := make([]string, 0, len(hashes))
	for path := range hashes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, path := range paths {
		fmt.Fprintf(&b, "%s:%s\n", path, hashes[path])
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// nextID returns the next zero-padded history id ("0001", ...).
func nextID(root string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(root, dir, "history"))
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	next := 1
	for _, entry := range entries {
		var n int
		if _, err := fmt.Sscanf(strings.TrimSuffix(entry.Name(), ".json"), "%d", &n); err == nil && n >= next {
			next = n + 1
		}
	}
	return fmt.Sprintf("%04d", next), nil
}

func (e Envelope) Summary() string {
	return fmt.Sprintf("Checkpoint: %s | Stage: %s | Complete: %t | Files: %d", e.ID, e.Stage, e.Complete, len(e.InventoryHashes))
}

func (s StatusFacts) Summary() string {
	if !s.Exists {
		return "No checkpoint"
	}
	return s.Checkpoint.Summary()
}

func (f Facts) Summary() string {
	if f.FirstRun {
		return fmt.Sprintf("First run: %d files added", len(f.Changes))
	}
	if len(f.Changes) == 0 {
		return "No changes"
	}
	return fmt.Sprintf("Changes: %d (see --json)", len(f.Changes))
}
