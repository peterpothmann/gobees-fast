//go:build ui || windows || (darwin && cgo)

package ui

import (
	"fmt"
	"image"
	"reflect"
	"slices"
	"strings"
	"time"

	"gobees/comp"
	"gobees/globals"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mlange-42/ark/ecs"
)

// debugState is what the debug tab shows of the model where it stopped. It is read while the model waits for the
// next command (driver), never while it runs.
type debugState struct {
	archetypes []archetypeRow // non-empty archetypes, most entities first
	entities   int
	memory     int         // memory of the component storage [bytes]
	resources  [][2]string // name and value of the resources of package globals
	bee        *beeView    // the bee of the last turn, nil if the model did not stop after a turn
}

// archetypeRow is one archetype of the world: entities with the same components.
type archetypeRow struct {
	components string
	entities   int
}

// beeView is the bee whose turn it was, after her turn.
type beeView struct {
	alive                bool // false: she died in her turn (on a foraging trip)
	number               int
	caste, stage, task   string
	colony               int64   // the number the window gives her colony, 0 without colony
	colonyX, colonyY     float64 // the nest
	start, end           float64 // personal time before and after the turn [s]
	thresholds, stimuli  comp.Thresholds
	hasStimuli           bool // the bee has a colony, so the stimuli are those of her colony
	energy, pollen       float64
	carriedN, carriedP   float64    // load of the bee or cohort: nectar [kJ], pollen [g]
	nectarSrc, pollenSrc ecs.Entity // masterpatches of the next nectar and pollen trip, zero for none
}

// debugReader reads the world for the debug tab.
type debugReader struct {
	world    *ecs.World
	daytime  *globals.Daytime
	bees     *ecs.Map3[comp.Bee, comp.Thresholds, comp.Forager]
	colonies *ecs.Map3[comp.Position, comp.ColonyStimuli, comp.ColonyStores]
	layers   *ecs.Map1[comp.Layer]
	globals  []any // pointers to the resources of package globals
}

func newDebugReader(w *ecs.World) *debugReader {
	return &debugReader{
		world:    w,
		daytime:  ecs.GetResource[globals.Daytime](w),
		bees:     ecs.NewMap3[comp.Bee, comp.Thresholds, comp.Forager](w),
		colonies: ecs.NewMap3[comp.Position, comp.ColonyStimuli, comp.ColonyStores](w),
		layers:   ecs.NewMap1[comp.Layer](w),
		globals: []any{
			ecs.GetResource[globals.Ticks](w), ecs.GetResource[globals.Calendar](w),
			ecs.GetResource[globals.Daytime](w), ecs.GetResource[globals.ActiveBees](w),
			ecs.GetResource[globals.Flowering](w), ecs.GetResource[globals.BroodTotals](w),
			ecs.GetResource[globals.Landscape](w), ecs.GetResource[globals.NestSites](w),
		},
	}
}

// read collects the debug state. bee is the bee of the last turn, zero if the model did not stop after a turn.
// known gives the window's colony numbers.
func (r *debugReader) read(bee ecs.Entity, known map[ecs.Entity]colonyID) debugState {
	var d debugState
	st := r.world.Stats()
	d.entities, d.memory = st.Entities.Used, st.MemoryUsed
	for _, a := range st.Archetypes {
		n := 0
		for _, t := range a.Tables {
			n += t.Size
		}
		if n == 0 {
			continue
		}
		names := make([]string, len(a.ComponentTypeNames))
		for i, name := range a.ComponentTypeNames {
			names[i] = strings.TrimPrefix(name, "comp.")
		}
		d.archetypes = append(d.archetypes, archetypeRow{strings.Join(names, " "), n})
	}
	slices.SortStableFunc(d.archetypes, func(a, b archetypeRow) int { return b.entities - a.entities })
	for _, res := range r.globals {
		d.resources = append(d.resources, resourceRows(res)...)
	}
	if !bee.IsZero() {
		d.bee = r.readBee(bee, known)
	}
	return d
}

// resourceRows formats the fields of a resource: one row per field, slices and maps by their length.
func resourceRows(res any) [][2]string {
	v := reflect.ValueOf(res).Elem()
	t := v.Type()
	out := [][2]string{}
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		fv := v.Field(i)
		val := ""
		switch fv.Kind() {
		case reflect.Slice, reflect.Map:
			val = fmt.Sprintf("%d items", fv.Len())
		case reflect.Float64:
			val = formatFloat(fv.Float())
		default:
			val = fmt.Sprint(fv.Interface())
		}
		out = append(out, [2]string{t.Name() + "." + f.Name, val})
	}
	return out
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%.6g", v)
}

func (r *debugReader) readBee(e ecs.Entity, known map[ecs.Entity]colonyID) *beeView {
	v := &beeView{start: r.daytime.Time}
	if !r.world.Alive(e) {
		return v
	}
	b, th, fo := r.bees.Get(e)
	v.alive, v.number, v.end = true, b.Number, b.PersonalTime
	v.caste, v.stage, v.task = b.Caste.String(), b.Stage.String(), b.Activity.String()
	v.thresholds = *th
	v.carriedN, v.carriedP = fo.CarriedNectar, fo.CarriedPollen
	if !fo.NectarSource.IsZero() && r.world.Alive(fo.NectarSource) {
		v.nectarSrc = r.layers.Get(fo.NectarSource).Masterpatch
	}
	if !fo.PollenSource.IsZero() && r.world.Alive(fo.PollenSource) {
		v.pollenSrc = r.layers.Get(fo.PollenSource).Masterpatch
	}
	if !b.Colony.IsZero() && r.world.Alive(b.Colony) {
		pos, stim, stores := r.colonies.Get(b.Colony)
		v.colony = known[b.Colony].Who
		v.colonyX, v.colonyY = pos.X, pos.Y
		v.stimuli = comp.Thresholds{Egglaying: stim.Egglaying, NectarForaging: stim.NectarForaging,
			PollenForaging: stim.PollenForaging, Nursing: stim.Nursing}
		v.hasStimuli = true
		v.energy, v.pollen = stores.Energy, stores.Pollen
	}
	return v
}

// clock formats a time of day in seconds as hh:mm.
func clock(s float64) string {
	m := int(s) / 60
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

// stepTurn, stepPart: the buttons and keys of the debug tab.
func (g *Game) stepTurn() { g.paused = true; g.stepTo(toTurn) }
func (g *Game) stepPart() { g.paused = true; g.stepTo(toPart) }

// drawDebug draws the debug tab from y down to bottom.
func (g *Game) drawDebug(screen *ebiten.Image, x, y, w, bottom float64) {
	// where the model is
	t := g.snap.Tick
	where := fmt.Sprintf("end of day %d (%d parts)", t, len(g.drv.lastDay))
	if s := g.stopAt; s != nil {
		where = fmt.Sprintf("day %d · part %d: %s.%s", t+1, s.partIndex, s.system, s.part)
		if s.kind == stopTurn {
			where = fmt.Sprintf("day %d · %s.%s · turn %d at %s", t+1, s.system, s.part, s.turn,
				clock(g.dbg.bee.start))
		}
	}
	drawText(screen, where, faceBold, x, y, colText)
	y += 24

	// step buttons
	labels := []struct {
		label  string
		action func()
	}{{"Next turn (T)", g.stepTurn}, {"Next part (P)", g.stepPart}, {"End of day (→)", g.stepOnce}}
	gap := 8.0
	bw := (w - gap*float64(len(labels)-1)) / float64(len(labels))
	for i, l := range labels {
		bx := x + float64(i)*(bw+gap)
		g.addButton(screen, image.Rect(int(bx), int(y), int(bx+bw), int(y+26)), l.label, l.action)
	}
	y += 36

	if b := g.dbg.bee; b != nil {
		y = g.drawBee(screen, x, y, w, b) + 10
	}

	// two columns: time per part on the left, entities and resources on the right
	colW := (w - 16) / 2
	area := screen.SubImage(image.Rect(int(x), int(y), int(x+w), int(bottom))).(*ebiten.Image)
	g.drawTimes(area, x, y, colW)
	ry := g.drawEntities(area, x+colW+16, y, colW)
	g.drawResources(area, x+colW+16, ry+10, colW)
}

// drawBee shows the bee of the last turn: what she chose and why (stimuli of her colony against her thresholds).
func (g *Game) drawBee(screen *ebiten.Image, x, y, w float64, b *beeView) float64 {
	if !b.alive {
		return y + drawText(screen, "The bee died in her turn (on a foraging trip).", faceSmall, x, y, colError)
	}
	who := "no colony"
	if b.colony != 0 {
		who = fmt.Sprintf("colony %d", b.colony)
	}
	head := fmt.Sprintf("%s %s ×%d · %s · task: %s · %s → %s", b.stage, b.caste, b.number, who, b.task,
		clock(b.start), clock(b.end))
	y += drawText(screen, truncate(head, w), faceSmall, x, y, colAccent) + 4
	if b.hasStimuli {
		rows := []struct {
			name     string
			stim, th float64
		}{
			{"egglaying", b.stimuli.Egglaying, b.thresholds.Egglaying},
			{"nursing", b.stimuli.Nursing, b.thresholds.Nursing},
			{"pollen foraging", b.stimuli.PollenForaging, b.thresholds.PollenForaging},
			{"nectar foraging", b.stimuli.NectarForaging, b.thresholds.NectarForaging},
		}
		drawText(screen, "task (the last one above its threshold wins)", faceSmall, x, y, colMuted)
		drawTextRight(screen, "stimulus", faceSmall, x+w-110, y, colMuted)
		drawTextRight(screen, "threshold", faceSmall, x+w, y, colMuted)
		y += 16
		for _, r := range rows {
			c := colText
			if r.stim > r.th {
				c = colAccent
			}
			drawText(screen, r.name, faceSmall, x+8, y, c)
			drawTextRight(screen, formatFloat(r.stim), faceSmall, x+w-110, y, c)
			drawTextRight(screen, formatFloat(r.th), faceSmall, x+w, y, c)
			y += 16
		}
		line := fmt.Sprintf("colony stores: %.1f kJ, %.2f g · load: nectar %.3g kJ, pollen %.3g g", b.energy,
			b.pollen, b.carriedN, b.carriedP)
		y += drawText(screen, truncate(line, w), faceSmall, x, y, colMuted)
	}
	return y
}

// drawTimes shows the time per part: of the day so far (or the last full day) and the mean over all days.
func (g *Game) drawTimes(dst *ebiten.Image, x, y, w float64) float64 {
	d := g.drv
	day := d.lastDay
	label := "last day"
	if g.stopAt != nil {
		day, label = d.today, "today"
	}
	drawText(dst, "Time per part [µs]", faceBold, x, y, colText)
	y += 20
	drawText(dst, "part", faceSmall, x, y, colMuted)
	drawTextRight(dst, label, faceSmall, x+w-60, y, colMuted)
	drawTextRight(dst, "mean", faceSmall, x+w, y, colMuted)
	y += 16
	took := map[string]time.Duration{}
	for _, p := range day {
		took[p.name] = p.took
	}
	var sumDay, sumMean float64
	for _, name := range d.order {
		short := name[strings.Index(name, ".")+1:]
		c := colText
		if s := g.stopAt; s != nil && s.system+"."+s.part == name {
			c = colAccent
		}
		dt, ok := took[name]
		dayText := "-"
		if ok {
			dayText = fmt.Sprintf("%.0f", us(dt))
			sumDay += us(dt)
		}
		mean := us(d.total[name]) / float64(max(d.fullDays, 1))
		sumMean += mean
		drawText(dst, truncate(short, w-120), faceSmall, x, y, c)
		drawTextRight(dst, dayText, faceSmall, x+w-60, y, c)
		drawTextRight(dst, fmt.Sprintf("%.0f", mean), faceSmall, x+w, y, c)
		y += 16
	}
	vector.StrokeLine(dst, float32(x), float32(y+1), float32(x+w), float32(y+1), 1, colBorder, false)
	y += 4
	drawText(dst, "sum", faceSmall, x, y, colText)
	drawTextRight(dst, fmt.Sprintf("%.0f", sumDay), faceSmall, x+w-60, y, colText)
	drawTextRight(dst, fmt.Sprintf("%.0f", sumMean), faceSmall, x+w, y, colText)
	return y + 16
}

// drawEntities shows the archetypes of the world: entities with the same components.
func (g *Game) drawEntities(dst *ebiten.Image, x, y, w float64) float64 {
	drawText(dst, fmt.Sprintf("Entities: %d · %.1f MB", g.dbg.entities, float64(g.dbg.memory)/1e6), faceBold, x, y,
		colText)
	y += 20
	for _, a := range g.dbg.archetypes {
		drawTextRight(dst, fmt.Sprint(a.entities), faceSmall, x+48, y, colText)
		y += drawText(dst, wrap(a.components, faceSmall, w-56), faceSmall, x+56, y, colMuted) + 2
	}
	return y
}

// drawResources shows the resources of package globals, the state of the run that is not part of an entity.
func (g *Game) drawResources(dst *ebiten.Image, x, y, w float64) float64 {
	drawText(dst, "Resources (globals)", faceBold, x, y, colText)
	y += 20
	for _, r := range g.dbg.resources {
		drawText(dst, truncate(r[0], w-90), faceSmall, x, y, colMuted)
		drawTextRight(dst, truncate(r[1], 86), faceSmall, x+w, y, colText)
		y += 16
	}
	return y
}

// us returns d in microseconds.
func us(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1000
}
