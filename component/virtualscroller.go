package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// VirtualScrollerStyle controls visual appearance of the virtual scroller.
type VirtualScrollerStyle struct {
	Normal     buffer.Style
	Selected   buffer.Style
	Border     buffer.Style
	Scrollbar  buffer.Style
	ScrollThumb buffer.Style
	Header     buffer.Style
}

// DefaultVirtualScrollerStyle returns a default style set.
func DefaultVirtualScrollerStyle() VirtualScrollerStyle {
	return VirtualScrollerStyle{
		Normal:      buffer.Style{Fg: buffer.RGB(248, 248, 242)},
		Selected:    buffer.Style{Fg: buffer.RGB(248, 248, 242), Bg: buffer.RGB(68, 71, 90), Flags: buffer.Bold},
		Border:      buffer.Style{Fg: buffer.RGB(98, 114, 164)},
		Scrollbar:   buffer.Style{Fg: buffer.RGB(98, 114, 164)},
		ScrollThumb: buffer.Style{Fg: buffer.RGB(139, 143, 164)},
		Header:      buffer.Style{Fg: buffer.RGB(189, 147, 249), Flags: buffer.Bold},
	}
}

// VirtualItem represents a single item in a virtual scroller.
type VirtualItem struct {
	ID    string
	Text  string
	Data  interface{}
}

// VirtualScroller is a high-performance list component that only renders
// visible items, making it suitable for datasets with 10,000+ entries.
// It implements the Component interface.
type VirtualScroller struct {
	BaseComponent
	mu sync.RWMutex

	items    []VirtualItem
	cursor   int
	scrollY  int
	style    VirtualScrollerStyle
	header   string
	showScroll bool
	showBorder bool

	// Callbacks
	OnSelect func(item VirtualItem)

	// Computed fields (updated on paint)
	visibleStart int
	visibleEnd   int
}

// NewVirtualScroller creates a new virtual scroller with default settings.
func NewVirtualScroller() *VirtualScroller {
	return &VirtualScroller{
		BaseComponent: BaseComponent{id: GenerateID("vscroll")},
		style:         DefaultVirtualScrollerStyle(),
		showScroll:    true,
		showBorder:    true,
	}
}

// --- Item Management ---

// SetItems replaces the entire item list.
func (vs *VirtualScroller) SetItems(items []VirtualItem) {
	vs.mu.Lock()
	vs.items = items
	vs.clampScrollLocked()
	vs.clampCursorLocked()
	vs.mu.Unlock()
}

// AddItem appends a single item to the list.
func (vs *VirtualScroller) AddItem(item VirtualItem) {
	vs.mu.Lock()
	vs.items = append(vs.items, item)
	vs.mu.Unlock()
}

// AddItems appends multiple items.
func (vs *VirtualScroller) AddItems(items []VirtualItem) {
	vs.mu.Lock()
	vs.items = append(vs.items, items...)
	vs.mu.Unlock()
}

// Items returns a copy of all items.
func (vs *VirtualScroller) Items() []VirtualItem {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	out := make([]VirtualItem, len(vs.items))
	copy(out, vs.items)
	return out
}

// ItemCount returns the total number of items.
func (vs *VirtualScroller) ItemCount() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return len(vs.items)
}

// ItemAt returns the item at the given index, or nil if out of range.
func (vs *VirtualScroller) ItemAt(idx int) *VirtualItem {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	if idx < 0 || idx >= len(vs.items) {
		return nil
	}
	return &vs.items[idx]
}

// Clear removes all items.
func (vs *VirtualScroller) Clear() {
	vs.mu.Lock()
	vs.items = nil
	vs.cursor = 0
	vs.scrollY = 0
	vs.mu.Unlock()
}

// --- Cursor Navigation ---

// Cursor returns the current cursor index.
func (vs *VirtualScroller) Cursor() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.cursor
}

// SetCursor sets the cursor position, clamping to valid range.
func (vs *VirtualScroller) SetCursor(idx int) {
	vs.mu.Lock()
	vs.clampCursorToLocked(idx)
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// MoveDown moves the cursor down by n positions.
func (vs *VirtualScroller) MoveDown(n int) {
	vs.mu.Lock()
	vs.clampCursorToLocked(vs.cursor + n)
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// MoveUp moves the cursor up by n positions.
func (vs *VirtualScroller) MoveUp(n int) {
	vs.mu.Lock()
	vs.clampCursorToLocked(vs.cursor - n)
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// MovePageDown moves down by one viewport height.
func (vs *VirtualScroller) MovePageDown() {
	vs.mu.Lock()
	h := vs.viewportHeightLocked()
	vs.clampCursorToLocked(vs.cursor + h)
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// MovePageUp moves up by one viewport height.
func (vs *VirtualScroller) MovePageUp() {
	vs.mu.Lock()
	h := vs.viewportHeightLocked()
	vs.clampCursorToLocked(vs.cursor - h)
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// MoveToStart moves cursor to the first item.
func (vs *VirtualScroller) MoveToStart() {
	vs.mu.Lock()
	vs.cursor = 0
	vs.scrollY = 0
	vs.mu.Unlock()
}

// MoveToEnd moves cursor to the last item.
func (vs *VirtualScroller) MoveToEnd() {
	vs.mu.Lock()
	if len(vs.items) > 0 {
		vs.cursor = len(vs.items) - 1
	} else {
		vs.cursor = 0
	}
	vs.ensureCursorVisibleLocked()
	vs.mu.Unlock()
}

// CurrentItem returns the item under the cursor, or nil if empty.
func (vs *VirtualScroller) CurrentItem() *VirtualItem {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	if len(vs.items) == 0 || vs.cursor < 0 || vs.cursor >= len(vs.items) {
		return nil
	}
	return &vs.items[vs.cursor]
}

// --- Scrolling ---

// ScrollY returns the current scroll offset.
func (vs *VirtualScroller) ScrollY() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.scrollY
}

// ScrollTo sets the scroll offset, clamping to valid range.
func (vs *VirtualScroller) ScrollTo(y int) {
	vs.mu.Lock()
	vs.scrollY = y
	vs.clampScrollLocked()
	vs.mu.Unlock()
}

// VisibleRange returns the [start, end) indices of visible items.
func (vs *VirtualScroller) VisibleRange() (int, int) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.visibleStart, vs.visibleEnd
}

// VisibleItems returns only the items currently in the viewport.
func (vs *VirtualScroller) VisibleItems() []VirtualItem {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	if len(vs.items) == 0 {
		return nil
	}
	start, end := vs.visibleStart, vs.visibleEnd
	if start < 0 {
		start = 0
	}
	if end > len(vs.items) {
		end = len(vs.items)
	}
	if start >= end {
		return nil
	}
	out := make([]VirtualItem, end-start)
	copy(out, vs.items[start:end])
	return out
}

// --- Filtering ---

// Filter returns indices of items matching the given substring (case-insensitive).
func (vs *VirtualScroller) Filter(query string) []int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	q := strings.ToLower(query)
	var result []int
	for i, item := range vs.items {
		if q == "" || strings.Contains(strings.ToLower(item.Text), q) {
			result = append(result, i)
		}
	}
	return result
}

// --- Configuration ---

// SetStyle sets the visual style.
func (vs *VirtualScroller) SetStyle(s VirtualScrollerStyle) {
	vs.mu.Lock()
	vs.style = s
	vs.mu.Unlock()
}

// Style returns the current style.
func (vs *VirtualScroller) Style() VirtualScrollerStyle {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.style
}

// SetHeader sets the header text (empty = no header).
func (vs *VirtualScroller) SetHeader(h string) {
	vs.mu.Lock()
	vs.header = h
	vs.mu.Unlock()
}

// Header returns the header text.
func (vs *VirtualScroller) Header() string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.header
}

// SetShowScrollbar controls scrollbar visibility.
func (vs *VirtualScroller) SetShowScrollbar(show bool) {
	vs.mu.Lock()
	vs.showScroll = show
	vs.mu.Unlock()
}

// SetShowBorder controls border visibility.
func (vs *VirtualScroller) SetShowBorder(show bool) {
	vs.mu.Lock()
	vs.showBorder = show
	vs.mu.Unlock()
}

// --- Component Interface ---

// Measure returns the desired size (fills available space).
func (vs *VirtualScroller) Measure(cs Constraints) Size {
	w, h := 40, 10
	if cs.HasWidth() {
		w = cs.MaxWidth
	}
	if cs.HasHeight() {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders only the visible items plus border and scrollbar.
func (vs *VirtualScroller) Paint(buf *buffer.Buffer) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	b := vs.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	// Compute viewport
	innerX, innerY, innerW, innerH := b.X, b.Y, b.W, b.H

	if vs.showBorder {
		drawVSBorder(buf, b, vs.style)
		innerX = b.X + 1
		innerY = b.Y + 1
		innerW = b.W - 2
		innerH = b.H - 2
	}

	// Header
	yOff := 0
	if vs.header != "" {
		for j, r := range vs.header {
			if innerX+j < b.X+b.W {
				buf.SetCell(innerX+j, innerY, buffer.NewCell(r, vs.style.Header))
			}
		}
		yOff = 1
		// Header separator
		if innerY+1 < b.Y+b.H {
			for x := innerX; x < innerX+innerW && x < b.X+b.W; x++ {
				buf.SetCell(x, innerY+1, buffer.NewCell('─', vs.style.Border))
			}
		}
	}

	availH := innerH - yOff
	if availH <= 0 {
		vs.visibleStart, vs.visibleEnd = 0, 0
		return
	}

	// Clamp scroll
	vs.clampScrollLocked()

	// Compute visible range
	start := vs.scrollY
	if start >= len(vs.items) {
		start = 0
		if len(vs.items) > 0 {
			start = len(vs.items) - 1
		}
	}
	end := start + availH
	if end > len(vs.items) {
		end = len(vs.items)
	}

	vs.visibleStart = start
	vs.visibleEnd = end

	// Render visible items
	for i := start; i < end; i++ {
		row := innerY + yOff + (i - start)
		if row >= b.Y+b.H {
			break
		}
		style := vs.style.Normal
		if i == vs.cursor {
			style = vs.style.Selected
		}

		// Item text
		text := vs.items[i].Text
		for j, r := range text {
			if innerX+j >= innerX+innerW || innerX+j >= b.X+b.W {
				break
			}
			buf.SetCell(innerX+j, row, buffer.NewCell(r, style))
		}

		// Fill rest of line with background for selected
		if i == vs.cursor {
			for x := innerX + len([]rune(text)); x < innerX+innerW && x < b.X+b.W; x++ {
				buf.SetCell(x, row, buffer.NewCell(' ', style))
			}
		}
	}

	// Scrollbar
	if vs.showScroll && len(vs.items) > availH {
		scrollX := b.X + b.W - 1
		totalH := b.H
		if vs.showBorder {
			totalH = b.H - 2
		}
		if totalH > 0 {
			thumbH := max(1, availH*totalH/len(vs.items))
			thumbY := b.Y
			if vs.showBorder {
				thumbY = b.Y + 1
			}
			if len(vs.items) > availH {
				thumbY += vs.scrollY * totalH / len(vs.items)
			}

			sbStart := b.Y
			if vs.showBorder {
				sbStart = b.Y + 1
			}
			sbEnd := sbStart + totalH
			if sbEnd > b.Y+b.H {
				sbEnd = b.Y + b.H
			}

			for y := sbStart; y < sbEnd; y++ {
				if y >= thumbY && y < thumbY+thumbH && y < b.Y+b.H {
					buf.SetCell(scrollX, y, buffer.NewCell('┃', vs.style.ScrollThumb))
				} else if vs.showBorder {
					// Don't overwrite border
				} else {
					buf.SetCell(scrollX, y, buffer.NewCell('│', vs.style.Scrollbar))
				}
			}
		}
	}
}

// String returns a debug description.
func (vs *VirtualScroller) String() string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return fmt.Sprintf("VirtualScroller{items:%d cursor:%d scrollY:%d visible:[%d,%d)}",
		len(vs.items), vs.cursor, vs.scrollY, vs.visibleStart, vs.visibleEnd)
}

// --- Internal helpers (must be called under Lock) ---

func (vs *VirtualScroller) viewportHeightLocked() int {
	h := vs.bounds.H
	if vs.showBorder {
		h -= 2
	}
	if vs.header != "" {
		h -= 2 // header line + separator
	}
	if h < 0 {
		h = 0
	}
	return h
}

func (vs *VirtualScroller) clampScrollLocked() {
	maxScroll := vs.maxScrollLocked()
	if vs.scrollY < 0 {
		vs.scrollY = 0
	}
	if vs.scrollY > maxScroll {
		vs.scrollY = maxScroll
	}
}

func (vs *VirtualScroller) maxScrollLocked() int {
	vh := vs.viewportHeightLocked()
	if len(vs.items) <= vh {
		return 0
	}
	return len(vs.items) - vh
}

func (vs *VirtualScroller) clampCursorLocked() {
	vs.clampCursorToLocked(vs.cursor)
}

func (vs *VirtualScroller) clampCursorToLocked(idx int) {
	if idx < 0 {
		idx = 0
	}
	n := len(vs.items)
	if n == 0 {
		vs.cursor = 0
		return
	}
	if idx >= n {
		idx = n - 1
	}
	vs.cursor = idx
}

func (vs *VirtualScroller) ensureCursorVisibleLocked() {
	vh := vs.viewportHeightLocked()
	if vh <= 0 {
		return
	}
	if vs.cursor < vs.scrollY {
		vs.scrollY = vs.cursor
	}
	if vs.cursor >= vs.scrollY+vh {
		vs.scrollY = vs.cursor - vh + 1
	}
	vs.clampScrollLocked()
}

// drawVSBorder draws a unicode box border.
func drawVSBorder(buf *buffer.Buffer, r Rect, style VirtualScrollerStyle) {
	if r.W < 2 || r.H < 2 {
		return
	}
	right := r.X + r.W - 1
	bottom := r.Y + r.H - 1

	buf.SetCell(r.X, r.Y, buffer.NewCell('┌', style.Border))
	buf.SetCell(right, r.Y, buffer.NewCell('┐', style.Border))
	buf.SetCell(r.X, bottom, buffer.NewCell('└', style.Border))
	buf.SetCell(right, bottom, buffer.NewCell('┘', style.Border))

	for x := r.X + 1; x < right; x++ {
		buf.SetCell(x, r.Y, buffer.NewCell('─', style.Border))
		buf.SetCell(x, bottom, buffer.NewCell('─', style.Border))
	}
	for y := r.Y + 1; y < bottom; y++ {
		buf.SetCell(r.X, y, buffer.NewCell('│', style.Border))
		buf.SetCell(right, y, buffer.NewCell('│', style.Border))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
