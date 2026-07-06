package block

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
	"github.com/topcheer/fluui/theme"
)

// thinkingSpinnerFrames provides braille spinner animation for thinking state.
var thinkingSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ThinkingBlock displays an AI's thinking/reasoning process.
// It is collapsible: collapsed shows a single summary line with animated spinner
// and content preview, expanded shows the full thinking content rendered as markdown.
//
// Thread-safe via BaseBlock's RWMutex.
type ThinkingBlock struct {
	BaseBlock

	content   strings.Builder
	style     buffer.Style
	collapsed bool
	spinnerF  int  // spinner frame index for animation
	renderer  *markdown.MarkdownRenderer

	// Caching for expanded view
	cachedBlocks  []*markdown.Block
	cachedText    string
	cachedW       int
	cacheFillDone bool
}

// NewThinkingBlock creates a new ThinkingBlock in streaming state.
func NewThinkingBlock(id string) *ThinkingBlock {
	bb := NewBaseBlock(id, TypeThinking)
	return &ThinkingBlock{
		BaseBlock: bb,
		collapsed: true,
		style:     buffer.Style{Fg: theme.Get().ThinkingFg},
	}
}

// AppendDelta appends streaming content to the thinking text.
func (b *ThinkingBlock) AppendDelta(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.WriteString(delta)
	b.cachedText = "" // invalidate cache
	b.markDirtyLocked()
}

// Toggle expands or collapses the thinking content.
func (b *ThinkingBlock) Toggle() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = !b.collapsed
	b.markDirtyLocked()
}

// Expand expands the thinking content.
func (b *ThinkingBlock) Expand() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.collapsed {
		return
	}
	b.collapsed = false
	b.markDirtyLocked()
}

// Collapse collapses the thinking content.
func (b *ThinkingBlock) Collapse() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.collapsed {
		return
	}
	b.collapsed = true
	b.markDirtyLocked()
}

// Collapsed returns whether the block is currently collapsed.
func (b *ThinkingBlock) Collapsed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.collapsed
}

// SetCollapsed sets the collapsed state.
func (b *ThinkingBlock) SetCollapsed(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = v
	b.markDirtyLocked()
}

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
	b.cachedText = "" // invalidate cache
	b.markDirtyLocked()
}

// AdvanceSpinner advances the spinner animation by one frame.
// Call this on every render tick for smooth animation.
func (b *ThinkingBlock) AdvanceSpinner() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spinnerF = (b.spinnerF + 1) % len(thinkingSpinnerFrames)
	b.markDirtyLocked()
}

// SpinnerFrame returns the current spinner frame string.
func (b *ThinkingBlock) SpinnerFrame() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.spinnerF < 0 || b.spinnerF >= len(thinkingSpinnerFrames) {
		return thinkingSpinnerFrames[0]
	}
	return thinkingSpinnerFrames[b.spinnerF]
}

// CharCount returns the number of characters in the thinking content.
func (b *ThinkingBlock) CharCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.content.Len()
}

// PreviewLine returns a truncated first line of content for collapsed preview.
func (b *ThinkingBlock) PreviewLine(maxLen int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	content := b.content.String()
	if content == "" {
		return ""
	}
	// Skip leading whitespace/newlines
	content = strings.TrimLeft(content, " \t\n\r")
	// Take first line
	if idx := strings.IndexByte(content, '\n'); idx >= 0 {
		content = content[:idx]
	}
	content = strings.TrimSpace(content)
	r := []rune(content)
	if len(r) > maxLen {
		return string(r[:maxLen-1]) + "…"
	}
	return content
}

// getCachedBlocks returns cached markdown blocks, re-rendering only when needed.
func (b *ThinkingBlock) getCachedBlocks(w int) []*markdown.Block {
	content := b.content.String()
	if b.cacheFillDone && content == b.cachedText && w == b.cachedW {
		return b.cachedBlocks
	}
	// Re-render
	if b.renderer == nil {
		b.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), w)
	} else if w != b.cachedW {
		b.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), w)
	}
	blocks, _ := b.renderer.Render(content)
	b.cachedBlocks = blocks
	b.cachedText = content
	b.cachedW = w
	b.cacheFillDone = true
	return blocks
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

	content := b.content.String()
	if content == "" {
		return component.Size{W: maxW, H: 1}
	}

	blocks := b.getCachedBlocks(maxW)
	totalH := 1 // header
	for _, blk := range blocks {
		totalH += len(blk.Cells)
	}
	if totalH < 1 {
		totalH = 1
	}
	return component.Size{W: maxW, H: totalH}
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
		b.paintExpanded(buf, bounds.X, bounds.Y, maxW, bounds.H)
	}
}

// paintCollapsed draws the single-line summary with animated spinner.
func (b *ThinkingBlock) paintCollapsed(buf *buffer.Buffer, x, y, w int) {
	dur := b.Duration()
	content := b.content.String()
	isStreaming := b.State() == BlockStreaming

	var text string
	if isStreaming {
		spinner := thinkingSpinnerFrames[b.spinnerF%len(thinkingSpinnerFrames)]
		if content == "" {
			text = spinner + " Thinking... (" + formatDuration(dur) + ")"
		} else {
			// Show preview of thinking content
			preview := b.PreviewLineUnlocked(w - 40)
			if preview == "" {
				text = spinner + " Thinking... (" + formatDuration(dur) + ")"
			} else {
				text = spinner + " " + preview + " (" + formatDuration(dur) + ")"
			}
		}
	} else {
		// Complete state
		preview := b.PreviewLineUnlocked(w - 40)
		if preview == "" {
			text = "✓ Thought for " + formatDuration(dur)
		} else {
			text = "✓ " + preview + " — " + formatDuration(dur)
		}
	}
	r := []rune(text)
	if len(r) > w {
		text = string(r[:w])
	}
	buf.DrawText(x, y, text, b.style)
}

// PreviewLineUnlocked is PreviewLine without locking (for internal use).
func (b *ThinkingBlock) PreviewLineUnlocked(maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	content := b.content.String()
	if content == "" {
		return ""
	}
	content = strings.TrimLeft(content, " \t\n\r")
	if idx := strings.IndexByte(content, '\n'); idx >= 0 {
		content = content[:idx]
	}
	content = strings.TrimSpace(content)
	r := []rune(content)
	if len(r) > maxLen {
		return string(r[:maxLen-1]) + "…"
	}
	return content
}

// paintExpanded draws the header line + markdown-rendered content.
func (b *ThinkingBlock) paintExpanded(buf *buffer.Buffer, x, y, w, h int) {
	dur := b.Duration()
	isStreaming := b.State() == BlockStreaming

	// Header
	var header string
	if isStreaming {
		spinner := thinkingSpinnerFrames[b.spinnerF%len(thinkingSpinnerFrames)]
		header = "▼ " + spinner + " Thinking... (" + formatDuration(dur) + ")"
	} else {
		header = "▼ ✓ Thought for " + formatDuration(dur)
	}
	hr := []rune(header)
	if len(hr) > w {
		header = string(hr[:w])
	}
	headerStyle := b.style
	headerStyle.Flags |= buffer.Bold
	buf.DrawText(x, y, header, headerStyle)

	// Content
	content := b.content.String()
	if content == "" {
		return
	}

	// Render markdown blocks
	blocks := b.getCachedBlocks(w)
	contentStyle := b.style
	contentStyle.Flags |= buffer.Italic

	drawY := y + 1
	for _, blk := range blocks {
		for _, row := range blk.Cells {
			if drawY >= y+h {
				return // out of bounds
			}
			// Draw with left margin for visual grouping
			margin := 2
			for i, cell := range row {
				drawX := x + margin + i
				if drawX >= x+w {
					break
				}
				if cell.Width > 0 {
					buf.SetCell(drawX, drawY, cell)
				}
			}
			drawY++
		}
		// Add spacing between blocks
		if drawY < y+h {
			drawY++
		}
	}

	// Streaming cursor at end
	if isStreaming && drawY < y+h {
		cursorStyle := b.style
		cursorStyle.Flags |= buffer.Reverse
		buf.DrawText(x+2, drawY, "▋", cursorStyle)
	}
}

// SerializeState serializes the thinking block's state to JSON.
func (b *ThinkingBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.Marshal(map[string]any{
		"collapsed": b.collapsed,
		"content":   b.content.String(),
		"spinnerF":  b.spinnerF,
	})
}

// DeserializeState restores the thinking block's state from JSON.
func (b *ThinkingBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Collapsed bool   `json:"collapsed"`
		Content   string `json:"content"`
		SpinnerF  int    `json:"spinnerF"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collapsed = s.Collapsed
	b.content.Reset()
	b.content.WriteString(s.Content)
	b.spinnerF = s.SpinnerF
	b.cachedText = "" // invalidate cache
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
