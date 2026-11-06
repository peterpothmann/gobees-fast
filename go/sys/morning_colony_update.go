package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// colonyUpdate updates every colony in the morning: range lists, deaths of whole colonies, statistics,
// eusocial phase, switch point, competition point, death of an empty colony.
//
// Part of the morning routine. Initialize runs it also at ticks = 0 (no colonies exist then).
//
// NetLogo: UpdateColoniesProc.
type colonyUpdate struct {
	rng        *rand.Rand
	ticks      *globals.Ticks
	calendar   *globals.Calendar
	today      *globals.Flowering
	parameters *params.Parameters
	constants  *params.Constants
	beeDeath   beeDeath
	colDeath   colonyDeath
	colonies   colonyFounder

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map3[comp.Colony, comp.ColonyStores, comp.ColonyCounts]
	beeMap       *ecs.Map1[comp.Bee]
	speciesMap   *speciesTable

	beesOf map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *colonyUpdate) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.today = ecs.GetResource[globals.Flowering](w)
	s.parameters = ecs.GetResource[params.Parameters](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.beeDeath = newBeeDeath(w)
	s.colonies = newColonyFounder(w, s.rng)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap3[comp.Colony, comp.ColonyStores, comp.ColonyCounts](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.speciesMap = newSpeciesTable(w)
	s.Update(w) // run at ticks = 0
}

func (s *colonyUpdate) Update(w *ecs.World) {
	// ORDER of the colonies: irrelevant for the state. Shuffled, as NetLogo asks the colonies in random order.
	colonies := []ecs.Entity{}
	query := s.colonyFilter.Query()
	for query.Next() {
		colonies = append(colonies, query.Entity())
	}
	shuffle(s.rng, colonies)

	// The bees are grouped by colony once, which makes updateColony run faster. This is possible as no bee is created
	// and no bee changes its colony here. Bees that die on the way are skipped with w.Alive.
	s.beesOf = beesByColony(s.beesOf, s.beeFilter, func(b *comp.Bee) bool { return true })

	for _, col := range colonies {
		s.updateColony(w, col, s.beesOf[col])
	}
}

func (s *colonyUpdate) Finalize(w *ecs.World) {}

// updateColony updates one colony. ORDER within the colony: MATTERS.
func (s *colonyUpdate) updateColony(w *ecs.World, col ecs.Entity, members []ecs.Entity) {
	ticks := s.ticks.Ticks
	// alive returns the living members for which keep is true.
	alive := func(keep func(b *comp.Bee) bool) []ecs.Entity {
		result := []ecs.Entity{}
		for _, e := range members {
			if w.Alive(e) && keep(s.beeMap.Get(e)) {
				result = append(result, e)
			}
		}
		return result
	}
	// ask is alive in random order, as NetLogo asks the bees in random order.
	ask := func(keep func(b *comp.Bee) bool) []ecs.Entity {
		result := alive(keep)
		shuffle(s.rng, result)
		return result
	}
	all := func(b *comp.Bee) bool { return true }

	if s.today.Changed {
		s.colonies.flowersInRange(col)
	}
	colony, stores, counts := s.colonyMap.Get(col)

	// after the competition point and without brood, the workers and queens older than 10 days die.
	// Males survive.
	// counts.Eggs, Larvae and Pupae are set further below in statistics. Here they still hold the values
	// of yesterday morning: brood that emerged or died since then still counts as brood. The adults therefore die
	// one day later than with today's numbers, as in NetLogo.
	// ORDER: irrelevant, each bee dies independently.
	if colony.CompetitionPointDate.Before(ticks) && counts.Eggs+counts.Larvae+counts.Pupae == 0 {
		for _, e := range ask(func(b *comp.Bee) bool {
			return b.AdultAge > s.constants.Colony.DeathAge && (b.Caste == caste.Worker || b.Caste == caste.Queen)
		}) {
			s.beeDeath.die(w, e, "Colony death after CP!")
		}
	}
	if stores.Energy <= 0 {
		for _, e := range ask(all) {
			s.beeDeath.die(w, e, "Colony's energy store depleted!")
		}
	}
	// adults: all bees in stage adult, males included
	adults := 0
	for _, e := range alive(func(b *comp.Bee) bool { return b.Stage == stage.Adult }) {
		adults += s.beeMap.Get(e).Number
	}
	if adults == 0 {
		for _, e := range ask(all) {
			s.beeDeath.die(w, e, "No adult bees left!")
		}
	}
	// get the components again: removing entities is a structural change of the world
	colony, stores, counts = s.colonyMap.Get(col)

	stores.IncubationToday = 0

	s.statistics(counts, alive(all))

	if !colony.EusocialPhaseDate.IsSet() && counts.AdultWorkers > 0 {
		colony.EusocialPhaseDate = day.At(ticks)
	}
	// ORDER: irrelevant, one independent draw per colony.
	if !colony.SwitchPointDate.IsSet() {
		if colony.EusocialPhaseDate.IsSet() && counts.LarvaWorkerRatio < s.constants.Colony.LarvaWorkerRatioLimit {
			if s.rng.Float64() <= s.constants.Colony.DailySwitchProbability {
				colony.SwitchPointDate = day.At(ticks)
			}
		}
	}
	if !colony.CompetitionPointDate.IsSet() &&
		colony.EusocialPhaseDate.IsSet() &&
		colony.QueenProductionDate.IsSet() {
		colony.CompetitionPointDate = day.At(competitionPointDate(colony, &s.constants.Colony))
	}

	if len(alive(all)) == 0 {
		s.colDeath.die(w, col, "Colony: No adults or brood left in this colony!")
	}
}

// statistics sets the colony statistics from the living members of the colony.
func (s *colonyUpdate) statistics(c *comp.ColonyCounts, members []ecs.Entity) {
	c.Eggs, c.Larvae, c.Pupae, c.Adults, c.AdultWorkers = 0, 0, 0, 0, 0
	for _, e := range members {
		b := s.beeMap.Get(e)
		switch b.Stage {
		case stage.Egg:
			c.Eggs += b.Number
		case stage.Larva:
			c.Larvae += b.Number
		case stage.Pupa:
			c.Pupae += b.Number
		case stage.Adult:
			c.Adults += b.Number
			if b.Caste == caste.Worker {
				c.AdultWorkers += b.Number
			}
		}
	}
	if c.AdultWorkers > 0 {
		c.LarvaWorkerRatio = float64(c.Larvae) / float64(c.AdultWorkers)
	} else {
		c.LarvaWorkerRatio = params.NotSetHigh
	}
}

// competitionPointDate returns the day of the competition point of a colony.
// After the competition point, workers and queens compete for male production: workers become aggressive, lay their
// own (male) eggs and eat the eggs of others. The model does not simulate this behaviour, only its consequences: after
// the competition point eggs no longer develop (broodMortality) and the colony ends once no brood is left.
// Both dates must be set.
//
// NetLogo: CompetitionPointDateREP.
func competitionPointDate(c *comp.Colony, cc *params.ColonyConstants) day.Abs {
	// x: days from the first worker to the first queen-destined eggs (about -5 to 40).
	// y: days from the first worker to the CP, linear in x (Duchateau & Velthuis 1988, Fig. 6).
	// The CP is rounded to a whole day (.5 up) and lies at most LatestCompetitionPoint days after the first worker.
	// The coefficients 0.7 and 15.5 of the regression stay in the code.
	eusocialPhaseDate := c.EusocialPhaseDate.Day()
	x := float64(c.QueenProductionDate.Day() - eusocialPhaseDate)
	// float64(...) forces Go to round after the multiplication, as Java does. Otherwise Go may compute
	// 0.7*x + 15.5 in one step (fused multiply-add), which can change the last bit and thus the rounding.
	y := float64(0.7*x) + 15.5
	latestCPAfter := day.Abs(cc.LatestCompetitionPoint) // [d]
	compDate := day.Abs(round(float64(eusocialPhaseDate) + y))
	// The CP is at most latestCPAfter days after the eusocial phase
	if compDate-eusocialPhaseDate > latestCPAfter {
		compDate = eusocialPhaseDate + latestCPAfter // latestCPAfter days after the eusocial phase (first worker), as in NetLogo
	}
	return compDate
}
