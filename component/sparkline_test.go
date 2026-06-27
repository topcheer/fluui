package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- Constructor and defaults ---

func TestSparkline_New(t *testing.T) {
	s := NewSparkline()
	if s.ID() == "" {
		t.Error("ID should not be empty")
	}
	if s.Count() != 0 {
		t.Errorf("Count: got %d, want 0", s.Count())
	}
	if s.colorMode != ColorSingle {
		t.Errorf("ColorMode: got %v, want ColorSingle", s.colorMode)
	}
	if !s.autoScale {
		t.Error("AutoScale should be true by default")
	}
}

func TestSparkline_SetData(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5})

	if s.Count() != 5 {
		t.Errorf("Count: got %d, want 5", s.Count())
	}

	data := s.Data()
	if len(data) != 5 || data[0] != 1 || data[4] != 5 {
		t.Errorf("Data: got %v", data)
	}
}

func TestSparkline_DataCopy(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	d := s.Data()
	d[0] = 999
	// Original should not change
	if s.Data()[0] != 1 {
		t.Error("Data should return a copy, not a reference")
	}
}

func TestSparkline_Clear(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	s.Clear()
	if s.Count() != 0 {
		t.Errorf("Count after Clear: got %d, want 0", s.Count())
	}
	if s.Min() != 0 || s.Max() != 0 {
		t.Errorf("Min/Max after Clear: got %f/%f", s.Min(), s.Max())
	}
}

// --- Push ---

func TestSparkline_Push(t *testing.T) {
	s := NewSparkline()
	s.Push(1.0, 0)
	s.Push(2.0, 0)
	s.Push(3.0, 0)

	if s.Count() != 3 {
		t.Errorf("Count: got %d, want 3", s.Count())
	}
}

func TestSparkline_PushWithMaxPoints(t *testing.T) {
	s := NewSparkline()
	for i := 0; i < 10; i++ {
		s.Push(float64(i), 5)
	}
	if s.Count() != 5 {
		t.Errorf("Count after trim: got %d, want 5", s.Count())
	}
	data := s.Data()
	// Should keep last 5 values: 5,6,7,8,9
	if data[0] != 5 || data[4] != 9 {
		t.Errorf("Data after trim: got %v, want [5,6,7,8,9]", data)
	}
}

func TestSparkline_PushZeroMax(t *testing.T) {
	s := NewSparkline()
	s.Push(1.0, 0) // maxPoints=0 means no trimming
	s.Push(2.0, 0)
	s.Push(3.0, 0)
	if s.Count() != 3 {
		t.Errorf("Count: got %d, want 3", s.Count())
	}
}

// --- Auto-scale / min-max ---

func TestSparkline_AutoScale(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{10, 20, 30, 40, 50})
	if s.Min() != 10 {
		t.Errorf("Min: got %f, want 10", s.Min())
	}
	if s.Max() != 50 {
		t.Errorf("Max: got %f, want 50", s.Max())
	}
}

func TestSparkline_AutoScaleNegative(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{-5, 0, 5})
	if s.Min() != -5 {
		t.Errorf("Min: got %f, want -5", s.Min())
	}
	if s.Max() != 5 {
		t.Errorf("Max: got %f, want 5", s.Max())
	}
}

func TestSparkline_AutoScaleEqualValues(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{7, 7, 7})
	// When all values are equal, max should be min+1 to avoid div by zero
	if s.Min() != 7 {
		t.Errorf("Min: got %f, want 7", s.Min())
	}
	if s.Max() != 8 {
		t.Errorf("Max: got %f, want 8 (should be min+1 for equal values)", s.Max())
	}
}

func TestSparkline_SetAutoScaleFalse(t *testing.T) {
	s := NewSparkline()
	s.SetAutoScale(false)
	s.SetData([]float64{10, 20, 30})
	// Without autoScale, min/max won't be recomputed
	if s.Min() != 0 || s.Max() != 0 {
		t.Errorf("Min/Max without autoScale: got %f/%f, want 0/0", s.Min(), s.Max())
	}
}

// --- Color modes ---

func TestSparkline_SetColorMode(t *testing.T) {
	s := NewSparkline()
	s.SetColorMode(ColorGradient)
	if s.colorMode != ColorGradient {
		t.Errorf("ColorMode: got %v, want ColorGradient", s.colorMode)
	}
}

func TestSparkline_SetColor(t *testing.T) {
	s := NewSparkline()
	c := buffer.NamedColor(buffer.NamedRed)
	s.SetColor(c)
	if !s.fgColor.Equal(c) {
		t.Error("fgColor not set correctly")
	}
}

func TestSparkline_SetStyle(t *testing.T) {
	s := NewSparkline()
	st := buffer.Style{Flags: buffer.Bold}
	s.SetStyle(st)
	if s.style.Flags != buffer.Bold {
		t.Error("style not set correctly")
	}
}

// --- Label ---

func TestSparkline_SetLabel(t *testing.T) {
	s := NewSparkline()
	s.SetLabel("CPU")
	if s.Label() != "CPU" {
		t.Errorf("Label: got %q, want %q", s.Label(), "CPU")
	}
}

func TestSparkline_SetShowMinMax(t *testing.T) {
	s := NewSparkline()
	s.SetShowMinMax(true)
	if !s.showMinMax {
		t.Error("showMinMax should be true")
	}
}

// --- Scroll ---

func TestSparkline_SetScrollX(t *testing.T) {
	s := NewSparkline()
	s.SetScrollX(5)
	if s.ScrollX() != 5 {
		t.Errorf("ScrollX: got %d, want 5", s.ScrollX())
	}
}

func TestSparkline_SetScrollXNegative(t *testing.T) {
	s := NewSparkline()
	s.SetScrollX(-5)
	if s.ScrollX() != 0 {
		t.Errorf("ScrollX: got %d, want 0 (clamped)", s.ScrollX())
	}
}

// --- Measure ---

func TestSparkline_MeasureEmpty(t *testing.T) {
	s := NewSparkline()
	size := s.Measure(Unbounded())
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}
	if size.W < 1 {
		t.Errorf("W should be at least 1")
	}
}

func TestSparkline_MeasureWithData(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5})
	size := s.Measure(Unbounded())
	if size.W != 5 {
		t.Errorf("W: got %d, want 5", size.W)
	}
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}
}

func TestSparkline_MeasureWithLabel(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	s.SetLabel("CPU")
	size := s.Measure(Unbounded())
	// 3 bars + 1 space + 3 label chars = 7
	if size.W != 7 {
		t.Errorf("W: got %d, want 7", size.W)
	}
}

func TestSparkline_MeasureClamped(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	size := s.Measure(Bounded(5, 1))
	if size.W != 5 {
		t.Errorf("W: got %d, want 5 (clamped)", size.W)
	}
}

// --- Paint ---

func TestSparkline_PaintBasic(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5})
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf)

	// First cell should be a sparkline char (not blank)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("cell[0,0] should have a sparkline rune")
	}
}

func TestSparkline_PaintEmpty(t *testing.T) {
	s := NewSparkline()
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	// Should not panic
	s.Paint(buf)
}

func TestSparkline_PaintZeroBounds(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	// Should not panic
	s.Paint(buf)
}

func TestSparkline_PaintLabel(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	s.SetLabel("CPU")
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})

	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)

	// After 3 bars + 1 space, "C" should be at position 4
	cell := buf.GetCell(4, 0)
	if cell.Rune != 'C' {
		t.Errorf("Expected 'C' at (4,0), got %q", string(cell.Rune))
	}
}

func TestSparkline_PaintGradientColors(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{0, 50, 100})
	s.SetColorMode(ColorGradient)
	s.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	s.Paint(buf)

	// Lowest value should be green, highest should be red
	lowCell := buf.GetCell(0, 0)
	highCell := buf.GetCell(2, 0)

	lowGreen := buffer.NamedColor(buffer.NamedGreen)
	highRed := buffer.NamedColor(buffer.NamedRed)

	if !lowCell.Fg.Equal(lowGreen) {
		t.Errorf("Low bar should be green, got %v", lowCell.Fg)
	}
	if !highCell.Fg.Equal(highRed) {
		t.Errorf("High bar should be red, got %v", highCell.Fg)
	}
}

func TestSparkline_PaintValueColors(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{10, 50, 90})
	s.SetColorMode(ColorValue)
	s.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	s.Paint(buf)

	// Should have different colors for different value ranges
	c0 := buf.GetCell(0, 0).Fg
	c2 := buf.GetCell(2, 0).Fg
	if c0.Equal(c2) {
		t.Error("Low and high values should have different colors")
	}
}

func TestSparkline_PaintSingleColor(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 5, 10})
	s.SetColor(buffer.NamedColor(buffer.NamedCyan))
	s.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	s.Paint(buf)

	for i := 0; i < 3; i++ {
		cell := buf.GetCell(i, 0)
		if !cell.Fg.Equal(buffer.NamedColor(buffer.NamedCyan)) {
			t.Errorf("Bar %d should be cyan, got %v", i, cell.Fg)
		}
	}
}

func TestSparkline_PaintScroll(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	s.SetScrollX(5) // start from index 5
	s.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})

	buf := buffer.NewBuffer(3, 1)
	s.Paint(buf)

	// Should render bars 5,6,7 (indices 5-7 of data)
	// All cells should have sparkline runes
	for i := 0; i < 3; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Rune == 0 {
			t.Errorf("cell[%d,0] should not be empty", i)
		}
	}
}

func TestSparkline_PaintWithMinMax(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{10, 20, 30})
	s.SetShowMinMax(true)
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})

	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf)

	// Should contain ".." somewhere (min..max)
	foundDots := false
	for i := 0; i < 30; i++ {
		if buf.GetCell(i, 0).Rune == '.' {
			foundDots = true
			break
		}
	}
	if !foundDots {
		t.Error("Paint with ShowMinMax should contain '..'")
	}
}

func TestSparkline_PaintAtOffset(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3})
	s.SetBounds(Rect{X: 5, Y: 2, W: 10, H: 1})

	buf := buffer.NewBuffer(20, 5)
	s.Paint(buf)

	cell := buf.GetCell(5, 2)
	if cell.Rune == 0 {
		t.Error("cell at offset (5,2) should not be empty")
	}
}

// --- Value to bar mapping ---

func TestSparkline_BarMapping(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{0, 1, 2, 3, 4, 5, 6, 7})
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf)

	// Lowest value (0) should map to lowest bar char '▁'
	// Highest value (7) should map to highest bar char '█'
	low := buf.GetCell(0, 0).Rune
	high := buf.GetCell(7, 0).Rune

	if low != '▁' {
		t.Errorf("Lowest bar should be '▁', got %q", string(low))
	}
	if high != '█' {
		t.Errorf("Highest bar should be '█', got %q", string(high))
	}
}

func TestSparkline_IncreasingBars(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4})
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf)

	// Each bar should be taller than the previous
	for i := 0; i < 3; i++ {
		curr := buf.GetCell(i, 0).Rune
		next := buf.GetCell(i+1, 0).Rune
		if curr >= next {
			t.Errorf("Bar %d (%q) should be shorter than bar %d (%q)", i, string(curr), i+1, string(next))
		}
	}
}

// --- Concurrency ---

func TestSparkline_ConcurrentAccess(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5})

	var wg sync.WaitGroup
	wg.Add(3)

	// Writer: Push data
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.Push(float64(i), 10)
		}
	}()

	// Reader: Read data and stats
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = s.Data()
			_ = s.Min()
			_ = s.Max()
			_ = s.Count()
		}
	}()

	// Reader: Measure
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.Measure(Unbounded())
		}
	}()

	wg.Wait()
}

func TestSparkline_ConcurrentPaint(t *testing.T) {
	s := NewSparkline()
	s.SetData([]float64{1, 2, 3, 4, 5})
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})

	var wg sync.WaitGroup
	wg.Add(2)

	// Painter
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(20, 1)
			s.Paint(buf)
		}
	}()

	// Mutator
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			s.SetLabel("test")
			s.SetColorMode(ColorGradient)
		}
	}()

	wg.Wait()
}
