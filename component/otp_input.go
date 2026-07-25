package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// OTPInput is a one-time password / PIN input with individual digit boxes.
// Common for 2FA, verification codes, and security PINs.
//
// Thread-safe. Zero-alloc Paint.
type OTPInput struct {
	BaseComponent
	mu      sync.Mutex
	length  int
	values  []rune
	focused int
	filled  bool
}

// NewOTPInput creates an OTP input with the given number of digits.
func NewOTPInput(length int) *OTPInput {
	if length < 1 {
		length = 1
	}
	if length > 16 {
		length = 16
	}
	return &OTPInput{
		BaseComponent: BaseComponent{id: GenerateID("otp")},
		length:        length,
		values:        make([]rune, length),
		focused:       0,
	}
}

// Length returns the number of digit boxes.
func (o *OTPInput) Length() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.length
}

// Value returns the current input as a string.
func (o *OTPInput) Value() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return string(o.values)
}

// SetValue sets the input from a string.
func (o *OTPInput) SetValue(s string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := 0; i < o.length; i++ {
		o.values[i] = 0
	}
	idx := 0
	for _, r := range s {
		if idx >= o.length {
			break
		}
		o.values[idx] = r
		idx++
	}
	o.focused = idx
	if o.focused >= o.length {
		o.focused = o.length - 1
	}
	o.filled = idx >= o.length
}

// IsFilled returns true when all boxes have a character.
func (o *OTPInput) IsFilled() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.filled
}

// Clear resets all boxes.
func (o *OTPInput) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := range o.values {
		o.values[i] = 0
	}
	o.focused = 0
	o.filled = false
}

// HandleKey processes keyboard input.
func (o *OTPInput) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	switch ev.Key {
	case term.KeyBackspace, term.KeyLeft:
		o.mu.Lock()
		if o.focused > 0 {
			if o.values[o.focused] == 0 {
				o.focused--
			}
			o.values[o.focused] = 0
		} else {
			o.values[0] = 0
		}
		o.filled = false
		o.mu.Unlock()
		return true
	case term.KeyRight:
		o.mu.Lock()
		if o.focused < o.length-1 {
			o.focused++
		}
		o.mu.Unlock()
		return true
	}
	// Digit or letter input
	if ev.Rune != 0 && (ev.Rune >= '0' && ev.Rune <= '9' || ev.Rune >= 'A' && ev.Rune <= 'Z' || ev.Rune >= 'a' && ev.Rune <= 'z') {
		o.mu.Lock()
		o.values[o.focused] = ev.Rune
		if o.focused < o.length-1 {
			o.focused++
		}
		// Check filled
		filled := true
		for _, v := range o.values {
			if v == 0 {
				filled = false
				break
			}
		}
		o.filled = filled
		o.mu.Unlock()
		return true
	}
	return false
}

// Measure returns the desired size.
func (o *OTPInput) Measure(cs Constraints) Size {
	o.mu.Lock()
	length := o.length
	o.mu.Unlock()
	w := length*3 + 1 // box width 2 + gap 1
	if w < 1 {
		w = 1
	}
	return Size{W: w, H: 1}
}

// Paint renders the OTP input.
func (o *OTPInput) Paint(buf *buffer.Buffer) {
	o.mu.Lock()
	values := make([]rune, len(o.values))
	copy(values, o.values)
	focused := o.focused
	o.mu.Unlock()

	b := o.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	boxStyle := buffer.Style{Fg: th.Border}
	activeStyle := buffer.Style{Fg: th.Accent}
	fillStyle := buffer.Style{Fg: th.Fg}
	mutedStyle := buffer.Style{Fg: th.Muted}

	x := b.X
	for i, v := range values {
		// Draw box borders: [X]
		style := boxStyle
		if i == focused {
			style = activeStyle
		}

		// Left bracket
		buf.DrawText(x, b.Y, "[", style)
		x++

		// Value or cursor
		if v != 0 {
			buf.DrawText(x, b.Y, string(v), fillStyle)
		} else if i == focused {
			buf.DrawText(x, b.Y, "\u2588", mutedStyle) // █ cursor
		} else {
			buf.DrawText(x, b.Y, " ", mutedStyle)
		}
		x++

		// Right bracket
		buf.DrawText(x, b.Y, "]", style)
		x += 2 // bracket + gap
	}
}

