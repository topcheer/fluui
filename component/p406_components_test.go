package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === CircularProgress tests ===

func TestP406_CircularProgress_New(t *testing.T) {
	c := NewCircularProgress(0.75)
	if c.Value() != 0.75 { t.Errorf("Value = %v", c.Value()) }
	if c.Style() != ProgressStyleRing { t.Errorf("Style = %v", c.Style()) }
	if c.BarWidth() != 5 { t.Errorf("BarWidth = %d", c.BarWidth()) }
	if c.ID() == "" { t.Error("ID empty") }
}

func TestP406_CircularProgress_SetValue(t *testing.T) {
	c := NewCircularProgress(0)
	c.SetValue(0.5)
	if c.Value() != 0.5 { t.Errorf("Value = %v", c.Value()) }
	c.SetValue(-1)
	if c.Value() != 0 { t.Errorf("clamped = %v", c.Value()) }
	c.SetValue(2)
	if c.Value() != 1 { t.Errorf("clamped = %v", c.Value()) }
}

func TestP406_CircularProgress_SetLabel(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetLabel("Loading")
	if c.Label() != "Loading" { t.Errorf("Label = %q", c.Label()) }
}

func TestP406_CircularProgress_SetStyle(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetStyle(ProgressStyleDots)
	if c.Style() != ProgressStyleDots { t.Error("should be Dots") }
	c.SetStyle(ProgressStyleBlock)
	if c.Style() != ProgressStyleBlock { t.Error("should be Block") }
}

func TestP406_CircularProgress_Measure(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetLabel("Progress")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 { t.Errorf("H = %d", s.H) }
	if s.W < 10 { t.Errorf("W = %d, too small", s.W) }
}

func TestP406_CircularProgress_Paint_Ring(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	c.Paint(buf) // should render arc icon + percentage
}

func TestP406_CircularProgress_Paint_Dots(t *testing.T) {
	c := NewCircularProgress(0.6)
	c.SetStyle(ProgressStyleDots)
	c.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	c.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '●' { t.Errorf("dot[0] = %q, want ●", string(c0.Rune)) }
}

func TestP406_CircularProgress_Paint_Block(t *testing.T) {
	c := NewCircularProgress(0.4)
	c.SetStyle(ProgressStyleBlock)
	c.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	c.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '▰' { t.Errorf("block[0] = %q, want ▰", string(c0.Rune)) }
}

func TestP406_CircularProgress_Paint_Complete(t *testing.T) {
	c := NewCircularProgress(1.0)
	c.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	c.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '✓' { t.Errorf("complete = %q, want ✓", string(c0.Rune)) }
}

func TestP406_CircularProgress_Paint_ZeroBounds(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	c.Paint(buf)
}

func TestP406_CircularProgress_Concurrent(t *testing.T) {
	c := NewCircularProgress(0.5)
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { c.SetValue(0.6) }; close(done) }()
	for i := 0; i < 500; i++ { _ = c.Value() }
	<-done
}

func BenchmarkP406_CircularProgress_Paint(b *testing.B) {
	c := NewCircularProgress(0.65)
	c.SetLabel("CPU")
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { c.Paint(buf) }
}

// === TogglePill tests ===

func TestP406_TogglePill_New(t *testing.T) {
	tp := NewTogglePill(true)
	if !tp.IsOn() { t.Error("should be on") }
	if tp.OnText() != "ON" { t.Errorf("OnText = %q", tp.OnText()) }
	if tp.OffText() != "OFF" { t.Errorf("OffText = %q", tp.OffText()) }
	if tp.ID() == "" { t.Error("ID empty") }
}

func TestP406_TogglePill_Toggle(t *testing.T) {
	tp := NewTogglePill(false)
	tp.Toggle()
	if !tp.IsOn() { t.Error("should be on after toggle") }
	tp.Toggle()
	if tp.IsOn() { t.Error("should be off after 2nd toggle") }
}

func TestP406_TogglePill_SetOn(t *testing.T) {
	tp := NewTogglePill(false)
	tp.SetOn(true)
	if !tp.IsOn() { t.Error("should be on") }
}

func TestP406_TogglePill_SetLabel(t *testing.T) {
	tp := NewTogglePill(true)
	tp.SetLabel("Notifications")
	if tp.Label() != "Notifications" { t.Errorf("Label = %q", tp.Label()) }
}

func TestP406_TogglePill_Measure(t *testing.T) {
	tp := NewTogglePill(true)
	s := tp.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W != 5 { t.Errorf("W = %d, want 5", s.W) }
	tp.SetLabel("Dark Mode")
	s = tp.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W < 10 { t.Errorf("W = %d, too small", s.W) }
}

func TestP406_TogglePill_Paint_On(t *testing.T) {
	tp := NewTogglePill(true)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	tp.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '[' { t.Errorf("cell[0] = %q, want '['", string(c.Rune)) }
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'O' { t.Errorf("cell[1] = %q, want 'O'", string(c1.Rune)) }
}

func TestP406_TogglePill_Paint_Off(t *testing.T) {
	tp := NewTogglePill(false)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	tp.Paint(buf)
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'O' { t.Errorf("cell[1] = %q, want 'O'", string(c1.Rune)) }
}

func TestP406_TogglePill_Paint_ZeroBounds(t *testing.T) {
	tp := NewTogglePill(true)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tp.Paint(buf)
}

func TestP406_TogglePill_Concurrent(t *testing.T) {
	tp := NewTogglePill(true)
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { tp.Toggle() }; close(done) }()
	for i := 0; i < 500; i++ { _ = tp.IsOn() }
	<-done
}

func BenchmarkP406_TogglePill_Paint(b *testing.B) {
	tp := NewTogglePill(true)
	tp.SetLabel("Auto-save")
	tp.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { tp.Paint(buf) }
}

// === Skeleton tests ===

func TestP406_Skeleton_New(t *testing.T) {
	s := NewSkeleton(20, 2)
	if s.Width() != 20 { t.Errorf("Width = %d", s.Width()) }
	if s.Height() != 2 { t.Errorf("Height = %d", s.Height()) }
	if !s.Animate() { t.Error("should animate by default") }
	if s.ID() == "" { t.Error("ID empty") }
}

func TestP406_Skeleton_SetWidth(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.SetWidth(10)
	if s.Width() != 10 { t.Errorf("Width = %d", s.Width()) }
	s.SetWidth(0)
	if s.Width() != 1 { t.Errorf("Width = %d, want 1", s.Width()) }
}

func TestP406_Skeleton_SetHeight(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.SetHeight(3)
	if s.Height() != 3 { t.Errorf("Height = %d", s.Height()) }
}

func TestP406_Skeleton_SetAnimate(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.SetAnimate(false)
	if s.Animate() { t.Error("should be false") }
}

func TestP406_Skeleton_Tick(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.Tick() // should not panic
}

func TestP406_Skeleton_Measure(t *testing.T) {
	s := NewSkeleton(20, 3)
	sz := s.Measure(Constraints{MaxWidth: 10, MaxHeight: 2})
	if sz.W != 10 { t.Errorf("W = %d, want 10", sz.W) }
	if sz.H != 2 { t.Errorf("H = %d, want 2", sz.H) }
}

func TestP406_Skeleton_Paint(t *testing.T) {
	s := NewSkeleton(10, 2)
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	buf := buffer.NewBuffer(10, 2)
	s.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune == 0 { t.Error("cell should not be empty") }
}

func TestP406_Skeleton_Paint_Static(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.SetAnimate(false)
	s.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	s.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2591' { t.Errorf("static cell = %q, want ░", string(c.Rune)) }
}

func TestP406_Skeleton_Paint_ZeroBounds(t *testing.T) {
	s := NewSkeleton(5, 1)
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	s.Paint(buf)
}

func TestP406_Skeleton_Concurrent(t *testing.T) {
	s := NewSkeleton(5, 1)
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { s.Tick() }; close(done) }()
	for i := 0; i < 500; i++ { _ = s.Animate() }
	<-done
}

func TestP406_Skeleton_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Skeleton)(nil)
}

func TestP406_CircularProgress_SatisfiesComponent(t *testing.T) {
	var _ Component = (*CircularProgress)(nil)
}

func TestP406_TogglePill_SatisfiesComponent(t *testing.T) {
	var _ Component = (*TogglePill)(nil)
}

func BenchmarkP406_Skeleton_Paint(b *testing.B) {
	s := NewSkeleton(20, 2)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { s.Paint(buf) }
}

func TestP406_FormatPercent(t *testing.T) {
	if got := FormatPercent(0.5); got != "50%" { t.Errorf("FormatPercent(0.5) = %q", got) }
	if got := FormatPercent(1.0); got != "100%" { t.Errorf("FormatPercent(1.0) = %q", got) }
}
