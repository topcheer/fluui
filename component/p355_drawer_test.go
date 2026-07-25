package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP355_Drawer_Create(t *testing.T) {
	d := NewDrawer(DrawerLeft, "Settings")
	if d.IsOpen() {
		t.Error("should start closed")
	}
	if d.Title() != "Settings" {
		t.Errorf("title = %q", d.Title())
	}
	if d.Side() != DrawerLeft {
		t.Error("should be left side")
	}
}

func TestP355_Drawer_OpenClose(t *testing.T) {
	d := NewDrawer(DrawerRight, "Menu")
	d.Open()
	if !d.IsOpen() {
		t.Error("should be open")
	}
	d.Close()
	if d.IsOpen() {
		t.Error("should be closed")
	}
}

func TestP355_Drawer_Toggle(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.Toggle()
	if !d.IsOpen() {
		t.Error("should be open after toggle")
	}
	d.Toggle()
	if d.IsOpen() {
		t.Error("should be closed after second toggle")
	}
}

func TestP355_Drawer_SetWidth(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.SetWidth(40)
	if d.Width() != 40 {
		t.Errorf("width = %d", d.Width())
	}
	d.SetWidth(1) // should clamp to 5
	if d.Width() != 5 {
		t.Errorf("width = %d, want 5 (clamped)", d.Width())
	}
}

func TestP355_Drawer_SetTitle(t *testing.T) {
	d := NewDrawer(DrawerLeft, "Old")
	d.SetTitle("New")
	if d.Title() != "New" {
		t.Errorf("title = %q", d.Title())
	}
}

func TestP355_Drawer_SetSide(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.SetSide(DrawerRight)
	if d.Side() != DrawerRight {
		t.Error("should be right")
	}
}

func TestP355_Drawer_Measure_Closed(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 0 || s.H != 0 {
		t.Errorf("closed measure = %dx%d, want 0x0", s.W, s.H)
	}
}

func TestP355_Drawer_Measure_Open(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.SetWidth(30)
	d.Open()
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 30 {
		t.Errorf("width = %d, want 30", s.W)
	}
	if s.H != 24 {
		t.Errorf("height = %d, want 24", s.H)
	}
}

func TestP355_Drawer_Paint_Left(t *testing.T) {
	d := NewDrawer(DrawerLeft, "Settings")
	d.Open()
	d.SetWidth(25)
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	d.Paint(buf)

	// Left edge should have border
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected border at top-left")
	}
}

func TestP355_Drawer_Paint_Right(t *testing.T) {
	d := NewDrawer(DrawerRight, "Menu")
	d.Open()
	d.SetWidth(25)
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	d.Paint(buf)

	// Right edge area should have content
	if buf.GetCell(55, 0).Rune == 0 {
		t.Error("expected border at right side")
	}
}

func TestP355_Drawer_Paint_Closed(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	d.Paint(buf) // closed → should render nothing
	// Buffer is pre-filled with zero cells; closed drawer leaves it untouched
	_ = buf
}

func TestP355_Drawer_Paint_LongTitle(t *testing.T) {
	d := NewDrawer(DrawerLeft, "This is a very long title that needs truncation")
	d.Open()
	d.SetWidth(15)
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	d.Paint(buf)
}

func TestP355_Drawer_Paint_NarrowWidth(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.Open()
	d.SetWidth(5) // minimum
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	d.Paint(buf)
}

func TestP355_Drawer_Paint_ZeroBounds(t *testing.T) {
	d := NewDrawer(DrawerLeft, "T")
	d.Open()
	d.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	d.Paint(buf) // should not panic
}

func BenchmarkDrawer_Paint(b *testing.B) {
	d := NewDrawer(DrawerLeft, "Settings Panel")
	d.Open()
	d.SetWidth(30)
	d.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Paint(buf)
	}
}
