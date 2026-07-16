package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// stubPanel for testing
type stubPanel struct{}

func (s *stubPanel) HandleKey(k interface{}) bool { return false }
func (s *stubPanel) Paint(buf *buffer.Buffer)     {}
func (s *stubPanel) Title() string                { return "stub" }

// P238: cover pure functions + app_shell branches

func TestSplitLines_MaxWZero_P238(t *testing.T) {
	lines := SplitLines("hello\nworld", 0)
	if len(lines) != 1 || lines[0] != "hello\nworld" {
		t.Errorf("maxW=0 should return single element, got %v", lines)
	}
}

func TestSplitLines_LongLine_P238(t *testing.T) {
	lines := SplitLines("abcdefghij", 3)
	want := []string{"abc", "def", "ghi", "j"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(want), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}

func TestSplitLines_Empty_P238(t *testing.T) {
	lines := SplitLines("", 5)
	if len(lines) != 1 {
		t.Errorf("empty text should return 1 line, got %d", len(lines))
	}
}

func TestReplacePrefix_InvalidCursor_P238(t *testing.T) {
	text, cursor := ReplacePrefix("hello", 0, "he", "HE")
	if text != "hello" || cursor != 0 {
		t.Error("cursor=0 should return unchanged")
	}
	text, cursor = ReplacePrefix("hello", 100, "he", "HE")
	if text != "hello" || cursor != 100 {
		t.Error("cursor>len should return unchanged")
	}
}

func TestReplacePrefix_StartNegative_P238(t *testing.T) {
	text, cursor := ReplacePrefix("ab", 1, "abc", "X")
	if text != "ab" || cursor != 1 {
		t.Error("start<0 should return unchanged")
	}
}

func TestReplacePrefix_PrefixMismatch_P238(t *testing.T) {
	text, cursor := ReplacePrefix("hello", 3, "he", "HE")
	if text != "hello" || cursor != 3 {
		t.Error("prefix mismatch should return unchanged")
	}
}

func TestReplacePrefix_Success_P238(t *testing.T) {
	text, cursor := ReplacePrefix("hello", 2, "he", "HE")
	if text != "HEllo" || cursor != 2 {
		t.Errorf("ReplacePrefix failed: text=%q cursor=%d", text, cursor)
	}
}
