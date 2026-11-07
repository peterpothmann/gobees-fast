package schedule

import (
	"reflect"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/input"
	"gobees/internal/check"
	"gobees/obs"
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark/ecs"
)

// exampleFarm returns the parameters of the example farm (the run file built into package input).
func exampleFarm(t testing.TB, seed float64) params.Params {
	t.Helper()
	p := input.ExampleFarm()
	p.Seed = int64(seed)
	return p
}

func TestDefaultYearCycle(t *testing.T) {
	// The full model over 3 years. check.Day checks the state at the end of every day: the bees of the colonies
	// and no colony on Day 364, which needs the end of season.
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
			active, males, hibernating := 0, 0, 0
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
				t.Fatalf("seed %v, step %d (Day %v): %d colonies, %d bees that do not hibernate, %d males; want 0",
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

// dailyRows is an app system that keeps the daily output (all columns) of every day.
type dailyRows struct {
	o    obs.DailyOutputExtended
	rows [][]float64
}

func (d *dailyRows) Initialize(w *ecs.World) {
	d.o.Columns = obs.AllColumns()
	d.o.Initialize(w)
}

func (d *dailyRows) Update(w *ecs.World) {
	d.o.Update(w)
	d.rows = append(d.rows, slices.Clone(d.o.Values(w)))
}

func (d *dailyRows) Finalize(w *ecs.World) {}

// TestHooks: with sys.Hooks the parts are reported in schedule order and every bee turn after its turn, in the order
// of the personal times, and the run is the same as without them (the hooks read the world).
func TestHooks(t *testing.T) {
	const days = 150
	// run runs the example farm, with hooks made by newHooks for its world if not nil.
	run := func(newHooks func(w *ecs.World) *sys.Hooks) [][]float64 {
		p := exampleFarm(t, 1)
		p.Termination.MaxTicks = days
		a := Default(p, nil)
		if newHooks != nil {
			ecs.AddResource(a.World, newHooks(a.World)) // after Default (it resets the app), before Initialize
		}
		out := &dailyRows{}
		a.AddSystem(out)
		a.Run()
		return out.rows
	}

	var parts []string
	turns, day, lastTime := 0, 0, 0.0
	withHooks := run(func(w *ecs.World) *sys.Hooks {
		bees := ecs.NewFilter1[comp.Bee](w)
		return &sys.Hooks{
			AfterPart: func(system, part string) {
				if part == "tick" { // a new day
					day++
					lastTime = 0
				}
				if day == 1 {
					parts = append(parts, system+"."+part)
				}
				q := bees.Query() // reading the world must not change the run
				q.Count()
				q.Close()
			},
			AfterTurn: func(bee ecs.Entity) {
				turns++
				daytime := ecs.GetResource[globals.Daytime](w).Time
				if daytime < lastTime {
					t.Fatalf("day %d: turn at %v after a turn at %v", day, daytime, lastTime)
				}
				lastTime = daytime
				_ = w.Alive(bee)
			},
		}
	})

	want := []string{"Morning.tick", "Morning.calendarUpdate", "Morning.flowerRefill", "Morning.seasonalEvents",
		"Morning.queenEmergence", "Morning.wakeUp", "Morning.cropRotation", "Morning.colonyUpdate",
		"Morning.broodTotals", "Morning.larvaeNeeds", "Day.activities", "Day.nightRest", "Evening.queensLeavingNest",
		"Evening.larvaeFeeding", "Evening.queenProductionDate", "Evening.development", "Evening.broodMortality",
		"Evening.badgers", "Evening.extinction"}
	if !slices.Equal(parts, want) {
		t.Errorf("parts of day 1:\n%v\nwant\n%v", parts, want)
	}
	if day != days || turns == 0 {
		t.Errorf("%d days, %d turns, want %d days and turns", day, turns, days)
	}
	if plain := run(nil); !reflect.DeepEqual(withHooks, plain) {
		t.Error("the daily output with hooks differs from the one without")
	}
}
