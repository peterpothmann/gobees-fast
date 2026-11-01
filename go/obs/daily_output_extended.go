package obs

import (
	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// DailyOutputExtended observes the columns of DailyOutput and further daily values of gobees-fast. It is the
// output of the comparisons with the reference commit (go/compare) and of the command line (-o).
//
// The column names stay the same in all steps of gobees-fast: the comparison matches the columns by name.
// The first 15 columns (ticks and 14 totals) are those of DailyOutput, so that a version without the further
// columns (e.g. gobees) can be compared on them. The further columns, all at the end of the day:
//
//   - Eggs, Larvae, Pupae, AdultWorkers: bees in these stages, counted by the observer (globals.Totals holds
//     morning values)
//   - EnergyStore_kJ, PollenStore_g: stores summed over all colonies
//   - ForagingTripsToday: foraging trips of today (globals.Totals.TotalForagingTripsToday)
//   - ForagerDeathsToday: bees that died on a foraging trip today
//   - TotalAdultQueensEverProduced: queens emerged from pupae since the start
//   - MeanCompetitionPointDate: mean competition point date (ticks) of the living colonies that have one, 0 if none
type DailyOutputExtended struct {
	daily    DailyOutput
	totals   ecs.Resource[globals.Totals]
	counters ecs.Resource[globals.Counters]
	bees     *ecs.Filter1[comp.Bee]
	colonies *ecs.Filter2[comp.Colony, comp.ColonyStores]
	values   []float64
	// lastForagerDeaths is the counter of foragers that died on a trip at the end of the previous day.
	lastForagerDeaths float64
}

func (o *DailyOutputExtended) Initialize(w *ecs.World) {
	o.daily.Initialize(w)
	o.totals = ecs.NewResource[globals.Totals](w)
	o.counters = ecs.NewResource[globals.Counters](w)
	o.bees = ecs.NewFilter1[comp.Bee](w)
	o.colonies = ecs.NewFilter2[comp.Colony, comp.ColonyStores](w)
	o.values = make([]float64, len(o.Header()))
}

func (o *DailyOutputExtended) Update(w *ecs.World) {}

func (o *DailyOutputExtended) Header() []string {
	return append(o.daily.Header(),
		"Eggs",
		"Larvae",
		"Pupae",
		"AdultWorkers",
		"EnergyStore_kJ",
		"PollenStore_g",
		"ForagingTripsToday",
		"ForagerDeathsToday",
		"TotalAdultQueensEverProduced",
		"MeanCompetitionPointDate",
	)
}

func (o *DailyOutputExtended) Values(w *ecs.World) []float64 {
	n := copy(o.values, o.daily.Values(w))

	var eggs, larvae, pupae, workers float64
	query := o.bees.Query()
	for query.Next() {
		b := query.Get()
		switch {
		case b.Stage == stage.Egg:
			eggs += b.Number
		case b.Stage == stage.Larva:
			larvae += b.Number
		case b.Stage == stage.Pupa:
			pupae += b.Number
		case b.Caste == caste.Worker:
			workers += b.Number
		}
	}

	var energy, pollen, cpSum, cpCount float64
	cq := o.colonies.Query()
	for cq.Next() {
		colony, stores := cq.Get()
		energy += stores.EnergyStore_kJ
		pollen += stores.PollenStore_g
		if colony.CompetitionPointDate != params.NotSetHigh {
			cpSum += colony.CompetitionPointDate
			cpCount++
		}
	}
	meanCP := 0.0
	if cpCount > 0 {
		meanCP = cpSum / cpCount
	}

	counters := o.counters.Get()
	foragerDeaths := counters.TotalForagersEverDiedForaging - o.lastForagerDeaths
	o.lastForagerDeaths = counters.TotalForagersEverDiedForaging

	for i, v := range []float64{
		eggs, larvae, pupae, workers,
		energy, pollen,
		o.totals.Get().TotalForagingTripsToday,
		foragerDeaths,
		counters.TotalAdultQueensEverProduced,
		meanCP,
	} {
		o.values[n+i] = v
	}
	return o.values
}
