package component

import (
	"sync"
	"unicode"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// AvatarSize controls the display size of an Avatar.
type AvatarSize int

const (
	// AvatarSmall renders a single-cell avatar (1 initial).
	AvatarSmall AvatarSize = iota
	// AvatarMedium renders a 3-cell-wide avatar (2 initials + padding).
	AvatarMedium
	// AvatarLarge renders a 3-cell-wide avatar with bold styling.
	AvatarLarge
)

// avatarPalette is the set of background colors used for auto-assignment.
// Each entry is a pleasant TrueColor that works on dark terminals.
var avatarPalette = [...]uint32{
	0x6366F1, // Indigo
	0x8B5CF6, // Violet
	0xEC4899, // Pink
	0xF43F5E, // Rose
	0xF97316, // Orange
	0xEAB308, // Yellow
	0x22C55E, // Green
	0x14B8A6, // Teal
	0x06B6D4, // Cyan
	0x3B82F6, // Blue
	0xA855F7, // Purple
	0xEF4444, // Red
}

// Avatar displays a user or entity avatar as a colored block with initials.
// The background color is auto-derived from a hash of the name for
// consistent identification (same name → same color, like GitHub/Slack).
// Pass an explicit color override via SetBg.
//
// Designed for AI chat interfaces, message bubbles, and session lists.
//
// Concurrent safe via sync.RWMutex.
type Avatar struct {
	BaseComponent
	mu       sync.RWMutex
	name     string
	initials string // override; empty = auto-extract from name
	icon     string // emoji/icon override; non-empty = use instead of initials
	size     AvatarSize
	bg       buffer.Color // 0 = auto from name hash
}

// NewAvatar creates an avatar for the given name.
func NewAvatar(name string) *Avatar {
	return &Avatar{
		BaseComponent: BaseComponent{id: GenerateID("avatar")},
		name:          name,
		size:          AvatarMedium,
	}
}

// Name returns the avatar's display name.
func (a *Avatar) Name() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.name
}

// SetName updates the name and resets auto-derived state.
func (a *Avatar) SetName(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.name = name
}

// Initials returns the display initials (override or auto-extracted).
func (a *Avatar) Initials() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.initialsLocked()
}

func (a *Avatar) initialsLocked() string {
	if a.initials != "" {
		return a.initials
	}
	var buf [2]byte
	n := extractInitialsInto(a.name, &buf)
	return string(buf[:n])
}

// SetInitials overrides the auto-extracted initials.
// Pass "" to revert to auto-extraction from Name.
func (a *Avatar) SetInitials(s string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.initials = s
}

// Icon returns the emoji/icon override (empty = none).
func (a *Avatar) Icon() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.icon
}

// SetIcon sets an emoji or icon character to display instead of initials.
// Pass "" to revert to initials.
func (a *Avatar) SetIcon(icon string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.icon = icon
}

// Size returns the avatar size.
func (a *Avatar) AvatarSize() AvatarSize {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.size
}

// SetSize updates the avatar size.
func (a *Avatar) SetSize(s AvatarSize) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.size = s
}

// Bg returns the background color (0 = auto).
func (a *Avatar) Bg() buffer.Color {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.bg
}

// SetBg overrides the background color. Pass buffer.Color{} (zero) for auto.
func (a *Avatar) SetBg(c buffer.Color) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bg = c
}

// resolveBgLocked returns the effective background color.
func (a *Avatar) resolveBgLocked() buffer.Color {
	if a.bg.Type != buffer.ColorNone {
		return a.bg
	}
	return avatarColorForName(a.name)
}

// Measure returns the preferred size.
// Small: 1x1, Medium: 3x1, Large: 3x1.
func (a *Avatar) Measure(cs Constraints) Size {
	a.mu.RLock()
	defer a.mu.RUnlock()

	w := a.measureWidthLocked()
	h := 1

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

func (a *Avatar) measureWidthLocked() int {
	switch a.size {
	case AvatarSmall:
		return 1
	default:
		return 3
	}
}

// Paint draws the avatar as a colored block with initials or icon.
// Zero allocations on the hot path.
func (a *Avatar) Paint(buf *buffer.Buffer) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	bounds := a.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	bg := a.resolveBgLocked()
	t := theme.Get()
	fg := t.Bg // text on colored bg = terminal background color

	flags := buffer.StyleFlags(0)
	if a.size == AvatarLarge {
		flags = buffer.Bold
	}
	style := buffer.Style{Fg: fg, Bg: bg, Flags: flags}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Draw content character(s)
	if a.icon != "" {
		// Draw icon centered
		x += buf.DrawText(x, y, a.icon, style)
		// Fill remaining width with bg
		for x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg, Flags: flags})
			x++
		}
		return
	}

	// Compute initials bytes on the stack (zero alloc)
	var initBytes [2]byte
	initN := 0
	if a.initials != "" {
		for i := 0; i < len(a.initials) && initN < 2; i++ {
			c := a.initials[i]
			if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
				if c >= 'a' {
					c -= 32 // to uppercase
				}
				initBytes[initN] = c
				initN++
			}
		}
	} else {
		initN = extractInitialsInto(a.name, &initBytes)
	}
	if initN == 0 {
		initBytes[0] = '?'
		initN = 1
	}

	if a.size == AvatarSmall {
		// Single character
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: rune(initBytes[0]), Width: 1, Fg: fg, Bg: bg, Flags: flags})
		}
		return
	}

	// Medium/Large: draw initials then pad with bg spaces
	for i := 0; i < initN && x < maxX; i++ {
		buf.SetCell(x, y, buffer.Cell{Rune: rune(initBytes[i]), Width: 1, Fg: fg, Bg: bg, Flags: flags})
		x++
	}
	for x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg, Flags: flags})
		x++
	}
}

// extractInitialsInto fills dst with 0-2 uppercase Latin initials from name.
// Only whitespace separates words; hyphens and apostrophes are part of names.
// Non-ASCII letters (CJK, etc.) are skipped.
// Returns the number of bytes written (0 if no Latin letters found).
func extractInitialsInto(name string, dst *[2]byte) int {
	if name == "" {
		return 0
	}

	n := 0
	inWord := false

	for _, r := range name {
		if unicode.IsLetter(r) {
			if !inWord && n < 2 {
				up := unicode.ToUpper(r)
				if up < 128 {
					dst[n] = byte(up)
					n++
				}
				// Non-ASCII letters: skip but still mark inWord
			}
			inWord = true
		} else if unicode.IsSpace(r) {
			inWord = false
			// hyphens, apostrophes, etc. stay in-word
		}
	}

	return n
}

// extractInitials extracts 1-2 uppercase initials from a name as a string.
func extractInitials(name string) string {
	var buf [2]byte
	n := extractInitialsInto(name, &buf)
	if n == 0 {
		return ""
	}
	return string(buf[:n])
}

// avatarColorForName returns a deterministic TrueColor for a name string
// using inline FNV-1a hash modulo the palette size (zero allocation).
func avatarColorForName(name string) buffer.Color {
	var h uint32 = 2166136261 // FNV offset basis
	for i := 0; i < len(name); i++ {
		h ^= uint32(name[i])
		h *= 16777619 // FNV prime
	}
	idx := int(h) % len(avatarPalette)
	return buffer.RGB(
		uint8(avatarPalette[idx]>>16),
		uint8(avatarPalette[idx]>>8),
		uint8(avatarPalette[idx]),
	)
}
