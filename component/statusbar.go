package component

import (
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// StatusItemAlignment controls where a StatusItem appears in the bar.
type StatusItemAlignment int

const (
	// StatusAlignLeft places the item in the left segment.
	StatusAlignLeft StatusItemAlignment = iota
	// StatusAlignCenter places the item in the center segment.
	StatusAlignCenter
	// StatusAlignRight places the item in the right segment.
	StatusAlignRight
)

// StatusItem is a single piece of information displayed in the StatusBar.
type StatusItem struct {
	ID    string
	Text  string
	Align StatusItemAlignment
	Style buffer.Style
}

// StatusBarStyle holds the visual styles for the StatusBar segments.
type StatusBarStyle struct {
	Background buffer.Style // background fill for the entire bar
	Left       buffer.Style // style for left-segment items
	Center     buffer.Style // style for center-segment items
	Right      buffer.Style // style for right-segment items
	Separator  buffer.Style // style for the separator character
}

// DefaultStatusBarStyle returns a Dracula-themed StatusBarStyle.
func DefaultStatusBarStyle() StatusBarStyle {
	bg := buffer.Style{
		Fg: buffer.RGB(0xf8, 0xf8, 0xf2),
		Bg: buffer.RGB(0x28, 0x2a, 0x36),
	}
	sep := buffer.Style{
		Fg: buffer.RGB(0x62, 0x72, 0xa4),
		Bg: buffer.RGB(0x28, 0x2a, 0x36),
	}
	return StatusBarStyle{
		Background: bg,
		Left:       bg,
		Center:     bg,
		Right:      bg,
		Separator:  sep,
	}
}

// StatusBar is a bottom-bar component with left/center/right segments.
// It implements the Component interface and is safe for concurrent use.
type StatusBar struct {
	BaseComponent
	mu          sync.RWMutex
	items       []StatusItem
	style       StatusBarStyle
	sep         string // separator between items within a segment
	height      int    // desired height (default 1)
	cachedLeft  string // pre-computed left segment text (avoids allocs in Paint)
	cachedCenter string
	cachedRight string
}

// NewStatusBar creates a StatusBar with default styling.
func NewStatusBar() *StatusBar {
	sb := &StatusBar{
		items:  make([]StatusItem, 0),
		style:  DefaultStatusBarStyle(),
		sep:    " │ ",
		height: 1,
	}
	sb.SetID(GenerateID("statusbar"))
	return sb
}

// --- Item management ---

// AddItem appends a StatusItem to the bar.
func (sb *StatusBar) AddItem(item StatusItem) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.items = append(sb.items, item)
	sb.recomputeTextsLocked()
}

// AddLeft is a convenience for adding a left-aligned item.
func (sb *StatusBar) AddLeft(id, text string) {
	sb.AddItem(StatusItem{ID: id, Text: text, Align: StatusAlignLeft})
}

// AddCenter is a convenience for adding a center-aligned item.
func (sb *StatusBar) AddCenter(id, text string) {
	sb.AddItem(StatusItem{ID: id, Text: text, Align: StatusAlignCenter})
}

// AddRight is a convenience for adding a right-aligned item.
func (sb *StatusBar) AddRight(id, text string) {
	sb.AddItem(StatusItem{ID: id, Text: text, Align: StatusAlignRight})
}

// RemoveItem removes the item with the given ID. Returns true if found.
func (sb *StatusBar) RemoveItem(id string) bool {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	for i, it := range sb.items {
		if it.ID == id {
			sb.items = append(sb.items[:i], sb.items[i+1:]...)
			sb.recomputeTextsLocked()
			return true
		}
	}
	return false
}

// SetItemText updates the text of the item with the given ID.
// If the item does not exist, it is added with AlignLeft.
func (sb *StatusBar) SetItemText(id, text string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	for i := range sb.items {
		if sb.items[i].ID == id {
			sb.items[i].Text = text
			// Only recompute the affected alignment instead of all three.
			switch sb.items[i].Align {
			case StatusAlignLeft:
				sb.cachedLeft = sb.buildTextLocked(StatusAlignLeft)
			case StatusAlignCenter:
				sb.cachedCenter = sb.buildTextLocked(StatusAlignCenter)
			case StatusAlignRight:
				sb.cachedRight = sb.buildTextLocked(StatusAlignRight)
			}
			return
		}
	}
	// Not found — add as left-aligned.
	sb.items = append(sb.items, StatusItem{ID: id, Text: text, Align: StatusAlignLeft})
	sb.recomputeTextsLocked()
}

// SetItemStyle updates the style of the item with the given ID.
func (sb *StatusBar) SetItemStyle(id string, style buffer.Style) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	for i := range sb.items {
		if sb.items[i].ID == id {
			sb.items[i].Style = style
		}
	}
}

// Items returns a copy of all items.
func (sb *StatusBar) Items() []StatusItem {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	out := make([]StatusItem, len(sb.items))
	copy(out, sb.items)
	return out
}

// ItemCount returns the number of items.
func (sb *StatusBar) ItemCount() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return len(sb.items)
}

// Clear removes all items.
func (sb *StatusBar) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.items = sb.items[:0]
	sb.recomputeTextsLocked()
}

// --- Configuration ---

// SetStyle sets the StatusBarStyle.
func (sb *StatusBar) SetStyle(s StatusBarStyle) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.style = s
}

// Style returns the current StatusBarStyle.
func (sb *StatusBar) Style() StatusBarStyle {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.style
}

// SetSeparator sets the separator string between items within a segment.
func (sb *StatusBar) SetSeparator(sep string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.sep = sep
	sb.recomputeTextsLocked()
}

// Separator returns the current separator string.
func (sb *StatusBar) Separator() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.sep
}

// SetHeight sets the desired height (default 1).
func (sb *StatusBar) SetHeight(h int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if h < 1 {
		h = 1
	}
	sb.height = h
}

// --- Helpers (must hold lock) ---

// recomputeTextsLocked rebuilds cached segment texts. Must hold Lock.
func (sb *StatusBar) recomputeTextsLocked() {
	sb.cachedLeft = sb.buildTextLocked(StatusAlignLeft)
	sb.cachedCenter = sb.buildTextLocked(StatusAlignCenter)
	sb.cachedRight = sb.buildTextLocked(StatusAlignRight)
}

// buildTextLocked joins texts for items matching the given alignment. Must hold Lock.
func (sb *StatusBar) buildTextLocked(align StatusItemAlignment) string {
	var parts []string
	for _, it := range sb.items {
		if it.Align == align {
			parts = append(parts, it.Text)
		}
	}
	return strings.Join(parts, sb.sep)
}

// leftTextLocked returns the cached left text. Must hold at least RLock.
func (sb *StatusBar) leftTextLocked() string {
	return sb.cachedLeft
}

// centerTextLocked returns the cached center text. Must hold at least RLock.
func (sb *StatusBar) centerTextLocked() string {
	return sb.cachedCenter
}

// rightTextLocked returns the cached right text. Must hold at least RLock.
func (sb *StatusBar) rightTextLocked() string {
	return sb.cachedRight
}

// LeftItems returns the concatenated text of all left-aligned items.
func (sb *StatusBar) LeftItems() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.leftTextLocked()
}

// CenterItems returns the concatenated text of all center-aligned items.
func (sb *StatusBar) CenterItems() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.centerTextLocked()
}

// RightItems returns the concatenated text of all right-aligned items.
func (sb *StatusBar) RightItems() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.rightTextLocked()
}

// String implements fmt.Stringer.
func (sb *StatusBar) String() string {
	return "StatusBar"
}

// --- Component interface ---

// Measure returns the desired size: full width, height rows.
func (sb *StatusBar) Measure(cs Constraints) Size {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	left := sb.leftTextLocked()
	center := sb.centerTextLocked()
	right := sb.rightTextLocked()

	w := buffer.StringWidth(left) + buffer.StringWidth(center) + buffer.StringWidth(right) + 4
	// Expand to fill available width
	if cs.MaxWidth > 0 && w < cs.MaxWidth {
		w = cs.MaxWidth
	}
	// Clamp to MaxWidth if content is wider
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w == 0 {
		w = 1
	}

	h := sb.height
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if h == 0 {
		h = 1
	}

	return Size{W: w, H: h}
}

// SetBounds sets the component's position and size.
func (sb *StatusBar) SetBounds(r Rect) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.BaseComponent.SetBounds(r)
}

// Bounds returns the component's current bounds.
func (sb *StatusBar) Bounds() Rect {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.BaseComponent.Bounds()
}

// Paint renders the StatusBar into the buffer.
func (sb *StatusBar) Paint(buf *buffer.Buffer) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	b := sb.BaseComponent.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	y := b.Y

	// Fill background
	bgCell := buffer.NewCell(' ', sb.style.Background)
	for x := b.X; x < b.X+b.W; x++ {
		for yy := y; yy < y+b.H; yy++ {
			buf.SetCell(x, yy, bgCell)
		}
	}

	left := sb.leftTextLocked()
	center := sb.centerTextLocked()
	right := sb.rightTextLocked()

	cw := buffer.StringWidth(center)
	rw := buffer.StringWidth(right)

	// Draw left segment starting at b.X + 1 (small left padding).
	if left != "" {
		buf.DrawTextClamped(b.X+1, y, left, sb.style.Left)
	}

	// Draw center segment — centered in the bar.
	if center != "" {
		centerX := b.X + (b.W-cw)/2
		buf.DrawTextClamped(centerX, y, center, sb.style.Center)
	}

	// Draw right segment — right-aligned with small right padding.
	if right != "" {
		rightX := b.X + b.W - rw - 1
		buf.DrawTextClamped(rightX, y, right, sb.style.Right)
	}
}

// Children returns nil — StatusBar has no child components.
func (sb *StatusBar) Children() []Component {
	return nil
}

// --- Convenience for AI Agent status ---

// SetModel updates a common "model" item in the left segment.
func (sb *StatusBar) SetModel(name string) {
	sb.SetItemText("model", name)
}

// SetTokenRate updates a "tokens/s" item in the right segment.
func (sb *StatusBar) SetTokenRate(perSec int) {
	sb.SetItemText("tokenrate", formatTokenRate(perSec))
}

// SetContextWindow updates a "context" item in the right segment.
func (sb *StatusBar) SetContextWindow(used, total int) {
	sb.SetItemText("context", formatContextWindow(used, total))
}

// SetClock updates a time item in the right segment.
func (sb *StatusBar) SetClock(t time.Time) {
	sb.SetItemText("clock", t.Format("15:04:05"))
}

// formatTokenRate formats token rate for display.
func formatTokenRate(perSec int) string {
	if perSec <= 0 {
		return "0 tok/s"
	}
	if perSec >= 1000 {
		whole := perSec / 1000
		frac := (perSec % 1000) / 100
		if frac == 0 {
			return itoa(whole) + "k tok/s"
		}
		return itoa(whole) + "." + itoa(frac) + "k tok/s"
	}
	return itoa(perSec) + " tok/s"
}

// formatContextWindow formats context window usage.
func formatContextWindow(used, total int) string {
	pct := 0
	if total > 0 {
		pct = used * 100 / total
	}
	return itoa(used) + "/" + itoa(total) + " (" + itoa(pct) + "%)"
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
