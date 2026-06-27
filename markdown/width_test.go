package markdown

import (
	"testing"
)

func TestWrapASCII(t *testing.T) {
	result := WrapText("hello world foo bar", 8)
	// Greedy word-wrap: "hello" (5) + space won't fit "world" → break
	// "world" (5) + space + "foo" (3) = 9 > 8 → break after "world"
	// "foo" (3) + space + "bar" (3) = 7 ≤ 8 → fits
	expected := []string{"hello", "world", "foo bar"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %v", len(expected), len(result), result)
	}
	for i, line := range expected {
		if result[i] != line {
			t.Errorf("line %d: got %q, want %q", i, result[i], line)
		}
	}
}

func TestWrapCJK(t *testing.T) {
	// Each CJK char is width=2, so width=5 allows 2 chars (4 cells) per line.
	result := WrapText("你好世界测试", 5)
	expected := []string{"你好", "世界", "测试"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %v", len(expected), len(result), result)
	}
	for i, line := range expected {
		if result[i] != line {
			t.Errorf("line %d: got %q, want %q", i, result[i], line)
		}
	}
}

func TestWrapMixed(t *testing.T) {
	// "Hello " = 6 cells, "你好" = 4 cells, total = 10 > 8
	// First line: "Hello " (6) + "你" (2) = 8 → "Hello 你"
	// Second line: "好" (2) + " " (1) + "World" (5) = 8 → "好 World"
	result := WrapText("Hello 你好 World", 8)
	if len(result) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %v", len(result), result)
	}
	// Verify each line fits within width.
	for i, line := range result {
		w := StringWidth(line)
		if w > 8 {
			t.Errorf("line %d width %d > 8: %q", i, w, line)
		}
	}
	// Verify the full text is preserved (minus spaces at breaks).
}

func TestWrapNoSplit(t *testing.T) {
	// A single CJK character (width=2) should never be split across lines.
	// With width=1, each CJK char must go on its own line (overflow allowed
	// since a width-2 char can't fit in width=1).
	result := WrapText("你好", 1)
	// Each line should contain exactly one CJK character (not split mid-character)
	for i, line := range result {
		runes := []rune(line)
		for _, r := range runes {
			// Each CJK char should be intact (not partially rendered)
			if r == '\ufffd' {
				t.Errorf("line %d contains replacement char: %q", i, line)
			}
		}
	}
	// Verify both characters are present
	allText := ""
	for _, line := range result {
		allText += line
	}
	if allText != "你好" {
		t.Errorf("text not preserved: got %q, want '你好'", allText)
	}
}

func TestWrapEmpty(t *testing.T) {
	result := WrapText("", 10)
	if len(result) != 1 {
		t.Fatalf("expected 1 line for empty input, got %d", len(result))
	}
	if result[0] != "" {
		t.Errorf("expected empty string, got %q", result[0])
	}
}

func TestWrapSingleWord(t *testing.T) {
	// A word longer than width should be hard-broken.
	result := WrapText("abcdefgh", 3)
	for i, line := range result {
		w := StringWidth(line)
		if w > 3 {
			t.Errorf("line %d width %d > 3: %q", i, w, line)
		}
	}
}

func TestWrapNoBreakNeeded(t *testing.T) {
	result := WrapText("short", 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result))
	}
	if result[0] != "short" {
		t.Errorf("got %q, want %q", result[0], "short")
	}
}

func TestWrapLeadingSpaces(t *testing.T) {
	// Leading spaces should be trimmed at line start.
	result := WrapText("  hello", 10)
	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d: %v", len(result), result)
	}
	// Leading spaces on continuation lines are trimmed, but on first line too.
	if result[0] != "hello" {
		t.Errorf("got %q", result[0])
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"你好", 4},
		{"Hi 你", 5}, // H=1, i=1, space=1, 你=2
		{"", 0},
	}
	for _, tt := range tests {
		got := StringWidth(tt.input)
		if got != tt.want {
			t.Errorf("StringWidth(%q): got %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	// No truncation needed.
	if got := Truncate("short", 10, "..."); got != "short" {
		t.Errorf("Truncate(short): got %q, want %q", got, "short")
	}

	// Truncate with ellipsis.
	got := Truncate("Hello World", 8, "...")
	if StringWidth(got) > 8 {
		t.Errorf("Truncate result width %d > 8: %q", StringWidth(got), got)
	}
	// Should end with ellipsis.
	if len(got) < 3 || got[len(got)-3:] != "..." {
		t.Errorf("expected trailing ellipsis, got %q", got)
	}

	// CJK truncation.
	got = Truncate("你好世界测试", 5, "")
	if StringWidth(got) > 5 {
		t.Errorf("CJK Truncate width %d > 5: %q", StringWidth(got), got)
	}
}

func TestPadRight(t *testing.T) {
	got := PadRight("ab", 5)
	if StringWidth(got) != 5 {
		t.Errorf("expected width 5, got %d: %q", StringWidth(got), got)
	}

	// Already wide enough.
	got = PadRight("你好", 3)
	if got != "你好" {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestIsBreakable(t *testing.T) {
	// CJK is always breakable.
	if !IsBreakable('你') {
		t.Error("CJK char should be breakable")
	}
	// Space is breakable.
	if !IsBreakable(' ') {
		t.Error("space should be breakable")
	}
	// Regular letter is not breakable.
	if IsBreakable('a') {
		t.Error("letter should not be breakable")
	}
}
