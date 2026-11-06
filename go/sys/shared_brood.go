package sys

import (
	"math/rand/v2"

	"gobees/comp"

	"github.com/mlange-42/ark/ecs"
)

// Helpers that collect the colonies and bees a part asks.
//
// askAllBees, askBees and beesByColony fill a buffer of the calling system instead of a new list. The
// caller keeps the returned list as its buffer for the next call. A list is valid until the next call with the same
// buffer.

// askColonies returns all colonies in random order. NetLogo: ask colonies.
func askColonies(filter *ecs.Filter1[comp.Colony], rng *rand.Rand) []ecs.Entity {
	result := []ecs.Entity{}
	query := filter.Query()
	for query.Next() {
		result = append(result, query.Entity())
	}
	shuffle(rng, result)
	return result
}

// askAllBees returns all bees in random order, in buf. NetLogo: ask bees.
func askAllBees(buf []ecs.Entity, filter *ecs.Filter1[comp.Bee], rng *rand.Rand) []ecs.Entity {
	result := buf[:0]
	query := filter.Query()
	for query.Next() {
		result = append(result, query.Entity())
	}
	shuffle(rng, result)
	return result
}

// askBees returns the bees for which keep is true in random order, in buf. The filter selects by component
// (all bees), keep by value, e.g. the stage. NetLogo: ask bees with [ ... ].
func askBees(buf []ecs.Entity, filter *ecs.Filter1[comp.Bee], rng *rand.Rand, keep func(b *comp.Bee) bool) []ecs.Entity {
	result := buf[:0]
	query := filter.Query()
	for query.Next() {
		if keep(query.Get()) {
			result = append(result, query.Entity())
		}
	}
	shuffle(rng, result)
	return result
}

// beesByColony groups the bees for which keep is true by their colony, in storage order, in groups (nil gives a
// new map). NetLogo: bees with [ ... and colonyID = myColony ], evaluated once for all colonies. Valid as long as
// no bee is created, dies or changes its colony or the property that keep tests.
func beesByColony(groups map[ecs.Entity][]ecs.Entity, filter *ecs.Filter1[comp.Bee], keep func(b *comp.Bee) bool) map[ecs.Entity][]ecs.Entity {
	if groups == nil {
		groups = map[ecs.Entity][]ecs.Entity{}
	}
	for col, bees := range groups {
		groups[col] = bees[:0]
	}
	query := filter.Query()
	for query.Next() {
		if b := query.Get(); !b.Colony.IsZero() && keep(b) {
			groups[b.Colony] = append(groups[b.Colony], query.Entity())
		}
	}
	// colonies without bees today, e.g. dead ones, would otherwise keep their lists for the rest of the run
	for col, bees := range groups {
		if len(bees) == 0 {
			delete(groups, col)
		}
	}
	return groups
}
