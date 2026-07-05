// Package main implements demo15 — LineChart Showcase.
//
// A print-based demo that renders sample charts: multi-series line chart,
// sine wave, stock simulation, and a simple bar-style chart.
//
// Usage: go run ./cmd/demo15/
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 15 — LineChart Showcase")
		fmt.Println("Usage: go run ./cmd/demo15/")
		return
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║          Fluui Demo 15 — LineChart Showcase                         ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── 1. Sine + Cosine waves ──
	demoSinCos()

	// ── 2. Multi-series with markers ──
	demoMultiSeries()

	// ── 3. Stock-like data ──
	demoStockData()

	// ── 4. No axes / minimal ──
	demoMinimal()

	fmt.Println("  All charts rendered with LineChart component (component/linechart.go)")
}

func demoSinCos() {
	fmt.Println("  ── Sine & Cosine Waves ──")
	fmt.Println()

	lc := component.NewLineChart()
	lc.SetTitle("Sine vs Cosine")
	lc.SetYAxis(component.ChartAxisConfig{Min: -1, Max: 1, LabelCount: 5})
	lc.SetXAxis(component.ChartAxisConfig{Min: 0, Max: 6.28, LabelCount: 7})

	var sineData, cosData []component.ChartPoint
	for i := 0; i <= 30; i++ {
		x := float64(i) * 6.28 / 30
		sineData = append(sineData, component.ChartPoint{X: x, Y: math.Sin(x)})
		cosData = append(cosData, component.ChartPoint{X: x, Y: math.Cos(x)})
	}

	lc.AddSeries(component.ChartSeries{
		Name:   "sin(x)",
		Data:   sineData,
		Color:  buffer.NamedColor(buffer.NamedCyan),
		Marker: component.ChartMarkerDot,
	})
	lc.AddSeries(component.ChartSeries{
		Name:   "cos(x)",
		Data:   cosData,
		Color:  buffer.NamedColor(buffer.NamedYellow),
		Marker: component.ChartMarkerPlus,
	})

	renderChart(lc, 70, 15)
	fmt.Println()
}

func demoMultiSeries() {
	fmt.Println("  ── Multi-Series with Trend Lines ──")
	fmt.Println()

	lc := component.NewLineChart()
	lc.SetTitle("Revenue Projections 2024")

	lc.AddSeries(component.ChartSeries{
		Name:  "Optimistic",
		Data: []component.ChartPoint{{X: 0, Y: 10}, {X: 1, Y: 25}, {X: 2, Y: 45}, {X: 3, Y: 60}, {X: 4, Y: 80}, {X: 5, Y: 95}},
		Color: buffer.NamedColor(buffer.NamedGreen),
	})
	lc.AddSeries(component.ChartSeries{
		Name:  "Realistic",
		Data: []component.ChartPoint{{X: 0, Y: 10}, {X: 1, Y: 20}, {X: 2, Y: 30}, {X: 3, Y: 40}, {X: 4, Y: 50}, {X: 5, Y: 55}},
		Color: buffer.NamedColor(buffer.NamedYellow),
	})
	lc.AddSeries(component.ChartSeries{
		Name:  "Pessimistic",
		Data: []component.ChartPoint{{X: 0, Y: 10}, {X: 1, Y: 12}, {X: 2, Y: 15}, {X: 3, Y: 18}, {X: 4, Y: 20}, {X: 5, Y: 22}},
		Color: buffer.NamedColor(buffer.NamedRed),
	})

	renderChart(lc, 70, 14)
	fmt.Println()
}

func demoStockData() {
	fmt.Println("  ── Stock Price Simulation ──")
	fmt.Println()

	lc := component.NewLineChart()
	lc.SetTitle("FLUU Stock — Daily Close")
	lc.SetShowGrid(true)

	// Simulate a week of stock prices
	prices := []float64{100, 102, 99, 105, 108, 104, 110, 115, 112, 118, 120, 116, 122, 125}
	data := make([]component.ChartPoint, len(prices))
	for i, p := range prices {
		data[i] = component.ChartPoint{X: float64(i), Y: p}
	}
	lc.AddSeries(component.ChartSeries{
		Name:   "Close",
		Data:   data,
		Color:  buffer.NamedColor(buffer.NamedBrightGreen),
		Marker: component.ChartMarkerStar,
	})

	renderChart(lc, 70, 14)
	fmt.Println()
}

func demoMinimal() {
	fmt.Println("  ── Minimal (No Axes, No Grid) ──")
	fmt.Println()

	lc := component.NewLineChart()
	lc.SetTitle("Pure Data")
	lc.SetShowAxes(false)
	lc.SetShowGrid(false)
	lc.SetShowLegend(false)

	var data []component.ChartPoint
	for i := 0; i <= 20; i++ {
		x := float64(i)
		y := math.Exp(float64(i)/10) * math.Sin(float64(i)/3)
		data = append(data, component.ChartPoint{X: x, Y: y})
	}
	lc.AddSeries(component.ChartSeries{
		Data:  data,
		Color: buffer.NamedColor(buffer.NamedMagenta),
	})

	renderChart(lc, 70, 10)
	fmt.Println()
}

func renderChart(lc *component.LineChart, w, h int) {
	lc.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
	buf := buffer.NewBuffer(w, h)
	lc.Paint(buf)

	// Render buffer as text with ANSI colors
	fmt.Print("  ")
	for y := 0; y < h; y++ {
		if y > 0 {
			fmt.Print("  ")
		}
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			// Apply ANSI color if available
			if cell.Fg.Type == buffer.ColorNamed && cell.Fg.Val > 0 {
				fmt.Printf("\033[3%dm%c\033[0m", cell.Fg.Val, r)
			} else {
				fmt.Printf("%c", r)
			}
		}
		fmt.Println()
	}
}
