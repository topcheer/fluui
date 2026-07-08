package markdown

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	ast "github.com/yuin/goldmark/ast"
	text "github.com/yuin/goldmark/text"
)

// BlockType identifies the kind of markdown block.
type BlockType uint8

const (
	BlockHeading    BlockType = iota
	BlockParagraph
	BlockList
	BlockCodeBlock
	BlockQuote
	BlockThematicBreak
	BlockTable
	BlockImage
)

// Block is a rendered markdown block.
type Block struct {
	Type  BlockType
	Level int             // heading level 1-6 (0 for non-headings)
	Cells [][]buffer.Cell // rendered lines
}

// MarkdownRenderer converts markdown source to styled Cell lines grouped into Blocks.
type MarkdownRenderer struct {
	theme        *MarkdownTheme
	width        int
	highlighter  *Highlighter
	md           goldmark.Markdown // parser with extensions
	linkRenderer *LinkRenderer
}

// NewMarkdownRenderer creates a renderer with the given theme and wrap width.
func NewMarkdownRenderer(theme *MarkdownTheme, width int) *MarkdownRenderer {
	if theme == nil {
		theme = DefaultTheme()
	}
	return &MarkdownRenderer{
		theme:        theme,
		width:        width,
		md:           goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough)),
		linkRenderer: NewLinkRenderer(false), // OSC8 disabled by default
	}
}

// SetLinkRenderer configures hyperlink rendering. Pass a LinkRenderer with
// OSC8 enabled to make links clickable in supporting terminals.
func (r *MarkdownRenderer) SetLinkRenderer(lr *LinkRenderer) {
	r.linkRenderer = lr
}

// SetHighlighter attaches a code highlighter for fenced code blocks.
func (r *MarkdownRenderer) SetHighlighter(h *Highlighter) {
	r.highlighter = h
}

// Render parses the markdown source and returns a slice of rendered Blocks.
func (r *MarkdownRenderer) Render(source string) ([]*Block, error) {
	// Use unsafe to get a read-only byte slice that shares the string's
	// backing array, avoiding a full copy of the source (56KB+ for large docs).
	// This is safe because goldmark's parser and all renderBlock methods only
	// READ from source bytes — they never modify them.
	srcBytes := unsafe.Slice(unsafe.StringData(source), len(source))
	reader := text.NewReader(srcBytes)
	doc := r.md.Parser().Parse(reader)

	// Pre-count top-level children for capacity hint.
	nBlocks := 0
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		nBlocks++
	}
	blocks := make([]*Block, 0, nBlocks)
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if blk := r.renderBlock(child, srcBytes); blk != nil {
			blocks = append(blocks, blk)
		}
	}
	return blocks, nil
}

// renderBlock dispatches to the appropriate renderer for a top-level block node.
func (r *MarkdownRenderer) renderBlock(node ast.Node, source []byte) *Block {
	switch n := node.(type) {
	case *ast.Heading:
		return r.renderHeading(n, source)
	case *ast.Paragraph:
		return r.renderParagraph(n, source)
	case *ast.List:
		return r.renderList(n, source)
	case *ast.FencedCodeBlock:
		return r.renderFencedCode(n, source)
	case *ast.CodeBlock:
		return r.renderCodeBlock(n, source)
	case *ast.Blockquote:
		return r.renderBlockquote(n, source)
	case *ast.ThematicBreak:
		return r.renderThematicBreak()
	case *extast.Table:
		return r.renderTable(n, source)
	case *ast.HTMLBlock:
		// Render raw HTML as plain text
		text := string(n.Text(source))
		lines := WrapText(text, r.width)
		var cells [][]buffer.Cell
		for _, line := range lines {
			cells = append(cells, r.textToCells(line, r.theme.Body, 0))
		}
		return &Block{Type: BlockParagraph, Cells: cells}
	}
	return nil
}

// renderHeading renders a heading node with appropriate color and bold.
func (r *MarkdownRenderer) renderHeading(n *ast.Heading, source []byte) *Block {
	text := string(n.Text(source))
	lines := WrapText(text, r.width)
	fg := r.theme.headingColor(n.Level)

	var cells [][]buffer.Cell
	for _, line := range lines {
		cells = append(cells, r.textToCells(line, fg, buffer.Bold))
	}
	return &Block{Type: BlockHeading, Level: n.Level, Cells: cells}
}

// renderInline renders inline content (text, emphasis, code, links) into cells.
// Uses recursive descent instead of ast.Walk to avoid double-processing.
func (r *MarkdownRenderer) renderInline(n ast.Node, source []byte) []buffer.Cell {
	// Fast path: single text child — return its cells directly, avoiding
	// an undersized intermediate slice that triggers 10+ reallocations
	// for large paragraphs (common in AI streaming).
	first := n.FirstChild()
	if first != nil && first.NextSibling() == nil {
		if t, ok := first.(*ast.Text); ok {
			text := string(t.Value(source))
			if HasInlineMath(text) {
				var sb strings.Builder
				sb.Grow(len(text) * 2)
				renderInlineMathToBuilder(text, &sb)
				return r.textToCells(sb.String(), r.theme.CodeFg, buffer.Italic)
			}
			return r.textToCells(text, r.theme.Body, 0)
		}
	}

	// Multi-child path: estimate capacity and concatenate.
	nChildren := 0
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		nChildren++
	}
	inline := make([]buffer.Cell, 0, nChildren*32)
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		inline = append(inline, r.renderInlineNode(child, source)...)
	}
	return inline
}

// renderInlineNode handles a single inline node.
func (r *MarkdownRenderer) renderInlineNode(n ast.Node, source []byte) []buffer.Cell {
	switch v := n.(type) {
	case *ast.Text:
		text := string(v.Value(source))
		// Inline math: $...$ or \(...\)
		if HasInlineMath(text) {
			converted := RenderInlineMath(text)
			return r.textToCells(converted, r.theme.CodeFg, buffer.Italic)
		}
		return r.textToCells(text, r.theme.Body, 0)
	case *ast.String:
		return r.textToCells(string(v.Value), r.theme.Body, 0)
	case *ast.CodeSpan:
		// Render all text children as code-styled
		var cells []buffer.Cell
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				cells = append(cells, r.textToCells(string(t.Value(source)), r.theme.CodeFg, 0)...)
			}
		}
		return cells
	case *ast.Link:
		// Render link text with link style
		url := string(v.Destination)
		var cells []buffer.Cell
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			cells = append(cells, r.renderInlineNode(child, source)...)
		}
		// Override style to link
		for i := range cells {
			cells[i].Fg = r.theme.LinkFg
			cells[i].Flags = buffer.Underline
			// Attach OSC8 hyperlink metadata when enabled
			if r.linkRenderer != nil && r.linkRenderer.Enabled() && url != "" {
				cells[i].Link = &buffer.Link{URL: url, Text: string(cells[i].Rune)}
			}
		}
		// When OSC8 is disabled and a URL exists, append " (url)" as plain text
		if r.linkRenderer != nil && !r.linkRenderer.Enabled() && url != "" {
			fallback := " (" + url + ")"
			cells = append(cells, r.textToCells(fallback, r.theme.LinkFg, 0)...)
		}
		return cells
	case *ast.Image:
		// Render image as "[image: alt] (url)" with image style.
		// Terminal can't display inline images in markdown flow, so we
		// show a styled indicator with alt text and the URL.
		url := string(v.Destination)
		altText := ""
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				if altText != "" {
					altText += " "
				}
				altText += string(t.Value(source))
			}
		}
		// Build display: [image: alt] (url)
		var cells []buffer.Cell
		if altText != "" {
			cells = append(cells, r.textToCells("[image: "+altText+"]", r.theme.ImageFg, buffer.Dim)...)
		} else {
			cells = append(cells, r.textToCells("[image]", r.theme.ImageFg, buffer.Dim)...)
		}
		if url != "" {
			cells = append(cells, r.textToCells(" ("+url+")", r.theme.LinkUrlFg, 0)...)
			// Attach OSC8 link so URL is clickable
			if r.linkRenderer != nil && r.linkRenderer.Enabled() {
				for i := range cells {
					cells[i].Link = &buffer.Link{URL: url, Text: string(cells[i].Rune)}
				}
			}
		}
		return cells
	case *ast.Emphasis:
		// Render children with bold/italic flag
		var cells []buffer.Cell
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			cells = append(cells, r.renderInlineNode(child, source)...)
		}
		flag := buffer.Italic
		if v.Level == 2 {
			flag = buffer.Bold
		}
		for i := range cells {
			cells[i].Flags |= flag
		}
		return cells
	case *extast.Strikethrough:
		// Render children with strikethrough flag
		var cells []buffer.Cell
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			cells = append(cells, r.renderInlineNode(child, source)...)
		}
		for i := range cells {
			cells[i].Flags |= buffer.Strikethrough
		}
		return cells
	default:
		// Fallback: try to get text content
		text := string(n.Text(source))
		if text != "" {
			return r.textToCells(text, r.theme.Body, 0)
		}
		return nil
	}
}

// renderParagraph renders a paragraph node with wrapping.
func (r *MarkdownRenderer) renderParagraph(n ast.Node, source []byte) *Block {
	inline := r.renderInline(n, source)
	cells := r.wrapCells(inline, r.width)
	return &Block{Type: BlockParagraph, Cells: cells}
}

// renderList renders a list with bullet/number prefixes.
func (r *MarkdownRenderer) renderList(n *ast.List, source []byte) *Block {
	var cells [][]buffer.Cell
	ordered := n.Marker != '-' && n.Marker != '+' && n.Marker != '*'
	index := n.Start
	if index == 0 {
		index = 1
	}

	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		var prefix string
		if ordered {
			prefix = fmt.Sprintf("%d. ", index)
		} else {
			// Check for GitHub-style task list: "- [ ] text" or "- [x] text"
			taskChar, isTask := detectTaskListItem(item, source)
			if isTask {
				if taskChar == ' ' {
					prefix = "\u2610 " // ☐ ballot box (unchecked)
				} else {
					prefix = "\u2611 " // ☑ ballot box with check (checked)
				}
			} else {
				prefix = "\u2022 " // bullet •
			}
		}
		index++

		// For task list items, strip the "[ ]" or "[x]" prefix from content
		itemSource := source
		stripTask := !ordered && isTaskListItem(item, source)

		// Render item content
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			switch v := child.(type) {
			case *ast.Paragraph:
				inline := r.renderInline(v, itemSource)
				if stripTask {
					inline = stripTaskPrefix(inline)
				}
				wrapped := r.wrapCellsWithPrefix(inline, prefix, "   ", r.width)
				cells = append(cells, wrapped...)
			case *ast.TextBlock:
				inline := r.renderInline(v, itemSource)
				if stripTask {
					inline = stripTaskPrefix(inline)
				}
				wrapped := r.wrapCellsWithPrefix(inline, prefix, "   ", r.width)
				cells = append(cells, wrapped...)
			default:
				if blk := r.renderBlock(child, source); blk != nil {
					cells = append(cells, blk.Cells...)
				}
			}
		}
	}

	return &Block{Type: BlockList, Cells: cells}
}

// wrapCellsWithPrefix wraps inline cells with a prefix on the first line
// and indentation on subsequent lines.
//
// Zero-copy optimization: Instead of allocating a combined allCells slice
// (which copies ALL cells), we wrap the content cells first, then prepend
// the prefix cells to the first line and continuation cells to subsequent
// lines. This avoids a full-slice copy for every paragraph/list item.
func (r *MarkdownRenderer) wrapCellsWithPrefix(cells []buffer.Cell, firstPrefix, contPrefix string, width int) [][]buffer.Cell {
	prefixCells := r.textToCells(firstPrefix, r.theme.Body, 0)
	contCells := r.textToCells(contPrefix, r.theme.Body, 0)

	// Compute effective wrapping width minus prefix
	prefixW := cellLineWidth(prefixCells)
	effWidth := width - prefixW
	if effWidth <= 0 {
		effWidth = width
	}

	// Wrap content without prefix (zero-copy into input cells)
	wrapped := r.wrapCells(cells, effWidth)

	// Prepend prefix to first line, continuation prefix to subsequent lines
	result := make([][]buffer.Cell, len(wrapped))
	for i, line := range wrapped {
		if i == 0 {
			result[i] = append(prefixCells, line...)
		} else {
			result[i] = append(contCells, line...)
		}
	}
	return result
}

// renderTable renders a markdown table with borders.
func (r *MarkdownRenderer) renderTable(n *extast.Table, source []byte) *Block {
	// First pass: collect all cell texts to compute column widths
	var rows [][]string
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		var rowCells []string
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			rowCells = append(rowCells, string(cell.Text(source)))
		}
		rows = append(rows, rowCells)
	}
	if len(rows) == 0 {
		return &Block{Type: BlockTable}
	}

	numCols := len(rows[0])
	colWidths := make([]int, numCols)
	for _, row := range rows {
		for i, text := range row {
			w := StringWidth(text)
			if i < numCols && w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	// Ensure table fits within width — scale down columns proportionally
	totalW := numCols + 1 // borders
	for _, w := range colWidths {
		totalW += w + 2 // 2 padding spaces per column
	}
	if totalW > r.width && r.width > 0 {
		// Reduce columns to fit
		availW := r.width - numCols - 1
		perCol := availW / numCols
		for i := range colWidths {
			if colWidths[i] > perCol {
				colWidths[i] = perCol
			}
		}
	}

	borderStyle := buffer.Style{Fg: r.theme.QuoteBar, Flags: 0}
	headerStyle := buffer.Style{Fg: r.theme.headingColor(4), Flags: buffer.Bold}
	cellStyle := buffer.Style{Fg: r.theme.Body, Flags: 0}

	var result [][]buffer.Cell

	// Top border: ┌─────┬─────┐
	topBorder := r.makeTableBorder('┌', '┬', '┐', colWidths, borderStyle)
	result = append(result, topBorder)

	for rowIdx, row := range rows {
		var cells []buffer.Cell
		for colIdx := 0; colIdx < numCols; colIdx++ {
			// Left border
			cells = append(cells, buffer.Cell{Rune: '│', Width: 1, Fg: borderStyle.Fg})
			cells = append(cells, r.textToCells(" ", r.theme.Body, 0)...)

			text := ""
			if colIdx < len(row) {
				text = row[colIdx]
			}
			// Truncate to col width
			textRunes := []rune(text)
			if len(textRunes) > colWidths[colIdx] {
				textRunes = textRunes[:colWidths[colIdx]]
			}
			textStr := string(textRunes)

			var style buffer.Style
			if rowIdx == 0 {
				style = headerStyle
			} else {
				style = cellStyle
			}

			// Compute left/right padding based on column alignment.
			padW := colWidths[colIdx] - len(textRunes)
			leftPad, rightPad := 0, padW // default: left-aligned
			if colIdx < len(n.Alignments) {
				switch n.Alignments[colIdx] {
				case extast.AlignCenter:
					leftPad = padW / 2
					rightPad = padW - leftPad
				case extast.AlignRight:
					leftPad = padW
					rightPad = 0
				}
			}

			// Left padding
			for j := 0; j < leftPad; j++ {
				cells = append(cells, buffer.Cell{Rune: ' ', Width: 1, Fg: style.Fg})
			}
			// Text
			cells = append(cells, r.textToCells(textStr, style.Fg, style.Flags)...)
			// Right padding
			for j := 0; j < rightPad; j++ {
				cells = append(cells, buffer.Cell{Rune: ' ', Width: 1, Fg: style.Fg})
			}
			cells = append(cells, r.textToCells(" ", r.theme.Body, 0)...)
		}
		// Right border
		cells = append(cells, buffer.Cell{Rune: '│', Width: 1, Fg: borderStyle.Fg})
		result = append(result, cells)

		// After header row: separator ┝━━━━━┿━━━━━┥ or ┝─────┼─────┥
		if rowIdx == 0 {
			sep := r.makeTableBorder('┝', '┿', '┥', colWidths, borderStyle)
			// Use heavy dashes for header separator
			for i := range sep {
				if sep[i].Rune == '─' {
					sep[i].Rune = '━'
				}
			}
			result = append(result, sep)
		}
	}

	// Bottom border: └─────┴─────┘
	botBorder := r.makeTableBorder('└', '┴', '┘', colWidths, borderStyle)
	result = append(result, botBorder)

	return &Block{Type: BlockTable, Cells: result}
}

// makeTableBorder generates a horizontal border line for a table.
func (r *MarkdownRenderer) makeTableBorder(left, mid, right rune, colWidths []int, style buffer.Style) []buffer.Cell {
	// Pre-compute exact capacity: left border + (colWidth+2 padding) per column
	// + separator between columns + right border.
	totalW := 2 // left + right
	for _, w := range colWidths {
		totalW += w + 2 // content + padding
	}
	totalW += len(colWidths) - 1 // separators
	if totalW < 1 {
		totalW = 1
	}
	cells := make([]buffer.Cell, 0, totalW)
	cells = append(cells, buffer.Cell{Rune: left, Width: 1, Fg: style.Fg})
	for i, w := range colWidths {
		for j := 0; j < w+2; j++ {
			cells = append(cells, buffer.Cell{Rune: '─', Width: 1, Fg: style.Fg})
		}
		if i < len(colWidths)-1 {
			cells = append(cells, buffer.Cell{Rune: mid, Width: 1, Fg: style.Fg})
		}
	}
	cells = append(cells, buffer.Cell{Rune: right, Width: 1, Fg: style.Fg})
	return cells
}

// renderFencedCode renders a fenced code block, optionally highlighting.
// Mermaid code blocks (```mermaid) are rendered as ASCII art diagrams.
func (r *MarkdownRenderer) renderFencedCode(n *ast.FencedCodeBlock, source []byte) *Block {
	code := extractCodeBlockText(n, source)
	lang := string(n.Language(source))

	// Mermaid diagram rendering
	if lang == "mermaid" {
		mermaidCells := RenderMermaidText(code, r.theme)
		if mermaidCells != nil {
			return &Block{Type: BlockCodeBlock, Cells: mermaidCells}
		}
		// Fall through to plain rendering if Mermaid parsing fails
	}

	// LaTeX math block rendering
	if lang == "math" || lang == "latex" {
		mathCells := RenderMathToCells(strings.TrimSpace(code), r.theme.CodeFg)
		return &Block{Type: BlockCodeBlock, Cells: [][]buffer.Cell{mathCells}}
	}

	var cells [][]buffer.Cell
	if r.highlighter != nil && lang != "" {
		var err error
		cells, err = r.highlighter.Highlight(code, lang)
		if err != nil {
			cells = nil // fall back to plain
		}
	}

	if cells == nil {
		for _, line := range strings.Split(code, "\n") {
			cells = append(cells, r.textToCells(line, r.theme.CodeFg, 0))
		}
		// Trim trailing empty line from final newline
		if len(cells) > 0 && len(cells[len(cells)-1]) == 0 {
			cells = cells[:len(cells)-1]
		}
	}

	return &Block{Type: BlockCodeBlock, Cells: cells}
}

// renderCodeBlock renders an indented code block (no language).
func (r *MarkdownRenderer) renderCodeBlock(n *ast.CodeBlock, source []byte) *Block {
	code := extractCodeBlockText(n, source)
	var cells [][]buffer.Cell
	for _, line := range strings.Split(code, "\n") {
		cells = append(cells, r.textToCells(line, r.theme.CodeFg, 0))
	}
	if len(cells) > 0 && len(cells[len(cells)-1]) == 0 {
		cells = cells[:len(cells)-1]
	}
	return &Block{Type: BlockCodeBlock, Cells: cells}
}

// renderBlockquote renders a blockquote with │ prefix and dim color.
func (r *MarkdownRenderer) renderBlockquote(n *ast.Blockquote, source []byte) *Block {
	var cells [][]buffer.Cell
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if blk := r.renderBlock(child, source); blk != nil {
			for _, line := range blk.Cells {
				barCell := buffer.Cell{Rune: '\u2502', Width: 1, Fg: r.theme.QuoteBar} // │
				spacing := buffer.Cell{Rune: ' ', Width: 1}
				line = append([]buffer.Cell{barCell, spacing}, line...)
				cells = append(cells, line)
			}
		}
	}
	return &Block{Type: BlockQuote, Cells: cells}
}

// renderThematicBreak renders a horizontal rule.
func (r *MarkdownRenderer) renderThematicBreak() *Block {
	w := r.width
	if w <= 0 {
		w = 80
	}
	line := make([]buffer.Cell, 0, w)
	for i := 0; i < w; i++ {
		line = append(line, buffer.Cell{Rune: '\u2500', Width: 1, Fg: r.theme.Hr}) // ─
	}
	return &Block{Type: BlockThematicBreak, Cells: [][]buffer.Cell{line}}
}

// --- helpers ---

// textToCells converts a string to a slice of Cells with uniform style.
func (r *MarkdownRenderer) textToCells(s string, fg buffer.Color, flags buffer.StyleFlags) []buffer.Cell {
	// Fast path: pure ASCII text (the common case for AI output).
	// For ASCII, len(s) == rune count and every rune has width 1,
	// so we skip the counting loop and RuneWidth calls entirely.
	if isAllASCII(s) {
		cells := make([]buffer.Cell, len(s))
		for i := 0; i < len(s); i++ {
			cells[i] = buffer.Cell{
				Rune:  rune(s[i]),
				Width: 1,
				Fg:    fg,
				Flags: flags,
			}
		}
		return cells
	}

	// Non-ASCII path: count runes for exact capacity, then fill.
	runeCount := 0
	for range s {
		runeCount++
	}
	cells := make([]buffer.Cell, 0, runeCount)
	for _, ch := range s {
		cells = append(cells, buffer.Cell{
			Rune:  ch,
			Width: uint8(buffer.RuneWidth(ch)),
			Fg:    fg,
			Flags: flags,
		})
	}
	return cells
}

// isAllASCII returns true if every byte in s is < 0x80.
// This is inlined by the compiler and vectorized on most architectures.
func isAllASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// wrapCells wraps a flat cell slice into lines of at most width display columns.
//
// Zero-copy strategy: instead of allocating a separate slab and copying cells,
// we create cap-limited sub-slices directly into the input cells slice. This
// eliminates the biggest single allocation in the render pipeline (~2.7MB for
// 56K chars).
//
// Cap-limited slicing ([start:end:end]) ensures that appending to any returned
// line triggers a fresh backing array (Go copy-on-write), preserving correctness
// even if callers mutate individual lines.
func (r *MarkdownRenderer) wrapCells(cells []buffer.Cell, width int) [][]buffer.Cell {
	if width <= 0 {
		return [][]buffer.Cell{cells}
	}
	if len(cells) == 0 {
		return [][]buffer.Cell{{}}
	}

	// Estimate line count for pre-allocation (avoids grow-and-copy).
	estLines := len(cells)/width + 1
	lines := make([][]buffer.Cell, 0, estLines)

	lineStart := 0 // start index in cells[] for the current line
	curWidth := 0   // display width of current line so far

	for i := 0; i < len(cells); i++ {
		c := &cells[i]
		if c.Width == 0 {
			continue // skip 0-width cells (combining marks, wide-char padding)
		}

		if curWidth+int(c.Width) > width && curWidth > 0 {
			// Need to wrap before this cell.
			if spaceIdx, afterSpaceW := lastSpaceCellAndWidth(cells[lineStart:i]); spaceIdx >= 0 {
				// Word-wrap: line ends before the space.
				end := lineStart + spaceIdx
				lines = append(lines, cells[lineStart:end:end])
				lineStart = end + 1 // skip the space cell
				curWidth = afterSpaceW
			} else {
				// Hard break: no space found.
				lines = append(lines, cells[lineStart:i:i])
				lineStart = i
				curWidth = 0
			}
		}
		curWidth += int(c.Width)
	}

	// Final line: remaining cells.
	if lineStart < len(cells) {
		end := len(cells)
		lines = append(lines, cells[lineStart:end:end])
	}

	if len(lines) == 0 {
		return [][]buffer.Cell{cells}
	}
	return lines
}

// lastSpaceCellAndWidth finds the last space cell and computes the total width
// of all cells after it in a single backward pass, avoiding a separate
// cellLineWidth call after word-wrap.
func lastSpaceCellAndWidth(cells []buffer.Cell) (int, int) {
	// Scan backward to find the last space while accumulating width of cells
	// encountered so far (these are the cells after the space).
	afterWidth := 0
	for i := len(cells) - 1; i >= 0; i-- {
		if cells[i].Rune == ' ' {
			return i, afterWidth
		}
		afterWidth += int(cells[i].Width)
	}
	return -1, 0
}

func cellLineWidth(cells []buffer.Cell) int {
	w := 0
	for _, c := range cells {
		w += int(c.Width)
	}
	return w
}

// extractCodeBlockText extracts all text lines from a code block node.
func extractCodeBlockText(node ast.Node, source []byte) string {
	var sb strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		sb.Write(seg.Value(source))
	}
	return sb.String()
}

// isTaskListItem returns true if the list item's first text starts with "[ ]" or "[x]"/"[X]".
func isTaskListItem(item ast.Node, source []byte) bool {
	_, ok := detectTaskListItem(item, source)
	return ok
}

// detectTaskListItem returns the checkbox character (' ' or 'x') and true if
// the list item's first child text starts with "[ ]", "[x]", or "[X]".
func detectTaskListItem(item ast.Node, source []byte) (byte, bool) {
	first := item.FirstChild()
	if first == nil {
		return 0, false
	}
	text := string(first.Text(source))
	if len(text) < 3 {
		return 0, false
	}
	if text[0] == '[' && text[2] == ']' {
		c := text[1]
		if c == ' ' || c == 'x' || c == 'X' {
			return c, true
		}
	}
	return 0, false
}

// stripTaskPrefix removes the leading "[ ] " or "[x] " cells (3 cells: '[', char, ']')
// plus the following space, returning cells starting from the actual content.
func stripTaskPrefix(cells []buffer.Cell) []buffer.Cell {
	// Pattern: [ <char> ] <space> <content...>
	// Need at least 5 cells: '[', char, ']', ' ', content
	if len(cells) < 4 {
		return cells
	}
	if cells[0].Rune != '[' || cells[2].Rune != ']' {
		return cells
	}
	// Skip "[x] " (4 cells including the space after ])
	if len(cells) >= 5 && cells[3].Rune == ' ' {
		return cells[4:]
	}
	// Skip "[x]" if no trailing space
	return cells[3:]
}
