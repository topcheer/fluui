package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStreamProgressBasic(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetExpected(500)
	if sp.Expected() != 500 {
		t.Errorf("Expected = %d, want 500", sp.Expected())
	}
	if sp.State() != StreamIdle {
		t.Errorf("State = %d, want StreamIdle", sp.State())
	}
}

func TestStreamProgressStartAddComplete(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetExpected(500)
	sp.Start()
	if sp.State() != StreamActive {
		t.Errorf("State after Start = %d, want StreamActive", sp.State())
	}
	sp.AddTokens(250)
	if sp.TokensReceived() != 250 {
		t.Errorf("TokensReceived = %d, want 250", sp.TokensReceived())
	}
	sp.Complete()
	if sp.State() != StreamDone {
		t.Errorf("State after Complete = %d, want StreamDone", sp.State())
	}
}

func TestStreamProgressFail(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.Start()
	sp.Fail()
	if sp.State() != StreamError {
		t.Errorf("State = %d, want StreamError", sp.State())
	}
}

func TestStreamProgressPercent(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetExpected(100)
	sp.AddTokens(50)
	if sp.Percent() != 50 {
		t.Errorf("Percent = %f, want 50", sp.Percent())
	}
	sp.AddTokens(100) // over
	if sp.Percent() != 100 {
		t.Errorf("Percent over = %f, want 100", sp.Percent())
	}
}

func TestStreamProgressPercentNoExpected(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.AddTokens(100)
	if sp.Percent() != 0 {
		t.Errorf("Percent without expected = %f, want 0", sp.Percent())
	}
}

func TestStreamProgressElapsed(t *testing.T) {
	sp := NewStreamProgressIndicator()
	// Before start
	if sp.Elapsed() != 0 {
		t.Errorf("Elapsed before start = %v, want 0", sp.Elapsed())
	}
	sp.Start()
	time.Sleep(10 * time.Millisecond)
	e := sp.Elapsed()
	if e < 5*time.Millisecond {
		t.Errorf("Elapsed = %v, want >= 5ms", e)
	}
	sp.Complete()
	// After complete, elapsed should be frozen
	e1 := sp.Elapsed()
	time.Sleep(10 * time.Millisecond)
	e2 := sp.Elapsed()
	if e2 != e1 {
		t.Errorf("Elapsed after complete changed: %v -> %v", e1, e2)
	}
}

func TestStreamProgressTokensPerSecond(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.Start()
	sp.AddTokens(100)
	time.Sleep(20 * time.Millisecond)
	tps := sp.TokensPerSecond()
	if tps <= 0 {
		t.Errorf("TokensPerSecond = %f, want > 0", tps)
	}
}

func TestStreamProgressMeasure(t *testing.T) {
	sp := NewStreamProgressIndicator()
	s := sp.Measure(Constraints{})
	if s.W < 30 {
		t.Errorf("W = %d, want >= 30", s.W)
	}
	if s.H < 4 {
		t.Errorf("H = %d, want >= 4", s.H)
	}
}

func TestStreamProgressPaint(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetExpected(500)
	sp.Start()
	sp.AddTokens(250)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})

	buf := buffer.NewBuffer(50, 4)
	sp.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Check filled bar chars
	filledCount := 0
	for x := 1; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '█' {
			filledCount++
		}
	}
	if filledCount == 0 {
		t.Error("no filled bar segments")
	}
}

func TestStreamProgressPaintDone(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetExpected(100)
	sp.Start()
	sp.AddTokens(100)
	sp.Complete()
	sp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})

	buf := buffer.NewBuffer(50, 4)
	sp.Paint(buf)

	// 100% → full bar
	filledCount := 0
	for x := 1; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '█' {
			filledCount++
		}
	}
	if filledCount < 20 {
		t.Errorf("filled = %d, want >= 20 for done state", filledCount)
	}
}

func TestStreamProgressPaintIdle(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	sp.Paint(buf) // should not panic
}

func TestStreamProgressChildren(t *testing.T) {
	sp := NewStreamProgressIndicator()
	if sp.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestStreamProgressBarWidth(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetBarWidth(10)
	sp.SetExpected(100)
	sp.Start()
	sp.AddTokens(50)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})

	buf := buffer.NewBuffer(50, 4)
	sp.Paint(buf)

	// 50% of 10 = 5 filled cells
	filledCount := 0
	for x := 1; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '█' {
			filledCount++
		}
	}
	if filledCount != 5 {
		t.Errorf("filled = %d, want 5", filledCount)
	}
}

func TestStreamProgressStyle(t *testing.T) {
	sp := NewStreamProgressIndicator()
	sp.SetStyle(StreamProgressStyle{
		Label:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Value:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Bar:    [3]buffer.Style{{}, {}, {}},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	sp.SetExpected(100)
	sp.Start()
	sp.AddTokens(50)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	sp.Paint(buf)
}
