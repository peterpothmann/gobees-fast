package params

import (
	"gobees/day"
)

// SpeciesFile are the values of one bee species in a run file (section species, package input), before the
// derived values. The keys (tag `yaml`) are the Go names in lowerCamelCase, the tag `netlogo` holds the column of the
// NetLogo BeespeciesFile where it differs (the converter reads the columns by it).
type SpeciesFile struct {
	// Name is the name in InitialQueens. NetLogo: the columns species and name, equal in all files.
	Name string `yaml:"name"`
	// Parasitic marks a cuckoo bumblebee species (Psithyrus). NetLogo compares the name with "Psithyrus", the
	// converter sets the key for this name.
	Parasitic                bool       `yaml:"parasitic,omitempty"`
	EmergenceDayMean         float64    `yaml:"emergenceDayMean" netlogo:"emergingDay_mean"` // mean day of the year of the emergence from hibernation [d]
	EmergenceDaySD           float64    `yaml:"emergenceDaySD" netlogo:"emergingDay_sd"`     // its standard deviation [d], rounded
	NestHabitats             []string   `yaml:"nestHabitats" netlogo:"nestHabitatsList"`     // habitat types of the nest sites, before the expansion, see globals.NestSites
	ProboscisMin             float64    `yaml:"proboscisMin" netlogo:"proboscis_min_mm"`     // [mm]
	ProboscisMax             float64    `yaml:"proboscisMax" netlogo:"proboscis_max_mm"`     // [mm]
	GrowthFactor             float64    `yaml:"growthFactor"`
	SeasonStop               day.OfYear `yaml:"seasonStop"`                                         // day of the year
	MaxLifespanWorkers       int        `yaml:"maxLifespanWorkers"`                                 // [d]
	BatchSize                int        `yaml:"batchSize" netlogo:"batchsize"`                      // eggs per batch
	FlightVelocity           float64    `yaml:"flightVelocity" netlogo:"flightVelocity_m/s"`        // [m/s]
	SearchLength             float64    `yaml:"searchLength" netlogo:"searchLength_m"`              // flight distance of a search trip [m]
	UnloadingTime            float64    `yaml:"unloadingTime" netlogo:"timeUnloading"`              // [s]
	MaxCropVolume            float64    `yaml:"maxCropVolume" netlogo:"specMax_cropVolume_myl"`     // crop volume of the largest bee [µl]
	MaxPollenPellets         float64    `yaml:"maxPollenPellets" netlogo:"specMax_pollenPellets_g"` // pollen load of the largest bee [g]
	MinToMaxFactor           float64    `yaml:"minToMaxFactor"`
	HatchingAgeMin           float64    `yaml:"hatchingAgeMin" netlogo:"devAgeHatchingMin_d"`        // [d]
	PupationAgeMin           float64    `yaml:"pupationAgeMin" netlogo:"devAgePupationMin_d"`        // [d]
	EmergenceAgeMin          float64    `yaml:"emergenceAgeMin" netlogo:"devAgeEmergingMin_d"`       // [d]
	EggWeight                float64    `yaml:"eggWeight" netlogo:"devWeightEgg_mg"`                 // [mg]
	PupationWeightMin        float64    `yaml:"pupationWeightMin" netlogo:"devWeightPupationMin_mg"` // [mg]
	PupationWeightMax        float64    `yaml:"pupationWeightMax" netlogo:"devWeightPupationMax_mg"` // [mg]
	PollenToBodymassFactor   float64    `yaml:"pollenToBodymassFactor"`
	QueenDeterminationWeight float64    `yaml:"queenDeterminationWeight" netlogo:"dev_Q_DeterminationWeight_mg"` // [mg]
	QueenPupationAgeMin      float64    `yaml:"queenPupationAgeMin" netlogo:"devAge_Q_PupationMin_d"`            // [d]
	QueenPupationWeightMin   float64    `yaml:"queenPupationWeightMin" netlogo:"devWeight_Q_PupationMin_mg"`     // [mg]
	QueenPupationWeightMax   float64    `yaml:"queenPupationWeightMax" netlogo:"devWeight_Q_PupationMax_mg"`     // [mg]
	QueenEmergenceAgeMin     float64    `yaml:"queenEmergenceAgeMin" netlogo:"devAge_Q_EmergingMin_d"`           // [d]
	DailyNestSiteChance      float64    `yaml:"dailyNestSiteChance"`
}

// Species holds the parameters of one bee species: the values of the run file (SpeciesFile) and the values derived
// from them in package input. The species lie in Input.Species, bees and colonies hold their index. Read-only
// during the run. The values that depend on the landscape are in globals.NestSites.
//
// NetLogo: species-own.
type Species struct {
	SpeciesFile // EmergenceDaySD is rounded

	FlightCosts              float64 // [kJ/(m mg)], derived from Constants.FlightMetabolicRate. NetLogo: flightCosts_kJ/m/mg
	DailyIncubationQuota     float64 // [kJ], this and the following: derived. NetLogo: devQuotaIncubationToday_kJ
	HatchingAgeMax           float64 // [d]. NetLogo: devAgeHatchingMax_d
	PupationAgeMax           float64 // [d]. NetLogo: devAgePupationMax_d
	EmergenceAgeMax          float64 // [d]. NetLogo: devAgeEmergingMax_d
	HatchingIncubation       float64 // incubation threshold of the hatching [kJ]. NetLogo: devIncubationHatchingTH_kJ
	PupationIncubation       float64 // incubation threshold of the pupation [kJ]. NetLogo: devIncubationPupationTH_kJ
	EmergenceIncubation      float64 // incubation threshold of the emergence [kJ]. NetLogo: devIncubationEmergingTH_kJ
	QueenDeterminationAge    int     // larval age of the caste determination [d]. NetLogo: dev_larvalAge_QueenDetermination_d
	QueenPupationAgeMax      float64 // [d]. NetLogo: devAge_Q_PupationMax_d
	QueenEmergenceAgeMax     float64 // [d]. NetLogo: devAge_Q_EmergingMax_d
	QueenPupationIncubation  float64 // [kJ]. NetLogo: devIncubation_Q_PupationTH_kJ
	QueenEmergenceIncubation float64 // [kJ]. NetLogo: devIncubation_Q_EmergingTH_kJ
	MinPollenStore           float64 // [g]. NetLogo: minPollenStore_g
}
