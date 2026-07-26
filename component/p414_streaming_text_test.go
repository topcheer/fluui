package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP414_NewStreamingText(t *testing.T) {
	st := NewStreamingText()
	if st.ID() == "" { t.Error("ID empty") }
	if st.Speed() != 2 { t.Errorf("Speed = %d", st.Speed()) }
	if !st.ShowCursor() { t.Error("should show cursor") }
	if !st.CursorOn() { t.Error("cursor should be on") }
}

func TestP414_SetText(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hello")
	if st.Text() != "Hello" { t.Errorf("Text = %q", st.Text()) }
	if st.Completed() { t.Error("should not be completed") }
}

func TestP414_AppendText(t *testing.T) {
	st := NewStreamingText()
	st.AppendText("Hello ")
	st.AppendText("World")
	if st.Text() != "Hello World" { t.Errorf("Text = %q", st.Text()) }
}

func TestP414_Tick(t *testing.T) {
	st := NewStreamingText()
	st.SetSpeed(2)
	st.SetText("abcde") // 5 chars
	st.Tick()           // visible=2
	if st.VisibleText() != "ab" { t.Errorf("Visible = %q, want 'ab'", st.VisibleText()) }
	st.Tick() // visible=4
	if st.VisibleText() != "abcd" { t.Errorf("Visible = %q", st.VisibleText()) }
	st.Tick() // visible=5 (capped)
	if !st.Completed() { t.Error("should be completed") }
	if st.VisibleText() != "abcde" { t.Errorf("Visible = %q", st.VisibleText()) }
}

func TestP414_Tick_MultiBlink(t *testing.T) {
	st := NewStreamingText()
	st.SetText("test")
	// Call tick 8 times to exercise blink toggle
	for i := 0; i < 8; i++ {
		st.Tick()
	}
	if !st.Completed() { t.Error("should complete after enough ticks") }
}

func TestP414_Skip(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hello")
	st.Skip()
	if !st.Completed() { t.Error("should be completed") }
	if st.VisibleText() != "Hello" { t.Errorf("Visible = %q", st.VisibleText()) }
}

func TestP414_Reset(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hello")
	st.Skip()
	st.Reset()
	if st.Completed() { t.Error("should not be completed after reset") }
	if st.VisibleText() != "" { t.Errorf("Visible = %q, want empty", st.VisibleText()) }
}

func TestP414_Progress(t *testing.T) {
	st := NewStreamingText()
	st.SetSpeed(2)
	st.SetText("abcde")
	if st.Progress() != 0 { t.Errorf("Progress = %v, want 0", st.Progress()) }
	st.Tick() // visible=2
	if st.Progress() != 0.4 { t.Errorf("Progress = %v, want 0.4", st.Progress()) }
	st.Skip()
	if st.Progress() != 1.0 { t.Errorf("Progress = %v, want 1.0", st.Progress()) }
}

func TestP414_SetSpeed(t *testing.T) {
	st := NewStreamingText()
	st.SetSpeed(5)
	if st.Speed() != 5 { t.Errorf("Speed = %d", st.Speed()) }
	st.SetSpeed(0)
	if st.Speed() != 1 { t.Errorf("Speed = %d, want 1 (clamped)", st.Speed()) }
}

func TestP414_SetShowCursor(t *testing.T) {
	st := NewStreamingText()
	st.SetShowCursor(false)
	if st.ShowCursor() { t.Error("should be false") }
}

func TestP414_Measure(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hello World")
	s := st.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W != 11 { t.Errorf("W = %d, want 11", s.W) }
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP414_Paint(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hello")
	st.Skip() // reveal all
	st.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	st.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'H' { t.Error("should draw 'H'") }
}

func TestP414_Paint_Partial(t *testing.T) {
	st := NewStreamingText()
	st.SetSpeed(2)
	st.SetText("Hello")
	st.Tick() // visible=2 → "He"
	st.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	st.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'H' { t.Error("should draw 'H'") }
	if buf.GetCell(1, 0).Rune != 'e' { t.Error("should draw 'e'") }
	// Cursor at position 2
	if buf.GetCell(2, 0).Rune != '\u2588' { t.Errorf("cursor = %q, want █", string(buf.GetCell(2, 0).Rune)) }
}

func TestP414_Paint_NoCursorWhenComplete(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hi")
	st.Skip()
	st.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	st.Paint(buf)
	// No cursor at position 2 (completed)
	if buf.GetCell(2, 0).Rune == '\u2588' { t.Error("should not show cursor when completed") }
}

func TestP414_Paint_ZeroBounds(t *testing.T) {
	st := NewStreamingText()
	st.SetText("x")
	st.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	st.Paint(buf)
}

func TestP414_Paint_WideChars(t *testing.T) {
	st := NewStreamingText()
	st.SetText("你好")
	st.Skip()
	st.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	st.Paint(buf) // should handle wide chars
}

func TestP414_Paint_NonZeroOffset(t *testing.T) {
	st := NewStreamingText()
	st.SetText("Hi")
	st.Skip()
	st.SetBounds(Rect{X: 5, Y: 3, W: 5, H: 1})
	buf := buffer.NewBuffer(15, 10)
	st.Paint(buf)
	if buf.GetCell(5, 3).Rune != 'H' { t.Error("offset cell should be 'H'") }
}

func TestP414_EstimateDuration(t *testing.T) {
	st := NewStreamingText()
	st.SetSpeed(2)
	st.SetText("abcde") // 5 chars, ceil(5/2)=3 ticks
	d := st.EstimateDuration(50 * time.Millisecond)
	if d != 150*time.Millisecond { t.Errorf("Duration = %v, want 150ms", d) }
}

func TestP414_FormatSpeed(t *testing.T) {
	if s := FormatSpeed(0); s != "instant" { t.Errorf("FormatSpeed(0) = %q", s) }
	if s := FormatSpeed(3); s != "3 chars/tick" { t.Errorf("FormatSpeed(3) = %q", s) }
	if s := FormatSpeed(10); s != "10 chars/tick" { t.Errorf("FormatSpeed(10) = %q", s) }
}

func TestP414_Concurrent(t *testing.T) {
	st := NewStreamingText()
	st.SetText("test")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { st.Tick() }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = st.VisibleText() }
	<-done
}

func TestP414_SatisfiesComponent(t *testing.T) {
	var _ Component = (*StreamingText)(nil)
}


func BenchmarkP414_StreamingText_Paint(b *testing.B) {
	st := NewStreamingText()
	st.SetText("The quick brown fox jumps over the lazy dog")
	st.Skip()
	st.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { st.Paint(buf) }
}
