package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- BarChart coverage ---

func TestP57_BarChart_VerticalPaint(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetShowLegend(true)
	bc.SetShowValues(true)
	bc.SetGap(2)
	bc.AddSeries(BarSeries{
		Name: "A",
		Data: []BarData{
			{Label: "x", Value: 10},
			{Label: "y", Value: 20},
			{Label: "z", Value: 30},
		},
	})
	bc.AddSeries(BarSeries{
		Name: "B",
		Data: []BarData{
			{Label: "x", Value: 5},
			{Label: "y", Value: 15},
			{Label: "z", Value: 25},
		},
	})

	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	bc.Paint(buf)
}

func TestP57_BarChart_SetGridStyle(t *testing.T) {
	bc := NewBarChart()
	bc.SetGridStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite)})
	bc.SetAxisStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite)})
}

func TestP57_BarChart_HorizontalMode(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowValues(true)
	bc.SetShowGrid(true)
	bc.AddSeries(BarSeries{
		Name: "Q1",
		Data: []BarData{
			{Label: "Jan", Value: 100},
			{Label: "Feb", Value: 200},
			{Label: "Mar", Value: 150},
		},
	})

	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP57_BarChart_SetMaxVal(t *testing.T) {
	bc := NewBarChart()
	bc.SetMaxVal(500)
	bc.AddSeries(BarSeries{
		Name: "Data",
		Data: []BarData{
			{Label: "a", Value: 100},
			{Label: "b", Value: 200},
			{Label: "c", Value: 300},
		},
	})

	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP57_BarChart_Empty(t *testing.T) {
	bc := NewBarChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	bc.Paint(buf)
}

// --- CodeBlock coverage ---

func TestP57_CodeBlock_TitleLineNumbers(t *testing.T) {
	cb := NewCodeBlock("python", "print('hello')\nx = 1\n")
	cb.SetShowLineNumbers(true)
	cb.SetShowTitle(true)
	cb.SetTitle("test.py")

	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP57_CodeBlock_ScrollTo(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {\n}\n")
	cb.ScrollTo(1)
	if cb.ScrollOffset() != 1 {
		t.Errorf("expected offset 1, got %d", cb.ScrollOffset())
	}
}

// --- DiffViewer coverage ---

func TestP57_DiffViewer_WithHeader(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("+added line\n-removed line\n context line")
	dv.SetShowHeader(true)

	dv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	dv.Paint(buf)
}

// --- Sparkline coverage ---

func TestP57_Sparkline_AutoScale(t *testing.T) {
	sl := NewSparkline()
	sl.SetAutoScale(true)
	sl.SetData([]float64{1, 5, 3, 8, 2, 6, 4})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	sl.Paint(buf)
}

// --- Table coverage ---

func TestP57_Table_LongHeaders(t *testing.T) {
	tbl := NewTable(
		[]string{"VeryLongColumnName1", "VeryLongColumnName2", "VeryLongColumnName3"},
		[]string{"a", "b", "c"},
	)
	tbl.SetSelectedRow(0)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tbl.Paint(buf)
}

// --- Dialog all types ---

func TestP57_Dialog_AllTypes(t *testing.T) {
	for _, dt := range []DialogType{DialogInfo, DialogConfirm, DialogPrompt, DialogCustom} {
		d := NewDialog(dt, "Test", "Message body")
		d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
		buf := buffer.NewBuffer(40, 10)
		d.Paint(buf)
	}
}

// --- ContextMenu cursor clamping ---

func TestP57_ContextMenu_CursorClamp(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("save", "Save"))
	cm.AddItem(NewMenuItem("load", "Load"))
	cm.AddItem(NewMenuItem("quit", "Quit"))
	cm.SetCursor(5)
	if cm.Cursor() > 2 {
		t.Errorf("expected cursor clamped to 2, got %d", cm.Cursor())
	}
}

// --- FilePicker key handling ---

func TestP57_FilePicker_KeyHandling(t *testing.T) {
	fp := NewFilePicker(".")
	fp.HandleKey(&term.KeyEvent{Rune: 'x'})
	fp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	fp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// --- Gauge ---

func TestP57_Gauge_HalfAndFull(t *testing.T) {
	g := NewGauge()
	g.SetValue(0.5)
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	g.Paint(buf)

	g.SetValue(1.0)
	g.Paint(buf)
}

// --- ProgressBar ---

func TestP57_ProgressBar_Partial(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(0.3)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf)
}

// --- Checkbox navigation ---

func TestP57_Checkbox_Navigate(t *testing.T) {
	cb := NewCheckbox([]string{"Item 1", "Item 2", "Item 3"})
	cb.MoveDown()
	if cb.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", cb.Cursor())
	}
	cb.MoveDown()
	if cb.Cursor() != 2 {
		t.Errorf("expected cursor 2, got %d", cb.Cursor())
	}
	cb.MoveDown()
	// Should wrap around
	if cb.Cursor() != 0 {
		t.Errorf("expected cursor 0 (wrap), got %d", cb.Cursor())
	}
}

// --- BaseComponent empty ---

func TestP57_BaseComponent_Empty(t *testing.T) {
	bc := &BaseComponent{}
	s := bc.Measure(Constraints{})
	if s.W != 0 || s.H != 0 {
		t.Errorf("expected 0x0, got %dx%d", s.W, s.H)
	}
	bc.Paint(buffer.NewBuffer(10, 5))
}

// --- LineChart markers ---

func TestP57_LineChart_DotMarker(t *testing.T) {
	lc := NewLineChart()
	lc.SetShowGrid(true)
	lc.SetShowLegend(true)
	lc.AddSeries(ChartSeries{
		Name:   "dots",
		Data:   []ChartPoint{{X: 0, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 3}},
		Color:  buffer.NamedColor(buffer.NamedCyan),
		Marker: ChartMarkerDot,
	})

	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	lc.Paint(buf)
}

func TestP57_LineChart_StarMarker(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Name:   "stars",
		Data:   []ChartPoint{{X: 0, Y: 0}, {X: 5, Y: 5}},
		Color:  buffer.NamedColor(buffer.NamedGreen),
		Marker: ChartMarkerStar,
	})

	lc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	lc.Paint(buf)
}

// --- SplitPane edge cases ---

func TestP57_SplitPane_NegativeRatio(t *testing.T) {
	sp := NewSplitPane(
		&BaseComponent{},
		&BaseComponent{},
	)
	sp.SetRatio(-0.5) // negative → should clamp
	sp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	sp.Paint(buf)
}

// --- HelpOverlay ---

func TestP57_HelpOverlay_Render(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{
			Name: "Navigation",
			Entries: []HelpEntry{
				{Keys: "j/k", Description: "Move up/down"},
				{Keys: "g/G", Description: "Top/Bottom"},
			},
		},
	})

	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ho.Paint(buf)
}

// --- Form ---

func TestP57_Form_Render(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "Name", ""))
	f.AddField(NewSelectField("color", "Color", []string{"Red", "Green", "Blue"}))

	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	f.Paint(buf)
}

// --- Tooltip ---

func TestP57_Tooltip_Render(t *testing.T) {
	tt := NewTooltip("This is a helpful tip")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	tt.Paint(buf)
}

// --- TextArea ---

func TestP57_TextArea_Render(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello\nWorld")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ta.Paint(buf)
}
