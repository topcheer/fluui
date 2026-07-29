package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownTableBasic(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetMarkdown(`| Name | Age |
|------|-----|
| Alice | 30 |
| Bob | 25 |`)

	if mt.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", mt.RowCount())
	}
	if mt.ColumnCount() != 2 {
		t.Errorf("ColumnCount = %d, want 2", mt.ColumnCount())
	}
}

func TestMarkdownTableEmpty(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetMarkdown("")
	if mt.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", mt.RowCount())
	}
}

func TestMarkdownTableAlignments(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetMarkdown(`| Left | Center | Right |
|:-----|:------:|------:|
| a | b | c |`)

	// Check column count
	if mt.ColumnCount() != 3 {
		t.Fatalf("ColumnCount = %d, want 3", mt.ColumnCount())
	}
}

func TestMarkdownTableParseAlign(t *testing.T) {
	if parseAlign(":---") != alignLeft {
		t.Error(":--- should be alignLeft")
	}
	if parseAlign("---:") != alignRight {
		t.Error("---: should be alignRight")
	}
	if parseAlign(":---:") != alignCenter {
		t.Error(":---: should be alignCenter")
	}
	if parseAlign("---") != alignLeft {
		t.Error("--- should be alignLeft (default)")
	}
}

func TestMarkdownTableSeparator(t *testing.T) {
	if !isSeparatorRow("|------|------|") {
		t.Error("valid separator not detected")
	}
	if isSeparatorRow("| data | data |") {
		t.Error("non-separator detected as separator")
	}
}

func TestMarkdownTableMeasure(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetMarkdown(`| Name | Age |
|------|-----|
| Alice | 30 |`)
	s := mt.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 3 {
		t.Errorf("H = %d, want >= 3", s.H)
	}
}

func TestMarkdownTablePaint(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetMarkdown(`| Name  | Score |
|:------|------:|
| Alice |   100 |
| Bob   |    85 |
| Carol |    92 |`)
	mt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})

	buf := buffer.NewBuffer(60, 10)
	mt.Paint(buf)

	// Check top border exists
	if buf.GetCell(0, 0).Rune != '─' {
		t.Error("top border missing")
	}
	// Check header text
	foundHeader := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 1).Rune == 'N' {
			foundHeader = true
			break
		}
	}
	if !foundHeader {
		t.Error("header text not found")
	}
	// Check pipe separators
	foundPipe := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 1).Rune == '|' {
			foundPipe = true
			break
		}
	}
	if !foundPipe {
		t.Error("pipe separator not found")
	}
}

func TestMarkdownTablePaintEmpty(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	mt.Paint(buf) // should not panic
}

func TestMarkdownTableChildren(t *testing.T) {
	mt := NewMarkdownTable()
	if mt.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownTableNoSeparator(t *testing.T) {
	mt := NewMarkdownTable()
	// Table without separator row
	mt.SetMarkdown(`| A | B |
| 1 | 2 |`)
	if mt.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 (no separator, row counts as data)", mt.RowCount())
	}
}

func TestMarkdownTableStyle(t *testing.T) {
	mt := NewMarkdownTable()
	mt.SetStyle(MarkdownTableStyle{
		Header:    buffer.Style{Fg: buffer.RGB(255, 0, 255), Flags: buffer.Bold},
		Cell:      buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Separator: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Border:    buffer.Style{Fg: buffer.RGB(50, 50, 50)},
		AltRow:    buffer.Style{Fg: buffer.RGB(180, 180, 180)},
	})
	mt.SetMarkdown(`| X | Y |
|---|---|
| 1 | 2 |
| 3 | 4 |`)
	mt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	mt.Paint(buf)
}
