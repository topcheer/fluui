package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PasswordStrength: Password Strength Meter ───
//
// PasswordStrength evaluates password strength and renders a colored bar
// with a label (Weak/Fair/Good/Strong) and optional requirement checklist.
//
// Usage:
//
//	ps := NewPasswordStrength()
//	ps.SetPassword("MyStr0ng!Pass")
//	ps.Paint(buf)

// PasswordLevel describes strength level.
type PasswordLevel int

const (
	PWVeryWeak PasswordLevel = iota
	PWWeak
	PWFair
	PWGood
	PWStrong
)

// PasswordStrengthStyle holds styling.
type PasswordStrengthStyle struct {
	Bar    [5]buffer.Style // [veryweak, weak, fair, good, strong]
	Label  buffer.Style
	Border buffer.Style
}

// DefaultPasswordStrengthStyle returns defaults.
func DefaultPasswordStrengthStyle() PasswordStrengthStyle {
	vw := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	wk := buffer.Style{Fg: buffer.RGB(249, 115, 22)}
	fr := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	gd := buffer.Style{Fg: buffer.RGB(132, 204, 22)}
	st := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return PasswordStrengthStyle{Bar: [5]buffer.Style{vw, wk, fr, gd, st}, Label: label, Border: border}
}

// pwLevelText returns text for a strength level.
func pwLevelText(l PasswordLevel) string {
	switch l {
	case PWVeryWeak: return "Very Weak"
	case PWWeak: return "Weak"
	case PWFair: return "Fair"
	case PWGood: return "Good"
	case PWStrong: return "Strong"
	default: return "Unknown"
	}
}

// evaluatePassword checks password strength criteria.
func evaluatePassword(pw string) PasswordLevel {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false
	length := 0
	for _, c := range pw {
		length++
		if c >= 'a' && c <= 'z' { hasLower = true }
		if c >= 'A' && c <= 'Z' { hasUpper = true }
		if c >= '0' && c <= '9' { hasDigit = true }
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) { hasSymbol = true }
	}
	score := 0
	if length >= 8 { score++ }
	if length >= 12 { score++ }
	if hasLower && hasUpper { score++ }
	if hasDigit { score++ }
	if hasSymbol { score++ }
	if score <= 1 { return PWVeryWeak }
	if score <= 2 { return PWWeak }
	if score <= 3 { return PWFair }
	if score <= 4 { return PWGood }
	return PWStrong
}

// PasswordStrength renders a password strength meter.
type PasswordStrength struct {
	BaseComponent
	mu sync.Mutex

	password string
	style    PasswordStrengthStyle
	// cached
	level     PasswordLevel
	levelText string
}

// NewPasswordStrength creates a PasswordStrength.
func NewPasswordStrength() *PasswordStrength {
	ps := &PasswordStrength{style: DefaultPasswordStrengthStyle()}
	ps.SetID(GenerateID("pwstrength"))
	return ps
}

// SetPassword sets the password to evaluate (caches level + text).
func (ps *PasswordStrength) SetPassword(pw string) *PasswordStrength {
	ps.mu.Lock()
	ps.password = pw
	ps.level = evaluatePassword(pw)
	ps.levelText = pwLevelText(ps.level)
	ps.mu.Unlock()
	return ps
}

// Password returns the current password.
func (ps *PasswordStrength) Password() string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.password
}

// Level returns the evaluated strength level.
func (ps *PasswordStrength) Level() PasswordLevel {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.level
}

// SetStyle sets custom style.
func (ps *PasswordStrength) SetStyle(s PasswordStrengthStyle) *PasswordStrength {
	ps.mu.Lock()
	ps.style = s
	ps.mu.Unlock()
	return ps
}

// Measure returns the preferred size.
func (ps *PasswordStrength) Measure(cs Constraints) Size {
	w := 30
	h := 3
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the password strength meter into the buffer.
func (ps *PasswordStrength) Paint(buf *buffer.Buffer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	b := ps.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 30 }
	if h < 3 { h = 3 }

	bs := ps.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	levelIdx := int(ps.level)
	if levelIdx < 0 || levelIdx > 4 { levelIdx = 0 }
	barStyle := ps.style.Bar[levelIdx]
	labelStyle := ps.style.Label

	// Bar segments (5 segments, filled up to level+1)
	col := x + 1
	segments := 5
	for i := 0; i < segments; i++ {
		if col >= x+w-1 || col >= buf.Width { break }
		var ch rune
		var style buffer.Style
		if i <= levelIdx {
			ch = '█'
			style = barStyle
		} else {
			ch = '░'
			style = labelStyle
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		col++
	}

	// Label text after bar
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+1, buffer.Cell{Rune: ' ', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
	for _, r := range ps.levelText {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ps *PasswordStrength) Children() []Component { return nil }
