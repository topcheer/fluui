package ai

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func newBufReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

// P213: lineScanner.scan() coverage

func TestLineScanner_NormalLine_P213(t *testing.T) {
	s := &lineScanner{reader: newBufReader("hello\nworld\n")}
	if !s.scan() {
		t.Error("should scan first line")
	}
	if s.text() != "hello" {
		t.Errorf("expected 'hello', got %q", s.text())
	}
	if !s.scan() {
		t.Error("should scan second line")
	}
	if s.text() != "world" {
		t.Errorf("expected 'world', got %q", s.text())
	}
}

func TestLineScanner_EmptyReader_P213(t *testing.T) {
	s := &lineScanner{reader: newBufReader("")}
	if s.scan() {
		t.Error("empty reader should return false")
	}
}

func TestLineScanner_CarriageReturn_P213(t *testing.T) {
	s := &lineScanner{reader: newBufReader("line\r\n")}
	s.scan()
	if s.text() != "line" {
		t.Errorf("expected 'line' (CR stripped), got %q", s.text())
	}
}

func TestLineScanner_PartialLine_P213(t *testing.T) {
	s := &lineScanner{reader: newBufReader("partial")}
	if !s.scan() {
		t.Error("should scan partial line")
	}
	if s.text() != "partial" {
		t.Errorf("expected 'partial', got %q", s.text())
	}
	// scan() consumes io.EOF internally; getErr may be nil
}

func TestLineScanner_ReadError_P214(t *testing.T) {
	// Create a reader that returns an error other than EOF
	s := &lineScanner{reader: bufio.NewReader(&errorReader{})}
	if s.scan() {
		t.Error("should return false on read error")
	}
	if s.getErr() == nil {
		t.Error("should have error set")
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
