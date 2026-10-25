package sys

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// qlnWorld is a colonyWorld with the QueensLeavingNest system.
func qlnWorld(t *testing.T, unlimitedMales bool) (*colonyWorld, *QueensLeavingNest) {
	t.Helper()
	cw := newColonyWorld(t, 1)
	cw.file.UnlimitedMales = unlimitedMales
	s := &QueensLeavingNest{}
	s.Initialize(cw.w)
	return cw, s
}

// addYoungQueen adds an unmated adult queen with the "youngQueen" thresholds [DEVP-04] to a colony.
func (cw *colonyWorld) addYoungQueen(col ecs.Entity, number float64) ecs.Entity {
	e := cw.addBee(col, stage.Adult, caste.Queen, number, 0)
	_, _, b, _, th, _, _ := cw.beeMap.Get(e)
	b.Mated = false
	*th = thresholds(sitYoungQueen)
	return e
}

// addMale adds an adult male with the given alleles (1: haploid, 2: diploid).
func (cw *colonyWorld) addMale(col ecs.Entity, alleles ...float64) ecs.Entity {
	e := cw.addBee(col, stage.Adult, caste.Male, 1, 0)
	_, _, _, g, _, _, _ := cw.beeMap.Get(e)
	g.AllelesList = alleles
	return e
}

func TestQueensLeavingNestMating(t *testing.T) {
	cw, s := qlnWorld(t, false)
	col := cw.addColony(nil)
	mother := cw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	young := cw.addYoungQueen(col, 3)
	queenPupa := cw.addBee(col, stage.Pupa, caste.Queen, 1, 0)
	// [QLN-02] the male can be in any colony, here in another one
	male := cw.addMale(cw.addColony(nil), 7)

	s.Update(cw.w)

	// [QLN-01] the cohort of 3 leaves as one agent
	_, _, b, g, th, _, _ := cw.beeMap.Get(young)
	if b.Number != 3 || !b.Mated || b.Activity != activity.Hibernate || !b.ColonyID.IsZero() {
		t.Errorf("young queen: number %v, mated %v, activity %v, colony %v; want 3, true, hibernate, none",
			b.Number, b.Mated, b.Activity, b.ColonyID)
	}
	if !slices.Equal(g.SpermathecaList, []float64{7}) {
		t.Errorf("spermathecaList %v, want [7]", g.SpermathecaList)
	}
	// [QLN-03] only the egg-laying threshold changes
	want := thresholds(sitYoungQueen)
	want.ThEgglaying = 0.1
	if *th != want {
		t.Errorf("thresholds %+v, want %+v", *th, want)
	}
	// the list is a copy of the male's list
	_, _, _, mg, _, _, _ := cw.beeMap.Get(male)
	mg.AllelesList[0] = 99
	if g.SpermathecaList[0] != 7 {
		t.Errorf("spermathecaList shares the male's allelesList")
	}
	// the mother queen and the queen pupa stay in the colony
	for _, e := range []ecs.Entity{mother, queenPupa} {
		_, _, b, _, _, _, _ := cw.beeMap.Get(e)
		if b.ColonyID != col || b.Activity == activity.Hibernate {
			t.Errorf("%v %v left the colony", b.Stage, b.Caste)
		}
	}
	if !cw.w.Alive(male) {
		t.Errorf("the male died")
	}
}

func TestQueensLeavingNestNoMale(t *testing.T) {
	// [QLN-02] no adult male of her species: with UnlimitedMales? = false she dies. A male of another species and
	// male brood of her species do not count.
	cw, s := qlnWorld(t, false)
	other := cw.speciesEntityNamed(t, "B_pascuorum")
	col := cw.addColony(nil)
	young := cw.addYoungQueen(col, 2)
	foreign := cw.addMale(col, 5)
	_, _, fb, _, _, _, _ := cw.beeMap.Get(foreign)
	fb.SpeciesID = other
	cw.addBee(col, stage.Pupa, caste.Male, 4, 0)

	s.Update(cw.w)
	if cw.w.Alive(young) {
		t.Errorf("young queen without a male survived")
	}
	if cw.counters.TotalBeesEverDied != 2 {
		t.Errorf("TotalBeesEverDied %v, want 2 (the cohort)", cw.counters.TotalBeesEverDied)
	}
}

func TestQueensLeavingNestUnlimitedMales(t *testing.T) {
	// [QLN-02] with UnlimitedMales? = true she mates with a male from outside: one allele -1 to -N_ForeignAlleles
	cw, s := qlnWorld(t, true)
	col := cw.addColony(nil)
	queens := []ecs.Entity{}
	for range 2000 {
		queens = append(queens, cw.addYoungQueen(col, 1))
	}
	s.Update(cw.w)
	seen := map[float64]bool{}
	for _, e := range queens {
		_, _, b, g, _, _, _ := cw.beeMap.Get(e)
		if len(g.SpermathecaList) != 1 || g.SpermathecaList[0] > -1 || g.SpermathecaList[0] < -24 ||
			g.SpermathecaList[0] != float64(int(g.SpermathecaList[0])) {
			t.Fatalf("spermathecaList %v, want one integer allele in [-24, -1]", g.SpermathecaList)
		}
		if !b.Mated || b.Activity != activity.Hibernate {
			t.Fatalf("queen not mated or not hibernating")
		}
		seen[g.SpermathecaList[0]] = true
	}
	if len(seen) != 24 {
		t.Errorf("%d different foreign alleles, want all 24", len(seen))
	}
}

func TestQueensLeavingNestDiploidMale(t *testing.T) {
	// [QLN-04] a queen that mated with a diploid male dies
	cw, s := qlnWorld(t, true)
	col := cw.addColony(nil)
	young := cw.addYoungQueen(col, 1)
	cw.addMale(col, 3, 3)
	s.Update(cw.w)
	if cw.w.Alive(young) {
		t.Errorf("queen mated with a diploid male survived")
	}
}

// speciesEntityNamed adds a second species with the values of B_terrestris and the given name.
func (cw *colonyWorld) speciesEntityNamed(t *testing.T, name string) ecs.Entity {
	t.Helper()
	sp := *terrestris(t)
	sp.Name = name
	return ecs.NewMap2[comp.Who, comp.Species](cw.w).NewEntity(&comp.Who{Who: cw.who.Next()}, &sp)
}
