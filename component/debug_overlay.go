package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── DebugOverlay: Performance Debug Information Overlay ───
//
// DebugOverlay renders a compact debug panel showing FPS, component count,
// memory usage, and render time. Useful for development profiling.
//
// Usage:
//
//	d := NewDebugOverlay()
//	d.SetMetrics(60, 142, 8192, 2500) // fps, components, memKB, renderUs
//	d.Paint(buf)

// DebugOverlayStyle holds styling.
type DebugOverlayStyle struct {
	Label buffer.Style
	Value buffer.Style
	Good  buffer.Style
	Warn  buffer.Style
	Bad   buffer.Style
}

// DefaultDebugOverlayStyle returns defaults.
func DefaultDebugOverlayStyle() DebugOverlayStyle {
	return DebugOverlayStyle{
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value: buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Good:  buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Warn:  buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Bad:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
	}
}

// DebugOverlay renders a debug info panel.
type DebugOverlay struct {
	BaseComponent
	mu sync.Mutex

	fps        int
	components int
	memKB      int
	renderUs   int
	style      DebugOverlayStyle
	// cached
	fpsStr   string
	compStr  string
	memStr   string
	renderStr string
	fpsStyle  buffer.Style
}

// NewDebugOverlay creates a DebugOverlay.
func NewDebugOverlay() *DebugOverlay {
	d := &DebugOverlay{style: DefaultDebugOverlayStyle()}
	d.SetID(GenerateID("debug"))
	d.recomputeLocked()
	return d
}

// SetMetrics sets FPS, component count, memory (KB), and render time (microseconds).
func (d *DebugOverlay) SetMetrics(fps, components, memKB, renderUs int) *DebugOverlay {
	d.mu.Lock()
	if fps < 0 { fps = 0 }
	if components < 0 { components = 0 }
	if memKB < 0 { memKB = 0 }
	if renderUs < 0 { renderUs = 0 }
	d.fps = fps
	d.components = components
	d.memKB = memKB
	d.renderUs = renderUs
	d.recomputeLocked()
	d.mu.Unlock()
	return d
}

func (d *DebugOverlay) recomputeLocked() {
	d.fpsStr = itoa(d.fps)
	d.compStr = itoa(d.components)
	d.memStr = itoa(d.memKB) + "KB"
	d.renderStr = itoa(d.renderUs) + "us"

	// Color-code FPS
	if d.fps >= 50 {
		d.fpsStyle = d.style.Good
	} else if d.fps >= 30 {
		d.fpsStyle = d.style.Warn
	} else {
		d.fpsStyle = d.style.Bad
	}
}

// FPS returns the current FPS value.
func (d *DebugOverlay) FPS() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.fps
}

// SetStyle sets custom style.
func (d *DebugOverlay) SetStyle(s DebugOverlayStyle) *DebugOverlay {
	d.mu.Lock()
	d.style = s
	d.recomputeLocked()
	d.mu.Unlock()
	return d
}

// Measure returns preferred size.
func (d *DebugOverlay) Measure(cs Constraints) Size {
	w := 26
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 4}
}

// Paint renders the debug overlay.
func (d *DebugOverlay) Paint(buf *buffer.Buffer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	b := d.Bounds()
	x, y := b.X, b.Y

	labelStyle := d.style.Label
	valueStyle := d.style.Value

	// Row 0: FPS
	col := x
	for _, r := range "FPS " {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range d.fpsStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: d.fpsStyle.Fg, Bg: d.fpsStyle.Bg, Flags: d.fpsStyle.Flags, Width: 1})
		col++
	}

	// Row 1: Components
	col = x
	for _, r := range "Components " {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range d.compStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Row 2: Memory
	col = x
	for _, r := range "Memory " {
		if col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range d.memStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Row 3: Render time
	col = x
	for _, r := range "Render " {
		if col >= buf.Width { break }
		buf.SetCell(col, y+3, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range d.renderStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y+3, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (d *DebugOverlay) Children() []Component { return nil }
