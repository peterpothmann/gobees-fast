package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CreateBadgers [INIT-08].
//
// Creates N_Badgers badgers, each in the centre of the patch cell of a "Scrub" masterpatch, at least 300 m from
// the badgers placed before. N_Badgers is 0 in netlogo/_SYSTEM_Example_Farm_Parameters.csv.
// Not translated: size, shape (GUI).
//
// NetLogo: CreateBadgersProc.
type CreateBadgers struct{}

func (s *CreateBadgers) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	landscape := ecs.GetResource[globals.Landscape](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))
	CreateBadgersProc(w, int(file.N_Badgers), landscape.Scaling_NLpatches_per_m, rng)
}

func (s *CreateBadgers) Update(w *ecs.World) {}

func (s *CreateBadgers) Finalize(w *ecs.World) {}

// CreateBadgersProc creates n badgers on "Scrub" masterpatches [BADC], finding 48 (see
// trans/netlogo-findings.md#finding-48).
func CreateBadgersProc(w *ecs.World, n int, scaling_NLpatches_per_m float64, rng *rand.Rand) {
	burrowHabitatsList := []string{"Scrub"} // case-sensitive, as member?
	const distanceLimit_m = 300
	distance_patches := distanceLimit_m * scaling_NLpatches_per_m

	type burrow struct {
		e   ecs.Entity
		pos comp.Position
	}
	// the Scrub masterpatches, in storage order (NetLogo: an agentset, one-of picks at random)
	sites := []burrow{}
	query := ecs.NewFilter3[comp.Position, comp.Foodsource, comp.Layer](w).Query()
	for query.Next() {
		pos, fs, layer := query.Get()
		if layer.Masterpatch && slices.Contains(burrowHabitatsList, fs.PatchType) {
			sites = append(sites, burrow{query.Entity(), *pos})
		}
	}

	// [BADC-01] create-badgers creates ALL badgers first, at the origin (0, 0), each with a who number (not
	// kept, step F12). Then each
	// runs the block, in random order (as ask).
	badgerMap := ecs.NewMap2[comp.Position, comp.Badger](w)
	badgers := make([]ecs.Entity, n)
	for i := range badgers {
		badgers[i] = badgerMap.NewEntity(&comp.Position{}, &comp.Badger{})
	}
	placed := []ecs.Entity{}
	util.Shuffle(rng, badgers)

	// burrowFsSet before the badgers were created: every Scrub masterpatch
	burrowFsSet := slices.Clone(sites)
	for _, b := range badgers {
		if len(burrowFsSet) == 0 {
			badgerDeath{}.DieProc(w, b, "Badger: not enough habitat!")
			continue
		}
		chosenFs := burrowFsSet[rng.IntN(len(burrowFsSet))]
		// the badger moves to the centre of the patch cell of the foodsource (pxcor, pycor)
		pos, _ := badgerMap.Get(b)
		pos.X, pos.Y = float64(util.Pcor(chosenFs.pos.X)), float64(util.Pcor(chosenFs.pos.Y))
		placed = append(placed, b)

		// recomputed with the badgers placed so far: no badger on the patch cell of the foodsource (badgers-here)
		// and none within distance_patches (in-radius: <=). NetLogo also counts the badgers that still sit at the
		// origin, which blocks the Scrub patches near the origin (finding 48). Fixed in step F5.
		burrowFsSet = burrowFsSet[:0]
		for _, site := range sites {
			free := true
			for _, other := range placed {
				p, _ := badgerMap.Get(other)
				samePatch := util.Pcor(p.X) == util.Pcor(site.pos.X) && util.Pcor(p.Y) == util.Pcor(site.pos.Y)
				if samePatch || util.Distance(site.pos.X, site.pos.Y, p.X, p.Y) <= distance_patches {
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
