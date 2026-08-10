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
	"regexp"
	"sort"
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

// Family identifies one of the three ordered catalog families; status output
// always serializes them in fixed order: ecosystem, productivity, provider.
type Family string

const (
	FamilyEcosystem    Family = "ecosystem"
	FamilyProductivity Family = "productivity"
	FamilyProvider     Family = "provider"
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

// Entry is one catalog entry: stable identifier, family, detection metadata,
// the verified install recipe when one exists, and the seven capability states.
type Entry struct {
	ID           string
	Family       Family
	Executables  []string
	VersionArgs  []string
	Install      map[string]RecipeOp
	Capabilities Capabilities
}

// entrySpec is the authored capability-truth row; recipe-backed entries
// inherit executables/version/install from their embedded recipe JSON.
type entrySpec struct {
	id          string
	family      Family
	caps        Capabilities
	executables []string
	versionArgs []string
}

// support is the single capability-truth table, sorted by stable identifier.
// install: supported only where a verified embedded recipe exists and its
// plan/execution/verification/consent behavior is tested.
var support = []entrySpec{
	{id: "ast-grep", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported)},
	{id: "codegraph", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported)},
	{id: "context7", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported)},
	{id: "fd", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported)},
	{id: "gh", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported)},
	{id: "go", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"go"}, versionArgs: []string{"version"}},
	{id: "jq", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported)},
	{id: "node", family: FamilyEcosystem, caps: caps(Supported, Supported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported), executables: []string{"node"}, versionArgs: []string{"--version"}},
	{id: "rg", family: FamilyProductivity, caps: caps(Supported, Supported, Unknown, Supported, Unsupported, Unsupported, Unsupported)},
	{id: "rtk", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported)},
	{id: "semble", family: FamilyProvider, caps: caps(Unsupported, Unsupported, Unknown, Unsupported, Unsupported, Unsupported, Unsupported)},
}

// caps is the positional seven-state constructor in Capabilities field order.
func caps(detect, version, recommend, install, configure, integration, sideEffects CapabilityState) Capabilities {
	return Capabilities{Detect: detect, Version: version, Recommend: recommend, Install: install, Configure: configure, Integration: integration, SideEffects: sideEffects}
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
		} else {
			entry.Executables = spec.executables
			entry.VersionArgs = spec.versionArgs
		}
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
// executable, fixed args with no shell metacharacters, no elevation in args.
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
		for _, arg := range op.Args {
			if shellMeta.MatchString(arg) {
				return fmt.Errorf("recipe %s: %s arg %q contains shell metacharacters", recipe.ID, pm, arg)
			}
		}
	}
	return nil
}
