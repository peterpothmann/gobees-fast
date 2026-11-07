package comp

import (
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// A bee entity has the components Position, Bee, Genetics, Thresholds, Forager and Development.
// It is one bee or a cohort of Number bees (trans/data-model.md).
// NetLogo: bees-own. Not translated: activityList (a log for the GUI, [GO-09]).

// Bee holds the state of a bee or cohort.
type Bee struct {
	Number         int               // bees in this agent: 1, or the cohort size
	Stage          stage.Stage       // egg, larva, pupa, adult (decision A). NetLogo brood?: Stage != stage.Adult (step F9)
	Caste          caste.Caste       // undefined, queen, worker, male
	Activity       activity.Activity // current task
	ColonyID       ecs.Entity        // the colony, zero entity for none (NetLogo -1)
	SpeciesID      int               // index in params.Input.Species (step F11). NetLogo speciesName: the Name of the species (step F9)
	Mated          bool              // mated?
	AdultAge       int               // [d]
	BroodAge       int               // [d]
	EmergingDate   day.Date          // day of emergence from hibernation, unset for brood (NetLogo: NotSetLow)
	PersonalTime_s float64           // the bee's own time of day, used by the daily activity queue
	Weight_mg      float64
}

// Genetics holds the alleles. Allele values are random-float 139.9 (a NetLogo colour, also used for display).
type Genetics struct {
	MtDNA           float64
	AllelesList     []float64 // 1 allele (haploid male) or 2 (female, diploid male). NetLogo ploidy: len(AllelesList) (step F9)
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
	CumulTimeLarva_d           int
}
