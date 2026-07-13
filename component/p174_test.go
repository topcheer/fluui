package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Digits tests ===

func TestDigits_New(t *testing.T) {
	d := NewDigits("42")
	if d.Value() != "42" {
		t.Errorf("expected '42', got %q", d.Value())
	}
}

func TestDigits_SetValue(t *testing.T) {
	d := NewDigits("0")
	d.SetValue("99")
	if d.Value() != "99" {
		t.Errorf("expected '99', got %q", d.Value())
	}
}

func TestDigits_SetValueInt(t *testing.T) {
	d := NewDigits("")
	d.SetValueInt(42)
	if d.Value() != "42" {
		t.Errorf("expected '42', got %q", d.Value())
	}
	d.SetValueInt(-7)
	if d.Value() != "-7" {
		t.Errorf("expected '-7', got %q", d.Value())
	}
}

func TestDigits_SetValueFormatted(t *testing.T) {
	d := NewDigits("")
	d.SetValueFormatted("12:30:45")
	if d.Value() != "12:30:45" {
		t.Errorf("expected '12:30:45', got %q", d.Value())
	}
	// Invalid chars filtered
	d.SetValueFormatted("abc123!@#")
	if d.Value() != "123" {
		t.Errorf("expected '123', got %q", d.Value())
	}
}

func TestDigits_Measure(t *testing.T) {
	d := NewDigits("123")
	s := d.Measure(Constraints{})
	// 3 chars × 3 width + 2 gaps × 1 = 11
	if s.W != 11 || s.H != 5 {
		t.Errorf("expected 11x5, got %dx%d", s.W, s.H)
	}
}

func TestDigits_MeasureEmpty(t *testing.T) {
	d := NewDigits("")
	s := d.Measure(Constraints{})
	if s.W != 0 || s.H != 5 {
		t.Errorf("expected 0x5, got %dx%d", s.W, s.H)
	}
}

func TestDigits_Paint(t *testing.T) {
	d := NewDigits("8")
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	d.Paint(buf)
	// Check that some cells are drawn
	drawn := 0
	for y := 0; y < 5; y++ {
		for x := 0; x < 3; x++ {
			if buf.GetCell(x, y).Rune == '█' {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Error("expected drawn cells for digit '8'")
	}
}

func TestDigits_PaintAllDigits(t *testing.T) {
	for i := 0; i <= 9; i++ {
		d := NewDigits(string(rune('0' + i)))
		d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
		buf := buffer.NewBuffer(5, 5)
		d.Paint(buf)
		// At least one cell should be drawn for each digit
		drawn := 0
		for y := 0; y < 5; y++ {
			for x := 0; x < 3; x++ {
				if buf.GetCell(x, y).Rune == '█' {
					drawn++
				}
			}
		}
		if drawn == 0 {
			t.Errorf("digit %d: expected drawn cells", i)
		}
	}
}

func TestDigits_PaintColon(t *testing.T) {
	d := NewDigits(":")
	d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	d.Paint(buf) // should not panic
}

func TestDigits_PaintMinus(t *testing.T) {
	d := NewDigits("-")
	d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	d.Paint(buf) // should not panic
}

func TestDigits_PaintShowDim(t *testing.T) {
	d := NewDigits("1")
	d.SetShowDim(true)
	d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	d.Paint(buf)
	// Should have dim cells
	dimCount := 0
	for y := 0; y < 5; y++ {
		for x := 0; x < 3; x++ {
			if buf.GetCell(x, y).Rune == '·' {
				dimCount++
			}
		}
	}
	if dimCount == 0 {
		t.Error("expected dim cells when showDim=true")
	}
}

func TestDigits_PaintNilBuffer(t *testing.T) {
	d := NewDigits("1")
	d.Paint(nil) // should not panic
}

func TestDigits_PaintEmptyValue(t *testing.T) {
	d := NewDigits("")
	d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	d.Paint(buf) // should not panic
}

func TestDigits_SetDigitGap(t *testing.T) {
	d := NewDigits("12")
	d.SetDigitGap(2)
	s := d.Measure(Constraints{})
	// 2 chars × 3 + 1 gap × 2 = 8
	if s.W != 8 {
		t.Errorf("expected width 8, got %d", s.W)
	}
}

func TestDigits_SetDigitGapNegative(t *testing.T) {
	d := NewDigits("12")
	d.SetDigitGap(-5)
	// Should clamp to 0
}

func TestDigits_Children(t *testing.T) {
	d := NewDigits("1")
	if d.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestDigits_Concurrent(t *testing.T) {
	d := NewDigits("0")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			d.SetValueInt(i)
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		buf := buffer.NewBuffer(10, 5)
		d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
		d.Paint(buf)
	}
	<-done
}

// === LoadingIndicator tests ===

func TestLoadingIndicator_New(t *testing.T) {
	l := NewLoadingIndicator("Loading data...")
	if l.Text() != "Loading data..." {
		t.Errorf("expected 'Loading data...', got %q", l.Text())
	}
}

func TestLoadingIndicator_SetText(t *testing.T) {
	l := NewLoadingIndicator("")
	l.SetText("Fetching...")
	if l.Text() != "Fetching..." {
		t.Errorf("expected 'Fetching...', got %q", l.Text())
	}
}

func TestLoadingIndicator_StartStop(t *testing.T) {
	l := NewLoadingIndicator("test")
	if l.IsRunning() {
		t.Error("should not be running initially")
	}
	l.Start()
	if !l.IsRunning() {
		t.Error("should be running after Start")
	}
	l.Stop()
	if l.IsRunning() {
		t.Error("should not be running after Stop")
	}
	// Stop is idempotent
	l.Stop()
}

func TestLoadingIndicator_AdvanceFrame(t *testing.T) {
	l := NewLoadingIndicator("test")
	orig := l.Frame()
	l.AdvanceFrame()
	if l.Frame() != orig+1 {
		t.Errorf("expected frame %d, got %d", orig+1, l.Frame())
	}
}

func TestLoadingIndicator_FrameWrap(t *testing.T) {
	l := NewLoadingIndicator("test")
	for i := 0; i < 100; i++ {
		l.AdvanceFrame()
	}
	if l.Frame() < 0 || l.Frame() >= 20 {
		t.Errorf("frame should wrap, got %d", l.Frame())
	}
}

func TestLoadingIndicator_Measure(t *testing.T) {
	l := NewLoadingIndicator("test")
	s := l.Measure(Constraints{})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", s.W, s.H)
	}
}

func TestLoadingIndicator_Paint(t *testing.T) {
	l := NewLoadingIndicator("Loading data...")
	l.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	l.Paint(buf)
	// Should have text
	if buf.GetCell(0, 0).Rune != 'L' {
		t.Error("expected 'L' at start of text")
	}
}

func TestLoadingIndicator_PaintNilBuffer(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Paint(nil) // should not panic
}

func TestLoadingIndicator_PaintNarrow(t *testing.T) {
	l := NewLoadingIndicator("Loading data...")
	l.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	l.Paint(buf) // should not panic
}

func TestLoadingIndicator_PaintEmptyText(t *testing.T) {
	l := NewLoadingIndicator("")
	l.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	l.Paint(buf)
	// Should show default "Loading..."
	if buf.GetCell(0, 0).Rune != 'L' {
		t.Error("expected default 'Loading...' text")
	}
}

func TestLoadingIndicator_Children(t *testing.T) {
	l := NewLoadingIndicator("test")
	if l.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestLoadingIndicator_HandleKey(t *testing.T) {
	l := NewLoadingIndicator("test")
	if l.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected false for HandleKey")
	}
}

func TestLoadingIndicator_Concurrent(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	defer l.Stop()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			l.AdvanceFrame()
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		buf := buffer.NewBuffer(30, 3)
		l.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
		l.Paint(buf)
	}
	<-done
}

// === TabbedContent tests ===

func TestTabbedContent_New(t *testing.T) {
	tc := NewTabbedContent()
	if tc.TabCount() != 0 {
		t.Errorf("expected 0 tabs, got %d", tc.TabCount())
	}
}

func TestTabbedContent_AddTab(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	if tc.TabCount() != 2 {
		t.Errorf("expected 2 tabs, got %d", tc.TabCount())
	}
}

func TestTabbedContent_SwitchTo(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SwitchTo("b")
	if tc.ActiveTab() != "b" {
		t.Errorf("expected 'b', got %q", tc.ActiveTab())
	}
}

func TestTabbedContent_NextTab(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.AddTab("c", "Tab C", NewParagraph("content C"))
	tc.SwitchTo("a")
	tc.NextTab()
	if tc.ActiveTab() != "b" {
		t.Errorf("expected 'b', got %q", tc.ActiveTab())
	}
}

func TestTabbedContent_PrevTab(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SwitchTo("b")
	tc.PrevTab()
	if tc.ActiveTab() != "a" {
		t.Errorf("expected 'a', got %q", tc.ActiveTab())
	}
}

func TestTabbedContent_RemoveTab(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.RemoveTab("a")
	if tc.TabCount() != 1 {
		t.Errorf("expected 1 tab, got %d", tc.TabCount())
	}
}

func TestTabbedContent_Paint(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	tc.Paint(buf) // should not panic
}

func TestTabbedContent_PaintNilBuffer(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.Paint(nil) // should not panic
}

func TestTabbedContent_Measure(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	s := tc.Measure(Constraints{MaxWidth: 40, MaxHeight: 20})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", s.W, s.H)
	}
}

func TestTabbedContent_HandleKey_Tab(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SwitchTo("a")
	tc.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if tc.ActiveTab() != "b" {
		t.Errorf("expected 'b' after Tab, got %q", tc.ActiveTab())
	}
}

func TestTabbedContent_HandleKey_Nil(t *testing.T) {
	tc := NewTabbedContent()
	if tc.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestTabbedContent_HandleMouse(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// Click on second tab area
	tc.HandleMouse(10, 0, 0) // should not panic
}

func TestTabbedContent_Children(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	children := tc.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children (tabs+switcher), got %d", len(children))
	}
}

func TestTabbedContent_SetStyle(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.SetStyle(DefaultTabBarStyle()) // should not panic
}

func TestTabbedContent_Concurrent(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			tc.NextTab()
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		buf := buffer.NewBuffer(40, 10)
		tc.Paint(buf)
	}
	<-done
}
