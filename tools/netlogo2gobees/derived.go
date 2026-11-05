package main

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

// QueensInitialProc parses BeeSpeciesInitialQueensListAsString, e.g. "B_hortorum 100 B_terrestris 500" [QINI].
//
// Items alternate between species name and number of queens. A trailing name without number is dropped.
func QueensInitialProc(s string) ([]globals.SpeciesQueens, []string, error) {
	queens := []globals.SpeciesQueens{}
	// [QINI-01]
	items := ListFromStringREP(s)
	// [QINI-02]
	for i := 0; i+1 < len(items); i += 2 {
		x, err := strconv.ParseFloat(items[i+1], 64) // NetLogo: read-from-string
		if err != nil {
			return nil, nil, fmt.Errorf("BeeSpeciesInitialQueensListAsString: %q is not a number", items[i+1])
		}
		n, ok := whole(x)
		if !ok {
			return nil, nil, fmt.Errorf("BeeSpeciesInitialQueensListAsString: %q is not a whole number", items[i+1])
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
