package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// Helpers shared by the brood systems of step 4b.

// askColonies returns all colonies in random order. NetLogo: ask colonies (decision C in trans/data-model.md).
func askColonies(filter *ecs.Filter1[comp.Colony], rng *rand.Rand) []ecs.Entity {
	result := []ecs.Entity{}
	query := filter.Query()
	for query.Next() {
		result = append(result, query.Entity())
	}
	util.Shuffle(rng, result)
	return result
}

// askAllBees returns all bees in random order. NetLogo: ask bees (decision C).
func askAllBees(filter *ecs.Filter1[comp.Bee], rng *rand.Rand) []ecs.Entity {
	result := []ecs.Entity{}
	query := filter.Query()
	for query.Next() {
		result = append(result, query.Entity())
	}
	util.Shuffle(rng, result)
	return result
}

// askBees returns the bees for which keep is true in random order. NetLogo: ask bees with [ ... ] (decision C).
// The filter selects by component (all bees), keep by value, e.g. the stage (decision A).
func askBees(filter *ecs.Filter1[comp.Bee], rng *rand.Rand, keep func(b *comp.Bee) bool) []ecs.Entity {
	result := []ecs.Entity{}
	query := filter.Query()
	for query.Next() {
		if keep(query.Get()) {
			result = append(result, query.Entity())
		}
	}
	util.Shuffle(rng, result)
	return result
}

// beesByColony groups the bees for which keep is true by their colony, in storage order.
// NetLogo: bees with [ ... and colonyID = myColony ], evaluated once for all colonies. Valid as long as no bee
// is created, dies or changes its colony or the property that keep tests.
func beesByColony(filter *ecs.Filter1[comp.Bee], keep func(b *comp.Bee) bool) map[ecs.Entity][]ecs.Entity {
	result := map[ecs.Entity][]ecs.Entity{}
	query := filter.Query()
	for query.Next() {
		if b := query.Get(); !b.ColonyID.IsZero() && keep(b) {
			result[b.ColonyID] = append(result[b.ColonyID], query.Entity())
		}
	}
	return result
}
