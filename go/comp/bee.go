package comp

import (
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// A bee entity has the components Position, Bee, Genetics, Thresholds, Forager and Development.
// It is one bee or a cohort of Number bees.
// NetLogo: bees-own.

// Bee holds the state of a bee or cohort.
type Bee struct {
	Number        int               // bees in this agent: 1, or the cohort size
	Stage         stage.Stage       // egg, larva, pupa, adult. NetLogo brood?: Stage != stage.Adult
	Caste         caste.Caste       // undefined, queen, worker, male
	Activity      activity.Activity // current task
	Colony        ecs.Entity        // the colony, zero entity for none. NetLogo: colonyID
	Species       int               // index in params.Input.Species. NetLogo: speciesID, speciesName
	Mated         bool              // mated?
	AdultAge      int               // [d]
	BroodAge      int               // [d]
	EmergenceDate day.Date          // day of emergence from hibernation, unset for brood. NetLogo: emergingDate
	PersonalTime  float64           // the bee's own time of day, used by the daily activity queue [s]. NetLogo: personalTime_s
	Weight        float64           // [mg]. NetLogo: weight_mg
}

// Genetics holds the alleles. Allele values are random numbers below 139.9 (in NetLogo a colour).
type Genetics struct {
	MtDNA       float64
	Alleles     []float64 // 1 allele (haploid male) or 2 (female, diploid male). NetLogo: allelesList, ploidy
	Spermatheca []float64 // allele(s) of the male the queen mated with. NetLogo: spermathecaList
}

// Thresholds are the task thresholds of a bee. NetLogo: thEgglaying, thForagingNectar, thForagingPollen,
// thNursing.
type Thresholds struct {
	Egglaying      float64
	NectarForaging float64
	PollenForaging float64
	Nursing        float64
}

// Forager holds the foraging state and capacities of a bee.
type Forager struct {
	CropVolume    float64 // crop capacity [µl]. NetLogo: cropvolume_myl
	PollenPellets float64 // pollen load capacity [g]. NetLogo: pollenPellets_g
	GlossaLength  float64 // [mm]. NetLogo: glossaLength_mm
	// IngestionFactor is the term 0.3 * W^0.3333 * G of the nectar ingestion time (forager.handlingTime), from the
	// weight W [g] and GlossaLength G [mm] of the adult bee, which do not change. 0: not computed yet, handlingTime
	// computes it at its first use. Whoever changes Weight or GlossaLength sets it to 0. Not in NetLogo.
	IngestionFactor    float64
	CurrentFoodsource  ecs.Entity
	NectarSource       ecs.Entity // the nectar layer of the next trip. NetLogo: nectarsourceToGoTo
	PollenSource       ecs.Entity // the pollen layer of the next trip. NetLogo: pollensourceToGoTo
	PollenForager      bool
	KnownNectarPatches []ecs.Entity // known masterpatches with nectar, sorted by distance. NetLogo: knownMasterpatchesNectarList
	KnownPollenPatches []ecs.Entity // known masterpatches with pollen, sorted by distance. NetLogo: knownMasterpatchesPollenList
	CarriedNectar      float64      // nectar load of the bee or cohort [kJ]. NetLogo: nectarLoadSquadron_kJ
	CarriedPollen      float64      // pollen load of the bee or cohort [g]. NetLogo: pollenLoadSquadron_g
	ExpectedNectarTrip float64      // expected duration of a nectar trip [s]. NetLogo: expectation_NectarTrip_s
	ExpectedPollenTrip float64      // expected duration of a pollen trip [s]. NetLogo: expectation_PollenTrip_s
}

// Development holds the brood development state of a bee.
type Development struct {
	IncubationReceived float64 // incubation received since the egg was laid [kJ]. NetLogo: cumulIncubationReceived_kJ
	LarvaTime          int     // days spent as a larva [d]. NetLogo: cumulTimeLarva_d
}
