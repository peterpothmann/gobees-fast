package params

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ParameterFile contains all parameters of the NetLogo parameter file (e.g. _SYSTEM_Parameters.csv).
//
// The tag `netlogo` holds the column name in the file. Parameters that are only used by the GUI
// are read (an unknown column is an error, as in NetLogo) but not used by gobees.
type ParameterFile struct {
	AbundanceBoost                      float64 `netlogo:"AbundanceBoost"`                      // Factor on the flower abundance of all foodsources.
	BeeSpeciesInitialQueensListAsString string  `netlogo:"BeeSpeciesInitialQueensListAsString"` // Species and initial queens, e.g. "B_terrestris 500" [QINI].
	BeespeciesFile                      string  `netlogo:"BeespeciesFile"`                      // Bee species table.
	CropRotationListAsString            string  `netlogo:"CropRotationListAsString"`            // Maps of the crop rotation, separated by blanks [PAR-07].
	CumulVisitsOnlyLastYear             bool    `netlogo:"CumulVisitsOnlyLastYear?"`            // Reset cumulative visit counters on 1 January.
	FlowerspeciesFile                   string  `netlogo:"FlowerspeciesFile"`                   // Flower species table. The column name in the file has a trailing blank.
	FoodSourceLimit                     float64 `netlogo:"FoodSourceLimit"`                     // Minimum size of a foodsource in queen trips (MinSizeFoodSources?).
	ForagingMortalityFactor             float64 `netlogo:"ForagingMortalityFactor"`             // Factor on the foraging mortality.
	ForagingMortalityModel              string  `netlogo:"ForagingMortalityModel"`              // "high", "intermediate" or "low" [PAR-04].
	HabitatsFile                        string  `netlogo:"HabitatsFile"`                        // Habitat table, provides the stewardship option names [CSO].
	InitialQueensModifierBehaviorSpace  float64 `netlogo:"InitialQueensModifierBehaviorSpace"`  // Factor on the initial number of queens.
	KeepDeadColonies                    bool    `netlogo:"KeepDeadColonies?"`                   // Keep dead colonies until 1 January.
	Lambda_detectProb                   float64 `netlogo:"Lambda_detectProb"`                   // Decay of the detection probability with distance.
	MapAreaIncluded                     string  `netlogo:"MapAreaIncluded"`                     // Part of the map that is simulated, e.g. "complete".
	MaxForagingRange_m                  float64 `netlogo:"MaxForagingRange_m"`                  // NOT USED by the model, see ForagingRangeMax_m in Constants.
	MaxHibernatingQueens                float64 `netlogo:"MaxHibernatingQueens"`                // Maximum number of hibernating queens on 1 January.
	MergeHedges                         bool    `netlogo:"MergeHedges?"`                        // Merge small hedge patches.
	MinSizeFoodSources                  bool    `netlogo:"MinSizeFoodSources?"`                 // Set small nectar or pollen amounts to 0.
	MyFarmlandPatches                   string  `netlogo:"MyFarmlandPatches"`                   // IDs of farmland foodsources, e.g. " 3 7 ".
	N_Badgers                           float64 `netlogo:"N_Badgers"`                           // Number of badgers.
	RemoveEmptyFoodSources              bool    `netlogo:"RemoveEmptyFoodSources?"`             // Remove foodsources without nectar and pollen.
	SexLocus                            bool    `netlogo:"SexLocus?"`                           // Simulate the sex locus (diploid males).
	StopExtinct                         bool    `netlogo:"StopExtinct?"`                        // Stop the run when the population is extinct [GO-20].
	UnlimitedMales                      bool    `netlogo:"UnlimitedMales?"`                     // Queens always find a mate.
	Weather                             string  `netlogo:"Weather"`                             // Daily foraging period, e.g. "Constant 8 hrs" [FPER].
	WinterMortality                     bool    `netlogo:"WinterMortality?"`                    // Apply winter mortality to hibernating queens.

	// GUI only, read but not used.
	Backgroundcolour      float64 `netlogo:"Backgroundcolour"`
	BackgroundImage       string  `netlogo:"BackgroundImage"`
	Black_th              float64 `netlogo:"Black_th"`
	Gridsize              float64 `netlogo:"Gridsize"`
	MasterSizeFactor      float64 `netlogo:"MasterSizeFactor"`
	MaxPatchRadius_m      float64 `netlogo:"MaxPatchRadius_m"`
	SatelliteFile         string  `netlogo:"SatelliteFile"`
	SetColour             string  `netlogo:"SetColour"`
	ShowCohorts           bool    `netlogo:"ShowCohorts?"`
	ShowDeadCols          bool    `netlogo:"ShowDeadCols?"`
	ShowFoodsources       bool    `netlogo:"ShowFoodsources?"`
	ShowGrid              bool    `netlogo:"ShowGrid?"`
	ShowInspectedColony   bool    `netlogo:"ShowInspectedColony?"`
	ShowMasterpatchesOnly bool    `netlogo:"ShowMasterpatchesOnly?"`
	ShowNests             bool    `netlogo:"ShowNests?"`
	ShowPlots             bool    `netlogo:"ShowPlots?"`
	ShowQueens            bool    `netlogo:"ShowQueens?"`
	ShowSearchingQueens   bool    `netlogo:"ShowSearchingQueens?"`
	ShowWeather           bool    `netlogo:"ShowWeather?"`
	White_th              float64 `netlogo:"White_th"`
}

// removedParameters are parameters of the NetLogo parameter files whose model branch gobees-fast does not have, with
// the only value it accepts (lower-case names). The NetLogo files keep these columns, so they are read and checked,
// but not stored.
var removedParameters = map[string]string{
	"colonies_ibm": "0", // individual-based colonies [CCOL-05], removed in step F1
}

// ReadParameterFile reads a NetLogo parameter file [RPAR].
func ReadParameterFile(path string) (ParameterFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return ParameterFile{}, err
	}
	defer f.Close()
	p, err := ParseParameterFile(f)
	if err != nil {
		return ParameterFile{}, fmt.Errorf("%s: %w", path, err)
	}
	return p, nil
}

// ParseParameterFile parses a NetLogo parameter file [RPAR].
//
// Row 1 holds the parameter names, row 2 the values, further rows are ignored [RPAR-02].
// A column missing in the file keeps its zero value (NetLogo: the value after clear-all).
func ParseParameterFile(r io.Reader) (ParameterFile, error) {
	p := ParameterFile{}

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
	t := reflect.TypeFor[ParameterFile]()
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
