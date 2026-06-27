package component

import (
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// TooltipPlacement determines where the tooltip appears relative to the anchor.
type TooltipPlacement int

const (
	// TooltipTop places the tooltip above the anchor.
	TooltipTop TooltipPlacement = iota
	// TooltipBottom places the tooltip below the anchor.
	TooltipBottom
	// TooltipRight places the tooltip to the right of the anchor.
	TooltipRight
	// TooltipLeft places the tooltip to the left of the anchor.
	TooltipLeft
)

// Tooltip is a small popup that displays helpful text when the user hovers
// over a region. It is typically rendered as an overlay with a border,
// positioned near the anchor point.
//
// The tooltip auto-hides after a configurable duration or when the mouse
// leaves the anchor region.
type Tooltip struct {
	BaseComponent

	mu sync.RWMutex

	text      string
	lines     []string // pre-split text lines
	placement TooltipPlacement

	// anchorX/anchorY define the position the tooltip is attached to.
	anchorX int
	anchorY int

	// showDelay is how long the mouse must hover before the tooltip appears.
	showDelay time.Duration

	// autoHide is how long the tooltip stays visible (0 = until dismissed).
	autoHide time.Duration

	// showTimer tracks elapsed hover time.
	showTimer time.Duration
	visible   bool

	// maxWidth limits the tooltip width (0 = no limit).
	maxWidth int

	// style for the tooltip text.
	textStyle  buffer.Style
	borderStyle buffer.Style
	showBorder  bool

	// smart positioning: if true, auto-flip when near screen edge.
	smartPosition bool
}

// NewTooltip creates a tooltip with the given text.
func NewTooltip(text string) *Tooltip {
	t := &Tooltip{
		text:          text,
		placement:     TooltipTop,
		showDelay:     500 * time.Millisecond,
		autoHide:      0,
		maxWidth:      0,
		textStyle:     buffer.Style{Fg: buffer.RGB(220, 220, 220)},
		borderStyle:   buffer.Style{Fg: buffer.RGB(120, 120, 120)},
		showBorder:    true,
		smartPosition: true,
	}
	t.splitText()
	t.SetID(GenerateID("tooltip"))
	return t
}

// SetText updates the tooltip text and re-splits lines.
func (t *Tooltip) SetText(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.text = text
	t.splitText()
}

// Text returns the current tooltip text.
func (t *Tooltip) Text() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.text
}

// SetPlacement sets where the tooltip appears relative to the anchor.
func (t *Tooltip) SetPlacement(p TooltipPlacement) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.placement = p
}

// Placement returns the current placement.
func (t *Tooltip) Placement() TooltipPlacement {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.placement
}

// SetAnchor sets the (x, y) position the tooltip is attached to.
func (t *Tooltip) SetAnchor(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.anchorX = x
	t.anchorY = y
}

// Anchor returns the current anchor position.
func (t *Tooltip) Anchor() (int, int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.anchorX, t.anchorY
}

// SetShowDelay sets the hover duration before the tooltip appears.
func (t *Tooltip) SetShowDelay(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showDelay = d
}

// SetAutoHide sets how long the tooltip stays visible (0 = persistent).
func (t *Tooltip) SetAutoHide(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.autoHide = d
}

// SetTextStyle sets the style for tooltip text.
func (t *Tooltip) SetTextStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.textStyle = s
}

// SetBorderStyle sets the style for the tooltip border.
func (t *Tooltip) SetBorderStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.borderStyle = s
}

// SetShowBorder toggles whether a border is drawn around the tooltip.
func (t *Tooltip) SetShowBorder(show bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showBorder = show
}

// SetMaxWidth sets the maximum line width (0 = no limit).
func (t *Tooltip) SetMaxWidth(w int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxWidth = w
	t.splitText()
}

// SetSmartPosition toggles automatic edge-aware positioning.
func (t *Tooltip) SetSmartPosition(smart bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.smartPosition = smart
}

// IsVisible reports whether the tooltip is currently shown.
func (t *Tooltip) IsVisible() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.visible
}

// Show immediately displays the tooltip.
func (t *Tooltip) Show() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.visible = true
	t.showTimer = 0
}

// Hide immediately hides the tooltip.
func (t *Tooltip) Hide() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.visible = false
	t.showTimer = 0
}

// Tick advances the tooltip's internal timer. Call on each animation frame
// or when mouse position changes.
// If hovering is true, the show timer counts up; when it exceeds showDelay,
// the tooltip becomes visible.
// If autoHide > 0 and the tooltip is visible, the timer counts down;
// when it reaches 0, the tooltip hides.
// Returns true if the visibility state changed.
func (t *Tooltip) Tick(elapsed time.Duration, hovering bool) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	changed := false
	if hovering {
		if !t.visible {
			t.showTimer += elapsed
			if t.showTimer >= t.showDelay {
				t.visible = true
				t.showTimer = 0
				changed = true
			}
		}
	} else {
		if t.visible {
			t.visible = false
			t.showTimer = 0
			changed = true
		} else {
			t.showTimer = 0
		}
	}
	return changed
}

// Lines returns the pre-split text lines.
func (t *Tooltip) Lines() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]string, len(t.lines))
	copy(result, t.lines)
	return result
}

// splitText splits the tooltip text into lines respecting maxWidth.
func (t *Tooltip) splitText() {
	maxW := t.maxWidth
	if maxW <= 0 {
		t.lines = strings.Split(t.text, "\n")
		return
	}
	// Word-wrap at maxW.
	var result []string
	for _, line := range strings.Split(t.text, "\n") {
		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			if len(current)+1+len(w) > maxW {
				result = append(result, current)
				current = w
			} else {
				current += " " + w
			}
		}
		result = append(result, current)
	}
	t.lines = result
}

// Measure returns the desired tooltip size based on text content.
func (t *Tooltip) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()

	maxLineW := 0
	for _, line := range t.lines {
		w := buffer.StringWidth(line)
		if w > maxLineW {
			maxLineW = w
		}
	}
	h := len(t.lines)
	if t.showBorder {
		maxLineW += 2 // 1 padding each side
		h += 2        // top + bottom border
	}
	return Size{W: maxLineW, H: h}
}

// Paint renders the tooltip into the buffer.
func (t *Tooltip) Paint(buf *buffer.Buffer) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.lines) == 0 {
		return
	}

	b := t.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	if t.showBorder {
		t.paintWithBorder(buf, b)
	} else {
		t.paintPlainText(buf, b)
	}
}

// paintWithBorder draws the tooltip with a Unicode box border.
func (t *Tooltip) paintWithBorder(buf *buffer.Buffer, b Rect) {
	w := b.W
	h := b.H

	// Top border.
	if h >= 1 {
		buf.SetCell(b.X, b.Y, buffer.Cell{Rune: '┌', Width: 1, Fg: t.borderStyle.Fg})
		buf.SetCell(b.X+w-1, b.Y, buffer.Cell{Rune: '┐', Width: 1, Fg: t.borderStyle.Fg})
		for x := b.X + 1; x < b.X+w-1; x++ {
			buf.SetCell(x, b.Y, buffer.Cell{Rune: '─', Width: 1, Fg: t.borderStyle.Fg})
		}
	}

	// Bottom border.
	if h >= 2 {
		bottomY := b.Y + h - 1
		buf.SetCell(b.X, bottomY, buffer.Cell{Rune: '└', Width: 1, Fg: t.borderStyle.Fg})
		buf.SetCell(b.X+w-1, bottomY, buffer.Cell{Rune: '┘', Width: 1, Fg: t.borderStyle.Fg})
		for x := b.X + 1; x < b.X+w-1; x++ {
			buf.SetCell(x, bottomY, buffer.Cell{Rune: '─', Width: 1, Fg: t.borderStyle.Fg})
		}
	}

	// Text lines.
	textW := w - 2 // minus borders
	for i, line := range t.lines {
		y := b.Y + 1 + i
		if y >= b.Y+h-1 {
			break
		}
		x := b.X + 1
		for _, r := range line {
			if x >= b.X+w-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: t.textStyle.Fg, Bg: t.textStyle.Bg})
			x++
		}
		// Fill remaining with background.
		for ; x < b.X+w-1; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: t.textStyle.Bg})
		}
	}

	// Middle fill for empty lines.
	for i := len(t.lines); i < h-2; i++ {
		y := b.Y + 1 + i
		for x := b.X + 1; x < b.X+w-1; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: t.textStyle.Bg})
		}
	}

	_ = textW
}

// paintPlainText draws just the text without a border.
func (t *Tooltip) paintPlainText(buf *buffer.Buffer, b Rect) {
	for i, line := range t.lines {
		y := b.Y + i
		if y >= b.Y+b.H {
			break
		}
		x := b.X
		for _, r := range line {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: t.textStyle.Fg, Bg: t.textStyle.Bg})
			x++
		}
	}
}

// ComputePosition calculates the optimal (x, y) for the tooltip given
// the anchor point, tooltip size, and screen dimensions. When smartPosition
// is enabled, the tooltip flips to avoid going off-screen.
func (t *Tooltip) ComputePosition(screenW, screenH int) (int, int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	size := t.Measure(Unbounded())
	tw := size.W
	th := size.H

	x, y := t.anchorX, t.anchorY
	p := t.placement

	if t.smartPosition {
		p = t.smartFlip(p, t.anchorX, t.anchorY, tw, th, screenW, screenH)
	}

	switch p {
	case TooltipTop:
		x = t.anchorX - tw/2
		y = t.anchorY - th - 1
	case TooltipBottom:
		x = t.anchorX - tw/2
		y = t.anchorY + 1
	case TooltipRight:
		x = t.anchorX + 1
		y = t.anchorY - th/2
	case TooltipLeft:
		x = t.anchorX - tw - 1
		y = t.anchorY - th/2
	}

	// Clamp to screen.
	if x < 0 {
		x = 0
	}
	if x+tw > screenW {
		x = screenW - tw
	}
	if y < 0 {
		y = 0
	}
	if y+th > screenH {
		y = screenH - th
	}

	return x, y
}

// smartFlip adjusts placement to avoid going off-screen.
func (t *Tooltip) smartFlip(p TooltipPlacement, ax, ay, tw, th, sw, sh int) TooltipPlacement {
	switch p {
	case TooltipTop:
		if ay-th-1 < 0 {
			return TooltipBottom
		}
	case TooltipBottom:
		if ay+1+th > sh {
			return TooltipTop
		}
	case TooltipRight:
		if ax+1+tw > sw {
			return TooltipLeft
		}
	case TooltipLeft:
		if ax-tw-1 < 0 {
			return TooltipRight
		}
	}
	return p
}
