package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintPasswordStrength(b *testing.B) {
	ps := NewPasswordStrength()
	ps.SetPassword("MySuperSecureP@ssw0rd2024!")
	ps.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ps.Paint(buf)
	}
}

func BenchmarkPaintAITokenFlow(b *testing.B) {
	tf := NewAITokenFlow()
	tf.AddStage("Input", 500, buffer.RGB(96, 165, 250))
	tf.AddStage("Tokenize", 480, buffer.RGB(167, 139, 250))
	tf.AddStage("Embedding", 480, buffer.RGB(167, 139, 250))
	tf.AddStage("Attention", 450, buffer.RGB(234, 179, 8))
	tf.AddStage("FFN", 430, buffer.RGB(251, 146, 60))
	tf.AddStage("Output", 200, buffer.RGB(34, 197, 94))
	tf.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tf.Paint(buf)
	}
}
