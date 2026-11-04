package sys

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// Parameters derives the globals of ParametersProc from the parameters [INIT-05].
//
// The hard-coded constants [PAR-03] are in params.Constants. The output variables [PAR-08]
// (PopulationSizeDay365List, Survived?) are never read and were removed in step F3.
// Display-only values and the check loop [PAR-09] are not translated.
//
// NetLogo: ParametersProc.
type Parameters struct {
	derived globals.Derived
	weather globals.Weather
}

func (s *Parameters) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	iface := ecs.GetResource[params.Interface](w)
	dir := ecs.GetResource[params.WorkingDirectory](w)

	// [PAR-01]
	queens, species, err := QueensInitialProc(file.BeeSpeciesInitialQueensListAsString)
	if err != nil {
		panic(err)
	}
	// [PAR-06]
	cso, err := AllCSOlistREP(filepath.Join(dir.Path, file.HabitatsFile))
	if err != nil {
		panic(err)
	}

	// [FPER] the weather, read once (NetLogo: Foraging_periodREP every morning)
	weather, err := weatherFromParameter(file.Weather, dir.Path)
	if err != nil {
		panic(err)
	}
	s.weather = weather
	ecs.AddResource(w, &s.weather)

	// [PAR-04]
	mortality, err := mortalityForager(file.ForagingMortalityModel)
	if err != nil {
		panic(err)
	}

	s.derived = globals.Derived{
		BeeSpeciesInitialQueensList: queens,
		SpeciesList:                 species,
		MortalityForager_per_s:      mortality,
		FoodsourcesFile:             strings.ReplaceAll(iface.MyMap, ".png", "") + "_Foodsources.txt", // [PAR-05]
		CSS_OptionsList:             cso,
		CropRotationList:            util.ListFromStringREP(file.CropRotationListAsString), // [PAR-07]
	}
	ecs.AddResource(w, &s.derived)
}

func (s *Parameters) Update(w *ecs.World) {}

func (s *Parameters) Finalize(w *ecs.World) {}

// mortalityForager returns the foraging mortality per second [PAR-04].
// Any other model name is an error. NetLogo gives no message and keeps the 0 of clear-all, so foragers and
// queens searching a nest never die (finding 5, see trans/netlogo-findings.md#finding-5). Fixed in step F5.
func mortalityForager(model string) (float64, error) {
	switch model {
	case "high":
		return 1.0e-05, nil // BEEHAVE value, Visscher & Dukas 1997
	case "intermediate":
		return 2.14e-06, nil // Schmid-Hempel & Heeb 1991
	case "low":
		return 2.75e-07, nil // Stelzer et al. 2010
	}
	return 0, fmt.Errorf("unknown ForagingMortalityModel %q (expected \"high\", \"intermediate\" or \"low\")", model)
}

// QueensInitialProc parses BeeSpeciesInitialQueensListAsString, e.g. "B_hortorum 100 B_terrestris 500" [QINI].
//
// Items alternate between species name and number of queens. A trailing name without number is dropped.
func QueensInitialProc(s string) ([]globals.SpeciesQueens, []string, error) {
	queens := []globals.SpeciesQueens{}
	// [QINI-01]
	items := util.ListFromStringREP(s)
	// [QINI-02]
	for i := 0; i+1 < len(items); i += 2 {
		n, err := strconv.ParseFloat(items[i+1], 64) // NetLogo: read-from-string
		if err != nil {
			return nil, nil, fmt.Errorf("BeeSpeciesInitialQueensListAsString: %q is not a number", items[i+1])
		}
		queens = append(queens, globals.SpeciesQueens{Species: items[i], Queens: n})
	}
	// [QINI-03]
	species := make([]string, len(queens))
	for i, q := range queens {
		species[i] = q.Species
	}
	return queens, species, nil
}

// AllCSOlistREP returns the stewardship option names from the HabitatsFile [CSO].
//
// These are the habitatType values of all rows whose habitatColourID contains "FlowerSpeciesList_", in file order.
// NetLogo shows a message and continues with an empty list if the file is missing. Here it is an error.
func AllCSOlistREP(path string) ([]string, error) {
	// [CSO-01]
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("HabitatsFile: %w", err)
	}
	defer f.Close()
	result, err := parseCSOList(f)
	if err != nil {
		return nil, fmt.Errorf("HabitatsFile %s: %w", path, err)
	}
	return result, nil
}

// parseCSOList reads the stewardship option names from the content of a HabitatsFile [CSO].
func parseCSOList(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	header := rows[0]
	header[0] = strings.TrimPrefix(header[0], "\ufeff")
	colID := slices.Index(header, "habitatColourID")
	colType := slices.Index(header, "habitatType")
	if colID < 0 || colType < 0 {
		return nil, fmt.Errorf("columns habitatColourID and habitatType required")
	}

	// [CSO-02] numeric IDs never contain the text, so no special case is needed for them.
	result := []string{}
	for _, row := range rows[1:] {
		if strings.Contains(row[colID], "FlowerSpeciesList_") {
			result = append(result, row[colType])
		}
	}
	return result, nil
}
