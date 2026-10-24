package model

import (
	"gobees/globals"
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default gobees model.
//
// The system order mirrors the NetLogo procedures SetupWithoutClearingProc (initialisation) and Go.
// Codes in brackets refer to the annotations in the NetLogo code, e.g. [GO-04].
// The parameter file must already be read into p.File [INIT-04].
//
// If the argument app is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used.
func Default(p params.Params, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	// ---- Setup (NetLogo: Setup, SetupWithoutClearingProc). Runs in the systems' Initialize, in this order.
	// [CLEAR] clear-all: not needed, a new world is empty. clear-all also resets the who counter.
	p.Apply(a.World) // [INIT-01] random seed, parameters as resources
	who := globals.WhoCounter{}
	ecs.AddResource(a.World, &who)
	counters := globals.Counters{}
	ecs.AddResource(a.World, &counters)
	activeBees := globals.ActiveBees{}
	ecs.AddResource(a.World, &activeBees)
	daytime := globals.Daytime{}
	ecs.AddResource(a.World, &daytime)
	// [INIT-02] choice of the parameter file: GUI logic, the caller passes the file.
	// [INIT-03] PanelSettingProc: GUI only, not translated.
	// [INIT-04] ReadAllParametersProc: params.ReadParameterFile, before calling Default.
	a.AddSystem(&sys.Parameters{})          // [INIT-05] ParametersProc
	a.AddSystem(&sys.CreateFoodsources{})   // [INIT-06] CreateFoodsourcesProc
	a.AddSystem(&sys.CreateSpecies{})       // [INIT-07] CreateSpeciesProc
	a.AddSystem(&sys.CreateBadgers{})       // [INIT-08] CreateBadgersProc
	a.AddSystem(&sys.CreateInitialQueens{}) // [INIT-09] CreateInitialQueensProc
	// [INIT-10] UpdateMorning_Proc at ticks = 0: runs in the Initialize of the morning systems below [GO-04].
	// [INIT-11] CreateSignsProc: GUI only, not translated.
	// [INIT-12] OutputDailyProc: will run once in OutputDaily.Initialize.
	// [INIT-13], [INIT-14]: GUI only, not translated.
	// [INIT-15] assertion on undefined habitat types: part of CreateFoodsources.

	// ---- Go, one tick = one day.
	// [GO-01] Assertions are not a system. Violations will panic where they occur (NetLogo: AssertionProc).
	// [GO-02] InspectTurtle: GUI only, not translated.
	a.AddSystem(&sys.Tick{}) // [GO-03] tick
	// [GO-04] UpdateMorning_Proc [MORN], as several systems in its order. Their Initialize runs the morning at
	// ticks = 0 [INIT-10].
	a.AddSystem(&sys.UpdateMorning{})     // [MORN-01] to [MORN-04] day of the year, foraging period
	a.AddSystem(&sys.UpdateFoodsources{}) // [MORN-05] UpdateFoodsourcesProc [UFS]
	// [MORN-06] UpdateSeasonalEventsProc: step 7.
	a.AddSystem(&sys.EmergenceNewQueens{}) // [MORN-07] EmergenceNewQueensProc [EMER]
	a.AddSystem(&sys.WakeUpBees{})         // [MORN-08] wake-up times and ActiveBeesSortedList
	// [MORN-09] CropRotationProc: step 7.
	a.AddSystem(&sys.UpdateColonies{}) // [MORN-10] UpdateColoniesProc [UCOL]
	// [MORN-11] UpdateColonyStoreBarsProc: GUI only, not translated.
	// [MORN-12] CheckNumbersProc [CHK]. The colony check on day 364 is switched off until step 7 translates the
	// end of season, without which no colony dies.
	a.AddSystem(&sys.CheckNumbers{SkipColonyCheck: true})
	a.AddSystem(&sys.NeedNectarPollenLarvaeToday{}) // [GO-05] NeedNectarPollenLarvaeTodayProc
	a.AddSystem(&sys.ActivityLoop{})                // [GO-06] to [GO-10] event loop [NEXT], [ACT]. Foraging: step 6
	a.AddSystem(&sys.EndOfDayResting{})             // [GO-11] all adult females resting
	a.AddSystem(&sys.QueensLeavingNest{})           // [GO-12] QueensLeavingNestProc
	a.AddSystem(&sys.FeedLarvae{})                  // [GO-13] FeedLarvaeProc
	a.AddSystem(&sys.QueenProductionDate{})         // [GO-14] QueenProductionDateProc
	a.AddSystem(&sys.Development{})                 // [GO-15] DevelopmentProc
	a.AddSystem(&sys.MortalityBrood{})              // [GO-16] MortalityBroodProc
	a.AddSystem(&sys.BadgersOnTheProwl{})           // [GO-17] BadgersOnTheProwlProc
	a.AddSystem(&sys.OutputDaily{})                 // [GO-18] OutputDailyProc (totals only, no plots)
	// [GO-19] DrawCohortsProc: GUI only, not translated.
	a.AddSystem(&sys.Extinction{}) // [GO-20] extinction check and run length

	return a
}
