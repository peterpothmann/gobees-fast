package comp

import (
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// A bee entity has the components Who, Position, Bee, Genetics, Thresholds, Forager and Development.
// It is one bee or a cohort of Number bees (trans/data-model.md).
// NetLogo: bees-own. Not translated: activityList (a log for the GUI, [GO-09]).

// Bee holds the state of a bee or cohort.
type Bee struct {
	Number         float64           // bees in this agent: 1, or the cohort size
	Stage          stage.Stage       // egg, larva, pupa, adult (decision A)
	Brood          bool              // brood? true for egg, larva, pupa
	Caste          caste.Caste       // undefined, queen, worker, male
	Activity       activity.Activity // current task
	ColonyID       ecs.Entity        // the colony, zero entity for none (NetLogo -1)
	SpeciesID      ecs.Entity        // the species entity
	SpeciesName    string            // name of the species
	Mated          bool              // mated?
	Ploidy         float64           // 1: haploid male, 2: diploid female or diploid male
	AdultAge       float64           // [d]
	BroodAge       float64           // [d]
	EmergingDate   float64           // tick of emergence from hibernation (absolute, compared with ticks)
	PersonalTime_s float64           // the bee's own time of day, used by the daily activity queue
	Weight_mg      float64
}

// Genetics holds the alleles. Allele values are random-float 139.9 (a NetLogo colour, also used for display).
type Genetics struct {
	MtDNA           float64
	AllelesList     []float64 // 1 allele (haploid male) or 2 (female, diploid male)
	SpermathecaList []float64 // allele(s) of the male the queen mated with
}

// Thresholds are the task thresholds of a bee [THR].
type Thresholds struct {
	ThEgglaying      float64
	ThForagingNectar float64
	ThForagingPollen float64
	ThNursing        float64
}

// Forager holds the foraging state and capacities of a bee.
type Forager struct {
	Cropvolume_myl               float64 // [LOAD]
	PollenPellets_g              float64 // [LOAD]
	GlossaLength_mm              float64 // [PROB]
	CurrentFoodsource            ecs.Entity
	NectarsourceToGoTo           ecs.Entity
	PollensourceToGoTo           ecs.Entity
	PollenForager                bool
	KnownMasterpatchesNectarList []ecs.Entity
	KnownMasterpatchesPollenList []ecs.Entity
	NectarLoadSquadron_kJ        float64
	PollenLoadSquadron_g         float64
	Expectation_NectarTrip_s     float64
	Expectation_PollenTrip_s     float64
}

// Development holds the brood development state of a bee (translated in step 4).
type Development struct {
	CumulIncubationReceived_kJ float64
	CumulTimeLarva_d           float64
}
