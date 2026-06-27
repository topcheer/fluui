package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// BadgeVariant defines the visual style of a badge.
type BadgeVariant int

const (
	BadgeInfo BadgeVariant = iota
	BadgeSuccess
	BadgeWarning
	BadgeError
	BadgeCritical
	BadgeNeutral
)

// BadgeSize controls the padding and appearance of a badge.
type BadgeSize int

const (
	BadgeSizeSmall BadgeSize = iota
	BadgeSizeNormal
	BadgeSizeLarge
)

// Badge is a compact, colored label component used to highlight status,
// categories, counts, or other short metadata. It renders as an inline
// pill-shaped tag with background and foreground colors determined by its
// variant. Badges are immutable display elements — they do not handle key
// or mouse events.
//
// Concurrent safe via sync.RWMutex.
type Badge struct {
	BaseComponent
	mu      sync.RWMutex
	text    string
	variant BadgeVariant
	size    BadgeSize
	icon    string // optional leading icon character
	style   *buffer.Style
}

// NewBadge creates a badge with the given text and variant.
func NewBadge(text string, variant BadgeVariant) *Badge {
	return &Badge{
		text:    text,
		variant: variant,
		size:    BadgeSizeNormal,
	}
}

// NewBadgeWithSize creates a badge with explicit size.
func NewBadgeWithSize(text string, variant BadgeVariant, size BadgeSize) *Badge {
	return &Badge{
		text:    text,
		variant: variant,
		size:    size,
	}
}

// Text returns the badge display text.
func (b *Badge) Text() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.text
}

// SetText updates the badge display text.
func (b *Badge) SetText(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.text = s
}

// Variant returns the current badge variant.
func (b *Badge) Variant() BadgeVariant {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.variant
}

// SetVariant updates the badge variant.
func (b *Badge) SetVariant(v BadgeVariant) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variant = v
}

// Size returns the current badge size.
func (b *Badge) Size() BadgeSize {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// SetSize updates the badge size.
func (b *Badge) SetSize(s BadgeSize) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.size = s
}

// Icon returns the optional leading icon.
func (b *Badge) Icon() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.icon
}

// SetIcon sets an optional leading icon character (e.g. "●", "★").
func (b *Badge) SetIcon(icon string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.icon = icon
}

// SetStyle overrides the badge's colors. Pass nil to revert to variant defaults.
func (b *Badge) SetStyle(s *buffer.Style) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.style = s
}

// resolveColors returns (fg, bg) for the current variant + theme.
func (b *Badge) resolveColors() (buffer.Color, buffer.Color) {
	if b.style != nil {
		return b.style.Fg, b.style.Bg
	}

	t := theme.Get()
	switch b.variant {
	case BadgeInfo:
		return t.Bg, t.Accent
	case BadgeSuccess:
		return t.Bg, t.Success
	case BadgeWarning:
		return t.Bg, t.Warning
	case BadgeError:
		return t.Bg, t.Error
	case BadgeCritical:
		return t.Fg, t.Error
	case BadgeNeutral:
		return t.Fg, t.Muted
	default:
		return t.Fg, t.Muted
	}
}

// resolveFlags returns style flags based on variant.
func (b *Badge) resolveFlags() buffer.StyleFlags {
	switch b.variant {
	case BadgeCritical:
		return buffer.Bold | buffer.Reverse
	case BadgeError, BadgeWarning:
		return buffer.Bold
	default:
		return 0
	}
}

// padding returns left/right padding based on size.
func (b *Badge) padding() int {
	switch b.size {
	case BadgeSizeSmall:
		return 0
	case BadgeSizeLarge:
		return 2
	default:
		return 1
	}
}

// contentWidth returns the total visual width of badge content (icon + text + padding).
func (b *Badge) contentWidth() int {
	w := 0
	if b.icon != "" {
		w += buffer.StringWidth(b.icon) + 1 // icon + space
	}
	w += buffer.StringWidth(b.text)
	w += b.padding() * 2
	return w
}

// Measure returns the preferred size of the badge.
func (b *Badge) Measure(cs Constraints) Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	w := b.contentWidth()
	if w < 2 {
		w = 2
	}
	h := 1
	if b.size == BadgeSizeLarge {
		h = 1 // still single line, just wider
	}

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
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

// Paint draws the badge as a colored pill into the buffer.
func (b *Badge) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	fg, bg := b.resolveColors()
	flags := b.resolveFlags()
	pad := b.padding()

	x := b.bounds.X
	y := b.bounds.Y
	maxX := b.bounds.X + b.bounds.W

	// Left padding
	for i := 0; i < pad && x < maxX; i++ {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg, Flags: flags})
		x++
	}

	// Optional icon
	if b.icon != "" {
		for _, r := range b.icon {
			if x >= maxX {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: fg, Bg: bg, Flags: flags})
			x++
		}
		// Space after icon
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg, Flags: flags})
			x++
		}
	}

	// Text
	for _, r := range b.text {
		if x >= maxX {
			break
		}
		buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: fg, Bg: bg, Flags: flags})
		x++
	}

	// Right padding (fill remaining bounds with bg)
	for x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg, Flags: flags})
		x++
	}
}

// Children returns nil — Badge is a leaf component.
func (b *Badge) Children() []Component { return nil }

// --- BadgeGroup ---

// BadgeGroup is a container that lays out multiple badges horizontally
// with configurable spacing. It manages positioning but does not own
// the badge instances.
type BadgeGroup struct {
	BaseComponent
	mu      sync.RWMutex
	badges  []*Badge
	spacing int
}

// NewBadgeGroup creates an empty badge group.
func NewBadgeGroup() *BadgeGroup {
	return &BadgeGroup{spacing: 1}
}

// Add appends a badge to the group.
func (g *BadgeGroup) Add(b *Badge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.badges = append(g.badges, b)
}

// Badges returns a copy of the badge slice.
func (g *BadgeGroup) Badges() []*Badge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]*Badge, len(g.badges))
	copy(result, g.badges)
	return result
}

// Count returns the number of badges.
func (g *BadgeGroup) Count() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.badges)
}

// SetSpacing sets the gap between badges (in character cells).
func (g *BadgeGroup) SetSpacing(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.spacing = n
}

// Clear removes all badges from the group.
func (g *BadgeGroup) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.badges = nil
}

// Measure returns the preferred size for the badge group.
func (g *BadgeGroup) Measure(cs Constraints) Size {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.badges) == 0 {
		return Size{W: 0, H: 1}
	}

	totalW := 0
	for i, b := range g.badges {
		s := b.Measure(Constraints{})
		totalW += s.W
		if i < len(g.badges)-1 {
			totalW += g.spacing
		}
	}

	w := totalW
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 {
		w = 1
	}
	return Size{W: w, H: 1}
}

// Paint renders all badges left-to-right with spacing.
func (g *BadgeGroup) Paint(buf *buffer.Buffer) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	x := g.bounds.X
	y := g.bounds.Y
	maxX := g.bounds.X + g.bounds.W

	for i, b := range g.badges {
		w := b.Measure(Constraints{}).W
		if x+w > maxX {
			w = maxX - x
		}
		if w <= 0 {
			break
		}
		b.SetBounds(Rect{X: x, Y: y, W: w, H: 1})
		b.Paint(buf)
		x += w

		// Draw spacing
		if i < len(g.badges)-1 {
			for s := 0; s < g.spacing && x < maxX; s++ {
				t := theme.Get()
				buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: t.Fg, Bg: t.Bg})
				x++
			}
		}
	}
}

// Children returns the badge components as children.
func (g *BadgeGroup) Children() []Component {
	g.mu.RLock()
	defer g.mu.RUnlock()

	children := make([]Component, len(g.badges))
	for i, b := range g.badges {
		children[i] = b
	}
	return children
}

// --- BadgeFactory convenience constructors ---

// NewInfoBadge creates an info-style badge.
func NewInfoBadge(text string) *Badge {
	return NewBadge(text, BadgeInfo)
}

// NewSuccessBadge creates a success-style badge.
func NewSuccessBadge(text string) *Badge {
	return NewBadge(text, BadgeSuccess)
}

// NewWarningBadge creates a warning-style badge.
func NewWarningBadge(text string) *Badge {
	return NewBadge(text, BadgeWarning)
}

// NewErrorBadge creates an error-style badge.
func NewErrorBadge(text string) *Badge {
	return NewBadge(text, BadgeError)
}

// NewCriticalBadge creates a critical-style badge with reverse video.
func NewCriticalBadge(text string) *Badge {
	return NewBadge(text, BadgeCritical)
}

// NewNeutralBadge creates a neutral/muted badge.
func NewNeutralBadge(text string) *Badge {
	return NewBadge(text, BadgeNeutral)
}

// --- helpers ---

// VariantName returns a human-readable name for a badge variant.
func VariantName(v BadgeVariant) string {
	switch v {
	case BadgeInfo:
		return "info"
	case BadgeSuccess:
		return "success"
	case BadgeWarning:
		return "warning"
	case BadgeError:
		return "error"
	case BadgeCritical:
		return "critical"
	case BadgeNeutral:
		return "neutral"
	default:
		return "unknown"
	}
}

// ParseVariant converts a string name to a BadgeVariant.
// Returns BadgeNeutral for unrecognized names.
func ParseVariant(name string) BadgeVariant {
	switch name {
	case "info":
		return BadgeInfo
	case "success":
		return BadgeSuccess
	case "warning":
		return BadgeWarning
	case "error":
		return BadgeError
	case "critical":
		return BadgeCritical
	case "neutral":
		return BadgeNeutral
	default:
		return BadgeNeutral
	}
}

// SizeName returns a human-readable name for a badge size.
func SizeName(s BadgeSize) string {
	switch s {
	case BadgeSizeSmall:
		return "small"
	case BadgeSizeNormal:
		return "normal"
	case BadgeSizeLarge:
		return "large"
	default:
		return "unknown"
	}
}
