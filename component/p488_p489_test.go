package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownStrikethroughBasic(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("This is ~~old~~ text.")
	if ms.StrikethroughCount() != 1 {
		t.Errorf("StrikethroughCount = %d, want 1", ms.StrikethroughCount())
	}
}

func TestMarkdownStrikethroughMultiple(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("~~a~~ and ~~b~~ and ~c~")
	if ms.StrikethroughCount() != 3 {
		t.Errorf("StrikethroughCount = %d, want 3", ms.StrikethroughCount())
	}
}

func TestMarkdownStrikethroughNoStrike(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("Just normal text here.")
	if ms.StrikethroughCount() != 0 {
		t.Errorf("StrikethroughCount = %d, want 0", ms.StrikethroughCount())
	}
}

func TestMarkdownStrikethroughEmpty(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("")
	if ms.StrikethroughCount() != 0 {
		t.Errorf("StrikethroughCount = %d, want 0", ms.StrikethroughCount())
	}
}

func TestMarkdownStrikethroughCount(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("~~one~~ normal ~~two~~")
	if ms.StrikethroughCount() != 2 {
		t.Errorf("StrikethroughCount = %d, want 2", ms.StrikethroughCount())
	}
}

func TestMarkdownStrikethroughMeasure(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("text ~~strike~~")
	s := ms.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 3 {
		t.Errorf("H = %d, want >= 3", s.H)
	}
}

func TestMarkdownStrikethroughPaint(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("Normal ~~struck~~ text")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	ms.Paint(buf)

	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'N' {
			foundText = true
			break
		}
	}
	if !foundText {
		t.Error("text 'Normal' not found")
	}
}

func TestMarkdownStrikethroughChildren(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	if ms.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownStrikethroughStyle(t *testing.T) {
	ms := NewMarkdownStrikethrough()
	ms.SetStyle(StrikethroughStyle{
		Text:          buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Strikethrough: buffer.Style{Fg: buffer.RGB(100, 100, 100), Flags: buffer.Strikethrough},
		Border:        buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ms.SetMarkdown("~~test~~")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	ms.Paint(buf)
}

// ─── MarkdownEmphasis tests ───

func TestMarkdownEmphasisBasic(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("This is **bold** text.")
	if me.BoldCount() != 1 {
		t.Errorf("BoldCount = %d, want 1", me.BoldCount())
	}
}

func TestMarkdownEmphasisItalic(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("This is *italic* text.")
	if me.ItalicCount() != 1 {
		t.Errorf("ItalicCount = %d, want 1", me.ItalicCount())
	}
}

func TestMarkdownEmphasisBoldItalic(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("***both*** markers")
	if me.BoldCount() != 1 {
		t.Errorf("BoldCount = %d, want 1", me.BoldCount())
	}
	if me.ItalicCount() != 1 {
		t.Errorf("ItalicCount = %d, want 1", me.ItalicCount())
	}
}

func TestMarkdownEmphasisNoEmphasis(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("Just plain text.")
	if me.BoldCount() != 0 {
		t.Errorf("BoldCount = %d, want 0", me.BoldCount())
	}
}

func TestMarkdownEmphasisEmpty(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("")
	if me.BoldCount() != 0 || me.ItalicCount() != 0 {
		t.Error("counts should be 0 for empty")
	}
}

func TestMarkdownEmphasisUnderscore(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("__bold__ and _italic_")
	if me.BoldCount() != 1 {
		t.Errorf("BoldCount = %d, want 1", me.BoldCount())
	}
}

func TestMarkdownEmphasisMeasure(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("**bold**")
	s := me.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
}

func TestMarkdownEmphasisPaint(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("Text **bold** more")
	me.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	me.Paint(buf)

	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
}

func TestMarkdownEmphasisChildren(t *testing.T) {
	me := NewMarkdownEmphasis()
	if me.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownEmphasisStyle(t *testing.T) {
	me := NewMarkdownEmphasis()
	me.SetStyle(EmphasisStyle{
		Text:       buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Bold:       buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
		Italic:     buffer.Style{Fg: buffer.RGB(200, 200, 200), Flags: buffer.Italic},
		BoldItalic: buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold | buffer.Italic},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	me.SetMarkdown("**b** *i* ***bi***")
	me.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	me.Paint(buf)
}
