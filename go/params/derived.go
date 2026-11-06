package params

// SpeciesQueens is one entry of Derived.InitialQueens, e.g. ["B_terrestris" 500] (NetLogo:
// BeeSpeciesInitialQueensList).
type SpeciesQueens struct {
	Species string `yaml:"species"`
	Queens  int    `yaml:"queens"`
}

// Derived holds the values derived from the run file. Package input computes them before the setup, Params.Apply
// adds them as resource. Read-only during the run. The landscape of the current year, which the crop rotation
// changes, is globals.Landscape.File.
type Derived struct {
	InitialQueens      []SpeciesQueens // NetLogo: BeeSpeciesInitialQueensList
	SpeciesNames       []string        // species names, same order. NetLogo: SpeciesList
	ForagerMortality   float64         // mortality of a forager per second of a trip [1/s]. NetLogo: MortalityForager_per_s
	StartLandscape     string          // landscape file of the start, as named in the run file. NetLogo: FoodsourcesFile
	StewardshipOptions []string        // stewardship option names. NetLogo: CSS_OptionsList
	CropRotation       []string        // landscape files, empty or one file: no crop rotation. NetLogo: CropRotationList
}

// ForagingPeriod holds the daily foraging period of the run file: the period per day of the year, for one or more
// years. Params.Apply adds it as resource. NetLogo: the parameter Weather.
type ForagingPeriod struct {
	Years [][]float64 // foraging period [year-1][day-1] [s]
}

// Get returns the foraging period [s] of the day of the year day (1 to 365) in the year year of the run
// (1, 2, ...). The years repeat.
func (w *ForagingPeriod) Get(year, day int) float64 {
	return w.Years[(year-1)%len(w.Years)][day-1]
}
