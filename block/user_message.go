package block

import (
	"encoding/json"
	"strings"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/theme"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
)

// UserMessageBlock displays a user's message in the conversation.
type UserMessageBlock struct {
	BaseBlock
	content string
}

// NewUserMessageBlock creates a user message block.
func NewUserMessageBlock(id, content string) *UserMessageBlock {
	b := &UserMessageBlock{
		BaseBlock: NewBaseBlock(id, TypeUserMessage),
		content:   content,
	}
	b.state = BlockComplete // User messages are immediately complete
	return b
}

// Content returns the message text.
func (b *UserMessageBlock) Content() string {
	return b.content
}

// SetContent replaces the message text and marks the block dirty.
func (b *UserMessageBlock) SetContent(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content = s
	b.markDirtyLocked()
}

// Measure returns the size based on wrapped content.
func (b *UserMessageBlock) Measure(cs component.Constraints) component.Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	if b.content == "" {
		return component.Size{W: maxW, H: 1}
	}
	totalLines := 0
	for _, line := range strings.Split(b.content, "\n") {
		wrapped := markdown.WrapText(line, maxW)
		totalLines += len(wrapped)
	}
	if totalLines == 0 {
		totalLines = 1
	}
	return component.Size{W: maxW, H: totalLines}
}

// SetBounds sets the bounds.
func (b *UserMessageBlock) SetBounds(r component.Rect) {
	b.BaseBlock.SetBounds(r)
}

// SerializeState serializes the user message block's content to JSON.
func (b *UserMessageBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"content": b.content,
	})
}

// DeserializeState restores the user message block's content from JSON.
func (b *UserMessageBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.content = s.Content
	b.MarkDirty()
	return nil
}

// Paint renders the user message in cyan.
func (b *UserMessageBlock) Paint(buf *buffer.Buffer) {
	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}
	if b.content == "" {
		return
	}
	style := buffer.Style{Fg: theme.Get().UserMsgFg}
	rowIdx := 0
	for _, rawLine := range strings.Split(b.content, "\n") {
		for _, wrappedLine := range markdown.WrapText(rawLine, bounds.W) {
			y := bounds.Y + rowIdx
			if y >= bounds.Y+bounds.H {
				return
			}
			buf.DrawText(bounds.X, y, wrappedLine, style)
			rowIdx++
		}
	}
}
