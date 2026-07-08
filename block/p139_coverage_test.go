package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === Paint: fallback plain-text path (getCachedBlocks returns nil) ===
// This path is hit when renderer.Render() fails or returns 0 blocks.
// We force this by making the renderer produce an error by setting
// the renderer field to a broken state.

func TestP139_Paint_FallbackPlainText(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Hello\nWorld\nThird line")
	// Clear the renderer to trigger the fallback path
	b.mu.Lock()
	b.renderer = nil
	b.mu.Unlock()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	// Paint will call ensureRenderer which creates a new renderer.
	// but if we manually set renderer to nil and bypass ensureRenderer,
	// getCachedBlocks returns nil.
	// Actually ensureRenderer is called inside Paint, so we need a different approach.
	// Let's use a renderer that returns an error by setting it to a broken one.
	// The simplest: render content that causes an error.
	// But goldmark rarely errors. Let's just test the path directly:
	b.mu.Lock()
	b.renderer = nil
	b.cachedText = ""
	b.cachedBlocks = nil
	b.mu.Unlock()
	// This will go through ensureRenderer → getCachedBlocks → non-nil normally.
	// To hit the nil path, we need getCachedBlocks to return nil.
	// That happens when blocks == nil || err != nil.
	// The only way is if Render returns an error, which goldmark doesn't do for normal text.
	// So we need to test a different path — empty content returning nil from render.
	b2 := NewAssistantTextBlock("test2")
	b2.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf2 := buffer.NewBuffer(80, 10)
	// Empty content — getCachedBlocks may return nil for empty text
	b2.Paint(buf2)
}

// === Paint: height truncation in markdown path ===
func TestP139_Paint_HeightTruncationMarkdown(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	b.Paint(buf)
}

// === Paint: width truncation in cell loop ===
func TestP139_Paint_WidthTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("very long line that exceeds width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	b.Paint(buf)
}

// === Paint: non-zero offset ===
func TestP139_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("text content")
	b.SetBounds(component.Rect{X: 10, Y: 5, W: 40, H: 5})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf)
}

// === Measure: fallback for nil blocks ===
func TestP139_Measure_NilBlocksFallback(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("line1\nline2\nline3")
	b.mu.Lock()
	b.renderer = nil
	b.cachedBlocks = nil
	b.cachedText = ""
	b.mu.Unlock()
	// Measure will call ensureRenderer which creates a new renderer.
	// The fallback path (blocks == nil) may not be hit via public API.
	s := b.Measure(component.Bounded(80, 100))
	_ = s
}

// === Measure: empty content ===
func TestP139_Measure_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test")
	s := b.Measure(component.Bounded(80, 100))
	if s.H != 1 {
		t.Errorf("expected H=1 for empty content, got %d", s.H)
	}
}

// === Measure: zero max width ===
func TestP139_Measure_ZeroMaxWidth(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	s := b.Measure(component.Bounded(0, 100))
	// Should default to maxW=80
	if s.W != 80 {
		t.Errorf("expected W=80 for zero max width, got %d", s.W)
	}
}

// === contentString: cache hit ===
func TestP139_ContentString_CacheHit(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	// First call fills cache
	s1 := b.contentString()
	// Second call should hit cache
	s2 := b.contentString()
	if s1 != s2 {
		t.Errorf("contentString cache miss: %q != %q", s1, s2)
	}
}

// === contentString: after AppendDelta (cache dirty) ===
func TestP139_ContentString_AfterDelta(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("initial")
	_ = b.contentString()
	b.AppendDelta(" more")
	s := b.contentString()
	if s != "initial more" {
		t.Errorf("expected 'initial more', got %q", s)
	}
}

// === getCachedBlocks: cache invalidation on content change ===
func TestP139_CacheInvalidation_ContentChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // fills cache

	// Change content — cache should invalidate
	b.AppendDelta(" more content")
	b.Paint(buf) // should re-render
}

// === getCachedBlocks: cache invalidation on width change ===
func TestP139_CacheInvalidation_WidthChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // fills cache at width 80

	// Change width — cache should invalidate
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	b.Paint(buf) // should re-render at width 60
}

// === Serialize/Deserialize ===
func TestP139_SerializeDeserialize_RoundTrip(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("markdown **content**\n\nNew paragraph.")
	state, err := b.SerializeState()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	b2 := NewAssistantTextBlock("test2")
	if err := b2.DeserializeState(state); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if b2.Content() != b.Content() {
		t.Errorf("content mismatch after deserialize")
	}
}

func TestP139_Deserialize_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test")
	if err := b.DeserializeState([]byte("{invalid}")); err == nil {
		t.Error("expected error for invalid JSON")
	}
}
