package validation

import (
	"fmt"
	"image/color"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// Colours of the plots: slots 1 (blue), 2 (orange) and 3 (dark grey) of the categorical reference palette of the
// dataviz guidelines, light mode, on an explicitly light surface (GitHub shows a PNG as it is, also in dark mode).
var (
	colourNetLogo     = color.NRGBA{R: 0x2a, G: 0x78, B: 0xd6, A: 0xff}
	colourNetLogoBand = color.NRGBA{R: 0x2a, G: 0x78, B: 0xd6, A: 0x38} // the band as a light wash
	colourGobees      = color.NRGBA{R: 0xeb, G: 0x68, B: 0x34, A: 0xff} // gobees-fast
	colourBase        = color.NRGBA{R: 0x3d, G: 0x3c, B: 0x38, A: 0xff} // gobees
	colourText        = color.NRGBA{R: 0x0b, G: 0x0b, B: 0x0b, A: 0xff}
	colourMuted       = color.NRGBA{R: 0x52, G: 0x51, B: 0x4e, A: 0xff}
	colourGrid        = color.NRGBA{R: 0xe4, G: 0xe3, B: 0xdf, A: 0xff}
	colourSurface     = color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfb, A: 0xff}
)

const (
	plotCols   = 3
	plotDPI    = 500           // twice the screen resolution, so that zooming in still shows details
	tileWidth  = 4.8 * vg.Inch // about 920 px at 192 dpi
	tileHeight = 3.0 * vg.Inch // about 580 px at 192 dpi

	fontTitle  = 13 // [pt] title of a panel
	fontTicks  = 11 // [pt] numbers at the axes
	fontLegend = 13 // [pt]
)

// plotCaption explains the plots below the image in the report.
const plotCaption = "Plots: blue, NetLogo median and 5-95 % band; dark grey, dashed, gobees median; orange, gobees-fast " +
	"median; x axis: day."

// writePlot draws one panel per row that has at least 2 days (Medians, Bands) into a PNG file: the NetLogo band,
// the NetLogo median and the gobees median over the days. The first tile holds the legend. It returns false and
// writes nothing if no row has 2 days.
func writePlot(path string, rows []ReportRow) (bool, error) {
	plots := []*plot.Plot{legendPlot()}
	for _, r := range rows {
		if len(r.Medians) < 2 {
			continue
		}
		p, err := metricPlot(r)
		if err != nil {
			return false, err
		}
		plots = append(plots, p)
	}
	if len(plots) == 1 {
		return false, nil
	}

	nRows := (len(plots) + plotCols - 1) / plotCols
	grid := make([][]*plot.Plot, nRows)
	for i := range grid {
		grid[i] = make([]*plot.Plot, plotCols)
	}
	for i, p := range plots {
		grid[i/plotCols][i%plotCols] = p
	}

	img := vgimg.NewWith(vgimg.UseWH(plotCols*tileWidth, vg.Length(nRows)*tileHeight), vgimg.UseDPI(plotDPI),
		vgimg.UseBackgroundColor(colourSurface))
	dc := draw.New(img)
	tiles := draw.Tiles{Rows: nRows, Cols: plotCols, PadTop: 6, PadBottom: 6, PadLeft: 6, PadRight: 12, PadX: 18, PadY: 12}
	canvases := plot.Align(grid, tiles, dc)
	for j := range grid {
		for i, p := range grid[j] {
			if p != nil {
				p.Draw(canvases[j][i])
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	if _, err := (vgimg.PngCanvas{Canvas: img}).WriteTo(f); err != nil {
		return false, err
	}
	return true, f.Close()
}

// newPlot returns a plot with the common style: light surface, small muted axis text, recessive grid.
func newPlot() *plot.Plot {
	p := plot.New()
	p.BackgroundColor = colourSurface
	p.Title.TextStyle.Color = colourText
	p.Title.TextStyle.Font.Size = vg.Points(fontTitle)
	for _, a := range []*plot.Axis{&p.X, &p.Y} {
		a.Color = colourMuted
		a.Tick.Color = colourMuted
		a.Tick.Label.Color = colourMuted
		a.Tick.Label.Font.Size = vg.Points(fontTicks)
		a.Padding = vg.Points(2)
		a.Tick.Marker = compactTicks{}
	}
	return p
}

// metricPlot returns the panel of one metric.
func metricPlot(r ReportRow) (*plot.Plot, error) {
	p := newPlot()
	p.Title.Text = r.Metric

	g := plotter.NewGrid()
	g.Vertical.Color, g.Horizontal.Color = colourGrid, colourGrid
	g.Vertical.Width, g.Horizontal.Width = vg.Points(0.5), vg.Points(0.5)
	p.Add(g)

	n := len(r.Medians)
	band := make(plotter.XYs, 0, 2*n)
	netlogo := make(plotter.XYs, n)
	gobees := make(plotter.XYs, n)
	for d := range n {
		band = append(band, plotter.XY{X: float64(d), Y: r.Bands[d].Q95})
		netlogo[d] = plotter.XY{X: float64(d), Y: r.Bands[d].Q50}
		gobees[d] = plotter.XY{X: float64(d), Y: r.Medians[d]}
	}
	for d := n - 1; d >= 0; d-- {
		band = append(band, plotter.XY{X: float64(d), Y: r.Bands[d].Q05})
	}

	poly, err := plotter.NewPolygon(band)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", r.Metric, err)
	}
	poly.Color = colourNetLogoBand
	poly.LineStyle.Width = 0
	nl, err := plotter.NewLine(netlogo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", r.Metric, err)
	}
	nl.Color, nl.Width = colourNetLogo, vg.Points(1.2)
	gb, err := plotter.NewLine(gobees)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", r.Metric, err)
	}
	gb.Color, gb.Width = colourGobees, vg.Points(1.2)
	p.Add(poly, nl)
	if len(r.BaseMedians) == n {
		base := make(plotter.XYs, n)
		for d := range n {
			base[d] = plotter.XY{X: float64(d), Y: r.BaseMedians[d]}
		}
		bl, err := plotter.NewLine(base)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.Metric, err)
		}
		bl.Color, bl.Width, bl.Dashes = colourBase, vg.Points(1.2), []vg.Length{vg.Points(4), vg.Points(3)}
		p.Add(bl)
	}
	p.Add(gb) // on top: gobees-fast is the version under test
	p.X.Min, p.X.Max = 0, float64(n-1)
	return p, nil
}

// legendPlot returns the first tile: no axes, only the legend of the three marks.
func legendPlot() *plot.Plot {
	p := newPlot()
	p.HideAxes()
	band, _ := plotter.NewPolygon(plotter.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}})
	band.Color = colourNetLogoBand
	band.LineStyle.Width = 0
	nl := &plotter.Line{LineStyle: draw.LineStyle{Color: colourNetLogo, Width: vg.Points(2)}}
	gb := &plotter.Line{LineStyle: draw.LineStyle{Color: colourGobees, Width: vg.Points(2)}}
	base := &plotter.Line{LineStyle: draw.LineStyle{Color: colourBase, Width: vg.Points(2),
		Dashes: []vg.Length{vg.Points(4), vg.Points(3)}}}
	p.Legend.Add("NetLogo 5-95 % band (100 runs)", band)
	p.Legend.Add("NetLogo median", nl)
	p.Legend.Add("gobees median (100 runs)", base)
	p.Legend.Add("gobees-fast median (100 runs)", gb)
	p.Legend.Top, p.Legend.Left = true, true
	p.Legend.TextStyle.Color = colourText
	p.Legend.TextStyle.Font.Size = vg.Points(fontLegend)
	p.Legend.ThumbnailWidth = vg.Points(30)
	p.Legend.Padding = vg.Points(6)
	p.Legend.YOffs = -vg.Points(12)
	return p
}

// compactTicks are the default ticks of gonum/plot with short labels: 500 instead of 500.00, 0.015 instead of
// 0.0150.
type compactTicks struct{}

func (compactTicks) Ticks(min, max float64) []plot.Tick {
	ticks := plot.DefaultTicks{}.Ticks(min, max)
	for i := range ticks {
		if ticks[i].Label != "" {
			ticks[i].Label = strconv.FormatFloat(ticks[i].Value, 'g', 6, 64)
		}
	}
	return ticks
}
