package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownImageBasic(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("![Logo](https://example.com/logo.png)")
	if mi.ImageCount() != 1 { t.Errorf("ImageCount = %d, want 1", mi.ImageCount()) }
}

func TestMarkdownImageMultiple(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("![A](url1) text ![B](url2)")
	if mi.ImageCount() != 2 { t.Errorf("ImageCount = %d, want 2", mi.ImageCount()) }
}

func TestMarkdownImageWithText(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("Before ![img](url) after")
	if mi.ImageCount() != 1 { t.Errorf("ImageCount = %d, want 1", mi.ImageCount()) }
}

func TestMarkdownImageNoImages(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("Just plain text")
	if mi.ImageCount() != 0 { t.Errorf("ImageCount = %d, want 0", mi.ImageCount()) }
}

func TestMarkdownImageEmpty(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("")
	if mi.ImageCount() != 0 { t.Errorf("ImageCount = %d, want 0", mi.ImageCount()) }
}

func TestMarkdownImageMalformed(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("![unclosed image")
	if mi.ImageCount() != 0 { t.Errorf("ImageCount = %d, want 0 (malformed)", mi.ImageCount()) }
}

func TestMarkdownImageMeasure(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("![x](y)")
	s := mi.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownImagePaint(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("Logo ![Fluui](https://fluui.dev/logo.png) here")
	mi.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	mi.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundBracket := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 1).Rune == '[' { foundBracket = true; break }
	}
	if !foundBracket { t.Error("image placeholder bracket not found") }
}

func TestMarkdownImageChildren(t *testing.T) {
	mi := NewMarkdownImage()
	if mi.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownImageStyle(t *testing.T) {
	mi := NewMarkdownImage()
	mi.SetStyle(MarkdownImageStyle{Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Alt: buffer.Style{Fg: buffer.RGB(0,255,0)}, URL: buffer.Style{Fg: buffer.RGB(100,100,100)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	mi.SetMarkdown("![x](y)")
	mi.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	mi.Paint(buf)
}

// ─── SpinnerDots tests ───

func TestSpinnerDotsBasic(t *testing.T) {
	sd := NewSpinnerDots()
	if sd.DotCount() != 3 { t.Errorf("DotCount = %d, want 3 (default)", sd.DotCount()) }
	if sd.Current() != 0 { t.Errorf("Current = %d, want 0", sd.Current()) }
}

func TestSpinnerDotsAdvance(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetDotCount(3)
	sd.Advance()
	if sd.Current() != 1 { t.Errorf("Current = %d, want 1", sd.Current()) }
	sd.Advance()
	if sd.Current() != 2 { t.Errorf("Current = %d, want 2", sd.Current()) }
}

func TestSpinnerDotsWrap(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetDotCount(3)
	sd.Advance()
	sd.Advance()
	sd.Advance() // should wrap to 0
	if sd.Current() != 0 { t.Errorf("After wrap: Current = %d, want 0", sd.Current()) }
}

func TestSpinnerDotsReset(t *testing.T) {
	sd := NewSpinnerDots()
	sd.Advance()
	sd.Reset()
	if sd.Current() != 0 { t.Errorf("After reset: Current = %d, want 0", sd.Current()) }
}

func TestSpinnerDotsLabel(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetLabel("Processing")
	if sd.Label() != "Processing" { t.Errorf("Label = %q", sd.Label()) }
}

func TestSpinnerDotsSetDotCount(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetDotCount(5)
	if sd.DotCount() != 5 { t.Errorf("DotCount = %d, want 5", sd.DotCount()) }
}

func TestSpinnerDotsMeasure(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetLabel("Loading")
	s := sd.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestSpinnerDotsPaint(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetLabel("Loading")
	sd.SetDotCount(3)
	sd.Advance() // active = dot 1
	sd.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	sd.Paint(buf)
	// Check active dot exists
	foundActive := false
	for x := 0; x < 30; x++ {
		if buf.GetCell(x, 0).Rune == '●' { foundActive = true; break }
	}
	if !foundActive { t.Error("active dot not found") }
	// Check inactive dots
	foundInactive := false
	for x := 0; x < 30; x++ {
		if buf.GetCell(x, 0).Rune == '○' { foundInactive = true; break }
	}
	if !foundInactive { t.Error("inactive dot not found") }
}

func TestSpinnerDotsChildren(t *testing.T) {
	sd := NewSpinnerDots()
	if sd.Children() != nil { t.Error("Children should be nil") }
}

func TestSpinnerDotsStyle(t *testing.T) {
	sd := NewSpinnerDots()
	sd.SetStyle(SpinnerDotsStyle{Active: buffer.Style{Fg: buffer.RGB(0,255,0), Flags: buffer.Bold}, Inactive: buffer.Style{Fg: buffer.RGB(50,50,50)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}})
	sd.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	sd.Paint(buf)
}
