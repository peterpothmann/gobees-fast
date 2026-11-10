//go:build ui || windows || (darwin && cgo)

package ui

import (
	"cmp"
	"image/color"
	"math"
	"slices"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// daysPerYear is the length of the model year (package day).
const daysPerYear = 365

// flowerPatch is one flower patch of the map with all its layers.
type flowerPatch struct {
	Entity    ecs.Entity // the flower patch
	X, Y      float64    // NetLogo patch coordinates
	Radius    float64    // NetLogo patches
	PatchType string
	Area      float64    // [m2]
	Colour    color.RGBA // habitat colour (habitatColour)
	Layers    []flowerLayer
}

// flowerLayer is one layer: one flower species on one flower patch.
type flowerLayer struct {
	Source            comp.Source
	Index             int    // position in snapshot.Layers
	FlowerSpecies     string // name of the flower species, "" for a flower patch without flower species
	StartDay, StopDay float64
	NectarMax         float64 // [µl]
	PollenMax         float64 // [g]
}

// habitat is one entry of the map legend.
type habitat struct {
	Name    string
	Colour  color.RGBA
	Patches int
}

// landscape is a snapshot of the static model data, taken after setup and after each crop rotation.
type landscape struct {
	Patches  []flowerPatch
	index    map[ecs.Entity]int  // flower patch -> index in Patches
	layers   map[comp.Source]int // layer -> flowerLayer.Index
	Habitats []habitat
	Scale    float64 // NetLogo patches per metre (globals.Landscape.Scale)

	MaxForagingRange float64 // [m] (constants.foraging.maxForagingRange)
	DetectionLambda  float64 // [1/m] (parameters.detectionLambda)
	BadgerRange      float64 // [m] (constants.badgers.range)
}

func newLandscape(w *ecs.World) *landscape {
	constants := ecs.GetResource[params.Constants](w)
	l := &landscape{
		Scale:            ecs.GetResource[globals.Landscape](w).Scale,
		MaxForagingRange: constants.Foraging.MaxForagingRange,
		DetectionLambda:  ecs.GetResource[params.Parameters](w).DetectionLambda,
		BadgerRange:      constants.Badgers.Range,
	}
	var flowerNames []string
	if f := ecs.GetResource[params.Input](w).Flowers; f != nil {
		flowerNames = f.Names
	}

	query := ecs.NewFilter2[comp.Position, comp.FlowerPatch](w).Query()
	for query.Next() {
		pos, fp := query.Get()
		p := flowerPatch{
			Entity:    query.Entity(),
			X:         pos.X,
			Y:         pos.Y,
			Radius:    fp.Radius * l.Scale,
			PatchType: fp.PatchType,
			Area:      fp.Area,
		}
		for i := range fp.Layers {
			ly := &fp.Layers[i]
			name := ""
			if ly.Flower >= 0 && ly.Flower < len(flowerNames) {
				name = flowerNames[ly.Flower]
			}
			p.Layers = append(p.Layers, flowerLayer{
				Source:        comp.Source{Patch: query.Entity(), Layer: i},
				FlowerSpecies: name,
				StartDay:      float64(ly.StartDay),
				StopDay:       float64(ly.StopDay),
				NectarMax:     ly.NectarMax,
				PollenMax:     ly.PollenMax,
			})
		}
		l.Patches = append(l.Patches, p)
	}
	counts := map[string]int{}
	for i := range l.Patches {
		p := &l.Patches[i]
		c := habitatColour(p.PatchType)
		p.Colour = c
		if counts[p.PatchType] == 0 {
			l.Habitats = append(l.Habitats, habitat{Name: p.PatchType, Colour: c})
		}
		counts[p.PatchType]++
	}
	for i := range l.Habitats {
		l.Habitats[i].Patches = counts[l.Habitats[i].Name]
	}
	slices.SortFunc(l.Habitats, func(a, b habitat) int { return cmp.Compare(b.Patches, a.Patches) })
	// Draw large patches first, so small ones stay visible and clickable on top.
	slices.SortFunc(l.Patches, func(a, b flowerPatch) int { return cmp.Compare(b.Radius, a.Radius) })
	l.index, l.layers = map[ecs.Entity]int{}, map[comp.Source]int{}
	for i := range l.Patches {
		l.index[l.Patches[i].Entity] = i
		slices.SortFunc(l.Patches[i].Layers, func(a, b flowerLayer) int {
			return cmp.Compare(a.FlowerSpecies, b.FlowerSpecies)
		})
		for j := range l.Patches[i].Layers {
			ly := &l.Patches[i].Layers[j]
			ly.Index = len(l.layers)
			l.layers[ly.Source] = ly.Index
		}
	}

	return l
}

// patch returns the flower patch of the entity patch, or nil.
func (l *landscape) patch(patch ecs.Entity) *flowerPatch {
	i, ok := l.index[patch]
	if !ok {
		return nil
	}
	return &l.Patches[i]
}

// layersInFlower counts the layers of the patch that offer nectar or pollen on the day of the snapshot.
func (p *flowerPatch) layersInFlower(s *snapshot) int {
	n := 0
	for i := range p.Layers {
		if s.inFlower(p.Layers[i].Index) {
			n++
		}
	}
	return n
}

// reach describes how a colony at a nest position reaches a flower patch.
type reach struct {
	Centre   float64 // distance nest to patch centre [m]
	Edge     float64 // distance nest to patch edge [m], 0 if the nest lies inside the patch
	InRange  bool    // patch centre within the foraging range
	FindProb float64 // probability that one search detects the patch
}

// reachFrom computes the reach of patch p from a nest at (x, y) in NetLogo coordinates.
//
// For display only, because the test nest is not part of the model. It mirrors the range condition of the
// model (patch centre within maxForagingRange) and the detection probability (exp(detectionLambda × distance to
// the patch edge)).
func (l *landscape) reachFrom(p *flowerPatch, x, y float64) reach {
	d := math.Hypot(p.X-x, p.Y-y)
	r := reach{Centre: d / l.Scale}
	r.InRange = d <= l.MaxForagingRange*l.Scale
	r.Edge = max(r.Centre-p.Radius/l.Scale, 0)
	if r.InRange && r.Edge <= l.MaxForagingRange {
		r.FindProb = math.Exp(l.DetectionLambda * r.Edge)
	}
	return r
}
