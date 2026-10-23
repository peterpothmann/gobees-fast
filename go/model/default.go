package model

import (
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
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
	// [CLEAR] clear-all: not needed, a new world is empty.
	p.Apply(a.World) // [INIT-01] random seed, parameters as resources
	// [INIT-02] choice of the parameter file: GUI logic, the caller passes the file.
	// [INIT-03] PanelSettingProc: GUI only, not translated.
	// [INIT-04] ReadAllParametersProc: params.ReadParameterFile, before calling Default.
	a.AddSystem(&sys.Parameters{})          // [INIT-05] ParametersProc
	a.AddSystem(&sys.CreateFoodsources{})   // [INIT-06] CreateFoodsourcesProc
	a.AddSystem(&sys.CreateSpecies{})       // [INIT-07] CreateSpeciesProc
	a.AddSystem(&sys.CreateBadgers{})       // [INIT-08] CreateBadgersProc
	a.AddSystem(&sys.CreateInitialQueens{}) // [INIT-09] CreateInitialQueensProc
	// [INIT-10] UpdateMorning_Proc at ticks = 0: will run once in UpdateMorning.Initialize.
	// [INIT-11] CreateSignsProc: GUI only, not translated.
	// [INIT-12] OutputDailyProc: will run once in OutputDaily.Initialize.
	// [INIT-13], [INIT-14]: GUI only, not translated.
	// [INIT-15] assertion on undefined habitat types: will be part of CreateFoodsources.

	// ---- Go, one tick = one day.
	// [GO-01] Assertions are not a system. Violations will panic where they occur (NetLogo: AssertionProc).
	// [GO-02] InspectTurtle: GUI only, not translated.
	a.AddSystem(&sys.Tick{})                        // [GO-03] tick
	a.AddSystem(&sys.UpdateMorning{})               // [GO-04] UpdateMorning_Proc
	a.AddSystem(&sys.NeedNectarPollenLarvaeToday{}) // [GO-05] NeedNectarPollenLarvaeTodayProc
	a.AddSystem(&sys.ActivityLoop{})                // [GO-06] to [GO-10] event loop, NextActiveBeeREP, ActivityProc
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
