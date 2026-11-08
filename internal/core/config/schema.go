package config

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	"github.com/glassops-platform/glassops/schemas"
)

const schemaURL = "https://github.com/glassops-platform/glassops/schemas/config.json"

var (
	compileOnce sync.Once
	compiled    *jsonschema.Schema
	compileErr  error
)

func configSchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemas.Config))
		if err != nil {
			compileErr = fmt.Errorf("load embedded schema: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		if err := c.AddResource(schemaURL, doc); err != nil {
			compileErr = fmt.Errorf("add embedded schema: %w", err)
			return
		}
		compiled, compileErr = c.Compile(schemaURL)
	})
	return compiled, compileErr
}

func validateSchema(doc *yaml.Node) error {
	sch, err := configSchema()
	if err != nil {
		return err
	}
	value, err := toJSONValue(doc)
	if err != nil {
		return err
	}
	if err := sch.Validate(value); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			return &ValidationError{Problems: schemaProblems(ve)}
		}
		return err
	}
	return nil
}

// schemaProblems flattens a validation error tree into one line per failing location.
func schemaProblems(ve *jsonschema.ValidationError) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			loc := "/" + strings.Join(e.InstanceLocation, "/")
			line := fmt.Sprintf("%s: %s", loc, e.ErrorKind)
			if !seen[line] {
				seen[line] = true
				out = append(out, line)
			}
			return
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)
	sort.Strings(out)
	return out
}

// toJSONValue converts a YAML node to the plain values JSON decoding would produce. Scalars
// keep their YAML type for ints, floats, bools and nulls. Everything else, including dates,
// stays a string, so the schema sees what the user wrote.
func toJSONValue(n *yaml.Node) (any, error) {
	w := &yamlWalk{onPath: map[*yaml.Node]bool{}}
	return w.value(n, 0)
}

// Limits for converting a config node tree. A config is a few hundred nodes, so these only stop
// a hostile file: an alias that refers to itself, or aliases that expand into millions of nodes.
const (
	maxConfigDepth = 64
	maxConfigNodes = 100000
)

// yamlWalk carries the alias nodes on the current path and the count of nodes visited, so a
// cyclic or exploding config is an error instead of a crash.
type yamlWalk struct {
	onPath  map[*yaml.Node]bool
	visited int
}

func (w *yamlWalk) value(n *yaml.Node, depth int) (any, error) {
	if depth > maxConfigDepth {
		return nil, fmt.Errorf("line %d: the config nests deeper than %d levels", n.Line, maxConfigDepth)
	}
	w.visited++
	if w.visited > maxConfigNodes {
		return nil, fmt.Errorf("line %d: the config expands to more than %d values, check its anchors and aliases", n.Line, maxConfigNodes)
	}
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil, nil
		}
		return w.value(n.Content[0], depth+1)
	case yaml.MappingNode:
		m := make(map[string]any, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("line %d: mapping keys must be plain strings", k.Line)
			}
			v, err := w.value(n.Content[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			m[k.Value] = v
		}
		return m, nil
	case yaml.SequenceNode:
		s := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := w.value(c, depth+1)
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	case yaml.AliasNode:
		if n.Alias == nil {
			return nil, fmt.Errorf("line %d: an alias that names no anchor", n.Line)
		}
		if w.onPath[n.Alias] {
			return nil, fmt.Errorf("line %d: alias *%s refers to itself", n.Line, n.Value)
		}
		w.onPath[n.Alias] = true
		v, err := w.value(n.Alias, depth+1)
		delete(w.onPath, n.Alias)
		return v, err
	case yaml.ScalarNode:
		switch n.ShortTag() {
		case "!!int":
			if i, err := strconv.ParseInt(n.Value, 0, 64); err == nil {
				return i, nil
			}
			return n.Value, nil
		case "!!float":
			if f, err := strconv.ParseFloat(n.Value, 64); err == nil {
				return f, nil
			}
			return n.Value, nil
		case "!!bool":
			var b bool
			if err := n.Decode(&b); err == nil {
				return b, nil
			}
			return n.Value, nil
		case "!!null":
			return nil, nil
		default:
			return n.Value, nil
		}
	}
	return nil, fmt.Errorf("line %d: unsupported YAML node", n.Line)
}
