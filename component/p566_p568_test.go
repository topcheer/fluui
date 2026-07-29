package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownCodeBlock Tests ───

func TestMarkdownCodeBlockBasic(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	cb.SetLanguage("go")
	cb.SetLines([]string{"func main() {}", "// test"})
	if n := cb.LineCount(); n != 2 {
		t.Errorf("LineCount = %d, want 2", n)
	}
}

func TestMarkdownCodeBlockEmpty(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	if n := cb.LineCount(); n != 0 {
		t.Errorf("LineCount = %d, want 0", n)
	}
}

func TestMarkdownCodeBlockNoLanguage(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	if cb.langLabel != "code" {
		t.Errorf("langLabel = %q, want 'code'", cb.langLabel)
	}
}

func TestMarkdownCodeBlockOverflow(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	lines := make([]string, codeBlockMaxLines+10)
	for i := range lines {
		lines[i] = "x"
	}
	cb.SetLines(lines)
	if n := cb.LineCount(); n != codeBlockMaxLines {
		t.Errorf("LineCount = %d, want %d (capped)", n, codeBlockMaxLines)
	}
}

func TestMarkdownCodeBlockLineNumbers(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	cb.SetShowLineNumbers(false)
	cb.SetShowLineNumbers(true)
	cb.SetLanguage("python")
	cb.SetLines([]string{"print('hi')"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
	// Header should show 'python'
	if r := buf.GetCell(0, 0).Rune; r != 'p' {
		t.Errorf("Header rune = %q, want 'p'", r)
	}
}

func TestMarkdownCodeBlockChildren(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	if c := cb.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMarkdownCodeBlockStyle(t *testing.T) {
	cb := NewMarkdownCodeBlock()
	cb.SetStyle(MarkdownCodeBlockStyle{
		Header:     buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Code:       buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		LineNumber: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		CopyIcon:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
	})
	cb.SetLanguage("rust")
	cb.SetLines([]string{"fn main() {}"})
	buf := buffer.NewBuffer(40, 3)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.Paint(buf)
}

// ─── AICitationList Tests ───

func TestAICitationListBasic(t *testing.T) {
	cl := NewAICitationList()
	cl.AddCitation("Paper A", "https://example.com/a")
	if n := cl.Count(); n != 1 {
		t.Errorf("Count = %d, want 1", n)
	}
}

func TestAICitationListMultiple(t *testing.T) {
	cl := NewAICitationList()
	cl.AddCitation("A", "")
	cl.AddCitation("B", "url1")
	cl.AddCitation("C", "url2")
	if n := cl.Count(); n != 3 {
		t.Errorf("Count = %d, want 3", n)
	}
}

func TestAICitationListOverflow(t *testing.T) {
	cl := NewAICitationList()
	for i := 0; i < citationMaxEntries+5; i++ {
		cl.AddCitation("x", "")
	}
	if n := cl.Count(); n != citationMaxEntries {
		t.Errorf("Count = %d, want %d (capped)", n, citationMaxEntries)
	}
}

func TestAICitationListClear(t *testing.T) {
	cl := NewAICitationList()
	cl.AddCitation("A", "")
	cl.Clear()
	if n := cl.Count(); n != 0 {
		t.Errorf("Count after Clear = %d, want 0", n)
	}
}

func TestAICitationListPaint(t *testing.T) {
	cl := NewAICitationList()
	cl.AddCitation("Important Paper", "https://arxiv.org/abs/1234")
	cl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	cl.Paint(buf)
	// Should start with '['
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
}

func TestAICitationListChildren(t *testing.T) {
	cl := NewAICitationList()
	if c := cl.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAICitationListStyle(t *testing.T) {
	cl := NewAICitationList()
	cl.SetStyle(AICitationStyle{
		Number:    buffer.Style{Fg: buffer.RGB(0, 255, 255)},
		Title:     buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		URL:       buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Separator: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	cl.AddCitation("Test", "url")
	buf := buffer.NewBuffer(50, 2)
	cl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	cl.Paint(buf)
}

// ─── HallucinationIndicator Tests ───

func TestHallucinationBasic(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(90, 85)
	if r := hi.Risk(); r != RiskLow {
		t.Errorf("Risk = %d, want RiskLow (%d)", r, RiskLow)
	}
}

func TestHallucinationHighRisk(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(20, 30)
	if r := hi.Risk(); r != RiskHigh {
		t.Errorf("Risk = %d, want RiskHigh (%d)", r, RiskHigh)
	}
}

func TestHallucinationMediumRisk(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(65, 60)
	if r := hi.Risk(); r != RiskMedium {
		t.Errorf("Risk = %d, want RiskMedium (%d)", r, RiskMedium)
	}
}

func TestHallucinationClamp(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(-10, 200)
	// Should clamp: confidence=0, factuality=100 -> riskScore=50 -> HIGH
	if r := hi.Risk(); r != RiskHigh {
		t.Errorf("Risk = %d, want RiskHigh for clamped scores", r)
	}
}

func TestHallucinationZero(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(0, 0)
	if r := hi.Risk(); r != RiskHigh {
		t.Errorf("Risk = %d, want RiskHigh for 0/0", r)
	}
}

func TestHallucinationPerfect(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(100, 100)
	if r := hi.Risk(); r != RiskLow {
		t.Errorf("Risk = %d, want RiskLow for 100/100", r)
	}
}

func TestHallucinationPaint(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetScores(85, 90)
	hi.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	buf := buffer.NewBuffer(22, 1)
	hi.Paint(buf)
	// Should have non-space content
	hasContent := false
	for i := 0; i < 22; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestHallucinationChildren(t *testing.T) {
	hi := NewHallucinationIndicator()
	if c := hi.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestHallucinationStyle(t *testing.T) {
	hi := NewHallucinationIndicator()
	hi.SetStyle(HallucinationStyle{
		Low:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Med:   buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		High:  buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	hi.SetScores(50, 50)
	buf := buffer.NewBuffer(22, 1)
	hi.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	hi.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintMarkdownCodeBlock(b *testing.B) {
	cb := NewMarkdownCodeBlock()
	cb.SetLanguage("go")
	cb.SetLines([]string{"package main", "", "func main() {", "    println(\"hello\")", "}"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 7})
	buf := buffer.NewBuffer(40, 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}

func BenchmarkPaintAICitationList(b *testing.B) {
	cl := NewAICitationList()
	cl.AddCitation("Paper A", "https://a.com")
	cl.AddCitation("Paper B", "https://b.com")
	cl.AddCitation("Paper C", "https://c.com")
	cl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.Paint(buf)
	}
}

func BenchmarkPaintHallucinationIndicator(b *testing.B) {
	hi := NewHallucinationIndicator()
	hi.SetScores(85, 90)
	hi.SetBounds(Rect{X: 0, Y: 0, W: 22, H: 1})
	buf := buffer.NewBuffer(22, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hi.Paint(buf)
	}
}
