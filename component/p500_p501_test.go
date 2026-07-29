package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownFootnoteBasic(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("Text[^1] here.\n\n[^1]: Definition")
	if mf.FootnoteCount() != 1 {
		t.Errorf("FootnoteCount = %d, want 1", mf.FootnoteCount())
	}
}

func TestMarkdownFootnoteMultiple(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("A[^1] B[^2]\n\n[^1]: First\n[^2]: Second")
	if mf.FootnoteCount() != 2 {
		t.Errorf("FootnoteCount = %d, want 2", mf.FootnoteCount())
	}
}

func TestMarkdownFootnoteNoFootnotes(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("Just text, no footnotes.")
	if mf.FootnoteCount() != 0 {
		t.Errorf("FootnoteCount = %d, want 0", mf.FootnoteCount())
	}
}

func TestMarkdownFootnoteEmpty(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("")
	if mf.FootnoteCount() != 0 {
		t.Errorf("FootnoteCount = %d, want 0", mf.FootnoteCount())
	}
}

func TestMarkdownFootnoteMeasure(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("Text[^1]\n[^1]: Def")
	s := mf.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownFootnotePaint(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("See[^1] for info.\n\n[^1]: https://example.com")
	mf.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	mf.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
}

func TestMarkdownFootnoteChildren(t *testing.T) {
	mf := NewMarkdownFootnote()
	if mf.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownFootnoteStyle(t *testing.T) {
	mf := NewMarkdownFootnote()
	mf.SetStyle(MarkdownFootnoteStyle{
		Text:       buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		RefMarker:  buffer.Style{Fg: buffer.RGB(0, 255, 0), Flags: buffer.Bold},
		Definition: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	mf.SetMarkdown("x[^1]\n[^1]: y")
	mf.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	mf.Paint(buf)
}

// ─── Legend tests ───

func TestLegendBasic(t *testing.T) {
	l := NewLegend()
	l.AddEntry("Revenue", buffer.RGB(34, 197, 94))
	l.AddEntry("Costs", buffer.RGB(239, 68, 68))
	if l.EntryCount() != 2 { t.Errorf("EntryCount = %d, want 2", l.EntryCount()) }
}

func TestLegendSetEntries(t *testing.T) {
	l := NewLegend()
	l.SetEntries([]LegendEntry{{Label: "A", Color: buffer.RGB(1, 2, 3)}, {Label: "B", Color: buffer.RGB(4, 5, 6)}})
	if l.EntryCount() != 2 { t.Errorf("EntryCount = %d, want 2", l.EntryCount()) }
}

func TestLegendClear(t *testing.T) {
	l := NewLegend()
	l.AddEntry("A", buffer.RGB(0, 0, 0))
	l.Clear()
	if l.EntryCount() != 0 { t.Errorf("EntryCount = %d, want 0", l.EntryCount()) }
}

func TestLegendEmpty(t *testing.T) {
	l := NewLegend()
	if l.EntryCount() != 0 { t.Errorf("EntryCount = %d, want 0", l.EntryCount()) }
}

func TestLegendMeasure(t *testing.T) {
	l := NewLegend()
	l.AddEntry("X", buffer.RGB(0, 0, 0))
	s := l.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestLegendPaint(t *testing.T) {
	l := NewLegend()
	l.AddEntry("Alpha", buffer.RGB(34, 197, 94))
	l.AddEntry("Beta", buffer.RGB(239, 68, 68))
	l.SetBounds(Rect{X: 0, Y: 0, W: 25, H: 5})
	buf := buffer.NewBuffer(25, 5)
	l.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	// Check color marker exists
	foundMarker := false
	for x := 0; x < 25; x++ {
		if buf.GetCell(x, 1).Rune == '█' { foundMarker = true; break }
	}
	if !foundMarker { t.Error("color marker not found") }
}

func TestLegendPaintEmpty(t *testing.T) {
	l := NewLegend()
	l.SetBounds(Rect{X: 0, Y: 0, W: 25, H: 3})
	buf := buffer.NewBuffer(25, 3)
	l.Paint(buf)
}

func TestLegendChildren(t *testing.T) {
	l := NewLegend()
	if l.Children() != nil { t.Error("Children should be nil") }
}

func TestLegendStyle(t *testing.T) {
	l := NewLegend()
	l.SetStyle(LegendStyle{Label: buffer.Style{Fg: buffer.RGB(255, 255, 255)}, Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)}})
	l.AddEntry("Test", buffer.RGB(255, 0, 0))
	l.SetBounds(Rect{X: 0, Y: 0, W: 25, H: 4})
	buf := buffer.NewBuffer(25, 4)
	l.Paint(buf)
}
