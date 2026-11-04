package input

import (
	"fmt"
	"io"

	"gobees/params"
)

// parseFlowerspeciesFile reads the FlowerspeciesFile [LAYR-02].
// The names are stored as """Name""" in the file; read-from-string turns them into plain strings.
func parseFlowerspeciesFile(r io.Reader) (params.FlowerSpeciesTable, error) {
	t, err := readCSVTable(r)
	if err != nil {
		return params.FlowerSpeciesTable{}, err
	}
	table := params.FlowerSpeciesTable{}
	for _, row := range t.rows {
		v, err := ReadFromString(row[0])
		if err != nil {
			return table, fmt.Errorf("flower species name %q: %w", row[0], err)
		}
		name, ok := v.(string)
		if !ok {
			return table, fmt.Errorf("flower species name %q: expected a string", row[0])
		}
		d := params.FlowerSpecies{Name: name}
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
		table.Names = append(table.Names, name)
		table.Rows = append(table.Rows, d)
	}
	return table, nil
}
