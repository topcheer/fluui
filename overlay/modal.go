package overlay

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ModalStyle holds the visual styling for a Modal.
type ModalStyle struct {
	// Border is the style for the modal border frame.
	Border buffer.Style
	// Title is the style for the title text on the border.
	Title buffer.Style
	// Body is the style for the body background.
	Body buffer.Style
	// Mask is the style for the semi-transparent mask covering the screen.
	Mask buffer.Style
	// ButtonNormal is the style for unselected buttons.
	ButtonNormal buffer.Style
	// ButtonSelected is the style for the highlighted button.
	ButtonSelected buffer.Style
}

// DefaultModalStyle returns a reasonable default style.
func DefaultModalStyle() ModalStyle {
	return ModalStyle{
		Border:        buffer.DefaultStyle,
		Title:         buffer.DefaultStyle.AddFlags(buffer.Bold),
		Body:          buffer.DefaultStyle,
		Mask:          buffer.DefaultStyle.WithBg(theme.Get().MaskBg),
		ButtonNormal:  buffer.DefaultStyle,
		ButtonSelected: buffer.DefaultStyle.AddFlags(buffer.Reverse),
	}
}

// Modal is a centered modal dialog with a semi-transparent mask,
// a titled border, body content, and a button bar at the bottom.
type Modal struct {
	BaseOverlay
	title    string
	body     component.Component
	buttons  []string
	style    ModalStyle
	width    int
	height   int
	selected int
}

// NewModal creates a centered modal dialog.
// The modal is modal=true, z-index=100 by default.
func NewModal(id, title string, body component.Component, buttons []string) *Modal {
	return &Modal{
		BaseOverlay: NewBaseOverlay(id, 100, true),
		title:       title,
		body:        body,
		buttons:     buttons,
		style:       DefaultModalStyle(),
	}
}

// SetStyle overrides the default modal style.
func (m *Modal) SetStyle(s ModalStyle) {
	m.style = s
}

// SelectedButton returns the index of the currently selected button.
func (m *Modal) SelectedButton() int {
	return m.selected
}

// Measure computes the modal's desired size.
// Defaults to 50% of the screen, clamped to [20, 80] x [7, 40].
func (m *Modal) Measure(cs component.Constraints) component.Size {
	maxW := cs.MaxWidth
	maxH := cs.MaxHeight
	if maxW <= 0 {
		maxW = 80
	}
	if maxH <= 0 {
		maxH = 24
	}

	w := maxW / 2
	if w < 20 {
		w = 20
	}
	if w > 80 {
		w = 80
	}

	// Body height + 2 (border) + 1 (button bar)
	bodyH := 3
	if m.body != nil {
		bodySize := m.body.Measure(component.Bounded(w-2, maxH-4))
		bodyH = bodySize.H
		if bodyH < 1 {
			bodyH = 1
		}
	}
	h := bodyH + 3 // top border + body + button bar + bottom border
	if h < 7 {
		h = 7
	}
	if h > maxH {
		h = maxH
	}

	m.width = w
	m.height = h
	return component.Size{W: w, H: h}
}

// SetBounds centers the modal within the given rect.
func (m *Modal) SetBounds(r component.Rect) {
	w := m.width
	h := m.height
	if w <= 0 {
		w = r.W / 2
	}
	if h <= 0 {
		h = r.H / 2
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
	m.BaseComponent.SetBounds(component.Rect{X: x, Y: y, W: w, H: h})
}

// Paint renders the modal: mask, border with title, body, and button bar.
func (m *Modal) Paint(buf *buffer.Buffer) {
	bounds := m.Bounds()
	if bounds.W < 3 || bounds.H < 5 {
		return
	}
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H

	// 1. Draw semi-transparent mask over the full buffer
	maskCell := buffer.Cell{Rune: ' ', Width: 1, Bg: m.style.Mask.Bg}
	for my := 0; my < buf.Height; my++ {
		for mx := 0; mx < buf.Width; mx++ {
			// Skip the modal interior (will be drawn below)
			if mx >= x && mx < x+w && my >= y && my < y+h {
				continue
			}
			buf.SetCell(mx, my, maskCell)
		}
	}

	// 2. Fill modal background
	buf.FillRect(buffer.Rect{X: x, Y: y, W: w, H: h}, buffer.Cell{
		Rune:  ' ',
		Width: 1,
		Fg:    m.style.Body.Fg,
		Bg:    m.style.Body.Bg,
	})

	// 3. Draw border frame
	s := m.style.Border
	// Corners
	buf.SetCell(x, y, buffer.Cell{Rune: '\u256d', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})       // ╭
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '\u256e', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})  // ╮
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '\u2570', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})  // ╰
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '\u256f', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // ╯

	// Top and bottom edges
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+i, y+h-1, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}

	// Left and right edges
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}

	// 4. Draw title on the top border
	if m.title != "" {
		titleText := " " + m.title + " "
		titleW := buffer.StringWidth(titleText)
		startX := x + (w-titleW)/2
		if startX < x+1 {
			startX = x + 1
		}
		buf.DrawText(startX, y, titleText, m.style.Title)
	}

	// 5. Layout body in inner area (below border top, above button bar)
	innerY := y + 1
	innerH := h - 3 // subtract top border + button bar + bottom border
	if innerH < 1 {
		innerH = 1
	}
	if m.body != nil {
		m.body.SetBounds(component.Rect{
			X: x + 1,
			Y: innerY,
			W: w - 2,
			H: innerH,
		})
		m.body.Paint(buf)
	}

	// 6. Draw button bar at the bottom
	btnY := y + h - 2
	if len(m.buttons) > 0 {
		// Calculate total button width with spacing
		totalBtnW := 0
		for _, btn := range m.buttons {
			totalBtnW += buffer.StringWidth(btn) + 4 // [ btn ]
		}
		totalBtnW += len(m.buttons) - 1 // spaces between buttons

		btnX := x + (w-totalBtnW)/2
		if btnX < x+1 {
			btnX = x + 1
		}

		for i, btn := range m.buttons {
			label := "[ " + btn + " ]"
			labelW := buffer.StringWidth(label)
			if i > 0 {
				btnX++ // space between buttons
			}
			style := m.style.ButtonNormal
			if i == m.selected {
				style = m.style.ButtonSelected
			}
			buf.DrawText(btnX, btnY, label, style)
			btnX += labelW
		}
	}
}

// HandleKey processes keyboard input for the modal.
// Left/Right: cycle button selection.
// Enter: confirm (hide modal).
// Esc: cancel (hide modal).
func (m *Modal) HandleKey(key interface{}) bool {
	ke, ok := key.(*term.KeyEvent)
	if !ok {
		return false
	}

	switch ke.Key {
	case term.KeyLeft:
		if len(m.buttons) > 0 {
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.buttons) - 1
			}
		}
		return true
	case term.KeyRight:
		if len(m.buttons) > 0 {
			m.selected++
			if m.selected >= len(m.buttons) {
				m.selected = 0
			}
		}
		return true
	case term.KeyEnter:
		m.SetVisible(false)
		return true
	case term.KeyEscape:
		m.SetVisible(false)
		return true
	}
	return false
}
