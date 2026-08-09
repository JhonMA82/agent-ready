package opencode

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/tailscale/hujson"
)

const skillPath = "./.agent-ready/skills"

type ConfigFile struct {
	Data []byte
	Mode fs.FileMode
}

type ConfigPlan struct {
	Path          string
	Before, After []byte
	Mode          fs.FileMode
}

type configShape struct {
	root                         *hujson.Object
	skills                       *hujson.Object
	paths                        *hujson.Array
	rootEnd, skillsEnd, pathsEnd int
	has                          bool
}

// PlanConfig selects and losslessly patches the repository OpenCode config.
// It only returns bytes; committing them belongs to the transaction layer.
func PlanConfig(jsonFile, jsoncFile *ConfigFile) (ConfigPlan, error) {
	if jsonFile == nil && jsoncFile == nil {
		return ConfigPlan{Path: "opencode.json", After: []byte("{\"skills\":{\"paths\":[\"" + skillPath + "\"]}}\n"), Mode: 0o644}, nil
	}
	type candidate struct {
		name  string
		file  *ConfigFile
		shape configShape
	}
	var candidates []candidate
	for _, c := range []candidate{{name: "opencode.json", file: jsonFile}, {name: "opencode.jsonc", file: jsoncFile}} {
		if c.file == nil {
			continue
		}
		shape, err := parseConfig(c.file.Data)
		if err != nil {
			return ConfigPlan{}, fmt.Errorf("%s: %w", c.name, err)
		}
		c.shape = shape
		candidates = append(candidates, c)
	}
	selected := 0
	if len(candidates) == 2 {
		switch {
		case candidates[0].shape.has && candidates[1].shape.has:
			return ConfigPlan{}, fmt.Errorf("ambiguous OpenCode configs: both define skills")
		case candidates[1].shape.has:
			selected = 1
		case !candidates[0].shape.has:
			selected = 1 // JSONC owns new definitions when neither file does.
		}
	}
	c := candidates[selected]
	after, err := addSkillPath(c.file.Data, c.shape)
	if err != nil {
		return ConfigPlan{}, fmt.Errorf("%s: %w", c.name, err)
	}
	return ConfigPlan{Path: c.name, Before: bytes.Clone(c.file.Data), After: after, Mode: c.file.Mode}, nil
}

func parseConfig(data []byte) (configShape, error) {
	v, err := hujson.Parse(data)
	if err != nil {
		return configShape{}, fmt.Errorf("invalid config: %w", err)
	}
	root, ok := v.Value.(*hujson.Object)
	if !ok {
		return configShape{}, fmt.Errorf("config root must be an object")
	}
	s := configShape{root: root, rootEnd: v.EndOffset - 1}
	for _, member := range root.Members {
		if literalString(member.Name) != "skills" {
			continue
		}
		if s.has {
			return configShape{}, fmt.Errorf("duplicate skills key")
		}
		s.has = true
		s.skills, ok = member.Value.Value.(*hujson.Object)
		if !ok {
			return configShape{}, fmt.Errorf("skills must be an object")
		}
		s.skillsEnd = member.Value.EndOffset - 1
	}
	if !s.has {
		return s, nil
	}
	for _, member := range s.skills.Members {
		if literalString(member.Name) != "paths" {
			continue
		}
		if s.paths != nil {
			return configShape{}, fmt.Errorf("duplicate paths key")
		}
		s.paths, ok = member.Value.Value.(*hujson.Array)
		if !ok {
			return configShape{}, fmt.Errorf("skills.paths must be a string array")
		}
		s.pathsEnd = member.Value.EndOffset - 1
		for _, element := range s.paths.Elements {
			if _, ok := element.Value.(hujson.Literal); !ok || element.Value.Kind() != '"' {
				return configShape{}, fmt.Errorf("skills.paths must be a string array")
			}
		}
	}
	return s, nil
}

func addSkillPath(data []byte, shape configShape) ([]byte, error) {
	if shape.paths != nil {
		matches := 0
		for _, element := range shape.paths.Elements {
			if normalizePath(literalString(element)) == normalizePath(skillPath) {
				matches++
			}
		}
		if matches > 1 {
			return nil, fmt.Errorf("duplicate normalized skills path")
		}
		if matches == 1 {
			return bytes.Clone(data), nil
		}
		return insertMember(data, shape.pathsEnd, len(shape.paths.Elements), len(shape.paths.Elements) > 0 && shape.paths.Elements[len(shape.paths.Elements)-1].AfterExtra != nil, `"`+skillPath+`"`), nil
	}
	if shape.skills != nil {
		return insertMember(data, shape.skillsEnd, len(shape.skills.Members), objectTrailing(shape.skills), `"paths":["`+skillPath+`"]`), nil
	}
	return insertMember(data, shape.rootEnd, len(shape.root.Members), objectTrailing(shape.root), `"skills":{"paths":["`+skillPath+`"]}`), nil
}

func insertMember(data []byte, offset, count int, trailing bool, value string) []byte {
	prefix := ""
	if count > 0 && !trailing {
		prefix = ","
	}
	suffix := ""
	if trailing {
		suffix = ","
	}
	out := make([]byte, 0, len(data)+len(prefix)+len(value)+len(suffix))
	out = append(out, data[:offset]...)
	out = append(out, prefix...)
	out = append(out, value...)
	out = append(out, suffix...)
	return append(out, data[offset:]...)
}

func objectTrailing(v *hujson.Object) bool {
	return len(v.Members) > 0 && v.Members[len(v.Members)-1].Value.AfterExtra != nil
}

func literalString(v hujson.Value) string {
	if literal, ok := v.Value.(hujson.Literal); ok {
		return literal.String()
	}
	return ""
}

func normalizePath(value string) string {
	clean := path.Clean(strings.ReplaceAll(value, "\\", "/"))
	return strings.TrimPrefix(clean, "./")
}
