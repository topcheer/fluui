package block

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/theme"
	"github.com/topcheer/fluui/internal/buffer"
)

// ToolResultBlock displays the output of a tool call.
// It is collapsible when output exceeds maxPreview lines.
type ToolResultBlock struct {
	BaseBlock
	output     strings.Builder
	collapsed  bool
	maxPreview int
}

// NewToolResultBlock creates a tool result block in streaming state.
func NewToolResultBlock(id string) *ToolResultBlock {
	return &ToolResultBlock{
		BaseBlock:  NewBaseBlock(id, TypeToolResult),
		collapsed:  false,
		maxPreview: 5,
	}
}

// AppendDelta appends output text.
func (b *ToolResultBlock) AppendDelta(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output.WriteString(delta)
	b.markDirtyLocked()
}

// Complete marks the block as done and auto-collapses long output.
func (b *ToolResultBlock) Complete() {
	b.state = BlockComplete
	b.endedAt = time.Now()
	lines := strings.Count(b.output.String(), "\n") + 1
	if lines > b.maxPreview {
		b.collapsed = true
	}
	b.markDirtyLocked()
}

// Toggle expands or collapses.
func (b *ToolResultBlock) Toggle() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = !b.collapsed
	b.markDirtyLocked()
}

// Collapsed returns current collapse state.
func (b *ToolResultBlock) Collapsed() bool { return b.collapsed }

// Output returns the full output text.
func (b *ToolResultBlock) Output() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.output.String()
}

// Measure returns the desired height.
func (b *ToolResultBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	contentLines := 0
	if s := b.output.String(); s != "" {
		contentLines = strings.Count(s, "\n") + 1
	}

	var h int
	if b.collapsed && contentLines > b.maxPreview {
		h = b.maxPreview + 2 // border top+bottom
	} else if contentLines == 0 {
		h = 2 // just borders, no content
	} else {
		h = contentLines + 2
	}
	if h < 1 {
		h = 1
	}

	return component.Size{W: maxW, H: h}
}

// SetBounds sets the bounds.
func (b *ToolResultBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// SerializeState serializes the tool result block's state to JSON.
func (b *ToolResultBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.Marshal(map[string]any{
		"result":    b.output.String(),
		"collapsed": b.collapsed,
	})
}

// DeserializeState restores the tool result block's state from JSON.
func (b *ToolResultBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Result    string `json:"result"`
		Collapsed bool   `json:"collapsed"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output.Reset()
	b.output.WriteString(s.Result)
	b.collapsed = s.Collapsed
	b.markDirtyLocked()
	return nil
}

// Paint renders the tool result with a rounded border.
//
// All width calculations use rune count, not byte length,
// because box-drawing characters are multi-byte UTF-8.
func (b *ToolResultBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.Bounds()
	w := bounds.W
	if w <= 0 || bounds.H <= 0 {
		return
	}

	borderStyle := buffer.Style{Fg: theme.Get().Border}
	contentStyle := buffer.Style{Fg: theme.Get().ToolResultFg}

	// --- Top border: ╭─ Result ───...───╮ ---
	// " Result " is 9 runes (with spaces), ╭ and ╮ are 1 cell each.
	labelRunes := []rune(" Result ")
	labelW := len(labelRunes) // visual width = rune count (all ASCII)
	dashesTotal := w - 2 - labelW // total ─ needed
	if dashesTotal < 0 {
		dashesTotal = 0
	}
	leftDashes := dashesTotal / 2
	rightDashes := dashesTotal - leftDashes

	// Draw top border cell by cell
	x := bounds.X
	buf.DrawText(x, bounds.Y, "╭", borderStyle)
	x++
	for i := 0; i < leftDashes; i++ {
		buf.DrawText(x, bounds.Y, "─", borderStyle)
		x++
	}
	for _, r := range labelRunes {
		buf.DrawText(x, bounds.Y, string(r), borderStyle)
		x++
	}
	for i := 0; i < rightDashes; i++ {
		buf.DrawText(x, bounds.Y, "─", borderStyle)
		x++
	}
	if x < bounds.X+w {
		buf.DrawText(x, bounds.Y, "╮", borderStyle)
	}

	// --- Content lines ---
	allLines := strings.Split(b.output.String(), "\n")
	var displayLines []string
	if b.collapsed && len(allLines) > b.maxPreview {
		displayLines = allLines[:b.maxPreview]
	} else {
		displayLines = allLines
	}

	for i, line := range displayLines {
		y := bounds.Y + 1 + i
		if y >= bounds.Y+bounds.H-1 {
			break
		}
		// Left border + content
		buf.DrawText(bounds.X, y, "│", borderStyle)
		// Truncate line to fit within content area (w - 2 for borders, -1 for left pad)
		contentW := w - 2 // space between │ borders
		if contentW < 1 {
			continue
		}
		lineRunes := []rune(line)
		if len(lineRunes) > contentW-1 {
			lineRunes = lineRunes[:contentW-1]
		}
		buf.DrawText(bounds.X+1, y, " "+string(lineRunes), contentStyle)
		// Right border
		buf.DrawText(bounds.X+w-1, y, "│", borderStyle)
	}

	// --- Bottom border: ╰──...──╯ ---
	botY := bounds.Y + bounds.H - 1
	buf.DrawText(bounds.X, botY, "╰", borderStyle)
	for i := 1; i < w-1; i++ {
		buf.DrawText(bounds.X+i, botY, "─", borderStyle)
	}
	if w > 1 {
		buf.DrawText(bounds.X+w-1, botY, "╯", borderStyle)
	}
}
