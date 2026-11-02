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

// NetLogoMetrics are the columns of DailyOutputExtended with the metrics of the NetLogo experiments in
// go/validation/experiments (step7 to step8): same names and definitions as the NetLogo reporters in the experiment
// files. go/compare compares them with the NetLogo reference of each scenario.
var NetLogoMetrics = []string{
	"colonies", "dead_colonies", "colonies_founded", "eggs", "larvae", "pupae", "workers", "males", "young_queens",
	"hibernating_queens", "energy_kJ", "pollen_g", "trips", "nectar_left_l", "pollen_left_kg",
	"badgers", "foodsources", "ibm_colonies", "bee_agents", "diploid_males",
	"colonies_B_terrestris", "colonies_B_pascuorum", "colonies_B_lapidarius", "colonies_B_hortorum",
	"colonies_B_pratorum", "colonies_B_hypnorum", "psithyrus",
	"foraging_period_s", "queens_removed",
}

// netLogoSpecies are the Bombus species of the metrics colonies_<species>.
var netLogoSpecies = []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum"}

// DailyOutputExtended observes the columns of DailyOutput, the NetLogo metrics and further daily values of
// gobees-fast. It is the output of the comparisons (go/compare) and of the command line (-o).
//
// The column names stay the same in all steps of gobees-fast: the comparison matches the columns by name.
// The first 15 columns (ticks and 14 totals) are those of DailyOutput, so that a version without the further
// columns (e.g. gobees) can be compared on them. Then follow the columns of NetLogoMetrics, then, all at the end of
// the day:
//
//   - ForagerDeathsToday: bees that died on a foraging trip today
//   - TotalAdultQueensEverProduced: queens emerged from pupae since the start
//   - MeanCompetitionPointDate: mean competition point date (ticks) of the living colonies that have one, 0 if none
type DailyOutputExtended struct {
	daily        DailyOutput
	totals       ecs.Resource[globals.Totals]
	counters     ecs.Resource[globals.Counters]
	deadColonies ecs.Resource[globals.DeadColonies]
	calendar     ecs.Resource[globals.Calendar]
	bees         *ecs.Filter1[comp.Bee]
	colonies     *ecs.Filter2[comp.Colony, comp.ColonyStores]
	foodsources  *ecs.Filter1[comp.Foodsource]
	stocks       *ecs.Filter1[comp.FoodsourceState]
	badgers      *ecs.Filter1[comp.Badger]
	values       []float64
	// lastForagerDeaths is the counter of foragers that died on a trip at the end of the previous day.
	lastForagerDeaths float64
}

func (o *DailyOutputExtended) Initialize(w *ecs.World) {
	o.daily.Initialize(w)
	o.totals = ecs.NewResource[globals.Totals](w)
	o.counters = ecs.NewResource[globals.Counters](w)
	o.deadColonies = ecs.NewResource[globals.DeadColonies](w)
	o.calendar = ecs.NewResource[globals.Calendar](w)
	o.bees = ecs.NewFilter1[comp.Bee](w)
	o.colonies = ecs.NewFilter2[comp.Colony, comp.ColonyStores](w)
	o.foodsources = ecs.NewFilter1[comp.Foodsource](w)
	o.stocks = ecs.NewFilter1[comp.FoodsourceState](w)
	o.badgers = ecs.NewFilter1[comp.Badger](w)
	o.values = make([]float64, len(o.Header()))
}

func (o *DailyOutputExtended) Update(w *ecs.World) {}

func (o *DailyOutputExtended) Header() []string {
	h := append(o.daily.Header(), NetLogoMetrics...)
	return append(h, "ForagerDeathsToday", "TotalAdultQueensEverProduced", "MeanCompetitionPointDate")
}

func (o *DailyOutputExtended) Values(w *ecs.World) []float64 {
	n := copy(o.values, o.daily.Values(w))

	var eggs, larvae, pupae, workers, males, youngQueens, hibernating, beeAgents, diploidMales, psithyrus float64
	bq := o.bees.Query()
	for bq.Next() {
		b := bq.Get()
		beeAgents++
		switch b.Stage {
		case stage.Egg:
			eggs += b.Number
		case stage.Larva:
			larvae += b.Number
		case stage.Pupa:
			pupae += b.Number
		case stage.Adult:
			switch {
			case b.Caste == caste.Worker:
				workers += b.Number
			case b.Caste == caste.Male:
				males += b.Number
			case b.Caste == caste.Queen && !b.Mated:
				youngQueens += b.Number
			}
		}
		if b.Activity == activity.Hibernate {
			hibernating += b.Number
		}
		if b.Caste == caste.Male && b.Ploidy == 2 {
			diploidMales += b.Number
		}
		if b.SpeciesName == "Psithyrus" {
			psithyrus += b.Number
		}
	}

	var colonies, ibmColonies, energy, pollen, cpSum, cpCount float64
	perSpecies := make([]float64, len(netLogoSpecies))
	cq := o.colonies.Query()
	for cq.Next() {
		colony, stores := cq.Get()
		colonies++
		if !colony.CohortBased {
			ibmColonies++
		}
		for i, s := range netLogoSpecies {
			if colony.SpeciesNameColony == s {
				perSpecies[i]++
			}
		}
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

	var nectar, pollenLeft float64
	sq := o.stocks.Query()
	for sq.Next() {
		s := sq.Get()
		nectar += s.Nectar_myl
		pollenLeft += s.Pollen_g
	}
	fq := o.foodsources.Query()
	foodsources := float64(fq.Count())
	fq.Close()
	badq := o.badgers.Query()
	badgers := float64(badq.Count())
	badq.Close()

	counters := o.counters.Get()
	foragerDeaths := counters.TotalForagersEverDiedForaging - o.lastForagerDeaths
	o.lastForagerDeaths = counters.TotalForagersEverDiedForaging

	values := []float64{
		colonies, float64(len(o.deadColonies.Get().List)), counters.TotalColoniesEverProduced,
		eggs, larvae, pupae, workers, males, youngQueens, hibernating,
		energy, pollen, o.totals.Get().TotalForagingTripsToday, nectar / 1000000, pollenLeft / 1000,
		badgers, foodsources, ibmColonies, beeAgents, diploidMales,
	}
	values = append(values, perSpecies...)
	values = append(values, psithyrus, o.calendar.Get().DailyForagingPeriod_s, counters.TotalHibernatingQueensEverRemoved,
		foragerDeaths, counters.TotalAdultQueensEverProduced, meanCP)
	copy(o.values[n:], values)
	return o.values
}
