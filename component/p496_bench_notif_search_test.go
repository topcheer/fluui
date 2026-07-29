package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintNotificationStack(b *testing.B) {
	ns := NewNotificationStack()
	ns.AddNotification("Build Complete", "All 37 packages passed", NotifSuccess)
	ns.AddNotification("Deprecation Warning", "os.Exit is deprecated in main.go", NotifWarning)
	ns.AddNotification("Test Failure", "TestScanner failed in compat/xterm", NotifError)
	ns.AddNotification("Info", "New version v1.0.0 available", NotifInfo)
	ns.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ns.Paint(buf)
	}
}

func BenchmarkPaintImagePreview(b *testing.B) {
	ip := NewImagePreview()
	ip.SetFormat("PNG")
	ip.SetDimensions(1920, 1080)
	ip.SetLabel("screenshot.png")
	ip.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ip.Paint(buf)
	}
}
