// Package obs holds observers, which read model output from the world. They implement the observer interfaces of
// ark-tools and are written to files by the reporters of ark-tools, e.g. reporter.CSV.
//
// The model holds only state that its dynamics read (step F4). The observers compute the output from this state,
// and count from Ark events what the state does not hold (Counters).
package obs

import (
	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// DailyTotals are the daily totals of OutputDailyProc [OUT-01], [OUT-02]. The names are the NetLogo names of the
// globals.
type DailyTotals struct {
	TotalQueens            float64 // all queens, all stages (queen brood too)
	TotalMatedQueens       float64
	TotalUnmatedQueens     float64
	TotalHibernatingQueens float64 // all bees with activity "hibernate" (only queens hibernate)
	TotalColonies          float64
	TotalBeeAgents         float64 // number of bee agents (cohorts count once)
	TotalMales             float64 // all males, all stages (male brood too)
	TotalActiveBees        float64 // length of ActiveBeesSortedList after the event loop
	MeanWorkerWeight_mg    float64 // 0 without adult workers
	MeanQueenWeight_mg     float64 // 0 without adult queens
	MeanAdultWeight_mg     float64 // 0 without adults
	TotalMapArea_km2       float64 // set once [OUT-02], [AREA]
	ColonyDensity_km2      float64 // [OUT-02]
}

// DailyOutput observes the daily totals of OutputDailyProc [GO-18], [INIT-12].
//
// NetLogo sets them at the end of each day and once in Setup. The observer computes them from the state when its
// values are read. Initialize computes the map area, i.e. after Setup when the observer is added after the model
// systems, as OutputDailyProc in Setup. NetLogo runs OutputDailyProc inside with-local-randomness: nothing here
// changes the random numbers of the model, and the observer draws none.
// Not translated: plots (PlottingProc) and weather signs [OUT-03], GUI only.
//
// One row per day: ticks, then the totals. The column names are the NetLogo names of the globals.
// reporter.CSV adds its own first column t, which counts its updates from 0 and is therefore ticks - 1 (see
// globals.Ticks). The column ticks gives the NetLogo day.
//
// NetLogo: OutputDailyProc, TotalMapAreaREP.
type DailyOutput struct {
	// Totals holds the totals of the last call of Values.
	Totals DailyTotals

	ticks     ecs.Resource[globals.Ticks]
	world     *params.World
	landscape *globals.Landscape
	active    *globals.ActiveBees

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	fsFilter     *ecs.Filter2[comp.Position, comp.Foodsource]
	values       []float64
}

func (o *DailyOutput) Initialize(w *ecs.World) {
	o.ticks = ecs.NewResource[globals.Ticks](w)
	o.world = ecs.GetResource[params.World](w)
	o.landscape = ecs.GetResource[globals.Landscape](w)
	o.active = ecs.GetResource[globals.ActiveBees](w)
	o.beeFilter = ecs.NewFilter1[comp.Bee](w)
	o.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	o.fsFilter = ecs.NewFilter2[comp.Position, comp.Foodsource](w)
	o.values = make([]float64, 14)
	o.mapArea() // [INIT-12]
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
	o.update()
	d := &o.Totals
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

// update sets the totals from the state. ORDER irrelevant, only sums.
func (o *DailyOutput) update() {
	d := &o.Totals
	// [OUT-01]
	// TotalIBMColonies is not translated: gobees-fast has only cohort-based colonies.
	colonies := o.colonyFilter.Query()
	d.TotalColonies = float64(colonies.Count())
	colonies.Close()

	d.TotalQueens, d.TotalMatedQueens, d.TotalUnmatedQueens, d.TotalHibernatingQueens = 0, 0, 0, 0
	d.TotalBeeAgents, d.TotalMales = 0, 0
	// mean [ weight_mg ] of bees with [ brood? = false ... ]: a mean over agents, not weighted by number
	var workerWeight, queenWeight, adultWeight float64
	var workers, queens, adults int
	bees := o.beeFilter.Query()
	for bees.Next() {
		b := bees.Get()
		d.TotalBeeAgents++
		switch b.Caste {
		case caste.Queen:
			d.TotalQueens += float64(b.Number)
			if b.Mated {
				d.TotalMatedQueens += float64(b.Number)
			} else {
				d.TotalUnmatedQueens += float64(b.Number)
			}
		case caste.Male:
			d.TotalMales += float64(b.Number)
		}
		if b.Activity == activity.Hibernate {
			d.TotalHibernatingQueens += float64(b.Number)
		}
		if b.Stage == stage.Adult { // brood? = false
			adults++
			adultWeight += b.Weight_mg
			switch b.Caste {
			case caste.Worker:
				workers++
				workerWeight += b.Weight_mg
			case caste.Queen:
				queens++
				queenWeight += b.Weight_mg
			}
		}
	}
	d.TotalActiveBees = float64(len(o.active.List))
	d.MeanWorkerWeight_mg, d.MeanQueenWeight_mg, d.MeanAdultWeight_mg = 0, 0, 0
	if workers > 0 {
		d.MeanWorkerWeight_mg = workerWeight / float64(workers)
	}
	if queens > 0 {
		d.MeanQueenWeight_mg = queenWeight / float64(queens)
	}
	if adults > 0 {
		d.MeanAdultWeight_mg = adultWeight / float64(adults)
	}

	// [OUT-02] the map area once (not again after a crop rotation), then the colony density
	o.mapArea()
	if d.TotalMapArea_km2 > 0 {
		d.ColonyDensity_km2 = d.TotalColonies / d.TotalMapArea_km2
	}
}

// mapArea sets the map area the first time that there are foodsources [OUT-02].
func (o *DailyOutput) mapArea() {
	if o.Totals.TotalMapArea_km2 != 0 {
		return
	}
	fsQuery := o.fsFilter.Query()
	foodsources := fsQuery.Count()
	fsQuery.Close()
	if foodsources > 0 {
		o.Totals.TotalMapArea_km2 = o.TotalMapAreaREP()
	}
}

// TotalMapAreaREP returns the area [km2] of the box around all foodsources (centre +- radius), limited to the
// world [AREA]. SIMPLIFY: output only.
func (o *DailyOutput) TotalMapAreaREP() float64 {
	// [AREA-01] max-one-of / min-one-of choose randomly among ties (with the local random seed of OutputDailyProc).
	// gobees takes the first one in storage order. Tied foodsources share the coordinate, but their radius_m can
	// differ, so the area can differ from NetLogo's in that case.
	type source struct{ x, y, r float64 }
	var top, bottom, right, left source
	first := true
	query := o.fsFilter.Query()
	for query.Next() {
		pos, fs := query.Get()
		src := source{pos.X, pos.Y, fs.Radius_m}
		if first {
			top, bottom, right, left = src, src, src, src
			first = false
			continue
		}
		if src.y > top.y {
			top = src
		}
		if src.y < bottom.y {
			bottom = src
		}
		if src.x > right.x {
			right = src
		}
		if src.x < left.x {
			left = src
		}
	}
	scaling := o.landscape.Scaling_NLpatches_per_m
	topCor := top.y + top.r*scaling
	bottomCor := bottom.y - bottom.r*scaling
	rightCor := right.x + right.r*scaling
	leftCor := left.x - left.r*scaling

	topCor = min(topCor, o.world.MaxPycor)
	bottomCor = max(bottomCor, o.world.MinPycor)
	rightCor = min(rightCor, o.world.MaxPxcor)
	leftCor = max(leftCor, o.world.MinPxcor)

	area_m2 := ((topCor - bottomCor) * (rightCor - leftCor)) / (scaling * scaling)
	return area_m2 / (1000 * 1000)
}
