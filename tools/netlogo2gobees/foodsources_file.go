package main

import (
	"fmt"
	"io"
	"strings"

	"gobees/params"
)

// netLogoPatch is one data line of a NetLogo foodsources file, with all columns. patchColour, perimeter_m and info
// are only for the NetLogo GUI and are not converted.
type netLogoPatch struct {
	Id_Beescout       float64
	PatchType         string
	PatchColour       float64 // GUI only
	X, Y              float64
	Area_sqm          float64
	FlowerSpeciesList []params.FlowerEntry // the items with a value
	Options           []string             // the items without a value of flowerSpeciesList, e.g. ["Plot"], not converted
	Perimeter_m       float64              // GUI only
	PatchInfo         string               // GUI only
}

// parseFoodsourcesFile reads a foodsources file [FSRC-02] to [FSRC-04].
//
// Returns the metres per NetLogo patch from line 1 and one flower patch per data line, in file order.
// Header handling as in NetLogo (finding 21, see trans/netlogo-findings.md#finding-21): line 2 is skipped, and line 3 too if line 2 has at most 10 characters.
// Every data line: id patchType patchColour xcor ycor size_sqm flowerSpeciesList perimeter_m info,
// columns 1-8 as NetLogo values, column 9 is the raw rest of the line.
func parseFoodsourcesFile(r io.Reader) (float64, []netLogoPatch, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) < 2 {
		return 0, nil, fmt.Errorf("expected at least 2 lines")
	}

	// [FSRC-02] line 1: metres per NetLogo patch
	v, _, err := ReadValue(lines[0], 0)
	if err != nil {
		return 0, nil, fmt.Errorf("line 1: %w", err)
	}
	metresPerPatch, ok := v.(float64)
	if !ok {
		return 0, nil, fmt.Errorf("line 1: expected a number")
	}

	// [FSRC-03] skip line 2, and line 3 if line 2 has at most 10 characters (old format: number of flower patches)
	first := 2
	if len(lines[1]) <= 10 {
		first = 3
	}

	// [FSRC-04] data lines. file-read skips white space incl. line breaks, so blank lines are skipped.
	patches := []netLogoPatch{}
	for i := first; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		p, err := parseFlowerPatch(lines[i])
		if err != nil {
			return 0, nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		patches = append(patches, p)
	}
	return metresPerPatch, patches, nil
}

func parseFlowerPatch(line string) (netLogoPatch, error) {
	p := netLogoPatch{}
	pos := 0
	numbers := []*float64{&p.Id_Beescout, nil, &p.PatchColour, &p.X, &p.Y, &p.Area_sqm, nil, &p.Perimeter_m}
	for col := range 8 {
		v, next, err := ReadValue(line, pos)
		if err != nil {
			return p, fmt.Errorf("column %d: %w", col+1, err)
		}
		pos = next
		switch col {
		case 1:
			s, ok := v.(string)
			if !ok {
				return p, fmt.Errorf("column 2 (patchType): expected a string")
			}
			p.PatchType = s
		case 6:
			list, options, err := toFlowerSpeciesList(v)
			if err != nil {
				return p, fmt.Errorf("column 7 (flowerSpeciesList): %w", err)
			}
			p.FlowerSpeciesList, p.Options = list, options
		default:
			x, ok := v.(float64)
			if !ok {
				return p, fmt.Errorf("column %d: expected a number", col+1)
			}
			*numbers[col] = x
		}
	}
	p.PatchInfo = line[pos:] // file-read-line: the raw rest of the line
	return p, nil
}

// toFlowerSpeciesList splits a flowerSpeciesList into the items with a value (flower species) and the names of the
// items without a value (stewardship options, e.g. ["Plot"]). In NetLogo an option marks the flower species after it
// as growing in the option (occurrence of the layer), which only the GUI reads [LAYR-11], [LAYR-12].
func toFlowerSpeciesList(v any) ([]params.FlowerEntry, []string, error) {
	list, ok := v.([]any)
	if !ok {
		return nil, nil, fmt.Errorf("expected a list")
	}
	entries := make([]params.FlowerEntry, 0, len(list))
	var options []string
	for _, item := range list {
		sub, ok := item.([]any)
		if !ok || len(sub) == 0 {
			return nil, nil, fmt.Errorf("expected [name value] items")
		}
		name, ok := sub[0].(string)
		if !ok {
			return nil, nil, fmt.Errorf("expected a name as first item")
		}
		if len(sub) == 1 {
			options = append(options, name)
			continue
		}
		x, ok := sub[1].(float64)
		if !ok {
			return nil, nil, fmt.Errorf("%s: expected a number as second item", name)
		}
		entries = append(entries, params.FlowerEntry{Name: name, Value: x})
	}
	return entries, options, nil
}
