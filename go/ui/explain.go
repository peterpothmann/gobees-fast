//go:build ui || windows || (darwin && cgo)

package ui

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The explanation window shows a schematic animation of the processes in BEE-STEWARD.
// It does not run the model: all numbers are made up to look plausible. Where a value is taken
// from the model (e.g. the 20 % daily chance to find a nest), the comment says so.

var (
	colQueen  = color.RGBA{0xf0, 0x6f, 0x9c, 0xff}
	colWorker = color.RGBA{0xe8, 0x9a, 0x3c, 0xff}
	colMale   = color.RGBA{0x6c, 0xb4, 0xee, 0xff}
	colSoil   = color.RGBA{0x3b, 0x2f, 0x26, 0xff}
	colDead   = color.RGBA{0xb0, 0x58, 0x58, 0xff}
	colNest   = color.RGBA{0xc8, 0x9b, 0x5a, 0xff}
)

// process is one entry of the process list next to the animation.
type process struct {
	title  string
	detail string
	proc   string  // NetLogo procedure, for orientation
	glow   float64 // 1 when the process just happened, fades to 0
}

func (p *process) fire() { p.glow = 1 }

// scene is one view of the explanation window.
type scene interface {
	update(dt float64) // dt: seconds of animation time
	draw(screen *ebiten.Image, area image.Rectangle)
	processes() []*process
	caption() string
	reset()
	// The time slider: range, current value, label of a value, and jumping to a value.
	timeRange() (lo, hi float64)
	timeValue() float64
	timeLabel(v float64) string
	seek(v float64)
	// click handles a click into the animation. It reports whether it used the click.
	click(x, y int) bool
}

type explainGame struct {
	shot    string // screenshot file, see Config.Screenshot
	frames  int
	shotErr error
	scenes  []scene
	names   []string
	current int
	paused  bool
	speed   float64
	buttons []button
	w, h    int

	slider   image.Rectangle // time slider of the current view
	dragging bool
	area     image.Rectangle // animation area
}

// RunExplain opens the explanation window and blocks until it is closed.
// For a screenshot, view selects the view, warm the seconds of animation to run before, shot the file.
func RunExplain(view int, warm float64, shot string) error {
	g := &explainGame{
		scenes: []scene{newYearScene(1), newDayScene(2)},
		names:  []string{"A year", "A day"},
		speed:  1,
		shot:   shot,
	}
	g.current = min(max(view, 0), len(g.scenes)-1)
	for range int(warm * 60) {
		g.scenes[g.current].update(1.0 / 60)
	}
	ebiten.SetWindowTitle("gobees: how the model works")
	ebiten.SetWindowSize(1320, 820)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(g); err != nil && !errors.Is(err, errScreenshotDone) {
		return err
	}
	return nil
}

func (g *explainGame) Update() error {
	if g.shot != "" {
		if g.frames >= 3 {
			if g.shotErr != nil {
				return g.shotErr
			}
			return errScreenshotDone
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.faster()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.slower()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.current = (g.current + 1) % len(g.scenes)
	}
	s := g.scenes[g.current]
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		used := false
		for _, b := range g.buttons {
			if b.contains(mx, my) {
				b.action()
				used = true
				break
			}
		}
		switch {
		case used:
		case image.Pt(mx, my).In(g.slider.Inset(-6)):
			g.dragging = true
		case image.Pt(mx, my).In(g.area):
			s.click(mx, my)
		}
	}
	if g.dragging {
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.dragging = false
		} else {
			lo, hi := s.timeRange()
			f := float64(mx-g.slider.Min.X) / float64(max(g.slider.Dx(), 1))
			s.seek(lo + min(max(f, 0), 1)*(hi-lo))
			return nil // no animation while dragging
		}
	}
	if !g.paused {
		dt := g.speed / float64(ebiten.TPS())
		s.update(dt)
		for _, p := range s.processes() {
			p.glow = max(p.glow-dt*0.6, 0)
		}
	}
	return nil
}

func (g *explainGame) faster() { g.speed = min(g.speed*1.5, 8) }
func (g *explainGame) slower() { g.speed = max(g.speed/1.5, 0.25) }

func (g *explainGame) Layout(w, h int) (int, int) {
	g.w, g.h = w, h
	return w, h
}

const explainPanelW = 380

func (g *explainGame) Draw(screen *ebiten.Image) {
	screen.Fill(colBackground)
	g.buttons = g.buttons[:0]

	// Header: title, scene tabs, controls.
	drawText(screen, "How the model works", faceTitle, pad, pad, colText)
	x := float64(pad + 270)
	for i, name := range g.names {
		i := i
		x = g.drawButton(screen, x, pad, name, i == g.current, func() { g.current = i }) + 8
	}
	play := "Pause"
	if g.paused {
		play = "Play"
	}
	right := float64(g.w - pad)
	defs := []struct {
		label  string
		action func()
	}{
		{"Restart", func() { g.scenes[g.current].reset() }},
		{"Faster", g.faster},
		{"Slower", g.slower},
		{play, func() { g.paused = !g.paused }},
	}
	for _, d := range defs {
		right = g.drawButtonRight(screen, right, pad, d.label, d.action) - 8
	}

	top := pad + 48
	bottom := g.h - pad - 22
	area := image.Rect(pad, top, g.w-explainPanelW-2*pad, bottom-90)
	g.area = area
	s := g.scenes[g.current]
	fillRect(screen, area, colPanel)
	s.draw(screen, area)
	strokeRect(screen, area, colBorder)

	g.drawSlider(screen, s, image.Rect(area.Min.X+90, area.Max.Y+18, area.Max.X-110, area.Max.Y+26))

	// Caption below the animation.
	capRect := image.Rect(area.Min.X, area.Max.Y+44, area.Max.X, bottom)
	drawText(screen, wrap(s.caption(), faceNormal, float64(capRect.Dx())), faceNormal, float64(capRect.Min.X), float64(capRect.Min.Y), colText)

	// Process list.
	panel := image.Rect(g.w-explainPanelW-pad, top, g.w-pad, bottom)
	fillRect(screen, panel, colPanel)
	strokeRect(screen, panel, colBorder)
	px := float64(panel.Min.X + pad)
	pw := float64(panel.Dx() - 2*pad)
	py := float64(panel.Min.Y + pad)
	drawText(screen, "Processes", faceBold, px, py, colText)
	py += 28
	for _, p := range s.processes() {
		c := withAlpha(colAccent, 0.25+0.75*p.glow)
		vector.DrawFilledCircle(screen, float32(px+5), float32(py+9), 5, c, true)
		tc := colMuted
		if p.glow > 0.05 {
			tc = colText
		}
		drawText(screen, p.title, faceBold, px+18, py, tc)
		// The NetLogo procedure goes on the title line if there is room, otherwise below the text.
		tw, _ := textMeasure(p.title)
		pw2, _ := textMeasure(p.proc)
		procInline := p.proc != "" && tw+pw2*faceSmall.Size/faceNormal.Size+30 < pw-18
		if procInline {
			drawTextRight(screen, p.proc, faceSmall, px+pw, py+2, withAlpha(colMuted, 0.6))
		}
		py += 20
		py += drawText(screen, wrap(p.detail, faceSmall, pw-18), faceSmall, px+18, py, colMuted)
		if p.proc != "" && !procInline {
			py += drawText(screen, p.proc, faceSmall, px+18, py, withAlpha(colMuted, 0.6))
		}
		py += 8
	}

	if g.shot != "" {
		g.frames++
		if g.frames == 3 {
			g.shotErr = saveScreenshot(screen, g.shot)
		}
	}

	foot := "Schematic illustration of the processes in BEE-STEWARD. Not model output.   Space: pause · ↑↓: speed · Tab: switch view · drag the slider to move in time"
	drawText(screen, foot, faceSmall, pad, float64(g.h-pad-16), colMuted)
}

// drawSlider draws the time slider. Drag the handle to move through time.
func (g *explainGame) drawSlider(screen *ebiten.Image, s scene, r image.Rectangle) {
	g.slider = r
	lo, hi := s.timeRange()
	f := (s.timeValue() - lo) / (hi - lo)
	f = min(max(f, 0), 1)
	drawText(screen, "time", faceSmall, float64(g.area.Min.X), float64(r.Min.Y)-5, colMuted)
	vector.DrawFilledRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), colButton, false)
	vector.DrawFilledRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(float64(r.Dx())*f), float32(r.Dy()), withAlpha(colAccent, 0.6), false)
	hx := float32(float64(r.Min.X) + float64(r.Dx())*f)
	hy := float32(r.Min.Y + r.Dy()/2)
	c := colText
	if g.dragging {
		c = colAccent
	}
	vector.DrawFilledCircle(screen, hx, hy, 9, c, true)
	drawText(screen, s.timeLabel(s.timeValue()), faceSmall, float64(r.Max.X)+14, float64(r.Min.Y)-5, colText)
}

func (g *explainGame) drawButton(screen *ebiten.Image, x, y float64, label string, active bool, action func()) float64 {
	w := measure(label) + 28
	r := image.Rect(int(x), int(y), int(x+w), int(y+30))
	g.buttons = append(g.buttons, button{rect: r, label: label, action: action})
	c := colButton
	mx, my := ebiten.CursorPosition()
	if active {
		c = colButtonHi
	} else if image.Pt(mx, my).In(r) {
		c = colButtonHi
	}
	fillRect(screen, r, c)
	if active {
		vector.StrokeLine(screen, float32(r.Min.X), float32(r.Max.Y)-1, float32(r.Max.X), float32(r.Max.Y)-1, 2, colAccent, false)
	}
	drawText(screen, label, faceNormal, x+14, y+6, colText)
	return x + w
}

func (g *explainGame) drawButtonRight(screen *ebiten.Image, right, y float64, label string, action func()) float64 {
	w := measure(label) + 28
	g.drawButton(screen, right-w, y, label, false, action)
	return right - w
}

// lerp interpolates linearly between a and b.
func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// smooth is a smooth step from 0 to 1.
func smooth(t float64) float64 {
	t = min(max(t, 0), 1)
	return t * t * (3 - 2*t)
}

// drawCross draws a small cross, e.g. for a dead bee.
func drawCross(screen *ebiten.Image, x, y, r float64, c color.Color) {
	vector.StrokeLine(screen, float32(x-r), float32(y-r), float32(x+r), float32(y+r), 2, c, true)
	vector.StrokeLine(screen, float32(x-r), float32(y+r), float32(x+r), float32(y-r), 2, c, true)
}

// drawHex draws a filled hexagon, the symbol of a nest.
func drawHex(screen *ebiten.Image, x, y, r float64, fill, edge color.RGBA) {
	var p vector.Path
	for i := 0; i < 6; i++ {
		a := float64(i)*1.0471976 + 0.5235988
		px, py := x+r*math.Cos(a), y+r*math.Sin(a)
		if i == 0 {
			p.MoveTo(float32(px), float32(py))
		} else {
			p.LineTo(float32(px), float32(py))
		}
	}
	p.Close()
	vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
	setColor(vs, fill)
	screen.DrawTriangles(vs, is, whitePixel, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	sop := &vector.StrokeOptions{Width: 1.5, LineJoin: vector.LineJoinRound}
	vs, is = p.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	setColor(vs, edge)
	screen.DrawTriangles(vs, is, whitePixel, &ebiten.DrawTrianglesOptions{AntiAlias: true})
}

var whitePixel = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}()

func setColor(vs []ebiten.Vertex, c color.RGBA) {
	r, g, b, a := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, float32(c.A)/255
	for i := range vs {
		vs[i].SrcX, vs[i].SrcY = 1, 1
		vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = r, g, b, a
	}
}
