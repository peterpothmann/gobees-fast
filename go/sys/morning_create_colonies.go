package sys

import (
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
	ticks      *globals.Ticks
	parameters *params.Parameters
	constants  *params.Constants
	landscape  *globals.Landscape

	beeFilter   *ecs.Filter1[comp.Bee]
	patchFilter *ecs.Filter2[comp.Position, comp.FlowerPatch]
	beeMap      *ecs.Map2[comp.Position, comp.Bee]
	colonyMap   *colonyMapper
	positionMap *ecs.Map1[comp.Position]
	rangeMap    *ecs.Map1[comp.ColonyRange]
	patchMap    *ecs.Map1[comp.FlowerPatch]
}

func newColonyFounder(w *ecs.World) colonyFounder {
	return colonyFounder{
		ticks:       ecs.GetResource[globals.Ticks](w),
		parameters:  ecs.GetResource[params.Parameters](w),
		constants:   ecs.GetResource[params.Constants](w),
		landscape:   ecs.GetResource[globals.Landscape](w),
		beeFilter:   ecs.NewFilter1[comp.Bee](w),
		patchFilter: ecs.NewFilter2[comp.Position, comp.FlowerPatch](w),
		beeMap:      ecs.NewMap2[comp.Position, comp.Bee](w),
		colonyMap:   ecs.NewMap6[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w),
		positionMap: ecs.NewMap1[comp.Position](w),
		rangeMap:    ecs.NewMap1[comp.ColonyRange](w),
		patchMap:    ecs.NewMap1[comp.FlowerPatch](w),
	}
}

// found creates one colony per bee in "nestConstruction".
func (c *colonyFounder) found(w *ecs.World) {
	// each new colony takes one-of the remaining queens. ORDER: irrelevant, every queen founds her own colony and no
	// site is taken. Not shuffled (NetLogo: one-of, random order).
	queens := []ecs.Entity{}
	query := c.beeFilter.Query()
	for query.Next() {
		if query.Get().Activity == activity.NestConstruction {
			queens = append(queens, query.Entity())
		}
	}

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

// patchesInRange sets allPatchesInRangeList of a colony: all flower patches within Constants.MaxForagingRange.
// NetLogo: all foodsources (every layer) in range. The layers of a flower patch share its position, so a flower
// patch is in range exactly if its layers are.
func (c *colonyFounder) patchesInRange(colony ecs.Entity) {
	pos := *c.positionMap.Get(colony)
	maxDist := c.constants.Foraging.MaxForagingRange * c.landscape.Scale

	// ORDER: irrelevant. flowersInRange keeps the list order in PollenPatches and NectarPatches, which only the search
	// of a forager reads, and the search chooses at random among the detected patches in any order
	// (forager.search). Storage order (NetLogo: fput in random ask order).
	list := []ecs.Entity{}
	query := c.patchFilter.Query()
	for query.Next() {
		if p, _ := query.Get(); distanceBetween(pos.X, pos.Y, p.X, p.Y) <= maxDist {
			list = append(list, query.Entity())
		}
	}
	c.rangeMap.Get(colony).Patches = list
}

// flowersInRange sets the lists of flower patches in range that offer pollen or nectar now. It runs in the morning
// after flowerRefill, before any bee forages, so the flags of the flower patches hold.
// NetLogo: the masterpatches of the layers in range with food, without duplicates. The layers of a flower patch
// are stored together, so the flower patches keep the order of their first layer.
func (c *colonyFounder) flowersInRange(colony ecs.Entity) {
	r := c.rangeMap.Get(colony)
	polM, necM := []ecs.Entity{}, []ecs.Entity{}
	// ORDER: irrelevant, foreach over a list.
	for _, e := range r.Patches {
		fp := c.patchMap.Get(e)
		if fp.PollenInFlower {
			polM = append(polM, e)
		}
		if fp.NectarInFlower {
			necM = append(necM, e)
		}
	}
	r.PollenPatches = polM
	r.NectarPatches = necM
	r.PollenDetection, r.NectarDetection = nil, nil // new lists: the search computes them again
}
