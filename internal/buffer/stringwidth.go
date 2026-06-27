package buffer

// StringWidth returns the display width of s by summing the
// width of each rune. Wide characters (East Asian Fullwidth) count as 2.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}
