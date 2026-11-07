//go:build ui || windows || (darwin && cgo)

// Package ui is the window of gobees --ui: it shows one run while it runs (map, daily plots, colony and patch
// info, going back in time). Built on Windows always, on macOS with cgo, on Linux only with the build tag ui
// (go run -tags ui . --ui), because there Ebitengine needs cgo and the X11 and OpenGL headers, which a cluster
// often lacks.
package ui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"slices"
	"sync/atomic"
	"time"

	"gobees/globals"
	"gobees/obs"
	"gobees/params"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Config holds the settings of the window.
type Config struct {
	Title string // the run, shown next to the name, e.g. the name of the run file

	// Screenshot: if set, the run goes to its end at once, then a screenshot is saved to this PNG file and the
	// window closes. For checks without clicking.
	Screenshot string
}

// ExplainFlag is the flag of gobees that opens only the explanation window. The button "How the model works"
// starts gobees again with it, because Ebitengine has one window per process.
const ExplainFlag = "--ui-explain"

// doubleClick is the maximum time between the two clicks of a double click.
const doubleClick = 400 * time.Millisecond

// speeds are the selectable simulation speeds in days per second.
var speeds = []float64{0.5, 1, 2, 5, 10, 20, 50, 100, 365}

// Game is the ebiten.Game of the UI. It steps one model run, day by day.
type Game struct {
	cfg Config

	app           *app.App
	ticks         *globals.Ticks
	world         *params.World
	land          *landscape
	model         *modelState
	hist          history
	snap          snapshot                 // model state at the end of the latest day
	days          []snapshot               // the latest days, oldest first, the last one is snap
	viewing       bool                     // an earlier day is shown (← key)
	viewTick      int64                    // the tick shown while viewing
	backPending   float64                  // fraction of a day not yet stepped back
	colHist       map[int64]*colonyHistory // by colony number (colonyView.Who), also for dead colonies
	speciesColour map[string]color.RGBA
	seasons       []season
	finished      bool
	runErr        error

	paused  bool
	speed   int     // index into speeds
	pending float64 // fraction of a day not yet simulated

	width, height  int
	layout         layout
	hover          *flowerPatch // patch under the mouse
	hoverColony    *colonyView  // colony under the mouse
	selected       *flowerPatch // clicked patch
	nest           *[2]float64  // test nest in NetLogo coordinates, nil if none
	lastRightClick time.Time
	buttons        []button

	selColony *colonyHistory // clicked colony, nil if none

	explain     *exec.Cmd   // process of the explanation window
	explainOpen atomic.Bool // true while the explanation window is open

	frames  int   // drawn frames, for the screenshot
	shotErr error // error of saving the screenshot
}

// Watch opens the window and steps the run a day by day, at the speed chosen in the window. a must be set up with all
// systems of the run but not initialized: Watch calls Initialize, Update once per day and Finalize, like a.Run, so
// the run gives the same results and output as without the window. counters must be a system of a (the foraging
// trips of the day). Watch returns when the window is closed: at the end of the run, or before, then the run ends
// on that day. It must be called from the main goroutine. A failed day (panic) is returned as error.
func Watch(a *app.App, counters *obs.Counters, cfg Config) error {
	g := &Game{cfg: cfg, paused: true, speed: 3}
	if err := g.setup(a, counters); err != nil {
		return err
	}
	if cfg.Screenshot != "" {
		for !g.finished {
			g.step()
		}
	}

	ebiten.SetWindowTitle("gobees-fast · " + cfg.Title)
	ebiten.SetWindowSize(1440, 860)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	err := ebiten.RunGame(g)
	if g.explainOpen.Load() {
		_ = g.explain.Process.Kill()
	}
	if !g.finished {
		g.app.Finalize() // the window was closed before the end of the run
	}
	if err != nil && !errors.Is(err, errScreenshotDone) {
		return err
	}
	return g.runErr
}

// openExplain starts the explanation window as a second process, because Ebitengine has one window
// per process. It does nothing if the window is already open.
func (g *Game) openExplain() {
	if g.explainOpen.Load() {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		g.runErr = err
		return
	}
	cmd := exec.Command(exe, ExplainFlag)
	if err := cmd.Start(); err != nil {
		g.runErr = err
		return
	}
	g.explain = cmd
	g.explainOpen.Store(true)
	go func() {
		_ = cmd.Wait()
		g.explainOpen.Store(false)
	}()
}

// setup initializes the run (the model setup) and reads its first state. A setup panic is returned as error.
func (g *Game) setup(a *app.App, counters *obs.Counters) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("setup failed: %v", r)
		}
	}()
	a.Initialize()

	g.app = a
	g.ticks = ecs.GetResource[globals.Ticks](a.World)
	g.world = ecs.GetResource[params.World](a.World)
	g.land = newLandscape(a.World)
	g.model = newModel(a.World, counters)
	g.hist = history{}
	g.colHist = map[int64]*colonyHistory{}
	species := ecs.GetResource[params.Input](a.World).Species
	g.speciesColour = speciesColours(species)
	g.seasons = speciesSeasons(species)
	g.snap = g.model.read(int64(g.ticks.Ticks), g.land)
	g.days, g.viewing = []snapshot{g.snap}, false
	g.hist.record(&g.snap)
	recordColonies(g.colHist, &g.snap, int64(g.ticks.Ticks))
	return nil
}

// step simulates one day.
func (g *Game) step() {
	if g.finished {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			g.runErr = fmt.Errorf("day %d: %v", g.ticks.Ticks, r)
			g.finished = true
		}
	}()
	if !g.app.Update() {
		g.finished = true
		g.app.Finalize()
	}
	// The crop rotation replaces all foodsources on day 365: read the new patches. The days before show the old
	// patches, so they can no longer be viewed.
	if len(g.land.Patches) > 0 && !g.app.World.Alive(g.land.Patches[0].Entity) {
		g.land = newLandscape(g.app.World)
		g.hover, g.selected = nil, nil
		g.days = nil
	}
	g.snap = g.model.read(int64(g.ticks.Ticks), g.land)
	g.days = append(g.days, g.snap)
	if len(g.days) > maxBackDays+1 {
		g.days = slices.Clone(g.days[1:]) // a copy, so the dropped days are freed
	}
	g.hist.record(&g.snap)
	recordColonies(g.colHist, &g.snap, int64(g.ticks.Ticks))
}

// maxBackDays is how many days the ← key goes back.
const maxBackDays = 365

// shown is the snapshot on screen: an earlier day while viewing, else the latest.
func (g *Game) shown() *snapshot {
	if g.viewing {
		if i := int(g.viewTick - g.days[0].Tick); i >= 0 && i < len(g.days) {
			return &g.days[i]
		}
	}
	return &g.snap
}

// viewBack shows the day before the shown one, as long as it is kept.
func (g *Game) viewBack() {
	t := g.shown().Tick
	if len(g.days) == 0 || t <= g.days[0].Tick {
		return
	}
	g.viewing, g.viewTick = true, t-1
}

// viewForward shows the next kept day. It reports false if the latest day is shown.
func (g *Game) viewForward() bool {
	if !g.viewing {
		return false
	}
	g.viewTick++
	if g.viewTick >= g.snap.Tick {
		g.viewing = false
	}
	return true
}

func (g *Game) Update() error {
	if g.frames >= 3 {
		if g.shotErr != nil {
			return g.shotErr
		}
		return errScreenshotDone
	}
	g.handleInput()

	if !g.paused && !g.finished {
		g.advance()
	}
	if g.finished {
		g.paused = true
	}
	return nil
}

// advance simulates the days that are due in this frame at the selected speed.
func (g *Game) advance() {
	g.pending += speeds[g.speed] / float64(ebiten.TPS())
	for g.pending >= 1 && !g.finished {
		if !g.viewForward() { // while an earlier day is shown, go through the kept days first
			g.step()
		}
		g.pending--
	}
}

// retreat steps back through the kept days at the selected speed (← held down).
func (g *Game) retreat() {
	g.backPending += speeds[g.speed] / float64(ebiten.TPS())
	for g.backPending >= 1 {
		g.viewBack()
		g.backPending--
	}
}

func (g *Game) handleInput() {
	g.handleKeys()
	g.handleMouse()
}

func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.togglePause()
	}
	// → simulates one day. Held down, it runs at the selected speed after a short delay, until it is released.
	switch d := inpututil.KeyPressDuration(ebiten.KeyRight); {
	case d == 1:
		g.stepOnce()
	case d > 15:
		g.paused = true
		g.advance()
	}
	// ← shows the day before. Held down, it goes back at the selected speed.
	switch d := inpututil.KeyPressDuration(ebiten.KeyLeft); {
	case d == 1:
		g.paused = true
		g.viewBack()
	case d > 15:
		g.paused = true
		g.retreat()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.faster()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.slower()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.selected, g.nest, g.selColony = nil, nil, nil
	}
}

func (g *Game) handleMouse() {
	mx, my := ebiten.CursorPosition()
	g.hoverColony = g.colonyAt(float64(mx), float64(my))
	g.hover = nil
	if g.hoverColony == nil {
		g.hover = g.patchAt(float64(mx), float64(my))
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for _, b := range g.buttons {
			if b.contains(mx, my) {
				b.action()
				return
			}
		}
		if image.Pt(mx, my).In(g.layout.mapRect) {
			g.selected, g.selColony = g.hover, nil
			if g.hoverColony != nil {
				g.selColony = g.colHist[g.hoverColony.Who]
			}
		}
	}
	// Right-click places the test nest. A double right-click removes it.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && image.Pt(mx, my).In(g.layout.mapRect) {
		now := time.Now()
		if g.nest != nil && now.Sub(g.lastRightClick) < doubleClick {
			g.nest = nil
		} else {
			x, y := g.layout.fromScreen(float64(mx), float64(my))
			g.nest = &[2]float64{x, y}
		}
		g.lastRightClick = now
	}
}

func (g *Game) togglePause() {
	if !g.finished {
		g.paused = !g.paused
	}
}

func (g *Game) stepOnce() {
	g.paused = true
	if !g.viewForward() {
		g.step()
	}
}

func (g *Game) faster() { g.speed = min(g.speed+1, len(speeds)-1) }
func (g *Game) slower() { g.speed = max(g.speed-1, 0) }

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width, g.height = outsideWidth, outsideHeight
	g.layout = newLayout(outsideWidth, outsideHeight, g.world)
	return outsideWidth, outsideHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colBackground)
	g.drawMap(screen)
	g.drawPanel(screen)
	if g.cfg.Screenshot != "" {
		g.frames++
		if g.frames == 3 { // the first frames may not have the final window size
			g.shotErr = saveScreenshot(screen, g.cfg.Screenshot)
		}
	}
}

// errScreenshotDone ends the game loop after a screenshot.
var errScreenshotDone = errors.New("screenshot saved")

func saveScreenshot(screen *ebiten.Image, path string) error {
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
