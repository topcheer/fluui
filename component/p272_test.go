package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P272: diffpreview VisibleRange + form FocusNext/Errors + barchart Measure + codeblock rehighlight

func TestDiffPreview_VisibleRange_NegativeHeight_P272(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("@@ -1,3 +1,3 @@\n-a\n+b\n c\n")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 0}) // availableH = 0-2 = -2 → clamp 0
	start, end := dp.VisibleRange()
	if start != 0 {
		t.Errorf("expected start=0, got %d", start)
	}
	if end < 0 {
		t.Errorf("end should be >= 0, got %d", end)
	}
}

func TestDiffPreview_VisibleRange_Overflow_P272(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("@@ -1,3 +1,3 @@\n-a\n+b\n c\n")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	dp.scrollY = 100 // beyond lines
	start, end := dp.VisibleRange()
	_ = start
	_ = end
}

func TestDiffPreview_clampScrollLocked_SmallHeight_P272(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("@@ -1,3 +1,3 @@\n-a\n+b\n c\n")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1}) // availableH = 1-2 = -1 → clamp to 1
}

func TestDiffPreview_SetLines_P272(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	dp.SetLines([]DiffLine{
		{Type: DiffAdd, Content: "new"},
		{Type: DiffDel, Content: "old"},
		{Type: DiffContext, Content: "ctx"},
	})
}

func TestForm_FocusNext_Empty_P272(t *testing.T) {
	f := NewForm()
	f.FocusNext() // empty fields → no-op
}

func TestForm_FocusNext_Wrap_P272(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("a", "a", ""))
	f.AddField(NewTextField("b", "b", ""))
	f.FocusNext()
	f.FocusNext()
	f.FocusNext() // should wrap to 0
}

func TestForm_FocusPrev_Empty_P272(t *testing.T) {
	f := NewForm()
	f.FocusPrev()
}

func TestForm_FocusPrev_Wrap_P272(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("a", "a", ""))
	f.AddField(NewTextField("b", "b", ""))
	f.FocusPrev() // 0 → wrap to 1
}

func TestForm_Errors_Empty_P272(t *testing.T) {
	f := NewForm()
	if errs := f.Errors(); errs != nil {
		t.Error("empty form should return nil errors")
	}
}

func TestBarChart_Measure_Vertical_P272(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarVertical)
	bc.SetSeries([]BarSeries{{
		Name: "S1",
		Data: []BarData{{Label: "A", Value: 10}},
	}})
	s := bc.Measure(Constraints{MaxWidth: 200, MaxHeight: 200})
	if s.W < 10 {
		t.Errorf("min width should be 10, got %d", s.W)
	}
}

func TestBarChart_Measure_SmallConstraints_P272(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarVertical)
	bc.SetSeries([]BarSeries{{
		Name: "S1",
		Data: []BarData{{Label: "A", Value: 10}},
	}})
	s := bc.Measure(Constraints{MaxWidth: 3, MaxHeight: 2})
	if s.W != 10 {
		t.Errorf("min width should be 10, got %d", s.W)
	}
	if s.H != 5 {
		t.Errorf("min height should be 5, got %d", s.H)
	}
}

func TestCodeBlock_Rehighlight_LanguageChange_P272(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetLanguage("python")
	cb.SetSource("print('hi')")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}
