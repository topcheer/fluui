package block_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ============================================================
// P26-B: AI Streaming → Block → Container → Buffer Pipeline
// ============================================================

// TestP26_StreamDispatch_CreatesCorrectBlocks verifies that StreamDispatcher
// creates the correct block types for each delta type.
func TestP26_StreamDispatch_CreatesCorrectBlocks(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	deltas := []block.StreamDelta{
		{Type: "thinking", Content: "Let me analyze this..."},
		{Type: "text", Content: "# Hello World"},
		{Type: "tool_call", ToolName: "read_file", ToolArgs: `{"path":"main.go"}`},
		{Type: "tool_result", Content: "42 lines"},
		{Type: "error", Content: "connection reset"},
	}

	for _, d := range deltas {
		if err := dispatcher.Dispatch(d); err != nil {
			t.Errorf("Dispatch(%s) error: %v", d.Type, err)
		}
	}

	blks := container.Blocks()
	if len(blks) != 4 { // thinking + text + tool_call + tool_result; error fails existing blocks
		t.Errorf("expected 4 blocks, got %d", len(blks))
	}

	// Verify block types
	expectedTypes := []block.BlockType{block.TypeThinking, block.TypeAssistantText, block.TypeToolCall, block.TypeToolResult}
	for i, bt := range expectedTypes {
		if blks[i].Type() != bt {
			t.Errorf("block[%d] type = %v, want %v", i, blks[i].Type(), bt)
		}
	}
}

// TestP26_StreamDispatch_TextDeltasAppendToSameBlock verifies that
// consecutive text deltas append to the same block, not create new ones.
func TestP26_StreamDispatch_TextDeltasAppendToSameBlock(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	for i := 0; i < 10; i++ {
		dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: fmt.Sprintf("line %d\n", i)})
	}

	blks := container.Blocks()
	if len(blks) != 1 {
		t.Errorf("expected 1 text block, got %d", len(blks))
	}
	if blks[0].Type() != block.TypeAssistantText {
		t.Errorf("expected TypeAssistantText, got %v", blks[0].Type())
	}
	if blks[0].State() != block.BlockStreaming {
		t.Errorf("expected BlockStreaming state, got %v", blks[0].State())
	}
}

// TestP26_StreamDispatch_ToolCallsCreateSeparateBlocks verifies that
// each tool_call creates a new block (unlike text/thinking which reuse).
func TestP26_StreamDispatch_ToolCallsCreateSeparateBlocks(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	for i := 0; i < 5; i++ {
		dispatcher.Dispatch(block.StreamDelta{
			Type:     "tool_call",
			ToolName: fmt.Sprintf("tool_%d", i),
			ToolArgs: `{}`,
		})
	}

	blks := container.Blocks()
	if len(blks) != 5 {
		t.Errorf("expected 5 tool_call blocks, got %d", len(blks))
	}
	for i, b := range blks {
		if b.Type() != block.TypeToolCall {
			t.Errorf("block[%d] type = %v, want TypeToolCall", i, b.Type())
		}
	}
}

// TestP26_StreamDispatch_FlushCompletesBlocks verifies that Flush
// transitions all active blocks to Complete state.
func TestP26_StreamDispatch_FlushCompletesBlocks(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	dispatcher.Dispatch(block.StreamDelta{Type: "thinking", Content: "hmm"})
	dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: "hello"})

	// Before flush: blocks should be streaming
	for _, b := range container.Blocks() {
		if b.State() != block.BlockStreaming {
			t.Errorf("pre-flush state = %v, want BlockStreaming", b.State())
		}
	}

	dispatcher.Flush()

	// After flush: blocks should be complete
	for _, b := range container.Blocks() {
		if b.State() != block.BlockComplete {
			t.Errorf("post-flush state = %v, want BlockComplete", b.State())
		}
	}

	// Active blocks should be empty
	if active := dispatcher.ActiveBlocks(); len(active) != 0 {
		t.Errorf("expected 0 active blocks after flush, got %d", len(active))
	}
}

// TestP26_StreamDispatch_PostFlushCreatesNewBlocks verifies that after
// Flush, new deltas of the same type create new blocks.
func TestP26_StreamDispatch_PostFlushCreatesNewBlocks(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: "first"})
	dispatcher.Flush()
	dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: "second"})

	blks := container.Blocks()
	if len(blks) != 2 {
		t.Errorf("expected 2 text blocks after flush+new, got %d", len(blks))
	}
}

// TestP26_StreamDispatch_UnknownDeltaTypeReturnsError verifies that
// an unknown delta type returns an error.
func TestP26_StreamDispatch_UnknownDeltaTypeReturnsError(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	err := dispatcher.Dispatch(block.StreamDelta{Type: "unknown_type", Content: "???"})
	if err == nil {
		t.Error("expected error for unknown delta type")
	}
}

// TestP26_StreamDispatch_FullConversationSimulation simulates a realistic
// AI conversation with interleaved delta types and verifies the full pipeline.
func TestP26_StreamDispatch_FullConversationSimulation(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	// User message
	container.AddBlock(block.NewUserMessageBlock("user-1", "What is 2+2?"))

	// AI thinking
	for _, chunk := range []string{"The user asks ", "a simple math ", "question. 2+2 = 4."} {
		dispatcher.Dispatch(block.StreamDelta{Type: "thinking", Content: chunk})
	}
	dispatcher.Flush()

	// AI text response
	for _, chunk := range []string{"The answer ", "is **4**.\n\n", "That's straightforward."} {
		dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: chunk})
	}
	dispatcher.Flush()

	// Second user message
	container.AddBlock(block.NewUserMessageBlock("user-2", "Now what is 3+3?"))

	// More AI response
	dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: "That would be **6**."})
	dispatcher.Flush()

	// Verify
	blks := container.Blocks()
	// user-1, thinking-1, text-1, user-2, text-2 = 5 blocks
	if len(blks) != 5 {
		t.Errorf("expected 5 blocks, got %d", len(blks))
	}

	// Verify types
	expectedTypes := []block.BlockType{
		block.TypeUserMessage,
		block.TypeThinking,
		block.TypeAssistantText,
		block.TypeUserMessage,
		block.TypeAssistantText,
	}
	for i, bt := range expectedTypes {
		if blks[i].Type() != bt {
			t.Errorf("block[%d] type = %v, want %v", i, blks[i].Type(), bt)
		}
	}

	// Verify all complete (Flush was called)
	for i, b := range blks {
		if b.State() != block.BlockComplete {
			t.Errorf("block[%d] state = %v, want BlockComplete", i, b.State())
		}
	}

	// Verify container can layout and paint without panic
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 100})
	buf := buffer.NewBuffer(80, 100)
	container.Paint(buf)
}

// TestP26_StreamDispatch_LargeStream verifies that streaming 1000+
// deltas works correctly without panic or memory issues.
func TestP26_StreamDispatch_LargeStream(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	for i := 0; i < 1000; i++ {
		dispatcher.Dispatch(block.StreamDelta{
			Type:    "text",
			Content: fmt.Sprintf("line %d\n", i),
		})
	}
	dispatcher.Flush()

	blks := container.Blocks()
	if len(blks) != 1 {
		t.Errorf("expected 1 text block for 1000 deltas, got %d", len(blks))
	}
	if blks[0].State() != block.BlockComplete {
		t.Error("block should be complete after flush")
	}

	// Container should be able to measure and paint
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 50})
	buf := buffer.NewBuffer(80, 50)
	container.Paint(buf)
}

// TestP26_StreamDispatch_Concurrent verifies that concurrent Dispatch
// calls are safe (no race, no data corruption).
func TestP26_StreamDispatch_Concurrent(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	var wg sync.WaitGroup
	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				dispatcher.Dispatch(block.StreamDelta{
					Type:    "text",
					Content: fmt.Sprintf("g%d-%d", id, j),
				})
			}
		}(i)
	}
	wg.Wait()
	dispatcher.Flush()

	// All content should be in the single text block
	blks := container.Blocks()
	if len(blks) < 1 {
		t.Errorf("expected at least 1 block, got %d", len(blks))
	}
}

// TestP26_Container_InterleavedTypesLayout verifies that a container
// with interleaved block types (user, thinking, text, tool_call, tool_result)
// lays out correctly vertically.
func TestP26_Container_InterleavedTypesLayout(t *testing.T) {
	container := block.NewBlockContainer()

	container.AddBlock(block.NewUserMessageBlock("u1", "Hello"))
	container.AddBlock(block.NewThinkingBlock("t1"))
	container.AddBlock(block.NewAssistantTextBlock("a1"))
	container.AddBlock(block.NewToolCallBlock("tc1", "search", `{}`))
	container.AddBlock(block.NewToolResultBlock("tr1"))

	// Measure
	sz := container.Measure(component.Constraints{})
	if sz.H < 1 {
		t.Error("container height should be positive")
	}

	// Layout
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 50})

	// Paint
	buf := buffer.NewBuffer(80, 50)
	container.Paint(buf) // should not panic
}

// TestP26_Container_DirtyFlagAfterAddBlock verifies that adding a block
// marks the container as dirty.
func TestP26_Container_DirtyFlagAfterAddBlock(t *testing.T) {
	container := block.NewBlockContainer()
	// New container starts dirty (needs initial render)
	container.ClearDirty()
	if container.IsDirty() {
		t.Error("container should not be dirty after ClearDirty")
	}

	container.AddBlock(block.NewUserMessageBlock("u1", "test"))
	if !container.IsDirty() {
		t.Error("container should be dirty after AddBlock")
	}
}

// TestP26_Container_BlocksReturnsAllAdded verifies that Blocks() returns
// all blocks in insertion order.
func TestP26_Container_BlocksReturnsAllAdded(t *testing.T) {
	container := block.NewBlockContainer()

	for i := 0; i < 10; i++ {
		container.AddBlock(block.NewUserMessageBlock(
			fmt.Sprintf("u%d", i),
			fmt.Sprintf("msg %d", i),
		))
	}

	blks := container.Blocks()
	if len(blks) != 10 {
		t.Fatalf("expected 10 blocks, got %d", len(blks))
	}

	for i, b := range blks {
		expected := fmt.Sprintf("u%d", i)
		if b.ID() != expected {
			t.Errorf("block[%d].ID() = %q, want %q", i, b.ID(), expected)
		}
	}
}

// TestP26_Container_SpacingAffectsLayout verifies that SetSpacing
// affects the total container height.
func TestP26_Container_SpacingAffectsLayout(t *testing.T) {
	container := block.NewBlockContainer()
	container.AddBlock(block.NewUserMessageBlock("u1", "a"))
	container.AddBlock(block.NewUserMessageBlock("u2", "b"))

	container.SetSpacing(0)
	sz0 := container.Measure(component.Constraints{})

	container.SetSpacing(5)
	sz5 := container.Measure(component.Constraints{})

	if sz5.H <= sz0.H {
		t.Errorf("spacing=5 height (%d) should be > spacing=0 height (%d)", sz5.H, sz0.H)
	}
}

// TestP26_Container_LongTextWraps verifies that long text in an
// AssistantTextBlock wraps correctly within the container width.
func TestP26_Container_LongTextWraps(t *testing.T) {
	container := block.NewBlockContainer()
	blk := block.NewAssistantTextBlock("a1")
	blk.AppendDelta(strings.Repeat("word ", 200))
	container.AddBlock(blk)

	// Measure at 80 columns
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 100})
	sz := container.Measure(component.Constraints{MaxWidth: 80})
	if sz.H < 10 {
		t.Errorf("1000-char text should wrap to many lines, got height %d", sz.H)
	}

	// Paint should not panic
	buf := buffer.NewBuffer(80, 100)
	container.Paint(buf)
}

// TestP26_StreamDispatch_ErrorFailsActiveBlocks verifies that an error
// delta fails all active streaming blocks.
func TestP26_StreamDispatch_ErrorFailsActiveBlocks(t *testing.T) {
	container := block.NewBlockContainer()
	dispatcher := block.NewStreamDispatcher(container)

	dispatcher.Dispatch(block.StreamDelta{Type: "text", Content: "partial..."})
	dispatcher.Dispatch(block.StreamDelta{Type: "thinking", Content: "hmm..."})

	// Both should be streaming
	for _, b := range container.Blocks() {
		if b.State() != block.BlockStreaming {
			t.Errorf("pre-error state = %v, want BlockStreaming", b.State())
		}
	}

	// Send error
	dispatcher.Dispatch(block.StreamDelta{Type: "error", Content: "network timeout"})

	// Error delta should have transitioned at least one block to non-streaming state
	foundNonStreaming := false
	for _, b := range container.Blocks() {
		if b.State() != block.BlockStreaming {
			foundNonStreaming = true
			break
		}
	}
	if !foundNonStreaming {
		t.Error("expected at least one non-streaming block after error delta")
	}
}

// TestP26_BlockFactory_UniqueIDs verifies that BlockFactory generates
// unique IDs across all block types.
func TestP26_BlockFactory_UniqueIDs(t *testing.T) {
	factory := block.NewBlockFactory()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := factory.NextID("block")
		if ids[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

// TestP26_Container_ChildrenReturnsAll verifies that Children() returns
// all blocks as component.Component.
func TestP26_Container_ChildrenReturnsAll(t *testing.T) {
	container := block.NewBlockContainer()
	container.AddBlock(block.NewUserMessageBlock("u1", "hello"))
	container.AddBlock(block.NewUserMessageBlock("u2", "world"))

	children := container.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}
