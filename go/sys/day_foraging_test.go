package sys

import (
	"math"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// Foraging tests: the colony is at (0, 0), Scaling_NLpatches_per_m = 0.1 (10 m per NetLogo patch).
// Flower species values from netlogo/_SYSTEM_Flowerspecies.csv.

// willowherb is American_willowherb: nectar_ml/flower 0.00035, concentration_mol/l 1.34, corollaDepth_mm 3.5,
// intFlowerTime_s 0.6. nectarMax and pollenMax are chosen by hand.
func willowherb() comp.Foodsource {
	return comp.Foodsource{NectarFlowerVolume_myl: 0.35,
		NectarConcentration_mol_per_l: 1.34, CorollaDepth_mm: 3.5, InterFlowerTime_s: 0.6,
		NectarMax_myl: 1000, PollenMax_g: 10}
}

// alsike is Alsike_clover: nectar_ml/flower 0.00046, concentration_mol/l 0.7, corollaDepth_mm 10, intFlowerTime_s 0.6.
func alsike() comp.Foodsource {
	return comp.Foodsource{NectarFlowerVolume_myl: 0.46,
		NectarConcentration_mol_per_l: 0.7, CorollaDepth_mm: 10, InterFlowerTime_s: 0.6,
		NectarMax_myl: 1000, PollenMax_g: 10}
}

// flowerPatch adds a flower patch at (x, y) [NetLogo patches] with radius_m, one layer per foodsource. The first
// layer is the masterpatch, all layers share LayersInPatchList (sorted by who = order of creation, as in [LAYR-10]).
func (aw *activityWorld) flowerPatch(x, y, radius_m float64, layers ...comp.Foodsource) []ecs.Entity {
	m := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](aw.w)
	result := []ecs.Entity{}
	for _, fs := range layers {
		fs.Radius_m = radius_m
		state := comp.FoodsourceState{Nectar_myl: fs.NectarMax_myl, Pollen_g: fs.PollenMax_g}
		result = append(result, m.NewEntity(&comp.Position{X: x, Y: y}, &fs, &state, &comp.Layer{}))
	}
	for i, e := range result {
		_, _, _, l := m.Get(e)
		l.Masterpatch = i == 0
		l.MasterpatchID = result[0]
		l.LayersInPatchList = result
	}
	return result
}

func (aw *activityWorld) state(e ecs.Entity) *comp.FoodsourceState {
	return ecs.NewMap1[comp.FoodsourceState](aw.w).Get(e)
}

// worker adds a forager of 100 mg of B_terrestris: glossa 7.072 mm [PROB], crop 40.2 microlitres, pollen load
// 0.01206 g [LOAD].
func (aw *activityWorld) worker(col ecs.Entity, number int) ecs.Entity {
	e := aw.adult(col, caste.Worker, number, 100, sitWorker)
	_, _, _, fo, _ := aw.bee(e)
	fo.GlossaLength_mm = ProboscisLengthREP(100, aw.sp)
	fo.Cropvolume_myl = CropAndPelletSizeREP("nectar", 100, aw.sp)
	fo.PollenPellets_g = CropAndPelletSizeREP("pollen", 100, aw.sp)
	fo.KnownMasterpatchesNectarList = []ecs.Entity{}
	fo.KnownMasterpatchesPollenList = []ecs.Entity{}
	return e
}

func newForagingWorld(t *testing.T, seed uint64) (*activityWorld, ecs.Entity, *forager) {
	t.Helper()
	aw := newActivityWorld(t, seed)
	// netlogo/_SYSTEM_Example_Farm_Parameters.csv: Lambda_detectProb -0.005, ForagingMortalityFactor 1
	aw.file.Lambda_detectProb = -0.005
	aw.file.ForagingMortalityFactor = 1
	ecs.GetResource[params.Derived](aw.w).MortalityForager_per_s = 0 // deterministic trips; mortality is tested separately
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { s.EnergyStore_kJ = 100 })
	return aw, col, &aw.loop.foraging
}

func TestDetectionProbREP(t *testing.T) {
	// [DET-01] to [DET-03] worked example: edge 100 m away: e ^ -0.5
	aw, _, f := newForagingWorld(t, 1)
	for _, tc := range []struct {
		name string
		x    float64
		want float64
	}{
		{"edge 100 m away", 15, 0.6065306597},
		{"colony inside the patch", 2, 1},
		{"edge 750 m away", 80, 0.0235177459},
		// [DET-02] NetLogo gives 0 beyond ForagingRangeMax_m (758 m). The colony lists hold only masterpatches whose
		// centre is in range [PRNG], so their edge is never beyond it (removed in step F3).
	} {
		p := aw.flowerPatch(tc.x, 0, 50, willowherb())[0]
		if got := f.DetectionProbREP(p, 0, 0); math.Abs(got-tc.want) > 1e-10 {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHandlingTime_s_REP(t *testing.T) {
	// [HAND-01] to [HAND-03] worked examples: worker of 100 mg
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	for _, tc := range []struct {
		name   string
		fs     comp.Foodsource
		pollen bool
		fill   float64
		want   float64
	}{
		{"willowherb nectar, full", willowherb(), false, 1, 166.2323791},
		{"willowherb nectar, half", willowherb(), false, 0.5, 285.6838077},
		{"alsike nectar, deep corolla", alsike(), false, 1, 655.1447617},
		{"nectar, empty", willowherb(), false, 0, 3600},
		{"pollen, full", willowherb(), true, 1, 292.2},
		{"pollen, half", willowherb(), true, 0.5, 327},
		{"pollen, empty", willowherb(), true, 0, 3600},
		{"pollen, almost empty", willowherb(), true, 0.001, 3600},
	} {
		p := aw.flowerPatch(5, 0, 10, tc.fs)[0]
		s := aw.state(p)
		s.Nectar_myl *= tc.fill
		s.Pollen_g *= tc.fill
		if got := f.HandlingTime_s_REP(bee, p, tc.pollen); math.Abs(got-tc.want) > 1e-6 {
			t.Errorf("%s: %v s, want %v", tc.name, got, tc.want)
		}
	}
}

func TestForaging_bestLayerREP(t *testing.T) {
	// [BLYR-02] pollen: the shortest handling time, [BLYR-03] nectar: the highest efficiency
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	_, _, _, fo, _ := aw.bee(bee)
	slow := willowherb()
	slow.InterFlowerTime_s = 2
	layers := aw.flowerPatch(30, 0, 50, slow, willowherb(), alsike())
	fo.PollenForager = true
	if got := f.Foraging_bestLayerREP(bee, layers[0]); got != layers[1] {
		t.Errorf("pollen: best layer %v, want the willowherb with 0.6 s (%v); ties go to the first", got, layers[1])
	}
	fo.PollenForager = false
	if got := f.Foraging_bestLayerREP(bee, layers[2]); got != layers[1] {
		t.Errorf("nectar: best layer %v, want the faster willowherb (%v)", got, layers[1])
	}
	aw.state(layers[1]).Nectar_myl = 0
	if got := f.Foraging_bestLayerREP(bee, layers[0]); got != layers[0] {
		t.Errorf("nectar without the emptied layer: %v, want the slow willowherb (%v)", got, layers[0])
	}
	for _, l := range layers {
		aw.state(l).Nectar_myl = 0
	}
	if got := f.Foraging_bestLayerREP(bee, layers[0]); !got.IsZero() {
		t.Errorf("no nectar left: %v, want none", got)
	}
}

func TestForaging_SortKnownPatchesListREP(t *testing.T) {
	// [SORT-01] duplicates removed, sorted by distance, stable for equal distances
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	far := aw.flowerPatch(50, 0, 10, willowherb())[0]
	near := aw.flowerPatch(0, 10, 10, willowherb())[0]
	sameA := aw.flowerPatch(30, 0, 10, willowherb())[0]
	sameB := aw.flowerPatch(0, -30, 10, willowherb())[0]
	got := f.Foraging_SortKnownPatchesListREP(bee, []ecs.Entity{sameB, far, near, sameA, far, sameB})
	want := []ecs.Entity{near, sameB, sameA, far}
	if len(got) != len(want) {
		t.Fatalf("%v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%v, want %v", got, want)
			break
		}
	}
}

func TestForagingTripPollen(t *testing.T) {
	// worked example [COST-02]: pollen foodsource 300 m away, handling time 292.2 s: trip 412.2 s, 0.0101 kJ,
	// then unloading 165 s [UNLD-01]
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	fs := willowherb()
	fs.PollenMax_g = 1000 // one load (0.01206 g) hardly changes the filling level
	p := aw.flowerPatch(30, 0, 10, fs)[0]
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.PollenForaging
	fo.PollensourceToGoTo = p
	fo.KnownMasterpatchesPollenList = []ecs.Entity{p}
	fo.Expectation_PollenTrip_s = 412.2

	f.ForagingProc(aw.w, bee)

	b, _, _, fo, _ = aw.bee(bee)
	_, s, _ := aw.colony(col)
	if !almostEqual(s.PollenStore_g, 0.01206) {
		t.Errorf("pollen store %v g, want 0.01206", s.PollenStore_g)
	}
	if math.Abs(100-s.EnergyStore_kJ-0.0101) > 0.0001 {
		t.Errorf("energy paid %v kJ, want about 0.0101", 100-s.EnergyStore_kJ)
	}
	if math.Abs(b.PersonalTime_s-(412.2+165)) > 0.01 {
		t.Errorf("personalTime_s %v, want about %v", b.PersonalTime_s, 412.2+165)
	}
	if st := aw.state(p); !almostEqual(st.Pollen_g, 1000-0.01206) {
		t.Errorf("foodsource: %v g", st.Pollen_g)
	}
	if fo.PollenLoadSquadron_g != 0 || aw.tally.trips != 1 {
		t.Errorf("load %v, trips %v", fo.PollenLoadSquadron_g, aw.tally.trips)
	}
	if !almostEqual(fo.Expectation_PollenTrip_s, (412.2+b.PersonalTime_s-165)/2) {
		t.Errorf("expectation %v", fo.Expectation_PollenTrip_s)
	}
}

func TestForagingEmptiesFoodsource(t *testing.T) {
	// finding 47 (see trans/netlogo-findings.md#finding-47): the last load gives the handling time 3600 s
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 3) // a cohort of 3 takes up to 120.6 microlitres
	p := aw.flowerPatch(30, 0, 10, willowherb())[0]
	aw.state(p).Nectar_myl = 100
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.NectarForaging
	fo.NectarsourceToGoTo = p
	fo.KnownMasterpatchesNectarList = []ecs.Entity{p}

	f.ForagingProc(aw.w, bee)

	b, _, _, _, _ = aw.bee(bee)
	trip := 2*300/5.0 + 3600
	if !almostEqual(b.PersonalTime_s, trip+165) {
		t.Errorf("personalTime_s %v, want %v (handling time 3600 s)", b.PersonalTime_s, trip+165)
	}
	_, s, _ := aw.colony(col)
	want := 100 + 100*0.00582*1.34 - (600+3600*5*0.3)*aw.sp.FlightCosts_kJ_per_m_per_mg*100*3
	if !almostEqual(s.EnergyStore_kJ, want) {
		t.Errorf("energy store %v kJ, want %v", s.EnergyStore_kJ, want)
	}
	if st := aw.state(p); st.Nectar_myl != 0 {
		t.Errorf("foodsource: %v microlitres, want 0", st.Nectar_myl)
	}
}

func TestForagingSearch(t *testing.T) {
	// [SRCH-02], [SRCH-03]: the colony lies inside the flower patch, so it is detected with probability 1.
	// [PCH-03] a novice is always unhappy after her first trip.
	aw, col, f := newForagingWorld(t, 1)
	layers := aw.flowerPatch(1, 0, 50, willowherb(), alsike())
	_, _, _, _, _, r := aw.colonyMap.Get(col)
	r.MasterpatchesWithNectarlayersInFlowerAndRangeList = []ecs.Entity{layers[0]}
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.NectarForaging
	nectarBefore := [2]float64{aw.state(layers[0]).Nectar_myl, aw.state(layers[1]).Nectar_myl}

	f.ForagingProc(aw.w, bee)

	b, _, _, fo, _ = aw.bee(bee)
	if len(fo.KnownMasterpatchesNectarList) != 1 || fo.KnownMasterpatchesNectarList[0] != layers[0] {
		t.Errorf("known nectar patches %v, want [%v]", fo.KnownMasterpatchesNectarList, layers[0])
	}
	if b.Activity != activity.ReturningUnhappyN {
		t.Errorf("activity %v, want returningUnhappyN (novice)", b.Activity)
	}
	// [BLYR-03] willowherb: more energy per load and a shorter handling time than alsike
	if aw.state(layers[0]).Nectar_myl >= nectarBefore[0] || aw.state(layers[1]).Nectar_myl != nectarBefore[1] {
		t.Error("the bee did not collect at the best layer (willowherb)")
	}
}

func TestForagingSearchFails(t *testing.T) {
	// [SRCH-03] nothing found: returningEmpty, [COST-01] 500 s and 0.0244 kJ, [PCH-02] next time a new search
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.PollenForaging
	f.ForagingProc(aw.w, bee)
	b, _, _, fo, _ = aw.bee(bee)
	_, s, _ := aw.colony(col)
	if b.Activity != activity.ReturningEmpty || b.PersonalTime_s != 500 || !fo.PollensourceToGoTo.IsZero() {
		t.Errorf("activity %v, personalTime_s %v, next foodsource %v", b.Activity, b.PersonalTime_s, fo.PollensourceToGoTo)
	}
	if !almostEqual(100-s.EnergyStore_kJ, 2500*aw.sp.FlightCosts_kJ_per_m_per_mg*100) {
		t.Errorf("energy paid %v kJ, want 0.0244", 100-s.EnergyStore_kJ)
	}
}

func TestForagingMortality(t *testing.T) {
	// [COST-01] 1 - (1 - 1e-05 * 100) ^ 500 = 0.394 with ForagingMortalityFactor = 100
	aw, col, f := newForagingWorld(t, 3)
	ecs.GetResource[params.Derived](aw.w).MortalityForager_per_s = 1e-05
	aw.file.ForagingMortalityFactor = 100
	n := 2000
	for range n {
		bee := aw.worker(col, 1)
		b, _, _, _, _ := aw.bee(bee)
		b.Activity = activity.NectarForaging
		aw.active.List = append(aw.active.List, bee)
		f.ForagingProc(aw.w, bee)
	}
	died := float64(aw.tally.beesDied) / float64(n)
	if want := 1 - math.Pow(1-1e-03, 500); math.Abs(died-want) > 0.03 {
		t.Errorf("%v of the foragers died, want about %v", died, want)
	}
	if len(aw.active.List) != n-aw.tally.beesDied {
		t.Errorf("dead foragers still in the activity queue")
	}
	// event.DiedForaging for the output (obs.Counters.TotalForagersEverDiedForaging), event.ForagingTrip for each trip
	if aw.tally.foragersDied != aw.tally.beesDied || aw.tally.trips != n {
		t.Errorf("%v foragers died on a trip, %v bees died, %v trips; want equal and %d trips",
			aw.tally.foragersDied, aw.tally.beesDied, aw.tally.trips, n)
	}
}

func TestForaging_PatchChoiceProc(t *testing.T) {
	// [PCH-03] worked example: expectation 300 s, trip 400 s: new expectation 350 s.
	// [PCH-05] after returningEmpty: 10 % go to the closest known patch, 90 % search.
	aw, col, f := newForagingWorld(t, 5)
	p := aw.flowerPatch(30, 0, 10, willowherb())[0]
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	fo.KnownMasterpatchesNectarList = []ecs.Entity{p}
	fo.CurrentFoodsource = p
	fo.Expectation_NectarTrip_s = 300
	b.Activity = activity.ExpForagingN
	f.Foraging_PatchChoiceProc(bee, 400)
	if fo.Expectation_NectarTrip_s != 350 {
		t.Errorf("expectation %v, want 350", fo.Expectation_NectarTrip_s)
	}

	closest := 0
	n := 4000
	for range n {
		b.Activity = activity.ReturningEmpty
		f.Foraging_PatchChoiceProc(bee, 500)
		if fo.NectarsourceToGoTo == p {
			closest++
		}
	}
	if share := float64(closest) / float64(n); math.Abs(share-0.1) > 0.02 {
		t.Errorf("after returningEmpty %v go to the known patch, want about 0.1", share)
	}
}
