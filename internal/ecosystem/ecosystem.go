// Package ecosystem derives deterministic repository evidence from bounded paths.
package ecosystem

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// Ecosystem identifies one evidenced ecosystem and its sorted source paths.
type Ecosystem struct {
	ID       string   `json:"id"`
	Evidence []string `json:"evidence"`
}

// Signal identifies a factual repository signal and its source path.
type Signal struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type Facts struct {
	Ecosystems       []Ecosystem        `json:"ecosystems,omitempty"`
	Manifests        []Signal           `json:"manifests,omitempty"`
	Lockfiles        []Signal           `json:"lockfiles,omitempty"`
	WorkspaceSignals []Signal           `json:"workspace_signals,omitempty"`
	ProjectWrappers  []Signal           `json:"project_wrappers,omitempty"`
	FrameworkSignals []Signal           `json:"framework_signals,omitempty"`
	FrameworkFacts   []FrameworkFact    `json:"framework_facts,omitempty"`
	BuildTools       []Signal           `json:"build_tools,omitempty"`
	TestTools        []Signal           `json:"test_tools,omitempty"`
	PackageManagers  []ManagerCandidate `json:"package_managers,omitempty"`
	ManagerConflicts []ManagerConflict  `json:"manager_conflicts,omitempty"`
}

type rule struct {
	id       string
	names    []string
	suffixes []string
}

var manifestRules = []rule{
	{"go", []string{"go.mod"}, nil},
	{"javascript", []string{"deno.json", "deno.jsonc", "package.json"}, nil},
	{"python", []string{"Pipfile", "pyproject.toml", "requirements.txt"}, nil},
	{"dotnet", nil, []string{".sln", ".slnx", ".csproj", ".fsproj"}},
	{"ruby", nil, []string{".gemspec"}},
	{"rust", []string{"rust-toolchain.toml"}, nil},
	{"swift", nil, []string{".xcodeproj", ".xcworkspace"}},
	{"terraform", nil, []string{".tf"}},
}

var lockfileRules = []rule{
	{"bun", []string{"bun.lock", "bun.lockb"}, nil},
	{"deno", []string{"deno.lock"}, nil},
	{"go", []string{"go.sum", "go.work.sum"}, nil},
	{"npm", []string{"npm-shrinkwrap.json", "package-lock.json"}, nil},
	{"pnpm", []string{"pnpm-lock.yaml"}, nil},
	{"poetry", []string{"poetry.lock"}, nil},
	{"python", []string{"Pipfile.lock"}, nil},
	{"uv", []string{"uv.lock"}, nil},
	{"yarn", []string{"yarn.lock"}, nil},
	{"bundler", []string{"Gemfile.lock"}, nil},
	{"cargo", []string{"Cargo.lock"}, nil},
	{"composer", []string{"composer.lock"}, nil},
	{"conan", []string{"conan.lock"}, nil},
	{"helm", []string{"Chart.lock"}, nil},
	{"mix", []string{"mix.lock"}, nil},
	{"nix", []string{"flake.lock"}, nil},
	{"nuget", []string{"packages.lock.json"}, nil},
	{"pdm", []string{"pdm.lock"}, nil},
	{"pub", []string{"pubspec.lock"}, nil},
	{"swift", []string{"Package.resolved"}, nil},
	{"terraform", []string{".terraform.lock.hcl"}, nil},
}

var workspaceRules = []rule{
	{"go", []string{"go.work"}, nil},
	{"lerna", []string{"lerna.json"}, nil},
	{"nx", []string{"nx.json"}, nil},
	{"pnpm", []string{"pnpm-workspace.yaml"}, nil},
	{"turbo", []string{"turbo.json"}, nil},
}

var wrapperRules = []rule{
	{"gradle", []string{"gradlew", "gradlew.bat"}, nil},
	{"maven", []string{"mvnw", "mvnw.cmd"}, nil},
}

var buildRules = []rule{
	{"cmake", []string{"CMakeLists.txt", "CMakeUserPresets.json"}, nil},
	{"make", []string{"Makefile"}, nil},
	{"task", []string{"Taskfile.yaml", "Taskfile.yml"}, nil},
	{"vite", []string{"vite.config.js", "vite.config.ts"}, nil},
	{"webpack", []string{"webpack.config.js", "webpack.config.ts"}, nil},
}

var testRules = []rule{
	{"jest", []string{"jest.config.js", "jest.config.ts"}, nil},
	{"phpunit", []string{"phpunit.xml.dist"}, nil},
	{"pytest", []string{"pytest.ini"}, nil},
	{"vitest", []string{"vitest.config.js", "vitest.config.ts"}, nil},
}

type FrameworkFact struct {
	Name              string   `json:"name"`
	Version           string   `json:"version,omitempty"`
	Evidence          []string `json:"evidence"`
	CentralitySignals []Signal `json:"centrality_signals"`
}

type frameworkMeta struct{ manifest, dep, token string }

var frameworkRules = []rule{
	{"angular", []string{"angular.json"}, nil},
	{"django", []string{"manage.py"}, nil},
	{"nextjs", []string{"next.config.js", "next.config.mjs", "next.config.ts"}, nil},
	{"ratatui", nil, nil},
}

var frameworkMetaByID = map[string]frameworkMeta{
	"angular": {"package.json", "@angular/core", "@angular/"},
	"django":  {"", "", "django"},
	"nextjs":  {"package.json", "next", "next/"},
	"ratatui": {"Cargo.toml", "ratatui", "ratatui"},
}

func Detect(root string, paths []string) Facts {
	paths = append([]string(nil), paths...)
	sort.Strings(paths)
	paths = slices.Compact(paths)
	facts := Facts{
		Manifests:        signals(paths, manifestRules),
		Lockfiles:        signals(paths, lockfileRules),
		WorkspaceSignals: signals(paths, workspaceRules),
		ProjectWrappers:  signals(paths, wrapperRules),
		FrameworkSignals: signals(paths, frameworkRules),
		BuildTools:       signals(paths, buildRules),
		TestTools:        signals(paths, testRules),
	}
	evidence := map[string][]string{}
	for _, signal := range facts.Manifests {
		evidence[signal.ID] = append(evidence[signal.ID], signal.Path)
	}
	for id, paths := range evidence {
		facts.Ecosystems = append(facts.Ecosystems, Ecosystem{id, paths})
	}
	sort.Slice(facts.Ecosystems, func(i, j int) bool { return facts.Ecosystems[i].ID < facts.Ecosystems[j].ID })
	facts.FrameworkFacts = frameworkFacts(root, paths)
	facts.PackageManagers, facts.ManagerConflicts = resolveManagers(paths)
	return facts
}

func signals(paths []string, rules []rule) []Signal {
	var out []Signal
	for _, source := range paths {
		name := path.Base(source)
		for _, rule := range rules {
			if slices.Contains(rule.names, name) {
				out = append(out, Signal{rule.id, source})
				continue
			}
			for _, suffix := range rule.suffixes {
				if strings.HasSuffix(name, suffix) {
					out = append(out, Signal{rule.id, source})
					break
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID || out[i].ID == out[j].ID && out[i].Path < out[j].Path
	})
	return out
}

func frameworkFacts(root string, paths []string) []FrameworkFact {
	facts := []FrameworkFact{}
	for _, fr := range frameworkRules {
		meta := frameworkMetaByID[fr.id]
		evidence := []string{}
		version := ""
		for _, source := range paths {
			if slices.Contains(fr.names, path.Base(source)) {
				evidence = append(evidence, source)
			} else if root != "" && meta.manifest != "" && path.Base(source) == meta.manifest {
				if v := declaredVersion(root, source, meta.dep); v != "" {
					version, evidence = v, append(evidence, source)
				}
			}
		}
		if len(evidence) == 0 {
			continue
		}
		var centrality []Signal
		if root != "" && meta.token != "" {
			for i, scanned := 0, 0; i < len(paths) && scanned < 20; i++ {
				if slices.Contains(evidence, paths[i]) {
					continue
				}
				scanned++
				if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(paths[i]))); err == nil && strings.Contains(string(data), meta.token) {
					centrality = append(centrality, Signal{ID: fr.id, Path: paths[i]})
				}
			}
		}
		facts = append(facts, FrameworkFact{fr.id, version, evidence, centrality})
	}
	return facts
}

func declaredVersion(root, manifest, dep string) string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(manifest)))
	if err != nil {
		return ""
	}
	if path.Base(manifest) == "Cargo.toml" {
		for _, m := range cargoDepRe.FindAllSubmatch(data, -1) {
			if string(m[1]) == dep {
				if len(m[2]) > 0 {
					return string(m[2])
				}
				return string(m[3])
			}
		}
		return ""
	}
	var deps map[string]map[string]string
	if err := json.Unmarshal(data, &deps); err != nil {
		return ""
	}
	for _, section := range []string{"dependencies", "devDependencies", "require", "require-dev"} {
		if v, ok := deps[section][dep]; ok {
			return v
		}
	}
	return ""
}

var cargoDepRe = regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_.-]+)\s*=\s*(?:"([^"]+)"|\{[^}]*version\s*=\s*"([^"]+)"[^}]*\})`)
