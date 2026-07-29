package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AsciiArtBox Tests ───

func TestAsciiArtBasic(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetText("123")
	if txt := aa.Text(); txt != "123" {
		t.Errorf("Text = %q, want '123'", txt)
	}
}

func TestAsciiArtEmpty(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetText("")
	if txt := aa.Text(); txt != "" {
		t.Errorf("Text = %q, want ''", txt)
	}
}

func TestAsciiArtWithSpaces(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetText("1 2 3")
	// Should not panic
}

func TestAsciiArtUnknownChars(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetText("ABC")
	// Should render ░ blocks for unknown letters
}

func TestAsciiArtPaint(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetText("42")
	aa.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	aa.Paint(buf)
	// Should have block characters in first row
	hasBlock := false
	for i := 0; i < 20; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '█' || r == '░' {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Error("Paint should show block characters")
	}
}

func TestAsciiArtChildren(t *testing.T) {
	aa := NewAsciiArtBox()
	if c := aa.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAsciiArtStyle(t *testing.T) {
	aa := NewAsciiArtBox()
	aa.SetStyle(AsciiArtStyle{
		Text:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Shadow: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	aa.SetText("9")
	buf := buffer.NewBuffer(10, 5)
	aa.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	aa.Paint(buf)
}

// ─── LogScale Tests ───

func TestLogScaleBasic(t *testing.T) {
	ls := NewLogScale()
	ls.SetValue(1000, 1, 1000000)
	if v := ls.Value(); v != 1000 {
		t.Errorf("Value = %d, want 1000", v)
	}
}

func TestLogScaleMinValue(t *testing.T) {
	ls := NewLogScale()
	ls.SetValue(1, 1, 1000000)
	if ls.fillPct != 0 {
		t.Errorf("fillPct at min = %d, want 0", ls.fillPct)
	}
}

func TestLogScaleMaxValue(t *testing.T) {
	ls := NewLogScale()
	ls.SetValue(1000000, 1, 1000000)
	if ls.fillPct != 100 {
		t.Errorf("fillPct at max = %d, want 100", ls.fillPct)
	}
}

func TestLogScaleClamp(t *testing.T) {
	ls := NewLogScale()
	ls.SetValue(-10, -5, -1)
	// Should clamp to minV=1, maxV=2
	if v := ls.Value(); v != 1 {
		t.Errorf("Value = %d, want 1 (clamped)", v)
	}
}

func TestLogScaleWidth(t *testing.T) {
	ls := NewLogScale()
	ls.SetWidth(40)
	if ls.width != 40 {
		t.Errorf("width = %d, want 40", ls.width)
	}
	ls.SetWidth(5)
	if ls.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", ls.width)
	}
}

func TestLogScalePaint(t *testing.T) {
	ls := NewLogScale()
	ls.SetValue(1000, 1, 1000000)
	ls.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	ls.Paint(buf)
	hasBar := false
	for i := 0; i < 20; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '█' || r == '░' {
			hasBar = true
			break
		}
	}
	if !hasBar {
		t.Error("Paint should show bar")
	}
}

func TestLogScaleChildren(t *testing.T) {
	ls := NewLogScale()
	if c := ls.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestLogScaleStyle(t *testing.T) {
	ls := NewLogScale()
	ls.SetStyle(LogScaleStyle{
		Fill:   buffer.Style{Fg: buffer.RGB(255, 0, 255)},
		Empty:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Marker: buffer.Style{Fg: buffer.RGB(255, 255, 0)},
	})
	ls.SetValue(500, 1, 10000)
	buf := buffer.NewBuffer(30, 2)
	ls.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	ls.Paint(buf)
}

// ─── LogViewer Tests ───

func TestLogViewerBasic(t *testing.T) {
	lv := NewLogViewer()
	lv.AddEntry(LVInfo, "Server started")
	if n := lv.Count(); n != 1 {
		t.Errorf("Count = %d, want 1", n)
	}
}

func TestLogViewerMultiple(t *testing.T) {
	lv := NewLogViewer()
	lv.AddEntry(LVInfo, "msg1")
	lv.AddEntry(LVWarn, "msg2")
	lv.AddEntry(LVError, "msg3")
	if n := lv.Count(); n != 3 {
		t.Errorf("Count = %d, want 3", n)
	}
}

func TestLogViewerOverflow(t *testing.T) {
	lv := NewLogViewer()
	for i := 0; i < logViewerMaxEntries+10; i++ {
		lv.AddEntry(LVDebug, "msg")
	}
	if n := lv.Count(); n != logViewerMaxEntries {
		t.Errorf("Count = %d, want %d (capped)", n, logViewerMaxEntries)
	}
}

func TestLogViewerClear(t *testing.T) {
	lv := NewLogViewer()
	lv.AddEntry(LVInfo, "msg")
	lv.Clear()
	if n := lv.Count(); n != 0 {
		t.Errorf("Count after Clear = %d, want 0", n)
	}
}

func TestLogViewerEmpty(t *testing.T) {
	lv := NewLogViewer()
	if n := lv.Count(); n != 0 {
		t.Errorf("Count = %d, want 0", n)
	}
}

func TestLogViewerSetMaxRows(t *testing.T) {
	lv := NewLogViewer()
	lv.SetMaxRows(5)
	if lv.maxRows != 5 {
		t.Errorf("maxRows = %d, want 5", lv.maxRows)
	}
	lv.SetMaxRows(0)
	if lv.maxRows != 1 {
		t.Errorf("maxRows = %d, want 1 (clamped)", lv.maxRows)
	}
}

func TestLogViewerPaint(t *testing.T) {
	lv := NewLogViewer()
	lv.AddEntry(LVInfo, "Server started")
	lv.AddEntry(LVWarn, "High memory")
	lv.AddEntry(LVError, "Crash!")
	lv.SetMaxRows(3)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	lv.Paint(buf)
	// First row should have ℹ icon (Info)
	if r := buf.GetCell(0, 0).Rune; r != 'ℹ' {
		t.Errorf("Row 0 icon = %q, want 'ℹ'", r)
	}
}

func TestLogViewerPaintEmpty(t *testing.T) {
	lv := NewLogViewer()
	lv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	lv.Paint(buf)
	// Should be empty
	if r := buf.GetCell(0, 0).Rune; r != 0 && r != ' ' {
		t.Errorf("Expected empty, got %q", r)
	}
}

func TestLogViewerChildren(t *testing.T) {
	lv := NewLogViewer()
	if c := lv.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestLogViewerStyle(t *testing.T) {
	lv := NewLogViewer()
	lv.SetStyle(LogViewerStyle{
		Debug: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Info:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Warn:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Error: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Text:  buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Time:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	lv.AddEntry(LVError, "test")
	buf := buffer.NewBuffer(50, 3)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	lv.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintAsciiArtBox(b *testing.B) {
	aa := NewAsciiArtBox()
	aa.SetText("2024")
	aa.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aa.Paint(buf)
	}
}

func BenchmarkPaintLogScale(b *testing.B) {
	ls := NewLogScale()
	ls.SetValue(1000, 1, 1000000)
	ls.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.Paint(buf)
	}
}

func BenchmarkPaintLogViewer(b *testing.B) {
	lv := NewLogViewer()
	for i := 0; i < 20; i++ {
		lv.AddEntry(LVInfo, "Sample log entry message")
	}
	lv.SetMaxRows(10)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lv.Paint(buf)
	}
}
