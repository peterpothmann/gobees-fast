package sys

import (
	"math"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestCreateInitialQueensExampleFarm(t *testing.T) {
	a := newExampleFarmSetup(t, 3, &initialQueensSetup{})
	w := a.World
	sp := terrestris(t)
	spIndex := speciesIndices(w)["B_terrestris"][0]

	query := ecs.NewFilter6[comp.Position, comp.Bee, comp.Genetics, comp.Thresholds, comp.Forager, comp.Development](w).Query()
	n := 0
	sumWeight, sumEmergence := 0.0, 0.0
	for query.Next() {
		_, b, g, th, f, _ := query.Get()
		n++
		// [IQUE-03], [IQUE-04]
		// brood? false: stage adult, ploidy 2: 2 alleles (below), speciesName: the species (step F9)
		if b.Caste != caste.Queen || b.Stage != stage.Adult || b.Activity != activity.Hibernate || !b.Mated ||
			b.Number != 1 || b.AdultAge != 180 || b.BroodAge != 36 || !b.Colony.IsZero() || b.Species != spIndex {
			t.Fatalf("queen state: %+v", b)
		}
		if len(g.Alleles) != 2 || len(g.Spermatheca) != 1 || g.MtDNA < 0 || g.MtDNA >= 139.9 {
			t.Fatalf("genetics: %+v", g)
		}
		// [IQUE-05]
		if *th != (comp.Thresholds{Egglaying: 0.1, NectarForaging: 0.7, PollenForaging: 0.7, Nursing: 0.5}) {
			t.Fatalf("thresholds: %+v", th)
		}
		// [IQUE-06] whole seconds from -3599 to 3600 (finding 31)
		if b.PersonalTime < -3599 || b.PersonalTime > 3600 || b.PersonalTime != math.Trunc(b.PersonalTime) {
			t.Fatalf("personalTime_s = %v", b.PersonalTime)
		}
		// [IQUE-07] whole days, > 0 and < seasonStop (305)
		if !b.EmergenceDate.IsSet() || b.EmergenceDate.Day() <= 0 || b.EmergenceDate.Day() >= 305 {
			t.Fatalf("emergingDate = %v", b.EmergenceDate)
		}
		// [IQUE-09]
		if b.Weight < 590 || b.Weight > 980 {
			t.Fatalf("weight = %v", b.Weight)
		}
		if f.GlossaLength != proboscisLength(b.Weight, sp) || f.CropVolume != 173 ||
			f.PollenPellets != pollenLoad(b.Weight, sp) || !f.CurrentFoodsource.IsZero() {
			t.Fatalf("forager: %+v", f)
		}
		sumWeight += b.Weight
		sumEmergence += float64(b.EmergenceDate.Day())
	}
	if n != 500 {
		t.Fatalf("got %d queens, want 500 (\"B_terrestris 500\")", n)
	}
	// statistics over 500 queens: mean 785 mg (sd 97.5, cut at 590 and 980) and mean emergence day 91 (sd 28)
	if m := sumWeight / 500; math.Abs(m-785) > 20 {
		t.Errorf("mean weight %v, want about 785", m)
	}
	if m := sumEmergence / 500; math.Abs(m-91) > 8 {
		t.Errorf("mean emergence day %v, want about 91", m)
	}
}

func TestCreateInitialQueensModifier(t *testing.T) {
	// [IQUE-01] modifier 0 counts as 1 and changes the parameter; the number is rounded, .5 up: 500 * 0.501 = 250.5 -> 251
	for _, c := range []struct{ modifier, want float64 }{{0, 500}, {0.501, 251}} {
		a := newExampleFarmSetupWithModifier(t, c.modifier)
		q := ecs.NewFilter1[comp.Bee](a.World).Query()
		n := q.Count()
		q.Close()
		if float64(n) != c.want {
			t.Errorf("modifier %v: %d queens, want %v", c.modifier, n, c.want)
		}
		if c.modifier == 0 && ecs.GetResource[params.Parameters](a.World).InitialQueensFactor != 1 {
			t.Error("modifier 0 must be set to 1")
		}
	}
}
