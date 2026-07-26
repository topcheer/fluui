package component

import (
	"testing"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP402_NewSearchBar(t *testing.T) {
	s := NewSearchBar("Search...")
	if s.Placeholder() != "Search..." { t.Errorf("Placeholder = %q", s.Placeholder()) }
	if s.Query() != "" { t.Error("Query should be empty initially") }
	if s.ID() == "" { t.Error("ID empty") }
}

func TestP402_SetQuery(t *testing.T) {
	s := NewSearchBar("Search")
	s.SetQuery("hello")
	if s.Query() != "hello" { t.Errorf("Query = %q", s.Query()) }
}

func TestP402_Clear(t *testing.T) {
	s := NewSearchBar("Search")
	s.SetQuery("test")
	s.Clear()
	if s.Query() != "" { t.Error("Query should be empty after Clear") }
}

func TestP402_SetFocus(t *testing.T) {
	s := NewSearchBar("Search")
	s.SetFocus(true)
	if !s.Focused() { t.Error("should be focused") }
	s.SetFocus(false)
	if s.Focused() { t.Error("should not be focused") }
}

func TestP402_Measure(t *testing.T) {
	s := NewSearchBar("Search")
	sz := s.Measure(Constraints{MaxWidth: 50, MaxHeight: 5})
	if sz.H != 1 { t.Errorf("H = %d", sz.H) }
	if sz.W != 50 { t.Errorf("W = %d", sz.W) }
}

func TestP402_Paint_WithQuery(t *testing.T) {
	s := NewSearchBar("Search...")
	s.SetQuery("hello")
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
	// Icon at 0, query text after
	c := buf.GetCell(2, 0)
	if c.Rune != 'h' { t.Errorf("query cell = %q, want 'h'", string(c.Rune)) }
}

func TestP402_Paint_Placeholder(t *testing.T) {
	s := NewSearchBar("Type to search")
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
	c := buf.GetCell(2, 0)
	if c.Rune != 'T' { t.Errorf("placeholder cell = %q, want 'T'", string(c.Rune)) }
}

func TestP402_Paint_FocusedCursor(t *testing.T) {
	s := NewSearchBar("Search")
	s.SetQuery("hi")
	s.SetFocus(true)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
	// Cursor at position after "⌕ hi" = position 4
	c := buf.GetCell(4, 0)
	if c.Rune != '\u2588' { t.Errorf("cursor cell = %q, want █", string(c.Rune)) }
}

func TestP402_Paint_ZeroBounds(t *testing.T) {
	s := NewSearchBar("Search")
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	s.Paint(buf)
}

func TestP402_Concurrent(t *testing.T) {
	s := NewSearchBar("Search")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { s.SetQuery("concurrent") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = s.Query() }
	<-done
}

func TestP402_SatisfiesComponent(t *testing.T) {
	var _ Component = (*SearchBar)(nil)
}

func BenchmarkP402_SearchBar_Paint(b *testing.B) {
	s := NewSearchBar("Search files...")
	s.SetQuery("main.go")
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { s.Paint(buf) }
}
