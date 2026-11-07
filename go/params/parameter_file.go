package params

// Parameters contains the model parameters of a run file (section "parameters", see package input) [RPAR].
//
// The keys (tag `yaml`) are the Go names in lowerCamelCase (step F15c). The tag `netlogo` holds the column of the
// NetLogo parameter file, so that the converter tools/netlogo2gobees can map the columns by name. The NetLogo columns with file names and lists (BeeSpeciesInitialQueensListAsString,
// BeespeciesFile, CropRotationListAsString, FlowerspeciesFile, HabitatsFile, Weather) have their own sections in the
// run file since step F7. The columns that only the NetLogo GUI reads and the parameters without effect
// (CumulVisitsOnlyLastYear?, MaxForagingRange_m, MyFarmlandPatches, see tools/netlogo2gobees) are not in the run file.
// Step F13: MapAreaIncluded is removed (only "complete"), KeepDeadColonies? is an output option (Output).
type Parameters struct {
	AbundanceBoost          float64 `yaml:"abundanceBoost" netlogo:"AbundanceBoost"`                          // Factor on the flower abundance of all foodsources.
	FoodsourceLimit         float64 `yaml:"foodsourceLimit" netlogo:"FoodSourceLimit"`                        // Minimum size of a foodsource in queen trips (LimitFoodsources).
	ForagingMortalityFactor float64 `yaml:"foragingMortalityFactor" netlogo:"ForagingMortalityFactor"`        // Factor on the foraging mortality.
	ForagingMortalityModel  string  `yaml:"foragingMortalityModel" netlogo:"ForagingMortalityModel"`          // "high", "intermediate" or "low" [PAR-04].
	InitialQueensFactor     float64 `yaml:"initialQueensFactor" netlogo:"InitialQueensModifierBehaviorSpace"` // Factor on the initial number of queens.
	DetectionLambda         float64 `yaml:"detectionLambda" netlogo:"Lambda_detectProb"`                      // Decay of the detection probability with distance [1/m].
	MaxHibernatingQueens    int     `yaml:"maxHibernatingQueens" netlogo:"MaxHibernatingQueens"`              // Maximum number of hibernating queens on 1 January.
	MergeHedges             bool    `yaml:"mergeHedges" netlogo:"MergeHedges?"`                               // Merge small hedge patches.
	LimitFoodsources        bool    `yaml:"limitFoodsources" netlogo:"MinSizeFoodSources?"`                   // Set small nectar or pollen amounts to 0 (FoodsourceLimit).
	Badgers                 int     `yaml:"badgers" netlogo:"N_Badgers"`                                      // Number of badgers.
	RemoveEmptyFoodsources  bool    `yaml:"removeEmptyFoodsources" netlogo:"RemoveEmptyFoodSources?"`         // Remove foodsources without nectar and pollen.
	SexLocus                bool    `yaml:"sexLocus" netlogo:"SexLocus?"`                                     // Simulate the sex locus (diploid males).
	StopExtinct             bool    `yaml:"stopExtinct" netlogo:"StopExtinct?"`                               // Stop the run when the population is extinct [GO-20].
	UnlimitedMales          bool    `yaml:"unlimitedMales" netlogo:"UnlimitedMales?"`                         // Queens always find a mate.
	WinterMortality         bool    `yaml:"winterMortality" netlogo:"WinterMortality?"`                       // Apply winter mortality to hibernating queens.
}

// Output contains the options of the output (section "output" of a run file). They change no model value (step F13).
type Output struct {
	// KeepDeadColonies counts the dead colonies of all years in the output column DeadColonies, not only those since
	// the last 1 January [UCOL-12]. NetLogo: the parameter KeepDeadColonies?, which keeps the dead colonies as agents.
	KeepDeadColonies bool `yaml:"keepDeadColonies"`
}
