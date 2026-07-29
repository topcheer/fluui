package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ErrorBoundary Tests ───

func TestErrorBoundaryBasic(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetError("Test error", "file.go:42")
	if !eb.HasError() {
		t.Error("Expected HasError to be true")
	}
}

func TestErrorBoundaryEmpty(t *testing.T) {
	eb := NewErrorBoundary()
	if eb.HasError() {
		t.Error("Expected HasError to be false initially")
	}
}

func TestErrorBoundaryClear(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetError("Error", "file.go:1")
	eb.Clear()
	if eb.HasError() {
		t.Error("Expected HasError to be false after Clear")
	}
}

func TestErrorBoundaryNoDetail(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetError("Just a message", "")
	if !eb.HasError() {
		t.Error("Expected HasError to be true")
	}
}

func TestErrorBoundaryWidth(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetWidth(60)
	if eb.width != 60 {
		t.Errorf("width = %d, want 60", eb.width)
	}
	eb.SetWidth(5)
	if eb.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", eb.width)
	}
}

func TestErrorBoundaryPaint(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetError("Render failed", "widget.go:42")
	eb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	eb.Paint(buf)
	// Should have border corner
	if r := buf.GetCell(0, 0).Rune; r != '╭' {
		t.Errorf("First rune = %q, want '╭'", r)
	}
}

func TestErrorBoundaryPaintEmpty(t *testing.T) {
	eb := NewErrorBoundary()
	eb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	eb.Paint(buf)
	// Empty - no content
	if r := buf.GetCell(0, 0).Rune; r != 0 && r != ' ' {
		t.Errorf("Expected empty buffer, got rune %q", r)
	}
}

func TestErrorBoundaryChildren(t *testing.T) {
	eb := NewErrorBoundary()
	if children := eb.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestErrorBoundaryStyle(t *testing.T) {
	eb := NewErrorBoundary()
	custom := ErrorBoundaryStyle{
		Icon:    buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Title:   buffer.Style{Fg: buffer.RGB(255, 100, 100)},
		Message: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Detail:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Border:  buffer.Style{Fg: buffer.RGB(100, 0, 0)},
	}
	eb.SetStyle(custom)
	eb.SetError("Test", "detail")
	buf := buffer.NewBuffer(40, 4)
	eb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	eb.Paint(buf)
}

// ─── DebugOverlay Tests ───

func TestDebugOverlayBasic(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(60, 142, 8192, 2500)
	if fps := d.FPS(); fps != 60 {
		t.Errorf("FPS = %d, want 60", fps)
	}
}

func TestDebugOverlayZero(t *testing.T) {
	d := NewDebugOverlay()
	if fps := d.FPS(); fps != 0 {
		t.Errorf("FPS = %d, want 0", fps)
	}
}

func TestDebugOverlayNegative(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(-10, -5, -100, -50)
	if fps := d.FPS(); fps != 0 {
		t.Errorf("FPS = %d, want 0 (clamped)", fps)
	}
}

func TestDebugOverlayLowFPS(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(15, 10, 100, 5000)
	// Should use bad style for low FPS
	if d.fpsStyle.Fg != d.style.Bad.Fg {
		t.Error("Expected Bad style for FPS < 30")
	}
}

func TestDebugOverlayMediumFPS(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(40, 10, 100, 5000)
	// Should use warn style for 30-49 FPS
	if d.fpsStyle.Fg != d.style.Warn.Fg {
		t.Error("Expected Warn style for FPS 30-49")
	}
}

func TestDebugOverlayHighFPS(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(60, 10, 100, 5000)
	// Should use good style for 50+ FPS
	if d.fpsStyle.Fg != d.style.Good.Fg {
		t.Error("Expected Good style for FPS >= 50")
	}
}

func TestDebugOverlayPaint(t *testing.T) {
	d := NewDebugOverlay()
	d.SetMetrics(60, 142, 8192, 2500)
	d.SetBounds(Rect{X: 0, Y: 0, W: 26, H: 4})
	buf := buffer.NewBuffer(26, 4)
	d.Paint(buf)
	// Should start with 'F' from "FPS"
	if r := buf.GetCell(0, 0).Rune; r != 'F' {
		t.Errorf("First rune = %q, want 'F'", r)
	}
}

func TestDebugOverlayChildren(t *testing.T) {
	d := NewDebugOverlay()
	if children := d.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestDebugOverlayStyle(t *testing.T) {
	d := NewDebugOverlay()
	custom := DebugOverlayStyle{
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Good:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Warn:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Bad:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
	}
	d.SetStyle(custom)
	d.SetMetrics(55, 10, 100, 500)
	buf := buffer.NewBuffer(26, 4)
	d.SetBounds(Rect{X: 0, Y: 0, W: 26, H: 4})
	d.Paint(buf)
}

// ─── ShortcutList Tests ───

func TestShortcutListBasic(t *testing.T) {
	h := NewShortcutList()
	h.AddBinding("Q", "Quit")
	if count := h.Count(); count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}
}

func TestShortcutListMultiple(t *testing.T) {
	h := NewShortcutList()
	h.AddBinding("Q", "Quit")
	h.AddBinding("Ctrl+S", "Save")
	h.AddBinding("Ctrl+C", "Copy")
	if count := h.Count(); count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestShortcutListClear(t *testing.T) {
	h := NewShortcutList()
	h.AddBinding("Q", "Quit")
	h.Clear()
	if count := h.Count(); count != 0 {
		t.Errorf("Count after Clear = %d, want 0", count)
	}
}

func TestShortcutListOverflow(t *testing.T) {
	h := NewShortcutList()
	for i := 0; i < helpMaxBindings+5; i++ {
		h.AddBinding("K", "Action")
	}
	if count := h.Count(); count != helpMaxBindings {
		t.Errorf("Count = %d, want %d (capped)", count, helpMaxBindings)
	}
}

func TestShortcutListEmpty(t *testing.T) {
	h := NewShortcutList()
	if count := h.Count(); count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}
}

func TestShortcutListWidth(t *testing.T) {
	h := NewShortcutList()
	h.SetWidth(50)
	if h.width != 50 {
		t.Errorf("width = %d, want 50", h.width)
	}
	h.SetWidth(5)
	if h.width != 15 {
		t.Errorf("width = %d, want 15 (clamped)", h.width)
	}
}

func TestShortcutListPaint(t *testing.T) {
	h := NewShortcutList()
	h.AddBinding("Q", "Quit application")
	h.AddBinding("S", "Save file")
	h.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 3})
	buf := buffer.NewBuffer(36, 3)
	h.Paint(buf)
	// Title row
	if r := buf.GetCell(0, 0).Rune; r != ' ' && r != 'S' {
		t.Errorf("First rune = %q, want ' ' or 'S'", r)
	}
	// Second row should have '[' for key bracket
	if r := buf.GetCell(0, 1).Rune; r != '[' {
		t.Errorf("Row 1 first rune = %q, want '['", r)
	}
}

func TestShortcutListChildren(t *testing.T) {
	h := NewShortcutList()
	if children := h.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestShortcutListStyle(t *testing.T) {
	h := NewShortcutList()
	custom := ShortcutListStyle{
		Key:     buffer.Style{Fg: buffer.RGB(100, 200, 255)},
		Sep:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Desc:    buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Title:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Bracket: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	h.SetStyle(custom)
	h.AddBinding("Ctrl+X", "Cut")
	buf := buffer.NewBuffer(36, 3)
	h.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 3})
	h.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintErrorBoundary(b *testing.B) {
	eb := NewErrorBoundary()
	eb.SetError("Failed to render component", "widget.go:42")
	eb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Paint(buf)
	}
}

func BenchmarkPaintDebugOverlay(b *testing.B) {
	d := NewDebugOverlay()
	d.SetMetrics(60, 142, 8192, 2500)
	d.SetBounds(Rect{X: 0, Y: 0, W: 26, H: 4})
	buf := buffer.NewBuffer(26, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Paint(buf)
	}
}

func BenchmarkPaintShortcutList(b *testing.B) {
	h := NewShortcutList()
	h.AddBinding("Q", "Quit application")
	h.AddBinding("Ctrl+S", "Save file")
	h.AddBinding("Ctrl+C", "Copy selection")
	h.AddBinding("Ctrl+V", "Paste")
	h.AddBinding("Ctrl+Z", "Undo")
	h.SetBounds(Rect{X: 0, Y: 0, W: 36, H: 6})
	buf := buffer.NewBuffer(36, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Paint(buf)
	}
}
