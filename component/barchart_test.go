package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func newBarChartForTest() *BarChart {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	return bc
}

// --- Constructor ---

func TestBarChart_New(t *testing.T) {
	bc := NewBarChart()
	if bc == nil {
		t.Fatal("NewBarChart returned nil")
	}
	if bc.Orientation() != BarVertical {
		t.Errorf("default orientation = %v, want BarVertical", bc.Orientation())
	}
	if !bc.showGrid {
		t.Error("default showGrid should be true")
	}
	if !bc.showAxes {
		t.Error("default showAxes should be true")
	}
	if !bc.showLegend {
		t.Error("default showLegend should be true")
	}
	if !bc.showTitle {
		t.Error("default showTitle should be true")
	}
	if bc.gap != 1 {
		t.Errorf("default gap = %d, want 1", bc.gap)
	}
}

// --- Configuration ---

func TestBarChart_SetTitle(t *testing.T) {
	bc := NewBarChart()
	bc.SetTitle("Revenue Q4")
	if bc.Title() != "Revenue Q4" {
		t.Errorf("Title() = %q, want %q", bc.Title(), "Revenue Q4")
	}
}

func TestBarChart_SetOrientation(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	if bc.Orientation() != BarHorizontal {
		t.Errorf("Orientation() = %v, want BarHorizontal", bc.Orientation())
	}
}

func TestBarChart_AddSeries(t *testing.T) {
	bc := NewBarChart()
	bc.AddSeries(BarSeries{
		Name:  "Sales",
		Data:  []BarData{{Label: "A", Value: 10}, {Label: "B", Value: 20}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	if bc.SeriesCount() != 1 {
		t.Fatalf("SeriesCount = %d, want 1", bc.SeriesCount())
	}
	s := bc.Series(0)
	if s.Name != "Sales" {
		t.Errorf("Series(0).Name = %q, want %q", s.Name, "Sales")
	}
	if len(s.Data) != 2 {
		t.Errorf("len(Data) = %d, want 2", len(s.Data))
	}
}

func TestBarChart_SetSeries(t *testing.T) {
	bc := NewBarChart()
	series := []BarSeries{
		{Name: "A", Data: []BarData{{Label: "x", Value: 1}}, Color: buffer.NamedColor(buffer.NamedRed)},
		{Name: "B", Data: []BarData{{Label: "y", Value: 2}}, Color: buffer.NamedColor(buffer.NamedGreen)},
	}
	bc.SetSeries(series)
	if bc.SeriesCount() != 2 {
		t.Fatalf("SeriesCount = %d, want 2", bc.SeriesCount())
	}
}

func TestBarChart_ClearSeries(t *testing.T) {
	bc := NewBarChart()
	bc.AddSeries(BarSeries{Name: "S", Data: []BarData{{Label: "x", Value: 1}}, Color: buffer.NamedColor(buffer.NamedRed)})
	bc.ClearSeries()
	if bc.SeriesCount() != 0 {
		t.Errorf("after Clear, SeriesCount = %d, want 0", bc.SeriesCount())
	}
}

func TestBarChart_SeriesOutOfRange(t *testing.T) {
	bc := NewBarChart()
	s := bc.Series(5)
	if s.Name != "" {
		t.Errorf("Series(5) on empty should return zero value, got Name=%q", s.Name)
	}
	s = bc.Series(-1)
	if s.Name != "" {
		t.Errorf("Series(-1) should return zero value, got Name=%q", s.Name)
	}
}

func TestBarChart_SetShowGrid(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowGrid(false)
	if bc.showGrid {
		t.Error("showGrid should be false")
	}
}

func TestBarChart_SetShowAxes(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowAxes(false)
	if bc.showAxes {
		t.Error("showAxes should be false")
	}
}

func TestBarChart_SetShowLegend(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowLegend(false)
	if bc.showLegend {
		t.Error("showLegend should be false")
	}
}

func TestBarChart_SetShowTitle(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowTitle(false)
	if bc.showTitle {
		t.Error("showTitle should be false")
	}
}

func TestBarChart_SetShowValues(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowValues(true)
	if !bc.showValues {
		t.Error("showValues should be true")
	}
}

func TestBarChart_SetMaxVal(t *testing.T) {
	bc := NewBarChart()
	bc.SetMaxVal(500)
	if bc.MaxVal() != 500 {
		t.Errorf("MaxVal = %v, want 500", bc.MaxVal())
	}
}

func TestBarChart_SetGap(t *testing.T) {
	bc := NewBarChart()
	bc.SetGap(3)
	if bc.Gap() != 3 {
		t.Errorf("Gap = %d, want 3", bc.Gap())
	}
}

func TestBarChart_SetGapNegative(t *testing.T) {
	bc := NewBarChart()
	bc.SetGap(-5)
	if bc.Gap() != 0 {
		t.Errorf("Gap(-5) = %d, want 0 (clamped)", bc.Gap())
	}
}

func TestBarChart_SetGridStyle(t *testing.T) {
	bc := NewBarChart()
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)}
	bc.SetGridStyle(s)
	if bc.gridStyle.Fg != buffer.NamedColor(buffer.NamedRed) {
		t.Error("gridStyle not set")
	}
}

func TestBarChart_SetAxisStyle(t *testing.T) {
	bc := NewBarChart()
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)}
	bc.SetAxisStyle(s)
	if bc.axisStyle.Fg != buffer.NamedColor(buffer.NamedGreen) {
		t.Error("axisStyle not set")
	}
}

func TestBarChart_SetTitleStyle(t *testing.T) {
	bc := NewBarChart()
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)}
	bc.SetTitleStyle(s)
	if bc.titleStyle.Fg != buffer.NamedColor(buffer.NamedYellow) {
		t.Error("titleStyle not set")
	}
}

// --- Measure ---

func TestBarChart_Measure_Default(t *testing.T) {
	bc := NewBarChart()
	s := bc.Measure(Unbounded())
	if s.W < 10 {
		t.Errorf("Measure W = %d, too small", s.W)
	}
	if s.H < 5 {
		t.Errorf("Measure H = %d, too small", s.H)
	}
}

func TestBarChart_Measure_WithConstraints(t *testing.T) {
	bc := NewBarChart()
	bc.AddSeries(BarSeries{
		Data: []BarData{{Label: "A", Value: 1}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	s := bc.Measure(Bounded(20, 10))
	if s.W > 20 {
		t.Errorf("Measure W = %d, should not exceed 20", s.W)
	}
	if s.H > 10 {
		t.Errorf("Measure H = %d, should not exceed 10", s.H)
	}
}

func TestBarChart_Measure_HorizontalOrientation(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	s := bc.Measure(Unbounded())
	if s.H < 10 {
		t.Errorf("Horizontal H = %d, should be >= 10", s.H)
	}
}

func TestBarChart_Measure_ManyCategories(t *testing.T) {
	bc := NewBarChart()
	data := make([]BarData, 30)
	for i := range data {
		data[i] = BarData{Label: "cat", Value: float64(i)}
	}
	bc.AddSeries(BarSeries{Data: data, Color: buffer.NamedColor(buffer.NamedCyan)})
	s := bc.Measure(Bounded(100, 30))
	if s.W < 40 {
		t.Errorf("W = %d for 30 categories, should be >= 40", s.W)
	}
}

// --- Paint ---

func TestBarChart_Paint_Empty(t *testing.T) {
	bc := newBarChartForTest()
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // should not panic with no series
}

func TestBarChart_Paint_Vertical(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetTitle("Sales")
	bc.AddSeries(BarSeries{
		Name:  "2024",
		Data:  []BarData{{Label: "Q1", Value: 50}, {Label: "Q2", Value: 80}, {Label: "Q3", Value: 30}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Should have some non-space cells
	nonEmpty := 0
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != ' ' && c.Rune != 0 {
				nonEmpty++
			}
		}
	}
	if nonEmpty < 10 {
		t.Errorf("expected non-empty cells, got %d", nonEmpty)
	}
}

func TestBarChart_Paint_Vertical_BarPresent(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetShowTitle(false)
	bc.SetShowLegend(false)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 100}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Look for full block chars
	foundBar := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == '█' {
				foundBar = true
				break
			}
		}
	}
	if !foundBar {
		t.Error("expected to find full block char '█' in painted buffer")
	}
}

func TestBarChart_Paint_Horizontal(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetOrientation(BarHorizontal)
	bc.SetTitle("Revenue")
	bc.AddSeries(BarSeries{
		Name:  "2024",
		Data:  []BarData{{Label: "Q1", Value: 50}, {Label: "Q2", Value: 80}},
		Color: buffer.NamedColor(buffer.NamedGreen),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Check for horizontal bar chars
	foundBar := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == '█' {
				foundBar = true
			}
		}
	}
	if !foundBar {
		t.Error("expected bar char in horizontal layout")
	}
}

func TestBarChart_Paint_MultiSeries(t *testing.T) {
	bc := newBarChartForTest()
	bc.AddSeries(BarSeries{
		Name:  "A",
		Data:  []BarData{{Label: "X", Value: 30}, {Label: "Y", Value: 60}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	bc.AddSeries(BarSeries{
		Name:  "B",
		Data:  []BarData{{Label: "X", Value: 40}, {Label: "Y", Value: 20}},
		Color: buffer.NamedColor(buffer.NamedRed),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Both series should produce bars
	cyanCount := 0
	redCount := 0
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == '█' {
				if c.Fg == buffer.NamedColor(buffer.NamedCyan) {
					cyanCount++
				}
				if c.Fg == buffer.NamedColor(buffer.NamedRed) {
					redCount++
				}
			}
		}
	}
	if cyanCount == 0 {
		t.Error("no cyan bars found")
	}
	if redCount == 0 {
		t.Error("no red bars found")
	}
}

func TestBarChart_Paint_WithFixedMaxVal(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetMaxVal(200)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 50}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // should not panic, bar should be 25% of height
}

func TestBarChart_Paint_ShowValues(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetShowValues(true)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 42}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // should not panic
}

func TestBarChart_Paint_ShowValuesHorizontal(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowValues(true)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 42}, {Label: "B", Value: 88}},
		Color: buffer.NamedColor(buffer.NamedGreen),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // should not panic
}

func TestBarChart_Paint_NoAxes(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetShowAxes(false)
	bc.SetShowTitle(false)
	bc.SetShowLegend(false)
	bc.SetShowGrid(false)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 50}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Should still draw bars
	foundBar := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == '█' {
				foundBar = true
			}
		}
	}
	if !foundBar {
		t.Error("expected bar even with no axes")
	}
}

func TestBarChart_Paint_ZeroValue(t *testing.T) {
	bc := newBarChartForTest()
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 0}, {Label: "B", Value: 100}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // zero value bar should be skipped, not panic
}

func TestBarChart_Paint_NegativeValue(t *testing.T) {
	bc := newBarChartForTest()
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: -50}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf) // negative value should not panic
}

func TestBarChart_Paint_TooSmallBounds(t *testing.T) {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 2})
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 10}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	buf := buffer.NewBuffer(2, 2)
	bc.Paint(buf) // should not panic, should just return
}

func TestBarChart_Paint_PartialBar(t *testing.T) {
	bc := newBarChartForTest()
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 1}}, // very small value → partial bar
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	bc.SetMaxVal(100)
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Should have either partial or full bar char
	foundAny := false
	partialChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			for _, pc := range partialChars {
				if c.Rune == pc {
					foundAny = true
					break
				}
			}
		}
	}
	if !foundAny {
		t.Error("expected partial or full bar char for small value")
	}
}

func TestBarChart_Paint_HorizontalPartialBar(t *testing.T) {
	bc := newBarChartForTest()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 1}},
		Color: buffer.NamedColor(buffer.NamedGreen),
	})
	bc.SetMaxVal(100)
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
	// Horizontal partial bar chars
	hPartialChars := []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}
	foundAny := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			for _, pc := range hPartialChars {
				if c.Rune == pc {
					foundAny = true
				}
			}
		}
	}
	if !foundAny {
		t.Error("expected horizontal partial bar char")
	}
}

// --- Children ---

func TestBarChart_Children(t *testing.T) {
	bc := NewBarChart()
	if bc.Children() != nil {
		t.Error("Children should return nil")
	}
}

// --- String ---

func TestBarChart_String(t *testing.T) {
	bc := NewBarChart()
	bc.SetTitle("Test Chart")
	bc.AddSeries(BarSeries{Name: "S", Color: buffer.NamedColor(buffer.NamedCyan)})
	s := bc.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

// --- Concurrent safety ---

func TestBarChart_ConcurrentAccess(t *testing.T) {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bc.AddSeries(BarSeries{
				Name:  "S",
				Data:  []BarData{{Label: "A", Value: float64(idx)}},
				Color: buffer.NamedColor(buffer.NamedCyan),
			})
		}(i)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bc.SeriesCount()
			_ = bc.Series(0)
			_ = bc.Title()
		}()
	}
	wg.Wait()
}

func TestBarChart_ConcurrentPaint(t *testing.T) {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Label: "A", Value: 50}, {Label: "B", Value: 80}},
		Color: buffer.NamedColor(buffer.NamedCyan),
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(60, 20)
			bc.Paint(buf)
		}()
	}
	wg.Wait()
}

func TestBarChart_ConcurrentConfig(t *testing.T) {
	bc := NewBarChart()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bc.SetShowGrid(idx%2 == 0)
			bc.SetShowAxes(idx%2 == 1)
			bc.SetShowLegend(true)
			bc.SetShowValues(idx%3 == 0)
			bc.SetMaxVal(float64(idx * 10))
			bc.SetGap(idx % 3)
		}(i)
	}
	wg.Wait()
}

// --- Helpers ---

func TestFormatBarVal(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0"},
		{5.5, "5.50"},
		{42, "42.0"},
		{500, "500"},
		{1500, "1.5K"},
		{1500000, "1.5M"},
		{1500000000, "1.5B"},
	}
	for _, tc := range tests {
		got := formatBarVal(tc.input)
		if got != tc.want {
			t.Errorf("formatBarVal(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestBarAutoColors(t *testing.T) {
	colors := barAutoColors(5)
	if len(colors) != 5 {
		t.Fatalf("barAutoColors(5) returned %d colors, want 5", len(colors))
	}
	// Colors should be distinct
	seen := make(map[buffer.Color]bool)
	for _, c := range colors {
		seen[c] = true
	}
	if len(seen) < 2 {
		t.Error("barAutoColors should produce distinct colors")
	}
}

func TestBarAutoColors_LargeN(t *testing.T) {
	colors := barAutoColors(20)
	if len(colors) != 20 {
		t.Fatalf("barAutoColors(20) returned %d colors, want 20", len(colors))
	}
}

// --- SetBounds / Bounds ---

func TestBarChart_SetBounds(t *testing.T) {
	bc := NewBarChart()
	r := Rect{X: 5, Y: 3, W: 50, H: 20}
	bc.SetBounds(r)
	got := bc.Bounds()
	if got != r {
		t.Errorf("Bounds = %v, want %v", got, r)
	}
}

// --- ID ---

func TestBarChart_ID(t *testing.T) {
	bc := NewBarChart()
	bc.SetID("my-barchart")
	if bc.ID() != "my-barchart" {
		t.Errorf("ID = %q, want %q", bc.ID(), "my-barchart")
	}
}
