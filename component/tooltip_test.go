package component

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewTooltip(t *testing.T) {
	tt := NewTooltip("hello")
	if tt.Text() != "hello" {
		t.Errorf("expected text 'hello', got '%s'", tt.Text())
	}
	if tt.Placement() != TooltipTop {
		t.Errorf("expected placement top")
	}
	if tt.IsVisible() {
		t.Errorf("expected hidden by default")
	}
}

func TestTooltip_SetText(t *testing.T) {
	tt := NewTooltip("original")
	tt.SetText("updated")
	if tt.Text() != "updated" {
		t.Errorf("expected 'updated', got '%s'", tt.Text())
	}
}

func TestTooltip_SetPlacement(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetPlacement(TooltipBottom)
	if tt.Placement() != TooltipBottom {
		t.Errorf("expected bottom")
	}
	tt.SetPlacement(TooltipLeft)
	if tt.Placement() != TooltipLeft {
		t.Errorf("expected left")
	}
	tt.SetPlacement(TooltipRight)
	if tt.Placement() != TooltipRight {
		t.Errorf("expected right")
	}
}

func TestTooltip_SetAnchor(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetAnchor(10, 20)
	x, y := tt.Anchor()
	if x != 10 || y != 20 {
		t.Errorf("expected (10,20), got (%d,%d)", x, y)
	}
}

func TestTooltip_ShowHide(t *testing.T) {
	tt := NewTooltip("x")
	tt.Show()
	if !tt.IsVisible() {
		t.Error("expected visible after Show()")
	}
	tt.Hide()
	if tt.IsVisible() {
		t.Error("expected hidden after Hide()")
	}
}

func TestTooltip_TickHoverShow(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowDelay(100 * time.Millisecond)

	// Hover for 50ms — not enough.
	changed := tt.Tick(50*time.Millisecond, true)
	if changed {
		t.Error("should not change before showDelay")
	}
	if tt.IsVisible() {
		t.Error("should not be visible yet")
	}

	// Hover for another 60ms — total 110ms > 100ms delay.
	changed = tt.Tick(60*time.Millisecond, true)
	if !changed {
		t.Error("should change when becoming visible")
	}
	if !tt.IsVisible() {
		t.Error("should be visible after showDelay exceeded")
	}
}

func TestTooltip_TickNotHovering(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowDelay(100 * time.Millisecond)

	// Not hovering — should stay hidden.
	tt.Tick(200*time.Millisecond, false)
	if tt.IsVisible() {
		t.Error("should stay hidden when not hovering")
	}
}

func TestTooltip_TickMouseLeaves(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowDelay(100 * time.Millisecond)
	tt.Show()
	if !tt.IsVisible() {
		t.Error("expected visible")
	}
	// Mouse leaves.
	changed := tt.Tick(0, false)
	if !changed {
		t.Error("should report change when hiding")
	}
	if tt.IsVisible() {
		t.Error("should hide when mouse leaves")
	}
}

func TestTooltip_TickResetTimer(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowDelay(100 * time.Millisecond)

	tt.Tick(80*time.Millisecond, true)
	tt.Tick(0, false) // leave — resets timer
	tt.Tick(80*time.Millisecond, true) // only 80ms since reset
	if tt.IsVisible() {
		t.Error("timer should have reset, not enough time")
	}
	tt.Tick(30*time.Millisecond, true) // total 110ms > 100ms
	if !tt.IsVisible() {
		t.Error("should be visible now")
	}
}

func TestTooltip_Lines(t *testing.T) {
	tt := NewTooltip("line1\nline2\nline3")
	lines := tt.Lines()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("lines mismatch: %v", lines)
	}
}

func TestTooltip_LinesCopy(t *testing.T) {
	tt := NewTooltip("original")
	lines := tt.Lines()
	lines[0] = "modified"
	// Original should be unchanged.
	if tt.Lines()[0] != "original" {
		t.Errorf("Lines() should return a copy")
	}
}

func TestTooltip_SetMaxWidth(t *testing.T) {
	tt := NewTooltip("the quick brown fox jumps over the lazy dog")
	tt.SetMaxWidth(15)
	lines := tt.Lines()
	for _, line := range lines {
		if len(line) > 15 {
			t.Errorf("line exceeds max width: '%s' (%d)", line, len(line))
		}
	}
}

func TestTooltip_Measure(t *testing.T) {
	tt := NewTooltip("hello world")
	tt.SetShowBorder(false)
	sz := tt.Measure(Unbounded())
	if sz.W < 11 {
		t.Errorf("expected width >= 11, got %d", sz.W)
	}
	if sz.H != 1 {
		t.Errorf("expected height 1, got %d", sz.H)
	}
}

func TestTooltip_MeasureWithBorder(t *testing.T) {
	tt := NewTooltip("hello")
	tt.SetShowBorder(true)
	sz := tt.Measure(Unbounded())
	if sz.W < 7 {
		t.Errorf("expected width >= 7 (text + 2 borders), got %d", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("expected height 3 (1 text + 2 borders), got %d", sz.H)
	}
}

func TestTooltip_MeasureMultiline(t *testing.T) {
	tt := NewTooltip("line1\nline2\nline3")
	tt.SetShowBorder(false)
	sz := tt.Measure(Unbounded())
	if sz.H != 3 {
		t.Errorf("expected height 3, got %d", sz.H)
	}
}

func TestTooltip_PaintWithBorder(t *testing.T) {
	tt := NewTooltip("Hi")
	tt.SetShowBorder(true)
	sz := tt.Measure(Unbounded())
	tt.SetBounds(Rect{X: 0, Y: 0, W: sz.W, H: sz.H})

	buf := newTestBuffer(sz.W, sz.H)
	tt.Paint(buf)

	// Check top-left corner.
	c := buf.GetCell(0, 0)
	if c.Rune != '┌' {
		t.Errorf("expected ┌ at (0,0), got %q", string(c.Rune))
	}
	// Check text content.
	c = buf.GetCell(1, 1)
	if c.Rune != 'H' {
		t.Errorf("expected 'H' at (1,1), got %q", string(c.Rune))
	}
	c = buf.GetCell(2, 1)
	if c.Rune != 'i' {
		t.Errorf("expected 'i' at (2,1), got %q", string(c.Rune))
	}
}

func TestTooltip_PaintPlainText(t *testing.T) {
	tt := NewTooltip("Hello")
	tt.SetShowBorder(false)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := newTestBuffer(10, 1)
	tt.Paint(buf)

	c := buf.GetCell(0, 0)
	if c.Rune != 'H' {
		t.Errorf("expected 'H' at (0,0), got %q", string(c.Rune))
	}
}

func TestTooltip_PaintEmpty(t *testing.T) {
	tt := NewTooltip("")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := newTestBuffer(10, 3)
	tt.Paint(buf) // should not panic
}

func TestTooltip_PaintZeroBounds(t *testing.T) {
	tt := NewTooltip("Hello")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := newTestBuffer(10, 10)
	tt.Paint(buf) // should not panic
}

func TestTooltip_ComputePositionTop(t *testing.T) {
	tt := NewTooltip("tip")
	tt.SetPlacement(TooltipTop)
	tt.SetAnchor(50, 20)
	tt.SetSmartPosition(false)

	sz := tt.Measure(Unbounded())
	x, y := tt.ComputePosition(100, 50)
	// Should be above anchor.
	if y >= 20 {
		t.Errorf("expected y < 20 for top placement, got %d", y)
	}
	_ = x
	_ = sz
}

func TestTooltip_ComputePositionBottom(t *testing.T) {
	tt := NewTooltip("tip")
	tt.SetPlacement(TooltipBottom)
	tt.SetAnchor(50, 20)
	tt.SetSmartPosition(false)

	x, y := tt.ComputePosition(100, 50)
	if y <= 20 {
		t.Errorf("expected y > 20 for bottom placement, got %d", y)
	}
	_ = x
}

func TestTooltip_ComputePositionSmartFlip(t *testing.T) {
	tt := NewTooltip("tip")
	tt.SetPlacement(TooltipTop)
	tt.SetAnchor(50, 0) // Very top of screen — no room above.
	tt.SetSmartPosition(true)

	_, y := tt.ComputePosition(100, 50)
	if y < 0 {
		t.Errorf("smart flip should prevent negative y, got %d", y)
	}
}

func TestTooltip_ComputePositionClamp(t *testing.T) {
	tt := NewTooltip("a long tooltip text here")
	tt.SetPlacement(TooltipRight)
	tt.SetAnchor(95, 25)
	tt.SetSmartPosition(false)

	x, _ := tt.ComputePosition(100, 50)
	if x+10 > 100 {
		t.Errorf("x should be clamped within screen, x=%d", x)
	}
}

func TestTooltip_SetShowDelay(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowDelay(1 * time.Second)
	// Just verify it doesn't panic.
}

func TestTooltip_SetAutoHide(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetAutoHide(5 * time.Second)
}

func TestTooltip_SetTextStyle(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetTextStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)})
}

func TestTooltip_SetBorderStyle(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetBorderStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)})
}

func TestTooltip_SetShowBorder(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetShowBorder(false)
	tt.SetShowBorder(true)
}

func TestTooltip_SetSmartPosition(t *testing.T) {
	tt := NewTooltip("x")
	tt.SetSmartPosition(false)
}

func TestTooltip_Concurrent(t *testing.T) {
	tt := NewTooltip("test tooltip")
	tt.SetShowDelay(50 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tt.Tick(time.Millisecond, n%2 == 0)
				_ = tt.IsVisible()
			}
		}(i)
	}
	wg.Wait()
}

func TestTooltip_ConcurrentPaint(t *testing.T) {
	tt := NewTooltip("paint test")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := newTestBuffer(20, 5)
				tt.Paint(buf)
			}
		}()
	}
	wg.Wait()
}
