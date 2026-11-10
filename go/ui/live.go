//go:build ui || windows || (darwin && cgo)

package ui

import (
	"image/color"
	"slices"
	"strings"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/obs"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// modelState gives the UI read access to the parts of the model state it shows. It is created after setup.
type modelState struct {
	species  []params.Species // params.Input.Species, by the index in comp.Bee.Species and comp.Colony.Species
	counters *obs.Counters    // foraging trips of the day
	bees     *ecs.Filter2[comp.Bee, comp.Forager]
	colonies *ecs.Filter5[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyCounts, comp.ColonyRange]
	patches  *ecs.Filter1[comp.FlowerPatch]
	badgers  *ecs.Filter2[comp.Position, comp.Badger]

	// known holds the number and the day of foundation of every colony the UI has seen, by its entity. The model
	// keeps neither (no who numbers, no foundation day), so the UI numbers the colonies in the order it first sees
	// them and takes that day as the day of foundation. The UI reads every day, so it sees a colony on the day the
	// colony is founded.
	known   map[ecs.Entity]colonyID
	nextWho int64
}

// colonyID is what the UI remembers of a colony from the day it first saw it.
type colonyID struct {
	Who     int64
	Founded float64
}

func newModel(w *ecs.World, counters *obs.Counters) *modelState {
	return &modelState{
		species:  ecs.GetResource[params.Input](w).Species,
		counters: counters,
		bees:     ecs.NewFilter2[comp.Bee, comp.Forager](w),
		colonies: ecs.NewFilter5[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyCounts, comp.ColonyRange](w),
		badgers:  ecs.NewFilter2[comp.Position, comp.Badger](w),
		patches:  ecs.NewFilter1[comp.FlowerPatch](w),
		known:    map[ecs.Entity]colonyID{},
		nextWho:  1,
	}
}

// inFlower reports whether a foodsource layer offers nectar or pollen at the end of the day of the snapshot, as
// the model sees it (the colony update: pollen > 0 or nectar > 0).
func (s *snapshot) inFlower(layer int) bool {
	if layer >= len(s.Layers) {
		return false
	}
	l := s.Layers[layer]
	return l.Nectar > 0 || l.Pollen > 0
}

// beeCounts are the numbers of bees at the end of the day, counted by the UI over all bee agents.
// Cuckoo bees (bees of a parasitic species, params.Species.Parasitic) in a colony count only as Cuckoos, so the
// other numbers are the colony's own bees.
type beeCounts struct {
	Hibernating, Searching float64 // adult queens: hibernating, or emerged and without colony (no cuckoo bees)
	CuckooSearching        float64 // adult cuckoo bee queens, emerged and without colony
	Cuckoos                float64 // cuckoo bees in colonies, brood included
	Eggs, Larvae, Pupae    float64 // brood, all castes
	Workers, Males, Queens float64 // other adults (queens: in a colony)
}

// colonyView is what the UI shows of a colony.
type colonyView struct {
	Who     int64 // number the UI gives the colony (colonyID), not a NetLogo who number
	X, Y    float64
	Founded float64 // tick of the foundation (colonyID)
	Species string
	Colony  comp.Colony
	Stores  comp.ColonyStores
	Counts  beeCounts
	InRange int                 // flower patches within the foraging range
	Flowers map[ecs.Entity]bool // masterpatches with layers in flower within the foraging range
	// Known holds the masterpatches the colony's bees know for nectar and pollen (comp.Forager.KnownNectarPatches,
	// KnownPollenPatches), with the number of bees that know them.
	KnownNectar, KnownPollen map[ecs.Entity]float64
	CuckooQueen              bool // an adult cuckoo bee queen is in the nest
}

// Size is the number of bees in the colony, brood and cuckoo bees included.
func (c *colonyView) Size() float64 {
	return c.Counts.size()
}

func (n *beeCounts) size() float64 {
	return n.Eggs + n.Larvae + n.Pupae + n.Workers + n.Males + n.Queens + n.Cuckoos
}

// snapshot is the model state the UI shows at the end of a day. The UI keeps the snapshots of the last days to
// show them again (← key).
type snapshot struct {
	Tick         int64
	Colonies     []colonyView
	Bees         beeCounts          // all bees, also without colony
	Badgers      [][2]float64       // NetLogo coordinates
	Energy       float64            // energy stores of all colonies [kJ]
	Pollen       float64            // pollen stores of all colonies [g]
	NectarLeft   float64            // nectar left in all foodsources at the end of the day [µl]
	PollenLeft   float64            // pollen left in all foodsources at the end of the day [g]
	ForagerTrips float64            // foraging trips of the day (obs.Counters)
	index        map[ecs.Entity]int // colony entity -> index in Colonies
	Layers       []layerDay         // by flowerLayer.Index
}

// layerDay is the state of a foodsource layer at the end of the day. The flower refill fills nectar and pollen
// up to their maximum every morning, foraging takes from them during the day. float32 keeps the snapshots of many
// days small.
type layerDay struct {
	Nectar float32 // left at the end of the day [µl]
	Pollen float32 // left at the end of the day [g]
}

// read collects the snapshot of the day. It runs once per simulated day.
func (m *modelState) read(tick int64, land *landscape) snapshot {
	s := snapshot{Tick: tick, index: map[ecs.Entity]int{}, Layers: make([]layerDay, len(land.layers)),
		ForagerTrips: m.counters.TotalForagingTripsToday}
	pq := m.patches.Query()
	for pq.Next() {
		fp := pq.Get()
		for j := range fp.Layers {
			st := &fp.Layers[j]
			s.NectarLeft += st.Nectar
			s.PollenLeft += st.Pollen
			if i, ok := land.layers[comp.Source{Patch: pq.Entity(), Layer: j}]; ok {
				s.Layers[i] = layerDay{Nectar: float32(st.Nectar), Pollen: float32(st.Pollen)}
			}
		}
	}
	q := m.colonies.Query()
	for q.Next() {
		pos, c, st, _, rng := q.Get()
		id, ok := m.known[q.Entity()]
		if !ok {
			id = colonyID{Who: m.nextWho, Founded: float64(tick)}
			m.known[q.Entity()] = id
			m.nextWho++
		}
		v := colonyView{
			Who: id.Who, X: pos.X, Y: pos.Y, Founded: id.Founded, Species: m.species[c.Species].Name,
			Colony: *c, Stores: *st, InRange: len(rng.Patches), Flowers: map[ecs.Entity]bool{},
			KnownNectar: map[ecs.Entity]float64{}, KnownPollen: map[ecs.Entity]float64{},
		}
		for _, e := range rng.NectarPatches {
			v.Flowers[e] = true
		}
		for _, e := range rng.PollenPatches {
			v.Flowers[e] = true
		}
		s.index[q.Entity()] = len(s.Colonies)
		s.Colonies = append(s.Colonies, v)
		s.Energy += st.Energy
		s.Pollen += st.Pollen
	}

	bq := m.bees.Query()
	for bq.Next() {
		b, fo := bq.Get()
		parasitic := m.species[b.Species].Parasitic
		s.Bees.add(b, parasitic)
		if i, ok := s.index[b.Colony]; ok && !b.Colony.IsZero() {
			c := &s.Colonies[i]
			c.Counts.add(b, parasitic)
			if parasitic && b.Stage == stage.Adult && b.Caste == caste.Queen {
				c.CuckooQueen = true
			}
			for _, e := range fo.KnownNectarPatches {
				c.KnownNectar[e] += float64(b.Number)
			}
			for _, e := range fo.KnownPollenPatches {
				c.KnownPollen[e] += float64(b.Number)
			}
		}
	}

	bd := m.badgers.Query()
	for bd.Next() {
		pos, _ := bd.Get()
		s.Badgers = append(s.Badgers, [2]float64{pos.X, pos.Y})
	}
	return s
}

// add counts the bees of agent b. parasitic tells whether b belongs to a parasitic species.
func (n *beeCounts) add(b *comp.Bee, parasitic bool) {
	num := float64(b.Number)
	if parasitic && b.Activity != activity.Hibernate {
		if b.Colony.IsZero() {
			n.CuckooSearching += num // only adult queens live outside colonies
		} else {
			n.Cuckoos += num
		}
		return
	}
	switch b.Stage {
	case stage.Egg:
		n.Eggs += num
	case stage.Larva:
		n.Larvae += num
	case stage.Pupa:
		n.Pupae += num
	case stage.Adult:
		switch {
		case b.Caste == caste.Queen && b.Activity == activity.Hibernate:
			n.Hibernating += num
		case b.Caste == caste.Queen && b.Colony.IsZero():
			n.Searching += num
		case b.Caste == caste.Queen:
			n.Queens += num
		case b.Caste == caste.Worker:
			n.Workers += num
		case b.Caste == caste.Male:
			n.Males += num
		}
	}
}

// history holds one value per day for the plots, index = tick.
type history struct {
	Nectar, Pollen                   []float64 // left in the landscape at the end of the day [l], [kg]
	Hibernating, Searching, Colonies []float64
	Eggs, Larvae, Pupae              []float64
	Workers, Males, Queens           []float64
	Energy, PollenStore, Trips       []float64
	CuckooSearching, Cuckoos         []float64
}

func (h *history) record(s *snapshot) {
	b := s.Bees
	h.Nectar = append(h.Nectar, s.NectarLeft/1e6)
	h.Pollen = append(h.Pollen, s.PollenLeft/1000)
	h.Hibernating = append(h.Hibernating, b.Hibernating)
	h.Searching = append(h.Searching, b.Searching)
	h.Colonies = append(h.Colonies, float64(len(s.Colonies)))
	h.Eggs = append(h.Eggs, b.Eggs)
	h.Larvae = append(h.Larvae, b.Larvae)
	h.Pupae = append(h.Pupae, b.Pupae)
	h.Workers = append(h.Workers, b.Workers)
	h.Males = append(h.Males, b.Males)
	h.Queens = append(h.Queens, b.Queens)
	h.Energy = append(h.Energy, s.Energy)
	h.PollenStore = append(h.PollenStore, s.Pollen)
	h.Trips = append(h.Trips, s.ForagerTrips)
	h.CuckooSearching = append(h.CuckooSearching, b.CuckooSearching)
	h.Cuckoos = append(h.Cuckoos, b.Cuckoos)
}

// colonyHistory holds one value per day for one colony, from its first day on (index 0 = tick Start).
// It stays after the colony died.
type colonyHistory struct {
	Start       int64      // tick of the first value
	Died        int64      // tick of the first day without the colony, 0 while it lives
	CuckooSince int64      // tick of the first day with an adult cuckoo bee queen in the nest, 0 if none
	Last        colonyView // the colony on its last day

	Size, Cuckoos, Eggs, Larvae, Pupae, Workers, Males, Queens []float64
	Energy, IdealEnergy, Pollen, IdealPollen                   []float64
}

// recordColonies adds the day to the history of every colony and marks the colonies that died.
func recordColonies(hs map[int64]*colonyHistory, s *snapshot, tick int64) {
	alive := map[int64]bool{}
	for i := range s.Colonies {
		c := &s.Colonies[i]
		alive[c.Who] = true
		h := hs[c.Who]
		if h == nil {
			h = &colonyHistory{Start: tick}
			hs[c.Who] = h
		}
		if c.CuckooQueen && h.CuckooSince == 0 {
			h.CuckooSince = tick
		}
		h.Last = *c
		n := c.Counts
		h.Size = append(h.Size, n.size())
		h.Cuckoos = append(h.Cuckoos, n.Cuckoos)
		h.Eggs = append(h.Eggs, n.Eggs)
		h.Larvae = append(h.Larvae, n.Larvae)
		h.Pupae = append(h.Pupae, n.Pupae)
		h.Workers = append(h.Workers, n.Workers)
		h.Males = append(h.Males, n.Males)
		h.Queens = append(h.Queens, n.Queens)
		h.Energy = append(h.Energy, c.Stores.Energy)
		h.IdealEnergy = append(h.IdealEnergy, c.Stores.IdealEnergy)
		h.Pollen = append(h.Pollen, c.Stores.Pollen)
		h.IdealPollen = append(h.IdealPollen, c.Stores.IdealPollen)
	}
	for who, h := range hs {
		if h.Died == 0 && !alive[who] {
			h.Died = tick
		}
	}
}

// season is the yearly phenology of a bee species: when its hibernated queens emerge on average, and the day of
// the year from which its colonies die (seasonStop).
type season struct {
	Name         string
	Parasitic    bool    // a cuckoo bumblebee species (params.Species.Parasitic)
	Emerge, Stop float64 // day of the year
}

func speciesSeasons(species []params.Species) []season {
	out := []season{}
	for _, s := range species {
		out = append(out, season{s.Name, s.Parasitic, s.EmergenceDayMean, float64(s.SeasonStop)})
	}
	slices.SortFunc(out, func(a, b season) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// speciesColours gives every bee species that founds colonies (not parasitic) its own colour, in the order of the
// names. With one species, its colonies keep the nest colour.
func speciesColours(species []params.Species) map[string]color.RGBA {
	names := []string{}
	for _, s := range species {
		if !s.Parasitic {
			names = append(names, s.Name)
		}
	}
	slices.Sort(names)
	palette := []color.RGBA{colNest, {0x7f, 0xc9, 0x7f, 0xff}, {0x9e, 0xb8, 0xf0, 0xff}, {0xf0, 0x80, 0x60, 0xff},
		{0xe0, 0xe0, 0x70, 0xff}, {0x60, 0xd0, 0xd0, 0xff}, {0xd8, 0xa8, 0xd8, 0xff}}
	out := map[string]color.RGBA{}
	for i, n := range names {
		out[n] = palette[i%len(palette)]
	}
	return out
}
