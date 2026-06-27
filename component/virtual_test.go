package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// mockVisiblePainter is a test type that implements VisiblePainter.
type mockVisiblePainter struct {
	BaseComponent
	painted bool
	y0, y1  int
}

func (m *mockVisiblePainter) Measure(cs Constraints) Size { return Size{W: 80, H: 100} }

func (m *mockVisiblePainter) Paint(buf *buffer.Buffer) {}

func (m *mockVisiblePainter) PaintVisible(buf *buffer.Buffer, visibleY0, visibleY1 int) {
	m.painted = true
	m.y0 = visibleY0
	m.y1 = visibleY1
}

// mockNonVisiblePainter is a test type that does NOT implement VisiblePainter.
type mockNonVisiblePainter struct {
	BaseComponent
}

func (m *mockNonVisiblePainter) Measure(cs Constraints) Size { return Size{W: 80, H: 100} }

func (m *mockNonVisiblePainter) Paint(buf *buffer.Buffer) {}

func TestIsVisiblePainter_True(t *testing.T) {
	mvp := &mockVisiblePainter{}
	if !IsVisiblePainter(mvp) {
		t.Error("IsVisiblePainter: expected true for mockVisiblePainter")
	}
}

func TestIsVisiblePainter_False(t *testing.T) {
	mnvp := &mockNonVisiblePainter{}
	if IsVisiblePainter(mnvp) {
		t.Error("IsVisiblePainter: expected false for mockNonVisiblePainter")
	}
}

func TestIsVisiblePainter_TextComponent(t *testing.T) {
	text := NewText("hello")
	if IsVisiblePainter(text) {
		t.Error("IsVisiblePainter: expected false for Text")
	}
}

func TestIsVisiblePainter_BorderComponent(t *testing.T) {
	border := NewBorder(NewText("x"))
	if IsVisiblePainter(border) {
		t.Error("IsVisiblePainter: expected false for Border")
	}
}

func TestVisiblePainterTypeAssertion(t *testing.T) {
	mvp := &mockVisiblePainter{}
	var c Component = mvp

	vp, ok := c.(VisiblePainter)
	if !ok {
		t.Fatal("type assertion to VisiblePainter failed")
	}

	buf := buffer.NewBuffer(10, 10)
	vp.PaintVisible(buf, 5, 15)

	if !mvp.painted {
		t.Error("PaintVisible was not called")
	}
	if mvp.y0 != 5 || mvp.y1 != 15 {
		t.Errorf("visible range: got [%d,%d), want [5,15)", mvp.y0, mvp.y1)
	}
}

func TestVisiblePainterTypeAssertionFails(t *testing.T) {
	var c Component = NewText("hello")

	vp, ok := c.(VisiblePainter)
	if ok {
		t.Error("type assertion should fail for Text")
	}
	if vp != nil {
		t.Error("expected nil interface on failed assertion")
	}
}
