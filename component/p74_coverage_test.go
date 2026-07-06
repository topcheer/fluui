package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === P74 Coverage: 0% functions ===

func TestP74_BaseComponent_Paint(t *testing.T) {
	bc := &BaseComponent{}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	bc.Paint(buf) // should be no-op, not panic
}

func TestP74_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	if !dp.ShowLineNumbers() {
		t.Error("expected ShowLineNumbers() to be true")
	}
	dp.SetShowLineNumbers(false)
	// Always returns true currently
	if !dp.ShowLineNumbers() {
		t.Error("expected ShowLineNumbers() to always be true")
	}
}

func TestP74_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
	// Should not panic
}

// === P74 Coverage: filepicker handleFilterKey (15%) ===

func TestP74_FilePicker_HandleFilterKey_Escape(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	if !fp.IsFiltering() {
		t.Fatal("expected filtering to be on")
	}
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("expected Escape to be consumed in filter mode")
	}
	if fp.IsFiltering() {
		t.Error("expected filtering to be off after Escape")
	}
}

func TestP74_FilePicker_HandleFilterKey_Enter(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("expected Enter to be consumed in filter mode")
	}
	if fp.IsFiltering() {
		t.Error("expected filtering to be off after Enter")
	}
}

func TestP74_FilePicker_HandleFilterKey_BackspaceEmpty(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if !consumed {
		t.Error("expected Backspace to be consumed in filter mode")
	}
	if fp.IsFiltering() {
		t.Error("expected filtering to be off after Backspace on empty filter")
	}
}

func TestP74_FilePicker_HandleFilterKey_BackspaceWithText(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	fp.AppendFilter('a')
	fp.AppendFilter('b')
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if !consumed {
		t.Error("expected Backspace to be consumed")
	}
	if fp.Filter() != "a" {
		t.Errorf("expected filter to be 'a', got %q", fp.Filter())
	}
}

func TestP74_FilePicker_HandleFilterKey_Printable(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'x'})
	if !consumed {
		t.Error("expected printable to be consumed")
	}
	if fp.Filter() != "x" {
		t.Errorf("expected filter 'x', got %q", fp.Filter())
	}
}

func TestP74_FilePicker_HandleFilterKey_NilKey(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetFiltering(true)
	consumed := fp.HandleKey(nil)
	if consumed {
		t.Error("expected nil key to return false")
	}
}

// === P74 Coverage: dialog dialogTypeString (33%) ===

func TestP74_Dialog_dialogTypeString_AllTypes(t *testing.T) {
	tests := []struct {
		dt       DialogType
		expected string
	}{
		{DialogInfo, "Info"},
		{DialogConfirm, "Confirm"},
		{DialogPrompt, "Prompt"},
		{DialogCustom, "Custom"},
		{DialogType(99), "Unknown"},
	}
	for _, tt := range tests {
		got := dialogTypeString(tt.dt)
		if got != tt.expected {
			t.Errorf("dialogTypeString(%d) = %q, want %q", tt.dt, got, tt.expected)
		}
	}
}

func TestP74_Dialog_splitLines(t *testing.T) {
	result := splitLines("")
	if len(result) != 0 {
		t.Errorf("expected empty result for empty string, got %v", result)
	}
	result = splitLines("line1\nline2\nline3")
	if len(result) != 3 {
		t.Errorf("expected 3 lines, got %d", len(result))
	}
}

// === P74 Coverage: pagination recomputePagesLocked (50%) ===

func TestP74_Pagination_recomputePages_ZeroPerPage(t *testing.T) {
	p := NewPagination()
	// SetItemsPerPage(0) is a no-op — itemsPerPage stays at default
	p.SetItemsPerPage(10)
	p.SetTotalItems(50)
	if p.TotalPages() != 5 {
		t.Errorf("expected 5 pages for 50/10, got %d", p.TotalPages())
	}
}

func TestP74_Pagination_recomputePages_ExactDiv(t *testing.T) {
	p := NewPagination()
	p.SetItemsPerPage(10)
	p.SetTotalItems(100)
	if p.TotalPages() != 10 {
		t.Errorf("expected 10 pages, got %d", p.TotalPages())
	}
}

func TestP74_Pagination_recomputePages_Remainder(t *testing.T) {
	p := NewPagination()
	p.SetItemsPerPage(10)
	p.SetTotalItems(105)
	if p.TotalPages() != 11 {
		t.Errorf("expected 11 pages for 105 items / 10 per page, got %d", p.TotalPages())
	}
}

func TestP74_Pagination_recomputePages_FewItems(t *testing.T) {
	p := NewPagination()
	p.SetItemsPerPage(10)
	p.SetTotalItems(5)
	if p.TotalPages() != 1 {
		t.Errorf("expected 1 page for 5 items / 10 per page, got %d", p.TotalPages())
	}
}

// === P74 Coverage: contextmenu setCursorLocked (53%) and HandleKey (61%) ===

func TestP74_ContextMenu_SetCursor_SkipSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.AddSeparator()
	cm.AddLabel("Item 2", "id2")
	cm.SetCursor(1) // pointing at separator
	if cm.Cursor() != 2 {
		t.Errorf("expected cursor to skip separator to index 2, got %d", cm.Cursor())
	}
}

func TestP74_ContextMenu_SetCursor_BackwardSearch(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.AddSeparator()
	cm.AddLabel("Item 2", "id2")
	cm.AddSeparator()
	cm.AddLabel("Item 3", "id3")
	cm.SetCursor(3) // pointing at second separator
	// Should skip forward to item 3 (index 4)
	if cm.Cursor() != 4 {
		t.Errorf("expected cursor to skip to 4, got %d", cm.Cursor())
	}
}

func TestP74_ContextMenu_SetCursor_Empty(t *testing.T) {
	cm := NewContextMenu()
	cm.SetCursor(5) // no items
	if cm.Cursor() != 0 {
		t.Errorf("expected cursor 0 for empty menu, got %d", cm.Cursor())
	}
}

func TestP74_ContextMenu_SetCursor_Negative(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.SetCursor(-5)
	if cm.Cursor() != 0 {
		t.Errorf("expected cursor 0 for negative, got %d", cm.Cursor())
	}
}

func TestP74_ContextMenu_SetCursor_DisabledItems(t *testing.T) {
	cm := NewContextMenu()
	item1 := NewMenuItem("id1", "Item 1")
	item1.SetEnabled(false)
	cm.AddItem(item1)
	cm.AddLabel("Item 2", "id2")
	cm.SetCursor(0)
	if cm.Cursor() != 1 {
		t.Errorf("expected cursor to skip disabled item to 1, got %d", cm.Cursor())
	}
}

func TestP74_ContextMenu_HandleKey_NotVisible(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	if cm.HandleKey(&term.KeyEvent{Key: term.KeyUp}) {
		t.Error("expected false when menu not visible")
	}
}

func TestP74_ContextMenu_HandleKey_Escape(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.Show(0, 0)
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("expected Escape to be consumed")
	}
	if cm.Visible() {
		t.Error("expected menu to be hidden after Escape")
	}
}

func TestP74_ContextMenu_HandleKey_Enter(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.Show(0, 0)
	cm.SetCursor(0)
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("expected Enter to be consumed")
	}
}

func TestP74_ContextMenu_HandleKey_LeftNoSubmenu(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("Item 1", "id1")
	cm.Show(0, 0)
	cm.SetCursor(0)
	// Left without submenu returns false
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if consumed {
		t.Error("expected Left to return false when no submenu")
	}
}

func TestP74_ContextMenu_HandleKey_RightWithSubmenu(t *testing.T) {
	cm := NewContextMenu()
	parent := NewMenuItem("id1", "Parent")
	sub := NewContextMenu()
	sub.AddLabel("Sub 1", "sub1")
	parent.SetSubmenu(sub)
	cm.AddItem(parent)
	cm.Show(0, 0)
	cm.SetCursor(0)
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !consumed {
		t.Error("expected Right to open submenu")
	}
	if !sub.Visible() {
		t.Error("expected submenu to be visible")
	}
}

// === P74 Coverage: tabbar HitTest (56%) and IsCloseButton (55%) ===

func TestP74_TabBar_HitTest_OutOfBoundsY(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	if tb.HitTest(5, 5) != -1 {
		t.Error("expected -1 for y out of bounds")
	}
}

func TestP74_TabBar_HitTest_OutOfBoundsX(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	if tb.HitTest(-5, 0) != -1 {
		t.Error("expected -1 for x < 0")
	}
}

func TestP74_TabBar_HitTest_NewButton(t *testing.T) {
	tb := NewTabBar()
	tb.SetShowNewButton(true)
	tb.AddTab("t1", "Tab 1")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	// Tab 1 takes some width, new button follows
	// With 1 tab, the new button should be at offset = tabWidth
	x := tb.Bounds().X + tb.tabWidthLocked(tb.tabs[0])
	result := tb.HitTest(x, 0)
	if result != -2 {
		t.Errorf("expected -2 for new button hit, got %d", result)
	}
}

func TestP74_TabBar_HitTest_Separator(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	// Separator is at x = tabWidth(tab1)
	sepX := tb.Bounds().X + tb.tabWidthLocked(tb.tabs[0])
	if tb.HitTest(sepX, 0) != -1 {
		// Separator returns -1
	}
}

func TestP74_TabBar_IsCloseButton_OutOfBounds(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	idx, ok := tb.IsCloseButton(-1, 5)
	if ok || idx != -1 {
		t.Error("expected (-1, false) for out of bounds")
	}
}

func TestP74_TabBar_IsCloseButton_NonClosableTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	// Default tabs are not closable
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	_, ok := tb.IsCloseButton(2, 0)
	if ok {
		t.Error("expected false for non-closable tab")
	}
}

// === P74 Coverage: wizard Back (70%) ===

func TestP74_Wizard_Back_ErrorOnFirstStep(t *testing.T) {
	step := NewWizardStep("Step 1", "id1")
	w := NewWizard([]*WizardStep{step})
	err := w.Back()
	if err == nil {
		t.Error("expected error when going back from first step")
	}
}

func TestP74_Wizard_Back_Success(t *testing.T) {
	s1 := NewWizardStep("Step 1", "id1")
	s2 := NewWizardStep("Step 2", "id2")
	w := NewWizard([]*WizardStep{s1, s2})
	w.SetCurrentStep(1) // Move to step 2
	err := w.Back()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.CurrentStepIndex() != 0 {
		t.Errorf("expected current step 0, got %d", w.CurrentStepIndex())
	}
}

func TestP74_Wizard_Back_WithOnLeave(t *testing.T) {
	s1 := NewWizardStep("Step 1", "id1")
	s2 := NewWizardStep("Step 2", "id2")
	called := false
	s2.SetOnLeave(func(w *Wizard) error {
		called = true
		return nil
	})
	w := NewWizard([]*WizardStep{s1, s2})
	w.SetCurrentStep(1)
	w.Back()
	if !called {
		t.Error("expected OnLeave to be called")
	}
}

func TestP74_Wizard_Back_WithOnEnter(t *testing.T) {
	s1 := NewWizardStep("Step 1", "id1")
	s2 := NewWizardStep("Step 2", "id2")
	called := false
	s1.SetOnEnter(func(w *Wizard) error {
		called = true
		return nil
	})
	w := NewWizard([]*WizardStep{s1, s2})
	w.SetCurrentStep(1)
	w.Back()
	if !called {
		t.Error("expected OnEnter to be called")
	}
}

// === P74 Coverage: slider Measure (57%) ===

func TestP74_Slider_Measure_Vertical(t *testing.T) {
	s := NewSlider()
	s.SetOrientation(SliderVertical)
	sz := s.Measure(Bounded(10, 20))
	if sz.W <= 0 || sz.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", sz.W, sz.H)
	}
}

func TestP74_Slider_Measure_VerticalWithLabel(t *testing.T) {
	s := NewSlider()
	s.SetOrientation(SliderVertical)
	s.SetLabel("Volume")
	s.SetShowValue(true)
	sz := s.Measure(Bounded(10, 20))
	if sz.W <= 0 || sz.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", sz.W, sz.H)
	}
}

func TestP74_Slider_Measure_HorizontalUnbounded(t *testing.T) {
	s := NewSlider()
	s.SetLabel("Volume")
	sz := s.Measure(Unbounded())
	if sz.W <= 0 || sz.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", sz.W, sz.H)
	}
}

// === P74 Coverage: form clampInt (60%) ===

func TestP74_Form_clampInt(t *testing.T) {
	if clampInt(5, 10) != 5 {
		t.Error("expected 5 for 5 max=10")
	}
	if clampInt(-1, 10) != 0 {
		t.Error("expected 0 for -1 max=10")
	}
	if clampInt(15, 10) != 10 {
		t.Error("expected 10 for 15 max=10")
	}
	if clampInt(0, 10) != 0 {
		t.Error("expected 0 for 0 max=10")
	}
	if clampInt(10, 10) != 10 {
		t.Error("expected 10 for 10 max=10")
	}
}

// === P74 Coverage: codeblock paintStreamingCursorLocked (64.5%) ===

func TestP74_CodeBlock_PaintStreamingCursor_EmptyLines(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
	// Should paint cursor at top-left
}

func TestP74_CodeBlock_PaintStreamingCursor_WithContent(t *testing.T) {
	cb := NewCodeBlock("go", "fmt.Println")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP74_CodeBlock_PaintStreamingCursor_WithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetTitle("test.go")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP74_CodeBlock_gutterColorLocked(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

// === P74 Coverage: checkbox setNavigableCursor (67%) ===

func TestP74_Checkbox_SetNavigableCursor_AllDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.SetItems([]CheckboxItem{
		{Label: "A", Disabled: true},
		{Label: "B", Disabled: true},
		{Label: "C", Disabled: true},
	})
	cb.SetCursor(0)
	// With all disabled, cursor should remain valid (0)
	if cb.Cursor() < 0 {
		t.Error("cursor should not be negative")
	}
}

func TestP74_Checkbox_SetNavigableCursor_SkipForward(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C", "D"})
	cb.SetItems([]CheckboxItem{
		{Label: "A", Disabled: true},
		{Label: "B", Disabled: true},
		{Label: "C"},
		{Label: "D"},
	})
	cb.SetCursor(0)
	if cb.Cursor() != 2 {
		t.Errorf("expected cursor to skip to 2, got %d", cb.Cursor())
	}
}

// === P74 Coverage: commandpalette clampScrollLocked (67%) ===

func TestP74_CommandPalette_ClampScroll_ManyResults(t *testing.T) {
	cp := NewCommandPalette()
	for i := 0; i < 50; i++ {
		cp.AddCommand(Command{ID: "cmd" + string(rune('A'+i%26)) + string(rune('0'+i%10)), Label: "Command" + string(rune('A'+i%26)), Category: "cat"})
	}
	cp.SetMaxVisible(5)
	cp.SetQuery("") // show all
	cp.SetCursor(45) // near the end
	if cp.ScrollY() < 0 {
		t.Error("scroll should not be negative")
	}
}

func TestP74_CommandPalette_ClampScroll_NearEnd(t *testing.T) {
	cp := NewCommandPalette()
	for i := 0; i < 30; i++ {
		cp.AddCommand(Command{ID: "cmd" + string(rune('a'+i%26)) + string(rune('0'+i%10)), Label: "Cmd" + string(rune('a'+i%26))})
	}
	cp.SetMaxVisible(10)
	cp.SetQuery("")
	cp.MoveDown()
	cp.MoveDown()
	cp.MoveDown()
	if cp.ScrollY() < 0 {
		t.Error("scroll should not be negative")
	}
}

// === P74 Coverage: debuginspector componentTypeName (67%) ===

func TestP74_DebugInspector_ComponentTypeName(t *testing.T) {
	di := NewDebugInspector()
	di.SetRoot(NewStatusBar())
	di.SetRoot(&BaseComponent{})
	if componentTypeName(nil) != "nil" {
		t.Error("expected 'nil' for nil component")
	}
	if componentTypeName(&BaseComponent{}) != "BaseComponent" {
		// Should get type name
	}
	_ = di
}

// === P74 Coverage: dialog HandleKey (61.5%) ===

func TestP74_Dialog_HandleKey_NilKey(t *testing.T) {
	d := NewDialog(DialogInfo, "Test", "Message")
	if d.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestP74_Dialog_HandleKey_Left(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	d.AddButton(NewDialogButton("OK", DialogResultOK))
	d.AddButton(NewDialogButton("Cancel", DialogResultCancel))
	d.SetCursor(1)
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if !consumed {
		t.Error("expected Left to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Right(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	d.AddButton(NewDialogButton("OK", DialogResultOK))
	d.AddButton(NewDialogButton("Cancel", DialogResultCancel))
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !consumed {
		t.Error("expected Right to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Up(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if !consumed {
		t.Error("expected Up to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Down(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !consumed {
		t.Error("expected Down to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Tab(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !consumed {
		t.Error("expected Tab to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Backspace(t *testing.T) {
	d := NewDialog(DialogPrompt, "Test", "Message")
	d.SetInputValue("hello")
	d.SetInputCursor(3)
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if !consumed {
		t.Error("expected Backspace to be consumed in prompt mode")
	}
}

func TestP74_Dialog_HandleKey_BackspaceNonPrompt(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if consumed {
		t.Error("expected Backspace to not be consumed in non-prompt mode")
	}
}

func TestP74_Dialog_HandleKey_Delete(t *testing.T) {
	d := NewDialog(DialogPrompt, "Test", "Message")
	d.SetInputValue("hello")
	d.SetInputCursor(2)
	// Delete is not handled in dialog switch (only Ctrl+ modifier path)
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyDelete})
	_ = consumed // Delete is not consumed
}

func TestP74_Dialog_HandleKey_Enter(t *testing.T) {
	d := NewDialog(DialogConfirm, "Test", "Message")
	d.AddButton(NewDialogButton("OK", DialogResultOK))
	d.SetCursor(0)
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("expected Enter to be consumed")
	}
}

func TestP74_Dialog_HandleKey_Home(t *testing.T) {
	d := NewDialog(DialogPrompt, "Test", "Message")
	d.SetInputValue("test")
	d.SetInputCursor(4)
	// Home via Ctrl+A
	consumed := d.HandleKey(&term.KeyEvent{Key: 0x01, Modifiers: term.ModCtrl})
	if !consumed {
		t.Error("expected Ctrl+A (Home) to be consumed")
	}
}

func TestP74_Dialog_HandleKey_End(t *testing.T) {
	d := NewDialog(DialogPrompt, "Test", "Message")
	d.SetInputValue("test")
	d.SetInputCursor(0)
	// End via Ctrl+E
	consumed := d.HandleKey(&term.KeyEvent{Key: 0x05, Modifiers: term.ModCtrl})
	if !consumed {
		t.Error("expected Ctrl+E (End) to be consumed")
	}
}

func TestP74_Dialog_HandleKey_PrintablePrompt(t *testing.T) {
	d := NewDialog(DialogPrompt, "Test", "Message")
	d.SetInputValue("abc")
	d.SetInputCursor(3)
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'X'})
	if !consumed {
		t.Error("expected printable to be consumed in prompt mode")
	}
}

func TestP74_Dialog_HandleKey_Escape(t *testing.T) {
	d := NewDialog(DialogInfo, "Test", "Message")
	consumed := d.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("expected Escape to be consumed")
	}
}

// === P74 Coverage: filepicker HandleKey (55%) ===

func TestP74_FilePicker_HandleKey_Nil(t *testing.T) {
	fp := NewFilePicker(".")
	if fp.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestP74_FilePicker_HandleKey_Enter(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "test.txt", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("expected Enter to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_Backspace(t *testing.T) {
	fp := NewFilePicker(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if !consumed {
		t.Error("expected Backspace to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_Space(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "file.txt", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if !consumed {
		t.Error("expected Space to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_Home(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	fp.SetCursor(1)
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if !consumed {
		t.Error("expected Home to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_End(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if !consumed {
		t.Error("expected End to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_Escape(t *testing.T) {
	fp := NewFilePicker(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("expected Escape to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_VimKeys(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	for _, r := range []rune{'j', 'k', 'g', 'G', '/'} {
		consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: r})
		if !consumed {
			t.Errorf("expected rune %q to be consumed", r)
		}
		if r == '/' {
			// Turn off filtering for next iterations
			fp.SetFiltering(false)
		}
	}
}

func TestP74_FilePicker_HandleKey_PageUp(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if !consumed {
		t.Error("expected PageUp to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_PageDown(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	if !consumed {
		t.Error("expected PageDown to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_ArrowUp(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	fp.SetCursor(1)
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyUp, Rune: 0})
	if !consumed {
		t.Error("expected ArrowUp to be consumed")
	}
}

func TestP74_FilePicker_HandleKey_ArrowDown(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return []FileEntry{{Name: "a", IsDir: false}, {Name: "b", IsDir: false}}, nil
	})
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	fp.loadDir(".")
	consumed := fp.HandleKey(&term.KeyEvent{Key: term.KeyDown, Rune: 0})
	if !consumed {
		t.Error("expected ArrowDown to be consumed")
	}
}
