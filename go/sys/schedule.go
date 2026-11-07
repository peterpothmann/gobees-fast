package sys

import (
	"reflect"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Hooks lets a caller stop the run after each part of Morning, Day and Evening and after each bee turn of the
// activity loop: the window of gobees --ui uses it to step through a day. Add it as resource before the systems are
// initialized. Without the resource (the normal run), nothing is called. A hook may read the world but must not
// change it or draw random numbers, so that the run stays the same.
type Hooks struct {
	AfterPart func(system, part string) // after the Update of a part, e.g. ("Morning", "flowerRefill")
	AfterTurn func(bee ecs.Entity)      // after the turn of a bee (activities.act), the bee may have died in it
}

// group runs its parts in order: Initialize of all parts in Initialize, Update of all parts in Update. A part is one
// NetLogo procedure with the methods of an Ark system, but only its group calls them, the app does not.
// All parts draw from the one random source (resource.Rand), so this order fixes the random numbers.
type group struct {
	name  string // the system, for the hooks
	parts []app.System
	names []string // the type names of the parts, for the hooks
	hooks *Hooks   // nil without the resource
}

func (g *group) Initialize(w *ecs.World) {
	g.hooks = ecs.GetResource[Hooks](w)
	g.names = make([]string, len(g.parts))
	for i, p := range g.parts {
		g.names[i] = reflect.TypeOf(p).Elem().Name()
		p.Initialize(w)
	}
}

func (g *group) Update(w *ecs.World) {
	if g.hooks == nil {
		for _, p := range g.parts {
			p.Update(w)
		}
		return
	}
	for i, p := range g.parts {
		p.Update(w)
		g.hooks.AfterPart(g.name, g.names[i])
	}
}

func (g *group) Finalize(w *ecs.World) {}

// Setup builds the world: the landscape, the species, the badgers and the initial queens.
// The parameters and the derived values are resources before (params.Params.Apply).
// The morning at ticks = 0 runs in Morning.Initialize, which follows Setup. The parts work only in Initialize.
//
// NetLogo: Setup, SetupWithoutClearingProc.
type Setup struct{ group }

func (s *Setup) Initialize(w *ecs.World) {
	s.name = "Setup"
	s.parts = []app.System{
		&landscapeSetup{},     // foodsources and their layers
		&nestSitesSetup{},     // nest sites of each species
		&badgersSetup{},       // badgers
		&initialQueensSetup{}, // hibernating queens of the first year
	}
	s.group.Initialize(w)
}

// Update does nothing: the parts of Setup work only in Initialize, their Update is empty. So the hooks report no
// setup parts during the days.
func (s *Setup) Update(w *ecs.World) {}

// Morning advances the day and runs the morning routine. Its Initialize also runs the morning
// at ticks = 0: calendarUpdate, flowerRefill, wakeUp, colonyUpdate and broodTotals call their
// Update at the end of their Initialize. The other parts do nothing at ticks = 0 (see their doc comments).
//
// NetLogo: tick, UpdateMorning_Proc, NeedNectarPollenLarvaeTodayProc.
type Morning struct{ group }

func (s *Morning) Initialize(w *ecs.World) {
	s.name = "Morning"
	s.parts = []app.System{
		&tick{},           // tick
		&calendarUpdate{}, // day of the year, foraging period
		&flowerRefill{},   // foodsources in flower are filled
		&seasonalEvents{}, // end of the season, death of the males, limit of hibernating queens
		&queenEmergence{}, // queens emerge from hibernation and found colonies
		&wakeUp{},         // wake-up times and ActiveBeesSortedList
		&cropRotation{},   // next landscape of the crop rotation on Day 365
		&colonyUpdate{},   // range lists, deaths and phases of the colonies
		&broodTotals{},    // daily brood totals
		&larvaeNeeds{},    // nectar and pollen needs of the larvae
	}
	s.group.Initialize(w)
}

// Day runs the activities of the bees in the order of their personal time, then lets all adult females rest.
//
// NetLogo: the event loop of Go.
type Day struct{ group }

func (s *Day) Initialize(w *ecs.World) {
	s.name = "Day"
	s.parts = []app.System{
		&activities{}, // event loop
		&nightRest{},  // all adult females resting
	}
	s.group.Initialize(w)
}

// Evening runs what happens after the activities: queens leave the nest, larvae are fed, the brood develops and
// dies, the badgers prowl, and the run ends if the population is extinct.
//
// NetLogo: the end of Go.
type Evening struct{ group }

func (s *Evening) Initialize(w *ecs.World) {
	s.name = "Evening"
	s.parts = []app.System{
		&queensLeavingNest{},   // young queens leave the nest and hibernate
		&larvaeFeeding{},       // larvae are fed
		&queenProductionDate{}, // start of the queen production
		&development{},         // all bees age by one day
		&broodMortality{},      // brood that is too old dies
		&badgers{},             // badgers dig up colonies
		&extinction{},          // extinction check and run length
	}
	// The daily output is in the observers (obs.DailyOutput). seasonalEvents counts the three totals it reads
	// itself.
	s.group.Initialize(w)
}
