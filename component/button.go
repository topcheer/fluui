package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ButtonVariant defines the visual style of a button.
type ButtonVariant uint8

const (
	ButtonDefault ButtonVariant = iota
	ButtonPrimary
	ButtonSuccess
	ButtonWarning
	ButtonDanger
)

// ButtonStyle holds styling for each button state.
type ButtonStyle struct {
	Fg     buffer.Color
	Bg     buffer.Color
	ActiveFg buffer.Color
	ActiveBg buffer.Color
}

// DefaultButtonStyle returns a Dracula-themed default button style.
func DefaultButtonStyle() ButtonStyle {
	return ButtonStyle{
		Fg:       buffer.NamedColor(buffer.NamedWhite),
		Bg:       buffer.NamedColor(buffer.NamedBrightBlack),
		ActiveFg: buffer.NamedColor(buffer.NamedBlack),
		ActiveBg: buffer.NamedColor(buffer.NamedWhite),
	}
}

// variantStyle returns a style for the given button variant.
func variantStyle(v ButtonVariant) ButtonStyle {
	switch v {
	case ButtonPrimary:
		return ButtonStyle{
			Fg:       buffer.NamedColor(buffer.NamedWhite),
			Bg:       buffer.NamedColor(buffer.NamedBlue),
			ActiveFg: buffer.NamedColor(buffer.NamedBlack),
			ActiveBg: buffer.NamedColor(buffer.NamedBrightBlue),
		}
	case ButtonSuccess:
		return ButtonStyle{
			Fg:       buffer.NamedColor(buffer.NamedBlack),
			Bg:       buffer.NamedColor(buffer.NamedGreen),
			ActiveFg: buffer.NamedColor(buffer.NamedBlack),
			ActiveBg: buffer.NamedColor(buffer.NamedBrightGreen),
		}
	case ButtonWarning:
		return ButtonStyle{
			Fg:       buffer.NamedColor(buffer.NamedBlack),
			Bg:       buffer.NamedColor(buffer.NamedYellow),
			ActiveFg: buffer.NamedColor(buffer.NamedBlack),
			ActiveBg: buffer.NamedColor(buffer.NamedBrightYellow),
		}
	case ButtonDanger:
		return ButtonStyle{
			Fg:       buffer.NamedColor(buffer.NamedWhite),
			Bg:       buffer.NamedColor(buffer.NamedRed),
			ActiveFg: buffer.NamedColor(buffer.NamedWhite),
			ActiveBg: buffer.NamedColor(buffer.NamedBrightRed),
		}
	default:
		return DefaultButtonStyle()
	}
}

// VariantName returns the name of the button variant.
func (v ButtonVariant) String() string {
	switch v {
	case ButtonPrimary:
		return "primary"
	case ButtonSuccess:
		return "success"
	case ButtonWarning:
		return "warning"
	case ButtonDanger:
		return "danger"
	default:
		return "default"
	}
}

// Button is a clickable button with text label and variant styling.
// It supports keyboard activation (Enter/Space) and mouse clicks.
type Button struct {
	BaseComponent

	label   string
	variant ButtonVariant
	style   ButtonStyle
	active  bool // pressed/focused state
	enabled bool

	onClick func()

	mu sync.RWMutex
}

// NewButton creates a button with the given label.
func NewButton(label string) *Button {
	b := &Button{
		label:   label,
		variant: ButtonDefault,
		style:   DefaultButtonStyle(),
		enabled: true,
	}
	b.SetID(GenerateID("button"))
	return b
}

// NewButtonWithVariant creates a button with a specific variant.
func NewButtonWithVariant(label string, variant ButtonVariant) *Button {
	b := NewButton(label)
	b.variant = variant
	b.style = variantStyle(variant)
	return b
}

// Label returns the button label.
func (b *Button) Label() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.label
}

// SetLabel sets the button label.
func (b *Button) SetLabel(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.label = s
}

// Variant returns the button variant.
func (b *Button) Variant() ButtonVariant {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.variant
}

// SetVariant sets the button variant and updates the style.
func (b *Button) SetVariant(v ButtonVariant) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variant = v
	b.style = variantStyle(v)
}

// IsActive returns whether the button is in active (pressed) state.
func (b *Button) IsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.active
}

// SetActive sets the active state.
func (b *Button) SetActive(active bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.active = active
}

// Enabled returns whether the button is enabled.
func (b *Button) Enabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// SetEnabled enables or disables the button.
func (b *Button) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

// SetOnClick sets the click handler.
func (b *Button) SetOnClick(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onClick = fn
}

// SetStyle sets a custom button style.
func (b *Button) SetStyle(s ButtonStyle) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.style = s
}

// Click triggers the onClick handler.
func (b *Button) Click() {
	b.mu.RLock()
	fn := b.onClick
	enabled := b.enabled
	b.mu.RUnlock()
	if enabled && fn != nil {
		fn()
	}
}

// HandleKey handles keyboard input. Enter and Space activate the button.
func (b *Button) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	b.mu.RLock()
	enabled := b.enabled
	b.mu.RUnlock()
	if !enabled {
		return false
	}
	if ev.Key == term.KeyEnter || (ev.Rune == ' ' && ev.Key == term.KeySpace) {
		b.Click()
		return true
	}
	if ev.Rune == ' ' {
		b.Click()
		return true
	}
	return false
}

// HandleMouse handles mouse input. Click within bounds activates the button.
func (b *Button) HandleMouse(x, y int, action string) bool {
	b.mu.RLock()
	enabled := b.enabled
	bounds := b.Bounds()
	b.mu.RUnlock()
	if !enabled {
		return false
	}
	// Check if click is within bounds
	if x < bounds.X || x >= bounds.X+bounds.W || y < bounds.Y || y >= bounds.Y+bounds.H {
		return false
	}
	if action == "click" || action == "release" {
		b.Click()
		return true
	}
	return false
}

// Measure returns the desired size: label length + 4 (padding + border).
func (b *Button) Measure(cs Constraints) Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	w := len([]rune(b.label)) + 4 // 2 padding each side
	if w < 6 {
		w = 6 // minimum width
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the button into the buffer.
func (b *Button) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := b.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w <= 0 {
		return
	}

	fg := b.style.Fg
	bg := b.style.Bg
	if b.active {
		fg = b.style.ActiveFg
		bg = b.style.ActiveBg
	}

	if !b.enabled {
		fg = buffer.NamedColor(buffer.NamedBrightBlack)
		bg = buffer.Color{}
	}

	// Render: " label " with background
	labelRunes := []rune(b.label)
	labelLen := len(labelRunes)

	// Calculate padding
	totalContent := labelLen + 2 // 1 space padding each side
	if totalContent > w {
		// Truncate label to fit
		maxLabel := w - 2
		if maxLabel < 0 {
			maxLabel = 0
		}
		if labelLen > maxLabel {
			labelRunes = labelRunes[:maxLabel]
			labelLen = maxLabel
		}
		totalContent = labelLen + 2
	}

	// Center the content within the button width
	startX := x + (w-totalContent)/2
	if startX < x {
		startX = x
	}

	// Draw left padding
	buf.SetCell(startX, y, buffer.Cell{
		Rune:  ' ',
		Width: 1,
		Fg:    fg,
		Bg:    bg,
	})

	// Draw label
	for i := 0; i < labelLen && startX+1+i < x+w; i++ {
		buf.SetCell(startX+1+i, y, buffer.Cell{
			Rune:  labelRunes[i],
			Width: 1,
			Fg:    fg,
			Bg:    bg,
		})
	}

	// Draw right padding
	rightX := startX + 1 + labelLen
	if rightX < x+w {
		buf.SetCell(rightX, y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    fg,
			Bg:    bg,
		})
	}

	// Fill remaining width with background
	for i := startX; i < x+w; i++ {
		if i < startX || i >= startX+totalContent {
			buf.SetCell(i, y, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Fg:    fg,
				Bg:    bg,
			})
		}
	}
}

// Children returns nil.
func (b *Button) Children() []Component { return nil }
