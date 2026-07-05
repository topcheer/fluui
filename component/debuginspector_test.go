package component

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestNewDebugInspector_Defaults(t *testing.T) {
	di := NewDebugInspector()
	if di.Visible() {
		t.Error("expected hidden by default")
	}
	if di.Mode() != InspectTree {
		t.Error("expected InspectTree mode by default")
	}
	if di.Title() != "Debug Inspector" {
		t.Errorf("expected 'Debug Inspector', got %q", di.Title())
	}
}

func TestDebugInspector_Show(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	if !di.Visible() {
		t.Error("expected visible after Show")
	}
}

func TestDebugInspector_Hide(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.Hide()
	if di.Visible() {
		t.Error("expected hidden after Hide")
	}
}

func TestDebugInspector_Toggle(t *testing.T) {
	di := NewDebugInspector()

	newState := di.Toggle()
	if !newState {
		t.Error("expected true after first toggle")
	}
	if !di.Visible() {
		t.Error("expected visible")
	}

	newState = di.Toggle()
	if newState {
		t.Error("expected false after second toggle")
	}
	if di.Visible() {
		t.Error("expected hidden")
	}
}

func TestDebugInspector_SetVisible(t *testing.T) {
	di := NewDebugInspector()
	di.SetVisible(true)
	if !di.Visible() {
		t.Error("expected visible")
	}
	di.SetVisible(false)
	if di.Visible() {
		t.Error("expected hidden")
	}
}

func TestDebugInspector_SetMode(t *testing.T) {
	di := NewDebugInspector()
	di.SetMode(InspectEvents)
	if di.Mode() != InspectEvents {
		t.Error("expected InspectEvents mode")
	}
	di.SetMode(InspectStats)
	if di.Mode() != InspectStats {
		t.Error("expected InspectStats mode")
	}
}

func TestDebugInspector_NextMode(t *testing.T) {
	di := NewDebugInspector()

	// Tree -> Events
	di.NextMode()
	if di.Mode() != InspectEvents {
		t.Error("expected InspectEvents after first NextMode")
	}

	// Events -> Stats
	di.NextMode()
	if di.Mode() != InspectStats {
		t.Error("expected InspectStats after second NextMode")
	}

	// Stats -> Tree (wrap)
	di.NextMode()
	if di.Mode() != InspectTree {
		t.Error("expected InspectTree after third NextMode")
	}
}

func TestDebugInspector_SetRoot(t *testing.T) {
	di := NewDebugInspector()
	child := newStubChild(10, 5, "test")
	di.SetRoot(child)
	// Just verify no panic
}

func TestDebugInspector_SetPanelSize(t *testing.T) {
	di := NewDebugInspector()
	di.SetPanelSize(60, 25)
	sz := di.Measure(Bounded(80, 30))
	if sz.W != 60 {
		t.Errorf("expected width 60, got %d", sz.W)
	}
	if sz.H != 25 {
		t.Errorf("expected height 25, got %d", sz.H)
	}
}

func TestDebugInspector_SetTitle(t *testing.T) {
	di := NewDebugInspector()
	di.SetTitle("Custom Inspector")
	if di.Title() != "Custom Inspector" {
		t.Errorf("expected 'Custom Inspector', got %q", di.Title())
	}
}

func TestDebugInspector_RecordKey(t *testing.T) {
	di := NewDebugInspector()
	di.RecordKey(&term.KeyEvent{Key: term.KeyEnter})
	di.RecordKey(&term.KeyEvent{Rune: 'a'})

	events := di.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventKey {
		t.Error("expected EventKey type")
	}
	if events[1].Type != EventKey {
		t.Error("expected EventKey type")
	}
}

func TestDebugInspector_RecordKey_Nil(t *testing.T) {
	di := NewDebugInspector()
	di.RecordKey(nil)
	if len(di.Events()) != 0 {
		t.Error("expected 0 events for nil key")
	}
}

func TestDebugInspector_RecordMouse(t *testing.T) {
	di := NewDebugInspector()
	di.RecordMouse(&term.MouseEvent{
		X:      10,
		Y:      5,
		Button: term.MouseLeft,
		Action: term.MouseDown,
	})

	events := di.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventMouse {
		t.Error("expected EventMouse type")
	}
}

func TestDebugInspector_RecordMouse_Nil(t *testing.T) {
	di := NewDebugInspector()
	di.RecordMouse(nil)
	if len(di.Events()) != 0 {
		t.Error("expected 0 events for nil mouse")
	}
}

func TestDebugInspector_RecordResize(t *testing.T) {
	di := NewDebugInspector()
	di.RecordResize(80, 24)

	events := di.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventResize {
		t.Error("expected EventResize type")
	}
}

func TestDebugInspector_RecordCustom(t *testing.T) {
	di := NewDebugInspector()
	di.RecordCustom("AI response started")

	events := di.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventCustom {
		t.Error("expected EventCustom type")
	}
	if events[0].Summary != "AI response started" {
		t.Errorf("expected 'AI response started', got %q", events[0].Summary)
	}
}

func TestDebugInspector_EventsRingBuffer(t *testing.T) {
	di := NewDebugInspector()
	di.maxEvents = 5

	for i := 0; i < 10; i++ {
		di.RecordCustom("event")
	}

	events := di.Events()
	if len(events) != 5 {
		t.Errorf("expected 5 events (ring buffer cap), got %d", len(events))
	}
}

func TestDebugInspector_Events_DefensiveCopy(t *testing.T) {
	di := NewDebugInspector()
	di.RecordCustom("test")

	events1 := di.Events()
	if len(events1) > 0 {
		events1[0].Summary = "modified"
	}

	events2 := di.Events()
	if events2[0].Summary == "modified" {
		t.Error("expected defensive copy to prevent mutation")
	}
}

func TestDebugInspector_ClearEvents(t *testing.T) {
	di := NewDebugInspector()
	di.RecordCustom("a")
	di.RecordCustom("b")
	di.ClearEvents()
	if len(di.Events()) != 0 {
		t.Error("expected 0 events after clear")
	}
}

func TestDebugInspector_RecordRender(t *testing.T) {
	di := NewDebugInspector()
	di.RecordRender(5*time.Millisecond, 1000, true)
	di.RecordRender(3*time.Millisecond, 800, false)
	di.RecordRender(4*time.Millisecond, 1200, true)

	stats := di.Stats()
	if stats.FrameCount != 3 {
		t.Errorf("expected 3 frames, got %d", stats.FrameCount)
	}
	if stats.TotalCells != 3000 {
		t.Errorf("expected 3000 total cells, got %d", stats.TotalCells)
	}
	if stats.DirtyCount != 2 {
		t.Errorf("expected 2 dirty frames, got %d", stats.DirtyCount)
	}
	if stats.LastRenderNs != 4*int64(time.Millisecond) {
		t.Errorf("unexpected last render ns: %d", stats.LastRenderNs)
	}
}

func TestDebugInspector_ResetStats(t *testing.T) {
	di := NewDebugInspector()
	di.RecordRender(1*time.Millisecond, 100, true)
	di.ResetStats()

	stats := di.Stats()
	if stats.FrameCount != 0 {
		t.Error("expected 0 frames after reset")
	}
}

func TestDebugInspector_Stats_DefensiveCopy(t *testing.T) {
	di := NewDebugInspector()
	di.RecordRender(1*time.Millisecond, 100, true)

	s1 := di.Stats()
	s1.FrameCount = 999

	s2 := di.Stats()
	if s2.FrameCount == 999 {
		t.Error("expected defensive copy")
	}
}

func TestDebugInspector_ScrollUp(t *testing.T) {
	di := NewDebugInspector()
	// Add many events
	for i := 0; i < 50; i++ {
		di.RecordCustom("event")
	}
	di.ScrollDown(20)
	di.ScrollUp(5)
	// Just verify no panic
}

func TestDebugInspector_ScrollDown(t *testing.T) {
	di := NewDebugInspector()
	di.ScrollDown(10)
	// Just verify no panic
}

func TestDebugInspector_ScrollUp_Clamp(t *testing.T) {
	di := NewDebugInspector()
	di.ScrollDown(10)
	di.ScrollUp(100) // try to scroll way past start
	// Should clamp to 0, no panic
}

func TestDebugInspector_HandleKey_Esc(t *testing.T) {
	di := NewDebugInspector()
	di.Show()

	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !handled {
		t.Error("expected Esc to be handled")
	}
	if di.Visible() {
		t.Error("expected hidden after Esc")
	}
}

func TestDebugInspector_HandleKey_Tab(t *testing.T) {
	di := NewDebugInspector()
	di.Show()

	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !handled {
		t.Error("expected Tab to be handled")
	}
	if di.Mode() != InspectEvents {
		t.Error("expected mode change to Events")
	}
}

func TestDebugInspector_HandleKey_Arrows(t *testing.T) {
	di := NewDebugInspector()
	di.Show()

	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if !handled {
		t.Error("expected Up to be handled")
	}

	handled = di.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !handled {
		t.Error("expected Down to be handled")
	}
}

func TestDebugInspector_HandleKey_PageUp(t *testing.T) {
	di := NewDebugInspector()
	di.Show()

	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if !handled {
		t.Error("expected PageUp to be handled")
	}
}

func TestDebugInspector_HandleKey_PageDown(t *testing.T) {
	di := NewDebugInspector()
	di.Show()

	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	if !handled {
		t.Error("expected PageDown to be handled")
	}
}

func TestDebugInspector_HandleKey_NotVisible(t *testing.T) {
	di := NewDebugInspector()
	// Not visible, should not handle any key
	handled := di.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if handled {
		t.Error("expected no handling when not visible")
	}
}

func TestDebugInspector_HandleKey_Nil(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	handled := di.HandleKey(nil)
	if handled {
		t.Error("expected nil key to not be handled")
	}
}

func TestDebugInspector_HandleKey_Unhandled(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	handled := di.HandleKey(&term.KeyEvent{Rune: 'x'})
	if handled {
		t.Error("expected 'x' key to not be handled")
	}
}

func TestDebugInspector_Measure(t *testing.T) {
	di := NewDebugInspector()
	di.SetPanelSize(40, 15)
	sz := di.Measure(Bounded(80, 24))
	if sz.W != 40 || sz.H != 15 {
		t.Errorf("expected 40x15, got %dx%d", sz.W, sz.H)
	}
}

func TestDebugInspector_Measure_Clamped(t *testing.T) {
	di := NewDebugInspector()
	di.SetPanelSize(100, 50)
	sz := di.Measure(Bounded(60, 30))
	if sz.W != 60 {
		t.Errorf("expected width clamped to 60, got %d", sz.W)
	}
	if sz.H != 30 {
		t.Errorf("expected height clamped to 30, got %d", sz.H)
	}
}

func TestDebugInspector_SetBounds(t *testing.T) {
	di := NewDebugInspector()
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	b := di.Bounds()
	if b.W != 80 || b.H != 24 {
		t.Errorf("unexpected bounds: %+v", b)
	}
}

func TestDebugInspector_Paint_Hidden(t *testing.T) {
	di := NewDebugInspector()
	// Not visible
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should not draw anything

	// Verify no content drawn (all blank)
	if buffer_CountNonEmpty(buf) > 0 {
		t.Error("expected blank buffer when hidden")
	}
}

func TestDebugInspector_Paint_Visible(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf)

	// Should have drawn content
	if buffer_CountNonEmpty(buf) == 0 {
		t.Error("expected non-empty buffer when visible")
	}
}

func TestDebugInspector_Paint_TreeMode(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetMode(InspectTree)
	di.SetRoot(newStubChild(10, 5, "root"))
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should not panic
}

func TestDebugInspector_Paint_EventsMode(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetMode(InspectEvents)
	di.RecordKey(&term.KeyEvent{Key: term.KeyEnter})
	di.RecordMouse(&term.MouseEvent{X: 5, Y: 10, Button: term.MouseLeft})
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should not panic
}

func TestDebugInspector_Paint_StatsMode(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetMode(InspectStats)
	di.RecordRender(5*time.Millisecond, 1000, true)
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should not panic
}

func TestDebugInspector_Paint_TreeNoRoot(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetMode(InspectTree)
	// No root set
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should show "(no root)" message
}

func TestDebugInspector_Paint_EventsEmpty(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetMode(InspectEvents)
	// No events recorded
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	di.Paint(buf) // should show "(no events recorded)"
}

func TestDebugInspector_Paint_TinyBounds(t *testing.T) {
	di := NewDebugInspector()
	di.Show()
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3}) // too small

	buf := buffer.NewBuffer(5, 3)
	di.Paint(buf) // should not panic, may skip drawing
}

func TestDebugInspector_Children(t *testing.T) {
	di := NewDebugInspector()
	if di.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestDebugInspector_Concurrent(t *testing.T) {
	di := NewDebugInspector()
	di.SetPanelSize(40, 15)
	di.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	var wg sync.WaitGroup

	// Concurrent event recorders
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				di.RecordKey(&term.KeyEvent{Rune: 'a'})
				di.RecordRender(1*time.Millisecond, 100, true)
				di.Toggle()
			}
		}(i)
	}

	// Concurrent painters
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				buf := buffer.NewBuffer(80, 24)
				di.Paint(buf)
				di.Events()
				di.Stats()
			}
		}()
	}

	wg.Wait()
}

func TestDebugInspector_KeyEventSummary(t *testing.T) {
	tests := []struct {
		name     string
		key      *term.KeyEvent
		contains string
	}{
		{"plain rune", &term.KeyEvent{Rune: 'a'}, "a"},
		{"ctrl+enter", &term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModCtrl}, "Ctrl"},
		{"alt+x", &term.KeyEvent{Rune: 'x', Modifiers: term.ModAlt}, "Alt"},
		{"shift+up", &term.KeyEvent{Key: term.KeyUp, Modifiers: term.ModShift}, "Shift"},
	}
	for _, tc := range tests {
		got := keyEventSummary(tc.key)
		if !containsSubstr(got, tc.contains) {
			t.Errorf("%s: expected %q in %q", tc.name, tc.contains, got)
		}
	}
}

func TestDebugInspector_KeyName(t *testing.T) {
	if keyName(term.KeyUp) != "Up" {
		t.Error("expected 'Up'")
	}
	if keyName(term.KeyEnter) != "Enter" {
		t.Error("expected 'Enter'")
	}
	if keyName(term.KeyF12) != "F12" {
		t.Error("expected 'F12'")
	}
}

func TestDebugInspector_Truncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error("expected no truncation")
	}
	if truncate("hello world", 8) != "hello w…" {
		t.Errorf("expected 'hello w…', got %q", truncate("hello world", 8))
	}
	if truncate("ab", 2) != "ab" {
		t.Error("expected no truncation for exact fit")
	}
	if truncate("abc", 1) != "…" {
		t.Errorf("expected '…', got %q", truncate("abc", 1))
	}
	if truncate("anything", 0) != "" {
		t.Error("expected empty for maxLen=0")
	}
}

// --- helpers ---

func buffer_CountNonEmpty(buf *buffer.Buffer) int {
	count := 0
	for _, c := range buf.Cells {
		if c.Rune != 0 && c.Rune != ' ' {
			count++
		}
	}
	return count
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
