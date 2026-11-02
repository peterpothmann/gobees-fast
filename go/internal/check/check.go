// Package check holds the checks of the model state that the tests run at the end of every day. They replace the
// assertions of CheckNumbersProc [CHK], which are not part of the model since step F4.
//
// Not kept: [CHK-01] (initial + produced - died = present bees) and the egg accounting of [CHK-04] (eggs laid =
// adults emerged + brood deaths by cause). Both checked only NetLogo's own counters, which the model no longer has.
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
//   - [CHK-03] every egg, larva, pupa, adult worker and adult male belongs to a living colony. CheckNumbersProc
//     compares the totals of the colonies (allEggs, ..., which [UCOL-06] counts from the bees of each colony) with
//     the totals of all bees, which is the same. Unlike these totals of the morning, it holds at any time of the
//     day: a colony dies only when no bee belongs to it [UCOL-11].
//   - [CHK-04] if seasonEnd: no colony is alive on day 364. NetLogo checks produced - died = 0, which is the same,
//     as only CreateColoniesProc creates and only DieProc removes colonies. Needs the end of the season [SEAS-01],
//     so runs without UpdateSeasonalEvents pass false.
func Day(w *ecs.World, seasonEnd bool) error {
	colonies := ecs.NewMap1[comp.Colony](w)
	query := ecs.NewFilter1[comp.Bee](w).Query()
	for query.Next() {
		b := query.Get()
		member := b.Stage != stage.Adult || b.Caste == caste.Worker || b.Caste == caste.Male
		if member && (b.ColonyID.IsZero() || !w.Alive(b.ColonyID) || !colonies.HasAll(b.ColonyID)) {
			err := fmt.Errorf("[CHK-03] bee %v (stage %v, caste %v, number %v) belongs to no living colony",
				query.Entity(), b.Stage, b.Caste, b.Number)
			query.Close()
			return err
		}
	}

	if seasonEnd && ecs.GetResource[globals.Calendar](w).Day == 364 {
		q := ecs.NewFilter1[comp.Colony](w).Query()
		n := q.Count()
		q.Close()
		if n != 0 {
			return fmt.Errorf("[CHK-04] %d colonies alive on day 364", n)
		}
	}
	return nil
}
