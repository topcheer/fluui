package block

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ToolCallBlock displays a tool invocation with rich visualization.
//
// When collapsed (default): single-line summary with spinner/status + duration.
// When expanded: header line + pretty-printed JSON arguments in a bordered area.
//
// Thread-safe via BaseBlock's RWMutex.
type ToolCallBlock struct {
	BaseBlock
	toolName  string
	rawArgs   string
	expanded  bool
	spinnerF  int  // spinner frame counter for animation
	hasPretty bool // rawArgs is valid JSON
	prettyArg string
}

// toolCallSpinnerFrames provides the animation frames for in-progress tool calls.
var toolCallSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewToolCallBlock creates a tool call block in streaming state.
func NewToolCallBlock(id, toolName, rawArgs string) *ToolCallBlock {
	b := &ToolCallBlock{
		BaseBlock: NewBaseBlock(id, TypeToolCall),
		toolName:  toolName,
		rawArgs:   rawArgs,
	}
	// Try to pretty-print args as JSON
	if rawArgs != "" {
		var parsed any
		if err := json.Unmarshal([]byte(rawArgs), &parsed); err == nil {
			if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				b.prettyArg = string(formatted)
				b.hasPretty = true
			}
		}
	}
	return b
}

// ToolName returns the tool being called.
func (b *ToolCallBlock) ToolName() string { return b.toolName }

// RawArgs returns the raw arguments string.
func (b *ToolCallBlock) RawArgs() string { return b.rawArgs }

// Expanded returns whether the block is showing expanded details.
func (b *ToolCallBlock) Expanded() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.expanded
}

// SetExpanded controls the expand/collapse state.
func (b *ToolCallBlock) SetExpanded(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.expanded = v
	b.markDirtyLocked()
}

// Toggle switches between collapsed and expanded views.
func (b *ToolCallBlock) Toggle() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.expanded = !b.expanded
	b.markDirtyLocked()
}

// AdvanceSpinner increments the spinner frame counter for animation.
// Call this on each render tick to animate the spinner.
func (b *ToolCallBlock) AdvanceSpinner() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spinnerF++
	b.markDirtyLocked()
}

// Duration returns the elapsed time since the tool call started.
func (b *ToolCallBlock) Duration() time.Duration {
	return b.BaseBlock.Duration()
}

// formatDuration is shared with thinking.go (declared there).
// This file uses it for tool call duration display.

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

// runeCount returns the visual width of a rune string in terminal cells.
func runeCount(s string) int { return len([]rune(s)) }

// SerializeState serializes the tool call block's state to JSON.
func (b *ToolCallBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.Marshal(map[string]any{
		"tool_name": b.toolName,
		"args":      b.rawArgs,
		"expanded":  b.expanded,
	})
}

// DeserializeState restores the tool call block's state from JSON.
func (b *ToolCallBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		ToolName string `json:"tool_name"`
		Args     string `json:"args"`
		Expanded bool   `json:"expanded"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.toolName = s.ToolName
	b.rawArgs = s.Args
	b.expanded = s.Expanded
	// Re-pretty-print
	if s.Args != "" {
		var parsed any
		if err := json.Unmarshal([]byte(s.Args), &parsed); err == nil {
			if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				b.prettyArg = string(formatted)
				b.hasPretty = true
			}
		}
	}
	b.markDirtyLocked()
	return nil
}

// Measure returns the desired size.
// When collapsed: 1 row. When expanded: header + args lines + borders.
func (b *ToolCallBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	if !b.expanded || b.rawArgs == "" {
		return component.Size{W: maxW, H: 1}
	}

	// Expanded: count arg lines
	argsToShow := b.prettyArg
	if !b.hasPretty {
		argsToShow = b.rawArgs
	}
	argLines := strings.Count(argsToShow, "\n") + 1
	// Header(1) + top border(1) + arg lines + bottom border(1)
	h := 1 + 1 + argLines + 1
	return component.Size{W: maxW, H: h}
}

// SetBounds sets the bounds.
func (b *ToolCallBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// Paint renders the tool call.
func (b *ToolCallBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	if b.expanded && b.rawArgs != "" && bounds.H > 1 {
		b.paintExpanded(buf, bounds)
	} else {
		b.paintCollapsed(buf, bounds)
	}
}

// paintCollapsed renders the single-line view:
//
//	⠋ tool_name(args) 1.2s
//	✓ tool_name(args) 1.2s
//	✗ tool_name(args) 1.2s
func (b *ToolCallBlock) paintCollapsed(buf *buffer.Buffer, bounds component.Rect) {
	state := b.state
	durStr := formatDuration(b.Duration())

	// Status icon
	var icon string
	switch state {
	case BlockStreaming, BlockPending:
		icon = toolCallSpinnerFrames[b.spinnerF%len(toolCallSpinnerFrames)]
	case BlockComplete:
		icon = "✓"
	case BlockError:
		icon = "✗"
	}

	// Build: "icon name(args) duration"
	// Truncate args to fit available width
	nameRunes := []rune(b.toolName)
	durRunes := []rune(durStr)
	// "icon " = 2, "(" = 1, ") " = 2, " duration" = 1+len(dur)
	overhead := 2 + 1 + 2 + 1 + len(durRunes)
	maxArgsRunes := bounds.W - len(nameRunes) - overhead
	if maxArgsRunes < 0 {
		maxArgsRunes = 0
	}
	argsPreview := truncateRunes(b.rawArgs, maxArgsRunes)

	text := icon + " " + b.toolName + "(" + argsPreview + ") " + durStr
	text = truncateRunes(text, bounds.W)

	// Style based on state
	var style buffer.Style
	switch state {
	case BlockStreaming, BlockPending:
		style = buffer.Style{Fg: theme.Get().Muted}
	case BlockComplete:
		style = buffer.Style{Fg: theme.Get().Success}
	case BlockError:
		style = buffer.Style{Fg: theme.Get().Error}
	}

	buf.DrawText(bounds.X, bounds.Y, text, style)
}

// paintExpanded renders the multi-line view:
//
//	⠋ tool_name  1.2s
//	╭─ args ────────────╮
//	│ {"key": "value"}  │
//	╰───────────────────╯
func (b *ToolCallBlock) paintExpanded(buf *buffer.Buffer, bounds component.Rect) {
	w := bounds.W
	state := b.state
	durStr := formatDuration(b.Duration())

	// Status icon
	var icon string
	switch state {
	case BlockStreaming, BlockPending:
		icon = toolCallSpinnerFrames[b.spinnerF%len(toolCallSpinnerFrames)]
	case BlockComplete:
		icon = "✓"
	case BlockError:
		icon = "✗"
	}

	// --- Header line ---
	var headerStyle buffer.Style
	switch state {
	case BlockStreaming, BlockPending:
		headerStyle = buffer.Style{Fg: theme.Get().Muted}
	case BlockComplete:
		headerStyle = buffer.Style{Fg: theme.Get().Success}
	case BlockError:
		headerStyle = buffer.Style{Fg: theme.Get().Error}
	}
	header := icon + " " + b.toolName + "  " + durStr
	header = truncateRunes(header, w)
	buf.DrawText(bounds.X, bounds.Y, header, headerStyle)

	// --- Args box ---
	borderStyle := buffer.Style{Fg: theme.Get().Border}
	contentStyle := buffer.Style{Fg: theme.Get().Muted}

	argsToShow := b.prettyArg
	if !b.hasPretty {
		argsToShow = b.rawArgs
	}
	argLines := strings.Split(argsToShow, "\n")

	// Top border: ╭─ args ──╮
	labelRunes := []rune(" args ")
	labelW := len(labelRunes)
	dashesTotal := w - 2 - labelW
	if dashesTotal < 0 {
		dashesTotal = 0
	}
	leftDashes := dashesTotal / 2
	rightDashes := dashesTotal - leftDashes

	x := bounds.X
	topY := bounds.Y + 1
	buf.DrawText(x, topY, "╭", borderStyle)
	x++
	for i := 0; i < leftDashes; i++ {
		buf.DrawText(x, topY, "─", borderStyle)
		x++
	}
	for _, r := range labelRunes {
		buf.DrawText(x, topY, string(r), borderStyle)
		x++
	}
	for i := 0; i < rightDashes; i++ {
		buf.DrawText(x, topY, "─", borderStyle)
		x++
	}
	if x < bounds.X+w {
		buf.DrawText(x, topY, "╮", borderStyle)
	}

	// Content lines
	contentW := w - 2 // space between │ borders
	if contentW < 1 {
		contentW = 1
	}
	for i, line := range argLines {
		y := topY + 1 + i
		if y >= bounds.Y+bounds.H-1 {
			break
		}
		buf.DrawText(bounds.X, y, "│", borderStyle)
		lineRunes := []rune(line)
		if len(lineRunes) > contentW-1 {
			lineRunes = lineRunes[:contentW-1]
		}
		buf.DrawText(bounds.X+1, y, " "+string(lineRunes), contentStyle)
		buf.DrawText(bounds.X+w-1, y, "│", borderStyle)
	}

	// Bottom border: ╰──╯
	botY := bounds.Y + bounds.H - 1
	buf.DrawText(bounds.X, botY, "╰", borderStyle)
	for i := 1; i < w-1; i++ {
		buf.DrawText(bounds.X+i, botY, "─", borderStyle)
	}
	if w > 1 {
		buf.DrawText(bounds.X+w-1, botY, "╯", borderStyle)
	}
}
