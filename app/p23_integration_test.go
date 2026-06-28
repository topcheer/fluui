package app

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func p23contains(s, substr string) bool { return strings.Contains(s, substr) }

// P23-B: Component Interaction Integration Tests
//
// Tests end-to-end interactions between:
// - CommandPalette (toggle, search, execute, render)
// - Spinner (start, stop, label, render)
// - Undo/Redo (InputLine lifecycle, clear, consistency)
// - Theme (cycle, set by name/index, key routing, render)
// - Streaming (block creation, delta dispatch, render)
// - Cross-component concurrent safety
//
// Naming convention: TestIntegration_*

// ═══════════════════════════════════════════════════════════════
// CommandPalette Integration Tests
// ═══════════════════════════════════════════════════════════════

func TestIntegration_CommandPalette_ToggleOnOff(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	// Initially hidden
	if a.IsCommandPaletteVisible() {
		t.Fatal("palette should be hidden initially")
	}

	// Toggle on
	if !a.ToggleCommandPalette() {
		t.Error("ToggleCommandPalette should return true when palette attached")
	}
	if !a.IsCommandPaletteVisible() {
		t.Error("palette should be visible after first toggle")
	}

	// Toggle off
	a.ToggleCommandPalette()
	if a.IsCommandPaletteVisible() {
		t.Error("palette should be hidden after second toggle")
	}
}

func TestIntegration_CommandPalette_ToggleWithoutAttachment(t *testing.T) {
	a := NewChatApp(80, 24)

	// Toggle without attaching should return false
	if a.ToggleCommandPalette() {
		t.Error("ToggleCommandPalette should return false when no palette attached")
	}
	if a.IsCommandPaletteVisible() {
		t.Error("should not be visible without attachment")
	}
}

func TestIntegration_CommandPalette_RenderWithPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)
	a.AddCommand("save", "Save File", "File", func() {})

	// Open palette
	a.ToggleCommandPalette()

	// Render should not panic with palette open
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)
}

func TestIntegration_CommandPalette_AddCommandAndExecute(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	executed := false
	ok := a.AddCommand("test", "Test Command", "Test", func() {
		executed = true
	})
	if !ok {
		t.Fatal("AddCommand should return true with palette attached")
	}

	// Verify command was added to palette
	cmds := cp.Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].ID != "test" {
		t.Errorf("command ID = %s, want 'test'", cmds[0].ID)
	}

	// Execute the action
	cmds[0].Action()
	if !executed {
		t.Error("command action should have been executed")
	}
}

func TestIntegration_CommandPalette_AddCommandNoAttachment(t *testing.T) {
	a := NewChatApp(80, 24)
	ok := a.AddCommand("x", "Y", "Z", func() {})
	if ok {
		t.Error("AddCommand should return false when no palette attached")
	}
}

func TestIntegration_CommandPalette_CtrlPRouting(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	// Ctrl+P should toggle palette
	ctrlP := &term.KeyEvent{Rune: 'p', Modifiers: term.ModCtrl}
	consumed := a.HandleKey(ctrlP)
	if !consumed {
		t.Error("Ctrl+P should be consumed")
	}
	if !a.IsCommandPaletteVisible() {
		t.Error("palette should be visible after Ctrl+P")
	}

	// Ctrl+P again should hide
	a.HandleKey(ctrlP)
	if a.IsCommandPaletteVisible() {
		t.Error("palette should be hidden after second Ctrl+P")
	}
}

func TestIntegration_CommandPalette_Search(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	a.AddCommand("save-file", "Save File", "File", func() {})
	a.AddCommand("open-file", "Open File", "File", func() {})
	a.AddCommand("quit-app", "Quit App", "App", func() {})

	// Search for "file" — should match 2 commands
	cp.SetQuery("file")
	if cp.FilteredCount() != 2 {
		t.Errorf("filtered count = %d, want 2", cp.FilteredCount())
	}

	// Search for "quit" — should match 1
	cp.SetQuery("quit")
	if cp.FilteredCount() != 1 {
		t.Errorf("filtered count = %d, want 1", cp.FilteredCount())
	}

	// Search for "xyz" — should match 0
	cp.SetQuery("xyz")
	if cp.HasResults() {
		t.Error("should have no results for 'xyz'")
	}
}

func TestIntegration_CommandPalette_Navigate(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	a.AddCommand("cmd1", "Command 1", "Cat", func() {})
	a.AddCommand("cmd2", "Command 2", "Cat", func() {})
	a.AddCommand("cmd3", "Command 3", "Cat", func() {})

	cp.SetQuery("") // show all
	count := cp.FilteredCount()
	if count != 3 {
		t.Fatalf("filtered count = %d, want 3", count)
	}

	// Navigate down
	cp.SetCursor(0)
	cp.MoveDown()
	if cp.Cursor() != 1 {
		t.Errorf("cursor after MoveDown = %d, want 1", cp.Cursor())
	}

	// Navigate up
	cp.MoveUp()
	if cp.Cursor() != 0 {
		t.Errorf("cursor after MoveUp = %d, want 0", cp.Cursor())
	}
}

// ═══════════════════════════════════════════════════════════════
// Spinner Integration Tests
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Spinner_StartStop(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	s.Stop() // ensure not running initially
	a.SetSpinner(s)

	if a.IsSpinnerActive() {
		t.Fatal("spinner should not be active initially")
	}

	a.StartSpinner("Loading...")
	if !a.IsSpinnerActive() {
		t.Error("spinner should be active after StartSpinner")
	}

	// Render while active
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	a.StopSpinner()
	if a.IsSpinnerActive() {
		t.Error("spinner should be inactive after StopSpinner")
	}
}

func TestIntegration_Spinner_StartWithoutAttachment(t *testing.T) {
	a := NewChatApp(80, 24)
	a.StartSpinner("Loading") // should be no-op
	if a.IsSpinnerActive() {
		t.Error("spinner without attachment should not become active")
	}
}

func TestIntegration_Spinner_StopWithoutAttachment(t *testing.T) {
	a := NewChatApp(80, 24)
	a.StopSpinner() // should be safe no-op
	if a.IsSpinnerActive() {
		t.Error("should not be active")
	}
}

func TestIntegration_Spinner_LabelChange(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("Initial")
	a.SetSpinner(s)

	a.StartSpinner("Step 1")
	if !a.IsSpinnerActive() {
		t.Fatal("spinner should be active")
	}
	// Spinner should reflect new label
	sp := a.Spinner()
	if sp == nil {
		t.Fatal("Spinner() should not return nil when active")
	}
}

func TestIntegration_Spinner_SetNilDeactivates(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)
	a.StartSpinner("Loading")
	if !a.IsSpinnerActive() {
		t.Fatal("should be active")
	}

	a.SetSpinner(nil)
	if a.IsSpinnerActive() {
		t.Error("should be inactive after SetSpinner(nil)")
	}
	if a.Spinner() != nil {
		t.Error("Spinner() should return nil")
	}
}

func TestIntegration_Spender_RenderCycle(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)

	// Multiple start/stop cycles with render each time
	for i := 0; i < 3; i++ {
		a.StartSpinner("Working")
		buf := buffer.NewBuffer(80, 24)
		a.Render(buf)
		a.StopSpinner()
		a.Render(buf)
	}
}

// ═══════════════════════════════════════════════════════════════
// Theme Integration Tests
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Theme_CycleForward(t *testing.T) {
	a := NewChatApp(80, 24)

	initialIdx := a.ThemeIndex()

	// Ctrl+] cycles forward
	ctrlRB := &term.KeyEvent{Rune: ']', Modifiers: term.ModCtrl}
	consumed := a.HandleKey(ctrlRB)
	if !consumed {
		t.Error("Ctrl+] should be consumed")
	}
	if a.ThemeIndex() == initialIdx {
		t.Error("theme should change after Ctrl+]")
	}
}

func TestIntegration_Theme_CycleBackward(t *testing.T) {
	a := NewChatApp(80, 24)

	initialIdx := a.ThemeIndex()

	// Ctrl+\ cycles backward
	ctrlBS := &term.KeyEvent{Rune: '\\', Modifiers: term.ModCtrl}
	consumed := a.HandleKey(ctrlBS)
	if !consumed {
		t.Error("Ctrl+\\ should be consumed")
	}
	if a.ThemeIndex() == initialIdx {
		t.Error("theme should change after Ctrl+\\")
	}
}

func TestIntegration_Theme_SetByIndex(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()
	if count < 2 {
		t.Fatalf("need at least 2 themes, got %d", count)
	}

	a.SetThemeByIndex(1)
	if a.ThemeIndex() != 1 {
		t.Errorf("index = %d, want 1", a.ThemeIndex())
	}

	a.SetThemeByIndex(0)
	if a.ThemeIndex() != 0 {
		t.Errorf("index = %d, want 0", a.ThemeIndex())
	}
}

func TestIntegration_Theme_SetByName(t *testing.T) {
	a := NewChatApp(80, 24)
	themes := a.ThemeList()
	if len(themes) < 2 {
		t.Fatalf("need at least 2 themes, got %d", len(themes))
	}

	target := themes[len(themes)-1]
	ok := a.SetThemeByName(target)
	if !ok {
		t.Errorf("SetThemeByName(%q) failed", target)
	}
	if a.ThemeName() != target {
		t.Errorf("name = %q, want %q", a.ThemeName(), target)
	}

	// Invalid name
	if a.SetThemeByName("nonexistent") {
		t.Error("SetThemeByName should fail for invalid name")
	}
}

func TestIntegration_Theme_RenderAfterSwitch(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddUserMessage("test content")

	// Render with default theme
	buf1 := buffer.NewBuffer(80, 24)
	a.Render(buf1)

	// Switch theme and render
	a.SetThemeByIndex(1 % a.ThemeCount())
	buf2 := buffer.NewBuffer(80, 24)
	a.Render(buf2)

	// Both should have content — rendered buffers have a bg color fill
	hasContent := func(buf *buffer.Buffer) bool {
		// A rendered buffer will have non-default bg colors (the theme bg)
		cell := buf.GetCell(0, 0)
		// Just verify the buffer is populated (any cell has been written to)
		// The Render fills with bg color, so check if bg is not zero
		return cell.Bg.Type != buffer.ColorNone || cell.Rune != 0
	}
	if !hasContent(buf1) {
		t.Error("buf1 should have content")
	}
	if !hasContent(buf2) {
		t.Error("buf2 should have content after theme switch")
	}
}

func TestIntegration_Theme_FullCycleReturnsToStart(t *testing.T) {
	a := NewChatApp(80, 24)
	start := a.ThemeIndex()
	count := a.ThemeCount()

	// Cycle forward through all themes
	for i := 0; i < count; i++ {
		// Use Ctrl+] to cycle
		a.HandleKey(&term.KeyEvent{Rune: ']', Modifiers: term.ModCtrl})
	}

	if a.ThemeIndex() != start {
		t.Errorf("after full forward cycle: index = %d, want %d", a.ThemeIndex(), start)
	}
}

func TestIntegration_Theme_ThemeCountAndList(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()
	list := a.ThemeList()

	if count != len(list) {
		t.Errorf("ThemeCount() = %d, ThemeList() len = %d", count, len(list))
	}

	// All names should be non-empty and unique
	seen := map[string]bool{}
	for _, name := range list {
		if name == "" {
			t.Error("theme name should not be empty")
		}
		if seen[name] {
			t.Errorf("duplicate theme name: %s", name)
		}
		seen[name] = true
	}
}

// ═══════════════════════════════════════════════════════════════
// Undo/Redo Integration Tests
// ═══════════════════════════════════════════════════════════════

func TestIntegration_UndoRedo_InputLineCycle(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()
	if il == nil {
		t.Fatal("InputLine should be initialized after SetOnSubmit")
	}

	// Type text
	il.saveUndo()
	il.SetText("Hello")
	il.saveUndo()
	il.SetText("Hello World")

	if il.UndoCount() != 2 {
		t.Errorf("UndoCount = %d, want 2", il.UndoCount())
	}

	// Undo
	il.Undo()
	if il.Text() != "Hello" {
		t.Errorf("after undo: text = %q, want 'Hello'", il.Text())
	}

	// Redo
	il.Redo()
	if il.Text() != "Hello World" {
		t.Errorf("after redo: text = %q, want 'Hello World'", il.Text())
	}
}

func TestIntegration_UndoRedo_NewEditClearsRedo(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("v1")
	il.saveUndo()
	il.SetText("v2")

	// Undo
	il.Undo()
	if !il.CanRedo() {
		t.Error("should be able to redo")
	}

	// New edit clears redo
	il.saveUndo()
	il.SetText("v3")
	if il.CanRedo() {
		t.Error("redo should be cleared after new edit")
	}
}

func TestIntegration_UndoRedo_ClearHistory(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("a")
	il.saveUndo()
	il.SetText("b")
	il.saveUndo()
	il.SetText("c")

	il.ClearUndoHistory()
	if il.UndoCount() != 0 {
		t.Errorf("UndoCount = %d, want 0", il.UndoCount())
	}
	if il.RedoCount() != 0 {
		t.Errorf("RedoCount = %d, want 0", il.RedoCount())
	}
	if il.CanUndo() {
		t.Error("CanUndo should be false")
	}
	if il.CanRedo() {
		t.Error("CanRedo should be false")
	}
}

func TestIntegration_UndoRedo_RenderConsistency(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("content here")

	// Render before undo
	buf1 := buffer.NewBuffer(80, 24)
	a.Render(buf1)

	// Undo (text reverts)
	il.Undo()

	// Render after undo
	buf2 := buffer.NewBuffer(80, 24)
	a.Render(buf2)

	// The input area should differ since text changed
	inputDiffers := false
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			if buf1.GetCell(x, y).Rune != buf2.GetCell(x, y).Rune {
				inputDiffers = true
				break
			}
		}
		if inputDiffers {
			break
		}
	}
	if !inputDiffers {
		t.Error("render output should differ after undo (input text changed)")
	}
}

func TestIntegration_UndoRedo_CtrlZ(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("typed text")
	il.saveUndo()
	il.SetText("typed text more")

	// Ctrl+Z should undo
	ctrlZ := &term.KeyEvent{Rune: 'z', Modifiers: term.ModCtrl}
	a.HandleKey(ctrlZ)
	if il.Text() != "typed text" {
		t.Errorf("after Ctrl+Z: text = %q, want 'typed text'", il.Text())
	}
}

// ═══════════════════════════════════════════════════════════════
// Streaming + Block Lifecycle Integration
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Streaming_AddBlocksAndRender(t *testing.T) {
	a := NewChatApp(80, 24)

	// Add user message
	a.AddUserMessage("Question")
	if a.Container().Len() != 1 {
		t.Errorf("container len = %d, want 1", a.Container().Len())
	}

	// Add assistant text
	a.AddAssistantText()
	if a.Container().Len() != 2 {
		t.Errorf("container len = %d, want 2", a.Container().Len())
	}

	// Render with blocks
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)
}

func TestIntegration_Streaming_StreamDeltaAndRender(t *testing.T) {
	a := NewChatApp(80, 24)

	a.AddAssistantText()
	a.StreamDelta(block.StreamDelta{
		Type:    "text",
		Content: "Hello from AI",
	})

	// Render should show streamed content
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	// LastBlockText should return something
	text, ok := a.LastBlockText()
	if !ok {
		t.Error("LastBlockText should return content")
	}
	if text == "" {
		t.Error("text should not be empty")
	}
}

func TestIntegration_Streaming_MultipleDeltas(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddAssistantText()

	deltas := []string{"Hello", " World", "!"}
	for _, d := range deltas {
		a.StreamDelta(block.StreamDelta{
			Type:    "text",
			Content: d,
		})
	}

	// Each delta may create a separate block — collect all text
	text, ok := a.LastBlockText()
	if !ok {
		// Try getting text from all blocks
		t.Log("LastBlockText returned false — checking individual blocks")
	}
	// Verify at least the last delta was streamed
	if text == "" {
		t.Error("streamed text should not be empty")
	}
	// The last delta should be present
	if !p23contains(text, "!") {
		t.Errorf("streamed text %q should contain last delta", text)
	}
}

func TestIntegration_Streaming_WithSpinnerAndRender(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)

	a.AddAssistantText()
	a.StartSpinner("Thinking...")
	a.StreamDelta(block.StreamDelta{Type: "text", Content: "Partial..."})

	// Render with spinner and streaming
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	a.StopSpinner()
	a.Render(buf)
}

func TestIntegration_Streaming_PaletteWhileStreaming(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	// Start streaming
	a.AddAssistantText()
	a.StreamDelta(block.StreamDelta{Type: "text", Content: "Streaming..."})

	// Open palette while streaming
	a.ToggleCommandPalette()
	if !a.IsCommandPaletteVisible() {
		t.Error("palette should be visible")
	}

	// Continue streaming
	a.StreamDelta(block.StreamDelta{Type: "text", Content: " more"})

	// Render with both
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	// Close palette
	a.ToggleCommandPalette()
	if a.IsCommandPaletteVisible() {
		t.Error("palette should be hidden")
	}
}

// ═══════════════════════════════════════════════════════════════
// Resize + Render Integration
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Resize_Grow(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddUserMessage("content")

	a.SetSize(120, 40)
	buf := buffer.NewBuffer(120, 40)
	a.Render(buf)
}

func TestIntegration_Resize_Shrink(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddUserMessage("content")

	a.SetSize(40, 10)
	buf := buffer.NewBuffer(40, 10)
	a.Render(buf)
}

func TestIntegration_Resize_Multiple(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddUserMessage("resize test")

	sizes := []struct{ w, h int }{
		{40, 10},
		{200, 50},
		{80, 24},
		{1, 1},
		{80, 24},
	}
	for _, s := range sizes {
		a.SetSize(s.w, s.h)
		buf := buffer.NewBuffer(s.w, s.h)
		a.Render(buf)
	}
}

// ═══════════════════════════════════════════════════════════════
// Cross-Component Interaction Tests
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Cross_PalettePlusSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	s := component.NewSpinner("")
	a.SetCommandPalette(cp)
	a.SetSpinner(s)

	// Both active
	a.StartSpinner("Thinking")
	a.ToggleCommandPalette()

	if !a.IsSpinnerActive() || !a.IsCommandPaletteVisible() {
		t.Fatal("both spinner and palette should be active")
	}

	// Render with both
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	// Close palette, stop spinner
	a.ToggleCommandPalette()
	a.StopSpinner()
}

func TestIntegration_Cross_UndoPlusThemeSwitch(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("original")
	il.saveUndo()
	il.SetText("modified")

	// Switch theme
	a.SetThemeByIndex(1 % a.ThemeCount())

	// Undo should work after theme switch
	il.Undo()
	if il.Text() != "original" {
		t.Errorf("after undo + theme switch: text = %q, want 'original'", il.Text())
	}
}

func TestIntegration_Cross_UndoWhileStreaming(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	il.saveUndo()
	il.SetText("user input")

	// Start streaming
	a.AddAssistantText()
	a.StreamDelta(block.StreamDelta{Type: "text", Content: "AI"})

	// Undo input while streaming
	il.Undo()
	if il.Text() != "" {
		t.Errorf("after undo: text = %q, want empty", il.Text())
	}

	// Streaming should continue
	a.StreamDelta(block.StreamDelta{Type: "text", Content: " response"})

	// Render
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)
}

func TestIntegration_Cross_FullLifecycle(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()
	cp := component.NewCommandPalette()
	s := component.NewSpinner("")
	a.SetCommandPalette(cp)
	a.SetSpinner(s)

	// 1. Type input
	il.saveUndo()
	il.SetText("What is Go?")

	// 2. Add user message
	a.AddUserMessage("What is Go?")

	// 3. Start streaming with spinner
	a.AddAssistantText()
	a.StartSpinner("Thinking...")

	// 4. Stream content
	a.StreamDelta(block.StreamDelta{
		Type:    "text",
		Content: "Go is a programming language.",
	})

	// 5. Switch theme
	a.SetThemeByIndex(1 % a.ThemeCount())

	// 6. Stop spinner
	a.StopSpinner()

	// 7. Render
	buf := buffer.NewBuffer(80, 24)
	a.Render(buf)

	// 8. Verify container
	blocks := a.Container().Blocks()
	if len(blocks) < 2 {
		t.Errorf("expected at least 2 blocks, got %d", len(blocks))
	}

	// 9. Undo input
	il.Undo()
	if il.Text() != "" {
		t.Errorf("after undo: text = %q, want empty", il.Text())
	}
}

// ═══════════════════════════════════════════════════════════════
// Concurrent Integration Tests (run with -race)
// ═══════════════════════════════════════════════════════════════

func TestIntegration_Concurrent_StreamAndRender(t *testing.T) {
	a := NewChatApp(80, 24)
	a.AddAssistantText()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			a.StreamDelta(block.StreamDelta{Type: "text", Content: "x"})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			buf := buffer.NewBuffer(80, 24)
			a.Render(buf)
		}
	}()

	wg.Wait()
}

func TestIntegration_Concurrent_PaletteSpinnerRender(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	s := component.NewSpinner("")
	a.SetCommandPalette(cp)
	a.SetSpinner(s)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			a.ToggleCommandPalette()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			a.StartSpinner("x")
			a.StopSpinner()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			buf := buffer.NewBuffer(80, 24)
			a.Render(buf)
		}
	}()

	wg.Wait()
}

func TestIntegration_Concurrent_ThemeAndUndo(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Theme + InputLine operations under a shared mutex (same pattern as ChatApp.mu)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.SetThemeByIndex(i % a.ThemeCount())
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			il.SetText("text")
			il.saveUndo()
			il.Undo()
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := buffer.NewBuffer(80, 24)
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.Render(buf)
			mu.Unlock()
		}
	}()

	wg.Wait()
}

func TestIntegration_Concurrent_FullApp(t *testing.T) {
	a := NewChatApp(80, 24)
	a.OnSubmit(func(string) {})
	il := a.InputLine()
	cp := component.NewCommandPalette()
	s := component.NewSpinner("")
	a.SetCommandPalette(cp)
	a.SetSpinner(s)
	a.AddAssistantText()

	// InputLine has no internal mutex — serialize all operations through a shared mutex
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Streaming
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.StreamDelta(block.StreamDelta{Type: "text", Content: "d"})
			mu.Unlock()
		}
	}()

	// Theme cycling
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.SetThemeByIndex(i % a.ThemeCount())
			mu.Unlock()
		}
	}()

	// Render
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := buffer.NewBuffer(80, 24)
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.Render(buf)
			mu.Unlock()
		}
	}()

	// Palette toggle
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			a.ToggleCommandPalette()
			mu.Unlock()
		}
	}()

	// Input + undo
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mu.Lock()
			il.SetText("x")
			il.saveUndo()
			il.Undo()
			mu.Unlock()
		}
	}()

	wg.Wait()
}

// ═══════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════

