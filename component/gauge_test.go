package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- Construction tests ---

func TestGauge_New(t *testing.T) {
	g := NewGauge()
	if g == nil {
		t.Fatal("expected non-nil Gauge")
	}
	if g.Value() != 0 {
		t.Fatalf("expected default value 0, got %f", g.Value())
	}
	if g.Ratio() != 0 {
		t.Fatalf("expected default ratio 0, got %f", g.Ratio())
	}
}

func TestGauge_ID(t *testing.T) {
	g := NewGauge()
	if g.ID() == "" {
		t.Fatal("expected non-empty ID from GenerateID")
	}
}

func TestGauge_ImplementsComponent(t *testing.T) {
	var _ Component = (*Gauge)(nil)
}

func TestGauge_Children(t *testing.T) {
	g := NewGauge()
	if g.Children() != nil {
		t.Fatal("Gauge should have no children")
	}
}

// --- SetValue tests ---

func TestGauge_SetValue(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	if g.Value() != 50 {
		t.Fatalf("expected value 50, got %f", g.Value())
	}
}

func TestGauge_SetValue_ClampAboveMax(t *testing.T) {
	g := NewGauge()
	g.SetValue(150)
	if g.Value() != 100 {
		t.Fatalf("expected clamped to 100, got %f", g.Value())
	}
}

func TestGauge_SetValue_ClampBelowMin(t *testing.T) {
	g := NewGauge()
	g.SetValue(-10)
	if g.Value() != 0 {
		t.Fatalf("expected clamped to 0, got %f", g.Value())
	}
}

// --- SetRange tests ---

func TestGauge_SetRange(t *testing.T) {
	g := NewGauge()
	g.SetRange(10, 20)
	g.SetValue(15)
	if g.Value() != 15 {
		t.Fatalf("expected value 15, got %f", g.Value())
	}
}

func TestGauge_SetRange_Reclamp(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetRange(0, 30)
	// Value should be clamped to new max.
	if g.Value() != 30 {
		t.Fatalf("expected re-clamped to 30, got %f", g.Value())
	}
}

func TestGauge_SetRange_IgnoreInvalid(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetRange(100, 50) // min >= max, should be ignored
	if g.Value() != 50 {
		t.Fatalf("expected value unchanged at 50, got %f", g.Value())
	}
}

func TestGauge_SetRange_NegativeRange(t *testing.T) {
	g := NewGauge()
	g.SetRange(-50, 50)
	g.SetValue(0)
	if g.Value() != 0 {
		t.Fatalf("expected value 0, got %f", g.Value())
	}
	if g.Ratio() != 0.5 {
		t.Fatalf("expected ratio 0.5, got %f", g.Ratio())
	}
}

// --- Ratio tests ---

func TestGauge_Ratio(t *testing.T) {
	g := NewGauge()
	g.SetValue(75)
	if g.Ratio() != 0.75 {
		t.Fatalf("expected ratio 0.75, got %f", g.Ratio())
	}
}

func TestGauge_Ratio_FullMax(t *testing.T) {
	g := NewGauge()
	g.SetValue(100)
	if g.Ratio() != 1.0 {
		t.Fatalf("expected ratio 1.0, got %f", g.Ratio())
	}
}

func TestGauge_Ratio_HalfRange(t *testing.T) {
	g := NewGauge()
	g.SetRange(0, 200)
	g.SetValue(100)
	if g.Ratio() != 0.5 {
		t.Fatalf("expected ratio 0.5, got %f", g.Ratio())
	}
}

// --- SetLabel tests ---

func TestGauge_SetLabel(t *testing.T) {
	g := NewGauge()
	g.SetLabel("CPU")
	// Just verify no panic; label is used in Paint/Measure.
}

// --- SetShowValue tests ---

func TestGauge_SetShowValue(t *testing.T) {
	g := NewGauge()
	g.SetShowValue(false)
	// Verify no panic.
	g.SetValue(50)
}

// --- Orientation / Radial tests ---

func TestGauge_SetOrientation(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	// Just verify no panic.
	g.SetValue(50)
}

func TestGauge_SetRadial(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetValue(50)
	// Should still have a valid ratio.
	if g.Ratio() != 0.5 {
		t.Fatalf("expected ratio 0.5, got %f", g.Ratio())
	}
}

func TestGauge_SetRadialBackToBar(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetOrientation(GaugeHorizontal) // should turn off radial
	g.SetValue(25)
}

// --- Thresholds tests ---

func TestGauge_SetThresholds(t *testing.T) {
	g := NewGauge()
	g.SetThresholds(DefaultThresholds())
	g.SetValue(90) // should fall in red zone (ratio 0.9 >= 0.85)
	// Verify no panic.
}

func TestGauge_SetThresholdsNil(t *testing.T) {
	g := NewGauge()
	g.SetThresholds(DefaultThresholds())
	g.SetThresholds(nil) // revert to gradient
	g.SetValue(50)
}

func TestDefaultThresholds(t *testing.T) {
	ts := DefaultThresholds()
	if len(ts) != 3 {
		t.Fatalf("expected 3 thresholds, got %d", len(ts))
	}
	// Green band [0, 0.6)
	if ts[0].Low != 0.0 || ts[0].High != 0.6 {
		t.Fatalf("unexpected green band: %v", ts[0])
	}
	// Yellow band [0.6, 0.85)
	if ts[1].Low != 0.6 || ts[1].High != 0.85 {
		t.Fatalf("unexpected yellow band: %v", ts[1])
	}
	// Red band [0.85, 1.01)
	if ts[2].Low != 0.85 {
		t.Fatalf("unexpected red band low: %v", ts[2])
	}
}

// --- Fill/Empty char tests ---

func TestGauge_SetFillChar(t *testing.T) {
	g := NewGauge()
	g.SetFillChar('#')
	g.SetValue(50)
}

func TestGauge_SetEmptyChar(t *testing.T) {
	g := NewGauge()
	g.SetEmptyChar('.')
	g.SetValue(50)
}

// --- gradientColor tests ---

func TestGradientColor_LowRatio(t *testing.T) {
	c := gradientColor(0.0)
	// At ratio=0: r=0, g=255 → pure green.
	if c.R() != 0 || c.G() != 255 {
		t.Fatalf("expected pure green at ratio 0, got R=%d G=%d", c.R(), c.G())
	}
}

func TestGradientColor_HighRatio(t *testing.T) {
	c := gradientColor(1.0)
	// At ratio=1: r=255, g=0 → pure red.
	if c.R() != 255 || c.G() != 0 {
		t.Fatalf("expected pure red at ratio 1, got R=%d G=%d", c.R(), c.G())
	}
}

func TestGradientColor_MidRatio(t *testing.T) {
	c := gradientColor(0.5)
	// At ratio=0.5: r=255, g=255 → yellow (boundary).
	if c.R() != 255 || c.G() != 255 {
		t.Fatalf("expected yellow at ratio 0.5, got R=%d G=%d", c.R(), c.G())
	}
}

func TestGradientColor_Clamped(t *testing.T) {
	c1 := gradientColor(-0.5)
	c2 := gradientColor(0.0)
	if c1 != c2 {
		t.Fatal("expected negative ratio clamped to 0")
	}

	c3 := gradientColor(1.5)
	c4 := gradientColor(1.0)
	if c3 != c4 {
		t.Fatal("expected ratio >1 clamped to 1")
	}
}

// --- clamp tests ---

func TestClamp(t *testing.T) {
	tests := []struct {
		v, lo, hi, want float64
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
		{5, 5, 5, 5},
		{3, 5, 10, 5},
	}
	for _, tc := range tests {
		got := clamp(tc.v, tc.lo, tc.hi)
		if got != tc.want {
			t.Fatalf("clamp(%v,%v,%v) = %v, want %v", tc.v, tc.lo, tc.hi, got, tc.want)
		}
	}
}

// --- Measure tests ---

func TestGauge_Measure_Horizontal(t *testing.T) {
	g := NewGauge()
	size := g.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.H != 1 {
		t.Fatalf("expected horizontal height 1, got %d", size.H)
	}
	if size.W != 40 {
		t.Fatalf("expected width 40, got %d", size.W)
	}
}

func TestGauge_Measure_HorizontalWithLabel(t *testing.T) {
	g := NewGauge()
	g.SetLabel("CPU")
	size := g.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.H != 2 {
		t.Fatalf("expected height 2 (label + bar), got %d", size.H)
	}
}

func TestGauge_Measure_Vertical(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	size := g.Measure(Constraints{MaxWidth: 10, MaxHeight: 10})
	if size.W != 1 {
		t.Fatalf("expected vertical width 1, got %d", size.W)
	}
	if size.H <= 0 {
		t.Fatalf("expected positive height, got %d", size.H)
	}
}

func TestGauge_Measure_Radial(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	size := g.Measure(Constraints{MaxWidth: 20, MaxHeight: 20})
	if size.W < 7 {
		t.Fatalf("expected radial width >= 7, got %d", size.W)
	}
	if size.H < 5 {
		t.Fatalf("expected radial height >= 5, got %d", size.H)
	}
}

func TestGauge_Measure_RadialWithLabel(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetLabel("Disk")
	size := g.Measure(Constraints{MaxWidth: 20, MaxHeight: 20})
	if size.H < 7 {
		t.Fatalf("expected radial+label height >= 7, got %d", size.H)
	}
}

func TestGauge_Measure_DefaultWidth(t *testing.T) {
	g := NewGauge()
	size := g.Measure(Constraints{}) // no MaxWidth
	if size.W != 40 {
		t.Fatalf("expected default width 40, got %d", size.W)
	}
}

// --- Paint tests ---

func TestGauge_Paint_Horizontal(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})

	buf := buffer.NewBuffer(20, 1)
	g.Paint(buf)

	// At 50% with width 20 (minus value display width), some cells should be fill.
	// Just verify no panic and buffer is populated.
	hasFill := false
	for x := 0; x < 20; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '█' {
			hasFill = true
			break
		}
	}
	if !hasFill {
		t.Fatal("expected at least one fill character in horizontal gauge")
	}
}

func TestGauge_Paint_HorizontalNoValue(t *testing.T) {
	g := NewGauge()
	g.SetShowValue(false)
	g.SetValue(0)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	g.Paint(buf)

	// All cells should be empty.
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '█' {
			t.Fatal("expected no fill at 0%")
		}
	}
}

func TestGauge_Paint_HorizontalFull(t *testing.T) {
	g := NewGauge()
	g.SetShowValue(false)
	g.SetValue(100)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	g.Paint(buf)

	// All cells should be fill.
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '█' {
			t.Fatalf("expected fill at x=%d, got %q", x, string(c.Rune))
		}
	}
}

func TestGauge_Paint_HorizontalWithLabel(t *testing.T) {
	g := NewGauge()
	g.SetLabel("CPU")
	g.SetShowValue(false)
	g.SetValue(75)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})

	buf := buffer.NewBuffer(20, 2)
	g.Paint(buf)

	// Label should be on row 0.
	found := false
	for x := 0; x < 20; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == 'C' {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected label 'C' in row 0")
	}
}

func TestGauge_Paint_Vertical(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetShowValue(false)
	g.SetValue(50)
	g.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})

	buf := buffer.NewBuffer(1, 10)
	g.Paint(buf)

	// Bottom ~5 cells should be fill, top ~5 empty.
	fillCount := 0
	for y := 0; y < 10; y++ {
		c := buf.GetCell(0, y)
		if c.Rune == '█' {
			fillCount++
		}
	}
	if fillCount == 0 {
		t.Fatal("expected some fill cells in vertical gauge")
	}
}

func TestGauge_Paint_VerticalFull(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetShowValue(false)
	g.SetValue(100)
	g.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})

	buf := buffer.NewBuffer(1, 5)
	g.Paint(buf)

	for y := 0; y < 5; y++ {
		c := buf.GetCell(0, y)
		if c.Rune != '█' {
			t.Fatalf("expected fill at y=%d, got %q", y, string(c.Rune))
		}
	}
}

func TestGauge_Paint_Radial(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetShowValue(false)
	g.SetValue(75)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})

	buf := buffer.NewBuffer(10, 10)
	g.Paint(buf)
	// Should not panic.
}

func TestGauge_Paint_RadialFull(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetShowValue(false)
	g.SetValue(100)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})

	buf := buffer.NewBuffer(10, 10)
	g.Paint(buf)
	// All segments should be filled — check some cells.
	hasContent := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != ' ' && c.Rune != 0 {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Fatal("expected non-empty cells in radial gauge")
	}
}

func TestGauge_Paint_ZeroBounds(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(1, 1)
	g.Paint(buf) // should not panic
}

func TestGauge_Paint_CustomChars(t *testing.T) {
	g := NewGauge()
	g.SetShowValue(false)
	g.SetFillChar('=')
	g.SetEmptyChar('-')
	g.SetValue(50)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	g.Paint(buf)

	// First 5 should be '=', last 5 should be '-'.
	for x := 0; x < 5; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '=' {
			t.Fatalf("expected '=' at x=%d, got %q", x, string(c.Rune))
		}
	}
}

func TestGauge_Paint_Thresholds(t *testing.T) {
	g := NewGauge()
	g.SetShowValue(false)
	g.SetThresholds(DefaultThresholds())
	g.SetValue(90) // red zone
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	g.Paint(buf)
	// Should not panic, fill color should be red.
}

// --- String test ---

func TestGauge_String(t *testing.T) {
	g := NewGauge()
	g.SetValue(42)
	s := g.String()
	if s == "" {
		t.Fatal("expected non-empty string representation")
	}
}

// --- Concurrency tests ---

func TestGauge_Concurrent(t *testing.T) {
	g := NewGauge()
	var wg sync.WaitGroup

	// Writer: SetValue
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			g.SetValue(float64(i))
		}
	}()

	// Reader: Value
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = g.Value()
		}
	}()

	// Reader: Ratio
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = g.Ratio()
		}
	}()

	// Reader: String
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = g.String()
		}
	}()

	wg.Wait()
}

func TestGauge_ConcurrentPaint(t *testing.T) {
	g := NewGauge()
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	g.SetLabel("CPU")

	var wg sync.WaitGroup

	// Painter
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(20, 2)
			g.Paint(buf)
		}
	}()

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			g.SetValue(float64(i))
		}
	}()

	// Reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = g.Ratio()
		}
	}()

	wg.Wait()
}

func TestGauge_ConcurrentSetRange(t *testing.T) {
	g := NewGauge()
	var wg sync.WaitGroup

	// Writer: SetRange
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			g.SetRange(0, float64(100+i))
		}
	}()

	// Writer: SetValue
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			g.SetValue(float64(i))
		}
	}()

	// Reader: Ratio
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			r := g.Ratio()
			if r < 0 || r > 1 {
				t.Errorf("ratio out of range: %f", r)
			}
		}
	}()

	wg.Wait()
}
