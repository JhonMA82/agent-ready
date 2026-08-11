package tools

import (
	"encoding/json"
	"github.com/JhonMA82/agent-ready/internal/inventory"
	"strings"
	"testing"
)

func stableBytes(t *testing.T, root string) string {
	t.Helper()
	first, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) {
		t.Fatal("recommend must be byte-stable across runs")
	}
	return string(a)
}

// TestMonorepoOracle (§43): a multi-workspace monorepo emits provider
func TestMonorepoOracle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pnpm-workspace.yaml", "packages:\n  - packages/*\n")
	writeFile(t, root, "go.work", "go 1.24\n\nuse ./services/api\n")
	writeFile(t, root, "packages/a/package.json", "{\"name\":\"a\",\"dependencies\":{}}\n")
	writeFile(t, root, "packages/b/package.json", "{\"name\":\"b\",\"dependencies\":{}}\n")
	writeFile(t, root, "services/api/go.mod", "module example.com/api\n\ngo 1.24\n")
	writeFile(t, root, "services/api/main.go", "package main\n\nfunc main() {}\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]Candidate{}
	for _, c := range facts.Candidates {
		byID[c.CatalogID] = c
	}
	codegraph, ok := byID["codegraph"]
	if !ok || codegraph.Capability != "dependency_graph" || !strings.Contains(codegraph.Observed, "pnpm-workspace.yaml") {
		t.Fatalf("codegraph must carry workspace evidence: %+v", facts.Candidates)
	}
	serena, ok := byID["serena"]
	if !ok || serena.Capability != "symbol_intelligence" || !strings.Contains(serena.Observed, "workspace_signals") {
		t.Fatalf("serena must carry LSP surface and module-boundary evidence: %+v", facts.Candidates)
	}
	if _, forced := byID["semble"]; forced {
		t.Fatalf("small monorepo must not be forced to Semble: %+v", facts.Candidates)
	}
	for _, c := range facts.Candidates {
		if c.Reason == "" || c.Observed == "" || c.Capabilities == nil {
			t.Fatalf("candidate %s incomplete: %+v", c.CatalogID, c)
		}
	}
	lower := strings.ToLower(stableBytes(t, root))
	for _, forbidden := range []string{"install everything", "must install", "primary", "budget", "verdict"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("recommend JSON contains model-owned token %q", forbidden)
		}
	}
}

// TestBoilerplateOracle (§44): a starter-template fixture exposes
func TestBoilerplateOracle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "extension-points.md", "# Extension points\n\nCustomize via app/config.\n")
	writeFile(t, root, "PROJECT.template.md", "# Template\n\nVariants: blue, green.\n")
	writeFile(t, root, "variants/blue/main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, root, "generated/models.g.dart", "// GENERATED CODE — do not edit.\nclass User {}\n")
	writeFile(t, root, "lib/main.dart", "import 'generated/models.g.dart';\n\nvoid main() {}\n")
	writeFile(t, root, "pubspec.yaml", "name: boilerplate\n")
	writeFile(t, root, ".gitattributes", "*.g.dart linguist-generated=true\n")
	facts, err := Recommend(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range facts.Candidates {
		for _, forbidden := range []string{"scaffold", "skill", "language", "generator", "template"} {
			if strings.Contains(strings.ToLower(c.Candidate+c.Capability+c.CatalogID), forbidden) {
				t.Fatalf("boilerplate oracle must not propose generic %q candidate: %+v", forbidden, c)
			}
		}
		if c.Capabilities == nil || c.Reason == "" {
			t.Fatalf("candidate %s incomplete: %+v", c.CatalogID, c)
		}
	}
	stableBytes(t, root)
	inspect, err := inventory.Inspect(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := inspect.Files.ByExtension["dart"]; got < 2 {
		t.Fatalf("generated + editable dart sources must both be inventoried: %+v", inspect.Files)
	}
	if got := inspect.Files.ByExtension["md"]; got < 2 {
		t.Fatalf("extension-point and template markers must be inventoried: %+v", inspect.Files)
	}
	pubInferred := false
	for _, m := range inspect.PackageManagers {
		pubInferred = pubInferred || m.ID == "pub" && m.Confidence == "inferred"
	}
	if !pubInferred {
		t.Fatalf("boilerplate pubspec must infer the pub manager: %+v", inspect.PackageManagers)
	}
	data, _ := json.Marshal(inspect)
	for _, forbidden := range []string{"primary", "preferred", "migration", "selected"} {
		if strings.Contains(strings.ToLower(string(data)), forbidden) {
			t.Fatalf("boilerplate facts contain decision token %q", forbidden)
		}
	}
}
