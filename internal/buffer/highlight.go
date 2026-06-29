package buffer

// TextMatch describes a text range within the buffer that should be
// highlighted as a search result.
type TextMatch struct {
	X      int // start column
	Y      int // row (line number)
	Length int // number of character cells to highlight
}

// HighlightMatches applies highlight styling (reverse video) to all
// cells covered by the given TextMatch ranges. Cells outside the ranges
// are left unchanged.
//
// This is used to visually emphasize search results in the rendered output.
func HighlightMatches(buf *Buffer, matches []TextMatch, style Style) {
	for _, m := range matches {
		for i := 0; i < m.Length; i++ {
			x := m.X + i
			idx := buf.idx(x, m.Y)
			if idx < 0 {
				break
			}
			cell := &buf.Cells[idx]
			cell.Fg = style.Fg
			cell.Bg = style.Bg
			cell.Flags |= style.Flags
		}
	}
}

// HighlightCurrentMatch is like HighlightMatches but applies a distinct
// style to the single current match (e.g. a brighter/different color).
// All other matches get the normal match style.
func HighlightCurrentMatch(buf *Buffer, matches []TextMatch, current int, normalStyle, currentStyle Style) {
	for i, m := range matches {
		s := normalStyle
		if i == current {
			s = currentStyle
		}
		for j := 0; j < m.Length; j++ {
			x := m.X + j
			idx := buf.idx(x, m.Y)
			if idx < 0 {
				break
			}
			cell := &buf.Cells[idx]
			cell.Fg = s.Fg
			cell.Bg = s.Bg
			cell.Flags |= s.Flags
		}
	}
}

// FindTextInRow searches a specific row of the buffer for all occurrences
// of a substring and returns TextMatch entries for each hit.
// The search is case-sensitive. Matches that extend beyond the buffer
// width are clamped.
func FindTextInRow(buf *Buffer, y int, query string) []TextMatch {
	if y < 0 || y >= buf.Height || len(query) == 0 {
		return nil
	}

	// Build the string from the row's cells.
	row := make([]rune, 0, buf.Width)
	for x := 0; x < buf.Width; x++ {
		cell := buf.GetCell(x, y)
		if cell.Width == 0 {
			// Skip combining characters in search
			continue
		}
		row = append(row, cell.Rune)
	}

	rowStr := string(row)
	var matches []TextMatch
	searchStart := 0

	for {
		idx := indexFromString(rowStr, query, searchStart)
		if idx < 0 {
			break
		}
		matches = append(matches, TextMatch{
			X:      idx,
			Y:      y,
			Length: len(query),
		})
		searchStart = idx + len(query)
		if searchStart >= len(rowStr) {
			break
		}
	}

	return matches
}

// FindTextInBuffer scans the entire buffer for occurrences of query
// and returns TextMatch entries for each row that contains matches.
func FindTextInBuffer(buf *Buffer, query string) []TextMatch {
	if len(query) == 0 {
		return nil
	}

	var allMatches []TextMatch
	for y := 0; y < buf.Height; y++ {
		matches := FindTextInRow(buf, y, query)
		allMatches = append(allMatches, matches...)
	}
	return allMatches
}

// indexFromString is a simple substring search that works with the
// rune-based row string. It returns the rune index of the first match
// at or after start, or -1.
func indexFromString(s, sub string, start int) int {
	if start < 0 {
		start = 0
	}
	runes := []rune(s)
	subRunes := []rune(sub)
	n := len(runes)
	m := len(subRunes)

	if m == 0 {
		return start
	}
	if m > n-start {
		return -1
	}

	for i := start; i <= n-m; i++ {
		match := true
		for j := 0; j < m; j++ {
			if runes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
