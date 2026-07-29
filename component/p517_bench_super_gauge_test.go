package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownSuperscript(b *testing.B) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("Euler proved that e^(i*pi)^ + 1 = 0. Also x^(n+2)^ = y^(2k)^ and a^(2)^ + b^(2)^ = c^(2)^ in Euclidean space.")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ms.Paint(buf)
	}
}

func BenchmarkPaintGaugeCluster(b *testing.B) {
	gc := NewGaugeCluster()
	gc.SetColumns(2)
	gc.SetBarWidth(12)
	gc.AddGauge("CPU Core 0", 85, 100)
	gc.AddGauge("CPU Core 1", 45, 100)
	gc.AddGauge("Memory", 72, 100)
	gc.AddGauge("Disk I/O", 30, 100)
	gc.AddGauge("Network", 95, 100)
	gc.AddGauge("GPU", 60, 100)
	gc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gc.Paint(buf)
	}
}
