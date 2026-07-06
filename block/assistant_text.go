package block

import (
	"encoding/json"
	"strings"
	"sync"

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

	// P52 content string cache: avoid repeated String() allocations.
	cachedContentStr string
	contentDirty     bool
}

// NewAssistantTextBlock creates an assistant text block in streaming state.
func NewAssistantTextBlock(id string) *AssistantTextBlock {
	b := &AssistantTextBlock{
		BaseBlock: NewBaseBlock(id, TypeAssistantText),
	}
	// Pre-grow to eliminate ~5 buffer growth allocations during typical streaming.
	// 256 bytes covers ~50 words of average length, the common streaming case.
	b.content.Grow(256)
	return b
}

// rendererCache caches MarkdownRenderers by width to avoid creating a new
// goldmark parser for every AssistantTextBlock. The goldmark parser allocates
// ~100 objects on creation (block parsers, inline parsers, AST transformers,
// table extension, etc.), so sharing one renderer across all blocks of the
// same width eliminates ~100K allocations for 1000 blocks.
var rendererCache sync.Map // int (width) → *markdown.MarkdownRenderer

// sharedHighlighter is a package-level highlighter shared across all blocks.
// The Highlighter is stateless (just a style reference), so sharing is safe.
var (
	highlighterOnce sync.Once
	sharedHighlighter *markdown.Highlighter
)

func getSharedHighlighter() *markdown.Highlighter {
	highlighterOnce.Do(func() {
		sharedHighlighter = markdown.NewHighlighter()
	})
	return sharedHighlighter
}

// ensureRenderer creates or reuses a markdown renderer for the current width.
func (b *AssistantTextBlock) ensureRenderer(width int) {
	if b.renderer != nil && b.renderW == width {
		return
	}
	// Try to reuse a cached renderer for this width
	if v, ok := rendererCache.Load(width); ok {
		b.renderer = v.(*markdown.MarkdownRenderer)
		b.highlighter = getSharedHighlighter()
		b.renderW = width
		return
	}
	// Create new renderer and cache it for future blocks
	r := markdown.NewMarkdownRenderer(markdown.DefaultTheme(), width)
	h := getSharedHighlighter()
	r.SetHighlighter(h)
	rendererCache.Store(width, r)
	b.renderer = r
	b.highlighter = h
	b.renderW = width
}

// AppendDelta appends streaming text.
func (b *AssistantTextBlock) AppendDelta(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.WriteString(delta)
	b.cachedText = ""     // invalidate render cache
	b.contentDirty = true // invalidate string cache
	b.markDirtyLocked()
}

// getCachedBlocks returns the rendered markdown blocks, using a cache to avoid
// re-parsing when neither the text nor the width has changed. This dramatically
// reduces allocations in Container.Paint (100+ blocks) where Paint/Measure are
// called repeatedly without content changes.
//
// Callers must hold at least RLock. The cache write is safe because
// AppendDelta (the only content mutation) takes a full Lock and clears
// cachedText, so under RLock the cache fields are stable.
// getCachedBlocks returns cached markdown blocks, rendering if needed.
// Must be called with at least RLock held. The cache write is safe because
// AppendDelta/SetContent (which invalidate the cache) always take a full Lock,
// so they can't run concurrently with Measure/Paint's RLock.
// The only race is between concurrent Measure+Paint calls, but both write the
// SAME values (same text, same width → same blocks), making the write idempotent.
// Go's race detector doesn't flag writes under RLock when other concurrent
// readers also write the same value, but to be fully safe we use a mutex
// around the cache fill.
var cacheFillMu sync.Mutex

func (b *AssistantTextBlock) getCachedBlocks(text string, width int) []*markdown.Block {
	// Fast path: cache hit
	if b.cachedText == text && b.cachedW == width && b.cachedBlocks != nil {
		return b.cachedBlocks
	}
	// Cache miss: serialize cache fills to avoid races between concurrent
	// Measure/Paint calls. The actual markdown rendering happens outside
	// the block's mutex, so it doesn't block other blocks.
	cacheFillMu.Lock()
	defer cacheFillMu.Unlock()
	// Re-check after acquiring fill lock (another goroutine may have filled)
	if b.cachedText == text && b.cachedW == width && b.cachedBlocks != nil {
		return b.cachedBlocks
	}
	blocks, err := b.renderer.Render(text)
	if err != nil || len(blocks) == 0 {
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

// contentString returns the cached content string, avoiding repeated
// strings.Builder.String() allocations when content hasn't changed.
// Caller must hold at least RLock. Uses cacheFillMu for write safety.
func (b *AssistantTextBlock) contentString() string {
	if !b.contentDirty && b.cachedContentStr != "" {
		return b.cachedContentStr
	}
	cacheFillMu.Lock()
	defer cacheFillMu.Unlock()
	if !b.contentDirty && b.cachedContentStr != "" {
		return b.cachedContentStr
	}
	b.cachedContentStr = b.content.String()
	b.contentDirty = false
	return b.cachedContentStr
}

// SetContent replaces the full text and marks the block dirty.
// This invalidates the render cache.
func (b *AssistantTextBlock) SetContent(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.Reset()
	b.content.WriteString(s)
	b.cachedText = ""   // invalidate render cache
	b.contentDirty = true // invalidate string cache
	b.markDirtyLocked()
}

// Measure returns the size based on rendered markdown lines.
func (b *AssistantTextBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	text := b.contentString()
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
		"content": b.contentString(),
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
	b.cachedText = ""     // invalidate render cache
	b.contentDirty = true // invalidate string cache
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
	text := b.contentString()
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
