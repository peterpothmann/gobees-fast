package input

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// variationsKey is the key of the section variations in a run file.
const variationsKey = "variations"

// Variation is one key of a run file and the values it takes in an experiment (section variations). The run file
// with its variations is an experiment: every combination of the values (full factorial) runs run.runs times, with the
// seeds of the section run (each combination the same seeds).
type Variation struct {
	// Key is the path of the key in the run file, its parts separated by dots, e.g. parameters.badgers or
	// constants.foraging.maxForagingRange. A number selects an item of a list, e.g. initialQueens.0.queens.
	Key string `yaml:"key"`
	// Values are the values of the key, each written as in the run file (a number, a text, a list, ...).
	Values []yaml.Node `yaml:"values,flow"`
}

// Combination is one combination of the values of the variations of a run file: the run file with these values and
// its landscape files, without the section variations.
type Combination struct {
	// Values are the values of the variations in this combination, as text, in the order of the variations.
	Values []string
	Files  Files
}

// checkVariations checks the section variations: each key once, with at least one value, not in the sections run
// and variations, not output.columns (all combinations of an experiment write the same columns into one file) and not
// landscape or cropRotation (the run info copies the landscape files of the run file only).
func checkVariations(vs []Variation) error {
	seen := map[string]bool{}
	for i, v := range vs {
		if v.Key == "" {
			return fmt.Errorf("%s[%d]: key is missing", variationsKey, i)
		}
		first, _, _ := strings.Cut(v.Key, ".")
		switch {
		case first == runKey || first == variationsKey || v.Key == "output.columns":
			return fmt.Errorf("%s: %s cannot vary: seed, days, runs and threads are the same for all combinations, "+
				"and so are the columns of the output", variationsKey, v.Key)
		case first == "landscape" || first == "cropRotation":
			return fmt.Errorf("%s: %s cannot vary: other landscapes are other run files, run them together with "+
				"-c a.yaml -c b.yaml", variationsKey, v.Key)
		}
		if seen[v.Key] {
			return fmt.Errorf("%s: %s is listed twice", variationsKey, v.Key)
		}
		seen[v.Key] = true
		if len(v.Values) == 0 {
			return fmt.Errorf("%s: %s has no values", variationsKey, v.Key)
		}
	}
	return nil
}

// Combinations returns the combinations of the values of the variations of the run file f, in the order of a nested
// loop: the first variation changes slowest. Each combination is the run file with these values, read and checked as
// a run file (unknown keys, whole numbers, constants), and its landscape files. A run file without variations gives
// one combination, the run file itself.
func (f Files) Combinations() ([]Combination, error) {
	vs := f.Config.Variations
	if len(vs) == 0 {
		return []Combination{{Files: f}}, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(f.source, &root); err != nil {
		return nil, err
	}

	n := 1
	for _, v := range vs {
		n *= len(v.Values)
	}
	combinations := make([]Combination, 0, n)
	index := make([]int, len(vs)) // the index of the value of each variation
	for range n {
		doc := copyNode(&root)
		deleteKey(doc.Content[0], variationsKey)
		values := make([]string, len(vs))
		for i, v := range vs {
			value := &v.Values[index[i]]
			if err := setPath(doc.Content[0], v.Key, value); err != nil {
				return nil, fmt.Errorf("%s: %w", variationsKey, err)
			}
			text, err := valueText(value)
			if err != nil {
				return nil, fmt.Errorf("%s: %s: %w", variationsKey, v.Key, err)
			}
			values[i] = text
		}
		data, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		files, err := readFrom(data, f.open, f)
		if err != nil {
			return nil, fmt.Errorf("%s (%s): %w", variationsKey, describe(vs, values), err)
		}
		combinations = append(combinations, Combination{Values: values, Files: files})

		// next combination: the last variation changes fastest
		for i := len(vs) - 1; i >= 0; i-- {
			index[i]++
			if index[i] < len(vs[i].Values) {
				break
			}
			index[i] = 0
		}
	}
	return combinations, nil
}

// describe returns the keys and values of a combination, e.g. "parameters.badgers=2, parameters.sexLocus=true".
func describe(vs []Variation, values []string) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = v.Key + "=" + values[i]
	}
	return strings.Join(parts, ", ")
}

// Describe returns the keys and values of the combination c of the variations of f, e.g. "parameters.badgers=2".
func (f Files) Describe(c Combination) string {
	return describe(f.Config.Variations, c.Values)
}

// valueText returns the value as one line of text: a scalar as written, a list or mapping in flow style. It is a
// column of the daily output, so a tab or a line break in it is an error.
func valueText(n *yaml.Node) (string, error) {
	text := n.Value
	if n.Kind != yaml.ScalarNode {
		c := copyNode(n)
		setFlow(c)
		b, err := yaml.Marshal(c)
		if err != nil {
			return "", err
		}
		text = strings.TrimSuffix(string(b), "\n")
	}
	if strings.ContainsAny(text, "\t\n\r") {
		return "", fmt.Errorf("value %q holds a tab or a line break", text)
	}
	return text, nil
}

// setFlow sets the flow style on n and its children.
func setFlow(n *yaml.Node) {
	n.Style |= yaml.FlowStyle
	for _, c := range n.Content {
		setFlow(c)
	}
}

// setPath sets the value at the path key (parts separated by dots) in the mapping root to a copy of value. A missing
// key of a mapping is added (the strict reading of the run file then rejects an unknown one), a number selects an
// existing item of a list.
func setPath(root *yaml.Node, key string, value *yaml.Node) error {
	parts := strings.Split(key, ".")
	n := root
	for i, part := range parts {
		last := i == len(parts)-1
		path := strings.Join(parts[:i+1], ".")
		switch n.Kind {
		case yaml.MappingNode:
			child := mappingValue(n, part)
			if child == nil {
				child = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: part}, child)
			}
			if last {
				*child = *copyNode(value)
				return nil
			}
			n = child
		case yaml.SequenceNode:
			item, err := strconv.Atoi(part)
			if err != nil || item < 0 || item >= len(n.Content) {
				return fmt.Errorf("%s: %s is a list of %d items, want an item number from 0 to %d", path,
					strings.Join(parts[:i], "."), len(n.Content), len(n.Content)-1)
			}
			if last {
				n.Content[item] = copyNode(value)
				return nil
			}
			n = n.Content[item]
		default:
			return fmt.Errorf("%s: %s is a value, it has no key %s", key, strings.Join(parts[:i], "."), part)
		}
	}
	return nil
}

// deleteKey removes key and its value from the mapping n.
func deleteKey(n *yaml.Node, key string) {
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			n.Content = append(n.Content[:i], n.Content[i+2:]...)
			return
		}
	}
}

// copyNode returns a deep copy of n.
func copyNode(n *yaml.Node) *yaml.Node {
	c := *n
	c.Content = make([]*yaml.Node, len(n.Content))
	for i, child := range n.Content {
		c.Content[i] = copyNode(child)
	}
	return &c
}

// readFrom reads the run file data and its landscape files with open, relative to the folder of the run file. The
// landscape files of known, read already, are taken from it.
func readFrom(data []byte, open func(string) (io.ReadCloser, error), known Files) (Files, error) {
	f := Files{Landscapes: map[string]Landscape{}, Raw: map[string][]byte{}, source: data, open: open}
	c, err := ReadConfig(bytes.NewReader(data))
	if err != nil {
		return f, err
	}
	f.Config = c
	for _, n := range landscapeNames(c) {
		if l, ok := known.Landscapes[n]; ok {
			f.Landscapes[n], f.Raw[n] = l, known.Raw[n]
			continue
		}
		if err := f.readLandscape(n); err != nil {
			return f, err
		}
	}
	return f, nil
}
