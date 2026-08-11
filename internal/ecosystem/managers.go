package ecosystem

import (
	"sort"
	"strings"
)

// Confidence ranks how strongly evidence supports a candidate: project
// wrappers (tier 3) are stronger execution evidence than a global binary,
// which path evidence never asserts; lockfiles (tier 2) confirm; specific
// manifests (tier 1) infer; generic manifests never confirm alone.
const (
	ConfidenceConfirmed = "confirmed"
	ConfidenceInferred  = "inferred"
	ConfidenceAmbiguous = "ambiguous"
)

// ManagerCandidate is one package-manager candidate with evidence paths and
// the confidence that evidence supports; no candidate is ever selected.
type ManagerCandidate struct {
	ID         string   `json:"id"`
	Confidence string   `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// ManagerConflict retains concurrently evidenced managers that disagree,
// with a reason and sorted manager IDs; conflicts are facts, not choices.
type ManagerConflict struct {
	Reason   string   `json:"reason"`
	Managers []string `json:"managers"`
}

const (
	tierWrapper  = 3
	tierLockfile = 2
	tierManifest = 1
)

// managerWrapperRules are project-local executable wrappers, confirmed at
// the top tier: stronger execution evidence than a global binary.
var managerWrapperRules = []rule{
	{"gradle", []string{"gradlew", "gradlew.bat"}, nil},
	{"maven", []string{"mvnw", "mvnw.cmd"}, nil},
}

// managerLockfileRules are manager-specific lockfiles (confirmed).
// Pipfile.lock confirms pipenv even though the Slice 3 signal keeps "python".
var managerLockfileRules = []rule{
	{"bun", []string{"bun.lock", "bun.lockb"}, nil},
	{"deno", []string{"deno.lock"}, nil},
	{"go", []string{"go.sum"}, nil},
	{"npm", []string{"npm-shrinkwrap.json", "package-lock.json"}, nil},
	{"pnpm", []string{"pnpm-lock.yaml"}, nil},
	{"poetry", []string{"poetry.lock"}, nil},
	{"pipenv", []string{"Pipfile.lock"}, nil},
	{"uv", []string{"uv.lock"}, nil},
	{"yarn", []string{"yarn.lock"}, nil},
}

// managerManifestRules name a single manager without confirming it (inferred).
var managerManifestRules = []rule{
	{"deno", []string{"deno.json", "deno.jsonc"}, nil},
	{"go", []string{"go.mod"}, nil},
	{"pip", []string{"requirements.txt"}, nil},
	{"pipenv", []string{"Pipfile"}, nil},
}

// genericManifestRules are manifests shared by several managers; when no
// family candidate has better evidence, the family is emitted as ambiguous,
// so pyproject.toml alone never confirms a manager.
var genericManifestRules = []rule{
	{"javascript", []string{"package.json"}, nil},
	{"python", []string{"pyproject.toml"}, nil},
}

// ambiguousFamilies maps a generic manifest family to the candidates its
// manifest alone cannot distinguish.
var ambiguousFamilies = map[string][]string{
	"javascript": {"bun", "npm", "pnpm", "yarn"},
	"python":     {"pip", "poetry", "pdm", "uv"},
}

// conflictFamilies group managers whose concurrent confirmed evidence is a
// conflict; distinct ecosystems (go vs npm, gradle vs uv) never conflict.
var conflictFamilies = map[string][]string{
	"javascript": {"bun", "deno", "npm", "pnpm", "yarn"},
	"python":     {"pip", "pipenv", "poetry", "pdm", "uv"},
	"jvm":        {"gradle", "maven"},
	"go":         {"go"},
}

// resolveManagers derives sorted manager candidates and conflicts from paths
// only; it never selects a manager, rewrites files, or emits migration steps.
func resolveManagers(paths []string) ([]ManagerCandidate, []ManagerConflict) {
	strength := map[string]int{}
	evidence := map[string][]string{}
	add := func(rules []rule, tier int) {
		for _, s := range signals(paths, rules) {
			if strength[s.ID] < tier {
				strength[s.ID] = tier
			}
			evidence[s.ID] = append(evidence[s.ID], s.Path)
		}
	}
	add(managerWrapperRules, tierWrapper)
	add(managerLockfileRules, tierLockfile)
	add(managerManifestRules, tierManifest)

	ids := make([]string, 0, len(strength))
	for id := range strength {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var candidates []ManagerCandidate
	for _, id := range ids {
		paths := evidence[id]
		sort.Strings(paths)
		candidates = append(candidates, ManagerCandidate{ID: id, Confidence: managerConfidence(strength[id]), Evidence: paths})
	}

	// Generic manifests: a family without better-evidenced members emits
	// every plausible candidate as ambiguous, never confirming one alone.
	for _, generic := range genericManifestRules {
		var genericPaths []string
		for _, s := range signals(paths, []rule{generic}) {
			genericPaths = append(genericPaths, s.Path)
		}
		if len(genericPaths) == 0 || familyHasEvidence(ambiguousFamilies[generic.id], strength) {
			continue
		}
		sort.Strings(genericPaths)
		for _, id := range ambiguousFamilies[generic.id] {
			candidates = append(candidates, ManagerCandidate{ID: id, Confidence: ConfidenceAmbiguous, Evidence: append([]string(nil), genericPaths...)})
		}
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	return candidates, managerConflicts(strength)
}

func managerConfidence(tier int) string {
	if tier >= tierLockfile {
		return ConfidenceConfirmed
	}
	return ConfidenceInferred
}

func familyHasEvidence(ids []string, strength map[string]int) bool {
	for _, id := range ids {
		if strength[id] >= tierManifest {
			return true
		}
	}
	return false
}

func managerConflicts(strength map[string]int) []ManagerConflict {
	var conflicts []ManagerConflict
	for _, family := range conflictFamilies {
		var confirmed []string
		wrappers := true
		for _, id := range family {
			if strength[id] >= tierLockfile {
				confirmed = append(confirmed, id)
				wrappers = wrappers && strength[id] >= tierWrapper
			}
		}
		if len(confirmed) < 2 {
			continue
		}
		sort.Strings(confirmed)
		kind := "lockfiles"
		if wrappers {
			kind = "project wrappers"
		}
		conflicts = append(conflicts, ManagerConflict{
			Reason:   kind + " evidence distinct managers: " + strings.Join(confirmed, " and "),
			Managers: append([]string(nil), confirmed...),
		})
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Reason < conflicts[j].Reason })
	return conflicts
}
