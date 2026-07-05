package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/markdown"
	"github.com/topcheer/fluui/theme"
)

// --- StatusBar coverage ---

func TestP42_StatusBar_Items(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("status", "Ready")
	sb.AddCenter("mode", "INSERT")
	sb.AddRight("time", "12:00")

	if sb.LeftItems() == "" {
		t.Error("expected non-empty left items")
	}
	if sb.CenterItems() == "" {
		t.Error("expected non-empty center items")
	}
	if sb.RightItems() == "" {
		t.Error("expected non-empty right items")
	}
}

func TestP42_StatusBar_String(t *testing.T) {
	sb := NewStatusBar()
	if sb.String() != "StatusBar" {
		t.Errorf("expected 'StatusBar', got %q", sb.String())
	}
}

// --- SplitPane coverage ---

func makeTestSplitPane(dir SplitDirection) *SplitPane {
	child1 := NewText("left")
	child2 := NewText("right")
	sp := NewSplitPane(child1, child2)
	sp.SetDirection(dir)
	sp.SetRatio(0.5)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	return sp
}

func TestP42_SplitPane_HandleMouse_Horizontal(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)

	// Paint to cache dividerPos
	buf := buffer.NewBuffer(80, 24)
	sp.Paint(buf)

	divPos := sp.DividerPos()
	handled := sp.HandleMouse(divPos, 5, 1)
	if !handled {
		t.Error("expected click on divider to start drag")
	}
	handled = sp.HandleMouse(divPos+10, 5, 0)
	if !handled {
		t.Error("expected drag to be handled")
	}
	handled = sp.HandleMouse(divPos+10, 5, -1)
	if !handled {
		t.Error("expected release to be handled")
	}
}

func TestP42_SplitPane_HandleMouse_Vertical(t *testing.T) {
	sp := makeTestSplitPane(SplitVertical)

	// Paint to cache dividerPos
	buf := buffer.NewBuffer(80, 24)
	sp.Paint(buf)

	divPos := sp.DividerPos()
	handled := sp.HandleMouse(5, divPos, 1)
	if !handled {
		t.Error("expected click on divider to start drag")
	}
	handled = sp.HandleMouse(5, divPos+5, 0)
	if !handled {
		t.Error("expected drag to be handled")
	}
	handled = sp.HandleMouse(5, divPos+5, -1)
	if !handled {
		t.Error("expected release to be handled")
	}
}

func TestP42_SplitPane_HandleMouse_NoDrag(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)

	handled := sp.HandleMouse(5, 5, 1)
	if handled {
		t.Error("expected click away from divider to not be handled")
	}
}

func TestP42_SplitPane_HandleKey_Horizontal(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)

	// Ctrl+Shift+Left = mods (1|4)=5, key=11
	handled := sp.HandleKey(11, 5)
	if !handled {
		t.Error("expected Ctrl+Shift+Left to be handled")
	}
	// Ctrl+Shift+Right = key=12
	handled = sp.HandleKey(12, 5)
	if !handled {
		t.Error("expected Ctrl+Shift+Right to be handled")
	}
}

func TestP42_SplitPane_HandleKey_Vertical(t *testing.T) {
	sp := makeTestSplitPane(SplitVertical)

	handled := sp.HandleKey(9, 5) // Up
	if !handled {
		t.Error("expected Ctrl+Shift+Up to be handled")
	}
	handled = sp.HandleKey(10, 5) // Down
	if !handled {
		t.Error("expected Ctrl+Shift+Down to be handled")
	}
}

func TestP42_SplitPane_HandleKey_NoMods(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)
	handled := sp.HandleKey(11, 0)
	if handled {
		t.Error("expected unhandled without Ctrl+Shift")
	}
}

func TestP42_SplitPane_DividerPos(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)
	dp := sp.DividerPos()
	if dp <= 0 {
		t.Errorf("expected positive divider pos, got %d", dp)
	}
}

func TestP42_SplitPane_StartUpdateEndDrag(t *testing.T) {
	sp := makeTestSplitPane(SplitHorizontal)
	sp.StartDrag(40)
	if !sp.IsDragging() {
		t.Error("expected dragging after StartDrag")
	}
	sp.UpdateDrag(45)
	sp.EndDrag()
	if sp.IsDragging() {
		t.Error("expected not dragging after EndDrag")
	}
}

// --- BarChart coverage ---

func TestP42_BarChart_SetTheme(t *testing.T) {
	bc := NewBarChart()
	bc.SetTheme(theme.Dracula())
}

func TestP42_BarChart_SetGridStyle(t *testing.T) {
	bc := NewBarChart()
	bc.SetGridStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)})
}

func TestP42_BarChart_SetAxisStyle(t *testing.T) {
	bc := NewBarChart()
	bc.SetAxisStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan)})
}

// --- CodeBlock coverage ---

func TestP42_CodeBlock_SetHighlighter(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetHighlighter(markdown.NewHighlighter())
}

func TestP42_CodeBlock_SetTheme(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetTheme(theme.Dracula())
}

func TestP42_CodeBlock_ScrollTo(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.ScrollTo(2)
	if cb.ScrollOffset() != 2 {
		t.Errorf("expected offset 2, got %d", cb.ScrollOffset())
	}
	cb.ScrollTo(-5)
	if cb.ScrollOffset() != 0 {
		t.Errorf("expected clamped to 0, got %d", cb.ScrollOffset())
	}
}

func TestP42_CodeBlock_SourceLanguage(t *testing.T) {
	cb := NewCodeBlock("python", "print('hello')")
	if cb.Source() != "print('hello')" {
		t.Errorf("unexpected source: %q", cb.Source())
	}
	if cb.Language() != "python" {
		t.Errorf("expected 'python', got %q", cb.Language())
	}
}

// --- DiffViewer coverage ---

func TestP42_DiffViewer_SetShowHeader(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetShowHeader(false)
	dv.SetShowHeader(true)
}

// --- SparkLine coverage ---

func TestP42_SparkLine_AutoScale(t *testing.T) {
	sl := NewSparkline()
	sl.SetAutoScale(true)
	sl.SetAutoScale(false)
}

// --- Tree coverage ---

func TestP42_Tree_Navigate(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	root.AddChild(NewTreeNode("child1", "Child 1"))
	root.AddChild(NewTreeNode("child2", "Child 2"))
	tree.SetRoot(root)

	tree.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	tree.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	tree.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
}

// --- Table coverage ---

func TestP42_Table_BasicOps(t *testing.T) {
	tbl := NewTable([]string{"A", "B"}, []string{"1", "2"}, []string{"3", "4"})
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.RowCount())
	}
	tbl.SetSelectedRow(0)
	if tbl.SelectedRow() != 0 {
		t.Errorf("expected selected 0, got %d", tbl.SelectedRow())
	}
}

// --- HelpOverlay coverage ---

func TestP42_HelpOverlay_Toggle(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Navigation", Entries: []HelpEntry{{Keys: "j", Description: "Down"}}},
	}
	h := NewHelpOverlay(groups)
	h.SetQuery("nav")
	if h.Query() != "nav" {
		t.Error("expected query 'nav'")
	}
	h.SelectNext()
	h.SelectPrev()
}

// --- Form coverage ---

func TestP42_Form_AddField(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("Name:", "name", ""))
	f.AddField(NewTextField("Email:", "email", ""))
	if f.FieldCount() != 2 {
		t.Errorf("expected 2 fields, got %d", f.FieldCount())
	}
}

// --- Dialog coverage ---

func TestP42_Dialog_Type(t *testing.T) {
	d := NewDialog(DialogInfo, "Title", "Message")
	if d.Type() != DialogInfo {
		t.Error("expected DialogInfo")
	}
	d2 := NewDialog(DialogConfirm, "Confirm", "Are you sure?")
	if d2.Type() != DialogConfirm {
		t.Error("expected DialogConfirm")
	}
}

// --- Badge coverage ---

func TestP42_Badge_Text(t *testing.T) {
	b := NewBadge("New", BadgeSuccess)
	if b.Text() != "New" {
		t.Errorf("expected 'New', got %q", b.Text())
	}
	b.SetText("Updated")
	if b.Text() != "Updated" {
		t.Errorf("expected 'Updated', got %q", b.Text())
	}
}

// --- TextArea coverage ---

func TestP42_TextArea_SetText(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello\nWorld")
	if ta.Text() != "Hello\nWorld" {
		t.Errorf("unexpected text: %q", ta.Text())
	}
}

// --- ContextMenu coverage ---

func TestP42_ContextMenu_Navigate(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("copy", "Copy"))
	cm.AddItem(NewMenuItem("paste", "Paste"))
	cm.AddItem(NewMenuItem("delete", "Delete"))

	cm.MoveDown()
	if cm.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", cm.Cursor())
	}
	cm.MoveDown()
	cm.MoveUp()
}

// --- FilePicker coverage ---

func TestP42_FilePicker_Navigate(t *testing.T) {
	fp := NewFilePicker(".")
	fp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	fp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// --- Tooltip coverage ---

func TestP42_Tooltip_Text(t *testing.T) {
	tp := NewTooltip("Click here")
	if tp.Text() != "Click here" {
		t.Errorf("expected 'Click here', got %q", tp.Text())
	}
}

// --- Gauge coverage ---

func TestP42_Gauge_Value(t *testing.T) {
	g := NewGauge()
	g.SetValue(0.75)
	if g.Value() != 0.75 {
		t.Errorf("expected 0.75, got %f", g.Value())
	}
}
