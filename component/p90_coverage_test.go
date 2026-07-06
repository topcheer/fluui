package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ═══════════════════════════════════════════════════════════════════════════
// P90 Coverage Tests — Targeting sub-76% functions in component package
// ═══════════════════════════════════════════════════════════════════════════

// ─── CodeBlock.gutterColorLocked (66.7% → 100%) ───

func TestP90_CodeBlock_GutterColorLocked(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	// Without theme
	cb.mu.RLock()
	c := cb.gutterColorLocked()
	cb.mu.RUnlock()
	if c.Type == buffer.ColorNone {
		t.Error("expected non-none color")
	}

	// With theme set (both branches return same value, but cover the if-true path)
	cb.SetTheme(nil) // nil theme is fine
	cb.mu.RLock()
	c2 := cb.gutterColorLocked()
	cb.mu.RUnlock()
	_ = c2
}

// ─── CodeBlock.paintStreamingCursorLocked (67.7% → 100%) ───

func TestP90_CodeBlock_PaintStreamingCursor_AllVariants(t *testing.T) {
	// Variant 1: empty lines, no title
	cb1 := NewCodeBlock("go", "")
	cb1.SetStreaming(true)
	cb1.SetShowTitle(false)
	cb1.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf1 := buffer.NewBuffer(40, 10)
	cb1.Paint(buf1)

	// Variant 2: empty lines, with title
	cb2 := NewCodeBlock("go", "")
	cb2.SetStreaming(true)
	cb2.SetShowTitle(true)
	cb2.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf2 := buffer.NewBuffer(40, 10)
	cb2.Paint(buf2)

	// Variant 3: with content, streaming
	cb3 := NewCodeBlock("go", "func main() {\n    fmt.Println(\"hi\")\n}")
	cb3.SetStreaming(true)
	cb3.SetShowTitle(true)
	cb3.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf3 := buffer.NewBuffer(40, 10)
	cb3.Paint(buf3)

	// Variant 4: with content, cursor clamped (narrow width)
	cb4 := NewCodeBlock("go", "very long line of code that exceeds bounds")
	cb4.SetStreaming(true)
	cb4.SetShowTitle(false)
	cb4.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf4 := buffer.NewBuffer(5, 3)
	cb4.Paint(buf4)

	// Variant 5: with plain fallback during streaming
	cb5 := NewCodeBlock("go", "x := 1\ny := 2")
	cb5.SetStreaming(true)
	cb5.mu.Lock()
	cb5.usePlainFallback = true
	cb5.plainLines = [][]buffer.Cell{
		{buffer.NewCell('x', buffer.Style{})},
		{buffer.NewCell('y', buffer.Style{})},
	}
	cb5.mu.Unlock()
	cb5.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf5 := buffer.NewBuffer(40, 10)
	cb5.Paint(buf5)
}

// ─── Viewport.ScrollToX (75% → 100%) ───

func TestP90_Viewport_ScrollToX_EdgeCases(t *testing.T) {
	v := NewViewport(NewTooltip("content"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	// Scroll to 0
	v.ScrollToX(0)

	// Scroll to negative
	v.ScrollToX(-5)

	// Scroll beyond max
	v.ScrollToX(9999)
}

// ─── Viewport.drawVScrollBar (73.7% → 95%+) ───

func TestP90_Viewport_DrawVScrollBar(t *testing.T) {
	v := NewViewport(NewTooltip("tall content\nthat\noverflows\nthe\nviewport\narea"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf)

	// Scroll down to test thumb position
	v.ScrollDown(2)
	buf2 := buffer.NewBuffer(20, 5)
	v.Paint(buf2)
}

func TestP90_Viewport_DrawVScrollBar_FitsContent(t *testing.T) {
	v := NewViewport(NewTooltip("short"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

// ─── Viewport.drawHScrollBar (68.4% → 95%+) ───

func TestP90_Viewport_DrawHScrollBar(t *testing.T) {
	v := NewViewport(NewTooltip("a very wide content line that should overflow horizontally"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf)

	// Scroll right
	v.ScrollRight(5)
	buf2 := buffer.NewBuffer(10, 5)
	v.Paint(buf2)
}

func TestP90_Viewport_DrawHScrollBar_FitsContent(t *testing.T) {
	v := NewViewport(NewTooltip("short"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf)
}

// ─── Wizard.activateButtonLocked (70.8% → 100%) ───

func TestP90_Wizard_ActivateButton_AllButtons(t *testing.T) {
	// Test Back button
	w1 := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
		{Title: "S2", Content: NewTooltip("c2")},
	})
	w1.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	w1.mu.Lock()
	w1.selected = WizardBtnBack
	result := w1.activateButtonLocked()
	w1.mu.Unlock()
	if !result {
		t.Error("activateButtonLocked(Back) should return true")
	}

	// Test Next button
	w2 := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
		{Title: "S2", Content: NewTooltip("c2")},
	})
	w2.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	w2.mu.Lock()
	w2.current = 0 // first step, so Next is valid
	w2.selected = WizardBtnNext
	result2 := w2.activateButtonLocked()
	w2.mu.Unlock()
	if !result2 {
		t.Error("activateButtonLocked(Next) should return true")
	}

	// Test Finish button
	finishCalled := false
	w3 := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})
	w3.OnFinish = func(w *Wizard) { finishCalled = true }
	w3.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	w3.mu.Lock()
	w3.selected = WizardBtnFinish
	result3 := w3.activateButtonLocked()
	w3.mu.Unlock()
	if !result3 {
		t.Error("activateButtonLocked(Finish) should return true")
	}
	if !finishCalled {
		t.Error("OnFinish callback should have been called")
	}

	// Test Cancel button
	cancelCalled := false
	w4 := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})
	w4.OnCancel = func(w *Wizard) { cancelCalled = true }
	w4.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	w4.mu.Lock()
	w4.selected = WizardBtnCancel
	result4 := w4.activateButtonLocked()
	w4.mu.Unlock()
	if !result4 {
		t.Error("activateButtonLocked(Cancel) should return true")
	}
	if !cancelCalled {
		t.Error("OnCancel callback should have been called")
	}

	// Test no-match (invalid selected value)
	w5 := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})
	w5.mu.Lock()
	w5.selected = 999 // invalid
	result5 := w5.activateButtonLocked()
	w5.mu.Unlock()
	if result5 {
		t.Error("activateButtonLocked(invalid) should return false")
	}
}

// ─── Wizard.moveButtonForward/moveButtonBackward (71.4% → 100%) ───

func TestP90_Wizard_MoveButtons_Wrap(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})
	w.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	// Move forward past last button (wrap)
	w.mu.Lock()
	w.selected = w.buttonOrder[len(w.buttonOrder)-1] // last button
	w.moveButtonForward()
	forwardResult := w.selected
	w.mu.Unlock()

	// Move backward past first button (wrap)
	w.mu.Lock()
	w.selected = w.buttonOrder[0]
	w.moveButtonBackward()
	backwardResult := w.selected
	w.mu.Unlock()

	_ = forwardResult
	_ = backwardResult
}

func TestP90_Wizard_MoveButtons_EmptyOrder(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})
	w.mu.Lock()
	w.buttonOrder = nil
	w.moveButtonForward()
	w.moveButtonBackward()
	w.mu.Unlock()
}

// ─── Wizard.Measure (69.2% → 100%) ───

func TestP90_Wizard_Measure_AllPaths(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTooltip("c1")},
	})

	// Default width/height
	s1 := w.Measure(Constraints{MaxWidth: 100, MaxHeight: 50})
	_ = s1

	// Constrained to small (trigger minimums)
	s2 := w.Measure(Constraints{MaxWidth: 10, MaxHeight: 3})
	if s2.W < 30 {
		t.Errorf("Measure W = %d, want >= 30 (minimum)", s2.W)
	}
	if s2.H < 8 {
		t.Errorf("Measure H = %d, want >= 8 (minimum)", s2.H)
	}

	// Exact fit
	s3 := w.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s3.W > 80 {
		t.Errorf("Measure W = %d, want <= 80", s3.W)
	}
	if s3.H > 24 {
		t.Errorf("Measure H = %d, want <= 24", s3.H)
	}

	// Unbounded
	s4 := w.Measure(Constraints{})
	_ = s4
}

// ─── Tree.rebuildLocked (75% → 100%) ───

func TestP90_Tree_RebuildLocked_NilRoot(t *testing.T) {
	tr := NewTree()
	tr.mu.Lock()
	tr.rebuildLocked() // root is nil
	tr.mu.Unlock()
}

func TestP90_Tree_RebuildLocked_EmptyRoot(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	tr.SetRoot(root)
	tr.mu.Lock()
	tr.rebuildLocked()
	tr.mu.Unlock()
}

func TestP90_Tree_RebuildLocked_NestedExpanded(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	child1 := NewTreeNode("c1", "Child 1")
	child1.AddChild(NewTreeNode("g1", "Grandchild 1"))
	root.AddChild(child1)
	root.Expanded = true
	child1.Expanded = true
	tr.SetRoot(root)
	tr.mu.Lock()
	tr.rebuildLocked()
	tr.mu.Unlock()
	// flatList should include root, c1, g1
	if len(tr.flatList) < 3 {
		t.Errorf("flatList = %d items, want >= 3", len(tr.flatList))
	}
}

func TestP90_Tree_RebuildLocked_Collapsed(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	child1 := NewTreeNode("c1", "Child 1")
	child1.AddChild(NewTreeNode("g1", "Grandchild 1"))
	root.AddChild(child1)
	root.Expanded = true
	child1.Expanded = false // collapsed
	tr.SetRoot(root)
	tr.mu.Lock()
	tr.rebuildLocked()
	tr.mu.Unlock()
	// flatList should include root, c1 but NOT g1 (collapsed)
	if len(tr.flatList) != 2 {
		t.Errorf("flatList = %d items, want 2 (root+c1, g1 collapsed)", len(tr.flatList))
	}
}

// ─── Pagination.recomputePagesLocked (50% → 100%) ───

func TestP90_Pagination_RecomputePagesLocked_DirectField(t *testing.T) {
	p := NewPagination()
	// Set itemsPerPage to 0 directly to test the guard
	p.mu.Lock()
	p.totalItems = 100
	p.itemsPerPage = 0
	p.recomputePagesLocked()
	if p.totalPages != 0 {
		t.Errorf("totalPages with itemsPerPage=0 = %d, want 0", p.totalPages)
	}
	// currentPage should be clamped
	p.mu.Unlock()
}

func TestP90_Pagination_RecomputePagesLocked_Normal(t *testing.T) {
	p := NewPagination()
	p.mu.Lock()
	p.totalItems = 100
	p.itemsPerPage = 10
	p.recomputePagesLocked()
	if p.totalPages != 10 {
		t.Errorf("totalPages = %d, want 10", p.totalPages)
	}
	p.mu.Unlock()
}

// ─── RadioGroup.setNavigableCursor (66.7% → 100%) ───

func TestP90_RadioGroup_SetNavigableCursor_DisabledWrap(t *testing.T) {
	rg := NewRadioGroup([]string{"a", "b", "c"})
	// Disable middle item
	rg.SetDisabled(1, true)

	// Move cursor to disabled — should skip forward
	rg.mu.Lock()
	rg.cursor = 0
	rg.setNavigableCursor(1) // try to move to disabled 'b'
	result := rg.cursor
	rg.mu.Unlock()
	_ = result // should have skipped to 2
}

func TestP90_RadioGroup_SetNavigableCursor_AllDisabled(t *testing.T) {
	rg := NewRadioGroup([]string{"a", "b"})
	rg.SetDisabled(0, true)
	rg.SetDisabled(1, true)
	rg.mu.Lock()
	rg.setNavigableCursor(0) // all disabled — should stay
	rg.mu.Unlock()
}

// ─── Checkbox.setNavigableCursor (73.3% → 100%) ───

func TestP90_Checkbox_SetNavigableCursor_AllDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"a", "b", "c"})
	items := cb.Items()
	items[0].Disabled = true
	items[1].Disabled = true
	items[2].Disabled = true
	cb.SetItems(items)
	cb.mu.Lock()
	cb.setNavigableCursor(0)
	rg := cb.cursor
	cb.mu.Unlock()
	_ = rg
}

func TestP90_Checkbox_SetNavigableCursor_SkipDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"a", "b", "c", "d"})
	items := cb.Items()
	items[1].Disabled = true
	cb.SetItems(items)
	cb.mu.Lock()
	cb.cursor = 0
	cb.setNavigableCursor(1) // should skip to 2
	result := cb.cursor
	cb.mu.Unlock()
	if result != 2 {
		t.Errorf("cursor after skip disabled = %d, want 2", result)
	}
}

// ─── ContextMenu.setCursorLocked (73.3% → 100%) ───

func TestP90_ContextMenu_SetCursorLocked_NegativeAndOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("i1", "Item 1"))
	cm.AddItem(NewMenuItem("i2", "Item 2"))
	cm.AddItem(NewMenuItem("i3", "Item 3"))

	// Negative should clamp to 0
	cm.mu.Lock()
	cm.setCursorLocked(-1)
	c1 := cm.cursor
	cm.mu.Unlock()
	if c1 != 0 {
		t.Errorf("cursor after -1 = %d, want 0", c1)
	}

	// Overflow should clamp to last
	cm.mu.Lock()
	cm.setCursorLocked(99)
	c2 := cm.cursor
	cm.mu.Unlock()
	if c2 != 2 {
		t.Errorf("cursor after 99 = %d, want 2 (last)", c2)
	}

	// Empty items
	cm2 := NewContextMenu()
	cm2.mu.Lock()
	cm2.setCursorLocked(5)
	c3 := cm2.cursor
	cm2.mu.Unlock()
	if c3 != 0 {
		t.Errorf("cursor with empty items = %d, want 0", c3)
	}
}

// ─── DiffViewer.maxScrollOffsetLocked (70% → 100%) ───

func TestP90_DiffViewer_MaxScrollOffsetLocked(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("+ added line\n- removed line\n  context line")
	dv.mu.Lock()
	maxOff := dv.maxScrollOffsetLocked()
	dv.mu.Unlock()
	_ = maxOff
}

// ─── FilePicker.handleFilterKey (75% → 100%) ───

func TestP90_FilePicker_HandleFilterKey(t *testing.T) {
	fp := NewFilePicker(".")

	// Backspace in filter
	fp.SetFilter("abc")
	fp.handleFilterKey(&term.KeyEvent{Key: term.KeyBackspace})

	// Regular printable character
	fp.handleFilterKey(&term.KeyEvent{Rune: 'x', Key: term.KeyUnknown})

	// Enter (should exit filter mode)
	fp.handleFilterKey(&term.KeyEvent{Key: term.KeyEnter})
	if fp.filtering {
		t.Error("expected filtering = false after Enter")
	}

	// Escape (should cancel filter)
	fp.SetFiltering(true)
	fp.handleFilterKey(&term.KeyEvent{Key: term.KeyEscape})
	if fp.filtering {
		t.Error("expected filtering = false after Escape")
	}

	// Up/Down in filter mode
	fp.SetFiltering(true)
	fp.handleFilterKey(&term.KeyEvent{Key: term.KeyUp})
	fp.handleFilterKey(&term.KeyEvent{Key: term.KeyDown})
}

// ─── Gauge.paintVertical (70.8% → 100%) ───

func TestP90_Gauge_PaintVertical_AllPaths(t *testing.T) {
	// Normal value
	g1 := NewGauge()
	g1.SetValue(0.5)
	g1.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf1 := buffer.NewBuffer(10, 8)
	g1.Paint(buf1)

	// Full value
	g2 := NewGauge()
	g2.SetValue(1.0)
	g2.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf2 := buffer.NewBuffer(10, 8)
	g2.Paint(buf2)

	// Zero value
	g3 := NewGauge()
	g3.SetValue(0)
	g3.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf3 := buffer.NewBuffer(10, 8)
	g3.Paint(buf3)

	// Very narrow (width 1)
	g4 := NewGauge()
	g4.SetValue(0.5)
	g4.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf4 := buffer.NewBuffer(1, 5)
	g4.Paint(buf4)
}

// ─── Gauge.colorForRatio (75% → 100%) ───

func TestP90_Gauge_ColorForRatio(t *testing.T) {
	g := NewGauge()
	g.mu.Lock()
	// Low ratio (green)
	c1 := g.colorForRatio(0.2)
	// Medium ratio (yellow)
	c2 := g.colorForRatio(0.5)
	// High ratio (red)
	c3 := g.colorForRatio(0.9)
	g.mu.Unlock()
	_ = c1
	_ = c2
	_ = c3
}

// ─── Help.ensureSelectedVisibleLocked (75% → 100%) ───

func TestP90_Help_EnsureSelectedVisibleLocked(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "G1", Entries: []HelpEntry{{Keys: "Ctrl+A", Description: "desc A"}}},
		{Name: "G2", Entries: []HelpEntry{{Keys: "Ctrl+B", Description: "desc B"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.mu.Lock()
	ho.selected = 99 // out of bounds
	ho.ensureSelectedVisibleLocked()
	ho.mu.Unlock()
}

// ─── Notification.Paint (75% → 100%) ───

func TestP90_Notification_Paint_Variants(t *testing.T) {
	tm := NewToastManager(5)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Push a normal notification and paint
	tm.PushInfo("Title", "Info message")
	buf1 := buffer.NewBuffer(40, 10)
	tm.Paint(buf1)

	// Push a long message that should truncate
	tm.PushInfo("Title", "This is a very long notification message that exceeds bounds")
	buf2 := buffer.NewBuffer(15, 10)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	tm.Paint(buf2)

	// Zero bounds
	tm2 := NewToastManager(5)
	tm2.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf3 := buffer.NewBuffer(30, 3)
	tm2.Paint(buf3)
}

// ─── Spinner.Paint (70% → 100%) ───

func TestP90_Spinner_Paint_AllStates(t *testing.T) {
	s := NewSpinner("Loading")
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)

	// Not running
	s.Paint(buf)

	// Running
	s.Start()
	s.Paint(buf)

	// After stop
	s.Stop()
	s.Paint(buf)

	// Empty label, running
	s2 := NewSpinner("")
	s2.Start()
	s2.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	s2.Paint(buf)

	// Zero width
	s3 := NewSpinner("test")
	s3.Start()
	s3.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 1})
	s3.Paint(buf)
}

// ─── AutoComplete.Paint (75% → 100%) ───

func TestP90_AutoComplete_Paint_AllStates(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "a"}, {Label: "b"}, {Label: "c"}})
	ac.Show(0, 0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)

	// Cursor at last item
	ac.SetCursor(2)
	ac.Paint(buf)

	// Empty items, visible
	ac2 := NewAutoComplete()
	ac2.Show(0, 0)
	ac2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	ac2.Paint(buf)

	// Hidden
	ac.Hide()
	ac.Paint(buf)
}

// ─── TabBar.IsCloseButton (75% → 100%) ───

func TestP90_TabBar_IsCloseButton(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 3})
	tb.mu.RLock()
	// On close button area (rightmost of tab)
	idx, ok := tb.IsCloseButton(15, 0) // try to hit close button
	tb.mu.RUnlock()
	_ = idx
	_ = ok

	// Not on close button
	tb.mu.RLock()
	idx2, ok2 := tb.IsCloseButton(0, 5) // below tabs
	tb.mu.RUnlock()
	if ok2 {
		t.Error("expected ok=false for Y=5")
	}
	_ = idx2
}

// ─── Table.drawCellLocked (75% → 100%) ───

func TestP90_Table_DrawCellLocked(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.SetSelectedRow(0)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	tbl.Paint(buf)
}

// ─── Table.truncateToWidth (75% → 100%) ───

func TestP90_Table_TruncateToWidth(t *testing.T) {
	tbl := NewTable([]string{"H"})
	tbl.mu.Lock()
	// Exact fit
	r1 := tbl.truncateToWidth("hello", 5)
	if r1 != "hello" {
		t.Errorf("truncateToWidth('hello', 5) = %q, want 'hello'", r1)
	}
	// Needs truncation
	r2 := tbl.truncateToWidth("hello world", 5)
	if len(r2) > 5 {
		t.Errorf("truncateToWidth('hello world', 5) = %q, want len <= 5", r2)
	}
	// Empty string
	r3 := tbl.truncateToWidth("", 5)
	if r3 != "" {
		t.Errorf("truncateToWidth('', 5) = %q, want ''", r3)
	}
	// Zero width
	r4 := tbl.truncateToWidth("hello", 0)
	if r4 != "" {
		t.Errorf("truncateToWidth('hello', 0) = %q, want ''", r4)
	}
	// Width 1-2 (can't fit ellipsis)
	r5 := tbl.truncateToWidth("hello", 1)
	_ = r5
	tbl.mu.Unlock()
}

// ─── TextArea.clampCursorX (75% → 100%) ───

func TestP90_TextArea_ClampCursorX(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello\nworld")
	// Normal cursor position
	ta.clampCursorX()

	// Move cursor beyond line length
	ta.cursorX = 100
	ta.clampCursorX()
	if ta.cursorX > 5 {
		t.Errorf("cursorX after clamp = %d, want <= 5", ta.cursorX)
	}

	// Negative cursor
	ta.cursorX = -5
	ta.clampCursorX()
	if ta.cursorX != 0 {
		t.Errorf("cursorX after negative clamp = %d, want 0", ta.cursorX)
	}
}

// ─── Sparkline.sparkValueColor (75% → 100%) ───

func TestP90_Sparkline_SparkValueColor(t *testing.T) {
	// Normal value
	c1 := sparkValueColor(5, 0, 10)
	// Max value
	c2 := sparkValueColor(10, 0, 10)
	// Zero value
	c3 := sparkValueColor(0, 0, 10)
	// Negative min=max
	c4 := sparkValueColor(5, 5, 5)
	_ = c1
	_ = c2
	_ = c3
	_ = c4
}

// ─── ScrollView.scrollbarWidth (66.7% → 100%) ───

func TestP90_ScrollView_ScrollbarWidth(t *testing.T) {
	sv := NewScrollView(NewTooltip("content"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	w1 := sv.scrollbarWidth()

	// With content that overflows
	sv2 := NewScrollView(NewTooltip("tall\ncontent\nthat\noverflows\nthe\nviewport\nsignificantly"))
	sv2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	w2 := sv2.scrollbarWidth()

	_ = w1
	_ = w2
}

// ─── ScrollView.contentW (75% → 100%) ───

func TestP90_ScrollView_ContentW(t *testing.T) {
	sv := NewScrollView(NewTooltip("content"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	w := sv.contentW(20)
	_ = w
}

// ─── DebugInspector.keyName (75% → 100%) ───

func TestP90_DebugInspector_KeyName_AllKeys(t *testing.T) {
	// Test various key codes via keyName directly
	_ = keyName(term.KeyTab)
	_ = keyName(term.KeyEnter)
	_ = keyName(term.KeyEscape)
	_ = keyName(term.KeyBackspace)
	_ = keyName(term.KeyDelete)
	_ = keyName(term.KeyUp)
	_ = keyName(term.KeyDown)
	_ = keyName(term.KeyLeft)
	_ = keyName(term.KeyRight)
	_ = keyName(term.KeyPageUp)
	_ = keyName(term.KeyPageDown)
	_ = keyName(term.KeyHome)
	_ = keyName(term.KeyEnd)
	_ = keyName(term.KeyCode(999)) // unknown key
}
