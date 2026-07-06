package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === CompletionManager coverage (66-75% → 90%+) ===

func TestP61_Completion_SelectedInactive(t *testing.T) {
	cm := NewCompletionManager(nil)
	_, ok := cm.Selected()
	if ok {
		t.Error("expected false when not active")
	}
}

func TestP61_Completion_CycleNextInactive(t *testing.T) {
	cm := NewCompletionManager(nil)
	_, ok := cm.CycleNext()
	if ok {
		t.Error("expected false when not active")
	}
}

func TestP61_Completion_CyclePrevInactive(t *testing.T) {
	cm := NewCompletionManager(nil)
	_, ok := cm.CyclePrev()
	if ok {
		t.Error("expected false when not active")
	}
}

func TestP61_Completion_AcceptInactive(t *testing.T) {
	cm := NewCompletionManager(nil)
	_, ok := cm.Accept()
	if ok {
		t.Error("expected false when not active")
	}
}

func TestP61_Completion_StartNilProvider(t *testing.T) {
	cm := NewCompletionManager(nil)
	ok := cm.Start("/test")
	if ok {
		t.Error("expected false when provider is nil")
	}
}

func TestP61_Completion_CycleNextWrap(t *testing.T) {
	p := NewSlashCommandProvider()
	p.AddCommand("zzz_custom", "Custom cmd")
	cm := NewCompletionManager(p)
	cm.Start("/")
	items := cm.Items()
	n := len(items)
	if n == 0 {
		t.Fatal("expected at least 1 item")
	}
	// Cycle through ALL items to verify wrap
	for i := 0; i < n; i++ {
		cm.CycleNext()
	}
	// After n cycles we should be back at 0
	if cm.SelectedIndex() != 0 {
		t.Errorf("expected wrap to 0 after %d cycles, got %d", n, cm.SelectedIndex())
	}
}

func TestP61_Completion_CyclePrevWrap(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)
	cm.Start("/")
	items := cm.Items()
	n := len(items)
	if n == 0 {
		t.Fatal("expected at least 1 item")
	}
	// Cycle prev from 0 should wrap to last item
	cm.CyclePrev()
	if cm.SelectedIndex() != n-1 {
		t.Errorf("expected wrap to %d, got %d", n-1, cm.SelectedIndex())
	}
}

func TestP61_Completion_AcceptWithItem(t *testing.T) {
	p := NewSlashCommandProvider()
	p.AddCommand("test", "Test cmd")
	cm := NewCompletionManager(p)
	cm.Start("/t")
	item, ok := cm.Accept()
	if !ok {
		t.Error("expected accept to return true")
	}
	if item.Insert == "" {
		t.Error("expected non-empty insert text")
	}
}

func TestP61_Completion_Cancel(t *testing.T) {
	p := NewSlashCommandProvider()
	p.AddCommand("test", "Test cmd")
	cm := NewCompletionManager(p)
	cm.Start("/t")
	cm.Cancel()
	if cm.Active() {
		t.Error("expected inactive after cancel")
	}
	if len(cm.Items()) != 0 {
		t.Error("expected empty items after cancel")
	}
}

func TestP61_Completion_SetProvider(t *testing.T) {
	cm := NewCompletionManager(nil)
	p := NewSlashCommandProvider()
	p.AddCommand("x", "X")
	cm.SetProvider(p)
	if cm.Provider() == nil {
		t.Error("expected non-nil provider")
	}
}

func TestP61_Completion_ReplacePrefixNoMatch(t *testing.T) {
	text, cursor := ReplacePrefix("hello", 5, "world", "inserted")
	if text != "hello" {
		t.Errorf("expected unchanged when prefix doesn't match, got %q", text)
	}
	if cursor != 5 {
		t.Errorf("expected cursor 5, got %d", cursor)
	}
}

func TestP61_Completion_ExtractPrefixCursorZero(t *testing.T) {
	result := ExtractCompletionPrefix("hello", 0)
	if result != "" {
		t.Errorf("expected empty for cursor=0, got %q", result)
	}
}

// === Undo/Redo with nil undo (66.7% → 90%+) ===

func TestP61_Undo_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.Undo() {
		t.Error("expected false for nil undo stack")
	}
}

func TestP61_Redo_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.Redo() {
		t.Error("expected false for nil undo stack")
	}
}

func TestP61_CanUndo_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.CanUndo() {
		t.Error("expected false for nil undo stack")
	}
}

func TestP61_CanRedo_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.CanRedo() {
		t.Error("expected false for nil undo stack")
	}
}

func TestP61_UndoCount_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.UndoCount() != 0 {
		t.Error("expected 0 for nil undo stack")
	}
}

func TestP61_RedoCount_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	if il.RedoCount() != 0 {
		t.Error("expected 0 for nil undo stack")
	}
}

func TestP61_ClearUndoHistory_NilUndoStack(t *testing.T) {
	il := &InputLine{}
	il.ClearUndoHistory() // should not panic
}

// === Search coverage (65-67% → 85%+) ===

func TestP61_Search_NextMatchEmpty(t *testing.T) {
	sm := &SearchMode{}
	sm.NextMatch() // no matches — should not panic
	if sm.CurrentIndex() != 0 {
		t.Error("expected 0 with empty matches")
	}
}

func TestP61_Search_PrevMatchEmpty(t *testing.T) {
	sm := &SearchMode{}
	sm.PrevMatch() // no matches — should not panic
	if sm.CurrentIndex() != 0 {
		t.Error("expected 0 with empty matches")
	}
}

func TestP61_Search_NextMatchWrap(t *testing.T) {
	sm := &SearchMode{
		matches:  []SearchMatch{{BlockID: "1"}, {BlockID: "2"}, {BlockID: "3"}},
		current:  2,
	}
	sm.NextMatch()
	if sm.current != 0 {
		t.Errorf("expected wrap to 0, got %d", sm.current)
	}
}

func TestP61_Search_PrevMatchFromStart(t *testing.T) {
	sm := &SearchMode{
		matches:  []SearchMatch{{BlockID: "1"}, {BlockID: "2"}, {BlockID: "3"}},
		current:  0,
	}
	sm.PrevMatch()
	if sm.current != 2 {
		t.Errorf("expected wrap to 2, got %d", sm.current)
	}
}

func TestP61_Search_RenderSearchBarInactive(t *testing.T) {
	sm := &SearchMode{}
	buf := buffer.NewBuffer(80, 24)
	sm.RenderSearchBar(buf, 80, 23)
	// Should not crash when inactive
}

func TestP61_Search_RenderSearchBarWithMatches(t *testing.T) {
	sm := &SearchMode{
		active: true,
		query:  "test",
		matches: []SearchMatch{{BlockID: "1"}, {BlockID: "2"}},
	}
	buf := buffer.NewBuffer(80, 24)
	sm.RenderSearchBar(buf, 80, 23)
	// Should render with match count
}

func TestP61_Search_RenderSearchBarNoMatches(t *testing.T) {
	sm := &SearchMode{
		active: true,
		query:  "nomatch",
	}
	buf := buffer.NewBuffer(80, 24)
	sm.RenderSearchBar(buf, 80, 23)
	// Should render with "no matches" style
}

func TestP61_Search_RenderSearchBarEmpty(t *testing.T) {
	sm := &SearchMode{
		active: true,
		query:  "",
	}
	buf := buffer.NewBuffer(80, 24)
	sm.RenderSearchBar(buf, 80, 23)
	// Should render empty prompt
}

func TestP61_Search_RenderSearchBarWidthZero(t *testing.T) {
	sm := &SearchMode{
		active: true,
		query:  "test",
	}
	buf := buffer.NewBuffer(80, 24)
	sm.RenderSearchBar(buf, 0, 23)
	// Should not crash with zero width
}

func TestP61_Search_HandleKeyCtrlShiftF(t *testing.T) {
	sm := &SearchMode{
		active:  true,
		query:   "x",
		matches: []SearchMatch{{BlockID: "1"}},
	}
	// Ctrl+Shift+F should go to previous match
	key := &term.KeyEvent{Rune: 'F', Modifiers: term.ModCtrl | term.ModShift}
	consumed := sm.HandleKey(key)
	if !consumed {
		t.Error("expected Ctrl+Shift+F consumed")
	}
}

func TestP61_Search_HandleKeyCtrlF(t *testing.T) {
	sm := &SearchMode{
		active:  true,
		query:   "x",
		matches: []SearchMatch{{BlockID: "1"}},
	}
	key := &term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl}
	consumed := sm.HandleKey(key)
	if !consumed {
		t.Error("expected Ctrl+F consumed")
	}
}

func TestP61_Search_HandleKeyBackspaceEmpty(t *testing.T) {
	sm := &SearchMode{active: true, query: ""}
	key := &term.KeyEvent{Key: term.KeyBackspace}
	sm.HandleKey(key)
	if sm.query != "" {
		t.Error("expected empty query after backspace")
	}
}

func TestP61_Search_StatusTextEmpty(t *testing.T) {
	sm := &SearchMode{}
	if sm.StatusText() != "" {
		t.Error("expected empty status for inactive")
	}
}

func TestP61_Search_StatusTextNoMatches(t *testing.T) {
	sm := &SearchMode{active: true, query: "test"}
	status := sm.StatusText()
	if status == "" {
		t.Error("expected non-empty status")
	}
}

// === MouseHandler coverage (44.8% → 70%+) ===

func TestP61_Mouse_ScrollbarDown(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	// Test mouse on scrollbar column
	barX := app.ScrollView().ScrollbarColumn()
	if barX < 0 {
		t.Skip("no scrollbar visible")
	}
	mouse := &term.MouseEvent{
		X:      barX,
		Y:      5,
		Action: term.MouseDown,
		Button: term.MouseLeft,
	}
	mh := NewMouseHandler(app)
	consumed := mh.Handle(mouse)
	if !consumed {
		t.Error("expected scrollbar click consumed")
	}
}

func TestP61_Mouse_ScrollbarDrag(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	barX := app.ScrollView().ScrollbarColumn()
	if barX < 0 {
		t.Skip("no scrollbar visible")
	}
	mh := NewMouseHandler(app)
	// MouseDown first
	mh.Handle(&term.MouseEvent{
		X: barX, Y: 5, Action: term.MouseDown, Button: term.MouseLeft,
	})
	// Then drag
	consumed := mh.Handle(&term.MouseEvent{
		X: barX, Y: 10, Action: term.MouseDrag, Button: term.MouseLeft,
	})
	if !consumed {
		t.Error("expected scrollbar drag consumed")
	}
}

func TestP61_Mouse_ScrollbarUp(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	barX := app.ScrollView().ScrollbarColumn()
	if barX < 0 {
		t.Skip("no scrollbar visible")
	}
	mh := NewMouseHandler(app)
	mh.Handle(&term.MouseEvent{
		X: barX, Y: 5, Action: term.MouseDown, Button: term.MouseLeft,
	})
	consumed := mh.Handle(&term.MouseEvent{
		X: barX, Y: 8, Action: term.MouseUp, Button: term.MouseLeft,
	})
	if !consumed {
		t.Error("expected scrollbar up consumed")
	}
}

func TestP61_Mouse_DragOutsideColumn(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	barX := app.ScrollView().ScrollbarColumn()
	if barX < 0 {
		t.Skip("no scrollbar visible")
	}
	mh := NewMouseHandler(app)
	// Start drag in scrollbar
	mh.Handle(&term.MouseEvent{
		X: barX, Y: 5, Action: term.MouseDown, Button: term.MouseLeft,
	})
	// Drag outside scrollbar column
	consumed := mh.Handle(&term.MouseEvent{
		X: 0, Y: 10, Action: term.MouseDrag, Button: term.MouseLeft,
	})
	if !consumed {
		t.Error("expected drag outside column consumed when dragging")
	}
}

func TestP61_Mouse_ReleaseDrag(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	barX := app.ScrollView().ScrollbarColumn()
	if barX < 0 {
		t.Skip("no scrollbar visible")
	}
	mh := NewMouseHandler(app)
	// Start drag
	mh.Handle(&term.MouseEvent{
		X: barX, Y: 5, Action: term.MouseDown, Button: term.MouseLeft,
	})
	// Release outside scrollbar column
	consumed := mh.Handle(&term.MouseEvent{
		X: 0, Y: 0, Action: term.MouseUp, Button: term.MouseLeft,
	})
	if !consumed {
		t.Error("expected mouse up consumed when dragging")
	}
}

func TestP61_Mouse_WheelUp(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{
		X: 5, Y: 5, Action: term.MouseWheel, Button: term.MouseWheelUp,
	})
	if !consumed {
		t.Error("expected wheel up consumed")
	}
}

func TestP61_Mouse_WheelDown(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{
		X: 5, Y: 5, Action: term.MouseWheel, Button: term.MouseWheelDown,
	})
	if !consumed {
		t.Error("expected wheel down consumed")
	}
}

func TestP61_Mouse_ClickNoRegions(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{
		X: 5, Y: 5, Action: term.MouseDown, Button: term.MouseLeft,
	})
	// No regions → not consumed
	if consumed {
		t.Error("expected not consumed with no regions")
	}
}

func TestP61_Mouse_UnknownButton(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputLine(NewInputLine("> "))
	app.SetSize(80, 24)

	mh := NewMouseHandler(app)
	consumed := mh.Handle(&term.MouseEvent{
		X: 5, Y: 5, Action: term.MouseWheel, Button: 99,
	})
	if consumed {
		t.Error("expected not consumed for unknown wheel button")
	}
}

// === toggleBlock ===

func TestP61_ToggleBlock_ThinkingBlock(t *testing.T) {
	tb := block.NewThinkingBlock("test")
	tb.Toggle()
	toggleBlock(tb) // should toggle again without panic
}

func TestP61_ToggleBlock_ToolResultBlock(t *testing.T) {
	tr := block.NewToolResultBlock("test")
	tr.Toggle()
	toggleBlock(tr) // should toggle again without panic
}

func TestP61_ToggleBlock_OtherType(t *testing.T) {
	atb := block.NewAssistantTextBlock("test")
	toggleBlock(atb) // should not panic for unsupported type
}
