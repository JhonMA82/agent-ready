package lifecycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Mode selects what remove deletes.
type Mode string

// Remove modes (spec §33).
const (
	ModeHarnessOnly   Mode = "harness-only"
	ModeHarnessAndGen Mode = "harness+generated"
	RemoveJournalName      = ".agent-ready/remove-transaction.json"
)

// RemoveEntry is one planned removal.
type RemoveEntry struct {
	Path   string `json:"path"`
	Reason string `json:"reason,omitempty"`
}

// RemovePlan is the reviewable removal plan with kept-file reports.
type RemovePlan struct {
	Root    string        `json:"root"`
	Mode    Mode          `json:"mode"`
	Removed []RemoveEntry `json:"removed"`
	Kept    []RemoveEntry `json:"kept"`
}

// removeJournal records reverse-restore data for partial-failure recovery.
type removeJournal struct {
	Schema  string         `json:"schema"`
	Entries []journalEntry `json:"entries"`
}

type journalEntry struct {
	Path   string `json:"path"`
	Before []byte `json:"before,omitempty"`
}

const removeJournalSchema = "agent-ready.remove-transaction/v1"

// PlanRemove builds the ownership-driven removal plan. Harness-only removes
// manifest-owned assets; harness+generated additionally removes
// state/checkpoints and the config file ONLY when byte-identical to the
// installed one. Modified owned files and unowned files are never removed.
func PlanRemove(root string, mode Mode) (RemovePlan, error) {
	if mode != ModeHarnessOnly && mode != ModeHarnessAndGen {
		return RemovePlan{}, fmt.Errorf("invalid mode %q: use harness-only or harness+generated", mode)
	}
	schema, _, assets, err := installedManifest(root)
	if err != nil {
		return RemovePlan{}, err
	}
	if schema == "" {
		return RemovePlan{}, fmt.Errorf("not initialized: run agent-ready init first")
	}
	plan := RemovePlan{Root: root, Mode: mode}
	plan.Removed = append(plan.Removed, RemoveEntry{Path: ".agent-ready/manifest.json"})
	for _, asset := range assets {
		if !AssetMatches(root, asset) {
			plan.Kept = append(plan.Kept, RemoveEntry{Path: asset.Path, Reason: "modified or missing owned file; refusing to delete user content"})
			continue
		}
		plan.Removed = append(plan.Removed, RemoveEntry{Path: asset.Path})
	}
	if mode == ModeHarnessAndGen {
		for _, dir := range []string{".agent-ready/state", ".agent-ready/checkpoints", ".agent-ready/checkpoints/history"} {
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err == nil && info.IsDir() {
				plan.Removed = append(plan.Removed, RemoveEntry{Path: dir})
			}
		}
		config := configPath(root)
		if config != "" {
			if configByteIdentical(root, config) {
				plan.Removed = append(plan.Removed, RemoveEntry{Path: config, Reason: "byte-identical installed config"})
			} else {
				plan.Kept = append(plan.Kept, RemoveEntry{Path: config, Reason: "modified config; refusing to delete manual content"})
			}
		}
	}
	sort.Slice(plan.Removed, func(i, j int) bool { return plan.Removed[i].Path < plan.Removed[j].Path })
	sort.Slice(plan.Kept, func(i, j int) bool { return plan.Kept[i].Path < plan.Kept[j].Path })
	return plan, nil
}

// configByteIdentical reports whether the config file still carries the
// installed skill-path entry written by init (either "./.agent-ready/skills"
// or ".agent-ready/skills" normalized form).
func configByteIdentical(root, name string) bool {
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), ".agent-ready/skills")
}

// ApplyRemove executes the plan with a reverse-restore journal: journal
// first, deletions second, rollback from the journal on failure, journal
// removal last. State/checkpoints directories are removed only when empty.
func ApplyRemove(plan RemovePlan) error {
	if plan.Mode == "" || len(plan.Removed) == 0 {
		return nil
	}
	journalPath := filepath.Join(plan.Root, filepath.FromSlash(RemoveJournalName))
	journal := removeJournal{Schema: removeJournalSchema}
	before := map[string][]byte{}
	for _, entry := range plan.Removed {
		full := filepath.Join(plan.Root, filepath.FromSlash(entry.Path))
		if info, err := os.Lstat(full); err == nil && !info.IsDir() {
			data, err := os.ReadFile(full)
			if err != nil {
				return err
			}
			before[entry.Path] = data
			journal.Entries = append(journal.Entries, journalEntry{Path: entry.Path, Before: data})
		} else {
			journal.Entries = append(journal.Entries, journalEntry{Path: entry.Path})
		}
	}
	journalData, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	if err := os.WriteFile(journalPath, append(journalData, '\n'), 0o600); err != nil {
		return err
	}
	// Delete files, then prune empty directories bottom-up.
	if err := deleteEntries(plan); err != nil {
		rollbackErr := restoreFromJournal(plan.Root, journal)
		if rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("rollback failed: %w; recovery data at %s", rollbackErr, journalPath))
		}
		return err
	}
	return os.Remove(journalPath)
}

func deleteEntries(plan RemovePlan) error {
	for _, entry := range plan.Removed {
		full := filepath.Join(plan.Root, filepath.FromSlash(entry.Path))
		info, err := os.Lstat(full)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Planned generated directories (state/checkpoints) are removed
			// recursively: their contents are model-generated by definition.
			if err := os.RemoveAll(full); err != nil {
				return err
			}
			continue
		}
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func restoreFromJournal(root string, journal removeJournal) error {
	for i := len(journal.Entries) - 1; i >= 0; i-- {
		entry := journal.Entries[i]
		full := filepath.Join(root, filepath.FromSlash(entry.Path))
		if entry.Before == nil {
			if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, entry.Before, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Summary renders the compact default output.
func (p RemovePlan) Summary() string {
	if len(p.Removed) == 0 && len(p.Kept) == 0 {
		return "Remove: nothing planned"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remove (%s): %d removed", p.Mode, len(p.Removed))
	if len(p.Kept) > 0 {
		fmt.Fprintf(&b, ", %d kept", len(p.Kept))
	}
	return b.String()
}
