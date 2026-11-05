package sys

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"
)

func TestEgglaying(t *testing.T) {
	// [EGG-01] worked example: 12 eggs cost 0.018 g pollen and 0.1116 kJ. [EGGP-01] to [EGGP-03]
	bw := newBroodWorld(t, 7)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 1, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	qb, g, _, _, _ := bw.bee(queen)
	qb.PersonalTime_s = 1000
	qb.EmergingDate = day.At(120)
	g.MtDNA, g.AllelesList, g.SpermathecaList = 42, []float64{10, 20}, []float64{30}

	bw.eggs.layEggs(queen)

	eggs := bw.eggsOf(col)
	if len(eggs) != 1 {
		t.Fatalf("%d egg cohorts, want 1", len(eggs))
	}
	b, g, th, f, d := bw.bee(eggs[0])
	if b.Number != 12 || b.Stage != stage.Egg || b.Caste != caste.Undefined ||
		b.ColonyID != col || b.SpeciesID != bw.species || b.Weight_mg != 1.5 || b.EmergingDate.IsSet() ||
		b.AdultAge != 0 || b.BroodAge != 0 || b.PersonalTime_s != 0 || b.Mated {
		t.Errorf("egg %+v", *b)
	}
	if g.MtDNA != 42 || len(g.AllelesList) != 2 || g.AllelesList[0] != 30 || (g.AllelesList[1] != 10 && g.AllelesList[1] != 20) ||
		len(g.SpermathecaList) != 0 {
		t.Errorf("egg genetics %+v: want mtDNA 42, alleles [30 10|20], no spermatheca", *g)
	}
	if *th != thresholds(sitEgg) || f.Cropvolume_myl != 0 || f.KnownMasterpatchesNectarList == nil || *d != (comp.Development{}) {
		t.Errorf("egg thresholds %+v, forager %+v, development %+v", *th, *f, *d)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenStore_g, 1-0.018) || !almostEqual(s.EnergyStore_kJ, 1-0.1116) {
		t.Errorf("stores %v g, %v kJ; want 0.982, 0.8884", s.PollenStore_g, s.EnergyStore_kJ)
	}
	// [EGG-03]
	if qb, _, _, _, _ := bw.bee(queen); qb.PersonalTime_s != 1000+86400 || qb.EmergingDate != day.At(120) {
		t.Errorf("queen personalTime_s %v, emergingDate %v", qb.PersonalTime_s, qb.EmergingDate)
	}
}

func TestEgglayingNotEnoughStores(t *testing.T) {
	// [EGG-02] strict >: a store equal to the cost lays nothing. [EGG-03] the 24 h are used anyway
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 0.018, 10
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	bw.eggs.layEggs(queen)
	if len(bw.eggsOf(col)) != 0 {
		t.Error("eggs laid without enough pollen")
	}
	if b, _, _, _, _ := bw.bee(queen); b.PersonalTime_s != 86400 {
		t.Errorf("personalTime_s %v, want 86400", b.PersonalTime_s)
	}
	if _, s, _ := bw.colony(col); s.PollenStore_g != 0.018 {
		t.Errorf("pollen store %v, want 0.018", s.PollenStore_g)
	}
}

func TestEgglayingPloidy(t *testing.T) {
	// [EGG-02] haploid eggs only after the switch point day. SexLocus? with equal alleles: diploid male [EGGP-01].
	// The ploidy is the number of alleles (step F9).
	for _, tc := range []struct {
		name      string
		switchDay day.Date
		sexLocus  bool
		caste     caste.Caste
		alleles   int
	}{
		{"on switch point day", day.At(200), false, caste.Undefined, 2},
		{"after switch point", day.At(199), false, caste.Male, 1},
		{"sex locus, homozygous", day.Date{}, true, caste.Male, 2},
	} {
		bw := newBroodWorld(t, 1)
		bw.file.SexLocus = tc.sexLocus
		col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			c.SwitchPointDate = tc.switchDay
			s.PollenStore_g, s.EnergyStore_kJ = 1, 1
		})
		queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
		_, g, _, _, _ := bw.bee(queen)
		g.AllelesList, g.SpermathecaList = []float64{5, 5}, []float64{5}
		bw.eggs.layEggs(queen)
		b, g, _, _, _ := bw.bee(bw.eggsOf(col)[0])
		if b.Caste != tc.caste || len(g.AllelesList) != tc.alleles {
			t.Errorf("%s: caste %v, alleles %v; want %v, %d alleles", tc.name, b.Caste, g.AllelesList, tc.caste, tc.alleles)
		}
	}
}

func TestEgglayingOneCohort(t *testing.T) {
	// [EGG-01] one cohort of batchsize (12) eggs for the cost of 12 eggs (0.0015 g pollen each). gobees-fast has
	// no individual-based colonies, which in NetLogo lay 12 cohorts of 1 egg for the cost of 1 egg (finding 43).
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 1, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	_, g, _, _, _ := bw.bee(queen)
	g.AllelesList, g.SpermathecaList = []float64{10, 20}, []float64{30}
	bw.eggs.layEggs(queen)
	eggs := bw.eggsOf(col)
	if len(eggs) != 1 {
		t.Fatalf("%d egg cohorts, want 1", len(eggs))
	}
	if b, _, _, _, _ := bw.bee(eggs[0]); b.Number != 12 {
		t.Errorf("cohort of %v eggs, want 12", b.Number)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenStore_g, 1-12*0.0015) || !almostEqual(s.EnergyStore_kJ, 1-12*0.0015*6.2) {
		t.Errorf("stores %v g, %v kJ; want the cost of 12 eggs", s.PollenStore_g, s.EnergyStore_kJ)
	}
}
