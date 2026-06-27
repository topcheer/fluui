package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestContainerWithRealBlocks(t *testing.T) {
	c := NewBlockContainer()

	user := NewUserMessageBlock("u1", "Hello")
	asst := NewAssistantTextBlock("a1")
	asst.AppendDelta("Hi there")
	tool := NewToolCallBlock("t1", "bash", "echo hi")
	tool.Complete()

	c.AddBlock(user)
	c.AddBlock(asst)
	c.AddBlock(tool)

	if c.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", c.Len())
	}

	// Measure: each block is 1 line, spacing=1 → total 3 + 2 = 5
	size := c.Measure(component.Bounded(80, 200))
	if size.H != 5 {
		t.Errorf("Measure H = %d, want 5 (3 blocks + 2 spacing)", size.H)
	}

	// SetBounds + Paint without panic
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	c.Paint(buf)

	// Verify content painted: user message 'H' at (0,0), assistant 'H' at (0,2)
	if r := buf.GetCell(0, 0).Rune; r != 'H' {
		t.Errorf("cell(0,0) = %q, want 'H' (user message)", r)
	}
	if r := buf.GetCell(0, 2).Rune; r != 'H' {
		t.Errorf("cell(0,2) = %q, want 'H' (assistant text)", r)
	}
	// Tool call at row 4 — should have some content (tool name)
	if r := buf.GetCell(0, 4).Rune; r == 0 || r == ' ' {
		t.Errorf("cell(0,4) appears empty, expected tool call content")
	}
}

func TestContainerSpacingChange(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("u1", "A"))
	c.AddBlock(NewUserMessageBlock("u2", "B"))
	c.AddBlock(NewUserMessageBlock("u3", "C"))

	// Default spacing=1: 3 blocks × 1 row + 2 gaps × 1 = 5
	size1 := c.Measure(component.Bounded(80, 200))
	if size1.H != 5 {
		t.Fatalf("spacing=1: H = %d, want 5", size1.H)
	}

	// Change spacing to 3: 3 + 2×3 = 9
	c.SetSpacing(3)
	size3 := c.Measure(component.Bounded(80, 200))
	if size3.H != 9 {
		t.Errorf("spacing=3: H = %d, want 9", size3.H)
	}

	// Change spacing to 0: 3 + 0 = 3
	c.SetSpacing(0)
	size0 := c.Measure(component.Bounded(80, 200))
	if size0.H != 3 {
		t.Errorf("spacing=0: H = %d, want 3", size0.H)
	}
}

func TestContainerRemoveMiddle(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("u1", "AAA"))
	c.AddBlock(NewUserMessageBlock("u2", "BBB"))
	c.AddBlock(NewUserMessageBlock("u3", "CCC"))

	if c.Len() != 3 {
		t.Fatalf("Len before = %d, want 3", c.Len())
	}

	ok := c.RemoveBlock("u2")
	if !ok {
		t.Fatal("RemoveBlock(u2) returned false")
	}
	if c.Len() != 2 {
		t.Fatalf("Len after = %d, want 2", c.Len())
	}

	// Verify remaining order: u1, u3
	blocks := c.Blocks()
	if blocks[0].ID() != "u1" || blocks[1].ID() != "u3" {
		t.Errorf("after remove: order = %s, %s; want u1, u3", blocks[0].ID(), blocks[1].ID())
	}

	// Layout still correct: 2 blocks + 1 spacing = 3
	size := c.Measure(component.Bounded(80, 200))
	if size.H != 3 {
		t.Errorf("after remove Measure H = %d, want 3", size.H)
	}

	// Remove non-existent should return false
	if c.RemoveBlock("nonexistent") {
		t.Error("RemoveBlock(nonexistent) should return false")
	}
}

func TestContainerMeasureClampWidth(t *testing.T) {
	c := NewBlockContainer()

	// Block 1: short content (width ~5)
	c.AddBlock(NewUserMessageBlock("u1", "Hi"))
	// Block 2: long content (will be wider)
	long := NewUserMessageBlock("u2", "This is a much longer message that exceeds short width")
	c.AddBlock(long)

	// Measure unbounded → width = max of block widths
	size := c.Measure(component.Unbounded())
	if size.W <= 5 {
		t.Errorf("unbounded W = %d, should be > 5 (longest block)", size.W)
	}

	// Measure with narrow constraint → width should clamp
	narrow := c.Measure(component.Bounded(10, 200))
	if narrow.W > 10 {
		t.Errorf("clamped W = %d, should be <= 10", narrow.W)
	}
}

func TestContainerDirtyPropagation(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("u1", "hello"))

	// Container should be dirty after AddBlock
	if !c.IsDirty() {
		t.Error("container should be dirty after AddBlock")
	}

	// Clear all dirty
	c.ClearDirty()
	if c.IsDirty() {
		t.Error("container should be clean after ClearDirty")
	}

	// Mark a child dirty via AppendDelta
	asst := NewAssistantTextBlock("a1")
	c.AddBlock(asst)
	c.ClearDirty()

	asst.AppendDelta("new content")
	if !asst.IsDirty() {
		t.Error("assistant block should be dirty after AppendDelta")
	}
	if !c.IsDirty() {
		t.Error("container should be dirty when child is dirty")
	}

	// Clear propagates to children
	c.ClearDirty()
	if asst.IsDirty() {
		t.Error("child should be clean after container ClearDirty")
	}
	if c.IsDirty() {
		t.Error("container should be clean after ClearDirty")
	}
}

func TestContainerEmptyLayout(t *testing.T) {
	c := NewBlockContainer()

	size := c.Measure(component.Bounded(80, 200))
	// Empty container: min 1×1
	if size.W < 1 || size.H < 1 {
		t.Errorf("empty container Measure = %v, want min 1×1", size)
	}

	if c.LastBlock() != nil {
		t.Error("LastBlock() on empty should be nil")
	}

	// Paint on empty should not panic
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	c.Paint(buf)
}

func TestContainerLastBlock(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewUserMessageBlock("u1", "first")
	b2 := NewAssistantTextBlock("a1")
	c.AddBlock(b1)
	c.AddBlock(b2)

	last := c.LastBlock()
	if last == nil || last.ID() != "a1" {
		t.Errorf("LastBlock() = %v, want a1", last)
	}
}

func TestContainerChildrenAsComponents(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("u1", "A"))
	c.AddBlock(NewAssistantTextBlock("a1"))

	children := c.Children()
	if len(children) != 2 {
		t.Fatalf("Children() returned %d, want 2", len(children))
	}
}
