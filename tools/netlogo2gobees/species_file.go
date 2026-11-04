package main

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"gobees/input"
)

// parseSpeciesFile reads all rows of the BeespeciesFile, in FILE order [SPEC-01], [SPEC-03]. The columns are found
// by the keys of input.Species, which are the NetLogo column names. gobees-fast selects the species of the initial
// queens [SPEC-02] and derives the other values [SPEC-05], [SPEC-07] when it reads the run file.
// nestHabitatsList is as in the file, not yet expanded [SPEC-04].
func parseSpeciesFile(r io.Reader) ([]input.Species, error) {
	t, err := readCSVTable(r)
	if err != nil {
		return nil, err
	}
	result := []input.Species{}
	for _, row := range t.rows {
		s := input.Species{}
		v := reflect.ValueOf(&s).Elem()
		for i, f := range reflect.VisibleFields(v.Type()) {
			column := f.Tag.Get("yaml")
			field := v.Field(i)
			switch field.Kind() {
			case reflect.String:
				cell, err := t.cell(row, column)
				if err != nil {
					return nil, err
				}
				field.SetString(cell)
			case reflect.Float64:
				x, err := t.number(row, column)
				if err != nil {
					return nil, err
				}
				field.SetFloat(x)
			case reflect.Slice: // nestHabitatsList, a list literal
				cell, err := t.cell(row, column)
				if err != nil {
					return nil, err
				}
				list, err := stringList(cell)
				if err != nil {
					return nil, fmt.Errorf("%s: %s: %w", row[0], column, err)
				}
				field.Set(reflect.ValueOf(list))
			default:
				return nil, fmt.Errorf("unsupported field type %s of %s", field.Kind(), f.Name)
			}
		}
		result = append(result, s)
	}
	return result, nil
}

// stringList reads a NetLogo list literal of strings, e.g. ["Grassland*" "Garden"]. NetLogo: read-from-string.
func stringList(s string) ([]string, error) {
	v, err := ReadFromString(s)
	if err != nil {
		return nil, err
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a list")
	}
	result := make([]string, len(items))
	for i, item := range items {
		if result[i], ok = item.(string); !ok {
			return nil, fmt.Errorf("expected strings in the list")
		}
	}
	return result, nil
}

// readSpeciesFile reads the BeespeciesFile (parseSpeciesFile) [SPEC-01].
func readSpeciesFile(path string) ([]input.Species, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("BeespeciesFile: %w", err)
	}
	defer f.Close()
	species, err := parseSpeciesFile(f)
	if err != nil {
		return nil, fmt.Errorf("BeespeciesFile %s: %w", path, err)
	}
	return species, nil
}
