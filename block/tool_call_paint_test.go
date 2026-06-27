package block

import (
	"errors"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// findRune searches for a rune in a buffer row, returns x or -1.
func findRune(buf *buffer.Buffer, y int, want rune) int {
	for x := 0; x < buf.Width; x++ {
		if buf.GetCell(x, y).Rune == want {
			return x
		}
	}
	return -1
}

func TestToolCallPaintDimStreaming(t *testing.T) {
	b := NewToolCallBlock("tcp-1", "ReadFile", `"test.go"`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Streaming state should use Fg = RGB(0x62, 0x72, 0xA4)
	x := findRune(buf, 0, 'R') // 'R' from "ReadFile"
	if x < 0 {
		t.Fatal("could not find 'R' from tool name in buffer")
	}
	cell := buf.GetCell(x, 0)
	want := buffer.RGB(0x62, 0x72, 0xA4)
	if !cell.Fg.Equal(want) {
		t.Errorf("streaming Fg = %s, want %s", cell.Fg, want)
	}
}

func TestToolCallPaintGreenComplete(t *testing.T) {
	b := NewToolCallBlock("tcp-2", "Edit", `"f.go"`)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Must contain ✓
	x := findRune(buf, 0, '✓')
	if x < 0 {
		t.Fatal("expected ✓ in buffer for complete state")
	}

	// The ✓ cell should be green
	cell := buf.GetCell(x, 0)
	want := buffer.RGB(0x50, 0xFA, 0x7B)
	if !cell.Fg.Equal(want) {
		t.Errorf("complete ✓ Fg = %s, want %s (green)", cell.Fg, want)
	}
}

func TestToolCallPaintRedError(t *testing.T) {
	b := NewToolCallBlock("tcp-3", "Bash", `"ls"`)
	b.Fail(errors.New("boom"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Must contain ✗
	x := findRune(buf, 0, '✗')
	if x < 0 {
		t.Fatal("expected ✗ in buffer for error state")
	}

	// The ✗ cell should be red
	cell := buf.GetCell(x, 0)
	want := buffer.RGB(0xFF, 0x55, 0x55)
	if !cell.Fg.Equal(want) {
		t.Errorf("error ✗ Fg = %s, want %s (red)", cell.Fg, want)
	}
}

func TestToolCallPaintTruncate(t *testing.T) {
	longArgs := `{"very_long_argument_name":"value_that_exceeds_buffer_width"}`
	b := NewToolCallBlock("tcp-4", "WriteFile", longArgs)

	// Very narrow width to force truncation
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)

	// Buffer should not overflow: every cell in row 0 should be from our paint
	// And we should find the truncation marker … or at least the tool name
	foundEllipsis := false
	for x := 0; x < 10; x++ {
		r := buf.GetCell(x, 0).Rune
		if r == '…' {
			foundEllipsis = true
			break
		}
	}
	// With width 10, maxArgsLen = 10 - 8 ("WriteFile" is 9, 9+8=17 > 10 so maxArgsLen=0)
	// argsPreview becomes "" when maxArgsLen <= 1
	// So text = "⏺ WriteFi" (truncated to width 10)
	// Just verify no panic and buffer is filled
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 0 {
			t.Errorf("cell (%d,0) rune is 0, should be painted or space", x)
		}
	}
	_ = foundEllipsis // ellipsis may not appear when width is too small
}

func TestToolCallPaintTruncateWithEllipsis(t *testing.T) {
	longArgs := `{"path":"/very/long/path/to/some/file/that/exceeds/the/width.go"}`
	b := NewToolCallBlock("tcp-5", "Rd", longArgs)

	// Width 20: maxArgsLen = 20 - 2 ("Rd") - 8 = 10
	// args (67 chars) > 10, so truncate to 9 chars + "…"
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 1})

	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)

	foundEllipsis := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '…' {
			foundEllipsis = true
			break
		}
	}
	if !foundEllipsis {
		t.Error("expected … truncation marker when args exceed maxArgsLen")
	}
}

func TestToolCallPaintToolName(t *testing.T) {
	b := NewToolCallBlock("tcp-6", "MyCustomTool", `{}`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Buffer should contain "MyCustomTool"
	rest := readRunes(buf, 2, 0, 12) // skip "⏺ " prefix
	if rest != "MyCustomTool" {
		t.Errorf("expected 'MyCustomTool' at x=2, got %q", rest)
	}
}

func TestToolCallPaintOffsetBounds(t *testing.T) {
	b := NewToolCallBlock("tcp-7", "Grep", `"pat"`)
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 1})

	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)

	// The painted text should be at (5,3), not (0,0)
	cell := buf.GetCell(5, 3)
	if cell.Rune == ' ' || cell.Rune == 0 {
		t.Errorf("expected painted content at (5,3), got rune %q", cell.Rune)
	}

	// (0,0) should still be blank
	blank := buf.GetCell(0, 0)
	if blank.Rune != ' ' {
		t.Errorf("expected blank at (0,0), got rune %q", blank.Rune)
	}
}
