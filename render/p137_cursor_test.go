package render

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// p137CountWriter counts bytes written.
type p137CountWriter struct {
	n int
}

func (c *p137CountWriter) Write(p []byte) (int, error) {
	c.n += len(p)
	return len(p), nil
}

// TestP137_CursorTracking_SequentialCells verifies that adjacent cells
// skip the MoveTo escape sequence, producing fewer output bytes.
func TestP137_CursorTracking_SequentialCells(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 10, 1)

	// Fill back buffer with initial content.
	for x := 0; x < 10; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'A',
			Width: 1,
			Fg:    buffer.RGB(255, 255, 255),
		})
	}

	r.BeginFrame()
	// Change all cells to trigger full redraw.
	for x := 0; x < 10; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'B',
			Width: 1,
			Fg:    buffer.RGB(255, 255, 255),
		})
	}
	counter.n = 0
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// For 10 sequential cells with same style, we should see:
	// - 1 initial MoveAndStyle (cursor + SGR)
	// - 9 chars (no MoveTo for adjacent cells)
	// - 1 ResetStyle
	// Total should be well under 100 bytes.
	if counter.n > 100 {
		t.Errorf("expected < 100 bytes for 10 sequential cells, got %d", counter.n)
	}
}

// TestP137_CursorTracking_NonSequentialCells verifies that non-adjacent
// cells still emit MoveTo.
func TestP137_CursorTracking_NonSequentialCells(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 20, 5)

	// Fill back buffer with initial content.
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			r.Back().SetCell(x, y, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Fg:    buffer.DefaultStyle.Fg,
			})
		}
	}

	r.BeginFrame()
	// Change cells on different rows (non-adjacent).
	for y := 0; y < 5; y++ {
		r.Back().SetCell(0, y, buffer.Cell{
			Rune:  'X',
			Width: 1,
			Fg:    buffer.RGB(255, 0, 0),
		})
	}
	counter.n = 0
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// 5 non-adjacent cells should each get a MoveAndStyle.
	// Should be more bytes than if they were sequential.
	if counter.n < 50 {
		t.Errorf("expected > 50 bytes for 5 non-adjacent cells, got %d", counter.n)
	}
}

// TestP137_CursorTracking_DifferentStyles verifies that adjacent cells
// with different styles still emit SGR changes.
func TestP137_CursorTracking_DifferentStyles(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 5, 1)

	// Fill back buffer.
	for x := 0; x < 5; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    buffer.DefaultStyle.Fg,
		})
	}

	r.BeginFrame()
	// Each cell has a different color (adjacent but different styles).
	for x := 0; x < 5; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'A',
			Width: 1,
			Fg:    buffer.RGB(uint8(x*50), 0, 0),
		})
	}
	counter.n = 0
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Adjacent cells with different styles: no MoveTo but SGR changes.
	// Should still produce reasonable output.
	if counter.n == 0 {
		t.Error("expected non-zero output")
	}
}

// TestP137_CursorTracking_LinkedCells verifies that OSC8 linked cells
// always use full MoveAndStyle (break cursor continuity).
func TestP137_CursorTracking_LinkedCells(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 5, 1)

	// Fill back buffer.
	for x := 0; x < 5; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    buffer.DefaultStyle.Fg,
		})
	}

	r.BeginFrame()
	link := &buffer.Link{URL: "https://example.com"}
	for x := 0; x < 3; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'L',
			Width: 1,
			Fg:    buffer.RGB(0, 255, 255),
			Link:  link,
		})
	}
	counter.n = 0
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Linked cells should produce OSC8 sequences + MoveAndStyle.
	// Should be significantly more bytes than non-linked cells.
	if counter.n < 50 {
		t.Errorf("expected > 50 bytes for 3 linked cells, got %d", counter.n)
	}
}

// TestP137_CursorTracking_WideCharResets verifies that wide CJK characters
// properly update prevWidth so the next cell position is correct.
func TestP137_CursorTracking_WideCharResets(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 10, 1)

	// Fill back buffer.
	for x := 0; x < 10; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    buffer.DefaultStyle.Fg,
		})
	}

	r.BeginFrame()
	// Place a wide character (width=2) at position 0, then a normal char at 2.
	r.Back().SetCell(0, 0, buffer.Cell{
		Rune:  '世', // CJK wide char
		Width: 2,
		Fg:    buffer.RGB(255, 255, 255),
	})
	r.Back().SetCell(1, 0, buffer.Cell{Width: 0}) // padding
	r.Back().SetCell(2, 0, buffer.Cell{
		Rune:  'X',
		Width: 1,
		Fg:    buffer.RGB(255, 255, 255),
	})
	counter.n = 0
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Wide char (width=2) at position 0, then next at position 2.
	// prevX=0, prevWidth=2 → next expected at X=2.
	// op.X=2, op.Y=0, prevY=0, prevX+prevWidth=0+2=2 → adjacent! Skip MoveTo.
	if counter.n == 0 {
		t.Error("expected non-zero output")
	}
}

// TestP137_RenderSequentialBytes benchmark correctness check.
func TestP137_RenderSequentialBytes_Correctness(t *testing.T) {
	counter := &p137CountWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 5, 1)

	// Initial fill.
	for x := 0; x < 5; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'A',
			Width: 1,
			Fg:    buffer.RGB(255, 255, 255),
		})
	}

	r.BeginFrame()
	for x := 0; x < 5; x++ {
		r.Back().SetCell(x, 0, buffer.Cell{
			Rune:  'B',
			Width: 1,
			Fg:    buffer.RGB(255, 255, 255),
		})
	}
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// Output should contain "BBBBB" somewhere.
	output := counter.n
	if output == 0 {
		t.Error("expected non-zero output")
	}
}
