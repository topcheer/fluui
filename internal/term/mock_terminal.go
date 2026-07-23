package term

import (
	"bytes"
	"io"
	"sync"
)

// MockTerminal is a test double for Terminal that requires no /dev/tty.
// It implements the same public API as Terminal (Write, WriteRaw, Read,
// Size, ColorProfile, Close) using in-memory buffers.
//
// Use it in tests where you need a Terminal-compatible object without
// real terminal I/O.
type MockTerminal struct {
	mu       sync.Mutex
	out      bytes.Buffer
	in       *bytes.Buffer
	width    int
	height   int
	profile  ColorProfile
	closed   bool
}

// NewMockTerminal creates a mock terminal with the given dimensions
// and 256-color profile.
func NewMockTerminal(width, height int) *MockTerminal {
	return &MockTerminal{
		width:   width,
		height:  height,
		profile: Profile256,
		in:      &bytes.Buffer{},
	}
}

// NewMockTerminalWithProfile creates a mock with a specific color profile.
func NewMockTerminalWithProfile(width, height int, profile ColorProfile) *MockTerminal {
	return &MockTerminal{
		width:   width,
		height:  height,
		profile: profile,
		in:      &bytes.Buffer{},
	}
}

// Write writes bytes to the internal output buffer.
func (t *MockTerminal) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, io.ErrClosedPipe
	}
	return t.out.Write(b)
}

// WriteRaw writes a string to the internal output buffer.
func (t *MockTerminal) WriteRaw(s string) {
	t.Write([]byte(s))
}

// Read reads from the internal input buffer.
func (t *MockTerminal) Read(b []byte) (int, error) {
	return t.in.Read(b)
}

// SetInput sets the data that subsequent Read calls will return.
func (t *MockTerminal) SetInput(data []byte) {
	t.in.Reset()
	t.in.Write(data)
}

// Size returns the configured dimensions.
func (t *MockTerminal) Size() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.width, t.height
}

// SetSize updates the terminal dimensions (for resize testing).
func (t *MockTerminal) SetSize(w, h int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.width = w
	t.height = h
}

// ColorProfile returns the configured profile.
func (t *MockTerminal) ColorProfile() ColorProfile {
	return t.profile
}

// SetColorProfile updates the color profile.
func (t *MockTerminal) SetColorProfile(p ColorProfile) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.profile = p
}

// Close marks the terminal as closed.
func (t *MockTerminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

// IsClosed returns whether Close was called.
func (t *MockTerminal) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

// Output returns all bytes written to the terminal.
func (t *MockTerminal) Output() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.out.Bytes()
}

// OutputString returns all written output as a string.
func (t *MockTerminal) OutputString() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.out.String()
}

// Reset clears all buffers and resets state.
func (t *MockTerminal) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.out.Reset()
	t.in.Reset()
	t.closed = false
}

// SupportsMouse always returns true for the mock.
func (t *MockTerminal) SupportsMouse() bool {
	return true
}

// ResizeChan returns nil (no real resize events in mock).
func (t *MockTerminal) ResizeChan() <-chan struct{} {
	return nil
}
