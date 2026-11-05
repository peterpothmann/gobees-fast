package params

import (
	"gobees/day"
)

// SpeciesFile are the values of one bee species in a run file [SPEC-03] (section species, package input), before the
// derived values. The keys are the column names of the NetLogo BeespeciesFile.
type SpeciesFile struct {
	// Name is the name in InitialQueens [SPEC-02], [IQUE-04]. NetLogo has two columns, species (compared with the
	// initial queens) and name (stored), equal in all files. The converter checks that they are equal (step F13).
	Name string `yaml:"name"`
	// Parasitic marks a cuckoo bumblebee species (Psithyrus) [ACT-04], [EMER-04], [PSI]. NetLogo compares the name
	// with "Psithyrus", the converter sets the key for this name (step F13).
	Parasitic                    bool       `yaml:"parasitic,omitempty"`
	EmergingDay_mean             float64    `yaml:"emergingDay_mean"`
	EmergingDay_sd               float64    `yaml:"emergingDay_sd"`
	NestHabitatsList             []string   `yaml:"nestHabitatsList"` // [SPEC-04] before the expansion, see globals.NestSites
	Proboscis_min_mm             float64    `yaml:"proboscis_min_mm"`
	Proboscis_max_mm             float64    `yaml:"proboscis_max_mm"`
	GrowthFactor                 float64    `yaml:"growthFactor"`
	SeasonStop                   day.OfYear `yaml:"seasonStop"`
	MaxLifespanWorkers           int        `yaml:"maxLifespanWorkers"`
	Batchsize                    int        `yaml:"batchsize"`
	FlightVelocity_m_per_s       float64    `yaml:"flightVelocity_m/s"`
	SearchLength_m               float64    `yaml:"searchLength_m"`
	TimeUnloading                float64    `yaml:"timeUnloading"`
	SpecMax_cropVolume_myl       float64    `yaml:"specMax_cropVolume_myl"`
	SpecMax_pollenPellets_g      float64    `yaml:"specMax_pollenPellets_g"`
	MinToMaxFactor               float64    `yaml:"minToMaxFactor"`
	DevAgeHatchingMin_d          float64    `yaml:"devAgeHatchingMin_d"`
	DevAgePupationMin_d          float64    `yaml:"devAgePupationMin_d"`
	DevAgeEmergingMin_d          float64    `yaml:"devAgeEmergingMin_d"`
	DevWeightEgg_mg              float64    `yaml:"devWeightEgg_mg"`
	DevWeightPupationMin_mg      float64    `yaml:"devWeightPupationMin_mg"`
	DevWeightPupationMax_mg      float64    `yaml:"devWeightPupationMax_mg"`
	PollenToBodymassFactor       float64    `yaml:"pollenToBodymassFactor"`
	Dev_Q_DeterminationWeight_mg float64    `yaml:"dev_Q_DeterminationWeight_mg"`
	DevAge_Q_PupationMin_d       float64    `yaml:"devAge_Q_PupationMin_d"`
	DevWeight_Q_PupationMin_mg   float64    `yaml:"devWeight_Q_PupationMin_mg"`
	DevWeight_Q_PupationMax_mg   float64    `yaml:"devWeight_Q_PupationMax_mg"`
	DevAge_Q_EmergingMin_d       float64    `yaml:"devAge_Q_EmergingMin_d"`
	DailyNestSiteChance          float64    `yaml:"dailyNestSiteChance"`
}

// Species holds the parameters of one bee species: the values of the run file (SpeciesFile) and the values derived
// from them in package input [SPEC-03], [SPEC-05], [SPEC-07]. The species lie in Input.Species, bees and colonies
// hold their index (step F11). Read-only during the run. The values that depend on the landscape are in
// globals.NestSites. Until step F14 the type was params.Species.
//
// NetLogo: species-own (oneSpecies agents), created in CreateSpeciesProc [SPEC].
type Species struct {
	SpeciesFile // EmergingDay_sd is rounded [SPEC-03], finding 20 (see trans/netlogo-findings.md#finding-20)

	FlightCosts_kJ_per_m_per_mg        float64 // derived from MetabolicRateFlight_W/kg
	DevQuotaIncubationToday_kJ         float64 // [SPEC-05] and following: derived
	DevAgeHatchingMax_d                float64
	DevAgePupationMax_d                float64
	DevAgeEmergingMax_d                float64
	DevIncubationHatchingTH_kJ         float64
	DevIncubationPupationTH_kJ         float64
	DevIncubationEmergingTH_kJ         float64
	Dev_larvalAge_QueenDetermination_d int
	DevAge_Q_PupationMax_d             float64
	DevAge_Q_EmergingMax_d             float64
	DevIncubation_Q_PupationTH_kJ      float64
	DevIncubation_Q_EmergingTH_kJ      float64
	MinPollenStore_g                   float64 // [SPEC-07]
}
