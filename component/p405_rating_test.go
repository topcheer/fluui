package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === Rating tests ===

func TestP405_NewRating(t *testing.T) {
	r := NewRating(3.5, 5)
	if r.Value() != 3.5 { t.Errorf("Value = %v", r.Value()) }
	if r.Max() != 5 { t.Errorf("Max = %d", r.Max()) }
	if r.ID() == "" { t.Error("ID empty") }
}

func TestP405_Rating_SetValue(t *testing.T) {
	r := NewRating(0, 5)
	r.SetValue(4.2)
	if r.Value() != 4.2 { t.Errorf("Value = %v", r.Value()) }
	r.SetValue(10) // clamped to max
	if r.Value() != 5 { t.Errorf("Value = %v, want 5", r.Value()) }
	r.SetValue(-1) // clamped to 0
	if r.Value() != 0 { t.Errorf("Value = %v, want 0", r.Value()) }
}

func TestP405_Rating_SetMax(t *testing.T) {
	r := NewRating(3, 5)
	r.SetMax(10)
	if r.Max() != 10 { t.Errorf("Max = %d", r.Max()) }
	r.SetMax(0)
	if r.Max() != 1 { t.Errorf("Max = %d, want 1 (clamped)", r.Max()) }
}

func TestP405_Rating_SetShowNumber(t *testing.T) {
	r := NewRating(4, 5)
	r.SetShowNumber(true)
	if !r.ShowNumber() { t.Error("should be true") }
}

func TestP405_Rating_SetStars(t *testing.T) {
	r := NewRating(3, 5)
	r.SetStars('A', 'B', 'C')
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	r.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'A' { t.Error("filled star should be 'A'") }
	if buf.GetCell(3, 0).Rune != 'C' { t.Error("empty star should be 'C'") }
}

func TestP405_Rating_Measure(t *testing.T) {
	r := NewRating(3, 5)
	s := r.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W != 5 { t.Errorf("W = %d, want 5", s.W) }
	if s.H != 1 { t.Errorf("H = %d", s.H) }

	r.SetShowNumber(true)
	s = r.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W != 8 { t.Errorf("W = %d, want 8", s.W) }
}

func TestP405_Rating_Paint_Full(t *testing.T) {
	r := NewRating(5, 5)
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	r.Paint(buf)
	for i := 0; i < 5; i++ {
		c := buf.GetCell(i, 0)
		if c.Rune != '\u2605' { t.Errorf("cell[%d] = %q, want ★", i, string(c.Rune)) }
	}
}

func TestP405_Rating_Paint_Half(t *testing.T) {
	r := NewRating(2.5, 5)
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	r.Paint(buf)
	// First 2 stars filled, 3rd half (>=0.25 but <0.75)
	c2 := buf.GetCell(2, 0)
	if c2.Rune != '\u2606' { t.Errorf("half cell = %q", string(c2.Rune)) }
}

func TestP405_Rating_Paint_Empty(t *testing.T) {
	r := NewRating(0, 5)
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	r.Paint(buf)
	for i := 0; i < 5; i++ {
		c := buf.GetCell(i, 0)
		if c.Rune != '\u2606' { t.Errorf("cell[%d] = %q, want ☆", i, string(c.Rune)) }
	}
}

func TestP405_Rating_Paint_WithNumber(t *testing.T) {
	r := NewRating(3.5, 5)
	r.SetShowNumber(true)
	r.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	r.Paint(buf)
	c := buf.GetCell(5, 0) // space after 5 stars
	if c.Rune != ' ' { t.Errorf("space cell = %q", string(c.Rune)) }
}

func TestP405_Rating_Paint_ZeroBounds(t *testing.T) {
	r := NewRating(3, 5)
	r.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	r.Paint(buf)
}

func TestP405_Rating_Paint_NarrowWidth(t *testing.T) {
	r := NewRating(3, 5)
	r.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	r.Paint(buf) // should clip
}

func TestP405_Rating_Paint_NonZeroOffset(t *testing.T) {
	r := NewRating(4, 5)
	r.SetBounds(Rect{X: 10, Y: 5, W: 10, H: 1})
	buf := buffer.NewBuffer(25, 10)
	r.Paint(buf)
	c := buf.GetCell(10, 5)
	if c.Rune != '\u2605' { t.Errorf("offset cell = %q", string(c.Rune)) }
}

func TestP405_Rating_Concurrent(t *testing.T) {
	r := NewRating(3, 5)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { r.SetValue(4) }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = r.Value() }
	<-done
}

func TestP405_Rating_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Rating)(nil)
}

func BenchmarkP405_Rating_Paint(b *testing.B) {
	r := NewRating(3.5, 5)
	r.SetShowNumber(true)
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { r.Paint(buf) }
}

// === Coverage: Badge Measure 80%, Accordion Measure 89.5%, Breadcrumb Paint 83.3% ===

func TestP405_Badge_Measure_Edges(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	// Zero constraints
	s := b.Measure(Constraints{})
	if s.W < 1 || s.H < 1 { t.Error("should be >= 1") }
	// Normal
	s = b.Measure(Constraints{MaxWidth: 20, MaxHeight: 5})
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP405_Accordion_Measure_Edges(t *testing.T) {
	a := NewAccordion([]AccordionItem{{Title: "Item 1", Content: "content"}})
	s := a.Measure(Constraints{})
	if s.H < 1 { t.Error("H should be >= 1") }
}

func TestP405_Breadcrumb_Paint_SingleItem(t *testing.T) {
	b := NewBreadcrumb([]string{"Home"})
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'H' { t.Errorf("cell = %q, want 'H'", string(c.Rune)) }
}

func TestP405_Breadcrumb_Paint_Empty(t *testing.T) {
	b := NewBreadcrumb(nil)
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf) // should not panic
}
