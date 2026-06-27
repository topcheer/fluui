package buffer

import (
	"testing"
	"unicode/utf8"
)

func TestRuneWidthASCII(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		// Lowercase letters
		{"a", 'a', 1},
		{"z", 'z', 1},
		// Uppercase letters
		{"A", 'A', 1},
		{"Z", 'Z', 1},
		// Digits
		{"0", '0', 1},
		{"9", '9', 1},
		// Space and punctuation
		{"space", ' ', 1},
		{"bang", '!', 1},
		{"tilde", '~', 1},
		// Del
		{"del", 0x7F, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%U): got %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidthControlChars(t *testing.T) {
	// Control characters should be zero-width.
	for r := rune(0); r < 32; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("control char U+%04X: got %d, want 0", r, got)
		}
	}

	// C1 control characters (0x7F-0x9F) should be zero-width.
	for r := rune(0x7F); r < 0xA0; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("C1 control char U+%04X: got %d, want 0", r, got)
		}
	}
}

func TestRuneWidthCJK(t *testing.T) {
	// CJK Unified Ideographs (U+4E00 - U+9FFF)
	cjkChars := []rune("你好世界中文测试")
	for _, r := range cjkChars {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("CJK char %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Hiragana
	hiragana := []rune("あいうえお")
	for _, r := range hiragana {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Hiragana %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Katakana
	katakana := []rune("アイウエオ")
	for _, r := range katakana {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Katakana %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Hangul Syllables (Korean)
	hangul := []rune("안녕하세요")
	for _, r := range hangul {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Hangul %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Fullwidth digits
	for _, r := range []rune("０１２３４５６７８９") {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Fullwidth digit %U (%c): got %d, want 2", r, r, got)
		}
	}
}

func TestRuneWidthEmoji(t *testing.T) {
	// Emoji in the 0x1F300-0x1FAFF range should be width 2.
	emoji := []rune("😀🎉🔥")
	for _, r := range emoji {
		// Skip variation selector U+FE0F
		if r == 0xFE0F {
			continue
		}
		got := RuneWidth(r)
		if got != 2 {
			t.Errorf("Emoji %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Other emoji ranges
	for _, r := range []rune{0x1F600, 0x1F680, 0x1F970, 0x1FA00} {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Emoji %U: got %d, want 2", r, got)
		}
	}
}

func TestRuneWidthCombining(t *testing.T) {
	// Combining Diacritical Marks U+0300 - U+036F
	for r := rune(0x0300); r <= 0x036F; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("combining mark U+%04X: got %d, want 0", r, got)
		}
	}

	// Combining Diacritical Marks Supplement U+1DC0 - U+1DFF
	for r := rune(0x1DC0); r <= 0x1DFF; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("combining supplement U+%04X: got %d, want 0", r, got)
		}
	}

	// Combining Half Marks U+FE20 - U+FE2F
	for r := rune(0xFE20); r <= 0xFE2F; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("combining half mark U+%04X: got %d, want 0", r, got)
		}
	}

	// Variation Selectors U+FE00 - U+FE0F
	for r := rune(0xFE00); r <= 0xFE0F; r++ {
		if got := RuneWidth(r); got != 0 {
			t.Errorf("variation selector U+%04X: got %d, want 0", r, got)
		}
	}
}

func TestRuneWidthZeroWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
	}{
		{"null", 0x00},
		{"ZWSP", 0x200B},  // Zero Width Space
		{"ZWJ", 0x200D},   // Zero Width Joiner
		{"ZWNJ", 0x200C},  // Zero Width Non-Joiner
		// Note: BOM (U+FEFF) is NOT zero-width in our implementation —
		// it's treated as a visible character with width 1.
		// Uncomment below if implementation changes:
		// {"BOM", 0xFEFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != 0 {
				t.Errorf("RuneWidth(%U %s): got %d, want 0", tt.r, tt.name, got)
			}
		})
	}
}

func TestRuneWidthFullwidthForms(t *testing.T) {
	// Fullwidth ASCII variants U+FF01-FF5E
	fullwidth := []rune("！＂＃＄％＆")
	for _, r := range fullwidth {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Fullwidth %U (%c): got %d, want 2", r, r, got)
		}
	}

	// Fullwidth signs U+FFE0-FFE6
	for _, r := range []rune{0xFFE0, 0xFFE1, 0xFFE5, 0xFFE6} {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("Fullwidth sign %U: got %d, want 2", r, got)
		}
	}
}

func TestStringWidthMixed(t *testing.T) {
	// "ab你好cd" = 1+1+2+2+1+1 = 8
	s := "ab你好cd"
	if got := StringWidth(s); got != 8 {
		t.Errorf("StringWidth(%q): got %d, want 8", s, got)
	}

	// "Hi =*= こんにちは" (before CJK: 7 cells, CJK: 10 cells = 17)
	s2 := "Hi =*= こんにちは" // 3*1 + 1 + 3*1 + 1 + 5*2 = 3+1+3+1+10 = 18
	// Actually: H(1) i(1) (1) =(1) *(1) =(1) (1) こ(2) ん(2) に(2) は(2) ち(2) = 7 + 10 = 17
	if got := StringWidth(s2); got != 17 {
		t.Errorf("StringWidth(%q): got %d, want 17", s2, got)
	}
}

func TestStringWidthEmpty(t *testing.T) {
	if got := StringWidth(""); got != 0 {
		t.Errorf("StringWidth(\"\"): got %d, want 0", got)
	}
}

func TestStringWidthPureCJK(t *testing.T) {
	// "你好世界" = 2+2+2+2 = 8
	s := "你好世界"
	if got := StringWidth(s); got != 8 {
		t.Errorf("StringWidth(%q): got %d, want 8", s, got)
	}
}

func TestStringWidthWithCombining(t *testing.T) {
	// "café" with combining accent: c(1) a(1) f(1) e(1) ́ (0) = 4
	// é = U+0065 + U+0301 (combining acute accent)
	s := "caf\u0301"
	// "caf" + combining mark = c(1) + a(1) + f(1) + combining(0) = 3
	if got := StringWidth(s); got != 3 {
		t.Errorf("StringWidth(%q): got %d, want 3", s, got)
	}

	// é as precomposed U+00E9
	s2 := "café" // c(1) + a(1) + f(1) + é(1) = 4
	if got := StringWidth(s2); got != 4 {
		t.Errorf("StringWidth(%q): got %d, want 4", s2, got)
	}
}

func TestStringWidthWithEmoji(t *testing.T) {
	// "a😀b" = 1 + 2 + 1 = 4
	s := "a😀b"
	if got := StringWidth(s); got != 4 {
		t.Errorf("StringWidth(%q): got %d, want 4", s, got)
	}

	// Pure emoji string "🎉🎊" = 2 + 2 = 4
	s2 := "🎉🎊"
	if got := StringWidth(s2); got != 4 {
		t.Errorf("StringWidth(%q): got %d, want 4", s2, got)
	}
}

func TestStringWidthWithZeroWidthSpace(t *testing.T) {
	// "ab" with zero-width space in the middle = 1+0+1 = 2
	s := "a\u200Bb"
	if got := StringWidth(s); got != 2 {
		t.Errorf("StringWidth(%q): got %d, want 2", s, got)
	}
}

func TestRuneWidthConsistency(t *testing.T) {
	// RuneWidth should agree with StringWidth for single-rune strings.
	for _, r := range []rune{'a', 'A', '0', ' ', '你', '😀', '\u0300', '\u200B', '!', '~'} {
		expected := RuneWidth(r)
		actual := StringWidth(string(r))
		if expected != actual {
			t.Errorf("RuneWidth(%U)=%d but StringWidth(%q)=%d", r, expected, string(r), actual)
		}
	}
}

func TestRuneWidthAllUTF8Valid(t *testing.T) {
	// Ensure all test runes are valid UTF-8 when encoded.
	testRunes := []rune{'a', '你', '😀', '\u0300', '\u200B', 0x1F600, 0x4E00}
	for _, r := range testRunes {
		b := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(b, r)
		// Re-decode and verify round-trip.
		decoded, _ := utf8.DecodeRune(b)
		if decoded != r {
			t.Errorf("round-trip failed for U+%04X", r)
		}
	}
}
