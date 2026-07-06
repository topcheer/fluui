package markdown

import (
	"fmt"
	"strings"

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
		md:           goldmark.New(goldmark.WithExtensions(extension.Table)),
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
	// Convert source to []byte ONCE — avoid repeated string→bytes copy per block.
	srcBytes := []byte(source)
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
				converted := RenderInlineMath(text)
				return r.textToCells(converted, r.theme.CodeFg, buffer.Italic)
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
			prefix = "\u2022 " // bullet •
		}
		index++

		// Render item content
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			switch v := child.(type) {
			case *ast.Paragraph:
				inline := r.renderInline(v, source)
				wrapped := r.wrapCellsWithPrefix(inline, prefix, "   ", r.width)
				cells = append(cells, wrapped...)
			case *ast.TextBlock:
				inline := r.renderInline(v, source)
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
func (r *MarkdownRenderer) wrapCellsWithPrefix(cells []buffer.Cell, firstPrefix, contPrefix string, width int) [][]buffer.Cell {
	prefixCells := r.textToCells(firstPrefix, r.theme.Body, 0)
	contCells := r.textToCells(contPrefix, r.theme.Body, 0)

	allCells := append(prefixCells, cells...)
	wrapped := r.wrapCells(allCells, width)

	// Prepend continuation prefix to lines after the first
	for i := 1; i < len(wrapped); i++ {
		wrapped[i] = append(contCells, wrapped[i]...)
	}
	return wrapped
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
	var cells []buffer.Cell
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
	// Count runes first to avoid over-allocation for multi-byte UTF-8.
	// len(s) over-allocates for non-ASCII text (e.g., 4x for emoji).
	runeCount := 0
	for range s {
		runeCount++
	}
	cells := make([]buffer.Cell, 0, runeCount)
	for _, ch := range s {
		cells = append(cells, buffer.Cell{
			Rune:  ch,
			Width: buffer.RuneWidth(ch),
			Fg:    fg,
			Flags: flags,
		})
	}
	return cells
}

// wrapCells wraps a flat cell slice into lines of at most width display columns.
// Uses a reusable line buffer to minimize allocations: each line produces exactly
// one heap allocation (the copy into lines), instead of 4-8 from append growth.
func (r *MarkdownRenderer) wrapCells(cells []buffer.Cell, width int) [][]buffer.Cell {
	if width <= 0 {
		return [][]buffer.Cell{cells}
	}
	lines := make([][]buffer.Cell, 0, len(cells)/width+1)
	// Pre-allocate line buffer with width capacity — lines never exceed width
	// columns, so this buffer never needs to grow.
	lineBuf := make([]buffer.Cell, 0, width)
	curWidth := 0

	for _, c := range cells {
		if c.Width == 0 {
			lineBuf = append(lineBuf, c)
			continue
		}

		if curWidth+c.Width > width && curWidth > 0 {
			if spaceIdx := lastSpaceCell(lineBuf); spaceIdx >= 0 {
				// Word-wrap: copy content up to space into a fresh slice.
				line := make([]buffer.Cell, spaceIdx)
				copy(line, lineBuf[:spaceIdx])
				lines = append(lines, line)
				// Move remaining content to front of lineBuf (reuse buffer).
				remaining := lineBuf[spaceIdx+1:]
				copy(lineBuf, remaining)
				lineBuf = lineBuf[:len(remaining)]
				curWidth = cellLineWidth(lineBuf)
			} else {
				// Hard break: no space found.
				line := make([]buffer.Cell, len(lineBuf))
				copy(line, lineBuf)
				lines = append(lines, line)
				lineBuf = lineBuf[:0]
				curWidth = 0
			}
		}

		lineBuf = append(lineBuf, c)
		curWidth += c.Width
	}

	if len(lineBuf) > 0 {
		// Copy final line into its own slice.
		line := make([]buffer.Cell, len(lineBuf))
		copy(line, lineBuf)
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, []buffer.Cell{})
	}
	return lines
}

func lastSpaceCell(cells []buffer.Cell) int {
	for i := len(cells) - 1; i >= 0; i-- {
		if cells[i].Rune == ' ' {
			return i
		}
	}
	return -1
}

func cellLineWidth(cells []buffer.Cell) int {
	w := 0
	for _, c := range cells {
		w += c.Width
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
