package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDrivenFixtures drives the §32 NixOS Wizard and §33 Laravel cohorts
func TestDrivenFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("driven audit executes the real OpenCode runtime")
	}
	model := os.Getenv("AGENT_READY_DRIVEN_MODEL")
	if model == "" {
		t.Skip("AGENT_READY_DRIVEN_MODEL unset; driven audit proof requires a model")
	}
	if _, err := exec.LookPath("opencode"); err != nil {
		t.Skipf("opencode binary required on PATH: %v", err)
	}
	authSrc := filepath.Join(os.Getenv("HOME"), ".local", "share", "opencode", "auth.json")
	if _, err := os.Stat(authSrc); err != nil {
		t.Skipf("opencode auth %s required for the driven run: %v", authSrc, err)
	}
	base := t.TempDir()
	binDir, binary := filepath.Join(base, "bin"), filepath.Join(base, "agent-ready")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/agent-ready")
	build.Dir = filepath.Clean(filepath.Join("..", ".."))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	home := filepath.Join(base, "home")
	authDir := filepath.Join(home, ".local", "share", "opencode")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatal(err)
	}
	authData, err := os.ReadFile(authSrc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth.json"), authData, 0o600); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME="+filepath.Join(home, ".local", "state"),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	cohorts := []struct {
		name     string
		required []string
	}{
		{"nixos-wizard", []string{"rust", "cargo", "cargo.lock", "nix", "flake.lock", "ratatui", "0.29", "pnpm"}},
		{"laravel", []string{"php", "laravel", "composer", "bun"}},
	}
	for _, cohort := range cohorts {
		t.Run(cohort.name, func(t *testing.T) {
			repo := filepath.Join(base, "repo-"+cohort.name)
			copyFixture(t, filepath.Join("testdata", "acceptance", "driven", cohort.name), repo)
			if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			if got, err := gitRoot("git", repo); err != nil || got != repo {
				t.Fatalf("gitRoot(%q) = %q, %v", repo, got, err)
			}
			if out, code := runCommand(t, binary, repo, env, "init", "--json"); code != 0 {
				t.Fatalf("init: code=%d %s", code, out)
			}
			baseline := snapshot(t, repo)
			ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "opencode", "run", "--dir", repo, "--model", model,
				"--format", "json", "--command", "agent-ready", "audit")
			cmd.Dir, cmd.Env = repo, env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("opencode run: %v\n%s", err, out)
			}
			events := observeAuditJSONL(t, out)
			after := snapshot(t, repo)
			if data, err := os.ReadFile(filepath.Join(repo, ".agent-ready", "state", "decisions.jsonl")); err != nil || len(data) == 0 {
				t.Fatalf("audit did not write state decisions.jsonl: %v", err)
			}
			doc := auditDocument(events, baseline, after)
			assertAuditStructure(t, events, doc)
			lower := strings.ToLower(doc)
			for _, token := range cohort.required {
				if !strings.Contains(lower, token) {
					t.Fatalf("cohort %s missing structural token %q", cohort.name, token)
				}
			}
			assertCohortOracle(t, cohort.name, lower, after)
		})
	}
}

// assertCohortOracle is the per-cohort structural oracle: §32 stale pnpm
func assertCohortOracle(t *testing.T, mode, lower string, after map[string]string) {
	t.Helper()
	for path := range after {
		if !strings.HasPrefix(path, ".opencode/skills/") {
			continue
		}
		for _, generic := range []string{"rust", "generic", "language", "scaffold", "best-practices"} {
			if strings.Contains(strings.ToLower(path), generic) {
				t.Fatalf("cohort %s created a generic skill artifact: %s", mode, path)
			}
		}
	}
	switch mode {
	case "nixos":
		if !containsAny(lower, []string{"stale", "outdated", "obsolete", "no longer", "no package.json", "not present"}) {
			t.Fatal("NixOS Wizard cohort must flag the pnpm guidance as stale")
		}
		if !containsAny(lower, []string{"external", "verified", "verification", "documentation", "crates.io", "cargo doc"}) {
			t.Fatal("NixOS Wizard cohort must consider external verification for Ratatui")
		}
	case "laravel":
		if strings.Contains(lower, "npm") && !strings.Contains(lower, "bun") {
			t.Fatal("Laravel cohort must not resolve the frontend to npm when bun.lock exists")
		}
		if !containsAny(lower, []string{"bun", "javascript", "node"}) {
			t.Fatal("Laravel cohort must discuss the JS frontend alongside PHP")
		}
	}
}
func containsAny(lower string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
