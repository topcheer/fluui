package component

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// ScrollBar style options.
type ScrollBarStyle struct {
	Visible   bool
	Style     buffer.Style
	ThumbChar rune // default '█'
	TrackChar rune // default '░'
}

// DefaultScrollBarStyle returns a reasonable default scrollbar style.
func DefaultScrollBarStyle() ScrollBarStyle {
	return ScrollBarStyle{
		Visible:   true,
		ThumbChar: '█',
		TrackChar: '░',
	}
}

// ScrollView wraps a child component with vertical scrolling capability.
// It manages a scroll offset and clips the child to the visible area.
//
// When the scrollbar is visible and content exceeds the viewport height,
// one column is reserved for the scrollbar on the right edge. Content is
// measured and painted with width = bounds.W - scrollbarWidth, so nothing
// is ever obscured.
type ScrollView struct {
	BaseComponent

	content   Component
	offset    int // vertical scroll offset (in rows from top)
	maxOffset int // computed max scroll offset
	scrollBar ScrollBarStyle

	// Content height tracking (updated during Measure)
	contentHeight int

	// Drag state
	dragging bool // true while user is dragging the scrollbar thumb
}

// scrollbarWidth returns 1 if scrollbar is visible, 0 otherwise.
func (sv *ScrollView) scrollbarWidth() int {
	if sv.scrollBar.Visible {
		return 1
	}
	return 0
}

// contentW returns the actual content width (bounds minus scrollbar).
func (sv *ScrollView) contentW(boundsW int) int {
	w := boundsW - sv.scrollbarWidth()
	if w < 1 {
		w = 1
	}
	return w
}

// NewScrollView creates a ScrollView wrapping the given child component.
func NewScrollView(child Component) *ScrollView {
	return &ScrollView{
		content:   child,
		scrollBar: DefaultScrollBarStyle(),
	}
}

// ScrollUp moves the viewport up by n rows.
func (sv *ScrollView) ScrollUp(n int) {
	sv.offset -= n
	if sv.offset < 0 {
		sv.offset = 0
	}
}

// ScrollDown moves the viewport down by n rows.
func (sv *ScrollView) ScrollDown(n int) {
	sv.offset += n
	if sv.offset > sv.maxOffset {
		sv.offset = sv.maxOffset
	}
}

// ScrollTo sets the scroll offset to the given value.
func (sv *ScrollView) ScrollTo(offset int) {
	if offset < 0 {
		offset = 0
	}
	if offset > sv.maxOffset {
		offset = sv.maxOffset
	}
	sv.offset = offset
}

// Offset returns the current scroll offset.
func (sv *ScrollView) Offset() int { return sv.offset }

// MaxOffset returns the maximum scroll offset.
func (sv *ScrollView) MaxOffset() int { return sv.maxOffset }

// SetBounds overrides BaseComponent.SetBounds to recalculate maxOffset
// when the viewport size changes.
func (sv *ScrollView) SetBounds(r Rect) {
	sv.BaseComponent.SetBounds(r)

	if sv.contentHeight > r.H {
		sv.maxOffset = sv.contentHeight - r.H
	} else {
		sv.maxOffset = 0
	}

	if sv.offset > sv.maxOffset {
		sv.offset = sv.maxOffset
	}
}

// Measure returns the ScrollView's desired size.
// Content is measured at the FULL width — scrollbar width is only
// reserved during Paint, never during Measure. This keeps Measure/Paint
// width consistent so blocks never get clipped.
func (sv *ScrollView) Measure(cs Constraints) Size {
	contentCS := Constraints{MaxWidth: cs.MaxWidth}
	contentSize := sv.content.Measure(contentCS)
	sv.contentHeight = contentSize.H

	viewportH := contentSize.H
	if cs.MaxHeight > 0 && viewportH > cs.MaxHeight {
		viewportH = cs.MaxHeight
	}
	if sv.contentHeight > viewportH {
		sv.maxOffset = sv.contentHeight - viewportH
	} else {
		sv.maxOffset = 0
	}

	w := contentSize.W
	h := contentSize.H
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint renders the visible portion of the child content into the buffer,
// then draws the scrollbar in the reserved rightmost column.
//
// Content is painted to a sub-buffer with width = bounds.W - scrollbarWidth,
// so content never overlaps with the scrollbar.
func (sv *ScrollView) Paint(buf *buffer.Buffer) {
	bounds := sv.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Clear the viewport area first.
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		for x := bounds.X; x < bounds.X+bounds.W; x++ {
			if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
				buf.Cells[y*buf.Width+x] = buffer.Cell{Rune: ' ', Width: 1}
			}
		}
	}

	// Determine if scrollbar should show.
	maxOff := 0
	if sv.contentHeight > bounds.H {
		maxOff = sv.contentHeight - bounds.H
	}
	showScrollbar := sv.scrollBar.Visible && maxOff > 0

	// Content width: reserve 1 column for scrollbar when visible.
	cw := bounds.W
	if showScrollbar {
		cw = sv.contentW(bounds.W)
	}

	offset := sv.offset
	if offset > maxOff {
		offset = maxOff
	}

	// Set child bounds to content width (excluding scrollbar column).
	sv.content.SetBounds(Rect{
		X: bounds.X,
		Y: bounds.Y - offset,
		W: cw,
		H: sv.contentHeight,
	})

	// Paint content into a sub-buffer of content width, then copy visible rows.
	sub := buffer.NewBuffer(cw, bounds.H)
	if vp, ok := sv.content.(VisiblePainter); ok {
		vp.PaintVisible(sub, offset, offset+bounds.H)
	} else {
		sv.content.Paint(sub)
	}

	for y := 0; y < bounds.H; y++ {
		for x := 0; x < cw; x++ {
			dstX := bounds.X + x
			dstY := bounds.Y + y
			if dstX >= 0 && dstX < buf.Width && dstY >= 0 && dstY < buf.Height {
				contentY := y + offset
				if contentY >= 0 && contentY < sv.contentHeight {
					buf.Cells[dstY*buf.Width+dstX] = sub.Cells[y*cw+x]
				}
			}
		}
	}

	// Draw scrollbar in the rightmost column (only when needed).
	if showScrollbar {
		barX := bounds.X + bounds.W - 1
		barH := bounds.H

		thumbRatio := float64(bounds.H) / float64(sv.contentHeight)
		if thumbRatio < 0.1 {
			thumbRatio = 0.1
		}
		thumbH := int(float64(barH) * thumbRatio)
		if thumbH < 1 {
			thumbH = 1
		}
		thumbStart := int(float64(barH) * float64(offset) / float64(sv.contentHeight))

		// Determine thumb highlight style when dragging.
		thumbStyle := sv.scrollBar.Style
		if sv.dragging {
			// Highlight: invert by using Fg as Bg, add Bold.
			thumbStyle = buffer.Style{
				Fg:    sv.scrollBar.Style.Fg,
				Bg:    sv.scrollBar.Style.Fg,
				Flags: sv.scrollBar.Style.Flags | buffer.Bold,
			}
		}

		for y := 0; y < barH; y++ {
			r := sv.scrollBar.TrackChar
			cellStyle := sv.scrollBar.Style
			if y >= thumbStart && y < thumbStart+thumbH {
				r = sv.scrollBar.ThumbChar
				cellStyle = thumbStyle
			}
			buf.SetCell(barX, bounds.Y+y, buffer.NewCell(r, cellStyle))
		}
	}
}

// --- Scrollbar interaction ---

// IsScrollbarVisible reports whether the scrollbar is currently shown.
func (sv *ScrollView) IsScrollbarVisible() bool {
	bounds := sv.Bounds()
	return sv.scrollBar.Visible && sv.contentHeight > bounds.H
}

// ScrollbarColumn returns the X coordinate of the scrollbar column.
// Returns -1 if the scrollbar is not visible.
func (sv *ScrollView) ScrollbarColumn() int {
	if !sv.IsScrollbarVisible() {
		return -1
	}
	bounds := sv.Bounds()
	return bounds.X + bounds.W - 1
}

// ScrollbarBounds returns the scrollbar region as (barStartY, barH, thumbStartY, thumbH).
// Returns zero values if the scrollbar is not visible.
func (sv *ScrollView) ScrollbarBounds() (barStartY, barH, thumbStartY, thumbH int) {
	bounds := sv.Bounds()
	if !sv.IsScrollbarVisible() {
		return 0, 0, 0, 0
	}
	barH = bounds.H
	if sv.contentHeight <= 0 {
		return bounds.Y, barH, 0, 1
	}

	thumbRatio := float64(bounds.H) / float64(sv.contentHeight)
	if thumbRatio < 0.1 {
		thumbRatio = 0.1
	}
	thumbH = int(float64(barH) * thumbRatio)
	if thumbH < 1 {
		thumbH = 1
	}
	thumbStartY = int(float64(barH) * float64(sv.offset) / float64(sv.contentHeight))

	return bounds.Y, barH, thumbStartY, thumbH
}

// IsDragging reports whether the user is currently dragging the scrollbar thumb.
func (sv *ScrollView) IsDragging() bool { return sv.dragging }

// HandleScrollbarDown processes a mouse-down event on the scrollbar.
// relY is the Y coordinate relative to the ScrollView's top (0-based).
// If the click lands on the thumb, dragging begins.
// If it lands on the track, the view jumps so the thumb centers on relY.
func (sv *ScrollView) HandleScrollbarDown(relY int) {
	if sv.contentHeight <= 0 {
		return
	}
	bounds := sv.Bounds()
	barH := bounds.H
	if barH <= 0 {
		return
	}

	_, _, thumbStart, thumbH := sv.ScrollbarBounds()

	// Click on thumb → start dragging.
	if relY >= thumbStart && relY < thumbStart+thumbH {
		sv.dragging = true
		return
	}

	// Click on track → jump so thumb centers on click position.
	if sv.maxOffset <= 0 {
		return
	}
	// Target: center the viewport on the clicked position.
	targetRel := relY - thumbH/2
	if targetRel < 0 {
		targetRel = 0
	}
	newOffset := int(float64(targetRel) * float64(sv.contentHeight) / float64(barH))
	sv.ScrollTo(newOffset)
}

// HandleScrollbarDrag processes a mouse-drag event while dragging the thumb.
// relY is the Y coordinate relative to the ScrollView's top (0-based).
// The thumb top follows the mouse position.
func (sv *ScrollView) HandleScrollbarDrag(relY int) {
	if !sv.dragging {
		return
	}
	sv.scrollThumbTo(relY)
}

// scrollThumbTo moves the thumb so its top aligns with relY.
func (sv *ScrollView) scrollThumbTo(relY int) {
	if sv.contentHeight <= 0 || sv.maxOffset <= 0 {
		return
	}
	bounds := sv.Bounds()
	barH := bounds.H
	if barH <= 0 {
		return
	}

	_, _, _, thumbH := sv.ScrollbarBounds()

	// Clamp relY to valid range.
	maxRelY := barH - thumbH
	if relY < 0 {
		relY = 0
	}
	if relY > maxRelY {
		relY = maxRelY
	}

	// Map relY (thumb top position) to scroll offset.
	newOffset := 0
	if maxRelY > 0 {
		newOffset = int(float64(relY) * float64(sv.maxOffset) / float64(maxRelY))
	}
	sv.ScrollTo(newOffset)
}

// HandleScrollbarUp ends a drag operation.
func (sv *ScrollView) HandleScrollbarUp() {
	sv.dragging = false
}

// Children returns the child component.
func (sv *ScrollView) Children() []Component {
	return []Component{sv.content}
}
