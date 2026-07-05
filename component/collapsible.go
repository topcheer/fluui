package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Collapsible wraps any child component with an expandable/collapsible header.
// When collapsed, only the header line is shown. When expanded, the child
// component is rendered below the header.
//
// This is essential for AI chat UIs where tool calls, thinking sections,
// code blocks, and detailed output can be collapsed to save space.
type Collapsible struct {
	BaseComponent
	mu sync.RWMutex

	title      string
	expanded   bool
	child      Component
	collapsed  bool // alias for !expanded, kept for API clarity

	// visual options
	showArrow  bool          // show ▸/▾ indicator
	indent     int           // left indent for header
	headerStyle buffer.Style // style for header text
	childStyle  buffer.Style // style applied to header when expanded

	// callbacks
	onToggle func(expanded bool)

	bounds Rect
}

// NewCollapsible creates a Collapsible with the given title and child.
// The child can be nil; set it later with SetChild.
func NewCollapsible(title string, child Component) *Collapsible {
	c := &Collapsible{
		title:      title,
		expanded:   true,
		child:      child,
		showArrow:  true,
		indent:     0,
		headerStyle: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedWhite),
			Flags: buffer.Bold,
		},
		childStyle: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedCyan),
			Flags: buffer.Bold,
		},
	}
	c.SetID(GenerateID("collapsible"))
	return c
}

// Expanded returns true if the collapsible is currently expanded.
func (c *Collapsible) Expanded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.expanded
}

// Collapsed returns true if the collapsible is currently collapsed.
func (c *Collapsible) Collapsed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.expanded
}

// SetExpanded sets the expansion state.
func (c *Collapsible) SetExpanded(expanded bool) {
	c.mu.Lock()
	c.expanded = expanded
	cb := c.onToggle
	c.mu.Unlock()

	if cb != nil {
		cb(expanded)
	}
}

// Toggle flips the expansion state and returns the new state.
func (c *Collapsible) Toggle() bool {
	c.mu.Lock()
	c.expanded = !c.expanded
	newState := c.expanded
	cb := c.onToggle
	c.mu.Unlock()

	if cb != nil {
		cb(newState)
	}
	return newState
}

// Expand sets the state to expanded.
func (c *Collapsible) Expand() {
	c.SetExpanded(true)
}

// Collapse sets the state to collapsed.
func (c *Collapsible) Collapse() {
	c.SetExpanded(false)
}

// Title returns the current title.
func (c *Collapsible) Title() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.title
}

// SetTitle sets the header title.
func (c *Collapsible) SetTitle(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.title = title
}

// Child returns the wrapped child component.
func (c *Collapsible) Child() Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.child
}

// SetChild sets the wrapped child component.
func (c *Collapsible) SetChild(child Component) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.child = child
}

// SetShowArrow toggles the ▸/▾ indicator display.
func (c *Collapsible) SetShowArrow(show bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showArrow = show
}

// ShowArrow returns whether the arrow indicator is shown.
func (c *Collapsible) ShowArrow() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.showArrow
}

// SetIndent sets the left indentation for the header.
func (c *Collapsible) SetIndent(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 0 {
		n = 0
	}
	c.indent = n
}

// Indent returns the current indentation level.
func (c *Collapsible) Indent() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.indent
}

// SetHeaderStyle sets the style for the header text.
func (c *Collapsible) SetHeaderStyle(s buffer.Style) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headerStyle = s
}

// SetExpandedHeaderStyle sets the style for the header when expanded.
func (c *Collapsible) SetExpandedHeaderStyle(s buffer.Style) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.childStyle = s
}

// OnToggle sets a callback fired whenever the expansion state changes.
func (c *Collapsible) OnToggle(fn func(expanded bool)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onToggle = fn
}

// HandleKey processes keyboard input for toggling.
// Tab/Enter/Space toggle expansion. Arrow keys are forwarded to child when expanded.
func (c *Collapsible) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}

	// Toggle keys: Enter, Space, Tab
	if k.Key == term.KeyEnter || k.Key == term.KeyTab {
		c.Toggle()
		return true
	}
	if k.Rune == ' ' || k.Rune == '\t' {
		c.Toggle()
		return true
	}

	// Forward to child when expanded
	c.mu.RLock()
	child := c.child
	expanded := c.expanded
	c.mu.RUnlock()

	if expanded && child != nil {
		if hk, ok := child.(interface{ HandleKey(*term.KeyEvent) bool }); ok {
			return hk.HandleKey(k)
		}
	}

	return false
}

// Measure returns the desired size.
// When collapsed, height is 1 (header only).
// When expanded, height is 1 + child height.
func (c *Collapsible) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w := 0
	h := 1 // header

	if c.expanded && c.child != nil {
		childSize := c.child.Measure(cs)
		w = childSize.W
		h += childSize.H
	}

	// Ensure header fits
	headerW := c.indent + 3 + len([]rune(c.title)) // indent + arrow + space + title
	if headerW > w {
		w = headerW
	}

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 0 {
		w = 0
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// SetBounds sets the position and size for this component.
func (c *Collapsible) SetBounds(r Rect) {
	c.mu.Lock()
	c.bounds = r
	child := c.child
	expanded := c.expanded
	c.mu.Unlock()

	// Propagate bounds to child when expanded
	if expanded && child != nil {
		childRect := Rect{
			X: r.X + c.indent,
			Y: r.Y + 1, // below header
			W: r.W - c.indent,
			H: r.H - 1,
		}
		if childRect.W < 0 {
			childRect.W = 0
		}
		if childRect.H < 0 {
			childRect.H = 0
		}
		child.SetBounds(childRect)
	}
}

// Bounds returns the current bounds.
func (c *Collapsible) Bounds() Rect {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.bounds
}

// Paint renders the header and (when expanded) the child component.
func (c *Collapsible) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Draw header on first row
	headerY := bounds.Y
	x := bounds.X + c.indent

	// Draw arrow indicator
	if c.showArrow && x < bounds.X+bounds.W {
		var arrow rune
		if c.expanded {
			arrow = '\u25be' // ▾ (small down triangle)
		} else {
			arrow = '\u25b8' // ▸ (small right triangle)
		}
		style := c.headerStyle
		if c.expanded {
			style = c.childStyle
		}
		buf.SetCell(x, headerY, buffer.NewCell(arrow, style))
		x++
	}

	// Space after arrow
	if x < bounds.X+bounds.W {
		buf.SetCell(x, headerY, buffer.NewCell(' ', buffer.DefaultStyle))
		x++
	}

	// Draw title
	style := c.headerStyle
	if c.expanded {
		style = c.childStyle
	}
	for _, r := range c.title {
		if x >= bounds.X+bounds.W {
			break
		}
		buf.SetCell(x, headerY, buffer.NewCell(r, style))
		x++
	}

	// Paint child when expanded
	if c.expanded && c.child != nil && bounds.H > 1 {
		c.child.Paint(buf)
	}
}

// Children returns the child component slice when expanded.
func (c *Collapsible) Children() []Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.expanded || c.child == nil {
		return nil
	}
	return []Component{c.child}
}
