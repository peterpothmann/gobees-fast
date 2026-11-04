package params

// ParameterFile contains the parameters of the NetLogo parameter file (e.g. _SYSTEM_Parameters.csv) that the model
// reads. Package input fills it [RPAR].
//
// The tag `netlogo` holds the column name in the file. The columns that only the NetLogo GUI reads are not stored
// (input.ignoredParameters).
type ParameterFile struct {
	AbundanceBoost                      float64 `netlogo:"AbundanceBoost"`                      // Factor on the flower abundance of all foodsources.
	BeeSpeciesInitialQueensListAsString string  `netlogo:"BeeSpeciesInitialQueensListAsString"` // Species and initial queens, e.g. "B_terrestris 500" [QINI].
	BeespeciesFile                      string  `netlogo:"BeespeciesFile"`                      // Bee species table.
	CropRotationListAsString            string  `netlogo:"CropRotationListAsString"`            // Maps of the crop rotation, separated by blanks [PAR-07].
	CumulVisitsOnlyLastYear             bool    `netlogo:"CumulVisitsOnlyLastYear?"`            // Reset cumulative visit counters on 1 January.
	FlowerspeciesFile                   string  `netlogo:"FlowerspeciesFile"`                   // Flower species table. The column name in the file has a trailing blank.
	FoodSourceLimit                     float64 `netlogo:"FoodSourceLimit"`                     // Minimum size of a foodsource in queen trips (MinSizeFoodSources?).
	ForagingMortalityFactor             float64 `netlogo:"ForagingMortalityFactor"`             // Factor on the foraging mortality.
	ForagingMortalityModel              string  `netlogo:"ForagingMortalityModel"`              // "high", "intermediate" or "low" [PAR-04].
	HabitatsFile                        string  `netlogo:"HabitatsFile"`                        // Habitat table, provides the stewardship option names [CSO].
	InitialQueensModifierBehaviorSpace  float64 `netlogo:"InitialQueensModifierBehaviorSpace"`  // Factor on the initial number of queens.
	KeepDeadColonies                    bool    `netlogo:"KeepDeadColonies?"`                   // Keep dead colonies until 1 January.
	Lambda_detectProb                   float64 `netlogo:"Lambda_detectProb"`                   // Decay of the detection probability with distance.
	MapAreaIncluded                     string  `netlogo:"MapAreaIncluded"`                     // Part of the map that is simulated, e.g. "complete".
	MaxForagingRange_m                  float64 `netlogo:"MaxForagingRange_m"`                  // NOT USED by the model, see ForagingRangeMax_m in Constants.
	MaxHibernatingQueens                float64 `netlogo:"MaxHibernatingQueens"`                // Maximum number of hibernating queens on 1 January.
	MergeHedges                         bool    `netlogo:"MergeHedges?"`                        // Merge small hedge patches.
	MinSizeFoodSources                  bool    `netlogo:"MinSizeFoodSources?"`                 // Set small nectar or pollen amounts to 0.
	MyFarmlandPatches                   string  `netlogo:"MyFarmlandPatches"`                   // IDs of farmland foodsources, e.g. " 3 7 ".
	N_Badgers                           float64 `netlogo:"N_Badgers"`                           // Number of badgers.
	RemoveEmptyFoodSources              bool    `netlogo:"RemoveEmptyFoodSources?"`             // Remove foodsources without nectar and pollen.
	SexLocus                            bool    `netlogo:"SexLocus?"`                           // Simulate the sex locus (diploid males).
	StopExtinct                         bool    `netlogo:"StopExtinct?"`                        // Stop the run when the population is extinct [GO-20].
	UnlimitedMales                      bool    `netlogo:"UnlimitedMales?"`                     // Queens always find a mate.
	Weather                             string  `netlogo:"Weather"`                             // Daily foraging period, e.g. "Constant 8 hrs" [FPER].
	WinterMortality                     bool    `netlogo:"WinterMortality?"`                    // Apply winter mortality to hibernating queens.
}
