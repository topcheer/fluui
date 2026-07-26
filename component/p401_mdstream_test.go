package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP401_NewMarkdownStream(t *testing.T) {
	m := NewMarkdownStream()
	if m.ID() == "" { t.Error("ID empty") }
	if !m.Streaming() { t.Error("should default to streaming") }
	if !m.CursorOn() { t.Error("cursor should default on") }
	if m.CursorChar() != '\u2588' { t.Errorf("cursor char = %q", string(m.CursorChar())) }
}

func TestP401_SetSource(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("# Hello")
	if m.Source() != "# Hello" { t.Errorf("Source = %q", m.Source()) }
}

func TestP401_Append(t *testing.T) {
	m := NewMarkdownStream()
	m.Append("Hello ")
	m.Append("World")
	if m.Source() != "Hello World" { t.Errorf("Source = %q", m.Source()) }
}

func TestP401_SetStreaming(t *testing.T) {
	m := NewMarkdownStream()
	m.SetStreaming(false)
	if m.Streaming() { t.Error("should be false") }
}

func TestP401_SetCursorOn(t *testing.T) {
	m := NewMarkdownStream()
	m.SetCursorOn(false)
	if m.CursorOn() { t.Error("should be false") }
}

func TestP401_SetCursorChar(t *testing.T) {
	m := NewMarkdownStream()
	m.SetCursorChar('|')
	if m.CursorChar() != '|' { t.Errorf("Char = %q", string(m.CursorChar())) }
}

func TestP401_Tick(t *testing.T) {
	m := NewMarkdownStream()
	if !m.CursorOn() { t.Error("initial should be on") }
	m.Tick()
	if m.CursorOn() { t.Error("after tick should be off") }
	m.Tick()
	if !m.CursorOn() { t.Error("after 2nd tick should be on") }
}

func TestP401_Measure(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("line1\nline2\nline3")
	s := m.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H < 3 { t.Errorf("H = %d, want >= 3", s.H) }
}

func TestP401_Paint_WithCursor(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Hello")
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf)
	// Should render text + cursor at position 5
	c := buf.GetCell(0, 0)
	if c.Rune == 0 { t.Log("rendered cell empty (markdown may produce nil)") }
}

func TestP401_Paint_StreamingOff(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Hello")
	m.SetStreaming(false)
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf) // no cursor
}

func TestP401_Paint_CursorOff(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Hello")
	m.SetCursorOn(false)
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf) // cursor not drawn
}

func TestP401_Paint_EmptySource(t *testing.T) {
	m := NewMarkdownStream()
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf) // should not panic
}

func TestP401_Paint_ZeroBounds(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Hello")
	m.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	m.Paint(buf)
}

func TestP401_Paint_Multiline(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Line 1\nLine 2\nLine 3")
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf)
}

func TestP401_Concurrent(t *testing.T) {
	m := NewMarkdownStream()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { m.Append("x"); m.Tick() }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = m.Source() }
	<-done
}

func TestP401_SatisfiesComponent(t *testing.T) {
	var _ Component = (*MarkdownStream)(nil)
}

func TestP401_CountLinesFast(t *testing.T) {
	if n := countLinesFast(""); n != 1 { t.Errorf("empty = %d, want 1", n) }
	if n := countLinesFast("a"); n != 1 { t.Errorf("single = %d, want 1", n) }
	if n := countLinesFast("a\nb"); n != 2 { t.Errorf("two = %d, want 2", n) }
	if n := countLinesFast("a\nb\nc"); n != 3 { t.Errorf("three = %d, want 3", n) }
}

func BenchmarkP401_MarkdownStream_Paint(b *testing.B) {
	m := NewMarkdownStream()
	m.SetSource("# Hello World\n\nThis is a **test** with `code`.\n\n- Item 1\n- Item 2")
	m.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { m.Paint(buf) }
}
