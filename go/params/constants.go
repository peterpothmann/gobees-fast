package params

// Constants are the values hard-coded in ParametersProc [PAR-03].
//
// In NetLogo they cannot be changed via files. Here they are parameters with the NetLogo values as defaults.
// Display-only values (symbol sizes, colours) are not translated.
// The foraging mortality is not a constant, it depends on ForagingMortalityModel [PAR-04].
type Constants struct {
	CallItaDay_s                                 float64 // End of the working day [s].
	DailySwitchProbability                       float64 // Daily probability of the switch point (Duchateau & Velthuis 1988).
	EnergyFactorOnFlower                         float64 // Energy factor while on a flower (Kacelnik et al. 1986).
	EnergyRequiredForPollenAssimilation_kJ_per_g float64 // [kJ/g] (Hrassnig & Crailsheim 2005).
	EnergySucrose_kJ_per_mymol                   float64 // [kJ/micromol] (from BEEHAVE).
	ForagingRangeMax_m                           float64 // Foraging range [m] (Knight et al. 2005). NetLogo never reads MaxForagingRange_m of the parameter file.
	GetUpTime_s                                  float64 // Beginning of the working day [s].
	LarvaWorkerRatioTH                           float64 // Threshold larva:worker ratio (Duchateau & Velthuis 1988).
	MaxLifespanMales                             float64 // [d] adult age (Duchateau & Marien 1995).
	MetabolicRateFlight_W_per_kg                 float64 // [W/kg] (Wolf et al. 1999).
	MortalityAdultsBackground_daily              float64 // Daily background mortality of adults (Plowright & Jay 1968).
	N_ForeignAlleles                             float64 // Number of sex alleles in the population (Duchateau et al. 1994).
	NestSearchTime_h                             float64 // Daily nest search time of queens [h].
	QueenDestinedEggsBeforeSP_d                  float64 // First queen-destined egg before the switch point [d] (Duchateau & Velthuis 1988).
	Sunrise_s                                    float64 // Start of the foraging period [s].
}

// DefaultConstants returns the values from ParametersProc [PAR-03].
func DefaultConstants() Constants {
	return Constants{
		CallItaDay_s:           24 * 3600,
		DailySwitchProbability: 0.13,
		EnergyFactorOnFlower:   0.3,
		EnergyRequiredForPollenAssimilation_kJ_per_g: 6.2,
		EnergySucrose_kJ_per_mymol:                   0.00582,
		ForagingRangeMax_m:                           758,
		GetUpTime_s:                                  1,
		LarvaWorkerRatioTH:                           3,
		MaxLifespanMales:                             30,
		MetabolicRateFlight_W_per_kg:                 488.6,
		MortalityAdultsBackground_daily:              0,
		N_ForeignAlleles:                             24,
		NestSearchTime_h:                             6,
		QueenDestinedEggsBeforeSP_d:                  5,
		Sunrise_s:                                    8 * 3600,
	}
}

// NotSetHigh and NotSetLow are the sentinel values of ParametersProc [PAR-03]
// for dates and other values that are not yet set.
const (
	NotSetHigh = 999999999999999.0
	NotSetLow  = -1 * NotSetHigh
)
