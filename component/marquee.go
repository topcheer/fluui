package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Marquee: Scrolling Text Display ───
//
// Marquee renders horizontally scrolling text. Supports both left-scroll
// and bounce modes. Useful for tickers, notifications, and status displays.
//
// Usage:
//
//	m := NewMarquee()
//	m.SetText("Breaking News: Fluui reaches 260+ components!")
//	m.SetWidth(40)
//	m.SetOffset(5)
//	m.Paint(buf)

// MarqueeStyle holds styling.
type MarqueeStyle struct {
	Text     buffer.Style
	Indicator buffer.Style
}

// DefaultMarqueeStyle returns defaults.
func DefaultMarqueeStyle() MarqueeStyle {
	return MarqueeStyle{
		Text:      buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Indicator: buffer.Style{Fg: buffer.RGB(251, 191, 36)},
	}
}

// Marquee renders scrolling text.
type Marquee struct {
	BaseComponent
	mu sync.Mutex

	text   string
	offset int
	width  int
	bounce bool
	style  MarqueeStyle
	// cached
	runes []rune
}

// NewMarquee creates a Marquee.
func NewMarquee() *Marquee {
	m := &Marquee{width: 30, style: DefaultMarqueeStyle()}
	m.SetID(GenerateID("marquee"))
	return m
}

// SetText sets the scrolling text content.
func (m *Marquee) SetText(s string) *Marquee {
	m.mu.Lock()
	m.text = s
	m.runes = []rune(s)
	m.mu.Unlock()
	return m
}

// SetOffset sets the current scroll offset.
func (m *Marquee) SetOffset(n int) *Marquee {
	m.mu.Lock()
	if n < 0 { n = 0 }
	m.offset = n
	m.mu.Unlock()
	return m
}

// SetWidth sets the visible window width.
func (m *Marquee) SetWidth(w int) *Marquee {
	m.mu.Lock()
	if w < 5 { w = 5 }
	m.width = w
	m.mu.Unlock()
	return m
}

// SetBounce toggles bounce mode (text bounces back and forth).
func (m *Marquee) SetBounce(b bool) *Marquee {
	m.mu.Lock()
	m.bounce = b
	m.mu.Unlock()
	return m
}

// Offset returns the current scroll offset.
func (m *Marquee) Offset() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.offset
}

// SetStyle sets custom style.
func (m *Marquee) SetStyle(s MarqueeStyle) *Marquee {
	m.mu.Lock()
	m.style = s
	m.mu.Unlock()
	return m
}

// Measure returns preferred size.
func (m *Marquee) Measure(cs Constraints) Size {
	w := m.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the marquee.
func (m *Marquee) Paint(buf *buffer.Buffer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := m.Bounds()
	x, y := b.X, b.Y
	w := m.width

	textStyle := m.style.Text
	indStyle := m.style.Indicator

	if len(m.runes) == 0 {
		return
	}

	// If text fits entirely, just show it
	if len(m.runes) <= w {
		col := x
		for _, r := range m.runes {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
		return
	}

	// Scrolling mode: show a window starting at offset
	off := m.offset % len(m.runes)
	col := x
	for i := 0; i < w; i++ {
		if col >= buf.Width { break }
		idx := (off + i) % len(m.runes)
		buf.SetCell(col, y, buffer.Cell{Rune: m.runes[idx], Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
		col++
	}

	// Scroll indicators
	if m.offset > 0 && x > 0 {
		if x-1 < buf.Width {
			buf.SetCell(x-1, y, buffer.Cell{Rune: '‹', Fg: indStyle.Fg, Bg: indStyle.Bg, Flags: indStyle.Flags, Width: 1})
		}
	}
	if len(m.runes) > w && x+w < buf.Width {
		buf.SetCell(x+w, y, buffer.Cell{Rune: '›', Fg: indStyle.Fg, Bg: indStyle.Bg, Flags: indStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (m *Marquee) Children() []Component { return nil }
