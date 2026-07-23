package markdown

import (
	"strings"
	"sync"
)

// StreamingRenderer wraps a MarkdownRenderer to provide incremental
// markdown rendering for AI streaming responses. It buffers incoming
// text deltas and re-renders at debounced intervals, caching the last
// render result to avoid redundant parsing.
//
// Thread-safe.
type StreamingRenderer struct {
	mu       sync.Mutex
	renderer *MarkdownRenderer

	source   strings.Builder
	cached   []*Block
	cacheErr error
	dirty    bool

	deltaCount int
	threshold  int // re-render every N deltas
}

// NewStreamingRenderer creates a streaming renderer wrapping the given
// MarkdownRenderer. The threshold controls how many AppendDelta calls
// are buffered before a re-render (default 4).
func NewStreamingRenderer(r *MarkdownRenderer, threshold int) *StreamingRenderer {
	if threshold < 1 {
		threshold = 4
	}
	return &StreamingRenderer{
		renderer:  r,
		threshold: threshold,
		dirty:     true,
	}
}

// AppendDelta appends a text chunk and marks the cache dirty.
// Re-rendering is debounced based on the threshold.
func (s *StreamingRenderer) AppendDelta(delta string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source.WriteString(delta)
	s.deltaCount++
	s.dirty = true

	// Re-render if threshold reached
	if s.deltaCount%s.threshold == 0 {
		s.renderLocked()
	}
}

// SetSource replaces the entire source text and forces a re-render.
func (s *StreamingRenderer) SetSource(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source.Reset()
	s.source.WriteString(text)
	s.deltaCount = 0
	s.dirty = true
	s.renderLocked()
}

// Blocks returns the current rendered blocks. If the cache is dirty,
// a re-render is triggered first.
func (s *StreamingRenderer) Blocks() ([]*Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dirty {
		s.renderLocked()
	}
	return s.cached, s.cacheErr
}

// Flush forces an immediate re-render regardless of debounce.
func (s *StreamingRenderer) Flush() ([]*Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderLocked()
	return s.cached, s.cacheErr
}

// Source returns the full accumulated source text.
func (s *StreamingRenderer) Source() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.source.String()
}

// Reset clears all content.
func (s *StreamingRenderer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source.Reset()
	s.cached = nil
	s.cacheErr = nil
	s.dirty = true
	s.deltaCount = 0
}

// LineCount returns the total number of rendered lines in the cache.
// Returns 0 if no rendering has occurred.
func (s *StreamingRenderer) LineCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dirty {
		s.renderLocked()
	}
	total := 0
	for _, b := range s.cached {
		total += len(b.Cells)
	}
	return total
}

// IsDirty returns whether the cache needs re-rendering.
func (s *StreamingRenderer) IsDirty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dirty
}

// SetWidth changes the rendering width and forces a re-render.
func (s *StreamingRenderer) SetWidth(w int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderer.width = w
	s.dirty = true
	s.renderLocked()
}

// renderLocked re-parses and renders the full source (caller must hold lock).
func (s *StreamingRenderer) renderLocked() {
	blocks, err := s.renderer.Render(s.source.String())
	s.cached = blocks
	s.cacheErr = err
	s.dirty = false
}
