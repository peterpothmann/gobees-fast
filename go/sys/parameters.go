package sys

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
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
// The hard-coded constants [PAR-03] are in params.Constants. Display-only values,
// output variables [PAR-08] and the check loop [PAR-09] are not translated.
//
// NetLogo: ParametersProc.
type Parameters struct {
	derived globals.Derived
}

func (s *Parameters) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	iface := ecs.GetResource[params.Interface](w)
	dir := ecs.GetResource[params.WorkingDirectory](w)

	// [PAR-01]
	queens, species, err := queensInitial(file.BeeSpeciesInitialQueensListAsString)
	if err != nil {
		panic(err)
	}
	// [PAR-06]
	cso, err := allCSOList(filepath.Join(dir.Path, file.HabitatsFile))
	if err != nil {
		panic(err)
	}

	s.derived = globals.Derived{
		BeeSpeciesInitialQueensList: queens,
		SpeciesList:                 species,
		MortalityForager_per_s:      mortalityForager(file.ForagingMortalityModel),                    // [PAR-04]
		FoodsourcesFile:             strings.ReplaceAll(iface.MyMap, ".png", "") + "_Foodsources.txt", // [PAR-05]
		CSS_OptionsList:             cso,
		CropRotationList:            util.ListFromString(file.CropRotationListAsString), // [PAR-07]
	}
	ecs.AddResource(w, &s.derived)
}

func (s *Parameters) Update(w *ecs.World) {}

func (s *Parameters) Finalize(w *ecs.World) {}

// mortalityForager returns the foraging mortality per second [PAR-04].
// Any other model name gives 0, the value after clear-all (finding 5).
// NetLogo gives no message. The warning here does not change the dynamics.
func mortalityForager(model string) float64 {
	switch model {
	case "high":
		return 1.0e-05 // BEEHAVE value, Visscher & Dukas 1997
	case "intermediate":
		return 2.14e-06 // Schmid-Hempel & Heeb 1991
	case "low":
		return 2.75e-07 // Stelzer et al. 2010
	}
	log.Printf("warning: unknown ForagingMortalityModel %q (expected \"high\", \"intermediate\" or \"low\"), foraging mortality is 0 as in NetLogo", model)
	return 0
}

// queensInitial parses BeeSpeciesInitialQueensListAsString, e.g. "B_hortorum 100 B_terrestris 500" [QINI].
//
// Items alternate between species name and number of queens. A trailing name without number is dropped.
//
// NetLogo: QueensInitialProc.
func queensInitial(s string) ([]globals.SpeciesQueens, []string, error) {
	queens := []globals.SpeciesQueens{}
	// [QINI-01]
	items := util.ListFromString(s)
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

// allCSOList returns the stewardship option names from the HabitatsFile [CSO].
//
// These are the habitatType values of all rows whose habitatColourID contains "FlowerSpeciesList_", in file order.
// NetLogo shows a message and continues with an empty list if the file is missing. Here it is an error.
//
// NetLogo: AllCSOlistREP.
func allCSOList(path string) ([]string, error) {
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
