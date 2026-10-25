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
// run in Initialize, and the totals of OutputDaily and CheckNumbers need not exist yet.
// Not translated: moving the hibernating queens to (0, 0) on 1 January (GUI only, their position is never read)
// and the message in the output area.
//
// NetLogo: UpdateSeasonalEventsProc.
type UpdateSeasonalEvents struct {
	rng      *rand.Rand
	calendar *globals.Calendar
	file     *params.ParameterFile
	counters *globals.Counters
	output   ecs.Resource[globals.DailyOutput] // set by OutputDaily
	totals   ecs.Resource[globals.Totals]      // set by CheckNumbers
	death    beeDeath

	speciesFilter *ecs.Filter1[comp.Species]
	beeFilter     *ecs.Filter1[comp.Bee]
	fsFilter      *ecs.Filter1[comp.FoodsourceState]
	speciesMap    *ecs.Map1[comp.Species]
	beeMap        *ecs.Map1[comp.Bee]
	countsMap     *ecs.Map1[comp.ColonyCounts]
}

func (s *UpdateSeasonalEvents) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.counters = ecs.GetResource[globals.Counters](w)
	s.output = ecs.NewResource[globals.DailyOutput](w)
	s.totals = ecs.NewResource[globals.Totals](w)
	s.death = newBeeDeath(w)
	s.speciesFilter = ecs.NewFilter1[comp.Species](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.fsFilter = ecs.NewFilter1[comp.FoodsourceState](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.countsMap = ecs.NewMap1[comp.ColonyCounts](w)
}

func (s *UpdateSeasonalEvents) Update(w *ecs.World) {
	// [SEAS-01] ORDER: irrelevant, each bee dies independently, the colony counters are sums.
	species := []ecs.Entity{}
	speciesQuery := s.speciesFilter.Query()
	for speciesQuery.Next() {
		species = append(species, speciesQuery.Entity())
	}
	util.Shuffle(s.rng, species)
	for _, sp := range species {
		if s.calendar.Day != s.speciesMap.Get(sp).SeasonStop {
			continue
		}
		for _, e := range askBees(s.beeFilter, s.rng, func(b *comp.Bee) bool {
			return b.SpeciesID == sp && b.Activity != activity.Hibernate
		}) {
			if b := s.beeMap.Get(e); b.Brood {
				s.countsMap.Get(b.ColonyID).BroodDeathEndSeason += b.Number
			}
			s.death.DieProc(w, e, "End of season")
		}
	}

	// [SEAS-02] totals of two different times: OutputDaily of the previous evening [OUT-01] and CheckNumbers of the
	// previous morning [CHK-02], finding 39 (see trans/netlogo-findings.md#finding-39). ORDER: irrelevant.
	out, totals := s.output.Get(), s.totals.Get()
	if out.TotalHibernatingQueens == out.TotalQueens && totals.TotalEggs+totals.TotalLarvae+totals.TotalPupae == 0 &&
		out.TotalMales > 0 {
		for _, e := range askBees(s.beeFilter, s.rng, func(b *comp.Bee) bool { return b.Caste == caste.Male }) {
			s.death.DieProc(w, e, "Males: all queens in hibernation!")
		}
	}

	// [SEAS-03]
	if s.calendar.Day == 1 {
		// output only
		population := 0.0
		query := s.beeFilter.Query()
		for query.Next() {
			population += query.Get().Number
		}
		s.counters.PopulationSizeDay365List = append(s.counters.PopulationSizeDay365List, population)

		// the limit of hibernating queens, with TotalHibernatingQueens of the previous evening [OUT-01]
		queensToKill := out.TotalHibernatingQueens - s.file.MaxHibernatingQueens
		if queensToKill > 0 {
			s.counters.TotalHibernatingQueensEverRemoved += queensToKill // output only, the planned number
		}
		// ORDER: the random choice of the queens matters only through the random numbers.
		// NetLogo: one-of the hibernating queens per round. Removing the chosen one from the list gives the same
		// distribution. A cohort can remove more queens than needed.
		hibernating := []ecs.Entity{}
		if queensToKill > 0 {
			hibernating = askBees(s.beeFilter, s.rng, func(b *comp.Bee) bool {
				return b.Caste == caste.Queen && b.Activity == activity.Hibernate
			})
		}
		for queensToKill > 0 {
			if len(hibernating) == 0 {
				// NetLogo: ask one-of an empty agentset is a runtime error. Cannot happen: no hibernating queen dies
				// between OutputDaily and this point.
				panic(fmt.Sprintf("UpdateSeasonalEventsProc: %v queens to kill, but no hibernating queen left", queensToKill))
			}
			i := s.rng.IntN(len(hibernating))
			e := hibernating[i]
			hibernating[i] = hibernating[len(hibernating)-1]
			hibernating = hibernating[:len(hibernating)-1]
			queensToKill -= s.beeMap.Get(e).Number
			s.death.DieProc(w, e, "max. number of hibernating queens")
		}

		// visit counters: read only by the GUI and output files. ORDER: irrelevant.
		if s.file.CumulVisitsOnlyLastYear {
			fs := s.fsFilter.Query()
			for fs.Next() {
				state := fs.Get()
				state.CumulNectarVisits = 0
				state.CumulPollenVisits = 0
			}
		}
	}
}

func (s *UpdateSeasonalEvents) Finalize(w *ecs.World) {}
