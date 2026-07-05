// Package main implements demo16 — Streaming Code Highlighting.
//
// Demonstrates Fluui's AI-native streaming code highlighting: code is
// rendered token-by-token with real-time syntax highlighting, auto-scrolling,
// and a streaming cursor indicator.
//
// Usage: go run ./cmd/demo16/
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// sampleGoCode is streamed token-by-token to simulate an AI response.
const sampleGoCode = `package main

import (
	"fmt"
	"strings"
)

// Handler processes incoming requests and returns formatted output.
type Handler struct {
	name    string
	counter int
}

// Process takes a message, transforms it, and returns the result.
func (h *Handler) Process(msg string) string {
	h.counter++
	parts := strings.Split(msg, " ")
	result := strings.Join(parts, "_")
	return fmt.Sprintf("[%s:%d] %s", h.name, h.counter, result)
}

func main() {
	h := &Handler{name: "demo"}
	words := []string{"hello", "world", "streaming", "code"}
	for _, w := range words {
		fmt.Println(h.Process(w))
	}
}`

// samplePythonCode shows multi-language support.
const samplePythonCode = `class DataProcessor:
    """Process and transform data streams."""

    def __init__(self, batch_size=100):
        self.batch_size = batch_size
        self.buffer = []

    def add(self, item):
        self.buffer.append(item)
        if len(self.buffer) >= self.batch_size:
            return self.flush()
        return None

    def flush(self):
        result = sorted(self.buffer)
        self.buffer.clear()
        return result`

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 16 — Streaming Code Highlighting")
		fmt.Println("Usage: go run ./cmd/demo16/")
		return
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║       Fluui Demo 16 — Streaming Code Highlighting                ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  Simulating AI streaming code token-by-token with real-time")
	fmt.Println("  syntax highlighting, auto-scroll, and cursor indicator.")
	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────────────────────")
	fmt.Println()

	// Demo 1: Stream Go code
	fmt.Println("  ▸ Demo 1: Streaming Go code")
	fmt.Println()
	streamCode("go", "main.go", sampleGoCode, 60, 20)

	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────────────────────")
	fmt.Println()

	// Demo 2: Stream Python code
	fmt.Println("  ▸ Demo 2: Streaming Python code")
	fmt.Println()
	streamCode("python", "processor.py", samplePythonCode, 60, 15)

	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  Streaming features demonstrated:")
	fmt.Println("    • Real-time syntax highlighting (chroma-based, 300+ languages)")
	fmt.Println("    • Auto-scroll to latest content during streaming")
	fmt.Println("    • Streaming cursor indicator (pink block cursor)")
	fmt.Println("    • Line numbers and title bar")
	fmt.Println("    • Thread-safe concurrent streaming support")
	fmt.Println()
}

// streamCode simulates AI streaming by appending tokens one at a time
// and rendering the CodeBlock state at key points.
func streamCode(language, title, code string, width, height int) {
	cb := component.NewCodeBlock(language, "")
	cb.SetShowTitle(true)
	cb.SetShowLineNumbers(true)
	cb.SetTitle(title)
	cb.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: height})
	cb.SetStreaming(true)

	// Split code into tokens (simulating word-by-word streaming)
	tokens := strings.Fields(code)

	// Show initial empty state with streaming cursor
	fmt.Println("  [streaming started]")
	printCodeBlock(cb, width, height)
	fmt.Println()

	// Stream tokens
	step := len(tokens) / 4
	if step == 0 {
		step = 1
	}
	for i := 0; i < len(tokens); i++ {
		// Append token + space
		delta := tokens[i]
		if i < len(tokens)-1 {
			delta += " "
		}
		// Handle newlines in the original code
		if strings.Contains(tokens[i], "\n") {
			parts := strings.SplitN(tokens[i], "\n", 2)
			delta = parts[0] + "\n"
			if len(parts) > 1 {
				delta += parts[1] + " "
			}
		}
		cb.AppendSource(delta)

		// Show progress at 25%, 50%, 75%
		if (i+1)%step == 0 {
			pct := (i + 1) * 100 / len(tokens)
			fmt.Printf("  [streaming... %d%%]\n", pct)
			printCodeBlock(cb, width, height)
			fmt.Println()
		}
	}

	// Final state
	cb.FinishStreaming()
	fmt.Println("  [streaming complete]")
	printCodeBlock(cb, width, height)

	// Small delay for visual effect
	time.Sleep(100 * time.Millisecond)
}

// printCodeBlock renders the current state of the CodeBlock to stdout.
func printCodeBlock(cb *component.CodeBlock, width, height int) {
	buf := buffer.NewBuffer(width, height)
	cb.Paint(buf)

	for y := 0; y < height; y++ {
		var line strings.Builder
		line.WriteString("  ")
		hasContent := false
		for x := 0; x < width; x++ {
			cell := buf.GetCell(x, y)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			line.WriteRune(r)
			if r != ' ' {
				hasContent = true
			}
		}
		if hasContent {
			fmt.Println(strings.TrimRight(line.String(), " "))
		}
	}
}
