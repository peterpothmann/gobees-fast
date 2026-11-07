package schedule

import (
	"gobees/globals"
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default gobees model.
//
// The system order mirrors the NetLogo procedures SetupWithoutClearingProc (initialisation) and Go: four systems
// (sys.Setup, sys.Morning, sys.Day, sys.Evening) call the NetLogo procedures in their order (step F14).
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
	// [CLEAR] clear-all: not needed, a new world is empty. clear-all also resets the who counter, which gobees-fast
	// does not keep (step F12).
	p.Apply(a.World) // [INIT-01] random seed, parameters as resources
	activeBees := globals.ActiveBees{}
	ecs.AddResource(a.World, &activeBees)
	daytime := globals.Daytime{}
	ecs.AddResource(a.World, &daytime)
	// [INIT-02] choice of the parameter file: GUI logic, the caller passes the file.
	// [INIT-03] PanelSettingProc: GUI only, not translated.
	// [INIT-04] ReadAllParametersProc: input.Load (a run file and its landscape files), before calling Default, so no
	// system opens a file.
	// [INIT-05] ParametersProc: package input derives the values, p.Apply adds them (params.Derived, params.ForagingPeriod).
	a.AddSystem(&sys.Setup{}) // [INIT-06] to [INIT-09]
	// [INIT-10] UpdateMorning_Proc at ticks = 0: in Morning.Initialize.
	// [INIT-11] CreateSignsProc: GUI only, not translated.
	// [INIT-12] OutputDailyProc: output only, in the observers (obs.DailyOutput) [GO-18].
	// [INIT-13], [INIT-14]: GUI only, not translated.
	// [INIT-15] assertion on undefined habitat types: part of landscapeSetup.

	// ---- Go, one tick = one day.
	// [GO-01] Assertions are not a system. Violations will panic where they occur (NetLogo: AssertionProc).
	// [GO-02] InspectTurtle: GUI only, not translated.
	a.AddSystem(&sys.Morning{}) // [GO-03] to [GO-05], and [INIT-10] in Initialize
	a.AddSystem(&sys.Day{})     // [GO-06] to [GO-11]
	a.AddSystem(&sys.Evening{}) // [GO-12] to [GO-20]
	return a
}
