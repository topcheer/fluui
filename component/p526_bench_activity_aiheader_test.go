package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintActivityFeed(b *testing.B) {
	af := NewActivityFeed()
	af.AddEntry("alice_dev", "pushed 3 commits to main", time.Now().Add(-2*time.Minute))
	af.AddEntry("bob_smith", "opened pull request #42", time.Now().Add(-15*time.Minute))
	af.AddEntry("carol_ng", "approved pull request #42", time.Now().Add(-30*time.Minute))
	af.AddEntry("dave_wilson", "merged pull request #42", time.Now().Add(-45*time.Minute))
	af.AddEntry("eve_brown", "created branch feature/auth", time.Now().Add(-2*time.Hour))
	af.AddEntry("frank_garcia", "closed issue #17", time.Now().Add(-5*time.Hour))
	af.AddEntry("grace_lee", "commented on issue #23", time.Now().Add(-1*24*time.Hour))
	af.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		af.Paint(buf)
	}
}

func BenchmarkPaintAIPanelHeader(b *testing.B) {
	h := NewAIPanelHeader()
	h.SetModel("GPT-4o-2024-08-06")
	h.SetProvider("OpenAI")
	h.SetTokenUsage(45230, 128000)
	h.SetStatus(AIStatusStreaming)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Paint(buf)
	}
}
