package component

import (
	"strconv"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ToastLevel controls the visual style of a Toast.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastPosition controls where the toast appears.
type ToastPosition int

const (
	ToastTopLeft ToastPosition = iota
	ToastTopRight
	ToastBottomLeft
	ToastBottomRight
	ToastTopCenter
	ToastBottomCenter
)

// Toast is a transient notification popup that auto-dismisses after a timeout.
// Similar to Android Toasts or VS Code notifications.
//
// Thread-safe.
type Toast struct {
	BaseComponent
	mu       sync.RWMutex
	message  string
	level    ToastLevel
	duration time.Duration
	pos      ToastPosition
	dismissed bool
}

// NewToast creates a toast with the given message and level.
func NewToast(message string, level ToastLevel) *Toast {
	return &Toast{
		BaseComponent: BaseComponent{id: GenerateID("toast")},
		message:       message,
		level:         level,
		duration:      3 * time.Second,
		pos:           ToastBottomRight,
	}
}

// Message returns the toast message.
func (t *Toast) Message() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.message
}

// SetMessage updates the toast message.
func (t *Toast) SetMessage(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.message = s
}

// Level returns the toast level.
func (t *Toast) Level() ToastLevel {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.level
}

// SetLevel updates the toast level.
func (t *Toast) SetLevel(l ToastLevel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.level = l
}

// Duration returns the auto-dismiss duration.
func (t *Toast) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.duration
}

// SetDuration sets the auto-dismiss duration.
func (t *Toast) SetDuration(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.duration = d
}

// Position returns the toast position.
func (t *Toast) Position() ToastPosition {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.pos
}

// SetPosition updates the toast position.
func (t *Toast) SetPosition(p ToastPosition) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pos = p
}

// Dismissed returns whether the toast has been dismissed.
func (t *Toast) Dismissed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.dismissed
}

// Dismiss marks the toast as dismissed.
func (t *Toast) Dismiss() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dismissed = true
}

// resolveColorsLocked returns (fg, bg) based on level.
func (t *Toast) resolveColorsLocked() (buffer.Color, buffer.Color) {
	tt := theme.Get()
	switch t.level {
	case ToastSuccess:
		return tt.Bg, tt.Success
	case ToastWarning:
		return tt.Bg, tt.Warning
	case ToastError:
		return tt.Bg, tt.Error
	default:
		return tt.Bg, tt.Accent
	}
}

// resolvePositionLocked computes the (x, y) for the toast within the given screen bounds.
func (t *Toast) resolvePositionLocked(screenW, screenH, toastW, toastH int) (int, int) {
	switch t.pos {
	case ToastTopLeft:
		return 0, 0
	case ToastTopRight:
		return screenW - toastW, 0
	case ToastBottomLeft:
		return 0, screenH - toastH
	case ToastTopCenter:
		return (screenW - toastW) / 2, 0
	case ToastBottomCenter:
		return (screenW - toastW) / 2, screenH - toastH
	default: // BottomRight
		return screenW - toastW, screenH - toastH
	}
}

// Measure returns the preferred size.
func (t *Toast) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()

	msgW := buffer.StringWidth(t.message)
	// " ⚠ message " with padding
	w := msgW + 4
	if w < 6 {
		w = 6
	}
	h := 1

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the toast. Zero allocations.
func (t *Toast) Paint(buf *buffer.Buffer) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.dismissed {
		return
	}

	bounds := t.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	fg, bg := t.resolveColorsLocked()

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	style := buffer.Style{Fg: fg, Bg: bg, Flags: buffer.Bold}

	// Left padding + icon space
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}

	// Level icon
	var icon string
	switch t.level {
	case ToastSuccess:
		icon = "\u2713 " // ✓
	case ToastWarning:
		icon = "\u26a0 " // ⚠
	case ToastError:
		icon = "\u2717 " // ✗
	default:
		icon = "\u2139 " // ℹ
	}
	x = buf.DrawText(x, y, icon, style)

	// Message (truncated to fit)
	msgW := buffer.StringWidth(t.message)
	avail := maxX - x - 1 // -1 for right padding
	if avail < 1 {
		avail = 1
	}
	if msgW > avail {
		// Truncate using rune scanning
	 drawn := 0
		for _, r := range t.message {
			if drawn >= avail-1 {
				buf.SetCell(x, y, buffer.Cell{Rune: '\u2026', Width: 1, Fg: fg, Bg: bg, Flags: buffer.Bold})
				x++
				break
			}
			rw := buffer.RuneWidth(r)
			if x+rw > maxX-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: uint8(rw), Fg: fg, Bg: bg, Flags: buffer.Bold})
			if rw == 2 && x+1 < maxX {
				buf.SetCell(x+1, y, buffer.Cell{Rune: 0, Width: 0, Bg: bg})
			}
			x += rw
			drawn++
		}
	} else {
		x = buf.DrawText(x, y, t.message, style)
	}

	// Right padding
	for x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}
}

// FormatDuration formats a duration for compact display (e.g., "3s", "1.5s").
func formatToastDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		var buf [8]byte
		b := buf[:0]
		b = strconv.AppendInt(b, ms, 10)
		b = append(b, "ms"...)
		return string(b)
	}
	secs := d.Seconds()
	var buf [8]byte
	b := buf[:0]
	b = strconv.AppendFloat(b, secs, 'f', 1, 64)
	b = append(b, 's')
	return string(b)
}
