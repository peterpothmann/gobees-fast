package params

// SpeciesQueens is one entry of Derived.InitialQueens, e.g. ["B_terrestris" 500] (NetLogo:
// BeeSpeciesInitialQueensList).
type SpeciesQueens struct {
	Species string `yaml:"species"`
	Queens  int    `yaml:"queens"`
}

// Derived holds the globals that ParametersProc [PAR] derives from the parameters. Package input computes them before
// the setup, Params.Apply adds them as resource. Read-only during the run. The landscape of the current year, which
// CropRotationProc changes [CROP-01], is globals.Landscape.File. Until step F14 the type was
// params.Derived.
type Derived struct {
	InitialQueens      []SpeciesQueens // [QINI-02]. NetLogo: BeeSpeciesInitialQueensList
	SpeciesNames       []string        // [QINI-03] species names, same order. NetLogo: SpeciesList
	ForagerMortality   float64         // mortality of a forager per second of a trip [1/s] [PAR-04]. NetLogo: MortalityForager_per_s
	StartLandscape     string          // [PAR-05] landscape file of the start, as named in the run file. NetLogo: FoodsourcesFile
	StewardshipOptions []string        // [PAR-06] stewardship option names. NetLogo: CSS_OptionsList
	CropRotation       []string        // [PAR-07] landscape files, empty or one file: no crop rotation. NetLogo: CropRotationList
}

// ForagingPeriod holds the daily foraging period [FPER] of the run file: the period per day of the year, for one or more
// years. Since step F7 the NetLogo options "Constant N hrs" are 365 equal values. Params.Apply adds it as resource
// (until step F14 params.Weather, added by the system Parameters; named Weather until step F15c, NetLogo: the
// parameter Weather).
type ForagingPeriod struct {
	Years [][]float64 // foraging period [year-1][day-1] [s]
}

// Get returns the foraging period [s] of the day of the year day (1 to 365) in the year year of the run
// (1, 2, ...). The years repeat.
func (w *ForagingPeriod) Get(year, day int) float64 {
	return w.Years[(year-1)%len(w.Years)][day-1]
}
