package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMergeViewBasic(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("ours", "line1\nline2\nline3")
	mv.SetRight("theirs", "line1\nchanged\nline3")

	if mv.LeftLabel() != "ours" {
		t.Errorf("LeftLabel = %q, want ours", mv.LeftLabel())
	}
	if mv.RightLabel() != "theirs" {
		t.Errorf("RightLabel = %q, want theirs", mv.RightLabel())
	}
	if mv.LeftContent() != "line1\nline2\nline3" {
		t.Errorf("LeftContent = %q", mv.LeftContent())
	}
	if mv.RightContent() != "line1\nchanged\nline3" {
		t.Errorf("RightContent = %q", mv.RightContent())
	}
}

func TestMergeViewConflictDetection(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("ours", "<<<<<<< HEAD\nline1\n=======\nline2\n>>>>>>> branch")
	mv.SetRight("theirs", "line1")

	if !mv.HasConflicts() {
		t.Error("HasConflicts should be true with conflict markers")
	}

	// No conflicts
	mv2 := NewMergeView()
	mv2.SetLeft("a", "hello")
	mv2.SetRight("b", "world")
	if mv2.HasConflicts() {
		t.Error("HasConflicts should be false without conflict markers")
	}
}

func TestMergeViewLineCount(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("a", "line1\nline2\nline3")
	mv.SetRight("b", "line1\nline2\nline3")

	// All equal — should have 3 lines
	if mv.LineCount() != 3 {
		t.Errorf("LineCount = %d, want 3", mv.LineCount())
	}
}

func TestMergeViewLineCountDiff(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("a", "same\nold\nsame")
	mv.SetRight("b", "same\nnew\nsame")

	// Line 0 equal, line 1 differs (removes+adds 2 lines), line 2 equal = 4 lines
	lc := mv.LineCount()
	if lc != 4 {
		t.Errorf("LineCount = %d, want 4 (equal + removed/added + equal)", lc)
	}
}

func TestMergeViewMeasure(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("a", "line1\nline2")
	mv.SetRight("b", "line1\nline2")

	s := mv.Measure(Constraints{})
	if s.W < 20 {
		t.Errorf("W = %d, want >= 20", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestMergeViewPaint(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("ours", "line1\nline2\nline3")
	mv.SetRight("theirs", "line1\nchanged\nline3")
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})

	buf := buffer.NewBuffer(60, 10)
	mv.Paint(buf)

	// Check border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	if buf.GetCell(59, 9).Rune != '┘' {
		t.Error("bottom-right corner missing")
	}

	// Check header "ours" exists
	foundHeader := false
	for x := 0; x < 30; x++ {
		c := buf.GetCell(x, 1)
		if c.Rune == 'o' || c.Rune == 'u' {
			foundHeader = true
			break
		}
	}
	if !foundHeader {
		t.Error("left header not found")
	}

	// Check separator column exists
	sepFound := false
	for y := 1; y < 9; y++ {
		if buf.GetCell(30, y).Rune == '│' {
			sepFound = true
			break
		}
	}
	if !sepFound {
		t.Error("separator not found")
	}
}

func TestMergeViewPaintConflict(t *testing.T) {
	mv := NewMergeView()
	mv.SetLeft("ours", "<<<<<<< HEAD\ntested\n=======\nbroken\n>>>>>>> branch")
	mv.SetRight("theirs", "line1")
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})

	buf := buffer.NewBuffer(60, 12)
	mv.Paint(buf) // should not panic
}

func TestMergeViewPaintEmpty(t *testing.T) {
	mv := NewMergeView()
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})

	buf := buffer.NewBuffer(60, 10)
	mv.Paint(buf) // should not panic with empty content
}

func TestMergeViewChildren(t *testing.T) {
	mv := NewMergeView()
	if mv.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMergeViewStyle(t *testing.T) {
	mv := NewMergeView()
	mv.SetStyle(MergeLineStyle{
		Equal:     buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Added:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Removed:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Conflict:  buffer.Style{Fg: buffer.RGB(255, 255, 0), Flags: buffer.Bold},
		Header:    buffer.Style{Fg: buffer.RGB(255, 0, 255), Flags: buffer.Bold},
		Border:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Separator: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	mv.SetLeft("L", "same\ndiff")
	mv.SetRight("R", "same\nchanged")
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	mv.Paint(buf) // should not panic
}
