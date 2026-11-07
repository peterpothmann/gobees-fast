package params

// SpeciesQueens is one entry of BeeSpeciesInitialQueensList, e.g. ["B_terrestris" 500].
type SpeciesQueens struct {
	Species string `yaml:"species"`
	Queens  int    `yaml:"queens"`
}

// Derived holds the globals that ParametersProc [PAR] derives from the parameters. Package input computes them before
// the setup, Params.Apply adds them as resource. Read-only during the run. The landscape of the current year, which
// CropRotationProc changes [CROP-01], is globals.Landscape.FoodsourcesFile. Until step F14 the type was
// params.Derived.
type Derived struct {
	BeeSpeciesInitialQueensList []SpeciesQueens // [QINI-02]
	SpeciesList                 []string        // [QINI-03] species names, same order
	MortalityForager_per_s      float64         // [PAR-04]
	FoodsourcesFile             string          // [PAR-05] landscape file of the start, as named in the run file
	CSS_OptionsList             []string        // [PAR-06] stewardship option names
	CropRotationList            []string        // [PAR-07] landscape files, empty or one file: no crop rotation
}

// Weather holds the daily foraging period [FPER] of the run file: the period per day of the year, for one or more
// years. Since step F7 the NetLogo options "Constant N hrs" are 365 equal values. Params.Apply adds it as resource
// (until step F14 params.Weather, added by the system Parameters).
type Weather struct {
	Years_s [][]float64 // foraging period [year-1][day-1] [s]
}

// ForagingPeriod_s returns the foraging period [s] of the day of the year day (1 to 365) in the year year of the run
// (1, 2, ...). The years repeat.
func (w *Weather) ForagingPeriod_s(year, day int) float64 {
	return w.Years_s[(year-1)%len(w.Years_s)][day-1]
}
