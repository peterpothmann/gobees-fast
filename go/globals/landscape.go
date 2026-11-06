package globals

import (
	"github.com/mlange-42/ark/ecs"
)

// Landscape holds the values of the current landscape.
type Landscape struct {
	Scale float64 // NetLogo patches per metre [1/m], 1 / metres per patch of the landscape file, 8 decimals. NetLogo: Scaling_NLpatches/m
	// File is the landscape file of the current year, changed by the crop rotation (params.Derived.StartLandscape
	// holds the one of the start). NetLogo: FoodsourcesFile.
	File string
	// Masterpatches holds all masterpatches in the order of creation, set by the landscape setup and replaced at a
	// crop rotation. The nest sites are sorted in this order.
	Masterpatches []ecs.Entity
}

// NestSites holds the values of the species that depend on the landscape, one entry per species, in the order of
// params.Input.Species, set in the setup and again at a crop rotation. NetLogo: species-own.
type NestSites struct {
	BySpecies []SpeciesNestSites
}

// SpeciesNestSites are the nest sites of one species in the current landscape.
type SpeciesNestSites struct {
	Habitats []string     // nest habitats, expanded by expandNestHabitats, kept at a crop rotation. NetLogo: nestHabitatsList
	Sites    []ecs.Entity // masterpatches in the order of creation. NetLogo: nestsiteFoodsourceList
	Area     float64      // summed area of the nest sites [m2]. NetLogo: nestSiteArea
}
