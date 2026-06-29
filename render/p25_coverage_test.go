package render

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P25 coverage tests for render package edge cases.

func newTestWriter() *term.Writer {
	return term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
}

func TestP25_RendererFront(t *testing.T) {
	r := New(newTestWriter(), 80, 24)
	if r == nil {
		t.Fatal("New should return non-nil")
	}
	f := r.Front()
	if f == nil {
		t.Fatal("Front() should return non-nil buffer")
	}
}

func TestP25_RendererBeginFrame(t *testing.T) {
	r := New(newTestWriter(), 10, 5)

	r.BeginFrame()

	f := r.Front()
	if f == nil {
		t.Fatal("Front() should return buffer after BeginFrame")
	}
}

func TestP25_RendererResize(t *testing.T) {
	r := New(newTestWriter(), 80, 24)

	r.Resize(40, 12)
	r.BeginFrame()

	f := r.Front()
	if f == nil {
		t.Fatal("Front() should return buffer after resize")
	}
}

func TestP25_RendererEndFrameNoChanges(t *testing.T) {
	r := New(newTestWriter(), 80, 24)

	r.BeginFrame()
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with no changes: %v", err)
	}
}

func TestP25_RendererEndFrameWithChanges(t *testing.T) {
	r := New(newTestWriter(), 80, 24)

	r.BeginFrame()
	f := r.Front()
	f.SetCell(0, 0, buffer.NewCell('X', buffer.Style{}))
	err := r.EndFrame()
	if err != nil {
		t.Fatalf("EndFrame with changes: %v", err)
	}
}

func TestP25_RendererMultipleFrames(t *testing.T) {
	r := New(newTestWriter(), 40, 10)

	for i := 0; i < 5; i++ {
		r.BeginFrame()
		f := r.Front()
		f.SetCell(0, 0, buffer.NewCell(rune('A'+i), buffer.Style{}))
		r.EndFrame()
	}
}

func TestP25_RendererBackBuffer(t *testing.T) {
	r := New(newTestWriter(), 80, 24)

	r.BeginFrame()
	f := r.Front()
	f.SetCell(0, 0, buffer.NewCell('Z', buffer.Style{}))
	r.EndFrame()

	r.BeginFrame()
	f2 := r.Front()
	if f2 == nil {
		t.Fatal("Front() should work on second frame")
	}
}
