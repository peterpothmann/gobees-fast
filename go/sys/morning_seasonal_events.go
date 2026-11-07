package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// seasonalEvents handles the end of the season per species, the death of the males once all queens
// hibernate, and on 1 January the limit of hibernating queens.
//
// Part of the morning routine, after flowerRefill and before queenEmergence.
// In Setup (Day 0) nothing happens: Day is neither seasonStop nor 1, and TotalMales is 0. So it does not
// run at ticks = 0, and the totals of broodTotals need not exist yet.
//
// NetLogo: UpdateSeasonalEventsProc.
type seasonalEvents struct {
	rng        *rand.Rand
	calendar   *globals.Calendar
	parameters *params.Parameters
	totals     ecs.Resource[globals.BroodTotals] // set by broodTotals
	death      beeDeath

	beeFilter  *ecs.Filter1[comp.Bee]
	speciesMap *speciesTable
	beeMap     *ecs.Map1[comp.Bee]

	bees []ecs.Entity // buffer of askBees, for one call after the other
}

func (s *seasonalEvents) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.parameters = ecs.GetResource[params.Parameters](w)
	s.totals = ecs.NewResource[globals.BroodTotals](w)
	s.death = newBeeDeath(w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.speciesMap = newSpeciesTable(w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *seasonalEvents) Update(w *ecs.World) {
	// NetLogo reads the totals of its daily output of the previous evening. They are counted here and have the same
	// values: between the daily output and this point only the extinction check, tick, the calendar and the
	// foodsources are updated, and no bee changes.
	totalQueens, totalHibernatingQueens, totalMales := s.dailyTotals()

	// ORDER: irrelevant, each bee dies independently, the colony counters are sums.
	species := s.speciesMap.all()
	shuffle(s.rng, species)
	for _, sp := range species {
		if s.calendar.Day != s.speciesMap.Get(sp).SeasonStop {
			continue
		}
		s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool {
			return b.Species == sp && b.Activity != activity.Hibernate
		})
		for _, e := range s.bees {
			s.death.die(w, e, "End of season")
		}
	}

	// totals of two different times: the daily output of the previous evening and broodTotals of the
	// previous morning. ORDER: irrelevant.
	totals := s.totals.Get()
	if totalHibernatingQueens == totalQueens && totals.Eggs+totals.Larvae+totals.Pupae == 0 &&
		totalMales > 0 {
		s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool { return b.Caste == caste.Male })
		for _, e := range s.bees {
			s.death.die(w, e, "Males: all queens in hibernation!")
		}
	}

	if s.calendar.Day == 1 {
		// the limit of hibernating queens, with TotalHibernatingQueens of the previous evening
		queensToKill := totalHibernatingQueens - s.parameters.MaxHibernatingQueens
		// ORDER: the random choice of the queens matters only through the random numbers.
		// NetLogo picks one-of the hibernating queens per round. Removing the chosen one from the list gives the same
		// distribution. A cohort can remove more queens than needed.
		hibernating := []ecs.Entity{}
		if queensToKill > 0 {
			s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool {
				return b.Caste == caste.Queen && b.Activity == activity.Hibernate
			})
			hibernating = s.bees
		}
		for queensToKill > 0 {
			if len(hibernating) == 0 {
				// NetLogo: ask one-of an empty agentset is a runtime error. Cannot happen: no hibernating queen dies
				// between dailyTotals at the start of Update and this point.
				panic(fmt.Sprintf("seasonalEvents: %v queens to kill, but no hibernating queen left", queensToKill))
			}
			i := s.rng.IntN(len(hibernating))
			e := hibernating[i]
			hibernating[i] = hibernating[len(hibernating)-1]
			hibernating = hibernating[:len(hibernating)-1]
			queensToKill -= s.beeMap.Get(e).Number
			s.death.die(w, e, "max. number of hibernating queens")
		}
	}
}

func (s *seasonalEvents) Finalize(w *ecs.World) {}

// dailyTotals returns the numbers of queens (all stages), hibernating bees (only queens hibernate) and males (all
// stages), as the daily output of NetLogo sets TotalQueens, TotalHibernatingQueens and TotalMales.
// ORDER: irrelevant, only sums.
func (s *seasonalEvents) dailyTotals() (queens, hibernating, males int) {
	query := s.beeFilter.Query()
	for query.Next() {
		b := query.Get()
		switch b.Caste {
		case caste.Queen:
			queens += b.Number
		case caste.Male:
			males += b.Number
		}
		if b.Activity == activity.Hibernate {
			hibernating += b.Number
		}
	}
	return queens, hibernating, males
}
