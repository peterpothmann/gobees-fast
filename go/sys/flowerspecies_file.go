package sys

import (
	"fmt"
	"io"
	"slices"

	"gobees/util"
)

// flowerSpeciesData is one row of the FlowerspeciesFile, i.e. the data per flower of one flower species [LAYR-02].
type flowerSpeciesData struct {
	Name                    string
	Pollen_g_per_flower     float64
	Nectar_ml_per_flower    float64
	ProteinPollenProp       float64
	Concentration_mol_per_l float64
	StartDay                float64
	StopDay                 float64
	CorollaDepth_mm         float64
	IntFlowerTime_s         float64
}

// flowerSpeciesTable holds all rows of the FlowerspeciesFile in file order.
type flowerSpeciesTable struct {
	names []string // allFlowerspeciesList
	rows  []flowerSpeciesData
}

// get returns the row of a flower species. NetLogo: item (position mySpecies allFlowerspeciesList + 1),
// i.e. the FIRST row with this name. A missing species is an error (NetLogo: message, then runtime error).
func (t *flowerSpeciesTable) get(name string) (*flowerSpeciesData, error) {
	i := slices.Index(t.names, name)
	if i < 0 {
		return nil, fmt.Errorf("flower species %q not in the FlowerspeciesFile", name)
	}
	return &t.rows[i], nil
}

// parseFlowerspeciesFile reads the FlowerspeciesFile [LAYR-02].
// The names are stored as """Name""" in the file; read-from-string turns them into plain strings.
func parseFlowerspeciesFile(r io.Reader) (flowerSpeciesTable, error) {
	t, err := readCSVTable(r)
	if err != nil {
		return flowerSpeciesTable{}, err
	}
	table := flowerSpeciesTable{}
	for _, row := range t.rows {
		v, err := util.ReadFromString(row[0])
		if err != nil {
			return table, fmt.Errorf("flower species name %q: %w", row[0], err)
		}
		name, ok := v.(string)
		if !ok {
			return table, fmt.Errorf("flower species name %q: expected a string", row[0])
		}
		d := flowerSpeciesData{Name: name}
		fields := []struct {
			column string
			value  *float64
		}{
			{"pollen_g/flower", &d.Pollen_g_per_flower},
			{"nectar_ml/flower", &d.Nectar_ml_per_flower},
			{"proteinPollenProp", &d.ProteinPollenProp},
			{"concentration_mol/l", &d.Concentration_mol_per_l},
			{"startDay", &d.StartDay},
			{"stopDay", &d.StopDay},
			{"corollaDepth_mm", &d.CorollaDepth_mm},
			{"intFlowerTime_s", &d.IntFlowerTime_s},
		}
		for _, f := range fields {
			if *f.value, err = t.number(row, f.column); err != nil {
				return table, err
			}
		}
		table.names = append(table.names, name)
		table.rows = append(table.rows, d)
	}
	return table, nil
}
