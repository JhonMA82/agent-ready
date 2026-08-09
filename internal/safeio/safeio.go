package safeio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/gentle-ai/agent-ready/internal/app"
)

const journalName = ".agent-ready/transaction.json"

type Hook func(phase, path string) error

type Options struct{ Hook Hook }

type Result struct{ RecoveryPath string }

type Error struct {
	Phase, Path, RecoveryPath string
	Cause                     error
}

func (e *Error) Error() string {
	if e.RecoveryPath != "" {
		return fmt.Sprintf("%s %s: %v; recovery required at %s", e.Phase, e.Path, e.Cause, e.RecoveryPath)
	}
	return fmt.Sprintf("%s %s: %v", e.Phase, e.Path, e.Cause)
}
func (e *Error) Unwrap() error { return e.Cause }

type entry struct {
	Path    string      `json:"path"`
	Before  []byte      `json:"before,omitempty"`
	Mode    fs.FileMode `json:"mode,omitempty"`
	Existed bool        `json:"existed"`
	Applied bool        `json:"applied"`
}
type journal struct {
	Schema  string  `json:"schema"`
	Entries []entry `json:"entries"`
}

func Commit(p app.Plan, opts Options) (Result, error) {
	changes := ordered(p.Changes())
	for _, c := range changes {
		if c.Kind() != "noop" {
			if err := validate(p.Root(), c); err != nil {
				return Result{}, &Error{Phase: "revalidate", Path: c.Path(), Cause: err}
			}
		}
	}
	j := journal{Schema: "agent-ready.transaction/v1"}
	active := make([]app.Change, 0, len(changes))
	for _, c := range changes {
		if c.Kind() == "noop" {
			continue
		}
		active = append(active, c)
		j.Entries = append(j.Entries, entry{Path: c.Path(), Before: c.Before(), Mode: c.Mode(), Existed: c.Before() != nil})
	}
	if len(active) == 0 {
		return Result{}, nil
	}
	journalPath := filepath.Join(p.Root(), filepath.FromSlash(journalName))
	if err := writeJournal(journalPath, j); err != nil {
		return Result{}, &Error{Phase: "journal", Path: journalName, Cause: err}
	}
	for i, c := range active {
		j.Entries[i].Applied = true
		if err := writeJournal(journalPath, j); err == nil {
			err = call(opts.Hook, "commit", c.Path())
			if err == nil {
				err = replace(filepath.Join(p.Root(), filepath.FromSlash(c.Path())), c.After(), c.Mode())
			}
			if err == nil {
				continue
			}
			return rollbackFailure(p.Root(), j, opts, "commit", c.Path(), err)
		} else {
			return rollbackFailure(p.Root(), j, opts, "journal", journalName, err)
		}
	}
	if err := removeJournal(journalPath); err != nil {
		return rollbackFailure(p.Root(), j, opts, "journal", journalName, err)
	}
	return Result{}, nil
}

func Recover(root string, opts Options) (Result, error) {
	path := filepath.Join(root, filepath.FromSlash(journalName))
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, &Error{Phase: "recovery", Path: journalName, Cause: err}
	}
	var j journal
	if err := json.Unmarshal(data, &j); err != nil || j.Schema != "agent-ready.transaction/v1" {
		if err == nil {
			err = errors.New("unsupported recovery journal")
		}
		return Result{}, &Error{Phase: "recovery", Path: journalName, Cause: err, RecoveryPath: path}
	}
	if err := rollback(root, j, opts); err != nil {
		return Result{}, &Error{Phase: "recovery", Path: journalName, Cause: err, RecoveryPath: path}
	}
	if err := removeJournal(path); err != nil {
		return Result{}, &Error{Phase: "recovery", Path: journalName, Cause: err, RecoveryPath: path}
	}
	return Result{}, nil
}

func ordered(changes []app.Change) []app.Change {
	sort.Slice(changes, func(i, k int) bool {
		im, km := changes[i].Path() == ".agent-ready/manifest.json", changes[k].Path() == ".agent-ready/manifest.json"
		return im != km && !im || im == km && changes[i].Path() < changes[k].Path()
	})
	return changes
}

func validate(root string, c app.Change) error {
	path := filepath.Join(root, filepath.FromSlash(c.Path()))
	info, err := os.Lstat(path)
	if c.Before() == nil {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		return errors.New("expected target to be absent")
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != c.Mode() {
		return errors.New("mode or file type changed after planning")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, c.Before()) {
		return errors.New("bytes changed after planning")
	}
	return nil
}

func replace(path string, data []byte, mode fs.FileMode) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".agent-ready-txn-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if _, err = f.Write(data); err == nil {
		err = f.Chmod(mode)
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err == nil {
		err = syncDir(filepath.Dir(path))
	}
	return err
}

func writeJournal(path string, j journal) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return replace(path, append(data, '\n'), 0o600)
}

func removeJournal(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func rollbackFailure(root string, j journal, opts Options, phase, path string, cause error) (Result, error) {
	journalPath := filepath.Join(root, filepath.FromSlash(journalName))
	if err := rollback(root, j, opts); err != nil {
		return Result{RecoveryPath: journalPath}, &Error{Phase: phase, Path: path, Cause: errors.Join(cause, err), RecoveryPath: journalPath}
	}
	if err := removeJournal(journalPath); err != nil {
		return Result{RecoveryPath: journalPath}, &Error{Phase: phase, Path: path, Cause: errors.Join(cause, err), RecoveryPath: journalPath}
	}
	return Result{}, &Error{Phase: phase, Path: path, Cause: cause}
}

func rollback(root string, j journal, opts Options) error {
	for i := len(j.Entries) - 1; i >= 0; i-- {
		e := j.Entries[i]
		if !e.Applied {
			continue
		}
		if err := call(opts.Hook, "rollback", e.Path); err != nil {
			return err
		}
		path := filepath.Join(root, filepath.FromSlash(e.Path))
		if e.Existed {
			if err := replace(path, e.Before, e.Mode); err != nil {
				return err
			}
		} else {
			err := os.Remove(path)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err == nil {
				if err := syncDir(filepath.Dir(path)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func call(h Hook, phase, path string) error {
	if h != nil {
		return h(phase, path)
	}
	return nil
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
