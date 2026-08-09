package opencode

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fake(t *testing.T, version string) string {
	if runtime.GOOS == "windows" {
		t.Skip("shell fake is Unix-only")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "opencode"), []byte("#!/bin/sh\nprintf '%s\\n' '"+version+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPreflight(t *testing.T) {
	compatible := fake(t, RequiredVersion())
	for _, tt := range []struct{ path, config, want string }{{compatible, `{"skills":{"paths":["./skills"]}}`, ""}, {fake(t, "0.0.0"), `{}`, RequiredVersion()}, {compatible, `{"skills":["./skills"]}`, "skills"}, {t.TempDir(), `{}`, "PATH"}} {
		t.Setenv("PATH", tt.path)
		_, err := Preflight(context.Background(), []byte(tt.config))
		if tt.want == "" && err != nil || tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("want %q, error %v", tt.want, err)
		}
	}
}

// TestPreflightIsolatesGlobalTrees proves a probe child that writes
// cache/config/data/state like real OpenCode 1.18.15 leaves caller trees unchanged.
func TestPreflightIsolatesGlobalTrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fake is Unix-only")
	}
	home, xdg := t.TempDir(), t.TempDir()
	globalEnv(t, home, xdg)
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' '" + RequiredVersion() + "'\nmkdir -p \"$HOME/.cache/opencode\" \"$XDG_DATA_HOME/opencode\" \"$XDG_STATE_HOME/opencode\" || exit 1\n: > \"$HOME/.cache/opencode/probe\" || exit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	if _, err := Preflight(context.Background(), []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	assertUnchanged(t, home, "HOME")
	assertUnchanged(t, xdg, "XDG")
}

// TestPreflightRealOpenCodeSmoke proves the real pinned OpenCode probe leaves
// caller global trees intact where the exact version is available.
func TestPreflightRealOpenCodeSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("real OpenCode smoke is not a short test")
	}
	binary, err := exec.LookPath("opencode")
	if err != nil {
		t.Skip("real OpenCode not installed")
	}
	out, err := exec.Command(binary, "--version").Output()
	if err != nil || strings.TrimSuffix(string(out), "\n") != RequiredVersion() {
		t.Skipf("real OpenCode %q is not the pinned %s", strings.TrimSpace(string(out)), RequiredVersion())
	}
	home, xdg := t.TempDir(), t.TempDir()
	globalEnv(t, home, xdg)
	if _, err := Preflight(context.Background(), []byte(`{"skills":{"paths":["./skills"]}}`)); err != nil {
		t.Fatal(err)
	}
	assertUnchanged(t, home, "HOME")
	assertUnchanged(t, xdg, "XDG")
}

func globalEnv(t *testing.T, home, xdg string) {
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(xdg, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(xdg, "cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(xdg, "data"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(xdg, "state"))
}

// assertUnchanged fails if any path was added or modified under root
// (the caller trees start empty, so any entry is a global mutation).
func assertUnchanged(t *testing.T, root, label string) {
	t.Helper()
	if got := treePaths(t, root); len(got) != 0 {
		t.Fatalf("%s tree mutated: %v", label, got)
	}
}

func treePaths(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		if !d.IsDir() {
			if rel, err := filepath.Rel(root, path); err == nil {
				paths = append(paths, rel)
			}
		}
		return nil
	})
	return paths
}
