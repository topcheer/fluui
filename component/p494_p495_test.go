package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNotificationStackBasic(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("Build", "Success", NotifSuccess)
	ns.AddNotification("Warning", "Deprecation", NotifWarning)
	if ns.Count() != 2 {
		t.Errorf("Count = %d, want 2", ns.Count())
	}
}

func TestNotificationStackDismiss(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("A", "msg", NotifInfo)
	ns.AddNotification("B", "msg", NotifError)
	ns.Dismiss(0)
	if ns.Count() != 1 {
		t.Errorf("Count after dismiss = %d, want 1", ns.Count())
	}
}

func TestNotificationStackDismissOutOfRange(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("A", "msg", NotifInfo)
	ns.Dismiss(5)
	ns.Dismiss(-1)
	if ns.Count() != 1 {
		t.Errorf("Count = %d, want 1 (no change)", ns.Count())
	}
}

func TestNotificationStackClear(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("A", "msg", NotifInfo)
	ns.AddNotification("B", "msg", NotifError)
	ns.Clear()
	if ns.Count() != 0 {
		t.Errorf("Count after clear = %d, want 0", ns.Count())
	}
}

func TestNotificationStackSeverities(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("i", "info", NotifInfo)
	ns.AddNotification("w", "warn", NotifWarning)
	ns.AddNotification("e", "error", NotifError)
	ns.AddNotification("s", "success", NotifSuccess)
	if ns.Count() != 4 {
		t.Errorf("Count = %d, want 4", ns.Count())
	}
}

func TestNotificationStackMeasure(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("A", "msg", NotifInfo)
	s := ns.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 3 {
		t.Errorf("H = %d, want >= 3", s.H)
	}
}

func TestNotificationStackPaint(t *testing.T) {
	ns := NewNotificationStack()
	ns.AddNotification("Build", "Compiled OK", NotifSuccess)
	ns.AddNotification("Error", "Test failed", NotifError)
	ns.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	ns.Paint(buf)

	// Check success icon
	foundCheck := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '✓' {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Error("success icon not found")
	}
}

func TestNotificationStackPaintEmpty(t *testing.T) {
	ns := NewNotificationStack()
	ns.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	ns.Paint(buf)
}

func TestNotificationStackChildren(t *testing.T) {
	ns := NewNotificationStack()
	if ns.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestNotificationStackStyle(t *testing.T) {
	ns := NewNotificationStack()
	ns.SetStyle(NotificationStackStyle{
		Title:   [4]buffer.Style{{Fg: buffer.RGB(0, 0, 255), Flags: buffer.Bold}, {}, {}, {}},
		Message: [4]buffer.Style{{Fg: buffer.RGB(200, 200, 200)}, {}, {}, {}},
		Icon:    [4]buffer.Style{{Fg: buffer.RGB(0, 255, 0)}, {}, {}, {}},
		Border:  [4]buffer.Style{{Fg: buffer.RGB(64, 64, 64)}, {}, {}, {}},
	})
	ns.AddNotification("T", "M", NotifInfo)
	ns.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ns.Paint(buf)
}

func TestNotifIcon(t *testing.T) {
	if notifIcon(NotifInfo) != 'ℹ' {
		t.Error("info icon should be ℹ")
	}
	if notifIcon(NotifWarning) != '⚠' {
		t.Error("warning icon should be ⚠")
	}
	if notifIcon(NotifError) != '✗' {
		t.Error("error icon should be ✗")
	}
	if notifIcon(NotifSuccess) != '✓' {
		t.Error("success icon should be ✓")
	}
}

// ─── ImagePreview tests ───

func TestImagePreviewBasic(t *testing.T) {
	ip := NewImagePreview()
	ip.SetFormat("JPEG")
	ip.SetDimensions(1920, 1080)
	ip.SetLabel("photo.jpg")
	if ip.Format() != "JPEG" {
		t.Errorf("Format = %q, want JPEG", ip.Format())
	}
	w, h := ip.Dimensions()
	if w != 1920 || h != 1080 {
		t.Errorf("Dimensions = (%d,%d), want (1920,1080)", w, h)
	}
	if ip.Label() != "photo.jpg" {
		t.Errorf("Label = %q, want photo.jpg", ip.Label())
	}
}

func TestImagePreviewMeasure(t *testing.T) {
	ip := NewImagePreview()
	s := ip.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestImagePreviewPaint(t *testing.T) {
	ip := NewImagePreview()
	ip.SetFormat("PNG")
	ip.SetDimensions(800, 600)
	ip.SetLabel("image.png")
	ip.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	ip.Paint(buf)

	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	// Check checkerboard pattern exists
	foundChecker := false
	for y := 2; y < 8; y++ {
		for x := 1; x < 29; x++ {
			r := buf.GetCell(x, y).Rune
			if r == '▓' || r == '░' {
				foundChecker = true
				break
			}
		}
	}
	if !foundChecker {
		t.Error("checkerboard pattern not found")
	}
}

func TestImagePreviewPaintEmpty(t *testing.T) {
	ip := NewImagePreview()
	ip.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	ip.Paint(buf)
}

func TestImagePreviewChildren(t *testing.T) {
	ip := NewImagePreview()
	if ip.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestImagePreviewStyle(t *testing.T) {
	ip := NewImagePreview()
	ip.SetStyle(ImagePreviewStyle{
		Label:     buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
		Dimension: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Format:    buffer.Style{Fg: buffer.RGB(0, 255, 0), Flags: buffer.Bold},
		CheckerA:  buffer.Style{Fg: buffer.RGB(50, 50, 50)},
		CheckerB:  buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Border:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ip.SetDimensions(100, 100)
	ip.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	ip.Paint(buf)
}
