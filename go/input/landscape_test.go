package input

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"gobees/params"
)

// mapOpener opens the files of files by name, for ReadLandscape.
func mapOpener(files map[string]string) func(string) (io.ReadCloser, error) {
	return func(name string) (io.ReadCloser, error) {
		text, ok := files[name]
		if !ok {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		return io.NopCloser(strings.NewReader(text)), nil
	}
}

// sameLandscape compares two landscapes, an empty list equal to a missing one.
func sameLandscape(a, b Landscape) bool {
	if len(a.Patches) != len(b.Patches) {
		return false
	}
	for i := range a.Patches {
		p, q := a.Patches[i], b.Patches[i]
		if p.ID != q.ID || p.PatchType != q.PatchType || p.X != q.X || p.Y != q.Y || p.Area != q.Area ||
			!slices.Equal(p.Flowers, q.Flowers) {
			return false
		}
	}
	return true
}

func TestWriteLandscape(t *testing.T) {
	// two hedges share the list of their patch type, a third has its own list, a scrub has no flowers. The written
	// tables, line by line, and read back the same landscape.
	hedge := []params.FlowerEntry{{Name: "Bramble", Value: 0.331}, {Name: "Dandelion", Value: 0.11}}
	l := Landscape{Patches: []Patch{
		{ID: 0, PatchType: "Hedgerow", X: 130, Y: 175, Area: 200, Flowers: hedge},
		{ID: 1, PatchType: "Hedgerow", X: 125.948, Y: 141.856, Area: 105900, Flowers: hedge},
		{ID: 2, PatchType: "Hedgerow", X: 1.5, Y: 2, Area: 300, Flowers: []params.FlowerEntry{{Name: "Bramble", Value: 1}}},
		{ID: 3, PatchType: "Scrub", X: 5, Y: 6, Area: 0.0001},
	}}
	files, err := WriteLandscape("farm.yaml", l)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"farm.yaml": "patches: farm_patches.csv\nhabitats: farm_habitats.csv\npatchFlowers: farm_patchFlowers.csv\n",
		"farm_patches.csv": "id,patchType,x,y,area\n0,Hedgerow,130,175,200\n" +
			"1,Hedgerow,125.948,141.856,105900\n2,Hedgerow,1.5,2,300\n3,Scrub,5,6,0.0001\n",
		"farm_habitats.csv":     "patchType,flower,abundance\nHedgerow,Bramble,0.331\nHedgerow,Dandelion,0.11\n",
		"farm_patchFlowers.csv": "id,flower,abundance\n2,Bramble,1\n",
	}
	got := map[string]string{}
	for n, b := range files {
		got[n] = string(b)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("written:\n%v\nwant:\n%v", got, want)
	}

	read, tables, raw, err := ReadLandscape("farm.yaml", mapOpener(want))
	if err != nil {
		t.Fatal(err)
	}
	if !sameLandscape(read, l) {
		t.Errorf("read back %+v, want %+v", read, l)
	}
	wantTables := []LandscapeTable{{"patches", "farm_patches.csv", "farm_patches.csv"},
		{"habitats", "farm_habitats.csv", "farm_habitats.csv"},
		{"patchFlowers", "farm_patchFlowers.csv", "farm_patchFlowers.csv"}}
	if !slices.Equal(tables, wantTables) || len(raw) != 4 || string(raw["farm_habitats.csv"]) != want["farm_habitats.csv"] {
		t.Errorf("tables %v, %d raw files", tables, len(raw))
	}
}

func TestWriteLandscapeWithoutHabitats(t *testing.T) {
	// a patch type with a patch without flowers has no list: the other patch of the type has its own list. Without
	// any list there is no habitat table and no table patchFlowers.
	l := Landscape{Patches: []Patch{
		{ID: 7, PatchType: "Scrub"},
		{ID: 8, PatchType: "Scrub", Flowers: []params.FlowerEntry{{Name: "Bramble", Value: 2}}},
	}}
	files, err := WriteLandscape("a.yaml", l)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := files["a_habitats.csv"]; ok || string(files["a_patchFlowers.csv"]) != "id,flower,abundance\n8,Bramble,2\n" {
		t.Errorf("files %v", files)
	}
	opened := map[string]string{}
	for n, b := range files {
		opened[n] = string(b)
	}
	if read, _, _, err := ReadLandscape("a.yaml", mapOpener(opened)); err != nil || !sameLandscape(read, l) {
		t.Errorf("read back %+v, %v", read, err)
	}

	files, err = WriteLandscape("b.yaml", Landscape{Patches: []Patch{{ID: 1, PatchType: "Scrub"}}})
	if err != nil || len(files) != 2 || !strings.Contains(string(files["b.yaml"]), "patches: b_patches.csv") {
		t.Errorf("files %v, %v; want the landscape file and the patch table only", files, err)
	}

	// an own list needs a unique id
	two := Landscape{Patches: []Patch{{ID: 1, PatchType: "Scrub", Flowers: []params.FlowerEntry{{Name: "A", Value: 1}}},
		{ID: 1, PatchType: "Scrub"}}}
	if _, err := WriteLandscape("c.yaml", two); err == nil {
		t.Error("an own list for an id of two patches: want an error")
	}
}

func TestReadLandscapeTableNames(t *testing.T) {
	// the tables are relative to the folder of the landscape file, so two sites can share one habitat table
	files := map[string]string{
		"sites/a/land.yaml":   "patches: patches.csv\nhabitats: ../habitats.csv\n",
		"sites/a/patches.csv": "area,id,patchType,x,y\n200,0,Hedgerow,1,2\n", // columns in any order
		"sites/habitats.csv":  string(rune(0xFEFF)) + "patchType,flower,abundance\r\nHedgerow,Bramble,0.5\r\n",
	}
	l, tables, raw, err := ReadLandscape("sites/a/land.yaml", mapOpener(files))
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 || tables[1].Name != "sites/habitats.csv" || tables[1].Ref != "../habitats.csv" || len(raw) != 3 {
		t.Errorf("tables %+v, %d raw files", tables, len(raw))
	}
	want := Patch{ID: 0, PatchType: "Hedgerow", X: 1, Y: 2, Area: 200,
		Flowers: []params.FlowerEntry{{Name: "Bramble", Value: 0.5}}}
	if len(l.Patches) != 1 || !sameLandscape(l, Landscape{Patches: []Patch{want}}) {
		t.Errorf("read %+v", l)
	}
}

func TestReadLandscapeErrors(t *testing.T) {
	land := "patches: p.csv\nhabitats: h.csv\npatchFlowers: f.csv\n"
	patches := "id,patchType,x,y,area\n1,Scrub,1,2,3\n"
	habitats := "patchType,flower,abundance\nScrub,Bramble,1\n"
	own := "id,flower,abundance\n1,Bramble,2\n"
	for _, tc := range []struct {
		name, want                     string
		land, patches, habitats, flows string
	}{
		{"unknown key", "field other", land + "other: 1\n", patches, habitats, own},
		{"no patch table", "patches: missing", "habitats: h.csv\n", patches, habitats, own},
		{"old landscape file", "field metresPerPatch", "metresPerPatch: 10\n" + land, patches, habitats, own},
		{"missing table", "p.csv", land, "", habitats, own},
		{"row names of R", `unknown column ""`, land, `"",id,patchType,x,y,area` + "\n", habitats, own},
		{"missing column", `column "area" missing`, land, "id,patchType,x,y\n", habitats, own},
		{"column twice", `column "id" twice`, land, "id,id,patchType,x,y,area\n", habitats, own},
		{"no number", `area: "big" is no number`, land, "id,patchType,x,y,area\n1,Scrub,1,2,big\n",
			habitats, own},
		{"no abundance", `abundance: "" is no number`, land, patches, "patchType,flower,abundance\nScrub,Bramble,\n", own},
		{"no flower", "flower: empty", land, patches, "patchType,flower,abundance\nScrub,,1\n", own},
		{"id no integer", `id: "1.5" is no integer`, land, "id,patchType,x,y,area\n1.5,Scrub,1,2,3\n", habitats, own},
		{"unknown id", "id 9 is no flower patch", land, patches, habitats, "id,flower,abundance\n9,Bramble,2\n"},
		{"id twice", "id 1 belongs to 2 flower patches", land, patches + "1,Scrub,5,6,7\n", habitats, own},
		{"empty table", "empty file", land, patches, "", own},
	} {
		files := map[string]string{"land.yaml": tc.land, "p.csv": tc.patches, "h.csv": tc.habitats, "f.csv": tc.flows}
		if tc.patches == "" {
			delete(files, "p.csv")
		}
		_, _, _, err := ReadLandscape("land.yaml", mapOpener(files))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestRewriteLandscapeFile(t *testing.T) {
	data := []byte("patches: p.csv\nhabitats: ../h.csv\n")
	got, err := RewriteLandscapeFile(data, map[string]string{"../h.csv": "h_2.csv"})
	if err != nil || string(got) != "patches: p.csv\nhabitats: h_2.csv\n" {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestExampleFarmTables(t *testing.T) {
	// the first lines of the tables of the example farm, as in go/input/examplefarm, read value by value
	// (_SYSTEM_Example_Farm_Foodsources.txt, 10 m per patch: 0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275]
	// ["Birdsfoot_trefoil" 0.037] ...] and 1 "Crop_Cereals" 75 125.948 141.856 105900 [["Crop_Cereals" 1]]).
	files := map[string]string{
		"land.yaml": "patches: p.csv\nhabitats: h.csv\n",
		"p.csv":     "id,patchType,x,y,area\n0,Hedgerow,1300,1750,200\n1,Crop_Cereals,1259.48,1418.56,105900\n",
		"h.csv": "patchType,flower,abundance\nHedgerow,Average_Willow,1.275\nHedgerow,Birdsfoot_trefoil,0.037\n" +
			"Crop_Cereals,Crop_Cereals,1\n",
	}
	for name, prefix := range map[string]string{
		"_SYSTEM_Example_Farm_Foodsources_patches.csv":  files["p.csv"],
		"_SYSTEM_Example_Farm_Foodsources_habitats.csv": "patchType,flower,abundance\nHedgerow,Average_Willow,1.275\nHedgerow,Birdsfoot_trefoil,0.037\n",
	} {
		b, err := exampleFarm.ReadFile("examplefarm/" + name)
		if err != nil || !bytes.HasPrefix(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), []byte(prefix)) {
			t.Errorf("%s does not start with\n%s", name, prefix)
		}
	}
	l, _, _, err := ReadLandscape("land.yaml", mapOpener(files))
	if err != nil {
		t.Fatal(err)
	}
	want := Landscape{Patches: []Patch{
		{ID: 0, PatchType: "Hedgerow", X: 1300, Y: 1750, Area: 200,
			Flowers: []params.FlowerEntry{{Name: "Average_Willow", Value: 1.275}, {Name: "Birdsfoot_trefoil", Value: 0.037}}},
		{ID: 1, PatchType: "Crop_Cereals", X: 1259.48, Y: 1418.56, Area: 105900,
			Flowers: []params.FlowerEntry{{Name: "Crop_Cereals", Value: 1}}},
	}}
	if !sameLandscape(l, want) {
		t.Errorf("read %+v, want %+v", l, want)
	}
}
