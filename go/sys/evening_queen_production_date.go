package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// queenProductionDate starts the queen production of a colony: QueenProduction becomes true, and
// QueenProductionDate is back-calculated as the laying day of the larvae now at the caste determination age. These
// larvae got their caste on the previous evening, as workers, so the date is the laying day of the last worker
// larvae. No queen is created here. Must run before development.
//
// NetLogo: QueenProductionDateProc.
type queenProductionDate struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	constants *params.Constants

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map2[comp.Colony, comp.ColonyCounts]
	beeMap       *ecs.Map2[comp.Bee, comp.Development]
	speciesMap   *speciesTable

	larvae map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *queenProductionDate) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.constants = ecs.GetResource[params.Constants](w)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap2[comp.Colony, comp.ColonyCounts](w)
	s.beeMap = ecs.NewMap2[comp.Bee, comp.Development](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *queenProductionDate) Update(w *ecs.World) {
	ticks := s.ticks.Ticks
	// ORDER: irrelevant, each colony only reads its own larvae and changes itself. Shuffled, as NetLogo asks the
	// colonies in random order.
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		colony, counts := s.colonyMap.Get(col)

		// 7 d + QueenDeterminationAge, or the largest broodAge of the larvae at larval day 3
		averageEggTime := 7 // [d]
		timeEggToLarvalAgeAtQueenDetermination := averageEggTime + s.speciesMap.Get(colony.Species).QueenDeterminationAge
		found := false
		maxBroodAge := 0
		for _, e := range s.larvae[col] {
			b, dev := s.beeMap.Get(e)
			// the species of the larva, as in NetLogo (the same as the colony's)
			if dev.LarvaTime == s.speciesMap.Get(b.Species).QueenDeterminationAge {
				if !found || b.BroodAge > maxBroodAge {
					maxBroodAge = b.BroodAge
				}
				found = true
			}
		}
		if found {
			timeEggToLarvalAgeAtQueenDetermination = maxBroodAge
		}

		// without switch point the condition is false, as with the sentinel NotSetHigh in NetLogo
		layingDate := ticks - day.Abs(timeEggToLarvalAgeAtQueenDetermination)
		if !colony.QueenProduction && colony.SwitchPointDate.IsSet() &&
			layingDate >= colony.SwitchPointDate.Day()-day.Abs(s.constants.QueenEggsBeforeSwitchPoint) &&
			counts.LarvaWorkerRatio < s.constants.LarvaWorkerRatioLimit {
			colony.QueenProduction = true
			colony.QueenProductionDate = day.At(layingDate)
		}
	}
}

func (s *queenProductionDate) Finalize(w *ecs.World) {}
