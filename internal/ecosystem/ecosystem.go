// Package ecosystem derives deterministic repository evidence from bounded paths.
package ecosystem

import (
	"path"
	"slices"
	"sort"
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

// Facts contains additive inspect evidence. Empty collections stay omitted so
// repositories without recognized signals preserve the legacy V1 JSON bytes.
type Facts struct {
	Ecosystems       []Ecosystem `json:"ecosystems,omitempty"`
	Manifests        []Signal    `json:"manifests,omitempty"`
	Lockfiles        []Signal    `json:"lockfiles,omitempty"`
	WorkspaceSignals []Signal    `json:"workspace_signals,omitempty"`
	ProjectWrappers  []Signal    `json:"project_wrappers,omitempty"`
	FrameworkSignals []Signal    `json:"framework_signals,omitempty"`
	BuildTools       []Signal    `json:"build_tools,omitempty"`
	TestTools        []Signal    `json:"test_tools,omitempty"`
}

type rule struct {
	id    string
	names []string
}

var manifestRules = []rule{
	{"go", []string{"go.mod"}},
	{"javascript", []string{"deno.json", "deno.jsonc", "package.json"}},
	{"python", []string{"Pipfile", "pyproject.toml", "requirements.txt"}},
}

var lockfileRules = []rule{
	{"bun", []string{"bun.lock", "bun.lockb"}},
	{"deno", []string{"deno.lock"}},
	{"go", []string{"go.sum"}},
	{"npm", []string{"npm-shrinkwrap.json", "package-lock.json"}},
	{"pnpm", []string{"pnpm-lock.yaml"}},
	{"poetry", []string{"poetry.lock"}},
	{"python", []string{"Pipfile.lock"}},
	{"uv", []string{"uv.lock"}},
	{"yarn", []string{"yarn.lock"}},
}

var workspaceRules = []rule{
	{"go", []string{"go.work"}},
	{"lerna", []string{"lerna.json"}},
	{"nx", []string{"nx.json"}},
	{"pnpm", []string{"pnpm-workspace.yaml"}},
	{"turbo", []string{"turbo.json"}},
}

var wrapperRules = []rule{
	{"gradle", []string{"gradlew", "gradlew.bat"}},
	{"maven", []string{"mvnw", "mvnw.cmd"}},
}

var frameworkRules = []rule{
	{"angular", []string{"angular.json"}},
	{"django", []string{"manage.py"}},
	{"nextjs", []string{"next.config.js", "next.config.mjs", "next.config.ts"}},
}

var buildRules = []rule{
	{"cmake", []string{"CMakeLists.txt"}},
	{"make", []string{"Makefile"}},
	{"task", []string{"Taskfile.yaml", "Taskfile.yml"}},
	{"vite", []string{"vite.config.js", "vite.config.ts"}},
	{"webpack", []string{"webpack.config.js", "webpack.config.ts"}},
}

var testRules = []rule{
	{"jest", []string{"jest.config.js", "jest.config.ts"}},
	{"pytest", []string{"pytest.ini"}},
	{"vitest", []string{"vitest.config.js", "vitest.config.ts"}},
}

// Detect derives ordered facts only from paths already admitted by inventory.
func Detect(paths []string) Facts {
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
		facts.Ecosystems = append(facts.Ecosystems, Ecosystem{ID: id, Evidence: paths})
	}
	sort.Slice(facts.Ecosystems, func(i, j int) bool { return facts.Ecosystems[i].ID < facts.Ecosystems[j].ID })
	return facts
}

func signals(paths []string, rules []rule) []Signal {
	var out []Signal
	for _, source := range paths {
		name := path.Base(source)
		for _, rule := range rules {
			if slices.Contains(rule.names, name) {
				out = append(out, Signal{ID: rule.id, Path: source})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID || out[i].ID == out[j].ID && out[i].Path < out[j].Path
	})
	return out
}
