package sys

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CreateFoodsources reads the foodsources file, merges small hedges and splits every flower patch
// into one foodsource layer per flower species [INIT-06].
//
// Not translated: the map image [FSRC-01], display settings, farmland? [FSRC-07], the brush [FSRC-11]
// and the stewardship area totals [LAYR-12].
//
// NetLogo: CreateFoodsourcesProc, MergeHedgesProc, CreateLayersProc.
type CreateFoodsources struct {
	landscape globals.Landscape
}

func (s *CreateFoodsources) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	dir := ecs.GetResource[params.WorkingDirectory](w)
	world := ecs.GetResource[params.World](w)
	derived := ecs.GetResource[globals.Derived](w)
	who := ecs.GetResource[globals.WhoCounter](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// [FSRC-02] to [FSRC-04]
	metresPerPatch, patches, err := readFoodsourcesFile(filepath.Join(dir.Path, derived.FoodsourcesFile))
	if err != nil {
		panic(err) // [FSRC-09] NetLogo: message and Setup continues without foodsources
	}
	s.landscape.Scaling_NLpatches_per_m = precision(1/metresPerPatch, 8)

	// [FSRC-04] to [FSRC-08] every flower patch is created and gets a who number, also if it dies right after.
	kept := make([]*flowerPatch, 0, len(patches))
	for i := range patches {
		p := &patches[i]
		p.who = who.Next()
		p.Radius_m = math.Sqrt(p.Area_sqm / math.Pi) // [FSRC-05]
		if !insideMapArea(file.MapAreaIncluded, p.X, p.Y, world) {
			continue // [FSRC-08]
		}
		kept = append(kept, p)
	}

	// [FSRC-10] Merge small hedges into the nearest non-hedge flower patch of at least 6500 m2, and merge all hedges linked to the same patch into the largest of them [HEDG], finding 14 (see trans/netlogo-findings.md#finding-14).
	if file.MergeHedges {
		MergeHedgesProc(kept, rng)
	}

	// [FSRC-12]
	var flowers *flowerSpeciesTable
	if file.FlowerspeciesFile != "No Input File" {
		t, err := readFlowerspeciesFile(filepath.Join(dir.Path, file.FlowerspeciesFile))
		if err != nil {
			panic(err) // NetLogo: message, no layers are created
		}
		flowers = &t
	}
	if err := CreateLayersProc(w, kept, flowers, file, derived.CSS_OptionsList, who, rng); err != nil {
		panic(err)
	}

	// [FSRC-13] output only
	// [INIT-15] assertion: every foodsource must have a habitat type
	query := ecs.NewFilter1[comp.Foodsource](w).Query()
	s.landscape.TotalFoodSources = query.Count()
	undefined := false
	for query.Next() {
		if query.Get().PatchType == "undefined" {
			undefined = true
		}
	}
	if undefined {
		panic("Undefined habitat patch!")
	}
	ecs.AddResource(w, &s.landscape)
}

func (s *CreateFoodsources) Update(w *ecs.World) {}

func (s *CreateFoodsources) Finalize(w *ecs.World) {}

func readFoodsourcesFile(path string) (float64, []flowerPatch, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, nil, fmt.Errorf("FoodsourcesFile: %w", err)
	}
	defer f.Close()
	m, p, err := parseFoodsourcesFile(f)
	if err != nil {
		return 0, nil, fmt.Errorf("FoodsourcesFile %s: %w", path, err)
	}
	return m, p, nil
}

func readFlowerspeciesFile(path string) (flowerSpeciesTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return flowerSpeciesTable{}, fmt.Errorf("FlowerspeciesFile: %w", err)
	}
	defer f.Close()
	t, err := parseFlowerspeciesFile(f)
	if err != nil {
		return t, fmt.Errorf("FlowerspeciesFile %s: %w", path, err)
	}
	return t, nil
}

// precision rounds x to the given number of decimals. NetLogo: precision.
func precision(x float64, places int) float64 {
	f := math.Pow(10, float64(places))
	return math.Round(x*f) / f
}

// insideMapArea reproduces the MapAreaIncluded conditions [FSRC-08] (max-pxcor / 2 = 150, max-pycor / 2 = 105).
func insideMapArea(area string, x, y float64, world *params.World) bool {
	hx, hy := world.MaxPxcor/2, world.MaxPycor/2
	switch area {
	case "top left quarter":
		return !(x > hx || y < hy)
	case "top right quarter":
		return !(x <= hx || y < hy)
	case "bottom left quarter":
		return !(x > hx || y >= hy)
	case "bottom right quarter":
		return !(x <= hx || y >= hy)
	case "top half":
		return !(y < hy)
	case "bottom half":
		return !(y >= hy)
	case "left half":
		return !(x > hx)
	case "right half":
		return !(x <= hx)
	}
	return true // "complete" and any other value
}

// MergeHedgesProc links every hedge to its nearest non-hedge flower patch of at least 6500 m2 and merges
// all hedges linked to the same patch into the largest of them [HEDG], finding 14 (see trans/netlogo-findings.md#finding-14).
func MergeHedgesProc(patches []*flowerPatch, rng *rand.Rand) {
	const areaTH_sqm = 6500 // [HEDG-01]

	hedges := []*flowerPatch{}
	fields := []*flowerPatch{}
	for _, p := range patches {
		if p.PatchType == "Hedgerow" {
			hedges = append(hedges, p)
		} else if p.Area_sqm >= areaTH_sqm {
			fields = append(fields, p)
		}
	}

	// [HEDG-02] largest first. ORDER: sort-on breaks ties randomly (finding 14, see trans/netlogo-findings.md#finding-14): shuffle, then stable sort.
	rng.Shuffle(len(hedges), func(i, j int) { hedges[i], hedges[j] = hedges[j], hedges[i] })
	slices.SortStableFunc(hedges, func(a, b *flowerPatch) int {
		switch {
		case a.Area_sqm > b.Area_sqm:
			return -1
		case a.Area_sqm < b.Area_sqm:
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
			d := util.Distance(h.X, h.Y, f.X, f.Y)
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
			master.Area_sqm += l.hedge.Area_sqm
			master.NectarMax_myl += l.hedge.NectarMax_myl // SIMPLIFY: always 0 + 0 here [HEDG-04]
			master.PollenMax_g += l.hedge.PollenMax_g     // SIMPLIFY: always 0 + 0 here [HEDG-04]
			l.hedge.Area_sqm = 0
		}
	}

	// [HEDG-05] new radius of all hedges
	for _, h := range hedges {
		h.Radius_m = math.Sqrt(h.Area_sqm / math.Pi)
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

// CreateLayersProc splits every flower patch into one foodsource per flower species [LAYR].
func CreateLayersProc(w *ecs.World, patches []*flowerPatch, flowers *flowerSpeciesTable, file *params.ParameterFile,
	cssOptions []string, who *globals.WhoCounter, rng *rand.Rand) error {

	mapper := ecs.NewMap5[comp.Who, comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)

	// [LAYR-01]
	minNectSize_myl := file.FoodSourceLimit * 180
	minPolsize_g := file.FoodSourceLimit * 0.05

	newFoodsource := func(p *flowerPatch, whoNum int64, fs comp.Foodsource) ecs.Entity {
		return mapper.NewEntity(&comp.Who{Who: whoNum}, &comp.Position{X: p.X, Y: p.Y}, &fs,
			&comp.FoodsourceState{}, &comp.Layer{})
	}
	layerMapper := ecs.NewMap1[comp.Layer](w)

	// [LAYR-03] ORDER: MATTERS. The who numbers of the layers follow the random order of ask (finding 16, see trans/netlogo-findings.md#finding-16).
	toSplit := []*flowerPatch{}
	for _, p := range patches {
		if flowers != nil && len(p.FlowerSpeciesList) > 0 {
			toSplit = append(toSplit, p)
		}
	}
	rng.Shuffle(len(toSplit), func(i, j int) { toSplit[i], toSplit[j] = toSplit[j], toSplit[i] })

	for _, p := range toSplit {
		var master ecs.Entity
		occurrence := "Patch"
		// [LAYR-04] ORDER: irrelevant, foreach over a list.
		for _, e := range p.FlowerSpeciesList {
			if slices.Contains(cssOptions, e.Name) {
				occurrence = e.Name
				continue
			}
			// [LAYR-05] hatch: a new who number, all values of the flower patch, then the flower data.
			whoNum := who.Next()
			data, err := flowers.get(e.Name)
			if err != nil {
				return err
			}
			if !e.HasValue {
				return fmt.Errorf("flower patch %v: %s has no value", p.Id_Beescout, e.Name)
			}
			relAbundance := file.AbundanceBoost * e.Value
			fs := comp.Foodsource{
				Id_Beescout:                   p.Id_Beescout,
				PatchType:                     p.PatchType,
				Area_sqm:                      p.Area_sqm,
				Radius_m:                      p.Radius_m,
				PatchInfo:                     p.PatchInfo,
				Occurrence:                    occurrence,
				FlowerSpecies:                 e.Name,
				RelativeAbundance:             e.Value,
				PollenMax_g:                   p.Area_sqm * relAbundance * data.Pollen_g_per_flower,
				NectarFlowerVolume_myl:        data.Nectar_ml_per_flower * 1000,
				NectarConcentration_mol_per_l: data.Concentration_mol_per_l,
				ProteinPollenProp:             data.ProteinPollenProp,
				StartDay:                      data.StartDay,
				StopDay:                       data.StopDay,
				CorollaDepth_mm:               data.CorollaDepth_mm,
				InterFlowerTime_s:             data.IntFlowerTime_s,
			}
			fs.NectarMax_myl = p.Area_sqm * relAbundance * fs.NectarFlowerVolume_myl

			// [LAYR-06]
			if file.MinSizeFoodSources && fs.NectarMax_myl < minNectSize_myl {
				fs.NectarMax_myl = 0
			}
			if file.MinSizeFoodSources && fs.PollenMax_g < minPolsize_g {
				fs.PollenMax_g = 0
			}
			// [LAYR-07] the removed layer has used up its who number
			if fs.NectarMax_myl == 0 && fs.PollenMax_g == 0 && file.RemoveEmptyFoodSources &&
				!strings.Contains(p.PatchType, "Crop_") {
				continue
			}

			// [LAYR-08] the first surviving layer is the masterpatch
			layer := newFoodsource(p, whoNum, fs)
			l := layerMapper.Get(layer)
			if master.IsZero() {
				master = layer
				l.Masterpatch = true
			}
			l.MasterpatchID = master
		}
		// [LAYR-09] the flower patch itself is removed (its who number stays used)
	}

	// [LAYR-09] a flower patch with an empty flowerSpeciesList is not split (finding 19, see trans/netlogo-findings.md#finding-19): it stays a masterpatch
	// without food (nectarMax_myl = pollenMax_g = 0, startDay = stopDay = 0).
	// Without a FlowerspeciesFile, no flower patch is split.
	for _, p := range patches {
		if flowers != nil && len(p.FlowerSpeciesList) > 0 {
			continue
		}
		e := newFoodsource(p, p.who, comp.Foodsource{
			Id_Beescout:       p.Id_Beescout,
			PatchType:         p.PatchType,
			Area_sqm:          p.Area_sqm,
			Radius_m:          p.Radius_m,
			PatchInfo:         p.PatchInfo,
			Occurrence:        "Patch",
			FlowerSpecies:     p.PatchType, // flowerSpecies_relativeAbundanceList (list patchType 1)
			RelativeAbundance: 1,
			NectarMax_myl:     p.NectarMax_myl,
			PollenMax_g:       p.PollenMax_g,
		})
		l := layerMapper.Get(e)
		l.Masterpatch = true
		l.MasterpatchID = e
	}

	setLayersInPatchList(w)
	return nil
}

// setLayersInPatchList sets layersInPatchList of EVERY foodsource to all foodsources on the same NetLogo
// patch cell, sorted by who [LAYR-10], finding 15 (see trans/netlogo-findings.md#finding-15). ORDER: irrelevant, each foodsource only changes itself.
func setLayersInPatchList(w *ecs.World) {
	type cell struct{ x, y int }
	type item struct {
		e   ecs.Entity
		who int64
	}
	byCell := map[cell][]item{}

	filter := ecs.NewFilter3[comp.Who, comp.Position, comp.Layer](w)
	query := filter.Query()
	for query.Next() {
		who, pos, _ := query.Get()
		c := cell{util.Pcor(pos.X), util.Pcor(pos.Y)}
		byCell[c] = append(byCell[c], item{query.Entity(), who.Who})
	}

	lists := map[cell][]ecs.Entity{}
	for c, items := range byCell {
		slices.SortFunc(items, func(a, b item) int { return int(a.who - b.who) })
		list := make([]ecs.Entity, len(items))
		for i, it := range items {
			list[i] = it.e
		}
		lists[c] = list
	}

	query = filter.Query()
	for query.Next() {
		_, pos, layer := query.Get()
		layer.LayersInPatchList = lists[cell{util.Pcor(pos.X), util.Pcor(pos.Y)}]
	}
}
