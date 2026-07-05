package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// SplitDirection determines the orientation of a SplitPane.
type SplitDirection int

const (
	// SplitHorizontal splits left/right (vertical divider line).
	SplitHorizontal SplitDirection = iota
	// SplitVertical splits top/bottom (horizontal divider line).
	SplitVertical
)

// SplitPane is a layout component that divides its area into two panes
// separated by a draggable divider. The user can resize the panes by
// dragging the divider with the mouse or using keyboard shortcuts.
type SplitPane struct {
	BaseComponent

	mu sync.RWMutex

	direction SplitDirection
	first     Component
	second    Component

	// ratio is the proportion [0,1] of space allocated to the first pane.
	ratio float64

	// dividerPos is the pixel position of the divider (cached during layout).
	dividerPos int

	// min/max ratios constrain how far the divider can move.
	minRatio float64
	maxRatio float64

	// divider thickness in cells (typically 1).
	dividerW int

	// drag state
	dragging    bool
	dragStart   int
	dragStartRatio float64

	// style for the divider line.
	dividerStyle buffer.Style
	dividerChar  rune

	// showHandle adds a visible grip on the divider.
	showHandle bool
}

// NewSplitPane creates a SplitPane with two child components.
// Default: horizontal split, 50/50 ratio.
func NewSplitPane(first, second Component) *SplitPane {
	sp := &SplitPane{
		direction:    SplitHorizontal,
		first:        first,
		second:       second,
		ratio:        0.5,
		minRatio:     0.1,
		maxRatio:     0.9,
		dividerW:     1,
		dividerChar:  '│',
		dividerStyle: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		showHandle:   true,
	}
	sp.SetID(GenerateID("splitpane"))
	return sp
}

// SetDirection sets the split orientation.
func (sp *SplitPane) SetDirection(d SplitDirection) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.direction = d
	if d == SplitVertical {
		sp.dividerChar = '─'
	} else {
		sp.dividerChar = '│'
	}
}

// Direction returns the current split orientation.
func (sp *SplitPane) Direction() SplitDirection {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.direction
}

// SetRatio sets the first pane's proportion (clamped to [minRatio, maxRatio]).
func (sp *SplitPane) SetRatio(r float64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.ratio = clampRatio(r, sp.minRatio, sp.maxRatio)
}

// Ratio returns the current split ratio.
func (sp *SplitPane) Ratio() float64 {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.ratio
}

// SetMinRatio sets the minimum allowed ratio for the first pane.
func (sp *SplitPane) SetMinRatio(r float64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if r < 0 {
		r = 0
	}
	sp.minRatio = r
	sp.ratio = clampRatio(sp.ratio, sp.minRatio, sp.maxRatio)
}

// SetMaxRatio sets the maximum allowed ratio for the first pane.
func (sp *SplitPane) SetMaxRatio(r float64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if r > 1 {
		r = 1
	}
	sp.maxRatio = r
	sp.ratio = clampRatio(sp.ratio, sp.minRatio, sp.maxRatio)
}

// SetDividerStyle sets the visual style of the divider line.
func (sp *SplitPane) SetDividerStyle(s buffer.Style) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.dividerStyle = s
}

// SetShowHandle toggles the visible grip on the divider.
func (sp *SplitPane) SetShowHandle(show bool) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.showHandle = show
}

// First returns the first (left/top) pane component.
func (sp *SplitPane) First() Component {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.first
}

// Second returns the second (right/bottom) pane component.
func (sp *SplitPane) Second() Component {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.second
}

// SetFirst replaces the first pane component.
func (sp *SplitPane) SetFirst(c Component) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.first = c
}

// SetSecond replaces the second pane component.
func (sp *SplitPane) SetSecond(c Component) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.second = c
}

// Children returns the two pane components.
func (sp *SplitPane) Children() []Component {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return []Component{sp.first, sp.second}
}

// IsDragging reports whether the divider is currently being dragged.
func (sp *SplitPane) IsDragging() bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.dragging
}

// StartDrag begins a drag operation. pos is the current mouse position
// along the divider axis (x for horizontal, y for vertical).
func (sp *SplitPane) StartDrag(pos int) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.dragging = true
	sp.dragStart = pos
	sp.dragStartRatio = sp.ratio
}

// UpdateDrag updates the ratio based on mouse movement during a drag.
// pos is the current mouse position along the divider axis.
func (sp *SplitPane) UpdateDrag(pos int) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if !sp.dragging {
		return
	}
	delta := pos - sp.dragStart
	if sp.direction == SplitHorizontal {
		avail := sp.bounds.W - sp.dividerW
		if avail <= 0 {
			return
		}
		sp.ratio = clampRatio(sp.dragStartRatio+float64(delta)/float64(avail), sp.minRatio, sp.maxRatio)
	} else {
		avail := sp.bounds.H - sp.dividerW
		if avail <= 0 {
			return
		}
		sp.ratio = clampRatio(sp.dragStartRatio+float64(delta)/float64(avail), sp.minRatio, sp.maxRatio)
	}
}

// EndDrag finishes a drag operation.
func (sp *SplitPane) EndDrag() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.dragging = false
}

// HandleMouse processes a mouse event. Returns true if the event was consumed.
// button=1 means left click (start drag), button=0 means move, button=-1 means release.
func (sp *SplitPane) HandleMouse(x, y, button int) bool {
	sp.mu.Lock()
	divPos := sp.computeDividerPosLocked()
	vert := sp.direction == SplitVertical
	sp.mu.Unlock()

	if vert {
		// Vertical split: divider is horizontal, check y
		if button == 1 && y == divPos {
			sp.StartDrag(y)
			return true
		}
		if button == 0 && sp.IsDragging() {
			sp.UpdateDrag(y)
			return true
		}
		if button == -1 && sp.IsDragging() {
			sp.EndDrag()
			return true
		}
		// Click near divider
		if button == 1 && abs(y-divPos) <= 0 {
			sp.StartDrag(y)
			return true
		}
	} else {
		// Horizontal split: divider is vertical, check x
		if button == 1 && x == divPos {
			sp.StartDrag(x)
			return true
		}
		if button == 0 && sp.IsDragging() {
			sp.UpdateDrag(x)
			return true
		}
		if button == -1 && sp.IsDragging() {
			sp.EndDrag()
			return true
		}
	}
	return false
}

// HandleKey processes keyboard resize commands.
// Ctrl+Shift+Left/Right (horizontal) or Ctrl+Shift+Up/Down (vertical) resize.
// Returns true if the key was consumed.
func (sp *SplitPane) HandleKey(key uint16, mods uint8) bool {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	const (
		modCtrl  = 1
		modShift = 4
		keyLeft  = 11
		keyRight = 12
		keyUp    = 9
		keyDown  = 10
	)

	if mods&(modCtrl|modShift) != (modCtrl | modShift) {
		return false
	}

	step := 0.05
	switch {
	case key == keyLeft && sp.direction == SplitHorizontal:
		sp.ratio = clampRatio(sp.ratio-step, sp.minRatio, sp.maxRatio)
		return true
	case key == keyRight && sp.direction == SplitHorizontal:
		sp.ratio = clampRatio(sp.ratio+step, sp.minRatio, sp.maxRatio)
		return true
	case key == keyUp && sp.direction == SplitVertical:
		sp.ratio = clampRatio(sp.ratio-step, sp.minRatio, sp.maxRatio)
		return true
	case key == keyDown && sp.direction == SplitVertical:
		sp.ratio = clampRatio(sp.ratio+step, sp.minRatio, sp.maxRatio)
		return true
	}
	return false
}

// Measure returns the desired size, which is the sum of both children's
// desired sizes plus the divider width.
func (sp *SplitPane) Measure(cs Constraints) Size {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	if sp.direction == SplitHorizontal {
		w := cs.MaxWidth
		if w <= 0 {
			w = 80
		}
		h := cs.MaxHeight
		if h <= 0 {
			if sp.first != nil {
				h = sp.first.Measure(Unbounded()).H
			}
			if sp.second != nil {
				h2 := sp.second.Measure(Unbounded()).H
				if h2 > h {
					h = h2
				}
			}
			if h <= 0 {
				h = 24
			}
		}
		return Size{W: w, H: h}
	}

	h := cs.MaxHeight
	if h <= 0 {
		h = 24
	}
	w := cs.MaxWidth
	if w <= 0 {
		w = 80
	}
	return Size{W: w, H: h}
}

// Paint renders both panes and the divider into the buffer.
func (sp *SplitPane) Paint(buf *buffer.Buffer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	b := sp.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	if sp.direction == SplitHorizontal {
		sp.paintHorizontal(buf, b)
	} else {
		sp.paintVertical(buf, b)
	}
}

// paintHorizontal renders left/right panes with a vertical divider.
func (sp *SplitPane) paintHorizontal(buf *buffer.Buffer, b Rect) {
	avail := b.W - sp.dividerW
	if avail <= 0 {
		return
	}
	firstW := int(float64(avail) * sp.ratio)
	secondW := avail - firstW
	divPos := b.X + firstW // local variable, no mutation of shared state

	// Layout first pane.
	if sp.first != nil {
		sp.first.SetBounds(Rect{X: b.X, Y: b.Y, W: firstW, H: b.H})
		sp.first.Paint(buf)
	}

	// Draw divider.
	for y := b.Y; y < b.Y+b.H; y++ {
		buf.SetCell(divPos, y, buffer.Cell{
			Rune:  sp.dividerChar,
			Width: 1,
			Fg:    sp.dividerStyle.Fg,
			Bg:    sp.dividerStyle.Bg,
		})
	}

	// Draw handle grip in the middle.
	if sp.showHandle && b.H > 4 {
		midY := b.Y + b.H/2
		buf.SetCell(divPos, midY, buffer.Cell{
			Rune:  '◆',
			Width: 1,
			Fg:    buffer.RGB(200, 200, 200),
		})
	}

	// Layout second pane.
	if sp.second != nil {
		secondX := divPos + sp.dividerW
		sp.second.SetBounds(Rect{X: secondX, Y: b.Y, W: secondW, H: b.H})
		sp.second.Paint(buf)
	}
}

// DividerPos returns the current divider position. Computed from bounds and ratio.
func (sp *SplitPane) DividerPos() int {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.computeDividerPosLocked()
}

// computeDividerPosLocked calculates the divider position without locking.
// Caller must hold at least RLock.
func (sp *SplitPane) computeDividerPosLocked() int {
	b := sp.bounds
	if sp.direction == SplitHorizontal {
		avail := b.W - sp.dividerW
		if avail <= 0 {
			return 0
		}
		return b.X + int(float64(avail)*sp.ratio)
	}
	avail := b.H - sp.dividerW
	if avail <= 0 {
		return 0
	}
	return b.Y + int(float64(avail)*sp.ratio)
}

// paintVertical renders top/bottom panes with a horizontal divider.
func (sp *SplitPane) paintVertical(buf *buffer.Buffer, b Rect) {
	avail := b.H - sp.dividerW
	if avail <= 0 {
		return
	}
	firstH := int(float64(avail) * sp.ratio)
	secondH := avail - firstH
	divPos := b.Y + firstH // local variable

	// Layout first pane.
	if sp.first != nil {
		sp.first.SetBounds(Rect{X: b.X, Y: b.Y, W: b.W, H: firstH})
		sp.first.Paint(buf)
	}

	// Draw divider.
	for x := b.X; x < b.X+b.W; x++ {
		buf.SetCell(x, divPos, buffer.Cell{
			Rune:  sp.dividerChar,
			Width: 1,
			Fg:    sp.dividerStyle.Fg,
			Bg:    sp.dividerStyle.Bg,
		})
	}

	// Draw handle grip in the middle.
	if sp.showHandle && b.W > 4 {
		midX := b.X + b.W/2
		buf.SetCell(midX, divPos, buffer.Cell{
			Rune:  '◆',
			Width: 1,
			Fg:    buffer.RGB(200, 200, 200),
		})
	}

	// Layout second pane.
	if sp.second != nil {
		secondY := divPos + sp.dividerW
		sp.second.SetBounds(Rect{X: b.X, Y: secondY, W: b.W, H: secondH})
		sp.second.Paint(buf)
	}
}

// clampRatio ensures ratio stays within [min, max].
func clampRatio(r, min, max float64) float64 {
	if r < min {
		return min
	}
	if r > max {
		return max
	}
	return r
}

// abs returns the absolute value.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
