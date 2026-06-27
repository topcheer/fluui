package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewSplitPane(t *testing.T) {
	first := NewText("left")
	second := NewText("right")
	sp := NewSplitPane(first, second)
	if sp.Direction() != SplitHorizontal {
		t.Errorf("expected horizontal, got vertical")
	}
	if sp.Ratio() != 0.5 {
		t.Errorf("expected ratio 0.5, got %f", sp.Ratio())
	}
	if sp.First() != first {
		t.Errorf("first mismatch")
	}
	if sp.Second() != second {
		t.Errorf("second mismatch")
	}
}

func TestSplitPane_SetDirection(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetDirection(SplitVertical)
	if sp.Direction() != SplitVertical {
		t.Errorf("expected vertical")
	}
	sp.SetDirection(SplitHorizontal)
	if sp.Direction() != SplitHorizontal {
		t.Errorf("expected horizontal")
	}
}

func TestSplitPane_SetRatio(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetMinRatio(0.2)
	sp.SetMaxRatio(0.8)
	cases := []struct{ in, want float64 }{
		{0.3, 0.3}, {0.0, 0.2}, {1.0, 0.8}, {0.5, 0.5}, {-0.1, 0.2}, {1.5, 0.8},
	}
	for _, c := range cases {
		sp.SetRatio(c.in)
		if got := sp.Ratio(); got != c.want {
			t.Errorf("SetRatio(%f)=%f want %f", c.in, got, c.want)
		}
	}
}

func TestSplitPane_SetMinRatio(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetRatio(0.5)
	sp.SetMinRatio(0.6)
	if sp.Ratio() != 0.6 {
		t.Errorf("expected 0.6, got %f", sp.Ratio())
	}
}

func TestSplitPane_SetMaxRatio(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetRatio(0.5)
	sp.SetMaxRatio(0.3)
	if sp.Ratio() != 0.3 {
		t.Errorf("expected 0.3, got %f", sp.Ratio())
	}
}

func TestSplitPane_Children(t *testing.T) {
	a := NewText("a")
	b := NewText("b")
	sp := NewSplitPane(a, b)
	kids := sp.Children()
	if len(kids) != 2 || kids[0] != a || kids[1] != b {
		t.Errorf("children mismatch")
	}
}

func TestSplitPane_SetFirstSecond(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	nf := NewText("X")
	ns := NewText("Y")
	sp.SetFirst(nf)
	sp.SetSecond(ns)
	if sp.First() != nf || sp.Second() != ns {
		t.Errorf("set mismatch")
	}
}

func TestSplitPane_DragHorizontal(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	sp.SetRatio(0.5)
	sp.StartDrag(49)
	if !sp.IsDragging() {
		t.Error("expected dragging")
	}
	sp.UpdateDrag(59)
	if sp.Ratio() <= 0.5 {
		t.Errorf("expected ratio > 0.5, got %f", sp.Ratio())
	}
	sp.EndDrag()
	if sp.IsDragging() {
		t.Error("expected not dragging")
	}
}

func TestSplitPane_DragVertical(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetDirection(SplitVertical)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 100})
	sp.SetRatio(0.5)
	sp.StartDrag(49)
	sp.UpdateDrag(39)
	if sp.Ratio() >= 0.5 {
		t.Errorf("expected ratio < 0.5, got %f", sp.Ratio())
	}
	sp.EndDrag()
}

func TestSplitPane_DragClamped(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	sp.SetMinRatio(0.3)
	sp.SetMaxRatio(0.7)
	sp.SetRatio(0.5)
	sp.StartDrag(49)
	sp.UpdateDrag(200)
	if r := sp.Ratio(); r > 0.7 {
		t.Errorf("clamped at 0.7, got %f", r)
	}
	sp.EndDrag()
}

func TestSplitPane_DragNotDragging(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.UpdateDrag(50)
	if sp.Ratio() != 0.5 {
		t.Errorf("ratio should not change")
	}
}

func TestSplitPane_DragZeroBounds(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	sp.StartDrag(0)
	sp.UpdateDrag(10)
	if sp.Ratio() != 0.5 {
		t.Errorf("ratio should not change with zero bounds")
	}
	sp.EndDrag()
}

func TestSplitPane_MeasureHorizontal(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sz := sp.Measure(Bounded(80, 24))
	if sz.W != 80 || sz.H != 24 {
		t.Errorf("expected 80x24, got %dx%d", sz.W, sz.H)
	}
}

func TestSplitPane_MeasureVertical(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetDirection(SplitVertical)
	sz := sp.Measure(Bounded(60, 30))
	if sz.W != 60 || sz.H != 30 {
		t.Errorf("expected 60x30, got %dx%d", sz.W, sz.H)
	}
}

func TestSplitPane_MeasureDefaults(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sz := sp.Measure(Unbounded())
	if sz.W <= 0 || sz.H <= 0 {
		t.Errorf("expected non-zero size")
	}
}

func TestSplitPane_PaintHorizontal(t *testing.T) {
	first := NewText("HELLO")
	second := NewText("WORLD")
	sp := NewSplitPane(first, second)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	sp.SetRatio(0.5)
	sp.SetShowHandle(false)
	buf := newTestBuffer(20, 1)
	sp.Paint(buf)
	got := cellRunes(buf, 0, 0, 9)
	if got != "HELLO" {
		t.Errorf("left: got %q", got)
	}
	if c := buf.GetCell(9, 0); c.Rune != '│' {
		t.Errorf("divider: got %q want │", string(c.Rune))
	}
	got = cellRunes(buf, 10, 0, 9)
	if got != "WORLD" {
		t.Errorf("right: got %q", got)
	}
}

func TestSplitPane_PaintVertical(t *testing.T) {
	first := NewText("TOP")
	second := NewText("BOT")
	sp := NewSplitPane(first, second)
	sp.SetDirection(SplitVertical)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	sp.SetRatio(0.5)
	sp.SetShowHandle(false)
	buf := newTestBuffer(10, 5)
	sp.Paint(buf)
	for x := 0; x < 10; x++ {
		if c := buf.GetCell(x, 2); c.Rune != '─' {
			t.Errorf("divider (%d,2): got %q", x, string(c.Rune))
			break
		}
	}
}

func TestSplitPane_PaintZeroBounds(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := newTestBuffer(10, 10)
	sp.Paint(buf)
}

func TestSplitPane_PaintHandle(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	sp.SetShowHandle(true)
	sp.SetRatio(0.5)
	buf := newTestBuffer(20, 10)
	sp.Paint(buf)
	if c := buf.GetCell(9, 5); c.Rune != '◆' {
		t.Errorf("handle: got %q", string(c.Rune))
	}
}

func TestSplitPane_PaintNilChildren(t *testing.T) {
	sp := NewSplitPane(nil, nil)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := newTestBuffer(20, 5)
	sp.Paint(buf)
}

func TestSplitPane_SetDividerStyle(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetDividerStyle(buffer.Style{})
}

func TestSplitPane_Concurrent(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				sp.SetRatio(float64(j%10) / 10.0)
				_ = sp.Ratio()
				_ = sp.IsDragging()
			}
		}()
	}
	wg.Wait()
}

func TestSplitPane_ConcurrentPaint(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := newTestBuffer(50, 10)
				sp.Paint(buf)
			}
		}()
	}
	wg.Wait()
}

func TestSplitPane_DragToMax(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	sp.SetMaxRatio(0.9)
	sp.SetRatio(0.5)
	sp.StartDrag(49)
	sp.UpdateDrag(95)
	if r := sp.Ratio(); r > 0.91 {
		t.Errorf("clamped near max, got %f", r)
	}
	sp.EndDrag()
}

func TestSplitPane_DragToMin(t *testing.T) {
	sp := NewSplitPane(NewText("a"), NewText("b"))
	sp.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	sp.SetMinRatio(0.1)
	sp.SetRatio(0.5)
	sp.StartDrag(49)
	sp.UpdateDrag(0)
	if r := sp.Ratio(); r < 0.09 {
		t.Errorf("clamped near min, got %f", r)
	}
	sp.EndDrag()
}
