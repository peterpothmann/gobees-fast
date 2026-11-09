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

// Variation is one key of a run file and the values it takes in an experiment (section variations), or a list of
// sites. The run file with its variations is an experiment: every combination of the values (full factorial) runs
// run.repetitions times, with the seeds of the section run (each combination the same seeds).
type Variation struct {
	// Key is the path of the key in the run file, its parts separated by dots, e.g. parameters.badgers or
	// constants.foraging.maxForagingRange. A number selects an item of a list, e.g. initialQueens.0.queens.
	Key string `yaml:"key,omitempty"`
	// Values are the values of the key, each written as in the run file (a number, a text, a list, ...).
	Values []yaml.Node `yaml:"values,omitempty,flow"`
	// Sites are the site files, instead of Key and Values: each site file sets the keys landscape and cropRotation
	// (Site), so that a landscape and its crop rotation vary together. Files.Sites holds the files they match.
	Sites Patterns `yaml:"sites,omitempty"`
}

// Column returns the column of the variation in the daily output: its key, or site for the sites.
func (v Variation) Column() string {
	if len(v.Sites) > 0 {
		return siteColumn
	}
	return v.Key
}

// Combination is one combination of the values of the variations of a run file: the run file with these values and
// its landscape files, without the section variations.
type Combination struct {
	// Values are the values of the variations in this combination, as text, in the order of the variations.
	Values []string
	Files  Files
}

// checkVariations checks the section variations: each key once, with at least one value, not name (the folder of the
// one output of all combinations), not in the sections run and variations, not output.columns and output.visits (all
// combinations of an experiment write the same columns into the same files) and not landscape or cropRotation (a
// landscape and its crop rotation vary together, as sites). The sites at most once, without key and values.
func checkVariations(vs []Variation) error {
	seen := map[string]bool{}
	sites := false
	for i, v := range vs {
		if len(v.Sites) > 0 {
			if v.Key != "" || len(v.Values) > 0 {
				return fmt.Errorf("%s[%d]: sites has no key and no values: each site file sets landscape and "+
					"cropRotation", variationsKey, i)
			}
			if sites {
				return fmt.Errorf("%s[%d]: sites is listed twice: list all site files in one sites", variationsKey, i)
			}
			sites = true
			continue
		}
		if v.Key == "" {
			return fmt.Errorf("%s[%d]: key is missing", variationsKey, i)
		}
		first, _, _ := strings.Cut(v.Key, ".")
		switch {
		case first == nameKey:
			return fmt.Errorf("%s: %s cannot vary: all combinations write into the one folder of the name",
				variationsKey, v.Key)
		case first == runKey || first == variationsKey || v.Key == "output.columns" || v.Key == "output.visits":
			return fmt.Errorf("%s: %s cannot vary: seed, days, repetitions and threads are the same for all "+
				"combinations, and so are the output files and their columns", variationsKey, v.Key)
		case first == "landscape" || first == "cropRotation":
			return fmt.Errorf("%s: %s cannot vary as a key: a landscape and its crop rotation vary together, list "+
				"them as sites (sites: sites/*/site.yaml)", variationsKey, v.Key)
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

// CombinationCount returns the number of combinations of the variations of the run file f: the product of the numbers
// of their values (of the sites: of the site files), 1 without variations.
func (f Files) CombinationCount() int {
	n := 1
	for _, v := range f.Config.Variations {
		n *= f.valueCount(v)
	}
	return n
}

// valueCount returns the number of values of the variation v of f: its values, or the site files of the sites.
func (f Files) valueCount(v Variation) int {
	if len(v.Sites) > 0 {
		return len(f.Sites)
	}
	return len(v.Values)
}

// Combination returns the combination number k (from 0) of the values of the variations of the run file f, in the
// order of a nested loop: the first variation changes slowest. It is the run file with these values, read and checked
// as a run file (unknown keys, whole numbers, constants), and its landscape files. A run file without variations has
// one combination, the run file itself.
//
// Each call builds the combination anew, so that an experiment holds only the combinations of the runs at the same
// time, not all of them. Combination only reads f, so several runs may call it at the same time.
func (f Files) Combination(k int) (Combination, error) {
	vs := f.Config.Variations
	if n := f.CombinationCount(); k < 0 || k >= n {
		return Combination{}, fmt.Errorf("combination %d, want 0 to %d", k, n-1)
	}
	if len(vs) == 0 {
		return Combination{Files: f}, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(f.source, &root); err != nil {
		return Combination{}, err
	}
	deleteKey(root.Content[0], variationsKey)

	// the index of the value of each variation: the digits of k, the last variation is the lowest digit
	index := make([]int, len(vs))
	for i := len(vs) - 1; i >= 0; i-- {
		index[i] = k % f.valueCount(vs[i])
		k /= f.valueCount(vs[i])
	}
	values := make([]string, len(vs))
	for i, v := range vs {
		if len(v.Sites) > 0 {
			s := f.Sites[index[i]]
			if err := setSite(root.Content[0], s); err != nil {
				return Combination{}, fmt.Errorf("%s: sites: %s: %w", variationsKey, s.File, err)
			}
			values[i] = s.Name
			continue
		}
		value := &v.Values[index[i]]
		if err := setPath(root.Content[0], v.Key, value); err != nil {
			return Combination{}, fmt.Errorf("%s: %w", variationsKey, err)
		}
		text, err := valueText(value)
		if err != nil {
			return Combination{}, fmt.Errorf("%s: %s: %w", variationsKey, v.Key, err)
		}
		values[i] = text
	}
	data, err := yaml.Marshal(&root)
	if err != nil {
		return Combination{}, err
	}
	files, err := readFrom(data, f.open, f)
	if err != nil {
		return Combination{}, fmt.Errorf("%s (%s): %w", variationsKey, describe(vs, values), err)
	}
	return Combination{Values: values, Files: files}, nil
}

// Combinations returns all combinations of the variations of the run file f at once, in the order of Combination.
// The model builds them one by one with Combination.
func (f Files) Combinations() ([]Combination, error) {
	n := f.CombinationCount()
	combinations := make([]Combination, 0, n)
	for k := range n {
		c, err := f.Combination(k)
		if err != nil {
			return nil, err
		}
		combinations = append(combinations, c)
	}
	return combinations, nil
}

// describe returns the keys and values of a combination, e.g. "parameters.badgers=2, parameters.sexLocus=true".
func describe(vs []Variation, values []string) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = v.Column() + "=" + values[i]
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
// landscape files of known, read already, are taken from it. A run file with sites has no landscape files of its
// own: each combination reads those of its site.
func readFrom(data []byte, open func(string) (io.ReadCloser, error), known Files) (Files, error) {
	f := Files{Landscapes: map[string]Landscape{}, Tables: map[string][]LandscapeTable{}, Raw: map[string][]byte{},
		source: data, open: open}
	c, err := ReadConfig(bytes.NewReader(data))
	if err != nil {
		return f, err
	}
	f.Config = c
	if hasSites(c.Variations) {
		return f, nil
	}
	for _, n := range landscapeNames(c) {
		if l, ok := known.Landscapes[n]; ok {
			f.Landscapes[n], f.Raw[n], f.Tables[n] = l, known.Raw[n], known.Tables[n]
			for _, t := range known.Tables[n] {
				f.Raw[t.Name] = known.Raw[t.Name]
			}
			continue
		}
		if err := f.readLandscape(n); err != nil {
			return f, err
		}
	}
	return f, nil
}
