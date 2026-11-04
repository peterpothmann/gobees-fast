package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// csvTable is a CSV file as read by csv:from-file: the header row and the data rows.
// Columns are found by name with item (position name header) row, i.e. the FIRST column with that name.
type csvTable struct {
	header []string
	rows   [][]string
}

func readCSVTable(r io.Reader) (csvTable, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return csvTable{}, err
	}
	if len(rows) == 0 {
		return csvTable{}, fmt.Errorf("empty file")
	}
	rows[0][0] = strings.TrimPrefix(rows[0][0], "\ufeff")
	return csvTable{header: rows[0], rows: rows[1:]}, nil
}

// cell returns the value of column name in row. NetLogo: item (position name header) row.
func (t *csvTable) cell(row []string, name string) (string, error) {
	col := slices.Index(t.header, name)
	if col < 0 {
		return "", fmt.Errorf("column %q not found", name)
	}
	if col >= len(row) {
		return "", fmt.Errorf("column %q missing in row %q", name, row[0])
	}
	return row[col], nil
}

// number returns the value of column name in row as a number, as csv:from-file converts it.
func (t *csvTable) number(row []string, name string) (float64, error) {
	s, err := t.cell(row, name)
	if err != nil {
		return 0, err
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("column %q in row %q: expected a number, got %q", name, row[0], s)
	}
	return x, nil
}
