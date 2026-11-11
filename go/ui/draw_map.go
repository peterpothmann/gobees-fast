//go:build ui || windows || (darwin && cgo)

package ui

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mlange-42/ark/ecs"
)

// minRadius keeps tiny patches visible and clickable [screen pixels].
const minRadius = 3

func (g *Game) drawMap(full *ebiten.Image) {
	r := g.layout.mapRect
	screen := full.SubImage(r).(*ebiten.Image)              // clip ranges and circles to the map
	fillRect(screen, r, color.RGBA{0x8c, 0x8c, 0x8c, 0xff}) // NetLogo Backgroundcolor 5 (grey)
	strokeRect(screen, r, colBorder)

	if g.nest != nil {
		x, y := g.layout.toScreen(g.nest[0], g.nest[1])
		g.drawRange(screen, x, y)
	}
	hc := g.focusColony()
	if hc != nil {
		x, y := g.layout.toScreen(hc.X, hc.Y)
		g.drawRange(screen, x, y)
	}

	for i := range g.land.Patches {
		p := &g.land.Patches[i]
		x, y := g.layout.toScreen(p.X, p.Y)
		rad := float32(max(p.Radius*g.layout.scale, minRadius))
		alpha := 1.0
		if g.nest != nil && !g.land.reachFrom(p, g.nest[0], g.nest[1]).InRange {
			alpha = 0.25 // out of reach of the test nest
		}
		if hc != nil && !hc.Flowers[p.Entity] {
			alpha = 0.25 // not in flower within the range of the colony under the mouse or selected
		}
		// The habitat, dimmed like a map image behind the patches, so the flowers stand out.
		vector.DrawFilledCircle(screen, float32(x), float32(y), rad, withAlpha(dimmed(p.Colour, 0.5), alpha), true)
		if n := p.layersInFlower(g.shown()); n > 0 {
			frac := float64(n) / float64(len(p.Layers))
			vector.DrawFilledCircle(screen, float32(x), float32(y), rad, withAlpha(colAccent, (0.2+0.55*frac)*alpha), true)
		}
		width := float32(2)
		if hc != nil && hc.Flowers[p.Entity] {
			vector.StrokeCircle(screen, float32(x), float32(y), rad+2, 2, colText, true)
		}
		if p == g.hover || p == g.selected {
			width = 4
			vector.StrokeCircle(screen, float32(x), float32(y), rad+2, 1.5, colText, true)
		}
		vector.StrokeCircle(screen, float32(x), float32(y), rad, width, withAlpha(p.Colour, alpha), true)
	}

	if g.nest != nil {
		g.drawSearchLines(screen)
	}
	g.drawTurnBee(screen)
	var labels []mapLabel // drawn last, on top of everything
	if hc != nil {
		labels = g.drawKnownPatches(screen, hc)
	} else if p := g.focusPatch(); p != nil {
		labels = g.drawPatchVisitors(screen, p)
	}
	g.drawBadgers(screen)
	g.drawColonies(screen)
	g.drawLegend(screen)
	g.drawScaleBar(screen)
	drawLabels(screen, labels)
}

// colonyRadius is the radius of the nest symbol of a colony on the map [screen pixels]. It grows with the
// number of bees, brood included, up to 400 bees.
func colonyRadius(c *colonyView) float64 {
	return 3.5 + 6*math.Sqrt(min(c.Size(), 400)/400)
}

// drawColonies draws every colony as a clump of three wax pots at its nest, the larger, the more bees.
// Colonies founded in the last three days get a ring.
func (g *Game) drawColonies(screen *ebiten.Image) {
	now := float64(g.shown().Tick)
	for i := range g.shown().Colonies {
		c := &g.shown().Colonies[i]
		x, y := g.layout.toScreen(c.X, c.Y)
		r := colonyRadius(c)
		if now-c.Founded <= 3 {
			vector.StrokeCircle(screen, float32(x), float32(y), float32(r+5), 2, colAccent, true)
		}
		edge := colBackground
		if c == g.hoverColony || g.isSelected(c.Who) {
			edge = colText
			vector.StrokeCircle(screen, float32(x), float32(y), float32(r+5), 2.5, colText, true)
		}
		if c.CuckooQueen {
			vector.StrokeCircle(screen, float32(x), float32(y), float32(r+2.5), 2, colCuckoo, true)
		}
		drawNest(screen, x, y, r, g.colonyColour(c.Species), edge)
		if c.CuckooQueen {
			vector.DrawFilledCircle(screen, float32(x), float32(y), float32(max(2, r*0.4)), colCuckoo, true)
		}
	}
}

// colonyColour is the fill of a colony on the map: the colour of its species.
func (g *Game) colonyColour(species string) color.RGBA {
	if c, ok := g.speciesColour[species]; ok {
		return c
	}
	return colNest
}

// focusColony is the colony whose range and patches the map shows: the one under the mouse, else the selected
// one while it lives.
func (g *Game) focusColony() *colonyView {
	if g.hoverColony != nil {
		return g.hoverColony
	}
	if g.selColony == nil {
		return nil
	}
	for i := range g.shown().Colonies {
		if g.shown().Colonies[i].Who == g.selColony.Last.Who {
			return &g.shown().Colonies[i]
		}
	}
	return nil
}

func (g *Game) isSelected(who int64) bool {
	return g.selColony != nil && g.selColony.Last.Who == who
}

// flow is a line from (x0, y0) to (x1, y1) for n bees, all in screen pixels.
type flow struct {
	x0, y0, x1, y1 float64
	n              float64
	col            color.RGBA
	side           float64 // -1 or 1: nectar and pollen lines between the same two places run side by side
}

// flowOrder sorts flows by the number of bees, then by their end and side. The full order matters: the flows
// come from maps with a random order, and lines with equal numbers would otherwise swap places and labels from
// frame to frame (flicker).
func flowOrder(a, b flow) int {
	return cmp.Or(cmp.Compare(a.n, b.n), cmp.Compare(a.x1, b.x1), cmp.Compare(a.y1, b.y1), cmp.Compare(a.side, b.side))
}

// drawFlows draws the lines. Width and opacity grow with the number of bees, relative to the strongest line, so
// the main sources stand out. The strongest lines lie on top, the five strongest show their number of bees.
func drawFlows(screen *ebiten.Image, flows []flow) {
	most := 0.0
	for _, f := range flows {
		most = max(most, f.n)
	}
	slices.SortFunc(flows, flowOrder)
	for _, f := range flows {
		d := math.Hypot(f.x1-f.x0, f.y1-f.y0)
		if d == 0 || f.n == 0 {
			continue
		}
		rel := f.n / most
		width := 1 + 7*rel
		off := f.side * (width/2 + 0.5)
		ox, oy := -(f.y1-f.y0)/d*off, (f.x1-f.x0)/d*off
		vector.StrokeLine(screen, float32(f.x0+ox), float32(f.y0+oy), float32(f.x1+ox), float32(f.y1+oy),
			float32(width), withAlpha(f.col, 0.35+0.6*rel), true)
	}
}

// flowLabels returns the labels of the five strongest flows: their number of bees, placed on their line.
func flowLabels(flows []flow) []mapLabel {
	out := []mapLabel{}
	for i := len(flows) - 1; i >= 0 && i >= len(flows)-5; i-- {
		f := flows[i]
		if f.n == 0 {
			break
		}
		out = append(out, mapLabel{x0: f.x0, y0: f.y0, x1: f.x1, y1: f.y1, text: beesText(f.n), col: f.col})
	}
	return out
}

// beesText is a number of bees, e.g. "1 bee", "25 bees".
func beesText(n float64) string {
	if n == 1 {
		return "1 bee"
	}
	return fmt.Sprintf("%.0f bees", n)
}

// mapLabel is a text on a line of the map from (x0, y0) to (x1, y1).
type mapLabel struct {
	x0, y0, x1, y1 float64
	text           string
	col            color.RGBA
}

// drawLabels draws the labels on top of everything, where placeLabels puts them.
func drawLabels(screen *ebiten.Image, labels []mapLabel) {
	for i, r := range placeLabels(labels) {
		fillRect(screen, r, withAlpha(colPanel, 0.9))
		drawText(screen, labels[i].text, faceSmall, float64(r.Min.X)+3, float64(r.Min.Y), labels[i].col)
	}
}

// placeLabels returns the box of every label. Each label takes the first place along its line where it does not
// cover a label placed before; if there is none, it moves below until it is free.
func placeLabels(labels []mapLabel) []image.Rectangle {
	placed := []image.Rectangle{}
	free := func(r image.Rectangle) bool {
		for _, p := range placed {
			if r.Overlaps(p.Inset(-2)) {
				return false
			}
		}
		return true
	}
	for _, l := range labels {
		lw, lh := measureFace(l.text, faceSmall)
		rectAt := func(t, dy float64) image.Rectangle {
			cx, cy := l.x0+t*(l.x1-l.x0), l.y0+t*(l.y1-l.y0)+dy
			return image.Rect(int(cx-lw/2-3), int(cy-lh/2-1), int(cx+lw/2+3), int(cy+lh/2+1))
		}
		r, found := image.Rectangle{}, false
		for _, t := range []float64{0.7, 0.55, 0.85, 0.4, 0.95, 0.3, 0.2} {
			if c := rectAt(t, 0); free(c) {
				r, found = c, true
				break
			}
		}
		for dy := lh + 4; !found; dy += lh + 4 {
			if c := rectAt(0.7, dy); free(c) {
				r, found = c, true
			}
		}
		placed = append(placed, r)
	}
	return placed
}

// drawKnownPatches draws a line from the colony to every flower patch its bees know as a nectar source (yellow)
// or pollen source (blue): where its bees fly to.
func (g *Game) drawKnownPatches(screen *ebiten.Image, c *colonyView) []mapLabel {
	cx, cy := g.layout.toScreen(c.X, c.Y)
	flows := []flow{}
	for _, k := range []struct {
		known map[ecs.Entity]float64
		col   color.RGBA
		side  float64
	}{{c.KnownNectar, colAccent, -1}, {c.KnownPollen, colSecond, 1}} {
		for e, n := range k.known {
			if p := g.land.patch(e); p != nil {
				px, py := g.layout.toScreen(p.X, p.Y)
				flows = append(flows, flow{cx, cy, px, py, n, k.col, k.side})
			}
		}
	}
	drawFlows(screen, flows)
	return flowLabels(flows)
}

// focusPatch is the flower patch under the mouse, else the clicked one.
func (g *Game) focusPatch() *flowerPatch {
	if g.hover != nil {
		return g.hover
	}
	return g.selected
}

// drawPatchVisitors draws a line from a flower patch to every colony whose bees know it as a nectar source
// (yellow) or pollen source (blue): where its bees come from.
func (g *Game) drawPatchVisitors(screen *ebiten.Image, p *flowerPatch) []mapLabel {
	px, py := g.layout.toScreen(p.X, p.Y)
	flows := []flow{}
	for i := range g.shown().Colonies {
		c := &g.shown().Colonies[i]
		cx, cy := g.layout.toScreen(c.X, c.Y)
		flows = append(flows,
			flow{px, py, cx, cy, c.KnownNectar[p.Entity], colAccent, -1},
			flow{px, py, cx, cy, c.KnownPollen[p.Entity], colSecond, 1})
	}
	drawFlows(screen, flows)
	return flowLabels(flows)
}

// drawBadgers draws every badger with the range in which it can dig up colonies.
func (g *Game) drawBadgers(screen *ebiten.Image) {
	rad := float32(g.land.BadgerRange * g.land.Scale * g.layout.scale)
	for _, b := range g.shown().Badgers {
		x, y := g.layout.toScreen(b[0], b[1])
		vector.StrokeCircle(screen, float32(x), float32(y), rad, 1, withAlpha(colText, 0.3), true)
		drawBadger(screen, x, y)
	}
}

// drawBadger draws the badger symbol: a dark disc with a white stripe.
func drawBadger(screen *ebiten.Image, x, y float64) {
	vector.DrawFilledCircle(screen, float32(x), float32(y), 6, colBadger, true)
	vector.StrokeLine(screen, float32(x), float32(y-5), float32(x), float32(y+5), 2, colText, true)
	vector.StrokeCircle(screen, float32(x), float32(y), 6, 1.5, colText, true)
}

// colonyAt returns the colony under the screen position, or nil.
func (g *Game) colonyAt(sx, sy float64) *colonyView {
	var best *colonyView
	bestD := math.Inf(1)
	for i := range g.shown().Colonies {
		c := &g.shown().Colonies[i]
		x, y := g.layout.toScreen(c.X, c.Y)
		if d := math.Hypot(sx-x, sy-y); d < max(7, colonyRadius(c)+2) && d < bestD {
			best, bestD = c, d
		}
	}
	return best
}

// drawRange draws the foraging range around a nest at screen position (x, y).
func (g *Game) drawRange(screen *ebiten.Image, x, y float64) {
	rad := float32(g.land.MaxForagingRange * g.land.Scale * g.layout.scale)
	vector.DrawFilledCircle(screen, float32(x), float32(y), rad, color.RGBA{0x18, 0x18, 0x18, 0x18}, true)
	const segments = 90
	for i := 0; i < segments; i += 2 { // dashed
		a0 := 2 * math.Pi * float64(i) / segments
		a1 := 2 * math.Pi * float64(i+1) / segments
		vector.StrokeLine(screen,
			float32(x)+rad*float32(math.Cos(a0)), float32(y)+rad*float32(math.Sin(a0)),
			float32(x)+rad*float32(math.Cos(a1)), float32(y)+rad*float32(math.Sin(a1)), 1.5, colText, true)
	}
	drawText(screen, fmt.Sprintf("%.0f m", g.land.MaxForagingRange), faceSmall, float64(x)+4, float64(y)-float64(rad)+4, colText) // inside the circle, stays on the map
}

// drawSearchLines draws a line from the test nest to the edge of every reachable patch in flower.
// The stronger the line, the more likely a searching bee finds the patch.
func (g *Game) drawSearchLines(screen *ebiten.Image) {
	nx, ny := g.layout.toScreen(g.nest[0], g.nest[1])
	for i := range g.land.Patches {
		p := &g.land.Patches[i]
		re := g.land.reachFrom(p, g.nest[0], g.nest[1])
		if !re.InRange || p.layersInFlower(g.shown()) == 0 || re.Edge == 0 {
			continue
		}
		px, py := g.layout.toScreen(p.X, p.Y)
		d := math.Hypot(px-nx, py-ny)
		rad := max(p.Radius*g.layout.scale, minRadius)
		ex, ey := px-(px-nx)/d*rad, py-(py-ny)/d*rad // point on the edge facing the nest
		strong := p == g.hover || p == g.selected
		c := withAlpha(colText, 0.15+0.75*re.FindProb)
		width := float32(1 + 3*re.FindProb)
		if strong {
			c, width = colText, width+1
		}
		vector.StrokeLine(screen, float32(nx), float32(ny), float32(ex), float32(ey), width, c, true)
	}
	vector.DrawFilledCircle(screen, float32(nx), float32(ny), 7, colBee, true)
	vector.StrokeCircle(screen, float32(nx), float32(ny), 7, 2, colText, true)
}

// drawLegend draws the habitat colours in the upper left corner of the map.
func (g *Game) drawLegend(screen *ebiten.Image) {
	x := float64(g.layout.mapRect.Min.X + 14)
	y := float64(g.layout.mapRect.Min.Y + 12)
	rows := len(g.land.Habitats) + 2
	if g.nest != nil {
		rows++
	}
	if len(g.shown().Badgers) > 0 {
		rows++
	}
	species := g.legendSpecies()
	rows += len(species)
	cuckoos := g.shown().Bees.Cuckoos+g.shown().Bees.CuckooSearching > 0
	if cuckoos {
		rows++
	}
	flows := g.focusColony() != nil || g.focusPatch() != nil // lines between colonies and patches are shown
	if flows {
		rows++
	}
	box := image.Rect(int(x-8), int(y-6), int(x-8)+340, int(y-6)+rows*19+16)
	if image.Pt(ebiten.CursorPosition()).In(box) {
		return // hidden while the mouse is on it, so the colonies and patches below can be reached
	}
	fillRect(screen, box, withAlpha(colPanel, 0.85))
	for _, h := range g.land.Habitats {
		vector.StrokeCircle(screen, float32(x+6), float32(y+9), 6, 2, h.Colour, true)
		drawText(screen, fmt.Sprintf("%s (%d)", prettyName(h.Name), h.Patches), faceSmall, x+18, y, colText)
		y += 19
	}
	y += 4
	vector.DrawFilledCircle(screen, float32(x+6), float32(y+9), 6, withAlpha(colAccent, 0.75), true)
	drawText(screen, "in flower today", faceSmall, x+18, y, colText)
	y += 19
	drawNest(screen, x+6, y+9, 5, colNest, colBackground)
	drawText(screen, fmt.Sprintf("colony (%d), larger with more bees, click to follow it", len(g.shown().Colonies)), faceSmall, x+18, y, colText)
	y += 19
	for _, s := range species {
		drawNest(screen, x+18, y+9, 5, g.colonyColour(s), colBackground)
		drawText(screen, prettyName(s), faceSmall, x+30, y, colText)
		y += 19
	}
	if cuckoos {
		drawNest(screen, x+6, y+9, 5, colNest, colBackground)
		vector.DrawFilledCircle(screen, float32(x+6), float32(y+9), 2, colCuckoo, true)
		vector.StrokeCircle(screen, float32(x+6), float32(y+9), 7, 1.5, colCuckoo, true)
		drawText(screen, "colony with a cuckoo bee queen", faceSmall, x+18, y, colText)
		y += 19
	}
	if len(g.shown().Badgers) > 0 {
		drawBadger(screen, x+6, y+9)
		drawText(screen, fmt.Sprintf("badger (%d) and the range it digs in", len(g.shown().Badgers)), faceSmall, x+18, y, colText)
		y += 19
	}
	if g.nest != nil {
		vector.DrawFilledCircle(screen, float32(x+6), float32(y+9), 5, colBee, true)
		drawText(screen, "test nest (double right-click removes it)", faceSmall, x+18, y, colText)
		y += 19
	}
	if flows {
		// One entry: a yellow and a blue line as icon, "nectar" and "pollen" in their colour.
		vector.StrokeLine(screen, float32(x), float32(y+6), float32(x+12), float32(y+6), 2.5, colAccent, true)
		vector.StrokeLine(screen, float32(x), float32(y+12), float32(x+12), float32(y+12), 2.5, colSecond, true)
		tx := x + 18
		for _, part := range []struct {
			text string
			col  color.RGBA
		}{{"living bees that know the patch for ", colText}, {"nectar", colAccent}, {" / ", colText}, {"pollen", colSecond}} {
			drawText(screen, part.text, faceSmall, tx, y, part.col)
			w, _ := measureFace(part.text, faceSmall)
			tx += w
		}
	}
}

// legendSpecies returns the species with their own colony colour, if there is more than one.
func (g *Game) legendSpecies() []string {
	if len(g.speciesColour) < 2 {
		return nil
	}
	out := []string{}
	for s := range g.speciesColour {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

// drawScaleBar draws a scale bar with a round length in metres.
func (g *Game) drawScaleBar(screen *ebiten.Image) {
	pxPerMetre := g.land.Scale * g.layout.scale
	if pxPerMetre <= 0 {
		return
	}
	target := 120 / pxPerMetre // metres for about 120 px
	step := math.Pow(10, math.Floor(math.Log10(target)))
	for _, f := range []float64{5, 2, 1} {
		if f*step <= target {
			step *= f
			break
		}
	}
	length := step * pxPerMetre
	x0 := float32(g.layout.mapRect.Min.X + 14)
	y0 := float32(g.layout.mapRect.Max.Y - 18)
	vector.StrokeLine(screen, x0, y0, x0+float32(length), y0, 2, colText, false)
	vector.StrokeLine(screen, x0, y0-5, x0, y0+1, 2, colText, false)
	vector.StrokeLine(screen, x0+float32(length), y0-5, x0+float32(length), y0+1, 2, colText, false)
	label := fmt.Sprintf("%.0f m", step)
	if step >= 1000 {
		label = fmt.Sprintf("%.0f km", step/1000)
	}
	drawText(screen, label, faceSmall, float64(x0), float64(y0)-22, colText)
}

// patchAt returns the smallest flower patch under the screen position, or nil.
func (g *Game) patchAt(sx, sy float64) *flowerPatch {
	if g.land == nil {
		return nil
	}
	var best *flowerPatch
	bestR := math.Inf(1)
	for i := range g.land.Patches {
		p := &g.land.Patches[i]
		x, y := g.layout.toScreen(p.X, p.Y)
		rad := max(p.Radius*g.layout.scale, minRadius) + 2
		if math.Hypot(sx-x, sy-y) <= rad && rad < bestR {
			best, bestR = p, rad
		}
	}
	return best
}

// dimmed returns c with its red, green and blue multiplied by f.
func dimmed(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), c.A}
}

// drawTurnBee marks, after a bee turn (debug tab), the nest of the bee and the flower patches of her next nectar
// trip (yellow) and pollen trip (blue).
func (g *Game) drawTurnBee(screen *ebiten.Image) {
	b := g.dbg.bee
	if g.viewing || g.stopAt == nil || g.stopAt.kind != stopTurn || b == nil || !b.alive || b.colony == 0 {
		return
	}
	cx, cy := g.layout.toScreen(b.colonyX, b.colonyY)
	for _, t := range []struct {
		patch ecs.Entity
		col   color.RGBA
	}{{b.nectarSrc, colAccent}, {b.pollenSrc, colSecond}} {
		if p := g.land.patch(t.patch); p != nil {
			px, py := g.layout.toScreen(p.X, p.Y)
			vector.StrokeLine(screen, float32(cx), float32(cy), float32(px), float32(py), 2, t.col, true)
			vector.StrokeCircle(screen, float32(px), float32(py), float32(max(p.Radius*g.layout.scale, minRadius)+3),
				2, t.col, true)
		}
	}
	vector.StrokeCircle(screen, float32(cx), float32(cy), 14, 3, colAccent, true)
}
