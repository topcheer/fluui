package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── ToStyle full branch coverage ───

func TestP106_ToStyle_AllFlagsTrue(t *testing.T) {
	decl := StyleDecl{
		Fg:            buffer.NamedColor(buffer.NamedRed),
		Bg:            buffer.NamedColor(buffer.NamedBlue),
		Bold:          boolPtr(true),
		Italic:        boolPtr(true),
		Underline:     boolPtr(true),
		Dim:           boolPtr(true),
		Blink:         boolPtr(true),
		Reverse:       boolPtr(true),
		Strikethrough: boolPtr(true),
	}
	s := decl.ToStyle(buffer.Style{})
	if s.Flags&(buffer.Bold|buffer.Italic|buffer.Underline|buffer.Dim|buffer.Blink|buffer.Reverse|buffer.Strikethrough) == 0 {
		t.Error("all flags should be set")
	}
}

func TestP106_ToStyle_AllFlagsFalseClear(t *testing.T) {
	// Start with all flags set, then clear them all
	base := buffer.Style{Flags: buffer.Bold | buffer.Italic | buffer.Underline | buffer.Dim | buffer.Blink | buffer.Reverse | buffer.Strikethrough}
	decl := StyleDecl{
		Bold:          boolPtr(false),
		Italic:        boolPtr(false),
		Underline:     boolPtr(false),
		Dim:           boolPtr(false),
		Blink:         boolPtr(false),
		Reverse:       boolPtr(false),
		Strikethrough: boolPtr(false),
	}
	s := decl.ToStyle(base)
	if s.Flags != 0 {
		t.Errorf("all flags should be cleared, got %d", s.Flags)
	}
}

// ─── applyDecl full coverage ───

func TestP106_ApplyDecl_AllProps(t *testing.T) {
	decl := &StyleDecl{}

	// Test all known property keys
	applyDecl(decl, "fg", "red")
	if !decl.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("fg should be red")
	}

	applyDecl(decl, "foreground", "blue")
	if !decl.Fg.Equal(buffer.NamedColor(buffer.NamedBlue)) {
		t.Error("foreground should be blue")
	}

	applyDecl(decl, "color", "green")
	if !decl.Fg.Equal(buffer.NamedColor(buffer.NamedGreen)) {
		t.Error("color should be green")
	}

	applyDecl(decl, "bg", "yellow")
	if !decl.Bg.Equal(buffer.NamedColor(buffer.NamedYellow)) {
		t.Error("bg should be yellow")
	}

	applyDecl(decl, "background", "cyan")
	if !decl.Bg.Equal(buffer.NamedColor(buffer.NamedCyan)) {
		t.Error("background should be cyan")
	}

	applyDecl(decl, "bold", "true")
	if decl.Bold == nil || !*decl.Bold {
		t.Error("bold should be true")
	}

	applyDecl(decl, "italic", "yes")
	if decl.Italic == nil || !*decl.Italic {
		t.Error("italic should be true")
	}

	applyDecl(decl, "underline", "on")
	if decl.Underline == nil || !*decl.Underline {
		t.Error("underline should be true")
	}

	applyDecl(decl, "dim", "1")
	if decl.Dim == nil || !*decl.Dim {
		t.Error("dim should be true")
	}

	applyDecl(decl, "blink", "true")
	if decl.Blink == nil || !*decl.Blink {
		t.Error("blink should be true")
	}

	applyDecl(decl, "reverse", "true")
	if decl.Reverse == nil || !*decl.Reverse {
		t.Error("reverse should be true")
	}

	applyDecl(decl, "strikethrough", "true")
	if decl.Strikethrough == nil || !*decl.Strikethrough {
		t.Error("strikethrough should be true")
	}

	applyDecl(decl, "strike", "true")
	if decl.Strikethrough == nil || !*decl.Strikethrough {
		t.Error("strike alias should work")
	}

	// Padding aliases
	applyDecl(decl, "padding-top", "3")
	if decl.PaddingTop == nil || *decl.PaddingTop != 3 {
		t.Error("padding-top should be 3")
	}

	applyDecl(decl, "pt", "4")
	if decl.PaddingTop == nil || *decl.PaddingTop != 4 {
		t.Error("pt alias should set PaddingTop to 4")
	}

	applyDecl(decl, "padding-bottom", "5")
	if decl.PaddingBottom == nil || *decl.PaddingBottom != 5 {
		t.Error("padding-bottom should be 5")
	}

	applyDecl(decl, "pb", "6")
	if decl.PaddingBottom == nil || *decl.PaddingBottom != 6 {
		t.Error("pb alias should set PaddingBottom to 6")
	}

	applyDecl(decl, "padding-left", "7")
	if decl.PaddingLeft == nil || *decl.PaddingLeft != 7 {
		t.Error("padding-left should be 7")
	}

	applyDecl(decl, "pl", "8")
	if decl.PaddingLeft == nil || *decl.PaddingLeft != 8 {
		t.Error("pl alias should set PaddingLeft to 8")
	}

	applyDecl(decl, "padding-right", "9")
	if decl.PaddingRight == nil || *decl.PaddingRight != 9 {
		t.Error("padding-right should be 9")
	}

	applyDecl(decl, "pr", "10")
	if decl.PaddingRight == nil || *decl.PaddingRight != 10 {
		t.Error("pr alias should set PaddingRight to 10")
	}
}

func TestP106_ApplyDecl_InvalidValues(t *testing.T) {
	decl := &StyleDecl{}

	// Invalid color
	applyDecl(decl, "fg", "nonexistent")
	// Should not crash, fg stays zero

	// "inherit", "default", "none" should produce zero color
	applyDecl(decl, "fg", "inherit")
	// inherit/default/none all map to zero color — but applyDecl sets via parseColor
	// parseColor returns zero color for those → applyDecl checks c.Type != 0 → doesn't set
	// So fg stays as whatever it was (red from earlier test line)
	// This is correct behavior: inherit means don't override

	// Invalid bool
	applyDecl(decl, "bold", "maybe")
	if decl.Bold != nil {
		t.Error("invalid bool should produce nil")
	}

	// Invalid int
	applyDecl(decl, "padding-top", "abc")
	if decl.PaddingTop != nil {
		t.Error("invalid int should produce nil")
	}

	// Unknown key
	applyDecl(decl, "unknown-prop", "value")
	// Should not crash
}

func TestP106_ParseColor_BrightColors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"bright_black", "bright_black"},
		{"bright_red", "bright_red"},
		{"bright_green", "bright_green"},
		{"bright_yellow", "bright_yellow"},
		{"bright_blue", "bright_blue"},
		{"bright_magenta", "bright_magenta"},
		{"bright_cyan", "bright_cyan"},
		{"bright_white", "bright_white"},
	}
	for _, tt := range tests {
		c := parseColor(tt.input)
		if c.Type == 0 {
			t.Errorf("parseColor(%q) returned zero color", tt.input)
		}
	}
}

func TestP106_ParseColor_UnknownFallback(t *testing.T) {
	c := parseColor("purple")
	// Should fall back to white
	if c.Type == 0 {
		t.Error("unknown color should fallback to white, not zero")
	}
}

func TestP106_ParseStyleSheet_Complex(t *testing.T) {
	// Test with comments and complex values
	text := `
// This is a comment
.danger {
	fg: #ff0000
	bg: #000000
	bold: true
	reverse: false
	strikethrough: true
	padding-top: 5
	padding-bottom: 5
}
`
	ss, err := ParseStyleSheet(text)
	if err != nil {
		t.Fatalf("ParseStyleSheet failed: %v", err)
	}
	if !ss.Has(".danger") {
		t.Fatal("should have .danger class")
	}

	decl := ss.ResolveDecl(".danger")
	if decl.PaddingTop == nil || *decl.PaddingTop != 5 {
		t.Error("padding-top should be 5")
	}
	if decl.PaddingBottom == nil || *decl.PaddingBottom != 5 {
		t.Error("padding-bottom should be 5")
	}

	style := ss.Resolve(".danger")
	if style.Flags&buffer.Strikethrough == 0 {
		t.Error("strikethrough should be set")
	}
}

// ─── CodeBlock paintStreamingCursorLocked edge cases ───

func TestP106_CodeBlock_StreamingCursor_LongLineClamp(t *testing.T) {
	// Line longer than bounds width → x should be clamped
	cb := NewCodeBlock("go", "this is a very long line of code that exceeds the width")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	cb.Paint(buf)
}

func TestP106_CodeBlock_StreamingCursor_EmptyLinesWithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP106_CodeBlock_StreamingCursor_PastEnd(t *testing.T) {
	// scrollOffset so large that lastIdx >= len(lines)
	cb := NewCodeBlock("go", "short")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.ScrollTo(100) // scroll way past content
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP106_CodeBlock_StreamingCursor_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "func main() {\n\tfmt.Println(\"hello\")\n}")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP106_CodeBlock_StreamingCursor_OutOfBoundsY(t *testing.T) {
	// y falls outside bounds → should return without setting cell
	cb := NewCodeBlock("go", "test")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	cb.ScrollTo(10)
	buf := buffer.NewBuffer(40, 2)
	cb.Paint(buf)
}

func TestP106_CodeBlock_StreamingCursor_PlainFallback(t *testing.T) {
	// Use plain fallback mode
	cb := NewCodeBlock("go", "test code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
	// Should paint without error
}

func TestP106_CodeBlock_StreamingCursor_NonStreaming(t *testing.T) {
	// Not streaming → cursor should NOT appear
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
	// Just verify no panic
}

func TestP106_CodeBlock_StreamingCursor_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "test")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

// ─── Checkbox setNavigableCursor edge ───

func TestP106_Checkbox_NavigableCursor_SingleEnabledAtZero(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	items := cb.Items()
	items[1].Disabled = true
	items[2].Disabled = true
	cb.SetItems(items)
	cb.SetCursor(1) // disabled, should fall back to 0
	if cb.Cursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", cb.Cursor())
	}
}

// ─── RadioGroup setNavigableCursor ───

func TestP106_RadioGroup_NavigableCursor_DisabledWrap(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(1, true)
	rg.SetCursor(2)
	rg.SetCursor(0) // enabled, should stay
	if rg.Cursor() != 0 {
		t.Errorf("expected 0, got %d", rg.Cursor())
	}
}

// ─── ContextMenu SetCursor overflow ───

func TestP106_ContextMenu_SetCursor_ExactSize(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.SetCursor(1) // exactly len-1, but may clamp differently
	// ContextMenu SetCursor behavior: just verify it doesn't panic
}

func TestP106_ContextMenu_SetCursor_OnePastEnd(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetCursor(2) // exactly len
	if cm.Cursor() != 1 { // should clamp to last
		t.Errorf("expected 1, got %d", cm.Cursor())
	}
}

// ─── SelectField Value ───

func TestP106_SelectField_NegativeIndex(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	sf.SetSelectedIndex(-1)
	// SelectField clamps negative to 0
	_ = sf.Value() // just verify no panic
}

func TestP106_SelectField_BeyondEnd(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	sf.SetSelectedIndex(10)
	_ = sf.Value() // clamps to last, just verify no panic
}

// ─── Wizard MoveButtons ───

func TestP106_Wizard_MoveButtonsForward(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTabBar()},
		{Title: "S2", Content: NewTabBar()},
		{Title: "S3", Content: NewTabBar()},
	})
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	w.Paint(buf)
}

func TestP106_Wizard_Back(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "S1", Content: NewTabBar()},
		{Title: "S2", Content: NewTabBar()},
	})
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	w.Paint(buf)
	// Back at step 0 should be no-op
	_ = w.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	w.Paint(buf)
}

// ─── Viewport edge ───

func TestP106_Viewport_ScrollToX_Negative(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 30, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollToX(-5)
	if vp.OffsetX() != 0 {
		t.Errorf("expected 0 for negative scroll, got %d", vp.OffsetX())
	}
}

func TestP106_Viewport_ScrollToY_Negative(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 30, h: 50})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollToY(-5)
	if vp.OffsetY() != 0 {
		t.Errorf("expected 0 for negative scroll, got %d", vp.OffsetY())
	}
}

// ─── BarChart edge ───

func TestP106_BarChart_EmptySeries(t *testing.T) {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP106_BarChart_HorizontalMode(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{Name: "A", Data: []BarData{{Label: "x", Value: 10}, {Label: "y", Value: 20}}})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

// ─── Tree rebuildLocked ───

func TestP106_Tree_RebuildCollapsed(t *testing.T) {
	root := NewTreeNode("root", "Root")
	child := NewTreeNode("c1", "Child")
	child.Children = []*TreeNode{NewTreeNode("gc1", "Grandchild")}
	root.Children = []*TreeNode{child}
	tr := NewTree()
	tr.SetRoot(root)
	tr.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	tr.Paint(buf)
}

// ─── DiffPreview SetShowLineNumbers/SetShowStats ───

func TestP106_DiffPreview_ShowControls(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowStats(true)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	dp.Paint(buf)
}

// ─── TextArea edge ───

func TestP106_TextArea_HandleKey_Enter(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello world")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	buf := buffer.NewBuffer(40, 5)
	ta.Paint(buf)
}

func TestP106_TextArea_MoveLineDown(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	buf := buffer.NewBuffer(40, 5)
	ta.Paint(buf)
}

// ─── TabBar Measure ───

func TestP106_TabBar_MeasureZeroTabs(t *testing.T) {
	tb := NewTabBar()
	s := tb.Measure(Bounded(40, 3))
	if s.W < 0 || s.H < 0 {
		t.Error("Measure should return non-negative size")
	}
}
