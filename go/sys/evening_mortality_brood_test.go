package sys

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

func TestMortalityBrood(t *testing.T) {
	// [MORB-01] to [MORB-04], maximum ages from the worked example: eggs 10, larvae 20.8, pupae 34.7, queen
	// larvae 29, queen pupae 47
	bw := newBroodWorld(t, 1)
	sp := bw.sp
	for _, x := range []struct{ got, want float64 }{{sp.DevAgeHatchingMax_d, 10}, {sp.DevAgePupationMax_d, 20.8},
		{sp.DevAgeEmergingMax_d, 34.7}, {sp.DevAge_Q_PupationMax_d, 29}, {sp.DevAge_Q_EmergingMax_d, 47}} {
		if !almostEqual(x.got, x.want) {
			t.Errorf("maximum age %v, want %v", x.got, x.want)
		}
	}
	col := bw.addColony(nil)
	oldEgg := bw.addBrood(col, stage.Egg, caste.Undefined, 12, 11, 1.5, 0)
	egg := bw.addBrood(col, stage.Egg, caste.Undefined, 12, 10, 1.5, 0)
	lightLarva := bw.addBrood(col, stage.Larva, caste.Worker, 3, 21, 31.2, 7.17) // weight 0.5, incubation 1
	coldQueen := bw.addBrood(col, stage.Larva, caste.Queen, 2, 30, 590, 4.72)    // weight 1, incubation 0.5
	queenLarva := bw.addBrood(col, stage.Larva, caste.Queen, 1, 29, 100, 1)      // not too old
	oldPupa := bw.addBrood(col, stage.Pupa, caste.Male, 4, 35, 100, 1)
	queenPupa := bw.addBrood(col, stage.Pupa, caste.Queen, 1, 47, 900, 1)
	oldQueenPupa := bw.addBrood(col, stage.Pupa, caste.Queen, 5, 48, 900, 1)
	bw.addBee(bw.addColony(nil), stage.Adult, caste.Queen, 1, 300)
	queenWithoutColony := bw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	b, _, _, _, _ := bw.bee(queenWithoutColony)
	b.ColonyID = ecs.Entity{}

	bw.morb.Update(bw.w)

	for _, tc := range []struct {
		name  string
		e     ecs.Entity
		alive bool
	}{{"egg 11", oldEgg, false}, {"egg 10", egg, true}, {"light larva", lightLarva, false},
		{"cold queen larva", coldQueen, false}, {"queen larva 29", queenLarva, true}, {"male pupa 35", oldPupa, false},
		{"queen pupa 47", queenPupa, true}, {"queen pupa 48", oldQueenPupa, false}, {"queen without colony", queenWithoutColony, true}} {
		if bw.w.Alive(tc.e) != tc.alive {
			t.Errorf("%s: alive %v, want %v", tc.name, bw.w.Alive(tc.e), tc.alive)
		}
	}
	if bw.tally.beesDied != 12+3+2+4+5 {
		t.Errorf("%v bees died, want 26", bw.tally.beesDied)
	}

	// [MORB-02] assertion: a larva that would have pupated
	expectPanic(t, "larva with enough weight and incubation", func() {
		bw.addBrood(col, stage.Larva, caste.Worker, 1, 21, 100, 100)
		bw.morb.Update(bw.w)
	})
}

func TestMortalityBroodCompetitionPoint(t *testing.T) {
	// [MORB-04] after the competition point every egg dies, larvae survive
	bw := newBroodWorld(t, 1)
	after := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.CompetitionPointDate = day.At(199) })
	onCP := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.CompetitionPointDate = day.At(200) })
	eggAfter := bw.addBrood(after, stage.Egg, caste.Undefined, 12, 1, 1.5, 0)
	larvaAfter := bw.addBrood(after, stage.Larva, caste.Undefined, 12, 6, 10, 3)
	eggOnCP := bw.addBrood(onCP, stage.Egg, caste.Undefined, 12, 1, 1.5, 0)
	bw.morb.Update(bw.w)
	if bw.w.Alive(eggAfter) || !bw.w.Alive(larvaAfter) || !bw.w.Alive(eggOnCP) {
		t.Errorf("alive: egg after CP %v, larva after CP %v, egg on CP day %v; want false, true, true",
			bw.w.Alive(eggAfter), bw.w.Alive(larvaAfter), bw.w.Alive(eggOnCP))
	}
	if bw.tally.beesDied != 12 {
		t.Errorf("%v bees died, want 12", bw.tally.beesDied)
	}
}
