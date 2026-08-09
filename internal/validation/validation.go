// Package validation emits deterministic skill-conformance facts (R2/R3/R1/R8).
package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const SchemaVersion = "agent-ready.validate/v1"

const (
	verdictPass  = "pass"
	verdictFail  = "fail"
	manifestName = "manifest.json"
	skillsDir    = ".agent-ready/skills"
)

type Check struct {
	Skill         string   `json:"skill"`
	NameValid     bool     `json:"name_valid"`
	PatternOK     bool     `json:"pattern_ok"`
	DirMatch      bool     `json:"dir_match"`
	DescriptionOK bool     `json:"description_ok"`
	Errors        []string `json:"errors"`
}

type Facts struct {
	SchemaVersion string  `json:"schema_version"`
	Target        string  `json:"target"`
	Checks        []Check `json:"checks"`
	Verdict       string  `json:"verdict"`
}

var (
	namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	nameLine    = regexp.MustCompile(`(?m)^name:\s*(.*)$`)
	descLine    = regexp.MustCompile(`(?m)^description:\s*"([^"]*)"`)
	// refMention matches flat references/<name>.md mentions; cross-tree paths
	// (../.., references/skill-system/) never match.
	refMention = regexp.MustCompile(`(?m)(^|[^./\w])references/([A-Za-z0-9._-]+)\.md`)
)

type manifest struct {
	Assets []struct {
		Path string `json:"path"`
	} `json:"assets"`
}

func Validate(root string) (Facts, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Facts{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Facts{}, err
	}
	if !info.IsDir() {
		return Facts{}, fmt.Errorf("not a directory: %s", root)
	}
	owned, disk, pool, err := scan(root)
	if err != nil {
		return Facts{}, err
	}
	names := make([]string, 0, len(disk)+len(owned))
	for name := range disk {
		names = append(names, name)
	}
	for name := range owned {
		names = append(names, name)
	}
	sort.Strings(names)
	names = slices.Compact(names)
	checks := make([]Check, 0, len(names))
	verdict := verdictPass
	for _, name := range names {
		c := checkSkill(root, name, owned[name], pool)
		if len(c.Errors) > 0 {
			verdict = verdictFail
		}
		checks = append(checks, c)
	}
	return Facts{SchemaVersion: SchemaVersion, Target: root, Checks: checks, Verdict: verdict}, nil
}

func scan(root string) (owned map[string][]string, disk map[string]bool, pool map[string]bool, err error) {
	owned, disk, pool = map[string][]string{}, map[string]bool{}, map[string]bool{}
	data, err := os.ReadFile(filepath.Join(root, ".agent-ready", manifestName))
	if os.IsNotExist(err) {
		err = nil
	} else if err == nil {
		var m manifest
		if err = json.Unmarshal(data, &m); err != nil {
			return nil, nil, nil, fmt.Errorf("parse %s: %w", manifestName, err)
		}
		for _, a := range m.Assets {
			rest, ok := strings.CutPrefix(a.Path, filepath.ToSlash(skillsDir)+"/")
			if !ok {
				continue
			}
			name, _, _ := strings.Cut(rest, "/")
			owned[name] = append(owned[name], a.Path)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, skillsDir))
	if err != nil {
		return owned, disk, pool, nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		disk[e.Name()] = true
		for _, ref := range refs(filepath.Join(root, skillsDir, e.Name(), "references")) {
			pool[ref] = true
		}
	}
	return owned, disk, pool, nil
}

func checkSkill(root, name string, owned []string, pool map[string]bool) Check {
	c := Check{Skill: name, Errors: []string{}}
	skDir := filepath.Join(root, skillsDir, name)
	data, err := os.ReadFile(filepath.Join(skDir, "SKILL.md"))
	if err == nil {
		c.NameValid, c.PatternOK, c.DirMatch, c.DescriptionOK = frontmatter(name, string(data), &c.Errors)
		disclosure(skDir, string(data), pool, &c.Errors)
	} else if os.IsNotExist(err) {
		c.Errors = append(c.Errors, "SKILL.md missing")
	} else {
		c.Errors = append(c.Errors, err.Error())
	}
	if owned == nil {
		c.Errors = append(c.Errors, "ownership: not listed in ownership manifest")
	} else {
		for _, path := range owned {
			if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(path))); os.IsNotExist(err) {
				c.Errors = append(c.Errors, "ownership: owned target missing: "+path)
			}
		}
	}
	sort.Strings(c.Errors)
	return c
}

// frontmatter validates one SKILL.md (R2): pattern, dir match, description.
func frontmatter(dir, doc string, errors *[]string) (nameValid, patternOK, dirMatch, descriptionOK bool) {
	name := ""
	if m := nameLine.FindStringSubmatch(doc); m != nil {
		name = strings.TrimSpace(m[1])
	}
	if name == "" {
		*errors = append(*errors, "name missing")
		return false, false, false, false
	}
	nameValid = true
	patternOK = namePattern.MatchString(name)
	if !patternOK {
		*errors = append(*errors, fmt.Sprintf("name %q violates ^[a-z0-9]+(-[a-z0-9]+)*$", name))
	}
	dirMatch = name == dir
	if !dirMatch {
		*errors = append(*errors, fmt.Sprintf("name %q != directory %q", name, dir))
	}
	m := descLine.FindStringSubmatch(doc)
	if m == nil {
		*errors = append(*errors, "description missing or unquoted")
	} else if descriptionOK = len(m[1]) >= 1 && len(m[1]) <= 1024; !descriptionOK {
		*errors = append(*errors, fmt.Sprintf("description must be 1-1024 chars (got %d)", len(m[1])))
	}
	return nameValid, patternOK, dirMatch, descriptionOK
}

// disclosure asserts progressive disclosure (R3): refs named in body, and
// every references/<name>.md mention resolves in the suite.
func disclosure(skDir, body string, pool map[string]bool, errors *[]string) {
	for _, ref := range refs(filepath.Join(skDir, "references")) {
		if !strings.Contains(body, "references/"+ref) {
			*errors = append(*errors, "progressive disclosure: body must name references/"+ref)
		}
	}
	seen := map[string]bool{}
	for _, m := range refMention.FindAllStringSubmatch(body, -1) {
		file := m[2] + ".md"
		if seen[file] || pool[file] {
			continue
		}
		seen[file] = true
		*errors = append(*errors, "progressive disclosure: references/"+file+" named in body but not found in the suite")
	}
}

// refs returns the sorted file names under root; empty when absent.
func refs(root string) []string {
	var out []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// Summary is the compact default rendering (D5).
func (f Facts) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Target: %s\n", f.Target)
	failed := 0
	for _, c := range f.Checks {
		if len(c.Errors) > 0 {
			failed++
			fmt.Fprintf(&b, "- %s: %s\n", c.Skill, strings.Join(c.Errors, "; "))
		}
	}
	fmt.Fprintf(&b, "Skills: %d checked, %d failed\nVerdict: %s", len(f.Checks), failed, f.Verdict)
	return b.String()
}
