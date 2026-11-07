package input

import (
	"bytes"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"

	"gobees/params"
)

// exampleSpecies returns the species of the example farm's run file by name (all rows of the BeespeciesFile).
func exampleSpecies(t *testing.T) map[string]Species {
	t.Helper()
	c, _ := ExampleFarmConfig()
	m := map[string]Species{}
	for _, s := range c.Species {
		m[s.Name] = s
	}
	return m
}

func TestExampleFarm(t *testing.T) {
	p := ExampleFarm()
	in := p.Input
	want := params.Derived{
		InitialQueens:      []params.SpeciesQueens{{Species: "B_terrestris", Queens: 500}},
		SpeciesNames:       []string{"B_terrestris"},
		ForagerMortality:   1.0e-05,
		StartLandscape:     "_SYSTEM_Example_Farm_Foodsources.yaml",
		StewardshipOptions: []string{"Legume", "Margin", "newCSO_Legume", "BirdBumble_Margin", "GameCover_Plot", "Plot"},
	}
	if !reflect.DeepEqual(in.Derived, want) {
		t.Errorf("Derived:\n got  %+v\n want %+v", in.Derived, want)
	}
	// "Constant 8 hrs": 365 times 28800 s
	if len(in.ForagingPeriod.Years) != 1 || len(in.ForagingPeriod.Years[0]) != 365 || slices.ContainsFunc(in.ForagingPeriod.Years[0],
		func(v float64) bool { return v != 28800 }) {
		t.Errorf("ForagingPeriod: want one year of 28800 s")
	}
	if len(in.Species) != 1 || in.Species[0].Name != "B_terrestris" {
		t.Errorf("Species: want B_terrestris only, got %d", len(in.Species))
	}
	if in.Flowers == nil || len(in.Flowers.Rows) != 79 || len(in.Flowers.Names) != 79 {
		t.Errorf("Flowers: want the 79 flower species of _SYSTEM_Flowerspecies.csv")
	}
	fs, ok := in.Foodsources[in.Derived.StartLandscape]
	if len(in.Foodsources) != 1 || !ok || fs.MetresPerPatch != 10 || len(fs.Patches) != 73 {
		t.Errorf("Foodsources: want only the example farm with 10 m per patch and 73 flower patches")
	}
	// first line of _SYSTEM_Example_Farm_Foodsources.txt: 0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275] ...] 20
	first := fs.Patches[0]
	if first.ID != 0 || first.PatchType != "Hedgerow" || first.X != 130 || first.Y != 175 || first.Area != 200 ||
		len(first.Flowers) != 33 ||
		first.Flowers[0] != (params.FlowerEntry{Name: "Average_Willow", Value: 1.275, HasValue: true}) {
		t.Errorf("first flower patch = %+v", first)
	}
	if !p.Parameters.StopExtinct || p.Parameters.ForagingMortalityModel != "high" || p.Parameters.FoodsourceLimit != 25 {
		t.Errorf("File = %+v, want the example farm parameters", p.Parameters)
	}
}

func TestDeriveSpecies(t *testing.T) {
	constants := params.DefaultConstants()
	species := exampleSpecies(t)
	// B_terrestris: 91,28,[...],6.9,11.1,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,62.4,249.5,1,0,17,590,980,32,0.2
	s := deriveSpecies(species["B_terrestris"], &constants)
	read := []float64{s.EmergenceDayMean, s.EmergenceDaySD, s.ProboscisMin, s.ProboscisMax, s.GrowthFactor, float64(s.SeasonStop),
		float64(s.MaxLifespanWorkers), float64(s.BatchSize), s.FlightVelocity, s.SearchLength, s.UnloadingTime, s.MaxCropVolume,
		s.MaxPollenPellets, s.MinToMaxFactor, s.HatchingAgeMin, s.PupationAgeMin, s.EmergenceAgeMin,
		s.EggWeight, s.PupationWeightMin, s.PupationWeightMax, s.PollenToBodymassFactor,
		s.QueenDeterminationWeight, s.QueenPupationAgeMin, s.QueenPupationWeightMin, s.QueenPupationWeightMax,
		s.QueenEmergenceAgeMin, s.DailyNestSiteChance}
	wantRead := []float64{91, 28, 6.9, 11.1, 1.88, 305, 60, 12, 5, 2500, 165, 173, 0.15, 2, 5, 12.9, 23.8, 1.5, 62.4, 249.5, 1, 0, 17, 590, 980, 32, 0.2}
	if s.Name != "B_terrestris" || !reflect.DeepEqual(read, wantRead) {
		t.Errorf("B_terrestris:\n got  %v %v\n want %v", s.Name, read, wantRead)
	}
	if !reflect.DeepEqual(s.NestHabitats, []string{"Grassland*", "Garden", "Hedgerow", "Scrub", "Woodland", "Nestboxes"}) {
		t.Errorf("nestHabitats = %q (not yet expanded)", s.NestHabitats)
	}

	// derived values, computed by hand
	derived := map[string][2]float64{
		"FlightCosts (488.6 / 5 / 1e9)":                {s.FlightCosts, 9.772e-8},
		"DailyIncubationQuota (10 / 18)":               {s.DailyIncubationQuota, 0.5555555555555556},
		"HatchingAgeMax (5 * 2)":                       {s.HatchingAgeMax, 10},
		"PupationAgeMax (12.9 * 2 - 5)":                {s.PupationAgeMax, 20.8},
		"EmergenceAgeMax (23.8 * 2 - 12.9)":            {s.EmergenceAgeMax, 34.7},
		"HatchingIncubation (10/18 * 5)":               {s.HatchingIncubation, 2.7777777777777777},
		"PupationIncubation (10/18 * 12.9)":            {s.PupationIncubation, 7.166666666666667},
		"EmergenceIncubation (10/18 * 23.8)":           {s.EmergenceIncubation, 13.222222222222221},
		"QueenDeterminationAge (constant 3)":           {float64(s.QueenDeterminationAge), 3},
		"QueenPupationAgeMax (17 * 2 - 5)":             {s.QueenPupationAgeMax, 29},
		"QueenEmergenceAgeMax (32 * 2 - 17)":           {s.QueenEmergenceAgeMax, 47},
		"QueenPupationIncubation (10/18 * 17)":         {s.QueenPupationIncubation, 9.444444444444445},
		"QueenEmergenceIncubation (10/18 * 32)":        {s.QueenEmergenceIncubation, 17.77777777777778},
		"MinPollenStore (0.5 * 0.001 * 62.4 * 12 / 1)": {s.MinPollenStore, 0.3744},
	}
	for name, v := range derived {
		if math.Abs(v[0]-v[1]) > 1e-9*math.Max(1, math.Abs(v[1])) {
			t.Errorf("%s = %v, want %v", name, v[0], v[1])
		}
	}

	// emergenceDaySD is rounded, .5 up
	for name, want := range map[string]float64{"B_pascuorum": 15, "B_lapidarius": 21, "B_pratorum": 31, "B_hortorum": 8} {
		if got := deriveSpecies(species[name], &constants).EmergenceDaySD; got != want {
			t.Errorf("%s: emergingDay_sd = %v, want %v", name, got, want)
		}
	}

	// only the Psithyrus row of the run file has the key parasitic, and the species keeps it
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
	// species are created in the order of the run file, a name missing in the file creates no species, names are case-sensitive.
	c, landscapes := ExampleFarmConfig()
	c.InitialQueens = []params.SpeciesQueens{{Species: "B_pascuorum", Queens: 1}, {Species: "B_terrestris", Queens: 1},
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
	if !reflect.DeepEqual(p.Input.Derived.SpeciesNames, []string{"B_pascuorum", "B_terrestris", "B_missing", "b_hortorum"}) {
		t.Errorf("SpeciesNames = %q, want the order of initialQueens", p.Input.Derived.SpeciesNames)
	}
}

// defaults are the NetLogo values of the constants.
var defaults = params.DefaultConstants()

func TestMortalityForagerUnknownModel(t *testing.T) {
	// case-sensitive, any other value is an error
	for _, m := range []string{"High", "medium", ""} {
		if got, err := mortalityForager(m, &defaults.Foraging.Mortality); err == nil {
			t.Errorf("mortalityForager(%q) = %v, want an error", m, got)
		}
	}
	for m, want := range map[string]float64{"high": 1.0e-05, "intermediate": 2.14e-06, "low": 2.75e-07} {
		if got, err := mortalityForager(m, &defaults.Foraging.Mortality); err != nil || got != want {
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

func TestForagingPeriod(t *testing.T) {
	// the years follow each other and then repeat
	w, err := weather([]Year{constantYear(1), constantYear(2)})
	if err != nil {
		t.Fatal(err)
	}
	for year, want := range map[int]float64{1: 1, 2: 2, 3: 1, 4: 2} {
		if got := w.Get(year, 100); got != want {
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
	// An unknown key is an error. A missing key keeps its zero value, except in the section constants.
	cases := map[string]string{
		"unknown section":   "foo: 1\n",
		"unknown parameter": "parameters:\n  Foo: 1\n",
		"GUI parameter":     "parameters:\n  ShowPlots?: true\n",
		"two documents":     "landscape: a.yaml\n---\nlandscape: b.yaml\n",
		"empty":             "",
		// counts and days are whole numbers
		"fraction of queens":       "initialQueens:\n  - species: B_terrestris\n    queens: 2.5\n",
		"fraction in batchSize":    "species:\n  - batchSize: 12.5\n",
		"fraction in seasonStop":   "species:\n  - seasonStop: 305.5\n",
		"fraction in a flower day": "flowers:\n  - startDay: 1.5\n",
	}
	for name, text := range cases {
		if _, err := ReadConfig(strings.NewReader(text)); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	c, err := ReadConfig(strings.NewReader(withConstants(t, "parameters:\n  badgers: 2\n  mergeHedges: true\n")))
	if err != nil || c.Parameters != (params.Parameters{Badgers: 2, MergeHedges: true}) {
		t.Errorf("got %+v, %v", c.Parameters, err)
	}
}

// withConstants appends the section constants with the NetLogo values to the run file text.
func withConstants(t *testing.T, text string) string {
	t.Helper()
	var b bytes.Buffer
	section := struct {
		Constants params.Constants `yaml:"constants"`
	}{defaults}
	if err := encode(&b, section); err != nil {
		t.Fatal(err)
	}
	return text + b.String()
}

func TestConstantsSection(t *testing.T) {
	// the section must hold every key
	if _, err := ReadConfig(strings.NewReader("parameters:\n  badgers: 2\n")); err == nil ||
		!strings.Contains(err.Error(), "constants: missing") {
		t.Errorf("no section constants: error %v", err)
	}
	full := withConstants(t, "")
	c, err := ReadConfig(strings.NewReader(full))
	if err != nil || c.Constants != defaults {
		t.Errorf("all keys: got %+v, %v, want the NetLogo values", c.Constants, err)
	}
	for _, key := range []string{"digUpProb", "high", "psithyrus", "queenSocialPhase"} {
		lines := []string{}
		for _, line := range strings.Split(full, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), key+":") {
				continue // drops the first and only key of this name below its parent
			}
			lines = append(lines, line)
		}
		text := strings.Join(lines, "\n")
		if key == "psithyrus" { // the section without its values
			text = strings.NewReplacer("    findNestProb: 0.05\n", "", "    accessProb: 0.25\n", "",
				"    killedProb: 0.25\n", "", "    killQueenProb: 0.5\n", "").Replace(text)
		}
		if _, err := ReadConfig(strings.NewReader(text)); err == nil || !strings.Contains(err.Error(), key+": missing") {
			t.Errorf("without %s: error %v", key, err)
		}
	}
	if _, err := ReadConfig(strings.NewReader(full + "  foo: 1\n")); err == nil {
		t.Error("unknown key in constants: want an error")
	}
}

func TestConstantsComments(t *testing.T) {
	// the written run file has the comment above the section and unit and source at each key, with the notes
	c, _ := ExampleFarmConfig()
	var b bytes.Buffer
	if err := WriteConfigNotes(&b, c, map[string]string{"foraging.maxForagingRange": "a note"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# Fixed in the NetLogo code, not settable in NetLogo's parameter files. Defaults: the NetLogo values.\n",
		"    maxForagingRange: 758 # [m] Knight et al. 2005. a note\n",
		"    digUpProb: 0.1 # arbitrary\n",
		"    nestSearchTime: 21600 # [s] no source\n",
		"      queenSocialPhase: 0 # no source\n",
	} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("written run file lacks %q", want)
		}
	}
	// every value of the section has a comment
	var check func(t reflect.Type, path string)
	check = func(t2 reflect.Type, path string) {
		for i := range t2.NumField() {
			f := t2.Field(i)
			if f.Type.Kind() == reflect.Struct {
				check(f.Type, path+"."+f.Name)
			} else if f.Tag.Get("doc") == "" {
				t.Errorf("%s.%s has no tag doc", path, f.Name)
			}
		}
	}
	check(reflect.TypeFor[params.Constants](), "Constants")
}

func TestFlowerEntryYAML(t *testing.T) {
	// A flower entry is written [name value] or [name] (Margin and Plot have no value).
	l := Landscape{MetresPerPatch: 10, Patches: []Patch{{ID: 3, PatchType: "Scrub", X: 1.5, Y: 2, Area: 300,
		Flowers: []FlowerEntry{{Name: "Bramble", Value: 0.331, HasValue: true}, {Name: "Margin"}}}}}
	var b bytes.Buffer
	if err := WriteLandscape(&b, l); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "flowers: [[Bramble, 0.331], [Margin]]") {
		t.Errorf("written:\n%s", b.String())
	}
	got, err := ReadLandscape(&b)
	if err != nil || !reflect.DeepEqual(got, l) {
		t.Errorf("read back %+v, %v, want %+v", got, err, l)
	}
	for _, text := range []string{"[]", "[a, 1, 2]", "a"} {
		if _, err := ReadLandscape(strings.NewReader("patches:\n  - flowers: [" + text + "]\n")); err == nil {
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
	// every landscape file of the crop rotation must be read before the setup.
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
	// one file: no crop rotation, the file is not needed
	c.CropRotation = []string{"missing.yaml"}
	if p, err := Build(c, landscapes); err != nil || len(p.Input.Foodsources) != 1 {
		t.Errorf("a crop rotation with one file must not be read: %v", err)
	}
}

func TestBuildNoFlowers(t *testing.T) {
	// without flower species there is no table.
	c, landscapes := ExampleFarmConfig()
	c.Flowers = nil
	p, err := Build(c, landscapes)
	if err != nil || p.Input.Flowers != nil {
		t.Errorf("Flowers = %v, %v, want nil without error", p.Input.Flowers, err)
	}
}

// TestReadRelativePath: a run file given by a relative path with folders is read with its landscape files (the
// command line gives an absolute path, the user interface the path as typed).
func TestReadRelativePath(t *testing.T) {
	t.Chdir("..")
	f, err := Read("compare/scenarios/step7_crop.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Landscapes) < 2 {
		t.Errorf("%d landscape files, want the start and the crop rotation", len(f.Landscapes))
	}
}
