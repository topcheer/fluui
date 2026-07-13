package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Header tests ===

func TestHeader_New(t *testing.T) {
	h := NewHeader("MyApp")
	if h.Title() != "MyApp" {
		t.Errorf("expected 'MyApp', got %q", h.Title())
	}
}

func TestHeader_SetTitle(t *testing.T) {
	h := NewHeader("")
	h.SetTitle("New Title")
	if h.Title() != "New Title" {
		t.Errorf("expected 'New Title', got %q", h.Title())
	}
}

func TestHeader_SetSubtitle(t *testing.T) {
	h := NewHeader("App")
	h.SetSubtitle("v1.0")
	if h.Subtitle() != "v1.0" {
		t.Errorf("expected 'v1.0', got %q", h.Subtitle())
	}
}

func TestHeader_Measure(t *testing.T) {
	h := NewHeader("App")
	s := h.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
}

func TestHeader_Paint(t *testing.T) {
	h := NewHeader("MyApp")
	h.SetSubtitle("v2.0")
	h.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	h.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'M' {
		t.Error("expected 'M' at start of title")
	}
}

func TestHeader_PaintNilBuffer(t *testing.T) {
	h := NewHeader("App")
	h.Paint(nil)
}

func TestHeader_PaintZeroWidth(t *testing.T) {
	h := NewHeader("App")
	h.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 1})
	buf := buffer.NewBuffer(1, 1)
	h.Paint(buf)
}

func TestHeader_PaintLongTitle(t *testing.T) {
	h := NewHeader("This is a very long title that exceeds width")
	h.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	h.Paint(buf)
}

func TestHeader_HandleKey(t *testing.T) {
	h := NewHeader("App")
	if h.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected false")
	}
}

func TestHeader_Children(t *testing.T) {
	h := NewHeader("App")
	if h.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestHeader_SetStyle(t *testing.T) {
	h := NewHeader("App")
	h.SetStyle(DefaultHeaderStyle())
	h.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	h.Paint(buffer.NewBuffer(20, 1))
}

func TestHeader_Concurrent(t *testing.T) {
	h := NewHeader("App")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			h.SetTitle("App")
			h.SetSubtitle("v1")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		h.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
		h.Paint(buffer.NewBuffer(20, 1))
	}
	<-done
}

// === Footer tests ===

func TestFooter_New(t *testing.T) {
	f := NewFooter()
	if len(f.Hints()) != 0 {
		t.Error("expected 0 hints")
	}
}

func TestFooter_SetHints(t *testing.T) {
	f := NewFooter()
	f.SetHints([]FooterHint{{Keys: "Ctrl+Q", Description: "Quit"}})
	if len(f.Hints()) != 1 {
		t.Errorf("expected 1 hint, got %d", len(f.Hints()))
	}
}

func TestFooter_AddHint(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+S", "Save")
	f.AddHint("Ctrl+Q", "Quit")
	if len(f.Hints()) != 2 {
		t.Errorf("expected 2 hints, got %d", len(f.Hints()))
	}
}

func TestFooter_ClearHints(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+S", "Save")
	f.ClearHints()
	if len(f.Hints()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestFooter_Measure(t *testing.T) {
	f := NewFooter()
	s := f.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
}

func TestFooter_Paint(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+Q", "Quit")
	f.AddHint("Ctrl+S", "Save")
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	f.Paint(buf)
	// Should have 'C' at start
	if buf.GetCell(0, 0).Rune != 'C' {
		t.Error("expected 'C' at start")
	}
}

func TestFooter_PaintNilBuffer(t *testing.T) {
	f := NewFooter()
	f.Paint(nil)
}

func TestFooter_PaintZeroWidth(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+Q", "Quit")
	f.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 1})
	f.Paint(buffer.NewBuffer(1, 1))
}

func TestFooter_PaintOverflow(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+Shift+Alt+Q", "A very long description that won't fit")
	f.AddHint("F12", "More")
	f.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	f.Paint(buf) // should not panic
}

func TestFooter_HandleKey(t *testing.T) {
	f := NewFooter()
	if f.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected false")
	}
}

func TestFooter_Children(t *testing.T) {
	f := NewFooter()
	if f.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestFooter_SetStyle(t *testing.T) {
	f := NewFooter()
	f.SetStyle(DefaultFooterStyle())
	f.AddHint("Ctrl+Q", "Quit")
	f.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	f.Paint(buffer.NewBuffer(20, 1))
}

func TestFooter_Concurrent(t *testing.T) {
	f := NewFooter()
	f.AddHint("Ctrl+Q", "Quit")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			f.AddHint("Ctrl+S", "Save")
			f.ClearHints()
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		f.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
		f.Paint(buffer.NewBuffer(30, 1))
	}
	<-done
}
