package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ColorWheel Tests ───

func TestColorWheelBasic(t *testing.T) {
	cw := NewColorWheel()
	cw.SetHue(180)
	if h := cw.Hue(); h != 180 {
		t.Errorf("Hue = %d, want 180", h)
	}
}

func TestColorWheelWrap(t *testing.T) {
	cw := NewColorWheel()
	cw.SetHue(-30)
	if h := cw.Hue(); h < 0 || h >= 360 {
		t.Errorf("Hue = %d, should wrap to [0,360)", h)
	}
}

func TestColorWheelZero(t *testing.T) {
	cw := NewColorWheel()
	cw.SetHue(0)
	if h := cw.Hue(); h != 0 {
		t.Errorf("Hue = %d, want 0", h)
	}
}

func TestColorWheelFull(t *testing.T) {
	cw := NewColorWheel()
	cw.SetHue(359)
	if h := cw.Hue(); h != 359 {
		t.Errorf("Hue = %d, want 359", h)
	}
}

func TestColorWheelPaint(t *testing.T) {
	cw := NewColorWheel()
	cw.SetHue(120) // green
	cw.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 1})
	buf := buffer.NewBuffer(16, 1)
	cw.Paint(buf)
	if r := buf.GetCell(4, 0).Rune; r == 0 || r == ' ' {
		t.Error("Paint should show color dots")
	}
}

func TestColorWheelChildren(t *testing.T) {
	cw := NewColorWheel()
	if c := cw.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestColorWheelStyle(t *testing.T) {
	cw := NewColorWheel()
	cw.SetStyle(ColorWheelStyle{
		Selected: buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Marker:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	cw.SetHue(60)
	buf := buffer.NewBuffer(16, 1)
	cw.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 1})
	cw.Paint(buf)
}

// ─── FileSizeBar Tests ───

func TestFileSizeBarBasic(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetSize(8589934592, 107374182400) // 8GB of 100GB
	if u := fb.Used(); u != 8589934592 {
		t.Errorf("Used = %d, want 8589934592", u)
	}
}

func TestFileSizeBarZero(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetSize(0, 100)
	if u := fb.Used(); u != 0 {
		t.Errorf("Used = %d, want 0", u)
	}
}

func TestFileSizeBarClamp(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetSize(200, 100)
	if u := fb.Used(); u != 100 {
		t.Errorf("Used = %d, want 100 (clamped)", u)
	}
}

func TestFileSizeBarHumanize(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetSize(1073741824, 10737418240) // 1GB of 10GB
	// usedStr should contain "1GB" or "1024MB"
	if fb.usedStr == "" {
		t.Error("usedStr should not be empty")
	}
}

func TestFileSizeBarPaint(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetSize(50, 100)
	fb.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 1})
	buf := buffer.NewBuffer(36, 1)
	fb.Paint(buf)
	hasBar := false
	for i := 0; i < 36; i++ {
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

func TestFileSizeBarChildren(t *testing.T) {
	fb := NewFileSizeBar()
	if c := fb.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestFileSizeBarStyle(t *testing.T) {
	fb := NewFileSizeBar()
	fb.SetStyle(FileSizeBarStyle{
		Fill:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Empty:  buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Suffix: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
	})
	fb.SetSize(1024, 4096)
	buf := buffer.NewBuffer(36, 1)
	fb.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 1})
	fb.Paint(buf)
}

// ─── EdgeLabel Tests ───

func TestEdgeLabelBasic(t *testing.T) {
	el := NewEdgeLabel()
	el.SetLabel("data")
	if l := el.Label(); l != "data" {
		t.Errorf("Label = %q, want 'data'", l)
	}
}

func TestEdgeLabelEmpty(t *testing.T) {
	el := NewEdgeLabel()
	if l := el.Label(); l != "" {
		t.Errorf("Label = %q, want ''", l)
	}
}

func TestEdgeLabelLength(t *testing.T) {
	el := NewEdgeLabel()
	el.SetLength(5)
	if el.length != 5 {
		t.Errorf("length = %d, want 5", el.length)
	}
	el.SetLength(1)
	if el.length != 3 {
		t.Errorf("length = %d, want 3 (clamped)", el.length)
	}
}

func TestEdgeLabelVertical(t *testing.T) {
	el := NewEdgeLabel()
	el.SetVertical(true)
	if !el.vertical {
		t.Error("Expected vertical=true")
	}
}

func TestEdgeLabelPaintHorizontal(t *testing.T) {
	el := NewEdgeLabel()
	el.SetLabel("sync")
	el.SetLength(16)
	el.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 1})
	buf := buffer.NewBuffer(16, 1)
	el.Paint(buf)
	// Should have line chars
	hasLine := false
	for i := 0; i < 16; i++ {
		if buf.GetCell(i, 0).Rune == '─' {
			hasLine = true
			break
		}
	}
	if !hasLine {
		t.Error("Paint should show horizontal line")
	}
}

func TestEdgeLabelPaintVertical(t *testing.T) {
	el := NewEdgeLabel()
	el.SetLabel("x")
	el.SetVertical(true)
	el.SetLength(8)
	el.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 8})
	buf := buffer.NewBuffer(3, 8)
	el.Paint(buf)
	// Should have vertical line chars
	hasLine := false
	for i := 0; i < 8; i++ {
		if buf.GetCell(0, i).Rune == '│' {
			hasLine = true
			break
		}
	}
	if !hasLine {
		t.Error("Paint should show vertical line")
	}
}

func TestEdgeLabelChildren(t *testing.T) {
	el := NewEdgeLabel()
	if c := el.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestEdgeLabelStyle(t *testing.T) {
	el := NewEdgeLabel()
	el.SetStyle(EdgeLabelStyle{
		Line:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	el.SetLabel("test")
	el.SetLength(20)
	buf := buffer.NewBuffer(20, 1)
	el.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	el.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintColorWheel(b *testing.B) {
	cw := NewColorWheel()
	cw.SetHue(180)
	cw.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 1})
	buf := buffer.NewBuffer(16, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cw.Paint(buf)
	}
}

func BenchmarkPaintFileSizeBar(b *testing.B) {
	fb := NewFileSizeBar()
	fb.SetSize(8589934592, 107374182400)
	fb.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 1})
	buf := buffer.NewBuffer(36, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fb.Paint(buf)
	}
}

func BenchmarkPaintEdgeLabel(b *testing.B) {
	el := NewEdgeLabel()
	el.SetLabel("data-flow")
	el.SetLength(20)
	el.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		el.Paint(buf)
	}
}
