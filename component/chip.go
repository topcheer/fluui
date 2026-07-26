package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ChipVariant controls the visual style of a Chip.
type ChipVariant int

const (
	// ChipFilled renders the chip with a solid background (default).
	ChipFilled ChipVariant = iota
	// ChipOutlined renders the chip with a border and transparent background.
	ChipOutlined
	// ChipSubtle renders the chip with a muted background.
	ChipSubtle
)

// Chip is a compact label/tag badge for displaying categories, labels,
// or entity classifications. Similar to Material Design chips or GitHub labels.
//
// Designed for AI chat UIs to show:
//   - Model tags (e.g., "gpt-4", "claude-3")
//   - Status labels (e.g., "streaming", "complete", "error")
//   - Topic tags (e.g., "python", "refactoring")
//   - Priority indicators
//
// Thread-safe.
type Chip struct {
	BaseComponent
	mu       sync.RWMutex
	text     string
	icon     string // optional emoji/icon prefix
	variant  ChipVariant
	customFg buffer.Color // zero = auto from variant
	customBg buffer.Color // zero = auto from variant
}

// NewChip creates a chip with the given text.
func NewChip(text string) *Chip {
	return &Chip{
		BaseComponent: BaseComponent{id: GenerateID("chip")},
		text:          text,
		variant:       ChipFilled,
	}
}

// Text returns the chip display text.
func (c *Chip) Text() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.text
}

// SetText updates the chip display text.
func (c *Chip) SetText(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = s
}

// Icon returns the chip icon prefix (empty = none).
func (c *Chip) Icon() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.icon
}

// SetIcon sets an emoji or icon prefix. Pass "" to remove.
func (c *Chip) SetIcon(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.icon = s
}

// Variant returns the current visual style.
func (c *Chip) Variant() ChipVariant {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.variant
}

// SetVariant updates the visual style.
func (c *Chip) SetVariant(v ChipVariant) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.variant = v
}

// SetColors overrides foreground and background colors.
// Pass buffer.Color{} (zero) for either to revert to auto.
func (c *Chip) SetColors(fg, bg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.customFg = fg
	c.customBg = bg
}

// resolveFgLocked returns the effective foreground color.
func (c *Chip) resolveFgLocked() buffer.Color {
	if c.customFg.Type != buffer.ColorNone {
		return c.customFg
	}
	t := theme.Get()
	switch c.variant {
	case ChipFilled:
		return t.Bg // text on colored bg
	case ChipOutlined:
		return t.Accent
	case ChipSubtle:
		return t.Fg
	default:
		return t.Fg
	}
}

// resolveBgLocked returns the effective background color.
func (c *Chip) resolveBgLocked() buffer.Color {
	if c.customBg.Type != buffer.ColorNone {
		return c.customBg
	}
	t := theme.Get()
	switch c.variant {
	case ChipFilled:
		return t.Accent
	case ChipOutlined:
		return buffer.NoColor() // transparent
	case ChipSubtle:
		return t.Muted
	default:
		return t.Muted
	}
}

// contentWidth returns the visual width of icon + text + padding.
func (c *Chip) contentWidth() int {
	w := 0
	if c.icon != "" {
		w += buffer.StringWidth(c.icon) + 1 // icon + space
	}
	w += buffer.StringWidth(c.text)
	return w + 2 // 1 cell padding each side
}

// Measure returns the preferred size (always 1 row).
func (c *Chip) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w := c.contentWidth()
	h := 1

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the chip. Zero allocations.
func (c *Chip) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	fg := c.resolveFgLocked()
	bg := c.resolveBgLocked()

	t := theme.Get()
	flags := buffer.StyleFlags(0)
	if c.variant == ChipFilled || c.variant == ChipOutlined {
		flags = buffer.Bold
	}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Left padding
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}

	// Icon
	if c.icon != "" && x < maxX {
		x = buf.DrawText(x, y, c.icon+" ", buffer.Style{Fg: fg, Bg: bg, Flags: flags})
	}

	// Text
	if x < maxX {
		textStyle := buffer.Style{Fg: fg, Bg: bg, Flags: flags}
		x = buf.DrawText(x, y, c.text, textStyle)
	}

	// Right padding
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}

	// For outlined variant, draw border on edges if bg is transparent
	if c.variant == ChipOutlined {
		borderStyle := buffer.Style{Fg: t.Accent, Flags: buffer.Bold}
		// Draw top and bottom borders using SetCell
		// Actually for a single-row chip, just draw the corner chars
		// Left bracket [
		buf.SetCell(bounds.X, y, buffer.Cell{Rune: '[', Width: 1, Fg: borderStyle.Fg, Bg: bg, Flags: borderStyle.Flags})
		// Right bracket ]
		endX := bounds.X + bounds.W - 1
		if endX > bounds.X && endX < maxX+1 {
			buf.SetCell(endX, y, buffer.Cell{Rune: ']', Width: 1, Fg: borderStyle.Fg, Bg: bg, Flags: borderStyle.Flags})
		}
	}
}

// Helper: format chip count for display (e.g., "python (12)")
func chipCountLabel(label string, count int) string {
	if count <= 0 {
		return label
	}
	var buf [32]byte
	b := buf[:0]
	b = append(b, label...)
	b = append(b, " ("...)
	b = strconv.AppendInt(b, int64(count), 10)
	b = append(b, ')')
	return string(b)
}
