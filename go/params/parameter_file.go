package params

// Parameters contains the model parameters of a run file (section "parameters", see package input).
//
// The keys (tag `yaml`) are the Go names in lowerCamelCase. The tag `netlogo` holds the column of the NetLogo
// parameter file, so that the converter tools/netlogo2gobees can map the columns by name.
type Parameters struct {
	AbundanceBoost          float64 `yaml:"abundanceBoost" netlogo:"AbundanceBoost"`                          // Factor on the flower abundance of all foodsources.
	FoodsourceLimit         float64 `yaml:"foodsourceLimit" netlogo:"FoodSourceLimit"`                        // Minimum size of a foodsource in queen trips (LimitFoodsources).
	ForagingMortalityFactor float64 `yaml:"foragingMortalityFactor" netlogo:"ForagingMortalityFactor"`        // Factor on the foraging mortality.
	ForagingMortalityModel  string  `yaml:"foragingMortalityModel" netlogo:"ForagingMortalityModel"`          // "high", "intermediate" or "low".
	InitialQueensFactor     float64 `yaml:"initialQueensFactor" netlogo:"InitialQueensModifierBehaviorSpace"` // Factor on the initial number of queens.
	DetectionLambda         float64 `yaml:"detectionLambda" netlogo:"Lambda_detectProb"`                      // Decay of the detection probability with distance [1/m].
	MaxHibernatingQueens    int     `yaml:"maxHibernatingQueens" netlogo:"MaxHibernatingQueens"`              // Maximum number of hibernating queens on 1 January.
	MergeHedges             bool    `yaml:"mergeHedges" netlogo:"MergeHedges?"`                               // Merge small hedge patches.
	LimitFoodsources        bool    `yaml:"limitFoodsources" netlogo:"MinSizeFoodSources?"`                   // Set small nectar or pollen amounts to 0 (FoodsourceLimit).
	Badgers                 int     `yaml:"badgers" netlogo:"N_Badgers"`                                      // Number of badgers.
	RemoveEmptyFoodsources  bool    `yaml:"removeEmptyFoodsources" netlogo:"RemoveEmptyFoodSources?"`         // Remove foodsources without nectar and pollen.
	SexLocus                bool    `yaml:"sexLocus" netlogo:"SexLocus?"`                                     // Simulate the sex locus (diploid males).
	StopExtinct             bool    `yaml:"stopExtinct" netlogo:"StopExtinct?"`                               // Stop the run when the population is extinct.
	UnlimitedMales          bool    `yaml:"unlimitedMales" netlogo:"UnlimitedMales?"`                         // Queens always find a mate.
	WinterMortality         bool    `yaml:"winterMortality" netlogo:"WinterMortality?"`                       // Apply winter mortality to hibernating queens.
}

// Output contains the options of the output (section "output" of a run file). They change no model value.
type Output struct {
	// KeepDeadColonies counts the dead colonies of all years in the output column DeadColonies, not only those since
	// the last 1 January. NetLogo: the parameter KeepDeadColonies?, which keeps the dead colonies as agents.
	KeepDeadColonies bool `yaml:"keepDeadColonies"`

	// Columns are the columns of the daily output, in this order (names: gobees --list-columns, obs.AllColumns). The
	// column ticks is always the first. No key: all columns. The flag --columns overrides it. The column t, which
	// the CSV reporter writes in front, is not listed.
	Columns []string `yaml:"columns,omitempty,flow"`

	// Visits writes the visits output: the visits of the foragers per foodsource and day. No key: false. It changes
	// no model value.
	Visits bool `yaml:"visits,omitempty"`
}
