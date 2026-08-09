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
