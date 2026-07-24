package markdown_test

import (
	"fmt"

	"github.com/topcheer/fluui/markdown"
)

// ExampleNewMarkdownRenderer demonstrates basic markdown rendering.
func ExampleNewMarkdownRenderer() {
	r := markdown.NewMarkdownRenderer(nil, 80) // default theme, 80 cols
	blocks, err := r.Render("# Hello\n\n**Bold** text with `code`.")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Blocks: %d\n", len(blocks))
	// Output: Blocks: 2
}

// ExampleStreamingRenderer demonstrates incremental AI markdown rendering.
func ExampleStreamingRenderer() {
	r := markdown.NewMarkdownRenderer(nil, 60)
	sr := markdown.NewStreamingRenderer(r, 4) // render every 4 deltas

	// Simulate streaming AI response
	sr.AppendDelta("# Title")
	sr.AppendDelta("\n\n")
	sr.AppendDelta("Content here")
	sr.Flush()

	blocks, _ := sr.Blocks()
	fmt.Printf("Rendered blocks: %d\n", len(blocks))
}

// ExampleDetectLanguage demonstrates code language detection.
func ExampleDetectLanguage() {
	lang := markdown.DetectLanguage("go")
	fmt.Println(lang)
	// Output: go
}

// ExampleNewHighlighter demonstrates code syntax highlighting.
func ExampleNewHighlighter() {
	h := markdown.NewHighlighter()
	lines, err := h.HighlightToLines("fmt.Println(\"hello\")", "go")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Highlighted lines: %d\n", len(lines))
	// Output: Highlighted lines: 1
}
