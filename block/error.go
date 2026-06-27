package block

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ErrorBlock displays an error message with a red border.
// It is always expanded by default — errors should be immediately visible.
type ErrorBlock struct {
	BaseBlock

	message   string
	timestamp time.Time
}

// NewErrorBlock creates an ErrorBlock in BlockError state.
func NewErrorBlock(id string) *ErrorBlock {
	return &ErrorBlock{
		BaseBlock: NewBaseBlock(id, TypeError),
		timestamp: time.Now(),
	}
}

// NewErrorBlockWithMessage creates an ErrorBlock with a pre-set message,
// immediately in the BlockError state.
func NewErrorBlockWithMessage(id string, message string) *ErrorBlock {
	b := NewErrorBlock(id)
	b.message = message
	b.state = BlockError
	b.endedAt = time.Now()
	b.dirty = true
	return b
}

// AppendDelta appends to the error message (used during streaming error output).
func (b *ErrorBlock) AppendDelta(delta string) {
	b.message += delta
	b.markDirtyLocked()
}

// Fail sets the error message and transitions to BlockError state.
func (b *ErrorBlock) Fail(err error) {
	b.message = err.Error()
	b.BaseBlock.Fail(err)
}

// Message returns the error message text.
func (b *ErrorBlock) Message() string { return b.message }

// Timestamp returns when the error occurred.
func (b *ErrorBlock) Timestamp() time.Time { return b.timestamp }

// Measure returns the desired size for the error block.
// Height = top border + content lines + bottom border.
func (b *ErrorBlock) Measure(cs component.Constraints) component.Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	contentLines := 0
	if b.message != "" {
		contentLines = strings.Count(b.message, "\n") + 1
	}

	h := contentLines + 2 // border top + bottom
	if h < 2 {
		h = 2
	}

	return component.Size{W: maxW, H: h}
}

// SetBounds sets the bounds.
func (b *ErrorBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// SerializeState serializes the error block's state to JSON.
func (b *ErrorBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"message": b.message,
	})
}

// DeserializeState restores the error block's state from JSON.
func (b *ErrorBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.message = s.Message
	b.MarkDirty()
	return nil
}

// Paint renders the error block with a red border.
func (b *ErrorBlock) Paint(buf *buffer.Buffer) {
	bounds := b.Bounds()
	w := bounds.W
	if w <= 0 || bounds.H <= 0 {
		return
	}

	t := theme.Get()
	borderStyle := buffer.Style{Fg: t.Error, Flags: buffer.Bold}
	labelStyle := buffer.Style{Fg: t.Error, Flags: buffer.Bold}
	contentStyle := buffer.Style{Fg: t.Error}

	// --- Top border: ╭─ ERROR ───...───╮ ---
	labelRunes := []rune(" ERROR ")
	labelW := len(labelRunes)
	dashesTotal := w - 2 - labelW
	if dashesTotal < 0 {
		dashesTotal = 0
	}
	leftDashes := dashesTotal / 2
	rightDashes := dashesTotal - leftDashes

	x := bounds.X
	buf.DrawText(x, bounds.Y, "╭", borderStyle)
	x++
	for i := 0; i < leftDashes; i++ {
		buf.DrawText(x, bounds.Y, "─", borderStyle)
		x++
	}
	for _, r := range labelRunes {
		buf.DrawText(x, bounds.Y, string(r), labelStyle)
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
	contentLines := strings.Split(b.message, "\n")
	contentW := w - 2 // space between │ borders
	for i, line := range contentLines {
		y := bounds.Y + 1 + i
		if y >= bounds.Y+bounds.H-1 {
			break
		}
		// Left border
		buf.DrawText(bounds.X, y, "│", borderStyle)
		// Content (truncated to fit)
		if contentW > 1 {
			lineRunes := []rune(line)
			if len(lineRunes) > contentW-1 {
				lineRunes = lineRunes[:contentW-1]
			}
			buf.DrawText(bounds.X+1, y, " "+string(lineRunes), contentStyle)
		}
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
