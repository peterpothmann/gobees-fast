package input

import (
	"bytes"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"

	"gobees/globals"
	"gobees/params"
)

// exampleSpecies returns the species of the example farm's run file by name (all rows of the BeespeciesFile).
func exampleSpecies(t *testing.T) map[string]Species {
	t.Helper()
	c, _ := ExampleFarmConfig()
	m := map[string]Species{}
	for _, s := range c.Species {
		m[s.Species] = s
	}
	return m
}

func TestExampleFarm(t *testing.T) {
	p := ExampleFarm()
	in := p.Input
	// [PAR-01], [PAR-04] to [PAR-07]
	want := globals.Derived{
		BeeSpeciesInitialQueensList: []globals.SpeciesQueens{{Species: "B_terrestris", Queens: 500}},
		SpeciesList:                 []string{"B_terrestris"},
		MortalityForager_per_s:      1.0e-05,
		FoodsourcesFile:             "_SYSTEM_Example_Farm_Foodsources.yaml",
		CSS_OptionsList:             []string{"Legume", "Margin", "newCSO_Legume", "BirdBumble_Margin", "GameCover_Plot", "Plot"},
	}
	if !reflect.DeepEqual(in.Derived, want) {
		t.Errorf("Derived:\n got  %+v\n want %+v", in.Derived, want)
	}
	// [FPER] "Constant 8 hrs": 365 times 28800 s
	if len(in.Weather.Years_s) != 1 || len(in.Weather.Years_s[0]) != 365 || slices.ContainsFunc(in.Weather.Years_s[0],
		func(v float64) bool { return v != 28800 }) {
		t.Errorf("Weather: want one year of 28800 s")
	}
	if len(in.Species) != 1 || in.Species[0].Name != "B_terrestris" {
		t.Errorf("Species: want B_terrestris only, got %d", len(in.Species))
	}
	if in.Flowers == nil || len(in.Flowers.Rows) != 79 || len(in.Flowers.Names) != 79 {
		t.Errorf("Flowers: want the 79 flower species of _SYSTEM_Flowerspecies.csv")
	}
	fs, ok := in.Foodsources[in.Derived.FoodsourcesFile]
	if len(in.Foodsources) != 1 || !ok || fs.MetresPerPatch != 10 || len(fs.Patches) != 73 {
		t.Errorf("Foodsources: want only the example farm with 10 m per patch and 73 flower patches")
	}
	// first line of _SYSTEM_Example_Farm_Foodsources.txt: 0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275] ...] 20
	first := fs.Patches[0]
	if first.Id_Beescout != 0 || first.PatchType != "Hedgerow" || first.X != 130 || first.Y != 175 || first.Area_sqm != 200 ||
		len(first.FlowerSpeciesList) != 33 ||
		first.FlowerSpeciesList[0] != (params.FlowerEntry{Name: "Average_Willow", Value: 1.275, HasValue: true}) {
		t.Errorf("first flower patch = %+v", first)
	}
	if !p.File.StopExtinct || p.File.ForagingMortalityModel != "high" || p.File.FoodSourceLimit != 25 {
		t.Errorf("File = %+v, want the example farm parameters", p.File)
	}
}

func TestDeriveSpecies(t *testing.T) {
	constants := params.DefaultConstants()
	species := exampleSpecies(t)
	// B_terrestris: 91,28,[...],6.9,11.1,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,62.4,249.5,1,0,17,590,980,32,0.2
	s := deriveSpecies(species["B_terrestris"], &constants)
	read := []float64{s.EmergingDay_mean, s.EmergingDay_sd, s.Proboscis_min_mm, s.Proboscis_max_mm, s.GrowthFactor, float64(s.SeasonStop),
		float64(s.MaxLifespanWorkers), float64(s.Batchsize), s.FlightVelocity_m_per_s, s.SearchLength_m, s.TimeUnloading, s.SpecMax_cropVolume_myl,
		s.SpecMax_pollenPellets_g, s.MinToMaxFactor, s.DevAgeHatchingMin_d, s.DevAgePupationMin_d, s.DevAgeEmergingMin_d,
		s.DevWeightEgg_mg, s.DevWeightPupationMin_mg, s.DevWeightPupationMax_mg, s.PollenToBodymassFactor,
		s.Dev_Q_DeterminationWeight_mg, s.DevAge_Q_PupationMin_d, s.DevWeight_Q_PupationMin_mg, s.DevWeight_Q_PupationMax_mg,
		s.DevAge_Q_EmergingMin_d, s.ChanceFindNest}
	wantRead := []float64{91, 28, 6.9, 11.1, 1.88, 305, 60, 12, 5, 2500, 165, 173, 0.15, 2, 5, 12.9, 23.8, 1.5, 62.4, 249.5, 1, 0, 17, 590, 980, 32, 0.2}
	if s.Name != "B_terrestris" || !reflect.DeepEqual(read, wantRead) {
		t.Errorf("B_terrestris:\n got  %v %v\n want %v", s.Name, read, wantRead)
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
		"Dev_larvalAge_QueenDetermination_d (constant 3)": {float64(s.Dev_larvalAge_QueenDetermination_d), 3},
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
	for name, want := range map[string]float64{"B_pascuorum": 15, "B_lapidarius": 21, "B_pratorum": 31, "B_hortorum": 8} {
		if got := deriveSpecies(species[name], &constants).EmergingDay_sd; got != want {
			t.Errorf("%s: emergingDay_sd = %v, want %v", name, got, want)
		}
	}

	// step F9: only Psithyrus is parasitic, by its name
	for name, sp := range species {
		if got := deriveSpecies(sp, &constants).Parasitic; got != (name == "Psithyrus") {
			t.Errorf("%s: Parasitic = %v", name, got)
		}
	}
	if _, ok := species["Psithyrus"]; !ok {
		t.Error("the example species have no Psithyrus row")
	}
}

func TestBuildSpeciesOrder(t *testing.T) {
	// finding 20 (see trans/netlogo-findings.md#finding-20): species are created in the order of the run file (the
	// order of the BeespeciesFile), a name missing in the file creates no species, names are case-sensitive.
	c, landscapes := ExampleFarmConfig()
	c.InitialQueens = []globals.SpeciesQueens{{Species: "B_pascuorum", Queens: 1}, {Species: "B_terrestris", Queens: 1},
		{Species: "B_missing", Queens: 1}, {Species: "b_hortorum", Queens: 1}}
	p, err := Build(c, landscapes)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, s := range p.Input.Species {
		names = append(names, s.Name)
	}
	if !reflect.DeepEqual(names, []string{"B_terrestris", "B_pascuorum"}) {
		t.Errorf("species = %q, want B_terrestris, B_pascuorum", names)
	}
	if !reflect.DeepEqual(p.Input.Derived.SpeciesList, []string{"B_pascuorum", "B_terrestris", "B_missing", "b_hortorum"}) {
		t.Errorf("SpeciesList = %q, want the order of initial_queens [QINI-03]", p.Input.Derived.SpeciesList)
	}
}

func TestMortalityForagerUnknownModel(t *testing.T) {
	// Finding 5 (see trans/netlogo-findings.md#finding-5), fixed in step F5: case-sensitive, any other value is an
	// error (NetLogo: 0).
	for _, m := range []string{"High", "medium", ""} {
		if got, err := mortalityForager(m); err == nil {
			t.Errorf("mortalityForager(%q) = %v, want an error", m, got)
		}
	}
	for m, want := range map[string]float64{"high": 1.0e-05, "intermediate": 2.14e-06, "low": 2.75e-07} {
		if got, err := mortalityForager(m); err != nil || got != want {
			t.Errorf("mortalityForager(%q) = %v, %v, want %v", m, got, err, want)
		}
	}
}

// constantYear returns one year with the foraging period v on every day.
func constantYear(v float64) Year {
	y := make(Year, daysPerYear)
	for i := range y {
		y[i] = v
	}
	return y
}

func TestWeather(t *testing.T) {
	// [FPER] the years follow each other and then repeat
	w, err := weather([]Year{constantYear(1), constantYear(2)})
	if err != nil {
		t.Fatal(err)
	}
	for year, want := range map[int]float64{1: 1, 2: 2, 3: 1, 4: 2} {
		if got := w.ForagingPeriod_s(year, 100); got != want {
			t.Errorf("year %d: %v s, want %v s", year, got, want)
		}
	}
	short := constantYear(1)[:364]
	negative, tooLong := constantYear(1), constantYear(1)
	negative[10], tooLong[20] = -1, 86401
	for name, years := range map[string][]Year{"no year": nil, "364 days": {short}, "-1 s": {negative},
		"86401 s": {constantYear(1), tooLong}} {
		if _, err := weather(years); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestReadConfigStrict(t *testing.T) {
	// An unknown key is an error, as an unknown NetLogo parameter. A missing key keeps its zero value.
	cases := map[string]string{
		"unknown section":   "foo: 1\n",
		"unknown parameter": "parameters:\n  Foo: 1\n",
		"GUI parameter":     "parameters:\n  ShowPlots?: true\n",
		"two documents":     "landscape: a.yaml\n---\nlandscape: b.yaml\n",
		"empty":             "",
		// counts and days are whole numbers (step F8)
		"fraction of queens":       "initial_queens:\n  - species: B_terrestris\n    queens: 2.5\n",
		"fraction in batchsize":    "species:\n  - batchsize: 12.5\n",
		"fraction in seasonStop":   "species:\n  - seasonStop: 305.5\n",
		"fraction in a flower day": "flowers:\n  - startDay: 1.5\n",
	}
	for name, text := range cases {
		if _, err := ReadConfig(strings.NewReader(text)); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	c, err := ReadConfig(strings.NewReader("parameters:\n  N_Badgers: 2\n  MergeHedges?: true\n"))
	if err != nil || c.Parameters != (params.ParameterFile{N_Badgers: 2, MergeHedges: true}) {
		t.Errorf("got %+v, %v", c.Parameters, err)
	}
}

func TestFlowerEntryYAML(t *testing.T) {
	// A flower entry is written as in NetLogo, [name value] or [name] (Margin and Plot have no value).
	l := Landscape{MetresPerPatch: 10, Patches: []Patch{{Id: 3, PatchType: "Scrub", X: 1.5, Y: 2, Area_sqm: 300,
		FlowerSpeciesList: []FlowerEntry{{Name: "Bramble", Value: 0.331, HasValue: true}, {Name: "Margin"}}}}}
	var b bytes.Buffer
	if err := WriteLandscape(&b, l); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "flowerSpeciesList: [[Bramble, 0.331], [Margin]]") {
		t.Errorf("written:\n%s", b.String())
	}
	got, err := ReadLandscape(&b)
	if err != nil || !reflect.DeepEqual(got, l) {
		t.Errorf("read back %+v, %v, want %+v", got, err, l)
	}
	for _, text := range []string{"[]", "[a, 1, 2]", "a"} {
		if _, err := ReadLandscape(strings.NewReader("patches:\n  - flowerSpeciesList: [" + text + "]\n")); err == nil {
			t.Errorf("flower entry %s: want an error", text)
		}
	}
}

func TestConfigRoundTrip(t *testing.T) {
	// Writing and reading the run file of the example farm keeps every value.
	c, _ := ExampleFarmConfig()
	var b bytes.Buffer
	if err := WriteConfig(&b, c); err != nil {
		t.Fatal(err)
	}
	got, err := ReadConfig(&b)
	if err != nil || !reflect.DeepEqual(got, c) {
		t.Errorf("round trip changed the run file: %v", err)
	}
}

func TestBuildCropRotation(t *testing.T) {
	// [CROP-01] every landscape file of the crop rotation must be read before the setup.
	c, landscapes := ExampleFarmConfig()
	c.CropRotation = []string{c.Landscape, "other.yaml"}
	if _, err := Build(c, landscapes); err == nil {
		t.Error("want an error for a landscape file of the crop rotation that was not read")
	}
	landscapes["other.yaml"] = Landscape{MetresPerPatch: 5}
	p, err := Build(c, landscapes)
	if err != nil || len(p.Input.Foodsources) != 2 || p.Input.Foodsources["other.yaml"].MetresPerPatch != 5 {
		t.Errorf("want both landscape files, got %d, %v", len(p.Input.Foodsources), err)
	}
	// one file: no crop rotation [MORN-09], the file is not needed
	c.CropRotation = []string{"missing.yaml"}
	if p, err := Build(c, landscapes); err != nil || len(p.Input.Foodsources) != 1 {
		t.Errorf("a crop rotation with one file must not be read: %v", err)
	}
}

func TestBuildNoFlowers(t *testing.T) {
	// [FSRC-12] without flower species (NetLogo: "No Input File") there is no table.
	c, landscapes := ExampleFarmConfig()
	c.Flowers = nil
	p, err := Build(c, landscapes)
	if err != nil || p.Input.Flowers != nil {
		t.Errorf("Flowers = %v, %v, want nil without error", p.Input.Flowers, err)
	}
}
