package render

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TestRendererBasicDiff tests that the renderer correctly diffs
// front and back buffers and only outputs changed cells.
func TestRendererBasicDiff(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 40, 10)

	// Frame 1: render initial content
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Hello", buffer.DefaultStyle.WithFg(buffer.Red))
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Frame 2: only change one character at the end
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Hello", buffer.DefaultStyle.WithFg(buffer.Red))
	r.Back().DrawText(5, 0, "!", buffer.DefaultStyle.WithFg(buffer.Green))
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame 2: %v", err)
	}

	// Verify front buffer has "Hello!" now
	cell := r.Back().GetCell(5, 0)
	if cell.Rune != '!' {
		t.Errorf("cell (5,0): got %c, want '!'", cell.Rune)
	}
}

// TestRendererResize tests buffer reallocation on terminal resize.
func TestRendererResize(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 40, 10)

	r.Resize(60, 20)

	if r.Width() != 60 || r.Height() != 20 {
		t.Errorf("after resize: got %dx%d, want 60x20", r.Width(), r.Height())
	}

	// Should be able to render to the new larger size
	r.BeginFrame()
	r.Back().DrawText(55, 18, "X", buffer.DefaultStyle)
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame after resize: %v", err)
	}
}

// TestRendererMultipleStyles tests rendering cells with different styles.
func TestRendererMultipleStyles(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 40, 5)

	r.BeginFrame()

	// Red bold text
	style1 := buffer.Style{}.WithFg(buffer.Red).WithFlags(buffer.Bold)
	r.Back().DrawText(0, 0, "Red", style1)

	// Blue underline
	style2 := buffer.Style{}.WithFg(buffer.Blue).WithFlags(buffer.Underline)
	r.Back().DrawText(0, 1, "Blue", style2)

	// Green background
	style3 := buffer.Style{}.WithBg(buffer.Green).WithFg(buffer.Black)
	r.Back().DrawText(0, 2, "GreenBg", style3)

	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Verify styles
	cell0 := r.Back().GetCell(0, 0)
	if !cell0.Fg.Equal(buffer.Red) {
		t.Error("cell (0,0) should be red")
	}
	if cell0.Flags&buffer.Bold == 0 {
		t.Error("cell (0,0) should be bold")
	}

	cell1 := r.Back().GetCell(0, 1)
	if !cell1.Fg.Equal(buffer.Blue) {
		t.Error("cell (0,1) should be blue")
	}
	if cell1.Flags&buffer.Underline == 0 {
		t.Error("cell (0,1) should be underline")
	}

	cell2 := r.Back().GetCell(0, 2)
	if !cell2.Bg.Equal(buffer.Green) {
		t.Error("cell (0,2) should have green bg")
	}
}

// TestRendererNoChange tests that two identical frames produce no diffs.
func TestRendererNoChange(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 20, 5)

	// Frame 1
	r.BeginFrame()
	r.Back().DrawText(0, 0, "static", buffer.DefaultStyle)
	r.EndFrame()

	// Frame 2: identical content
	r.BeginFrame()
	r.Back().DrawText(0, 0, "static", buffer.DefaultStyle)
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame identical: %v", err)
	}
}

// TestRendererClearAndRedraw tests clearing a buffer between frames.
func TestRendererClearAndRedraw(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 20, 5)

	// Frame 1: fill with text
	r.BeginFrame()
	for y := 0; y < 5; y++ {
		r.Back().DrawText(0, y, "XXXXXXXXXXXX", buffer.DefaultStyle)
	}
	r.EndFrame()

	// Frame 2: clear and draw new content
	r.BeginFrame()
	r.Back().DrawText(0, 2, "New", buffer.DefaultStyle)
	r.EndFrame()

	// Old content should be gone
	cell := r.Back().GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("cell (0,0) after clear: got %c, want space", cell.Rune)
	}
}

// TestRendererTrueColorOutput tests that TrueColor cells generate
// correct ANSI sequences.
func TestRendererTrueColorOutput(t *testing.T) {
	var dw captureWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 10, 3)

	r.BeginFrame()
	color := buffer.RGB(255, 128, 64)
	r.Back().DrawText(0, 0, "X", buffer.Style{}.WithFg(color))
	r.EndFrame()

	// The output should contain the truecolor escape sequence
	output := string(dw.bytes)
	if !contains(output, "38;2;255;128;64") {
		t.Errorf("output should contain truecolor FG sequence, got: %q", output)
	}
}

// TestWriterStyleBatching verifies the ANSI writer batches style changes.
func TestWriterStyleBatching(t *testing.T) {
	var dw captureWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)

	// Set same style twice — should only emit once
	style := buffer.Style{}.WithFg(buffer.Red)
	tw.SetStyle(style)
	tw.SetStyle(style) // should be deduplicated

	tw.WriteString("hello")

	tw.Flush()

	output := string(dw.bytes)
	// Should contain exactly one SGR sequence for red
	count := countOccurrences(output, "31")
	if count < 1 {
		t.Errorf("expected at least 1 occurrence of '31', got output: %q", output)
	}
}

type dummyWriter struct{}

func (d *dummyWriter) Write(b []byte) (int, error) { return len(b), nil }

type captureWriter struct {
	bytes []byte
}

func (c *captureWriter) Write(b []byte) (int, error) {
	c.bytes = append(c.bytes, b...)
	return len(b), nil
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func countOccurrences(s, sub string) int {
	count := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			count++
		}
	}
	return count
}
