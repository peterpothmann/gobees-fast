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
// into one foodsource layer per flower species.
//
// NetLogo: CreateFoodsourcesProc, MergeHedgesProc, CreateLayersProc.
type landscapeSetup struct {
	landscape globals.Landscape
}

func (s *landscapeSetup) Initialize(w *ecs.World) {
	derived := ecs.GetResource[params.Derived](w)
	createFoodsources(w, derived.StartLandscape, &s.landscape)
	ecs.AddResource(w, &s.landscape)

	// assertion: every foodsource must have a habitat type. Part of the setup, not of createFoodsources,
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

// createFoodsources creates the foodsources of the given foodsources file and sets the landscape globals.
// Called by landscapeSetup and by cropRotation.rotate.
func createFoodsources(w *ecs.World, foodsourcesFile string, landscape *globals.Landscape) {
	file := ecs.GetResource[params.Parameters](w)
	constants := ecs.GetResource[params.Constants](w)
	in := ecs.GetResource[params.Input](w)
	derived := ecs.GetResource[params.Derived](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// read by package input, which reads the landscape of the start and every landscape of the
	// crop rotation.
	read, ok := in.Foodsources[foodsourcesFile]
	if !ok {
		panic(fmt.Errorf("landscape file %q was not read by package input", foodsourcesFile))
	}
	landscape.File = foodsourcesFile // the landscape of the current year
	landscape.Scale = precision(1/read.MetresPerPatch, 8)
	// copies, because the setup changes them (radius, merged hedges), and a crop rotation can use a file twice
	patches := make([]flowerPatch, len(read.Patches))
	for i := range read.Patches {
		patches[i] = flowerPatch{FlowerPatch: read.Patches[i]}
	}

	// every flower patch gets its radius.
	// The whole landscape file is simulated. A part of a map is its own landscape file.
	kept := make([]*flowerPatch, len(patches))
	for i := range patches {
		p := &patches[i]
		p.Radius = math.Sqrt(p.Area / math.Pi)
		kept[i] = p
	}

	// Merge small hedges into the nearest non-hedge flower patch of at least MinMergeArea, and merge all hedges linked to the
	// same patch into the largest of them.
	if file.MergeHedges {
		mergeHedges(kept, constants.Landscape.MinMergeArea, rng)
	}

	// in.Flowers is nil if FlowerspeciesFile is "No Input File".
	masterpatches, err := createLayers(w, kept, in.Flowers, file, &constants.Landscape, derived.StewardshipOptions, rng)
	if err != nil {
		panic(err)
	}
	landscape.Masterpatches = masterpatches
}

// flowerPatch is a flower patch of a foodsources file during the setup. It is a plain struct. Entities are
// created only for the layers (and for flower patches with an empty list).
//
// NetLogo: the foodsource turtle of a flower patch, which dies in CreateLayersProc.
type flowerPatch struct {
	params.FlowerPatch         // the line of the file
	Radius             float64 // [m]
	NectarMax          float64 // always 0, the food of a flower patch is in its layers
	PollenMax          float64 // always 0, the food of a flower patch is in its layers
}

// precision rounds x to the given number of decimals. NetLogo: precision.
func precision(x float64, places int) float64 {
	f := math.Pow(10, float64(places))
	return math.Round(x*f) / f
}

// mergeHedges links every hedge to its nearest non-hedge flower patch of at least minArea [m2] and merges
// all hedges linked to the same patch into the largest of them.
func mergeHedges(patches []*flowerPatch, minArea float64, rng *rand.Rand) {

	hedges := []*flowerPatch{}
	fields := []*flowerPatch{}
	for _, p := range patches {
		if p.PatchType == "Hedgerow" {
			hedges = append(hedges, p)
		} else if p.Area >= minArea {
			fields = append(fields, p)
		}
	}

	// largest first. ORDER: ties are broken randomly, as by NetLogo's sort-on: shuffle, then stable sort.
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

	// nearest field per hedge. A random one among equally near fields, as by NetLogo's min-one-of.
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

	// the first linked hedge of a field is the master hedge, the others add their area and get 0.
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
			master.Area += l.hedge.Area
			l.hedge.Area = 0
		}
	}

	// new radius of all hedges
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

// createLayers splits every flower patch into one foodsource per flower species. It returns the
// masterpatches in NetLogo's who order: first the flower patches that are not split, in the order of
// patches (in NetLogo they get their who numbers before all layers), then the masterpatches of the split flower
// patches, in the order of their creation.
func createLayers(w *ecs.World, patches []*flowerPatch, flowers *params.FlowerSpeciesTable, file *params.Parameters,
	c *params.LandscapeConstants, cssOptions []string, rng *rand.Rand) ([]ecs.Entity, error) {

	mapper := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)

	minNectar := file.FoodsourceLimit * c.TripNectar // [µl]
	minPollen := file.FoodsourceLimit * c.TripPollen // [g]

	newFoodsource := func(p *flowerPatch, fs comp.Foodsource) ecs.Entity {
		return mapper.NewEntity(&comp.Position{X: p.X, Y: p.Y}, &fs, &comp.FoodsourceState{}, &comp.Layer{})
	}
	layerMapper := ecs.NewMap1[comp.Layer](w)
	splitMasterpatches := []ecs.Entity{}

	// ORDER: MATTERS. The flower patches are split in random order, as NetLogo asks them in random order.
	// The order of creation of the layers is NetLogo's who order.
	toSplit := []*flowerPatch{}
	for _, p := range patches {
		if flowers != nil && len(p.Flowers) > 0 {
			toSplit = append(toSplit, p)
		}
	}
	rng.Shuffle(len(toSplit), func(i, j int) { toSplit[i], toSplit[j] = toSplit[j], toSplit[i] })

	for _, p := range toSplit {
		var master ecs.Entity
		// the layers of this flower patch, in order of creation (= who order)
		layers := []ecs.Entity{}
		// ORDER: irrelevant, foreach over a list. A stewardship option is no flower species and gets no
		// layer.
		for _, e := range p.Flowers {
			if slices.Contains(cssOptions, e.Name) {
				continue
			}
			// a new layer: all values of the flower patch, then the flower data.
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
				Flower:              slices.Index(flowers.Names, e.Name), // the row of flowers.Get
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

			if file.LimitFoodsources && fs.NectarMax < minNectar {
				fs.NectarMax = 0
			}
			if file.LimitFoodsources && fs.PollenMax < minPollen {
				fs.PollenMax = 0
			}
			if fs.NectarMax == 0 && fs.PollenMax == 0 && file.RemoveEmptyFoodsources &&
				!strings.Contains(p.PatchType, "Crop_") {
				continue
			}

			// the first surviving layer is the masterpatch
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
		// every layer of the flower patch gets the same list
		for _, layer := range layers {
			layerMapper.Get(layer).PatchLayers = layers
		}
		// the flower patch itself is removed
	}

	// a flower patch with an empty flower species list is not split: it stays a masterpatch
	// without food (NectarMax = PollenMax = 0, StartDay = StopDay = 0).
	// Without a FlowerspeciesFile, no flower patch is split.
	masterpatches := []ecs.Entity{}
	for _, p := range patches {
		if flowers != nil && len(p.Flowers) > 0 {
			continue
		}
		e := newFoodsource(p, comp.Foodsource{
			ID:        p.ID,
			Flower:    -1,
			PatchType: p.PatchType,
			Area:      p.Area,
			Radius:    p.Radius,
			NectarMax: p.NectarMax,
			PollenMax: p.PollenMax,
		})
		l := layerMapper.Get(e)
		l.IsMasterpatch = true
		l.Masterpatch = e
		l.PatchLayers = []ecs.Entity{e}
		masterpatches = append(masterpatches, e)
	}

	// PatchLayers of EVERY foodsource holds all layers of its flower patch (the same Masterpatch), in the order of
	// creation (NetLogo's who order). Set above while the layers are created. NetLogo takes all foodsources on the
	// same NetLogo patch cell, also the layers of another flower patch in this cell. The lists equal NetLogo's if no
	// patch cell holds two flower patches, as in the example farm.
	return append(masterpatches, splitMasterpatches...), nil
}
