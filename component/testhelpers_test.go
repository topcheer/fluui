package component

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
)

// newTestBuffer creates a buffer filled with spaces for testing Paint output.
func newTestBuffer(w, h int) *buffer.Buffer {
	buf := buffer.NewBuffer(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.NewCell(' ', buffer.DefaultStyle))
		}
	}
	return buf
}

// cellRunes reads runes starting at (x, y) for maxW cells and returns them as a string.
func cellRunes(buf *buffer.Buffer, x, y, maxW int) string {
	var sb strings.Builder
	for i := x; i < x+maxW; i++ {
		c := buf.GetCell(i, y)
		if c.Width == 0 {
			continue
		}
		sb.WriteRune(c.Rune)
	}
	return strings.TrimRight(sb.String(), " ")
}
