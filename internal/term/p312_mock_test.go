package term

import (
	"io"
	"strings"
	"testing"
)

func TestP312_NewMockTerminal(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	if mt == nil {
		t.Fatal("nil")
	}
	w, h := mt.Size()
	if w != 80 || h != 24 {
		t.Errorf("size = %dx%d", w, h)
	}
	if mt.ColorProfile() != Profile256 {
		t.Errorf("profile = %v", mt.ColorProfile())
	}
}

func TestP312_Write(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.Write([]byte("hello"))
	if mt.OutputString() != "hello" {
		t.Errorf("output = %q", mt.OutputString())
	}
}

func TestP312_WriteRaw(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.WriteRaw("\x1b[2J\x1b[H")
	if !strings.Contains(mt.OutputString(), "\x1b[2J") {
		t.Error("missing escape sequence")
	}
}

func TestP312_Read(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.SetInput([]byte("input data"))
	buf := make([]byte, 10)
	n, err := mt.Read(buf)
	if err != nil || n != 10 {
		t.Errorf("Read: n=%d err=%v", n, err)
	}
	if string(buf) != "input data" {
		t.Errorf("read = %q", string(buf))
	}
}

func TestP312_SetSize(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.SetSize(120, 40)
	w, h := mt.Size()
	if w != 120 || h != 40 {
		t.Errorf("size = %dx%d", w, h)
	}
}

func TestP312_SetColorProfile(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.SetColorProfile(ProfileTrue)
	if mt.ColorProfile() != ProfileTrue {
		t.Errorf("profile = %v", mt.ColorProfile())
	}
}

func TestP312_NewMockWithProfile(t *testing.T) {
	mt := NewMockTerminalWithProfile(80, 24, ProfileTrue)
	if mt.ColorProfile() != ProfileTrue {
		t.Errorf("profile = %v", mt.ColorProfile())
	}
}

func TestP312_Close(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.Close()
	if !mt.IsClosed() {
		t.Error("should be closed")
	}
	// Write after close should error
	_, err := mt.Write([]byte("x"))
	if err != io.ErrClosedPipe {
		t.Errorf("write after close: err = %v", err)
	}
}

func TestP312_Reset(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.Write([]byte("data"))
	mt.Close()
	mt.Reset()
	if mt.IsClosed() {
		t.Error("should not be closed after reset")
	}
	if mt.OutputString() != "" {
		t.Error("output should be empty after reset")
	}
}

func TestP312_Output(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	mt.WriteRaw("abc")
	mt.WriteRaw("def")
	out := mt.Output()
	if string(out) != "abcdef" {
		t.Errorf("output = %q", string(out))
	}
}

func TestP312_SupportsMouse(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	if !mt.SupportsMouse() {
		t.Error("mock should support mouse")
	}
}

func TestP312_ResizeChan(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	if mt.ResizeChan() != nil {
		t.Error("mock resize chan should be nil")
	}
}

func TestP312_ThreadSafe(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			mt.WriteRaw("x")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		w, h := mt.Size()
		_ = w
		_ = h
		_ = mt.OutputString()
	}
	<-done
}

func TestP312_ProtocolSequenceOutput(t *testing.T) {
	mt := NewMockTerminal(80, 24)
	// Simulate writing protocol init sequences
	mt.WriteRaw("\x1b[?1049h") // alt screen
	mt.WriteRaw("\x1b[?25l")   // hide cursor
	mt.WriteRaw("\x1b[?2004h") // bracketed paste

	out := mt.OutputString()
	if !strings.Contains(out, "\x1b[?1049h") {
		t.Error("missing alt screen sequence")
	}
	if !strings.Contains(out, "\x1b[?25l") {
		t.Error("missing hide cursor sequence")
	}
	if !strings.Contains(out, "\x1b[?2004h") {
		t.Error("missing bracketed paste sequence")
	}
}
