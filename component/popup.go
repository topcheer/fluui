package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Popup is a floating container that renders content with a border,
// positioned at an absolute (x, y) coordinate. Unlike Dialog, it has no
// buttons or modal overlay — it's a lightweight positioning wrapper.
//
// Thread-safe.
type Popup struct {
	BaseComponent
	mu      sync.RWMutex
	content Component
	title   string
	visible bool
	shadow  bool
}

// NewPopup creates a popup wrapping the given content.
func NewPopup(content Component) *Popup {
	return &Popup{
		BaseComponent: BaseComponent{id: GenerateID("popup")},
		content:       content,
		visible:       true,
	}
}

func (p *Popup) Content() Component { p.mu.RLock(); defer p.mu.RUnlock(); return p.content }

func (p *Popup) SetContent(c Component) { p.mu.Lock(); defer p.mu.Unlock(); p.content = c }

func (p *Popup) Title() string { p.mu.RLock(); defer p.mu.RUnlock(); return p.title }

func (p *Popup) SetTitle(s string) { p.mu.Lock(); defer p.mu.Unlock(); p.title = s }

func (p *Popup) Visible() bool { p.mu.RLock(); defer p.mu.RUnlock(); return p.visible }

func (p *Popup) SetVisible(b bool) { p.mu.Lock(); defer p.mu.Unlock(); p.visible = b }

func (p *Popup) Show() { p.mu.Lock(); defer p.mu.Unlock(); p.visible = true }

func (p *Popup) Hide() { p.mu.Lock(); defer p.mu.Unlock(); p.visible = false }

func (p *Popup) Shadow() bool { p.mu.RLock(); defer p.mu.RUnlock(); return p.shadow }

func (p *Popup) SetShadow(b bool) { p.mu.Lock(); defer p.mu.Unlock(); p.shadow = b }

func (p *Popup) Measure(cs Constraints) Size {
	p.mu.RLock()
	defer p.mu.RUnlock()
	w := cs.MaxWidth
	if w <= 0 { w = 20 }
	h := cs.MaxHeight
	if h <= 0 { h = 5 }
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 2 { w = 2 }
	if h < 2 { h = 2 }
	return Size{W: w, H: h}
}

func (p *Popup) Paint(buf *buffer.Buffer) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.visible { return }

	bounds := p.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	border := buffer.Style{Fg: tt.Border}
	titleStyle := buffer.Style{Fg: tt.Accent, Flags: buffer.Bold}

	// Shadow
	if p.shadow {
		shadowStyle := buffer.Style{Fg: tt.Muted}
		for i := 1; i < bounds.W && bounds.X+i < buf.Width; i++ {
			if bounds.Y+bounds.H < buf.Height {
				buf.SetCell(bounds.X+i, bounds.Y+bounds.H, buffer.Cell{Rune: '\u2592', Width: 1, Fg: shadowStyle.Fg})
			}
		}
		for i := 0; i < bounds.H && bounds.Y+i < buf.Height; i++ {
			if bounds.X+bounds.W < buf.Width {
				buf.SetCell(bounds.X+bounds.W, bounds.Y+i, buffer.Cell{Rune: '\u2592', Width: 1, Fg: shadowStyle.Fg})
			}
		}
	}

	// Top border with title
	y := bounds.Y
	if bounds.H < 2 { return } // need at least 2 rows for border+content

	// ╭─ title ─╮
	buf.SetCell(bounds.X, y, buffer.Cell{Rune: '\u256d', Width: 1, Fg: border.Fg})
	x := bounds.X + 1
	if p.title != "" {
		x = buf.DrawText(x, y, " "+p.title+" ", titleStyle)
	}
	for ; x < bounds.X+bounds.W-1; x++ {
		buf.SetCell(x, y, buffer.Cell{Rune: '\u2500', Width: 1, Fg: border.Fg})
	}
	buf.SetCell(bounds.X+bounds.W-1, y, buffer.Cell{Rune: '\u256e', Width: 1, Fg: border.Fg})

	// Side borders
	for row := 1; row < bounds.H-1; row++ {
		buf.SetCell(bounds.X, bounds.Y+row, buffer.Cell{Rune: '\u2502', Width: 1, Fg: border.Fg})
		buf.SetCell(bounds.X+bounds.W-1, bounds.Y+row, buffer.Cell{Rune: '\u2502', Width: 1, Fg: border.Fg})
	}

	// Bottom border
	botY := bounds.Y + bounds.H - 1
	buf.SetCell(bounds.X, botY, buffer.Cell{Rune: '\u2570', Width: 1, Fg: border.Fg})
	for x := bounds.X + 1; x < bounds.X+bounds.W-1; x++ {
		buf.SetCell(x, botY, buffer.Cell{Rune: '\u2500', Width: 1, Fg: border.Fg})
	}
	buf.SetCell(bounds.X+bounds.W-1, botY, buffer.Cell{Rune: '\u256f', Width: 1, Fg: border.Fg})

	// Content
	if p.content != nil {
		contentBounds := Rect{X: bounds.X + 1, Y: bounds.Y + 1, W: bounds.W - 2, H: bounds.H - 2}
		if contentBounds.W > 0 && contentBounds.H > 0 {
			p.content.SetBounds(contentBounds)
			p.content.Paint(buf)
		}
	}
}
