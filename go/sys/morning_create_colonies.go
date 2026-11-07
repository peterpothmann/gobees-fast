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
type colonyMapper = ecs.Map6[
	comp.Position,
	comp.Colony,
	comp.ColonyStores,
	comp.ColonyStimuli,
	comp.ColonyCounts,
	comp.ColonyRange,
]

// colonyFounder creates colonies [CCOL] and their foodsource lists [PRNG], [FRNG].
type colonyFounder struct {
	rng        *rand.Rand
	ticks      *globals.Ticks
	parameters *params.Parameters
	constants  *params.Constants
	landscape  *globals.Landscape

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
		parameters:       ecs.GetResource[params.Parameters](w),
		constants:        ecs.GetResource[params.Constants](w),
		landscape:        ecs.GetResource[globals.Landscape](w),
		beeFilter:        ecs.NewFilter1[comp.Bee](w),
		foodsourceFilter: ecs.NewFilter2[comp.Position, comp.Foodsource](w),
		beeMap:           ecs.NewMap2[comp.Position, comp.Bee](w),
		colonyMap:        ecs.NewMap6[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w),
		positionMap:      ecs.NewMap1[comp.Position](w),
		rangeMap:         ecs.NewMap1[comp.ColonyRange](w),
		foodsourceMap:    ecs.NewMap2[comp.FoodsourceState, comp.Layer](w),
	}
}

// found creates one colony per bee in "nestConstruction" [CCOL].
// Not translated: the store bars [CCOL-01], [CCOL-07] (GUI).
func (c *colonyFounder) found(w *ecs.World) {
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
		speciesID := qBee.Species

		// [CCOL-05] not translated: gobees-fast has only cohort-based colonies (finding 30, see
		// trans/netlogo-findings.md#finding-30).
		// [CCOL-04]
		// The four dates start unset (NetLogo: NotSetHigh), the zero value of day.Date.
		colony := comp.Colony{
			Species:         speciesID,
			QueenProduction: false,
		}
		stores := comp.ColonyStores{Energy: 100 * c.constants.SucroseEnergy * 1.5}
		counts := comp.ColonyCounts{}
		e := c.colonyMap.NewEntity(&pos, &colony, &stores, &comp.ColonyStimuli{}, &counts, &comp.ColonyRange{})

		// [CCOL-03] the queen joins the colony (get her again, creating the colony may move components)
		_, qBee = c.beeMap.Get(q)
		qBee.Colony = e
		qBee.Activity = activity.Resting

		// [CCOL-06]
		c.patchesInRange(e)
		c.flowersInRange(e)

		// [CCOL-08] totalColoniesEverProduced: output and [CHK-04] only, removed in step F4. The observers count
		// the new colonies (ecs.OnCreateEntity, obs.Counters).
	}
}

// patchesInRange sets allPatchesInRangeList of a colony: all foodsources (every layer) within Constants.MaxForagingRange [PRNG].
// masterpatchesInRangeList is not translated (finding 23, see trans/netlogo-findings.md#finding-23).
func (c *colonyFounder) patchesInRange(colony ecs.Entity) {
	pos := *c.positionMap.Get(colony)
	maxDist := c.constants.MaxForagingRange * c.landscape.Scale

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
	c.rangeMap.Get(colony).Layers = list
}

// flowersInRange sets the lists of masterpatches in range that offer pollen or nectar now [FRNG].
// Not translated: pollenInFlowerAndRangeList, nectarInFlowerAndRangeList and allSourcesInFlowerAndRangeList (never
// read, removed in step F3).
func (c *colonyFounder) flowersInRange(colony ecs.Entity) {
	r := c.rangeMap.Get(colony)
	polM, necM := []ecs.Entity{}, []ecs.Entity{}
	// [FRNG-01] ORDER: irrelevant, foreach over a list.
	for _, fs := range r.Layers {
		state, layer := c.foodsourceMap.Get(fs)
		if state.Pollen > 0 {
			polM = append(polM, layer.Masterpatch)
		}
		if state.Nectar > 0 {
			necM = append(necM, layer.Masterpatch)
		}
	}
	r.PollenPatches = removeDuplicates(polM)
	r.NectarPatches = removeDuplicates(necM)
}
