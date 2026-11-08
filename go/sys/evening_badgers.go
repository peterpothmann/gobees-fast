package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// badgers lets every badger dig up each colony closer than Constants.Badgers.Range (735 m) with probability
// EncounterProb * DigUpProb (0.19 * 0.1 = 0.019) per day: the
// colony loses its stores and all its bees. The colony itself dies the next morning in colonyUpdate.
//
// NetLogo: BadgersOnTheProwlProc.
type badgers struct {
	rng       *rand.Rand
	landscape *globals.Landscape
	death     beeDeath
	constants *params.Constants

	badgerFilter *ecs.Filter2[comp.Position, comp.Badger]
	colonyFilter *ecs.Filter2[comp.Position, comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map1[comp.ColonyStores]

	beesOf map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *badgers) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.death = newBeeDeath(w)
	s.badgerFilter = ecs.NewFilter2[comp.Position, comp.Badger](w)
	s.colonyFilter = ecs.NewFilter2[comp.Position, comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap1[comp.ColonyStores](w)
}

func (s *badgers) Update(w *ecs.World) {
	// variables, not Go constants: NetLogo multiplies two doubles, 0.19 * 0.1 = 0.019000000000000003, whereas Go
	// would multiply constants exactly (0.019).
	foragingRange := s.constants.Badgers.Range         // [m], Kruuk & Parish 1982: territory ca. 170 ha, radius ca. 735 m
	encounterProb := s.constants.Badgers.EncounterProb // Kowalczyk et al. 2006: daily range 19 % of the home range
	digUpProb := s.constants.Badgers.DigUpProb

	// ORDER: irrelevant. Each badger and colony in range draws once, a colony dug up twice in one evening has nothing
	// left to lose, and all its bees die. Badgers, colonies and victims are not shuffled (NetLogo: random order).
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

	// the members of each colony, grouped once. Bees that died on the way are skipped with w.Alive.
	s.beesOf = beesByColony(s.beesOf, s.beeFilter, func(b *comp.Bee) bool { return true })

	for _, b := range badgers {
		// colonies closer than the range (strict), in patch units, no wrapping, as in NetLogo
		inRange := []ecs.Entity{}
		cq := s.colonyFilter.Query()
		for cq.Next() {
			pos, _ := cq.Get()
			if distanceBetween(pos.X, pos.Y, b.x, b.y) < s.landscape.Scale*foragingRange {
				inRange = append(inRange, cq.Entity())
			}
		}

		for _, col := range inRange {
			if s.rng.Float64() >= encounterProb*digUpProb {
				continue
			}
			stores := s.colonyMap.Get(col)
			stores.Energy = 0
			stores.Pollen = 0
			// every bee of the colony dies
			for _, e := range s.beesOf[col] {
				if w.Alive(e) {
					s.death.die(w, e, "Colony killed by badger!")
				}
			}
		}
	}
}

func (s *badgers) Finalize(w *ecs.World) {}
