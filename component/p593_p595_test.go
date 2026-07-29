package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownMath Tests ───

func TestMarkdownMathBasic(t *testing.T) {
	mm := NewMarkdownMath()
	mm.SetExpression("E = mc^2")
	if e := mm.Expression(); e != "E = mc^2" {
		t.Errorf("Expression = %q, want 'E = mc^2'", e)
	}
}

func TestMarkdownMathEmpty(t *testing.T) {
	mm := NewMarkdownMath()
	mm.SetExpression("")
	mm.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	mm.Paint(buf)
}

func TestMarkdownMathSymbolTransform(t *testing.T) {
	mm := NewMarkdownMath()
	mm.SetExpression("3 * 4 - 2")
	mm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	mm.Paint(buf)
	// '*' should be transformed to '×'
	hasMul := false
	for i := 0; i < 20; i++ {
		if buf.GetCell(i, 0).Rune == '×' {
			hasMul = true
			break
		}
	}
	if !hasMul {
		t.Error("Expected × symbol for *")
	}
}

func TestMarkdownMathBlock(t *testing.T) {
	mm := NewMarkdownMath()
	mm.SetExpression("x^2 + y^2")
	mm.SetBlock(true)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	mm.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != '⟦' {
		t.Errorf("First rune = %q, want '⟦' for block mode", r)
	}
}

func TestMarkdownMathChildren(t *testing.T) {
	mm := NewMarkdownMath()
	if c := mm.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMarkdownMathStyle(t *testing.T) {
	mm := NewMarkdownMath()
	mm.SetStyle(MarkdownMathStyle{
		Text:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Variable:  buffer.Style{Fg: buffer.RGB(147, 197, 253)},
		Operator:  buffer.Style{Fg: buffer.RGB(252, 165, 165)},
		Delimiter: buffer.Style{Fg: buffer.RGB(168, 85, 247)},
	})
	mm.SetExpression("a + b")
	buf := buffer.NewBuffer(10, 1)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	mm.Paint(buf)
}

// ─── AIResponseScore Tests ───

func TestAIResponseScoreBasic(t *testing.T) {
	rs := NewAIResponseScore()
	rs.SetScores(90, 85, 95, 80, 88)
	if avg := rs.Average(); avg != 87 {
		t.Errorf("Average = %d, want 87", avg)
	}
}

func TestAIResponseScoreZero(t *testing.T) {
	rs := NewAIResponseScore()
	if avg := rs.Average(); avg != 0 {
		t.Errorf("Average = %d, want 0", avg)
	}
}

func TestAIResponseScoreClamp(t *testing.T) {
	rs := NewAIResponseScore()
	rs.SetScores(-10, 200, 50, 50, 50)
	if avg := rs.Average(); avg != 50 {
		t.Errorf("Average = %d, want 50 (clamped)", avg)
	}
}

func TestAIResponseScorePerfect(t *testing.T) {
	rs := NewAIResponseScore()
	rs.SetScores(100, 100, 100, 100, 100)
	if avg := rs.Average(); avg != 100 {
		t.Errorf("Average = %d, want 100", avg)
	}
}

func TestAIResponseScorePaint(t *testing.T) {
	rs := NewAIResponseScore()
	rs.SetScores(80, 70, 90, 60, 85)
	rs.SetBounds(Rect{X: 0, Y: 0, W: 24, H: 5})
	buf := buffer.NewBuffer(24, 5)
	rs.Paint(buf)
	// Should have label 'A' in first row
	if r := buf.GetCell(0, 0).Rune; r != 'A' {
		t.Errorf("First rune = %q, want 'A'", r)
	}
}

func TestAIResponseScoreChildren(t *testing.T) {
	rs := NewAIResponseScore()
	if c := rs.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIResponseScoreStyle(t *testing.T) {
	rs := NewAIResponseScore()
	rs.SetStyle(AIResponseScoreStyle{
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Bar:     buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Empty:   buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Score:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Average: buffer.Style{Fg: buffer.RGB(255, 215, 0)},
	})
	rs.SetScores(85, 90, 75, 80, 88)
	buf := buffer.NewBuffer(24, 5)
	rs.SetBounds(Rect{X: 0, Y: 0, W: 24, H: 5})
	rs.Paint(buf)
}

// ─── StreamingDiff Tests ───

func TestStreamingDiffBasic(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetOldText("hello\nworld")
	sd.SetNewText("hello\nuniverse")
	if a := sd.AddCount(); a != 1 {
		t.Errorf("AddCount = %d, want 1", a)
	}
}

func TestStreamingDiffIdentical(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetOldText("same text")
	sd.SetNewText("same text")
	if a := sd.AddCount(); a != 0 {
		t.Errorf("AddCount = %d, want 0", a)
	}
	if d := sd.DelCount(); d != 0 {
		t.Errorf("DelCount = %d, want 0", d)
	}
}

func TestStreamingDiffEmpty(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetOldText("")
	sd.SetNewText("")
	if a := sd.AddCount(); a != 0 {
		t.Errorf("AddCount = %d, want 0", a)
	}
}

func TestStreamingDiffMultiLine(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetOldText("line1\nline2\nline3")
	sd.SetNewText("line1\nchanged\nline3")
	if a := sd.AddCount(); a != 1 {
		t.Errorf("AddCount = %d, want 1", a)
	}
	if d := sd.DelCount(); d != 1 {
		t.Errorf("DelCount = %d, want 1", d)
	}
}

func TestStreamingDiffPaint(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetOldText("old line")
	sd.SetNewText("new line")
	sd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	sd.Paint(buf)
	// Should have prefix characters
	hasPrefix := false
	for i := 0; i < 5; i++ {
		r := buf.GetCell(0, i).Rune
		if r == '+' || r == '-' || r == ' ' {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		t.Error("Paint should show diff prefixes")
	}
}

func TestStreamingDiffChildren(t *testing.T) {
	sd := NewStreamingDiff()
	if c := sd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestStreamingDiffStyle(t *testing.T) {
	sd := NewStreamingDiff()
	sd.SetStyle(StreamingDiffStyle{
		Added:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Removed: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Same:    buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Prefix:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	sd.SetOldText("a").SetNewText("b")
	buf := buffer.NewBuffer(40, 5)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	sd.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintMarkdownMath(b *testing.B) {
	mm := NewMarkdownMath()
	mm.SetExpression("E = mc^2 + integral(f(x)dx)")
	mm.SetBlock(true)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.Paint(buf)
	}
}

func BenchmarkPaintAIResponseScore(b *testing.B) {
	rs := NewAIResponseScore()
	rs.SetScores(90, 85, 95, 80, 88)
	rs.SetBounds(Rect{X: 0, Y: 0, W: 24, H: 5})
	buf := buffer.NewBuffer(24, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Paint(buf)
	}
}

func BenchmarkPaintStreamingDiff(b *testing.B) {
	sd := NewStreamingDiff()
	sd.SetOldText("line1\nline2\nline3\nline4\nline5")
	sd.SetNewText("line1\nchanged\nline3\nadded\nline5")
	sd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.Paint(buf)
	}
}
