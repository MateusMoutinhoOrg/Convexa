package erpview

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
)

// A chart is drawn by its template as inline SVG: the view computes every
// coordinate here — the template only places the shapes — so no markup is
// ever built as a string, and every value stays escaped.

// ChartLine is one series of a line chart, as the points attribute of a
// polyline.
type ChartLine struct {
	Class  string
	Points string
}

// ChartBar is one bar of a bar chart.
type ChartBar struct {
	Class string
	X     int64
	Y     int64
	W     int64
	H     int64
	// Title is shown when the pointer rests on the bar.
	Title string
}

// ChartTick is one labeled line of an axis.
type ChartTick struct {
	X     int64
	Y     int64
	Label string
	// Anchor is the text-anchor of the label: the first and the last label
	// of the x axis lean inward, so neither is cut at the edge.
	Anchor string
}

// Chart is what a chart template draws.
type Chart struct {
	Width  int64
	Height int64
	// ViewBox is the SVG viewBox.
	ViewBox string
	Lines   []ChartLine
	Bars    []ChartBar
	YTicks  []ChartTick
	XTicks  []ChartTick
	// Zero is the y of the value 0, drawn when it is inside the chart.
	Zero     int64
	HasZero  bool
	Empty    bool
	PlotLeft int64
	PlotTop  int64
	PlotW    int64
	PlotH    int64
}

// The margins around the plot, in SVG units.
const (
	chartLeft   = 74
	chartRight  = 12
	chartTop    = 12
	chartBottom = 26
)

// scale is the range of values a chart spans: from min to max, widened to
// hold 0 and to never be empty.
func scale(sandbox *api.Sandbox, values []int64) (int64, int64) {
	low, high := int64(0), int64(0)
	for _, value := range values {
		if value < low {
			low = value
		}
		if value > high {
			high = value
		}
	}
	if low == high {
		high = low + 100
	}
	span := high - low
	return low - span/20, high + span/20
}

// newChart is the empty Chart of width by height spanning low to high,
// with its value ticks labeled by label.
func newChart(sandbox *api.Sandbox, width int64, height int64, low int64, high int64, label func(int64) string) Chart {
	chart := Chart{Width: width, Height: height, PlotLeft: chartLeft, PlotTop: chartTop, PlotW: width - chartLeft - chartRight, PlotH: height - chartTop - chartBottom}
	chart.ViewBox = sandbox.Deps.StdDeps.Sprintf("0 0 %d %d", width, height)
	for step := int64(0); step <= 4; step++ {
		value := low + (high-low)*step/4
		chart.YTicks = append(chart.YTicks, ChartTick{Y: yOf(sandbox, chart, low, high, value), Label: label(value)})
	}
	if low < 0 && high > 0 {
		chart.Zero, chart.HasZero = yOf(sandbox, chart, low, high, 0), true
	}
	return chart
}

// yOf is the y of value on chart, which spans low to high.
func yOf(sandbox *api.Sandbox, chart Chart, low int64, high int64, value int64) int64 {
	if high == low {
		return chart.PlotTop + chart.PlotH/2
	}
	return chart.PlotTop + chart.PlotH - (value-low)*chart.PlotH/(high-low)
}

// LineChart is a chart of width by height drawing each series of values —
// one value per x, every series as long — classed by classes, its values
// labeled by label, the x of index i labeled xLabel(i) every so often.
func LineChart(sandbox *api.Sandbox, width int64, height int64, series [][]int64, classes []string, label func(int64) string, xLabel func(int) string) Chart {
	all := []int64{}
	count := 0
	for _, values := range series {
		all = append(all, values...)
		if len(values) > count {
			count = len(values)
		}
	}
	low, high := scale(sandbox, all)
	chart := newChart(sandbox, width, height, low, high, label)
	if count == 0 {
		chart.Empty = true
		return chart
	}
	xOf := func(i int) int64 {
		if count == 1 {
			return chart.PlotLeft + chart.PlotW/2
		}
		return chart.PlotLeft + int64(i)*chart.PlotW/int64(count-1)
	}
	for index, values := range series {
		points := []string{}
		for i, value := range values {
			points = append(points, sandbox.Deps.StdDeps.Sprintf("%d,%d", xOf(i), yOf(sandbox, chart, low, high, value)))
		}
		class := ""
		if index < len(classes) {
			class = classes[index]
		}
		chart.Lines = append(chart.Lines, ChartLine{Class: class, Points: sandbox.Deps.StringsDeps.Join(points, " ")})
	}
	ticks := 6
	if count < ticks {
		ticks = count
	}
	for t := 0; t < ticks; t++ {
		i := 0
		if ticks > 1 {
			i = t * (count - 1) / (ticks - 1)
		}
		anchor := "middle"
		if ticks > 1 && t == 0 {
			anchor = "start"
		} else if ticks > 1 && t == ticks-1 {
			anchor = "end"
		}
		chart.XTicks = append(chart.XTicks, ChartTick{X: xOf(i), Y: chart.PlotTop + chart.PlotH + 18, Label: xLabel(i), Anchor: anchor})
	}
	return chart
}

// BarChart is a chart of width by height drawing, for each group, one bar
// per series side by side — groups[g][s] the value of series s in group g —
// classed by classes, titled by title, labeled by label and xLabel.
func BarChart(sandbox *api.Sandbox, width int64, height int64, groups [][]int64, classes []string, label func(int64) string, xLabel func(int) string, title func(int, int) string) Chart {
	all := []int64{}
	for _, group := range groups {
		all = append(all, group...)
	}
	low, high := scale(sandbox, all)
	chart := newChart(sandbox, width, height, low, high, label)
	if len(groups) == 0 {
		chart.Empty = true
		return chart
	}
	slot := chart.PlotW / int64(len(groups))
	zero := yOf(sandbox, chart, low, high, 0)
	for g, group := range groups {
		if len(group) == 0 {
			continue
		}
		bar := (slot * 7 / 10) / int64(len(group))
		if bar < 1 {
			bar = 1
		}
		left := chart.PlotLeft + int64(g)*slot + (slot-bar*int64(len(group)))/2
		for s, value := range group {
			y := yOf(sandbox, chart, low, high, value)
			top, h := y, zero-y
			if value < 0 {
				top, h = zero, y-zero
			}
			class := ""
			if s < len(classes) {
				class = classes[s]
			}
			chart.Bars = append(chart.Bars, ChartBar{Class: class, X: left + int64(s)*bar, Y: top, W: bar - 1, H: h, Title: title(g, s)})
		}
		every := len(groups)/8 + 1
		if g%every == 0 {
			chart.XTicks = append(chart.XTicks, ChartTick{X: chart.PlotLeft + int64(g)*slot + slot/2, Y: chart.PlotTop + chart.PlotH + 18, Label: xLabel(g), Anchor: "middle"})
		}
	}
	return chart
}

// CompactMoney is cents spelled short for an axis: 1.2k, 3.4M.
func CompactMoney(sandbox *api.Sandbox, cents int64) string {
	units := cents / 100
	sign := ""
	if units < 0 {
		sign, units = "-", -units
	}
	switch {
	case units >= 1_000_000:
		return sandbox.Deps.StdDeps.Sprintf("%s%d.%dM", sign, units/1_000_000, (units%1_000_000)/100_000)
	case units >= 1_000:
		return sandbox.Deps.StdDeps.Sprintf("%s%d.%dk", sign, units/1_000, (units%1_000)/100)
	}
	return sandbox.Deps.StdDeps.Sprintf("%s%d", sign, units)
}
