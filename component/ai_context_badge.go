package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIContextBadge: AI Context Source Type Badge ───
//
// AIContextBadge renders a compact badge showing the type of context
// source an AI response used (e.g., RAG, fine-tune, system prompt, tool).
// Each type has a distinct color and icon.
//
// Usage:
//
//	cb := NewAIContextBadge()
//	cb.SetSource(ContextRAG, "docs.md")
//	cb.Paint(buf)

// ContextSourceType represents the context source.
type ContextSourceType int

const (
	ContextSystem   ContextSourceType = 0
	ContextRAG      ContextSourceType = 1
	ContextTool     ContextSourceType = 2
	ContextFineTune ContextSourceType = 3
	ContextMemory   ContextSourceType = 4
)

var contextIcons = [...]rune{'⚙', '📄', '🔧', '🎯', '🧠'}
var contextLabels = [...]string{"System", "RAG", "Tool", "FineTune", "Memory"}

// AIContextBadgeStyle holds styling.
type AIContextBadgeStyle struct {
	System   buffer.Style
	RAG      buffer.Style
	Tool     buffer.Style
	FineTune buffer.Style
	Memory   buffer.Style
	Name     buffer.Style
	Bracket  buffer.Style
}

// DefaultAIContextBadgeStyle returns defaults.
func DefaultAIContextBadgeStyle() AIContextBadgeStyle {
	return AIContextBadgeStyle{
		System:   buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Bold},
		RAG:      buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Tool:     buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		FineTune: buffer.Style{Fg: buffer.RGB(168, 85, 247), Flags: buffer.Bold},
		Memory:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Name:     buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		Bracket:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// AIContextBadge renders a context source badge.
type AIContextBadge struct {
	BaseComponent
	mu sync.Mutex

	sourceType ContextSourceType
	name       string
	style      AIContextBadgeStyle
	// cached
	labelStr string
	curStyle buffer.Style
}

// NewAIContextBadge creates an AIContextBadge.
func NewAIContextBadge() *AIContextBadge {
	cb := &AIContextBadge{sourceType: ContextSystem, style: DefaultAIContextBadgeStyle()}
	cb.SetID(GenerateID("ctxbadge"))
	cb.recomputeLocked()
	return cb
}

// SetSource sets the source type and optional name.
func (cb *AIContextBadge) SetSource(st ContextSourceType, name string) *AIContextBadge {
	cb.mu.Lock()
	if int(st) < 0 || int(st) >= len(contextLabels) {
		st = ContextSystem
	}
	cb.sourceType = st
	cb.name = name
	cb.recomputeLocked()
	cb.mu.Unlock()
	return cb
}

func (cb *AIContextBadge) recomputeLocked() {
	cb.labelStr = contextLabels[cb.sourceType]
	switch cb.sourceType {
	case ContextSystem:
		cb.curStyle = cb.style.System
	case ContextRAG:
		cb.curStyle = cb.style.RAG
	case ContextTool:
		cb.curStyle = cb.style.Tool
	case ContextFineTune:
		cb.curStyle = cb.style.FineTune
	case ContextMemory:
		cb.curStyle = cb.style.Memory
	}
}

// SourceType returns the current source type.
func (cb *AIContextBadge) SourceType() ContextSourceType {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.sourceType
}

// SetStyle sets custom style.
func (cb *AIContextBadge) SetStyle(s AIContextBadgeStyle) *AIContextBadge {
	cb.mu.Lock()
	cb.style = s
	cb.recomputeLocked()
	cb.mu.Unlock()
	return cb
}

// Measure returns preferred size.
func (cb *AIContextBadge) Measure(cs Constraints) Size {
	w := 16
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the context badge.
func (cb *AIContextBadge) Paint(buf *buffer.Buffer) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	b := cb.Bounds()
	x, y := b.X, b.Y

	sourceStyle := cb.curStyle
	nameStyle := cb.style.Name
	bracketStyle := cb.style.Bracket

	col := x

	// Icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: contextIcons[cb.sourceType], Fg: sourceStyle.Fg, Bg: sourceStyle.Bg, Flags: sourceStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
		col++
	}

	// Label
	for _, r := range cb.labelStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: sourceStyle.Fg, Bg: sourceStyle.Bg, Flags: sourceStyle.Flags, Width: 1})
		col++
	}

	// Name in brackets
	if cb.name != "" {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
			col++
		}
		for _, r := range cb.name {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (cb *AIContextBadge) Children() []Component { return nil }
