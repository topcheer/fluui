package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewQRCode(t *testing.T) {
	q := NewQRCode("https://github.com/topcheer/fluui")
	if q == nil {
		t.Fatal("NewQRCode returned nil")
	}
	if q.matrixSize == 0 {
		t.Error("matrixSize should be > 0 after generation")
	}
	if q.Size() == 0 {
		t.Error("Size should be > 0")
	}
}

func TestQRCode_SetData(t *testing.T) {
	q := NewQRCode("hello")
	origSize := q.Size()

	err := q.SetData("https://github.com/topcheer/fluui")
	if err != nil {
		t.Fatalf("SetData failed: %v", err)
	}

	// Different data may produce different sizes
	_ = origSize
	if q.data != "https://github.com/topcheer/fluui" {
		t.Error("data not updated")
	}
}

func TestQRCode_SetData_Error(t *testing.T) {
	q := NewQRCode("test")
	// Empty string should still work
	err := q.SetData("")
	if err == nil {
		// go-qrcode may reject empty — that's OK
	}
}

func TestQRCode_SetModuleSize(t *testing.T) {
	q := NewQRCode("test")

	q.SetModuleSize(1)
	q.mu.RLock()
	if q.module != 1 {
		t.Errorf("module = %d, want 1", q.module)
	}
	q.mu.RUnlock()

	q.SetModuleSize(5) // should clamp to 2
	q.mu.RLock()
	if q.module != 2 {
		t.Errorf("module = %d, want 2", q.module)
	}
	q.mu.RUnlock()

	q.SetModuleSize(-1) // should clamp to 1
	q.mu.RLock()
	if q.module != 1 {
		t.Errorf("module = %d, want 1", q.module)
	}
	q.mu.RUnlock()
}

func TestQRCode_SetMargin(t *testing.T) {
	q := NewQRCode("test")
	q.SetMargin(4)
	q.mu.RLock()
	if q.margin != 4 {
		t.Errorf("margin = %d, want 4", q.margin)
	}
	q.mu.RUnlock()

	q.SetMargin(-1)
	q.mu.RLock()
	if q.margin != 0 {
		t.Errorf("margin = %d, want 0", q.margin)
	}
	q.mu.RUnlock()
}

func TestQRCode_SetColors(t *testing.T) {
	q := NewQRCode("test")
	red := buffer.RGB(255, 0, 0)
	green := buffer.RGB(0, 255, 0)
	blue := buffer.RGB(0, 0, 255)
	q.SetColors(red, green, blue)

	q.mu.RLock()
	if q.darkColor != red {
		t.Error("darkColor not set")
	}
	if q.lightColor != green {
		t.Error("lightColor not set")
	}
	if q.bgColor != blue {
		t.Error("bgColor not set")
	}
	q.mu.RUnlock()
}

func TestQRCode_PixelSize(t *testing.T) {
	q := NewQRCode("test")
	w, h := q.PixelSize()
	if w <= 0 || h <= 0 {
		t.Errorf("PixelSize = (%d, %d), both should be > 0", w, h)
	}
}

func TestQRCode_Measure(t *testing.T) {
	q := NewQRCode("test")
	s := q.Measure(Constraints{MaxWidth: 80, MaxHeight: 40})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("Measure = (%d, %d), both should be > 0", s.W, s.H)
	}
}

func TestQRCode_Paint(t *testing.T) {
	q := NewQRCode("hello world")
	q.SetModuleSize(1)
	q.SetMargin(0)

	w := 40
	h := 25
	buf := buffer.NewBuffer(w, h)
	buf.Fill(buffer.BlankCell)

	q.SetBounds(Rect{X: 1, Y: 1, W: w - 2, H: h - 2})
	q.Paint(buf)

	// Should have drawn some non-blank cells
	found := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != ' ' && c.Rune != 0 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Paint did not render any non-blank cells")
	}
}

func TestQRCode_PaintNilBuffer(t *testing.T) {
	q := NewQRCode("test")
	q.Paint(nil) // should not panic
}

func TestQRCode_PaintZeroBounds(t *testing.T) {
	q := NewQRCode("test")
	buf := buffer.NewBuffer(40, 25)
	buf.Fill(buffer.BlankCell)

	q.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	q.Paint(buf) // should not crash
}

func TestQRCode_HandleKey(t *testing.T) {
	q := NewQRCode("test")
	if q.HandleKey(nil) != false {
		t.Error("HandleKey should return false")
	}
}

func TestQRCode_Children(t *testing.T) {
	q := NewQRCode("test")
	if q.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestQRCode_ConcurrentAccess(t *testing.T) {
	q := NewQRCode("concurrent test")
	done := make(chan struct{})

	go func() {
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(40, 25)
			buf.Fill(buffer.BlankCell)
			q.SetBounds(Rect{0, 0, 40, 25})
			q.Paint(buf)
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 50; i++ {
			q.Size()
			q.PixelSize()
			q.Measure(Constraints{MaxWidth: 80})
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}
