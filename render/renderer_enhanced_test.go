package render

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Enhanced Test 1: Resize from 80x24 to 120x40 ===

func TestEnhancedRendererResize(t *testing.T) {
	var dw captureWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 80, 24)

	// Frame 1: draw content at original size
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Hello World", buffer.DefaultStyle.WithFg(buffer.Red))
	r.Back().DrawText(0, 1, "Second Line", buffer.DefaultStyle.WithFg(buffer.Green))
	r.EndFrame()

	if r.Width() != 80 || r.Height() != 24 {
		t.Fatalf("before resize: expected 80x24, got %dx%d", r.Width(), r.Height())
	}

	// Resize to 120x40
	r.Resize(120, 40)

	if r.Width() != 120 || r.Height() != 40 {
		t.Fatalf("after resize: expected 120x40, got %dx%d", r.Width(), r.Height())
	}

	// After resize, front buffer should match new dimensions
	if r.front.Width != 120 || r.front.Height != 40 {
		t.Errorf("front buffer: expected 120x40, got %dx%d", r.front.Width, r.front.Height)
	}
	if r.back.Width != 120 || r.back.Height != 40 {
		t.Errorf("back buffer: expected 120x40, got %dx%d", r.back.Width, r.back.Height)
	}

	// Render at new size — content should be drawable at previously out-of-bounds positions
	r.BeginFrame()
	r.Back().DrawText(100, 30, "Far Corner", buffer.DefaultStyle.WithFg(buffer.Cyan))
	r.EndFrame()

	// Verify the cell is actually there
	cell := r.Back().GetCell(100, 30)
	if cell.Rune != 'F' {
		t.Errorf("cell at (100,30): got %q, want 'F'", string(cell.Rune))
	}
	if !cell.Fg.Equal(buffer.Cyan) {
		t.Errorf("cell at (100,30): expected cyan, got %v", cell.Fg)
	}

	// Shrink back down
	r.Resize(40, 10)
	if r.Width() != 40 || r.Height() != 10 {
		t.Errorf("after shrink: expected 40x10, got %dx%d", r.Width(), r.Height())
	}
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Small", buffer.DefaultStyle)
	r.EndFrame()
}

// === Enhanced Test 2: CJK double-width rendering ===

func TestEnhancedRendererCJK(t *testing.T) {
	var dw dummyWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 40, 5)

	r.BeginFrame()

	// Draw CJK text — each character is width 2
	// 你好世界 = 4 CJK characters = display width 8
	r.Back().DrawText(0, 0, "你好世界", buffer.DefaultStyle.WithFg(buffer.Red))

	// Draw ASCII after CJK to verify positioning
	// DrawText advances by rune width, so "AB" should start at x=8
	r.Back().DrawText(8, 0, "AB", buffer.DefaultStyle.WithFg(buffer.Green))

	r.EndFrame()

	// Verify CJK characters have Width=2
	cjkCell := r.Back().GetCell(0, 0)
	if cjkCell.Rune != '你' {
		t.Errorf("cell (0,0): got %q, want '你'", string(cjkCell.Rune))
	}
	if cjkCell.Width != 2 {
		t.Errorf("cell (0,0): expected width 2, got %d", cjkCell.Width)
	}

	// Second CJK char should be at x=2
	cjkCell2 := r.Back().GetCell(2, 0)
	if cjkCell2.Rune != '好' {
		t.Errorf("cell (2,0): got %q, want '好'", string(cjkCell2.Rune))
	}

	// ASCII "A" should be at x=8 (after 4 CJK chars * width 2)
	asciiCell := r.Back().GetCell(8, 0)
	if asciiCell.Rune != 'A' {
		t.Errorf("cell (8,0): got %q, want 'A'", string(asciiCell.Rune))
	}
	if asciiCell.Width != 1 {
		t.Errorf("cell (8,0): expected width 1, got %d", asciiCell.Width)
	}
	if !asciiCell.Fg.Equal(buffer.Green) {
		t.Errorf("cell (8,0): expected green, got %v", asciiCell.Fg)
	}

	// Mixed CJK + ASCII on next line
	r.BeginFrame()
	r.Back().DrawText(0, 1, "A你B好C", buffer.DefaultStyle.WithFg(buffer.Yellow))
	r.EndFrame()

	// A at x=0 (width 1), 你 at x=1 (width 2), B at x=3 (width 1), 好 at x=4 (width 2), C at x=6
	checks := []struct {
		x     int
		rune  rune
		width int
	}{
		{0, 'A', 1},
		{1, '你', 2},
		{3, 'B', 1},
		{4, '好', 2},
		{6, 'C', 1},
	}
	for _, chk := range checks {
		cell := r.Back().GetCell(chk.x, 1)
		if cell.Rune != chk.rune {
			t.Errorf("cell (%d,1): got %q, want %q", chk.x, string(cell.Rune), string(chk.rune))
		}
		if cell.Width != chk.width {
			t.Errorf("cell (%d,1): expected width %d, got %d", chk.x, chk.width, cell.Width)
		}
	}
}

// === Enhanced Test 3: Style overflow (all flags at once) ===

func TestEnhancedRendererStyleOverflow(t *testing.T) {
	var dw captureWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 20, 3)

	// Set all style flags simultaneously
	allFlags := buffer.Bold | buffer.Italic | buffer.Underline | buffer.Reverse |
		buffer.Strikethrough | buffer.Dim

	r.BeginFrame()
	r.Back().DrawText(0, 0, "X", buffer.Style{}.WithFg(buffer.Red).WithFlags(allFlags))
	r.EndFrame()

	// Verify the cell has all flags set
	cell := r.Back().GetCell(0, 0)
	if cell.Flags != allFlags {
		t.Errorf("expected flags %d, got %d", allFlags, cell.Flags)
	}

	// Verify the ANSI output contains all SGR codes
	output := string(dw.bytes)

	// Bold=1, Italic=3, Underline=4, Reverse=7, Strikethrough=9, Dim=2
	expectedSGR := []string{"1", "3", "4", "7", "9", "2"}
	for _, sgr := range expectedSGR {
		if !contains(output, sgr+";") && !strings.HasSuffix(output, sgr) {
			// The SGR parameter might be separated by ; or at end of sequence
			// Check it appears in the output as part of an SGR sequence
			if !contains(output, "\x1b["+sgr) && !contains(output, ";"+sgr+";") && !contains(output, ";"+sgr+"\x1b") {
				t.Errorf("expected SGR code %s in output: %q", sgr, output)
			}
		}
	}

	// Verify the cell renders correctly across diff frames
	r.BeginFrame()
	// Change only the rune, keep all flags
	r.Back().SetCell(0, 0, buffer.Cell{
		Rune:  'Y',
		Width: 1,
		Fg:    buffer.Red,
		Flags: allFlags,
	})
	r.EndFrame()

	cell2 := r.Back().GetCell(0, 0)
	if cell2.Rune != 'Y' {
		t.Errorf("after update: expected 'Y', got %q", string(cell2.Rune))
	}
	if cell2.Flags != allFlags {
		t.Errorf("after update: flags changed, expected %d, got %d", allFlags, cell2.Flags)
	}
}

// === Enhanced Test 4: Diff sparsity — 1 cell change vs full change ===

func TestEnhancedRendererDiffSparsity(t *testing.T) {
	// Use two separate writers to measure output length
	var dw1, dw2 captureWriter

	// Renderer 1: change only 1 cell
	tw1 := term.NewWriter(&dw1, term.ProfileTrue)
	r1 := New(tw1, 20, 5)

	// Initial frame
	r1.BeginFrame()
	for y := 0; y < 5; y++ {
		r1.Back().DrawText(0, y, "AAAAAAAAAAAA", buffer.DefaultStyle)
	}
	r1.EndFrame()

	// Reset capture to measure only the diff frame
	dw1.bytes = nil

	// Frame with 1 cell change
	r1.BeginFrame()
	for y := 0; y < 5; y++ {
		r1.Back().DrawText(0, y, "AAAAAAAAAAAA", buffer.DefaultStyle)
	}
	r1.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1, Fg: buffer.Red})
	r1.EndFrame()

	sparseOutput := len(dw1.bytes)

	// Renderer 2: change all cells
	tw2 := term.NewWriter(&dw2, term.ProfileTrue)
	r2 := New(tw2, 20, 5)

	// Same initial frame
	r2.BeginFrame()
	for y := 0; y < 5; y++ {
		r2.Back().DrawText(0, y, "AAAAAAAAAAAA", buffer.DefaultStyle)
	}
	r2.EndFrame()

	// Reset capture to measure only the diff
	dw2.bytes = nil

	// Change all cells
	r2.BeginFrame()
	for y := 0; y < 5; y++ {
		r2.Back().DrawText(0, y, "BBBBBBBBBBBB", buffer.DefaultStyle.WithFg(buffer.Red))
	}
	r2.EndFrame()

	fullOutput := len(dw2.bytes)

	// Sparse change should produce significantly less output
	if sparseOutput >= fullOutput {
		t.Errorf("sparse diff (%d bytes) should be smaller than full diff (%d bytes)",
			sparseOutput, fullOutput)
	}

	// Sparse change should be much smaller (at least 3x less) than a full redraw
	if fullOutput > 0 && sparseOutput > fullOutput/2 {
		t.Errorf("sparse diff (%d bytes) should be at least 2x smaller than full diff (%d bytes), ratio=%.1f",
			sparseOutput, fullOutput, float64(sparseOutput)/float64(fullOutput))
	}
}

// === Enhanced Test 5: Multiple frames with small changes ===

func TestEnhancedRendererMultipleFrames(t *testing.T) {
	var dw captureWriter
	tw := term.NewWriter(&dw, term.ProfileTrue)
	r := New(tw, 40, 5)

	// Frame 1: draw a counter
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Count: 0", buffer.DefaultStyle.WithFg(buffer.White))
	r.Back().DrawText(0, 1, "Static text line", buffer.DefaultStyle.WithFg(buffer.Green))
	r.EndFrame()

	// Frame 2: change only the counter digit
	dw.bytes = nil // reset — measure only frame 2 output
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Count: 1", buffer.DefaultStyle.WithFg(buffer.White))
	r.Back().DrawText(0, 1, "Static text line", buffer.DefaultStyle.WithFg(buffer.Green))
	r.EndFrame()

	// Capture output length after frame 2 (should be small — just 1 char change)
	frame2OutputLen := len(dw.bytes)

	// Frame 3: change counter again
	dw.bytes = nil // reset capture
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Count: 2", buffer.DefaultStyle.WithFg(buffer.White))
	r.Back().DrawText(0, 1, "Static text line", buffer.DefaultStyle.WithFg(buffer.Green))
	r.EndFrame()

	frame3OutputLen := len(dw.bytes)

	// Both incremental frames should produce similar (small) output
	if frame2OutputLen == 0 {
		t.Error("frame 2 should produce some output for the changed cell")
	}
	if frame3OutputLen == 0 {
		t.Error("frame 3 should produce some output for the changed cell")
	}

	// The outputs should be comparable in size (both are single-char diffs)
	if frame2OutputLen > 0 && frame3OutputLen > 0 {
		ratio := float64(frame3OutputLen) / float64(frame2OutputLen)
		if ratio < 0.5 || ratio > 2.0 {
			t.Errorf("frame output sizes vary too much: frame2=%d, frame3=%d, ratio=%.1f",
				frame2OutputLen, frame3OutputLen, ratio)
		}
	}

	// Verify final state is correct
	cell := r.Back().GetCell(7, 0)
	if cell.Rune != '2' {
		t.Errorf("final counter: expected '2', got %q", string(cell.Rune))
	}

	// Static line should be unchanged
	staticCell := r.Back().GetCell(0, 1)
	if staticCell.Rune != 'S' {
		t.Errorf("static line: expected 'S', got %q", string(staticCell.Rune))
	}
	if !staticCell.Fg.Equal(buffer.Green) {
		t.Errorf("static line: expected green, got %v", staticCell.Fg)
	}

	// Frame 4: no change at all — should produce zero or minimal output
	dw.bytes = nil
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Count: 2", buffer.DefaultStyle.WithFg(buffer.White))
	r.Back().DrawText(0, 1, "Static text line", buffer.DefaultStyle.WithFg(buffer.Green))
	r.EndFrame()

	noChangeOutputLen := len(dw.bytes)
	// No-change frame should produce no diff ops (or just reset sequence)
	if noChangeOutputLen > 10 {
		t.Errorf("no-change frame should produce minimal output, got %d bytes", noChangeOutputLen)
	}
}
