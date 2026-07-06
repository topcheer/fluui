package block

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ToolResultBlock displays the output of a tool call.
//
// Features:
//   - Rounded border with title showing line count
//   - Auto-collapse when output exceeds maxPreview lines
//   - "show N more lines…" hint in collapsed view
//   - Duration tracking via BaseBlock
//   - State indicators: ✓ complete, ✗ error, ⠋ streaming
//
// Thread-safe via BaseBlock's RWMutex.
type ToolResultBlock struct {
	BaseBlock
	output      strings.Builder
	collapsed   bool
	maxPreview  int
	statusCode  int    // HTTP-style status code (0 = unknown, 200 = success, etc.)
	contentType string // "json", "text", "html", "xml", etc.
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

// SetOutput replaces the entire output.
func (b *ToolResultBlock) SetOutput(output string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output.Reset()
	b.output.WriteString(output)
	b.markDirtyLocked()
}

// SetStatusCode sets the result status code (e.g. 200 for HTTP success).
func (b *ToolResultBlock) SetStatusCode(code int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.statusCode = code
	b.markDirtyLocked()
}

// StatusCode returns the current status code.
func (b *ToolResultBlock) StatusCode() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.statusCode
}

// SetContentType sets the content type for rendering hints.
func (b *ToolResultBlock) SetContentType(ct string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.contentType = ct
	b.markDirtyLocked()
}

// ContentType returns the content type.
func (b *ToolResultBlock) ContentType() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.contentType
}

// Complete marks the block as done and auto-collapses long output.
func (b *ToolResultBlock) Complete() {
	b.mu.Lock()
	defer b.mu.Unlock()
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
func (b *ToolResultBlock) Collapsed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.collapsed
}

// SetCollapsed explicitly sets the collapsed state.
func (b *ToolResultBlock) SetCollapsed(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = v
	b.markDirtyLocked()
}

// Output returns the full output text.
func (b *ToolResultBlock) Output() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.output.String()
}

// LineCount returns the number of lines in the output.
func (b *ToolResultBlock) LineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	s := b.output.String()
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// MaxPreview returns the max lines shown when collapsed.
func (b *ToolResultBlock) MaxPreview() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.maxPreview
}

// SetMaxPreview sets the max lines shown when collapsed.
func (b *ToolResultBlock) SetMaxPreview(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n < 1 {
		n = 1
	}
	b.maxPreview = n
	b.markDirtyLocked()
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
		h = b.maxPreview + 2 + 1 // border top + preview lines + hint line + border bottom
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
		"result":       b.output.String(),
		"collapsed":    b.collapsed,
		"status_code":  b.statusCode,
		"content_type": b.contentType,
	})
}

// DeserializeState restores the tool result block's state from JSON.
func (b *ToolResultBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Result      string `json:"result"`
		Collapsed   bool   `json:"collapsed"`
		StatusCode  int    `json:"status_code"`
		ContentType string `json:"content_type"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output.Reset()
	b.output.WriteString(s.Result)
	b.collapsed = s.Collapsed
	b.statusCode = s.StatusCode
	b.contentType = s.ContentType
	b.markDirtyLocked()
	return nil
}

// Paint renders the tool result with a rounded border.
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
	mutedStyle := buffer.Style{Fg: theme.Get().Muted}

	// --- Status indicator ---
	var statusIcon string
	switch b.state {
	case BlockStreaming, BlockPending:
		statusIcon = "⠋"
	case BlockComplete:
		if b.statusCode >= 400 {
			statusIcon = "✗"
		} else {
			statusIcon = "✓"
		}
	case BlockError:
		statusIcon = "✗"
	}

	// --- Title: " Result (N lines) 1.2s " ---
	lineCount := 0
	if s := b.output.String(); s != "" {
		lineCount = strings.Count(s, "\n") + 1
	}
	durStr := formatDuration(b.Duration())

	labelText := " Result"
	if lineCount > 0 {
		labelText += " (" + itoa(lineCount) + " line"
		if lineCount != 1 {
			labelText += "s"
		}
		labelText += ")"
	}
	labelText += "  " + durStr + " "

	fullLabel := statusIcon + labelText
	labelRunes := []rune(fullLabel)
	labelW := len(labelRunes)

	// Draw top border: ╭ <label> ─── ╮
	dashesTotal := w - 2 - labelW
	if dashesTotal < 0 {
		// Label too long — truncate
		maxLabel := w - 2
		if maxLabel < 1 {
			maxLabel = 1
		}
		labelRunes = labelRunes[:maxLabel]
		dashesTotal = 0
	}

	// Status icon style
	var iconStyle buffer.Style
	switch b.state {
	case BlockComplete:
		if b.statusCode >= 400 {
			iconStyle = buffer.Style{Fg: theme.Get().Error}
		} else {
			iconStyle = buffer.Style{Fg: theme.Get().Success}
		}
	case BlockError:
		iconStyle = buffer.Style{Fg: theme.Get().Error}
	default:
		iconStyle = mutedStyle
	}

	x := bounds.X
	buf.DrawText(x, bounds.Y, "╭", borderStyle)
	x++
	// Draw status icon with icon style
	if len(labelRunes) > 0 {
		buf.DrawText(x, bounds.Y, string(labelRunes[0]), iconStyle)
		x++
	}
	// Draw rest of label with border style
	for i := 1; i < len(labelRunes); i++ {
		buf.DrawText(x, bounds.Y, string(labelRunes[i]), borderStyle)
		x++
	}
	// Draw remaining dashes
	for i := 0; i < dashesTotal; i++ {
		buf.DrawText(x, bounds.Y, "─", borderStyle)
		x++
	}
	if x <= bounds.X+w-1 {
		buf.DrawText(x, bounds.Y, "╮", borderStyle)
	}

	// --- Content lines ---
	allLines := strings.Split(b.output.String(), "\n")
	var displayLines []string
	isCollapsed := b.collapsed && len(allLines) > b.maxPreview
	if isCollapsed {
		displayLines = allLines[:b.maxPreview]
	} else {
		displayLines = allLines
	}

	contentW := w - 2 // space between │ borders
	if contentW < 1 {
		contentW = 1
	}

	for i, line := range displayLines {
		y := bounds.Y + 1 + i
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

	// --- "Show more" hint in collapsed view ---
	if isCollapsed {
		hintY := bounds.Y + 1 + len(displayLines)
		if hintY < bounds.Y+bounds.H-1 {
			hidden := len(allLines) - b.maxPreview
			hint := "  ⤷ " + itoa(hidden) + " more line"
			if hidden != 1 {
				hint += "s"
			}
			hint += "… (press Enter to expand)"
			hintRunes := []rune(hint)
			if len(hintRunes) > contentW-1 {
				hintRunes = hintRunes[:contentW-1]
			}
			buf.DrawText(bounds.X, hintY, "│", borderStyle)
			buf.DrawText(bounds.X+1, hintY, string(hintRunes), mutedStyle)
			buf.DrawText(bounds.X+w-1, hintY, "│", borderStyle)
		}
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

// itoa converts an int to its decimal string representation.
// Avoids strconv.Itoa import for this simple case.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
