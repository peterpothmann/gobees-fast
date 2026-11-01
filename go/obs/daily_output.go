// Package obs holds observers, which read model output from the world. They implement the observer interfaces of
// ark-tools and are written to files by the reporters of ark-tools, e.g. reporter.CSV.
package obs

import (
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// DailyOutput observes the daily totals of OutputDailyProc (globals.DailyOutput) [GO-18], [INIT-12].
//
// One row per day: ticks, then the totals. The column names are the NetLogo names of the globals.
// reporter.CSV adds its own first column t, which counts its updates from 0 and is therefore ticks - 1 (see
// globals.Ticks). The column ticks gives the NetLogo day.
type DailyOutput struct {
	ticks  ecs.Resource[globals.Ticks]
	output ecs.Resource[globals.DailyOutput]
	values []float64
}

func (o *DailyOutput) Initialize(w *ecs.World) {
	o.ticks = ecs.NewResource[globals.Ticks](w)
	o.output = ecs.NewResource[globals.DailyOutput](w)
	o.values = make([]float64, 14)
}

func (o *DailyOutput) Update(w *ecs.World) {}

func (o *DailyOutput) Header() []string {
	return []string{
		"ticks",
		"TotalQueens",
		"TotalMatedQueens",
		"TotalUnmatedQueens",
		"TotalHibernatingQueens",
		"TotalColonies",
		"TotalBeeAgents",
		"TotalMales",
		"TotalActiveBees",
		"MeanWorkerWeight_mg",
		"MeanQueenWeight_mg",
		"MeanAdultWeight_mg",
		"TotalMapArea_km2",
		"ColonyDensity_km2",
	}
}

func (o *DailyOutput) Values(w *ecs.World) []float64 {
	d := o.output.Get()
	o.values[0] = float64(o.ticks.Get().Ticks)
	o.values[1] = d.TotalQueens
	o.values[2] = d.TotalMatedQueens
	o.values[3] = d.TotalUnmatedQueens
	o.values[4] = d.TotalHibernatingQueens
	o.values[5] = d.TotalColonies
	o.values[6] = d.TotalBeeAgents
	o.values[7] = d.TotalMales
	o.values[8] = d.TotalActiveBees
	o.values[9] = d.MeanWorkerWeight_mg
	o.values[10] = d.MeanQueenWeight_mg
	o.values[11] = d.MeanAdultWeight_mg
	o.values[12] = d.TotalMapArea_km2
	o.values[13] = d.ColonyDensity_km2
	return o.values
}
