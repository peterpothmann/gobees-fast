package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// UpdateColonies updates every colony in the morning [UCOL]: range lists, deaths of whole colonies, statistics,
// eusocial phase, switch point, competition point, death of an empty colony.
//
// Part of the morning routine [MORN-10]. Initialize runs it at ticks = 0 [INIT-10] (no colonies exist then).
// Not translated: heading and label (GUI), the patch counters nMalesProduced, nQueensProduced, nColonies
// (never read).
//
// NetLogo: UpdateColoniesProc, CompetitionPointDateREP.
type UpdateColonies struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	calendar  *globals.Calendar
	today     *globals.FoodsourcesToday
	file      *params.ParameterFile
	constants *params.Constants
	counters  *globals.Counters
	dead      globals.DeadColonies
	beeDeath  beeDeath
	colDeath  colonyDeath
	colonies  colonyFounder

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map3[comp.Colony, comp.ColonyStores, comp.ColonyCounts]
	beeMap       *ecs.Map1[comp.Bee]
	speciesMap   *ecs.Map1[comp.Species]
}

func (s *UpdateColonies) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.today = ecs.GetResource[globals.FoodsourcesToday](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.counters = ecs.GetResource[globals.Counters](w)
	ecs.AddResource(w, &s.dead)
	s.beeDeath = newBeeDeath(w)
	s.colDeath = newColonyDeath(w)
	s.colonies = newColonyFounder(w, s.rng)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap3[comp.Colony, comp.ColonyStores, comp.ColonyCounts](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
	s.Update(w)
}

func (s *UpdateColonies) Update(w *ecs.World) {
	// ORDER of the colonies: irrelevant for the state, see [UCOL-01] to [UCOL-11]. Shuffled (decision C).
	colonies := []ecs.Entity{}
	query := s.colonyFilter.Query()
	for query.Next() {
		colonies = append(colonies, query.Entity())
	}
	util.Shuffle(s.rng, colonies)

	// NetLogo: bees with [ colonyID = whoCol ]. The bees in the same colony are grouped once, as no bee is created and no bee
	// changes its colony here. Bees that die on the way are skipped with w.Alive.
	// makes updateColony run faster
	beesOf := map[ecs.Entity][]ecs.Entity{}
	bees := s.beeFilter.Query()
	for bees.Next() {
		if b := bees.Get(); !b.ColonyID.IsZero() {
			beesOf[b.ColonyID] = append(beesOf[b.ColonyID], bees.Entity())
		}
	}

	for _, col := range colonies {
		s.updateColony(w, col, beesOf[col])
	}

	// [UCOL-12] the record then holds the same colonies as NetLogo's deadCols.
	if !s.file.KeepDeadColonies && s.calendar.Day == 1 {
		s.dead.List = s.dead.List[:0]
	}
}

func (s *UpdateColonies) Finalize(w *ecs.World) {}

// updateColony runs [UCOL-01] to [UCOL-11] for one colony. ORDER within the colony: MATTERS
// (finding 37, see trans/netlogo-findings.md#finding-37).
func (s *UpdateColonies) updateColony(w *ecs.World, col ecs.Entity, members []ecs.Entity) {
	ticks := float64(s.ticks.Ticks)
	// alive returns the living members for which keep is true (NetLogo: bees with [ colonyID = whoCol ... ]).
	alive := func(keep func(b *comp.Bee) bool) []ecs.Entity {
		result := []ecs.Entity{}
		for _, e := range members {
			if w.Alive(e) && keep(s.beeMap.Get(e)) {
				result = append(result, e)
			}
		}
		return result
	}
	// ask is alive in random order, for NetLogo's ask (decision C).
	ask := func(keep func(b *comp.Bee) bool) []ecs.Entity {
		result := alive(keep)
		util.Shuffle(s.rng, result)
		return result
	}
	all := func(b *comp.Bee) bool { return true }

	countBroodMort_NA := 0.0
	countBroodMort_ES := 0.0

	// [UCOL-01]
	if s.today.FoodsourcesInFlowerUpdate {
		s.colonies.FoodsourcesInFlowerAndRangeProc(col)
	}
	colony, stores, counts := s.colonyMap.Get(col)

	// [UCOL-02] after the competition point and without brood, the workers and queens older than 10 days die.
	// Males survive (finding 38, see trans/netlogo-findings.md#finding-38).
	// counts.AllEggs, AllLarvae and AllPupae are set further below in [UCOL-06]. Here they still hold the values
	// of yesterday morning: brood that emerged or died since then still counts as brood. The adults therefore die
	// one day later than with today's numbers (finding 37, see trans/netlogo-findings.md#finding-37).
	// ORDER: irrelevant, each bee dies independently.
	if ticks > colony.CompetitionPointDate && counts.AllEggs+counts.AllLarvae+counts.AllPupae == 0 {
		for _, e := range ask(func(b *comp.Bee) bool {
			return b.AdultAge > 10 && (b.Caste == caste.Worker || b.Caste == caste.Queen)
		}) {
			s.beeDeath.DieProc(w, e, "Colony death after CP!")
		}
	}
	// [UCOL-03]
	if stores.EnergyStore_kJ <= 0 {
		for _, e := range ask(all) {
			if b := s.beeMap.Get(e); b.Brood {
				countBroodMort_ES += b.Number
			}
			s.beeDeath.DieProc(w, e, "Colony's energy store depleted!")
		}
	}
	// [UCOL-04] adult = brood? false, males included (finding 38, see trans/netlogo-findings.md#finding-38)
	adults := 0.0
	for _, e := range alive(func(b *comp.Bee) bool { return !b.Brood }) {
		adults += s.beeMap.Get(e).Number
	}
	if adults == 0 {
		for _, e := range ask(all) {
			if b := s.beeMap.Get(e); b.Brood {
				countBroodMort_NA += b.Number
			}
			s.beeDeath.DieProc(w, e, "No adult bees left!")
		}
	}
	// get the components again: removing entities is a structural change of the world
	colony, stores, counts = s.colonyMap.Get(col)

	// [UCOL-05]
	counts.BroodDeathsNoAdults += countBroodMort_NA
	counts.BroodDeathsEnergyStores += countBroodMort_ES
	stores.SummedIncubationToday_kJ = 0

	// [UCOL-06]
	s.statistics(counts, alive(all))

	// [UCOL-07]
	if colony.EusocialPhaseDate == params.NotSetHigh && counts.AllAdultWorkers > 0 {
		colony.EusocialPhaseDate = ticks
	}
	// [UCOL-08] ORDER: irrelevant, one independent draw per colony.
	if colony.SwitchPointDate == params.NotSetHigh {
		if colony.EusocialPhaseDate < params.NotSetHigh && counts.LarvaWorkerRatio < s.constants.LarvaWorkerRatioTH {
			if s.rng.Float64() <= s.constants.DailySwitchProbability {
				colony.SwitchPointDate = ticks
			}
		}
	}
	// [UCOL-09]
	if colony.CompetitionPointDate == params.NotSetHigh &&
		colony.EusocialPhaseDate < params.NotSetHigh &&
		colony.QueenProductionDate < params.NotSetHigh {
		colony.CompetitionPointDate = CompetitionPointDateREP(colony)
	}

	// [UCOL-10] queenright?: never read, removed in step F3

	// [UCOL-11]
	if len(alive(all)) == 0 {
		if s.calendar.Day >= s.speciesMap.Get(colony.SpeciesIDcolony).SeasonStop {
			s.counters.ColonyDeathsEndSeason++
		} else {
			s.counters.ColonyDeathsNoBees++
		}
		s.colDeath.DieProc(w, col, "Colony: No adults or brood left in this colony!")
	} // otherwise colonyAge + 1: never read, removed in step F3
}

// statistics sets the colony statistics from the living members of the colony [UCOL-06].
func (s *UpdateColonies) statistics(c *comp.ColonyCounts, members []ecs.Entity) {
	c.AllEggs, c.AllLarvae, c.AllPupae, c.AllAdults = 0, 0, 0, 0
	c.AllAdultWorkers, c.AllAdultMales = 0, 0
	for _, e := range members {
		b := s.beeMap.Get(e)
		switch b.Stage {
		case stage.Egg:
			c.AllEggs += b.Number
		case stage.Larva:
			c.AllLarvae += b.Number
		case stage.Pupa:
			c.AllPupae += b.Number
		case stage.Adult:
			c.AllAdults += b.Number
			if b.Caste == caste.Worker {
				c.AllAdultWorkers += b.Number
			}
		}
		if !b.Brood && b.Caste == caste.Male {
			c.AllAdultMales += b.Number
		}
	}
	if c.AllAdultWorkers > 0 {
		c.LarvaWorkerRatio = c.AllLarvae / c.AllAdultWorkers
	} else {
		c.LarvaWorkerRatio = params.NotSetHigh
	}
}

// CompetitionPointDateREP returns the day of the competition point of a colony [CPD].
// After the competition point, workers and queens compete for male production: workers become aggressive, lay their own (male) eggs and eat the eggs of others. The model does not simulate this behaviour, only its consequences: after the competition point eggs no longer develop (MortalityBroodProc) and the colony ends once no brood is left [UCOL-02].
func CompetitionPointDateREP(c *comp.Colony) float64 {
	// [CPD-01] x: days from the first worker to the first queen-destined eggs (about -5 to 40).
	// y: days from the first worker to the CP, linear in x (Duchateau & Velthuis 1988, Fig. 6).
	// The CP is rounded to a whole day (.5 up) and lies at most 45 days after the first worker.
	// NetLogo first sets compDate to NotSetHigh and overwrites it below (removed in step F3).
	x := c.QueenProductionDate - c.EusocialPhaseDate
	// float64(...) forces Go to round after the multiplication, as Java does. Otherwise Go may compute
	// 0.7*x + 15.5 in one step (fused multiply-add), which can change the last bit and thus the rounding.
	y := float64(0.7*x) + 15.5
	latestCPafter_d := 45.0
	compDate := util.Round(c.EusocialPhaseDate + y)
	// The CP is at most 45 days after the eusocial phase
	if compDate-c.EusocialPhaseDate > latestCPafter_d {
		compDate = c.EusocialPhaseDate + latestCPafter_d // set to 45 days after eusocial phase (first worker), as in NetLogo
	}
	return compDate
}
