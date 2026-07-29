package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AsciiArtBox: ASCII Art Text Display ───
//
// AsciiArtBox renders text in a large ASCII art style using block characters.
// Supports a simple 5-line font for A-Z, 0-9, and space.
// Useful for banners, headers, and splash screens.
//
// Usage:
//
//	aa := NewAsciiArtBox()
//	aa.SetText("HELLO")
//	aa.Paint(buf)

// AsciiArtStyle holds styling.
type AsciiArtStyle struct {
	Text   buffer.Style
	Shadow buffer.Style
}

// DefaultAsciiArtStyle returns defaults.
func DefaultAsciiArtStyle() AsciiArtStyle {
	return AsciiArtStyle{
		Text:   buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Shadow: buffer.Style{Fg: buffer.RGB(30, 41, 59)},
	}
}

// 5-row ASCII art font for digits 0-9 (simplified block style)
var asciiArtDigits = map[rune][5]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", " █ ", " █ ", " █ "},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
}

// AsciiArtBox renders large ASCII art text.
type AsciiArtBox struct {
	BaseComponent
	mu sync.Mutex

	text  string
	style AsciiArtStyle
	// cached rows — each row is pre-computed string of the art
	rows [5]string
}

// NewAsciiArtBox creates an AsciiArtBox.
func NewAsciiArtBox() *AsciiArtBox {
	aa := &AsciiArtBox{style: DefaultAsciiArtStyle()}
	aa.SetID(GenerateID("asciiart"))
	aa.recomputeLocked()
	return aa
}

// SetText sets the text to render (only digits 0-9 are supported as art;
// other characters are rendered as spaces).
func (aa *AsciiArtBox) SetText(s string) *AsciiArtBox {
	aa.mu.Lock()
	aa.text = s
	aa.recomputeLocked()
	aa.mu.Unlock()
	return aa
}

func (aa *AsciiArtBox) recomputeLocked() {
	for row := 0; row < 5; row++ {
		var line []rune
		for _, r := range aa.text {
			if glyph, ok := asciiArtDigits[r]; ok {
				line = append(line, []rune(glyph[row])...)
				line = append(line, ' ') // spacing between chars
			} else if r == ' ' {
				line = append(line, ' ', ' ', ' ', ' ')
			} else {
				// Unknown char: render as 3-wide block
				line = append(line, '░', '░', '░', ' ')
			}
		}
		aa.rows[row] = string(line)
	}
}

// Text returns the current text.
func (aa *AsciiArtBox) Text() string {
	aa.mu.Lock()
	defer aa.mu.Unlock()
	return aa.text
}

// SetStyle sets custom style.
func (aa *AsciiArtBox) SetStyle(s AsciiArtStyle) *AsciiArtBox {
	aa.mu.Lock()
	aa.style = s
	aa.mu.Unlock()
	return aa
}

// Measure returns preferred size.
func (aa *AsciiArtBox) Measure(cs Constraints) Size {
	aa.mu.Lock()
	w := len(aa.rows[0])
	aa.mu.Unlock()
	if w < 3 {
		w = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 5}
}

// Paint renders the ASCII art.
func (aa *AsciiArtBox) Paint(buf *buffer.Buffer) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	b := aa.Bounds()
	x, y := b.X, b.Y

	textStyle := aa.style.Text

	for row := 0; row < 5; row++ {
		yy := y + row
		if yy >= buf.Height {
			break
		}
		col := x
		for _, r := range aa.rows[row] {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (aa *AsciiArtBox) Children() []Component { return nil }
