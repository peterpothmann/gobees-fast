package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/globals"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// BadgersOnTheProwl [GO-17].
//
// Every badger digs up each colony closer than 735 m with probability 0.19 * 0.1 = 0.019 per day: the colony
// loses its stores and all its bees. The colony itself dies the next morning in [UCOL-11].
// Not translated: color (GUI).
//
// NetLogo: BadgersOnTheProwlProc.
type BadgersOnTheProwl struct {
	rng       *rand.Rand
	landscape *globals.Landscape
	death     beeDeath

	badgerFilter *ecs.Filter2[comp.Position, comp.Badger]
	colonyFilter *ecs.Filter2[comp.Position, comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map2[comp.ColonyStores, comp.ColonyCounts]
	beeMap       *ecs.Map1[comp.Bee]
}

func (s *BadgersOnTheProwl) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.death = newBeeDeath(w)
	s.badgerFilter = ecs.NewFilter2[comp.Position, comp.Badger](w)
	s.colonyFilter = ecs.NewFilter2[comp.Position, comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap2[comp.ColonyStores, comp.ColonyCounts](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *BadgersOnTheProwl) Update(w *ecs.World) {
	// variables, not Go constants: NetLogo multiplies two doubles, 0.19 * 0.1 = 0.019000000000000003, whereas Go
	// would multiply constants exactly (0.019).
	foragingRange_m := 735.0 // Kruuk & Parish 1982: territory ca. 170 ha, radius ca. 735 m
	encounterProb := 0.19    // Kowalczyk et al. 2006: daily range 19 % of the home range
	digUpProb := 0.1         // ARBITRARY VALUE (NetLogo comment)

	// [BADG-01] ORDER: irrelevant for the state. A colony dug up twice in one evening has nothing left to lose.
	type badger struct{ x, y float64 }
	badgers := []badger{}
	bq := s.badgerFilter.Query()
	for bq.Next() {
		pos, _ := bq.Get()
		badgers = append(badgers, badger{pos.X, pos.Y})
	}
	if len(badgers) == 0 {
		return
	}
	util.Shuffle(s.rng, badgers)

	// the members of each colony, grouped once. Bees that died on the way are skipped with w.Alive.
	beesOf := beesByColony(s.beeFilter, func(b *comp.Bee) bool { return true })

	for _, b := range badgers {
		// colonies with distancexy < range, strict, in NetLogo patch units, no wrapping
		inRange := []ecs.Entity{}
		cq := s.colonyFilter.Query()
		for cq.Next() {
			pos, _ := cq.Get()
			if util.Distance(pos.X, pos.Y, b.x, b.y) < s.landscape.Scaling_NLpatches_per_m*foragingRange_m {
				inRange = append(inRange, cq.Entity())
			}
		}
		util.Shuffle(s.rng, inRange)

		for _, col := range inRange {
			if s.rng.Float64() >= encounterProb*digUpProb { // NetLogo: random-float 1 < encounterProb * digUpProb
				continue
			}
			stores, _ := s.colonyMap.Get(col)
			stores.EnergyStore_kJ = 0
			stores.PollenStore_g = 0
			// every bee of the colony dies, in random order. Each bee dies independently.
			victims := []ecs.Entity{}
			for _, e := range beesOf[col] {
				if w.Alive(e) {
					victims = append(victims, e)
				}
			}
			util.Shuffle(s.rng, victims)
			memoBroodDeaths := 0.0
			for _, e := range victims {
				if bee := s.beeMap.Get(e); bee.Brood {
					memoBroodDeaths += bee.Number
				}
				s.death.DieProc(w, e, "Colony killed by badger!")
			}
			_, counts := s.colonyMap.Get(col)
			counts.BroodDeathBadger += memoBroodDeaths
		}
	}
}

func (s *BadgersOnTheProwl) Finalize(w *ecs.World) {}
