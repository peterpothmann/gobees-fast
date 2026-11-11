//go:build ui || windows || (darwin && cgo)

package ui

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"maps"
	"math"
	"slices"
	"strings"

	"gobees/day"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawPanel(screen *ebiten.Image) {
	r := g.layout.panelRect
	fillRect(screen, r, colPanel)
	strokeRect(screen, r, colBorder)

	x := float64(r.Min.X + pad)
	w := float64(r.Dx() - 2*pad)
	y := float64(r.Min.Y + pad)

	drawText(screen, "gobees-fast", faceTitle, x, y, colText)
	tw, _ := measureFace("gobees-fast", faceTitle)
	// the run file, up to the explanation button (drawExplainButton)
	room := w - tw - 12 - (measure("How the model works") + 24) - 8
	drawText(screen, truncate(g.cfg.Title, room), faceSmall, x+tw+12, y+8, colMuted)
	titleY := y
	y += 36

	// Status line, for the day on screen
	t := g.shown().Tick
	state := "running"
	switch {
	case g.runErr != nil:
		state = "error"
	case g.viewing:
		state = fmt.Sprintf("back in time, latest tick %d", g.snap.Tick)
	case g.finished:
		state = "finished"
	case g.paused:
		state = "paused"
	}
	c := colText
	if g.viewing {
		c = colAccent
	}
	drawText(screen, fmt.Sprintf("ticks %d   ·   day %d   ·   year %d", t, tickDay(t), tickYear(t)), faceBold, x, y, c)
	drawTextRight(screen, fmt.Sprintf("%s · %g d/s", state, speeds[g.speed]), faceNormal, x+w, y, colMuted)
	y += 28

	y = g.drawButtons(screen, x, y, w)
	g.drawExplainButton(screen, x+w, titleY)
	if g.runErr != nil {
		y += drawText(screen, wrap(g.runErr.Error(), faceSmall, w), faceSmall, x, y, colError) + 8
	}
	y = g.drawTabs(screen, x, y, w) + 12

	help := "Space play/pause · → to the end of the day, ← one day back (hold: more) · T next bee turn · P next part · ↑↓ speed · click: keep patch or follow colony · right-click: test nest · double right-click or Esc: remove · closing the window ends the run at the end of the day"
	help = wrap(help, faceSmall, w)
	lines := float64(strings.Count(help, "\n") + 1)
	helpY := float64(r.Max.Y-pad) - lines*faceSmall.Size*1.35
	drawText(screen, help, faceSmall, x, helpY, colMuted)

	if g.tab == 1 {
		g.drawDebug(screen, x, y, w, helpY-10)
		return
	}
	if g.selColony != nil {
		y = g.drawColonySeries(screen, x, y, w, g.selColony)
	} else {
		y = g.drawSeries(screen, x, y, w)
	}
	y += 14
	info := screen.SubImage(image.Rect(r.Min.X, int(y), r.Max.X, int(helpY-6))).(*ebiten.Image) // stop above the help
	switch {
	case g.hoverColony != nil:
		g.drawColonyInfo(info, x, y, w, g.hoverColony)
	case g.selColony != nil:
		g.drawColonyInfo(info, x, y, w, g.selectedView())
	default:
		g.drawPatchInfo(info, x, y, w)
	}
}

// drawTabs draws the tabs "Model output" and "Debug".
func (g *Game) drawTabs(screen *ebiten.Image, x, y, w float64) float64 {
	labels := []string{"Model output", "Debug"}
	tw := w / float64(len(labels))
	mx, my := ebiten.CursorPosition()
	for i, l := range labels {
		rect := image.Rect(int(x+float64(i)*tw), int(y), int(x+float64(i+1)*tw), int(y+28))
		b := button{rect: rect, label: l, action: func() { g.setTab(i) }}
		g.buttons = append(g.buttons, b)
		c := colMuted
		if g.tab == i {
			c = colText
			vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Max.Y-3), float32(rect.Dx()), 3, colAccent, false)
		} else if b.contains(mx, my) {
			c = colText
		}
		lw := measure(l)
		drawText(screen, l, faceNormal, float64(rect.Min.X)+(tw-lw)/2, y+4, c)
	}
	vector.StrokeLine(screen, float32(x), float32(y+28), float32(x+w), float32(y+28), 1, colBorder, false)
	return y + 28
}

// addButton draws a button with a label in the small font and registers it for clicks.
func (g *Game) addButton(dst *ebiten.Image, rect image.Rectangle, label string, action func()) {
	b := button{rect: rect, label: label, action: action}
	g.buttons = append(g.buttons, b)
	c := colButton
	mx, my := ebiten.CursorPosition()
	if b.contains(mx, my) {
		c = colButtonHi
	}
	fillRect(dst, rect, c)
	tw, th := measureFace(label, faceSmall)
	drawText(dst, label, faceSmall, float64(rect.Min.X)+(float64(rect.Dx())-tw)/2,
		float64(rect.Min.Y)+(float64(rect.Dy())-th)/2-1, colText)
}

func (g *Game) drawButtons(screen *ebiten.Image, x, y, w float64) float64 {
	play := "Play"
	if !g.paused {
		play = "Pause"
	}
	defs := []struct {
		label  string
		action func()
	}{
		{play, g.togglePause},
		{"Step", g.stepOnce},
		{"Slower", g.slower},
		{"Faster", g.faster},
	}
	gap := 8.0
	bw := (w - gap*float64(len(defs)-1)) / float64(len(defs))
	g.buttons = g.buttons[:0]
	mx, my := ebiten.CursorPosition()
	for i, d := range defs {
		bx := x + float64(i)*(bw+gap)
		rect := image.Rect(int(bx), int(y), int(bx+bw), int(y+30))
		b := button{rect: rect, label: d.label, action: d.action}
		g.buttons = append(g.buttons, b)
		c := colButton
		if b.contains(mx, my) {
			c = colButtonHi
		}
		vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), c, false)
		tw := measure(d.label)
		drawText(screen, d.label, faceNormal, bx+(bw-tw)/2, y+6, colText)
	}
	return y + 30 + 8
}

// drawExplainButton draws the button that opens the explanation window, right-aligned at x.
func (g *Game) drawExplainButton(screen *ebiten.Image, x, y float64) {
	label := "How the model works"
	bw := measure(label) + 24
	rect := image.Rect(int(x-bw), int(y), int(x), int(y+28))
	b := button{rect: rect, label: label, action: g.openExplain}
	g.buttons = append(g.buttons, b)
	c := colButton
	mx, my := ebiten.CursorPosition()
	if b.contains(mx, my) {
		c = colButtonHi
	}
	fillRect(screen, rect, c)
	tc := colText
	if g.explainOpen.Load() {
		tc = colMuted
	}
	drawText(screen, label, faceNormal, x-bw+12, y+5, tc)
}

func measure(s string) float64 {
	w, _ := textMeasure(s)
	return w
}

// line is one series of a plot.
type line struct {
	name   string
	values []float64
	format string // for the value today, e.g. "%.2f l"
	col    color.RGBA
	start  int64 // tick of values[0]
}

// plotEvent is a vertical line in a plot, e.g. the foundation of a colony.
type plotEvent struct {
	t     int64
	label string
	col   color.RGBA // zero: grey
}

// drawSeries draws the daily model output since the start of the run, counted at the end of each day.
func (g *Game) drawSeries(screen *ebiten.Image, x, y, w float64) float64 {
	h := &g.hist
	drawText(screen, "Model output per day", faceBold, x, y, colText)
	y += 22
	legend := "Vertical lines each year: E queens emerge (mean), S season ends (SeasonStop)"
	if len(g.seasons) > 1 {
		legend += ". Coloured lines differ between species and have the colour of the species"
	}
	y += drawText(screen, wrap(legend, faceSmall, w), faceSmall, x, y, colMuted) + 8
	plots := []struct {
		lines    []line
		ownScale bool
	}{
		{[]line{
			{"nectar left in the landscape", h.Nectar, "%.2f l", colAccent, 0},
			{"pollen left in the landscape", h.Pollen, "%.2f kg", colSecond, 0},
		}, true},
		{[]line{
			{"queens hibernating", h.Hibernating, "%.0f", colMuted, 0},
			{"queens searching a nest", h.Searching, "%.0f", colBee, 0},
			{"cuckoo bee queens searching", h.CuckooSearching, "%.0f", colCuckoo, 0},
			{"colonies", h.Colonies, "%.0f", colNest, 0},
		}, false},
		{[]line{
			{"eggs", h.Eggs, "%.0f", colBrood, 0},
			{"larvae", h.Larvae, "%.0f", colLarva, 0},
			{"pupae", h.Pupae, "%.0f", colPupa, 0},
		}, false},
		{[]line{
			{"workers", h.Workers, "%.0f", colAccent, 0},
			{"males", h.Males, "%.0f", colSecond, 0},
			{"queens in colonies", h.Queens, "%.0f", colBee, 0},
			{"cuckoo bees in colonies", h.Cuckoos, "%.0f", colCuckoo, 0},
		}, false},
		{[]line{
			{"energy stores", h.Energy, "%.0f kJ", colAccent, 0},
			{"pollen stores", h.PollenStore, "%.1f g", colSecond, 0},
			{"foraging trips", h.Trips, "%.0f", colLarva, 0},
		}, true},
	}
	cuckoos := slices.Max(h.CuckooSearching) > 0 || slices.Max(h.Cuckoos) > 0
	xr := [2]float64{0, max(float64(len(h.Nectar)-1), daysPerYear)}
	events := g.seasonEvents(xr)
	for i, p := range plots {
		lines := p.lines
		if !cuckoos { // only runs with cuckoo bees show their lines
			lines = slices.DeleteFunc(slices.Clone(lines), func(l line) bool { return l.col == colCuckoo })
		}
		ev := events
		if i > 0 {
			ev = unlabelled(events)
		}
		y = g.drawPlot(screen, x, y, w, lines, p.ownScale, xr, ev) + 12
	}
	return y
}

// drawColonySeries draws the daily output of the selected colony, with the events of its life.
func (g *Game) drawColonySeries(screen *ebiten.Image, x, y, w float64, h *colonyHistory) float64 {
	c := &h.Last
	drawText(screen, fmt.Sprintf("Colony %d · %s", c.Who, prettyName(c.Species)), faceBold, x, y, colText)
	drawTextRight(screen, "Esc or click the map: all colonies", faceSmall, x+w, y+2, colMuted)
	y += 22

	events := []plotEvent{{t: h.Start, label: "F"}}
	col := c.Colony
	for _, e := range []struct {
		t     day.Date
		label string
	}{{col.EusocialPhaseDate, "W"}, {col.QueenProductionDate, "Q"}, {col.SwitchPointDate, "M"}, {col.CompetitionPointDate, "C"}} {
		if e.t.IsSet() {
			events = append(events, plotEvent{t: int64(e.t.Day()), label: e.label})
		}
	}
	if h.CuckooSince != 0 {
		events = append(events, plotEvent{t: h.CuckooSince, label: "K"})
	}
	if h.Died != 0 {
		events = append(events, plotEvent{t: h.Died, label: "†"})
	}
	legend := "F founded · W first workers · Q eggs for young queens · M male eggs · C competition point"
	if h.CuckooSince != 0 {
		legend += " · K cuckoo bee"
	}
	if h.Died != 0 {
		legend += " · † died"
	}
	y += drawText(screen, wrap(legend, faceSmall, w), faceSmall, x, y, colMuted) + 8

	end := int64(g.ticks.Ticks)
	if h.Died != 0 {
		end = h.Died
	}
	xr := [2]float64{float64(h.Start - 1), float64(max(end, h.Start+180))}
	s := h.Start
	bees := []line{{"bees, brood included", h.Size, "%.0f", colNest, s}}
	if h.CuckooSince != 0 {
		bees = append(bees, line{"cuckoo bees", h.Cuckoos, "%.0f", colCuckoo, s})
	}
	plots := []struct {
		lines    []line
		ownScale bool
	}{
		{bees, false},
		{[]line{
			{"eggs", h.Eggs, "%.0f", colBrood, s},
			{"larvae", h.Larvae, "%.0f", colLarva, s},
			{"pupae", h.Pupae, "%.0f", colPupa, s},
		}, false},
		{[]line{
			{"workers", h.Workers, "%.0f", colAccent, s},
			{"males", h.Males, "%.0f", colSecond, s},
			{"queens", h.Queens, "%.0f", colBee, s},
		}, false},
		{[]line{
			{"energy store", h.Energy, "%.1f kJ", colAccent, s},
			{"it aims at", h.IdealEnergy, "%.1f kJ", colMuted, s},
		}, false},
		{[]line{
			{"pollen store", h.Pollen, "%.2f g", colSecond, s},
			{"it aims at", h.IdealPollen, "%.2f g", colMuted, s},
		}, false},
	}
	for i, p := range plots {
		ev := events
		if i > 0 {
			ev = unlabelled(events)
		}
		y = g.drawPlot(screen, x, y, w, p.lines, p.ownScale, xr, ev) + 12
	}
	return y
}

// unlabelled returns the events without their labels: only the first plot names them.
func unlabelled(events []plotEvent) []plotEvent {
	out := make([]plotEvent, len(events))
	for i, e := range events {
		out[i] = plotEvent{t: e.t, col: e.col}
	}
	return out
}

// seasonEvents returns for every year in the range xr (ticks) and every species the day its queens emerge on
// average and the day its season ends.
func (g *Game) seasonEvents(xr [2]float64) []plotEvent {
	out := []plotEvent{}
	for year := 0.0; year*daysPerYear <= xr[1]; year++ {
		out = append(out, g.speciesEvents(year*daysPerYear, "E", func(s season) float64 { return math.Round(s.Emerge) })...)
		out = append(out, g.speciesEvents(year*daysPerYear, "S", func(s season) float64 { return s.Stop })...)
	}
	return out
}

// speciesEvents returns one line per species at start + day(species), in the colour of the species. If all
// species have the same day, it returns one grey line.
func (g *Game) speciesEvents(start float64, label string, day func(season) float64) []plotEvent {
	if len(g.seasons) == 0 {
		return nil
	}
	same := true
	for _, s := range g.seasons {
		same = same && day(s) == day(g.seasons[0])
	}
	if same {
		return []plotEvent{{t: int64(start + day(g.seasons[0])), label: label}}
	}
	out := []plotEvent{}
	for _, s := range g.seasons {
		col := g.colonyColour(s.Name)
		if s.Parasitic {
			col = colCuckoo
		}
		out = append(out, plotEvent{t: int64(start + day(s)), label: label, col: col})
	}
	return out
}

// drawPlot draws one plot with the legend on the left. With ownScale, every line is scaled to its own maximum
// (different units). xr is the range of the x axis in ticks. Events are drawn as vertical lines.
func (g *Game) drawPlot(screen *ebiten.Image, x, y, w float64, lines []line, ownScale bool, xr [2]float64, events []plotEvent) float64 {
	plotH := max(54.0, 18*float64(len(lines)))
	labelW := 210.0
	px, pw := x+labelW, w-labelW
	tx := func(t float64) float32 { return float32(px + (t-xr[0])/(xr[1]-xr[0])*pw) }
	stride := max(1, int((xr[1]-xr[0])/pw)) // at most about one segment per pixel

	maxAll := 0.0
	for _, l := range lines {
		for _, v := range l.values {
			maxAll = max(maxAll, v)
		}
	}
	top, bottom := y, y+plotH
	vector.StrokeLine(screen, float32(px), float32(bottom), float32(px+pw), float32(bottom), 1, colBorder, false)
	// Year boundaries: dashed lines below the data, only where a new year begins inside the plot.
	for yr := math.Floor(xr[0]/daysPerYear+1) * daysPerYear; yr < xr[1]; yr += daysPerYear {
		for dy := 0.0; dy < plotH; dy += 6 {
			yx := pixelCentre(tx(yr))
			vector.StrokeLine(screen, yx, float32(top+dy), yx, float32(top+min(dy+3, plotH)), 1, colBorder, false)
		}
	}
	labelEnd := math.Inf(-1) // right end of the last label: labels closer than that are left out
	for _, e := range slices.SortedStableFunc(slices.Values(events), func(a, b plotEvent) int { return cmp.Compare(a.t, b.t) }) {
		if float64(e.t) < xr[0] || float64(e.t) > xr[1] {
			continue
		}
		ex := pixelCentre(tx(float64(e.t)))
		lc, tc := withAlpha(colText, 0.35), colMuted
		if e.col.A != 0 {
			lc, tc = withAlpha(e.col, 0.6), e.col
		}
		vector.StrokeLine(screen, ex, float32(top), ex, float32(bottom), 1, lc, false)
		if e.label != "" && float64(ex) > labelEnd {
			drawText(screen, e.label, faceSmall, float64(ex)+2, top-2, tc)
			lw, _ := measureFace(e.label, faceSmall)
			labelEnd = float64(ex) + lw + 4
		}
	}
	for i, l := range lines {
		m := maxAll
		if ownScale {
			m = 0
			for _, v := range l.values {
				m = max(m, v)
			}
		}
		if m == 0 {
			m = 1
		}
		ty := func(v float64) float32 { return float32(bottom - v/m*plotH) }
		t0 := float64(l.start)
		prev := 0
		for t := stride; t < len(l.values)+stride-1; t += stride {
			t = min(t, len(l.values)-1)
			vector.StrokeLine(screen, tx(t0+float64(prev)), ty(l.values[prev]), tx(t0+float64(t)), ty(l.values[t]), 1.5, l.col, true)
			prev = t
		}
		// Legend with today's value.
		ly := top + float64(i)*18
		vector.DrawFilledRect(screen, float32(x), float32(ly+6), 10, 3, l.col, false)
		// The value of the shown day: the latest, or the viewed one while going back.
		text := l.name + "  –"
		if i := len(l.values) - 1; i >= 0 {
			if g.viewing {
				i = int(g.viewTick - l.start)
			}
			if i >= 0 && i < len(l.values) {
				text = fmt.Sprintf("%s  "+l.format, l.name, l.values[i])
			}
		}
		drawText(screen, text, faceSmall, x+16, ly, colText)
	}
	if g.viewing && float64(g.viewTick) >= xr[0] && float64(g.viewTick) <= xr[1] {
		vx := pixelCentre(tx(float64(g.viewTick)))
		vector.StrokeLine(screen, vx, float32(top), vx, float32(bottom), 1, colAccent, false)
	}
	if !ownScale {
		label := fmt.Sprintf("max %.0f", maxAll)
		if maxAll < 10 {
			label = fmt.Sprintf("max %.2g", maxAll)
		}
		lw, lh := text.Measure(label, faceSmall, 0)
		vector.DrawFilledRect(screen, float32(px+pw-lw-4), float32(top-4), float32(lw+4), float32(lh+2), colPanel, false) // keep the label readable over the lines
		drawTextRight(screen, label, faceSmall, px+pw, top-4, colMuted)
	}
	return bottom + 4
}

// pixelCentre moves x to the centre of its pixel. A vertical 1-pixel line there covers exactly one pixel column;
// between two columns it would be drawn 1 or 2 pixels wide, depending on the position.
func pixelCentre(x float32) float32 {
	return float32(math.Floor(float64(x))) + 0.5
}

// tickDay is NetLogo's Day of a tick: 1 to 365, 0 before the first day.
func tickDay(t int64) int64 {
	if t <= 0 {
		return 0
	}
	return (t-1)%daysPerYear + 1
}

// tickYear is the year of a tick, 0 before the first day.
func tickYear(t int64) int64 {
	if t <= 0 {
		return 0
	}
	return (t-1)/daysPerYear + 1
}

// selectedView is the selected colony on the shown day, or on its last day if it did not live then.
func (g *Game) selectedView() *colonyView {
	for i, c := range g.shown().Colonies {
		if c.Who == g.selColony.Last.Who {
			return &g.shown().Colonies[i]
		}
	}
	return &g.selColony.Last
}

// tickDate formats a tick as day of the year, with the year from year 2 on.
func tickDate(t float64) string {
	day := math.Mod(t-1, daysPerYear) + 1
	year := math.Floor((t-1)/daysPerYear) + 1
	if year <= 1 {
		return fmt.Sprintf("day %.0f", day)
	}
	return fmt.Sprintf("day %.0f of year %.0f", day, year)
}

// drawColonyInfo shows a colony: the one under the mouse, or the selected one.
func (g *Game) drawColonyInfo(screen *ebiten.Image, x, y, w float64, c *colonyView) float64 {
	h := g.colHist[c.Who]
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+w), float32(y), 1, colBorder, false)
	y += 10
	drawNest(screen, x+7, y+10, 7, g.colonyColour(c.Species), colText)
	drawText(screen, fmt.Sprintf("Colony %d · %s · %.0f bees", c.Who, prettyName(c.Species), c.Size()), faceBold, x+22, y, colText)
	y += 24

	// Bees and stores as a small table. The plots of the selected colony show them already.
	n := c.Counts
	st := c.Stores
	rows := [][2]string{
		{"eggs · larvae · pupae", fmt.Sprintf("%.0f · %.0f · %.0f", n.Eggs, n.Larvae, n.Pupae)},
		{"workers · males · queens", fmt.Sprintf("%.0f · %.0f · %.0f", n.Workers, n.Males, n.Queens)},
		{"cuckoo bees", fmt.Sprintf("%.0f", n.Cuckoos)},
		{"energy store", fmt.Sprintf("%.1f kJ (it aims at %.1f kJ)", st.Energy, st.IdealEnergy)},
		{"pollen store", fmt.Sprintf("%.2f g (it aims at %.2f g)", st.Pollen, st.IdealPollen)},
	}
	if g.isSelected(c.Who) {
		rows = nil
	}
	for _, r := range rows {
		drawText(screen, r[0], faceSmall, x, y, colMuted)
		drawText(screen, r[1], faceSmall, x+170, y, colText)
		y += 18
	}
	y += 6

	// The colony's life so far, in the order of the events.
	type event struct {
		t    float64
		text string
	}
	col := c.Colony
	events := []event{{c.Founded, fmt.Sprintf("Founded on %s by a queen that found a nest site here.", tickDate(c.Founded))}}
	for _, e := range []struct {
		t    day.Date
		text string
	}{
		{col.EusocialPhaseDate, "First workers on %s."},
		{col.QueenProductionDate, "Raises young queens from eggs laid on %s."},
		{col.SwitchPointDate, "The queen lays male eggs from %s on."},
		{col.CompetitionPointDate, "Workers compete with the queen from %s on: new eggs no longer develop, " +
			"and the colony ends when its brood is gone."},
	} {
		if e.t.IsSet() {
			t := float64(e.t.Day())
			events = append(events, event{t, fmt.Sprintf(e.text, tickDate(t))})
		}
	}
	slices.SortStableFunc(events, func(a, b event) int { return cmp.Compare(a.t, b.t) })
	life := []string{}
	for _, e := range events {
		life = append(life, e.text)
	}
	if h != nil && h.CuckooSince != 0 && h.CuckooSince <= g.shown().Tick {
		life = append(life, fmt.Sprintf("A cuckoo bee queen entered the nest by %s. She lays her own eggs here, "+
			"and the colony raises her brood.", tickDate(float64(h.CuckooSince))))
	}
	// The model keeps no flag for a lost queen (NetLogo: queenright?), the UI tells only that no adult queen is in
	// the colony. Young queens count as well, so the colony can have lost its founding queen before.
	if c.Counts.Queens == 0 {
		life = append(life, "No adult queen is in the colony.")
	}
	if h != nil && h.Died != 0 && h.Died <= g.shown().Tick {
		life = append(life, fmt.Sprintf("The colony died on %s. The numbers are those of its last day.", tickDate(float64(h.Died))))
	}
	text := strings.Join(life, " ")
	y += drawText(screen, wrap(text, faceSmall, w), faceSmall, x, y, colText) + 8

	text = fmt.Sprintf("Within its foraging range of %.0f m: %d flower patches in flower today (white rings). "+
		"Its bees know %d of them for nectar (yellow lines) and %d for pollen (blue lines). "+
		"The more bees know a patch, the stronger the line. The five strongest lines show their number of bees.",
		g.land.MaxForagingRange, len(c.Flowers), len(c.KnownNectar), len(c.KnownPollen))
	return y + drawText(screen, wrap(text, faceSmall, w), faceSmall, x, y, colMuted)
}

// drawPatchInfo shows the selected patch, or the one under the mouse.
func (g *Game) drawPatchInfo(screen *ebiten.Image, x, y, w float64) float64 {
	p := g.hover
	if p == nil {
		p = g.selected
	}
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+w), float32(y), 1, colBorder, false)
	y += 10
	if p == nil {
		note := "Each circle is one flower patch, each small cell a colony. Point at them for details."
		if g.nest == nil {
			note += " Right-click the map to place a test nest."
		} else {
			n := 0
			for i := range g.land.Patches {
				if g.land.reachFrom(&g.land.Patches[i], g.nest[0], g.nest[1]).InRange {
					n++
				}
			}
			note += fmt.Sprintf("\n\nTest nest: %d of %d patches are within reach (centre at most %.0f m away). "+
				"The stronger a line, the more likely a searching bee finds the patch.",
				n, len(g.land.Patches), g.land.MaxForagingRange)
		}
		return y + drawText(screen, wrap(note, faceSmall, w), faceSmall, x, y, colMuted)
	}

	vector.StrokeCircle(screen, float32(x+7), float32(y+10), 7, 2.5, p.Colour, true)
	drawText(screen, fmt.Sprintf("%s · %.2f ha", prettyName(p.PatchType), p.Area/10000), faceBold, x+22, y, colText)
	y += 22
	info := fmt.Sprintf("%d plant species, %d in flower today", len(p.Layers), p.layersInFlower(g.shown()))
	y += drawText(screen, info, faceSmall, x, y, colMuted) + 4
	y = g.drawPatchFood(screen, x, y, w, p)
	if g.nest != nil {
		re := g.land.reachFrom(p, g.nest[0], g.nest[1])
		// Two distances: the centre decides whether the patch is within reach,
		// the edge decides how likely a searching bee finds it.
		centre := fmt.Sprintf("Centre: %.0f m from the test nest, ", re.Centre)
		if re.InRange {
			centre += fmt.Sprintf("within reach (max %.0f m).", g.land.MaxForagingRange)
		} else {
			centre += fmt.Sprintf("out of reach (max %.0f m).", g.land.MaxForagingRange)
		}
		y += drawText(screen, wrap(centre, faceSmall, w), faceSmall, x, y, colText) + 2
		if re.InRange {
			edge := fmt.Sprintf("Edge: %.0f m. A search finds the patch with %.0f %%.", re.Edge, 100*re.FindProb)
			if re.Edge == 0 {
				edge = "Edge: the test nest lies inside the patch. Every search finds it."
			}
			y += drawText(screen, wrap(edge, faceSmall, w), faceSmall, x, y, colText) + 4
		}
	}
	y += 4

	// Layers with their flowering period and what is left today, those in flower first.
	day := float64(tickDay(g.shown().Tick))
	maxRows := int((float64(screen.Bounds().Max.Y) - y - 16 - 18) / 16) // room for the heading and "… n more"
	rows := 0
	nameW, leftW := 150.0, 120.0
	bx, bw := x+nameW, w-nameW-leftW
	drawTextRight(screen, "left today", faceSmall, x+w, y-2, colMuted)
	y += 16
	dayX := func(d float64) float32 { return float32(bx + (d-1)/(daysPerYear-1)*bw) }
	for pass := 0; pass < 2; pass++ {
		for i := range p.Layers {
			ly := &p.Layers[i]
			if g.shown().inFlower(ly.Index) != (pass == 0) {
				continue
			}
			if rows >= maxRows {
				break
			}
			c := colMuted
			if pass == 0 {
				c = colAccent
			}
			name := prettyName(ly.FlowerSpecies)
			if name == "" {
				name = "no flower species"
			}
			drawText(screen, name, faceSmall, x, y, c)
			cy := float32(y + 9)
			vector.StrokeLine(screen, dayX(1), cy, dayX(daysPerYear), cy, 1, colBorder, false)
			a, b := math.Max(ly.StartDay, 1), math.Min(ly.StopDay-1, daysPerYear)
			if b >= a {
				vector.StrokeLine(screen, dayX(a), cy, dayX(b), cy, 4, withAlpha(c, 0.8), false)
			}
			if pass == 0 {
				l := g.shown().Layers[ly.Index]
				drawTextRight(screen, fmt.Sprintf("%.1f ml · %.2f g", l.Nectar/1000, l.Pollen), faceSmall, x+w, y, colText)
			}
			y += 16
			rows++
		}
	}
	if day >= 1 {
		vector.StrokeLine(screen, dayX(day), float32(y-float64(rows)*16), dayX(day), float32(y), 1, colText, false)
	}
	if rows < len(p.Layers) {
		y += drawText(screen, fmt.Sprintf("… %d more", len(p.Layers)-rows), faceSmall, x, y, colMuted)
	}
	return y
}

// drawPatchFood shows what the flower patch offers today and which colonies know it. The model counts no visits
// per patch, so the UI shows none.
func (g *Game) drawPatchFood(screen *ebiten.Image, x, y, w float64, p *flowerPatch) float64 {
	var nectar, nectarMax, pollen, pollenMax float64
	for _, ly := range p.Layers {
		l := g.shown().Layers[ly.Index]
		nectar, pollen = nectar+float64(l.Nectar), pollen+float64(l.Pollen)
		if g.shown().inFlower(ly.Index) {
			nectarMax, pollenMax = nectarMax+ly.NectarMax, pollenMax+ly.PollenMax
		}
	}
	lines := []string{}
	if nectarMax+pollenMax > 0 {
		lines = append(lines, fmt.Sprintf("Left today: nectar %.1f / %.1f ml · pollen %.2f / %.2f g (refilled each morning)",
			nectar/1000, nectarMax/1000, pollen, pollenMax))
	}

	// Colonies whose bees know the patch, by species.
	colonies, beesN, beesP := 0, 0.0, 0.0
	bySpecies := map[string]int{}
	for i := range g.shown().Colonies {
		c := &g.shown().Colonies[i]
		n, q := c.KnownNectar[p.Entity], c.KnownPollen[p.Entity]
		if n+q == 0 {
			continue
		}
		colonies++
		beesN, beesP = beesN+n, beesP+q
		bySpecies[prettyName(c.Species)]++
	}
	if colonies == 0 {
		lines = append(lines, "No colony's bees know this patch")
	} else {
		known := fmt.Sprintf("Known to %d colonies: %.0f bees for nectar (yellow) · %.0f for pollen (blue)",
			colonies, beesN, beesP)
		if len(bySpecies) > 1 {
			names := slices.Sorted(maps.Keys(bySpecies))
			parts := []string{}
			for _, n := range names {
				parts = append(parts, fmt.Sprintf("%s %d", n, bySpecies[n]))
			}
			known += " · " + strings.Join(parts, ", ")
		}
		lines = append(lines, known)
	}
	for _, l := range lines {
		y += drawText(screen, wrap(l, faceSmall, w), faceSmall, x, y, colText) + 2
	}
	return y + 4
}
