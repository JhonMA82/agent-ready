package bootstrap

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestSkillQualityRubricContent locks the normative rubric contract (R5): the
// embedded skill-quality-rubric.md must declare all seven criteria with the
// exact weights (summing to 100), the exact verdict boundaries (>=85 PASS,
// 70-84 REVISE, <70 REJECT), and a justification-forcing scoring contract.
// It also verifies the full PR2 reference set is embedded and routed to the
// design's skill-system layout (D2).
func TestSkillQualityRubricContent(t *testing.T) {
	const rubric = "assets/references/skill-system/skill-quality-rubric.md"
	data, err := fs.ReadFile(assetsFS, rubric)
	if err != nil {
		t.Fatalf("read %s: %v", rubric, err)
	}
	doc := string(data)

	criteria := map[string]int{
		"necessity":              25,
		"repository_specificity": 20,
		"discovery_description":  15,
		"procedural_value":       15,
		"progressive_disclosure": 10,
		"evidence_grounding":     10,
		"validation":             5,
	}

	row := regexp.MustCompile(`\|\s*(\d+)\s*\|\s*([a-z_]+)\s*\|`)
	found := map[string]bool{}
	sum := 0
	for _, m := range row.FindAllStringSubmatch(doc, -1) {
		name := m[2]
		if _, ok := criteria[name]; !ok {
			continue
		}
		want, err := strconv.Atoi(m[1])
		if err != nil || want != criteria[name] {
			t.Fatalf("criterion %q weight = %q, want %d", name, m[1], criteria[name])
		}
		found[name] = true
		sum += want
	}
	for name := range criteria {
		if !found[name] {
			t.Fatalf("criterion %q missing from rubric", name)
		}
	}
	if sum != 100 {
		t.Fatalf("criterion weights sum to %d, want 100", sum)
	}

	for _, boundary := range []string{">= 85", "PASS", "70-84", "REVISE", "< 70", "REJECT"} {
		if !strings.Contains(doc, boundary) {
			t.Fatalf("threshold marker %q missing from rubric", boundary)
		}
	}
	if !strings.Contains(doc, "85 is PASS, 70 is REVISE, 69 is REJECT") {
		t.Fatal("explicit boundary line missing from rubric")
	}
	if !strings.Contains(doc, "Justification contract") || !strings.Contains(doc, "per-criterion") {
		t.Fatal("justification-forcing contract missing from rubric")
	}

	// The full PR2 reference set must be embedded and routed per design D2.
	refs := []string{
		"skill-quality-rubric.md",
		"skill-authoring-guide.md",
		"progressive-disclosure.md",
		"anti-patterns.md",
	}
	for _, name := range refs {
		rel := "references/skill-system/" + name
		if _, err := fs.ReadFile(assetsFS, "assets/"+rel); err != nil {
			t.Fatalf("embedded reference %s: %v", rel, err)
		}
		target, err := route(rel)
		if err != nil {
			t.Fatalf("route %s: %v", rel, err)
		}
		if want := filepath.FromSlash(".agent-ready/references/skill-system/" + name); target != want {
			t.Fatalf("route(%s) = %q, want %q", rel, target, want)
		}
	}
}

func readAsset(t *testing.T, rel string) string {
	t.Helper()
	data, err := fs.ReadFile(assetsFS, rel)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func mustRoute(t *testing.T, rel string) string {
	t.Helper()
	target, err := route(rel)
	if err != nil {
		t.Fatalf("route %s: %v", rel, err)
	}
	return target
}

func frontmatterDescription(t *testing.T, doc string) string {
	t.Helper()
	m := regexp.MustCompile(`(?m)^description: "([^"]+)"`).FindStringSubmatch(doc)
	if m == nil {
		t.Fatal("quoted description line missing from frontmatter")
	}
	return m[1]
}

// TestSkillCreatorReviewerContent locks the PR3 skill pair (R4/R6): both
// skills are embedded and routed to .agent-ready/skills/, frontmatter is
// valid for the pinned runtime, the creator never decides necessity, the
// reviewer is the mandatory acceptance gate, and each skill's reference
// resolves per design D2.
func TestSkillCreatorReviewerContent(t *testing.T) {
	namePattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, name := range []string{"skill-creator", "skill-reviewer"} {
		rel := "skills/" + name + "/SKILL.md"
		doc := readAsset(t, "assets/"+rel)
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
		if !strings.Contains(doc, "name: "+name) {
			t.Fatalf("%s: frontmatter name missing", name)
		}
		if !namePattern.MatchString(name) {
			t.Fatalf("%s: name pattern violation", name)
		}
		if desc := frontmatterDescription(t, doc); len(desc) < 1 || len(desc) > 250 || !strings.HasPrefix(desc, "Trigger:") {
			t.Fatalf("%s: description must be 1-250 chars and trigger-first", name)
		}
	}
	if creator := readAsset(t, "assets/skills/skill-creator/SKILL.md"); !strings.Contains(creator, "Never decide necessity") {
		t.Fatal("skill-creator must never decide necessity (R4)")
	}
	if reviewer := readAsset(t, "assets/skills/skill-reviewer/SKILL.md"); !strings.Contains(reviewer, "mandatory gate") {
		t.Fatal("skill-reviewer must be the mandatory acceptance gate (R6)")
	}
	for _, rel := range []string{
		"skills/skill-creator/references/authoring-procedure.md",
		"skills/skill-reviewer/references/review-procedure.md",
	} {
		if readAsset(t, "assets/"+rel) == "" {
			t.Fatalf("%s is empty", rel)
		}
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
	}
}

// TestCanonicalExamplesContent locks the PR3 canonical examples (R6): all
// four are embedded with SKILL.md + NOTES.md and routed under
// references/skill-system/examples/ (D2); the excellent pair demonstrates
// PASS-level calibration (trigger-first, resolving references), and the bad
// pair documents its own REJECT score sheet.
func TestCanonicalExamplesContent(t *testing.T) {
	for _, name := range []string{"excellent-simple", "excellent-complex", "bad-generic", "unnecessary-skill"} {
		for _, file := range []string{"SKILL.md", "NOTES.md"} {
			rel := "references/skill-system/examples/" + name + "/" + file
			if readAsset(t, "assets/"+rel) == "" {
				t.Fatalf("%s is empty", rel)
			}
			if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
				t.Fatalf("route(%s) != %q", rel, want)
			}
		}
	}
	simple := readAsset(t, "assets/references/skill-system/examples/excellent-simple/SKILL.md")
	if desc := frontmatterDescription(t, simple); !strings.HasPrefix(desc, "Trigger:") {
		t.Fatal("excellent-simple description must be trigger-first")
	}
	complex := readAsset(t, "assets/references/skill-system/examples/excellent-complex/SKILL.md")
	if !strings.Contains(complex, "references/checklist.md") {
		t.Fatal("excellent-complex must load its checklist on demand")
	}
	if readAsset(t, "assets/references/skill-system/examples/excellent-complex/references/checklist.md") == "" {
		t.Fatal("excellent-complex checklist is empty")
	}
	bad := readAsset(t, "assets/references/skill-system/examples/bad-generic/NOTES.md")
	for _, marker := range []string{"REJECT", "30"} {
		if !strings.Contains(bad, marker) {
			t.Fatalf("bad-generic NOTES.md missing %q", marker)
		}
	}
	unnecessary := readAsset(t, "assets/references/skill-system/examples/unnecessary-skill/NOTES.md")
	for _, marker := range []string{"duplicates", "REJECT", "NO_ACTION"} {
		if !strings.Contains(unnecessary, marker) {
			t.Fatalf("unnecessary-skill NOTES.md missing %q", marker)
		}
	}
}
