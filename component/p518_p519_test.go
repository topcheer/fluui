package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownSubscriptBasic(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("H~2~O")
	if ms.SubscriptCount() != 1 { t.Errorf("SubscriptCount = %d, want 1", ms.SubscriptCount()) }
}

func TestMarkdownSubscriptMultiple(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("H~2~O and CO~2~")
	if ms.SubscriptCount() != 2 { t.Errorf("SubscriptCount = %d, want 2", ms.SubscriptCount()) }
}

func TestMarkdownSubscriptNoMarker(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("Just text")
	if ms.SubscriptCount() != 0 { t.Errorf("SubscriptCount = %d, want 0", ms.SubscriptCount()) }
}

func TestMarkdownSubscriptEmpty(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("")
	if ms.SubscriptCount() != 0 { t.Errorf("SubscriptCount = %d, want 0", ms.SubscriptCount()) }
}

func TestMarkdownSubscriptUnclosed(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("Unclosed ~tilde")
	if ms.SubscriptCount() != 0 { t.Errorf("SubscriptCount = %d, want 0 (unclosed)", ms.SubscriptCount()) }
}

func TestMarkdownSubscriptMeasure(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("x~2~")
	s := ms.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownSubscriptPaint(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("H~2~O formula")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	ms.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'H' { foundText = true; break }
	}
	if !foundText { t.Error("text not found") }
}

func TestMarkdownSubscriptChildren(t *testing.T) {
	ms := NewMarkdownSubscript()
	if ms.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownSubscriptStyle(t *testing.T) {
	ms := NewMarkdownSubscript()
	ms.SetStyle(SubscriptStyle{Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Subscript: buffer.Style{Fg: buffer.RGB(0,255,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ms.SetMarkdown("x~2~")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	ms.Paint(buf)
}

// ─── MultiSelect tests ───

func TestMultiSelectBasic(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.AddOption("C")
	if ms.OptionCount() != 3 { t.Errorf("OptionCount = %d, want 3", ms.OptionCount()) }
}

func TestMultiSelectToggle(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.Toggle(0)
	if ms.SelectedCount() != 1 { t.Errorf("SelectedCount = %d, want 1", ms.SelectedCount()) }
	ms.Toggle(0)
	if ms.SelectedCount() != 0 { t.Errorf("SelectedCount = %d, want 0 after untoggle", ms.SelectedCount()) }
}

func TestMultiSelectSelectAll(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.AddOption("C")
	ms.SelectAll()
	if ms.SelectedCount() != 3 { t.Errorf("SelectedCount = %d, want 3", ms.SelectedCount()) }
}

func TestMultiSelectDeselectAll(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.SelectAll()
	ms.DeselectAll()
	if ms.SelectedCount() != 0 { t.Errorf("SelectedCount = %d, want 0", ms.SelectedCount()) }
}

func TestMultiSelectSelectedIndices(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.AddOption("C")
	ms.Toggle(0)
	ms.Toggle(2)
	idxs := ms.SelectedIndices()
	if len(idxs) != 2 || idxs[0] != 0 || idxs[1] != 2 { t.Errorf("SelectedIndices = %v", idxs) }
}

func TestMultiSelectCursor(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	ms.AddOption("B")
	ms.AddOption("C")
	ms.MoveCursorDown()
	if ms.Cursor() != 1 { t.Errorf("Cursor = %d, want 1", ms.Cursor()) }
	ms.MoveCursorDown()
	ms.MoveCursorDown() // at end, shouldn't go past
	if ms.Cursor() != 2 { t.Errorf("Cursor = %d, want 2 (clamped)", ms.Cursor()) }
	ms.MoveCursorUp()
	if ms.Cursor() != 1 { t.Errorf("Cursor = %d, want 1", ms.Cursor()) }
}

func TestMultiSelectEmpty(t *testing.T) {
	ms := NewMultiSelect()
	if ms.OptionCount() != 0 { t.Errorf("OptionCount = %d, want 0", ms.OptionCount()) }
}

func TestMultiSelectMeasure(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("A")
	s := ms.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMultiSelectPaint(t *testing.T) {
	ms := NewMultiSelect()
	ms.AddOption("Apple")
	ms.AddOption("Banana")
	ms.Toggle(0)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ms.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundChecked := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 30; x++ {
			if buf.GetCell(x, y).Rune == '☑' { foundChecked = true; break }
		}
	}
	if !foundChecked { t.Error("checked checkbox not found") }
}

func TestMultiSelectPaintEmpty(t *testing.T) {
	ms := NewMultiSelect()
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	ms.Paint(buf)
}

func TestMultiSelectChildren(t *testing.T) {
	ms := NewMultiSelect()
	if ms.Children() != nil { t.Error("Children should be nil") }
}

func TestMultiSelectStyle(t *testing.T) {
	ms := NewMultiSelect()
	ms.SetStyle(MultiSelectStyle{Normal: buffer.Style{Fg: buffer.RGB(200,200,200)}, Selected: buffer.Style{Fg: buffer.RGB(0,255,0)}, Highlight: buffer.Style{Fg: buffer.RGB(0,0,255), Flags: buffer.Bold}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ms.AddOption("X")
	ms.Toggle(0)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	ms.Paint(buf)
}
