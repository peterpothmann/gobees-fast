package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/internal/check"
	"gobees/internal/examplefiles"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// exampleFarm returns the parameters of the example farm, with the input files written to a temporary directory
// (also the second map of the crop rotation experiment).
func exampleFarm(t testing.TB, seed float64) params.Params {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"_SYSTEM_Habitats.csv":                 examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt": examplefiles.ExampleFarmFoodsources,
		"_SYSTEM_Flowerspecies.csv":            examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":      examplefiles.BumbleSpecies,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	file, err := params.ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	p := params.Default()
	p.WorkingDirectory.Path = dir
	p.File = file
	p.Interface.RAND_SEED = seed
	return p
}

func TestDefaultYearCycle(t *testing.T) {
	// The full model over 3 years. check.Day checks the state at the end of every day: the bees of the colonies
	// [CHK-03] and no colony on Day 364 [CHK-04], which needs the end of season [SEAS-01].
	const years = 3
	for seed := 1.0; seed <= 3; seed++ {
		p := exampleFarm(t, seed)
		p.Termination.MaxTicks = 365 * years
		a := Default(p, nil)
		a.Initialize()
		calendar := ecs.GetResource[globals.Calendar](a.World)
		colonies := ecs.NewFilter1[comp.Colony](a.World)
		bees := ecs.NewFilter1[comp.Bee](a.World)

		for step := 1; step <= 365*years; step++ {
			a.Update()
			if err := check.Day(a.World, true); err != nil {
				t.Fatalf("seed %v, step %d: %v", seed, step, err)
			}
			if calendar.Day < 305 {
				continue
			}
			// from Day seasonStop (305) on: no colony, only hibernating queens, which are mated
			q := colonies.Query()
			nColonies := q.Count()
			q.Close()
			active, males, hibernating := 0, 0.0, 0.0
			bq := bees.Query()
			for bq.Next() {
				b := bq.Get()
				if b.Activity != activity.Hibernate || !b.Mated {
					active++
				}
				if b.Caste == caste.Male {
					males += b.Number
				}
				if b.Activity == activity.Hibernate {
					hibernating += b.Number
				}
			}
			if nColonies != 0 || active != 0 || males != 0 {
				t.Fatalf("seed %v, step %d (Day %v): %d colonies, %d bees that do not hibernate, %v males; want 0",
					seed, step, calendar.Day, nColonies, active, males)
			}
			if calendar.Day == 365 {
				t.Logf("seed %v, year %d: %v hibernating queens", seed, step/365, hibernating)
				if hibernating == 0 {
					t.Fatalf("seed %v: no hibernating queens at the end of year %d", seed, step/365)
				}
			}
		}
	}
}
