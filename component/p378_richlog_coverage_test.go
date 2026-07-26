package component

import (
	"testing"
)

// P378: RichLog countVisibleLinesLocked coverage (78.6% → 95%+)

func TestP378_Measure_BeforeBounds(t *testing.T) {
	// Measure called before SetBounds → bounds.H <= 0 path
	rl := NewRichLog()
	rl.Info("test message")
	s := rl.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H < 1 {
		t.Errorf("H = %d, want >= 1", s.H)
	}
}

func TestP378_Measure_NarrowWidth(t *testing.T) {
	// Very narrow width → contentW < 1 path
	rl := NewRichLog()
	rl.Info("a long message that would wrap")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 20}) // width=1, hdrWidth=15 → contentW = 1-15 < 0 → clamped to 1
	s := rl.Measure(Constraints{MaxWidth: 1, MaxHeight: 20})
	if s.H < 1 {
		t.Errorf("H = %d, want >= 1", s.H)
	}
}

func TestP378_Measure_FilteredByLevel(t *testing.T) {
	// Entries below minLevel are skipped
	rl := NewRichLog()
	rl.SetMinLevel(LogWarn)
	rl.Write(LogDebug, "debug msg")  // should be hidden
	rl.Write(LogInfo, "info msg")    // should be hidden
	rl.Warn("warn msg")              // should be visible
	rl.Error("error msg")            // should be visible
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	s := rl.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	// Should count only 2 visible entries (warn + error)
	if s.H < 2 {
		t.Errorf("H = %d, want >= 2 (only warn+error visible)", s.H)
	}
}

func TestP378_Measure_EmptyEntries(t *testing.T) {
	rl := NewRichLog()
	s := rl.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H != 1 {
		t.Errorf("empty H = %d, want 1", s.H)
	}
}
