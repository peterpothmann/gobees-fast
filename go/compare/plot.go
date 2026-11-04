package compare

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

// Colours of the plots: blue and red of the categorical reference palette of the dataviz guidelines, light mode,
// and its secondary text grey, on an explicitly light surface (GitHub shows a PNG as it is, also in dark mode).
var (
	colourNetLogo     = color.NRGBA{R: 0x2a, G: 0x78, B: 0xd6, A: 0xff}
	colourNetLogoBand = color.NRGBA{R: 0x2a, G: 0x78, B: 0xd6, A: 0x38} // the band as a light wash
	// gobees: thick and 50 % opaque, so that the NetLogo median below it stays visible
	colourGobees = color.NRGBA{R: 0xe3, G: 0x49, B: 0x48, A: 0x80}
	// gobees-fast: thin, dashed and 60 % opaque, so that gobees below it stays visible in the gaps and at the edges
	colourGobeesFast = color.NRGBA{R: 0x52, G: 0x51, B: 0x4e, A: 0x99}
	colourText       = color.NRGBA{R: 0x0b, G: 0x0b, B: 0x0b, A: 0xff}
	colourMuted      = color.NRGBA{R: 0x52, G: 0x51, B: 0x4e, A: 0xff}
	colourGrid       = color.NRGBA{R: 0xe4, G: 0xe3, B: 0xdf, A: 0xff}
	colourSurface    = color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfb, A: 0xff}
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

// Lines of the medians [pt].
var (
	widthNetLogo     = vg.Points(1.2)
	widthGobees      = vg.Points(2.8)
	widthGobeesFast  = vg.Points(1.6)
	dashesGobeesFast = []vg.Length{vg.Points(5), vg.Points(3)}
)

// plotCaption explains the plots below the image in the report.
const plotCaption = "Plots: blue, NetLogo median and 5-95 % band; red, gobees median; grey dashed, gobees-fast " +
	"median; x axis: day."

// writePlot draws one panel per metric that has at least 2 days (Medians, Bands) into a PNG file: the NetLogo band,
// the NetLogo median, the gobees median (if the metric has BaseMedians) and the gobees-fast median over the days.
// The first tile holds the legend. It returns false and writes nothing if no row has 2 days.
func writePlot(path string, rows []NetLogoMetric) (bool, error) {
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
func metricPlot(r NetLogoMetric) (*plot.Plot, error) {
	p := newPlot()
	p.Title.Text = r.Metric

	g := plotter.NewGrid()
	g.Vertical.Color, g.Horizontal.Color = colourGrid, colourGrid
	g.Vertical.Width, g.Horizontal.Width = vg.Points(0.5), vg.Points(0.5)
	p.Add(g)

	n := len(r.Medians)
	band := make(plotter.XYs, 0, 2*n)
	netlogo := make(plotter.XYs, n)
	gobeesFast := make(plotter.XYs, n)
	for d := range n {
		band = append(band, plotter.XY{X: float64(d + 1), Y: r.Bands[d].Q95})
		netlogo[d] = plotter.XY{X: float64(d + 1), Y: r.Bands[d].Q50}
		gobeesFast[d] = plotter.XY{X: float64(d + 1), Y: r.Medians[d]}
	}
	for d := n - 1; d >= 0; d-- {
		band = append(band, plotter.XY{X: float64(d + 1), Y: r.Bands[d].Q05})
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
	nl.Color, nl.Width = colourNetLogo, widthNetLogo
	p.Add(poly, nl)
	if len(r.BaseMedians) == n {
		gobees := make(plotter.XYs, n)
		for d, median := range r.BaseMedians {
			gobees[d] = plotter.XY{X: float64(d + 1), Y: median}
		}
		gb, err := plotter.NewLine(gobees)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.Metric, err)
		}
		gb.Color, gb.Width = colourGobees, widthGobees
		p.Add(gb)
	}
	gf, err := plotter.NewLine(gobeesFast)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", r.Metric, err)
	}
	gf.Color, gf.Width, gf.Dashes = colourGobeesFast, widthGobeesFast, dashesGobeesFast
	p.Add(gf) // gobees-fast on top: the version under test
	p.X.Min, p.X.Max = 1, float64(n)
	return p, nil
}

// legendPlot returns the first tile: no axes, only the legend of the four marks.
func legendPlot() *plot.Plot {
	p := newPlot()
	p.HideAxes()
	band, _ := plotter.NewPolygon(plotter.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}})
	band.Color = colourNetLogoBand
	band.LineStyle.Width = 0
	nl := &plotter.Line{LineStyle: draw.LineStyle{Color: colourNetLogo, Width: vg.Points(2)}}
	gb := &plotter.Line{LineStyle: draw.LineStyle{Color: colourGobees, Width: widthGobees}}
	gf := &plotter.Line{LineStyle: draw.LineStyle{Color: colourGobeesFast, Width: widthGobeesFast,
		Dashes: dashesGobeesFast}}
	p.Legend.Add("NetLogo 5-95 % band (100 runs)", band)
	p.Legend.Add("NetLogo median", nl)
	p.Legend.Add("gobees median (100 runs)", gb)
	p.Legend.Add("gobees-fast median (100 runs)", gf)
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
