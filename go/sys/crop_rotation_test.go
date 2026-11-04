package sys

import (
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// smallFoodsources is a foodsources file made by hand for the crop rotation tests, in the format of
// netlogo/_SYSTEM_Example_Farm_Foodsources.txt: two flower patches with large areas, so that their layers are not
// removed as too small [LAYR-06], [LAYR-07].
const smallFoodsources = `10
2
id patchType patchColour xcor ycor size_sqm flowerSpeciesList perimeter_m info
0 "Scrub" 125 50 50 100000 [["Bramble" 0.211] ["Dandelion" 0.113]] 1200  "no info"
1 "Garden" 45 120 80 50000 [["Dandelion" 0.5]] 900  "no info"
`

// cropWorld runs the example farm setup with the crop rotation list "_SYSTEM_Example_Farm_Foodsources.txt
// small_Foodsources.txt" and writes small_Foodsources.txt into the working directory.
func cropWorld(t *testing.T, list string) (*app.App, *CropRotation) {
	t.Helper()
	cr := &CropRotation{}
	a := newExampleFarmSetupWith(t, 1, func(p *params.Params) { p.File.CropRotationListAsString = list },
		&Tick{}, &UpdateMorning{}, &CreateInitialQueens{}, cr)
	dir := ecs.GetResource[params.WorkingDirectory](a.World).Path
	if err := os.WriteFile(filepath.Join(dir, "small_Foodsources.txt"), []byte(smallFoodsources), 0o644); err != nil {
		t.Fatal(err)
	}
	return a, cr
}

// foodsourceSummary returns the number of foodsources, the masterpatches by patch type and the largest who number.
func foodsourceSummary(w *ecs.World) (int, map[string]int, int64) {
	n, maxWho := 0, int64(-1)
	masters := map[string]int{}
	q := ecs.NewFilter3[comp.Who, comp.Foodsource, comp.Layer](w).Query()
	for q.Next() {
		who, fs, layer := q.Get()
		n++
		maxWho = max(maxWho, who.Who)
		if layer.Masterpatch {
			masters[fs.PatchType]++
		}
	}
	return n, masters, maxWho
}

// setDay sets ticks and Day.
func setDay(w *ecs.World, ticks int64) {
	ecs.GetResource[globals.Ticks](w).Ticks = ticks
	ecs.GetResource[globals.Calendar](w).Day = dayOfYear(ticks)
}

func TestCropRotation(t *testing.T) {
	a, cr := cropWorld(t, "_SYSTEM_Example_Farm_Foodsources.txt small_Foodsources.txt")
	w := a.World
	derived := ecs.GetResource[globals.Derived](w)
	nBefore, _, maxWhoBefore := foodsourceSummary(w)

	// a bee that knows a foodsource, and a colony with range lists
	var anyFs ecs.Entity
	fsq := ecs.NewFilter1[comp.Foodsource](w).Query()
	for fsq.Next() {
		anyFs = fsq.Entity()
	}
	var bee ecs.Entity
	bq := ecs.NewFilter1[comp.Forager](w).Query()
	for bq.Next() {
		bee = bq.Entity()
	}
	f := ecs.NewMap1[comp.Forager](w).Get(bee)
	f.KnownMasterpatchesNectarList = []ecs.Entity{anyFs}
	f.KnownMasterpatchesPollenList = []ecs.Entity{anyFs}
	f.NectarsourceToGoTo, f.PollensourceToGoTo, f.CurrentFoodsource = anyFs, anyFs, anyFs
	inRange := []ecs.Entity{anyFs}
	col := ecs.NewMap7[comp.Who, comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w).NewEntity(
		&comp.Who{}, &comp.Position{}, &comp.Colony{}, &comp.ColonyStores{}, &comp.ColonyStimuli{}, &comp.ColonyCounts{},
		&comp.ColonyRange{AllPatchesInRangeList: inRange,
			MasterpatchesWithPollenlayersInFlowerAndRangeList: inRange, MasterpatchesWithNectarlayersInFlowerAndRangeList: inRange})

	// [MORN-09] only on Day 365
	setDay(w, 364)
	cr.Update(w)
	if derived.FoodsourcesFile != "_SYSTEM_Example_Farm_Foodsources.txt" {
		t.Fatalf("crop rotation on day 364")
	}

	// [CROP-01] ticks 365: year 1, item 1
	setDay(w, 365)
	cr.Update(w)
	if derived.FoodsourcesFile != "small_Foodsources.txt" {
		t.Errorf("FoodsourcesFile %q, want small_Foodsources.txt", derived.FoodsourcesFile)
	}
	// [CROP-03] only the foodsources of the new file, with new who numbers
	n, masters, _ := foodsourceSummary(w)
	if !maps.Equal(masters, map[string]int{"Scrub": 1, "Garden": 1}) || n != 3 {
		t.Errorf("%d foodsources, masterpatches %v; want 3 layers, one Scrub and one Garden masterpatch", n, masters)
	}
	minWho := int64(-1)
	q := ecs.NewFilter2[comp.Who, comp.Foodsource](w).Query()
	for q.Next() {
		who, _ := q.Get()
		if minWho < 0 || who.Who < minWho {
			minWho = who.Who
		}
	}
	if minWho <= maxWhoBefore {
		t.Errorf("new foodsources start at who %d, want more than %d", minWho, maxWhoBefore)
	}
	foodsources := ecs.NewFilter1[comp.Foodsource](w).Query()
	if got := foodsources.Count(); got != 3 {
		t.Errorf("%d foodsources, want 3", got)
	}
	foodsources.Close()

	// [CROP-02] the bee forgets her foodsources
	f = ecs.NewMap1[comp.Forager](w).Get(bee)
	if len(f.KnownMasterpatchesNectarList)+len(f.KnownMasterpatchesPollenList) != 0 ||
		!f.NectarsourceToGoTo.IsZero() || !f.PollensourceToGoTo.IsZero() || !f.CurrentFoodsource.IsZero() {
		t.Errorf("bee still knows foodsources: %+v", *f)
	}
	// the colony empties its in-flower lists, but keeps allPatchesInRangeList
	r := ecs.NewMap1[comp.ColonyRange](w).Get(col)
	if len(r.MasterpatchesWithPollenlayersInFlowerAndRangeList)+len(r.MasterpatchesWithNectarlayersInFlowerAndRangeList) != 0 {
		t.Errorf("in-flower lists not empty: %+v", *r)
	}
	if !slices.Equal(r.AllPatchesInRangeList, inRange) || w.Alive(anyFs) {
		t.Errorf("allPatchesInRangeList must keep the dead foodsource")
	}
	// [FRNG] reads the dead foodsources: NetLogo runtime error, gobees panics
	founder := newColonyFounder(w, rand.New(ecs.GetResource[resource.Rand](w)))
	expectPanic(t, "FoodsourcesInFlowerAndRangeProc after a crop rotation", func() { founder.FoodsourcesInFlowerAndRangeProc(col) })

	// [SPEC-06], [SPEC-07] for B_terrestris: Scrub and Garden are nest habitats
	sq := ecs.NewFilter1[comp.Species](w).Query()
	for sq.Next() {
		sp := sq.Get()
		if len(sp.NestsiteFoodsourceList) != 2 || sp.NestSiteArea != 150000 {
			t.Errorf("%s: %d nest sites, area %v; want 2 and 150000", sp.Name, len(sp.NestsiteFoodsourceList), sp.NestSiteArea)
		}
		if want := 0.5 * 0.001 * sp.DevWeightPupationMin_mg * sp.Batchsize / sp.PollenToBodymassFactor; sp.MinPollenStore_g != want {
			t.Errorf("minPollenStore_g %v, want %v", sp.MinPollenStore_g, want)
		}
	}

	// ticks 730: year 2, item 0, the example farm again
	setDay(w, 730)
	cr.Update(w)
	if n, _, _ := foodsourceSummary(w); derived.FoodsourcesFile != "_SYSTEM_Example_Farm_Foodsources.txt" || n != nBefore {
		t.Errorf("year 2: FoodsourcesFile %q with %d foodsources, want the example farm with %d", derived.FoodsourcesFile, n, nBefore)
	}
}

func TestCropRotationDefaultList(t *testing.T) {
	// finding 3 (see trans/netlogo-findings.md#finding-3): "" gives [""] with length 1, so no crop rotation
	a, cr := cropWorld(t, "")
	w := a.World
	nBefore, _, maxWho := foodsourceSummary(w)
	setDay(w, 365)
	cr.Update(w)
	if n, _, m := foodsourceSummary(w); n != nBefore || m != maxWho {
		t.Errorf("foodsources changed without crop rotation")
	}
}
