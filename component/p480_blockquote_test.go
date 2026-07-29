package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownBlockquoteBasic(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> This is a quote.")
	if bq.LineCount() != 1 {
		t.Errorf("LineCount = %d, want 1", bq.LineCount())
	}
}

func TestMarkdownBlockquoteMultiLine(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> First line.\n> Second line.\n> Third line.")
	if bq.LineCount() != 3 {
		t.Errorf("LineCount = %d, want 3", bq.LineCount())
	}
}

func TestMarkdownBlockquoteNested(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> Outer.\n>> Inner nested.")
	if bq.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", bq.LineCount())
	}
	// Verify nesting parsed correctly
	bq.mu.Lock()
	if bq.cachedLines[1].Indent != 2 {
		t.Errorf("nested indent = %d, want 2", bq.cachedLines[1].Indent)
	}
	if bq.cachedLines[1].Text != "Inner nested." {
		t.Errorf("nested text = %q, want 'Inner nested.'", bq.cachedLines[1].Text)
	}
	bq.mu.Unlock()
}

func TestMarkdownBlockquoteEmpty(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("")
	if bq.LineCount() != 0 {
		t.Errorf("LineCount = %d, want 0", bq.LineCount())
	}
}

func TestMarkdownBlockquoteMeasure(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> Line 1\n> Line 2")
	s := bq.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 4 {
		t.Errorf("H = %d, want >= 4", s.H)
	}
}

func TestMarkdownBlockquotePaint(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> This is a quote.\n> Second line.")
	bq.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})

	buf := buffer.NewBuffer(50, 5)
	bq.Paint(buf)

	// Check border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Check accent bar on line 1
	foundBar := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '▎' {
			foundBar = true
			break
		}
	}
	if !foundBar {
		t.Error("accent bar not found")
	}

	// Check text content
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'T' {
			foundText = true
			break
		}
	}
	if !foundText {
		t.Error("quote text not found")
	}
}

func TestMarkdownBlockquotePaintNested(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> Outer.\n>> Inner.")
	bq.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})

	buf := buffer.NewBuffer(50, 5)
	bq.Paint(buf)

	// Row 2 (inner) should have TWO accent bars
	barCount := 0
	for x := 0; x < 10; x++ {
		if buf.GetCell(x, 2).Rune == '▎' {
			barCount++
		}
	}
	if barCount < 2 {
		t.Errorf("nested accent bars = %d, want >= 2", barCount)
	}
}

func TestMarkdownBlockquoteChildren(t *testing.T) {
	bq := NewMarkdownBlockquote()
	if bq.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownBlockquoteStyle(t *testing.T) {
	bq := NewMarkdownBlockquote()
	bq.SetStyle(BlockquoteStyle{
		Text:   buffer.Style{Fg: buffer.RGB(200, 200, 200), Flags: buffer.Italic},
		Accent: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Nested: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	bq.SetMarkdown("> Test\n>> Nested")
	bq.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	bq.Paint(buf)
}
