package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// UpdateSeasonalEvents handles the end of the season per species, the death of the males once all queens
// hibernate, and on 1 January the limit of hibernating queens [SEAS].
//
// Part of the morning routine [MORN-06], after UpdateFoodsources and before EmergenceNewQueens.
// In Setup (Day 0, [INIT-10]) nothing happens: Day is neither seasonStop nor 1, and TotalMales is 0. So it does not
// run in Initialize, and the totals of CheckNumbers need not exist yet.
// Not translated: moving the hibernating queens to (0, 0) on 1 January (GUI only, their position is never read)
// and the message in the output area.
//
// NetLogo: UpdateSeasonalEventsProc.
type UpdateSeasonalEvents struct {
	rng      *rand.Rand
	calendar *globals.Calendar
	file     *params.ParameterFile
	totals   ecs.Resource[globals.Totals] // set by CheckNumbers
	death    beeDeath

	beeFilter  *ecs.Filter1[comp.Bee]
	speciesMap *speciesTable
	beeMap     *ecs.Map1[comp.Bee]

	bees []ecs.Entity // buffer of askBees, for one call after the other
}

func (s *UpdateSeasonalEvents) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.totals = ecs.NewResource[globals.Totals](w)
	s.death = newBeeDeath(w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.speciesMap = newSpeciesTable(w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *UpdateSeasonalEvents) Update(w *ecs.World) {
	// The totals of OutputDailyProc of the previous evening [OUT-01] that [SEAS-02] and [SEAS-03] read. gobees counts
	// them here, before [SEAS-01] (step F4). The same values: between OutputDailyProc and this point only the
	// extinction check, tick, the calendar and the foodsources are updated, and no bee changes.
	totalQueens, totalHibernatingQueens, totalMales := s.dailyTotals()

	// [SEAS-01] ORDER: irrelevant, each bee dies independently, the colony counters are sums.
	species := s.speciesMap.all()
	util.Shuffle(s.rng, species)
	for _, sp := range species {
		if s.calendar.Day != s.speciesMap.Get(sp).SeasonStop {
			continue
		}
		s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool {
			return b.SpeciesID == sp && b.Activity != activity.Hibernate
		})
		for _, e := range s.bees {
			// broodDeathEndSeason of the colony: read only by the egg accounting [CHK-04], removed in step F4
			s.death.DieProc(w, e, "End of season")
		}
	}

	// [SEAS-02] totals of two different times: OutputDailyProc of the previous evening [OUT-01] and CheckNumbers of the
	// previous morning [CHK-02], finding 39 (see trans/netlogo-findings.md#finding-39). ORDER: irrelevant.
	totals := s.totals.Get()
	if totalHibernatingQueens == totalQueens && totals.TotalEggs+totals.TotalLarvae+totals.TotalPupae == 0 &&
		totalMales > 0 {
		s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool { return b.Caste == caste.Male })
		for _, e := range s.bees {
			s.death.DieProc(w, e, "Males: all queens in hibernation!")
		}
	}

	// [SEAS-03]
	if s.calendar.Day == 1 {
		// populationSizeDay365List: output only, never read, removed in step F3

		// the limit of hibernating queens, with TotalHibernatingQueens of the previous evening [OUT-01]
		// totalHibernatingQueensEverRemoved (the planned number): output only, computed by the observers (step F4)
		queensToKill := totalHibernatingQueens - s.file.MaxHibernatingQueens
		// ORDER: the random choice of the queens matters only through the random numbers.
		// NetLogo: one-of the hibernating queens per round. Removing the chosen one from the list gives the same
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
				panic(fmt.Sprintf("UpdateSeasonalEventsProc: %v queens to kill, but no hibernating queen left", queensToKill))
			}
			i := s.rng.IntN(len(hibernating))
			e := hibernating[i]
			hibernating[i] = hibernating[len(hibernating)-1]
			hibernating = hibernating[:len(hibernating)-1]
			queensToKill -= s.beeMap.Get(e).Number
			s.death.DieProc(w, e, "max. number of hibernating queens")
		}
		// reset of the visit counters with CumulVisitsOnlyLastYear?: the counters are never read, removed in step F3
	}
}

func (s *UpdateSeasonalEvents) Finalize(w *ecs.World) {}

// dailyTotals returns the numbers of queens (all stages), hibernating bees (only queens hibernate) and males (all
// stages), as OutputDailyProc sets TotalQueens, TotalHibernatingQueens and TotalMales [OUT-01].
// ORDER: irrelevant, only sums.
func (s *UpdateSeasonalEvents) dailyTotals() (queens, hibernating, males int) {
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
