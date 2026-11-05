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
	a := newExampleFarmSetup(t, 3, &CreateInitialQueens{})
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
			b.Number != 1 || b.AdultAge != 180 || b.BroodAge != 36 || !b.ColonyID.IsZero() || b.SpeciesID != spIndex {
			t.Fatalf("queen state: %+v", b)
		}
		if len(g.AllelesList) != 2 || len(g.SpermathecaList) != 1 || g.MtDNA < 0 || g.MtDNA >= 139.9 {
			t.Fatalf("genetics: %+v", g)
		}
		// [IQUE-05]
		if *th != (comp.Thresholds{ThEgglaying: 0.1, ThForagingNectar: 0.7, ThForagingPollen: 0.7, ThNursing: 0.5}) {
			t.Fatalf("thresholds: %+v", th)
		}
		// [IQUE-06] whole seconds from -3599 to 3600 (finding 31)
		if b.PersonalTime_s < -3599 || b.PersonalTime_s > 3600 || b.PersonalTime_s != math.Trunc(b.PersonalTime_s) {
			t.Fatalf("personalTime_s = %v", b.PersonalTime_s)
		}
		// [IQUE-07] whole days, > 0 and < seasonStop (305)
		if !b.EmergingDate.IsSet() || b.EmergingDate.Day() <= 0 || b.EmergingDate.Day() >= 305 {
			t.Fatalf("emergingDate = %v", b.EmergingDate)
		}
		// [IQUE-09]
		if b.Weight_mg < 590 || b.Weight_mg > 980 {
			t.Fatalf("weight = %v", b.Weight_mg)
		}
		if f.GlossaLength_mm != ProboscisLengthREP(b.Weight_mg, sp) || f.Cropvolume_myl != 173 ||
			f.PollenPellets_g != CropAndPelletSizeREP("pollen", b.Weight_mg, sp) || !f.CurrentFoodsource.IsZero() {
			t.Fatalf("forager: %+v", f)
		}
		sumWeight += b.Weight_mg
		sumEmergence += float64(b.EmergingDate.Day())
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
		if c.modifier == 0 && ecs.GetResource[params.ParameterFile](a.World).InitialQueensModifierBehaviorSpace != 1 {
			t.Error("modifier 0 must be set to 1")
		}
	}
}
