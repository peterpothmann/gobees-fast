package sys

import (
	"math/rand/v2"
	"testing"

	"gobees/comp"

	"github.com/mlange-42/ark/ecs"
)

// addScrub adds a Scrub flower patch at (x, y).
func addScrub(w *ecs.World, x, y float64) ecs.Entity {
	return ecs.NewMap2[comp.Position, comp.FlowerPatch](w).NewEntity(
		&comp.Position{X: x, Y: y}, &comp.FlowerPatch{PatchType: "Scrub"})
}

// badgerPositions returns the positions of all badgers.
func badgerPositions(w *ecs.World) []comp.Position {
	result := []comp.Position{}
	q := ecs.NewFilter2[comp.Position, comp.Badger](w).Query()
	for q.Next() {
		pos, _ := q.Get()
		result = append(result, *pos)
	}
	return result
}

func TestCreateBadgersNearOrigin(t *testing.T) {
	// the badgers that have not moved yet sit at the origin, but only the placed badgers exclude Scrub patches.
	// Scrub A lies 146 m from the origin, so both badgers can choose it:
	// P(A occupied) = 1/3 + 2/3 * 1/2 = 2/3, sd 0.009. NetLogo: only the first badger can choose it, P = 1/3.
	const runs = 3000
	occupiedA := 0
	for seed := range uint64(runs) {
		w := ecs.NewWorld()
		addScrub(w, 103, 104)   // A
		addScrub(w, 1000, 1000) // B
		addScrub(w, 2000, 1000) // C
		createBadgers(w, 2, 300, rand.New(rand.NewPCG(seed, 5)))

		pos := badgerPositions(w)
		if len(pos) != 2 {
			t.Fatalf("%d badgers, want 2", len(pos))
		}
		// at the position of the flower patch, more than 300 m apart (NetLogo: at the centre of its patch cell)
		if distanceBetween(pos[0].X, pos[0].Y, pos[1].X, pos[1].Y) <= 300 {
			t.Fatalf("badgers at %v and %v are not more than 300 m apart", pos[0], pos[1])
		}
		for _, p := range pos {
			if p == (comp.Position{X: 103, Y: 104}) {
				occupiedA++
			}
		}
	}
	if f := float64(occupiedA) / runs; f < 0.63 || f > 0.70 {
		t.Errorf("Scrub A occupied in %.3f of the runs, want about 2/3", f)
	}
}

func TestCreateBadgersNotEnoughHabitat(t *testing.T) {
	// no candidate left: the badger dies.
	w := ecs.NewWorld()
	addScrub(w, 1000, 1000)
	addScrub(w, 2000, 1000)
	addScrub(w, 2050, 1000) // within 300 m of the previous one
	createBadgers(w, 3, 300, rand.New(rand.NewPCG(1, 5)))
	if n := len(badgerPositions(w)); n != 2 {
		t.Errorf("%d badgers, want 2", n)
	}
	// at most one badger per flower patch, also without a minimum distance
	w = ecs.NewWorld()
	addScrub(w, 1000, 1000)
	createBadgers(w, 2, 0, rand.New(rand.NewPCG(1, 5)))
	if n := len(badgerPositions(w)); n != 1 {
		t.Errorf("%d badgers on one Scrub patch, want 1", n)
	}
}
