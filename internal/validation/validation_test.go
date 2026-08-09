package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, data string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// manifestJSON builds an ownership manifest over the given paths.
func manifestJSON(paths ...string) string {
	entries := make([]string, len(paths))
	for i, p := range paths {
		entries[i] = fmt.Sprintf(`{"path":%q}`, p)
	}
	return fmt.Sprintf(`{"schema":"agent-ready.manifest/v1","assets":[%s]}`, strings.Join(entries, ",")) + "\n"
}

func install(t *testing.T, root, name string, refs ...string) {
	t.Helper()
	var b strings.Builder
	fmt.Fprintf(&b, "---\nname: %s\ndescription: \"Trigger: fixture\"\n---\n", name)
	for _, ref := range refs {
		b.WriteString("Load references/" + ref + " when needed.\n")
		write(t, root, ".agent-ready/skills/"+name+"/references/"+ref, ref+"\n")
	}
	write(t, root, ".agent-ready/skills/"+name+"/SKILL.md", b.String())
}

func mustContain(t *testing.T, got []string, want string) {
	t.Helper()
	if !strings.Contains(strings.Join(got, "\n"), want) {
		t.Fatalf("want %q in %v", want, got)
	}
}

func TestValidateFactsPass(t *testing.T) {
	root := t.TempDir()
	install(t, root, "good-skill", "guide.md")
	write(t, root, ".agent-ready/manifest.json", manifestJSON(
		".agent-ready/skills/good-skill/SKILL.md",
		".agent-ready/skills/good-skill/references/guide.md",
	))
	facts, err := Validate(root)
	if err != nil || facts.Verdict != "pass" || facts.SchemaVersion != SchemaVersion || facts.Target != root {
		t.Fatalf("pass facts mismatch: %+v, %v", facts, err)
	}
	if c := facts.Checks[0]; !c.NameValid || !c.PatternOK || !c.DirMatch || !c.DescriptionOK || len(c.Errors) != 0 {
		t.Fatalf("check mismatch: %+v", c)
	}
}

func TestValidateFrontmatterFailures(t *testing.T) {
	long := strings.Repeat("x", 1025)
	for _, tt := range []struct{ dir, body, wantErr string }{
		{"bad_name", "---\nname: Bad_Name\ndescription: \"Trigger: x\"\n---\n", `name "Bad_Name" violates ^[a-z0-9]+(-[a-z0-9]+)*$`},
		{"other-skill", "---\nname: good-skill\ndescription: \"Trigger: x\"\n---\n", `name "good-skill" != directory "other-skill"`},
		{"longdesc", "---\nname: longdesc\ndescription: \"" + long + "\"\n---\n", "description must be 1-1024 chars (got 1025)"},
	} {
		root := t.TempDir()
		write(t, root, ".agent-ready/skills/"+tt.dir+"/SKILL.md", tt.body)
		facts, err := Validate(root)
		if err != nil || facts.Verdict != "fail" {
			t.Fatalf("%s: want fail verdict, got %+v, %v", tt.dir, facts.Checks[0], err)
		}
		mustContain(t, facts.Checks[0].Errors, tt.wantErr)
	}
}

func TestValidateProgressiveDisclosure(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agent-ready/skills/a-skill/SKILL.md", "---\nname: a-skill\ndescription: \"Trigger: disclosure fixture\"\n---\nRead references/ghost.md; scoring uses ../../references/skill-system/skill-quality-rubric.md.\n")
	write(t, root, ".agent-ready/skills/a-skill/references/extra.md", "extra\n")
	facts, err := Validate(root)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, facts.Checks[0].Errors, "progressive disclosure: body must name references/extra.md")
	mustContain(t, facts.Checks[0].Errors, "progressive disclosure: references/ghost.md named in body but not found in the suite")
	if strings.Contains(strings.Join(facts.Checks[0].Errors, "\n"), "skill-quality-rubric.md") {
		t.Fatalf("cross-tree mention must be ignored: %v", facts.Checks[0].Errors)
	}
}

func TestValidateOwnership(t *testing.T) {
	{
		root := t.TempDir()
		install(t, root, "orphan")
		install(t, root, "owned-skill")
		write(t, root, ".agent-ready/manifest.json", manifestJSON(
			".agent-ready/skills/owned-skill/SKILL.md",
			".agent-ready/skills/owned-skill/references/ghost.md",
		))
		facts, err := Validate(root)
		if err != nil || facts.Verdict != "fail" || len(facts.Checks) != 2 {
			t.Fatalf("ownership failures: %+v, %v", facts.Checks, err)
		}
		mustContain(t, facts.Checks[0].Errors, "ownership: not listed")
		mustContain(t, facts.Checks[1].Errors, "ownership: owned target missing: .agent-ready/skills/owned-skill/references/ghost.md")
	}
	{
		root := t.TempDir()
		write(t, root, ".agent-ready/manifest.json", "{not json")
		if _, err := Validate(root); err == nil {
			t.Fatal("expected error")
		}
	}
}
