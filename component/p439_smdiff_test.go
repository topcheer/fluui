package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStreamingMarkdownDiff_New_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	if !d.IsUnified() {
		t.Error("default should be unified")
	}
	if d.LineCount() != 0 {
		t.Errorf("LineCount = %d, want 0", d.LineCount())
	}
}

func TestStreamingMarkdownDiff_SetOldNew_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("line1\nline2")
	d.SetNew("line1\nline2\nline3")
	if d.OldText() != "line1\nline2" {
		t.Errorf("OldText = %q", d.OldText())
	}
	if d.NewText() != "line1\nline2\nline3" {
		t.Errorf("NewText = %q", d.NewText())
	}
}

func TestStreamingMarkdownDiff_Stats_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("a\nb\nc")
	d.SetNew("a\nx\nc\nd")
	added, removed, unchanged := d.Stats()
	if added != 2 {
		t.Errorf("added = %d, want 2", added)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
	if unchanged != 2 {
		t.Errorf("unchanged = %d, want 2", unchanged)
	}
}

func TestStreamingMarkdownDiff_DiffLines_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("hello\nworld")
	d.SetNew("hello\nGo")
	lines := d.DiffLines()
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	if lines[0].Type != DiffLineContext || lines[0].Content != "hello" {
		t.Errorf("line 0: type=%v content=%q", lines[0].Type, lines[0].Content)
	}
	if lines[1].Type != DiffLineRemoved || lines[1].Content != "world" {
		t.Errorf("line 1: type=%v content=%q", lines[1].Type, lines[1].Content)
	}
	if lines[2].Type != DiffLineAdded || lines[2].Content != "Go" {
		t.Errorf("line 2: type=%v content=%q", lines[2].Type, lines[2].Content)
	}
}

func TestStreamingMarkdownDiff_Empty_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	lines := d.DiffLines()
	if len(lines) != 0 {
		t.Errorf("empty diff should have 0 lines, got %d", len(lines))
	}
}

func TestStreamingMarkdownDiff_Identical_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("same\nlines")
	d.SetNew("same\nlines")
	added, removed, _ := d.Stats()
	if added != 0 || removed != 0 {
		t.Errorf("identical texts: added=%d removed=%d, want 0/0", added, removed)
	}
}

func TestStreamingMarkdownDiff_CacheRebuild_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("a\nb")
	d.SetNew("a\nc")
	lc1 := d.LineCount()
	d.SetNew("a\nc\nd\ne")
	lc2 := d.LineCount()
	if lc1 >= lc2 {
		t.Errorf("cache should rebuild: %d -> %d", lc1, lc2)
	}
}

func TestStreamingMarkdownDiff_UnifiedToggle_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetUnified(false)
	if d.IsUnified() {
		t.Error("should be false after SetUnified(false)")
	}
}

func TestStreamingMarkdownDiff_Style_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	st := DefaultStreamingMarkdownDiffStyle()
	d.SetStyle(st)
	if d.Style().Added.Fg != st.Added.Fg {
		t.Error("style mismatch")
	}
}

func TestStreamingMarkdownDiff_Measure_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("a\nb\nc")
	d.SetNew("a\nb\nc")
	sz := d.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
}

func TestStreamingMarkdownDiff_Paint_NoPanic_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("# Title\n\nOld paragraph.")
	d.SetNew("# Title\n\nNew **bold** paragraph.\n\nExtra line.")
	d.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	d.Paint(buf)
}

func TestStreamingMarkdownDiff_Paint_ZeroBounds_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("a")
	d.SetNew("b")
	d.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	d.Paint(buf)
}

func TestStreamingMarkdownDiff_Children_P439(t *testing.T) {
	if NewStreamingMarkdownDiff().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestStreamingMarkdownDiff_Fluent_P439(t *testing.T) {
	d := NewStreamingMarkdownDiff().SetOld("a").SetNew("b").SetUnified(false)
	if d.OldText() != "a" || d.NewText() != "b" || d.IsUnified() {
		t.Error("fluent chain failed")
	}
}

func BenchmarkStreamingMarkdownDiff_DiffCompute_P439(b *testing.B) {
	old := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"
	newT := "line1\nCHANGED\nline3\nline4\nline5\nline6\nADDED\nline7\nline8\nline9\nline10\nNEW"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeUnifiedDiff(old, newT)
	}
}

func BenchmarkStreamingMarkdownDiff_Paint_P439(b *testing.B) {
	d := NewStreamingMarkdownDiff()
	d.SetOld("# Title\n\nOld text here.\nLine 3.")
	d.SetNew("# Title\n\nNew **bold** text.\nLine 3.\nAdded line.")
	d.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Paint(buf)
	}
}
