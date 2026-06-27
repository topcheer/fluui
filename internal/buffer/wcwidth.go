package buffer

// RuneWidth returns the display width of a rune:
//   - 0 for combining/zero-width characters
//   - 1 for normal-width characters
//   - 2 for East Asian Wide / Fullwidth characters
//
// This is a simplified implementation based on Unicode ranges.
// For production, consider embedding a full wcwidth table.
func RuneWidth(r rune) int {
	switch {
	case r == 0:
		return 0

	// Control characters
	case r < 32 || (r >= 0x7f && r < 0xa0):
		return 0

	// Zero-width characters
	case isZeroWidth(r):
		return 0
	}

	// East Asian Wide / Fullwidth ranges
	if isWide(r) {
		return 2
	}

	return 1
}

func isZeroWidth(r rune) bool {
	// Combining Diacritical Marks
	if r >= 0x0300 && r <= 0x036F {
		return true
	}
	// Combining Diacritical Marks Supplement
	if r >= 0x1DC0 && r <= 0x1DFF {
		return true
	}
	// Variation Selectors
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	// Zero Width Joiner / Non-Joiner
	if r == 0x200D || r == 0x200C {
		return true
	}
	// Zero Width Space
	if r == 0x200B {
		return true
	}
	// Combining Half Marks
	if r >= 0xFE20 && r <= 0xFE2F {
		return true
	}
	return false
}

// isWide returns true for characters that take 2 cells.
func isWide(r rune) bool {
	switch {
	// CJK Unified Ideographs and common ranges
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
		return true
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals, Kangxi
		return true
	case r >= 0x3040 && r <= 0x33BF: // Hiragana, Katakana, CJK symbols, Hangul compat
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Unified Ideographs Extension A
		return true
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
		return true
	case r >= 0xA000 && r <= 0xA4CF: // Yi Syllables, Yi Radicals
		return true
	case r >= 0xAC00 && r <= 0xD7A3: // Hangul Syllables
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case r >= 0xFE30 && r <= 0xFE4F: // CJK Compatibility Forms
		return true
	case r >= 0xFF00 && r <= 0xFF60: // Fullwidth Forms
		return true
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth Signs
		return true
	case r >= 0x1F300 && r <= 0x1FAFF: // Emoji and symbols (Misc Symbols and Pictographs, Emoticons, etc.)
		return true
	case r >= 0x20000 && r <= 0x3FFFD: // CJK Unified Ideographs Extension B-F
		return true
	}
	return false
}
