package check

import (
	"fmt"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// Day returns an error if the state at the end of a day violates a check, nil otherwise.
//
//   - every egg, larva, pupa, adult worker and adult male belongs to a living colony. It holds at any time of the
//     day: a colony dies only when no bee belongs to it.
//   - every mated bee is an adult queen. The output relies on it: the adult queens are young_queens (adult, not
//     mated) + TotalMatedQueens.
//   - if seasonEnd: no colony is alive on day 364. Needs the end of the season, so runs without seasonalEvents
//     pass false.
func Day(w *ecs.World, seasonEnd bool) error {
	colonies := ecs.NewMap1[comp.Colony](w)
	query := ecs.NewFilter1[comp.Bee](w).Query()
	for query.Next() {
		b := query.Get()
		member := b.Stage != stage.Adult || b.Caste == caste.Worker || b.Caste == caste.Male
		if member && (b.Colony.IsZero() || !w.Alive(b.Colony) || !colonies.HasAll(b.Colony)) {
			err := fmt.Errorf("bee %v (stage %v, caste %v, number %v) belongs to no living colony",
				query.Entity(), b.Stage, b.Caste, b.Number)
			query.Close()
			return err
		}
		if b.Mated && (b.Stage != stage.Adult || b.Caste != caste.Queen) {
			err := fmt.Errorf("bee %v (stage %v, caste %v, number %v) is mated", query.Entity(), b.Stage, b.Caste,
				b.Number)
			query.Close()
			return err
		}
	}

	if seasonEnd && ecs.GetResource[globals.Calendar](w).Day == 364 {
		q := ecs.NewFilter1[comp.Colony](w).Query()
		n := q.Count()
		q.Close()
		if n != 0 {
			return fmt.Errorf("%d colonies alive on day 364", n)
		}
	}
	return nil
}
