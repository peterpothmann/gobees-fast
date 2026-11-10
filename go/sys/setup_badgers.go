package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// badgersSetup creates Parameters.Badgers badgers, each in the centre of the patch cell of a "Scrub" masterpatch,
// more than Constants.Badgers.MinDistance from the badgers placed before. The example farm has no badgers.
//
// NetLogo: CreateBadgersProc.
type badgersSetup struct{}

func (s *badgersSetup) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.Parameters](w)
	landscape := ecs.GetResource[globals.Landscape](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))
	constants := ecs.GetResource[params.Constants](w)
	createBadgers(w, file.Badgers, landscape.Scale, constants.Badgers.MinDistance, rng)
}

func (s *badgersSetup) Update(w *ecs.World) {}

func (s *badgersSetup) Finalize(w *ecs.World) {}

// createBadgers creates n badgers on "Scrub" masterpatches.
// scale is in NetLogo patches per metre [1/m], distanceLimit in [m].
func createBadgers(w *ecs.World, n int, scale, distanceLimit float64, rng *rand.Rand) {
	burrowHabitatsList := []string{"Scrub"} // case-sensitive, as member?
	limitDistance := distanceLimit * scale  // [NetLogo patches]

	type burrow struct {
		e   ecs.Entity
		pos comp.Position
	}
	// the Scrub flower patches, in storage order (NetLogo: the Scrub masterpatches, an agentset, one-of picks at
	// random)
	sites := []burrow{}
	query := ecs.NewFilter2[comp.Position, comp.FlowerPatch](w).Query()
	for query.Next() {
		pos, fp := query.Get()
		if slices.Contains(burrowHabitatsList, fp.PatchType) {
			sites = append(sites, burrow{query.Entity(), *pos})
		}
	}

	// ALL badgers are created first, at the origin (0, 0). Then each is placed. ORDER: irrelevant, the badgers are
	// identical until placed. Not shuffled (NetLogo: random order).
	badgerMap := ecs.NewMap2[comp.Position, comp.Badger](w)
	badgers := make([]ecs.Entity, n)
	for i := range badgers {
		badgers[i] = badgerMap.NewEntity(&comp.Position{}, &comp.Badger{})
	}
	placed := []ecs.Entity{}

	// burrowFsSet before the badgers were created: every Scrub flower patch
	burrowFsSet := slices.Clone(sites)
	for _, b := range badgers {
		if len(burrowFsSet) == 0 {
			badgerDeath{}.die(w, b, "Badger: not enough habitat!")
			continue
		}
		chosenFs := burrowFsSet[rng.IntN(len(burrowFsSet))]
		// the badger moves to the centre of the patch cell of the foodsource (pxcor, pycor)
		pos, _ := badgerMap.Get(b)
		pos.X, pos.Y = float64(pcor(chosenFs.pos.X)), float64(pcor(chosenFs.pos.Y))
		placed = append(placed, b)

		// recomputed with the badgers placed so far: no badger on the patch cell of the foodsource (badgers-here)
		// and none within limitDistance (in-radius: <=). Only the placed badgers count. NetLogo also counts the
		// badgers that still sit at the origin, which blocks the Scrub patches near the origin.
		burrowFsSet = burrowFsSet[:0]
		for _, site := range sites {
			free := true
			for _, other := range placed {
				p, _ := badgerMap.Get(other)
				samePatch := pcor(p.X) == pcor(site.pos.X) && pcor(p.Y) == pcor(site.pos.Y)
				if samePatch || distanceBetween(site.pos.X, site.pos.Y, p.X, p.Y) <= limitDistance {
					free = false
					break
				}
			}
			if free {
				burrowFsSet = append(burrowFsSet, site)
			}
		}
	}
}
