package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/globals"
	"gobees/params"

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

// colonyFounder creates colonies and their foodsource lists.
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

// found creates one colony per bee in "nestConstruction".
func (c *colonyFounder) found(w *ecs.World) {
	// each new colony takes one-of the remaining queens: shuffle the queens, pair them in order.
	queens := []ecs.Entity{}
	query := c.beeFilter.Query()
	for query.Next() {
		if query.Get().Activity == activity.NestConstruction {
			queens = append(queens, query.Entity())
		}
	}
	shuffle(c.rng, queens)

	for _, q := range queens {
		qPos, qBee := c.beeMap.Get(q)
		pos := *qPos
		speciesID := qBee.Species

		// All colonies are cohort-based.
		// The four dates start unset (NotSetHigh in NetLogo), the zero value of day.Date.
		colony := comp.Colony{
			Species:         speciesID,
			QueenProduction: false,
		}
		stores := comp.ColonyStores{Energy: c.constants.Colony.StartNectar * c.constants.Foraging.SucroseEnergy *
			c.constants.Colony.StartConcentration}
		counts := comp.ColonyCounts{}
		e := c.colonyMap.NewEntity(&pos, &colony, &stores, &comp.ColonyStimuli{}, &counts, &comp.ColonyRange{})

		// the queen joins the colony (get her again, creating the colony may move components)
		_, qBee = c.beeMap.Get(q)
		qBee.Colony = e
		qBee.Activity = activity.Resting

		c.patchesInRange(e)
		c.flowersInRange(e)

		// The observers count the new colonies (ecs.OnCreateEntity, obs.Counters).
	}
}

// patchesInRange sets allPatchesInRangeList of a colony: all foodsources (every layer) within Constants.MaxForagingRange.
func (c *colonyFounder) patchesInRange(colony ecs.Entity) {
	pos := *c.positionMap.Get(colony)
	maxDist := c.constants.Foraging.MaxForagingRange * c.landscape.Scale

	// ORDER: MATTERS for the list order, which flowersInRange keeps.
	// NetLogo builds the list with fput in random ask order. Here: shuffle, then append. Both give a list in random
	// order.
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
	shuffle(c.rng, all)
	list := []ecs.Entity{}
	for _, it := range all {
		if distanceBetween(pos.X, pos.Y, it.x, it.y) <= maxDist {
			list = append(list, it.e)
		}
	}
	c.rangeMap.Get(colony).Layers = list
}

// flowersInRange sets the lists of masterpatches in range that offer pollen or nectar now.
func (c *colonyFounder) flowersInRange(colony ecs.Entity) {
	r := c.rangeMap.Get(colony)
	polM, necM := []ecs.Entity{}, []ecs.Entity{}
	// ORDER: irrelevant, foreach over a list.
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
