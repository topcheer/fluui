package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewCodeBlock_Basic(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}")
	if cb == nil {
		t.Fatal("NewCodeBlock returned nil")
	}
	if cb.Language() != "go" {
		t.Errorf("Language: got %q want %q", cb.Language(), "go")
	}
	if cb.Source() == "" {
		t.Error("Source should not be empty")
	}
}

func TestCodeBlock_LineCount(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	if cb.LineCount() != 3 {
		t.Errorf("LineCount: got %d want 3", cb.LineCount())
	}
}

func TestCodeBlock_LineCount_EmptySource(t *testing.T) {
	cb := NewCodeBlock("go", "")
	// Empty string split by \n produces [""], which is 1 line
	if cb.LineCount() != 1 {
		t.Errorf("empty LineCount: got %d want 1", cb.LineCount())
	}
}

func TestCodeBlock_SetSource(t *testing.T) {
	cb := NewCodeBlock("go", "old code")
	cb.SetSource("new code\nsecond line")
	if cb.Source() != "new code\nsecond line" {
		t.Error("SetSource failed")
	}
	if cb.LineCount() != 2 {
		t.Errorf("LineCount after SetSource: got %d want 2", cb.LineCount())
	}
}

func TestCodeBlock_SetLanguage(t *testing.T) {
	cb := NewCodeBlock("go", "print('hello')")
	cb.SetLanguage("python")
	if cb.Language() != "python" {
		t.Errorf("SetLanguage: got %q", cb.Language())
	}
}

func TestCodeBlock_SetTitle(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetTitle("main.go")
	// showTitle should be auto-enabled when title is set
	cb.mu.RLock()
	show := cb.showTitle
	cb.mu.RUnlock()
	if !show {
		t.Error("SetTitle should auto-enable showTitle")
	}
}

func TestCodeBlock_SetShowTitle(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetShowTitle(true)
	cb.mu.RLock()
	show := cb.showTitle
	cb.mu.RUnlock()
	if !show {
		t.Error("SetShowTitle(true) failed")
	}
	cb.SetShowTitle(false)
	cb.mu.RLock()
	show = cb.showTitle
	cb.mu.RUnlock()
	if show {
		t.Error("SetShowTitle(false) failed")
	}
}

func TestCodeBlock_SetShowLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetShowLineNumbers(true)
	cb.mu.RLock()
	show := cb.showLineNumbers
	cb.mu.RUnlock()
	if !show {
		t.Error("SetShowLineNumbers(true) failed")
	}
}

func TestCodeBlock_ScrollDown(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollDown(2)
	if cb.ScrollOffset() != 2 {
		t.Errorf("ScrollDown: offset = %d, want 2", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollDown_Clamped(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollDown(100)
	// max scroll = 5 lines - 3 visible = 2
	if cb.ScrollOffset() != 2 {
		t.Errorf("ScrollDown clamp: offset = %d, want 2", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollUp(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollDown(2)
	cb.ScrollUp(1)
	if cb.ScrollOffset() != 1 {
		t.Errorf("ScrollUp: offset = %d, want 1", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollUp_Clamped(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.ScrollUp(10)
	if cb.ScrollOffset() != 0 {
		t.Errorf("ScrollUp clamp: offset = %d, want 0", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollTo(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollTo(3)
	if cb.ScrollOffset() != 3 {
		t.Errorf("ScrollTo: offset = %d, want 3", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollTo_Negative(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.ScrollTo(-5)
	if cb.ScrollOffset() != 0 {
		t.Errorf("ScrollTo negative: offset = %d, want 0", cb.ScrollOffset())
	}
}

func TestCodeBlock_ScrollTo_TooFar(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollTo(100)
	// max = 5 - 3 = 2
	if cb.ScrollOffset() != 2 {
		t.Errorf("ScrollTo too far: offset = %d, want 2", cb.ScrollOffset())
	}
}

func TestCodeBlock_Measure_BasicWidth(t *testing.T) {
	cb := NewCodeBlock("text", "hello world")
	size := cb.Measure(Unbounded())
	if size.W < 11 {
		t.Errorf("Measure width: got %d, want >= 11", size.W)
	}
	if size.H < 1 {
		t.Errorf("Measure height: got %d, want >= 1", size.H)
	}
}

func TestCodeBlock_Measure_Multiline(t *testing.T) {
	cb := NewCodeBlock("text", "a\nbb\nccc")
	size := cb.Measure(Unbounded())
	if size.H != 3 {
		t.Errorf("Measure multiline height: got %d, want 3", size.H)
	}
	if size.W != 3 {
		t.Errorf("Measure multiline width: got %d, want 3", size.W)
	}
}

func TestCodeBlock_Measure_WithTitle(t *testing.T) {
	cb := NewCodeBlock("text", "code")
	cb.SetShowTitle(true)
	size := cb.Measure(Unbounded())
	if size.H != 2 { // 1 title + 1 code
		t.Errorf("Measure with title: H = %d, want 2", size.H)
	}
}

func TestCodeBlock_Measure_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("text", "line1\nline2\nline3")
	cb.SetShowLineNumbers(true)
	size := cb.Measure(Unbounded())
	// gutter = 2 digits + 1 separator = 3; width = 5 + 3 = 8
	if size.W != 8 {
		t.Errorf("Measure with line numbers: W = %d, want 8", size.W)
	}
}

func TestCodeBlock_Measure_ConstrainedWidth(t *testing.T) {
	cb := NewCodeBlock("text", "a very long line of text")
	size := cb.Measure(Constraints{MaxWidth: 5})
	if size.W > 5 {
		t.Errorf("Constrained width: got %d, want <= 5", size.W)
	}
}

func TestCodeBlock_Paint_NoPanic(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {\n  println(\"hi\")\n}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
	// No panic = success
}

func TestCodeBlock_Paint_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("text", "hello\nworld")
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
	// Check that line number "1" is at position (1,0) — right-justified
	cell := buf.GetCell(1, 0)
	if cell.Rune != '1' {
		t.Errorf("Expected line number '1' at (1,0), got rune %d", cell.Rune)
	}
}

func TestCodeBlock_Paint_WithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetTitle("main.go")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 3})
	buf := buffer.NewBuffer(15, 3)
	cb.Paint(buf)
	// Title bar should have content at row 0
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("Expected title content at (0,0)")
	}
}

func TestCodeBlock_Paint_Scrolled(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	cb.ScrollTo(2)
	buf := buffer.NewBuffer(10, 2)
	cb.Paint(buf)
	// After scrolling 2 lines, visible content should be lines 3,4
	// Line 3 = "3", should be at row 0
	cell := buf.GetCell(0, 0)
	if cell.Rune != '3' {
		t.Errorf("Expected '3' at (0,0) after scroll, got %+v", cell)
	}
}

func TestCodeBlock_Paint_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "x")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf) // should not panic
}

func TestCodeBlock_Paint_FillsBlankLines(t *testing.T) {
	cb := NewCodeBlock("text", "short")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	cb.Paint(buf)
	// Rows 1-4 should be blank-filled
	for y := 1; y < 5; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == 0 {
				t.Errorf("Missing cell at (%d,%d)", x, y)
			}
		}
	}
}

func TestCodeBlock_Highlighter_ProducesColoredCells(t *testing.T) {
	// Go keywords should have a color set by the chroma highlighter
	cb := NewCodeBlock("go", "func main() {}")
	cb.mu.RLock()
	lines := cb.lines
	cb.mu.RUnlock()
	if len(lines) == 0 {
		t.Fatal("no highlighted lines")
	}
	// The first cell should be 'f' of 'func' and should have a non-zero color
	found := false
	for _, line := range lines {
		for _, cell := range line {
			if cell.Fg != (buffer.Color{}) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("Expected at least one colored cell from syntax highlighting")
	}
}

func TestCodeBlock_UnknownLanguage_NoError(t *testing.T) {
	cb := NewCodeBlock("nonexistent-lang-xyz", "some code")
	if cb.LineCount() != 1 {
		t.Errorf("Unknown language LineCount: got %d want 1", cb.LineCount())
	}
}

func TestCodeBlock_ConcurrentAccess(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	done := make(chan struct{})

	// Concurrent reader
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			cb.Source()
			cb.Language()
			cb.LineCount()
			cb.ScrollOffset()
			cb.Measure(Unbounded())
			buf := buffer.NewBuffer(40, 10)
			cb.Paint(buf)
		}
	}()

	// Concurrent writer
	for i := 0; i < 100; i++ {
		cb.SetSource("package main\nfunc f" + string(rune('0'+i%10)) + "() {}")
		cb.ScrollDown(1)
		cb.ScrollUp(1)
	}

	<-done
}

func TestCodeBlock_SetSource_ResetsScroll(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5\n6\n7\n8")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollDown(3)
	if cb.ScrollOffset() != 3 {
		t.Fatalf("pre-condition: offset = %d want 3", cb.ScrollOffset())
	}
	cb.SetSource("a\nb")
	if cb.ScrollOffset() != 0 {
		t.Errorf("SetSource should reset scroll, got offset %d", cb.ScrollOffset())
	}
}

func TestCodeBlock_SetLanguage_ResetsScroll(t *testing.T) {
	cb := NewCodeBlock("text", "1\n2\n3\n4\n5\n6\n7\n8")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.ScrollDown(3)
	cb.SetLanguage("go")
	if cb.ScrollOffset() != 0 {
		t.Errorf("SetLanguage should reset scroll, got offset %d", cb.ScrollOffset())
	}
}
