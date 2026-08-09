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

// Catalog returns the embedded verified recipes sorted by id.
func Catalog() []Recipe {
	recipes, err := loadAll()
	if err != nil {
		panic("invalid embedded recipe catalog: " + err.Error())
	}
	return recipes
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
