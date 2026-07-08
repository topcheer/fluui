package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func testMenus() []Menu {
	return []Menu{
		{
			ID: "file", Title: "File",
			Items: []MenuEntry{
				{ID: "new", Label: "New", Shortcut: "Ctrl+N"},
				{ID: "open", Label: "Open", Shortcut: "Ctrl+O"},
				{ID: "sep1", Label: "", Separator: true},
				{ID: "quit", Label: "Quit", Shortcut: "Ctrl+Q"},
			},
		},
		{
			ID: "edit", Title: "Edit",
			Items: []MenuEntry{
				{ID: "cut", Label: "Cut", Shortcut: "Ctrl+X"},
				{ID: "copy", Label: "Copy", Shortcut: "Ctrl+C"},
				{ID: "paste", Label: "Paste", Shortcut: "Ctrl+V"},
			},
		},
		{
			ID: "help", Title: "Help",
			Items: []MenuEntry{
				{ID: "about", Label: "About"},
				{ID: "disabled", Label: "Disabled", Disabled: true},
			},
		},
	}
}

func TestMenuBar_Construction(t *testing.T) {
	mb := NewMenuBar(testMenus())
	if len(mb.Menus()) != 3 {
		t.Errorf("expected 3 menus, got %d", len(mb.Menus()))
	}
	if mb.IsOpen() {
		t.Error("expected menu closed on construction")
	}
	if mb.ActiveMenu() != -1 {
		t.Error("expected activeIdx -1")
	}
}

func TestMenuBar_SetStyle(t *testing.T) {
	mb := NewMenuBar(testMenus())
	s := DefaultMenuBarStyle()
	s.Bar.Fg = buffer.RGB(255, 0, 0)
	mb.SetStyle(s)
	if mb.Style().Bar.Fg != buffer.RGB(255, 0, 0) {
		t.Error("style not set")
	}
}

func TestMenuBar_SetMenus(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetMenus([]Menu{{ID: "x", Title: "X", Items: []MenuEntry{{ID: "a", Label: "A"}}}})
	if len(mb.Menus()) != 1 {
		t.Errorf("expected 1 menu, got %d", len(mb.Menus()))
	}
	if mb.IsOpen() {
		t.Error("expected closed after SetMenus")
	}
}

func TestMenuBar_Measure(t *testing.T) {
	mb := NewMenuBar(testMenus())
	s := mb.Measure(Bounded(200, 10))
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
	if s.W <= 0 {
		t.Error("expected positive width")
	}
}

func TestMenuBar_Measure_Clamped(t *testing.T) {
	mb := NewMenuBar(testMenus())
	s := mb.Measure(Bounded(5, 1))
	if s.W > 5 {
		t.Errorf("expected width <= 5, got %d", s.W)
	}
}

func TestMenuBar_Paint_NoPanic(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 10)
	mb.Paint(buf)
}

func TestMenuBar_Paint_ZeroBounds(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 10)
	mb.Paint(buf) // should not panic
}

func TestMenuBar_Paint_WithOpenMenu(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(80, 10)
	mb.Paint(buf) // should paint dropdown
}

func TestMenuBar_OpenCloseMenu(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.OpenMenu(1)
	if !mb.IsOpen() {
		t.Error("expected open")
	}
	if mb.SelectedItem() < 0 {
		t.Error("expected selected item after open")
	}
	mb.CloseMenu()
	if mb.IsOpen() {
		t.Error("expected closed")
	}
}

func TestMenuBar_OpenMenu_InvalidIndex(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.OpenMenu(99) // should not panic
	if mb.IsOpen() {
		t.Error("expected not open for invalid index")
	}
}

func TestMenuBar_OpenMenu_SkipsSeparator(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.OpenMenu(0) // File menu, first item is "new" (index 0)
	// Separator at index 2 should be skipped when navigating
	idx := mb.SelectedItem()
	if idx == 2 {
		t.Error("separator should not be selected")
	}
}

func TestMenuBar_HandleKey_AltF(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	consumed := mb.HandleKey(&term.KeyEvent{
		Rune:      'f',
		Modifiers: term.ModAlt,
	})
	if !consumed {
		t.Error("expected Alt+F consumed")
	}
	if !mb.IsOpen() {
		t.Error("expected File menu open")
	}
}

func TestMenuBar_HandleKey_AltE(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	consumed := mb.HandleKey(&term.KeyEvent{
		Rune:      'e',
		Modifiers: term.ModAlt,
	})
	if !consumed || !mb.IsOpen() {
		t.Error("expected Edit menu open via Alt+E")
	}
}

func TestMenuBar_HandleKey_DownOpensMenu(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	// Right to activate first menu
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	// Down to open dropdown
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !mb.IsOpen() {
		t.Error("expected menu open after Down")
	}
}

func TestMenuBar_HandleKey_NavigateDown(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0) // File menu

	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // new → open
	if mb.SelectedItem() != 1 {
		t.Errorf("expected item 1 after Down, got %d", mb.SelectedItem())
	}
}

func TestMenuBar_HandleKey_NavigateDownSkipsSeparator(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // 0→1
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // 1→3 (skip 2=separator)
	if mb.SelectedItem() != 3 {
		t.Errorf("expected item 3 (skip separator 2), got %d", mb.SelectedItem())
	}
}

func TestMenuBar_HandleKey_NavigateDownSkipsDisabled(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(2) // Help menu

	// Item 0 = about, Item 1 = disabled
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // should wrap 0 → 0 (skip disabled 1)
	if mb.SelectedItem() == 1 {
		t.Error("disabled item should not be selected")
	}
}

func TestMenuBar_HandleKey_NavigateUp(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // 0→1
	mb.HandleKey(&term.KeyEvent{Key: term.KeyUp})   // 1→0
	if mb.SelectedItem() != 0 {
		t.Errorf("expected item 0, got %d", mb.SelectedItem())
	}
}

func TestMenuBar_HandleKey_LeftRightSwitchMenus(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0) // File

	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // → Edit
	if !mb.IsOpen() {
		t.Error("expected still open")
	}

	mb.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) // → File
	mb.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) // → Help (wrap)
}

func TestMenuBar_HandleKey_EscapeCloses(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	mb.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if mb.IsOpen() {
		t.Error("expected closed after Escape")
	}
}

func TestMenuBar_HandleKey_EnterFiresAction(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	var firedMenu, firedItem string
	var firedMu sync.Mutex
	mb.OnAction = func(menuID, itemID string) {
		firedMu.Lock()
		firedMenu = menuID
		firedItem = itemID
		firedMu.Unlock()
	}

	mb.OpenMenu(0) // File
	mb.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) // select first item

	firedMu.Lock()
	defer firedMu.Unlock()
	if firedMenu != "file" || firedItem != "new" {
		t.Errorf("expected file/new, got %s/%s", firedMenu, firedItem)
	}
}

func TestMenuBar_HandleKey_EnterOnDisabled(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	fired := false
	mb.OnAction = func(menuID, itemID string) {
		fired = true
	}

	mb.OpenMenu(2) // Help menu — first selectable is 0 (about)
	mb.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	// 'about' is valid, so action should fire
	if !fired {
		t.Error("should fire action for 'about' (valid, non-disabled)")
	}
}

func TestMenuBar_HandleKey_LeftRightMenuBar(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	// Left/Right without menu open
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // activate first
	if mb.ActiveMenu() != 0 {
		t.Errorf("expected active 0, got %d", mb.ActiveMenu())
	}
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if mb.ActiveMenu() != 1 {
		t.Errorf("expected active 1, got %d", mb.ActiveMenu())
	}
	mb.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if mb.ActiveMenu() != 0 {
		t.Errorf("expected active 0, got %d", mb.ActiveMenu())
	}
}

func TestMenuBar_HandleKey_LeftWrap(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	mb.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) // wrap to last
	if mb.ActiveMenu() != 2 {
		t.Errorf("expected wrap to 2, got %d", mb.ActiveMenu())
	}
}

func TestMenuBar_HandleKey_RightWrap(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // 0
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // 1
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // 2
	mb.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // wrap to 0
	if mb.ActiveMenu() != 0 {
		t.Errorf("expected wrap to 0, got %d", mb.ActiveMenu())
	}
}

func TestMenuBar_HandleKey_UnknownKey(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	consumed := mb.HandleKey(&term.KeyEvent{Key: term.KeyCode(999)})
	if consumed {
		t.Error("expected unknown key not consumed")
	}
}

func TestMenuBar_HandleKey_EmptyMenus(t *testing.T) {
	mb := NewMenuBar(nil)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	consumed := mb.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if consumed {
		t.Error("expected not consumed with empty menus")
	}
}

func TestMenuBar_HandleMouse_ClickMenuTitle(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	// Click on " File " (x=0..5)
	consumed := mb.HandleMouse(2, 0, term.MouseDown)
	if !consumed {
		t.Error("expected click consumed")
	}
	if !mb.IsOpen() {
		t.Error("expected menu open")
	}
}

func TestMenuBar_HandleMouse_ClickToggleClose(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	mb.OpenMenu(0)
	mb.HandleMouse(2, 0, term.MouseDown) // click same menu → close
	if mb.IsOpen() {
		t.Error("expected closed after toggle click")
	}
}

func TestMenuBar_HandleMouse_ClickSecondMenu(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	// " File " is 6 chars + 1 gap = 7. " Edit " starts at x=7
	consumed := mb.HandleMouse(8, 0, term.MouseDown)
	if !consumed {
		t.Error("expected click consumed")
	}
	if !mb.IsOpen() {
		t.Error("expected Edit menu open")
	}
}

func TestMenuBar_HandleMouse_ClickOutsideCloses(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	consumed := mb.HandleMouse(50, 5, term.MouseDown)
	if !consumed {
		t.Error("expected click outside consumed")
	}
	if mb.IsOpen() {
		t.Error("expected closed after click outside")
	}
}

func TestMenuBar_HandleMouse_ClickDropdownItem(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	var firedMenu, firedItem string
	var firedMu sync.Mutex
	mb.OnAction = func(menuID, itemID string) {
		firedMu.Lock()
		firedMenu = menuID
		firedItem = itemID
		firedMu.Unlock()
	}

	mb.OpenMenu(0) // File
	// Click first item (y=1)
	mb.HandleMouse(3, 1, term.MouseDown)

	firedMu.Lock()
	defer firedMu.Unlock()
	if firedMenu != "file" || firedItem != "new" {
		t.Errorf("expected file/new, got %s/%s", firedMenu, firedItem)
	}
}

func TestMenuBar_HandleMouse_ClickOnSeparator(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	fired := false
	mb.OnAction = func(menuID, itemID string) { fired = true }

	// Separator is at index 2, so y=3
	mb.HandleMouse(3, 3, term.MouseDown)
	if fired {
		t.Error("should not fire on separator")
	}
}

func TestMenuBar_HandleMouse_ClickEmptyBarArea(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	// Click far right of bar
	mb.HandleMouse(75, 0, term.MouseDown)
	if mb.IsOpen() {
		t.Error("expected closed after clicking empty area")
	}
}

func TestMenuBar_ConcurrentAccess(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			mb.Paint(buffer.NewBuffer(80, 10))
		}()
		go func() {
			defer wg.Done()
			mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
		}()
		go func() {
			defer wg.Done()
			_ = mb.IsOpen()
			_ = mb.SelectedItem()
		}()
	}
	wg.Wait()
}
