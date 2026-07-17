package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestBarChart_PaintVertical_WithGrid_P265(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarVertical)
	bc.SetShowGrid(true)
	bc.SetTitle("Test Chart")
	bc.SetSeries([]BarSeries{{
		Name: "Q1",
		Data: []BarData{
			{Label: "A", Value: 30},
			{Label: "B", Value: 50},
			{Label: "C", Value: 80},
		},
	}})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	bc.Paint(buf)
}

func TestBarChart_PaintVertical_NegativeValue_P265(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarVertical)
	bc.SetSeries([]BarSeries{{
		Name: "Q1",
		Data: []BarData{
			{Label: "A", Value: -10},
			{Label: "B", Value: 20},
		},
	}})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	bc.Paint(buf)
}

func TestDiffViewer_Measure_ClampWidth_P265(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("@@ -1,3 +1,3 @@\n-old\n+new\n context\n")
	s := dv.Measure(Constraints{MaxWidth: 20, MaxHeight: 3})
	if s.W > 20 {
		t.Errorf("width should be clamped to 20, got %d", s.W)
	}
}

func TestDiffViewer_Paint_WithLineNumbers_P265(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetShowLineNumbers(true)
	dv.SetContent("@@ -1,3 +1,4 @@\n context\n-old line\n+new line\n+added\n context2\n")
	dv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	dv.Paint(buf)
}

func TestDiffViewer_Paint_WithHeader_P265(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetShowHeader(true)
	dv.SetTitle("app.go")
	dv.SetContent("@@ -1,2 +1,2 @@\n-a\n+b\n")
	dv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	dv.Paint(buf)
}

func TestCommandPalette_ClampScroll_Navigate_P265(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{
		{Label: "cmd1"}, {Label: "cmd2"}, {Label: "cmd3"},
	})
	cp.SetQuery("cmd")
	cp.Show(0, 0)
	cp.SetMaxVisible(2)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}
