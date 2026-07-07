package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ─── ThemeStudio initSlots coverage ───

func TestP103_InitSlots_AllGettersSetters(t *testing.T) {
	tm := theme.Default()
	ts := NewThemeStudio(tm)

	if len(ts.slots) == 0 {
		t.Fatal("initSlots produced no slots")
	}

	// Test every getter/setter closure
	for i, slot := range ts.slots {
		// Getter should return a valid color
		orig := slot.getter(tm)
		_ = orig // just verify it doesn't panic

		// Setter should modify the theme field
		newColor := buffer.RGB(0xAB, 0xCD, 0xEF)
		slot.setter(tm, newColor)
		got := slot.getter(tm)
		if got.Type != newColor.Type || got.Val != newColor.Val {
			t.Errorf("slot[%d] %s: setter did not modify theme field (got %+v, want %+v)",
				i, slot.Name, got, newColor)
		}

		// Restore
		slot.setter(tm, orig)
	}
}

func TestP103_InitSlots_Sorted(t *testing.T) {
	ts := NewThemeStudio(theme.Default())

	// Verify slots are sorted by category then name
	for i := 1; i < len(ts.slots); i++ {
		if ts.slots[i-1].Category > ts.slots[i].Category {
			t.Errorf("slots not sorted by category at %d", i)
		}
		if ts.slots[i-1].Category == ts.slots[i].Category && ts.slots[i-1].Name > ts.slots[i].Name {
			t.Errorf("slots not sorted by name at %d", i)
		}
	}
}

func TestP103_CountCategories(t *testing.T) {
	ts := NewThemeStudio(theme.Default())
	cats := countCategories(ts.slots)
	if cats < 5 {
		t.Errorf("expected at least 5 categories, got %d", cats)
	}
}

func TestP103_CopyTheme(t *testing.T) {
	// nil
	if copyTheme(nil) != nil {
		t.Error("copyTheme(nil) should return nil")
	}
	// non-nil
	tm := theme.Default()
	cp := copyTheme(tm)
	if cp == nil || cp == tm {
		t.Error("copyTheme should return a different pointer with same value")
	}
}

// ─── Viewport drawVScrollBar / drawHScrollBar coverage ───

func TestP103_Viewport_DrawVScrollBar_ContentOverflow(t *testing.T) {
	child := &fixedSize{w: 20, h: 50}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	buf := buffer.NewBuffer(10, 5)
	vp.Paint(buf)

	// Should have track characters on the right column
	found := false
	for y := 0; y < 5; y++ {
		c := buf.GetCell(9, y)
		if c.Rune != 0 && c.Rune != ' ' {
			found = true
		}
	}
	if !found {
		t.Error("vertical scrollbar not drawn on right edge")
	}
}

func TestP103_Viewport_DrawVScrollBar_ThumbClamped(t *testing.T) {
	child := &fixedSize{w: 20, h: 50}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.offsetY = 40 // scroll near bottom to trigger thumb clamping

	buf := buffer.NewBuffer(10, 5)
	vp.Paint(buf)
	// Just verify it doesn't panic
}

func TestP103_Viewport_DrawHScrollBar_ContentOverflow(t *testing.T) {
	child := &fixedSize{w: 50, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})

	buf := buffer.NewBuffer(10, 8)
	vp.Paint(buf)

	// Should have track characters on the bottom row
	found := false
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 7)
		if c.Rune != 0 && c.Rune != ' ' {
			found = true
		}
	}
	if !found {
		t.Error("horizontal scrollbar not drawn on bottom edge")
	}
}

func TestP103_Viewport_DrawHScrollBar_ThumbClamped(t *testing.T) {
	child := &fixedSize{w: 50, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	vp.offsetX = 40 // scroll near right to trigger thumb clamping

	buf := buffer.NewBuffer(10, 8)
	vp.Paint(buf)
	// Just verify it doesn't panic
}

func TestP103_Viewport_DrawVScrollBar_ZeroBarHeight(t *testing.T) {
	child := &fixedSize{w: 20, h: 50}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	vp.showVBar = true
	vp.showHBar = false

	buf := buffer.NewBuffer(10, 1)
	vp.Paint(buf)
	// barH <= 0 should return early without panic
}

func TestP103_Viewport_DrawHScrollBar_ZeroBarWidth(t *testing.T) {
	child := &fixedSize{w: 50, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 8})
	vp.showHBar = true
	vp.showVBar = false

	buf := buffer.NewBuffer(1, 8)
	vp.Paint(buf)
	// barW <= 0 should return early without panic
}

// ─── CodeBlock paintStreamingCursorLocked coverage ───

func TestP103_CodeBlock_StreamingCursor_EmptyLinesWithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
	// Should place cursor at top-left of code area (after title)
	// Just verify it doesn't panic
}

func TestP103_CodeBlock_StreamingCursor_WithLines(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP103_CodeBlock_StreamingCursor_ScrolledPastVisible(t *testing.T) {
	cb := NewCodeBlock("go", "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.ScrollTo(8) // scroll past many lines

	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP103_CodeBlock_StreamingCursor_NarrowBounds(t *testing.T) {
	cb := NewCodeBlock("go", "hello world this is a long line of code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})

	buf := buffer.NewBuffer(5, 3)
	cb.Paint(buf)
}

func TestP103_CodeBlock_StreamingCursor_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "test")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

// ─── Checkbox setNavigableCursor edge cases ───

func TestP103_Checkbox_SetNavigableCursor_AllDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	items := cb.Items()
	for i := range items {
		items[i].Disabled = true
	}
	cb.SetItems(items)

	// Should not panic even if all items are disabled
	cb.SetCursor(1)
}

func TestP103_Checkbox_SetNavigableCursor_SingleEnabled(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	items := cb.Items()
	items[0].Disabled = true
	items[2].Disabled = true
	cb.SetItems(items)

	cb.SetCursor(2) // disabled, should skip to B
	if cb.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", cb.Cursor())
	}
}

// ─── RadioGroup setNavigableCursor edge cases ───

func TestP103_RadioGroup_SetNavigableCursor_AllDisabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(0, true)
	rg.SetDisabled(1, true)
	rg.SetDisabled(2, true)

	rg.SetCursor(1)
	// Should not panic
}

func TestP103_RadioGroup_SetNavigableCursor_DisabledSkip(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(0, true)
	rg.SetDisabled(2, true)

	rg.SetCursor(0) // disabled, should skip to B (1)
	if rg.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", rg.Cursor())
	}
}

// ─── ContextMenu setCursorLocked edge cases ───

func TestP103_ContextMenu_SetCursor_LargeNegative(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))

	cm.SetCursor(-100)
	if cm.Cursor() != 0 { // clamps to first
		t.Errorf("expected cursor at 0, got %d", cm.Cursor())
	}
}

func TestP103_ContextMenu_SetCursor_LargeOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))

	cm.SetCursor(100)
	if cm.Cursor() != 1 { // clamps to last
		t.Errorf("expected cursor at 1, got %d", cm.Cursor())
	}
}

// ─── Viewport scroll edge cases ───

func TestP103_Viewport_ScrollLeft(t *testing.T) {
	child := &fixedSize{w: 30, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollToX(5)
	vp.ScrollLeft(2)
	if vp.OffsetX() != 3 {
		t.Errorf("expected OffsetX=3, got %d", vp.OffsetX())
	}
}

func TestP103_Viewport_HScrollbarRow(t *testing.T) {
	child := &fixedSize{w: 30, h: 30}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	if vp.HScrollbarRow() < 0 {
		t.Error("HScrollbarRow should be >= 0 when scrollbar visible")
	}
}

func TestP103_Viewport_VScrollbarColumn(t *testing.T) {
	child := &fixedSize{w: 30, h: 30}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	if vp.VScrollbarColumn() < 0 {
		t.Error("VScrollbarColumn should be >= 0 when scrollbar visible")
	}
}

// ─── Table truncateToWidth edge cases ───

func TestP103_Table_TruncateToWidth(t *testing.T) {
	tbl := NewTable([]string{"H"})
	tbl.SetRows([][]string{{"hello world"}})

	// Verify truncate works by measuring
	s := tbl.Measure(Bounded(5, 5))
	if s.W <= 0 || s.H <= 0 {
		t.Error("Table Measure returned zero size")
	}
}

// ─── Badge SizeName / VariantName ───

func TestP103_Badge_SizeName(t *testing.T) {
	if SizeName(BadgeSizeSmall) != "small" {
		t.Errorf("expected 'small', got %s", SizeName(BadgeSizeSmall))
	}
	if SizeName(BadgeSizeNormal) != "normal" {
		t.Errorf("expected 'normal', got %s", SizeName(BadgeSizeNormal))
	}
	if SizeName(BadgeSizeLarge) != "large" {
		t.Errorf("expected 'large', got %s", SizeName(BadgeSizeLarge))
	}
}

func TestP103_Badge_VariantName(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical}
	for _, v := range variants {
		name := VariantName(v)
		if name == "" {
			t.Errorf("VariantName(%d) returned empty", v)
		}
	}
}

// ─── Gauge colorForRatio ───

func TestP103_Gauge_ColorForRatio(t *testing.T) {
	g := NewGauge()
	g.SetRange(0, 100)
	g.SetValue(10)  // low
	g.SetValue(50)  // mid
	g.SetValue(90)  // high
	g.SetValue(100) // max

	// Just verify it doesn't panic
	g.Measure(Bounded(20, 3))
	buf := buffer.NewBuffer(20, 3)
	g.Paint(buf)
}

// ─── SelectField Value edge cases ───

func TestP103_SelectField_ValueEdgeCases(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	// Default selected should be 0
	if sf.Value() != "A" {
		t.Errorf("expected 'A', got %s", sf.Value())
	}
	// Select last
	sf.SetSelectedIndex(2)
	if sf.Value() != "C" {
		t.Errorf("expected 'C', got %s", sf.Value())
	}
}

// ─── Sparkline format helpers ───

func TestP103_Sparkline_SetData(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
	sp.SetAutoScale(true)
	sp.Measure(Bounded(20, 3))
	buf := buffer.NewBuffer(20, 3)
	sp.Paint(buf)
}

// ─── DebugInspector key recording ───

func TestP103_DebugInspector_RecordKey(t *testing.T) {
	di := NewDebugInspector()
	di.RecordKey(&term.KeyEvent{Rune: 'a'})
	di.RecordKey(&term.KeyEvent{Rune: 'b'})
	if len(di.Events()) != 2 {
		t.Errorf("expected 2 events, got %d", len(di.Events()))
	}
}

// ─── Tree Paint with indentation ───

func TestP103_Tree_PaintNested(t *testing.T) {
	root := NewTreeNode("root", "Root")
	child1 := NewTreeNode("c1", "Child 1")
	child1.Children = []*TreeNode{NewTreeNode("gc1", "Grandchild 1")}
	root.Children = []*TreeNode{child1, NewTreeNode("c2", "Child 2")}

	tr := NewTree()
	tr.SetRoot(root)
	tr.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	tr.Measure(Bounded(30, 10))
	buf := buffer.NewBuffer(30, 10)
	tr.Paint(buf)
}
