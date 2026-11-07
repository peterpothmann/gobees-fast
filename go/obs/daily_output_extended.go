package obs

import (
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// NetLogoMetrics are the columns of DailyOutputExtended with the metrics of the NetLogo experiments in
// go/validation/experiments: same names and definitions as the NetLogo reporters in the experiment
// files. go/compare compares them with the NetLogo reference of each scenario.
var NetLogoMetrics = []string{
	"colonies", "dead_colonies", "colonies_founded", "eggs", "larvae", "pupae", "workers", "males", "young_queens",
	"hibernating_queens", "energy_kJ", "pollen_g", "trips", "nectar_left_l", "pollen_left_kg",
	"badgers", "foodsources", "bee_agents", "diploid_males",
	"colonies_B_terrestris", "colonies_B_pascuorum", "colonies_B_lapidarius", "colonies_B_hortorum",
	"colonies_B_pratorum", "colonies_B_hypnorum", "psithyrus",
	"foraging_period_s", "queens_removed",
}

// netLogoSpecies are the Bombus species of the metrics colonies_<species>.
var netLogoSpecies = []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum"}

// DailyOutputExtended observes the columns of DailyOutput, the NetLogo metrics and further daily values of
// gobees-fast. It is the output of the comparisons (go/compare) and of the command line (-o).
//
// The comparison matches the columns by name. A column is removed only together with its model branch, and is
// then listed in REMOVED_COLUMNS of go/compare/commits.env (e.g. TotalIBMColonies and ibm_colonies).
// The first 14 columns (ticks and 13 totals) are those of DailyOutput, so that a version without the further
// columns (e.g. gobees) can be compared on them. Then follow the columns of NetLogoMetrics, then, all at the end of
// the day:
//
//   - ForagerDeathsToday: bees that died on a foraging trip today
//   - TotalAdultQueensEverProduced: queens emerged from pupae since the start
//   - MeanCompetitionPointDate: mean competition point date (ticks) of the living colonies that have one, 0 if none
//
// Update must run once at the end of every day (reporter.CSV calls it on every tick), for the counters.
//
// Columns selects the columns written (see SelectColumns). Every value is computed on every day, also those of
// columns not selected, because some are differences to the previous day.
type DailyOutputExtended struct {
	// Columns are the names of the columns written, in this order. Set it to the result of SelectColumns before
	// Initialize. nil: all columns.
	Columns []string

	daily       DailyOutput
	counters    Counters
	calendar    ecs.Resource[globals.Calendar]
	bees        *ecs.Filter2[comp.Bee, comp.Genetics]
	colonies    *ecs.Filter2[comp.Colony, comp.ColonyStores]
	input       ecs.Resource[params.Input] // the species
	foodsources *ecs.Filter1[comp.Foodsource]
	stocks      *ecs.Filter1[comp.FoodsourceState]
	badgers     *ecs.Filter1[comp.Badger]
	values      []float64 // all columns of the day
	selected    []int     // index in values of each column written, nil: all
	out         []float64 // the columns written
	// lastForagerDeaths is the counter of foragers that died on a trip at the end of the previous day.
	lastForagerDeaths float64
}

func (o *DailyOutputExtended) Initialize(w *ecs.World) {
	o.daily.Initialize(w)
	o.counters.Initialize(w)
	o.calendar = ecs.NewResource[globals.Calendar](w)
	o.bees = ecs.NewFilter2[comp.Bee, comp.Genetics](w)
	o.colonies = ecs.NewFilter2[comp.Colony, comp.ColonyStores](w)
	o.input = ecs.NewResource[params.Input](w)
	o.foodsources = ecs.NewFilter1[comp.Foodsource](w)
	o.stocks = ecs.NewFilter1[comp.FoodsourceState](w)
	o.badgers = ecs.NewFilter1[comp.Badger](w)
	all := AllColumns()
	o.values = make([]float64, len(all))
	if len(o.Columns) > 0 {
		o.selected = make([]int, len(o.Columns))
		for i, name := range o.Columns {
			o.selected[i] = slices.Index(all, name)
			if o.selected[i] < 0 {
				panic("obs: unknown output column " + name)
			}
		}
		o.out = make([]float64, len(o.Columns))
	}
}

func (o *DailyOutputExtended) Update(w *ecs.World) {
	o.counters.Update(w)
}

// Header returns the names of the columns written: Columns, or all columns.
func (o *DailyOutputExtended) Header() []string {
	if len(o.Columns) > 0 {
		return slices.Clone(o.Columns)
	}
	return AllColumns()
}

// Values returns the values of the columns written, in the order of Header.
func (o *DailyOutputExtended) Values(w *ecs.World) []float64 {
	all := o.allValues(w)
	if o.selected == nil {
		return all
	}
	for i, j := range o.selected {
		o.out[i] = all[j]
	}
	return o.out
}

// allValues returns the values of all columns, in the order of AllColumns.
func (o *DailyOutputExtended) allValues(w *ecs.World) []float64 {
	n := copy(o.values, o.daily.Values(w))

	// the species with Parasitic (Psithyrus), usually 0 or 1
	species := o.input.Get().Species
	var parasitic []int
	for i := range species {
		if species[i].Parasitic {
			parasitic = append(parasitic, i)
		}
	}

	var eggs, larvae, pupae, workers, males, youngQueens, hibernating, beeAgents, diploidMales, psithyrus float64
	bq := o.bees.Query()
	for bq.Next() {
		b, g := bq.Get()
		beeAgents++
		switch b.Stage {
		case stage.Egg:
			eggs += float64(b.Number)
		case stage.Larva:
			larvae += float64(b.Number)
		case stage.Pupa:
			pupae += float64(b.Number)
		case stage.Adult:
			switch {
			case b.Caste == caste.Worker:
				workers += float64(b.Number)
			case b.Caste == caste.Male:
				males += float64(b.Number)
			case b.Caste == caste.Queen && !b.Mated:
				youngQueens += float64(b.Number)
			}
		}
		if b.Activity == activity.Hibernate {
			hibernating += float64(b.Number)
		}
		// NetLogo: ploidy = 2
		if b.Caste == caste.Male && len(g.Alleles) == 2 {
			diploidMales += float64(b.Number)
		}
		// NetLogo: speciesName = "Psithyrus"
		if slices.Contains(parasitic, b.Species) {
			psithyrus += float64(b.Number)
		}
	}

	var colonies, energy, pollen, cpSum, cpCount float64
	perSpecies := make([]float64, len(netLogoSpecies))
	cq := o.colonies.Query()
	for cq.Next() {
		colony, stores := cq.Get()
		colonies++
		name := species[colony.Species].Name
		for i, s := range netLogoSpecies {
			if name == s {
				perSpecies[i]++
			}
		}
		energy += stores.Energy
		pollen += stores.Pollen
		if colony.CompetitionPointDate.IsSet() {
			cpSum += float64(colony.CompetitionPointDate.Day())
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
		nectar += s.Nectar
		pollenLeft += s.Pollen
	}
	fq := o.foodsources.Query()
	foodsources := float64(fq.Count())
	fq.Close()
	badq := o.badgers.Query()
	badgers := float64(badq.Count())
	badq.Close()

	counters := &o.counters
	foragerDeaths := counters.TotalForagersEverDiedForaging - o.lastForagerDeaths
	o.lastForagerDeaths = counters.TotalForagersEverDiedForaging

	values := []float64{
		colonies, float64(counters.DeadColonies), counters.TotalColoniesEverProduced,
		eggs, larvae, pupae, workers, males, youngQueens, hibernating,
		energy, pollen, counters.TotalForagingTripsToday, nectar / 1000000, pollenLeft / 1000,
		badgers, foodsources, beeAgents, diploidMales,
	}
	values = append(values, perSpecies...)
	values = append(values, psithyrus, o.calendar.Get().ForagingPeriod, float64(counters.TotalHibernatingQueensEverRemoved),
		foragerDeaths, counters.TotalAdultQueensEverProduced, meanCP)
	copy(o.values[n:], values)
	return o.values
}
