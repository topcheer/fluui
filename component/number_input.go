package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// NumberInput is a numeric input with increment/decrement buttons.
// Useful for forms, configuration panels, and setting parameters.
//
// Features:
//   - Min/max bounds with clamping
//   - Configurable step size
//   - Keyboard: Up/Down to increment/decrement, type digits
//   - Thread-safe
type NumberInput struct {
	BaseComponent
	mu sync.Mutex

	value    int
	min      int
	max      int
	step     int
	suffix   string // optional suffix (e.g., "ms", "tokens")
	prefix   string // optional prefix (e.g., "$")
	focused  bool
}

// NewNumberInput creates a number input with the given value and bounds.
func NewNumberInput(value, min, max int) *NumberInput {
	if max < min {
		max = min
	}
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	return &NumberInput{
		BaseComponent: BaseComponent{id: GenerateID("number")},
		value:         value,
		min:           min,
		max:           max,
		step:          1,
	}
}

// Value returns the current value.
func (n *NumberInput) Value() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.value
}

// SetValue sets the value, clamped to min/max.
func (n *NumberInput) SetValue(v int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.value = clampIntRange(v, n.min, n.max)
}

// SetStep changes the increment/decrement step size.
func (n *NumberInput) SetStep(s int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if s < 1 {
		s = 1
	}
	n.step = s
}

// SetBounds changes the min/max range.
func (n *NumberInput) SetRange(min, max int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if max < min {
		max = min
	}
	n.min = min
	n.max = max
	n.value = clampIntRange(n.value, min, max)
}

// SetSuffix sets an optional suffix displayed after the value.
func (n *NumberInput) SetSuffix(s string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.suffix = s
}

// SetPrefix sets an optional prefix displayed before the value.
func (n *NumberInput) SetPrefix(s string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.prefix = s
}

// Increment adds the step to the value.
func (n *NumberInput) Increment() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.value = clampIntRange(n.value+n.step, n.min, n.max)
}

// Decrement subtracts the step from the value.
func (n *NumberInput) Decrement() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.value = clampIntRange(n.value-n.step, n.min, n.max)
}

// SetFocused controls whether the input is highlighted.
func (n *NumberInput) SetFocused(v bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.focused = v
}

// IsFocused returns whether the input is focused.
func (n *NumberInput) IsFocused() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.focused
}

// HandleKey processes keyboard input.
func (n *NumberInput) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	switch ev.Key {
	case term.KeyUp:
		n.Increment()
		return true
	case term.KeyDown:
		n.Decrement()
		return true
	case term.KeyLeft, term.KeyRight:
		return true
	}
	// Digit input
	if ev.Rune >= '0' && ev.Rune <= '9' {
		n.mu.Lock()
		newVal := n.value*10 + int(ev.Rune-'0')
		clamped := clampIntRange(newVal, n.min, n.max)
		n.value = clamped
		n.mu.Unlock()
		return true
	}
	return false
}

// Measure returns the desired size.
func (n *NumberInput) Measure(cs Constraints) Size {
	n.mu.Lock()
	suffix := n.suffix
	prefix := n.prefix
	n.mu.Unlock()

	w := 8 // "[-12345]" base width
	w += utf8.RuneCountInString(prefix)
	w += utf8.RuneCountInString(suffix)
	maxW := cs.MaxWidth
	if maxW > 0 && w > maxW {
		w = maxW
	}
	if w < 4 {
		w = 4
	}
	return Size{W: w, H: 1}
}

// Paint renders the number input.
func (n *NumberInput) Paint(buf *buffer.Buffer) {
	n.mu.Lock()
	value := n.value
	min, max := n.min, n.max
	suffix := n.suffix
	prefix := n.prefix
	focused := n.focused
	n.mu.Unlock()

	b := n.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	fg := th.Fg
	borderFg := th.Border
	if focused {
		borderFg = th.Accent
	}

	borderStyle := buffer.Style{Fg: borderFg}
	valueStyle := buffer.Style{Fg: fg}
	btnStyle := buffer.Style{Fg: th.Accent}

	x := b.X
	y := b.Y

	// [- prefix value suffix -]
	buf.DrawText(x, y, "[", borderStyle)
	x++

	if prefix != "" {
		x = buf.DrawText(x, y, prefix, valueStyle)
	}

	// Format value with zero allocation
	var vbuf [16]byte
	vb := intToBuf(vbuf[:0], value)
	x = buf.DrawText(x, y, string(vb), valueStyle)

	if suffix != "" {
		x = buf.DrawText(x, y, " "+suffix, valueStyle)
	}

	// Padding to fill width
	for x < b.X+b.W-1 {
		buf.DrawText(x, y, " ", valueStyle)
		x++
	}

	buf.DrawText(x, y, "]", borderStyle)

	// Show range hint below if space allows and we have a second line
	if b.H > 1 {
		hintStyle := buffer.Style{Fg: th.Muted}
		var hbuf [32]byte
		hb := hbuf[:0]
		hb = append(hb, "min "...)
		hb = intToBuf(hb, min)
		hb = append(hb, " max "...)
		hb = intToBuf(hb, max)
		buf.DrawText(b.X, b.Y+1, string(hb), hintStyle)
	}

	// Indicators at far right of the border if focused
	if focused && b.W >= 3 {
		buf.DrawText(b.X, y, "\u25c0", btnStyle) // ◀
		buf.DrawText(b.X+b.W-1, y, "\u25b6", btnStyle) // ▶
	}
}

// intToBuf appends an integer to the byte slice without allocation.
func intToBuf(b []byte, v int) []byte {
	if v == 0 {
		return append(b, '0')
	}
	if v < 0 {
		b = append(b, '-')
		v = -v
	}
	// Convert digits
	var digits [20]byte
	n := 0
	for v > 0 {
		digits[n] = byte('0' + v%10)
		v /= 10
		n++
	}
	// Reverse into output
	for i := n - 1; i >= 0; i-- {
		b = append(b, digits[i])
	}
	return b
}

// clampIntRange clamps v to the [min, max] range.
func clampIntRange(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
