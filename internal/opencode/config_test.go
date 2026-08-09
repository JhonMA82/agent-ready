package opencode

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPlanConfigPreservesFixtureAndMode(t *testing.T) {
	in, err := os.ReadFile("testdata/config/preserve.jsonc")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/config/preserve.want.jsonc")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanConfig(nil, &ConfigFile{Data: in, Mode: 0o640})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Path != "opencode.jsonc" || plan.Mode != 0o640 || !bytes.Equal(plan.After, want) {
		t.Fatalf("plan = %#v\n%s", plan, plan.After)
	}
}

func TestPlanConfigSelectionAndValidation(t *testing.T) {
	file := func(s string) *ConfigFile { return &ConfigFile{Data: []byte(s), Mode: 0o600} }
	for _, tt := range []struct {
		name        string
		json, jsonc *ConfigFile
		path, want  string
		wantErr     string
	}{
		{"create JSON", nil, nil, "opencode.json", `{"skills":{"paths":["./.agent-ready/skills"]}}` + "\n", ""},
		{"single JSON", file(`{"x":1}`), nil, "opencode.json", `{"x":1,"skills":{"paths":["./.agent-ready/skills"]}}`, ""},
		{"both no owner prefer JSONC", file(`{"x":1}`), file("{\r\n  \"y\": 2\r\n}"), "opencode.jsonc", "{\r\n  \"y\": 2\r\n,\"skills\":{\"paths\":[\"./.agent-ready/skills\"]}}", ""},
		{"sole JSON owner", file(`{"skills":{}}`), file(`{"x":1}`), "opencode.json", `{"skills":{"paths":["./.agent-ready/skills"]}}`, ""},
		{"sole JSONC owner", file(`{"x":1}`), file(`{"skills":{}}`), "opencode.jsonc", `{"skills":{"paths":["./.agent-ready/skills"]}}`, ""},
		{"equivalent normalized", file(`{"skills":{"paths":[".agent-ready/skills/"]}}`), nil, "opencode.json", `{"skills":{"paths":[".agent-ready/skills/"]}}`, ""},
		{"equivalent escaped separators", file(`{"skills":{"paths":[".\\.agent-ready\\skills"]}}`), nil, "opencode.json", `{"skills":{"paths":[".\\.agent-ready\\skills"]}}`, ""},
		{"duplicate normalized", file(`{"skills":{"paths":["./.agent-ready/skills",".agent-ready/skills/"]}}`), nil, "", "", "duplicate normalized"},
		{"dual owners", file(`{"skills":{}}`), file(`{"skills":{}}`), "", "", "both define skills"},
		{"invalid non-owner", file(`{"skills":{}}`), file(`[`), "", "", "opencode.jsonc"},
		{"unsupported root", file(`[]`), nil, "", "", "root must be an object"},
		{"unsupported skills", file(`{"skills":[]}`), nil, "", "", "skills must be an object"},
		{"unsupported member", file(`{"skills":{"paths":[1]}}`), nil, "", "", "string array"},
		{"duplicate skills", file(`{"skills":{},"skills":{}}`), nil, "", "", "duplicate skills"},
		{"duplicate paths", file(`{"skills":{"paths":[],"paths":[]}}`), nil, "", "", "duplicate paths"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlanConfig(tt.json, tt.jsonc)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got.Path != tt.path || string(got.After) != tt.want {
				t.Fatalf("PlanConfig = %#v, %v", got, err)
			}
		})
	}
}

func FuzzPlanConfigPreservation(f *testing.F) {
	for _, seed := range []string{`{}`, `{"skills":{}}`, `{"skills":{"paths":[]}}`, "{\r\n// c\r\n\"x\":1,\r\n}"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		original := []byte(input)
		first, err := PlanConfig(&ConfigFile{Data: original, Mode: 0o640}, nil)
		if err != nil {
			return
		}
		if first.Path != "opencode.json" || first.Mode != 0o640 || !bytes.Equal(first.Before, original) {
			t.Fatalf("plan metadata or before bytes changed: %#v", first)
		}
		second, err := PlanConfig(&ConfigFile{Data: first.After, Mode: first.Mode}, nil)
		if err != nil || !bytes.Equal(first.After, second.After) {
			t.Fatalf("non-idempotent plan: %v\n%s\n%s", err, first.After, second.After)
		}
	})
}
