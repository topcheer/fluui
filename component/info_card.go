package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// InfoCardVariant controls the color scheme of an InfoCard.
type InfoCardVariant int

const (
	InfoCardDefault InfoCardVariant = iota
	InfoCardSuccess
	InfoCardWarning
	InfoCardError
	InfoCardAccent
)

// InfoCard renders a compact card with icon, title, and body text.
// Useful for displaying AI analysis results, system status, or
// contextual information in a visually distinct container.
//
// Thread-safe.
type InfoCard struct {
	BaseComponent
	mu      sync.RWMutex
	icon    string
	title   string
	body    string
	variant InfoCardVariant
}

// NewInfoCard creates an info card with icon, title, and body.
func NewInfoCard(icon, title, body string) *InfoCard {
	return &InfoCard{
		BaseComponent: BaseComponent{id: GenerateID("infocard")},
		icon:          icon,
		title:         title,
		body:          body,
		variant:       InfoCardDefault,
	}
}

func (c *InfoCard) Icon() string { c.mu.RLock(); defer c.mu.RUnlock(); return c.icon }
func (c *InfoCard) SetIcon(s string) { c.mu.Lock(); defer c.mu.Unlock(); c.icon = s }

func (c *InfoCard) Title() string { c.mu.RLock(); defer c.mu.RUnlock(); return c.title }
func (c *InfoCard) SetTitle(s string) { c.mu.Lock(); defer c.mu.Unlock(); c.title = s }

func (c *InfoCard) Body() string { c.mu.RLock(); defer c.mu.RUnlock(); return c.body }
func (c *InfoCard) SetBody(s string) { c.mu.Lock(); defer c.mu.Unlock(); c.body = s }

func (c *InfoCard) Variant() InfoCardVariant { c.mu.RLock(); defer c.mu.RUnlock(); return c.variant }
func (c *InfoCard) SetVariant(v InfoCardVariant) { c.mu.Lock(); defer c.mu.Unlock(); c.variant = v }

func (c *InfoCard) resolveAccentLocked() buffer.Color {
	t := theme.Get()
	switch c.variant {
	case InfoCardSuccess: return t.Success
	case InfoCardWarning: return t.Warning
	case InfoCardError: return t.Error
	case InfoCardAccent: return t.Accent
	default: return t.Muted
	}
}

func (c *InfoCard) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w := 0
	if c.icon != "" { w += len(c.icon) + 1 }
	if tw := len(c.title); tw > w { w = tw }
	if bw := len(c.body); bw > w { w = bw }
	w += 4 // padding
	h := 1 // title row
	if c.body != "" { h++ }
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 4 { w = 4 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

func (c *InfoCard) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	t := theme.Get()
	accent := c.resolveAccentLocked()
	titleStyle := buffer.Style{Fg: accent, Flags: buffer.Bold}
	bodyStyle := buffer.Style{Fg: t.Fg}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Title row: icon + title
	if c.icon != "" {
		x = buf.DrawText(x, y, c.icon+" ", titleStyle)
	}
	if c.title != "" {
		x = buf.DrawText(x, y, c.title, titleStyle)
	}

	// Body row
	if c.body != "" && bounds.H > 1 {
		by := y + 1
		bw := bounds.W
		if c.icon != "" { bw -= 2; buf.DrawText(bounds.X, by, "  ", bodyStyle) }
		avail := maxX - bounds.X - 2
		if avail < 0 { avail = 0 }
		// Draw body text, truncated to width
		drawn := 0
		for _, r := range c.body {
			if drawn >= avail { break }
			rw := buffer.RuneWidth(r)
			if bounds.X+drawn+rw > maxX { break }
			buf.SetCell(bounds.X+(func() int { if c.icon != "" { return 2 }; return 0 }())+drawn, by, buffer.Cell{
				Rune: r, Width: uint8(rw), Fg: bodyStyle.Fg,
			})
			if rw == 2 { drawn += 2 } else { drawn++ }
		}
	}
}
