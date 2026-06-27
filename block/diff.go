package block

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DiffType identifies the type of a unified diff line.
type DiffType uint8

const (
	DiffContext DiffType = iota // unchanged context line
	DiffAdd                     // added line (+)
	DiffDel                     // removed line (-)
	DiffHunk                    // hunk header (@@ ... @@)
	DiffFile                    // file header (diff --git)
	DiffMeta                    // metadata (index, ---, +++)
)

// DiffLine represents a single parsed line from a unified diff.
type DiffLine struct {
	Type    DiffType
	Content string
}

// diffColorAdd/Del/Hunk/File/Meta are resolved lazily from the active
// theme so that SetTheme() takes effect immediately.

// DetectDiff reports whether text looks like a unified diff.
// It checks for the canonical "diff --git" header or a hunk header
// combined with +/- lines.
func DetectDiff(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			return true
		}
	}
	// Also accept patches that start with hunk headers and have add/del lines
	hasHunk := false
	hasAddDel := false
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			hasHunk = true
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			hasAddDel = true
		}
	}
	return hasHunk && hasAddDel
}

// ParseDiff splits unified diff text into typed DiffLines.
// Non-diff text simply returns every line as DiffContext.
func ParseDiff(text string) []DiffLine {
	rawLines := strings.Split(text, "\n")
	result := make([]DiffLine, 0, len(rawLines))

	for _, line := range rawLines {
		result = append(result, DiffLine{
			Type:    classifyDiffLine(line),
			Content: line,
		})
	}
	return result
}

// classifyDiffLine determines the DiffType from the line prefix.
func classifyDiffLine(line string) DiffType {
	switch {
	case strings.HasPrefix(line, "diff --git"):
		return DiffFile
	case strings.HasPrefix(line, "@@"):
		return DiffHunk
	case strings.HasPrefix(line, "index "),
		strings.HasPrefix(line, "---"),
		strings.HasPrefix(line, "+++"):
		return DiffMeta
	case strings.HasPrefix(line, "+"):
		return DiffAdd
	case strings.HasPrefix(line, "-"):
		return DiffDel
	default:
		return DiffContext
	}
}

// DiffStyle returns the buffer.Style for a given DiffType.
//   - DiffAdd:   green foreground
//   - DiffDel:   red foreground
//   - DiffHunk:  cyan foreground
//   - DiffFile:  purple foreground, Bold
//   - DiffMeta:  dim gray-blue foreground, Dim flag
//   - DiffContext: default (no color)
func DiffStyle(dt DiffType) buffer.Style {
	t := theme.Get()
	switch dt {
	case DiffAdd:
		return buffer.Style{Fg: t.DiffAdd}
	case DiffDel:
		return buffer.Style{Fg: t.DiffDel}
	case DiffHunk:
		return buffer.Style{Fg: t.DiffHunk}
	case DiffFile:
		return buffer.Style{Fg: t.DiffFile, Flags: buffer.Bold}
	case DiffMeta:
		return buffer.Style{Fg: t.DiffMeta, Flags: buffer.Dim}
	default:
		return buffer.DefaultStyle
	}
}
