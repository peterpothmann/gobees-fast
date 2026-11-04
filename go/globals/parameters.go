package globals

// SpeciesQueens is one entry of BeeSpeciesInitialQueensList, e.g. ["B_terrestris" 500].
type SpeciesQueens struct {
	Species string
	Queens  float64
}

// Derived holds the globals that ParametersProc [PAR] derives from the parameters.
// Derived get their values in Parameters.Initialize() and are read-only afterwards, except FoodsourcesFile, which
// CropRotationProc changes [CROP-01].
type Derived struct {
	BeeSpeciesInitialQueensList []SpeciesQueens // [QINI-02]
	SpeciesList                 []string        // [QINI-03] species names, same order
	MortalityForager_per_s      float64         // [PAR-04]
	FoodsourcesFile             string          // [PAR-05] "X.png" -> "X_Foodsources.txt"
	CSS_OptionsList             []string        // [PAR-06] stewardship option names from the HabitatsFile
	CropRotationList            []string        // [PAR-07] [""] by default, i.e. no crop rotation
}
