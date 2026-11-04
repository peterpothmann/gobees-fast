package globals

// SpeciesQueens is one entry of BeeSpeciesInitialQueensList, e.g. ["B_terrestris" 500].
type SpeciesQueens struct {
	Species string
	Queens  float64
}

// Derived holds the globals that ParametersProc [PAR] derives from the parameters.
// Package input computes them before the setup (params.Input), sys.Parameters adds a copy as resource. They are
// read-only afterwards, except FoodsourcesFile, which CropRotationProc changes [CROP-01].
type Derived struct {
	BeeSpeciesInitialQueensList []SpeciesQueens // [QINI-02]
	SpeciesList                 []string        // [QINI-03] species names, same order
	MortalityForager_per_s      float64         // [PAR-04]
	FoodsourcesFile             string          // [PAR-05] the example farm at the start (input.StartFoodsourcesFile)
	CSS_OptionsList             []string        // [PAR-06] stewardship option names from the HabitatsFile
	CropRotationList            []string        // [PAR-07] [""] by default, i.e. no crop rotation
}
