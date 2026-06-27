package app_test

import (
	"testing"

	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
)

// === P4 Integration Test 1: Full chat flow with all block types ===

func TestP4IntegrationFullChatFlow(t *testing.T) {
	chat := app.NewChatApp(80, 24)

	// User message
	chat.AddUserMessage("What files are in this project?")

	// Thinking
	thinking := chat.AddThinking()
	thinking.AppendDelta("Let me check the project structure.")
	thinking.AppendDelta(" I'll use list_files.")
	thinking.Complete()

	// Tool call
	tc := chat.AddToolCall("list_files", `{"dir": "."}`)
	tc.Complete()

	// Tool result
	tr := chat.AddToolResult()
	tr.AppendDelta("main.go\ngo.mod\nREADME.md\nbuffer.go")
	tr.Complete()

	// Assistant text
	at := chat.AddAssistantText()
	at.AppendDelta("The project has 4 files: main.go, go.mod, README.md, and buffer.go.")
	at.Complete()

	// Verify container
	if chat.Container().Len() != 5 {
		t.Fatalf("expected 5 blocks, got %d", chat.Container().Len())
	}

	// Verify block types
	expectedTypes := []block.BlockType{
		block.TypeUserMessage,
		block.TypeThinking,
		block.TypeToolCall,
		block.TypeToolResult,
		block.TypeAssistantText,
	}
	for i, et := range expectedTypes {
		b := chat.Container().Blocks()[i]
		if b.Type() != et {
			t.Errorf("block %d: expected %s, got %s", i, et, b.Type())
		}
	}

	// Render to buffer — verify content
	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)

	hasContent := false
	for y := 0; y < 24 && !hasContent; y++ {
		for x := 0; x < 80; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != 0 && c.Rune != ' ' {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("buffer should have rendered content")
	}

	// Clear
	chat.Clear()
	if chat.Container().Len() != 0 {
		t.Errorf("after Clear: expected 0 blocks, got %d", chat.Container().Len())
	}
}

// === P4 Integration Test 2: Overlay modal ===

func TestP4IntegrationOverlayModal(t *testing.T) {
	chat := app.NewChatApp(80, 24)
	chat.AddUserMessage("Hello")

	// No modal initially
	if chat.Overlays().HasModal() {
		t.Error("should not have modal initially")
	}

	// Create and add modal
	body := component.NewText("Are you sure you want to quit?")
	modal := overlay.NewModal("quit-modal", "Confirm Quit", body, []string{"OK", "Cancel"})
	modal.Measure(component.Bounded(80, 24))
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	chat.Overlays().Add(modal)

	if !chat.Overlays().HasModal() {
		t.Error("should have modal after Add")
	}

	// Render — should not panic
	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)

	// Esc closes modal
	consumed := chat.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("HandleKey(Esc) should return true when modal is visible")
	}
	if modal.Visible() {
		t.Error("modal should be hidden after Esc")
	}

	// Render again — no modal
	chat.Render(buf)
	if chat.Overlays().HasModal() {
		t.Error("should not have modal after Esc")
	}
}

// === P4 Integration Test 3: Scroll with many blocks ===

func TestP4IntegrationScrollAndRender(t *testing.T) {
	chat := app.NewChatApp(80, 24)

	// Add 30 blocks
	for i := 0; i < 30; i++ {
		chat.AddUserMessage("Message line that fills the screen")
	}

	if chat.Container().Len() != 30 {
		t.Fatalf("expected 30 blocks, got %d", chat.Container().Len())
	}

	// Scroll down multiple times
	for i := 0; i < 10; i++ {
		chat.ScrollDown()
	}

	// Render should not panic
	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)

	// Scroll back up
	for i := 0; i < 10; i++ {
		chat.ScrollUp()
	}
	chat.Render(buf)

	// Set input height — render should not panic
	chat.SetInputHeight(2)
	chat.Render(buf)
}

// === P4 Integration Test 4: StreamDispatcher with all delta types ===

func TestP4IntegrationStreamDispatcher(t *testing.T) {
	chat := app.NewChatApp(80, 24)

	// Stream thinking
	chat.StreamDelta(block.StreamDelta{Type: "thinking", Content: "Analyzing..."})
	chat.StreamDelta(block.StreamDelta{Type: "thinking", Content: " Done."})

	// Stream text
	chat.StreamDelta(block.StreamDelta{Type: "text", Content: "Here is "})
	chat.StreamDelta(block.StreamDelta{Type: "text", Content: "my response."})

	// Stream tool call
	chat.StreamDelta(block.StreamDelta{
		Type:     "tool_call",
		ToolName: "read_file",
		ToolArgs: `{"path": "main.go"}`,
	})

	// Stream tool result
	chat.StreamDelta(block.StreamDelta{
		Type:    "tool_result",
		Content: "package main\nfunc main() {}",
	})

	// Should have at least 4 blocks: thinking, text, tool_call, tool_result
	blocks := chat.Container().Blocks()
	if len(blocks) < 4 {
		t.Fatalf("expected >= 4 blocks, got %d", len(blocks))
	}

	// Verify types are present
	typesSeen := make(map[block.BlockType]bool)
	for _, b := range blocks {
		typesSeen[b.Type()] = true
	}
	for _, et := range []block.BlockType{block.TypeThinking, block.TypeAssistantText, block.TypeToolCall, block.TypeToolResult} {
		if !typesSeen[et] {
			t.Errorf("expected block type %s in container", et)
		}
	}

	// Render should work
	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)
}

// === P4 Integration Test 5: Keyboard handling ===

func TestP4IntegrationKeyboard(t *testing.T) {
	chat := app.NewChatApp(80, 24)
	chat.AddUserMessage("Test message")
	chat.AddAssistantText()

	quitCalled := false
	chat.OnQuit(func() { quitCalled = true })

	// Up/Down scroll — should not panic
	chat.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	chat.HandleKey(&term.KeyEvent{Key: term.KeyDown})

	// PageUp/PageDown
	chat.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	chat.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})

	// Home/End
	chat.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	chat.HandleKey(&term.KeyEvent{Key: term.KeyEnd})

	// 'q' should trigger quit
	chat.HandleKey(&term.KeyEvent{Rune: 'q'})
	if !quitCalled {
		t.Error("expected quit to be called on 'q'")
	}

	// With modal, Esc should go to modal not quit
	quitCalled = false
	modal := overlay.NewModal("m1", "Test", component.NewText("body"), []string{"OK"})
	modal.Measure(component.Bounded(80, 24))
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	chat.Overlays().Add(modal)

	chat.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if quitCalled {
		t.Error("should not quit when modal is visible and Esc is pressed")
	}
	if modal.Visible() {
		t.Error("modal should be hidden after Esc")
	}
}
