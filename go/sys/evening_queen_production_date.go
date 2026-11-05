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

// queenProductionDate [GO-14].
//
// Starts the queen production of a colony [QPD]: queenProduction? becomes true, and queenProductionDate is
// back-calculated as the laying day of the larvae now at the caste determination age (finding 42, see
// trans/netlogo-findings.md#finding-42). No queen is created here. Must run before development.
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
	// ORDER: irrelevant, each colony only reads its own larvae and changes itself. Shuffled (decision C).
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		colony, counts := s.colonyMap.Get(col)

		// [QPD-01] 7 d + dev_larvalAge_QueenDetermination_d, or the largest broodAge of the larvae at larval day 3
		averageCumulTimeEgg_d := 7
		timeEggToLarvalAgeAtQueenDetermination := averageCumulTimeEgg_d + s.speciesMap.Get(colony.SpeciesIDcolony).Dev_larvalAge_QueenDetermination_d
		found := false
		maxBroodAge := 0
		for _, e := range s.larvae[col] {
			b, dev := s.beeMap.Get(e)
			// speciesID of the larva, as in NetLogo (the same species as the colony's)
			if dev.CumulTimeLarva_d == s.speciesMap.Get(b.SpeciesID).Dev_larvalAge_QueenDetermination_d {
				if !found || b.BroodAge > maxBroodAge {
					maxBroodAge = b.BroodAge
				}
				found = true
			}
		}
		if found {
			timeEggToLarvalAgeAtQueenDetermination = maxBroodAge
		}

		// [QPD-02] without switch point (NetLogo: NotSetHigh) the condition is false
		layingDate := ticks - day.Abs(timeEggToLarvalAgeAtQueenDetermination)
		if !colony.QueenProduction && colony.SwitchPointDate.IsSet() &&
			layingDate >= colony.SwitchPointDate.Day()-day.Abs(s.constants.QueenDestinedEggsBeforeSP_d) &&
			counts.LarvaWorkerRatio < s.constants.LarvaWorkerRatioTH {
			colony.QueenProduction = true
			colony.QueenProductionDate = day.At(layingDate)
		}
	}
}

func (s *queenProductionDate) Finalize(w *ecs.World) {}
