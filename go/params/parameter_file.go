package params

// ParameterFile contains the model parameters of a run file (section "parameters", see package input) [RPAR].
//
// The keys are the names of the NetLogo parameter file (tag `yaml`), so that the converter tools/netlogo2gobees can
// map the columns by name. The NetLogo columns with file names and lists (BeeSpeciesInitialQueensListAsString,
// BeespeciesFile, CropRotationListAsString, FlowerspeciesFile, HabitatsFile, Weather) have their own sections in the
// run file since step F7. The columns that only the NetLogo GUI reads and the parameters without effect
// (CumulVisitsOnlyLastYear?, MaxForagingRange_m, MyFarmlandPatches, see tools/netlogo2gobees) are not in the run file.
type ParameterFile struct {
	AbundanceBoost                     float64 `yaml:"AbundanceBoost"`                     // Factor on the flower abundance of all foodsources.
	FoodSourceLimit                    float64 `yaml:"FoodSourceLimit"`                    // Minimum size of a foodsource in queen trips (MinSizeFoodSources?).
	ForagingMortalityFactor            float64 `yaml:"ForagingMortalityFactor"`            // Factor on the foraging mortality.
	ForagingMortalityModel             string  `yaml:"ForagingMortalityModel"`             // "high", "intermediate" or "low" [PAR-04].
	InitialQueensModifierBehaviorSpace float64 `yaml:"InitialQueensModifierBehaviorSpace"` // Factor on the initial number of queens.
	KeepDeadColonies                   bool    `yaml:"KeepDeadColonies?"`                  // Keep dead colonies until 1 January.
	Lambda_detectProb                  float64 `yaml:"Lambda_detectProb"`                  // Decay of the detection probability with distance.
	MapAreaIncluded                    string  `yaml:"MapAreaIncluded"`                    // Part of the map that is simulated, e.g. "complete".
	MaxHibernatingQueens               int     `yaml:"MaxHibernatingQueens"`               // Maximum number of hibernating queens on 1 January.
	MergeHedges                        bool    `yaml:"MergeHedges?"`                       // Merge small hedge patches.
	MinSizeFoodSources                 bool    `yaml:"MinSizeFoodSources?"`                // Set small nectar or pollen amounts to 0.
	N_Badgers                          float64 `yaml:"N_Badgers"`                          // Number of badgers.
	RemoveEmptyFoodSources             bool    `yaml:"RemoveEmptyFoodSources?"`            // Remove foodsources without nectar and pollen.
	SexLocus                           bool    `yaml:"SexLocus?"`                          // Simulate the sex locus (diploid males).
	StopExtinct                        bool    `yaml:"StopExtinct?"`                       // Stop the run when the population is extinct [GO-20].
	UnlimitedMales                     bool    `yaml:"UnlimitedMales?"`                    // Queens always find a mate.
	WinterMortality                    bool    `yaml:"WinterMortality?"`                   // Apply winter mortality to hibernating queens.
}
