package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// colonyMapper creates colony entities with all their components.
type colonyMapper = ecs.Map7[
	comp.Who,
	comp.Position,
	comp.Colony,
	comp.ColonyStores,
	comp.ColonyStimuli,
	comp.ColonyCounts,
	comp.ColonyRange,
]

// colonyFounder creates colonies [CCOL] and their foodsource lists [PRNG], [FRNG].
type colonyFounder struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	file      *params.ParameterFile
	constants *params.Constants
	landscape *globals.Landscape
	counters  *globals.Counters
	who       *globals.WhoCounter

	beeFilter        *ecs.Filter1[comp.Bee]
	foodsourceFilter *ecs.Filter2[comp.Position, comp.Foodsource]
	beeMap           *ecs.Map2[comp.Position, comp.Bee]
	colonyMap        *colonyMapper
	positionMap      *ecs.Map1[comp.Position]
	rangeMap         *ecs.Map1[comp.ColonyRange]
	foodsourceMap    *ecs.Map2[comp.FoodsourceState, comp.Layer]
}

func newColonyFounder(w *ecs.World, rng *rand.Rand) colonyFounder {
	return colonyFounder{
		rng:              rng,
		ticks:            ecs.GetResource[globals.Ticks](w),
		file:             ecs.GetResource[params.ParameterFile](w),
		constants:        ecs.GetResource[params.Constants](w),
		landscape:        ecs.GetResource[globals.Landscape](w),
		counters:         ecs.GetResource[globals.Counters](w),
		who:              ecs.GetResource[globals.WhoCounter](w),
		beeFilter:        ecs.NewFilter1[comp.Bee](w),
		foodsourceFilter: ecs.NewFilter2[comp.Position, comp.Foodsource](w),
		beeMap:           ecs.NewMap2[comp.Position, comp.Bee](w),
		colonyMap:        ecs.NewMap7[comp.Who, comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w),
		positionMap:      ecs.NewMap1[comp.Position](w),
		rangeMap:         ecs.NewMap1[comp.ColonyRange](w),
		foodsourceMap:    ecs.NewMap2[comp.FoodsourceState, comp.Layer](w),
	}
}

// CreateColoniesProc creates one colony per bee in "nestConstruction" [CCOL].
// Not translated: the store bars [CCOL-01], [CCOL-07] (GUI).
func (c *colonyFounder) CreateColoniesProc(w *ecs.World) {
	// [CCOL-02] each new colony takes one-of the remaining queens: shuffle the queens, pair them in order.
	queens := []ecs.Entity{}
	query := c.beeFilter.Query()
	for query.Next() {
		if query.Get().Activity == activity.NestConstruction {
			queens = append(queens, query.Entity())
		}
	}
	util.Shuffle(c.rng, queens)

	for _, q := range queens {
		qPos, qBee := c.beeMap.Get(q)
		pos := *qPos
		speciesID := qBee.SpeciesID

		// [CCOL-05] not translated: gobees-fast has only cohort-based colonies (finding 30, see
		// trans/netlogo-findings.md#finding-30).
		// [CCOL-04]
		colony := comp.Colony{
			SpeciesIDcolony:      speciesID,
			QueenProduction:      false,
			QueenProductionDate:  params.NotSetHigh,
			SwitchPointDate:      params.NotSetHigh,
			CompetitionPointDate: params.NotSetHigh,
			EusocialPhaseDate:    params.NotSetHigh,
		}
		stores := comp.ColonyStores{EnergyStore_kJ: 100 * c.constants.EnergySucrose_kJ_per_mymol * 1.5}
		counts := comp.ColonyCounts{}
		e := c.colonyMap.NewEntity(&comp.Who{Who: c.who.Next()}, &pos, &colony, &stores, &comp.ColonyStimuli{}, &counts, &comp.ColonyRange{})

		// [CCOL-03] the queen joins the colony (get her again, creating the colony may move components)
		_, qBee = c.beeMap.Get(q)
		qBee.ColonyID = e
		qBee.Activity = activity.Resting

		// [CCOL-06]
		c.PatchesInRangeProc(e)
		c.FoodsourcesInFlowerAndRangeProc(e)

		// [CCOL-08]
		c.counters.TotalColoniesEverProduced++
	}
}

// PatchesInRangeProc sets allPatchesInRangeList of a colony: all foodsources (every layer) within ForagingRangeMax_m [PRNG].
// masterpatchesInRangeList is not translated (finding 23, see trans/netlogo-findings.md#finding-23).
func (c *colonyFounder) PatchesInRangeProc(colony ecs.Entity) {
	pos := *c.positionMap.Get(colony)
	maxDist := c.constants.ForagingRangeMax_m * c.landscape.Scaling_NLpatches_per_m

	// [PRNG-01] ORDER: MATTERS for the list order (finding 29, see trans/netlogo-findings.md#finding-29).
	// NetLogo: fput in random ask order. Here: shuffle, then append. Both give a list in random order.
	type item struct {
		e    ecs.Entity
		x, y float64
	}
	all := []item{}
	query := c.foodsourceFilter.Query()
	for query.Next() {
		p, _ := query.Get()
		all = append(all, item{query.Entity(), p.X, p.Y})
	}
	util.Shuffle(c.rng, all)
	list := []ecs.Entity{}
	for _, it := range all {
		if util.Distance(pos.X, pos.Y, it.x, it.y) <= maxDist {
			list = append(list, it.e)
		}
	}
	c.rangeMap.Get(colony).AllPatchesInRangeList = list
}

// FoodsourcesInFlowerAndRangeProc sets the lists of masterpatches in range that offer pollen or nectar now [FRNG].
// Not translated: pollenInFlowerAndRangeList, nectarInFlowerAndRangeList and allSourcesInFlowerAndRangeList (never
// read, removed in step F3).
func (c *colonyFounder) FoodsourcesInFlowerAndRangeProc(colony ecs.Entity) {
	r := c.rangeMap.Get(colony)
	polM, necM := []ecs.Entity{}, []ecs.Entity{}
	// [FRNG-01] ORDER: irrelevant, foreach over a list.
	for _, fs := range r.AllPatchesInRangeList {
		state, layer := c.foodsourceMap.Get(fs)
		if state.Pollen_g > 0 {
			polM = append(polM, layer.MasterpatchID)
		}
		if state.Nectar_myl > 0 {
			necM = append(necM, layer.MasterpatchID)
		}
	}
	r.MasterpatchesWithPollenlayersInFlowerAndRangeList = removeDuplicates(polM)
	r.MasterpatchesWithNectarlayersInFlowerAndRangeList = removeDuplicates(necM)
}
