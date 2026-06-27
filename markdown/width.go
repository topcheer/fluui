package markdown

import (
	"github.com/topcheer/fluui/internal/buffer"
	"strings"
	"unicode"
)

// WrapText wraps text to fit within a given display width, respecting
// CJK wide characters (width=2) and ASCII (width=1).
//
// Rules:
//   - Break at spaces when possible (greedy word wrap).
//   - CJK characters can break between any two characters.
//   - Never split a single rune across lines.
//   - Zero-width characters (combining marks) stay attached to the previous rune.
func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []string{""}
	}

	var lines []string
	var lineRunes []rune
	lineWidth := 0

	i := 0
	for i < len(runes) {
		r := runes[i]
		rw := buffer.RuneWidth(r)

		// Zero-width character: always attach to current line.
		if rw == 0 {
			lineRunes = append(lineRunes, r)
			i++
			continue
		}

		// Check if this rune fits on the current line.
		if lineWidth+rw > width && lineWidth > 0 {
			// Need to break.
			// If current char is a space, just end the line here (skip the space).
			if r == ' ' || r == '\t' {
				lines = append(lines, string(lineRunes))
				lineRunes = nil
				lineWidth = 0
				i++ // skip the whitespace
				continue
			}

			// Try to break at the last space in the current line.
			if spaceIdx := lastSpaceIndex(lineRunes); spaceIdx >= 0 {
				lines = append(lines, strings.TrimSpace(string(lineRunes[:spaceIdx])))
				// Remaining runes after the space.
				remaining := lineRunes[spaceIdx+1:]
				lineRunes = append([]rune{}, remaining...)
				lineWidth = runesWidth(remaining)
				// Don't increment i — process current rune again.
				continue
			}

			// No space to break at. For CJK characters, we can break anywhere.
			// But never split a width-2 rune into a width-1 line.
			if rw >= width {
				// Character is wider than the line — put it on its own line.
				lines = append(lines, string(lineRunes))
				lineRunes = []rune{r}
				lineWidth = rw
				i++
				continue
			}

			// For a single word that's too long, break mid-word.
			lines = append(lines, string(lineRunes))
			lineRunes = nil
			lineWidth = 0
			// Don't increment i — process current rune again.
			continue
		}

		// If we're at the start of a line and the character is a space, skip it.
		if lineWidth == 0 && (r == ' ' || r == '\t') {
			i++
			continue
		}

		lineRunes = append(lineRunes, r)
		lineWidth += rw
		i++
	}

	// Flush remaining runes.
	if len(lineRunes) > 0 {
		lines = append(lines, string(lineRunes))
	}

	// Handle empty input.
	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return lines
}

// runesWidth calculates the total display width of a slice of runes.
func runesWidth(runes []rune) int {
	w := 0
	for _, r := range runes {
		w += buffer.RuneWidth(r)
	}
	return w
}

// lastSpaceIndex returns the index of the last space/tab in the rune slice, or -1.
func lastSpaceIndex(runes []rune) int {
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == ' ' || runes[i] == '\t' {
			return i
		}
	}
	return -1
}

// StringWidth returns the display width of a string (sum of rune widths).
func StringWidth(s string) int {
	return runesWidth([]rune(s))
}

// Truncate truncates a string to fit within width, appending an ellipsis
// if characters were removed. The ellipsis is counted towards the width.
func Truncate(s string, width int, ellipsis string) string {
	fullWidth := StringWidth(s)
	if fullWidth <= width {
		return s
	}

	ellipsisWidth := StringWidth(ellipsis)
	if width <= ellipsisWidth {
		// Width too small for ellipsis — just truncate hard.
		var result []rune
		curW := 0
		for _, r := range s {
			rw := buffer.RuneWidth(r)
			if curW+rw > width {
				break
			}
			result = append(result, r)
			curW += rw
		}
		return string(result)
	}

	targetWidth := width - ellipsisWidth
	var result []rune
	curW := 0
	for _, r := range s {
		rw := buffer.RuneWidth(r)
		if curW+rw > targetWidth {
			break
		}
		result = append(result, r)
		curW += rw
	}
	return string(result) + ellipsis
}

// PadRight pads a string with spaces on the right to reach the given width.
func PadRight(s string, width int) string {
	sw := StringWidth(s)
	if sw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}

// IsBreakable returns true if a line break is preferred after this rune.
// CJK characters (width=2) are always breakable. ASCII characters are
// breakable only if they're whitespace or punctuation.
func IsBreakable(r rune) bool {
	if buffer.RuneWidth(r) == 2 {
		return true // CJK characters can break after any character
	}
	return unicode.IsSpace(r) || unicode.IsPunct(r)
}
