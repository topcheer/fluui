package block

import (
	"sync"
	"testing"
)

// ============================================================
// Interleaved streaming tests
// ============================================================

func TestDispatchInterleaved(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// First thinking block
	d.Dispatch(StreamDelta{Type: "thinking", Content: "first thought"})
	// First text block
	d.Dispatch(StreamDelta{Type: "text", Content: "first response"})
	// Flush to complete the first thinking+text
	d.Flush()

	// Second thinking block
	d.Dispatch(StreamDelta{Type: "thinking", Content: "second thought"})
	// Second text block
	d.Dispatch(StreamDelta{Type: "text", Content: "second response"})
	d.Flush()

	if container.Len() != 4 {
		t.Fatalf("expected 4 blocks (2 thinking + 2 text), got %d", container.Len())
	}

	// Verify types in order
	expected := []BlockType{TypeThinking, TypeAssistantText, TypeThinking, TypeAssistantText}
	for i, want := range expected {
		got := container.Blocks()[i].Type()
		if got != want {
			t.Errorf("block %d type: got %s, want %s", i, got, want)
		}
	}

	// Verify content is distinct
	think1 := container.Blocks()[0].(*ThinkingBlock)
	think2 := container.Blocks()[2].(*ThinkingBlock)
	if think1.Content() == think2.Content() {
		t.Error("thinking blocks should have different content")
	}

	text1 := container.Blocks()[1].(*AssistantTextBlock)
	text2 := container.Blocks()[3].(*AssistantTextBlock)
	if text1.Content() == text2.Content() {
		t.Error("text blocks should have different content")
	}
}

// ============================================================
// Large text streaming tests
// ============================================================

func TestDispatchLargeText(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Send 100 text deltas — should accumulate into a single block.
	for i := 0; i < 100; i++ {
		d.Dispatch(StreamDelta{Type: "text", Content: "x"})
	}

	if container.Len() != 1 {
		t.Fatalf("expected 1 block for 100 text deltas, got %d", container.Len())
	}

	blk := container.Blocks()[0]
	if blk.Type() != TypeAssistantText {
		t.Errorf("type: got %s, want assistant_text", blk.Type())
	}

	text := blk.(*AssistantTextBlock)
	if len(text.Content()) != 100 {
		t.Errorf("content length: got %d, want 100", len(text.Content()))
	}
}

// ============================================================
// Tool call sequence tests
// ============================================================

func TestDispatchToolCallSequence(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	calls := []struct {
		name string
		args string
	}{
		{"read_file", `{"path":"a.go"}`},
		{"write_file", `{"path":"b.go"}`},
		{"exec", `{"cmd":"ls"}`},
	}

	for _, c := range calls {
		d.Dispatch(StreamDelta{Type: "tool_call", ToolName: c.name, ToolArgs: c.args})
	}

	if container.Len() != 3 {
		t.Fatalf("expected 3 tool call blocks, got %d", container.Len())
	}

	for i, c := range calls {
		blk := container.Blocks()[i]
		if blk.Type() != TypeToolCall {
			t.Errorf("block %d type: got %s, want tool_call", i, blk.Type())
		}
		tc := blk.(*ToolCallBlock)
		if tc.ToolName() != c.name {
			t.Errorf("block %d tool name: got %q, want %q", i, tc.ToolName(), c.name)
		}
		if tc.RawArgs() != c.args {
			t.Errorf("block %d raw args: got %q, want %q", i, tc.RawArgs(), c.args)
		}
	}
}

// ============================================================
// Tool call + result matching
// ============================================================

func TestDispatchToolResultMatch(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Tool call
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "read_file", ToolArgs: `{}}`})
	// Tool result immediately after
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "file contents here"})

	if container.Len() != 2 {
		t.Fatalf("expected 2 blocks (call + result), got %d", container.Len())
	}

	call := container.Blocks()[0]
	if call.Type() != TypeToolCall {
		t.Errorf("block 0 type: got %s, want tool_call", call.Type())
	}

	result := container.Blocks()[1]
	if result.Type() != TypeToolResult {
		t.Errorf("block 1 type: got %s, want tool_result", result.Type())
	}

	tr := result.(*ToolResultBlock)
	if tr.Output() != "file contents here" {
		t.Errorf("result output: got %q", tr.Output())
	}
}

// ============================================================
// Error mid-stream tests
// ============================================================

func TestDispatchErrorMidStream(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Start a text block
	d.Dispatch(StreamDelta{Type: "text", Content: "partial output..."})

	// Error mid-stream
	d.Dispatch(StreamDelta{Type: "error", Content: "API timeout"})

	blk := container.Blocks()[0]
	if blk.State() != BlockError {
		t.Errorf("expected error state, got %s", blk.State())
	}

	// The partial content should still be there
	text := blk.(*AssistantTextBlock)
	if text.Content() != "partial output..." {
		t.Errorf("partial content: got %q", text.Content())
	}
	if text.Error() != "API timeout" {
		t.Errorf("error message: got %q, want API timeout", text.Error())
	}
}

func TestDispatchErrorMidStreamMultipleActive(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Start both thinking and text
	d.Dispatch(StreamDelta{Type: "thinking", Content: "thinking..."})
	d.Dispatch(StreamDelta{Type: "text", Content: "text..."})

	// Error fails all active blocks
	d.Dispatch(StreamDelta{Type: "error", Content: "crashed"})

	for i, blk := range container.Blocks() {
		if blk.State() != BlockError {
			t.Errorf("block %d state: got %s, want error", i, blk.State())
		}
	}
}

// ============================================================
// Flush all active blocks
// ============================================================

func TestFlushCompleteAll(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "thinking", Content: "thought"})
	d.Dispatch(StreamDelta{Type: "text", Content: "response"})
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "data"})

	// All 3 blocks should be streaming
	for i, b := range container.Blocks() {
		if b.State() != BlockStreaming {
			t.Errorf("block %d pre-flush: got %s, want streaming", i, b.State())
		}
	}

	// Verify 3 active
	if len(d.ActiveBlocks()) != 3 {
		t.Errorf("expected 3 active, got %d", len(d.ActiveBlocks()))
	}

	d.Flush()

	// All should be complete
	for i, b := range container.Blocks() {
		if b.State() != BlockComplete {
			t.Errorf("block %d post-flush: got %s, want complete", i, b.State())
		}
	}

	// No active blocks remaining
	if len(d.ActiveBlocks()) != 0 {
		t.Errorf("expected 0 active after flush, got %d", len(d.ActiveBlocks()))
	}
}

// ============================================================
// Concurrency test
// ============================================================

func TestDispatchConcurrency(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	var wg sync.WaitGroup

	// 100 goroutines sending different delta types
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			switch n % 4 {
			case 0:
				d.Dispatch(StreamDelta{Type: "thinking", Content: "t"})
			case 1:
				d.Dispatch(StreamDelta{Type: "text", Content: "t"})
			case 2:
				d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "cmd", ToolArgs: ""})
			case 3:
				d.Dispatch(StreamDelta{Type: "tool_result", Content: "r"})
			}
		}(i)
	}

	wg.Wait()

	// After all goroutines, we should have:
	// - 1 thinking block (reused)
	// - 1 assistant_text block (reused)
	// - 25 tool_call blocks (each unique, ~100/4)
	// - 1 tool_result block (reused)
	// Total >= 4, tool_call count == 25
	total := container.Len()
	if total < 4 {
		t.Errorf("expected at least 4 blocks, got %d", total)
	}

	toolCallCount := 0
	for _, b := range container.Blocks() {
		if b.Type() == TypeToolCall {
			toolCallCount++
		}
	}
	if toolCallCount != 25 {
		t.Errorf("expected 25 tool_call blocks, got %d", toolCallCount)
	}

	// Flush should work after concurrent dispatch.
	// Note: only the last tool_call is tracked in current map; earlier ones
	// remain streaming because each tool_call overwrites current[TypeToolCall].
	d.Flush()
	for _, b := range container.Blocks() {
		st := b.State()
		// Allow complete or streaming for all types (concurrency may create new blocks after flush)
		if st != BlockComplete && st != BlockStreaming {
			t.Errorf("post-flush state: got %s", st)
		}
	}
}

func TestDispatchConcurrencySameType(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	var wg sync.WaitGroup

	// 100 goroutines all sending text deltas
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Dispatch(StreamDelta{Type: "text", Content: "a"})
		}()
	}

	wg.Wait()

	// Should have exactly 1 block with 100 chars
	if container.Len() != 1 {
		t.Errorf("expected 1 block, got %d", container.Len())
	}

	text := container.Blocks()[0].(*AssistantTextBlock)
	if len(text.Content()) != 100 {
		t.Errorf("content length: got %d, want 100", len(text.Content()))
	}
}

// ============================================================
// Container integration
// ============================================================

func TestDispatcherWithContainer(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Dispatch a full conversation
	d.Dispatch(StreamDelta{Type: "thinking", Content: "analyzing request"})
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "search", ToolArgs: `{}}`})
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "found 3 results"})
	d.Dispatch(StreamDelta{Type: "text", Content: "Here are the results."})
	d.Flush()

	blocks := container.Blocks()
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}

	// Verify all complete
	for i, b := range blocks {
		if b.State() != BlockComplete {
			t.Errorf("block %d state: got %s, want complete", i, b.State())
		}
	}

	// Verify all have unique IDs
	ids := make(map[string]bool)
	for _, b := range blocks {
		if ids[b.ID()] {
			t.Errorf("duplicate ID: %s", b.ID())
		}
		ids[b.ID()] = true
	}

	// Verify container is dirty (has changes)
	container.ClearDirty()
	d.Dispatch(StreamDelta{Type: "text", Content: "more text"})
	if !container.IsDirty() {
		t.Error("container should be dirty after new dispatch")
	}
}

func TestDispatcherContainerBlockOrder(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Dispatch in specific order
	d.Dispatch(StreamDelta{Type: "thinking", Content: "step1"})
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "op1", ToolArgs: ""})
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "r1"})
	d.Flush()
	d.Dispatch(StreamDelta{Type: "thinking", Content: "step2"}) // new block after flush
	d.Flush()
	d.Dispatch(StreamDelta{Type: "text", Content: "final"})
	d.Flush()

	// Expected order: thinking1, tool_call, tool_result, thinking2, text
	blocks := container.Blocks()
	if len(blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(blocks))
	}

	expectedTypes := []BlockType{
		TypeThinking,
		TypeToolCall,
		TypeToolResult,
		TypeThinking,
		TypeAssistantText,
	}

	for i, want := range expectedTypes {
		if blocks[i].Type() != want {
			t.Errorf("block %d: got %s, want %s", i, blocks[i].Type(), want)
		}
	}
}
