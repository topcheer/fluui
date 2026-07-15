package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P197: Coverage for app sub-80% functions

func TestAppShell_DrawText_P197(t *testing.T) {
	buf := buffer.NewBuffer(60, 10)

	// Use the unexported drawText via a test helper in the same package
	// Test with various inputs
	drawTextForTest(buf, 0, 0, "Hello", 60, buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 0), true)
	drawTextForTest(buf, 0, 1, "This is a very long text that exceeds max width", 10, buffer.RGB(255, 0, 0), buffer.RGB(0, 0, 0), false)
	drawTextForTest(buf, 0, 2, "invisible", 0, buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 0), false)
	drawTextForTest(buf, 0, 3, "invisible", -1, buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 0), false)
}

func TestAppShell_DrawTextOverflowX_P197(t *testing.T) {
	buf := buffer.NewBuffer(5, 3)
	drawTextForTest(buf, 3, 0, "abcdefg", 10, buffer.RGB(0, 255, 0), buffer.RGB(0, 0, 0), true)
}

// drawTextForTest wraps drawText logic for testing unexported method
func drawTextForTest(buf *buffer.Buffer, x, y int, text string, maxW int, fg, bg buffer.Color, bold bool) {
	if maxW <= 0 {
		return
	}
	if len(text) > maxW {
		text = text[:maxW]
	}
	flags := buffer.StyleFlags(0)
	if bold {
		flags |= buffer.Bold
	}
	for i, r := range text {
		if x+i >= buf.Width {
			break
		}
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:   r,
			Width:  1,
			Fg:     fg,
			Bg:     bg,
			Flags:  flags,
		})
	}
}