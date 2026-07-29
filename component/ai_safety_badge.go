package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AISafetyBadge: AI Content Safety Classification Badge ───
//
// AISafetyBadge renders a compact badge showing the safety classification
// of an AI response (safe, warning, blocked) with a category label.
// Useful for content moderation dashboards.
//
// Usage:
//
//	b := NewAISafetyBadge()
//	b.SetClassification(SafetySafe, "clean")
//	b.Paint(buf)

// SafetyLevel represents content safety level.
type SafetyLevel int

const (
	SafetySafe    SafetyLevel = 0
	SafetyWarning SafetyLevel = 1
	SafetyBlocked SafetyLevel = 2
	SafetyUnknown SafetyLevel = 3
)

// AISafetyBadgeStyle holds styling.
type AISafetyBadgeStyle struct {
	Safe    buffer.Style
	Warning buffer.Style
	Blocked buffer.Style
	Unknown buffer.Style
	Label   buffer.Style
	Bracket buffer.Style
}

// DefaultAISafetyBadgeStyle returns defaults.
func DefaultAISafetyBadgeStyle() AISafetyBadgeStyle {
	return AISafetyBadgeStyle{
		Safe:    buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Warning: buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Blocked: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Unknown: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Bracket: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

var safetyLevelLabels = [...]string{"SAFE", "WARN", "BLOCK", "????"}
var safetyLevelIcons = [...]rune{'✓', '⚠', '✗', '?'}

// AISafetyBadge renders a content safety badge.
type AISafetyBadge struct {
	BaseComponent
	mu sync.Mutex

	level    SafetyLevel
	category string
	style    AISafetyBadgeStyle
	// cached
	labelStr string
	icon     rune
}

// NewAISafetyBadge creates an AISafetyBadge.
func NewAISafetyBadge() *AISafetyBadge {
	b := &AISafetyBadge{level: SafetyUnknown, category: "", style: DefaultAISafetyBadgeStyle()}
	b.SetID(GenerateID("safety"))
	b.recomputeLocked()
	return b
}

// SetClassification sets the safety level and category.
func (b *AISafetyBadge) SetClassification(level SafetyLevel, category string) *AISafetyBadge {
	b.mu.Lock()
	if int(level) < 0 || int(level) > 3 { level = SafetyUnknown }
	b.level = level
	b.category = category
	b.recomputeLocked()
	b.mu.Unlock()
	return b
}

func (b *AISafetyBadge) recomputeLocked() {
	b.labelStr = safetyLevelLabels[b.level]
	b.icon = safetyLevelIcons[b.level]
	if b.category != "" {
		b.labelStr += ":" + b.category
	}
}

// Level returns the current safety level.
func (b *AISafetyBadge) Level() SafetyLevel {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.level
}

// SetStyle sets custom style.
func (b *AISafetyBadge) SetStyle(s AISafetyBadgeStyle) *AISafetyBadge {
	b.mu.Lock()
	b.style = s
	b.mu.Unlock()
	return b
}

// Measure returns preferred size.
func (b *AISafetyBadge) Measure(cs Constraints) Size {
	w := len(b.labelStr) + 4
	if w < 8 { w = 8 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the safety badge.
func (b *AISafetyBadge) Paint(buf *buffer.Buffer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	bx := b.Bounds()
	x, y := bx.X, bx.Y

	var levelStyle buffer.Style
	switch b.level {
	case SafetySafe:
		levelStyle = b.style.Safe
	case SafetyWarning:
		levelStyle = b.style.Warning
	case SafetyBlocked:
		levelStyle = b.style.Blocked
	default:
		levelStyle = b.style.Unknown
	}
	bracketStyle := b.style.Bracket

	col := x

	// [ prefix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
	// Icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: b.icon, Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
		col++
	}
	// space
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
		col++
	}
	// Label
	for _, r := range b.labelStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
		col++
	}
	// ] suffix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (b *AISafetyBadge) Children() []Component { return nil }
