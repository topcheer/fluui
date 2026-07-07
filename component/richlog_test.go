package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestRichLog_New(t *testing.T) {
	rl := NewRichLog()
	if rl.EntryCount() != 0 {
		t.Errorf("expected 0 entries, got %d", rl.EntryCount())
	}
	if !rl.AutoScroll() {
		t.Error("expected auto-scroll on by default")
	}
	if !rl.ShowLevels() {
		t.Error("expected show levels on by default")
	}
	if !rl.ShowTime() {
		t.Error("expected show time on by default")
	}
	if rl.MaxSize() != 10000 {
		t.Errorf("expected maxSize 10000, got %d", rl.MaxSize())
	}
	if rl.MinLevel() != LogDebug {
		t.Errorf("expected minLevel LogDebug, got %d", rl.MinLevel())
	}
}

func TestRichLog_Write(t *testing.T) {
	rl := NewRichLog()
	rl.Info("hello")
	rl.Warn("warning")
	rl.Error("error")
	rl.Debug("debug")
	rl.Fatal("fatal")

	if rl.EntryCount() != 5 {
		t.Errorf("expected 5 entries, got %d", rl.EntryCount())
	}

	entries := rl.Entries()
	if entries[0].Text != "hello" {
		t.Errorf("expected 'hello', got %q", entries[0].Text)
	}
	if entries[1].Level != LogWarn {
		t.Errorf("expected LogWarn, got %d", entries[1].Level)
	}
	if entries[2].Level != LogError {
		t.Errorf("expected LogError, got %d", entries[2].Level)
	}
	if entries[4].Level != LogFatal {
		t.Errorf("expected LogFatal, got %d", entries[4].Level)
	}
}

func TestRichLog_WriteLine(t *testing.T) {
	rl := NewRichLog()
	rl.WriteLine("test line")
	if rl.EntryCount() != 1 {
		t.Errorf("expected 1 entry, got %d", rl.EntryCount())
	}
	if rl.Entries()[0].Level != LogInfo {
		t.Errorf("expected LogInfo, got %d", rl.Entries()[0].Level)
	}
}

func TestRichLog_Writef(t *testing.T) {
	rl := NewRichLog()
	rl.Infof("value=%d name=%s", 42, "test")
	if rl.Entries()[0].Text != "value=42 name=test" {
		t.Errorf("expected formatted text, got %q", rl.Entries()[0].Text)
	}
}

func TestRichLog_MaxSize(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(3)
	rl.Info("first")
	rl.Info("second")
	rl.Info("third")
	rl.Info("fourth")
	if rl.EntryCount() != 3 {
		t.Errorf("expected 3 after trim, got %d", rl.EntryCount())
	}
	entries := rl.Entries()
	if entries[0].Text != "second" {
		t.Errorf("expected 'second', got %q", entries[0].Text)
	}
}

func TestRichLog_Clear(t *testing.T) {
	rl := NewRichLog()
	rl.Info("a")
	rl.Info("b")
	rl.Clear()
	if rl.EntryCount() != 0 {
		t.Errorf("expected 0 after clear, got %d", rl.EntryCount())
	}
}

func TestRichLog_MinLevel(t *testing.T) {
	rl := NewRichLog()
	rl.SetMinLevel(LogWarn)
	rl.Info("hidden")
	rl.Warn("visible")
	rl.Error("visible")
	// EntryCount counts ALL entries, but paint only shows >= minLevel
	if rl.EntryCount() != 3 {
		t.Errorf("expected 3 total entries, got %d", rl.EntryCount())
	}
}

func TestRichLog_Paint(t *testing.T) {
	rl := NewRichLog()
	rl.Info("hello world")
	rl.Warn("warning text")
	rl.Error("error here")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_WithBounds(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 10; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_ZeroBounds(t *testing.T) {
	rl := NewRichLog()
	rl.Info("test")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	rl.Paint(buf) // should not panic
}

func TestRichLog_Paint_NoTimeNoLevel(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(false)
	rl.SetShowLevels(false)
	rl.Info("plain text")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	rl.Paint(buf)
}

func TestRichLog_Paint_Narrow(t *testing.T) {
	rl := NewRichLog()
	rl.Info("a very long line that needs to wrap in a narrow terminal window")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	rl.Paint(buf)
}

func TestRichLog_Paint_Wrapping(t *testing.T) {
	rl := NewRichLog()
	rl.Info("the quick brown fox jumps over the lazy dog multiple times")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	rl.Paint(buf)
}

func TestRichLog_ScrollUp(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	rl.ScrollUp(5)
	if rl.ScrollY() != 5 {
		t.Errorf("expected scrollY 5, got %d", rl.ScrollY())
	}
	if rl.Following() {
		t.Error("expected following=false after scroll up")
	}
}

func TestRichLog_ScrollDown(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	rl.ScrollUp(10)
	rl.ScrollDown(3)
	if rl.ScrollY() != 7 {
		t.Errorf("expected scrollY 7, got %d", rl.ScrollY())
	}
}

func TestRichLog_ScrollDown_ClampToZero(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.ScrollUp(5)
	rl.ScrollDown(10)
	if rl.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", rl.ScrollY())
	}
	if !rl.Following() {
		t.Error("expected following=true after scrolling back to bottom")
	}
}

func TestRichLog_ScrollToTop(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.ScrollToTop()
	if rl.Following() {
		t.Error("expected following=false at top")
	}
}

func TestRichLog_ScrollToBottom(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.ScrollUp(10)
	rl.ScrollToBottom()
	if rl.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", rl.ScrollY())
	}
	if !rl.Following() {
		t.Error("expected following=true at bottom")
	}
}

func TestRichLog_AutoScroll(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 10; i++ {
		rl.Info("line")
	}
	rl.ScrollUp(3)
	// New entry while auto-scroll is on should re-follow
	rl.Info("new")
	if !rl.Following() {
		t.Error("expected following=true after write with auto-scroll on")
	}
}

func TestRichLog_AutoScrollOff(t *testing.T) {
	rl := NewRichLog()
	rl.SetAutoScroll(false)
	for i := 0; i < 10; i++ {
		rl.Info("line")
	}
	rl.ScrollUp(3)
	rl.Info("new")
	if rl.Following() {
		t.Error("expected following=false with auto-scroll off")
	}
}

func TestRichLog_HandleKey_Arrows(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})

	if !rl.HandleKey(&term.KeyEvent{Key: term.KeyUp}) {
		t.Error("expected KeyUp consumed")
	}
	if rl.ScrollY() != 1 {
		t.Errorf("expected scrollY 1, got %d", rl.ScrollY())
	}

	if !rl.HandleKey(&term.KeyEvent{Key: term.KeyDown}) {
		t.Error("expected KeyDown consumed")
	}
	if rl.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", rl.ScrollY())
	}
}

func TestRichLog_HandleKey_PageUp(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if rl.ScrollY() != 5 {
		t.Errorf("expected scrollY 5, got %d", rl.ScrollY())
	}
}

func TestRichLog_HandleKey_VimKeys(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})

	rl.HandleKey(&term.KeyEvent{Rune: 'k'})
	if rl.ScrollY() != 1 {
		t.Errorf("expected scrollY 1, got %d", rl.ScrollY())
	}
	rl.HandleKey(&term.KeyEvent{Rune: 'j'})
	if rl.ScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", rl.ScrollY())
	}
	rl.HandleKey(&term.KeyEvent{Rune: 'g'})
	if rl.Following() {
		t.Error("expected following=false after 'g'")
	}
	rl.HandleKey(&term.KeyEvent{Rune: 'G'})
	if !rl.Following() {
		t.Error("expected following=true after 'G'")
	}
}

func TestRichLog_HandleKey_Unknown(t *testing.T) {
	rl := NewRichLog()
	if rl.HandleKey(&term.KeyEvent{Rune: 'x'}) {
		t.Error("expected unknown key not consumed")
	}
}

func TestRichLog_SetShowLevels(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowLevels(false)
	if rl.ShowLevels() {
		t.Error("expected show levels false")
	}
}

func TestRichLog_SetShowTime(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(false)
	if rl.ShowTime() {
		t.Error("expected show time false")
	}
}

func TestRichLog_SetStyle(t *testing.T) {
	rl := NewRichLog()
	s := DefaultRichLogStyle()
	s.InfoStyle = buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)}
	rl.SetStyle(s)
	if rl.Style().InfoStyle.Fg.Val != buffer.NamedGreen {
		t.Error("style not set correctly")
	}
}

func TestRichLog_Measure(t *testing.T) {
	rl := NewRichLog()
	rl.Info("hello")
	rl.Info("world")
	s := rl.Measure(Bounded(80, 24))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero measure, got %v", s)
	}
}

func TestRichLog_LevelName(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogDebug, "DEBUG"},
		{LogInfo, " INFO"},
		{LogWarn, " WARN"},
		{LogError, "ERROR"},
		{LogFatal, "FATAL"},
		{LogLevel(99), "?????"},
	}
	for _, tc := range tests {
		got := LevelName(tc.level)
		if got != tc.want {
			t.Errorf("LevelName(%d) = %q, want %q", tc.level, got, tc.want)
		}
	}
}

func TestRichLog_LevelColor(t *testing.T) {
	style := DefaultRichLogStyle()
	for _, level := range []LogLevel{LogDebug, LogInfo, LogWarn, LogError, LogFatal, LogLevel(99)} {
		s := LevelColor(level, style)
		_ = s
	}
}

func TestRichLog_Entries_DefensiveCopy(t *testing.T) {
	rl := NewRichLog()
	rl.Info("hello")
	e1 := rl.Entries()
	e1[0].Text = "modified"
	if rl.Entries()[0].Text != "hello" {
		t.Error("Entries() should return defensive copy")
	}
}

func TestRichLog_ConcurrentWrite(t *testing.T) {
	rl := NewRichLog()
	done := make(chan struct{})
	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rl.Info("concurrent")
			}
			done <- struct{}{}
		}()
	}
	// Concurrent reader
	go func() {
		for j := 0; j < 100; j++ {
			rl.Entries()
		}
		done <- struct{}{}
	}()
	for i := 0; i < 6; i++ {
		<-done
	}
	if rl.EntryCount() != 500 {
		t.Errorf("expected 500 entries, got %d", rl.EntryCount())
	}
}

func TestRichLog_Paint_HighVolume(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	for i := 0; i < 200; i++ {
		rl.Infof("log entry number %d", i)
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	rl.Paint(buf)
	if rl.EntryCount() != 100 {
		t.Errorf("expected 100 entries after trim, got %d", rl.EntryCount())
	}
}

func TestRichLog_Paint_PaintAfterScroll(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 30; i++ {
		rl.Infof("entry %d", i)
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	rl.ScrollUp(10)
	buf := buffer.NewBuffer(60, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_WithUnicode(t *testing.T) {
	rl := NewRichLog()
	rl.Info("Unicode test: café naïve 日本語")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_WithEmptyEntries(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	rl.Paint(buf) // should not panic
}

func TestWrapText_Simple(t *testing.T) {
	lines := wrapText("hello world", 80)
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestWrapText_Wrapping(t *testing.T) {
	lines := wrapText("aaa bbb ccc ddd", 7)
	// Should wrap to multiple lines
	if len(lines) < 2 {
		t.Errorf("expected >= 2 lines, got %d", len(lines))
	}
}

func TestWrapText_Empty(t *testing.T) {
	lines := wrapText("", 80)
	if len(lines) != 1 {
		t.Errorf("expected 1 line for empty, got %d", len(lines))
	}
}

func TestWrapText_ZeroWidth(t *testing.T) {
	lines := wrapText("hello", 0)
	if len(lines) != 1 {
		t.Errorf("expected passthrough for width 0, got %d", len(lines))
	}
}

func TestVisibleWidth(t *testing.T) {
	if visibleWidth("hello") != 5 {
		t.Error("expected 5")
	}
	if visibleWidth("") != 0 {
		t.Error("expected 0")
	}
}

// Ensure output contains expected substrings
func TestRichLog_Paint_OutputContent(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(false)
	rl.Info("testmarker")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	rl.Paint(buf)

	var out strings.Builder
	for x := 0; x < 80; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune != 0 && cell.Rune != ' ' {
			out.WriteRune(cell.Rune)
		}
	}
	if !strings.Contains(out.String(), "testmarker") {
		t.Errorf("expected 'testmarker' in output, got %q", out.String())
	}
}
