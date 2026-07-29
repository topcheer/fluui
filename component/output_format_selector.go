package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── OutputFormatSelector: AI Output Format Display ───
//
// OutputFormatSelector renders a compact selector showing the current
// output format for AI responses. Supports cycling through formats
// like JSON, Text, Markdown, and Code.
//
// Usage:
//
//	of := NewOutputFormatSelector()
//	of.SetActive(FormatMarkdown)
//	of.Paint(buf)

// OutputFormat represents supported output formats.
type OutputFormat int

const (
	FormatText     OutputFormat = 0
	FormatJSON     OutputFormat = 1
	FormatMarkdown OutputFormat = 2
	FormatCode     OutputFormat = 3
	FormatHTML     OutputFormat = 4
)

var formatNames = [...]string{"Text", "JSON", "Markdown", "Code", "HTML"}

// OutputFormatStyle holds styling.
type OutputFormatStyle struct {
	Active  buffer.Style
	Inactive buffer.Style
	Bracket buffer.Style
	Separator buffer.Style
}

// DefaultOutputFormatStyle returns defaults.
func DefaultOutputFormatStyle() OutputFormatStyle {
	return OutputFormatStyle{
		Active:    buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Inactive:  buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Bracket:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Separator: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// OutputFormatSelector renders a format selector.
type OutputFormatSelector struct {
	BaseComponent
	mu sync.Mutex

	active OutputFormat
	style  OutputFormatStyle
}

// NewOutputFormatSelector creates an OutputFormatSelector.
func NewOutputFormatSelector() *OutputFormatSelector {
	of := &OutputFormatSelector{style: DefaultOutputFormatStyle()}
	of.SetID(GenerateID("outfmt"))
	return of
}

// SetActive sets the active format.
func (of *OutputFormatSelector) SetActive(f OutputFormat) *OutputFormatSelector {
	of.mu.Lock()
	if int(f) < 0 || int(f) >= len(formatNames) {
		f = FormatText
	}
	of.active = f
	of.mu.Unlock()
	return of
}

// CycleNext advances to the next format.
func (of *OutputFormatSelector) CycleNext() *OutputFormatSelector {
	of.mu.Lock()
	of.active = (of.active + 1) % OutputFormat(len(formatNames))
	of.mu.Unlock()
	return of
}

// Active returns the active format.
func (of *OutputFormatSelector) Active() OutputFormat {
	of.mu.Lock()
	defer of.mu.Unlock()
	return of.active
}

// SetStyle sets custom style.
func (of *OutputFormatSelector) SetStyle(s OutputFormatStyle) *OutputFormatSelector {
	of.mu.Lock()
	of.style = s
	of.mu.Unlock()
	return of
}

// Measure returns preferred size.
func (of *OutputFormatSelector) Measure(cs Constraints) Size {
	w := 30
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the format selector.
func (of *OutputFormatSelector) Paint(buf *buffer.Buffer) {
	of.mu.Lock()
	defer of.mu.Unlock()

	b := of.Bounds()
	x, y := b.X, b.Y

	activeStyle := of.style.Active
	inactiveStyle := of.style.Inactive
	bracketStyle := of.style.Bracket
	sepStyle := of.style.Separator

	col := x

	// [ prefix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}

	for i := 0; i < len(formatNames); i++ {
		var style_ buffer.Style
		if OutputFormat(i) == of.active {
			style_ = activeStyle
		} else {
			style_ = inactiveStyle
		}

		for _, r := range formatNames[i] {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
			col++
		}

		if i < len(formatNames)-1 {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: '|', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}
	}

	// ] suffix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (of *OutputFormatSelector) Children() []Component { return nil }
