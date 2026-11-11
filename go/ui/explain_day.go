//go:build ui || windows || (darwin && cgo)

package ui

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// dayScene shows one day of one colony in early summer, before males and young queens are raised: the event loop over the bees' personal times, the task choice
// by stimulus and threshold, foraging, egg laying, and the end-of-day processes.
//
// Every day has its own random sequence. Seeking within the day restores the state of the morning and
// runs the day again up to the chosen time, so the slider shows exactly what playing shows.

const (
	dayHoursPerSecond = 0.9
	eveningStepS      = 3.0 // seconds per end-of-day step
	sunrise           = 8.0 // Sunrise_s = 8:00 in the model
	foragingEnd       = 22.0
	nCells            = 30
	eggEvery          = 0.25 // hours between two eggs in the animation
)

// Tasks in the order in which ActivityProc tests them. The last task whose stimulus exceeds the
// bee's threshold wins (finding 11).
var dayTasks = []string{"egg laying", "nursing", "pollen foraging", "nectar foraging"}

type dayBee struct {
	name             string
	queen            bool
	t                float64 // personal time [h]
	task             string
	actStart, actEnd float64
	patch            int
	unloaded         bool
	th               [4]float64
	hx, hy           float64 // resting place in the nest, normalised
	died             bool
	layCells         []int // cells the queen fills today, in this order
	layStart         float64
}

type broodCell struct {
	stage  int     // 0 empty, 1 egg, 2 larva, 3 pupa
	age    int     // days in this stage
	laidAt float64 // egg laid today at this time (for the animation), -1 otherwise
}

type dayPatch struct {
	x, y, r float64
	name    string
	nectar  float64 // 0..1
}

type logEntry struct {
	t    float64
	bee  string
	text string
}

// dayState is everything that changes during a day. It is copied in the morning for seeking.
type dayState struct {
	bees     []*dayBee
	cells    []broodCell
	patches  []dayPatch
	pollen   float64 // 0..1
	energy   float64 // 0..1
	eggsLaid bool
	log      []logEntry
	dayNo    int
	used     float64 // time up to which the colony's consumption is counted
}

func (d *dayState) clone() dayState {
	c := *d
	c.bees = make([]*dayBee, len(d.bees))
	for i, b := range d.bees {
		nb := *b
		nb.layCells = append([]int(nil), b.layCells...)
		c.bees[i] = &nb
	}
	c.cells = append([]broodCell(nil), d.cells...)
	c.patches = append([]dayPatch(nil), d.patches...)
	c.log = append([]logEntry(nil), d.log...)
	return c
}

type beeHit struct {
	name string
	x, y float64
}

type dayScene struct {
	dayState
	seed     uint64
	rng      *rand.Rand
	start    dayState // state before the morning of the current day
	clock    float64  // hours
	evening  int      // -1 during the day, else index of the end-of-day step
	eveTimer float64
	eveNote  string
	last     string // name of the bee that acted last
	follow   string // name of the bee the user follows, "" if none
	hits     []beeHit

	pMorning, pNeed, pLoop, pChoice, pEgg, pNurse, pForage, pEvening, pDevelop, pBadger process
}

func newDayScene(seed uint64) *dayScene {
	s := &dayScene{seed: seed}
	s.pMorning = process{title: "Morning", detail: "Flowers refill. Every bee gets a wake-up time.", proc: "UpdateMorning_Proc"}
	s.pNeed = process{title: "Larval needs", detail: "Nectar and pollen the larvae need today.", proc: "NeedNectarPollenLarvaeTodayProc"}
	s.pLoop = process{title: "Activity loop", detail: "The bee with the earliest personal time acts next.", proc: "NextActiveBeeREP"}
	s.pChoice = process{title: "Task choice", detail: "A task is chosen if the colony's need (stimulus) is above the bee's threshold. Otherwise the bee rests for 30 min.", proc: "ActivityProc"}
	s.pEgg = process{title: "Egg laying", detail: "Once a day the queen lays a batch of eggs, if pollen and energy suffice. Then her day is over.", proc: "EgglayingProc"}
	s.pNurse = process{title: "Brood care", detail: "Workers warm the brood.", proc: "BroodIncubationProc"}
	s.pForage = process{title: "Foraging", detail: "Find a flower patch, collect nectar or pollen, fly home, unload. Only during daylight.", proc: "ForagingProc"}
	s.pEvening = process{title: "Evening", detail: "Young queens leave the nest. The larvae are fed from the stores.", proc: "QueensLeavingNestProc, FeedLarvaeProc"}
	s.pDevelop = process{title: "Development and mortality", detail: "Everybody ages one day. Brood that develops too slowly dies.", proc: "DevelopmentProc, MortalityBroodProc"}
	s.pBadger = process{title: "Badgers", detail: "A badger nearby can dig up the colony.", proc: "BadgersOnTheProwlProc"}
	s.reset()
	return s
}

func (s *dayScene) processes() []*process {
	return []*process{&s.pMorning, &s.pNeed, &s.pLoop, &s.pChoice, &s.pEgg, &s.pNurse, &s.pForage, &s.pEvening, &s.pDevelop, &s.pBadger}
}

func (s *dayScene) reset() {
	s.rng = rand.New(rand.NewPCG(s.seed, 0))
	s.dayState = dayState{dayNo: 1, pollen: 0.45, energy: 0.55}
	s.patches = []dayPatch{
		{0.11, 0.12, 0.055, "Hedgerow", 1},
		{0.10, 0.32, 0.075, "Clover", 1},
		{0.12, 0.54, 0.05, "Bramble", 1},
	}
	s.bees = []*dayBee{{name: "Q", queen: true, th: [4]float64{0.5, 9, 9, 9}}}
	for i := range 7 {
		s.bees = append(s.bees, s.newWorker(fmt.Sprintf("W%d", i+1)))
	}
	s.cells = make([]broodCell, nCells)
	for i := range s.cells {
		s.cells[i].laidAt = -1
		switch {
		case i < 6:
			s.cells[i].stage, s.cells[i].age = 1, s.rng.IntN(3)
		case i < 14:
			s.cells[i].stage, s.cells[i].age = 2, s.rng.IntN(8)
		case i < 19:
			s.cells[i].stage, s.cells[i].age = 3, s.rng.IntN(6)
		}
	}
	s.follow = ""
	s.start = s.clone()
	s.morning()
}

func (s *dayScene) newWorker(name string) *dayBee {
	b := &dayBee{name: name}
	// Every worker has its own thresholds, so the tasks are shared between the workers.
	b.th = [4]float64{9, 0.2 + 0.6*s.rng.Float64(), 0.2 + 0.6*s.rng.Float64(), 0.2 + 0.6*s.rng.Float64()}
	return b
}

// morning: flowers refill, every active bee gets a wake-up time between 0:00 and 0:30
// (GetUpTime_s + random 1800 s in the model).
func (s *dayScene) morning() {
	s.rng = rand.New(rand.NewPCG(s.seed, uint64(s.dayNo)))
	s.clock, s.evening, s.eggsLaid, s.last, s.eveNote, s.used = 0, -1, false, "", "", 0
	s.log = nil
	for i := range s.patches {
		s.patches[i].nectar = 1
	}
	for i, b := range s.bees {
		b.t = s.rng.Float64() * 0.5
		b.task, b.actStart, b.actEnd, b.unloaded, b.layCells = "", 0, 0, true, nil
		b.hx = 0.47 + 0.06*float64(i%4)
		b.hy = 0.52 + 0.06*float64(i/4)
	}
	s.pMorning.fire()
	s.pNeed.fire()
}

func (s *dayScene) timeRange() (float64, float64) { return 0, 24 }
func (s *dayScene) timeValue() float64            { return s.clock }
func (s *dayScene) timeLabel(v float64) string {
	if s.evening >= 0 {
		return fmt.Sprintf("day %d · evening", s.dayNo)
	}
	return fmt.Sprintf("day %d · %s", s.dayNo, hhmm(v))
}

// seek jumps to a time of the current day: back to the morning, then run the day up to v.
func (s *dayScene) seek(v float64) {
	s.dayState = s.start.clone()
	s.morning()
	s.advance(v)
}

func (s *dayScene) click(x, y int) bool {
	for _, h := range s.hits {
		if math.Hypot(float64(x)-h.x, float64(y)-h.y) < 11 {
			if s.follow == h.name {
				s.follow = ""
			} else {
				s.follow = h.name
			}
			return true
		}
	}
	s.follow = ""
	return false
}

func hhmm(t float64) string {
	h := int(t)
	m := int((t - float64(h)) * 60)
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (s *dayScene) count(stage int) int {
	n := 0
	for _, c := range s.cells {
		if c.stage == stage {
			n++
		}
	}
	return n
}

// stimuli are made up, but behave like in the model: brood raises the need for brood care, low stores
// raise the need for foraging, and foraging needs daylight.
func (s *dayScene) stimuli(t float64) [4]float64 {
	var st [4]float64
	if !s.eggsLaid && s.count(0) >= 4 {
		st[0] = 1
	}
	brood := float64(s.count(1) + s.count(2) + s.count(3))
	st[1] = 0.15 + 0.7*brood/nCells
	if t >= sunrise && t < foragingEnd {
		st[2] = 0.2 + 0.8*(1-s.pollen)*float64(s.count(2))/8
		st[3] = 0.2 + 0.8*(1-s.energy)
	}
	for i := range st {
		st[i] = min(st[i], 1)
	}
	return st
}

func (s *dayScene) next() *dayBee {
	var best *dayBee
	for _, b := range s.bees {
		if b.died {
			continue
		}
		if best == nil || b.t < best.t {
			best = b
		}
	}
	return best
}

func (s *dayScene) bee(name string) *dayBee {
	for _, b := range s.bees {
		if b.name == name && !b.died {
			return b
		}
	}
	return nil
}

func (s *dayScene) update(dt float64) {
	if s.evening >= 0 {
		s.eveTimer += dt
		if s.eveTimer >= eveningStepS {
			s.eveTimer = 0
			s.evening++
			s.runEveningStep()
		}
		return
	}
	// Slow while the queen lays her eggs, so it can be seen. Faster in the night, when the bees only rest.
	rate := dayHoursPerSecond
	switch {
	case s.queenLaying():
		rate = 0.25
		s.pEgg.fire()
	case s.clock < sunrise-0.5 && s.clock > 1:
		rate = 2
	}
	s.advance(s.clock + dt*rate)
	if s.clock >= 24 {
		if b := s.next(); b == nil || b.t >= 24 { // CallItaDay_s: even the earliest bee is beyond the end of the day
			s.evening, s.eveTimer = 0, 0
			s.runEveningStep()
		}
	}
}

// advance runs the day up to the given time. The bees act in the order of their personal times.
func (s *dayScene) advance(target float64) {
	target = min(target, 24)
	for {
		b := s.next()
		if b == nil || b.t >= 24 || b.t > target {
			break
		}
		s.consumeUntil(b.t)
		s.unloadUntil(b.t)
		s.act(b)
	}
	s.clock = max(s.clock, target)
	s.unloadUntil(s.clock)
}

// consumeUntil lets the colony use energy and pollen up to time t (made up rates). It is called only at
// the times when a bee acts, so that seeking and playing give the same result.
func (s *dayScene) consumeUntil(t float64) {
	dt := t - s.used
	s.used = t
	s.energy = max(s.energy-0.03*dt, 0)
	s.pollen = max(s.pollen-0.025*dt, 0)
}

// unloadUntil lets foragers that are back by time t unload into the stores.
func (s *dayScene) unloadUntil(t float64) {
	for _, b := range s.bees {
		if !b.unloaded && b.actEnd <= t {
			b.unloaded = true
			if b.task == "pollen foraging" {
				s.pollen = min(s.pollen+0.035, 1)
			} else {
				s.energy = min(s.energy+0.03, 1)
			}
		}
	}
}

// act lets one bee perform one task. Its personal time moves on by the duration of the task.
func (s *dayScene) act(b *dayBee) {
	s.pLoop.fire()
	s.pChoice.fire()
	st := s.stimuli(b.t)
	task := "resting"
	for i, name := range dayTasks {
		if st[i] > b.th[i] {
			task = name
		}
	}
	b.task, b.actStart, b.unloaded = task, b.t, true
	dur := 0.5
	switch task {
	case "egg laying":
		dur = 24 // the queen is done for the day
		b.layCells, b.layStart = nil, b.t
		for i := range s.cells {
			if s.cells[i].stage == 0 && len(b.layCells) < 6 && s.pollen > 0.05 && s.energy > 0.05 {
				s.cells[i] = broodCell{stage: 1, laidAt: b.t + eggEvery*float64(len(b.layCells)+1)}
				b.layCells = append(b.layCells, i)
			}
		}
		s.pollen = max(s.pollen-0.05, 0)
		s.energy = max(s.energy-0.03, 0)
		s.eggsLaid = true
		s.pEgg.fire()
	case "nursing":
		s.pNurse.fire()
	case "pollen foraging", "nectar foraging":
		dur = 0.5 + s.rng.Float64()
		b.patch = s.rng.IntN(len(s.patches))
		b.unloaded = false
		p := &s.patches[b.patch]
		p.nectar = max(p.nectar-0.08, 0)
		s.pForage.fire()
	}
	b.actEnd = b.t + dur
	b.t += dur
	s.last = b.name
	text := task
	if task == "egg laying" {
		text = fmt.Sprintf("egg laying (%d eggs)", len(b.layCells))
	}
	s.log = append(s.log, logEntry{b.actStart, b.name, text})
}

func (s *dayScene) runEveningStep() {
	switch s.evening {
	case 0:
		for _, b := range s.bees {
			b.task, b.unloaded = "resting", true
		}
		for i := range s.cells {
			s.cells[i].laidAt = -1
		}
		s.eveNote = "End of the day: all bees rest."
		s.pEvening.fire()
	case 1:
		s.eveNote = "Young queens would leave the nest now to mate and hibernate. This colony has none yet."
		s.pEvening.fire()
	case 2:
		larvae := float64(s.count(2))
		s.pollen = max(s.pollen-0.025*larvae, 0)
		s.energy = max(s.energy-0.02*larvae, 0)
		s.eveNote = fmt.Sprintf("The %d larvae are fed from the stores: pollen and energy go down.", int(larvae))
		s.pEvening.fire()
	case 3:
		emerged, died := 0, 0
		for i := range s.cells {
			c := &s.cells[i]
			if c.stage == 0 {
				continue
			}
			c.age++
			switch {
			case c.stage == 1 && c.age >= 3:
				c.stage, c.age = 2, 0
			case c.stage == 2 && c.age >= 8:
				c.stage, c.age = 3, 0
			case c.stage == 3 && c.age >= 6:
				*c = broodCell{laidAt: -1}
				emerged++
			}
		}
		for range emerged {
			if s.aliveWorkers() < 11 {
				s.bees = append(s.bees, s.newWorker(fmt.Sprintf("W%d", len(s.bees))))
			}
		}
		for _, b := range s.bees {
			if !b.queen && !b.died && s.rng.Float64() < 0.06 {
				b.died = true
				died++
			}
		}
		s.eveNote = fmt.Sprintf("Everybody ages one day: eggs hatch, larvae pupate, %d new workers emerge, %d old workers die.", emerged, died)
		s.pDevelop.fire()
	case 4:
		died := 0
		if s.pollen < 0.15 {
			for i := range s.cells {
				if s.cells[i].stage == 2 && died < 2 {
					s.cells[i] = broodCell{laidAt: -1}
					died++
				}
			}
		}
		if died > 0 {
			s.eveNote = fmt.Sprintf("Pollen is short: %d larvae did not grow enough and die.", died)
		} else {
			s.eveNote = "All brood developed in time, none dies today."
		}
		s.pDevelop.fire()
	case 5:
		s.eveNote = "No badger came near the colony today."
		s.pBadger.fire()
	default:
		s.dayNo++
		var alive []*dayBee
		for _, b := range s.bees {
			if !b.died {
				alive = append(alive, b)
			}
		}
		s.bees = alive
		if s.bee(s.follow) == nil {
			s.follow = ""
		}
		s.start = s.clone()
		s.morning()
	}
}

// queenLaying reports whether the queen is laying eggs at the current time (animation).
func (s *dayScene) queenLaying() bool {
	q := s.bee("Q")
	return q != nil && s.evening < 0 && len(q.layCells) > 0 &&
		s.clock < q.layStart+eggEvery*float64(len(q.layCells)+1)
}

func (s *dayScene) aliveWorkers() int {
	n := 0
	for _, b := range s.bees {
		if !b.queen && !b.died {
			n++
		}
	}
	return n
}

func (s *dayScene) caption() string {
	if s.evening >= 0 && s.eveNote != "" {
		return s.eveNote
	}
	if s.queenLaying() {
		return "The queen wakes up first and lays a batch of eggs, one per empty cell. Egg laying costs pollen and energy. Afterwards her personal time jumps by 24 h: her day is over."
	}
	if s.clock < 0.6 {
		return "Morning: the flowers are refilled, every bee gets a wake-up time between 0:00 and 0:30. The bees are sorted by these times."
	}
	if s.clock < sunrise {
		return "The bee with the earliest personal time acts next. Before sunrise there is no foraging, so most bees rest for 30 min and come back later."
	}
	return "Each bee compares the colony's needs (stimuli) with its own thresholds and does one task. Its personal time then moves on by the duration of the task. Click a bee to follow it."
}

// layout of the nest, shared by drawing and the queen's flight to the cells.
type nestLayout struct {
	x0, y0, x1, y1, cr float64
}

func (s *dayScene) nest(pt func(x, y float64) (float64, float64)) nestLayout {
	x0, y0 := pt(0.30, 0.04)
	x1, y1 := pt(0.70, 0.66)
	return nestLayout{x0, y0, x1, y1, math.Min((x1-x0-40)/10/2, 13)}
}

func (n nestLayout) cell(i int) (float64, float64) {
	const cols = 10
	cx := n.x0 + 28 + float64(i%cols)*n.cr*2.05
	cy := n.y0 + 36 + float64(i/cols)*n.cr*1.85
	if i/cols%2 == 1 {
		cx += n.cr
	}
	return cx, cy
}

func (s *dayScene) draw(screen *ebiten.Image, area image.Rectangle) {
	inner := image.Rect(area.Min.X+pad, area.Min.Y+pad, area.Max.X-pad, area.Max.Y-pad)
	W, H := float64(inner.Dx()), float64(inner.Dy())
	pt := func(x, y float64) (float64, float64) {
		return float64(inner.Min.X) + x*W, float64(inner.Min.Y) + y*H
	}
	s.hits = s.hits[:0]

	// Flower patches with their nectar level.
	for _, p := range s.patches {
		x, y := pt(p.x, p.y)
		r := p.r * W
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(r), withAlpha(colAccent, 0.15+0.6*p.nectar), true)
		vector.StrokeCircle(screen, float32(x), float32(y), float32(r), 2, colAccent, true)
		drawText(screen, p.name, faceSmall, x-r, y+r+2, colMuted)
	}

	// Nest with brood cells and stores.
	n := s.nest(pt)
	vector.DrawFilledRect(screen, float32(n.x0), float32(n.y0), float32(n.x1-n.x0), float32(n.y1-n.y0), withAlpha(colNest, 0.12), false)
	vector.StrokeRect(screen, float32(n.x0), float32(n.y0), float32(n.x1-n.x0), float32(n.y1-n.y0), 1.5, colNest, false)
	drawText(screen, "nest", faceSmall, n.x0+6, n.y0+4, colNest)
	for i, c := range s.cells {
		cx, cy := n.cell(i)
		drawCell(screen, cx, cy, n.cr, withAlpha(colNest, 0.25), withAlpha(colNest, 0.7))
		switch c.stage {
		case 1:
			if c.laidAt < 0 || s.clock >= c.laidAt {
				vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(n.cr*0.25), colText, true)
			}
		case 2:
			vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(n.cr*(0.35+0.05*float64(c.age))), withAlpha(colText, 0.8), true)
		case 3:
			drawCell(screen, cx, cy, n.cr*0.8, withAlpha(colNest, 0.95), colNest)
		}
	}
	ly := n.y0 + 36 + 3*n.cr*1.85 + 4
	legend := []struct {
		stage int
		label string
	}{{1, "egg"}, {2, "larva"}, {3, "pupa"}}
	lx := n.x0 + 16
	for _, l := range legend {
		drawCell(screen, lx, ly+8, 7, withAlpha(colNest, 0.25), withAlpha(colNest, 0.7))
		switch l.stage {
		case 1:
			vector.DrawFilledCircle(screen, float32(lx), float32(ly+8), 2, colText, true)
		case 2:
			vector.DrawFilledCircle(screen, float32(lx), float32(ly+8), 4, colText, true)
		case 3:
			drawCell(screen, lx, ly+8, 5.5, colNest, colNest)
		}
		drawText(screen, fmt.Sprintf("%s (%d)", l.label, s.count(l.stage)), faceSmall, lx+11, ly, colMuted)
		lx += 90
	}
	// Stores
	sx := n.x0 + 16
	sy := ly + 30
	bw := (n.x1 - n.x0 - 48) / 2
	for i, st := range []struct {
		label string
		v     float64
	}{{"energy store (nectar)", s.energy}, {"pollen store", s.pollen}} {
		x := sx + float64(i)*(bw+16)
		drawText(screen, st.label, faceSmall, x, sy, colMuted)
		vector.DrawFilledRect(screen, float32(x), float32(sy+20), float32(bw), 10, colButton, false)
		vector.DrawFilledRect(screen, float32(x), float32(sy+20), float32(bw*st.v), 10, colAccent, false)
	}

	// Bees
	for _, b := range s.bees {
		if b.died {
			continue
		}
		x, y := s.beePos(b, pt, n)
		c := colWorker
		r := float32(6)
		if b.queen {
			c, r = colQueen, 8
		}
		vector.DrawFilledCircle(screen, float32(x), float32(y), r, c, true)
		if b.name == s.follow {
			vector.StrokeCircle(screen, float32(x), float32(y), r+5, 3, colAccent, true)
		} else if b.name == s.last && s.evening < 0 {
			vector.StrokeCircle(screen, float32(x), float32(y), r+4, 1.5, colText, true)
		}
		drawText(screen, b.name, faceSmall, x+9, y-8, colText)
		s.hits = append(s.hits, beeHit{b.name, x, y})
	}

	s.drawNeeds(screen, pt)
	s.drawQueue(screen, pt)
}

// beePos is where a bee is at the current time: at a flower patch, at the brood, laying eggs, or resting.
func (s *dayScene) beePos(b *dayBee, pt func(x, y float64) (float64, float64), n nestLayout) (float64, float64) {
	hx, hy := pt(b.hx, b.hy)
	if s.evening >= 0 {
		return hx, hy
	}
	if b.queen && len(b.layCells) > 0 {
		// The queen flies from cell to cell while she lays, then goes back to her resting place.
		k := int((s.clock - b.layStart) / eggEvery)
		if k >= 0 && k < len(b.layCells) {
			f := smooth((s.clock-b.layStart)/eggEvery - float64(k))
			x0, y0 := hx, hy
			if k > 0 {
				x0, y0 = n.cell(b.layCells[k-1])
				y0 -= n.cr
			}
			x1, y1 := n.cell(b.layCells[k])
			y1 -= n.cr
			return lerp(x0, x1, f), lerp(y0, y1, f)
		}
		if k == len(b.layCells) {
			x0, y0 := n.cell(b.layCells[k-1])
			f := smooth((s.clock-b.layStart)/eggEvery - float64(k))
			return lerp(x0, hx, f), lerp(y0-n.cr, hy, f)
		}
		return hx, hy
	}
	switch b.task {
	case "pollen foraging", "nectar foraging":
		if s.clock >= b.actStart && s.clock < b.actEnd {
			f := (s.clock - b.actStart) / (b.actEnd - b.actStart) * 2
			if f > 1 {
				f = 2 - f
			}
			p := s.patches[b.patch]
			px, py := pt(p.x, p.y)
			return lerp(hx, px, smooth(f)), lerp(hy, py, smooth(f))
		}
	case "nursing":
		if s.clock >= b.actStart && s.clock < b.actEnd {
			return pt(b.hx-0.1, 0.13+(b.hy-0.52)) // above the brood cells, spread like the resting places
		}
	}
	return hx, hy
}

// drawNeeds shows the colony's needs now, the thresholds and the actions of the followed bee.
func (s *dayScene) drawNeeds(screen *ebiten.Image, pt func(x, y float64) (float64, float64)) {
	x, y := pt(0.75, 0.04)
	x1, _ := pt(1, 0)
	w := x1 - x
	f := s.bee(s.follow)
	drawText(screen, "Colony needs now", faceBold, x, y, colText)
	y += 26
	st := s.stimuli(s.clock)
	for i, name := range dayTasks {
		active := f != nil && f.task == name && s.clock < f.actEnd && s.evening < 0
		c := colMuted
		if active {
			c = colText
		}
		drawText(screen, name, faceSmall, x, y, c)
		y += 18
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 9, colButton, false)
		bc := withAlpha(colAccent, 0.5)
		if active {
			bc = colAccent
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w*st[i]), 9, bc, false)
		if f != nil && f.th[i] <= 1 {
			tx := float32(x + w*f.th[i])
			vector.StrokeLine(screen, tx, float32(y-3), tx, float32(y+12), 2, colText, false)
		}
		y += 18
	}
	if f == nil {
		y += drawText(screen, wrap("Click a bee to follow it and see its thresholds.", faceSmall, w), faceSmall, x, y, colMuted) + 12
		drawText(screen, "Last actions", faceBold, x, y, colText)
	} else {
		drawText(screen, fmt.Sprintf("bar: need · line: threshold of %s", f.name), faceSmall, x, y, colMuted)
		y += 26
		drawText(screen, "Following "+f.name, faceBold, x, y, colAccent)
		y += 22
		var now string
		switch {
		case s.evening >= 0:
			now = "resting (end of the day)"
		case f.queen && f.task == "egg laying":
			now = "laid her eggs, done for today (personal time +24 h)"
		case f.task == "" || f.t <= s.clock:
			now = fmt.Sprintf("waiting, acts at %s", hhmm(f.t))
		case s.clock < f.actEnd && f.task != "resting":
			now = fmt.Sprintf("%s until %s", f.task, hhmm(f.actEnd))
		default:
			now = fmt.Sprintf("resting, acts again at %s", hhmm(f.t))
		}
		y += drawText(screen, wrap(now, faceSmall, w), faceSmall, x, y, colText) + 8
		drawText(screen, "Actions of "+f.name, faceBold, x, y, colText)
	}
	y += 22
	shown := 0
	for i := len(s.log) - 1; i >= 0 && shown < 7; i-- {
		e := s.log[i]
		if f != nil && e.bee != f.name {
			continue
		}
		label := e.text
		if f == nil {
			label = e.bee + "  " + e.text
		}
		drawText(screen, fmt.Sprintf("%s  %s", hhmm(e.t), label), faceSmall, x, y, colMuted)
		y += 17
		shown++
	}
}

// drawQueue draws the time ruler with the personal time of every bee. The bee furthest left acts next.
func (s *dayScene) drawQueue(screen *ebiten.Image, pt func(x, y float64) (float64, float64)) {
	x0, y := pt(0.02, 0.84)
	x1, _ := pt(0.98, 0)
	hx := func(h float64) float32 { return float32(x0 + h/24*(x1-x0)) }
	drawText(screen, fmt.Sprintf("Day %d in early summer: only workers are raised, males and young queens come later", s.dayNo), faceBold, x0, y-92, colText)
	drawText(screen, "Personal time of each bee. The bee furthest left acts next.", faceSmall, x0, y-70, colMuted)
	vector.DrawFilledRect(screen, hx(sunrise), float32(y-8), hx(foragingEnd)-hx(sunrise), 16, withAlpha(colAccent, 0.18), false)
	drawText(screen, "foraging possible", faceSmall, float64(hx(sunrise))+4, y-28, colAccent)
	vector.StrokeLine(screen, hx(0), float32(y), hx(24), float32(y), 2, colBorder, false)
	for h := 0; h <= 24; h += 3 {
		vector.StrokeLine(screen, hx(float64(h)), float32(y-4), hx(float64(h)), float32(y+4), 1, colMuted, false)
		drawText(screen, fmt.Sprintf("%d:00", h), faceSmall, float64(hx(float64(h)))-14, y+10, colMuted)
	}
	// Bees on the ruler, stacked when close together.
	used := map[int]int{}
	for _, b := range s.bees {
		if b.died {
			continue
		}
		t := math.Min(b.t, 24)
		slot := int(t * 4)
		k := used[slot]
		used[slot]++
		bx := hx(t)
		by := float32(y) - 14 - float32(k)*11
		c := colWorker
		if b.queen {
			c = colQueen
		}
		vector.DrawFilledCircle(screen, bx, by, 4.5, c, true)
		if b.name == s.follow {
			vector.StrokeCircle(screen, bx, by, 8, 2.5, colAccent, true)
		}
		s.hits = append(s.hits, beeHit{b.name, float64(bx), float64(by)})
	}
	cx := hx(math.Min(s.clock, 24))
	vector.StrokeLine(screen, cx, float32(y-40), cx, float32(y+8), 2, colText, false)
	label := "now " + hhmm(s.clock)
	lx := float64(cx) + 4
	if s.clock > 20 { // keep the label inside the view at the end of the day
		lx = float64(cx) - 4 - measure(label)*faceSmall.Size/faceNormal.Size
	}
	drawText(screen, label, faceSmall, lx, y-44, colText)
}
