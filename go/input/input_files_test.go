package input

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"gobees/internal/examplefiles"
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
		want := params.FlowerSpecies{Name: "Alsike_clover", Pollen_g_per_flower: 0.0005, Nectar_ml_per_flower: 0.00046, ProteinPollenProp: 0.2, Concentration_mol_per_l: 0.7, StartDay: 181, StopDay: 242, CorollaDepth_mm: 10, IntFlowerTime_s: 0.6}
		if table.Rows[0] != want {
			t.Errorf("first row = %+v, want %+v", table.Rows[0], want)
		}
		// """Average_Willow""",0.01,0.00261,0.25,1.13,59,119,0,0.6
		got, err := table.Get("Average_Willow")
		if err != nil || *got != (params.FlowerSpecies{Name: "Average_Willow", Pollen_g_per_flower: 0.01, Nectar_ml_per_flower: 0.00261, ProteinPollenProp: 0.25, Concentration_mol_per_l: 1.13, StartDay: 59, StopDay: 119, CorollaDepth_mm: 0, IntFlowerTime_s: 0.6}) {
			t.Errorf("Average_Willow = %+v, %v", got, err)
		}
		// last line: """Yellow_rattle""",0.01,0.00025,0.2,0.917323,121,243,13,0.6
		if table.Rows[78] != (params.FlowerSpecies{Name: "Yellow_rattle", Pollen_g_per_flower: 0.01, Nectar_ml_per_flower: 0.00025, ProteinPollenProp: 0.2, Concentration_mol_per_l: 0.917323, StartDay: 121, StopDay: 243, CorollaDepth_mm: 13, IntFlowerTime_s: 0.6}) {
			t.Errorf("last row = %+v", table.Rows[78])
		}
		if _, err := table.Get("Not_a_flower"); err == nil {
			t.Error("expected an error for a flower species missing in the file")
		}
	}
}

func TestParseSpeciesFileExample(t *testing.T) {
	constants := params.DefaultConstants()
	all := []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum", "Psithyrus"}
	for _, lineEnd := range []string{"\n", "\r\n"} { // the original file has CRLF
		species, err := parseSpeciesFile(strings.NewReader(strings.ReplaceAll(examplefiles.BumbleSpecies, "\n", lineEnd)), all, &constants)
		if err != nil {
			t.Fatal(err)
		}
		if len(species) != 7 {
			t.Fatalf("got %d species, want 7", len(species))
		}
		// B_terrestris: 91,28,["Grassland*" ...],6.9,11.1,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,62.4,249.5,1,0,17,590,980,32,0.2
		s := species[0]
		read := []float64{s.EmergingDay_mean, s.EmergingDay_sd, s.Proboscis_min_mm, s.Proboscis_max_mm, s.GrowthFactor, s.SeasonStop,
			s.MaxLifespanWorkers, s.Batchsize, s.FlightVelocity_m_per_s, s.SearchLength_m, s.TimeUnloading, s.SpecMax_cropVolume_myl,
			s.SpecMax_pollenPellets_g, s.MinToMaxFactor, s.DevAgeHatchingMin_d, s.DevAgePupationMin_d, s.DevAgeEmergingMin_d,
			s.DevWeightEgg_mg, s.DevWeightPupationMin_mg, s.DevWeightPupationMax_mg, s.PollenToBodymassFactor,
			s.Dev_Q_DeterminationWeight_mg, s.DevAge_Q_PupationMin_d, s.DevWeight_Q_PupationMin_mg, s.DevWeight_Q_PupationMax_mg,
			s.DevAge_Q_EmergingMin_d, s.ChanceFindNest}
		wantRead := []float64{91, 28, 6.9, 11.1, 1.88, 305, 60, 12, 5, 2500, 165, 173, 0.15, 2, 5, 12.9, 23.8, 1.5, 62.4, 249.5, 1, 0, 17, 590, 980, 32, 0.2}
		if s.Name != "B_terrestris" || !reflect.DeepEqual(read, wantRead) {
			t.Errorf("B_terrestris columns:\n got  %v %v\n want %v", s.Name, read, wantRead)
		}
		if !reflect.DeepEqual(s.NestHabitatsList, []string{"Grassland*", "Garden", "Hedgerow", "Scrub", "Woodland", "Nestboxes"}) {
			t.Errorf("nestHabitatsList = %q (not yet expanded)", s.NestHabitatsList)
		}

		// [SPEC-05], [SPEC-07] derived values, computed by hand
		derived := map[string][2]float64{
			"FlightCosts_kJ_per_m_per_mg (488.6 / 5 / 1e9)":   {s.FlightCosts_kJ_per_m_per_mg, 9.772e-8},
			"DevQuotaIncubationToday_kJ (10 / 18)":            {s.DevQuotaIncubationToday_kJ, 0.5555555555555556},
			"DevAgeHatchingMax_d (5 * 2)":                     {s.DevAgeHatchingMax_d, 10},
			"DevAgePupationMax_d (12.9 * 2 - 5)":              {s.DevAgePupationMax_d, 20.8},
			"DevAgeEmergingMax_d (23.8 * 2 - 12.9)":           {s.DevAgeEmergingMax_d, 34.7},
			"DevIncubationHatchingTH_kJ (10/18 * 5)":          {s.DevIncubationHatchingTH_kJ, 2.7777777777777777},
			"DevIncubationPupationTH_kJ (10/18 * 12.9)":       {s.DevIncubationPupationTH_kJ, 7.166666666666667},
			"DevIncubationEmergingTH_kJ (10/18 * 23.8)":       {s.DevIncubationEmergingTH_kJ, 13.222222222222221},
			"Dev_larvalAge_QueenDetermination_d (constant 3)": {s.Dev_larvalAge_QueenDetermination_d, 3},
			"DevAge_Q_PupationMax_d (17 * 2 - 5)":             {s.DevAge_Q_PupationMax_d, 29},
			"DevAge_Q_EmergingMax_d (32 * 2 - 17)":            {s.DevAge_Q_EmergingMax_d, 47},
			"DevIncubation_Q_PupationTH_kJ (10/18 * 17)":      {s.DevIncubation_Q_PupationTH_kJ, 9.444444444444445},
			"DevIncubation_Q_EmergingTH_kJ (10/18 * 32)":      {s.DevIncubation_Q_EmergingTH_kJ, 17.77777777777778},
			"MinPollenStore_g (0.5 * 0.001 * 62.4 * 12 / 1)":  {s.MinPollenStore_g, 0.3744},
		}
		for name, v := range derived {
			if math.Abs(v[0]-v[1]) > 1e-9*math.Max(1, math.Abs(v[1])) {
				t.Errorf("%s = %v, want %v", name, v[0], v[1])
			}
		}

		// finding 20 (see trans/netlogo-findings.md#finding-20): emergingDay_sd is rounded, .5 up
		wantSD := map[string]float64{"B_pascuorum": 15, "B_lapidarius": 21, "B_pratorum": 31, "B_hortorum": 8}
		for _, sp := range species {
			if want, ok := wantSD[sp.Name]; ok && sp.EmergingDay_sd != want {
				t.Errorf("%s: emergingDay_sd = %v, want %v", sp.Name, sp.EmergingDay_sd, want)
			}
		}
	}
}

func TestParseSpeciesFileOrder(t *testing.T) {
	// finding 20 (see trans/netlogo-findings.md#finding-20): species are created in FILE order, a name missing in the file creates no species.
	constants := params.DefaultConstants()
	species, err := parseSpeciesFile(strings.NewReader(examplefiles.BumbleSpecies),
		[]string{"B_pascuorum", "B_terrestris", "B_missing", "b_hortorum"}, &constants)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, s := range species {
		names = append(names, s.Name)
	}
	if !reflect.DeepEqual(names, []string{"B_terrestris", "B_pascuorum"}) {
		t.Errorf("species = %q, want B_terrestris, B_pascuorum (file order, case-sensitive)", names)
	}
}
