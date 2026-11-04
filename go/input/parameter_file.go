package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gobees/params"
)

// removedParameters are parameters of the NetLogo parameter files whose model branch gobees-fast does not have, with
// the only value it accepts (lower-case names). The NetLogo files keep these columns, so they are read and checked,
// but not stored.
var removedParameters = map[string]string{
	"colonies_ibm": "0", // individual-based colonies [CCOL-05], removed in step F1
}

// ignoredParameters are the columns of the NetLogo parameter files that only the NetLogo GUI reads (lower-case
// names). They are read, but neither stored nor checked. Removed from params.ParameterFile in step F6.
var ignoredParameters = map[string]bool{
	"backgroundcolour":       true,
	"backgroundimage":        true,
	"black_th":               true,
	"gridsize":               true,
	"mastersizefactor":       true,
	"maxpatchradius_m":       true,
	"satellitefile":          true,
	"setcolour":              true,
	"showcohorts?":           true,
	"showdeadcols?":          true,
	"showfoodsources?":       true,
	"showgrid?":              true,
	"showinspectedcolony?":   true,
	"showmasterpatchesonly?": true,
	"shownests?":             true,
	"showplots?":             true,
	"showqueens?":            true,
	"showsearchingqueens?":   true,
	"showweather?":           true,
	"white_th":               true,
}

// ReadParameterFile reads a NetLogo parameter file [RPAR].
func ReadParameterFile(path string) (params.ParameterFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return params.ParameterFile{}, err
	}
	defer f.Close()
	p, err := ParseParameterFile(f)
	if err != nil {
		return params.ParameterFile{}, fmt.Errorf("%s: %w", path, err)
	}
	return p, nil
}

// ParseParameterFile parses a NetLogo parameter file [RPAR].
//
// Row 1 holds the parameter names, row 2 the values, further rows are ignored [RPAR-02].
// A column missing in the file keeps its zero value (NetLogo: the value after clear-all).
func ParseParameterFile(r io.Reader) (params.ParameterFile, error) {
	p := params.ParameterFile{}

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return p, err
	}
	if len(rows) < 2 {
		return p, fmt.Errorf("expected a header row and a value row, got %d rows", len(rows))
	}
	header, values := rows[0], rows[1]
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}

	fields := fieldsByName()
	for _, name := range header {
		// [RPAR-03] position finds the FIRST column with this name.
		idx := firstIndex(header, name)
		if idx >= len(values) {
			return p, fmt.Errorf("parameter %q has no value", name)
		}
		literal := parameterLiteral(name, values[idx])

		key := strings.ToLower(strings.TrimSpace(name)) // NetLogo names are case-insensitive.
		if ignoredParameters[key] {
			continue
		}
		if allowed, removed := removedParameters[key]; removed {
			if strings.TrimSpace(literal) != allowed {
				return p, fmt.Errorf("parameter %q = %s is not supported by gobees-fast, only %s", name, literal,
					allowed)
			}
			continue
		}
		field, ok := fields[key]
		if !ok {
			// NetLogo: run "set <name> ..." fails for an unknown variable.
			return p, fmt.Errorf("unknown parameter %q", name)
		}
		if err := setField(reflect.ValueOf(&p).Elem().FieldByIndex(field.Index), literal); err != nil {
			return p, fmt.Errorf("parameter %q: %w", name, err)
		}
	}
	return p, nil
}

// parameterLiteral reproduces how ReadAllParametersProc builds the NetLogo source text of a value [RPAR-03].
// A value containing doubled quotes "" is reduced to one pair of quotes, except for MyFarmlandPatches.
func parameterLiteral(name, value string) string {
	if strings.Contains(value, `""`) && name != "MyFarmlandPatches" {
		return `"` + strings.ReplaceAll(value, `"`, "") + `"`
	}
	return value
}

// setField assigns a NetLogo literal (number, true/false, or "string") to a struct field.
func setField(v reflect.Value, literal string) error {
	isString := len(literal) >= 2 && strings.HasPrefix(literal, `"`) && strings.HasSuffix(literal, `"`)

	switch v.Kind() {
	case reflect.String:
		if !isString {
			return fmt.Errorf("expected a string in quotes, got %s", literal)
		}
		v.SetString(literal[1 : len(literal)-1])
	case reflect.Float64:
		x, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			return fmt.Errorf("expected a number, got %s", literal)
		}
		v.SetFloat(x)
	case reflect.Bool:
		switch strings.ToLower(literal) {
		case "true":
			v.SetBool(true)
		case "false":
			v.SetBool(false)
		default:
			return fmt.Errorf("expected true or false, got %s", literal)
		}
	default:
		return fmt.Errorf("unsupported field type %s", v.Kind())
	}
	return nil
}

// fieldsByName maps the lower-case NetLogo name of each field to the field.
func fieldsByName() map[string]reflect.StructField {
	t := reflect.TypeFor[params.ParameterFile]()
	fields := make(map[string]reflect.StructField, t.NumField())
	for f := range t.Fields() {
		fields[strings.ToLower(f.Tag.Get("netlogo"))] = f
	}
	return fields
}

func firstIndex(list []string, s string) int {
	for i, x := range list {
		if x == s {
			return i
		}
	}
	return -1
}
