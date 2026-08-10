package bootstrap

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
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

// skillDirs returns the sorted names of the directories directly under root
// that contain a SKILL.md, walked over the embedded assets FS.
func skillDirs(t *testing.T, root string) []string {
	t.Helper()
	var dirs []string
	err := fs.WalkDir(assetsFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root || !d.IsDir() {
			return nil
		}
		if _, statErr := fs.Stat(assetsFS, path+"/SKILL.md"); statErr == nil {
			dirs = append(dirs, strings.TrimPrefix(path, root+"/"))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(dirs)
	return dirs
}

// referenceFiles returns the sorted relative names of the files under
// root/references/ (empty when the skill ships no references/ subtree).
func referenceFiles(t *testing.T, root string) []string {
	t.Helper()
	if _, err := fs.Stat(assetsFS, root+"/references"); err != nil {
		return nil
	}
	var refs []string
	err := fs.WalkDir(assetsFS, root+"/references", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		refs = append(refs, strings.TrimPrefix(path, root+"/references/"))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root+"/references", err)
	}
	sort.Strings(refs)
	return refs
}

// checkFrontmatter validates one SKILL.md against the pinned OpenCode 1.18.15
// rules (R2): name matching ^[a-z0-9]+(-[a-z0-9]+)*$, optional match with the
// containing directory name, and description 1-1024 chars.
func checkFrontmatter(t *testing.T, dir, doc string, requireDirMatch bool) {
	t.Helper()
	m := regexp.MustCompile(`(?m)^name: ([a-z0-9-]+)$`).FindStringSubmatch(doc)
	if m == nil {
		t.Fatalf("%s: frontmatter name missing", dir)
	}
	name := m[1]
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(name) {
		t.Fatalf("%s: name %q violates ^[a-z0-9]+(-[a-z0-9]+)*$", dir, name)
	}
	if requireDirMatch && name != dir {
		t.Fatalf("%s: frontmatter name %q != containing directory name", dir, name)
	}
	if desc := frontmatterDescription(t, doc); len(desc) < 1 || len(desc) > 1024 {
		t.Fatalf("%s: description must be 1-1024 chars", dir)
	}
}

// checkDisclosure asserts progressive disclosure (R3) for one skill: every
// file under its references/ subtree must be named in the SKILL.md body,
// non-empty, and routed to .agent-ready/ per design D2.
func checkDisclosure(t *testing.T, dir, doc, root string) {
	t.Helper()
	for _, ref := range referenceFiles(t, root) {
		if !strings.Contains(doc, "references/"+ref) {
			t.Fatalf("%s: body must name references/%s (progressive disclosure R3)", dir, ref)
		}
		rel := strings.TrimPrefix(root, "assets/") + "/references/" + ref
		if readAsset(t, root+"/references/"+ref) == "" {
			t.Fatalf("%s references/%s is empty", dir, ref)
		}
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
	}
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

// TestOrchestratorAnalysisContent locks the PR4 skill pair (R3, R11-R15): the
// orchestrator is the real adaptive-loop skill (mode dispatch, no rigid
// sequence, no override keys) with its four references (audit-flow,
// resume-rules, stop-conditions, token-discipline); repository-analysis ships
// the FACT/INFERENCE/UNKNOWN discipline with its two references
// (evidence-labels, inventory-facts). Frontmatter is valid for the pinned
// runtime and every reference resolves and routes per design D2.
func TestOrchestratorAnalysisContent(t *testing.T) {
	namePattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, name := range []string{"agent-ready-orchestrator", "repository-analysis"} {
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

	orch := readAsset(t, "assets/skills/agent-ready-orchestrator/SKILL.md")
	if !strings.Contains(orch, "## Mode Dispatch") {
		t.Fatal("orchestrator must dispatch modes")
	}
	for _, mode := range []string{"audit", "sync", "review", "status"} {
		if !regexp.MustCompile(`(?m)^- ` + mode + `\b`).MatchString(orch) {
			t.Fatalf("orchestrator mode dispatch missing %q", mode)
		}
	}
	if !strings.Contains(orch, "## Adaptive Loop") {
		t.Fatal("orchestrator must run an adaptive loop")
	}
	for _, marker := range []string{"evidence", "confidence", "NO_ACTION"} {
		if !strings.Contains(orch, marker) {
			t.Fatalf("orchestrator adaptive loop missing %q", marker)
		}
	}
	if !strings.Contains(orch, "rigid sequence") {
		t.Fatal("orchestrator must state the loop is not a rigid sequence")
	}
	if regexp.MustCompile(`(?m)^(model|agent|subtask):`).MatchString(orch) {
		t.Fatal("orchestrator must not specify model/agent/subtask overrides (R7)")
	}
	for _, name := range []string{"audit-flow", "resume-rules", "stop-conditions", "token-discipline"} {
		rel := "skills/agent-ready-orchestrator/references/" + name + ".md"
		if readAsset(t, "assets/"+rel) == "" {
			t.Fatalf("%s is empty", rel)
		}
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
		if !strings.Contains(orch, "references/"+name+".md") {
			t.Fatalf("orchestrator body must name references/%s.md", name)
		}
	}
	auditFlow := readAsset(t, "assets/skills/agent-ready-orchestrator/references/audit-flow.md")
	for _, marker := range []string{"exploration_plan", "NO_ACTION", "checkpoint save", "tools status --json", "degrade gracefully"} {
		if !strings.Contains(auditFlow, marker) {
			t.Fatalf("audit-flow missing %q", marker)
		}
	}
	resume := readAsset(t, "assets/skills/agent-ready-orchestrator/references/resume-rules.md")
	for _, marker := range []string{"checkpoint", "hash changed", "NO_ACTION"} {
		if !strings.Contains(resume, marker) {
			t.Fatalf("resume-rules missing %q", marker)
		}
	}
	stop := readAsset(t, "assets/skills/agent-ready-orchestrator/references/stop-conditions.md")
	for _, marker := range []string{"ASK_USER", "STOP_WITH_CONCERNS", "no-new-evidence"} {
		if !strings.Contains(stop, marker) {
			t.Fatalf("stop-conditions missing %q", marker)
		}
	}
	tokens := readAsset(t, "assets/skills/agent-ready-orchestrator/references/token-discipline.md")
	for _, marker := range []string{"smallest useful context", "Reuse checkpointed evidence", "on demand"} {
		if !strings.Contains(tokens, marker) {
			t.Fatalf("token-discipline missing %q", marker)
		}
	}

	analysis := readAsset(t, "assets/skills/repository-analysis/SKILL.md")
	for _, label := range []string{"FACT", "INFERENCE", "UNKNOWN"} {
		if !strings.Contains(analysis, label) {
			t.Fatalf("repository-analysis missing evidence label %q", label)
		}
	}
	for _, name := range []string{"evidence-labels", "inventory-facts"} {
		rel := "skills/repository-analysis/references/" + name + ".md"
		if readAsset(t, "assets/"+rel) == "" {
			t.Fatalf("%s is empty", rel)
		}
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
		if !strings.Contains(analysis, "references/"+name+".md") {
			t.Fatalf("repository-analysis body must name references/%s.md", name)
		}
	}
	labels := readAsset(t, "assets/skills/repository-analysis/references/evidence-labels.md")
	for _, label := range []string{"FACT", "INFERENCE", "UNKNOWN"} {
		if !strings.Contains(labels, label) {
			t.Fatalf("evidence-labels missing %q", label)
		}
	}
	inventory := readAsset(t, "assets/skills/repository-analysis/references/inventory-facts.md")
	for _, source := range []string{"inspect", "state", "changes", "checkpoint status"} {
		if !strings.Contains(inventory, source) {
			t.Fatalf("inventory-facts missing source %q", source)
		}
	}

	// Slice 7 audit-evidence-gates contract: every successful initial audit
	// mandates a categorized Tool / Capability Assessment (ecosystem,
	// productivity, provider) with reasons, or reasoned NO_ADDITIONAL_TOOLS;
	// the "tool management is out of scope" rule is replaced by mandatory
	// assessment/evaluation.
	for _, marker := range []string{"Tool / Capability Assessment", "ecosystem", "productivity", "provider", "reason"} {
		if !strings.Contains(orch, marker) {
			t.Fatalf("orchestrator missing audit-gate marker %q", marker)
		}
	}
	if strings.Contains(orch, "tool management is out of scope") {
		t.Fatal("orchestrator must not defer tool assessment as out of scope")
	}
	for _, marker := range []string{"ecosystem", "productivity", "provider", "NO_ADDITIONAL_TOOLS", "reason"} {
		if !strings.Contains(auditFlow, marker) {
			t.Fatalf("audit-flow missing audit-gate marker %q", marker)
		}
	}
	for _, marker := range []string{"Tool / Capability Assessment", "tools status", "reason"} {
		if !strings.Contains(analysis, marker) {
			t.Fatalf("repository-analysis missing audit-gate marker %q", marker)
		}
	}
	for _, marker := range []string{"tools status", "tools recommend", "NO_ADDITIONAL_TOOLS"} {
		if !strings.Contains(inventory, marker) {
			t.Fatalf("inventory-facts missing audit-gate marker %q", marker)
		}
	}
}

// TestResearchDesignEvolutionContent locks the PR5 skill trio (R1, R11):
// targeted-research, artifact-design, and incremental-evolution are embedded
// and routed to .agent-ready/skills/ with valid frontmatter; targeted-research
// owns the evidence-gap search ladder (repo first, exact version, provenance,
// stopping condition), artifact-design owns the six artifact decisions
// (CREATE/UPDATE/REUSE/REMOVE/NO_ACTION/ASK_USER) with no artifact spam, and
// incremental-evolution owns selective sync (ChangeSet interpretation, no full
// re-audit, not every dependency requires change). Each reference resolves and
// routes per design D2.
func TestResearchDesignEvolutionContent(t *testing.T) {
	namePattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, name := range []string{"targeted-research", "artifact-design", "incremental-evolution"} {
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

	research := readAsset(t, "assets/skills/targeted-research/SKILL.md")
	if !strings.Contains(research, "references/search-strategies.md") {
		t.Fatal("targeted-research body must name references/search-strategies.md")
	}
	for _, marker := range []string{"concrete question", "exact version", "source and version", "Stop when the question is answered", "never blocks"} {
		if !strings.Contains(research, marker) {
			t.Fatalf("targeted-research missing %q", marker)
		}
	}

	design := readAsset(t, "assets/skills/artifact-design/SKILL.md")
	if !strings.Contains(design, "references/artifact-decisions.md") {
		t.Fatal("artifact-design body must name references/artifact-decisions.md")
	}
	for _, decision := range []string{"CREATE", "UPDATE", "REUSE", "REMOVE", "NO_ACTION", "ASK_USER"} {
		if !strings.Contains(design, decision) {
			t.Fatalf("artifact-design missing decision %q", decision)
		}
	}
	for _, marker := range []string{"labeled evidence", "artifact spam", "UNKNOWN-only", "decisions.jsonl"} {
		if !strings.Contains(design, marker) {
			t.Fatalf("artifact-design missing %q", marker)
		}
	}

	evolution := readAsset(t, "assets/skills/incremental-evolution/SKILL.md")
	if !strings.Contains(evolution, "references/sync-flow.md") {
		t.Fatal("incremental-evolution body must name references/sync-flow.md")
	}
	for _, marker := range []string{"ChangeSet", "changed paths", "Never re-run a full audit", "Not every dependency", "changes` and `checkpoint status"} {
		if !strings.Contains(evolution, marker) {
			t.Fatalf("incremental-evolution missing %q", marker)
		}
	}

	for _, rel := range []string{
		"skills/targeted-research/references/search-strategies.md",
		"skills/artifact-design/references/artifact-decisions.md",
		"skills/incremental-evolution/references/sync-flow.md",
	} {
		if readAsset(t, "assets/"+rel) == "" {
			t.Fatalf("%s is empty", rel)
		}
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
	}

	strategies := readAsset(t, "assets/skills/targeted-research/references/search-strategies.md")
	for _, marker := range []string{"Repository itself", "Version metadata", "Official documentation", "specialized provider", "Broader web", "Stopping condition", "Exact version", "Provenance", "tools status --json", "candidate evidence", "never invent tool state"} {
		if !strings.Contains(strategies, marker) {
			t.Fatalf("search-strategies missing %q", marker)
		}
	}

	decisions := readAsset(t, "assets/skills/artifact-design/references/artifact-decisions.md")
	for _, decision := range []string{"CREATE", "UPDATE", "REUSE", "REMOVE", "NO_ACTION", "ASK_USER"} {
		if !strings.Contains(decisions, decision) {
			t.Fatalf("artifact-decisions missing decision %q", decision)
		}
	}
	if !strings.Contains(decisions, "artifact spam") || !strings.Contains(decisions, "N skills generated") {
		t.Fatal("artifact-decisions must forbid artifact spam and count-based evidence (R11)")
	}

	syncFlow := readAsset(t, "assets/skills/incremental-evolution/references/sync-flow.md")
	for _, entry := range []string{"`added`", "`changed`", "`removed`", "`first_run`"} {
		if !strings.Contains(syncFlow, entry) {
			t.Fatalf("sync-flow missing ChangeSet entry %q", entry)
		}
	}
	for _, marker := range []string{"Selective updates", "No full re-audit", "Not every dependency", "NO_ACTION"} {
		if !strings.Contains(syncFlow, marker) {
			t.Fatalf("sync-flow missing %q", marker)
		}
	}
}

// TestFullSuiteFrontmatterAndProgressiveDisclosure locks the full-suite
// contract (R2, R3): the harness ships exactly seven skills (R1), every
// shipped skill frontmatter validates against the pinned OpenCode 1.18.15
// rules (name pattern ^[a-z0-9]+(-[a-z0-9]+)*$, name == containing directory,
// description 1-1024 chars), and every skill progressively discloses (R3): all
// references/ files are named in the body, non-empty, and routed per design
// D2. The examples subtree is validated for name pattern and description only;
// directory-name matching is excluded there by design (examples demonstrate
// skills whose frontmatter names differ from their demo directories).
func TestFullSuiteFrontmatterAndProgressiveDisclosure(t *testing.T) {
	shipped := skillDirs(t, "assets/skills")
	if len(shipped) != 7 {
		t.Fatalf("harness must ship exactly 7 skills, found %d: %v", len(shipped), shipped)
	}
	for _, dir := range shipped {
		rel := "skills/" + dir + "/SKILL.md"
		doc := readAsset(t, "assets/"+rel)
		if want := filepath.FromSlash(".agent-ready/" + rel); mustRoute(t, rel) != want {
			t.Fatalf("route(%s) != %q", rel, want)
		}
		checkFrontmatter(t, dir, doc, true)
		checkDisclosure(t, dir, doc, "assets/skills/"+dir)
	}

	// Examples subtree: name pattern + description bounds only, no dir-match.
	for _, dir := range skillDirs(t, "assets/references/skill-system/examples") {
		doc := readAsset(t, "assets/references/skill-system/examples/"+dir+"/SKILL.md")
		checkFrontmatter(t, dir, doc, false)
		checkDisclosure(t, dir, doc, "assets/references/skill-system/examples/"+dir)
	}
}
