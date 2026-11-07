package sys

import (
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// group runs its parts in order: Initialize of all parts in Initialize, Update of all parts in Update. A part is one
// NetLogo procedure (step F14) with the methods of an Ark system, but only its group calls them, the app does not.
// All parts draw from the one random source (resource.Rand), so this order fixes the random numbers.
type group struct {
	parts []app.System
}

func (g *group) Initialize(w *ecs.World) {
	for _, p := range g.parts {
		p.Initialize(w)
	}
}

func (g *group) Update(w *ecs.World) {
	for _, p := range g.parts {
		p.Update(w)
	}
}

func (g *group) Finalize(w *ecs.World) {}

// Setup builds the world: the landscape, the species, the badgers and the initial queens [INIT-06] to [INIT-09].
// The parameters and the values that ParametersProc derives [INIT-05] are resources before (params.Params.Apply).
// The morning at ticks = 0 [INIT-10] runs in Morning.Initialize, which follows Setup.
//
// NetLogo: Setup, SetupWithoutClearingProc. The parts work only in Initialize.
type Setup struct{ group }

func (s *Setup) Initialize(w *ecs.World) {
	s.parts = []app.System{
		&CreateFoodsources{},   // [INIT-06] CreateFoodsourcesProc
		&CreateSpecies{},       // [INIT-07] CreateSpeciesProc
		&CreateBadgers{},       // [INIT-08] CreateBadgersProc
		&CreateInitialQueens{}, // [INIT-09] CreateInitialQueensProc
	}
	s.group.Initialize(w)
}

// Morning advances the day and runs the morning routine [GO-03] to [GO-05]. Its Initialize also runs the morning
// at ticks = 0 [INIT-10]: UpdateMorning, UpdateFoodsources, WakeUpBees, UpdateColonies and CheckNumbers call their
// Update at the end of their Initialize. The other parts do nothing at ticks = 0 (see their doc comments).
//
// NetLogo: tick, UpdateMorning_Proc [MORN], NeedNectarPollenLarvaeTodayProc.
type Morning struct{ group }

func (s *Morning) Initialize(w *ecs.World) {
	s.parts = []app.System{
		&Tick{},                        // [GO-03] tick
		&UpdateMorning{},               // [MORN-01] to [MORN-04] day of the year, foraging period
		&UpdateFoodsources{},           // [MORN-05] UpdateFoodsourcesProc [UFS]
		&UpdateSeasonalEvents{},        // [MORN-06] UpdateSeasonalEventsProc [SEAS]
		&EmergenceNewQueens{},          // [MORN-07] EmergenceNewQueensProc [EMER]
		&WakeUpBees{},                  // [MORN-08] wake-up times and ActiveBeesSortedList
		&CropRotation{},                // [MORN-09] CropRotationProc [CROP] on Day 365
		&UpdateColonies{},              // [MORN-10] UpdateColoniesProc [UCOL]
		&CheckNumbers{},                // [MORN-12] CheckNumbersProc [CHK], totals only, no checks
		&NeedNectarPollenLarvaeToday{}, // [GO-05] NeedNectarPollenLarvaeTodayProc
	}
	// [MORN-11] UpdateColonyStoreBarsProc: GUI only, not translated.
	s.group.Initialize(w)
}

// Day runs the activities of the bees in the order of their personal time, then lets all adult females rest
// [GO-06] to [GO-11].
//
// NetLogo: the event loop of Go [NEXT], [ACT].
type Day struct{ group }

func (s *Day) Initialize(w *ecs.World) {
	s.parts = []app.System{
		&ActivityLoop{},    // [GO-06] to [GO-10] event loop [NEXT], [ACT]
		&EndOfDayResting{}, // [GO-11] all adult females resting
	}
	s.group.Initialize(w)
}

// Evening runs what happens after the activities: queens leave the nest, larvae are fed, the brood develops and
// dies, the badgers prowl, and the run ends if the population is extinct [GO-12] to [GO-20].
//
// NetLogo: the end of Go.
type Evening struct{ group }

func (s *Evening) Initialize(w *ecs.World) {
	s.parts = []app.System{
		&QueensLeavingNest{},   // [GO-12] QueensLeavingNestProc
		&FeedLarvae{},          // [GO-13] FeedLarvaeProc
		&QueenProductionDate{}, // [GO-14] QueenProductionDateProc
		&Development{},         // [GO-15] DevelopmentProc
		&MortalityBrood{},      // [GO-16] MortalityBroodProc
		&BadgersOnTheProwl{},   // [GO-17] BadgersOnTheProwlProc
		&Extinction{},          // [GO-20] extinction check and run length
	}
	// [GO-18] OutputDailyProc: output only, in the observers (obs.DailyOutput). UpdateSeasonalEvents counts the
	// three totals it reads itself.
	// [GO-19] DrawCohortsProc: GUI only, not translated.
	s.group.Initialize(w)
}
