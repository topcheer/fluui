package overlay

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// PopupStyle holds the visual styling for a Popup.
type PopupStyle struct {
	Border buffer.Style
	Title  buffer.Style
	Body   buffer.Style
}

// DefaultPopupStyle returns a reasonable default style.
func DefaultPopupStyle() PopupStyle {
	return PopupStyle{
		Border: buffer.DefaultStyle,
		Title:  buffer.DefaultStyle.AddFlags(buffer.Bold),
		Body:   buffer.DefaultStyle,
	}
}

// Popup is a full-screen overlay used for content viewers (e.g. code reader,
// diff viewer). It takes ~90% of the screen with a titled border.
type Popup struct {
	BaseOverlay
	title   string
	content component.Component
	style   PopupStyle
	width   int
	height  int
}

// NewPopup creates a full-screen popup overlay.
// z-index defaults to 90 (below modal at 100).
func NewPopup(id, title string, content component.Component) *Popup {
	return &Popup{
		BaseOverlay: NewBaseOverlay(id, 90, false),
		title:       title,
		content:     content,
		style:       DefaultPopupStyle(),
	}
}

// SetStyle overrides the default popup style.
func (p *Popup) SetStyle(s PopupStyle) {
	p.style = s
}

// Measure returns ~90% of the available space, clamped to reasonable bounds.
func (p *Popup) Measure(cs component.Constraints) component.Size {
	maxW := cs.MaxWidth
	maxH := cs.MaxHeight
	if maxW <= 0 {
		maxW = 80
	}
	if maxH <= 0 {
		maxH = 24
	}

	w := maxW * 9 / 10
	if w < 20 {
		w = 20
	}
	h := maxH * 9 / 10
	if h < 5 {
		h = 5
	}

	p.width = w
	p.height = h
	return component.Size{W: w, H: h}
}

// SetBounds centers the popup within the given rect.
func (p *Popup) SetBounds(r component.Rect) {
	w := p.width
	h := p.height
	if w <= 0 {
		w = r.W * 9 / 10
	}
	if h <= 0 {
		h = r.H * 9 / 10
	}
	if w > r.W {
		w = r.W
	}
	if h > r.H {
		h = r.H
	}
	x := r.X + (r.W-w)/2
	y := r.Y + (r.H-h)/2
	if x < r.X {
		x = r.X
	}
	if y < r.Y {
		y = r.Y
	}
	p.BaseComponent.SetBounds(component.Rect{X: x, Y: y, W: w, H: h})
}

// Paint renders the popup: titled border frame + content.
func (p *Popup) Paint(buf *buffer.Buffer) {
	bounds := p.Bounds()
	if bounds.W < 3 || bounds.H < 3 {
		return
	}
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H

	// Fill background
	buf.FillRect(buffer.Rect{X: x, Y: y, W: w, H: h}, buffer.Cell{
		Rune:  ' ',
		Width: 1,
		Fg:    p.style.Body.Fg,
		Bg:    p.style.Body.Bg,
	})

	// Border frame
	s := p.style.Border
	buf.SetCell(x, y, buffer.Cell{Rune: '\u250c', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})       // ┌
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '\u2510', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})  // ┐
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '\u2514', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})  // └
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '\u2518', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // ┘

	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+i, y+h-1, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}

	// Title on top border
	if p.title != "" {
		titleText := " " + p.title + " "
		titleW := buffer.StringWidth(titleText)
		startX := x + (w-titleW)/2
		if startX < x+1 {
			startX = x + 1
		}
		buf.DrawText(startX, y, titleText, p.style.Title)
	}

	// Content in inner area
	if p.content != nil {
		p.content.SetBounds(component.Rect{
			X: x + 1,
			Y: y + 1,
			W: w - 2,
			H: h - 2,
		})
		p.content.Paint(buf)
	}
}

// HandleKey processes keyboard input for the popup.
// Esc: close the popup.
func (p *Popup) HandleKey(key interface{}) bool {
	ke, ok := key.(*term.KeyEvent)
	if !ok {
		return false
	}
	if ke.Key == term.KeyEscape {
		p.SetVisible(false)
		return true
	}
	return false
}
