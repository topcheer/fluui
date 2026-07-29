package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── P487: Zero-alloc Paint benchmarks for P485-P486 components ───

func BenchmarkPaintMarkdownInlineCode(b *testing.B) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Use `fmt.Println` and `os.Exit` to control flow.\n```go\nfunc main() {\n    fmt.Println(\"hello\")\n    os.Exit(0)\n}\n```\nThen run `go build` and `./myapp`.")
	mic.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mic.Paint(buf)
	}
}

func BenchmarkPaintStepProgress(b *testing.B) {
	sp := NewStepProgress()
	sp.AddStep("Account Setup")
	sp.AddStep("Profile Details")
	sp.AddStep("Payment Method")
	sp.AddStep("Confirmation")
	sp.AddStep("Welcome Aboard")
	sp.SetCurrentStep(2)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sp.Paint(buf)
	}
}
