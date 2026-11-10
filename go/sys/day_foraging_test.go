package sys

import (
	"math"
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// Foraging tests: the colony is at (0, 0), Landscape.Scale = 0.1 (10 m per NetLogo patch).
// Flower species values from netlogo/_SYSTEM_Flowerspecies.csv.

// willowherb is American_willowherb: nectar_ml/flower 0.00035, concentration_mol/l 1.34, corollaDepth_mm 3.5,
// intFlowerTime_s 0.6. nectarMax and pollenMax are chosen by hand.
func willowherb() comp.Foodsource {
	return comp.Foodsource{NectarPerFlower: 0.35,
		NectarConcentration: 1.34, CorollaDepth: 3.5, InterFlowerTime: 0.6,
		NectarMax: 1000, PollenMax: 10}
}

// alsike is Alsike_clover: nectar_ml/flower 0.00046, concentration_mol/l 0.7, corollaDepth_mm 10, intFlowerTime_s 0.6.
func alsike() comp.Foodsource {
	return comp.Foodsource{NectarPerFlower: 0.46,
		NectarConcentration: 0.7, CorollaDepth: 10, InterFlowerTime: 0.6,
		NectarMax: 1000, PollenMax: 10}
}

// flowerPatch adds a flower patch at (x, y) [NetLogo patches] with radius_m, one layer per foodsource. The first
// layer is the masterpatch, all layers share PatchLayers (in order of creation, as NetLogo sorts them by who).
func (aw *activityWorld) flowerPatch(x, y, radius float64, layers ...comp.Foodsource) []ecs.Entity {
	m := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](aw.w)
	result := []ecs.Entity{}
	for _, fs := range layers {
		fs.Radius = radius
		state := comp.FoodsourceState{Nectar: fs.NectarMax, Pollen: fs.PollenMax}
		result = append(result, m.NewEntity(&comp.Position{X: x, Y: y}, &fs, &state, &comp.Layer{}))
	}
	for i, e := range result {
		_, _, _, l := m.Get(e)
		l.IsMasterpatch = i == 0
		l.Masterpatch = result[0]
		l.PatchLayers = result
	}
	return result
}

func (aw *activityWorld) state(e ecs.Entity) *comp.FoodsourceState {
	return ecs.NewMap1[comp.FoodsourceState](aw.w).Get(e)
}

// worker adds a forager of 100 mg of B_terrestris: glossa 7.072 mm, crop 40.2 microlitres, pollen load
// 0.01206 g.
func (aw *activityWorld) worker(col ecs.Entity, number int) ecs.Entity {
	e := aw.adult(col, caste.Worker, number, 100, sitWorker)
	_, _, _, fo, _ := aw.bee(e)
	fo.GlossaLength = proboscisLength(100, aw.sp)
	fo.CropVolume = cropVolume(100, aw.sp, &defaults.Foraging)
	fo.PollenPellets = pollenLoad(100, aw.sp, &defaults.Foraging)
	fo.KnownNectarPatches = []ecs.Entity{}
	fo.KnownPollenPatches = []ecs.Entity{}
	return e
}

func newForagingWorld(t *testing.T, seed uint64) (*activityWorld, ecs.Entity, *forager) {
	t.Helper()
	aw := newActivityWorld(t, seed)
	// netlogo/_SYSTEM_Example_Farm_Parameters.csv: Lambda_detectProb -0.005, ForagingMortalityFactor 1
	aw.parameters.DetectionLambda = -0.005
	aw.parameters.ForagingMortalityFactor = 1
	ecs.GetResource[params.Derived](aw.w).ForagerMortality = 0 // deterministic trips, mortality is tested separately
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { s.Energy = 100 })
	return aw, col, &aw.loop.foraging
}

func TestDetectionProb(t *testing.T) {
	// worked example: edge 100 m away: e ^ -0.5
	aw, _, f := newForagingWorld(t, 1)
	for _, tc := range []struct {
		name string
		x    float64
		want float64
	}{
		{"edge 100 m away", 15, 0.6065306597},
		{"colony inside the patch", 2, 1},
		{"edge 750 m away", 80, 0.0235177459},
		// NetLogo gives 0 beyond ForagingRangeMax_m (758 m). The colony lists hold only masterpatches whose
		// centre is in range, so their edge is never beyond it.
	} {
		p := aw.flowerPatch(tc.x, 0, 50, willowherb())[0]
		if got := f.detectionProb(p, 0, 0); math.Abs(got-tc.want) > 1e-10 {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHandlingTime(t *testing.T) {
	// worked examples: worker of 100 mg
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
		s.Nectar *= tc.fill
		s.Pollen *= tc.fill
		if got := f.handlingTime(f.tripOf(bee), p, tc.pollen); math.Abs(got-tc.want) > 1e-6 {
			t.Errorf("%s: %v s, want %v", tc.name, got, tc.want)
		}
	}
}

// TestHandlingTimeIngestionCache checks the cache of the ingestion time (comp.Forager.IngestionTime): a bee that
// comes back to a foodsource, at another filling level or after another foodsource, gets bit for bit the handling
// time of a new bee without a cache.
func TestHandlingTimeIngestionCache(t *testing.T) {
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	willow := aw.flowerPatch(5, 0, 10, willowherb())[0]
	deep := aw.flowerPatch(50, 0, 10, alsike())[0]
	fsMap := ecs.NewMap1[comp.Foodsource](aw.w)
	for i, v := range []struct {
		p    ecs.Entity
		fill float64
	}{{willow, 1}, {willow, 0.5}, {deep, 1}, {willow, 0.5}, {willow, 0.25}, {deep, 0.5}} {
		aw.state(v.p).Nectar = fsMap.Get(v.p).NectarMax * v.fill
		got := f.handlingTime(f.tripOf(bee), v.p, false)
		if want := f.handlingTime(f.tripOf(aw.worker(col, 1)), v.p, false); got != want {
			t.Errorf("visit %d: %v s, want %v (new bee)", i+1, got, want)
		}
	}
	if _, _, _, fo, _ := aw.bee(bee); fo.IngestionSource != deep {
		t.Errorf("IngestionSource %v, want the last foodsource %v", fo.IngestionSource, deep)
	}
}

func TestPowMinus04(t *testing.T) {
	// bit for bit equal to math.Pow: the base of handlingTime lies in [0.001, 1.41], checked far beyond that
	rng := rand.New(rand.NewPCG(1, 2))
	check := func(x float64) {
		if got, want := powMinus04(x), math.Pow(x, -0.4); math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("powMinus04(%v) = %v, math.Pow = %v", x, got, want)
		}
	}
	for _, x := range []float64{0.001, 0.5, 1, 1.41, math.SmallestNonzeroFloat64, math.MaxFloat64} {
		check(x)
	}
	for range 1_000_000 {
		check(0.001 + 1.409*rng.Float64())
		check(math.Exp(1400 * (rng.Float64() - 0.5))) // about 1e-304 to 1e304
	}
}

func TestBestLayer(t *testing.T) {
	// pollen: the shortest handling time nectar: the highest efficiency
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	_, _, _, fo, _ := aw.bee(bee)
	slow := willowherb()
	slow.InterFlowerTime = 2
	layers := aw.flowerPatch(30, 0, 50, slow, willowherb(), alsike())
	fo.PollenForager = true
	if got := f.bestLayer(f.tripOf(bee), layers[0]); got != layers[1] {
		t.Errorf("pollen: best layer %v, want the willowherb with 0.6 s (%v); ties go to the first", got, layers[1])
	}
	fo.PollenForager = false
	if got := f.bestLayer(f.tripOf(bee), layers[2]); got != layers[1] {
		t.Errorf("nectar: best layer %v, want the faster willowherb (%v)", got, layers[1])
	}
	aw.state(layers[1]).Nectar = 0
	if got := f.bestLayer(f.tripOf(bee), layers[0]); got != layers[0] {
		t.Errorf("nectar without the emptied layer: %v, want the slow willowherb (%v)", got, layers[0])
	}
	for _, l := range layers {
		aw.state(l).Nectar = 0
	}
	if got := f.bestLayer(f.tripOf(bee), layers[0]); !got.IsZero() {
		t.Errorf("no nectar left: %v, want none", got)
	}
}

func TestSortKnownPatches(t *testing.T) {
	// duplicates removed, sorted by distance, stable for equal distances
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	far := aw.flowerPatch(50, 0, 10, willowherb())[0]
	near := aw.flowerPatch(0, 10, 10, willowherb())[0]
	sameA := aw.flowerPatch(30, 0, 10, willowherb())[0]
	sameB := aw.flowerPatch(0, -30, 10, willowherb())[0]
	got := f.sortKnownPatches(f.tripOf(bee), []ecs.Entity{sameB, far, near, sameA, far, sameB})
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
	// worked example: pollen foodsource 300 m away, handling time 292.2 s: trip 412.2 s, 0.0101 kJ,
	// then unloading 165 s
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	fs := willowherb()
	fs.PollenMax = 1000 // one load (0.01206 g) hardly changes the filling level
	p := aw.flowerPatch(30, 0, 10, fs)[0]
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.PollenForaging
	fo.PollenSource = p
	fo.KnownPollenPatches = []ecs.Entity{p}
	fo.ExpectedPollenTrip = 412.2

	f.forage(aw.w, bee)

	b, _, _, fo, _ = aw.bee(bee)
	_, s, _ := aw.colony(col)
	if !almostEqual(s.Pollen, 0.01206) {
		t.Errorf("pollen store %v g, want 0.01206", s.Pollen)
	}
	if math.Abs(100-s.Energy-0.0101) > 0.0001 {
		t.Errorf("energy paid %v kJ, want about 0.0101", 100-s.Energy)
	}
	if math.Abs(b.PersonalTime-(412.2+165)) > 0.01 {
		t.Errorf("PersonalTime %v, want about %v", b.PersonalTime, 412.2+165)
	}
	if st := aw.state(p); !almostEqual(st.Pollen, 1000-0.01206) {
		t.Errorf("foodsource: %v g", st.Pollen)
	}
	if fo.CarriedPollen != 0 || aw.tally.trips != 1 {
		t.Errorf("load %v, trips %v", fo.CarriedPollen, aw.tally.trips)
	}
	if !almostEqual(fo.ExpectedPollenTrip, (412.2+b.PersonalTime-165)/2) {
		t.Errorf("expectation %v", fo.ExpectedPollenTrip)
	}
}

func TestForagingEmptiesFoodsource(t *testing.T) {
	// the last load empties the foodsource and gives the handling time 3600 s
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 3) // a cohort of 3 takes up to 120.6 microlitres
	p := aw.flowerPatch(30, 0, 10, willowherb())[0]
	aw.state(p).Nectar = 100
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.NectarForaging
	fo.NectarSource = p
	fo.KnownNectarPatches = []ecs.Entity{p}

	f.forage(aw.w, bee)

	b, _, _, _, _ = aw.bee(bee)
	trip := 2*300/5.0 + 3600
	if !almostEqual(b.PersonalTime, trip+165) {
		t.Errorf("PersonalTime %v, want %v (handling time 3600 s)", b.PersonalTime, trip+165)
	}
	_, s, _ := aw.colony(col)
	want := 100 + 100*0.00582*1.34 - (600+3600*5*0.3)*aw.sp.FlightCosts*100*3
	if !almostEqual(s.Energy, want) {
		t.Errorf("energy store %v kJ, want %v", s.Energy, want)
	}
	if st := aw.state(p); st.Nectar != 0 {
		t.Errorf("foodsource: %v microlitres, want 0", st.Nectar)
	}
}

func TestForagingSearch(t *testing.T) {
	// the colony lies inside the flower patch, so it is detected with probability 1.
	// a novice is always unhappy after her first trip.
	aw, col, f := newForagingWorld(t, 1)
	layers := aw.flowerPatch(1, 0, 50, willowherb(), alsike())
	_, _, _, _, _, r := aw.colonyMap.Get(col)
	r.NectarPatches = []ecs.Entity{layers[0]}
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.NectarForaging
	nectarBefore := [2]float64{aw.state(layers[0]).Nectar, aw.state(layers[1]).Nectar}

	f.forage(aw.w, bee)

	b, _, _, fo, _ = aw.bee(bee)
	if len(fo.KnownNectarPatches) != 1 || fo.KnownNectarPatches[0] != layers[0] {
		t.Errorf("known nectar patches %v, want [%v]", fo.KnownNectarPatches, layers[0])
	}
	if b.Activity != activity.ReturningUnhappyNectar {
		t.Errorf("activity %v, want returningUnhappyN (novice)", b.Activity)
	}
	// willowherb: more energy per load and a shorter handling time than alsike
	if aw.state(layers[0]).Nectar >= nectarBefore[0] || aw.state(layers[1]).Nectar != nectarBefore[1] {
		t.Error("the bee did not collect at the best layer (willowherb)")
	}
}

func TestForagingSearchFails(t *testing.T) {
	// nothing found: returningEmpty 500 s and 0.0244 kJ next time a new search
	aw, col, f := newForagingWorld(t, 1)
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	b.Activity = activity.PollenForaging
	f.forage(aw.w, bee)
	b, _, _, fo, _ = aw.bee(bee)
	_, s, _ := aw.colony(col)
	if b.Activity != activity.ReturningEmpty || b.PersonalTime != 500 || !fo.PollenSource.IsZero() {
		t.Errorf("activity %v, PersonalTime %v, next foodsource %v", b.Activity, b.PersonalTime, fo.PollenSource)
	}
	if !almostEqual(100-s.Energy, 2500*aw.sp.FlightCosts*100) {
		t.Errorf("energy paid %v kJ, want 0.0244", 100-s.Energy)
	}
}

func TestForagingMortality(t *testing.T) {
	// 1 - (1 - 1e-05 * 100) ^ 500 = 0.394 with ForagingMortalityFactor = 100
	aw, col, f := newForagingWorld(t, 3)
	ecs.GetResource[params.Derived](aw.w).ForagerMortality = 1e-05
	aw.parameters.ForagingMortalityFactor = 100
	n := 2000
	for range n {
		bee := aw.worker(col, 1)
		b, _, _, _, _ := aw.bee(bee)
		b.Activity = activity.NectarForaging
		aw.active.Set(append(aw.active.Entities(), bee), func(ecs.Entity) float64 { return 0 })
		f.forage(aw.w, bee)
	}
	died := float64(aw.tally.beesDied) / float64(n)
	if want := 1 - math.Pow(1-1e-03, 500); math.Abs(died-want) > 0.03 {
		t.Errorf("%v of the foragers died, want about %v", died, want)
	}
	if aw.active.Len() != n-aw.tally.beesDied {
		t.Errorf("dead foragers still in the activity queue")
	}
	// event.DiedForaging for the output (obs.Counters.TotalForagersEverDiedForaging), event.ForagingTrip for each trip
	if aw.tally.foragersDied != aw.tally.beesDied || aw.tally.trips != n {
		t.Errorf("%v foragers died on a trip, %v bees died, %v trips; want equal and %d trips",
			aw.tally.foragersDied, aw.tally.beesDied, aw.tally.trips, n)
	}
}

func TestChoosePatch(t *testing.T) {
	// worked example: expectation 300 s, trip 400 s: new expectation 350 s.
	// after returningEmpty: 10 % go to the closest known patch, 90 % search.
	aw, col, f := newForagingWorld(t, 5)
	p := aw.flowerPatch(30, 0, 10, willowherb())[0]
	bee := aw.worker(col, 1)
	b, _, _, fo, _ := aw.bee(bee)
	fo.KnownNectarPatches = []ecs.Entity{p}
	fo.CurrentFoodsource = p
	fo.ExpectedNectarTrip = 300
	b.Activity = activity.ReturningHappyNectar
	f.choosePatch(f.tripOf(bee), 400)
	if fo.ExpectedNectarTrip != 350 {
		t.Errorf("expectation %v, want 350", fo.ExpectedNectarTrip)
	}

	closest := 0
	n := 4000
	for range n {
		b.Activity = activity.ReturningEmpty
		f.choosePatch(f.tripOf(bee), 500)
		if fo.NectarSource == p {
			closest++
		}
	}
	if share := float64(closest) / float64(n); math.Abs(share-0.1) > 0.02 {
		t.Errorf("after returningEmpty %v go to the known patch, want about 0.1", share)
	}
}
