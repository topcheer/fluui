package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelSwitcher Tests ───

func TestModelSwitcherBasic(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetModels("gpt-4", "claude-3", "gemini-pro")
	if idx := ms.ActiveIndex(); idx != 0 {
		t.Errorf("ActiveIndex = %d, want 0", idx)
	}
}

func TestModelSwitcherSetActive(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetModels("a", "b", "c")
	ms.SetActive(2)
	if idx := ms.ActiveIndex(); idx != 2 {
		t.Errorf("ActiveIndex = %d, want 2", idx)
	}
}

func TestModelSwitcherCycleNext(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetModels("a", "b", "c")
	ms.CycleNext()
	if idx := ms.ActiveIndex(); idx != 1 {
		t.Errorf("After CycleNext: ActiveIndex = %d, want 1", idx)
	}
	ms.CycleNext()
	ms.CycleNext() // wraps to 0
	if idx := ms.ActiveIndex(); idx != 0 {
		t.Errorf("After wrap: ActiveIndex = %d, want 0", idx)
	}
}

func TestModelSwitcherEmpty(t *testing.T) {
	ms := NewModelSwitcher()
	if idx := ms.ActiveIndex(); idx != 0 {
		t.Errorf("Empty ActiveIndex = %d, want 0", idx)
	}
}

func TestModelSwitcherInvalidActive(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetModels("a", "b")
	ms.SetActive(-1)
	if idx := ms.ActiveIndex(); idx != 0 {
		t.Errorf("ActiveIndex = %d, want 0 (clamped)", idx)
	}
	ms.SetActive(99)
	if idx := ms.ActiveIndex(); idx != 1 {
		t.Errorf("ActiveIndex = %d, want 1 (clamped)", idx)
	}
}

func TestModelSwitcherOverflowModels(t *testing.T) {
	ms := NewModelSwitcher()
	models := make([]string, modelSwitcherMax+5)
	for i := range models {
		models[i] = "model"
	}
	ms.SetModels(models...)
	if ms.count != modelSwitcherMax {
		t.Errorf("count = %d, want %d", ms.count, modelSwitcherMax)
	}
}

func TestModelSwitcherPaint(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetModels("gpt-4", "claude-3")
	ms.SetActive(1)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ms.Paint(buf)
	// Should start with ▶ arrow
	if r := buf.GetCell(0, 0).Rune; r != '▶' {
		t.Errorf("First rune = %q, want '▶'", r)
	}
}

func TestModelSwitcherPaintEmpty(t *testing.T) {
	ms := NewModelSwitcher()
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ms.Paint(buf)
	// Empty - should render nothing
	if r := buf.GetCell(0, 0).Rune; r != 0 && r != ' ' {
		t.Errorf("Expected empty, got rune %q", r)
	}
}

func TestModelSwitcherChildren(t *testing.T) {
	ms := NewModelSwitcher()
	if children := ms.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestModelSwitcherStyle(t *testing.T) {
	ms := NewModelSwitcher()
	custom := ModelSwitcherStyle{
		Active:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Inactive: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Arrow:    buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Bracket:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	ms.SetStyle(custom)
	ms.SetModels("model-a", "model-b")
	buf := buffer.NewBuffer(30, 1)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	ms.Paint(buf)
}

// ─── ConversationDepthBar Tests ───

func TestConversationDepthBasic(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetDepth(8, 20)
	if d := c.Depth(); d != 8 {
		t.Errorf("Depth = %d, want 8", d)
	}
}

func TestConversationDepthFull(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetDepth(20, 20)
	if d := c.Depth(); d != 20 {
		t.Errorf("Depth = %d, want 20", d)
	}
}

func TestConversationDepthZero(t *testing.T) {
	c := NewConversationDepthBar()
	if d := c.Depth(); d != 0 {
		t.Errorf("Depth = %d, want 0", d)
	}
}

func TestConversationDepthNegative(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetDepth(-5, -1)
	if d := c.Depth(); d != 0 {
		t.Errorf("Depth = %d, want 0 (clamped)", d)
	}
}

func TestConversationDepthOverflow(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetDepth(30, 20)
	if d := c.Depth(); d != 20 {
		t.Errorf("Depth = %d, want 20 (capped)", d)
	}
}

func TestConversationDepthWidth(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetWidth(50)
	if c.width != 50 {
		t.Errorf("width = %d, want 50", c.width)
	}
	c.SetWidth(5)
	if c.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", c.width)
	}
}

func TestConversationDepthPaint(t *testing.T) {
	c := NewConversationDepthBar()
	c.SetDepth(5, 10)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
	// Should start with 'T' from "Turns"
	if r := buf.GetCell(0, 0).Rune; r != 'T' {
		t.Errorf("First rune = %q, want 'T'", r)
	}
}

func TestConversationDepthChildren(t *testing.T) {
	c := NewConversationDepthBar()
	if children := c.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestConversationDepthStyle(t *testing.T) {
	c := NewConversationDepthBar()
	custom := ConversationDepthStyle{
		Filled:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Empty:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Counter: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	c.SetStyle(custom)
	c.SetDepth(3, 10)
	buf := buffer.NewBuffer(40, 1)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	c.Paint(buf)
}

// ─── TokenRing Tests ───

func TestTokenRingBasic(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(500, 1000)
	if pct := r.Percent(); pct != 50 {
		t.Errorf("Percent = %d, want 50", pct)
	}
}

func TestTokenRingFull(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(1000, 1000)
	if pct := r.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100", pct)
	}
}

func TestTokenRingZero(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(0, 1000)
	if pct := r.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestTokenRingOverflow(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(2000, 1000)
	if pct := r.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100 (capped)", pct)
	}
}

func TestTokenRingNegative(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(-10, -5)
	if pct := r.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestTokenRingQuadrants(t *testing.T) {
	r := NewTokenRing()
	// 0% → all ○
	r.SetUsage(0, 100)
	if r.ringChars[0] != '○' {
		t.Errorf("0%% ring[0] = %q, want '○'", r.ringChars[0])
	}
	// 25% → first quadrant filled
	r.SetUsage(25, 100)
	if r.ringChars[0] != '◐' {
		t.Errorf("25%% ring[0] = %q, want '◐'", r.ringChars[0])
	}
	// 75% → three quadrants
	r.SetUsage(75, 100)
	if r.ringChars[2] != '◓' {
		t.Errorf("75%% ring[2] = %q, want '◓'", r.ringChars[2])
	}
	// 100% → all filled
	r.SetUsage(100, 100)
	if r.ringChars[3] != '●' {
		t.Errorf("100%% ring[3] = %q, want '●'", r.ringChars[3])
	}
}

func TestTokenRingPaint(t *testing.T) {
	r := NewTokenRing()
	r.SetUsage(500, 1000)
	r.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	r.Paint(buf)
	// Should start with ring char
	c := buf.GetCell(0, 0)
	if c.Rune == 0 || c.Rune == ' ' {
		t.Error("Paint produced no ring character")
	}
}

func TestTokenRingChildren(t *testing.T) {
	r := NewTokenRing()
	if children := r.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestTokenRingStyle(t *testing.T) {
	r := NewTokenRing()
	custom := TokenRingStyle{
		Fill:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Empty:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Percent: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	r.SetStyle(custom)
	r.SetUsage(750, 1000)
	buf := buffer.NewBuffer(20, 1)
	r.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	r.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintModelSwitcher(b *testing.B) {
	ms := NewModelSwitcher()
	ms.SetModels("gpt-4", "claude-3", "gemini-pro", "llama-3")
	ms.SetActive(1)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.Paint(buf)
	}
}

func BenchmarkPaintConversationDepthBar(b *testing.B) {
	c := NewConversationDepthBar()
	c.SetDepth(8, 20)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintTokenRing(b *testing.B) {
	r := NewTokenRing()
	r.SetUsage(750, 1000)
	r.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Paint(buf)
	}
}
