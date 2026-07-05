package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// --- Constructor tests ---

func TestNewLineChart_Defaults(t *testing.T) {
	lc := NewLineChart()
	if !lc.showGrid {
		t.Error("showGrid should default to true")
	}
	if !lc.showAxes {
		t.Error("showAxes should default to true")
	}
	if !lc.showLegend {
		t.Error("showLegend should default to true")
	}
	if !lc.showTitle {
		t.Error("showTitle should default to true")
	}
	if !lc.autoScale {
		t.Error("autoScale should default to true")
	}
	if lc.xAxis.LabelCount != 5 {
		t.Errorf("xAxis.LabelCount = %d, want 5", lc.xAxis.LabelCount)
	}
	if lc.yAxis.LabelCount != 5 {
		t.Errorf("yAxis.LabelCount = %d, want 5", lc.yAxis.LabelCount)
	}
	if lc.SeriesCount() != 0 {
		t.Errorf("SeriesCount = %d, want 0", lc.SeriesCount())
	}
}

// --- Series management tests ---

func TestLineChart_AddSeries(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Name:  "S1",
		Data:  []ChartPoint{{1, 2}, {3, 4}},
		Color: buffer.NamedColor(buffer.NamedRed),
	})
	if lc.SeriesCount() != 1 {
		t.Fatalf("SeriesCount = %d, want 1", lc.SeriesCount())
	}
	s := lc.Series(0)
	if s.Name != "S1" {
		t.Errorf("Name = %q, want %q", s.Name, "S1")
	}
	if len(s.Data) != 2 {
		t.Errorf("Data len = %d, want 2", len(s.Data))
	}
}

func TestLineChart_AddMultipleSeries(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "A", Data: []ChartPoint{{0, 0}}})
	lc.AddSeries(ChartSeries{Name: "B", Data: []ChartPoint{{0, 1}}})
	lc.AddSeries(ChartSeries{Name: "C", Data: []ChartPoint{{0, 2}}})
	if lc.SeriesCount() != 3 {
		t.Errorf("SeriesCount = %d, want 3", lc.SeriesCount())
	}
}

func TestLineChart_SetSeries(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "old"})
	lc.SetSeries([]ChartSeries{
		{Name: "new1"},
		{Name: "new2"},
	})
	if lc.SeriesCount() != 2 {
		t.Fatalf("SeriesCount = %d, want 2", lc.SeriesCount())
	}
	if lc.Series(0).Name != "new1" {
		t.Errorf("Series(0).Name = %q", lc.Series(0).Name)
	}
}

func TestLineChart_ClearSeries(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "S1"})
	lc.AddSeries(ChartSeries{Name: "S2"})
	lc.ClearSeries()
	if lc.SeriesCount() != 0 {
		t.Errorf("SeriesCount = %d after ClearSeries, want 0", lc.SeriesCount())
	}
}

func TestLineChart_Series_OutOfBounds(t *testing.T) {
	lc := NewLineChart()
	s := lc.Series(-1)
	if s.Name != "" {
		t.Error("Series(-1) should return zero value")
	}
	s = lc.Series(0)
	if s.Name != "" {
		t.Error("Series(0) with no series should return zero value")
	}
}

func TestLineChart_SeriesNames(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "alpha"})
	lc.AddSeries(ChartSeries{Name: "beta"})
	names := lc.SeriesNames()
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("SeriesNames = %v", names)
	}
}

func TestLineChart_PointCount(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "S1", Data: []ChartPoint{{0, 0}, {1, 1}, {2, 2}}})
	lc.AddSeries(ChartSeries{Name: "S2", Data: []ChartPoint{{0, 0}, {1, 1}}})
	if lc.PointCount() != 5 {
		t.Errorf("PointCount = %d, want 5", lc.PointCount())
	}
}

// --- Configuration tests ---

func TestLineChart_SetTitle(t *testing.T) {
	lc := NewLineChart()
	lc.SetTitle("My Chart")
	if lc.Title() != "My Chart" {
		t.Errorf("Title() = %q, want %q", lc.Title(), "My Chart")
	}
}

func TestLineChart_SetShowFlags(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowGrid(false)
	lc.SetShowAxes(false)
	lc.SetShowLegend(false)
	lc.SetShowTitle(false)
	if lc.showGrid || lc.showAxes || lc.showLegend || lc.showTitle {
		t.Error("show flags should all be false")
	}

	lc.SetShowGrid(true)
	lc.SetShowAxes(true)
	lc.SetShowLegend(true)
	lc.SetShowTitle(true)
	if !lc.showGrid || !lc.showAxes || !lc.showLegend || !lc.showTitle {
		t.Error("show flags should all be true")
	}
}

func TestLineChart_SetAxis(t *testing.T) {
	lc := NewLineChart()
	lc.SetXAxis(ChartAxisConfig{Title: "X", Min: 0, Max: 100, LabelCount: 3})
	lc.SetYAxis(ChartAxisConfig{Title: "Y", Min: -10, Max: 10, LabelCount: 5})

	if lc.xAxis.Title != "X" || lc.xAxis.Min != 0 || lc.xAxis.Max != 100 {
		t.Errorf("xAxis = %+v", lc.xAxis)
	}
	if lc.yAxis.Title != "Y" || lc.yAxis.Min != -10 || lc.yAxis.Max != 10 {
		t.Errorf("yAxis = %+v", lc.yAxis)
	}
}

func TestLineChart_SetAutoScale(t *testing.T) {
	lc := NewLineChart()
	lc.SetAutoScale(false)
	if lc.autoScale {
		t.Error("autoScale should be false")
	}
	lc.SetAutoScale(true)
	if !lc.autoScale {
		t.Error("autoScale should be true")
	}
}

func TestLineChart_SetGridStyle(t *testing.T) {
	lc := NewLineChart()
	style := buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)}
	lc.SetGridStyle(style)
	if lc.gridStyle.Fg != buffer.NamedColor(buffer.NamedBlue) {
		t.Error("gridStyle not set correctly")
	}
}

func TestLineChart_SetTheme(t *testing.T) {
	lc := NewLineChart()
	tm := theme.Default()
	lc.SetTheme(tm)
	if lc.currentTheme != tm {
		t.Error("theme not set")
	}
}

// --- Range computation tests ---

func TestLineChart_computeRanges_AutoScale(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Data: []ChartPoint{{1, 10}, {5, 50}, {10, 100}},
	})
	xMin, xMax, yMin, yMax := lc.computeRanges()
	if xMin != 1 || xMax != 10 {
		t.Errorf("X range = [%f, %f], want [1, 10]", xMin, xMax)
	}
	if yMin != 10 || yMax != 100 {
		t.Errorf("Y range = [%f, %f], want [10, 100]", yMin, yMax)
	}
}

func TestLineChart_computeRanges_FixedAxis(t *testing.T) {
	lc := NewLineChart()
	lc.SetYAxis(ChartAxisConfig{Min: 0, Max: 200})
	lc.AddSeries(ChartSeries{
		Data: []ChartPoint{{0, 50}, {1, 100}},
	})
	_, _, yMin, yMax := lc.computeRanges()
	if yMin != 0 || yMax != 200 {
		t.Errorf("Y range = [%f, %f], want [0, 200]", yMin, yMax)
	}
}

func TestLineChart_computeRanges_NoData(t *testing.T) {
	lc := NewLineChart()
	xMin, xMax, yMin, yMax := lc.computeRanges()
	if xMin != 0 || xMax != 1 || yMin != 0 || yMax != 1 {
		t.Errorf("range with no data = [%f,%f,%f,%f], want [0,1,0,1]", xMin, xMax, yMin, yMax)
	}
}

func TestLineChart_computeRanges_EqualMinMax(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Data: []ChartPoint{{5, 5}, {5, 5}},
	})
	xMin, xMax, yMin, yMax := lc.computeRanges()
	if xMax <= xMin {
		t.Error("X range should be expanded for equal values")
	}
	if yMax <= yMin {
		t.Error("Y range should be expanded for equal values")
	}
}

func TestLineChart_Range(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Data: []ChartPoint{{0, 0}, {10, 100}},
	})
	xMin, xMax, yMin, yMax := lc.Range()
	if xMin != 0 || xMax != 10 || yMin != 0 || yMax != 100 {
		t.Errorf("Range() = [%f,%f,%f,%f]", xMin, xMax, yMin, yMax)
	}
}

// --- Measure tests ---

func TestLineChart_Measure_Default(t *testing.T) {
	lc := NewLineChart()
	s := lc.Measure(Unbounded())
	if s.W < 10 || s.H < 5 {
		t.Errorf("Measure = %dx%d, too small", s.W, s.H)
	}
}

func TestLineChart_Measure_Constrained(t *testing.T) {
	lc := NewLineChart()
	s := lc.Measure(Constraints{MaxWidth: 20, MaxHeight: 8})
	if s.W > 20 || s.H > 8 {
		t.Errorf("Measure = %dx%d, exceeds constraints", s.W, s.H)
	}
}

func TestLineChart_Measure_MinSize(t *testing.T) {
	lc := NewLineChart()
	s := lc.Measure(Constraints{MaxWidth: 3, MaxHeight: 2})
	if s.W < 10 {
		t.Errorf("Measure W = %d, should clamp to min 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("Measure H = %d, should clamp to min 5", s.H)
	}
}

// --- Paint tests ---

func TestLineChart_Paint_TooSmall(t *testing.T) {
	lc := NewLineChart()
	lc.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 2})
	buf := buffer.NewBuffer(3, 2)
	// Should not panic and should do nothing
	lc.Paint(buf)
}

func TestLineChart_Paint_Basic(t *testing.T) {
	lc := NewLineChart()
	lc.SetTitle("Test Chart")
	lc.AddSeries(ChartSeries{
		Name:  "data",
		Data:  []ChartPoint{{0, 0}, {5, 10}, {10, 5}},
		Color: buffer.NamedColor(buffer.NamedGreen),
		Marker: ChartMarkerDot,
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	lc.Paint(buf)

	// Check that at least some non-space cells were drawn
	hasContent := false
	for y := 0; y < 15; y++ {
		for x := 0; x < 50; x++ {
			if buf.GetCell(x, y).Rune != ' ' && buf.GetCell(x, y).Rune != 0 {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no visible content")
	}
}

func TestLineChart_Paint_EmptySeries(t *testing.T) {
	lc := NewLineChart()
	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 12})
	buf := buffer.NewBuffer(40, 12)
	// Should not panic with no series
	lc.Paint(buf)
}

func TestLineChart_Paint_MultipleSeries(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowTitle(false)
	lc.SetShowLegend(false)
	lc.AddSeries(ChartSeries{
		Name:  "A",
		Data:  []ChartPoint{{0, 0}, {10, 100}},
		Color: buffer.NamedColor(buffer.NamedRed),
	})
	lc.AddSeries(ChartSeries{
		Name:  "B",
		Data:  []ChartPoint{{0, 100}, {10, 0}},
		Color: buffer.NamedColor(buffer.NamedBlue),
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	lc.Paint(buf)

	// Check that both red and blue cells exist
	hasRed := false
	hasBlue := false
	for y := 0; y < 15; y++ {
		for x := 0; x < 50; x++ {
			c := buf.GetCell(x, y)
			if c.Fg == buffer.NamedColor(buffer.NamedRed) {
				hasRed = true
			}
			if c.Fg == buffer.NamedColor(buffer.NamedBlue) {
				hasBlue = true
			}
		}
	}
	if !hasRed {
		t.Error("Expected red series cells in buffer")
	}
	if !hasBlue {
		t.Error("Expected blue series cells in buffer")
	}
}

func TestLineChart_Paint_NoAxes(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowAxes(false)
	lc.SetShowTitle(false)
	lc.SetShowLegend(false)
	lc.AddSeries(ChartSeries{
		Data:  []ChartPoint{{0, 0}, {10, 10}},
		Color: buffer.NamedColor(buffer.NamedWhite),
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	lc.Paint(buf)
	// Should not panic
}

func TestLineChart_Paint_NoGrid(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowGrid(false)
	lc.AddSeries(ChartSeries{
		Data:  []ChartPoint{{0, 0}, {10, 10}},
		Color: buffer.NamedColor(buffer.NamedWhite),
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 12})
	buf := buffer.NewBuffer(40, 12)
	lc.Paint(buf)
	// Should not panic, and grid dots should not be present
}

func TestLineChart_Paint_WithMarkers(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowTitle(false)
	lc.SetShowLegend(false)
	lc.SetShowAxes(false)
	lc.SetShowGrid(false)
	lc.AddSeries(ChartSeries{
		Data:   []ChartPoint{{0, 5}, {5, 5}, {10, 5}},
		Color:  buffer.NamedColor(buffer.NamedYellow),
		Marker: ChartMarkerDot,
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	lc.Paint(buf)

	// Check that at least one marker character exists
	hasDot := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			if buf.GetCell(x, y).Rune == '●' {
				hasDot = true
				break
			}
		}
	}
	if !hasDot {
		t.Error("Expected dot markers in output")
	}
}

func TestLineChart_Paint_AllMarkerTypes(t *testing.T) {
	markers := []ChartMarker{ChartMarkerDot, ChartMarkerPlus, ChartMarkerStar, ChartMarkerNone}
	expectedRunes := []rune{'●', '+', '*', 0}

	for i, m := range markers {
		lc := NewLineChart()
		lc.SetShowTitle(false)
		lc.SetShowLegend(false)
		lc.SetShowAxes(false)
		lc.SetShowGrid(false)
		lc.AddSeries(ChartSeries{
			Data:   []ChartPoint{{0, 5}, {10, 5}},
			Color:  buffer.NamedColor(buffer.NamedWhite),
			Marker: m,
		})
		lc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
		buf := buffer.NewBuffer(20, 10)
		lc.Paint(buf)

		if expectedRunes[i] != 0 {
			found := false
			for y := 0; y < 10; y++ {
				for x := 0; x < 20; x++ {
					if buf.GetCell(x, y).Rune == expectedRunes[i] {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("Marker %d: expected rune %q not found", i, expectedRunes[i])
			}
		}
	}
}

// --- Coordinate mapping tests ---

func TestLineChart_dataToScreen(t *testing.T) {
	lc := NewLineChart()
	// Chart area: cx=5, cy=0, cw=20, ch=10
	// Data range: xMin=0, xMax=10, yMin=0, yMax=100
	pt := lc.dataToScreen(0, 0, 5, 0, 20, 10, 0, 10, 0, 100)
	// (0,0) should map to bottom-left corner: (5, 9)
	if pt[0] != 5 || pt[1] != 9 {
		t.Errorf("dataToScreen(0,0) = (%d,%d), want (5,9)", pt[0], pt[1])
	}

	pt = lc.dataToScreen(10, 100, 5, 0, 20, 10, 0, 10, 0, 100)
	// (10,100) should map to top-right corner: (24, 0)
	if pt[0] != 24 || pt[1] != 0 {
		t.Errorf("dataToScreen(10,100) = (%d,%d), want (24,0)", pt[0], pt[1])
	}
}

func TestLineChart_dataToScreen_Clamp(t *testing.T) {
	lc := NewLineChart()
	// Value beyond range should be clamped
	pt := lc.dataToScreen(-100, -100, 5, 0, 20, 10, 0, 10, 0, 100)
	if pt[0] != 5 || pt[1] != 9 {
		t.Errorf("dataToScreen(-100,-100) = (%d,%d), want clamped (5,9)", pt[0], pt[1])
	}

	pt = lc.dataToScreen(100, 200, 5, 0, 20, 10, 0, 10, 0, 100)
	if pt[0] != 24 || pt[1] != 0 {
		t.Errorf("dataToScreen(100,200) = (%d,%d), want clamped (24,0)", pt[0], pt[1])
	}
}

// --- Helper function tests ---

func TestFormatAxisVal(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0"},
		{10, "10"},
		{-5, "-5"},
		{1000000, "1000000"},
		{3.14, "3.14"},
		{12.5, "12.5"},
		{123.456, "123"},
	}
	for _, tt := range tests {
		got := formatAxisVal(tt.input)
		if got != tt.want {
			t.Errorf("formatAxisVal(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSeriesMarkerRune(t *testing.T) {
	if seriesMarkerRune(ChartMarkerDot) != '●' {
		t.Error("Dot marker rune mismatch")
	}
	if seriesMarkerRune(ChartMarkerPlus) != '+' {
		t.Error("Plus marker rune mismatch")
	}
	if seriesMarkerRune(ChartMarkerStar) != '*' {
		t.Error("Star marker rune mismatch")
	}
	if seriesMarkerRune(ChartMarkerNone) != '─' {
		t.Error("None marker rune mismatch")
	}
}

func TestAbsInt(t *testing.T) {
	if absInt(-5) != 5 {
		t.Error("absInt(-5) should be 5")
	}
	if absInt(5) != 5 {
		t.Error("absInt(5) should be 5")
	}
	if absInt(0) != 0 {
		t.Error("absInt(0) should be 0")
	}
}

func TestSignInt(t *testing.T) {
	if signInt(5) != 1 {
		t.Error("signInt(5) should be 1")
	}
	if signInt(-5) != -1 {
		t.Error("signInt(-5) should be -1")
	}
	if signInt(0) != 0 {
		t.Error("signInt(0) should be 0")
	}
}

func TestMaxInt(t *testing.T) {
	if maxInt(3, 7) != 7 {
		t.Error("maxInt(3,7) should be 7")
	}
	if maxInt(10, 2) != 10 {
		t.Error("maxInt(10,2) should be 10")
	}
}

func TestMinInt(t *testing.T) {
	if minInt(3, 7) != 3 {
		t.Error("minInt(3,7) should be 3")
	}
	if minInt(10, 2) != 2 {
		t.Error("minInt(10,2) should be 2")
	}
}

// --- String method test ---

func TestLineChart_String(t *testing.T) {
	lc := NewLineChart()
	lc.SetTitle("Revenue")
	lc.AddSeries(ChartSeries{Name: "Q1", Data: []ChartPoint{{0, 0}, {1, 1}}})
	s := lc.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

// --- Concurrent access test -----

func TestLineChart_ConcurrentAccess(t *testing.T) {
	lc := NewLineChart()
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			lc.AddSeries(ChartSeries{
				Name:  "S",
				Data:  []ChartPoint{{0, float64(n)}, {1, float64(n + 1)}},
				Color: buffer.NamedColor(buffer.NamedGreen),
			})
			lc.SeriesCount()
			lc.SeriesNames()
			lc.Range()
			lc.PointCount()
		}(i)
	}
	wg.Wait()
}

func TestLineChart_ConcurrentPaint(t *testing.T) {
	lc := NewLineChart()
	lc.SetTitle("Concurrent")
	lc.AddSeries(ChartSeries{
		Name:   "data",
		Data:   []ChartPoint{{0, 0}, {5, 10}, {10, 5}},
		Color:  buffer.NamedColor(buffer.NamedCyan),
		Marker: ChartMarkerDot,
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(60, 20)
			lc.Paint(buf)
		}()
	}
	wg.Wait()
}

func TestLineChart_ConcurrentReadWrite(t *testing.T) {
	lc := NewLineChart()
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	lc.AddSeries(ChartSeries{
		Name: "init",
		Data: []ChartPoint{{0, 0}, {1, 1}},
	})

	var wg sync.WaitGroup
	// Writer goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			lc.AddSeries(ChartSeries{
				Name: "writer",
				Data: []ChartPoint{{0, float64(n)}},
			})
		}(i)
	}
	// Reader goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lc.SeriesCount()
			lc.Range()
			lc.SeriesNames()
			buf := buffer.NewBuffer(60, 20)
			lc.Paint(buf)
		}()
	}
	wg.Wait()
}

// --- Edge case tests -----

func TestLineChart_Paint_NegativeValues(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowTitle(false)
	lc.SetShowLegend(false)
	lc.AddSeries(ChartSeries{
		Data:  []ChartPoint{{0, -50}, {5, 0}, {10, 50}},
		Color: buffer.NamedColor(buffer.NamedGreen),
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	lc.Paint(buf)
	// Should handle negative values without panic
}

func TestLineChart_Paint_SinglePoint(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowTitle(false)
	lc.SetShowLegend(false)
	lc.AddSeries(ChartSeries{
		Data:   []ChartPoint{{5, 5}},
		Color:  buffer.NamedColor(buffer.NamedRed),
		Marker: ChartMarkerDot,
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	lc.Paint(buf)
	// Should render a single marker without panic
}

func TestLineChart_Paint_ManyPoints(t *testing.T) {
	lc := NewLineChart()
	data := make([]ChartPoint, 100)
	for i := 0; i < 100; i++ {
		data[i] = ChartPoint{X: float64(i), Y: float64(i % 20)}
	}
	lc.AddSeries(ChartSeries{
		Name:  "wave",
		Data:  data,
		Color: buffer.NamedColor(buffer.NamedCyan),
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	lc.Paint(buf)
	// Should handle many points without panic
}

func TestLineChart_Paint_LegendOverflow(t *testing.T) {
	lc := NewLineChart()
	// Add many series with long names to overflow the legend
	for i := 0; i < 10; i++ {
		lc.AddSeries(ChartSeries{
			Name:  "LongSeriesNameThatOverflows",
			Data:  []ChartPoint{{0, float64(i)}},
			Color: buffer.NamedColor(buffer.NamedWhite),
		})
	}
	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	lc.Paint(buf)
	// Should not panic when legend overflows
}

// --- drawLineSegment character selection tests -----

func TestLineChart_DrawLineSegment_Horizontal(t *testing.T) {
	lc := NewLineChart()
	buf := buffer.NewBuffer(20, 5)
	lc.drawLineSegment(buf, 2, 2, 10, 2, buffer.NamedColor(buffer.NamedGreen), buffer.Color{})
	// Horizontal line should use '─'
	if buf.GetCell(5, 2).Rune != '─' {
		t.Errorf("Horizontal line char = %q, want '─'", buf.GetCell(5, 2).Rune)
	}
}

func TestLineChart_DrawLineSegment_Vertical(t *testing.T) {
	lc := NewLineChart()
	buf := buffer.NewBuffer(5, 20)
	lc.drawLineSegment(buf, 2, 2, 2, 10, buffer.NamedColor(buffer.NamedGreen), buffer.Color{})
	// Vertical line should use '│'
	if buf.GetCell(2, 5).Rune != '│' {
		t.Errorf("Vertical line char = %q, want '│'", buf.GetCell(2, 5).Rune)
	}
}

func TestLineChart_DrawLineSegment_Ascending(t *testing.T) {
	lc := NewLineChart()
	buf := buffer.NewBuffer(20, 20)
	// Ascending: from (1,10) to (10,1) — going up-right
	lc.drawLineSegment(buf, 1, 10, 10, 1, buffer.NamedColor(buffer.NamedGreen), buffer.Color{})
	// Should use '╱'
	found := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if buf.GetCell(x, y).Rune == '╱' {
				found = true
			}
		}
	}
	if !found {
		t.Error("Expected '╱' character for ascending line")
	}
}

func TestLineChart_DrawLineSegment_Descending(t *testing.T) {
	lc := NewLineChart()
	buf := buffer.NewBuffer(20, 20)
	// Descending: from (1,1) to (10,10) — going down-right
	lc.drawLineSegment(buf, 1, 1, 10, 10, buffer.NamedColor(buffer.NamedGreen), buffer.Color{})
	// Should use '╲'
	found := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if buf.GetCell(x, y).Rune == '╲' {
				found = true
			}
		}
	}
	if !found {
		t.Error("Expected '╲' character for descending line")
	}
}

func TestLineChart_DrawLineSegment_SamePoint(t *testing.T) {
	lc := NewLineChart()
	buf := buffer.NewBuffer(5, 5)
	// Same start and end point
	lc.drawLineSegment(buf, 2, 2, 2, 2, buffer.NamedColor(buffer.NamedGreen), buffer.Color{})
	// Should draw something (a dot)
	if buf.GetCell(2, 2).Rune == ' ' {
		t.Error("Expected non-space at same point")
	}
}
