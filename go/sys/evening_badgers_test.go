package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// prowlWorld is a colonyWorld with badgers at the given positions [m] and the badgers system.
func prowlWorld(t *testing.T, seed uint64, positions ...comp.Position) (*colonyWorld, *badgers) {
	t.Helper()
	cw := newColonyWorld(t, seed)
	m := ecs.NewMap2[comp.Position, comp.Badger](cw.w)
	for _, p := range positions {
		m.NewEntity(&p, &comp.Badger{})
	}
	s := &badgers{}
	s.Initialize(cw.w)
	return cw, s
}

// addColonyAt adds a colony at (x, y).
func (cw *colonyWorld) addColonyAt(x, y float64, change func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts)) ecs.Entity {
	col := cw.addColony(change)
	pos, _, _, _, _, _ := cw.colonyMap.Get(col)
	pos.X, pos.Y = x, y
	return col
}

func TestBadgersDigUpColony(t *testing.T) {
	// a dug-up colony loses its stores and all its bees. The colony dies the next morning in colonyUpdate.
	cw, s := prowlWorld(t, 1, comp.Position{X: 0, Y: 0})
	near := cw.addColonyAt(100, 0, func(c *comp.Colony, st *comp.ColonyStores, n *comp.ColonyCounts) {
		st.Energy, st.Pollen = 50, 3
	})
	cw.addBee(near, stage.Adult, caste.Queen, 1, 30)
	cw.addBee(near, stage.Adult, caste.Worker, 3, 5)
	cw.addBee(near, stage.Egg, caste.Undefined, 5, 0)
	cw.addBee(near, stage.Larva, caste.Worker, 2, 0)
	far := cw.addColonyAt(800, 0, nil) // 800 m
	farQueen := cw.addBee(far, stage.Adult, caste.Queen, 1, 30)
	hibernating := cw.addHibernatingQueen(1) // colonyID -1 and position (0, 0): safe

	for day := 0; day < 2000; day++ {
		if _, st, _ := cw.colony(near); st.Energy == 0 {
			break
		}
		s.Update(cw.w)
	}
	_, st, _ := cw.colony(near)
	if st.Energy != 0 || st.Pollen != 0 {
		t.Fatalf("near colony: energy %v, pollen %v; want 0, 0", st.Energy, st.Pollen)
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
	// p = 0.19 * 0.1 per badger, colony and day. Range strictly < 735 m.
	cases := []struct {
		name    string
		badgers []comp.Position
		dist    float64
		want    float64
	}{
		{"one badger", []comp.Position{{X: 0, Y: 0}}, 500, 0.019},
		{"two badgers", []comp.Position{{X: 0, Y: 0}, {X: 0, Y: 10}}, 500, 1 - 0.981*0.981},
		{"exactly 735 m", []comp.Position{{X: 0, Y: 0}}, 735, 0},
	}
	for _, c := range cases {
		cw, s := prowlWorld(t, 2, c.badgers...)
		const n = 20000
		cols := make([]ecs.Entity, n)
		for i := range cols {
			cols[i] = cw.addColonyAt(c.dist, 0, func(_ *comp.Colony, st *comp.ColonyStores, _ *comp.ColonyCounts) {
				st.Energy = 1
			})
		}
		s.Update(cw.w)
		dug := 0
		for _, col := range cols {
			if _, st, _ := cw.colony(col); st.Energy == 0 {
				dug++
			}
		}
		got := float64(dug) / n
		if got < c.want*0.8 || got > c.want*1.2 {
			t.Errorf("%s: %.4f of the colonies dug up, want about %.4f", c.name, got, c.want)
		}
	}
}
