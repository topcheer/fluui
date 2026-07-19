package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP297_NewCitationsBlock(t *testing.T) {
	cits := []Citation{
		{Index: 1, Title: "Go Docs", URL: "https://go.dev/doc", Snippet: "Go is a compiled language."},
		{Index: 2, Title: "Wikipedia", URL: "https://wikipedia.org", Snippet: "The free encyclopedia."},
	}
	cb := NewCitationsBlock(cits)
	if cb.Count() != 2 {
		t.Errorf("Count = %d, want 2", cb.Count())
	}
	if cb.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestP297_Citations(t *testing.T) {
	cb := NewCitationsBlock(nil)
	if cb.Count() != 0 {
		t.Errorf("Count = %d, want 0", cb.Count())
	}
	cb.AddCitation(Citation{Index: 1, Title: "Test", URL: "https://example.com"})
	cb.AddCitation(Citation{Index: 2, Title: "Test2", URL: "https://example2.com"})
	if cb.Count() != 2 {
		t.Errorf("Count = %d, want 2", cb.Count())
	}
	cits := cb.Citations()
	if len(cits) != 2 {
		t.Errorf("len(Citations) = %d", len(cits))
	}
	// verify it's a copy
	cits[0].Title = "modified"
	orig := cb.Citations()
	if orig[0].Title == "modified" {
		t.Error("Citations() should return a copy")
	}
}

func TestP297_AddCitation_AutoIndex(t *testing.T) {
	cb := NewCitationsBlock(nil)
	cb.AddCitation(Citation{Title: "No Index", URL: "https://x.com"})
	cits := cb.Citations()
	if cits[0].Index != 1 {
		t.Errorf("auto index = %d, want 1", cits[0].Index)
	}
}

func TestP297_SetCitations(t *testing.T) {
	cb := NewCitationsBlock([]Citation{{Index: 1, Title: "Old"}})
	cb.SetCitations([]Citation{
		{Index: 1, Title: "New1"},
		{Index: 2, Title: "New2"},
	})
	if cb.Count() != 2 {
		t.Errorf("Count = %d, want 2", cb.Count())
	}
}

func TestP297_Toggle(t *testing.T) {
	cb := NewCitationsBlock([]Citation{{Index: 1, Title: "T"}})
	if cb.Expanded() {
		t.Error("should start collapsed")
	}
	cb.Toggle()
	if !cb.Expanded() {
		t.Error("should be expanded")
	}
	cb.Toggle()
	if cb.Expanded() {
		t.Error("should be collapsed")
	}
}

func TestP297_SetExpanded(t *testing.T) {
	cb := NewCitationsBlock([]Citation{{Index: 1}})
	cb.SetExpanded(true)
	if !cb.Expanded() {
		t.Error("should be expanded")
	}
}

func TestP297_SetMaxSnippet(t *testing.T) {
	cb := NewCitationsBlock(nil)
	cb.SetMaxSnippet(50)
	cb.mu.Lock()
	if cb.maxSnippet != 50 {
		t.Errorf("maxSnippet = %d, want 50", cb.maxSnippet)
	}
	cb.mu.Unlock()
	// clamp
	cb.SetMaxSnippet(0)
	cb.mu.Lock()
	if cb.maxSnippet != 1 {
		t.Errorf("maxSnippet = %d, want 1 (clamped)", cb.maxSnippet)
	}
	cb.mu.Unlock()
}

func TestP297_Measure_Empty(t *testing.T) {
	cb := NewCitationsBlock(nil)
	s := cb.Measure(Constraints{MaxWidth: 80})
	if s.H != 0 {
		t.Errorf("empty H = %d, want 0", s.H)
	}
}

func TestP297_Measure_Collapsed(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "A", URL: "https://a.com"},
		{Index: 2, Title: "B", URL: "https://b.com"},
	})
	s := cb.Measure(Constraints{MaxWidth: 80})
	if s.H != 1 {
		t.Errorf("collapsed H = %d, want 1", s.H)
	}
}

func TestP297_Measure_Expanded(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "A", URL: "u", Snippet: "s"},
		{Index: 2, Title: "B", URL: "u", Snippet: "s"},
	})
	cb.SetExpanded(true)
	s := cb.Measure(Constraints{MaxWidth: 80})
	// header(1) + 2 citations × 3 lines = 7
	if s.H != 7 {
		t.Errorf("expanded H = %d, want 7", s.H)
	}
}

func TestP297_Measure_DefaultWidth(t *testing.T) {
	cb := NewCitationsBlock([]Citation{{Index: 1, Title: "T"}})
	s := cb.Measure(Constraints{})
	if s.W != 80 {
		t.Errorf("default W = %d, want 80", s.W)
	}
}

func TestP297_Paint_Collapsed(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Go", URL: "https://go.dev", Snippet: "Go docs"},
		{Index: 2, Title: "Rust", URL: "https://rust-lang.org", Snippet: "Rust docs"},
	})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	cb.Paint(buf)
}

func TestP297_Paint_Expanded(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Go Documentation", URL: "https://go.dev/doc", Snippet: "Go is a statically typed, compiled programming language."},
		{Index: 2, Title: "Wikipedia", URL: "https://wikipedia.org", Snippet: "The free encyclopedia that anyone can edit."},
	})
	cb.SetExpanded(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cb.Paint(buf)
}

func TestP297_Paint_ExpandedWithSnippet(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Long Title Here", URL: "https://example.com/very/long/path", Snippet: "This is a snippet that might be longer than the available width and should get truncated"},
	})
	cb.SetExpanded(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP297_Paint_Empty(t *testing.T) {
	cb := NewCitationsBlock(nil)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	cb.Paint(buf) // should not panic
}

func TestP297_Paint_ZeroBounds(t *testing.T) {
	cb := NewCitationsBlock([]Citation{{Index: 1, Title: "T"}})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf) // should not panic
}

func TestP297_Paint_NonZeroOffset(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Test", URL: "https://test.com", Snippet: "Snippet"},
	})
	cb.SetExpanded(true)
	cb.SetBounds(Rect{X: 5, Y: 3, W: 50, H: 5})
	buf := buffer.NewBuffer(60, 10)
	cb.Paint(buf)
}

func TestP297_Paint_NarrowWidth(t *testing.T) {
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Very Long Title", URL: "https://example.com"},
	})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	cb.Paint(buf)
}

func TestP297_TruncateStr(t *testing.T) {
	tests := []struct {
		input    string
		maxRunes int
		want     string
	}{
		{"hello", 10, "hello"},
		{"hello", 3, "he…"},
		{"hi", 3, "hi"},
		{"ab", 1, "…"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncateStr(tt.input, tt.maxRunes)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.maxRunes, got, tt.want)
		}
	}
}

func TestP297_Concurrent(t *testing.T) {
	cb := NewCitationsBlock(nil)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			cb.AddCitation(Citation{Index: i + 1, Title: "T"})
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		_ = cb.Count()
		_ = cb.Citations()
	}
	<-done
}

func TestP297_SatisfiesComponent(t *testing.T) {
	var _ Component = (*CitationsBlock)(nil)
}
