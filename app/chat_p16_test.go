package app

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- StatusBar Integration ---

func TestP16_SetStatusBar(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)

	if app.StatusBar() == nil {
		t.Fatal("StatusBar should not be nil after SetStatusBar")
	}
	if app.statusBarHeight != 1 {
		t.Errorf("statusBarHeight = %d, want 1", app.statusBarHeight)
	}
}

func TestP16_SetStatusBar_Nil(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.SetStatusBar(nil)

	if app.StatusBar() != nil {
		t.Error("StatusBar should be nil after SetStatusBar(nil)")
	}
	if app.statusBarHeight != 0 {
		t.Errorf("statusBarHeight = %d, want 0", app.statusBarHeight)
	}
}

func TestP16_SetModel(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.SetModel("GPT-4")

	// StatusBar should have a model item
	if sb.ItemCount() == 0 {
		t.Error("StatusBar should have items after SetModel")
	}
}

func TestP16_SetModel_NoStatusBar(t *testing.T) {
	app := NewChatApp(80, 24)
	// Should not panic
	app.SetModel("GPT-4")
}

func TestP16_SetTokenRate(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.SetTokenRate(1500)

	if sb.ItemCount() == 0 {
		t.Error("StatusBar should have items after SetTokenRate")
	}
}

func TestP16_SetContextWindow(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.SetContextWindow(5000, 128000)

	if sb.ItemCount() == 0 {
		t.Error("StatusBar should have items after SetContextWindow")
	}
}

func TestP16_UpdateClock(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.UpdateClock()

	if sb.ItemCount() == 0 {
		t.Error("StatusBar should have items after UpdateClock")
	}
}

// --- TabBar Integration ---

func TestP16_SetTabBar(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	if app.TabBar() == nil {
		t.Fatal("TabBar should not be nil after SetTabBar")
	}
	if app.tabBarHeight != 1 {
		t.Errorf("tabBarHeight = %d, want 1", app.tabBarHeight)
	}
}

func TestP16_SetTabBar_Nil(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.SetTabBar(nil)

	if app.TabBar() != nil {
		t.Error("TabBar should be nil after SetTabBar(nil)")
	}
	if app.tabBarHeight != 0 {
		t.Errorf("tabBarHeight = %d, want 0", app.tabBarHeight)
	}
}

func TestP16_AddSession(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	idx := app.AddSession("Chat 1")
	if idx != 0 {
		t.Errorf("first AddSession index = %d, want 0", idx)
	}
	if app.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", app.SessionCount())
	}
}

func TestP16_AddSession_NoTabBar(t *testing.T) {
	app := NewChatApp(80, 24)
	idx := app.AddSession("Chat 1")
	if idx != -1 {
		t.Errorf("AddSession without TabBar = %d, want -1", idx)
	}
}

func TestP16_AddSession_Multiple(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	if app.SessionCount() != 3 {
		t.Errorf("SessionCount = %d, want 3", app.SessionCount())
	}
}

func TestP16_NextSession(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	app.NextSession()
	if app.ActiveSession() != 1 {
		t.Errorf("ActiveSession = %d, want 1", app.ActiveSession())
	}
}

func TestP16_PrevSession(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	app.PrevSession()
	if app.ActiveSession() != 2 {
		t.Errorf("ActiveSession = %d, want 2 (wrap around)", app.ActiveSession())
	}
}

func TestP16_SwitchSession(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	app.SwitchSession(2)
	if app.ActiveSession() != 2 {
		t.Errorf("ActiveSession = %d, want 2", app.ActiveSession())
	}
}

func TestP16_CloseSession(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("A")
	app.AddSession("B")

	app.CloseSession()
	if app.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1 after close", app.SessionCount())
	}
}

func TestP16_Sessions(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("Alpha")
	app.AddSession("Beta")

	sessions := app.Sessions()
	if len(sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].Name != "Alpha" {
		t.Errorf("sessions[0].Name = %q, want 'Alpha'", sessions[0].Name)
	}
	if sessions[1].Name != "Beta" {
		t.Errorf("sessions[1].Name = %q, want 'Beta'", sessions[1].Name)
	}
}

func TestP16_ActiveSessionName(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)

	app.AddSession("Alpha")
	app.AddSession("Beta")

	if app.ActiveSessionName() != "Alpha" {
		t.Errorf("ActiveSessionName = %q, want 'Alpha'", app.ActiveSessionName())
	}

	app.NextSession()
	if app.ActiveSessionName() != "Beta" {
		t.Errorf("ActiveSessionName after Next = %q, want 'Beta'", app.ActiveSessionName())
	}
}

func TestP16_ActiveSessionName_NoTabBar(t *testing.T) {
	app := NewChatApp(80, 24)
	if app.ActiveSessionName() != "" {
		t.Errorf("ActiveSessionName without TabBar = %q, want ''", app.ActiveSessionName())
	}
}

// --- SelectionManager Integration ---

func TestP16_SetSelectionManager(t *testing.T) {
	app := NewChatApp(80, 24)
	sm := NewSelectionManager()
	app.SetSelectionManager(sm)

	if app.SelectionManager() == nil {
		t.Fatal("SelectionManager should not be nil")
	}
}

func TestP16_HasSelection_NoSelection(t *testing.T) {
	app := NewChatApp(80, 24)
	sm := NewSelectionManager()
	app.SetSelectionManager(sm)

	if app.HasSelection() {
		t.Error("HasSelection should be false initially")
	}
}

func TestP16_HasSelection_NoMgr(t *testing.T) {
	app := NewChatApp(80, 24)
	if app.HasSelection() {
		t.Error("HasSelection should be false without SelectionManager")
	}
}

func TestP16_ClearSelection(t *testing.T) {
	app := NewChatApp(80, 24)
	sm := NewSelectionManager()
	app.SetSelectionManager(sm)

	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 0)
	sm.EndSelection()

	app.ClearSelection()
	if app.HasSelection() {
		t.Error("HasSelection should be false after ClearSelection")
	}
}

func TestP16_ClearSelection_NoMgr(t *testing.T) {
	app := NewChatApp(80, 24)
	// Should not panic
	app.ClearSelection()
}

// --- HandleMouseP16 ---

func TestP16_HandleMouseP16_NilMouse(t *testing.T) {
	app := NewChatApp(80, 24)
	// Should not panic
	result := app.HandleMouseP16(nil)
	if result {
		t.Error("HandleMouseP16(nil) should return false")
	}
}

func TestP16_HandleMouseP16_TabClick(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.AddSession("Tab1")
	app.AddSession("Tab2")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

	// Click on second tab area
	mouse := &term.MouseEvent{
		Action: term.MouseDown,
		Button: term.MouseLeft,
		X:      10,
		Y:      0,
	}
	app.HandleMouseP16(mouse)
	// Tab click should switch active
	if app.ActiveSession() != 0 && app.ActiveSession() != 1 {
		t.Errorf("ActiveSession = %d, expected 0 or 1", app.ActiveSession())
	}
}

func TestP16_HandleMouseP16_WheelScroll(t *testing.T) {
	app := NewChatApp(80, 24)

	mouse := &term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelDown,
		X:      10,
		Y:      10,
	}
	result := app.HandleMouseP16(mouse)
	if !result {
		t.Error("HandleMouseP16 with wheel should return true")
	}
}

// --- handleP16Keys ---

func TestP16_HandleP16Keys_AltBracket(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	// Alt+] → next session
	key := &term.KeyEvent{
		Modifiers: term.ModAlt,
		Rune:     ']',
	}
	result := app.handleP16Keys(key)
	if !result {
		t.Error("handleP16Keys with Alt+] should return true")
	}
	if app.ActiveSession() != 1 {
		t.Errorf("ActiveSession = %d, want 1 after Alt+]", app.ActiveSession())
	}
}

func TestP16_HandleP16Keys_AltDigit(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.AddSession("A")
	app.AddSession("B")
	app.AddSession("C")

	// Alt+3 → switch to session index 2
	key := &term.KeyEvent{
		Modifiers: term.ModAlt,
		Rune:     '3',
	}
	result := app.handleP16Keys(key)
	if !result {
		t.Error("handleP16Keys with Alt+3 should return true")
	}
	if app.ActiveSession() != 2 {
		t.Errorf("ActiveSession = %d, want 2 after Alt+3", app.ActiveSession())
	}
}

func TestP16_HandleP16Keys_NoTabBar(t *testing.T) {
	app := NewChatApp(80, 24)

	key := &term.KeyEvent{
		Modifiers: term.ModAlt,
		Rune:     ']',
	}
	result := app.handleP16Keys(key)
	if result {
		t.Error("handleP16Keys without TabBar should return false")
	}
}

func TestP16_HandleP16Keys_AltW(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.AddSession("A")
	app.AddSession("B")

	key := &term.KeyEvent{
		Modifiers: term.ModAlt,
		Rune:     'w',
	}
	result := app.handleP16Keys(key)
	if !result {
		t.Error("handleP16Keys with Alt+W should return true")
	}
	if app.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1 after Alt+W", app.SessionCount())
	}
}

// --- Concurrency ---

func TestP16_ConcurrentAccess(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	sm := NewSelectionManager()
	app.SetSelectionManager(sm)

	var wg sync.WaitGroup

	// Writers: add/switch sessions
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				app.AddSession("session")
				app.NextSession()
				app.PrevSession()
				app.ActiveSession()
				app.SessionCount()
			}
		}(i)
	}

	// Readers
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				app.ActiveSession()
				app.SessionCount()
				app.HasSelection()
				app.StatusBar()
				app.TabBar()
			}
		}()
	}

	// Status bar updates
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				app.SetModel("GPT-4")
				app.SetTokenRate(1000)
				app.UpdateClock()
			}
		}()
	}

	wg.Wait()
}

// --- RenderP16 ---

func TestP16_RenderP16_NoPanic(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)

	app.AddSession("Tab1")
	app.SetModel("GPT-4")

	buf := buffer.NewBuffer(80, 24)
	app.renderP16(buf, 80, 24)
}

func TestP16_RenderP16_NoComponents(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	// Should not panic even without any P15 components
	app.renderP16(buf, 80, 24)
}

// --- Integration ---

func TestP16_RealWorldScenario(t *testing.T) {
	app := NewChatApp(120, 40)

	// Attach StatusBar with AI agent info
	sb := component.NewStatusBar()
	app.SetStatusBar(sb)
	app.SetModel("GPT-4")
	app.SetTokenRate(0)
	app.SetContextWindow(0, 128000)
	app.UpdateClock()

	// Attach TabBar with multiple sessions
	tb := component.NewTabBar()
	app.SetTabBar(tb)
	app.AddSession("Research")
	app.AddSession("Coding")
	app.AddSession("Writing")

	// Attach SelectionManager
	sm := NewSelectionManager()
	app.SetSelectionManager(sm)

	// Verify setup
	if app.SessionCount() != 3 {
		t.Errorf("SessionCount = %d, want 3", app.SessionCount())
	}
	if app.ActiveSessionName() != "Research" {
		t.Errorf("ActiveSessionName = %q, want 'Research'", app.ActiveSessionName())
	}

	// Switch to Coding session
	app.NextSession()
	if app.ActiveSessionName() != "Coding" {
		t.Errorf("ActiveSessionName = %q, want 'Coding'", app.ActiveSessionName())
	}

	// Verify status bar has content
	if sb.ItemCount() == 0 {
		t.Error("StatusBar should have items")
	}

	// Simulate text selection
	sm.StartSelection(0, 5)
	sm.ExtendSelection(10, 5)
	sm.EndSelection()

	if !app.HasSelection() {
		t.Error("Should have selection after StartSelection+ExtendSelection+EndSelection")
	}

	// Clear selection
	app.ClearSelection()
	if app.HasSelection() {
		t.Error("Should not have selection after ClearSelection")
	}

	// Render everything
	buf := buffer.NewBuffer(120, 40)
	app.renderP16(buf, 120, 40)
}
