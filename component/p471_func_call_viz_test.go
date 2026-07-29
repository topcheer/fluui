package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestFunctionCallVisualizerBasic(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("search_web", `{"q":"go"}`, 120*time.Millisecond, CallSuccess)

	if fcv.CallCount() != 1 {
		t.Errorf("CallCount = %d, want 1", fcv.CallCount())
	}
}

func TestFunctionCallVisualizerMultiple(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("search_web", `{"q":"go tui"}`, 120*time.Millisecond, CallSuccess)
	fcv.AddCall("read_file", `{"path":"main.go"}`, 5*time.Millisecond, CallSuccess)
	fcv.AddCall("write_file", `{"path":"out.go"}`, 15*time.Millisecond, CallError)

	if fcv.CallCount() != 3 {
		t.Errorf("CallCount = %d, want 3", fcv.CallCount())
	}
}

func TestFunctionCallVisualizerNested(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("outer", `{}`, 100*time.Millisecond, CallSuccess)
	fcv.AddNestedCall("inner_1", `{}`, 30*time.Millisecond, CallSuccess, 2)
	fcv.AddNestedCall("inner_2", `{}`, 40*time.Millisecond, CallError, 2)

	if fcv.CallCount() != 3 {
		t.Errorf("CallCount = %d, want 3", fcv.CallCount())
	}
}

func TestFunctionCallVisualizerClear(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("a", `{}`, 10*time.Millisecond, CallSuccess)
	fcv.AddCall("b", `{}`, 20*time.Millisecond, CallSuccess)
	if fcv.CallCount() != 2 {
		t.Fatalf("CallCount before clear = %d, want 2", fcv.CallCount())
	}
	fcv.Clear()
	if fcv.CallCount() != 0 {
		t.Errorf("CallCount after clear = %d, want 0", fcv.CallCount())
	}
}

func TestFunctionCallVisualizerMeasure(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("test", `{}`, 10*time.Millisecond, CallSuccess)
	s := fcv.Measure(Constraints{})
	if s.W < 20 {
		t.Errorf("W = %d, want >= 20", s.W)
	}
	if s.H < 3 {
		t.Errorf("H = %d, want >= 3", s.H)
	}
}

func TestFunctionCallVisualizerPaint(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("search_web", `{"q":"go tui"}`, 120*time.Millisecond, CallSuccess)
	fcv.AddCall("read_file", `{"path":"main.go"}`, 5*time.Millisecond, CallError)
	fcv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})

	buf := buffer.NewBuffer(60, 5)
	fcv.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	// Check a success icon exists somewhere
	foundCheck := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 1).Rune == '✓' {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Error("success icon not found")
	}
}

func TestFunctionCallVisualizerPaintEmpty(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	fcv.Paint(buf) // should not panic
}

func TestFunctionCallVisualizerPaintNested(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.AddCall("outer", `{}`, 100*time.Millisecond, CallRunning)
	fcv.AddNestedCall("inner", `{}`, 30*time.Millisecond, CallSuccess, 2)
	fcv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	fcv.Paint(buf)

	// Should find pipe indent chars on row 2
	foundIndent := false
	for x := 0; x < 10; x++ {
		if buf.GetCell(x, 2).Rune == '│' {
			foundIndent = true
			break
		}
	}
	if !foundIndent {
		t.Error("indent pipe not found for nested call")
	}
}

func TestFunctionCallVisualizerChildren(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	if fcv.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestFunctionCallVisualizerStyle(t *testing.T) {
	fcv := NewFunctionCallVisualizer()
	fcv.SetStyle(FuncCallVizStyle{
		Name:     buffer.Style{Fg: buffer.RGB(255, 0, 0), Flags: buffer.Bold},
		Args:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Duration: buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Status:   [3]buffer.Style{{}, {}, {}},
		Indent:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Border:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	fcv.AddCall("test", `{}`, 5*time.Millisecond, CallSuccess)
	fcv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	fcv.Paint(buf)
}

func TestCallStatusHelpers(t *testing.T) {
	if callStatusIndex(CallRunning) != 0 {
		t.Error("running index should be 0")
	}
	if callStatusIndex(CallSuccess) != 1 {
		t.Error("success index should be 1")
	}
	if callStatusIndex(CallError) != 2 {
		t.Error("error index should be 2")
	}
	if callStatusIcon(CallSuccess) != '✓' {
		t.Error("success icon should be ✓")
	}
	if callStatusIcon(CallError) != '✗' {
		t.Error("error icon should be ✗")
	}
	if callStatusIcon(CallRunning) != '⟳' {
		t.Error("running icon should be ⟳")
	}
}
