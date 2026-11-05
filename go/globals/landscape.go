package globals

import "github.com/mlange-42/ark/ecs"

// WhoCounter hands out NetLogo who numbers. NetLogo numbers all turtles of all breeds with one counter and
// never reuses a number within a run. A turtle that dies right after its creation has used up its number
// (finding 18, see trans/netlogo-findings.md#finding-18), so Next must be called for it too.
type WhoCounter struct {
	next int64
}

// Next returns the who number of the next created turtle.
func (c *WhoCounter) Next() int64 {
	n := c.next
	c.next++
	return n
}

// Landscape holds the globals that CreateFoodsourcesProc [FSRC] sets.
type Landscape struct {
	Scaling_NLpatches_per_m float64 // NetLogo patches per metre, 1 / line 1 of the foodsources file, 8 decimals [FSRC-02]
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
