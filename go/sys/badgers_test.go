package sys

import (
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
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
		who := &globals.WhoCounter{}
		CreateBadgersProc(w, 2, 0.1, who, rand.New(rand.NewPCG(seed, 5)))

		pos := badgerPositions(w)
		if len(pos) != 2 || who.Next() != 2 {
			t.Fatalf("%d badgers, want 2 with who numbers 0 and 1", len(pos))
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
	// [BADC-01] no candidate left: the badger dies. All three have used up a who number.
	w := ecs.NewWorld()
	addScrub(w, 100, 100)
	addScrub(w, 200, 100)
	addScrub(w, 205, 100) // within 300 m of the previous one
	who := &globals.WhoCounter{}
	CreateBadgersProc(w, 3, 0.1, who, rand.New(rand.NewPCG(1, 5)))
	if n := len(badgerPositions(w)); n != 2 {
		t.Errorf("%d badgers, want 2", n)
	}
	if next := who.Next(); next != 3 {
		t.Errorf("next who number %d, want 3", next)
	}
}

// prowlWorld is a colonyWorld (Scaling 0.1 patches per m: 735 m = 73.5 patches) with badgers at the given
// positions and the BadgersOnTheProwl system.
func prowlWorld(t *testing.T, seed uint64, badgers ...comp.Position) (*colonyWorld, *BadgersOnTheProwl) {
	t.Helper()
	cw := newColonyWorld(t, seed)
	m := ecs.NewMap3[comp.Who, comp.Position, comp.Badger](cw.w)
	for _, p := range badgers {
		m.NewEntity(&comp.Who{Who: cw.who.Next()}, &p, &comp.Badger{})
	}
	s := &BadgersOnTheProwl{}
	s.Initialize(cw.w)
	return cw, s
}

// addColonyAt adds a colony at (x, y).
func (cw *colonyWorld) addColonyAt(x, y float64, change func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts)) ecs.Entity {
	col := cw.addColony(change)
	_, pos, _, _, _, _, _ := cw.colonyMap.Get(col)
	pos.X, pos.Y = x, y
	return col
}

func TestBadgersDigUpColony(t *testing.T) {
	// [BADG-01] a dug-up colony loses its stores and all its bees. The colony dies the next morning in [UCOL-11].
	cw, s := prowlWorld(t, 1, comp.Position{X: 0, Y: 0})
	near := cw.addColonyAt(10, 0, func(c *comp.Colony, st *comp.ColonyStores, n *comp.ColonyCounts) {
		st.EnergyStore_kJ, st.PollenStore_g = 50, 3
	})
	cw.addBee(near, stage.Adult, caste.Queen, 1, 30)
	cw.addBee(near, stage.Adult, caste.Worker, 3, 5)
	cw.addBee(near, stage.Egg, caste.Undefined, 5, 0)
	cw.addBee(near, stage.Larva, caste.Worker, 2, 0)
	far := cw.addColonyAt(80, 0, nil) // 800 m
	farQueen := cw.addBee(far, stage.Adult, caste.Queen, 1, 30)
	hibernating := cw.addHibernatingQueen(1) // colonyID -1 and position (0, 0): safe

	for day := 0; day < 2000; day++ {
		if _, st, _ := cw.colony(near); st.EnergyStore_kJ == 0 {
			break
		}
		s.Update(cw.w)
	}
	_, st, _ := cw.colony(near)
	if st.EnergyStore_kJ != 0 || st.PollenStore_g != 0 {
		t.Fatalf("near colony: energy %v, pollen %v; want 0, 0", st.EnergyStore_kJ, st.PollenStore_g)
	}
	if left := len(beesByColony(nil, ecs.NewFilter1[comp.Bee](cw.w), func(b *comp.Bee) bool { return true })[near]); left != 0 {
		t.Errorf("%d bee agents left in the dug-up colony", left)
	}
	if !cw.w.Alive(near) || !cw.w.Alive(farQueen) || !cw.w.Alive(hibernating) {
		t.Errorf("colony alive %v, far queen alive %v, hibernating queen alive %v; want all true",
			cw.w.Alive(near), cw.w.Alive(farQueen), cw.w.Alive(hibernating))
	}
	if cw.tally.beesDied != 11 {
		t.Errorf("%v bees died, want 11", cw.tally.beesDied)
	}
	// next morning: the empty colony dies
	cw.nextMorning()
	if cw.w.Alive(near) || cw.tally.coloniesDied != 1 {
		t.Errorf("colony alive %v, %d colonies died; want false and 1", cw.w.Alive(near), cw.tally.coloniesDied)
	}
}

func TestBadgersDigUpProbability(t *testing.T) {
	// [BADG-01] p = 0.19 * 0.1 per badger, colony and day. Range strictly < 735 m.
	cases := []struct {
		name    string
		badgers []comp.Position
		dist    float64
		want    float64
	}{
		{"one badger", []comp.Position{{X: 0, Y: 0}}, 50, 0.019},
		{"two badgers", []comp.Position{{X: 0, Y: 0}, {X: 0, Y: 1}}, 50, 1 - 0.981*0.981},
		{"exactly 735 m", []comp.Position{{X: 0, Y: 0}}, 73.5, 0},
	}
	for _, c := range cases {
		cw, s := prowlWorld(t, 2, c.badgers...)
		const n = 20000
		cols := make([]ecs.Entity, n)
		for i := range cols {
			cols[i] = cw.addColonyAt(c.dist, 0, func(_ *comp.Colony, st *comp.ColonyStores, _ *comp.ColonyCounts) {
				st.EnergyStore_kJ = 1
			})
		}
		s.Update(cw.w)
		dug := 0
		for _, col := range cols {
			if _, st, _ := cw.colony(col); st.EnergyStore_kJ == 0 {
				dug++
			}
		}
		got := float64(dug) / n
		if got < c.want*0.8 || got > c.want*1.2 {
			t.Errorf("%s: %.4f of the colonies dug up, want about %.4f", c.name, got, c.want)
		}
	}
}
