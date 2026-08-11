package opencode

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

// bumpPatch returns the tested baseline with its patch version incremented so
// the above-floor case stays above the floor and differs from the baseline
// regardless of the embedded compatibility values.
func bumpPatch(v string) string {
	parts := strings.Split(v, ".")
	n, _ := strconv.Atoi(parts[len(parts)-1])
	parts[len(parts)-1] = strconv.Itoa(n + 1)
	return strings.Join(parts, ".")
}

func TestParseVersion(t *testing.T) {
	for _, tt := range []struct{ out, want string }{
		{"1.18.15\n", "1.18.15"},
		{"opencode 1.18.16\n", "1.18.16"},
		{"v1.19.0\n", "1.19.0"},
		{"1.18.16-1\n", "1.18.16"},
		{"1.18\n", "1.18.0"},
		{"not a version\n", ""},
	} {
		got, err := parseVersion(tt.out)
		if tt.want == "" && err == nil || tt.want != "" && (err != nil || got != tt.want) {
			t.Errorf("parseVersion(%q) = %q, %v; want %q", tt.out, got, err, tt.want)
		}
	}
}

func TestCompareVersion(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want int
	}{
		{"1.18.15", "1.18.15", 0},
		{"1.18.16", "1.18.15", 1},
		{"1.18.15", "1.18.16", -1},
		{"1.19.0", "1.18.15", 1},
		{"2.0.0", "1.99.99", 1},
	} {
		if got := compareVersion(tt.a, tt.b); got != tt.want {
			t.Errorf("compareVersion(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestPreflight(t *testing.T) {
	floor := fake(t, MinimumVersion())
	above := fake(t, bumpPatch(TestedVersion()))
	// Cases: exact floor match accepted, above-floor accepted with a recorded
	// fact and drift warning (not failure), below-floor fails closed with
	// guidance, missing binary fails closed, schema validation preserved.
	for _, tt := range []struct {
		path, config, want, version, drift string
	}{
		{floor, `{"skills":{"paths":["./skills"]}}`, "", MinimumVersion(), ""},
		{above, `{}`, "", bumpPatch(TestedVersion()), TestedVersion()},
		{fake(t, "1.18.14"), `{}`, "install at least " + MinimumVersion(), "", ""},
		{fake(t, "0.0.0"), `{}`, "install at least", "", ""},
		{floor, `{"skills":["./skills"]}`, "skills", "", ""},
		{t.TempDir(), `{}`, "PATH", "", ""},
	} {
		t.Setenv("PATH", tt.path)
		res, err := Preflight(context.Background(), []byte(tt.config))
		if tt.want == "" && err != nil || tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("want %q, error %v", tt.want, err)
		}
		if tt.want == "" {
			if res.Version != tt.version || (res.Drift != "") != (tt.drift != "") {
				t.Errorf("version=%q drift=%q, want version %q drift %q", res.Version, res.Drift, tt.version, tt.drift)
			}
			if tt.drift != "" && !strings.Contains(res.Drift, tt.drift) {
				t.Errorf("drift warning must name the tested baseline %q: %q", tt.drift, res.Drift)
			}
			if res.Binary == "" {
				t.Error("successful preflight must return the binary path")
			}
		}
	}
}

// TestPreflightIsolatesGlobalTrees proves a probe child that writes
// cache/config/data/state like real OpenCode leaves caller trees unchanged.
func TestPreflightIsolatesGlobalTrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fake is Unix-only")
	}
	home, xdg := t.TempDir(), t.TempDir()
	globalEnv(t, home, xdg)
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' '" + TestedVersion() + "'\nmkdir -p \"$HOME/.cache/opencode\" \"$XDG_DATA_HOME/opencode\" \"$XDG_STATE_HOME/opencode\" || exit 1\n: > \"$HOME/.cache/opencode/probe\" || exit 1\n"
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

// TestPreflightRealOpenCodeSmoke proves the real installed OpenCode probe
// leaves caller global trees intact when the runtime is at or above the
// minimum compatible floor (any host version, not just the tested baseline).
func TestPreflightRealOpenCodeSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("real OpenCode smoke is not a short test")
	}
	binary, err := exec.LookPath("opencode")
	if err != nil {
		t.Skip("real OpenCode not installed")
	}
	out, err := exec.Command(binary, "--version").Output()
	if err != nil {
		t.Skipf("real OpenCode version probe failed: %v", err)
	}
	version, err := parseVersion(string(out))
	if err != nil {
		t.Skipf("real OpenCode version %q unrecognized", strings.TrimSpace(string(out)))
	}
	if compareVersion(version, MinimumVersion()) < 0 {
		t.Skipf("real OpenCode %s is below the %s floor", version, MinimumVersion())
	}
	home, xdg := t.TempDir(), t.TempDir()
	globalEnv(t, home, xdg)
	res, err := Preflight(context.Background(), []byte(`{"skills":{"paths":["./skills"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Version != version {
		t.Fatalf("preflight recorded %q, want %q", res.Version, version)
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
