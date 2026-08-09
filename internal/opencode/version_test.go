package opencode

import (
	"context"
	"os"
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
