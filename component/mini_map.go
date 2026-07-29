package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MiniMap: Compact Content Overview Map ───
//
// MiniMap renders a miniature overview of a larger content area, showing
// visible viewport position as a highlighted block within the full content.
// Useful for scrollable content to show position context.
//
// Usage:
//
//	mm := NewMiniMap()
//	mm.SetContent(500, 80, 100) // totalLines=500, viewStart=80, viewHeight=100
//	mm.Paint(buf)

// MiniMapStyle holds styling.
type MiniMapStyle struct {
	Full    buffer.Style
	Visible buffer.Style
	Border  buffer.Style
}

// DefaultMiniMapStyle returns defaults.
func DefaultMiniMapStyle() MiniMapStyle {
	return MiniMapStyle{
		Full:    buffer.Style{Fg: buffer.RGB(51, 65, 85)},
		Visible: buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Border:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const miniMapHeight = 10

// MiniMap renders a compact content overview.
type MiniMap struct {
	BaseComponent
	mu sync.Mutex

	totalLines int
	viewStart  int
	viewHeight int
	style      MiniMapStyle
	// cached
	visStartRow int
	visEndRow   int
}

// NewMiniMap creates a MiniMap.
func NewMiniMap() *MiniMap {
	mm := &MiniMap{totalLines: 100, viewHeight: 20, style: DefaultMiniMapStyle()}
	mm.SetID(GenerateID("minimap"))
	mm.recomputeLocked()
	return mm
}

// SetContent sets the total content size, visible start, and visible height.
func (mm *MiniMap) SetContent(total, start, height int) *MiniMap {
	mm.mu.Lock()
	if total < 1 {
		total = 1
	}
	if height < 1 {
		height = 1
	}
	if start < 0 {
		start = 0
	}
	if start > total-height {
		start = total - height
	}
	if start < 0 {
		start = 0
	}
	mm.totalLines = total
	mm.viewStart = start
	mm.viewHeight = height
	mm.recomputeLocked()
	mm.mu.Unlock()
	return mm
}

func (mm *MiniMap) recomputeLocked() {
	mm.visStartRow = mm.viewStart * miniMapHeight / mm.totalLines
	mm.visEndRow = (mm.viewStart + mm.viewHeight) * miniMapHeight / mm.totalLines
	if mm.visEndRow > miniMapHeight {
		mm.visEndRow = miniMapHeight
	}
	if mm.visStartRow >= mm.visEndRow && mm.visStartRow > 0 {
		mm.visStartRow = mm.visEndRow - 1
	}
	if mm.visStartRow < 0 {
		mm.visStartRow = 0
	}
}

// ViewStart returns the current viewport start line.
func (mm *MiniMap) ViewStart() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return mm.viewStart
}

// SetStyle sets custom style.
func (mm *MiniMap) SetStyle(s MiniMapStyle) *MiniMap {
	mm.mu.Lock()
	mm.style = s
	mm.mu.Unlock()
	return mm
}

// Measure returns preferred size.
func (mm *MiniMap) Measure(cs Constraints) Size {
	w := 3
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: miniMapHeight}
}

// Paint renders the mini map.
func (mm *MiniMap) Paint(buf *buffer.Buffer) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	b := mm.Bounds()
	x, y := b.X, b.Y

	fullStyle := mm.style.Full
	visStyle := mm.style.Visible

	for row := 0; row < miniMapHeight; row++ {
		yy := y + row
		if yy >= buf.Height {
			break
		}

		var r rune
		var st buffer.Style
		if row >= mm.visStartRow && row < mm.visEndRow {
			r = '█'
			st = visStyle
		} else {
			r = '░'
			st = fullStyle
		}

		if x < buf.Width {
			buf.SetCell(x, yy, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (mm *MiniMap) Children() []Component { return nil }
