// Package obs holds observers, which read model output from the world. They implement the observer interfaces of
// ark-tools and are written to files by the reporters of ark-tools, e.g. reporter.CSV.
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

// outputDailyColumns are the columns of the totals of OutputDailyProc (globals.DailyOutput) [GO-18], [INIT-12]:
// ticks, then the totals under the NetLogo names of the globals.
var outputDailyColumns = []string{
	"ticks",
	"TotalIBMColonies",
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

// NetLogoMetrics are the columns of DailyOutput with the metrics of the NetLogo experiments in
// go/validation/experiments (step7 to step8): same names and definitions as the NetLogo reporters in the experiment
// files and as the metrics of the validation tests (e.g. model/validation_step7_test.go).
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

// DailyOutput observes the daily output of the command line (-o). One row per day:
//
//   - ticks and the totals of OutputDailyProc (globals.DailyOutput) [GO-18], [INIT-12], under the NetLogo names of
//     the globals
//   - the metrics of the NetLogo experiments (NetLogoMetrics), under the names of their reporters
//   - TotalAdultQueensEverProduced: queens emerged from pupae since the start
//   - MeanCompetitionPointDate: mean competition point date (ticks) of the living colonies that have one, 0 if none
//
// All at the end of the day. reporter.CSV adds its own first column t, which counts its updates from 0 and is
// therefore ticks - 1 (see globals.Ticks). The column ticks gives the NetLogo day.
//
// gobees-fast writes the same columns (plus ForagerDeathsToday) and compares its steps with gobees on them, matched
// by name, so the names must not change. DailyOutput only reads the world: the runs are the same with and without it.
type DailyOutput struct {
	ticks        ecs.Resource[globals.Ticks]
	output       ecs.Resource[globals.DailyOutput]
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
}

func (o *DailyOutput) Initialize(w *ecs.World) {
	o.ticks = ecs.NewResource[globals.Ticks](w)
	o.output = ecs.NewResource[globals.DailyOutput](w)
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

func (o *DailyOutput) Update(w *ecs.World) {}

func (o *DailyOutput) Header() []string {
	h := append(append([]string{}, outputDailyColumns...), NetLogoMetrics...)
	return append(h, "TotalAdultQueensEverProduced", "MeanCompetitionPointDate")
}

func (o *DailyOutput) Values(w *ecs.World) []float64 {
	d := o.output.Get()
	o.values[0] = float64(o.ticks.Get().Ticks)
	o.values[1] = d.TotalIBMColonies
	o.values[2] = d.TotalQueens
	o.values[3] = d.TotalMatedQueens
	o.values[4] = d.TotalUnmatedQueens
	o.values[5] = d.TotalHibernatingQueens
	o.values[6] = d.TotalColonies
	o.values[7] = d.TotalBeeAgents
	o.values[8] = d.TotalMales
	o.values[9] = d.TotalActiveBees
	o.values[10] = d.MeanWorkerWeight_mg
	o.values[11] = d.MeanQueenWeight_mg
	o.values[12] = d.MeanAdultWeight_mg
	o.values[13] = d.TotalMapArea_km2
	o.values[14] = d.ColonyDensity_km2
	n := len(outputDailyColumns)

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
	values := []float64{
		colonies, float64(len(o.deadColonies.Get().List)), counters.TotalColoniesEverProduced,
		eggs, larvae, pupae, workers, males, youngQueens, hibernating,
		energy, pollen, o.totals.Get().TotalForagingTripsToday, nectar / 1000000, pollenLeft / 1000,
		badgers, foodsources, ibmColonies, beeAgents, diploidMales,
	}
	values = append(values, perSpecies...)
	values = append(values, psithyrus, o.calendar.Get().DailyForagingPeriod_s, counters.TotalHibernatingQueensEverRemoved,
		counters.TotalAdultQueensEverProduced, meanCP)
	copy(o.values[n:], values)
	return o.values
}
