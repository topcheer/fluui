package block

import (
	"encoding/json"
	"strings"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/theme"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
)

// AssistantTextBlock displays streaming markdown text from the AI assistant.
// It renders headings, lists, code blocks, bold/italic, links, and inline code
// with proper syntax highlighting and Dracula theme colors.
type AssistantTextBlock struct {
	BaseBlock
	content     strings.Builder
	renderer    *markdown.MarkdownRenderer
	highlighter *markdown.Highlighter
	renderW     int // width the renderer was created for

	// P24-D render cache: avoid re-parsing markdown on every Paint/Measure call.
	// The cache is invalidated when content changes or render width changes.
	cachedBlocks []*markdown.Block
	cachedText   string
	cachedW      int
}

// NewAssistantTextBlock creates an assistant text block in streaming state.
func NewAssistantTextBlock(id string) *AssistantTextBlock {
	return &AssistantTextBlock{
		BaseBlock: NewBaseBlock(id, TypeAssistantText),
	}
}

// ensureRenderer creates or recreates the markdown renderer for the current width.
func (b *AssistantTextBlock) ensureRenderer(width int) {
	if b.renderer != nil && b.renderW == width {
		return
	}
	b.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), width)
	if b.highlighter == nil {
		b.highlighter = markdown.NewHighlighter()
	}
	b.renderer.SetHighlighter(b.highlighter)
	b.renderW = width
}

// AppendDelta appends streaming text.
func (b *AssistantTextBlock) AppendDelta(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.WriteString(delta)
	b.cachedText = "" // invalidate cache
	b.markDirtyLocked()
}

// getCachedBlocks returns the rendered markdown blocks, using a cache to avoid
// re-parsing when neither the text nor the width has changed. This dramatically
// reduces allocations in Container.Paint (100+ blocks) where Paint/Measure are
// called repeatedly without content changes.
func (b *AssistantTextBlock) getCachedBlocks(text string, width int) []*markdown.Block {
	if len(b.cachedBlocks) > 0 && b.cachedText == text && b.cachedW == width {
		return b.cachedBlocks
	}
	blocks, err := b.renderer.Render(text)
	if err != nil || len(blocks) == 0 {
		b.cachedBlocks = nil
		b.cachedText = ""
		b.cachedW = 0
		return nil
	}
	b.cachedBlocks = blocks
	b.cachedText = text
	b.cachedW = width
	return blocks
}

// Content returns the full text so far.
func (b *AssistantTextBlock) Content() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.content.String()
}

// Measure returns the size based on rendered markdown lines.
func (b *AssistantTextBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	text := b.content.String()
	if text == "" {
		return component.Size{W: maxW, H: 1}
	}

	b.ensureRenderer(maxW)
	blocks := b.getCachedBlocks(text, maxW)
	if blocks == nil {
		// Fallback: plain text line count
		lines := strings.Count(text, "\n") + 1
		return component.Size{W: maxW, H: lines}
	}

	totalLines := 0
	for _, blk := range blocks {
		totalLines += len(blk.Cells)
	}
	if totalLines == 0 {
		totalLines = 1
	}
	return component.Size{W: maxW, H: totalLines}
}

// SetBounds sets the bounds.
func (b *AssistantTextBlock) SetBounds(r component.Rect) {
	b.BaseComponent.SetBounds(r)
}

// SerializeState serializes the assistant text block's content to JSON.
func (b *AssistantTextBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.Marshal(map[string]any{
		"content": b.content.String(),
	})
}

// DeserializeState restores the assistant text block's content from JSON.
func (b *AssistantTextBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.Reset()
	b.content.WriteString(s.Content)
	b.markDirtyLocked()
	return nil
}

// Paint renders the assistant text as styled markdown.
func (b *AssistantTextBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}
	text := b.content.String()
	if text == "" {
		return
	}

	// Render markdown to blocks of styled cell lines (cached)
	b.ensureRenderer(bounds.W)
	mdBlocks := b.getCachedBlocks(text, bounds.W)
	if mdBlocks == nil {
		// Fallback: plain text render
		style := buffer.Style{Fg: theme.Get().AssistantFg}
		rowIdx := 0
		for _, rawLine := range strings.Split(text, "\n") {
			for _, wrappedLine := range markdown.WrapText(rawLine, bounds.W) {
				y := bounds.Y + rowIdx
				if y >= bounds.Y+bounds.H {
					return
				}
				buf.DrawText(bounds.X, y, wrappedLine, style)
				rowIdx++
			}
		}
		return
	}

	// Draw each markdown block's cell lines
	rowIdx := 0
	for _, blk := range mdBlocks {
		for _, cellLine := range blk.Cells {
			y := bounds.Y + rowIdx
			if y >= bounds.Y+bounds.H {
				return
			}
			x := bounds.X
			for _, cell := range cellLine {
				if x >= bounds.X+bounds.W {
					break
				}
				buf.SetCell(x, y, cell)
				x += cell.Width
				if cell.Width == 0 {
					x++
				}
			}
			rowIdx++
		}
	}
}
