package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
)

func TestLarvaeFeeding(t *testing.T) {
	// worked example: pollen 0.3 of 0.6 g, energy 20 of 10 kJ: factor 0.5
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.Pollen, s.IdealPollen, s.Energy, s.IdealEnergy = 0.3, 0.6, 20, 10
	})
	larva := bw.addBrood(col, stage.Larva, caste.Undefined, 12, 6, 10, 0)
	pupa := bw.addBrood(col, stage.Pupa, caste.Worker, 3, 15, 200, 0)
	// an ideal store of 0: no feeding
	noIdeal := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.Pollen, s.Energy, s.IdealEnergy = 5, 50, 10
	})
	unfed := bw.addBrood(noIdeal, stage.Larva, caste.Worker, 1, 10, 10, 0)
	// the cap compares ONE larva with the whole store: each of 12 larvae gets the full 1 mg
	small := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.Pollen, s.IdealPollen, s.Energy, s.IdealEnergy = 0.001, 0.001, 50, 10
	})
	capped := bw.addBrood(small, stage.Larva, caste.Worker, 12, 10, 10, 0)

	bw.feed.Update(bw.w)

	// gain 8.8 mg, half of it: 4.4 mg per larva, 52.8 mg for the colony
	if b, _, _, _, _ := bw.bee(larva); !almostEqual(b.Weight, 14.4) {
		t.Errorf("larva weight %v mg, want 14.4", b.Weight)
	}
	if b, _, _, _, _ := bw.bee(pupa); b.Weight != 200 {
		t.Errorf("pupa was fed: %v mg", b.Weight)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.Pollen, 0.3-0.0528) || !almostEqual(s.Energy, 20-0.0528*6.2) {
		t.Errorf("stores %v g, %v kJ; want %v, %v", s.Pollen, s.Energy, 0.3-0.0528, 20-0.0528*6.2)
	}
	if b, _, _, _, _ := bw.bee(unfed); b.Weight != 10 {
		t.Errorf("larva without ideal store was fed: %v mg", b.Weight)
	}
	if b, _, _, _, _ := bw.bee(capped); !almostEqual(b.Weight, 11) {
		t.Errorf("capped larva %v mg, want 11", b.Weight)
	}
	if _, s, _ := bw.colony(small); !almostEqual(s.Pollen, 0.001-0.012) {
		t.Errorf("small store %v g, want %v (negative)", s.Pollen, 0.001-0.012)
	}
}
