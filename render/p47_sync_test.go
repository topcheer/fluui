package render

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP47_SetSyncOutput(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 10, 5)

	if r.SyncOutput() {
		t.Error("expected sync disabled by default")
	}

	r.SetSyncOutput(true)
	if !r.SyncOutput() {
		t.Error("expected sync enabled")
	}

	r.SetSyncOutput(false)
	if r.SyncOutput() {
		t.Error("expected sync disabled")
	}
}

func TestP47_EndFrame_SyncDisabled(t *testing.T) {
	bw := &bytesBuf{}
	tw := term.NewWriter(bw, term.ProfileTrue)
	r := New(tw, 10, 5)

	// Set a cell to trigger diff
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	_ = r.EndFrame()

	output := bw.String()
	// Should NOT contain DCS sync sequences
	if strings.Contains(output, "\x1bP=1s") {
		t.Error("should not contain sync begin when disabled")
	}
	if strings.Contains(output, "\x1bP=2s") {
		t.Error("should not contain sync end when disabled")
	}
}

func TestP47_EndFrame_SyncEnabled(t *testing.T) {
	bw := &bytesBuf{}
	tw := term.NewWriter(bw, term.ProfileTrue)
	r := New(tw, 10, 5)
	r.SetSyncOutput(true)

	// Set a cell to trigger diff
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	_ = r.EndFrame()

	output := bw.String()
	// Should contain DCS sync sequences
	if !strings.Contains(output, "\x1bP=1s") {
		t.Error("expected sync begin (BSU) sequence in output")
	}
	if !strings.Contains(output, "\x1bP=2s") {
		t.Error("expected sync end (ESU) sequence in output")
	}
}

func TestP47_EndFrame_SyncNoChanges(t *testing.T) {
	bw := &bytesBuf{}
	tw := term.NewWriter(bw, term.ProfileTrue)
	r := New(tw, 10, 5)
	r.SetSyncOutput(true)

	// No changes — EndFrame should return early, no sync sequences
	r.BeginFrame()
	_ = r.EndFrame()

	output := bw.String()
	if output != "" {
		t.Errorf("expected empty output for no changes, got %q", output)
	}
}

func TestP47_EndFrame_SyncOrder(t *testing.T) {
	bw := &bytesBuf{}
	tw := term.NewWriter(bw, term.ProfileTrue)
	r := New(tw, 10, 5)
	r.SetSyncOutput(true)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	_ = r.EndFrame()

	output := bw.String()
	// BSU should come before content, ESU after
	bsuIdx := strings.Index(output, "\x1bP=1s")
	esuIdx := strings.Index(output, "\x1bP=2s")
	if bsuIdx < 0 || esuIdx < 0 {
		t.Fatal("expected both sync sequences")
	}
	if bsuIdx >= esuIdx {
		t.Error("BSU should come before ESU")
	}
	// BSU must precede the content ('A')
	contentIdx := strings.Index(output, "A")
	if contentIdx < 0 {
		t.Fatal("expected content 'A' in output")
	}
	if bsuIdx >= contentIdx {
		t.Errorf("BSU (at %d) should come before content 'A' (at %d)", bsuIdx, contentIdx)
	}
	// ESU must follow the content
	if esuIdx <= contentIdx {
		t.Errorf("ESU (at %d) should come after content 'A' (at %d)", esuIdx, contentIdx)
	}
}

// bytesBuf is a simple io.Writer for capturing output in tests.
type bytesBuf struct {
	data []byte
}

func (b *bytesBuf) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuf) String() string {
	return string(b.data)
}
