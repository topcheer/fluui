package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownEmojiBasic(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown("Hello :smile: world")
	if me.EmojiCount() != 1 { t.Errorf("EmojiCount = %d, want 1", me.EmojiCount()) }
}

func TestMarkdownEmojiMultiple(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown(":heart: :fire: :rocket:")
	if me.EmojiCount() != 3 { t.Errorf("EmojiCount = %d, want 3", me.EmojiCount()) }
}

func TestMarkdownEmojiUnknown(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown(":unknowncode: text")
	if me.EmojiCount() != 0 { t.Errorf("EmojiCount = %d, want 0", me.EmojiCount()) }
}

func TestMarkdownEmojiNoEmoji(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown("Just plain text")
	if me.EmojiCount() != 0 { t.Errorf("EmojiCount = %d, want 0", me.EmojiCount()) }
}

func TestMarkdownEmojiEmpty(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown("")
	if me.EmojiCount() != 0 { t.Errorf("EmojiCount = %d, want 0", me.EmojiCount()) }
}

func TestMarkdownEmojiMeasure(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown(":heart:")
	s := me.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownEmojiPaint(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetMarkdown("Test :fire: emoji")
	me.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	me.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'T' { foundText = true; break }
	}
	if !foundText { t.Error("text not found") }
}

func TestMarkdownEmojiChildren(t *testing.T) {
	me := NewMarkdownEmoji()
	if me.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownEmojiStyle(t *testing.T) {
	me := NewMarkdownEmoji()
	me.SetStyle(EmojiStyle{Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Emoji: buffer.Style{Fg: buffer.RGB(255,0,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	me.SetMarkdown(":star:")
	me.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	me.Paint(buf)
}

// ─── ChipBadge tests ───

func TestChipBadgeBasic(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("go")
	cb.AddChip("tui")
	if cb.ChipCount() != 2 { t.Errorf("ChipCount = %d, want 2", cb.ChipCount()) }
}

func TestChipBadgeRemove(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("a")
	cb.AddChip("b")
	cb.AddChip("c")
	cb.RemoveChip(1)
	if cb.ChipCount() != 2 { t.Errorf("ChipCount = %d, want 2", cb.ChipCount()) }
}

func TestChipBadgeRemoveOutOfRange(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("a")
	cb.RemoveChip(5)
	cb.RemoveChip(-1)
	if cb.ChipCount() != 1 { t.Errorf("ChipCount = %d, want 1", cb.ChipCount()) }
}

func TestChipBadgeToggle(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("test")
	cb.ToggleSelected(0)
	cb.mu.Lock()
	if !cb.chips[0].Selected { t.Error("chip should be selected after toggle") }
	cb.mu.Unlock()
	cb.ToggleSelected(0)
	cb.mu.Lock()
	if cb.chips[0].Selected { t.Error("chip should be unselected after second toggle") }
	cb.mu.Unlock()
}

func TestChipBadgeClear(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("a")
	cb.Clear()
	if cb.ChipCount() != 0 { t.Errorf("ChipCount = %d, want 0", cb.ChipCount()) }
}

func TestChipBadgeEmpty(t *testing.T) {
	cb := NewChipBadge()
	if cb.ChipCount() != 0 { t.Errorf("ChipCount = %d, want 0", cb.ChipCount()) }
}

func TestChipBadgeMeasure(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("test")
	s := cb.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestChipBadgePaint(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("go")
	cb.AddChip("tui")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
	foundBracket := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '[' { foundBracket = true; break }
	}
	if !foundBracket { t.Error("chip bracket not found") }
}

func TestChipBadgePaintSelected(t *testing.T) {
	cb := NewChipBadge()
	cb.AddChip("active")
	cb.ToggleSelected(0)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf1 := buffer.NewBuffer(40, 1)
	cb.Paint(buf1)
	selColor := buf1.GetCell(0, 0).Fg

	cb.ToggleSelected(0)
	buf2 := buffer.NewBuffer(40, 1)
	cb.Paint(buf2)
	normColor := buf2.GetCell(0, 0).Fg

	if selColor.Equal(normColor) { t.Error("expected different colors for selected vs normal") }
}

func TestChipBadgePaintEmpty(t *testing.T) {
	cb := NewChipBadge()
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
}

func TestChipBadgeChildren(t *testing.T) {
	cb := NewChipBadge()
	if cb.Children() != nil { t.Error("Children should be nil") }
}

func TestChipBadgeStyle(t *testing.T) {
	cb := NewChipBadge()
	cb.SetStyle(ChipBadgeStyle{Normal: buffer.Style{Fg: buffer.RGB(100,100,100)}, Selected: buffer.Style{Fg: buffer.RGB(0,255,0)}, Remove: buffer.Style{Fg: buffer.RGB(255,0,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	cb.AddChip("x")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
}
