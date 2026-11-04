package globals

// SpeciesQueens is one entry of BeeSpeciesInitialQueensList, e.g. ["B_terrestris" 500].
type SpeciesQueens struct {
	Species string  `yaml:"species"`
	Queens  float64 `yaml:"queens"`
}

// Derived holds the globals that ParametersProc [PAR] derives from the parameters.
// Package input computes them before the setup (params.Input), sys.Parameters adds a copy as resource. They are
// read-only afterwards, except FoodsourcesFile, which CropRotationProc changes [CROP-01].
type Derived struct {
	BeeSpeciesInitialQueensList []SpeciesQueens // [QINI-02]
	SpeciesList                 []string        // [QINI-03] species names, same order
	MortalityForager_per_s      float64         // [PAR-04]
	FoodsourcesFile             string          // [PAR-05] landscape file of the start, as named in the run file
	CSS_OptionsList             []string        // [PAR-06] stewardship option names
	CropRotationList            []string        // [PAR-07] landscape files, empty or one file: no crop rotation
}
