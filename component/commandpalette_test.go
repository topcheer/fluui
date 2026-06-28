package component

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ============================================================
// P19-B: CommandPalette Component Tests
// ============================================================

func newTestCommands(n int) []Command {
	cmds := make([]Command, n)
	for i := 0; i < n; i++ {
		cmds[i] = Command{
			ID:    fmt.Sprintf("cmd-%d", i),
			Label: fmt.Sprintf("Command %d", i),
		}
	}
	return cmds
}

// ─── Construction ────────────────────────────────────────────────

func TestCommandPalette_New(t *testing.T) {
	cp := NewCommandPalette()
	if cp == nil {
		t.Fatal("NewCommandPalette returned nil")
	}
	if cp.ID() == "" {
		t.Error("ID should not be empty")
	}
	if cp.Visible() {
		t.Error("should start hidden")
	}
	if cp.CommandCount() != 0 {
		t.Error("should start with no commands")
	}
	if cp.FilteredCount() != 0 {
		t.Error("should start with no filtered results")
	}
	if cp.MaxVisible() != 10 {
		t.Errorf("MaxVisible = %d, want 10", cp.MaxVisible())
	}
}

func TestCommandPalette_UniqueID(t *testing.T) {
	cp1 := NewCommandPalette()
	cp2 := NewCommandPalette()
	if cp1.ID() == cp2.ID() {
		t.Error("IDs should be unique")
	}
}

func TestCommandPalette_ImplementsComponent(t *testing.T) {
	var _ Component = NewCommandPalette()
}

// ─── Commands ────────────────────────────────────────────────────

func TestCommandPalette_SetCommands(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	if cp.CommandCount() != 5 {
		t.Errorf("CommandCount = %d, want 5", cp.CommandCount())
	}
	if cp.FilteredCount() != 5 {
		t.Errorf("FilteredCount = %d, want 5 (empty query shows all)", cp.FilteredCount())
	}
}

func TestCommandPalette_AddCommand(t *testing.T) {
	cp := NewCommandPalette()
	cp.AddCommand(Command{ID: "save", Label: "Save File"})
	cp.AddCommand(Command{ID: "open", Label: "Open File"})
	if cp.CommandCount() != 2 {
		t.Errorf("CommandCount = %d, want 2", cp.CommandCount())
	}
}

func TestCommandPalette_Commands_ReturnsCopy(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cmds := cp.Commands()
	cmds[0].Label = "mutated"
	if cp.Commands()[0].Label == "mutated" {
		t.Error("Commands() should return a copy")
	}
}

// ─── Query & filtering ───────────────────────────────────────────

func TestCommandPalette_SetQuery_FuzzyMatch(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{ID: "save", Label: "Save File"},
		{ID: "search", Label: "Search in Files"},
		{ID: "close", Label: "Close Tab"},
		{ID: "settings", Label: "Settings"},
	})
	cp.SetQuery("sf")
	if cp.FilteredCount() == 0 {
		t.Error("expected at least 1 result for 'sf' (Save File)")
	}
}

func TestCommandPalette_SetQuery_NoMatch(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{ID: "save", Label: "Save File"},
	})
	cp.SetQuery("xyz")
	if cp.HasResults() {
		t.Error("expected no results for 'xyz'")
	}
}

func TestCommandPalette_SetQuery_EmptyMatchesAll(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(10))
	cp.SetQuery("")
	if cp.FilteredCount() != 10 {
		t.Errorf("FilteredCount = %d, want 10", cp.FilteredCount())
	}
}

func TestCommandPalette_InsertRune(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{ID: "save", Label: "Save File"},
	})
	cp.InsertRune('s')
	cp.InsertRune('f')
	if cp.Query() != "sf" {
		t.Errorf("Query = %q, want 'sf'", cp.Query())
	}
}

func TestCommandPalette_Backspace(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetQuery("hello")
	cp.Backspace()
	if cp.Query() != "hell" {
		t.Errorf("Query = %q, want 'hell'", cp.Query())
	}
}

func TestCommandPalette_Backspace_Empty(t *testing.T) {
	cp := NewCommandPalette()
	cp.Backspace() // should not panic
	if cp.Query() != "" {
		t.Error("Query should remain empty")
	}
}

func TestCommandPalette_Reset(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.SetQuery("test")
	cp.Reset()
	if cp.Query() != "" {
		t.Errorf("Query = %q, want ''", cp.Query())
	}
	if cp.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", cp.Cursor())
	}
}

// ─── Cursor navigation ───────────────────────────────────────────

func TestCommandPalette_MoveDown(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	if cp.Cursor() != 0 {
		t.Errorf("initial cursor = %d", cp.Cursor())
	}
	cp.MoveDown()
	if cp.Cursor() != 1 {
		t.Errorf("after MoveDown cursor = %d, want 1", cp.Cursor())
	}
}

func TestCommandPalette_MoveDown_WrapAround(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cp.SetCursor(2)
	cp.MoveDown()
	if cp.Cursor() != 0 {
		t.Errorf("after wrap cursor = %d, want 0", cp.Cursor())
	}
}

func TestCommandPalette_MoveUp(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.SetCursor(2)
	cp.MoveUp()
	if cp.Cursor() != 1 {
		t.Errorf("after MoveUp cursor = %d, want 1", cp.Cursor())
	}
}

func TestCommandPalette_MoveUp_WrapAround(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cp.MoveUp() // from 0, should wrap to 2
	if cp.Cursor() != 2 {
		t.Errorf("after wrap cursor = %d, want 2", cp.Cursor())
	}
}

func TestCommandPalette_SetCursor_Wrap(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	// Negative wraps to bottom
	cp.SetCursor(-1)
	if cp.Cursor() != 2 {
		t.Errorf("cursor = %d after SetCursor(-1), want 2 (wrap to bottom)", cp.Cursor())
	}
	// Positive overflow wraps to top
	cp.SetCursor(10)
	if cp.Cursor() != 0 {
		t.Errorf("cursor = %d after SetCursor(10), want 0 (wrap to top)", cp.Cursor())
	}
	// In-range sets directly
	cp.SetCursor(2)
	if cp.Cursor() != 2 {
		t.Errorf("cursor = %d after SetCursor(2), want 2", cp.Cursor())
	}
}

func TestCommandPalette_CurrentCommand(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{ID: "first", Label: "First"},
		{ID: "second", Label: "Second"},
	})
	cmd := cp.CurrentCommand()
	if cmd == nil {
		t.Fatal("CurrentCommand should not be nil")
	}
	if cmd.ID != "first" {
		t.Errorf("CurrentCommand.ID = %q, want 'first'", cmd.ID)
	}
}

func TestCommandPalette_CurrentCommand_Empty(t *testing.T) {
	cp := NewCommandPalette()
	if cp.CurrentCommand() != nil {
		t.Error("CurrentCommand should be nil when empty")
	}
}

// ─── Visibility ──────────────────────────────────────────────────

func TestCommandPalette_ShowHide(t *testing.T) {
	cp := NewCommandPalette()
	if cp.Visible() {
		t.Error("should start hidden")
	}
	cp.Show(10, 20)
	if !cp.Visible() {
		t.Error("should be visible after Show")
	}
	x, y := cp.Position()
	if x != 10 || y != 20 {
		t.Errorf("Position = (%d,%d), want (10,20)", x, y)
	}
	cp.Hide()
	if cp.Visible() {
		t.Error("should be hidden after Hide")
	}
}

func TestCommandPalette_Show_ResetsQuery(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cp.SetQuery("old query")
	cp.Show(0, 0)
	if cp.Query() != "" {
		t.Errorf("Query = %q after Show, want ''", cp.Query())
	}
}

func TestCommandPalette_SetPosition(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetPosition(5, 15)
	x, y := cp.Position()
	if x != 5 || y != 15 {
		t.Errorf("Position = (%d,%d), want (5,15)", x, y)
	}
}

func TestCommandPalette_Hide_OnDismiss(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cp.Show(0, 0)

	var dismissed bool
	cp.SetOnDismiss(func() {
		dismissed = true
	})
	cp.Hide()
	if !dismissed {
		t.Error("OnDismiss should have been called")
	}
}

// ─── Selection ───────────────────────────────────────────────────

func TestCommandPalette_Select(t *testing.T) {
	cp := NewCommandPalette()
	executed := false
	cp.SetCommands([]Command{
		{ID: "cmd1", Label: "Command 1", Action: func() { executed = true }},
	})
	cp.Show(0, 0)
	cp.Select()
	if !executed {
		t.Error("command Action should have been called")
	}
	if cp.Visible() {
		t.Error("should be hidden after Select")
	}
}

func TestCommandPalette_Select_OnExecute(t *testing.T) {
	cp := NewCommandPalette()
	var executedCmd Command
	cp.SetCommands([]Command{
		{ID: "cmd1", Label: "Command 1"},
	})
	cp.SetOnExecute(func(cmd Command) {
		executedCmd = cmd
	})
	cp.Show(0, 0)
	cp.Select()
	if executedCmd.ID != "cmd1" {
		t.Errorf("OnExecute cmd = %q, want 'cmd1'", executedCmd.ID)
	}
}

func TestCommandPalette_Select_Empty(t *testing.T) {
	cp := NewCommandPalette()
	cp.Select() // should not panic
}

// ─── Configuration ───────────────────────────────────────────────

func TestCommandPalette_SetMaxVisible(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetMaxVisible(5)
	if cp.MaxVisible() != 5 {
		t.Errorf("MaxVisible = %d, want 5", cp.MaxVisible())
	}
}

func TestCommandPalette_SetMaxVisible_ClampToMin1(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetMaxVisible(0)
	if cp.MaxVisible() != 1 {
		t.Errorf("MaxVisible = %d, want 1 (clamped)", cp.MaxVisible())
	}
}

func TestCommandPalette_SetStyle(t *testing.T) {
	cp := NewCommandPalette()
	custom := CommandPaletteStyle{
		Normal:  buffer.Style{Fg: buffer.Color256Val(255)},
		Matched: buffer.Style{Fg: buffer.Color256Val(39), Flags: buffer.Bold},
	}
	cp.SetStyle(custom)
	if cp.Style().Normal.Fg != buffer.Color256Val(255) {
		t.Error("style not set")
	}
}

func TestCommandPalette_DefaultStyle(t *testing.T) {
	s := DefaultCommandPaletteStyle()
	_ = s // should not panic
}

// ─── Scroll ──────────────────────────────────────────────────────

func TestCommandPalette_ScrollY(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(30))
	cp.SetMaxVisible(5)
	if cp.ScrollY() != 0 {
		t.Errorf("initial ScrollY = %d, want 0", cp.ScrollY())
	}
	cp.SetCursor(10)
	if cp.ScrollY() != 6 {
		t.Errorf("ScrollY = %d, want 6", cp.ScrollY())
	}
}

// ─── Keyboard ────────────────────────────────────────────────────

func TestCommandPalette_HandleKey_Navigation(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))

	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyDown}) {
		t.Error("HandleKey Down should return true")
	}
	if cp.Cursor() != 1 {
		t.Errorf("cursor = %d after Down, want 1", cp.Cursor())
	}
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyUp}) {
		t.Error("HandleKey Up should return true")
	}
	if cp.Cursor() != 0 {
		t.Errorf("cursor = %d after Up, want 0", cp.Cursor())
	}
}

func TestCommandPalette_HandleKey_Enter(t *testing.T) {
	cp := NewCommandPalette()
	executed := false
	cp.SetCommands([]Command{
		{ID: "cmd", Label: "Test", Action: func() { executed = true }},
	})
	cp.Show(0, 0)
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("HandleKey Enter should return true")
	}
	if !executed {
		t.Error("command should have been executed")
	}
}

func TestCommandPalette_HandleKey_Escape(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	cp.Show(0, 0)
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("HandleKey Escape should return true")
	}
	if cp.Visible() {
		t.Error("should be hidden after Escape")
	}
}

func TestCommandPalette_HandleKey_Tab(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyTab}) {
		t.Error("HandleKey Tab should return true")
	}
	if cp.Cursor() != 1 {
		t.Errorf("cursor = %d after Tab, want 1", cp.Cursor())
	}
}

func TestCommandPalette_HandleKey_Backspace(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.SetQuery("test")
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace}) {
		t.Error("HandleKey Backspace should return true")
	}
	if cp.Query() != "tes" {
		t.Errorf("Query = %q, want 'tes'", cp.Query())
	}
}

func TestCommandPalette_HandleKey_Printable(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	if !cp.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'a'}) {
		t.Error("HandleKey printable should return true")
	}
	if cp.Query() != "a" {
		t.Errorf("Query = %q, want 'a'", cp.Query())
	}
}

func TestCommandPalette_HandleKey_Unhandled(t *testing.T) {
	cp := NewCommandPalette()
	if cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) {
		t.Error("HandleKey Left should return false")
	}
}

func TestCommandPalette_HandleKey_Nil(t *testing.T) {
	cp := NewCommandPalette()
	if cp.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

// ─── Measure ─────────────────────────────────────────────────────

func TestCommandPalette_Measure_Basic(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	size := cp.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 20 {
		t.Errorf("width = %d, want >= 20", size.W)
	}
	if size.H < 4 {
		t.Errorf("height = %d, want >= 4", size.H)
	}
}

func TestCommandPalette_Measure_ClampedToConstraints(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(20))
	size := cp.Measure(Constraints{MaxWidth: 30, MaxHeight: 8})
	if size.W > 30 {
		t.Errorf("width = %d, want <= 30", size.W)
	}
	if size.H > 8 {
		t.Errorf("height = %d, want <= 8", size.H)
	}
}

// ─── Paint ───────────────────────────────────────────────────────

func TestCommandPalette_Paint_NoPanic(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	cp.Show(0, 0)
	buf := buffer.NewBuffer(40, 15)
	cp.Paint(buf) // should not panic
}

func TestCommandPalette_Paint_Invisible(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	cp.Paint(buf) // should not draw when hidden
}

func TestCommandPalette_Paint_ZeroBounds(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(5))
	cp.Show(0, 0)
	buf := buffer.NewBuffer(40, 15)
	cp.Paint(buf) // should not panic with zero bounds
}

func TestCommandPalette_Paint_RendersBorder(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{{ID: "test", Label: "Test"}})
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 12})
	cp.Show(0, 0)
	buf := buffer.NewBuffer(40, 12)
	cp.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Errorf("top-left corner = %q, want '┌'", string(buf.GetCell(0, 0).Rune))
	}
}

func TestCommandPalette_Paint_NoResults(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{{ID: "test", Label: "Test"}})
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 12})
	cp.Show(0, 0)
	cp.SetQuery("xyz") // no matches
	buf := buffer.NewBuffer(40, 12)
	cp.Paint(buf) // should not panic
}

// ─── Misc ────────────────────────────────────────────────────────

func TestCommandPalette_FilteredCommands(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{ID: "save", Label: "Save File"},
		{ID: "search", Label: "Search"},
		{ID: "close", Label: "Close"},
	})
	cp.SetQuery("s")
	cmds := cp.FilteredCommands()
	if len(cmds) < 2 {
		t.Errorf("expected >= 2 filtered commands, got %d", len(cmds))
	}
}

func TestCommandPalette_Children(t *testing.T) {
	cp := NewCommandPalette()
	if cp.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestCommandPalette_String(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(3))
	if cp.String() == "" {
		t.Error("String should not be empty")
	}
}

// ─── Concurrency ─────────────────────────────────────────────────

func TestCommandPalette_ConcurrentAccess(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(100))

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				cp.SetQuery(fmt.Sprintf("Command %d", (n*50+j)%100))
				cp.MoveDown()
				cp.Cursor()
				cp.HasResults()
				cp.CurrentCommand()
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cp.FilteredCommands()
				cp.FilteredCount()
				cp.Commands()
			}
		}()
	}
	wg.Wait()
}

func TestCommandPalette_ConcurrentPaint(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands(newTestCommands(50))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	cp.Show(0, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(40, 15)
			cp.Paint(buf)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cp.SetQuery(fmt.Sprintf("Command %d", i%50))
			cp.MoveDown()
		}
	}()
	wg.Wait()
}