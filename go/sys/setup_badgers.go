package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// badgersSetup creates Parameters.Badgers badgers, each at the position of a "Scrub" flower patch without a badger,
// more than Constants.Badgers.MinDistance from the badgers placed before. The example farm has no badgers.
//
// NetLogo: CreateBadgersProc.
type badgersSetup struct{}

func (s *badgersSetup) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.Parameters](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))
	constants := ecs.GetResource[params.Constants](w)
	createBadgers(w, file.Badgers, constants.Badgers.MinDistance, rng)
}

func (s *badgersSetup) Update(w *ecs.World) {}

func (s *badgersSetup) Finalize(w *ecs.World) {}

// createBadgers creates n badgers on "Scrub" flower patches. distanceLimit in [m].
func createBadgers(w *ecs.World, n int, distanceLimit float64, rng *rand.Rand) {
	burrowHabitatsList := []string{"Scrub"} // case-sensitive, as member?

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
	burrows := []ecs.Entity{} // the flower patch of each placed badger

	// burrowFsSet before the badgers were created: every Scrub flower patch
	burrowFsSet := slices.Clone(sites)
	for _, b := range badgers {
		if len(burrowFsSet) == 0 {
			badgerDeath{}.die(w, b, "Badger: not enough habitat!")
			continue
		}
		chosenFs := burrowFsSet[rng.IntN(len(burrowFsSet))]
		// the badger moves to the position of the flower patch. NetLogo: the centre of its patch cell (pxcor,
		// pycor).
		pos, _ := badgerMap.Get(b)
		pos.X, pos.Y = chosenFs.pos.X, chosenFs.pos.Y
		placed = append(placed, b)
		burrows = append(burrows, chosenFs.e)

		// recomputed with the badgers placed so far: no badger on the flower patch (NetLogo: on its patch cell,
		// badgers-here) and none within distanceLimit (in-radius: <=). Only the placed badgers count. NetLogo also
		// counts the badgers that still sit at the origin, which blocks the Scrub patches near the origin.
		burrowFsSet = burrowFsSet[:0]
		for _, site := range sites {
			free := true
			for i, other := range placed {
				p, _ := badgerMap.Get(other)
				if burrows[i] == site.e || distanceBetween(site.pos.X, site.pos.Y, p.X, p.Y) <= distanceLimit {
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
