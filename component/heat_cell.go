package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── HeatCell: Single Cell Heat Display ───
//
// HeatCell renders a single colored cell representing a heat value.
// Color maps from cold (blue) to hot (red) across 8 levels.
// Useful as building blocks for heatmap grids and status walls.
//
// Usage:
//
//	hc := NewHeatCell()
//	hc.SetValue(75) // 0-100
//	hc.Paint(buf)

type HeatCellStyle struct {
	Cold  buffer.Style
	Cool  buffer.Style
	Warm  buffer.Style
	Hot   buffer.Style
	Empty buffer.Style
	Label buffer.Style
}

func DefaultHeatCellStyle() HeatCellStyle {
	return HeatCellStyle{
		Cold:  buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Cool:  buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Warm:  buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Hot:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Empty: buffer.Style{Fg: buffer.RGB(30, 41, 59)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

var heatChars = [9]rune{'░', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// HeatCell renders a single heat cell.
type HeatCell struct {
	BaseComponent
	mu sync.Mutex

	value int // 0-100, 0=empty
	label string
	style HeatCellStyle
	// cached
	level    int
	curRune  rune
	curStyle buffer.Style
}

func NewHeatCell() *HeatCell {
	hc := &HeatCell{style: DefaultHeatCellStyle()}
	hc.SetID(GenerateID("heatcell"))
	hc.recomputeLocked()
	return hc
}

func (hc *HeatCell) SetValue(v int) *HeatCell {
	hc.mu.Lock()
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	hc.value = v
	hc.recomputeLocked()
	hc.mu.Unlock()
	return hc
}

func (hc *HeatCell) SetLabel(l string) *HeatCell {
	hc.mu.Lock()
	hc.label = l
	hc.mu.Unlock()
	return hc
}

func (hc *HeatCell) recomputeLocked() {
	if hc.value == 0 {
		hc.level = 0
		hc.curRune = '░'
		hc.curStyle = hc.style.Empty
		return
	}
	hc.level = hc.value * 8 / 100
	if hc.level > 8 {
		hc.level = 8
	}
	if hc.level < 1 {
		hc.level = 1
	}
	hc.curRune = heatChars[hc.level]

	switch {
	case hc.level <= 2:
		hc.curStyle = hc.style.Cold
	case hc.level <= 4:
		hc.curStyle = hc.style.Cool
	case hc.level <= 6:
		hc.curStyle = hc.style.Warm
	default:
		hc.curStyle = hc.style.Hot
	}
}

func (hc *HeatCell) Value() int {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	return hc.value
}

func (hc *HeatCell) SetStyle(s HeatCellStyle) *HeatCell {
	hc.mu.Lock()
	hc.style = s
	hc.recomputeLocked()
	hc.mu.Unlock()
	return hc
}

func (hc *HeatCell) Measure(cs Constraints) Size {
	return Size{W: 4, H: 1}
}

func (hc *HeatCell) Paint(buf *buffer.Buffer) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	b := hc.Bounds()
	x, y := b.X, b.Y

	if x < buf.Width {
		buf.SetCell(x, y, buffer.Cell{Rune: hc.curRune, Fg: hc.curStyle.Fg, Bg: hc.curStyle.Bg, Flags: hc.curStyle.Flags, Width: 1})
	}
	if hc.label != "" && x+1 < buf.Width {
		for i, r := range hc.label {
			if x+1+i >= buf.Width {
				break
			}
			buf.SetCell(x+1+i, y, buffer.Cell{Rune: r, Fg: hc.style.Label.Fg, Bg: hc.style.Label.Bg, Flags: hc.style.Label.Flags, Width: 1})
		}
	}
}

func (hc *HeatCell) Children() []Component { return nil }
