package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCarouselBasic(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("Title 1", "Content 1")
	c.AddSlide("Title 2", "Content 2")

	if c.SlideCount() != 2 {
		t.Errorf("SlideCount = %d, want 2", c.SlideCount())
	}
	if c.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex = %d, want 0", c.CurrentIndex())
	}
}

func TestCarouselNext(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")
	c.AddSlide("C", "c")

	c.Next()
	if c.CurrentIndex() != 1 {
		t.Errorf("After next: CurrentIndex = %d, want 1", c.CurrentIndex())
	}
	c.Next()
	if c.CurrentIndex() != 2 {
		t.Errorf("After next: CurrentIndex = %d, want 2", c.CurrentIndex())
	}
}

func TestCarouselWrapNext(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")

	c.Next() // 0→1
	c.Next() // 1→0 (wrap)
	if c.CurrentIndex() != 0 {
		t.Errorf("Wrap next: CurrentIndex = %d, want 0", c.CurrentIndex())
	}
}

func TestCarouselPrev(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")

	c.Next() // 0→1
	c.Prev() // 1→0
	if c.CurrentIndex() != 0 {
		t.Errorf("After prev: CurrentIndex = %d, want 0", c.CurrentIndex())
	}
}

func TestCarouselWrapPrev(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")
	c.AddSlide("C", "c")

	c.Prev() // 0→2 (wrap backwards)
	if c.CurrentIndex() != 2 {
		t.Errorf("Wrap prev: CurrentIndex = %d, want 2", c.CurrentIndex())
	}
}

func TestCarouselSetCurrent(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")
	c.AddSlide("C", "c")

	c.SetCurrent(2)
	if c.CurrentIndex() != 2 {
		t.Errorf("SetCurrent(2): CurrentIndex = %d, want 2", c.CurrentIndex())
	}
	c.SetCurrent(-5)
	if c.CurrentIndex() != 0 {
		t.Errorf("SetCurrent(-5): CurrentIndex = %d, want 0", c.CurrentIndex())
	}
	c.SetCurrent(100)
	if c.CurrentIndex() != 2 {
		t.Errorf("SetCurrent(100): CurrentIndex = %d, want 2 (clamped)", c.CurrentIndex())
	}
}

func TestCarouselEmpty(t *testing.T) {
	c := NewCarousel()
	if c.SlideCount() != 0 {
		t.Errorf("SlideCount = %d, want 0", c.SlideCount())
	}
	c.Next() // should not panic
	c.Prev() // should not panic
}

func TestCarouselMeasure(t *testing.T) {
	c := NewCarousel()
	s := c.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestCarouselPaint(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("Welcome", "Get started!")
	c.AddSlide("Features", "160+ components")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})

	buf := buffer.NewBuffer(40, 8)
	c.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Title text
	foundTitle := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 1).Rune == 'W' {
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Error("title text not found")
	}
}

func TestCarouselPaintDots(t *testing.T) {
	c := NewCarousel()
	c.AddSlide("A", "a")
	c.AddSlide("B", "b")
	c.AddSlide("C", "c")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})

	buf := buffer.NewBuffer(40, 8)
	c.Paint(buf)

	// Find navigation dots on row 6
	activeDotCount := 0
	inactiveDotCount := 0
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 6)
		if cell.Rune == '●' {
			activeDotCount++
		}
		if cell.Rune == '○' {
			inactiveDotCount++
		}
	}
	if activeDotCount != 1 {
		t.Errorf("active dots = %d, want 1", activeDotCount)
	}
	if inactiveDotCount != 2 {
		t.Errorf("inactive dots = %d, want 2", inactiveDotCount)
	}
}

func TestCarouselPaintEmpty(t *testing.T) {
	c := NewCarousel()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	c.Paint(buf) // should not panic
}

func TestCarouselChildren(t *testing.T) {
	c := NewCarousel()
	if c.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestCarouselStyle(t *testing.T) {
	c := NewCarousel()
	c.SetStyle(CarouselStyle{
		Title:   buffer.Style{Fg: buffer.RGB(255, 0, 255), Flags: buffer.Bold},
		Content: buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Dots:    buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Active:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Nav:     buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Border:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	c.AddSlide("Test", "Content")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	c.Paint(buf)
}
