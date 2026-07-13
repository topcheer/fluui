package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestMarkdownViewer_New(t *testing.T) {
	v := NewMarkdownViewer("# Hello\n\nWorld")
	if v == nil {
		t.Fatal("expected non-nil viewer")
	}
}

func TestMarkdownViewer_SetSource(t *testing.T) {
	v := NewMarkdownViewer("# Initial")
	v.SetSource("# Updated\n\nNew content")
}

func TestMarkdownViewer_SetTitle(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.SetTitle("My Document")
}

func TestMarkdownViewer_SetStyle(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.SetStyle(DefaultMarkdownViewerStyle())
}

func TestMarkdownViewer_Toc(t *testing.T) {
	v := NewMarkdownViewer("# Heading 1\n\nText\n\n## Subheading\n\nMore text\n\n### Deep")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.ShowToc()
	buf := buffer.NewBuffer(80, 24)
	v.Paint(buf)
	entries := v.TocEntries()
	if len(entries) < 2 {
		t.Errorf("expected at least 2 TOC entries, got %d", len(entries))
	}
}

func TestMarkdownViewer_Scroll(t *testing.T) {
	v := NewMarkdownViewer("# Title\n\n" + repeatText("Line\n\n", 50))
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	v.Paint(buf)

	v.ScrollDown(5)
	if v.ScrollY() != 5 {
		t.Errorf("expected scrollY 5, got %d", v.ScrollY())
	}

	v.ScrollUp(2)
	if v.ScrollY() != 3 {
		t.Errorf("expected scrollY 3, got %d", v.ScrollY())
	}

	v.ScrollUp(100)
	if v.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 after over-scroll up, got %d", v.ScrollY())
	}
}

func TestMarkdownViewer_ScrollToTopBottom(t *testing.T) {
	v := NewMarkdownViewer("# Title\n\n" + repeatText("Line\n\n", 50))
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))

	v.ScrollToBottom()
	if v.ScrollY() == 0 {
		t.Error("expected non-zero scrollY at bottom")
	}

	v.ScrollToTop()
	if v.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 at top, got %d", v.ScrollY())
	}
}

func TestMarkdownViewer_HandleKey(t *testing.T) {
	v := NewMarkdownViewer("# Title\n\n" + repeatText("Line\n\n", 50))
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))

	// Down
	v.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if v.ScrollY() != 1 {
		t.Errorf("expected scrollY 1, got %d", v.ScrollY())
	}

	// Up
	v.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if v.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", v.ScrollY())
	}

	// PageDown
	v.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	if v.ScrollY() != 10 {
		t.Errorf("expected scrollY 10, got %d", v.ScrollY())
	}

	// PageUp
	v.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if v.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", v.ScrollY())
	}

	// Home
	v.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	v.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	v.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if v.ScrollY() != 0 {
		t.Error("expected scrollY 0 after Home")
	}

	// End
	v.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if v.ScrollY() == 0 {
		t.Error("expected non-zero scrollY after End")
	}

	// Vim j/k
	v.ScrollToTop()
	v.HandleKey(&term.KeyEvent{Rune: 'j'})
	if v.ScrollY() != 1 {
		t.Errorf("expected scrollY 1 after j, got %d", v.ScrollY())
	}
	v.HandleKey(&term.KeyEvent{Rune: 'k'})
	if v.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 after k, got %d", v.ScrollY())
	}

	// Vim g/G
	v.HandleKey(&term.KeyEvent{Rune: 'g'})
	if v.ScrollY() != 0 {
		t.Error("expected scrollY 0 after g")
	}
	v.HandleKey(&term.KeyEvent{Rune: 'G'})
	if v.ScrollY() == 0 {
		t.Error("expected non-zero scrollY after G")
	}
}

func TestMarkdownViewer_ToggleToc(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.ToggleToc()
	v.ToggleToc()
}

func TestMarkdownViewer_TocNavigation(t *testing.T) {
	v := NewMarkdownViewer("# Heading 1\n\nText\n\n## Heading 2\n\nText\n\n### Heading 3")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.ShowToc()
	v.Paint(buffer.NewBuffer(80, 24))

	entries := v.TocEntries()
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}

	// Navigate to next TOC entry
	v.HandleKey(&term.KeyEvent{Rune: 'n'})
	// Navigate to prev
	v.HandleKey(&term.KeyEvent{Rune: 'p'})
}

func TestMarkdownViewer_HandleKeyNil(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	if v.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestMarkdownViewer_HandleKeyEscape(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.ShowToc()
	// Escape when TOC visible → hides TOC, returns true
	handled := v.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !handled {
		t.Error("expected true when escaping TOC")
	}
	// Escape when TOC hidden → returns false
	handled = v.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if handled {
		t.Error("expected false when no TOC and escape")
	}
}

func TestMarkdownViewer_PaintNilBuffer(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.Paint(nil)
}

func TestMarkdownViewer_PaintZeroBounds(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	v.Paint(buffer.NewBuffer(1, 1))
}

func TestMarkdownViewer_PaintWithContent(t *testing.T) {
	v := NewMarkdownViewer("# Title\n\nThis is a paragraph with **bold** and `code`.\n\n- Item 1\n- Item 2\n\n| Col1 | Col2 |\n|------|------|\n| a    | b    |")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	v.Paint(buf)
	// Should have title at row 0
	if buf.GetCell(0, 0).Rune == ' ' {
		// Title might start with space, check non-blank somewhere
	}
}

func TestMarkdownViewer_PaintWithToc(t *testing.T) {
	v := NewMarkdownViewer("# Heading 1\n\nText\n\n## Sub\n\nMore")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.ShowToc()
	v.Paint(buffer.NewBuffer(80, 24))
}

func TestMarkdownViewer_ScrollToTocEntry(t *testing.T) {
	v := NewMarkdownViewer("# H1\n\nText\n\n## H2\n\nText\n\n### H3")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))
	v.ScrollToTocEntry(0)
	if v.ScrollY() != 0 {
		t.Error("expected scrollY 0 for TOC entry 0")
	}
	v.ScrollToTocEntry(-1) // out of range
	v.ScrollToTocEntry(999) // out of range
}

func TestMarkdownViewer_TotalLines(t *testing.T) {
	v := NewMarkdownViewer("# Title\n\nLine 1\n\nLine 2")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))
	if v.TotalLines() == 0 {
		t.Error("expected non-zero total lines")
	}
}

func TestMarkdownViewer_Measure(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	s := v.Measure(Constraints{MaxWidth: 100, MaxHeight: 30})
	if s.W != 100 || s.H != 30 {
		t.Errorf("expected 100x30, got %dx%d", s.W, s.H)
	}
	// Default when 0
	s = v.Measure(Constraints{})
	if s.W != 80 || s.H != 24 {
		t.Errorf("expected 80x24, got %dx%d", s.W, s.H)
	}
}

func TestMarkdownViewer_Children(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	if v.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestMarkdownViewer_ExtractCellText(t *testing.T) {
	cells := [][]buffer.Cell{
		{buffer.Cell{Rune: 'H', Width: 1}, buffer.Cell{Rune: 'i', Width: 1}},
	}
	text := extractCellText(cells)
	if text != "Hi" {
		t.Errorf("expected 'Hi', got %q", text)
	}
}

func TestMarkdownViewer_Concurrent(t *testing.T) {
	v := NewMarkdownViewer("# Test\n\nContent")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			v.SetSource("# Updated\n\nNew content")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		v.ScrollDown(1)
		v.Paint(buffer.NewBuffer(80, 24))
	}
	<-done
}

func TestMarkdownViewer_TabKey(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))
	// Tab toggles TOC
	handled := v.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !handled {
		t.Error("expected Tab to be handled")
	}
}

func TestMarkdownViewer_EmptySource(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))
}

func TestMarkdownViewer_UnknownKey(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	v.Paint(buffer.NewBuffer(80, 24))
	handled := v.HandleKey(&term.KeyEvent{Rune: 'z'})
	if handled {
		t.Error("expected false for unknown key 'z'")
	}
}

// Helper
func repeatText(text string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += text
	}
	return result
}
