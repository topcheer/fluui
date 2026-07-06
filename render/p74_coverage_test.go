package render

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === P74: BeginFrame coverage (66.7% → 90%+) ===

func TestP74_BeginFrame_CreateBuffer(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	// First BeginFrame creates the back buffer
	r.BeginFrame()
}

func TestP74_BeginFrame_FillSameSize(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	// First call creates buffer
	r.BeginFrame()
	// Second call with same size should Fill (not create new)
	r.BeginFrame()
}

func TestP74_BeginFrame_ResizeRecreate(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.BeginFrame()
	// Change dimensions via Resize
	r.Resize(100, 30)
	// BeginFrame should detect mismatch and recreate
	r.BeginFrame()
}

// === P74: EndFrame with OSC8 links ===

func TestP74_EndFrame_WithLink(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.BeginFrame()
	buf := r.Back()
	link := &buffer.Link{URL: "https://example.com"}
	buf.SetCell(5, 5, buffer.Cell{Rune: 'L', Fg: buffer.NamedColor(buffer.NamedBlue), Link: link})
	buf.SetCell(6, 5, buffer.Cell{Rune: 'i', Fg: buffer.NamedColor(buffer.NamedBlue), Link: link})
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with link failed: %v", err)
	}
}

func TestP74_EndFrame_SyncEnabled(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.SetSyncOutput(true)
	r.BeginFrame()
	buf := r.Back()
	buf.SetCell(0, 0, buffer.Cell{Rune: 'X', Fg: buffer.NamedColor(buffer.NamedWhite)})
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with sync failed: %v", err)
	}
}

func TestP74_EndFrame_MultipleCells(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.BeginFrame()
	buf := r.Back()
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: rune('A' + y), Fg: buffer.NamedColor(buffer.NamedGreen)})
		}
	}
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with multiple cells failed: %v", err)
	}
}

// === P74: Second EndFrame should be fast path (no changes) ===

func TestP74_EndFrame_SecondFrameNoChanges(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.BeginFrame()
	buf := r.Back()
	buf.SetCell(0, 0, buffer.Cell{Rune: 'A', Fg: buffer.NamedColor(buffer.NamedWhite)})
	r.EndFrame()
	// Second frame — no changes
	r.BeginFrame()
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("second EndFrame failed: %v", err)
	}
}

func TestP74_EndFrame_CellWidth0(t *testing.T) {
	wr := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(wr, 80, 24)
	r.BeginFrame()
	buf := r.Back()
	buf.SetCell(5, 5, buffer.Cell{Rune: ' ', Width: 0})
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with Width=0 cell failed: %v", err)
	}
}
