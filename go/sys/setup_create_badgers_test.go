package sys

import (
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// addScrub adds a Scrub masterpatch at (x, y).
func addScrub(w *ecs.World, x, y float64) ecs.Entity {
	return ecs.NewMap3[comp.Position, comp.Foodsource, comp.Layer](w).NewEntity(
		&comp.Position{X: x, Y: y}, &comp.Foodsource{PatchType: "Scrub"}, &comp.Layer{Masterpatch: true})
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
	// [BADC-01] finding 48 (see trans/netlogo-findings.md#finding-48), fixed in step F5: the badgers that have not
	// moved yet sit at the origin, but only the placed badgers exclude Scrub patches. Scrub A lies 14 patches (140 m)
	// from the origin, so both badgers can choose it: P(A occupied) = 1/3 + 2/3 * 1/2 = 2/3, sd 0.009. NetLogo: only
	// the first badger can choose it, P = 1/3.
	const runs = 3000
	occupiedA := 0
	for seed := range uint64(runs) {
		w := ecs.NewWorld()
		addScrub(w, 10.3, 10.4) // A, patch cell (10, 10)
		addScrub(w, 100, 100)   // B
		addScrub(w, 200, 100)   // C
		CreateBadgersProc(w, 2, 0.1, rand.New(rand.NewPCG(seed, 5)))

		pos := badgerPositions(w)
		if len(pos) != 2 {
			t.Fatalf("%d badgers, want 2", len(pos))
		}
		// at the centre of the patch cell, at least 300 m apart
		if util.Distance(pos[0].X, pos[0].Y, pos[1].X, pos[1].Y) <= 30 {
			t.Fatalf("badgers at %v and %v are not more than 300 m apart", pos[0], pos[1])
		}
		for _, p := range pos {
			if p == (comp.Position{X: 10, Y: 10}) {
				occupiedA++
			}
		}
	}
	if f := float64(occupiedA) / runs; f < 0.63 || f > 0.70 {
		t.Errorf("Scrub A occupied in %.3f of the runs, want about 2/3", f)
	}
}

func TestCreateBadgersNotEnoughHabitat(t *testing.T) {
	// [BADC-01] no candidate left: the badger dies.
	w := ecs.NewWorld()
	addScrub(w, 100, 100)
	addScrub(w, 200, 100)
	addScrub(w, 205, 100) // within 300 m of the previous one
	CreateBadgersProc(w, 3, 0.1, rand.New(rand.NewPCG(1, 5)))
	if n := len(badgerPositions(w)); n != 2 {
		t.Errorf("%d badgers, want 2", n)
	}
}
