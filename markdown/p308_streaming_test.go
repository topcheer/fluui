package markdown

import (
	"strings"
	"testing"
)

func TestP308_NewStreamingRenderer(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 4)
	if sr == nil {
		t.Fatal("nil")
	}
	if !sr.IsDirty() {
		t.Error("should start dirty")
	}
}

func TestP308_AppendDelta_Debounce(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 3)
	sr.AppendDelta("# Hello")
	// 1 delta < threshold(3), still dirty
	if !sr.IsDirty() {
		t.Error("should be dirty before threshold")
	}
	sr.AppendDelta(" World")
	sr.AppendDelta("!")
	// 3 deltas = threshold → renders
	if sr.IsDirty() {
		t.Error("should not be dirty after threshold render")
	}
}

func TestP308_Blocks_ForceRender(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 100) // high threshold
	sr.AppendDelta("# Title")
	blocks, err := sr.Blocks() // Blocks() forces render
	if err != nil {
		t.Fatalf("Blocks error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected at least 1 block")
	}
}

func TestP308_Flush(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 100)
	sr.AppendDelta("# Title\n\nSome text")
	blocks, err := sr.Flush()
	if err != nil {
		t.Fatalf("Flush error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected blocks after flush")
	}
}

func TestP308_Source(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 10)
	sr.AppendDelta("hello ")
	sr.AppendDelta("world")
	if sr.Source() != "hello world" {
		t.Errorf("Source = %q", sr.Source())
	}
}

func TestP308_SetSource(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 10)
	sr.AppendDelta("old")
	sr.SetSource("# New Content")
	if sr.Source() != "# New Content" {
		t.Errorf("Source = %q", sr.Source())
	}
}

func TestP308_Reset(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 1)
	sr.AppendDelta("content")
	sr.Reset()
	if sr.Source() != "" {
		t.Errorf("Source after reset = %q", sr.Source())
	}
	if !sr.IsDirty() {
		t.Error("should be dirty after reset")
	}
}

func TestP308_LineCount(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 1)
	sr.AppendDelta("# Title\n\nParagraph")
	lc := sr.LineCount()
	if lc == 0 {
		t.Error("expected non-zero line count")
	}
}

func TestP308_SetWidth(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 10)
	sr.AppendDelta("text")
	sr.SetWidth(40)
	// Should not panic, should re-render
	blocks, _ := sr.Blocks()
	if blocks == nil {
		t.Error("expected non-nil blocks after SetWidth")
	}
}

func TestP308_StreamingFullDocument(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	sr := NewStreamingRenderer(r, 2)

	tokens := strings.Fields("# AI Response\n\nThis is **bold** text with `code`.\n\n- Item 1\n- Item 2\n\n```go\nfunc main() {}\n```")
	for _, tk := range tokens {
		sr.AppendDelta(tk + " ")
	}
	sr.Flush()

	if sr.Source() == "" {
		t.Error("source should not be empty")
	}
	lc := sr.LineCount()
	if lc < 2 {
		t.Errorf("expected >= 2 lines, got %d", lc)
	}
}

func TestP308_ThreadSafe(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 1)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			sr.AppendDelta("x")
		}
		close(done)
	}()

	for i := 0; i < 50; i++ {
		_, _ = sr.Blocks()
		_ = sr.Source()
	}
	<-done
}

func TestP308_DefaultThreshold(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	sr := NewStreamingRenderer(r, 0) // should default to 4
	sr.mu.Lock()
	if sr.threshold != 4 {
		t.Errorf("threshold = %d, want 4", sr.threshold)
	}
	sr.mu.Unlock()
}
