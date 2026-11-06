package sys

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// landscapeSetup reads the foodsources file, merges small hedges and splits every flower patch
// into one foodsource layer per flower species [INIT-06].
//
// Not translated: the map image [FSRC-01], display settings, farmland? [FSRC-07], the brush [FSRC-11]
// and the stewardship area totals [LAYR-12].
//
// NetLogo: CreateFoodsourcesProc, MergeHedgesProc, CreateLayersProc.
type landscapeSetup struct {
	landscape globals.Landscape
}

func (s *landscapeSetup) Initialize(w *ecs.World) {
	derived := ecs.GetResource[params.Derived](w)
	createFoodsources(w, derived.StartLandscape, &s.landscape)
	ecs.AddResource(w, &s.landscape)

	// [INIT-15] assertion: every foodsource must have a habitat type. Part of Setup, not of CreateFoodsourcesProc,
	// so it does not run after a crop rotation.
	query := ecs.NewFilter1[comp.Foodsource](w).Query()
	undefined := false
	for query.Next() {
		if query.Get().PatchType == "undefined" {
			undefined = true
		}
	}
	if undefined {
		panic("Undefined habitat patch!")
	}
}

func (s *landscapeSetup) Update(w *ecs.World) {}

func (s *landscapeSetup) Finalize(w *ecs.World) {}

// createFoodsources creates the foodsources of the given foodsources file and sets the landscape globals
// [FSRC]. Called in Setup [INIT-06] and by CropRotationProc [CROP-03].
func createFoodsources(w *ecs.World, foodsourcesFile string, landscape *globals.Landscape) {
	file := ecs.GetResource[params.Parameters](w)
	in := ecs.GetResource[params.Input](w)
	derived := ecs.GetResource[params.Derived](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// [FSRC-02] to [FSRC-04] read by package input, which reads the landscape of the start and every landscape of the
	// crop rotation.
	read, ok := in.Foodsources[foodsourcesFile]
	if !ok {
		panic(fmt.Errorf("landscape file %q was not read by package input", foodsourcesFile))
	}
	landscape.File = foodsourcesFile // the landscape of the current year [PAR-05], [CROP-01]
	landscape.Scale = precision(1/read.MetresPerPatch, 8)
	// copies, because the setup changes them (radius, merged hedges), and a crop rotation can use a file twice
	patches := make([]flowerPatch, len(read.Patches))
	for i := range read.Patches {
		patches[i] = flowerPatch{FlowerPatch: read.Patches[i]}
	}

	// [FSRC-04] to [FSRC-08] every flower patch is created. NetLogo gives it a who number here, before all layers.
	// gobees-fast keeps no who number (step F12), only the order that follows from it (see CreateLayersProc).
	// [FSRC-08] MapAreaIncluded is removed (step F13): the whole landscape file is simulated (NetLogo: "complete"). A
	// part of a map is its own landscape file.
	kept := make([]*flowerPatch, len(patches))
	for i := range patches {
		p := &patches[i]
		p.Radius = math.Sqrt(p.Area / math.Pi) // [FSRC-05]
		kept[i] = p
	}

	// [FSRC-10] Merge small hedges into the nearest non-hedge flower patch of at least 6500 m2, and merge all hedges linked to the same patch into the largest of them [HEDG], finding 14 (see trans/netlogo-findings.md#finding-14).
	if file.MergeHedges {
		mergeHedges(kept, rng)
	}

	// [FSRC-12] in.Flowers is nil if FlowerspeciesFile is "No Input File".
	masterpatches, err := createLayers(w, kept, in.Flowers, file, derived.StewardshipOptions, rng)
	if err != nil {
		panic(err)
	}
	landscape.Masterpatches = masterpatches
	// [FSRC-13] TotalFoodSources: output only, never read, removed in step F3.
}

// flowerPatch is a flower patch of a foodsources file during the setup. In NetLogo it is a foodsource turtle that
// dies in CreateLayersProc [LAYR-09]. gobees keeps it as a plain struct during setup and creates entities only for the
// layers (and for flower patches with an empty list).
type flowerPatch struct {
	params.FlowerPatch         // the line of the file [FSRC-04]
	Radius             float64 // [FSRC-05], [HEDG-05]
	NectarMax          float64 // 0 until CreateLayersProc, see [HEDG-04]
	PollenMax          float64 // 0 until CreateLayersProc, see [HEDG-04]
}

// precision rounds x to the given number of decimals. NetLogo: precision.
func precision(x float64, places int) float64 {
	f := math.Pow(10, float64(places))
	return math.Round(x*f) / f
}

// mergeHedges links every hedge to its nearest non-hedge flower patch of at least 6500 m2 and merges
// all hedges linked to the same patch into the largest of them [HEDG], finding 14 (see trans/netlogo-findings.md#finding-14).
func mergeHedges(patches []*flowerPatch, rng *rand.Rand) {
	const minArea = 6500 // [m2], [HEDG-01]

	hedges := []*flowerPatch{}
	fields := []*flowerPatch{}
	for _, p := range patches {
		if p.PatchType == "Hedgerow" {
			hedges = append(hedges, p)
		} else if p.Area >= minArea {
			fields = append(fields, p)
		}
	}

	// [HEDG-02] largest first. ORDER: sort-on breaks ties randomly (finding 14, see trans/netlogo-findings.md#finding-14): shuffle, then stable sort.
	rng.Shuffle(len(hedges), func(i, j int) { hedges[i], hedges[j] = hedges[j], hedges[i] })
	slices.SortStableFunc(hedges, func(a, b *flowerPatch) int {
		switch {
		case a.Area > b.Area:
			return -1
		case a.Area < b.Area:
			return 1
		}
		return 0
	})

	// [HEDG-03] nearest field per hedge. min-one-of picks a random one among equally near fields.
	type link struct{ field, hedge *flowerPatch }
	links := []link{}
	fieldsWithHedges := []*flowerPatch{}
	for _, h := range hedges {
		var nearest []*flowerPatch
		best := math.Inf(1)
		for _, f := range fields {
			d := distanceBetween(h.X, h.Y, f.X, f.Y)
			if d < best {
				best, nearest = d, []*flowerPatch{f}
			} else if d == best {
				nearest = append(nearest, f)
			}
		}
		if len(nearest) == 0 {
			continue
		}
		field := nearest[rng.IntN(len(nearest))]
		links = append(links, link{field, h})
		fieldsWithHedges = append(fieldsWithHedges, field)
	}
	fieldsWithHedges = removeDuplicates(fieldsWithHedges)

	// [HEDG-04] the first linked hedge of a field is the master hedge, the others add their area and get 0.
	for _, field := range fieldsWithHedges {
		var master *flowerPatch
		for _, l := range links {
			if l.field != field {
				continue
			}
			if master == nil {
				master = l.hedge
				continue
			}
			// [HEDG-04] NetLogo also adds nectarMax_myl and pollenMax_g, which are always 0 here (removed in step F3)
			master.Area += l.hedge.Area
			l.hedge.Area = 0
		}
	}

	// [HEDG-05] new radius of all hedges
	for _, h := range hedges {
		h.Radius = math.Sqrt(h.Area / math.Pi)
	}
}

// removeDuplicates keeps the first occurrence of each element. NetLogo: remove-duplicates.
func removeDuplicates[T comparable](list []T) []T {
	seen := map[T]bool{}
	out := list[:0:0]
	for _, x := range list {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// createLayers splits every flower patch into one foodsource per flower species [LAYR]. It returns the
// masterpatches in NetLogo's who order (step F12): first the flower patches that are not split, in the order of
// patches (they took their who numbers in [FSRC-04], before all layers), then the masterpatches of the split flower
// patches, in the order of their creation.
func createLayers(w *ecs.World, patches []*flowerPatch, flowers *params.FlowerSpeciesTable, file *params.Parameters,
	cssOptions []string, rng *rand.Rand) ([]ecs.Entity, error) {

	mapper := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)

	// [LAYR-01]
	minNectar := file.FoodsourceLimit * 180  // [µl]
	minPollen := file.FoodsourceLimit * 0.05 // [g]

	newFoodsource := func(p *flowerPatch, fs comp.Foodsource) ecs.Entity {
		return mapper.NewEntity(&comp.Position{X: p.X, Y: p.Y}, &fs, &comp.FoodsourceState{}, &comp.Layer{})
	}
	layerMapper := ecs.NewMap1[comp.Layer](w)
	splitMasterpatches := []ecs.Entity{}

	// [LAYR-03] ORDER: MATTERS. The who numbers of the layers follow the random order of ask (finding 16, see trans/netlogo-findings.md#finding-16).
	// gobees-fast keeps no who number (step F12), only the order of creation, which is the who order.
	toSplit := []*flowerPatch{}
	for _, p := range patches {
		if flowers != nil && len(p.Flowers) > 0 {
			toSplit = append(toSplit, p)
		}
	}
	rng.Shuffle(len(toSplit), func(i, j int) { toSplit[i], toSplit[j] = toSplit[j], toSplit[i] })

	for _, p := range toSplit {
		var master ecs.Entity
		// [LAYR-10] the layers of this flower patch, in order of creation (= who order)
		layers := []ecs.Entity{}
		// [LAYR-04] ORDER: irrelevant, foreach over a list. A stewardship option is no flower species and gets no
		// layer. NetLogo stores it as occurrence of the following layers, which nothing reads (removed in step F3).
		for _, e := range p.Flowers {
			if slices.Contains(cssOptions, e.Name) {
				continue
			}
			// [LAYR-05] hatch: a new who number (not kept, step F12), all values of the flower patch, then the flower data.
			data, err := flowers.Get(e.Name)
			if err != nil {
				return nil, err
			}
			if !e.HasValue {
				return nil, fmt.Errorf("flower patch %v: %s has no value", p.ID, e.Name)
			}
			relAbundance := file.AbundanceBoost * e.Value
			fs := comp.Foodsource{
				ID:                  p.ID,
				PatchType:           p.PatchType,
				Area:                p.Area,
				Radius:              p.Radius,
				PollenMax:           p.Area * relAbundance * data.PollenPerFlower,
				NectarPerFlower:     data.NectarPerFlower * 1000,
				NectarConcentration: data.Concentration,
				StartDay:            data.StartDay,
				StopDay:             data.StopDay,
				CorollaDepth:        data.CorollaDepth,
				InterFlowerTime:     data.InterFlowerTime,
			}
			fs.NectarMax = p.Area * relAbundance * fs.NectarPerFlower

			// [LAYR-06]
			if file.LimitFoodsources && fs.NectarMax < minNectar {
				fs.NectarMax = 0
			}
			if file.LimitFoodsources && fs.PollenMax < minPollen {
				fs.PollenMax = 0
			}
			// [LAYR-07] the removed layer has used up its who number (finding 18, no effect without who numbers)
			if fs.NectarMax == 0 && fs.PollenMax == 0 && file.RemoveEmptyFoodsources &&
				!strings.Contains(p.PatchType, "Crop_") {
				continue
			}

			// [LAYR-08] the first surviving layer is the masterpatch
			layer := newFoodsource(p, fs)
			l := layerMapper.Get(layer)
			if master.IsZero() {
				master = layer
				l.IsMasterpatch = true
				splitMasterpatches = append(splitMasterpatches, layer)
			}
			l.Masterpatch = master
			layers = append(layers, layer)
		}
		// [LAYR-10] every layer of the flower patch gets the same list
		for _, layer := range layers {
			layerMapper.Get(layer).PatchLayers = layers
		}
		// [LAYR-09] the flower patch itself is removed
	}

	// [LAYR-09] a flower patch with an empty flowerSpeciesList is not split (finding 19, see trans/netlogo-findings.md#finding-19): it stays a masterpatch
	// without food (nectarMax_myl = pollenMax_g = 0, startDay = stopDay = 0).
	// Without a FlowerspeciesFile, no flower patch is split.
	masterpatches := []ecs.Entity{}
	for _, p := range patches {
		if flowers != nil && len(p.Flowers) > 0 {
			continue
		}
		e := newFoodsource(p, comp.Foodsource{
			ID:        p.ID,
			PatchType: p.PatchType,
			Area:      p.Area,
			Radius:    p.Radius,
			NectarMax: p.NectarMax,
			PollenMax: p.PollenMax,
		})
		l := layerMapper.Get(e)
		l.IsMasterpatch = true
		l.Masterpatch = e
		l.PatchLayers = []ecs.Entity{e} // [LAYR-10]
		masterpatches = append(masterpatches, e)
	}

	// [LAYR-10] layersInPatchList of EVERY foodsource holds all layers of its flower patch (the same masterpatchID),
	// sorted by who. Set above while the layers are created: the layers of one flower patch are created one after the
	// other, so their order of creation is their who order (step F12). NetLogo takes all foodsources on the same
	// NetLogo patch cell, also the layers of another flower patch in this cell (finding 15, see
	// trans/netlogo-findings.md#finding-15). Fixed in step F5. The same lists as NetLogo's if no patch cell holds two
	// flower patches, as in the example farm.
	return append(masterpatches, splitMasterpatches...), nil
}
