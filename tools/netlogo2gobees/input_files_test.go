package main

import (
	"reflect"
	"strings"
	"testing"

	"gobees/examplefiles"
	"gobees/params"
)

// Tests of the readers for the step-2 input files, with the complete example files (examplefiles).

func TestParseFoodsourcesFileExampleFarmPlotTest(t *testing.T) {
	// the second map of the crop rotation experiment (go/validation/experiments/step7_crop.xml)
	metres, patches, err := parseFoodsourcesFile(strings.NewReader(examplefiles.ExampleFarmPlotTestFoodsources))
	if err != nil {
		t.Fatal(err)
	}
	// [FSRC-02] line 1: "10" metres per patch. [FSRC-03] line 2 "36" (<= 10 characters), so line 3 is skipped too.
	if metres != 10 || len(patches) != 36 {
		t.Fatalf("metres per patch %v, %d flower patches; want 10 and 36", metres, len(patches))
	}
	// Line 4: 73 "Crop_Cereals" 75 97.97 91.255 36500 [["Crop_Cereals" 1]] 940  "no info "
	p := patches[0]
	got := []any{p.Id_Beescout, p.PatchType, p.PatchColour, p.X, p.Y, p.Area_sqm, p.FlowerSpeciesList, p.Perimeter_m, p.PatchInfo}
	want := []any{73.0, "Crop_Cereals", 75.0, 97.97, 91.255, 36500.0, []params.FlowerEntry{{Name: "Crop_Cereals", Value: 1, HasValue: true}}, 940.0, `  "no info " `}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("patch 0:\n got  %#v\n want %#v", got, want)
	}
	// Last line: 1767 "Hedgerow" 105 135 59.684 3400 [...] 190  "no info "
	p = patches[35]
	got = []any{p.Id_Beescout, p.PatchType, p.X, p.Y, p.Area_sqm, p.Perimeter_m, p.PatchInfo}
	want = []any{1767.0, "Hedgerow", 135.0, 59.684, 3400.0, 190.0, `  "no info " `}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("patch 35:\n got  %#v\n want %#v", got, want)
	}
}

func TestParseFoodsourcesFileExampleFarm(t *testing.T) {
	metres, patches, err := parseFoodsourcesFile(strings.NewReader(examplefiles.ExampleFarmFoodsources))
	if err != nil {
		t.Fatal(err)
	}
	// [FSRC-02] line 1: "10" metres per patch. [FSRC-03] line 2 "73" (<= 10 characters), so line 3 is skipped too.
	if metres != 10 {
		t.Errorf("metres per patch = %v, want 10", metres)
	}
	if len(patches) != 73 {
		t.Fatalf("got %d flower patches, want 73 (finding 21, see trans/netlogo-findings.md#finding-21)", len(patches))
	}

	// Line 4: 0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275] ... ["Wild_teasel" 0.023]] 20  "no info"
	p := patches[0]
	got := []any{p.Id_Beescout, p.PatchType, p.PatchColour, p.X, p.Y, p.Area_sqm, len(p.FlowerSpeciesList), p.Perimeter_m, p.PatchInfo}
	want := []any{0.0, "Hedgerow", 105.0, 130.0, 175.0, 200.0, 33, 20.0, `  "no info" `}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("patch 0:\n got  %#v\n want %#v", got, want)
	}
	wantEntries := map[int]params.FlowerEntry{
		0:  {Name: "Average_Willow", Value: 1.275, HasValue: true},
		1:  {Name: "Birdsfoot_trefoil", Value: 0.037, HasValue: true},
		2:  {Name: "Blackthorn", Value: 0.139, HasValue: true},
		32: {Name: "Wild_teasel", Value: 0.023, HasValue: true},
	}
	for i, e := range wantEntries {
		if p.FlowerSpeciesList[i] != e {
			t.Errorf("patch 0, flower %d = %+v, want %+v", i, p.FlowerSpeciesList[i], e)
		}
	}

	// Line 5: 1 "Crop_Cereals" 75 125.948 141.856 105900 [["Crop_Cereals" 1]] 2080  "no info"
	p = patches[1]
	got = []any{p.Id_Beescout, p.PatchType, p.PatchColour, p.X, p.Y, p.Area_sqm, p.FlowerSpeciesList, p.Perimeter_m, p.PatchInfo}
	want = []any{1.0, "Crop_Cereals", 75.0, 125.948, 141.856, 105900.0, []params.FlowerEntry{{Name: "Crop_Cereals", Value: 1, HasValue: true}}, 2080.0, `  "no info" `}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("patch 1:\n got  %#v\n want %#v", got, want)
	}

	// Last line: 72 "Woodland" 35 212.2 34.4 500 [...] 50  "no info"
	p = patches[72]
	got = []any{p.Id_Beescout, p.PatchType, p.X, p.Y, p.Area_sqm, p.Perimeter_m}
	want = []any{72.0, "Woodland", 212.2, 34.4, 500.0, 50.0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("patch 72:\n got  %#v\n want %#v", got, want)
	}
}

func TestParseFoodsourcesFileHeaderFormats(t *testing.T) {
	line := `5 "Scrub" 55 1 2 300 [["Bramble" 0.5]] 10  "no info" ` + "\n"
	header := "id patchType patchColour xcor ycor size_sqm flowerSpeciesList perimeter_m info\n"
	cases := map[string]string{
		"old format (line 2: number of patches, line 3: header)": "15\n1\n" + header + line,
		"new format (line 2: header, > 10 characters)":           "15\n" + header + line,
		"CRLF line endings": strings.ReplaceAll("15\n1\n"+header+line, "\n", "\r\n"),
	}
	for name, text := range cases {
		metres, patches, err := parseFoodsourcesFile(strings.NewReader(text))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if metres != 15 || len(patches) != 1 || patches[0].PatchType != "Scrub" || patches[0].PatchInfo != `  "no info" ` {
			t.Errorf("%s: got %v m, %+v", name, metres, patches)
		}
	}
}

func TestParseFlowerspeciesFileExample(t *testing.T) {
	for _, lineEnd := range []string{"\n", "\r\n"} { // the original file has CRLF
		table, err := parseFlowerspeciesFile(strings.NewReader(strings.ReplaceAll(examplefiles.Flowerspecies, "\n", lineEnd)))
		if err != nil {
			t.Fatal(err)
		}
		if len(table.Rows) != 79 {
			t.Fatalf("got %d flower species, want 79", len(table.Rows))
		}
		// first line: """Alsike_clover""",0.0005,0.00046,0.2,0.7,181,242,10,0.6
		want := params.FlowerSpecies{Name: "Alsike_clover", PollenPerFlower: 0.0005, NectarPerFlower: 0.00046, Concentration: 0.7, StartDay: 181, StopDay: 242, CorollaDepth: 10, InterFlowerTime: 0.6}
		if table.Rows[0] != want {
			t.Errorf("first row = %+v, want %+v", table.Rows[0], want)
		}
		// """Average_Willow""",0.01,0.00261,0.25,1.13,59,119,0,0.6
		got, err := table.Get("Average_Willow")
		if err != nil || *got != (params.FlowerSpecies{Name: "Average_Willow", PollenPerFlower: 0.01, NectarPerFlower: 0.00261, Concentration: 1.13, StartDay: 59, StopDay: 119, CorollaDepth: 0, InterFlowerTime: 0.6}) {
			t.Errorf("Average_Willow = %+v, %v", got, err)
		}
		// last line: """Yellow_rattle""",0.01,0.00025,0.2,0.917323,121,243,13,0.6
		if table.Rows[78] != (params.FlowerSpecies{Name: "Yellow_rattle", PollenPerFlower: 0.01, NectarPerFlower: 0.00025, Concentration: 0.917323, StartDay: 121, StopDay: 243, CorollaDepth: 13, InterFlowerTime: 0.6}) {
			t.Errorf("last row = %+v", table.Rows[78])
		}
		if _, err := table.Get("Not_a_flower"); err == nil {
			t.Error("expected an error for a flower species missing in the file")
		}
	}
}

func TestParseSpeciesFileExample(t *testing.T) {
	// All rows in file order. gobees-fast selects the species and derives the other values (package input).
	for _, lineEnd := range []string{"\n", "\r\n"} { // the original file has CRLF
		species, err := parseSpeciesFile(strings.NewReader(strings.ReplaceAll(examplefiles.BumbleSpecies, "\n", lineEnd)))
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, s := range species {
			names = append(names, s.Name)
		}
		want := []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum", "Psithyrus"}
		if !reflect.DeepEqual(names, want) {
			t.Fatalf("species = %q, want %q", names, want)
		}
		// B_terrestris: 91,28,["Grassland*" ...],6.9,11.1,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,62.4,249.5,1,0,17,590,980,32,0.2
		s := species[0]
		read := []float64{s.EmergenceDayMean, s.EmergenceDaySD, s.ProboscisMin, s.ProboscisMax, s.GrowthFactor, float64(s.SeasonStop),
			float64(s.MaxLifespanWorkers), float64(s.BatchSize), s.FlightVelocity, s.SearchLength, s.UnloadingTime, s.MaxCropVolume,
			s.MaxPollenPellets, s.MinToMaxFactor, s.HatchingAgeMin, s.PupationAgeMin, s.EmergenceAgeMin,
			s.EggWeight, s.PupationWeightMin, s.PupationWeightMax, s.PollenToBodymassFactor,
			s.QueenDeterminationWeight, s.QueenPupationAgeMin, s.QueenPupationWeightMin, s.QueenPupationWeightMax,
			s.QueenEmergenceAgeMin, s.DailyNestSiteChance}
		wantRead := []float64{91, 28, 6.9, 11.1, 1.88, 305, 60, 12, 5, 2500, 165, 173, 0.15, 2, 5, 12.9, 23.8, 1.5, 62.4, 249.5, 1, 0, 17, 590, 980, 32, 0.2}
		if s.Name != "B_terrestris" || !reflect.DeepEqual(read, wantRead) {
			t.Errorf("B_terrestris columns:\n got  %v %v\n want %v", s.Name, read, wantRead)
		}
		if !reflect.DeepEqual(s.NestHabitats, []string{"Grassland*", "Garden", "Hedgerow", "Scrub", "Woodland", "Nestboxes"}) {
			t.Errorf("nestHabitatsList = %q (not yet expanded)", s.NestHabitats)
		}
		// only the row named Psithyrus is parasitic
		for _, s := range species {
			if s.Parasitic != (s.Name == "Psithyrus") {
				t.Errorf("%s: parasitic = %v", s.Name, s.Parasitic)
			}
		}
		// finding 20 (see trans/netlogo-findings.md#finding-20): emergingDay_sd as in the file, gobees-fast rounds it
		if species[1].EmergenceDaySD != 14.5 {
			t.Errorf("B_pascuorum: emergingDay_sd = %v, want 14.5 as in the file", species[1].EmergenceDaySD)
		}
	}
}
