package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── VolumeMeter: Audio Volume Level Meter ───
//
// VolumeMeter renders a horizontal volume level meter with segment-based
// fill and optional mute indicator. Color shifts from green (low) to
// red (high/clipping).
//
// Usage:
//
//	vm := NewVolumeMeter()
//	vm.SetLevel(75) // 75%
//	vm.SetMuted(false)
//	vm.Paint(buf)

// VolumeMeterStyle holds styling.
type VolumeMeterStyle struct {
	Low    buffer.Style
	Medium buffer.Style
	High   buffer.Style
	Muted  buffer.Style
	Label  buffer.Style
	Value  buffer.Style
}

// DefaultVolumeMeterStyle returns defaults.
func DefaultVolumeMeterStyle() VolumeMeterStyle {
	return VolumeMeterStyle{
		Low:    buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Medium: buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		High:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Muted:  buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:  buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

const volumeMeterSegments = 20

// VolumeMeter renders a volume level meter.
type VolumeMeter struct {
	BaseComponent
	mu sync.Mutex

	level  int // 0-100
	muted  bool
	width  int
	style  VolumeMeterStyle
	// cached
	levelStr  string
	barFill   int
	curStyle  buffer.Style
	displayStr string
}

// NewVolumeMeter creates a VolumeMeter.
func NewVolumeMeter() *VolumeMeter {
	vm := &VolumeMeter{width: 28, style: DefaultVolumeMeterStyle()}
	vm.SetID(GenerateID("volmeter"))
	vm.recomputeLocked()
	return vm
}

// SetLevel sets the volume level (0-100).
func (vm *VolumeMeter) SetLevel(n int) *VolumeMeter {
	vm.mu.Lock()
	if n < 0 { n = 0 }
	if n > 100 { n = 100 }
	vm.level = n
	vm.recomputeLocked()
	vm.mu.Unlock()
	return vm
}

// SetMuted toggles the mute state.
func (vm *VolumeMeter) SetMuted(m bool) *VolumeMeter {
	vm.mu.Lock()
	vm.muted = m
	vm.recomputeLocked()
	vm.mu.Unlock()
	return vm
}

func (vm *VolumeMeter) recomputeLocked() {
	vm.levelStr = itoa(vm.level) + "%"
	vm.barFill = vm.level * volumeMeterSegments / 100

	if vm.muted {
		vm.curStyle = vm.style.Muted
		vm.displayStr = "MUTED"
	} else {
		if vm.level >= 80 {
			vm.curStyle = vm.style.High
		} else if vm.level >= 50 {
			vm.curStyle = vm.style.Medium
		} else {
			vm.curStyle = vm.style.Low
		}
		vm.displayStr = vm.levelStr
	}
}

// Level returns the current volume level.
func (vm *VolumeMeter) Level() int {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.level
}

// SetWidth sets the meter width.
func (vm *VolumeMeter) SetWidth(w int) *VolumeMeter {
	vm.mu.Lock()
	if w < 10 { w = 10 }
	vm.width = w
	vm.mu.Unlock()
	return vm
}

// SetStyle sets custom style.
func (vm *VolumeMeter) SetStyle(s VolumeMeterStyle) *VolumeMeter {
	vm.mu.Lock()
	vm.style = s
	vm.recomputeLocked()
	vm.mu.Unlock()
	return vm
}

// Measure returns preferred size.
func (vm *VolumeMeter) Measure(cs Constraints) Size {
	w := vm.width + 6
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the volume meter.
func (vm *VolumeMeter) Paint(buf *buffer.Buffer) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	b := vm.Bounds()
	x, y := b.X, b.Y

	labelStyle := vm.style.Label
	barStyle := vm.curStyle
	valueStyle := vm.style.Value

	col := x

	// Volume icon
	var icon rune
	if vm.muted {
		icon = '🔇'
	} else if vm.level == 0 {
		icon = '🔇'
	} else if vm.level < 33 {
		icon = '🔈'
	} else if vm.level < 66 {
		icon = '🔉'
	} else {
		icon = '🔊'
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: icon, Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Segmented bar
	for i := 0; i < vm.barFill; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
	for i := vm.barFill; i < volumeMeterSegments; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Value label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range vm.displayStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (vm *VolumeMeter) Children() []Component { return nil }
