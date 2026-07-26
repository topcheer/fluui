package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DividerOrientation controls divider direction.
type DividerOrientation int

const (
	DividerHorizontal DividerOrientation = iota
	DividerVertical
)

// Divider renders a horizontal or vertical separator line with an optional
// inline label. Simpler than Rule — designed for section breaks in forms
// and settings panels.
//
// Thread-safe.
type Divider struct {
	BaseComponent
	mu      sync.RWMutex
	label   string
	orient  DividerOrientation
	char    rune
	customFg buffer.Color
}

// NewDivider creates a horizontal divider with optional label.
func NewDivider(label string) *Divider {
	return &Divider{
		BaseComponent: BaseComponent{id: GenerateID("divider")},
		label:         label,
		orient:        DividerHorizontal,
		char:          '\u2500', // ─
	}
}

func (d *Divider) Label() string { d.mu.RLock(); defer d.mu.RUnlock(); return d.label }
func (d *Divider) SetLabel(s string) { d.mu.Lock(); defer d.mu.Unlock(); d.label = s }

func (d *Divider) Orientation() DividerOrientation { d.mu.RLock(); defer d.mu.RUnlock(); return d.orient }
func (d *Divider) SetOrientation(o DividerOrientation) { d.mu.Lock(); defer d.mu.Unlock(); d.orient = o }

func (d *Divider) Char() rune { d.mu.RLock(); defer d.mu.RUnlock(); return d.char }
func (d *Divider) SetChar(r rune) { d.mu.Lock(); defer d.mu.Unlock(); d.char = r }

func (d *Divider) SetColor(c buffer.Color) { d.mu.Lock(); defer d.mu.Unlock(); d.customFg = c }

func (d *Divider) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.orient == DividerVertical {
		w := 1
		h := cs.MaxHeight
		if h <= 0 { h = 1 }
		if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
		if w < 1 { w = 1 }
		if h < 1 { h = 1 }
		return Size{W: w, H: h}
	}
	// Horizontal
	w := cs.MaxWidth
	if w <= 0 { w = 20 }
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	return Size{W: w, H: h}
}

func (d *Divider) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bounds := d.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	fg := d.customFg
	if fg.Type == buffer.ColorNone { fg = tt.Muted }

	if d.orient == DividerVertical {
		for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
			buf.SetCell(bounds.X, y, buffer.Cell{Rune: '\u2502', Width: 1, Fg: fg}) // │
		}
		return
	}

	// Horizontal
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	if d.label == "" {
		for ; x < maxX; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: d.char, Width: 1, Fg: fg})
		}
		return
	}

	// Label: ── label ──
	labelW := len(d.label) + 2 // " label "
	avail := maxX - x
	if labelW >= avail {
		// Just draw the label
		buf.DrawText(x, y, " "+d.label+" ", buffer.Style{Fg: fg})
		return
	}
	dashes := avail - labelW
	left := dashes / 2

	for i := 0; i < left && x < maxX; i++ {
		buf.SetCell(x, y, buffer.Cell{Rune: d.char, Width: 1, Fg: fg})
		x++
	}
	x = buf.DrawText(x, y, " "+d.label+" ", buffer.Style{Fg: fg})
	for ; x < maxX; x++ {
		buf.SetCell(x, y, buffer.Cell{Rune: d.char, Width: 1, Fg: fg})
	}
}
