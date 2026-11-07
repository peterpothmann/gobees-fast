package params

// Constants are the values that NetLogo hard-codes: the values set in ParametersProc and the local values of the
// procedures. NetLogo's parameter files cannot change them. In gobees-fast they are the section constants of the run
// file, which must hold every key (package input). DefaultConstants holds the NetLogo values, which the converter
// writes into the run files.
//
// The tag doc is the comment that the run file shows at the key: unit and source. "no source" means that NetLogo
// gives none, "heuristic" and "arbitrary" are NetLogo's words.
//
// Coefficients of fitted equations stay in the code (handling time, competition point, winter survival), as one of
// them alone cannot be changed without breaking the fit.
type Constants struct {
	Activities    ActivityConstants     `yaml:"activities"`
	Stimuli       StimulusConstants     `yaml:"stimuli"`
	Foraging      ForagingConstants     `yaml:"foraging"`
	Colony        ColonyConstants       `yaml:"colony"`
	Bees          BeeConstants          `yaml:"bees"`
	Psithyrus     PsithyrusConstants    `yaml:"psithyrus"`
	Badgers       BadgerConstants       `yaml:"badgers"`
	Landscape     LandscapeConstants    `yaml:"landscape"`
	InitialQueens InitialQueenConstants `yaml:"initialQueens"`
	Thresholds    Thresholds            `yaml:"thresholds"`
}

// ActivityConstants are the times of the working day and the durations of the tasks.
type ActivityConstants struct {
	GetUpTime        float64 `yaml:"getUpTime" doc:"[s] no source"`                                        // beginning of the working day [s]. NetLogo: GetUpTime_s
	GetUpSpread      float64 `yaml:"getUpSpread" doc:"[s] no source"`                                      // a bee wakes up at GetUpTime + random [0, GetUpSpread) [s]. NetLogo: randomTimeToGetUp_s
	Sunrise          float64 `yaml:"sunrise" doc:"[s] no source"`                                          // start of the foraging period [s]. NetLogo: Sunrise_s
	CallItADay       float64 `yaml:"callItADay" doc:"[s] no source"`                                       // end of the working day [s]. NetLogo: CallItaDay_s
	Pause            float64 `yaml:"pause" doc:"[s] no source"`                                            // rest of a bee without a task [s]. NetLogo: break_s
	EggLayingTime    float64 `yaml:"eggLayingTime" doc:"[s] no source"`                                    // time of a queen for one batch of eggs [s]
	NursingTime      float64 `yaml:"nursingTime" doc:"[s] derivation in the NetLogo comment"`              // time of one incubation [s]. NetLogo: heatingPeriod_s
	IncubationEnergy float64 `yaml:"incubationEnergy" doc:"[kJ/(mg s)] derivation in the NetLogo comment"` // heat per mg of incubating bee and second [kJ/(mg s)]. NetLogo: incubationEnergy_kJ_per_mg_s
}

// StimulusConstants are the values of the nectar and pollen stimuli.
type StimulusConstants struct {
	NectarStoreSize   float64 `yaml:"nectarStoreSize" doc:"[d] no source"` // days of energy need in the ideal nectar store [d]. NetLogo: storeSize_d
	MinNectarStore    float64 `yaml:"minNectarStore" doc:"[kJ] no source"` // the ideal nectar store holds at least this [kJ]. NetLogo: minNectarStore_kJ
	IdealEnergyFactor float64 `yaml:"idealEnergyFactor" doc:"heuristic"`   // factor on the energy need of the ideal store
	NectarThreshold   float64 `yaml:"nectarThreshold" doc:"heuristic"`     // relative deficit of the nectar store above which the stimulus is 1. NetLogo: nectarStimTH
	PollenStoreSize   float64 `yaml:"pollenStoreSize" doc:"[d] no source"` // days of pollen need in the ideal pollen store [d]. NetLogo: storeSize_d
	PollenThreshold   float64 `yaml:"pollenThreshold" doc:"heuristic"`     // relative deficit of the pollen store above which the stimulus is 1. NetLogo: pollenStimTH
}

// ForagingConstants are the values of foraging.
type ForagingConstants struct {
	MaxForagingRange     float64           `yaml:"maxForagingRange" doc:"[m] Knight et al. 2005"`                      // foraging range [m]. NetLogo: ForagingRangeMax_m
	EnergyFactorOnFlower float64           `yaml:"energyFactorOnFlower" doc:"Kacelnik et al. 1986 (honeybees)"`        // energy use on a flower relative to flight
	SucroseEnergy        float64           `yaml:"sucroseEnergy" doc:"[kJ/µmol] BEEHAVE"`                              // energy of sucrose [kJ/µmol]. NetLogo: EnergySucrose_kJ/mymol
	FlightMetabolicRate  float64           `yaml:"flightMetabolicRate" doc:"[W/kg] Wolf et al. 1999"`                  // metabolic rate in flight [W/kg]. NetLogo: MetabolicRateFlight_W/kg
	MaxHandlingTime      float64           `yaml:"maxHandlingTime" doc:"[s] Ings et al. 2006, Stelzer et al. 2010"`    // handling time of an empty or hardly accessible foodsource [s]. NetLogo: maxHandlingTime_s
	PollenLoadTime       float64           `yaml:"pollenLoadTime" doc:"[s] Raine & Chittka 2007"`                      // time on flowers for one pollen load [s]. NetLogo: timeInFlowers_s
	FlowersPerPollenLoad float64           `yaml:"flowersPerPollenLoad" doc:"Raine & Chittka 2007"`                    // flowers visited for one pollen load. NetLogo: flowersNeededForPollenLoad
	PreferCloseProb      float64           `yaml:"preferCloseProb" doc:"heuristic"`                                    // probability that a new patch is chosen among the closer ones. NetLogo: preferenceClosePatchesProb
	SearchProb           float64           `yaml:"searchProb" doc:"no source"`                                         // base probability to search a new patch. NetLogo: searchProbBase
	CropFactor           float64           `yaml:"cropFactor" doc:"[µl/mg] Ings et al. 2006"`                          // crop volume per mg of bee [µl/mg]. NetLogo: beeWeightToLoadFactor
	PelletFactor         float64           `yaml:"pelletFactor" doc:"[g/µl] BEEHAVE: 0.015 g pollen per 50 µl nectar"` // pollen load per µl of crop volume [g/µl]. NetLogo: beehaveCropToPelletFactor_ul-to-g
	Mortality            ForagingMortality `yaml:"mortality"`                                                          // foraging mortality of the three models of the parameter foragingMortalityModel
}

// ForagingMortality is the mortality of a forager per second of a trip [1/s] for each model of the parameter
// foragingMortalityModel. NetLogo: MortalityForager_per_s
type ForagingMortality struct {
	High         float64 `yaml:"high" doc:"[1/s] Visscher & Dukas 1997 (BEEHAVE)"`   // [1/s]
	Intermediate float64 `yaml:"intermediate" doc:"[1/s] Schmid-Hempel & Heeb 1991"` // [1/s]
	Low          float64 `yaml:"low" doc:"[1/s] Stelzer et al. 2010"`                // [1/s]
}

// ColonyConstants are the values of the colony cycle and the brood care.
type ColonyConstants struct {
	PollenAssimilationEnergy   float64 `yaml:"pollenAssimilationEnergy" doc:"[kJ/g] Hrassnig & Crailsheim 2005"` // energy for the assimilation of pollen [kJ/g]. NetLogo: EnergyRequiredForPollenAssimilation_kJ_per_g
	LarvaWorkerRatioLimit      float64 `yaml:"larvaWorkerRatioLimit" doc:"Duchateau & Velthuis 1988"`            // larva:worker ratio below which the switch point and queen eggs can come. NetLogo: LarvaWorkerRatioTH
	DailySwitchProbability     float64 `yaml:"dailySwitchProbability" doc:"[1/d] Duchateau & Velthuis 1988"`     // daily probability of the switch point [1/d]
	QueenEggsBeforeSwitchPoint int     `yaml:"queenEggsBeforeSwitchPoint" doc:"[d] Duchateau & Velthuis 1988"`   // first queen-destined egg before the switch point [d]. NetLogo: QueenDestinedEggsBeforeSP_d
	EggTime                    int     `yaml:"eggTime" doc:"[d] NetLogo: better sex ratio than 6"`               // egg phase in the queen production date [d]. NetLogo: averageCumulTimeEgg_d
	LatestCompetitionPoint     int     `yaml:"latestCompetitionPoint" doc:"[d] no source"`                       // competition point at the latest this long after the eusocial phase [d]. NetLogo: latestCPafter_d
	DeathAge                   int     `yaml:"deathAge" doc:"[d] no source"`                                     // after the competition point without brood, workers and queens older than this die [d]
	StartNectar                float64 `yaml:"startNectar" doc:"[µl] NetLogo comment: about one crop"`           // nectar store of a new colony [µl]
	StartConcentration         float64 `yaml:"startConcentration" doc:"[mol/l] NetLogo comment"`                 // sugar concentration of the start nectar [mol/l]
	DailyIncubation            float64 `yaml:"dailyIncubation" doc:"[kJ/d] Silvola 1984"`                        // incubation energy of a queen per day [kJ/d]
	IncubationBatches          float64 `yaml:"incubationBatches" doc:"Silvola 1984"`                             // batches that a queen incubates with DailyIncubation
	MinPollenStoreBatches      float64 `yaml:"minPollenStoreBatches" doc:"no source"`                            // minimum pollen store in batches of pupae
	QueenDeterminationAge      int     `yaml:"queenDeterminationAge" doc:"[d] Roeseler 1970"`                    // larval age of the caste determination [d]. NetLogo: dev_larvalAge_QueenDetermination_d
}

// BeeConstants are the values of single bees.
type BeeConstants struct {
	AdultBackgroundMortality float64 `yaml:"adultBackgroundMortality" doc:"[1/d] Plowright & Jay 1968"` // daily background mortality of adults [1/d]. NetLogo: MortalityAdultsBackground_daily
	MaxLifespanMales         int     `yaml:"maxLifespanMales" doc:"[d] Duchateau & Marien 1995"`        // adult life span of males [d]
	ForeignAlleles           float64 `yaml:"foreignAlleles" doc:"Duchateau et al. 1994"`                // number of sex alleles in the population. NetLogo: N_ForeignAlleles
	NestSearchTime           float64 `yaml:"nestSearchTime" doc:"[s] no source"`                        // daily nest search time of a queen [s]. NetLogo: NestSearchTime_h (in h)
}

// PsithyrusConstants are the probabilities of a cuckoo bee queen searching a host colony.
type PsithyrusConstants struct {
	FindNestProb  float64 `yaml:"findNestProb" doc:"arbitrary"`  // to find one given nest. NetLogo: findSingleNestProb
	AccessProb    float64 `yaml:"accessProb" doc:"arbitrary"`    // to get access to the nest found. NetLogo: getAccessProb
	KilledProb    float64 `yaml:"killedProb" doc:"arbitrary"`    // to be killed by the host. NetLogo: getKilledProb
	KillQueenProb float64 `yaml:"killQueenProb" doc:"arbitrary"` // to kill a host queen
}

// BadgerConstants are the values of the badgers.
type BadgerConstants struct {
	Range         float64 `yaml:"range" doc:"[m] Kruuk & Parish 1982"`             // colonies closer than this can be dug up [m]. NetLogo: foragingRange_m
	EncounterProb float64 `yaml:"encounterProb" doc:"[1/d] Kowalczyk et al. 2006"` // daily probability to come across a nest [1/d]
	DigUpProb     float64 `yaml:"digUpProb" doc:"arbitrary"`                       // probability to dig up a nest come across
	MinDistance   float64 `yaml:"minDistance" doc:"[m] no source"`                 // minimum distance between badger setts [m]. NetLogo: 300 in CreateBadgersProc
}

// LandscapeConstants are the values of the setup of the landscape.
type LandscapeConstants struct {
	MinMergeArea float64 `yaml:"minMergeArea" doc:"[m2] no source"` // a hedge is merged into a patch of at least this area [m2]. NetLogo: areaTH_sqm
	TripNectar   float64 `yaml:"tripNectar" doc:"[µl] no source"`   // crop volume of a queen, unit of the minimum nectar of a layer [µl]
	TripPollen   float64 `yaml:"tripPollen" doc:"[g] no source"`    // pollen pellets of a queen, unit of the minimum pollen of a layer [g]
}

// InitialQueenConstants are the values of the queens of the setup.
type InitialQueenConstants struct {
	AdultAge    int     `yaml:"adultAge" doc:"[d] NetLogo: exact age doesn't matter"` // [d]
	BroodAge    int     `yaml:"broodAge" doc:"[d] NetLogo: exact age doesn't matter"` // [d]
	AlleleRange float64 `yaml:"alleleRange" doc:"NetLogo colours"`                    // alleles and mtDNA are random in [0, AlleleRange)
}

// Thresholds are the task thresholds of the situations in which a task is possible. In all other situations the
// threshold is NotSetHigh, i.e. never. NetLogo: ThresholdLevelREP
type Thresholds struct {
	EggLaying      EggLayingThresholds `yaml:"eggLaying"`
	PollenForaging ForagingThresholds  `yaml:"pollenForaging"`
	NectarForaging ForagingThresholds  `yaml:"nectarForaging"`
	Nursing        NursingThresholds   `yaml:"nursing"`
}

// EggLayingThresholds are the egg-laying thresholds.
type EggLayingThresholds struct {
	QueenInitiationPhase float64 `yaml:"queenInitiationPhase" doc:"no source"`
	QueenSocialPhase     float64 `yaml:"queenSocialPhase" doc:"no source"`
	Psithyrus            float64 `yaml:"psithyrus" doc:"no source"`
}

// ForagingThresholds are the thresholds of pollen or nectar foraging.
type ForagingThresholds struct {
	Worker               float64 `yaml:"worker" doc:"no source"`
	QueenInitiationPhase float64 `yaml:"queenInitiationPhase" doc:"no source"`
}

// NursingThresholds are the nursing thresholds.
type NursingThresholds struct {
	Worker               float64 `yaml:"worker" doc:"no source"`
	QueenInitiationPhase float64 `yaml:"queenInitiationPhase" doc:"no source"`
	QueenSocialPhase     float64 `yaml:"queenSocialPhase" doc:"no source"`
}

// DefaultConstants returns the NetLogo values.
func DefaultConstants() Constants {
	return Constants{
		Activities: ActivityConstants{
			GetUpTime:        1,
			GetUpSpread:      1800,
			Sunrise:          8 * 3600,
			CallItADay:       24 * 3600,
			Pause:            0.5 * 3600,
			EggLayingTime:    24 * 3600,
			NursingTime:      2880,
			IncubationEnergy: 0.000000217013888,
		},
		Stimuli: StimulusConstants{
			NectarStoreSize:   5,
			MinNectarStore:    20,
			IdealEnergyFactor: 6,
			NectarThreshold:   0.005,
			PollenStoreSize:   5,
			PollenThreshold:   0.005,
		},
		Foraging: ForagingConstants{
			MaxForagingRange:     758,
			EnergyFactorOnFlower: 0.3,
			SucroseEnergy:        0.00582,
			FlightMetabolicRate:  488.6,
			MaxHandlingTime:      60 * 60,
			PollenLoadTime:       257.4,
			FlowersPerPollenLoad: 58,
			PreferCloseProb:      0.9,
			SearchProb:           0.1,
			CropFactor:           0.402,
			PelletFactor:         0.015 / 50,
			Mortality: ForagingMortality{
				High:         1.0e-05,
				Intermediate: 2.14e-06,
				Low:          2.75e-07,
			},
		},
		Colony: ColonyConstants{
			PollenAssimilationEnergy:   6.2,
			LarvaWorkerRatioLimit:      3,
			DailySwitchProbability:     0.13,
			QueenEggsBeforeSwitchPoint: 5,
			EggTime:                    7,
			LatestCompetitionPoint:     45,
			DeathAge:                   10,
			StartNectar:                100,
			StartConcentration:         1.5,
			DailyIncubation:            10,
			IncubationBatches:          1.5,
			MinPollenStoreBatches:      0.5,
			QueenDeterminationAge:      3,
		},
		Bees: BeeConstants{
			AdultBackgroundMortality: 0,
			MaxLifespanMales:         30,
			ForeignAlleles:           24,
			NestSearchTime:           6 * 3600,
		},
		Psithyrus: PsithyrusConstants{
			FindNestProb:  0.05,
			AccessProb:    0.25,
			KilledProb:    0.25,
			KillQueenProb: 0.5,
		},
		Badgers: BadgerConstants{
			Range:         735,
			EncounterProb: 0.19,
			DigUpProb:     0.1,
			MinDistance:   300,
		},
		Landscape: LandscapeConstants{
			MinMergeArea: 6500,
			TripNectar:   180,
			TripPollen:   0.05,
		},
		InitialQueens: InitialQueenConstants{
			AdultAge:    180,
			BroodAge:    36,
			AlleleRange: 139.9,
		},
		Thresholds: Thresholds{
			EggLaying:      EggLayingThresholds{QueenInitiationPhase: 0.1, QueenSocialPhase: 0, Psithyrus: 0.2},
			PollenForaging: ForagingThresholds{Worker: 0.9, QueenInitiationPhase: 0.7},
			NectarForaging: ForagingThresholds{Worker: 0.9, QueenInitiationPhase: 0.7},
			Nursing:        NursingThresholds{Worker: 0.9, QueenInitiationPhase: 0.5, QueenSocialPhase: 0.9},
		},
	}
}

// NotSetHigh and NotSetLow are the sentinel values for values that are not yet set.
const (
	NotSetHigh = 999999999999999.0
	NotSetLow  = -1 * NotSetHigh
)
