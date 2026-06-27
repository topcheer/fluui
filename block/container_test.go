package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// mockBlock is a minimal Block implementation for container testing.
type mockBlock struct {
	BaseBlock
	content string
	width   int
	height  int
}

func newMockBlock(id string, bt BlockType, w, h int) *mockBlock {
	m := &mockBlock{
		BaseBlock: NewBaseBlock(id, bt),
		width:     w,
		height:    h,
	}
	return m
}

func (m *mockBlock) Measure(cs component.Constraints) component.Size {
	w, h := m.width, m.height
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return component.Size{W: w, H: h}
}

func (m *mockBlock) SetBounds(r component.Rect) {
	m.BaseComponent.SetBounds(r)
}

func (m *mockBlock) Paint(buf *buffer.Buffer) {
	// no-op
}

func TestContainerEmpty(t *testing.T) {
	c := NewBlockContainer()
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
	if c.LastBlock() != nil {
		t.Error("LastBlock() on empty container should be nil")
	}
}

func TestContainerAddBlock(t *testing.T) {
	c := NewBlockContainer()
	b1 := newMockBlock("b1", TypeThinking, 10, 2)
	b2 := newMockBlock("b2", TypeAssistantText, 20, 5)

	c.AddBlock(b1)
	c.AddBlock(b2)

	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
	if c.LastBlock().ID() != "b2" {
		t.Errorf("LastBlock().ID() = %q, want 'b2'", c.LastBlock().ID())
	}
}

func TestContainerRemoveBlock(t *testing.T) {
	c := NewBlockContainer()
	b1 := newMockBlock("b1", TypeThinking, 10, 2)
	b2 := newMockBlock("b2", TypeAssistantText, 20, 5)

	c.AddBlock(b1)
	c.AddBlock(b2)

	if !c.RemoveBlock("b1") {
		t.Error("RemoveBlock('b1') should return true")
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
	if c.LastBlock().ID() != "b2" {
		t.Errorf("LastBlock().ID() = %q, want 'b2'", c.LastBlock().ID())
	}

	if c.RemoveBlock("nonexistent") {
		t.Error("RemoveBlock('nonexistent') should return false")
	}
}

func TestContainerMeasure(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 2))
	c.AddBlock(newMockBlock("b2", TypeAssistantText, 20, 5))

	// spacing=1: total height = 2 + 1 + 5 = 8, max width = 20
	sz := c.Measure(component.Unbounded())
	if sz.H != 8 {
		t.Errorf("height = %d, want 8", sz.H)
	}
	if sz.W != 20 {
		t.Errorf("width = %d, want 20", sz.W)
	}
}

func TestContainerMeasureWithConstraints(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(newMockBlock("b1", TypeThinking, 50, 2))

	sz := c.Measure(component.Bounded(30, 10))
	if sz.W != 30 {
		t.Errorf("width = %d, want 30 (clamped)", sz.W)
	}
}

func TestContainerSetBounds(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	b1 := newMockBlock("b1", TypeThinking, 10, 3)
	b2 := newMockBlock("b2", TypeAssistantText, 10, 5)
	c.AddBlock(b1)
	c.AddBlock(b2)

	c.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 20})

	// b1 should be at Y=0, b2 at Y=3
	if r := b1.Bounds(); r.Y != 0 {
		t.Errorf("b1 Y = %d, want 0", r.Y)
	}
	if r := b2.Bounds(); r.Y != 3 {
		t.Errorf("b2 Y = %d, want 3", r.Y)
	}
}

func TestContainerSetBoundsWithSpacing(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(2)
	b1 := newMockBlock("b1", TypeThinking, 10, 3)
	b2 := newMockBlock("b2", TypeAssistantText, 10, 5)
	c.AddBlock(b1)
	c.AddBlock(b2)

	c.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 20})

	// b1 at Y=0, b2 at Y=3+2=5
	if r := b2.Bounds(); r.Y != 5 {
		t.Errorf("b2 Y = %d, want 5 (with spacing=2)", r.Y)
	}
}

func TestContainerBlocks(t *testing.T) {
	c := NewBlockContainer()
	b1 := newMockBlock("b1", TypeThinking, 10, 2)
	b2 := newMockBlock("b2", TypeAssistantText, 20, 5)
	c.AddBlock(b1)
	c.AddBlock(b2)

	blocks := c.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("len(Blocks()) = %d, want 2", len(blocks))
	}
	if blocks[0].ID() != "b1" || blocks[1].ID() != "b2" {
		t.Error("Blocks() order mismatch")
	}
}

func TestContainerChildren(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 2))

	children := c.Children()
	if len(children) != 1 {
		t.Errorf("len(Children()) = %d, want 1", len(children))
	}
}

func TestContainerIsDirty(t *testing.T) {
	c := NewBlockContainer()
	b1 := newMockBlock("b1", TypeThinking, 10, 2)
	c.AddBlock(b1)

	// Container should be dirty after AddBlock
	if !c.IsDirty() {
		t.Error("container should be dirty after AddBlock")
	}

	// Clear dirty
	c.ClearDirty()
	if c.IsDirty() {
		t.Error("container should not be dirty after ClearDirty")
	}

	// Mark child dirty → container dirty
	b1.MarkDirty()
	if !c.IsDirty() {
		t.Error("container should be dirty when child is dirty")
	}
}

func TestContainerPaint(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 2))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})

	buf := buffer.NewBuffer(10, 5)
	// Should not panic
	c.Paint(buf)
}
