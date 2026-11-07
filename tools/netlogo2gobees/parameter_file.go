package main

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

// netLogoParameters are the parameters of a NetLogo parameter file: the model parameters of a run file and the
// columns with file names and lists, which have their own sections in a run file (convert), and the output option
// KeepDeadColonies? (section output, step F13).
type netLogoParameters struct {
	params.Parameters
	BeeSpeciesInitialQueensListAsString string `netlogo:"BeeSpeciesInitialQueensListAsString"` // Species and initial queens, e.g. "B_terrestris 500" [QINI].
	BeespeciesFile                      string `netlogo:"BeespeciesFile"`                      // Bee species table.
	CropRotationListAsString            string `netlogo:"CropRotationListAsString"`            // Maps of the crop rotation, separated by blanks [PAR-07].
	FlowerspeciesFile                   string `netlogo:"FlowerspeciesFile"`                   // Flower species table. The column name in the file has a trailing blank.
	HabitatsFile                        string `netlogo:"HabitatsFile"`                        // Habitat table, provides the stewardship option names [CSO].
	Weather                             string `netlogo:"Weather"`                             // Daily foraging period, e.g. "Constant 8 hrs" [FPER].
	KeepDeadColonies                    bool   `netlogo:"KeepDeadColonies?"`                   // Output option since step F13 (params.Output).
}

// removedParameters are parameters of the NetLogo parameter files whose model branch gobees-fast does not have, with
// the only value it accepts (lower-case names). The NetLogo files keep these columns, so they are read and checked,
// but not stored.
var removedParameters = map[string]string{
	"colonies_ibm":    "0",          // individual-based colonies [CCOL-05], removed in step F1
	"mapareaincluded": `"complete"`, // a part of the map [FSRC-08], removed in step F13 (a part is its own landscape file)
}

// noEffectParameters are parameters of the NetLogo parameter files that the model never reads (lower-case names). As
// the GUI columns, they are read, but neither stored nor checked, so that the run file holds only values that act
// (step F7).
var noEffectParameters = map[string]string{
	"cumulvisitsonlylastyear?": "resets visit counters that only the output read, removed in step F3",
	"maxforagingrange_m":       "the model uses the constant ForagingRangeMax_m = 758 m [PAR-03]",
	"myfarmlandpatches":        "sets farmland?, which is never read [FSRC-07]",
}

// ignoredParameters are the columns of the NetLogo parameter files that only the NetLogo GUI reads (lower-case
// names). They are read, but neither stored nor checked (step F6). The run files of gobees-fast do not have them.
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
func ReadParameterFile(path string) (netLogoParameters, error) {
	f, err := os.Open(path)
	if err != nil {
		return netLogoParameters{}, err
	}
	defer f.Close()
	p, err := ParseParameterFile(f)
	if err != nil {
		return netLogoParameters{}, fmt.Errorf("%s: %w", path, err)
	}
	return p, nil
}

// ParseParameterFile parses a NetLogo parameter file [RPAR].
//
// Row 1 holds the parameter names, row 2 the values, further rows are ignored [RPAR-02].
// A column missing in the file keeps its zero value (NetLogo: the value after clear-all).
func ParseParameterFile(r io.Reader) (netLogoParameters, error) {
	p := netLogoParameters{}

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
		if _, noEffect := noEffectParameters[key]; ignoredParameters[key] || noEffect {
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
	case reflect.Int: // counts, e.g. MaxHibernatingQueens
		x, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			return fmt.Errorf("expected a number, got %s", literal)
		}
		n, ok := whole(x)
		if !ok {
			return fmt.Errorf("expected a whole number, got %s", literal)
		}
		v.SetInt(int64(n))
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

// fieldsByName maps the lower-case NetLogo name of each field of netLogoParameters to the field: the tag `netlogo`
// of its own fields, the tag `yaml` of the embedded params.Parameters (the run file keeps the NetLogo names).
func fieldsByName() map[string]reflect.StructField {
	fields := map[string]reflect.StructField{}
	for _, f := range reflect.VisibleFields(reflect.TypeFor[netLogoParameters]()) {
		name := f.Tag.Get("netlogo")
		if name == "" {
			name = f.Tag.Get("yaml")
		}
		if f.Anonymous || name == "" {
			continue
		}
		fields[strings.ToLower(name)] = f
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
