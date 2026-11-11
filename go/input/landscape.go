package input

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gobees/params"
)

// Landscape is a landscape as the model reads it: the flower patches, each with its own flower list (NetLogo:
// foodsources file). On disk it is a landscape file (YAML) with tables (CSV), see ReadLandscape and WriteLandscape.
type Landscape struct {
	Patches []Patch // flower patches, in this order
}

// Patch is one flower patch of a landscape. The names are the columns of the NetLogo foodsources file, except x, y
// (xcor, ycor in NetLogo patches), area (size_sqm) and flowers (flowerSpeciesList).
type Patch struct {
	ID        int
	PatchType string
	X, Y      float64              // centre [m]
	Area      float64              // [m2]
	Flowers   []params.FlowerEntry // the flower species with their relative abundance, in this order
}

// landscapeFile is the content of a landscape file (YAML): the names of its tables, relative to the folder of the
// landscape file. One habitat table can serve many landscapes, e.g. ../habitats.csv.
type landscapeFile struct {
	Patches      string `yaml:"patches"`                // table of the flower patches: id, patchType, x, y, area
	Habitats     string `yaml:"habitats,omitempty"`     // table of the flower lists by patch type: patchType, flower, abundance
	PatchFlowers string `yaml:"patchFlowers,omitempty"` // table of own flower lists: id, flower, abundance
}

// Columns of the tables of a landscape file. Each table needs a header with all its columns, in any order.
var (
	patchColumns       = []string{"id", "patchType", "x", "y", "area"}
	habitatColumns     = []string{"patchType", "flower", "abundance"}
	patchFlowerColumns = []string{"id", "flower", "abundance"}
)

// LandscapeTable is a table of a landscape file, as read.
type LandscapeTable struct {
	Key  string // key in the landscape file: patches, habitats or patchFlowers
	Ref  string // value of the key, relative to the folder of the landscape file
	Name string // name of the table file as opened, relative to the folder of the run file
}

// ReadLandscape reads the landscape file name and its tables with open. The names of the tables are relative to the
// folder of the landscape file. It returns the landscape, its tables and the bytes of every file read, by its name as
// opened. An unknown key of the landscape file, an unknown or missing column of a table and a value that is not a
// number (an integer for id) where a number is needed are errors.
//
// The flower list of a patch is the list of its patch type in the habitat table (the rows of the patch type, in their
// order), or, if the table patchFlowers has rows of the patch, these rows.
func ReadLandscape(name string, open func(string) (io.ReadCloser, error)) (Landscape, []LandscapeTable, map[string][]byte, error) {
	raw := map[string][]byte{}
	var lf landscapeFile
	data, err := readFile(name, open, func(r io.Reader) error { return decodeStrict(r, &lf) })
	if err != nil {
		return Landscape{}, nil, nil, err
	}
	raw[name] = data
	if lf.Patches == "" {
		return Landscape{}, nil, nil, fmt.Errorf("%s: patches: missing, the landscape file needs a table of the flower patches", name)
	}

	tables := []LandscapeTable{}
	read := func(key, ref string, columns []string) ([]map[string]string, error) {
		if ref == "" {
			return nil, nil
		}
		t := LandscapeTable{Key: key, Ref: ref, Name: tableName(name, ref)}
		tables = append(tables, t)
		var rows []map[string]string
		b, err := readFile(t.Name, open, func(r io.Reader) (err error) { rows, err = readTable(r, columns); return })
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", name, key, err)
		}
		raw[t.Name] = b
		return rows, nil
	}
	patchRows, err := read("patches", lf.Patches, patchColumns)
	if err != nil {
		return Landscape{}, nil, nil, err
	}
	habitatRows, err := read("habitats", lf.Habitats, habitatColumns)
	if err != nil {
		return Landscape{}, nil, nil, err
	}
	ownRows, err := read("patchFlowers", lf.PatchFlowers, patchFlowerColumns)
	if err != nil {
		return Landscape{}, nil, nil, err
	}

	l, err := resolveLandscape(patchRows, habitatRows, ownRows)
	if err != nil {
		return Landscape{}, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	return l, tables, raw, nil
}

// tableName returns the name of the table ref of the landscape file name: relative to the folder of the landscape
// file, unless ref is absolute.
func tableName(name, ref string) string {
	if filepath.IsAbs(ref) {
		return ref
	}
	return path.Join(path.Dir(filepath.ToSlash(name)), filepath.ToSlash(ref))
}

// resolveLandscape builds the landscape from the rows of its tables.
func resolveLandscape(patchRows, habitatRows, ownRows []map[string]string) (Landscape, error) {
	l := Landscape{Patches: make([]Patch, len(patchRows))}
	byType := map[string][]params.FlowerEntry{}
	for i, r := range habitatRows {
		e, err := flowerEntry(r)
		if err != nil {
			return l, fmt.Errorf("habitats, row %d: %w", i+2, err)
		}
		byType[r["patchType"]] = append(byType[r["patchType"]], e)
	}
	own := map[int][]params.FlowerEntry{}
	ownOrder := []int{}
	for i, r := range ownRows {
		id, err := integer(r, "id")
		if err != nil {
			return l, fmt.Errorf("patchFlowers, row %d: %w", i+2, err)
		}
		e, err := flowerEntry(r)
		if err != nil {
			return l, fmt.Errorf("patchFlowers, row %d: %w", i+2, err)
		}
		if _, ok := own[id]; !ok {
			ownOrder = append(ownOrder, id)
		}
		own[id] = append(own[id], e)
	}

	ids := map[int]int{}
	for i, r := range patchRows {
		p := Patch{PatchType: r["patchType"]}
		var err error
		if p.ID, err = integer(r, "id"); err != nil {
			return l, fmt.Errorf("patches, row %d: %w", i+2, err)
		}
		for _, c := range []struct {
			key string
			v   *float64
		}{{"x", &p.X}, {"y", &p.Y}, {"area", &p.Area}} {
			if *c.v, err = number(r, c.key); err != nil {
				return l, fmt.Errorf("patches, row %d: %w", i+2, err)
			}
		}
		p.Flowers = slices.Clone(byType[p.PatchType])
		if list, ok := own[p.ID]; ok {
			p.Flowers = slices.Clone(list)
		}
		ids[p.ID]++
		l.Patches[i] = p
	}
	for _, id := range ownOrder {
		switch ids[id] {
		case 0:
			return l, fmt.Errorf("patchFlowers: id %d is no flower patch of the table patches", id)
		case 1:
		default:
			return l, fmt.Errorf("patchFlowers: id %d belongs to %d flower patches of the table patches", id, ids[id])
		}
	}
	return l, nil
}

// flowerEntry reads the columns flower and abundance of a row.
func flowerEntry(r map[string]string) (params.FlowerEntry, error) {
	v, err := number(r, "abundance")
	if err != nil {
		return params.FlowerEntry{}, err
	}
	if r["flower"] == "" {
		return params.FlowerEntry{}, fmt.Errorf("flower: empty")
	}
	return params.FlowerEntry{Name: r["flower"], Value: v}, nil
}

// number reads the column key of a row as a number.
func number(r map[string]string, key string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(r[key]), 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is no number", key, r[key])
	}
	return v, nil
}

// integer reads the column key of a row as an integer.
func integer(r map[string]string, key string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(r[key]))
	if err != nil {
		return 0, fmt.Errorf("%s: %q is no integer", key, r[key])
	}
	return v, nil
}

// readTable reads a CSV table with a header of exactly the given columns, in any order. It returns the rows by
// column name.
func readTable(r io.Reader, columns []string) ([]map[string]string, error) {
	cr := csv.NewReader(r)
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty file, want the header %s", strings.Join(columns, ","))
	}
	header := records[0]
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], string(rune(0xFEFF))) // byte order mark of a file from Excel
	}
	for i, h := range header {
		header[i] = strings.TrimSpace(h)
		if !slices.Contains(columns, header[i]) {
			return nil, fmt.Errorf("unknown column %q, want the columns %s (R: write.csv(..., row.names = FALSE))",
				header[i], strings.Join(columns, ","))
		}
		if slices.Index(header, header[i]) != i {
			return nil, fmt.Errorf("column %q twice", header[i])
		}
	}
	for _, c := range columns {
		if !slices.Contains(header, c) {
			return nil, fmt.Errorf("column %q missing, want the columns %s", c, strings.Join(columns, ","))
		}
	}
	rows := make([]map[string]string, len(records)-1)
	for i, rec := range records[1:] {
		rows[i] = map[string]string{}
		for j, h := range header {
			rows[i][h] = rec[j]
		}
	}
	return rows, nil
}

// LandscapeFiles are the files of a landscape: the landscape file and its tables, by their names relative to the
// folder of the landscape file.
type LandscapeFiles map[string][]byte

// WriteLandscape returns the files of the landscape l under the name name (a file name without folder): the
// landscape file and its tables name_patches.csv, name_habitats.csv and, if a patch has its own list,
// name_patchFlowers.csv (name without .yaml).
//
// The habitat table holds one list per patch type, the most frequent list of its patches (tie: the first in the
// file). A patch type with a patch without flowers has no list. Every patch with another list has its own list in
// the table patchFlowers, so ReadLandscape gives l again.
func WriteLandscape(name string, l Landscape) (LandscapeFiles, error) {
	base := strings.TrimSuffix(name, ".yaml")
	lf := landscapeFile{Patches: base + "_patches.csv"}

	types := []string{}
	lists := map[string][][]params.FlowerEntry{}
	for _, p := range l.Patches {
		if _, ok := lists[p.PatchType]; !ok {
			types = append(types, p.PatchType)
		}
		lists[p.PatchType] = append(lists[p.PatchType], p.Flowers)
	}
	habitat := map[string][]params.FlowerEntry{}
	for _, t := range types {
		habitat[t] = mostFrequent(lists[t])
	}

	patches := [][]string{patchColumns}
	habitats := [][]string{habitatColumns}
	own := [][]string{patchFlowerColumns}
	ids := map[int]int{}
	for _, p := range l.Patches {
		ids[p.ID]++
	}
	for _, t := range types {
		for _, e := range habitat[t] {
			habitats = append(habitats, []string{t, e.Name, fmtNumber(e.Value)})
		}
	}
	for _, p := range l.Patches {
		patches = append(patches, []string{strconv.Itoa(p.ID), p.PatchType, fmtNumber(p.X), fmtNumber(p.Y),
			fmtNumber(p.Area)})
		if slices.Equal(p.Flowers, habitat[p.PatchType]) {
			continue
		}
		if ids[p.ID] > 1 {
			return nil, fmt.Errorf("patch %d: its own flower list needs a unique id, %d patches have it",
				p.ID, ids[p.ID])
		}
		for _, e := range p.Flowers {
			own = append(own, []string{strconv.Itoa(p.ID), e.Name, fmtNumber(e.Value)})
		}
	}

	files := LandscapeFiles{}
	if len(habitats) > 1 {
		lf.Habitats = base + "_habitats.csv"
		files[lf.Habitats] = writeTable(habitats)
	}
	if len(own) > 1 {
		lf.PatchFlowers = base + "_patchFlowers.csv"
		files[lf.PatchFlowers] = writeTable(own)
	}
	files[lf.Patches] = writeTable(patches)
	var b bytes.Buffer
	if err := encode(&b, lf); err != nil {
		return nil, err
	}
	files[name] = b.Bytes()
	return files, nil
}

// mostFrequent returns the most frequent list of lists (tie: the first), or nil if one list is empty: an empty list
// cannot be an own list of the table patchFlowers.
func mostFrequent(lists [][]params.FlowerEntry) []params.FlowerEntry {
	best, bestCount := -1, 0
	for i, a := range lists {
		if len(a) == 0 {
			return nil
		}
		n := 0
		for _, b := range lists {
			if slices.Equal(a, b) {
				n++
			}
		}
		if n > bestCount {
			best, bestCount = i, n
		}
	}
	if best < 0 {
		return nil
	}
	return lists[best]
}

// RewriteLandscapeFile returns the landscape file data with the tables renamed by names (old name: new name), e.g.
// for the copies next to the provenance.
func RewriteLandscapeFile(data []byte, names map[string]string) ([]byte, error) {
	var lf landscapeFile
	if err := decodeStrict(bytes.NewReader(data), &lf); err != nil {
		return nil, err
	}
	for _, ref := range []*string{&lf.Patches, &lf.Habitats, &lf.PatchFlowers} {
		if n, ok := names[*ref]; ok {
			*ref = n
		}
	}
	var b bytes.Buffer
	err := encode(&b, lf)
	return b.Bytes(), err
}

// writeTable writes rows as CSV with line feeds.
func writeTable(rows [][]string) []byte {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	_ = w.WriteAll(rows) // a bytes.Buffer does not fail
	return b.Bytes()
}

// fmtNumber writes v with as few digits as read back exactly, without an exponent.
func fmtNumber(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
