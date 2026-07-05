package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Viewport wraps any component with bidirectional (vertical + horizontal)
// scrolling, keyboard navigation, and optional scrollbars.
//
// Unlike ScrollView (which is vertical-only and ChatApp-specific),
// Viewport is a general-purpose scrollable container that supports:
//   - Vertical and horizontal scrolling
//   - Keyboard navigation (arrows, PageUp/Down, Home/End, vim j/k/h/l/g/G)
//   - Mouse wheel scrolling
//   - Optional vertical/horizontal scrollbars
//   - Thread-safe via sync.RWMutex
type Viewport struct {
	BaseComponent

	mu sync.RWMutex

	content Component

	// Scroll offsets
	offsetX int // horizontal scroll offset (columns from left)
	offsetY int // vertical scroll offset (rows from top)

	// Computed during Measure
	contentW int
	contentH int

	// Scrollbar visibility
	showVBar bool // vertical scrollbar (auto-detects overflow)
	showHBar bool // horizontal scrollbar (auto-detects overflow)

	// Drag state
	draggingV bool
	draggingH bool

	// Style
	vBarStyle ScrollBarStyle
	hBarStyle ScrollBarStyle
}

// NewViewport creates a Viewport wrapping the given child component.
func NewViewport(child Component) *Viewport {
	return &Viewport{
		content:   child,
		vBarStyle: DefaultScrollBarStyle(),
		hBarStyle: DefaultScrollBarStyle(),
	}
}

// Content returns the wrapped child component.
func (v *Viewport) Content() Component {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.content
}

// SetContent replaces the child component.
func (v *Viewport) SetContent(c Component) {
	v.mu.Lock()
	v.content = c
	v.offsetX = 0
	v.offsetY = 0
	v.mu.Unlock()
}

// OffsetX returns the horizontal scroll offset.
func (v *Viewport) OffsetX() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.offsetX
}

// OffsetY returns the vertical scroll offset.
func (v *Viewport) OffsetY() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.offsetY
}

// ContentWidth returns the measured width of the child content.
func (v *Viewport) ContentWidth() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.contentW
}

// ContentHeight returns the measured height of the child content.
func (v *Viewport) ContentHeight() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.contentH
}

// MaxOffsetX returns the maximum horizontal scroll offset.
func (v *Viewport) MaxOffsetX() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	bw := v.bounds.W
	if v.showVBar {
		bw--
	}
	max := v.contentW - bw
	if max < 0 {
		return 0
	}
	return max
}

// MaxOffsetY returns the maximum vertical scroll offset.
func (v *Viewport) MaxOffsetY() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	bh := v.bounds.H
	if v.showHBar {
		bh--
	}
	max := v.contentH - bh
	if max < 0 {
		return 0
	}
	return max
}

// ScrollUp scrolls up (decreases offsetY) by n rows.
func (v *Viewport) ScrollUp(n int) {
	v.mu.Lock()
	v.offsetY -= n
	if v.offsetY < 0 {
		v.offsetY = 0
	}
	v.mu.Unlock()
}

// ScrollDown scrolls down (increases offsetY) by n rows.
func (v *Viewport) ScrollDown(n int) {
	v.mu.Lock()
	max := v.maxOffsetYLocked()
	v.offsetY += n
	if v.offsetY > max {
		v.offsetY = max
	}
	v.mu.Unlock()
}

// ScrollLeft scrolls left (decreases offsetX) by n columns.
func (v *Viewport) ScrollLeft(n int) {
	v.mu.Lock()
	v.offsetX -= n
	if v.offsetX < 0 {
		v.offsetX = 0
	}
	v.mu.Unlock()
}

// ScrollRight scrolls right (increases offsetX) by n columns.
func (v *Viewport) ScrollRight(n int) {
	v.mu.Lock()
	max := v.maxOffsetXLocked()
	v.offsetX += n
	if v.offsetX > max {
		v.offsetX = max
	}
	v.mu.Unlock()
}

// ScrollToX sets the horizontal scroll offset, clamped to valid range.
func (v *Viewport) ScrollToX(x int) {
	v.mu.Lock()
	if x < 0 {
		x = 0
	}
	max := v.maxOffsetXLocked()
	if x > max {
		x = max
	}
	v.offsetX = x
	v.mu.Unlock()
}

// ScrollToY sets the vertical scroll offset, clamped to valid range.
func (v *Viewport) ScrollToY(y int) {
	v.mu.Lock()
	if y < 0 {
		y = 0
	}
	max := v.maxOffsetYLocked()
	if y > max {
		y = max
	}
	v.offsetY = y
	v.mu.Unlock()
}

// ScrollToTop scrolls to the top.
func (v *Viewport) ScrollToTop() {
	v.mu.Lock()
	v.offsetY = 0
	v.mu.Unlock()
}

// ScrollToBottom scrolls to the bottom.
func (v *Viewport) ScrollToBottom() {
	v.mu.Lock()
	v.offsetY = v.maxOffsetYLocked()
	v.mu.Unlock()
}

// ScrollToLeft scrolls to the leftmost position.
func (v *Viewport) ScrollToLeft() {
	v.mu.Lock()
	v.offsetX = 0
	v.mu.Unlock()
}

// ScrollToRight scrolls to the rightmost position.
func (v *Viewport) ScrollToRight() {
	v.mu.Lock()
	v.offsetX = v.maxOffsetXLocked()
	v.mu.Unlock()
}

func (v *Viewport) maxOffsetXLocked() int {
	bw := v.bounds.W
	if v.showVBar {
		bw--
	}
	max := v.contentW - bw
	if max < 0 {
		return 0
	}
	return max
}

func (v *Viewport) maxOffsetYLocked() int {
	bh := v.bounds.H
	if v.showHBar {
		bh--
	}
	max := v.contentH - bh
	if max < 0 {
		return 0
	}
	return max
}

// vBarWidth returns 1 if vertical scrollbar takes space, 0 otherwise.
func (v *Viewport) vBarWidth() int {
	if v.showVBar {
		return 1
	}
	return 0
}

// hBarHeight returns 1 if horizontal scrollbar takes space, 0 otherwise.
func (v *Viewport) hBarHeight() int {
	if v.showHBar {
		return 1
	}
	return 0
}

// Measure measures the viewport and its content.
func (v *Viewport) Measure(cs Constraints) Size {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.content == nil {
		return Size{W: cs.MaxWidth, H: cs.MaxHeight}
	}

	// First pass: measure child without scrollbar constraint
	childSize := v.content.Measure(Unbounded())
	v.contentW = childSize.W
	v.contentH = childSize.H

	// Determine if scrollbars are needed
	v.showVBar = v.contentH > v.bounds.H && v.bounds.H > 0
	v.showHBar = v.contentW > v.bounds.W && v.bounds.W > 0

	// Return the bounds size (viewport is a fixed-size container)
	w := cs.MaxWidth
	h := cs.MaxHeight
	if w == 0 {
		w = v.bounds.W
	}
	if h == 0 {
		h = v.bounds.H
	}
	return Size{W: w, H: h}
}

// SetBounds sets the viewport's bounds.
func (v *Viewport) SetBounds(r Rect) {
	v.mu.Lock()
	v.bounds = r
	v.mu.Unlock()

	v.Measure(Unbounded())

	v.mu.Lock()
	if v.content != nil {
		// Child gets bounds minus scrollbar space
		cw := r.W - v.vBarWidth()
		ch := r.H - v.hBarHeight()
		if cw < 1 {
			cw = 1
		}
		if ch < 1 {
			ch = 1
		}
		v.content.SetBounds(Rect{X: 0, Y: 0, W: cw, H: ch})
	}
	// Clamp offsets
	if v.offsetY > v.maxOffsetYLocked() {
		v.offsetY = v.maxOffsetYLocked()
	}
	if v.offsetX > v.maxOffsetXLocked() {
		v.offsetX = v.maxOffsetXLocked()
	}
	v.mu.Unlock()
}

// Children returns the child components.
func (v *Viewport) Children() []Component {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.content == nil {
		return nil
	}
	return []Component{v.content}
}

// Paint renders the visible portion of the child content plus scrollbars.
func (v *Viewport) Paint(buf *buffer.Buffer) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.content == nil || v.bounds.W <= 0 || v.bounds.H <= 0 {
		return
	}

	// Create a sub-buffer for the content area, offset by scroll position
	contentW := v.bounds.W - v.vBarWidth()
	contentH := v.bounds.H - v.hBarHeight()
	if contentW < 1 || contentH < 1 {
		return
	}

	// Paint the child into a temporary buffer
	childBuf := buffer.NewBuffer(v.contentW, v.contentH)
	childBuf.Fill(buffer.BlankCell)
	v.content.Paint(childBuf)

	// Blit the visible portion to the main buffer
	for row := 0; row < contentH; row++ {
		srcY := v.offsetY + row
		if srcY >= v.contentH {
			break
		}
		for col := 0; col < contentW; col++ {
			srcX := v.offsetX + col
			if srcX >= v.contentW {
				break
			}
			cell := childBuf.GetCell(srcX, srcY)
			if cell.Width == 0 {
				continue // skip padding cells
			}
			buf.SetCell(v.bounds.X+col, v.bounds.Y+row, cell)
		}
	}

	// Draw vertical scrollbar
	if v.showVBar {
		v.drawVScrollBar(buf)
	}

	// Draw horizontal scrollbar
	if v.showHBar {
		v.drawHScrollBar(buf)
	}
}

// drawVScrollBar draws the vertical scrollbar on the right edge.
func (v *Viewport) drawVScrollBar(buf *buffer.Buffer) {
	barX := v.bounds.X + v.bounds.W - 1
	barH := v.bounds.H - v.hBarHeight()
	if barH <= 0 {
		return
	}

	// Draw track
	for y := 0; y < barH; y++ {
		buf.SetCell(barX, v.bounds.Y+y, buffer.NewCell(v.vBarStyle.TrackChar, v.vBarStyle.Style))
	}

	// Draw thumb
	maxOff := v.maxOffsetYLocked()
	if maxOff == 0 {
		// Content fits, thumb = full bar
		for y := 0; y < barH; y++ {
			buf.SetCell(barX, v.bounds.Y+y, buffer.NewCell(v.vBarStyle.ThumbChar, v.vBarStyle.Style))
		}
		return
	}

	thumbH := max0(barH * barH / v.contentH)
	if thumbH < 1 {
		thumbH = 1
	}
	thumbY := v.bounds.Y + (v.offsetY*barH)/v.contentH
	if thumbY+thumbH > v.bounds.Y+barH {
		thumbY = v.bounds.Y + barH - thumbH
	}
	for y := 0; y < thumbH; y++ {
		buf.SetCell(barX, thumbY+y, buffer.NewCell(v.vBarStyle.ThumbChar, v.vBarStyle.Style))
	}
}

// drawHScrollBar draws the horizontal scrollbar on the bottom edge.
func (v *Viewport) drawHScrollBar(buf *buffer.Buffer) {
	barY := v.bounds.Y + v.bounds.H - 1
	barW := v.bounds.W - v.vBarWidth()
	if barW <= 0 {
		return
	}

	// Draw track
	for x := 0; x < barW; x++ {
		buf.SetCell(v.bounds.X+x, barY, buffer.NewCell(v.hBarStyle.TrackChar, v.hBarStyle.Style))
	}

	// Draw thumb
	maxOff := v.maxOffsetXLocked()
	if maxOff == 0 {
		for x := 0; x < barW; x++ {
			buf.SetCell(v.bounds.X+x, barY, buffer.NewCell(v.hBarStyle.ThumbChar, v.hBarStyle.Style))
		}
		return
	}

	thumbW := max0(barW * barW / v.contentW)
	if thumbW < 1 {
		thumbW = 1
	}
	thumbX := v.bounds.X + (v.offsetX*barW)/v.contentW
	if thumbX+thumbW > v.bounds.X+barW {
		thumbX = v.bounds.X + barW - thumbW
	}
	for x := 0; x < thumbW; x++ {
		buf.SetCell(thumbX+x, barY, buffer.NewCell(v.hBarStyle.ThumbChar, v.hBarStyle.Style))
	}
}

// HandleKey processes keyboard input for scrolling.
// Returns true if the key was consumed.
func (v *Viewport) HandleKey(key *term.KeyEvent) bool {
	switch {
	// Vertical: arrows
	case key.Key == term.KeyUp:
		v.ScrollUp(1)
		return true
	case key.Key == term.KeyDown:
		v.ScrollDown(1)
		return true
	// Horizontal: arrows
	case key.Key == term.KeyLeft:
		v.ScrollLeft(1)
		return true
	case key.Key == term.KeyRight:
		v.ScrollRight(1)
		return true
	// Page navigation
	case key.Key == term.KeyPageUp:
		v.mu.RLock()
		h := v.bounds.H
		v.mu.RUnlock()
		v.ScrollUp(h)
		return true
	case key.Key == term.KeyPageDown:
		v.mu.RLock()
		h := v.bounds.H
		v.mu.RUnlock()
		v.ScrollDown(h)
		return true
	case key.Key == term.KeyHome:
		v.ScrollToTop()
		v.ScrollToLeft()
		return true
	case key.Key == term.KeyEnd:
		v.ScrollToBottom()
		return true
	// Vim keys: j/k for vertical, h/l for horizontal, g/G for top/bottom
	case key.Rune == 'j' && key.Modifiers == 0:
		v.ScrollDown(1)
		return true
	case key.Rune == 'k' && key.Modifiers == 0:
		v.ScrollUp(1)
		return true
	case key.Rune == 'h' && key.Modifiers == 0:
		v.ScrollLeft(1)
		return true
	case key.Rune == 'l' && key.Modifiers == 0:
		v.ScrollRight(1)
		return true
	case key.Rune == 'g' && key.Modifiers == 0:
		v.ScrollToTop()
		return true
	case key.Rune == 'G' && key.Modifiers == 0:
		v.ScrollToBottom()
		return true
	case key.Rune == '0' && key.Modifiers == 0:
		v.ScrollToLeft()
		return true
	case key.Rune == '$' && key.Modifiers == 0:
		v.ScrollToRight()
		return true
	}
	return false
}

// IsDraggingV returns whether the vertical scrollbar is being dragged.
func (v *Viewport) IsDraggingV() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.draggingV
}

// IsDraggingH returns whether the horizontal scrollbar is being dragged.
func (v *Viewport) IsDraggingH() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.draggingH
}

// VScrollbarColumn returns the x position of the vertical scrollbar, or -1 if not visible.
func (v *Viewport) VScrollbarColumn() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.showVBar {
		return -1
	}
	return v.bounds.X + v.bounds.W - 1
}

// HScrollbarRow returns the y position of the horizontal scrollbar, or -1 if not visible.
func (v *Viewport) HScrollbarRow() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.showHBar {
		return -1
	}
	return v.bounds.Y + v.bounds.H - 1
}

// SetVBarStyle customizes the vertical scrollbar appearance.
func (v *Viewport) SetVBarStyle(s ScrollBarStyle) {
	v.mu.Lock()
	v.vBarStyle = s
	v.mu.Unlock()
}

// SetHBarStyle customizes the horizontal scrollbar appearance.
func (v *Viewport) SetHBarStyle(s ScrollBarStyle) {
	v.mu.Lock()
	v.hBarStyle = s
	v.mu.Unlock()
}
