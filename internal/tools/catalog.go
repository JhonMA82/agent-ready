// Package tools implements the deterministic Tool Manager surface
// (spec sections 13–15, 18–20): facts about tool presence/version, doctor
// checks, §36 recommendation evidence, and consent-gated installation via
// embedded verified recipes. Go never performs semantic routing.
package tools

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Recipe is one embedded verified install recipe: executable + fixed args
// only, never shell strings, elevation modeled separately (never in args).
type Recipe struct {
	ID          string              `json:"id"`
	Executables []string            `json:"executables"`
	VersionArgs []string            `json:"version_args"`
	Install     map[string]RecipeOp `json:"install"`
}

// RecipeOp is one package-manager operation with fixed arguments.
type RecipeOp struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
}

//go:embed recipes
var recipesFS embed.FS

// shellMeta rejects unverified free-form shell content in recipe args.
var shellMeta = regexp.MustCompile(`[;&|<>$\x60\n]`)

// shellInterpreters are never valid recipe executables (spec §21).
var shellInterpreters = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "fish": true,
	"dash": true, "ksh": true, "pwsh": true, "cmd": true,
}

// Family identifies one of the three ordered catalog families; status output
// always serializes them in fixed order: ecosystem, productivity, provider.
type Family string

const (
	FamilyEcosystem    Family = "ecosystem"
	FamilyProductivity Family = "productivity"
	FamilyProvider     Family = "provider"
)

// SafetyLevel is the §20 install safety classification; additive metadata,
// never required to read presence or version.
type SafetyLevel string

const (
	SafetySafeRecipe              SafetyLevel = "SAFE_RECIPE"
	SafetyVersionSensitive        SafetyLevel = "VERSION_SENSITIVE"
	SafetyProjectWrapperPreferred SafetyLevel = "PROJECT_WRAPPER_PREFERRED"
)

// CapabilityState is one independent support truth value; unsupported and
// unknown stay distinguishable from supported (presence or recommendation
// never implies install/configure/integration support).
type CapabilityState string

const (
	Supported   CapabilityState = "supported"
	Unsupported CapabilityState = "unsupported"
	Unknown     CapabilityState = "unknown"
)

// Capabilities is the seven independent capability support states, always
// serialized in this fixed order: detect, version, recommend, install,
// configure, integration, side_effects.
type Capabilities struct {
	Detect      CapabilityState `json:"detect"`
	Version     CapabilityState `json:"version"`
	Recommend   CapabilityState `json:"recommend"`
	Install     CapabilityState `json:"install"`
	Configure   CapabilityState `json:"configure"`
	Integration CapabilityState `json:"integration"`
	SideEffects CapabilityState `json:"side_effects"`
}

// ProviderMetadata is the §48 metadata-first declaration carried by provider
// entries: install method, project-init requirement, OpenCode integration
// mode, global/local side effects, uninstall, and health check. Anything
// without a verified recipe stays declared unsupported (honesty invariant:
// install: supported ⟺ verified).
type ProviderMetadata struct {
	InstallMethod   string `json:"install_method,omitempty"`
	ProjectInit     string `json:"project_init,omitempty"`
	IntegrationMode string `json:"integration_mode,omitempty"`
	SideEffects     string `json:"side_effects,omitempty"`
	Uninstall       string `json:"uninstall,omitempty"`
	Health          string `json:"health,omitempty"`
}

// Entry is one catalog entry: stable identifier, family, detection metadata,
// the verified install recipe when one exists, the seven capability states,
// §20 additive safety metadata (level, methods, side effects, integration),
// and the §14 provider declarations (capability ids + §48 metadata).
type Entry struct {
	ID              string
	Family          Family
	Executables     []string
	VersionArgs     []string
	Install         map[string]RecipeOp
	Capabilities    Capabilities
	SafetyLevel     SafetyLevel
	Methods         []string
	SideEffects     string
	IntegrationMode string
	CapabilityIDs   []string
	Provider        *ProviderMetadata
}

// entrySpec is the authored capability-truth row; recipe-backed entries
// inherit executables/version/install from their embedded recipe JSON.
type entrySpec struct {
	id            string
	family        Family
	caps          Capabilities
	executables   []string
	versionArgs   []string
	level         SafetyLevel
	sideEffects   string
	integration   string
	capabilityIDs []string
	provider      *ProviderMetadata
}

// support is the single capability-truth table, sorted by stable identifier.
// install: supported only where a verified embedded recipe exists and its
// plan/execution/verification/consent behavior is tested.
var support = []entrySpec{
	{id: "ast-grep", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "bun", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"bun"}, versionArgs: []string{"--version"}},
	{id: "bundle", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"bundle"}, versionArgs: []string{"--version"}},
	{id: "cargo", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"cargo"}, versionArgs: []string{"--version"}},
	{id: "cmake", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"cmake"}, versionArgs: []string{"--version"}},
	{id: "codegraph", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), capabilityIDs: []string{"dependency_graph", "call_graph", "blast_radius", "architecture_query"}, provider: providerMeta("unsupported (no verified V1 recipe)", "required: project index (.codegraph/)", "none (local index only)", "executable + version + project index")},
	{id: "composer", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetyVersionSensitive},
	{id: "conan", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"conan"}, versionArgs: []string{"--version"}},
	{id: "context7", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), capabilityIDs: []string{"versioned_documentation"}, provider: providerMeta("unsupported (no verified V1 recipe)", "none", "none (client-side queries only)", "executable + version")},
	{id: "dart", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"dart"}, versionArgs: []string{"--version"}},
	{id: "deno", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"deno"}, versionArgs: []string{"--version"}},
	{id: "dotnet", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"dotnet"}, versionArgs: []string{"--version"}},
	{id: "fd", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "flutter", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"flutter"}, versionArgs: []string{"--version"}},
	{id: "gh", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "headroom", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), capabilityIDs: []string{"general_context_compression"}, provider: providerMeta("unsupported (no verified V1 recipe)", "none", "none (client-side queries only)", "executable + version")},
	{id: "go", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"go"}, versionArgs: []string{"version"}},
	{id: "gradle", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"gradle"}, versionArgs: []string{"--version"}, level: SafetyProjectWrapperPreferred},
	{id: "jq", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "maven", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"mvn"}, versionArgs: []string{"--version"}, level: SafetyProjectWrapperPreferred},
	{id: "mix", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"mix"}, versionArgs: []string{"--version"}},
	{id: "nix", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"nix"}, versionArgs: []string{"--version"}},
	{id: "node", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"node"}, versionArgs: []string{"--version"}},
	{id: "npm", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"npm"}, versionArgs: []string{"--version"}},
	{id: "pdm", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"pdm"}, versionArgs: []string{"--version"}},
	{id: "pip", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"pip"}, versionArgs: []string{"--version"}, level: "runtime-coupled"},
	{id: "pipenv", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"pipenv"}, versionArgs: []string{"--version"}},
	{id: "pnpm", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"pnpm"}, versionArgs: []string{"--version"}},
	{id: "poetry", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"poetry"}, versionArgs: []string{"--version"}},
	{id: "rg", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "rtk", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe, sideEffects: "GLOBAL_SIDE_EFFECT", integration: "opt-in"},
	{id: "rustup", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"rustup"}, versionArgs: []string{"--version"}, level: SafetyVersionSensitive},
	{id: "semble", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), capabilityIDs: []string{"semantic_retrieval"}, provider: providerMeta("unsupported (no verified V1 recipe)", "none", "none (client-side queries only)", "executable + version")},
	{id: "serena", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), capabilityIDs: []string{"symbol_intelligence", "semantic_navigation", "symbolic_editing"}, provider: providerMeta("unsupported (no verified V1 recipe)", "none", "none (client-side queries only)", "executable + version")},
	{id: "terraform", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"terraform"}, versionArgs: []string{"version"}},
	{id: "tofu", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"tofu"}, versionArgs: []string{"version"}},
	{id: "uv", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported), level: SafetySafeRecipe},
	{id: "yarn", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"yarn"}, versionArgs: []string{"--version"}},
}

// caps is the positional seven-state constructor in Capabilities field order.
func caps(detect, version, recommend, install, configure, integration, sideEffects CapabilityState) Capabilities {
	return Capabilities{Detect: detect, Version: version, Recommend: recommend, Install: install, Configure: configure, Integration: integration, SideEffects: sideEffects}
}

// providerMeta builds the §48 metadata declaration for a provider entry.
// Nothing is invented: fields without a verified implementation declare
// unsupported; uninstall stays not-applicable because providers are never
// auto-installed.
func providerMeta(install, init, effects, health string) *ProviderMetadata {
	return &ProviderMetadata{
		InstallMethod:   install,
		ProjectInit:     init,
		IntegrationMode: "none (unsupported in V1)",
		SideEffects:     effects,
		Uninstall:       "not applicable (never auto-installed)",
		Health:          health,
	}
}

// Catalog returns the single support-truth catalog: every entry with its
// family and capability states, sorted by stable identifier; recipe-backed
// entries carry their embedded verified recipe as the install contract.
func Catalog() []Entry {
	recipes, err := loadAll()
	if err != nil {
		panic("invalid embedded recipe catalog: " + err.Error())
	}
	byID := make(map[string]Recipe, len(recipes))
	for _, recipe := range recipes {
		byID[recipe.ID] = recipe
	}
	entries := make([]Entry, 0, len(support))
	for _, spec := range support {
		entry := Entry{ID: spec.id, Family: spec.family, Capabilities: spec.caps}
		if recipe, ok := byID[spec.id]; ok {
			entry.Executables = recipe.Executables
			entry.VersionArgs = recipe.VersionArgs
			entry.Install = recipe.Install
			for pm := range recipe.Install {
				entry.Methods = append(entry.Methods, pm)
			}
			sort.Strings(entry.Methods)
		} else {
			entry.Executables = spec.executables
			entry.VersionArgs = spec.versionArgs
		}
		entry.SafetyLevel = spec.level
		entry.SideEffects = spec.sideEffects
		entry.IntegrationMode = spec.integration
		entry.CapabilityIDs = spec.capabilityIDs
		entry.Provider = spec.provider
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries
}

// loadAll parses every embedded recipe file under recipes/.
func loadAll() ([]Recipe, error) {
	entries, err := fs.ReadDir(recipesFS, "recipes")
	if err != nil {
		return nil, err
	}
	var recipes []Recipe
	for _, entry := range entries {
		if entry.IsDir() || !hasJSONSuffix(entry.Name()) {
			continue
		}
		data, err := fs.ReadFile(recipesFS, "recipes/"+entry.Name())
		if err != nil {
			return nil, err
		}
		recipe, err := parseRecipe(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		recipes = append(recipes, recipe)
	}
	sort.Slice(recipes, func(i, j int) bool { return recipes[i].ID < recipes[j].ID })
	return recipes, nil
}

func hasJSONSuffix(name string) bool {
	return len(name) > 5 && name[len(name)-5:] == ".json"
}

func parseRecipe(data []byte) (Recipe, error) {
	var recipe Recipe
	if err := json.Unmarshal(data, &recipe); err != nil {
		return Recipe{}, fmt.Errorf("invalid recipe JSON: %w", err)
	}
	if err := ValidateRecipe(recipe); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

// ValidateRecipe enforces the recipe contract: id non-empty, at least one
// executable, no shell interpreters as executables, fixed args with no shell
// metacharacters or pipe patterns, no elevation in args.
func ValidateRecipe(recipe Recipe) error {
	if recipe.ID == "" {
		return fmt.Errorf("recipe id required")
	}
	if len(recipe.Executables) == 0 {
		return fmt.Errorf("recipe %s: at least one executable required", recipe.ID)
	}
	for pm, op := range recipe.Install {
		if op.Executable == "" {
			return fmt.Errorf("recipe %s: %s executable required", recipe.ID, pm)
		}
		if shellInterpreters[filepath.Base(op.Executable)] {
			return fmt.Errorf("recipe %s: %s executable %q is a shell interpreter; shell recipes are rejected", recipe.ID, pm, op.Executable)
		}
		for _, arg := range op.Args {
			if shellMeta.MatchString(arg) {
				return fmt.Errorf("recipe %s: %s arg %q contains shell metacharacters", recipe.ID, pm, arg)
			}
		}
	}
	return nil
}

// ExplainFacts is the agent-ready.explain/v1 schema: one entry's declared
// facts, rendered without executing or installing anything.
type ExplainFacts struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Kind          string       `json:"kind"`
	Capabilities  Capabilities `json:"capabilities"`
	SafetyLevel   SafetyLevel  `json:"safety_level,omitempty"`
	Methods       []string     `json:"methods"`
	SideEffects   string       `json:"side_effects,omitempty"`
	Integration   string       `json:"integration,omitempty"`
}

// ExplainSchemaVersion is the agent-ready.explain/v1 schema.
const ExplainSchemaVersion = "agent-ready.explain/v1"

// Explain returns one catalog entry's declared facts (D7) without executing
// or installing anything; an unknown id fails naming the id.
func Explain(toolID string) (ExplainFacts, error) {
	for _, entry := range Catalog() {
		if entry.ID == toolID {
			return ExplainFacts{SchemaVersion: ExplainSchemaVersion, ID: entry.ID, Kind: string(entry.Family),
				Capabilities: entry.Capabilities, SafetyLevel: entry.SafetyLevel, Methods: entry.Methods,
				SideEffects: entry.SideEffects, Integration: entry.IntegrationMode}, nil
		}
	}
	return ExplainFacts{}, fmt.Errorf("tool %q is not in the catalog", toolID)
}

// Summary renders the compact default output.
func (e ExplainFacts) Summary() string {
	level, methods := string(e.SafetyLevel), "none"
	if level == "" {
		level = "none"
	}
	if len(e.Methods) > 0 {
		methods = strings.Join(e.Methods, ", ")
	}
	return fmt.Sprintf("%s: kind=%s safety_level=%s methods=%s detect=%s version=%s install=%s integration=%s side_effects=%s",
		e.ID, e.Kind, level, methods, e.Capabilities.Detect, e.Capabilities.Version, e.Capabilities.Install,
		e.Capabilities.Integration, e.Capabilities.SideEffects)
}
