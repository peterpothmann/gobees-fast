package params

// Constants are the values hard-coded in ParametersProc [PAR-03].
//
// In NetLogo they cannot be changed via files. Here they are parameters with the NetLogo values as defaults.
// Display-only values (symbol sizes, colours) are not translated.
// The foraging mortality is not a constant, it depends on ForagingMortalityModel [PAR-04].
type Constants struct {
	CallItADay                 float64 // End of the working day [s]. NetLogo: CallItaDay_s
	DailySwitchProbability     float64 // Daily probability of the switch point (Duchateau & Velthuis 1988).
	EnergyFactorOnFlower       float64 // Energy factor while on a flower (Kacelnik et al. 1986).
	PollenAssimilationEnergy   float64 // Energy for the assimilation of pollen [kJ/g] (Hrassnig & Crailsheim 2005). NetLogo: EnergyRequiredForPollenAssimilation_kJ_per_g
	SucroseEnergy              float64 // Energy of sucrose [kJ/µmol] (from BEEHAVE). NetLogo: EnergySucrose_kJ/mymol
	MaxForagingRange           float64 // Foraging range [m] (Knight et al. 2005). NetLogo: ForagingRangeMax_m. NetLogo never reads MaxForagingRange_m of the parameter file.
	GetUpTime                  float64 // Beginning of the working day [s]. NetLogo: GetUpTime_s
	LarvaWorkerRatioLimit      float64 // Threshold larva:worker ratio (Duchateau & Velthuis 1988). NetLogo: LarvaWorkerRatioTH
	MaxLifespanMales           int     // [d] adult age (Duchateau & Marien 1995).
	FlightMetabolicRate        float64 // Metabolic rate in flight [W/kg] (Wolf et al. 1999). NetLogo: MetabolicRateFlight_W/kg
	AdultBackgroundMortality   float64 // Daily background mortality of adults [1/d] (Plowright & Jay 1968). NetLogo: MortalityAdultsBackground_daily
	ForeignAlleles             float64 // Number of sex alleles in the population (Duchateau et al. 1994). NetLogo: N_ForeignAlleles
	NestSearchTime             float64 // Daily nest search time of queens [h]. NetLogo: NestSearchTime_h
	QueenEggsBeforeSwitchPoint int     // First queen-destined egg before the switch point [d] (Duchateau & Velthuis 1988). NetLogo: QueenDestinedEggsBeforeSP_d
	Sunrise                    float64 // Start of the foraging period [s]. NetLogo: Sunrise_s
}

// DefaultConstants returns the values from ParametersProc [PAR-03].
func DefaultConstants() Constants {
	return Constants{
		CallItADay:                 24 * 3600,
		DailySwitchProbability:     0.13,
		EnergyFactorOnFlower:       0.3,
		PollenAssimilationEnergy:   6.2,
		SucroseEnergy:              0.00582,
		MaxForagingRange:           758,
		GetUpTime:                  1,
		LarvaWorkerRatioLimit:      3,
		MaxLifespanMales:           30,
		FlightMetabolicRate:        488.6,
		AdultBackgroundMortality:   0,
		ForeignAlleles:             24,
		NestSearchTime:             6,
		QueenEggsBeforeSwitchPoint: 5,
		Sunrise:                    8 * 3600,
	}
}

// NotSetHigh and NotSetLow are the sentinel values of ParametersProc [PAR-03]
// for dates and other values that are not yet set.
const (
	NotSetHigh = 999999999999999.0
	NotSetLow  = -1 * NotSetHigh
)
