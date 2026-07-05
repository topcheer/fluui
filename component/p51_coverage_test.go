package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- CodeBlock plainLinesLocked ---

func TestP51_CodeBlock_PlainLinesFallback(t *testing.T) {
	cb := NewCodeBlock("text", "hello\nworld")
	cb.SetHighlighter(nil) // force plain text rendering
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	b := buffer.NewBuffer(20, 5)
	cb.Paint(b)
	// Just verify no panic
}

// --- DiffPreview ---

func TestP51_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
}

func TestP51_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

// --- LineChart formatSeriesValue ---

func TestP51_LineChart_FormatValue(t *testing.T) {
	lc := NewLineChart()
	// formatSeriesValue is private, test through Paint
	lc.AddSeries(ChartSeries{
		Name: "test",
		Data: []ChartPoint{
			{X: 0, Y: 1.5},
			{X: 1, Y: 2.5},
		},
		Color:  buffer.NamedColor(buffer.NamedCyan),
		Marker: ChartMarkerDot,
	})
	lc.SetShowLegend(true)
	lc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	b := buffer.NewBuffer(40, 10)
	lc.Paint(b)
}

// --- Notification truncateString ---

func TestP51_Notification_LongText(t *testing.T) {
	nm := NewToastManager(3)
	nm.PushInfo("This is a very long notification message that should be truncated when displayed in a narrow terminal window", "detail")
	nm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	b := buffer.NewBuffer(20, 5)
	nm.Paint(b)
}

// --- ScrollView Children ---

func TestP51_ScrollView_Children(t *testing.T) {
	inner := NewStatusBar()
	sv := NewScrollView(inner)
	children := sv.Children()
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}
}

// --- SparkLine formatFloat ---

func TestP51_SparkLine_FormatValues(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{1.5, 2.7, 3.14, 99.99})
	sl.SetAutoScale(true)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	b := buffer.NewBuffer(20, 3)
	sl.Paint(b)
}

// --- SplitPane abs ---

func TestP51_SplitPane_NegativeDrag(t *testing.T) {
	sp := NewSplitPane(NewStatusBar(), NewStatusBar())
	sp.SetDirection(SplitVertical)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	// Drag to negative position — abs() should handle it
	sp.SetRatio(-0.5)
	sp.SetRatio(0.5) // restore
}

// --- Table truncateToWidth ---

func TestP51_Table_LongHeaders(t *testing.T) {
	t1 := NewTable([]string{"VeryLongHeaderThatExceedsWidth", "B"}, []string{"val", "val"})
	t1.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	b := buffer.NewBuffer(10, 5)
	t1.Paint(b)
}

// --- Dialog dialogTypeString ---

func TestP51_Dialog_AllTypes(t *testing.T) {
	for _, dt := range []DialogType{DialogInfo, DialogConfirm, DialogPrompt, DialogCustom} {
		d := NewDialog(dt, "Title", "Message")
		d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
		b := buffer.NewBuffer(40, 10)
		d.Paint(b)
	}
}

// --- ContextMenu ---

func TestP51_ContextMenu_SetCursor(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.AddItem(NewMenuItem("item2", "Item 2"))
	cm.SetCursor(0)
	if cm.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", cm.Cursor())
	}
	cm.SetCursor(1)
	if cm.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", cm.Cursor())
	}
	// Out of bounds should clamp
	cm.SetCursor(100)
	if cm.Cursor() > 1 {
		t.Errorf("expected clamped cursor, got %d", cm.Cursor())
	}
}

// --- BaseComponent Measure/Paint ---

func TestP51_BaseComponent_EmptyMeasure(t *testing.T) {
	bc := &BaseComponent{}
	s := bc.Measure(Unbounded())
	if s.W != 0 || s.H != 0 {
		t.Errorf("expected 0x0, got %dx%d", s.W, s.H)
	}
}

func TestP51_BaseComponent_EmptyPaint(t *testing.T) {
	bc := &BaseComponent{}
	b := buffer.NewBuffer(10, 5)
	bc.Paint(b) // should be no-op
}

// --- FilePicker handleFilterKey ---

func TestP51_FilePicker_FilterKey(t *testing.T) {
	fp := NewFilePicker(".")
	// Type into filter — handleFilterKey processes printable chars
	fp.HandleKey(&term.KeyEvent{Rune: 'a', Key: term.KeyUnknown})
	fp.HandleKey(&term.KeyEvent{Rune: 'b', Key: term.KeyUnknown})
	fp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
}

// --- Gauge edge cases ---

func TestP51_Gauge_ZeroAndFull(t *testing.T) {
	g := NewGauge()
	g.SetValue(0)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	b := buffer.NewBuffer(20, 1)
	g.Paint(b)

	g.SetValue(1.0)
	g.Paint(b)

	g.SetValue(0.5)
	g.Paint(b)
}

// --- ProgressBar ---

func TestP51_ProgressBar_Values(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(0)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	b := buffer.NewBuffer(20, 1)
	pb.Paint(b)

	pb.SetProgress(0.5)
	pb.Paint(b)

	pb.SetProgress(1.0)
	pb.Paint(b)
}
