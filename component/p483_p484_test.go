package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownHorizontalRuleBasic(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("Before\n---\nAfter")
	if hr.LineCount() != 3 {
		t.Errorf("LineCount = %d, want 3", hr.LineCount())
	}
	if hr.RuleCount() != 1 {
		t.Errorf("RuleCount = %d, want 1", hr.RuleCount())
	}
}

func TestMarkdownHorizontalRuleMultipleRules(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("A\n---\nB\n***\nC\n___\nD")
	if hr.LineCount() != 7 {
		t.Errorf("LineCount = %d, want 7", hr.LineCount())
	}
	if hr.RuleCount() != 3 {
		t.Errorf("RuleCount = %d, want 3", hr.RuleCount())
	}
}

func TestMarkdownHorizontalRuleTitle(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("# Section Title\n---\nContent")
	hr.mu.Lock()
	if len(hr.cachedLines) != 3 {
		t.Fatalf("lines = %d, want 3", len(hr.cachedLines))
	}
	if hr.cachedLines[0].Type != hrTitle {
		t.Errorf("line 0 type = %d, want hrTitle", hr.cachedLines[0].Type)
	}
	if hr.cachedLines[0].Text != "Section Title" {
		t.Errorf("title = %q, want 'Section Title'", hr.cachedLines[0].Text)
	}
	hr.mu.Unlock()
}

func TestMarkdownHorizontalRuleEmpty(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("")
	if hr.LineCount() != 0 {
		t.Errorf("LineCount = %d, want 0", hr.LineCount())
	}
}

func TestIsHorizontalRule(t *testing.T) {
	if !isHorizontalRule("---") {
		t.Error("--- should be a horizontal rule")
	}
	if !isHorizontalRule("***") {
		t.Error("*** should be a horizontal rule")
	}
	if !isHorizontalRule("___") {
		t.Error("___ should be a horizontal rule")
	}
	if !isHorizontalRule("----------") {
		t.Error("---------- should be a horizontal rule")
	}
	if isHorizontalRule("--a") {
		t.Error("--a should NOT be a horizontal rule")
	}
	if isHorizontalRule("text") {
		t.Error("text should NOT be a horizontal rule")
	}
	if isHorizontalRule("--") {
		t.Error("-- (too short) should NOT be a horizontal rule")
	}
}

func TestMarkdownHorizontalRuleMeasure(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("Line 1\n---\nLine 2")
	s := hr.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestMarkdownHorizontalRulePaint(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("Section A\n---\nSection B")
	hr.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 6})

	buf := buffer.NewBuffer(50, 6)
	hr.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Find rule line (row 2 should have ─ chars)
	ruleCount := 0
	for x := 1; x < 49; x++ {
		if buf.GetCell(x, 2).Rune == '─' {
			ruleCount++
		}
	}
	if ruleCount == 0 {
		t.Error("horizontal rule line not found")
	}
}

func TestMarkdownHorizontalRulePaintEmpty(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	hr.Paint(buf) // should not panic
}

func TestMarkdownHorizontalRuleChildren(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	if hr.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownHorizontalRuleStyle(t *testing.T) {
	hr := NewMarkdownHorizontalRule()
	hr.SetStyle(HorizontalRuleStyle{
		Text:   buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Rule:   buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Title:  buffer.Style{Fg: buffer.RGB(255, 0, 255), Flags: buffer.Bold},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	hr.SetMarkdown("# Title\n---\nText")
	hr.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	hr.Paint(buf)
}

func BenchmarkPaintMarkdownTaskList(b *testing.B) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Task alpha\n- [x] Task beta\n- [ ] Task gamma\n- [x] Task delta\n- Plain item\n- [ ] Task epsilon")
	tl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tl.Paint(buf)
	}
}

func BenchmarkPaintMarkdownHorizontalRule(b *testing.B) {
	hr := NewMarkdownHorizontalRule()
	hr.SetMarkdown("# Section A\nText line 1\nText line 2\n---\n# Section B\nMore text\n***\nFinal section")
	hr.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hr.Paint(buf)
	}
}
