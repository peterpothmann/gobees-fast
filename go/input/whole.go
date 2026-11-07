package input

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

var unmarshalerType = reflect.TypeFor[yaml.Unmarshaler]()

// checkWhole returns an error if the YAML tree n holds a number with a fraction where the Go type t has an integer.
// yaml.v3 truncates such a number silently (12.5 becomes 12). Counts and days are whole numbers, so a
// fraction there is an error in the file. A whole number written as a float (12.0) is accepted. path names the key
// in the error message. Types that read themselves (yaml.Unmarshaler) are not checked.
func checkWhole(n *yaml.Node, t reflect.Type, path string) error {
	if n.Kind == yaml.DocumentNode {
		for _, c := range n.Content {
			if err := checkWhole(c, t, path); err != nil {
				return err
			}
		}
		return nil
	}
	if reflect.PointerTo(t).Implements(unmarshalerType) {
		return nil
	}
	switch t.Kind() {
	case reflect.Pointer:
		return checkWhole(n, t.Elem(), path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!float" {
			x, err := strconv.ParseFloat(n.Value, 64)
			if err != nil || x != math.Trunc(x) {
				return fmt.Errorf("line %d: %s: %s is not a whole number", n.Line, path, n.Value)
			}
		}
	case reflect.Slice, reflect.Array:
		if n.Kind == yaml.SequenceNode {
			for i, c := range n.Content {
				if err := checkWhole(c, t.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		if n.Kind == yaml.MappingNode {
			for i := 1; i < len(n.Content); i += 2 {
				if err := checkWhole(n.Content[i], t.Elem(), path+"."+n.Content[i-1].Value); err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		if n.Kind == yaml.MappingNode {
			fields := yamlFields(t)
			for i := 1; i < len(n.Content); i += 2 {
				key := n.Content[i-1].Value
				if ft, ok := fields[key]; ok {
					if err := checkWhole(n.Content[i], ft, strings.TrimPrefix(path+"."+key, ".")); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// yamlFields returns the types of the exported fields of struct type t by their YAML key: the name in the tag yaml,
// else the field name in lower case, as yaml.v3 does.
func yamlFields(t reflect.Type) map[string]reflect.Type {
	fields := map[string]reflect.Type{}
	for _, f := range reflect.VisibleFields(t) {
		if !f.IsExported() || f.Anonymous {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		fields[name] = f.Type
	}
	return fields
}
