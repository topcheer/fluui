package block

import (
	"testing"
)

// ============================================================
// BlockFactory tests
// ============================================================

func TestFactoryNextID(t *testing.T) {
	f := NewBlockFactory()

	id1 := f.NextID("thinking")
	id2 := f.NextID("thinking")
	id3 := f.NextID("text")

	if id1 != "thinking-1" {
		t.Errorf("first ID: got %q, want thinking-1", id1)
	}
	if id2 != "thinking-2" {
		t.Errorf("second ID: got %q, want thinking-2", id2)
	}
	if id3 != "text-3" {
		t.Errorf("third ID: got %q, want text-3", id3)
	}
}

func TestFactoryNextIDConcurrency(t *testing.T) {
	f := NewBlockFactory()
	done := make(chan string, 100)

	for i := 0; i < 100; i++ {
		go func() {
			done <- f.NextID("block")
		}()
	}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := <-done
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
	}
}

func TestFactoryCreateThinkingBlock(t *testing.T) {
	f := NewBlockFactory()
	blk := f.CreateThinkingBlock()

	if blk.ID() == "" {
		t.Error("ID should not be empty")
	}
	if blk.Type() != TypeThinking {
		t.Errorf("type: got %s, want %s", blk.Type(), TypeThinking)
	}
	if blk.State() != BlockStreaming {
		t.Errorf("state: got %s, want streaming", blk.State())
	}
}

func TestFactoryCreateToolCallBlock(t *testing.T) {
	f := NewBlockFactory()
	blk := f.CreateToolCallBlock("grep", "{pattern: test}")

	if blk.Type() != TypeToolCall {
		t.Errorf("type: got %s, want %s", blk.Type(), TypeToolCall)
	}
	if blk.ToolName() != "grep" {
		t.Errorf("tool name: got %q, want grep", blk.ToolName())
	}
	if blk.RawArgs() != "{pattern: test}" {
		t.Errorf("raw args: got %q", blk.RawArgs())
	}
}

func TestFactoryCreateToolResultBlock(t *testing.T) {
	f := NewBlockFactory()
	blk := f.CreateToolResultBlock()

	if blk.Type() != TypeToolResult {
		t.Errorf("type: got %s, want %s", blk.Type(), TypeToolResult)
	}
	if blk.State() != BlockStreaming {
		t.Errorf("state: got %s, want streaming", blk.State())
	}
}

func TestFactoryCreateAssistantTextBlock(t *testing.T) {
	f := NewBlockFactory()
	blk := f.CreateAssistantTextBlock()

	if blk.Type() != TypeAssistantText {
		t.Errorf("type: got %s, want %s", blk.Type(), TypeAssistantText)
	}
	if blk.State() != BlockStreaming {
		t.Errorf("state: got %s, want streaming", blk.State())
	}
}

func TestFactoryCreateUserMessageBlock(t *testing.T) {
	f := NewBlockFactory()
	blk := f.CreateUserMessageBlock("Hello AI")

	if blk.Type() != TypeUserMessage {
		t.Errorf("type: got %s, want %s", blk.Type(), TypeUserMessage)
	}
	if blk.Content() != "Hello AI" {
		t.Errorf("content: got %q, want Hello AI", blk.Content())
	}
	if blk.State() != BlockComplete {
		t.Errorf("state: got %s, want complete", blk.State())
	}
}

// ============================================================
// StreamDispatcher tests
// ============================================================

func TestDispatchThinking(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	err := d.Dispatch(StreamDelta{Type: "thinking", Content: "I need to"})
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}
	err = d.Dispatch(StreamDelta{Type: "thinking", Content: " analyze this"})
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if container.Len() != 1 {
		t.Fatalf("expected 1 block, got %d", container.Len())
	}

	blk := container.Blocks()[0]
	if blk.Type() != TypeThinking {
		t.Errorf("type: got %s, want thinking", blk.Type())
	}
	if blk.State() != BlockStreaming {
		t.Errorf("state: got %s, want streaming", blk.State())
	}

	thinking := blk.(*ThinkingBlock)
	if thinking.Content() != "I need to analyze this" {
		t.Errorf("content: got %q", thinking.Content())
	}
}

func TestDispatchText(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "text", Content: "Hello "})
	d.Dispatch(StreamDelta{Type: "text", Content: "world!"})

	if container.Len() != 1 {
		t.Fatalf("expected 1 block, got %d", container.Len())
	}

	blk := container.Blocks()[0]
	if blk.Type() != TypeAssistantText {
		t.Errorf("type: got %s, want assistant_text", blk.Type())
	}

	text := blk.(*AssistantTextBlock)
	if text.Content() != "Hello world!" {
		t.Errorf("content: got %q", text.Content())
	}
}

func TestDispatchToolCall(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{
		Type:     "tool_call",
		ToolName: "read_file",
		ToolArgs: `{"path": "/tmp/test.go"}`,
	})

	if container.Len() != 1 {
		t.Fatalf("expected 1 block, got %d", container.Len())
	}

	blk := container.Blocks()[0]
	if blk.Type() != TypeToolCall {
		t.Errorf("type: got %s, want tool_call", blk.Type())
	}

	tc := blk.(*ToolCallBlock)
	if tc.ToolName() != "read_file" {
		t.Errorf("tool name: got %q", tc.ToolName())
	}
	if tc.RawArgs() != `{"path": "/tmp/test.go"}` {
		t.Errorf("raw args: got %q", tc.RawArgs())
	}
}

func TestDispatchToolResult(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "tool_result", Content: "line 1\n"})
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "line 2"})

	if container.Len() != 1 {
		t.Fatalf("expected 1 block, got %d", container.Len())
	}

	blk := container.Blocks()[0]
	if blk.Type() != TypeToolResult {
		t.Errorf("type: got %s, want tool_result", blk.Type())
	}

	tr := blk.(*ToolResultBlock)
	if tr.Output() != "line 1\nline 2" {
		t.Errorf("output: got %q", tr.Output())
	}
}

func TestDispatchMultipleTypes(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Thinking
	d.Dispatch(StreamDelta{Type: "thinking", Content: "Let me think..."})
	// Text response
	d.Dispatch(StreamDelta{Type: "text", Content: "Here is the answer."})
	// Tool call
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "ls", ToolArgs: "-la"})
	// Tool result
	d.Dispatch(StreamDelta{Type: "tool_result", Content: "file1.txt"})

	if container.Len() != 4 {
		t.Fatalf("expected 4 blocks, got %d", container.Len())
	}

	types := []BlockType{TypeThinking, TypeAssistantText, TypeToolCall, TypeToolResult}
	for i, want := range types {
		if container.Blocks()[i].Type() != want {
			t.Errorf("block %d type: got %s, want %s", i, container.Blocks()[i].Type(), want)
		}
	}
}

func TestDispatchSameTypeReusesBlock(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Multiple thinking deltas should reuse the same block.
	d.Dispatch(StreamDelta{Type: "thinking", Content: "part 1 "})
	d.Dispatch(StreamDelta{Type: "thinking", Content: "part 2 "})
	d.Dispatch(StreamDelta{Type: "thinking", Content: "part 3"})

	if container.Len() != 1 {
		t.Errorf("expected 1 block for 3 thinking deltas, got %d", container.Len())
	}
}

func TestDispatchMultipleToolCallsCreateSeparateBlocks(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "cmd1", ToolArgs: ""})
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "cmd2", ToolArgs: ""})
	d.Dispatch(StreamDelta{Type: "tool_call", ToolName: "cmd3", ToolArgs: ""})

	if container.Len() != 3 {
		t.Errorf("expected 3 tool call blocks, got %d", container.Len())
	}
}

func TestDispatchUnknownType(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	err := d.Dispatch(StreamDelta{Type: "unknown_type"})
	if err == nil {
		t.Error("expected error for unknown delta type")
	}
}

func TestDispatchError(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Create an active block first.
	d.Dispatch(StreamDelta{Type: "text", Content: "partial..."})

	// Send an error delta.
	d.Dispatch(StreamDelta{Type: "error", Content: "connection lost"})

	// The active text block should be in error state.
	blk := container.Blocks()[0]
	if blk.State() != BlockError {
		t.Errorf("expected error state, got %s", blk.State())
	}
	// Error() is on BaseBlock, accessible via concrete type.
	tb := blk.(*AssistantTextBlock)
	if tb.Error() != "connection lost" {
		t.Errorf("error msg: got %q", tb.Error())
	}
}

// ============================================================
// Flush tests
// ============================================================

func TestFlush(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "thinking", Content: "done thinking"})
	d.Dispatch(StreamDelta{Type: "text", Content: "done text"})

	// Verify blocks are streaming.
	for _, b := range container.Blocks() {
		if b.State() != BlockStreaming {
			t.Errorf("pre-flush state: got %s, want streaming", b.State())
		}
	}

	d.Flush()

	// All blocks should be completed.
	for _, b := range container.Blocks() {
		if b.State() != BlockComplete {
			t.Errorf("post-flush state: got %s, want complete", b.State())
		}
	}

	// Active blocks should be cleared.
	active := d.ActiveBlocks()
	if len(active) != 0 {
		t.Errorf("expected 0 active blocks after flush, got %d", len(active))
	}
}

func TestFlushEmpty(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// Flush with no active blocks should not panic.
	d.Flush()
	if container.Len() != 0 {
		t.Errorf("expected 0 blocks, got %d", container.Len())
	}
}

func TestFlushIdempotent(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	d.Dispatch(StreamDelta{Type: "text", Content: "hello"})
	d.Flush()
	d.Flush() // second flush should be safe

	if container.Len() != 1 {
		t.Errorf("expected 1 block, got %d", container.Len())
	}
	if container.Blocks()[0].State() != BlockComplete {
		t.Error("block should be complete after flush")
	}
}

// ============================================================
// ActiveBlocks tests
// ============================================================

func TestActiveBlocks(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	if len(d.ActiveBlocks()) != 0 {
		t.Error("expected 0 active blocks initially")
	}

	d.Dispatch(StreamDelta{Type: "thinking", Content: "thinking..."})
	if len(d.ActiveBlocks()) != 1 {
		t.Errorf("expected 1 active block, got %d", len(d.ActiveBlocks()))
	}

	d.Dispatch(StreamDelta{Type: "text", Content: "text..."})
	if len(d.ActiveBlocks()) != 2 {
		t.Errorf("expected 2 active blocks, got %d", len(d.ActiveBlocks()))
	}

	d.Flush()
	if len(d.ActiveBlocks()) != 0 {
		t.Errorf("expected 0 active blocks after flush, got %d", len(d.ActiveBlocks()))
	}
}

// ============================================================
// Integration: full streaming session
// ============================================================

func TestStreamDispatcherFullSession(t *testing.T) {
	container := NewBlockContainer()
	d := NewStreamDispatcher(container)

	// User asks a question
	userBlk := d.Factory().CreateUserMessageBlock("What is 2+2?")
	container.AddBlock(userBlk)

	// AI thinks
	d.Dispatch(StreamDelta{Type: "thinking", Content: "2+2 = "})
	d.Dispatch(StreamDelta{Type: "thinking", Content: "4"})

	// AI responds
	d.Dispatch(StreamDelta{Type: "text", Content: "The answer is 4."})

	// Flush (stream complete)
	d.Flush()

	// Verify
	if container.Len() != 3 { // user + thinking + text
		t.Fatalf("expected 3 blocks, got %d", container.Len())
	}

	// User message should be complete.
	if container.Blocks()[0].State() != BlockComplete {
		t.Error("user block should be complete")
	}

	// All other blocks should be complete after flush.
	for i := 1; i < container.Len(); i++ {
		b := container.Blocks()[i]
		if b.State() != BlockComplete {
			t.Errorf("block %d should be complete, got %s", i, b.State())
		}
	}
}
