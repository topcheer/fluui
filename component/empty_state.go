package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// EmptyState renders a centered placeholder when there's no data to display.
// It shows an icon, title, and optional description/hint. This is the
// standard pattern in modern apps for empty lists, search results with
// no matches, or before any conversation has started.
//
// Thread-safe. Zero-alloc Paint.
type EmptyState struct {
	BaseComponent
	mu sync.Mutex

	icon        string
	title       string
	description string
	hint        string
}

// NewEmptyState creates an empty state with title and optional description.
func NewEmptyState(title, description string) *EmptyState {
	return &EmptyState{
		BaseComponent: BaseComponent{id: GenerateID("empty")},
		title:         title,
		description:   description,
		icon:          "\u2728", // ✨ default sparkles icon
	}
}

// SetIcon changes the display icon.
func (e *EmptyState) SetIcon(icon string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.icon = icon
}

// SetTitle changes the title text.
func (e *EmptyState) SetTitle(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.title = s
}

// SetDescription changes the description text.
func (e *EmptyState) SetDescription(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.description = s
}

// SetHint changes the optional hint text (shown below description).
func (e *EmptyState) SetHint(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hint = s
}

// Title returns the current title.
func (e *EmptyState) Title() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.title
}

// Measure returns the desired size (fills available space).
func (e *EmptyState) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 5
	}
	return Size{W: maxW, H: maxH}
}

// Paint renders the empty state centered in the bounds.
func (e *EmptyState) Paint(buf *buffer.Buffer) {
	e.mu.Lock()
	icon := e.icon
	title := e.title
	desc := e.description
	hint := e.hint
	e.mu.Unlock()

	b := e.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	iconStyle := buffer.Style{Fg: th.Accent}
	titleStyle := buffer.Style{Fg: th.Fg}
	descStyle := buffer.Style{Fg: th.Muted}
	hintStyle := buffer.Style{Fg: th.BorderMuted}

	// Calculate vertical centering
	lines := 1 // title always present
	if icon != "" {
		lines++
	}
	if desc != "" {
		lines++
	}
	if hint != "" {
		lines++
	}

	y := b.Y + (b.H-lines)/2
	if y < b.Y {
		y = b.Y
	}

	// Draw icon
	if icon != "" {
		iconW := utf8.RuneCountInString(icon)
		x := b.X + (b.W-iconW)/2
		if x < b.X {
			x = b.X
		}
		buf.DrawText(x, y, icon, iconStyle)
		y++
	}

	// Draw title
	if title != "" {
		titleW := utf8.RuneCountInString(title)
		x := b.X + (b.W-titleW)/2
		if x < b.X {
			x = b.X
		}
		// Truncate if too wide
		if titleW > b.W {
			title = truncateRunes(title, b.W-1) + "\u2026"
		}
		buf.DrawText(x, y, title, titleStyle)
		y++
	}

	// Draw description
	if desc != "" {
		descW := utf8.RuneCountInString(desc)
		x := b.X + (b.W-descW)/2
		if x < b.X {
			x = b.X
		}
		if descW > b.W {
			desc = truncateRunes(desc, b.W-1) + "\u2026"
		}
		buf.DrawText(x, y, desc, descStyle)
		y++
	}

	// Draw hint
	if hint != "" {
		hintW := utf8.RuneCountInString(hint)
		x := b.X + (b.W-hintW)/2
		if x < b.X {
			x = b.X
		}
		if hintW > b.W {
			hint = truncateRunes(hint, b.W-1) + "\u2026"
		}
		buf.DrawText(x, y, hint, hintStyle)
	}
}

// CalloutType determines the visual style of a Callout.
type CalloutType int

const (
	CalloutInfo    CalloutType = iota
	CalloutWarning
	CalloutError
	CalloutSuccess
)

// Callout renders a prominent inline notice with an icon, title, and message.
// Used for warnings, errors, tips, and important information that the user
// must not miss. Inspired by GitHub callouts, Ant Design Alerts, and
// Textual's Static with styling.
//
// Thread-safe. Zero-alloc Paint.
type Callout struct {
	BaseComponent
	mu sync.Mutex

	calloutType CalloutType
	title       string
	message     string
}

// NewCallout creates a callout with the given type and message.
func NewCallout(ct CalloutType, message string) *Callout {
	return &Callout{
		BaseComponent: BaseComponent{id: GenerateID("callout")},
		calloutType:   ct,
		message:       message,
	}
}

// SetMessage changes the callout message.
func (c *Callout) SetMessage(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.message = s
}

// SetTitle sets an optional bold title shown before the message.
func (c *Callout) SetTitle(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.title = s
}

// SetType changes the callout type (info/warning/error/success).
func (c *Callout) SetType(ct CalloutType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calloutType = ct
}

// Message returns the current message.
func (c *Callout) Message() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.message
}

// Measure returns the desired size.
func (c *Callout) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 60
	}
	h := 1
	if c.message != "" {
		h = 2
	}
	maxH := cs.MaxHeight
	if maxH > 0 && h > maxH {
		h = maxH
	}
	return Size{W: maxW, H: h}
}

// Paint renders the callout.
func (c *Callout) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	ct := c.calloutType
	title := c.title
	msg := c.message
	c.mu.Unlock()

	b := c.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()

	// Determine icon and colors by type
	var icon string
	var fg, bg buffer.Color

	switch ct {
	case CalloutWarning:
		icon = "\u26a0" // ⚠
		fg = th.Warning
		bg = th.Warning
	case CalloutError:
		icon = "\u2716" // ✖
		fg = th.Error
		bg = th.Error
	case CalloutSuccess:
		icon = "\u2714" // ✔
		fg = th.Success
		bg = th.Success
	default: // CalloutInfo
		icon = "\u2139" // ℹ
		fg = th.Accent
		bg = th.Accent
	}

	iconStyle := buffer.Style{Fg: fg}
	titleStyle := buffer.Style{Fg: fg}
	msgStyle := buffer.Style{Fg: th.Muted}

	y := b.Y
	x := b.X

	// Title line with icon
	if title != "" {
		x += buf.DrawText(x, y, icon+" ", iconStyle)
		buf.DrawText(x, y, title, titleStyle)
		if b.H > 1 {
			y++
			x = b.X
		}
	} else {
		// No title — icon + message on same line
		x += buf.DrawText(x, y, icon+" ", iconStyle)
	}

	// Message
	if msg != "" {
		msgW := utf8.RuneCountInString(msg)
		availW := b.W - (x - b.X)
		if msgW > availW {
			msg = truncateRunes(msg, availW-1) + "\u2026"
		}
		buf.DrawText(x, y, msg, msgStyle)
	}

	// Left border bar
	for i := 0; i < b.H; i++ {
		by := b.Y + i
		if by < b.Y+b.H {
			buf.SetCell(b.X, by, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    bg,
			})
		}
	}
}
