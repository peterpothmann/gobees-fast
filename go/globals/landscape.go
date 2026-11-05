package globals

import (
	"github.com/mlange-42/ark/ecs"
)

// Landscape holds the globals that CreateFoodsourcesProc [FSRC] sets.
type Landscape struct {
	Scaling_NLpatches_per_m float64 // NetLogo patches per metre, 1 / line 1 of the foodsources file, 8 decimals [FSRC-02]
	// FoodsourcesFile is the landscape file of the current year [PAR-05], changed by CropRotationProc [CROP-01]. Until
	// step F14 a field of the derived globals (params.Derived.FoodsourcesFile holds the one of the start).
	FoodsourcesFile string
	// Masterpatches holds all masterpatches in NetLogo's who order, set by CreateLayersProc and replaced at a crop
	// rotation. gobees-fast keeps no who numbers (step F12), the model reads only this order: the nest sites
	// [SPEC-06] are sorted by who.
	Masterpatches []ecs.Entity
}

// NestSites holds the values of the species that depend on the landscape, one entry per species, in the order of
// params.Input.Species (step F11). NetLogo: species-own, set in CreateSpeciesProc [SPEC-04], [SPEC-06] and again in
// CropRotationProc.
type NestSites struct {
	BySpecies []SpeciesNestSites
}

// SpeciesNestSites are the nest sites of one species in the current landscape.
type SpeciesNestSites struct {
	NestHabitatsList       []string     // expanded by NestHabitatsListREP [SPEC-04], kept at a crop rotation
	NestsiteFoodsourceList []ecs.Entity // [SPEC-06] masterpatches, sorted by who (an agentset in NetLogo)
	NestSiteArea           float64      // [SPEC-06] summed area of the nest sites
}
