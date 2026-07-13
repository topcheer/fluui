package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Switch: On/Off Toggle Widget ───
//
// Switch is a visual on/off toggle (like iOS toggle switch).
// Distinct from Checkbox in visual style — a slider rather than a box.
//
// Usage:
//	sw := NewSwitch()
//	sw.SetOn(true)
//	sw.Toggle()
//	if sw.IsOn() { ... }
//	sw.SetOnChange(func(on bool) { ... })

// Switch is a visual toggle switch widget.
type Switch struct {
	mu       sync.RWMutex
	BaseComponent
	on       bool
	label    string
	style    SwitchStyle
	onChange func(bool)
}

// SwitchStyle holds colors for the switch.
type SwitchStyle struct {
	OnBg       buffer.Color
	OffBg      buffer.Color
	KnobFg     buffer.Color
	LabelFg    buffer.Color
	DisabledFg buffer.Color
}

func defaultSwitchStyle() SwitchStyle {
	return SwitchStyle{
		OnBg:       buffer.RGB(80, 250, 123), // Dracula green
		OffBg:      buffer.RGB(68, 71, 90),   // Dracula current line
		KnobFg:     buffer.NamedColor(buffer.NamedWhite),
		LabelFg:    buffer.NamedColor(buffer.NamedWhite),
		DisabledFg: buffer.RGB(98, 114, 164), // Dracula comment
	}
}

// NewSwitch creates a switch with default style and optional label.
func NewSwitch(label string) *Switch {
	return &Switch{
		label: label,
		style: defaultSwitchStyle(),
	}
}

// IsOn returns the current state.
func (s *Switch) IsOn() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.on
}

// SetOn sets the state explicitly.
func (s *Switch) SetOn(on bool) {
	s.mu.Lock()
	s.on = on
	s.mu.Unlock()
	s.notifyChange()
}

// Toggle flips the state.
func (s *Switch) Toggle() {
	s.mu.Lock()
	s.on = !s.on
	s.mu.Unlock()
	s.notifyChange()
}

// SetLabel sets the label text.
func (s *Switch) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

// Label returns the label text.
func (s *Switch) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetStyle customizes colors.
func (s *Switch) SetStyle(style SwitchStyle) {
	s.mu.Lock()
	s.style = style
	s.mu.Unlock()
}

// SetOnChange sets a callback fired on state change.
func (s *Switch) SetOnChange(fn func(bool)) {
	s.mu.Lock()
	s.onChange = fn
	s.mu.Unlock()
}

func (s *Switch) notifyChange() {
	s.mu.RLock()
	cb := s.onChange
	on := s.on
	s.mu.RUnlock()
	if cb != nil {
		cb(on)
	}
}

// HandleKey toggles on Enter or Space.
func (s *Switch) HandleKey(ev *term.KeyEvent) bool {
	if ev.Key == term.KeyEnter || ev.Key == term.KeySpace || ev.Rune == ' ' {
		s.Toggle()
		return true
	}
	return false
}

func (s *Switch) Measure(constraints Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	labelW := len([]rune(s.label))
	// Switch visual: "[ ON ]" or "[ OFF ]" = 8 chars + label
	w := 8
	if s.label != "" {
		w = labelW + 1 + 8 // label + space + switch
	}
	if w > constraints.MaxWidth && constraints.MaxWidth > 0 {
		w = constraints.MaxWidth
	}
	return Size{W: w, H: 1}
}

func (s *Switch) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	x := bounds.X
	y := bounds.Y

	// Draw label if present
	if s.label != "" {
		for i, r := range s.label {
			if x+i >= bounds.X+bounds.W {
				break
			}
			buf.SetCell(x+i, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    s.style.LabelFg,
			})
		}
		x += len([]rune(s.label)) + 1
	}

	// Draw switch: [ ON ] or [ OFF ]
	bg := s.style.OffBg
	text := " OFF "
	if s.on {
		bg = s.style.OnBg
		text = " ON  "
	}

	// Bracket
	buf.SetCell(x, y, buffer.Cell{Rune: '[', Width: 1, Fg: s.style.KnobFg, Bg: bg})
	for i, r := range text {
		if x+1+i >= bounds.X+bounds.W {
			break
		}
		buf.SetCell(x+1+i, y, buffer.Cell{Rune: r, Width: 1, Fg: s.style.KnobFg, Bg: bg})
	}
	if x+6 < bounds.X+bounds.W {
		buf.SetCell(x+6, y, buffer.Cell{Rune: ']', Width: 1, Fg: s.style.KnobFg, Bg: bg})
	}
}