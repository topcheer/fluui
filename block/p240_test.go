package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/termcompat"
)

func TestContainer_ScrollToBottomAlreadyAtBottom_P240(t *testing.T) {
	c := NewBlockContainer()
	c.SetAutoScroll(true)
	c.mu.Lock()
	c.scrollOffset = 0
	c.mu.Unlock()
	if c.ScrollToBottom() {
		t.Error("should return false when already at bottom")
	}
}

func TestContainer_ScrollToBottomAfterScrollUp_P240(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("id1", "hello"))
	c.SetAutoScroll(false)
	// Manually set scrollOffset by scrolling up
	c.mu.Lock()
	c.scrollOffset = 5
	c.mu.Unlock()
	if !c.ScrollToBottom() {
		t.Error("should return true after scroll up")
	}
}

func TestContainer_MeasureWithMaxHeight_P240(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("id1", "hello"))
	c.AddBlock(NewUserMessageBlock("id2", "world"))
	s := c.Measure(component.Constraints{MaxWidth: 80, MaxHeight: 1})
	if s.H > 1 {
		t.Errorf("height should be clamped to 1, got %d", s.H)
	}
}

func TestImageBlock_WithProtocol_P240(t *testing.T) {
	b := NewImageBlock("img1", "test.png", []byte("data"))
	b.mu.Lock()
	b.protocol = termcompat.ImageKitty
	b.mu.Unlock()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestImageBlock_WithDisplayH_P240(t *testing.T) {
	b := NewImageBlock("img2", "test.png", []byte("data"))
	b.SetDisplaySize(20, 10)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	b.Paint(buf)
}

func TestImageBlock_SequenceCached_P240(t *testing.T) {
	b := NewImageBlock("img3", "test.png", []byte("data"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	b.Paint(buf) // second call hits seqCached path
}
