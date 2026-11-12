//go:build ui || windows || (darwin && cgo)

package ui

import (
	"bytes"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	colBackground = color.RGBA{0x1d, 0x20, 0x24, 0xff}
	colPanel      = color.RGBA{0x26, 0x2a, 0x30, 0xff}
	colBorder     = color.RGBA{0x3a, 0x40, 0x48, 0xff}
	colText       = color.RGBA{0xe6, 0xe8, 0xeb, 0xff}
	colMuted      = color.RGBA{0x9a, 0xa2, 0xad, 0xff}
	colAccent     = color.RGBA{0xf5, 0xb7, 0x2a, 0xff} // flowers, current day
	colSecond     = color.RGBA{0x6c, 0xb4, 0xee, 0xff} // second series
	colBee        = color.RGBA{0xe8, 0x7a, 0x5d, 0xff} // bee species
	colError      = color.RGBA{0xff, 0x6b, 0x6b, 0xff}
	colBadger     = color.RGBA{0x30, 0x30, 0x30, 0xff}
	colCuckoo     = color.RGBA{0xe0, 0x50, 0xe0, 0xff} // cuckoo bumblebees (Psithyrus)
	colBrood      = color.RGBA{0xf0, 0xe6, 0xc8, 0xff} // eggs
	colLarva      = color.RGBA{0xb8, 0xd8, 0x8a, 0xff}
	colPupa       = color.RGBA{0xa8, 0x8a, 0xd8, 0xff}
	colButton     = color.RGBA{0x34, 0x3a, 0x42, 0xff}
	colButtonHi   = color.RGBA{0x44, 0x4c, 0x56, 0xff}
)

var (
	faceSmall  = newFace(goregular.TTF, 13)
	faceNormal = newFace(goregular.TTF, 15)
	faceBold   = newFace(gobold.TTF, 15)
	faceTitle  = newFace(gobold.TTF, 22)
)

func newFace(ttf []byte, size float64) *text.GoTextFace {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		panic(err)
	}
	return &text.GoTextFace{Source: src, Size: size}
}

// drawText draws s with its top left corner at (x, y) and returns the line height.
func drawText(dst *ebiten.Image, s string, face *text.GoTextFace, x, y float64, clr color.Color) float64 {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	op.LineSpacing = face.Size * 1.35
	text.Draw(dst, s, face, op)
	_, h := text.Measure(s, face, op.LineSpacing)
	return h
}

// drawTextRight draws s right-aligned at x.
func drawTextRight(dst *ebiten.Image, s string, face *text.GoTextFace, x, y float64, clr color.Color) {
	w, _ := text.Measure(s, face, 0)
	drawText(dst, s, face, x-w, y, clr)
}

// wrap breaks s into lines no wider than width.
func wrap(s string, face *text.GoTextFace, width float64) string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			try := word
			if line != "" {
				try = line + " " + word
			}
			if w, _ := text.Measure(try, face, 0); w > width && line != "" {
				out = append(out, line)
				line = word
			} else {
				line = try
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func fillRect(dst *ebiten.Image, r image.Rectangle, clr color.Color) {
	vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), clr, false)
}

func strokeRect(dst *ebiten.Image, r image.Rectangle, clr color.Color) {
	vector.StrokeRect(dst, float32(r.Min.X)+0.5, float32(r.Min.Y)+0.5, float32(r.Dx())-1, float32(r.Dy())-1, 1, clr, false)
}

func withAlpha(c color.RGBA, a float64) color.RGBA {
	// Premultiplied alpha, as image/color expects.
	return color.RGBA{uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(255 * a)}
}

const (
	panelWidth = 520
	pad        = 16
)

// bounds is the part of the landscape the map shows [m].
type bounds struct {
	MinX, MaxX, MinY, MaxY float64
}

// layout holds the screen regions and the transformation from model coordinates [m] to the screen.
type layout struct {
	mapRect   image.Rectangle
	panelRect image.Rectangle
	scale     float64 // screen pixels per metre
	bounds    bounds
}

func newLayout(w, h int, b bounds) layout {
	l := layout{bounds: b}
	worldW := b.MaxX - b.MinX
	worldH := b.MaxY - b.MinY
	availW := float64(w - panelWidth - 3*pad)
	availH := float64(h - 2*pad)
	l.scale = max(min(availW/worldW, availH/worldH), 0.001)
	mw, mh := int(worldW*l.scale), int(worldH*l.scale)
	x0 := pad + (int(availW)-mw)/2
	y0 := pad + (int(availH)-mh)/2
	l.mapRect = image.Rect(x0, y0, x0+mw, y0+mh)
	l.panelRect = image.Rect(w-panelWidth-pad, pad, w-pad, h-pad)
	return l
}

// toScreen converts model coordinates [m] (y upwards) to screen pixels.
func (l *layout) toScreen(x, y float64) (float64, float64) {
	sx := float64(l.mapRect.Min.X) + (x-l.bounds.MinX)*l.scale
	sy := float64(l.mapRect.Min.Y) + (l.bounds.MaxY-y)*l.scale
	return sx, sy
}

// fromScreen converts screen pixels to model coordinates [m].
func (l *layout) fromScreen(sx, sy float64) (float64, float64) {
	x := (sx-float64(l.mapRect.Min.X))/l.scale + l.bounds.MinX
	y := l.bounds.MaxY - (sy-float64(l.mapRect.Min.Y))/l.scale
	return x, y
}

// button is a clickable rectangle in the panel.
type button struct {
	rect   image.Rectangle
	label  string
	action func()
}

func (b *button) contains(x, y int) bool { return image.Pt(x, y).In(b.rect) }

func textMeasure(s string) (float64, float64) {
	return text.Measure(s, faceNormal, 0)
}

func measureFace(s string, face *text.GoTextFace) (float64, float64) {
	return text.Measure(s, face, 0)
}

// truncate shortens s with "…" to at most width pixels in the small font.
func truncate(s string, width float64) string {
	if w, _ := measureFace(s, faceSmall); w <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 {
		r = r[:len(r)-1]
		if w, _ := measureFace(string(r)+"…", faceSmall); w <= width {
			return string(r) + "…"
		}
	}
	return ""
}
