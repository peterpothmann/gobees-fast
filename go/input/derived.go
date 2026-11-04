package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"gobees/globals"
)

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
	items := ListFromStringREP(s)
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
