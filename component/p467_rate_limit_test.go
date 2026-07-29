package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestRateLimitIndicatorBasic(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(5000)
	rl.SetRemaining(3200)
	rl.SetResetTime(time.Now().Add(30 * time.Minute))

	if rl.Limit() != 5000 {
		t.Errorf("Limit = %d, want 5000", rl.Limit())
	}
	if rl.Remaining() != 3200 {
		t.Errorf("Remaining = %d, want 3200", rl.Remaining())
	}
	pct := rl.UsagePercent()
	// used = 1800, 1800/5000 = 36%
	if pct < 35 || pct > 37 {
		t.Errorf("UsagePercent = %f, want ~36", pct)
	}
	if rl.IsRateLimited() {
		t.Error("should not be rate limited")
	}
}

func TestRateLimitIndicatorRateLimited(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(1000)
	rl.SetRemaining(0)

	if !rl.IsRateLimited() {
		t.Error("should be rate limited when remaining=0")
	}
}

func TestRateLimitIndicatorRetryAfter(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(1000)
	rl.SetRemaining(500)
	rl.SetRetryAfter(30 * time.Second)

	if !rl.IsRateLimited() {
		t.Error("should be rate limited when retryAfter > 0")
	}
	if rl.RetryAfter() != 30*time.Second {
		t.Errorf("RetryAfter = %v, want 30s", rl.RetryAfter())
	}
}

func TestRateLimitIndicatorUsagePercent(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(100)
	rl.SetRemaining(100) // 0% used
	if rl.UsagePercent() != 0 {
		t.Errorf("UsagePercent = %f, want 0", rl.UsagePercent())
	}

	rl.SetRemaining(0) // 100% used
	if rl.UsagePercent() != 100 {
		t.Errorf("UsagePercent = %f, want 100", rl.UsagePercent())
	}
}

func TestRateLimitIndicatorZeroLimit(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(0)
	if rl.UsagePercent() != 0 {
		t.Errorf("UsagePercent with zero limit = %f, want 0", rl.UsagePercent())
	}
}

func TestRateLimitIndicatorMeasure(t *testing.T) {
	rl := NewRateLimitIndicator()
	s := rl.Measure(Constraints{})
	if s.W < 20 {
		t.Errorf("W = %d, want >= 20", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestRateLimitIndicatorPaint(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(1000)
	rl.SetRemaining(800)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	rl.Paint(buf)

	// Should have status icon at position 0
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("position 0 should have a cell")
	}
}

func TestRateLimitIndicatorPaintRetryAfter(t *testing.T) {
	rl := NewRateLimitIndicator()
	rl.SetLimit(1000)
	rl.SetRemaining(0)
	rl.SetRetryAfter(5 * time.Second)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	rl.Paint(buf)
	// Just verify it doesn't panic
}

func TestRateLimitIndicatorChildren(t *testing.T) {
	rl := NewRateLimitIndicator()
	if rl.Children() != nil {
		t.Error("Children should be nil")
	}
}
