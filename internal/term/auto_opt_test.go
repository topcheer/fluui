//go:build !windows

package term

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// ─── Terminal method tests via NewTestTerminal ─────────────────────

func TestTerminal_Write(t *testing.T) {
	var buf bytes.Buffer
	term := NewTestTerminal(nil, &buf, 80, 24)

	n, err := term.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write() n = %d, want 5", n)
	}
	if buf.String() != "hello" {
		t.Errorf("buf = %q, want 'hello'", buf.String())
	}
}

func TestTerminal_Write_Empty(t *testing.T) {
	var buf bytes.Buffer
	term := NewTestTerminal(nil, &buf, 80, 24)

	n, err := term.Write(nil)
	if err != nil {
		t.Fatalf("Write(nil) error: %v", err)
	}
	if n != 0 {
		t.Errorf("Write(nil) n = %d, want 0", n)
	}
}

func TestTerminal_Write_Concurrent(t *testing.T) {
	var buf bytes.Buffer
	term := NewTestTerminal(nil, &buf, 80, 24)

	done := make(chan struct{}, 2)
	go func() {
		term.Write([]byte("AAA"))
		done <- struct{}{}
	}()
	go func() {
		term.Write([]byte("BBB"))
		done <- struct{}{}
	}()

	<-done
	<-done
	// Should be 6 bytes total (concurrent writes shouldn't lose data)
	if buf.Len() != 6 {
		t.Errorf("buf.Len() = %d, want 6", buf.Len())
	}
}

func TestTerminal_WriteRaw(t *testing.T) {
	var buf bytes.Buffer
	term := NewTestTerminal(nil, &buf, 80, 24)

	term.WriteRaw("test\x1b[2J")
	if buf.String() != "test\x1b[2J" {
		t.Errorf("buf = %q, want 'test\\x1b[2J'", buf.String())
	}
}

func TestTerminal_WriteRaw_Empty(t *testing.T) {
	var buf bytes.Buffer
	term := NewTestTerminal(nil, &buf, 80, 24)

	term.WriteRaw("")
	if buf.Len() != 0 {
		t.Errorf("buf.Len() = %d, want 0", buf.Len())
	}
}

func TestTerminal_Read(t *testing.T) {
	term := NewTestTerminal(strings.NewReader("input"), nil, 80, 24)

	buf := make([]byte, 5)
	n, err := term.Read(buf)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if n != 5 {
		t.Errorf("Read() n = %d, want 5", n)
	}
	if string(buf) != "input" {
		t.Errorf("buf = %q, want 'input'", string(buf))
	}
}

func TestTerminal_Read_Partial(t *testing.T) {
	term := NewTestTerminal(strings.NewReader("hello"), nil, 80, 24)

	buf := make([]byte, 3)
	n, err := term.Read(buf)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if n != 3 {
		t.Errorf("Read() n = %d, want 3", n)
	}
	if string(buf) != "hel" {
		t.Errorf("buf = %q, want 'hel'", string(buf))
	}
}

func TestTerminal_Read_Empty(t *testing.T) {
	term := NewTestTerminal(strings.NewReader(""), nil, 80, 24)

	buf := make([]byte, 10)
	n, _ := term.Read(buf)
	// strings.NewReader returns 0, io.EOF on empty
	if n != 0 {
		t.Errorf("Read() n = %d, want 0", n)
	}
}

func TestTerminal_Size(t *testing.T) {
	term := NewTestTerminal(nil, nil, 120, 40)
	w, h := term.Size()
	if w != 120 {
		t.Errorf("Size() w = %d, want 120", w)
	}
	if h != 40 {
		t.Errorf("Size() h = %d, want 40", h)
	}
}

func TestTerminal_Size_Default(t *testing.T) {
	term := NewTestTerminal(nil, nil, 80, 24)
	w, h := term.Size()
	if w != 80 {
		t.Errorf("Size() w = %d, want 80", w)
	}
	if h != 24 {
		t.Errorf("Size() h = %d, want 24", h)
	}
}

func TestTerminal_Size_Zero(t *testing.T) {
	term := NewTestTerminal(nil, nil, 0, 0)
	w, h := term.Size()
	if w != 0 {
		t.Errorf("Size() w = %d, want 0", w)
	}
	if h != 0 {
		t.Errorf("Size() h = %d, want 0", h)
	}
}

func TestTerminal_ColorProfile(t *testing.T) {
	term := NewTestTerminal(nil, nil, 80, 24)
	// NewTestTerminal sets profile to Profile256
	profile := term.ColorProfile()
	if profile != Profile256 {
		t.Errorf("ColorProfile() = %v, want Profile256", profile)
	}
}

func TestTerminal_SupportsMouse(t *testing.T) {
	term := NewTestTerminal(nil, nil, 80, 24)
	if !term.SupportsMouse() {
		t.Error("SupportsMouse() should always return true")
	}
}

func TestTerminal_ResizeCh(t *testing.T) {
	term := NewTestTerminal(nil, nil, 80, 24)
	ch := term.ResizeCh()
	if ch == nil {
		t.Fatal("ResizeCh() returned nil")
	}

	// Channel should initially be empty (receive-only from outside)
	select {
	case <-ch:
		t.Error("ResizeCh() should be empty initially")
	default:
		// Good — no signal pending
	}

	// Simulate a resize by sending on the internal resizeCh
	term.resizeCh <- struct{}{}

	// Now should be able to receive
	select {
	case <-ch:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("ResizeCh() should have a pending signal after internal send")
	}

	// Channel should be empty again
	select {
	case <-ch:
		t.Error("ResizeCh() should be empty after receiving")
	default:
		// Good
	}
}

// ─── parseCSI edge cases ───────────────────────────────────────────

func TestParseCSI_EmptyBuf(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI(nil)
	if ev != nil {
		t.Error("parseCSI(nil) should return nil")
	}
}

func TestParseCSI_UnknownTilde(t *testing.T) {
	p := NewParser()
	// Tilde sequence with unknown param (e.g., 99) — falls through, returns nil
	p.state = stateCSI
	ev := p.parseCSI([]byte("99~"))
	if ev != nil {
		t.Error("parseCSI with unknown tilde 99 should return nil")
	}
}

func TestParseCSI_ZeroTilde(t *testing.T) {
	p := NewParser()
	p.state = stateCSI
	ev := p.parseCSI([]byte("0~"))
	if ev != nil {
		t.Error("parseCSI with tilde 0 should return nil")
	}
}

// ─── parseSGRMouse edge cases ──────────────────────────────────────

func TestParseSGRMouse_MotionWithButton(t *testing.T) {
	// Button 35 = 32(motion) + 3(btn=3) — SGR mouse with motion flag
	p := NewParser()
	ev := p.parseSGRMouse([]byte("35;10;20M"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("parseSGRMouse should return event")
	}
	// Verify coordinates are parsed correctly (x-1, y-1)
	if ev.Mouse.X != 9 {
		t.Errorf("X = %d, want 9", ev.Mouse.X)
	}
	if ev.Mouse.Y != 19 {
		t.Errorf("Y = %d, want 19", ev.Mouse.Y)
	}
}

func TestParseSGRMouse_ReleaseWithModifier(t *testing.T) {
	// Button 35 + release 'm' suffix
	p := NewParser()
	ev := p.parseSGRMouse([]byte("35;5;5m"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("parseSGRMouse should return event")
	}
}

// ─── FeedTimeout edge cases ────────────────────────────────────────

func TestFeedTimeout_NotInEscapeState(t *testing.T) {
	escTimeoutForTest = 10 * time.Millisecond
	p := NewParser()
	// Parser is in stateNormal, not stateEscape
	ev := p.FeedTimeout()
	if ev != nil {
		t.Error("FeedTimeout in stateNormal should return nil")
	}
}

func TestFeedTimeout_ImmediateNotInEscape(t *testing.T) {
	escTimeoutForTest = 0
	p := NewParser()
	// stateNormal, immediate mode
	ev := p.FeedTimeout()
	if ev != nil {
		t.Error("FeedTimeout in stateNormal with immediate mode should return nil")
	}
}

// ─── ParseOSC52Response edge case ──────────────────────────────────

func TestParseOSC52Response_MalformedPrefix(t *testing.T) {
	// Valid format starts with "52;" but we test the path where prefix is wrong
	resp := "notosc52data"
	// ParseOSC52Response should handle gracefully
	result, ok := ParseOSC52Response(resp)
	if ok {
		t.Error("ParseOSC52Response should return ok=false for malformed prefix")
	}
	if result != "" {
		t.Errorf("ParseOSC52Response should return empty string for malformed prefix, got %q", result)
	}
}
