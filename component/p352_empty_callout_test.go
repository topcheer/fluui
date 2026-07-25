package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── EmptyState Tests ─────────────────────────────────────

func TestP352_EmptyState_Create(t *testing.T) {
	e := NewEmptyState("No messages", "Start a conversation to see messages here")
	if e.Title() != "No messages" {
		t.Errorf("title = %q", e.Title())
	}
}

func TestP352_EmptyState_SetIcon(t *testing.T) {
	e := NewEmptyState("Empty", "")
	e.SetIcon("\U0001f4ad") // 💭
}

func TestP352_EmptyState_SetTitle(t *testing.T) {
	e := NewEmptyState("Old", "")
	e.SetTitle("New Title")
	if e.Title() != "New Title" {
		t.Errorf("title = %q", e.Title())
	}
}

func TestP352_EmptyState_SetDescription(t *testing.T) {
	e := NewEmptyState("T", "")
	e.SetDescription("New desc")
}

func TestP352_EmptyState_SetHint(t *testing.T) {
	e := NewEmptyState("T", "")
	e.SetHint("Press / for commands")
}

func TestP352_EmptyState_Measure(t *testing.T) {
	e := NewEmptyState("T", "")
	s := e.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 || s.H != 5 {
		t.Errorf("defaults = %dx%d, want 40x5", s.W, s.H)
	}
	s = e.Measure(Constraints{MaxWidth: 60, MaxHeight: 10})
	if s.W != 60 || s.H != 10 {
		t.Errorf("explicit = %dx%d, want 60x10", s.W, s.H)
	}
}

func TestP352_EmptyState_Paint(t *testing.T) {
	e := NewEmptyState("No messages yet", "Start chatting!")
	e.SetHint("Press / for commands")
	e.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	e.Paint(buf)

	// Should render centered content
	filled := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			if buf.GetCell(x, y).Rune != 0 {
				filled++
			}
		}
	}
	if filled < 3 {
		t.Errorf("expected at least 3 filled cells, got %d", filled)
	}
}

func TestP352_EmptyState_Paint_ZeroBounds(t *testing.T) {
	e := NewEmptyState("T", "")
	e.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	e.Paint(buf) // should not panic
}

func TestP352_EmptyState_Paint_LongText(t *testing.T) {
	e := NewEmptyState(
		"This is a very long title that needs truncation",
		"And this is an equally long description that should also be truncated with ellipsis",
	)
	e.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	e.Paint(buf) // should truncate without panic
}

func TestP352_EmptyState_Paint_NoIcon(t *testing.T) {
	e := NewEmptyState("Just title", "")
	e.SetIcon("")
	e.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	e.Paint(buf)
}

// ─── Callout Tests ────────────────────────────────────────

func TestP352_Callout_Create(t *testing.T) {
	c := NewCallout(CalloutWarning, "This action cannot be undone")
	if c.Message() != "This action cannot be undone" {
		t.Errorf("msg = %q", c.Message())
	}
}

func TestP352_Callout_SetMessage(t *testing.T) {
	c := NewCallout(CalloutInfo, "old")
	c.SetMessage("new message")
	if c.Message() != "new message" {
		t.Errorf("msg = %q", c.Message())
	}
}

func TestP352_Callout_SetTitle(t *testing.T) {
	c := NewCallout(CalloutError, "error")
	c.SetTitle("Critical Error")
}

func TestP352_Callout_SetType(t *testing.T) {
	c := NewCallout(CalloutInfo, "msg")
	c.SetType(CalloutSuccess)
}

func TestP352_Callout_Measure(t *testing.T) {
	c := NewCallout(CalloutInfo, "message")
	s := c.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 60 {
		t.Errorf("default width = %d, want 60", s.W)
	}
	if s.H != 2 {
		t.Errorf("height with message = %d, want 2", s.H)
	}

	c2 := NewCallout(CalloutInfo, "")
	s2 := c2.Measure(Constraints{MaxWidth: 40, MaxHeight: 0})
	if s2.H != 1 {
		t.Errorf("height without message = %d, want 1", s2.H)
	}
}

func TestP352_Callout_Paint_Info(t *testing.T) {
	c := NewCallout(CalloutInfo, "This is an informational message")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	c.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP352_Callout_Paint_Warning(t *testing.T) {
	c := NewCallout(CalloutWarning, "Warning message")
	c.SetTitle("Caution")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	c.Paint(buf)
}

func TestP352_Callout_Paint_Error(t *testing.T) {
	c := NewCallout(CalloutError, "Error occurred")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	c.Paint(buf)
}

func TestP352_Callout_Paint_Success(t *testing.T) {
	c := NewCallout(CalloutSuccess, "Operation completed")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	c.Paint(buf)
}

func TestP352_Callout_Paint_ZeroBounds(t *testing.T) {
	c := NewCallout(CalloutInfo, "msg")
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(50, 2)
	c.Paint(buf) // should not panic
}

func TestP352_Callout_Paint_LongMessage(t *testing.T) {
	c := NewCallout(CalloutInfo, "This is a very long message that exceeds the available width and should be truncated")
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	c.Paint(buf) // should truncate without panic
}

// ─── TagInput Measure coverage fix ────────────────────────

func TestP352_TagInput_Measure_Default(t *testing.T) {
	ti := NewTagInput("")
	// maxW <= 0 branch
	s := ti.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 {
		t.Errorf("default width = %d, want 40", s.W)
	}
}

func BenchmarkEmptyState_Paint(b *testing.B) {
	e := NewEmptyState("No messages", "Start a conversation")
	e.SetHint("Press / for commands")
	e.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e.Paint(buf)
	}
}

func BenchmarkCallout_Paint(b *testing.B) {
	c := NewCallout(CalloutWarning, "This action cannot be undone")
	c.SetTitle("Warning")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}
