package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AITokenCounter Tests ───

func TestAITokenCounterBasic(t *testing.T) {
	tc := NewAITokenCounter()
	tc.SetCounts(500, 350)
	if total := tc.TotalTokens(); total != 850 {
		t.Errorf("TotalTokens = %d, want 850", total)
	}
}

func TestAITokenCounterZero(t *testing.T) {
	tc := NewAITokenCounter()
	if total := tc.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0", total)
	}
}

func TestAITokenCounterNegative(t *testing.T) {
	tc := NewAITokenCounter()
	tc.SetCounts(-10, -20)
	if total := tc.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0 (clamped)", total)
	}
}

func TestAITokenCounterRate(t *testing.T) {
	tc := NewAITokenCounter()
	tc.SetRate(1000) // $10 per million
	tc.SetCounts(1000000, 0)
	if tc.costStr == "$0.00" {
		t.Error("Expected non-zero cost for 1M tokens at $10/M")
	}
}

func TestAITokenCounterPaint(t *testing.T) {
	tc := NewAITokenCounter()
	tc.SetCounts(500, 350)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 32, H: 1})
	buf := buffer.NewBuffer(32, 1)
	tc.Paint(buf)
	hasContent := false
	for i := 0; i < 32; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestAITokenCounterChildren(t *testing.T) {
	tc := NewAITokenCounter()
	if c := tc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAITokenCounterStyle(t *testing.T) {
	tc := NewAITokenCounter()
	tc.SetStyle(AITokenCounterStyle{
		Prompt:     buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Completion: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Total:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Cost:       buffer.Style{Fg: buffer.RGB(255, 215, 0)},
		Separator:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:      buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	tc.SetCounts(100, 200)
	buf := buffer.NewBuffer(32, 1)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 32, H: 1})
	tc.Paint(buf)
}

// ─── MarkdownAlert Tests ───

func TestMarkdownAlertBasic(t *testing.T) {
	a := NewMarkdownAlert()
	a.SetLevel(AlertWarning)
	a.SetText("Be careful!")
	if l := a.Level(); l != AlertWarning {
		t.Errorf("Level = %d, want AlertWarning(%d)", l, AlertWarning)
	}
}

func TestMarkdownAlertAllLevels(t *testing.T) {
	levels := []AlertLevel{AlertNote, AlertTip, AlertImportant, AlertWarning, AlertCaution}
	for _, l := range levels {
		a := NewMarkdownAlert()
		a.SetLevel(l)
		a.SetText("test")
		if a.Level() != l {
			t.Errorf("Level = %d, want %d", a.Level(), l)
		}
	}
}

func TestMarkdownAlertInvalidLevel(t *testing.T) {
	a := NewMarkdownAlert()
	a.SetLevel(AlertLevel(99))
	if l := a.Level(); l != AlertNote {
		t.Errorf("Level = %d, want AlertNote (clamped)", l)
	}
}

func TestMarkdownAlertEmpty(t *testing.T) {
	a := NewMarkdownAlert()
	a.SetText("")
	a.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	a.Paint(buf) // should not panic
}

func TestMarkdownAlertPaint(t *testing.T) {
	a := NewMarkdownAlert()
	a.SetLevel(AlertCaution)
	a.SetText("Danger zone!")
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	a.Paint(buf)
	// Should have left border
	if r := buf.GetCell(0, 0).Rune; r != '▌' {
		t.Errorf("First rune = %q, want '▌'", r)
	}
}

func TestMarkdownAlertChildren(t *testing.T) {
	a := NewMarkdownAlert()
	if c := a.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMarkdownAlertStyle(t *testing.T) {
	a := NewMarkdownAlert()
	a.SetStyle(MarkdownAlertStyle{
		Note:      buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Tip:       buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Important: buffer.Style{Fg: buffer.RGB(255, 0, 255)},
		Warning:   buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Caution:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Text:      buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Border:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	a.SetLevel(AlertTip)
	a.SetText("Pro tip!")
	buf := buffer.NewBuffer(30, 2)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	a.Paint(buf)
}

// ─── AIPolishIndicator Tests ───

func TestAIPolishBasic(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetScores(85, 90, 80)
	if s := p.Stars(); s != 4 {
		t.Errorf("Stars = %d, want 4", s)
	}
}

func TestAIPolishZero(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetScores(0, 0, 0)
	if s := p.Stars(); s != 0 {
		t.Errorf("Stars = %d, want 0", s)
	}
}

func TestAIPolishPerfect(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetScores(100, 100, 100)
	if s := p.Stars(); s != 5 {
		t.Errorf("Stars = %d, want 5", s)
	}
	if p.labelStr != "Excellent" {
		t.Errorf("labelStr = %q, want 'Excellent'", p.labelStr)
	}
}

func TestAIPolishClamp(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetScores(-10, 200, 50)
	if s := p.Stars(); s < 0 || s > 5 {
		t.Errorf("Stars = %d, out of range", s)
	}
}

func TestAIPolishPaint(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetScores(70, 80, 75)
	p.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	buf := buffer.NewBuffer(22, 1)
	p.Paint(buf)
	// Should have star characters
	hasStar := false
	for i := 0; i < 5; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '★' || r == '☆' {
			hasStar = true
			break
		}
	}
	if !hasStar {
		t.Error("Paint should show star characters")
	}
}

func TestAIPolishChildren(t *testing.T) {
	p := NewAIPolishIndicator()
	if c := p.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIPolishStyle(t *testing.T) {
	p := NewAIPolishIndicator()
	p.SetStyle(AIPolishStyle{
		Star:  buffer.Style{Fg: buffer.RGB(255, 215, 0)},
		Empty: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Score: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	p.SetScores(90, 85, 88)
	buf := buffer.NewBuffer(22, 1)
	p.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	p.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintAITokenCounter(b *testing.B) {
	tc := NewAITokenCounter()
	tc.SetCounts(500, 350)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 32, H: 1})
	buf := buffer.NewBuffer(32, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Paint(buf)
	}
}

func BenchmarkPaintMarkdownAlert(b *testing.B) {
	a := NewMarkdownAlert()
	a.SetLevel(AlertWarning)
	a.SetText("This is a warning message for the user.")
	a.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Paint(buf)
	}
}

func BenchmarkPaintAIPolishIndicator(b *testing.B) {
	p := NewAIPolishIndicator()
	p.SetScores(85, 90, 80)
	p.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	buf := buffer.NewBuffer(22, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Paint(buf)
	}
}
