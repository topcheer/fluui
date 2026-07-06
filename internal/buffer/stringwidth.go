package buffer

// StringWidth returns the display width of s by summing the
// width of each rune. Wide characters (East Asian Fullwidth) count as 2.
func StringWidth(s string) int {
	// Fast path: pure ASCII — skip UTF-8 decode and RuneWidth calls.
	// Scan for non-ASCII byte; if found, fall back to rune-based loop.
	w := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 0x80 {
			// Non-ASCII byte found — fall back to full computation.
			return w + stringWidthSlow(s[i:])
		}
		// Printable ASCII = width 1, control = width 0.
		if b >= 0x20 && b < 0x7f {
			w++
		}
	}
	return w
}

// stringWidthSlow computes string width using full rune-width logic.
func stringWidthSlow(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}
