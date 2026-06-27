package block

import (
	"errors"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestErrorBlockCreation(t *testing.T) {
	eb := NewErrorBlock("err-1")

	if eb.ID() != "err-1" {
		t.Fatalf("expected ID 'err-1', got %q", eb.ID())
	}
	if eb.Type() != TypeError {
		t.Fatalf("expected type %s, got %s", TypeError, eb.Type())
	}
	if eb.Message() != "" {
		t.Fatalf("expected empty message, got %q", eb.Message())
	}
	if eb.timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestErrorBlockWithMessage(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "connection refused")

	if eb.Message() != "connection refused" {
		t.Fatalf("expected message 'connection refused', got %q", eb.Message())
	}
	if eb.State() != BlockError {
		t.Fatalf("expected state %s, got %s", BlockError, eb.State())
	}
}

func TestErrorBlockAppendDelta(t *testing.T) {
	eb := NewErrorBlock("err-1")
	eb.AppendDelta("line 1\n")
	eb.AppendDelta("line 2")

	if eb.Message() != "line 1\nline 2" {
		t.Fatalf("expected 'line 1\nline 2', got %q", eb.Message())
	}
}

func TestErrorBlockFail(t *testing.T) {
	eb := NewErrorBlock("err-1")
	eb.Fail(errors.New("timeout"))

	if eb.State() != BlockError {
		t.Fatalf("expected state %s, got %s", BlockError, eb.State())
	}
	if eb.Message() != "timeout" {
		t.Fatalf("expected message 'timeout', got %q", eb.Message())
	}
}

func TestErrorBlockMeasure(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "single line error")

	size := eb.Measure(component.Constraints{MaxWidth: 80})
	// content(1) + border top(1) + bottom(1) = 3
	if size.H != 3 {
		t.Fatalf("expected height 3, got %d", size.H)
	}
	if size.W != 80 {
		t.Fatalf("expected width 80, got %d", size.W)
	}
}

func TestErrorBlockMeasureMultiLine(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "line 1\nline 2\nline 3")

	size := eb.Measure(component.Constraints{MaxWidth: 80})
	// content(3) + border(2) = 5
	if size.H != 5 {
		t.Fatalf("expected height 5, got %d", size.H)
	}
}

func TestErrorBlockMeasureEmpty(t *testing.T) {
	eb := NewErrorBlock("err-1")

	size := eb.Measure(component.Constraints{MaxWidth: 80})
	// content(0) + border(2) = 2
	if size.H != 2 {
		t.Fatalf("expected height 2 for empty message, got %d", size.H)
	}
}

func TestErrorBlockPaint(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "test error")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})

	buf := buffer.NewBuffer(40, 3)
	eb.Paint(buf)

	// Top-left corner should be ╭
	cell := buf.GetCell(0, 0)
	if cell.Rune != '╭' {
		t.Fatalf("expected '╭' at (0,0), got %c", cell.Rune)
	}

	// The label "ERROR" should appear in the top border
	found := false
	for x := 0; x < 40; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == 'E' {
			// Check next chars for RROR
			c2 := buf.GetCell(x+1, 0)
			if c2.Rune == 'R' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected 'ERROR' label in top border")
	}

	// Content on row 1 should contain 't' from "test error"
	found = false
	for x := 0; x < 40; x++ {
		c := buf.GetCell(x, 1)
		if c.Rune == 't' {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected content 'test error' on row 1")
	}

	// Bottom-left corner should be ╰
	cell = buf.GetCell(0, 2)
	if cell.Rune != '╰' {
		t.Fatalf("expected '╰' at (0,2), got %c", cell.Rune)
	}
}

func TestErrorBlockPaintEmpty(t *testing.T) {
	eb := NewErrorBlock("err-1")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(40, 10)
	eb.Paint(buf) // should not panic
}

func TestErrorBlockPaintMultiLine(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "line 1\nline 2\nline 3")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 5})

	buf := buffer.NewBuffer(30, 5)
	eb.Paint(buf)

	// Check all 3 content lines have left borders
	for i := 0; i < 3; i++ {
		cell := buf.GetCell(0, 1+i)
		if cell.Rune != '│' {
			t.Fatalf("expected '│' at (0,%d), got %c", 1+i, cell.Rune)
		}
	}

	// Bottom border
	cell := buf.GetCell(0, 4)
	if cell.Rune != '╰' {
		t.Fatalf("expected '╰' at (0,4), got %c", cell.Rune)
	}
}

func TestErrorBlockString(t *testing.T) {
	eb := NewErrorBlock("err-1")
	s := eb.String()

	if !strings.Contains(s, "err-1") {
		t.Fatalf("expected ID in String(), got %q", s)
	}
	if !strings.Contains(s, "error") {
		t.Fatalf("expected type 'error' in String(), got %q", s)
	}
}

func TestErrorBlockComplete(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-1", "some error")
	eb.Complete()

	if eb.State() != BlockComplete {
		t.Fatalf("expected state %s, got %s", BlockComplete, eb.State())
	}
}

func TestErrorBlockIsDirty(t *testing.T) {
	eb := NewErrorBlock("err-1")
	if !eb.IsDirty() {
		t.Fatal("expected dirty on creation")
	}

	eb.ClearDirty()
	if eb.IsDirty() {
		t.Fatal("expected not dirty after ClearDirty")
	}

	eb.AppendDelta("more error")
	if !eb.IsDirty() {
		t.Fatal("expected dirty after AppendDelta")
	}
}
