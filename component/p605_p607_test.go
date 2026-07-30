package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CompassDial Tests ───

func TestCompassDialBasic(t *testing.T) {
	cd := NewCompassDial()
	cd.SetHeading(90)
	if h := cd.Heading(); h != 90 {
		t.Errorf("Heading = %d, want 90", h)
	}
}

func TestCompassDialZero(t *testing.T) {
	cd := NewCompassDial()
	cd.SetHeading(0)
	if h := cd.Heading(); h != 0 {
		t.Errorf("Heading = %d, want 0", h)
	}
}

func TestCompassDialFull(t *testing.T) {
	cd := NewCompassDial()
	cd.SetHeading(359)
	if h := cd.Heading(); h != 359 {
		t.Errorf("Heading = %d, want 359", h)
	}
}

func TestCompassDialNegative(t *testing.T) {
	cd := NewCompassDial()
	cd.SetHeading(-90)
	if h := cd.Heading(); h != 270 {
		t.Errorf("Heading = %d, want 270 (wrapped)", h)
	}
}

func TestCompassDialPaint(t *testing.T) {
	cd := NewCompassDial()
	cd.SetHeading(45) // NE = ↗
	cd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	cd.Paint(buf)
	// Should have N at start
	if r := buf.GetCell(0, 0).Rune; r != 'N' {
		t.Errorf("First rune = %q, want 'N'", r)
	}
}

func TestCompassDialChildren(t *testing.T) {
	cd := NewCompassDial()
	if c := cd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestCompassDialStyle(t *testing.T) {
	cd := NewCompassDial()
	cd.SetStyle(CompassDialStyle{
		Needle:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Cardinal: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Hub:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	cd.SetHeading(180)
	buf := buffer.NewBuffer(10, 1)
	cd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	cd.Paint(buf)
}

// ─── AIContextBadge Tests ───

func TestAIContextBadgeBasic(t *testing.T) {
	cb := NewAIContextBadge()
	cb.SetSource(ContextRAG, "docs.md")
	if st := cb.SourceType(); st != ContextRAG {
		t.Errorf("SourceType = %d, want ContextRAG(%d)", st, ContextRAG)
	}
}

func TestAIContextBadgeAllTypes(t *testing.T) {
	types := []ContextSourceType{ContextSystem, ContextRAG, ContextTool, ContextFineTune, ContextMemory}
	for _, st := range types {
		cb := NewAIContextBadge()
		cb.SetSource(st, "test")
		if cb.SourceType() != st {
			t.Errorf("SourceType = %d, want %d", cb.SourceType(), st)
		}
	}
}

func TestAIContextBadgeInvalid(t *testing.T) {
	cb := NewAIContextBadge()
	cb.SetSource(ContextSourceType(99), "test")
	if st := cb.SourceType(); st != ContextSystem {
		t.Errorf("SourceType = %d, want ContextSystem (clamped)", st)
	}
}

func TestAIContextBadgeNoName(t *testing.T) {
	cb := NewAIContextBadge()
	cb.SetSource(ContextTool, "")
	if cb.name != "" {
		t.Errorf("name = %q, want ''", cb.name)
	}
}

func TestAIContextBadgePaint(t *testing.T) {
	cb := NewAIContextBadge()
	cb.SetSource(ContextRAG, "file.txt")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	cb.Paint(buf)
	hasContent := false
	for i := 0; i < 20; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestAIContextBadgeChildren(t *testing.T) {
	cb := NewAIContextBadge()
	if c := cb.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIContextBadgeStyle(t *testing.T) {
	cb := NewAIContextBadge()
	cb.SetStyle(AIContextBadgeStyle{
		System:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		RAG:      buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Tool:     buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		FineTune: buffer.Style{Fg: buffer.RGB(255, 0, 255)},
		Memory:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Name:     buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Bracket:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	cb.SetSource(ContextMemory, "ctx")
	buf := buffer.NewBuffer(20, 1)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	cb.Paint(buf)
}

// ─── MiniGantt Tests ───

func TestMiniGanttBasic(t *testing.T) {
	mg := NewMiniGantt()
	mg.AddTask("A", 0, 20, buffer.RGB(59, 130, 246))
	if n := mg.TaskCount(); n != 1 {
		t.Errorf("TaskCount = %d, want 1", n)
	}
}

func TestMiniGanttOverflow(t *testing.T) {
	mg := NewMiniGantt()
	for i := 0; i < miniGanttMaxTasks+5; i++ {
		mg.AddTask("T", 0, 10, buffer.RGB(255, 0, 0))
	}
	if n := mg.TaskCount(); n != miniGanttMaxTasks {
		t.Errorf("TaskCount = %d, want %d (capped)", n, miniGanttMaxTasks)
	}
}

func TestMiniGanttClear(t *testing.T) {
	mg := NewMiniGantt()
	mg.AddTask("A", 0, 10, buffer.RGB(0, 0, 0))
	mg.Clear()
	if n := mg.TaskCount(); n != 0 {
		t.Errorf("TaskCount after Clear = %d, want 0", n)
	}
}

func TestMiniGanttRange(t *testing.T) {
	mg := NewMiniGantt()
	mg.SetRange(50, 50) // min==max
	if mg.rangeMax != 51 {
		t.Errorf("rangeMax = %d, want 51 (clamped)", mg.rangeMax)
	}
}

func TestMiniGanttPaint(t *testing.T) {
	mg := NewMiniGantt()
	mg.SetRange(0, 100)
	mg.AddTask("Design", 0, 20, buffer.RGB(59, 130, 246))
	mg.AddTask("Build", 15, 50, buffer.RGB(34, 197, 94))
	mg.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	mg.Paint(buf)
	// Axis row should have ─ characters
	if r := buf.GetCell(5, 0).Rune; r != '─' && r != '├' && r != '┤' && r != '0' {
		t.Errorf("Axis rune = %q, want '─' or markers", r)
	}
}

func TestMiniGanttChildren(t *testing.T) {
	mg := NewMiniGantt()
	if c := mg.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMiniGanttStyle(t *testing.T) {
	mg := NewMiniGantt()
	mg.SetStyle(MiniGanttStyle{
		Axis:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Empty: buffer.Style{Fg: buffer.RGB(30, 30, 30)},
	})
	mg.AddTask("X", 0, 50, buffer.RGB(255, 0, 0))
	buf := buffer.NewBuffer(30, 3)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	mg.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintCompassDial(b *testing.B) {
	cd := NewCompassDial()
	cd.SetHeading(135)
	cd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.Paint(buf)
	}
}

func BenchmarkPaintAIContextBadge(b *testing.B) {
	cb := NewAIContextBadge()
	cb.SetSource(ContextRAG, "docs.md")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}

func BenchmarkPaintMiniGantt(b *testing.B) {
	mg := NewMiniGantt()
	mg.SetRange(0, 100)
	mg.AddTask("A", 0, 20, buffer.RGB(59, 130, 246))
	mg.AddTask("B", 20, 30, buffer.RGB(34, 197, 94))
	mg.AddTask("C", 50, 40, buffer.RGB(245, 158, 11))
	mg.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mg.Paint(buf)
	}
}
