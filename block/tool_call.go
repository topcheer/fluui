package block

import (
	"encoding/json"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ToolCallBlock displays a single tool invocation with its status.
type ToolCallBlock struct {
	BaseBlock
	toolName string
	rawArgs  string
}

// NewToolCallBlock creates a tool call block in streaming state.
func NewToolCallBlock(id, toolName, rawArgs string) *ToolCallBlock {
	return &ToolCallBlock{
		BaseBlock: NewBaseBlock(id, TypeToolCall),
		toolName:  toolName,
		rawArgs:   rawArgs,
	}
}

// ToolName returns the tool being called.
func (b *ToolCallBlock) ToolName() string { return b.toolName }

// RawArgs returns the raw arguments string.
func (b *ToolCallBlock) RawArgs() string { return b.rawArgs }

// Measure returns the desired size (always 1 row tall).
func (b *ToolCallBlock) Measure(cs component.Constraints) component.Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	return component.Size{W: maxW, H: 1}
}

// SetBounds sets the bounds.
func (b *ToolCallBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// truncateRunes truncates a string to maxRunes runes, appending "…" if truncated.
func truncateRunes(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	return string(r[:maxRunes-1]) + "…"
}

// runeWidth returns the visual width of a rune string in terminal cells.
// For simplicity, we use rune count (works for ASCII + box-drawing + emoji).
func runeCount(s string) int { return len([]rune(s)) }

// SerializeState serializes the tool call block's state to JSON.
func (b *ToolCallBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"tool_name": b.toolName,
		"args":      b.rawArgs,
	})
}

// DeserializeState restores the tool call block's state from JSON.
func (b *ToolCallBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		ToolName string `json:"tool_name"`
		Args     string `json:"args"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.toolName = s.ToolName
	b.rawArgs = s.Args
	b.MarkDirty()
	return nil
}

// Paint renders the tool call line.
// All width math uses rune counts, not byte lengths.
func (b *ToolCallBlock) Paint(buf *buffer.Buffer) {
	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	nameRunes := []rune(b.toolName)
	// Layout: "⏺ name(args) ✓" — count fixed overhead runes
	// "⏺ " = 2, "(" = 1, ") " = 2, status = 1 (✓/✗/…)
	overhead := 2 + 1 + 2 + 1 // = 6 runes for "⏺ () ✓"
	maxArgsRunes := bounds.W - len(nameRunes) - overhead
	if maxArgsRunes < 0 {
		maxArgsRunes = 0
	}
	argsPreview := truncateRunes(b.rawArgs, maxArgsRunes)

	var statusRune string
	switch b.State() {
	case BlockStreaming, BlockPending:
		statusRune = "…"
	case BlockComplete:
		statusRune = "✓"
	case BlockError:
		statusRune = "✗"
	}

	text := "⏺ " + b.toolName + "(" + argsPreview + ") " + statusRune

	// Truncate to bounds if still too long (rune-aware)
	text = truncateRunes(text, bounds.W)

	var style buffer.Style
	switch b.State() {
	case BlockStreaming, BlockPending:
		style = buffer.Style{Fg: theme.Get().Muted}
	case BlockComplete:
		style = buffer.Style{Fg: theme.Get().Success}
	case BlockError:
		style = buffer.Style{Fg: theme.Get().Error}
	}

	buf.DrawText(bounds.X, bounds.Y, text, style)
}
