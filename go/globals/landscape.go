package globals

import (
	"github.com/mlange-42/ark/ecs"
)

// Landscape holds the globals that CreateFoodsourcesProc [FSRC] sets.
type Landscape struct {
	Scale float64 // NetLogo patches per metre [1/m], 1 / line 1 of the foodsources file, 8 decimals [FSRC-02]. NetLogo: Scaling_NLpatches/m
	// File is the landscape file of the current year [PAR-05], changed by CropRotationProc [CROP-01]. Until
	// step F14 a field of the derived globals (params.Derived.StartLandscape holds the one of the start). NetLogo:
	// FoodsourcesFile.
	File string
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
	Habitats []string     // nest habitats, expanded by expandNestHabitats [SPEC-04], kept at a crop rotation. NetLogo: nestHabitatsList
	Sites    []ecs.Entity // [SPEC-06] masterpatches, sorted by who (an agentset in NetLogo). NetLogo: nestsiteFoodsourceList
	Area     float64      // [SPEC-06] summed area of the nest sites [m2]. NetLogo: nestSiteArea
}
