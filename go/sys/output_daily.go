package sys

import (
	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// OutputDaily [GO-18].
//
// Sets the daily totals (globals.DailyOutput). TotalHibernatingQueens, TotalQueens and TotalMales are read by
// UpdateSeasonalEvents on the next morning [SEAS-02], [SEAS-03]. Initialize runs it once in Setup [INIT-12].
// NetLogo runs it inside with-local-randomness with random-seed ticks: nothing here changes the random numbers of
// the model, and gobees draws none.
// Not translated: plots (PlottingProc) and weather signs [OUT-03], GUI only.
//
// NetLogo: OutputDailyProc, TotalMapAreaREP.
type OutputDaily struct {
	world     *params.World
	landscape *globals.Landscape
	active    *globals.ActiveBees
	output    globals.DailyOutput

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	fsFilter     *ecs.Filter2[comp.Position, comp.Foodsource]
}

func (s *OutputDaily) Initialize(w *ecs.World) {
	s.world = ecs.GetResource[params.World](w)
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.active = ecs.GetResource[globals.ActiveBees](w)
	ecs.AddResource(w, &s.output)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.fsFilter = ecs.NewFilter2[comp.Position, comp.Foodsource](w)
	s.Update(w) // [INIT-12]
}

// Update: ORDER irrelevant, only sums.
func (s *OutputDaily) Update(w *ecs.World) {
	o := &s.output
	// [OUT-01]
	o.TotalIBMColonies, o.TotalColonies = 0, 0
	colonies := s.colonyFilter.Query()
	for colonies.Next() {
		o.TotalColonies++
		if !colonies.Get().CohortBased {
			o.TotalIBMColonies++
		}
	}

	o.TotalQueens, o.TotalMatedQueens, o.TotalUnmatedQueens, o.TotalHibernatingQueens = 0, 0, 0, 0
	o.TotalBeeAgents, o.TotalMales = 0, 0
	// mean [ weight_mg ] of bees with [ brood? = false ... ]: a mean over agents, not weighted by number
	var workerWeight, queenWeight, adultWeight float64
	var workers, queens, adults int
	bees := s.beeFilter.Query()
	for bees.Next() {
		b := bees.Get()
		o.TotalBeeAgents++
		switch b.Caste {
		case caste.Queen:
			o.TotalQueens += b.Number
			if b.Mated {
				o.TotalMatedQueens += b.Number
			} else {
				o.TotalUnmatedQueens += b.Number
			}
		case caste.Male:
			o.TotalMales += b.Number
		}
		if b.Activity == activity.Hibernate {
			o.TotalHibernatingQueens += b.Number
		}
		if !b.Brood {
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
	o.TotalActiveBees = float64(len(s.active.List))
	o.MeanWorkerWeight_mg, o.MeanQueenWeight_mg, o.MeanAdultWeight_mg = 0, 0, 0
	if workers > 0 {
		o.MeanWorkerWeight_mg = workerWeight / float64(workers)
	}
	if queens > 0 {
		o.MeanQueenWeight_mg = queenWeight / float64(queens)
	}
	if adults > 0 {
		o.MeanAdultWeight_mg = adultWeight / float64(adults)
	}

	// [OUT-02] the map area once (not again after a crop rotation), then the colony density. Output only.
	fsQuery := s.fsFilter.Query()
	foodsources := fsQuery.Count()
	fsQuery.Close()
	if o.TotalMapArea_km2 == 0 && foodsources > 0 {
		o.TotalMapArea_km2 = s.TotalMapAreaREP()
	}
	if o.TotalMapArea_km2 > 0 {
		o.ColonyDensity_km2 = o.TotalColonies / o.TotalMapArea_km2
	}
}

func (s *OutputDaily) Finalize(w *ecs.World) {}

// TotalMapAreaREP returns the area [km2] of the box around all foodsources (centre +- radius), limited to the
// world [AREA]. Output only. SIMPLIFY: not needed for the model.
func (s *OutputDaily) TotalMapAreaREP() float64 {
	// [AREA-01] max-one-of / min-one-of choose randomly among ties (with the local random seed of OutputDailyProc).
	// gobees takes the first one in storage order. Tied foodsources share the coordinate, but their radius_m can
	// differ, so the area can differ from NetLogo's in that case.
	type source struct{ x, y, r float64 }
	var top, bottom, right, left source
	first := true
	query := s.fsFilter.Query()
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
	scaling := s.landscape.Scaling_NLpatches_per_m
	topCor := top.y + top.r*scaling
	bottomCor := bottom.y - bottom.r*scaling
	rightCor := right.x + right.r*scaling
	leftCor := left.x - left.r*scaling

	topCor = min(topCor, s.world.MaxPycor)
	bottomCor = max(bottomCor, s.world.MinPycor)
	rightCor = min(rightCor, s.world.MaxPxcor)
	leftCor = max(leftCor, s.world.MinPxcor)

	area_m2 := ((topCor - bottomCor) * (rightCor - leftCor)) / (scaling * scaling)
	return area_m2 / (1000 * 1000)
}
