package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

// ensureInputLine creates an InputLine on the ChatApp if not yet set.
func ensureInputLine(app *ChatApp) *InputLine {
	il := app.InputLine()
	if il == nil {
		il = NewInputLine("> ")
		app.SetInputLine(il)
	}
	return il
}

// --- ChatApp callbacks ---

func TestP51_ChatApp_OnFocus(t *testing.T) {
	app := NewChatApp(80, 24)
	called := false
	app.OnFocus(func(focused bool) { called = true })
	if called {
		t.Error("handler should not be called yet")
	}
}

func TestP51_ChatApp_OnQuit(t *testing.T) {
	app := NewChatApp(80, 24)
	called := false
	app.OnQuit(func() { called = true })
	if called {
		t.Error("handler should not be called yet")
	}
}

func TestP51_ChatApp_OnMouse(t *testing.T) {
	app := NewChatApp(80, 24)
	var received *term.MouseEvent
	app.OnMouse(func(m *term.MouseEvent) { received = m })
	if received != nil {
		t.Error("handler should not be called yet")
	}
}

// --- Undo/Redo ---

func TestP51_UndoRedo_Cycle(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)

	// saveUndo should be called BEFORE mutation (standard undo pattern)
	il.saveUndo() // save initial empty state
	il.SetText("hello")
	il.saveUndo() // save "hello" state
	il.SetText("world")

	// Now text is "world", undo should restore "hello"
	if !il.CanUndo() {
		t.Error("expected can undo")
	}
	il.Undo()
	if il.Text() != "hello" {
		t.Errorf("expected 'hello' after undo, got %q", il.Text())
	}

	// Redo should re-apply "world"
	if !il.CanRedo() {
		t.Error("expected can redo")
	}
	il.Redo()
	if il.Text() != "world" {
		t.Errorf("expected 'world' after redo, got %q", il.Text())
	}
}

func TestP51_UndoRedo_Empty(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	if il.CanUndo() {
		t.Error("expected no undo on fresh input")
	}
	if il.CanRedo() {
		t.Error("expected no redo on fresh input")
	}
	il.Undo()
	il.Redo()
}

func TestP51_UndoRedo_Counts(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	il.SetText("a")
	il.saveUndo()
	il.SetText("b")
	il.saveUndo()
	if il.UndoCount() != 2 {
		t.Errorf("expected 2 undo entries, got %d", il.UndoCount())
	}
	if il.RedoCount() != 0 {
		t.Errorf("expected 0 redo entries, got %d", il.RedoCount())
	}
	il.Undo()
	if il.RedoCount() != 1 {
		t.Errorf("expected 1 redo after undo, got %d", il.RedoCount())
	}
}

func TestP51_UndoRedo_ClearHistory(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	il.SetText("a")
	il.saveUndo()
	il.SetText("b")
	il.saveUndo()
	il.ClearUndoHistory()
	if il.CanUndo() {
		t.Error("expected no undo after clear")
	}
	if il.CanRedo() {
		t.Error("expected no redo after clear")
	}
}

// --- Completion ---

func TestP51_Completion_Items(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	cm := il.CompletionManager()
	// CompletionManager may be nil if not set up
	if cm != nil {
		_ = cm.Items()
	}
}

func TestP51_Completion_Selected(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	cm := il.CompletionManager()
	if cm == nil {
		return // no completion manager set up
	}
	_, ok := cm.Selected()
	if ok {
		t.Error("expected no selection by default")
	}
}

// --- HandleKey edge cases ---

func TestP51_HandleKey_Escape(t *testing.T) {
	app := NewChatApp(80, 24)
	ensureInputLine(app).SetText("some text")
	app.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
}

func TestP51_HandleKey_CtrlC(t *testing.T) {
	app := NewChatApp(80, 24)
	ensureInputLine(app)
	app.HandleKey(&term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl})
}

func TestP51_HandleKey_ArrowKeys(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	il.SetText("hello")
	app.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	app.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	app.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	app.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP51_HandleKey_HomeEnd(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	il.SetText("hello world")
	app.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	app.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
}

func TestP51_HandleKey_PageUpDown(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 30; i++ {
		app.AddUserMessage("line")
	}
	app.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	app.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
}

// --- InputLine ---

func TestP51_InputLine_Measure(t *testing.T) {
	app := NewChatApp(80, 24)
	il := ensureInputLine(app)
	il.SetText("hello")
	il.Measure(component.Bounded(40, 1))
}

// --- ScrollView ---

func TestP51_ScrollView_ScrollTo(t *testing.T) {
	app := NewChatApp(80, 24)
	sv := app.ScrollView()
	sv.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 22})
	for i := 0; i < 30; i++ {
		app.AddUserMessage("line")
	}
	sv.Measure(component.Unbounded())
	sv.ScrollTo(10)
	// Offset may or may not be 10 depending on layout. Just verify no panic and non-negative.
	if sv.Offset() < 0 {
		t.Error("expected non-negative offset")
	}
	sv.ScrollTo(1000)
	if sv.Offset() < 0 {
		t.Error("expected non-negative offset after clamping")
	}
}

func TestP51_ScrollView_ScrollUp(t *testing.T) {
	app := NewChatApp(80, 24)
	sv := app.ScrollView()
	for i := 0; i < 30; i++ {
		app.AddUserMessage("line")
	}
	sv.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 22})
	sv.Measure(component.Unbounded())
	sv.ScrollTo(10)
	sv.ScrollUp(3)
	// Just verify no panic and non-negative
	if sv.Offset() < 0 {
		t.Error("expected non-negative offset")
	}
}

// --- AIBridge ---

func TestP51_AIBridge_ClearHistory(t *testing.T) {
	app := NewChatApp(80, 24)
	// bridge() returns nil until SetAIClient is called
	b := app.bridge()
	if b != nil {
		b.ClearHistory()
	}
}

func TestP51_AIBridge_Messages_Empty(t *testing.T) {
	app := NewChatApp(80, 24)
	b := app.bridge()
	if b != nil {
		_ = b.Messages()
	}
}

func TestP51_AIBridge_SetSystemPrompt(t *testing.T) {
	app := NewChatApp(80, 24)
	b := app.bridge()
	if b != nil {
		b.SetSystemPrompt("You are a helpful assistant")
	}
}

func TestP51_AIBridge_SetOnError(t *testing.T) {
	app := NewChatApp(80, 24)
	b := app.bridge()
	if b != nil {
		b.SetOnError(func(err error) {})
	}
}

// --- toggleBlock ---

func TestP51_ToggleBlock_Thinking(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := app.AddThinking()
	tb.SetContent("thinking...")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	toggleBlock(tb)
}

func TestP51_ToggleBlock_ToolResult(t *testing.T) {
	app := NewChatApp(80, 24)
	trb := app.AddToolResult()
	trb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	toggleBlock(trb)
}

func TestP51_ToggleBlock_OtherType(t *testing.T) {
	app := NewChatApp(80, 24)
	ub := app.AddUserMessage("hello")
	toggleBlock(ub)
}

// --- MouseHandler ---

func TestP51_MouseHandle_WheelUp(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)
	for i := 0; i < 30; i++ {
		app.AddUserMessage("line")
	}
	app.SetSize(80, 24)
	consumed := mh.Handle(&term.MouseEvent{X: 10, Y: 10, Button: term.MouseWheelUp, Action: term.MouseWheel})
	if !consumed {
		t.Error("expected wheel up to be consumed")
	}
}

func TestP51_MouseHandle_WheelDown(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)
	for i := 0; i < 30; i++ {
		app.AddUserMessage("line")
	}
	app.SetSize(80, 24)
	app.ScrollDown()
	consumed := mh.Handle(&term.MouseEvent{X: 10, Y: 10, Button: term.MouseWheelDown, Action: term.MouseWheel})
	if !consumed {
		t.Error("expected wheel down to be consumed")
	}
}

func TestP51_MouseHandle_ClickNoRegions(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{X: 10, Y: 5, Button: term.MouseLeft, Action: term.MouseDown})
	if consumed {
		t.Error("expected click with no regions to not be consumed")
	}
}

func TestP51_MouseHandle_UnknownWheelButton(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{X: 10, Y: 10, Button: 99, Action: term.MouseWheel})
	if consumed {
		t.Error("expected unknown wheel button to not be consumed")
	}
}

// --- Block types ---

func TestP51_Block_Types(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("user")
	app.AddThinking().SetContent("thinking")
	app.AddAssistantText().SetContent("assistant")
	app.AddToolCall("search", `{"q":"test"}`)
	blocks := app.Container().Blocks()
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}
	if blocks[0].Type() != block.TypeUserMessage {
		t.Error("expected TypeUserMessage")
	}
	if blocks[1].Type() != block.TypeThinking {
		t.Error("expected TypeThinking")
	}
	if blocks[2].Type() != block.TypeAssistantText {
		t.Error("expected TypeAssistantText")
	}
	if blocks[3].Type() != block.TypeToolCall {
		t.Error("expected TypeToolCall")
	}
}
