package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownDefinitionListBasic(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("Go\n: A compiled language")
	if dl.TermCount() != 1 { t.Errorf("TermCount = %d, want 1", dl.TermCount()) }
}

func TestMarkdownDefinitionListMultiple(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("Go\n: Compiled\nRust\n: Systems\nPython\n: Interpreted")
	if dl.TermCount() != 3 { t.Errorf("TermCount = %d, want 3", dl.TermCount()) }
}

func TestMarkdownDefinitionListNoDefs(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("Just text without definitions")
	if dl.TermCount() != 0 { t.Errorf("TermCount = %d, want 0", dl.TermCount()) }
}

func TestMarkdownDefinitionListEmpty(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("")
	if dl.TermCount() != 0 { t.Errorf("TermCount = %d, want 0", dl.TermCount()) }
}

func TestMarkdownDefinitionListMeasure(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("A\n: B")
	s := dl.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownDefinitionListPaint(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("Go\n: A compiled language")
	dl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	dl.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundTerm := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'G' { foundTerm = true; break }
	}
	if !foundTerm { t.Error("term text not found") }
}

func TestMarkdownDefinitionListChildren(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	if dl.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownDefinitionListStyle(t *testing.T) {
	dl := NewMarkdownDefinitionList()
	dl.SetStyle(DefinitionListStyle{
		Term: buffer.Style{Fg: buffer.RGB(255, 0, 255), Flags: buffer.Bold},
		Definition: buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	dl.SetMarkdown("X\n: Y")
	dl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	dl.Paint(buf)
}

// ─── StatusBarSegment tests ───

func TestStatusBarSegmentBasic(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.AddSegment("Branch", "main", buffer.RGB(255, 255, 255), buffer.RGB(34, 197, 94))
	if sb.SegmentCount() != 1 { t.Errorf("SegmentCount = %d, want 1", sb.SegmentCount()) }
}

func TestStatusBarSegmentMultiple(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.AddSegment("A", "1", buffer.RGB(0, 0, 0), buffer.RGB(255, 0, 0))
	sb.AddSegment("B", "2", buffer.RGB(0, 0, 0), buffer.RGB(0, 255, 0))
	sb.AddSegment("C", "3", buffer.RGB(0, 0, 0), buffer.RGB(0, 0, 255))
	if sb.SegmentCount() != 3 { t.Errorf("SegmentCount = %d, want 3", sb.SegmentCount()) }
}

func TestStatusBarSegmentClear(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.AddSegment("A", "1", buffer.RGB(0, 0, 0), buffer.RGB(0, 0, 0))
	sb.Clear()
	if sb.SegmentCount() != 0 { t.Errorf("SegmentCount = %d, want 0", sb.SegmentCount()) }
}

func TestStatusBarSegmentEmpty(t *testing.T) {
	sb := NewStatusBarSegment()
	if sb.SegmentCount() != 0 { t.Errorf("SegmentCount = %d, want 0", sb.SegmentCount()) }
}

func TestStatusBarSegmentMeasure(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.AddSegment("X", "Y", buffer.RGB(0, 0, 0), buffer.RGB(0, 0, 0))
	s := sb.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestStatusBarSegmentPaint(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.AddSegment("Branch", "main", buffer.RGB(255, 255, 255), buffer.RGB(34, 197, 94))
	sb.AddSegment("Errors", "0", buffer.RGB(255, 255, 255), buffer.RGB(239, 68, 68))
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	sb.Paint(buf)
	foundLabel := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == 'B' { foundLabel = true; break }
	}
	if !foundLabel { t.Error("segment label not found") }
}

func TestStatusBarSegmentPaintEmpty(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	sb.Paint(buf)
}

func TestStatusBarSegmentChildren(t *testing.T) {
	sb := NewStatusBarSegment()
	if sb.Children() != nil { t.Error("Children should be nil") }
}

func TestStatusBarSegmentStyle(t *testing.T) {
	sb := NewStatusBarSegment()
	sb.SetStyle(StatusBarSegmentStyle{Separator: buffer.Style{Fg: buffer.RGB(100, 100, 100)}, Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)}})
	sb.AddSegment("X", "1", buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 255))
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	sb.Paint(buf)
}
