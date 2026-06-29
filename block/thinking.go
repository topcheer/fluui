package block

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
	"github.com/topcheer/fluui/theme"
)

// ThinkingBlock displays an AI's thinking/reasoning process.
// It is collapsible: collapsed shows a single summary line,
// expanded shows the full thinking content.
type ThinkingBlock struct {
	BaseBlock

	content  strings.Builder
	style    buffer.Style
	collapsed bool
	spinner  *animation.Spinner
	renderer *markdown.MarkdownRenderer
}

// NewThinkingBlock creates a new ThinkingBlock in streaming state.
func NewThinkingBlock(id string) *ThinkingBlock {
	bb := NewBaseBlock(id, TypeThinking)
	return &ThinkingBlock{
		BaseBlock: bb,
		collapsed: true, // Default: collapsed
		spinner:   animation.NewSpinner("dots"),
		style:     buffer.Style{Fg: theme.Get().ThinkingFg},
	}
}

// AppendDelta appends streaming content to the thinking text.
func (b *ThinkingBlock) AppendDelta(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.WriteString(delta)
	b.markDirtyLocked()
}

// Toggle expands or collapses the thinking content.
func (b *ThinkingBlock) Toggle() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = !b.collapsed
	b.markDirtyLocked()
}

// Collapsed returns whether the block is currently collapsed.
func (b *ThinkingBlock) Collapsed() bool { return b.collapsed }

// Content returns the full thinking text accumulated so far.
func (b *ThinkingBlock) Content() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.content.String()
}

// SetContent replaces the thinking text and marks the block dirty.
func (b *ThinkingBlock) SetContent(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.Reset()
	b.content.WriteString(s)
	b.markDirtyLocked()
}

// Measure returns the desired size.
// Collapsed: 1 row. Expanded: 1 header + content lines.
func (b *ThinkingBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	if b.collapsed {
		return component.Size{W: maxW, H: 1}
	}

	// Expanded: header (1) + content lines
	content := b.content.String()
	if content == "" {
		return component.Size{W: maxW, H: 1}
	}

	lines := markdown.WrapText(content, maxW)
	return component.Size{W: maxW, H: 1 + len(lines)}
}

// SetBounds sets the bounds and re-measures content width.
func (b *ThinkingBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// Paint renders the thinking block into the buffer.
func (b *ThinkingBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	maxW := bounds.W

	if b.collapsed {
		b.paintCollapsed(buf, bounds.X, bounds.Y, maxW)
	} else {
		b.paintExpanded(buf, bounds.X, bounds.Y, maxW)
	}
}

// paintCollapsed draws the single-line summary.
func (b *ThinkingBlock) paintCollapsed(buf *buffer.Buffer, x, y, w int) {
	var text string
	dur := b.Duration()
	if b.State() == BlockStreaming {
		text = "💭 Thinking... (" + formatDuration(dur) + ")"
	} else {
		text = "💭 Thought for " + formatDuration(dur)
	}
	r := []rune(text)
	if len(r) > w {
		text = string(r[:w])
	}
	buf.DrawText(x, y, text, b.style)
}

// paintExpanded draws the header line + content lines.
func (b *ThinkingBlock) paintExpanded(buf *buffer.Buffer, x, y, w int) {
	// Header
	var header string
	dur := b.Duration()
	if b.State() == BlockStreaming {
		header = "▼ 💭 Thinking... (" + formatDuration(dur) + ")"
	} else {
		header = "▼ 💭 Thought for " + formatDuration(dur)
	}
	hr := []rune(header)
	if len(hr) > w {
		header = string(hr[:w])
	}
	buf.DrawText(x, y, header, b.style)

	// Content
	content := b.content.String()
	if content == "" {
		return
	}

	lines := markdown.WrapText(content, w)
	contentStyle := b.style
	contentStyle.Flags |= buffer.Italic
	for i, line := range lines {
		drawY := y + 1 + i
		// Prefix with │ for visual grouping
		prefix := "│ "
		fullLine := prefix + line
		fr := []rune(fullLine)
		if len(fr) > w {
			fullLine = string(fr[:w])
		}
		buf.DrawText(x, drawY, fullLine, contentStyle)
	}
}

// SerializeState serializes the thinking block's state to JSON.
func (b *ThinkingBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.Marshal(map[string]any{
		"collapsed": b.collapsed,
		"content":   b.content.String(),
	})
}

// DeserializeState restores the thinking block's state from JSON.
func (b *ThinkingBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Collapsed bool   `json:"collapsed"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = s.Collapsed
	b.content.Reset()
	b.content.WriteString(s.Content)
	b.markDirtyLocked()
	return nil
}

// formatDuration formats a duration for display (e.g. "2.3s").
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
