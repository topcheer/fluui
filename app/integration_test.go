package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
)

// TestIntegrationFullChatFlow tests a complete chat conversation lifecycle:
// UserMessage → Thinking(stream) → ToolCall → ToolResult → AssistantText(stream) → Render
func TestIntegrationFullChatFlow(t *testing.T) {
	app := NewChatApp(80, 24)

	// User message
	app.AddUserMessage("What files are in the project?")

	// Thinking stream
	app.StreamDelta(block.StreamDelta{Type: "thinking", Content: "Let me check the project structure."})
	app.StreamDelta(block.StreamDelta{Type: "thinking", Content: " I'll list the files."})

	// Tool call
	app.StreamDelta(block.StreamDelta{
		Type:     "tool_call",
		ToolName: "list_files",
		ToolArgs: `{path: "."}`,
	})

	// Tool result
	app.StreamDelta(block.StreamDelta{Type: "tool_result", Content: "main.go\ninput.go\nbuffer.go"})

	// Assistant text stream
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "The project has 3 files: "})
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "main.go, input.go, and buffer.go."})

	// Verify container has blocks
	if app.Container().Len() < 4 {
		t.Errorf("Container.Len() = %d, want >= 4", app.Container().Len())
	}

	// Render — should not panic and should produce content
	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	hasContent := false
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			if c := buf.GetCell(x, y); c.Rune != 0 && c.Rune != ' ' {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if !hasContent {
		t.Error("Render should produce visible content")
	}
}

// TestIntegrationStreamAndRender verifies that streaming deltas set dirty,
// ClearDirty clears it, and new deltas set it again.
func TestIntegrationStreamAndRender(t *testing.T) {
	app := NewChatApp(80, 24)

	// Stream 5 text deltas
	for i := 0; i < 5; i++ {
		app.StreamDelta(block.StreamDelta{Type: "text", Content: "line\n"})
	}

	// Should be dirty
	if !app.IsDirty() {
		t.Error("should be dirty after streaming")
	}

	// Render to buffer
	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	// Clear dirty
	app.ClearDirty()
	if app.IsDirty() {
		t.Error("should not be dirty after ClearDirty")
	}

	// Stream one more delta
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "more text"})
	if !app.IsDirty() {
		t.Error("should be dirty again after new StreamDelta")
	}
}

// TestIntegrationOverlayModal tests ChatApp + Modal overlay integration.
func TestIntegrationOverlayModal(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("Hello")

	// No modal initially
	if app.Overlays().HasModal() {
		t.Error("should not have modal initially")
	}

	// Add a modal
	modal := overlay.NewModal("test-modal", "Confirm", component.NewText("Are you sure?"), []string{"OK", "Cancel"})
	modal.Measure(component.Bounded(80, 24))
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.Overlays().Add(modal)

	if !app.Overlays().HasModal() {
		t.Error("should have modal after Add")
	}

	// Render should not panic
	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	// Remove modal
	app.Overlays().Remove("test-modal")
	if app.Overlays().HasModal() {
		t.Error("should not have modal after Remove")
	}
}

// TestIntegrationKeyRouting verifies keyboard event routing:
// With modal visible, Esc goes to modal (not app quit).
// Without modal, Esc triggers app quit.
func TestIntegrationKeyRouting(t *testing.T) {
	app := NewChatApp(80, 24)
	quitCalled := false
	app.OnQuit(func() { quitCalled = true })

	// Add modal
	modal := overlay.NewModal("m1", "T", component.NewText("body"), []string{"OK"})
	modal.Measure(component.Bounded(80, 24))
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.Overlays().Add(modal)

	// Esc with modal present — modal should consume it, app should NOT quit
	consumed := app.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("HandleKey should return true (modal consumed Esc)")
	}
	if quitCalled {
		t.Error("app should not quit when modal is visible")
	}
	if modal.Visible() {
		t.Error("modal should be hidden after Esc")
	}

	// Now Esc without modal — should trigger quit
	quitCalled = false
	app.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !quitCalled {
		t.Error("app should quit when no modal is visible")
	}
}

// TestIntegrationScrollContent adds 50+ blocks and verifies scroll + render work.
func TestIntegrationScrollContent(t *testing.T) {
	app := NewChatApp(80, 24)

	// Add 50 user message blocks
	for i := 0; i < 50; i++ {
		app.AddUserMessage("Message line")
	}

	if app.Container().Len() != 50 {
		t.Errorf("Container.Len() = %d, want 50", app.Container().Len())
	}

	// Scroll down several times
	for i := 0; i < 10; i++ {
		app.ScrollDown()
	}

	// Render should not panic
	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	// Scroll back up
	for i := 0; i < 10; i++ {
		app.ScrollUp()
	}

	// Render again
	app.Render(buf)
}

// TestIntegrationMultipleSessions verifies Clear + restart works correctly.
func TestIntegrationMultipleSessions(t *testing.T) {
	app := NewChatApp(80, 24)

	// Session 1
	app.AddUserMessage("First question")
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "First answer"})
	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)
	firstCount := app.Container().Len()
	if firstCount < 2 {
		t.Errorf("session 1: Len = %d, want >= 2", firstCount)
	}

	// Clear
	app.Clear()
	if app.Container().Len() != 0 {
		t.Errorf("after Clear: Len = %d, want 0", app.Container().Len())
	}

	// Session 2
	app.AddUserMessage("Second question")
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "Second answer"})
	app.StreamDelta(block.StreamDelta{Type: "thinking", Content: "Hmm..."})
	app.Render(buf)
	secondCount := app.Container().Len()
	if secondCount < 2 {
		t.Errorf("session 2: Len = %d, want >= 2", secondCount)
	}

	// Session 3 (empty then add)
	app.Clear()
	app.AddUserMessage("Third")
	app.Render(buf)
	if app.Container().Len() != 1 {
		t.Errorf("session 3: Len = %d, want 1", app.Container().Len())
	}
}
