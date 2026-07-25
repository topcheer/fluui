package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Banner tests
func TestP357_Banner_Create(t *testing.T) {
	b := NewBanner(BannerInfo, "AI model connected")
	if b.Message() != "AI model connected" {
		t.Errorf("msg = %q", b.Message())
	}
	if b.IsDismissed() {
		t.Error("should not be dismissed")
	}
}

func TestP357_Banner_DismissShow(t *testing.T) {
	b := NewBanner(BannerWarning, "Slow response")
	b.Dismiss()
	if !b.IsDismissed() {
		t.Error("should be dismissed")
	}
	b.Show()
	if b.IsDismissed() {
		t.Error("should not be dismissed after Show")
	}
}

func TestP357_Banner_SetMessage(t *testing.T) {
	b := NewBanner(BannerInfo, "old")
	b.SetMessage("new")
	if b.Message() != "new" {
		t.Errorf("msg = %q", b.Message())
	}
}

func TestP357_Banner_SetVariant(t *testing.T) {
	b := NewBanner(BannerInfo, "msg")
	b.SetVariant(BannerDanger)
}

func TestP357_Banner_SetAction(t *testing.T) {
	b := NewBanner(BannerInfo, "msg")
	b.SetAction("Press R to retry")
}

func TestP357_Banner_Measure(t *testing.T) {
	b := NewBanner(BannerInfo, "msg")
	s := b.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 60 || s.H != 1 {
		t.Errorf("defaults = %dx%d, want 60x1", s.W, s.H)
	}
}

func TestP357_Banner_Paint(t *testing.T) {
	b := NewBanner(BannerSuccess, "Operation completed successfully")
	b.SetAction("Dismiss: Esc")
	b.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.Paint(buf)
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP357_Banner_Paint_AllVariants(t *testing.T) {
	for _, v := range []BannerVariant{BannerNeutral, BannerInfo, BannerSuccess, BannerWarning, BannerDanger} {
		b := NewBanner(v, "test message")
		b.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
		buf := buffer.NewBuffer(40, 1)
		b.Paint(buf)
	}
}

func TestP357_Banner_Paint_Dismissed(t *testing.T) {
	b := NewBanner(BannerInfo, "msg")
	b.Dismiss()
	b.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)
}

func TestP357_Banner_Paint_LongMessage(t *testing.T) {
	b := NewBanner(BannerInfo, "This is a very long message that exceeds width and needs truncation")
	b.SetAction("Retry")
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
}

func TestP357_Banner_Paint_ZeroBounds(t *testing.T) {
	b := NewBanner(BannerInfo, "msg")
	b.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)
}

func BenchmarkBanner_Paint(b *testing.B) {
	bn := NewBanner(BannerWarning, "Rate limit approaching — 80% of quota used")
	bn.SetAction("Press D for details")
	bn.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bn.Paint(buf)
	}
}
