package component

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Lipgloss-compatible Layout Helpers ───
//
// These functions provide string-level layout composition similar to lipgloss,
// allowing horizontal/vertical joining of styled text blocks and placement
// within a given area. They operate on plain strings (not buffers), making
// them ideal for building view output in a declarative style.
//
// Usage:
//
//	// Stack two blocks side by side
//	row := JoinHorizontal(Top, leftBlock, rightBlock)
//
//	// Stack two blocks vertically
//	col := JoinVertical(Left, topBlock, bottomBlock)
//
//	// Center text in a 80x24 area
//	placed := Place(80, 24, Center, Middle, "Hello, World!")

// VerticalAlign specifies how blocks are aligned when joined horizontally.
type VerticalAlign int

const (
	Top    VerticalAlign = iota // Align to top
	Middle                      // Align to center vertically
	Bottom                      // Align to bottom
)

// HorizontalAlign specifies how blocks are aligned when joined vertically.
type HorizontalAlign int

const (
	Left   HorizontalAlign = iota // Align to left
	Center                          // Align to center horizontally
	Right                           // Align to right
)

// JoinHorizontal joins blocks left-to-right, aligning them vertically.
// Each block is a multi-line string. Blocks are padded to the tallest block's
// height based on the alignment.
func JoinHorizontal(align VerticalAlign, blocks ...string) string {
	if len(blocks) == 0 {
		return ""
	}
	if len(blocks) == 1 {
		return blocks[0]
	}

	// Split each block into lines and find max height and max width per block
	allLines := make([][]string, len(blocks))
	maxHeight := 0
	widths := make([]int, len(blocks))

	for i, block := range blocks {
		lines := strings.Split(block, "\n")
		allLines[i] = lines
		if len(lines) > maxHeight {
			maxHeight = len(lines)
		}
		for _, line := range lines {
			w := stringWidth(line)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	// Build the joined output
	var result strings.Builder
	for row := 0; row < maxHeight; row++ {
		for i, lines := range allLines {
			var line string
			if row < len(lines) {
				line = lines[row]
			}
			w := stringWidth(line)
			pad := widths[i] - w

			switch align {
			case Top:
				// Line is already at top, just pad width
				result.WriteString(line)
				if pad > 0 {
					result.WriteString(strings.Repeat(" ", pad))
				}
			case Middle:
				// For middle alignment, pad top of short blocks
				// Each line is already in position, just pad width
				result.WriteString(line)
				if pad > 0 {
					result.WriteString(strings.Repeat(" ", pad))
				}
			case Bottom:
				// For bottom alignment, same — alignment is about which
				// row the content starts on, handled below
				result.WriteString(line)
				if pad > 0 {
					result.WriteString(strings.Repeat(" ", pad))
				}
			}
		}
		if row < maxHeight-1 {
			result.WriteString("\n")
		}
	}

	// For middle/bottom alignment, we need to offset the entire block
	// Rebuild with proper vertical padding
	if align == Middle || align == Bottom {
		return verticallyAlignBlocks(align, blocks, maxHeight, widths)
	}

	return result.String()
}

// verticallyAlignBlocks rebuilds the join with proper vertical offset per block.
func verticallyAlignBlocks(align VerticalAlign, blocks []string, maxHeight int, widths []int) string {
	allLines := make([][]string, len(blocks))
	for i, block := range blocks {
		allLines[i] = strings.Split(block, "\n")
	}

	var result strings.Builder
	for row := 0; row < maxHeight; row++ {
		for i, lines := range allLines {
			blockHeight := len(lines)
			var offset int
			switch align {
			case Middle:
				offset = (maxHeight - blockHeight) / 2
			case Bottom:
				offset = maxHeight - blockHeight
			}

			var line string
			if row >= offset && row < offset+blockHeight {
				line = lines[row-offset]
			}
			w := stringWidth(line)
			pad := widths[i] - w
			result.WriteString(line)
			if pad > 0 {
				result.WriteString(strings.Repeat(" ", pad))
			}
		}
		if row < maxHeight-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// JoinVertical joins blocks top-to-bottom, aligning them horizontally.
// Each block is a multi-line string. Blocks are padded to the widest block's
// width based on the alignment.
func JoinVertical(align HorizontalAlign, blocks ...string) string {
	if len(blocks) == 0 {
		return ""
	}
	if len(blocks) == 1 {
		return blocks[0]
	}

	// Find max width
	maxWidth := 0
	allLines := make([][]string, len(blocks))
	for i, block := range blocks {
		lines := strings.Split(block, "\n")
		allLines[i] = lines
		for _, line := range lines {
			w := stringWidth(line)
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	var result strings.Builder
	for i, lines := range allLines {
		for j, line := range lines {
			w := stringWidth(line)
			pad := maxWidth - w
			switch align {
			case Left:
				result.WriteString(line)
				if pad > 0 {
					result.WriteString(strings.Repeat(" ", pad))
				}
			case Center:
				leftPad := pad / 2
				if leftPad > 0 {
					result.WriteString(strings.Repeat(" ", leftPad))
				}
				result.WriteString(line)
				rightPad := pad - leftPad
				if rightPad > 0 {
					result.WriteString(strings.Repeat(" ", rightPad))
				}
			case Right:
				if pad > 0 {
					result.WriteString(strings.Repeat(" ", pad))
				}
				result.WriteString(line)
			}
			// Add newline between lines within a block
			if j < len(lines)-1 {
				result.WriteString("\n")
			}
		}
		// Add newline between blocks
		if i < len(allLines)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// Place positions a string block within a width x height area.
// The hAlign and vAlign determine where the content is placed.
func Place(width, height int, hAlign HorizontalAlign, vAlign VerticalAlign, content string) string {
	lines := strings.Split(content, "\n")
	contentWidth := 0
	for _, line := range lines {
		w := stringWidth(line)
		if w > contentWidth {
			contentWidth = w
		}
	}
	contentHeight := len(lines)

	// Calculate vertical offset
	vOffset := 0
	switch vAlign {
	case Middle:
		vOffset = (height - contentHeight) / 2
	case Bottom:
		vOffset = height - contentHeight
	}
	if vOffset < 0 {
		vOffset = 0
	}

	// Calculate horizontal padding per line
	var result strings.Builder
	for row := 0; row < height; row++ {
		var line string
		if row >= vOffset && row < vOffset+contentHeight {
			line = lines[row-vOffset]
		}

		lineWidth := stringWidth(line)
		hPad := width - lineWidth
		if hPad < 0 {
			hPad = 0
		}

		switch hAlign {
		case Left:
			result.WriteString(line)
			if hPad > 0 {
				result.WriteString(strings.Repeat(" ", hPad))
			}
		case Center:
			leftPad := hPad / 2
			if leftPad > 0 {
				result.WriteString(strings.Repeat(" ", leftPad))
			}
			result.WriteString(line)
			rightPad := hPad - leftPad
			if rightPad > 0 {
				result.WriteString(strings.Repeat(" ", rightPad))
			}
		case Right:
			if hPad > 0 {
				result.WriteString(strings.Repeat(" ", hPad))
			}
			result.WriteString(line)
		}
		if row < height-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// PlaceHorizontal places content horizontally within a given width.
func PlaceHorizontal(width int, align HorizontalAlign, content string) string {
	w := stringWidth(content)
	pad := width - w
	if pad <= 0 {
		return content
	}
	switch align {
	case Left:
		return content + strings.Repeat(" ", pad)
	case Center:
		left := pad / 2
		return strings.Repeat(" ", left) + content + strings.Repeat(" ", pad-left)
	case Right:
		return strings.Repeat(" ", pad) + content
	}
	return content
}

// PlaceVertical places content vertically within a given height.
func PlaceVertical(height int, align VerticalAlign, content string) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	pad := height - contentHeight
	if pad <= 0 {
		return content
	}

	var result strings.Builder
	switch align {
	case Top:
		result.WriteString(content)
		for i := 0; i < pad; i++ {
			result.WriteString("\n")
		}
	case Middle:
		top := pad / 2
		for i := 0; i < top; i++ {
			result.WriteString("\n")
		}
		result.WriteString(content)
		for i := 0; i < pad-top; i++ {
			result.WriteString("\n")
		}
	case Bottom:
		for i := 0; i < pad; i++ {
			result.WriteString("\n")
		}
		result.WriteString(content)
	}
	return result.String()
}

// Width returns the display width of a string (accounting for wide chars).
func Width(s string) int {
	return stringWidth(s)
}

// stringWidth computes the display width of a string using the buffer package.
func stringWidth(s string) int {
	return buffer.StringWidth(s)
}

// MaxWidth returns the maximum line width in a multi-line string.
func MaxWidth(s string) int {
	maxW := 0
	for _, line := range strings.Split(s, "\n") {
		w := stringWidth(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

// Height returns the number of lines in a string.
func Height(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}