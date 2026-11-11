//go:build ui || windows || (darwin && cgo)

package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// yearScene shows the life cycle over one year: hibernation, emergence, nest search, colony growth,
// reproduction, season end. Coordinates are normalised: x in [0, 1], y in [0, 1] for the landscape,
// y in [soilTop, 1] for the soil below it.

const (
	yearDaysPerSecond = 7
	soilTop           = 0.84
	seasonEnd         = 305 // schematic, the model uses SeasonStop of the species
	forageRange       = 0.3 // in units of the landscape width
)

const (
	qHibernating = iota
	qSearching
	qFounding // flying to the nest site
	qInNest
	qLeaving // young queen flying to the soil
	qDead
)

type toyPatch struct {
	x, y, r     float64
	habitat     string
	col         color.RGBA
	nest        bool // nest habitat
	start, stop float64
}

type toyQueen struct {
	state    int
	x, y     float64
	fx, fy   float64 // start of the current flight
	tx, ty   float64 // target of the current flight
	t0       float64 // day the flight started
	emerge   float64
	survival float64
	died     float64
	young    bool // emerged this year from a colony, hibernates until next year
}

type toyColony struct {
	x, y       float64
	founded    float64
	workers    float64
	males      float64
	switchDay  float64
	queensLeft int
	dead       bool
	died       float64
	cause      string
}

type yearScene struct {
	seed     uint64
	rng      *rand.Rand // one random sequence per year, so that seeking repeats the year exactly
	year     int
	start    []*toyQueen // queens at the start of the year, for seeking
	day      float64
	lastStep int
	patches  []toyPatch
	queens   []*toyQueen
	colonies []*toyColony
	badgers  []*toyColony // colonies dug up by a badger, for the animation

	pHibernate, pEmerge, pSearch, pFound, pLife, pRepro, pLeave, pBadger, pSeason process
}

func newYearScene(seed uint64) *yearScene {
	s := &yearScene{seed: seed}
	s.pHibernate = process{title: "Hibernation", detail: "Queens rest in the soil over winter.", proc: "activity \"hibernate\""}
	s.pEmerge = process{title: "Emergence", detail: "Each queen emerges on her own day. Heavier queens survive the winter more often.", proc: "EmergenceNewQueensProc"}
	s.pSearch = process{title: "Nest search", detail: "A queen finds a nest site with 20 % per day, otherwise she may die. Sites are chosen by habitat area, not by distance.", proc: "NestSitesSearchingProc"}
	s.pFound = process{title: "Colony foundation", detail: "The queen founds a colony at the nest site.", proc: "CreateColoniesProc"}
	s.pLife = process{title: "Colony life", detail: "Every day: egg laying, brood care, foraging, development. See \"A day\".", proc: "Go"}
	s.pRepro = process{title: "Males and new queens", detail: "Later in the season the queen lays male eggs, and larvae can develop into queens.", proc: "QueenProductionDateProc"}
	s.pLeave = process{title: "Young queens leave", detail: "They leave the nest, mate and hibernate in the soil.", proc: "QueensLeavingNestProc"}
	s.pBadger = process{title: "Badgers", detail: "A badger can dig up a colony nearby.", proc: "BadgersOnTheProwlProc"}
	s.pSeason = process{title: "Season end", detail: "All bees die except the queens in the soil.", proc: "UpdateSeasonalEventsProc"}
	s.reset()
	return s
}

func (s *yearScene) processes() []*process {
	return []*process{&s.pHibernate, &s.pEmerge, &s.pSearch, &s.pFound, &s.pLife, &s.pRepro, &s.pLeave, &s.pBadger, &s.pSeason}
}

func (s *yearScene) reset() {
	s.queens = nil
	s.beginYear(1)
	s.day = 1
	// Colours as on the map: NetLogo colours of the habitat types.
	type def struct {
		x, y, r     float64
		habitat     string
		c           float64
		nest        bool
		start, stop float64
	}
	defs := []def{
		{0.16, 0.22, 0.09, "Woodland", 35, true, 60, 150},
		{0.42, 0.18, 0.12, "Crop", 75, false, 0, 0},
		{0.70, 0.20, 0.08, "Oilseed rape", 95, false, 100, 140},
		{0.88, 0.40, 0.07, "Grassland", 135, true, 120, 270},
		{0.30, 0.50, 0.05, "Hedgerow", 105, true, 60, 300},
		{0.55, 0.46, 0.10, "Grassland", 135, true, 150, 280},
		{0.12, 0.62, 0.06, "Hedgerow", 105, true, 60, 300},
		{0.78, 0.66, 0.09, "Crop", 75, false, 0, 0},
		{0.45, 0.72, 0.04, "Scrub", 125, true, 90, 200},
		{0.62, 0.74, 0.05, "Clover field", 115, false, 160, 260},
	}
	s.patches = s.patches[:0]
	for _, d := range defs {
		s.patches = append(s.patches, toyPatch{d.x, d.y, d.r, d.habitat, netlogoColor(d.c), d.nest, d.start, d.stop})
	}
}

// beginYear keeps the young queens in the soil for the new year and starts its random sequence.
func (s *yearScene) beginYear(year int) {
	s.year = year
	s.start = nil
	for _, q := range s.queens {
		if q.state == qHibernating && q.young {
			c := *q
			s.start = append(s.start, &c)
		}
	}
	s.restart()
}

// restart puts the scene back to the start of the current year.
func (s *yearScene) restart() {
	s.rng = rand.New(rand.NewPCG(s.seed, uint64(s.year)))
	s.queens = nil
	for _, q := range s.start {
		c := *q
		s.queens = append(s.queens, &c)
	}
	s.colonies, s.badgers = nil, nil
	s.lastStep = 0
}

func (s *yearScene) timeRange() (float64, float64) { return 1, daysPerYear + 0.999 }
func (s *yearScene) timeValue() float64            { return s.day }
func (s *yearScene) timeLabel(v float64) string {
	return fmt.Sprintf("day %d · year %d", int(v), s.year)
}
func (s *yearScene) click(x, y int) bool { return false }

// seek jumps to a day of the current year by running the year again from its start.
func (s *yearScene) seek(v float64) {
	s.restart()
	s.day = v
	for s.lastStep < int(v) {
		s.lastStep++
		s.stepDay(s.lastStep)
	}
	s.updateFlights()
}

// prepareHibernation puts a queen into the soil with an emergence day and a winter survival chance.
func (s *yearScene) prepareHibernation(q *toyQueen) {
	q.state = qHibernating
	q.x, q.y = 0.05+0.9*s.rng.Float64(), soilTop+0.04+0.1*s.rng.Float64()
	// Emergence day ~ N(91, 28): EmergingDay_mean and EmergingDay_sd of B. terrestris in the species file.
	q.emerge = min(max(91+28*s.rng.NormFloat64(), 40), 200)
	q.survival = 0.3 + 0.5*s.rng.Float64() // made up, stands for the weight-dependent survival
}

func (s *yearScene) update(dt float64) {
	s.day += dt * yearDaysPerSecond
	for int(s.day) > s.lastStep {
		if s.lastStep == daysPerYear {
			s.day -= daysPerYear
			s.beginYear(s.year + 1)
			continue
		}
		s.lastStep++
		s.stepDay(s.lastStep)
	}
	s.updateFlights()
}

func (s *yearScene) updateFlights() {
	for _, q := range s.queens {
		switch q.state {
		case qSearching, qFounding, qLeaving:
			dur := 1.0
			if q.state == qLeaving {
				dur = 3
			}
			t := smooth((s.dayAbs() - q.t0) / dur)
			q.x, q.y = lerp(q.fx, q.tx, t), lerp(q.fy, q.ty, t)
		}
	}
}

// dayAbs is the day without the wrap at the end of the year, for animations.
func (s *yearScene) dayAbs() float64 { return float64(s.year-1)*daysPerYear + s.day - 1 }

func (s *yearScene) fly(q *toyQueen, now, tx, ty float64) {
	q.fx, q.fy, q.tx, q.ty, q.t0 = q.x, q.y, tx, ty, now
}

func (s *yearScene) stepDay(d int) {
	if d == 1 {
		s.newYear()
	}
	day := float64(d)
	now := float64(s.year-1)*daysPerYear + day - 1 // whole days, so that seeking gives the same result
	if d < 60 || d > 250 {
		for _, q := range s.queens {
			if q.state == qHibernating {
				s.pHibernate.fire()
				break
			}
		}
	}
	for _, q := range s.queens {
		switch q.state {
		case qHibernating:
			if !q.young && day >= q.emerge {
				s.pEmerge.fire()
				if s.rng.Float64() < q.survival {
					q.state = qSearching
					s.fly(q, now, 0.05+0.9*s.rng.Float64(), 0.05+0.7*s.rng.Float64())
				} else {
					q.state, q.died = qDead, now
				}
			}
		case qSearching:
			s.pSearch.fire()
			switch {
			case d >= seasonEnd:
				q.state, q.died = qDead, now
			case s.rng.Float64() < 0.2: // chanceFindNest of B. terrestris [NEST-01]
				p := s.nestSite()
				a, r := 2*math.Pi*s.rng.Float64(), p.r*math.Sqrt(s.rng.Float64())
				q.state = qFounding
				s.fly(q, now, p.x+r*math.Cos(a), p.y+r*math.Sin(a)*s.aspect())
			case s.rng.Float64() < 0.194: // death during the nest search with the default values [NEST-04]
				q.state, q.died = qDead, now
			default:
				s.fly(q, now, min(max(q.x+0.25*s.rng.NormFloat64(), 0.03), 0.97), min(max(q.y+0.2*s.rng.NormFloat64(), 0.03), 0.78))
			}
		case qFounding:
			if now-q.t0 >= 1 {
				q.state = qInNest
				q.x, q.y = q.tx, q.ty
				s.colonies = append(s.colonies, &toyColony{
					x: q.x, y: q.y, founded: day,
					switchDay:  day + 55 + 20*s.rng.Float64(),
					queensLeft: 1 + s.rng.IntN(3),
				})
				s.pFound.fire()
			}
		case qLeaving:
			if now-q.t0 >= 3 {
				q.state = qHibernating
				q.x, q.y = q.tx, q.ty
			}
		}
	}

	for _, c := range s.colonies {
		if c.dead {
			continue
		}
		if d >= seasonEnd {
			c.dead, c.died, c.cause = true, now, "season end"
			s.pSeason.fire()
			s.killQueenAt(c, now)
			continue
		}
		age := day - c.founded
		switch {
		case age < 22: // the queen raises the first brood alone
		case day < c.switchDay:
			c.workers = min(c.workers*1.06+1, 140)
		default:
			c.workers *= 0.985
			c.males += c.workers * 0.04
			s.pRepro.fire()
			if day >= c.switchDay+8 && c.queensLeft > 0 && s.rng.Float64() < 0.25 {
				c.queensLeft--
				q := &toyQueen{state: qLeaving, x: c.x, y: c.y, young: true}
				s.fly(q, now, 0.05+0.9*s.rng.Float64(), soilTop+0.04+0.1*s.rng.Float64())
				s.queens = append(s.queens, q)
				s.pLeave.fire()
			}
		}
		if age >= 0 {
			s.pLife.fire()
		}
		if c.workers > 5 && s.rng.Float64() < 0.002 { // made up
			c.dead, c.died, c.cause = true, now, "badger"
			s.badgers = append(s.badgers, c)
			s.pBadger.fire()
			s.killQueenAt(c, now)
		}
	}
}

func (s *yearScene) killQueenAt(c *toyColony, now float64) {
	for _, q := range s.queens {
		if q.state == qInNest && q.x == c.x && q.y == c.y {
			q.state, q.died = qDead, now
		}
	}
}

// newYear keeps the young queens in the soil for next year.
func (s *yearScene) newYear() {
	var next []*toyQueen
	for _, q := range s.queens {
		if q.state == qHibernating && q.young {
			q.young = false
			s.prepareHibernation(q)
			next = append(next, q)
		}
	}
	s.rng.Shuffle(len(next), func(i, j int) { next[i], next[j] = next[j], next[i] })
	if len(next) > 18 { // keep the picture readable
		next = next[:18]
	}
	for len(next) < 10 { // keep the picture alive
		q := &toyQueen{}
		s.prepareHibernation(q)
		next = append(next, q)
	}
	s.queens = next
}

// nestSite chooses a nest habitat weighted by its area, like NestSiteFoodSourceREP [NSFS].
func (s *yearScene) nestSite() *toyPatch {
	total := 0.0
	for i := range s.patches {
		if s.patches[i].nest {
			total += s.patches[i].r * s.patches[i].r
		}
	}
	p := s.rng.Float64() * total
	for i := range s.patches {
		if !s.patches[i].nest {
			continue
		}
		p -= s.patches[i].r * s.patches[i].r
		if p <= 0 {
			return &s.patches[i]
		}
	}
	return &s.patches[len(s.patches)-1]
}

// aspect converts a radius in width units into height units (set while drawing).
func (s *yearScene) aspect() float64 { return yearAspect }

var yearAspect = 1.5

func (s *yearScene) caption() string {
	d := s.day
	switch {
	case d < 60:
		return "Winter. The queens hibernate in the soil. Their body weight decides how likely they survive."
	case d < 130:
		return "Spring. Each queen emerges on her own day and searches a nest site. She finds one with 20 % per day and can die while searching."
	case d < 200:
		return "Early summer. Each colony starts with the queen alone. Once the first workers emerge, they forage nectar and pollen on flowering patches within reach, and the colony grows."
	case d < seasonEnd:
		return "Late summer. The colonies produce males and new queens. Young queens leave the nest, mate and hibernate in the soil until next spring."
	}
	return "Season end. All bees die except the young queens in the soil. In spring the cycle starts again."
}

func (s *yearScene) draw(screen *ebiten.Image, area image.Rectangle) {
	// Timeline
	tl := image.Rect(area.Min.X+pad, area.Min.Y+10, area.Max.X-pad, area.Min.Y+58)
	s.drawTimeline(screen, tl)

	land := image.Rect(area.Min.X+pad, tl.Max.Y+8, area.Max.X-pad, area.Max.Y-pad)
	W, H := float64(land.Dx()), float64(land.Dy())
	yearAspect = W / H
	pt := func(x, y float64) (float32, float32) {
		return float32(float64(land.Min.X) + x*W), float32(float64(land.Min.Y) + y*H)
	}

	// Soil
	sx, sy := pt(0, soilTop)
	vector.DrawFilledRect(screen, sx, sy, float32(W), float32(H*(1-soilTop)), colSoil, false)
	drawText(screen, "soil", faceSmall, float64(sx)+6, float64(sy)+4, colMuted)

	// Flower patches
	for i := range s.patches {
		p := &s.patches[i]
		x, y := pt(p.x, p.y)
		r := float32(p.r * W)
		if s.day >= p.start && s.day < p.stop {
			vector.DrawFilledCircle(screen, x, y, r, withAlpha(colAccent, 0.55), true)
		}
		vector.StrokeCircle(screen, x, y, r, 2.5, p.col, true)
		label := p.habitat
		if p.nest {
			label += " · nest site"
		}
		w := measure(label) * faceSmall.Size / faceNormal.Size
		drawText(screen, label, faceSmall, float64(x)-w/2, float64(y+r)+3, colMuted)
	}

	// Colonies and their foragers
	for ci, c := range s.colonies {
		x, y := pt(c.x, c.y)
		if c.dead {
			age := s.dayAbs() - c.died
			if age < 8 {
				a := 1 - age/8
				drawNest(screen, float64(x), float64(y), 9, withAlpha(colNest, 0.4*a), withAlpha(colDead, a))
				drawText(screen, c.cause, faceSmall, float64(x)+12, float64(y)-8, withAlpha(colDead, a))
			}
			continue
		}
		s.drawForagers(screen, c, ci, pt, W)
		size := 8 + math.Sqrt(c.workers)*1.2
		drawNest(screen, float64(x), float64(y), size, withAlpha(colNest, 0.9), colText)
		for m := range min(int(c.males/15), 6) {
			a := s.dayAbs()*1.7 + float64(m)
			vector.DrawFilledCircle(screen, x+float32((size+6)*math.Cos(a)), y+float32((size+6)*math.Sin(a)), 2.5, colMale, true)
		}
		if c.workers >= 1 {
			drawText(screen, fmt.Sprintf("%.0f", c.workers), faceSmall, float64(x)+size+3, float64(y)-8, colText)
		}
	}

	// Badgers
	for _, c := range s.badgers {
		age := s.dayAbs() - c.died
		if age > 6 {
			continue
		}
		x, y := pt(c.x, c.y)
		a := float32(1 - age/6)
		vector.DrawFilledCircle(screen, x-14, y+10, 7, withAlpha(color.RGBA{0xdd, 0xdd, 0xdd, 0xff}, float64(a)), true)
		vector.DrawFilledRect(screen, x-18, y+8, 8, 3, withAlpha(color.RGBA{0x22, 0x22, 0x22, 0xff}, float64(a)), false)
	}

	// Queens
	for _, q := range s.queens {
		x, y := pt(q.x, q.y)
		switch q.state {
		case qHibernating:
			vector.DrawFilledCircle(screen, x, y, 4.5, withAlpha(colQueen, 0.75), true)
		case qSearching, qFounding, qLeaving:
			vector.DrawFilledCircle(screen, x, y, 5, colQueen, true)
			vector.StrokeCircle(screen, x, y, 7, 1, withAlpha(colQueen, 0.5), true)
		case qDead:
			age := s.dayAbs() - q.died
			if age < 6 {
				drawCross(screen, float64(x), float64(y), 4, withAlpha(colDead, 1-age/6))
			}
		}
	}

	// Legend
	lx, ly := float64(land.Max.X-150), float64(land.Min.Y+4)
	items := []struct {
		c     color.RGBA
		label string
	}{{colQueen, "queen"}, {colWorker, "worker"}, {colMale, "male"}, {colAccent, "in flower"}}
	for _, it := range items {
		vector.DrawFilledCircle(screen, float32(lx+5), float32(ly+9), 4.5, it.c, true)
		drawText(screen, it.label, faceSmall, lx+16, ly, colText)
		ly += 18
	}
	drawNest(screen, lx+5, ly+9, 6, colNest, colText)
	drawText(screen, "colony", faceSmall, lx+16, ly, colText)
	ly += 18
	drawCross(screen, lx+5, ly+9, 4, colDead)
	drawText(screen, "died", faceSmall, lx+16, ly, colText)
}

// drawForagers draws workers flying between the colony and flowering patches within reach.
func (s *yearScene) drawForagers(screen *ebiten.Image, c *toyColony, ci int, pt func(x, y float64) (float32, float32), W float64) {
	if c.workers < 1 {
		return
	}
	var targets []*toyPatch
	for i := range s.patches {
		p := &s.patches[i]
		dx, dy := (p.x - c.x), (p.y-c.y)/yearAspect
		if s.day >= p.start && s.day < p.stop && math.Hypot(dx, dy) < forageRange {
			targets = append(targets, p)
		}
	}
	if len(targets) == 0 {
		return
	}
	n := min(int(c.workers/10)+1, 10)
	cx, cy := pt(c.x, c.y)
	for i := range n {
		p := targets[(i*7+ci*3)%len(targets)]
		phase := math.Mod(s.dayAbs()*0.35+float64(i)*0.61+float64(ci)*0.23, 1)
		f := phase * 2
		if f > 1 {
			f = 2 - f
		}
		px, py := pt(p.x, p.y)
		x := lerp(float64(cx), float64(px), f)
		y := lerp(float64(cy), float64(py), f) + 6*math.Sin(f*math.Pi)
		vector.DrawFilledCircle(screen, float32(x), float32(y), 2.5, colWorker, true)
	}
}

func (s *yearScene) drawTimeline(screen *ebiten.Image, r image.Rectangle) {
	x0, w := float64(r.Min.X), float64(r.Dx())
	dayX := func(d float64) float32 { return float32(x0 + (d-1)/(daysPerYear-1)*w) }
	y := float32(r.Min.Y + 20)
	vector.StrokeLine(screen, dayX(1), y, dayX(daysPerYear), y, 2, colBorder, false)
	// Seasons of the animation
	bands := []struct {
		a, b  float64
		label string
	}{{1, 60, "hibernation"}, {60, 130, "emergence, nest search"}, {130, 200, "colony growth"}, {200, seasonEnd, "males, new queens"}, {seasonEnd, daysPerYear, "season end"}}
	for i, b := range bands {
		c := withAlpha(colMuted, 0.15+0.1*float64(i%2))
		vector.DrawFilledRect(screen, dayX(b.a), y-6, dayX(b.b)-dayX(b.a), 12, c, false)
		drawText(screen, b.label, faceSmall, float64(dayX(b.a))+4, float64(y)-24, colMuted)
	}
	months := []string{"J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"}
	for i, m := range months {
		d := 1 + float64(i)*daysPerYear/12
		drawText(screen, m, faceSmall, float64(dayX(d))+2, float64(y)+6, colMuted)
	}
	dx := dayX(s.day)
	vector.StrokeLine(screen, dx, y-10, dx, y+10, 3, colText, false)
	drawText(screen, fmt.Sprintf("day %d", int(s.day)), faceSmall, float64(dx)-18, float64(y)+22, colText)
}
