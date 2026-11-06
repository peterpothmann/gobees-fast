package main

import (
	"fmt"
	"io"

	"gobees/day"
	"gobees/params"
)

// parseFlowerspeciesFile reads the FlowerspeciesFile [LAYR-02].
// The names are stored as """Name""" in the file; read-from-string turns them into plain strings.
// The column proteinPollenProp is not read: NetLogo never reads its value (removed in step F13).
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
			{"pollen_g/flower", &d.PollenPerFlower},
			{"nectar_ml/flower", &d.NectarPerFlower},
			{"concentration_mol/l", &d.Concentration},
			{"corollaDepth_mm", &d.CorollaDepth},
			{"intFlowerTime_s", &d.InterFlowerTime},
		}
		for _, f := range fields {
			if *f.value, err = t.number(row, f.column); err != nil {
				return table, err
			}
		}
		days := []struct {
			column string
			value  *day.OfYear
		}{
			{"startDay", &d.StartDay},
			{"stopDay", &d.StopDay},
		}
		for _, f := range days {
			n, err := t.wholeNumber(row, f.column)
			if err != nil {
				return table, err
			}
			*f.value = day.OfYear(n)
		}
		table.Names = append(table.Names, name)
		table.Rows = append(table.Rows, d)
	}
	return table, nil
}
