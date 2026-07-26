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

// decodeRuneFromBytes decodes the first UTF-8 rune from b.
// Returns the rune and its byte size. Zero allocation.
func decodeRuneFromBytes(b []byte) (rune, int) {
	if len(b) == 0 {
		return 0, 0
	}
	c := b[0]
	if c < 0x80 {
		return rune(c), 1
	}
	if c&0xE0 == 0xC0 && len(b) >= 2 {
		return rune(c&0x1F)<<6 | rune(b[1]&0x3F), 2
	}
	if c&0xF0 == 0xE0 && len(b) >= 3 {
		return rune(c&0x0F)<<12 | rune(b[1]&0x3F)<<6 | rune(b[2]&0x3F), 3
	}
	if c&0xF8 == 0xF0 && len(b) >= 4 {
		return rune(c&0x07)<<18 | rune(b[1]&0x3F)<<12 | rune(b[2]&0x3F)<<6 | rune(b[3]&0x3F), 4
	}
	return rune(c), 1 // invalid UTF-8
}
