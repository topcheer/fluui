package markdown

import (
	"testing"
)

// BenchmarkHighlight_GoCode measures the syntax highlighting hot path
// for typical Go code (all ASCII).
func BenchmarkHighlight_GoCode(b *testing.B) {
	h := NewHighlighter()
	source := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	for i := 0; i < 10; i++ {
		fmt.Printf("Count: %d\n", i)
	}
}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h.Highlight(source, "go")
	}
}
