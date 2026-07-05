package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// fixedSize is a test component with a fixed size.
type fixedSize struct {
	BaseComponent
	w, h int
}

func (f *fixedSize) Measure(cs Constraints) Size { return Size{W: f.w, H: f.h} }
func (f *fixedSize) Paint(buf *buffer.Buffer)    {}
func (f *fixedSize) SetBounds(r Rect)            { f.bounds = r }
func (f *fixedSize) Children() []Component       { return nil }

// --- Construction ---

func TestViewport_New(t *testing.T) {
	child := &fixedSize{w: 100, h: 50}
	vp := NewViewport(child)
	if vp == nil {
		t.Fatal("expected non-nil viewport")
	}
	if vp.Content() != child {
		t.Error("expected content to be the child")
	}
}

func TestViewport_NewNil(t *testing.T) {
	vp := NewViewport(nil)
	if vp.Content() != nil {
		t.Error("expected nil content")
	}
}

// --- Scroll offsets ---

func TestViewport_ScrollDown(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(5)
	if vp.OffsetY() != 5 {
		t.Errorf("expected offsetY 5, got %d", vp.OffsetY())
	}
}

func TestViewport_ScrollDown_Clamp(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(1000)
	maxOff := vp.MaxOffsetY()
	if vp.OffsetY() != maxOff {
		t.Errorf("expected offsetY %d, got %d", maxOff, vp.OffsetY())
	}
}

func TestViewport_ScrollUp(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(10)
	vp.ScrollUp(3)
	if vp.OffsetY() != 7 {
		t.Errorf("expected offsetY 7, got %d", vp.OffsetY())
	}
}

func TestViewport_ScrollUp_Zero(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollUp(5) // can't go below 0
	if vp.OffsetY() != 0 {
		t.Errorf("expected offsetY 0, got %d", vp.OffsetY())
	}
}

func TestViewport_ScrollRight(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollRight(5)
	if vp.OffsetX() != 5 {
		t.Errorf("expected offsetX 5, got %d", vp.OffsetX())
	}
}

func TestViewport_ScrollLeft(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollRight(10)
	vp.ScrollLeft(3)
	if vp.OffsetX() != 7 {
		t.Errorf("expected offsetX 7, got %d", vp.OffsetX())
	}
}

func TestViewport_ScrollToTop(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(20)
	vp.ScrollToTop()
	if vp.OffsetY() != 0 {
		t.Error("expected offsetY 0 after ScrollToTop")
	}
}

func TestViewport_ScrollToBottom(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollToBottom()
	if vp.OffsetY() != vp.MaxOffsetY() {
		t.Errorf("expected offsetY %d, got %d", vp.MaxOffsetY(), vp.OffsetY())
	}
}

func TestViewport_ScrollToLeft(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollRight(10)
	vp.ScrollToLeft()
	if vp.OffsetX() != 0 {
		t.Error("expected offsetX 0 after ScrollToLeft")
	}
}

func TestViewport_ScrollToRight(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollToRight()
	if vp.OffsetX() != vp.MaxOffsetX() {
		t.Errorf("expected offsetX %d, got %d", vp.MaxOffsetX(), vp.OffsetX())
	}
}

func TestViewport_ScrollToY(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollToY(15)
	if vp.OffsetY() != 15 {
		t.Errorf("expected offsetY 15, got %d", vp.OffsetY())
	}

	// Clamp negative
	vp.ScrollToY(-5)
	if vp.OffsetY() != 0 {
		t.Error("expected clamped to 0")
	}

	// Clamp overflow
	vp.ScrollToY(10000)
	if vp.OffsetY() != vp.MaxOffsetY() {
		t.Error("expected clamped to MaxOffsetY")
	}
}

func TestViewport_ScrollToX(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollToX(15)
	if vp.OffsetX() != 15 {
		t.Errorf("expected offsetX 15, got %d", vp.OffsetX())
	}
}

// --- Content size ---

func TestViewport_ContentSize(t *testing.T) {
	child := &fixedSize{w: 50, h: 30}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	if vp.ContentWidth() != 50 {
		t.Errorf("expected contentW 50, got %d", vp.ContentWidth())
	}
	if vp.ContentHeight() != 30 {
		t.Errorf("expected contentH 30, got %d", vp.ContentHeight())
	}
}

func TestViewport_MaxOffsets(t *testing.T) {
	child := &fixedSize{w: 50, h: 30}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	// Content overflows both dimensions → scrollbars take 1 col + 1 row each
	// MaxOffsetX = contentW - (bounds.W - vBarWidth) = 50 - 19 = 31
	// MaxOffsetY = contentH - (bounds.H - hBarHeight) = 30 - 9 = 21
	if vp.MaxOffsetX() != 31 {
		t.Errorf("expected MaxOffsetX 31, got %d", vp.MaxOffsetX())
	}
	if vp.MaxOffsetY() != 21 {
		t.Errorf("expected MaxOffsetY 21, got %d", vp.MaxOffsetY())
	}
}

// --- HandleKey ---

func TestViewport_HandleKey_ArrowDown(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	consumed := vp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !consumed {
		t.Error("expected Down to be consumed")
	}
	if vp.OffsetY() != 1 {
		t.Errorf("expected offsetY 1, got %d", vp.OffsetY())
	}
}

func TestViewport_HandleKey_ArrowUp(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(5)
	vp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if vp.OffsetY() != 4 {
		t.Errorf("expected offsetY 4, got %d", vp.OffsetY())
	}
}

func TestViewport_HandleKey_ArrowRight(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	consumed := vp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !consumed {
		t.Error("expected Right to be consumed")
	}
	if vp.OffsetX() != 1 {
		t.Errorf("expected offsetX 1, got %d", vp.OffsetX())
	}
}

func TestViewport_HandleKey_ArrowLeft(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollRight(5)
	vp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if vp.OffsetX() != 4 {
		t.Errorf("expected offsetX 4, got %d", vp.OffsetX())
	}
}

func TestViewport_HandleKey_PageDown(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	if vp.OffsetY() != 10 {
		t.Errorf("expected offsetY 10, got %d", vp.OffsetY())
	}
}

func TestViewport_HandleKey_PageUp(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(20)
	vp.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if vp.OffsetY() != 10 {
		t.Errorf("expected offsetY 10, got %d", vp.OffsetY())
	}
}

func TestViewport_HandleKey_Home(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(20)
	vp.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if vp.OffsetY() != 0 {
		t.Error("expected offsetY 0 after Home")
	}
}

func TestViewport_HandleKey_End(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if vp.OffsetY() != vp.MaxOffsetY() {
		t.Error("expected offsetY at MaxOffsetY after End")
	}
}

func TestViewport_HandleKey_Vim_j(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	consumed := vp.HandleKey(&term.KeyEvent{Rune: 'j'})
	if !consumed || vp.OffsetY() != 1 {
		t.Errorf("expected j to scroll down, consumed=%v offsetY=%d", consumed, vp.OffsetY())
	}
}

func TestViewport_HandleKey_Vim_k(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(5)
	consumed := vp.HandleKey(&term.KeyEvent{Rune: 'k'})
	if !consumed || vp.OffsetY() != 4 {
		t.Errorf("expected k to scroll up, consumed=%v offsetY=%d", consumed, vp.OffsetY())
	}
}

func TestViewport_HandleKey_Vim_g(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.ScrollDown(20)
	vp.HandleKey(&term.KeyEvent{Rune: 'g'})
	if vp.OffsetY() != 0 {
		t.Error("expected g to scroll to top")
	}
}

func TestViewport_HandleKey_Vim_G(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	vp.HandleKey(&term.KeyEvent{Rune: 'G'})
	if vp.OffsetY() != vp.MaxOffsetY() {
		t.Error("expected G to scroll to bottom")
	}
}

func TestViewport_HandleKey_Vim_h_l(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.HandleKey(&term.KeyEvent{Rune: 'l'})
	if vp.OffsetX() != 1 {
		t.Errorf("expected l to scroll right, got %d", vp.OffsetX())
	}
	vp.HandleKey(&term.KeyEvent{Rune: 'h'})
	if vp.OffsetX() != 0 {
		t.Errorf("expected h to scroll left, got %d", vp.OffsetX())
	}
}

func TestViewport_HandleKey_Vim_0_dollar(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})

	vp.ScrollRight(10)
	vp.HandleKey(&term.KeyEvent{Rune: '0'})
	if vp.OffsetX() != 0 {
		t.Error("expected 0 to scroll to left")
	}

	vp.HandleKey(&term.KeyEvent{Rune: '$'})
	if vp.OffsetX() != vp.MaxOffsetX() {
		t.Error("expected $ to scroll to right")
	}
}

func TestViewport_HandleKey_Unhandled(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	consumed := vp.HandleKey(&term.KeyEvent{Rune: 'x', Modifiers: term.ModCtrl})
	if consumed {
		t.Error("expected Ctrl+X to not be consumed")
	}
}

// --- Paint ---

func TestViewport_Paint_NoOverflow(t *testing.T) {
	child := &fixedSize{w: 10, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	buf.Fill(buffer.BlankCell)
	vp.Paint(buf) // should not panic
}

func TestViewport_Paint_Overflow(t *testing.T) {
	child := &fixedSize{w: 50, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollDown(5)
	vp.ScrollRight(10)

	buf := buffer.NewBuffer(20, 10)
	buf.Fill(buffer.BlankCell)
	vp.Paint(buf) // should not panic
}

func TestViewport_Paint_Nil(t *testing.T) {
	vp := NewViewport(nil)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf) // should not panic
}

// --- Children ---

func TestViewport_Children(t *testing.T) {
	child := &fixedSize{w: 20, h: 10}
	vp := NewViewport(child)
	children := vp.Children()
	if len(children) != 1 || children[0] != child {
		t.Error("expected 1 child matching the input")
	}
}

func TestViewport_Children_Nil(t *testing.T) {
	vp := NewViewport(nil)
	if vp.Children() != nil {
		t.Error("expected nil children for nil content")
	}
}

// --- SetContent ---

func TestViewport_SetContent(t *testing.T) {
	child1 := &fixedSize{w: 10, h: 10}
	child2 := &fixedSize{w: 20, h: 20}
	vp := NewViewport(child1)

	vp.ScrollDown(5)
	vp.SetContent(child2)

	if vp.Content() != child2 {
		t.Error("expected new content")
	}
	if vp.OffsetY() != 0 {
		t.Error("expected offsetY reset to 0")
	}
}

// --- Scrollbar visibility ---

func TestViewport_VScrollbarColumn(t *testing.T) {
	child := &fixedSize{w: 20, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 5, Y: 0, W: 20, H: 10})

	col := vp.VScrollbarColumn()
	if col < 0 {
		t.Error("expected scrollbar column >= 0 when content overflows")
	}
}

func TestViewport_VScrollbarColumn_NoOverflow(t *testing.T) {
	child := &fixedSize{w: 10, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	col := vp.VScrollbarColumn()
	if col >= 0 {
		t.Error("expected scrollbar column -1 when no overflow")
	}
}

func TestViewport_HScrollbarRow(t *testing.T) {
	child := &fixedSize{w: 100, h: 20}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 5, W: 20, H: 10})

	row := vp.HScrollbarRow()
	if row < 0 {
		t.Error("expected scrollbar row >= 0 when content overflows")
	}
}

// --- Style ---

func TestViewport_SetVBarStyle(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 10, h: 10})
	style := ScrollBarStyle{Visible: true, ThumbChar: '#', TrackChar: '-'}
	vp.SetVBarStyle(style)
	// Just verify no panic
}

func TestViewport_SetHBarStyle(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 10, h: 10})
	style := ScrollBarStyle{Visible: true, ThumbChar: '#', TrackChar: '-'}
	vp.SetHBarStyle(style)
}

// --- Concurrency ---

func TestViewport_Concurrent(t *testing.T) {
	child := &fixedSize{w: 100, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vp.ScrollDown(1)
			vp.ScrollUp(1)
			_ = vp.OffsetX()
			_ = vp.OffsetY()
			_ = vp.MaxOffsetX()
			_ = vp.MaxOffsetY()
		}()
	}
	wg.Wait()
}

func TestViewport_ConcurrentPaint(t *testing.T) {
	child := &fixedSize{w: 50, h: 50}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vp.ScrollDown(1)
			buf := buffer.NewBuffer(20, 10)
			vp.Paint(buf)
		}()
	}
	wg.Wait()
}

// --- Measure ---

func TestViewport_Measure(t *testing.T) {
	child := &fixedSize{w: 100, h: 100}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	s := vp.Measure(Unbounded())
	if s.W == 0 || s.H == 0 {
		t.Error("expected non-zero size")
	}
}

func TestViewport_Measure_NilContent(t *testing.T) {
	vp := NewViewport(nil)
	s := vp.Measure(Bounded(20, 10))
	if s.W != 20 || s.H != 10 {
		t.Errorf("expected 20x10, got %dx%d", s.W, s.H)
	}
}
