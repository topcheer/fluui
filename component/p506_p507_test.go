package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMiniGaugeBasic(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetValue(50)
	if mg.Value() != 50 { t.Errorf("Value = %f", mg.Value()) }
}

func TestMiniGaugeSetMax(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetMax(200)
	mg.SetValue(100)
	// 100/200 = 50%
}

func TestMiniGaugeSetWidth(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetWidth(20)
	mg.SetValue(50)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	mg.Paint(buf)
	// Check filled chars
	filledCount := 0
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '▰' { filledCount++ }
	}
	if filledCount != 10 { t.Errorf("filled = %d, want 10", filledCount) }
}

func TestMiniGaugeThresholdColors(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetWidth(10)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})

	// Normal (50%)
	mg.SetValue(50)
	buf1 := buffer.NewBuffer(30, 1)
	mg.Paint(buf1)
	normalColor := buf1.GetCell(0, 0).Fg

	// Critical (90%)
	mg.SetValue(90)
	buf2 := buffer.NewBuffer(30, 1)
	mg.Paint(buf2)
	critColor := buf2.GetCell(0, 0).Fg

	if normalColor.Equal(critColor) { t.Error("expected different colors for normal vs critical") }
}

func TestMiniGaugeLabel(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetLabel("CPU")
	mg.SetValue(50)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	mg.Paint(buf)
	foundLabel := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == 'C' { foundLabel = true; break }
	}
	if !foundLabel { t.Error("label not found") }
}

func TestMiniGaugeMeasure(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetWidth(15)
	s := mg.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestMiniGaugeClamp(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetValue(150) // over max(100)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	mg.Paint(buf) // should not panic
}

func TestMiniGaugeZero(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetValue(0)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	mg.Paint(buf)
	// No filled chars expected
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '▰' { t.Error("should have no filled at 0%%"); break }
	}
}

func TestMiniGaugeChildren(t *testing.T) {
	mg := NewMiniGauge()
	if mg.Children() != nil { t.Error("Children should be nil") }
}

func TestMiniGaugeStyle(t *testing.T) {
	mg := NewMiniGauge()
	mg.SetStyle(MiniGaugeStyle{
		Normal: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Warning: buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Critical: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Percent: buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
	})
	mg.SetValue(50)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	mg.Paint(buf)
}

// ─── MarkdownHeading tests ───

func TestMarkdownHeadingBasic(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("# Title")
	if mh.Level() != 1 { t.Errorf("Level = %d, want 1", mh.Level()) }
	if mh.Text() != "Title" { t.Errorf("Text = %q", mh.Text()) }
}

func TestMarkdownHeadingH2(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("## Section")
	if mh.Level() != 2 { t.Errorf("Level = %d, want 2", mh.Level()) }
}

func TestMarkdownHeadingH3to6(t *testing.T) {
	for level := 3; level <= 6; level++ {
		mh := NewMarkdownHeading()
		prefix := ""
		for i := 0; i < level; i++ { prefix += "#" }
		mh.SetMarkdown(prefix + " Text")
		if mh.Level() != level { t.Errorf("Level = %d, want %d", mh.Level(), level) }
	}
}

func TestMarkdownHeadingNotHeading(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("Just plain text")
	if mh.Level() != 0 { t.Errorf("Level = %d, want 0", mh.Level()) }
}

func TestMarkdownHeadingEmpty(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("")
	if mh.Level() != 0 { t.Errorf("Level = %d, want 0", mh.Level()) }
}

func TestMarkdownHeadingMeasure(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("# Title")
	s := mh.Measure(Constraints{})
	if s.W < 5 { t.Errorf("W = %d", s.W) }
	if s.H != 2 { t.Errorf("H = %d, want 2 (H1 has underline)", s.H) }
}

func TestMarkdownHeadingPaint(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("# Big Title")
	mh.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	mh.Paint(buf)
	// Check text exists
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 0).Rune == 'B' { foundText = true; break }
	}
	if !foundText { t.Error("heading text not found") }
	// Check H1 underline (═ chars on row 1)
	foundUnderline := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '═' { foundUnderline = true; break }
	}
	if !foundUnderline { t.Error("H1 underline not found") }
}

func TestMarkdownHeadingPaintH2(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("## Section")
	mh.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 2})
	buf := buffer.NewBuffer(50, 2)
	mh.Paint(buf)
	// H2 uses ─ not ═
	foundDash := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '─' { foundDash = true; break }
	}
	if !foundDash { t.Error("H2 dash underline not found") }
}

func TestMarkdownHeadingPaintH3(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("### Subsection")
	mh.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	mh.Paint(buf)
	// H3 has no underline, text only
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 0).Rune == 'S' { foundText = true; break }
	}
	if !foundText { t.Error("H3 text not found") }
}

func TestMarkdownHeadingChildren(t *testing.T) {
	mh := NewMarkdownHeading()
	if mh.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownHeadingStyle(t *testing.T) {
	mh := NewMarkdownHeading()
	mh.SetStyle(MarkdownHeadingStyle{
		Levels: [6]buffer.Style{{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold}, {Fg: buffer.RGB(200, 200, 200)}, {Fg: buffer.RGB(150, 150, 150)}, {}, {}, {}},
		Underline: [2]buffer.Style{{Fg: buffer.RGB(100, 100, 100)}, {Fg: buffer.RGB(80, 80, 80)}},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	mh.SetMarkdown("# Test")
	mh.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	mh.Paint(buf)
}
