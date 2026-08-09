package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Selection struct {
	Invocation string
	Root       string
}

func Discover(ctx context.Context, invocation, gitBinary string) (Selection, error) {
	abs, err := filepath.Abs(invocation)
	if err != nil {
		return Selection{}, err
	}
	out, err := exec.CommandContext(ctx, gitBinary, "-C", abs, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return Selection{}, fmt.Errorf("not inside a Git worktree: %w", err)
	}
	root := strings.TrimSuffix(string(out), "\n")
	if !filepath.IsAbs(root) {
		return Selection{}, fmt.Errorf("Git returned non-absolute root %q", root)
	}
	return Selection{Invocation: filepath.Clean(abs), Root: filepath.Clean(root)}, nil
}

func Contained(root, relative string) (string, error) {
	clean := filepath.Clean(relative)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes repository: %q", relative)
	}
	current := filepath.Clean(root)
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("managed path contains symlink: %q", current)
		}
	}
	return current, nil
}
