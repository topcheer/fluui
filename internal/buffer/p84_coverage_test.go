package buffer

import "testing"

// ─── isWide (53.3% → 100%) ───

func TestP84_IsWide_AllRanges(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		// ASCII — not wide
		{"ASCII a", 'a', false},
		{"ASCII space", ' ', false},
		{"ASCII digit", '5', false},

		// Hangul Jamo (0x1100-0x115F)
		{"Hangul Jamo start", 0x1100, true},
		{"Hangul Jamo end", 0x115F, true},

		// CJK Radicals (0x2E80-0x303E)
		{"CJK Radical start", 0x2E80, true},
		{"CJK Radical end", 0x303E, true},

		// Hiragana, Katakana, CJK symbols (0x3040-0x33BF)
		{"Hiragana start", 0x3040, true},
		{"Katakana end", 0x33BF, true},
		{"Hiragana あ", 'あ', true},
		{"Katakana カ", 'カ', true},

		// CJK Extension A (0x3400-0x4DBF)
		{"CJK Ext A start", 0x3400, true},
		{"CJK Ext A end", 0x4DBF, true},

		// CJK Unified (0x4E00-0x9FFF)
		{"CJK start", 0x4E00, true},
		{"CJK end", 0x9FFF, true},
		{"CJK 字", '字', true},
		{"CJK 日本", '日', true},

		// Yi Syllables (0xA000-0xA4CF)
		{"Yi start", 0xA000, true},
		{"Yi end", 0xA4CF, true},

		// Hangul Syllables (0xAC00-0xD7A3)
		{"Hangul Syllable start", 0xAC00, true},
		{"Hangul Syllable 가", '가', true},

		// CJK Compatibility (0xF900-0xFAFF)
		{"CJK Compat start", 0xF900, true},
		{"CJK Compat end", 0xFAFF, true},

		// CJK Compat Forms (0xFE30-0xFE4F)
		{"CJK Compat Form start", 0xFE30, true},
		{"CJK Compat Form end", 0xFE4F, true},

		// Fullwidth Forms (0xFF00-0xFF60)
		{"Fullwidth start", 0xFF01, true},
		{"Fullwidth ！", '！', true},
		{"Fullwidth end", 0xFF60, true},

		// Fullwidth Signs (0xFFE0-0xFFE6)
		{"Fullwidth Sign start", 0xFFE0, true},
		{"Fullwidth Sign ￠", '￠', true},
		{"Fullwidth Sign end", 0xFFE6, true},

		// Emoji (0x1F300-0x1FAFF)
		{"Emoji start", 0x1F300, true},
		{"Emoji 😀", 0x1F600, true},
		{"Emoji end", 0x1FAFF, true},

		// CJK Extension B-F (0x20000-0x3FFFD)
		{"CJK Ext B start", 0x20000, true},
		{"CJK Ext B mid", 0x2F800, true},
		{"CJK Ext F end", 0x3FFFD, true},

		// Negative: just outside ranges
		{"Before Hangul Jamo", 0x10FF, false},
		{"After Hangul Jamo", 0x1160, false},
		{"Before CJK Radical", 0x2E7F, false},
		{"After CJK Radical range", 0x303F, false},
		{"After Yi Syllables", 0xA4D0, false},
		{"Before Emoji", 0x1F2FF, false},
		{"After Emoji", 0x1FB00, false},
		{"Before CJK Ext B", 0x1FFFF, false},
		{"After CJK Ext F", 0x40000, false},

		// Latin accented — not wide
		{"é", 'é', false},
		{"ñ", 'ñ', false},
		{"ü", 'ü', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWide(tt.r); got != tt.want {
				t.Errorf("isWide(U+%04X) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

// ─── isZeroWidth additional coverage ───

func TestP84_IsZeroWidth_CombiningMarks(t *testing.T) {
	// Combining Diacritical Marks (0x0300-0x036F)
	if !isZeroWidth(0x0300) {
		t.Error("combining grave accent should be zero-width")
	}
	if !isZeroWidth(0x036F) {
		t.Error("0x036F should be zero-width")
	}
	// Just outside range
	if isZeroWidth(0x02FF) {
		t.Error("0x02FF should NOT be zero-width")
	}
	if isZeroWidth(0x0370) {
		t.Error("0x0370 should NOT be zero-width")
	}
}

func TestP84_IsZeroWidth_VariationSelectors(t *testing.T) {
	// Variation Selectors (0xFE00-0xFE0F)
	if !isZeroWidth(0xFE00) {
		t.Error("VS1 should be zero-width")
	}
	if !isZeroWidth(0xFE0F) {
		t.Error("VS16 should be zero-width")
	}
	// Just outside
	if isZeroWidth(0xFDFF) {
		t.Error("0xFDFF should NOT be zero-width")
	}
}

func TestP84_IsZeroWidth_CombiningHalfMarks(t *testing.T) {
	// Combining Half Marks (0xFE20-0xFE2F)
	if !isZeroWidth(0xFE20) {
		t.Error("0xFE20 should be zero-width")
	}
	if !isZeroWidth(0xFE2F) {
		t.Error("0xFE2F should be zero-width")
	}
}

func TestP84_IsZeroWidth_NormalChar(t *testing.T) {
	if isZeroWidth('a') {
		t.Error("'a' should NOT be zero-width")
	}
	if isZeroWidth(' ') {
		t.Error("space should NOT be zero-width")
	}
}
