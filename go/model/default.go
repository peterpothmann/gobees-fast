package model

import (
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
)

// Default sets up the default gobees model.
//
// The system order mirrors the NetLogo procedure Go. Codes in brackets refer to the
// annotations in the NetLogo code, e.g. [GO-04].
//
// If the argument app is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used.
func Default(p params.Params, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(a.World)

	// Initialization (NetLogo: Setup) is not yet translated.

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
