package component

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- Debounce behavior ---

func TestP60_SetStreamDebounce(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreamDebounce(5)
	if cb.StreamDebounce() != 5 {
		t.Errorf("expected debounce=5, got %d", cb.StreamDebounce())
	}
}

func TestP60_DefaultDebounce(t *testing.T) {
	cb := NewCodeBlock("go", "")
	if cb.StreamDebounce() != 10 {
		t.Errorf("expected default debounce=10, got %d", cb.StreamDebounce())
	}
}

func TestP60_Debounce_UsesPlainFallback(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(5)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Append 1-4: should use plain fallback (debounce=5)
	for i := 0; i < 4; i++ {
		cb.AppendSource("x = 1\n")
	}
	if cb.StreamAppendCount() != 4 {
		t.Errorf("expected 4 appends, got %d", cb.StreamAppendCount())
	}

	// Append 5: should trigger full re-highlight (debounce interval)
	cb.AppendSource("y = 2\n")
	if cb.StreamAppendCount() != 5 {
		t.Errorf("expected 5 appends, got %d", cb.StreamAppendCount())
	}
}

func TestP60_Debounce_Disabled(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(0) // 0 = always re-highlight
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	cb.AppendSource("x = 1\n")
	cb.AppendSource("y = 2\n")
	cb.AppendSource("z = 3\n")

	// With debounce=0, every append does full re-highlight
	// Source should be correct
	src := cb.Source()
	if !strings.Contains(src, "x = 1") {
		t.Errorf("expected source to contain appended text, got %q", src)
	}
}

func TestP60_FinishStreaming_ClearsFallback(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(5)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Append during streaming (triggers plain fallback)
	cb.AppendSource("package main\n")
	cb.AppendSource("func main() {}\n")

	// Finish streaming
	cb.FinishStreaming()

	// After finish, source should still be correct
	src := cb.Source()
	if !strings.Contains(src, "package main") {
		t.Errorf("expected source preserved, got %q", src)
	}
}

// --- Paint with streaming ---

func TestP60_PaintDuringStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(3)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Append several tokens
	cb.AppendSource("package main\n")
	cb.AppendSource("func main() {\n")
	cb.AppendSource("    println(\"hello\")\n")

	// Paint should work without panic even during debounced streaming
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP60_PaintAfterFinish(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	cb.AppendSource("package main\n")
	cb.AppendSource("func main() {}\n")
	cb.FinishStreaming()

	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP60_StreamAppendCounter(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)

	for i := 0; i < 10; i++ {
		cb.AppendSource("x\n")
	}
	if cb.StreamAppendCount() != 10 {
		t.Errorf("expected 10 appends, got %d", cb.StreamAppendCount())
	}
}

// --- Concurrent streaming + painting ---

func TestP60_ConcurrentStreamingAndPaint(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(2)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup

	// Concurrent appenders
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				cb.AppendSource("x = 1\n")
			}
		}(i)
	}

	// Concurrent painters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				buf := buffer.NewBuffer(60, 20)
				cb.Paint(buf)
			}
		}()
	}

	wg.Wait()
}

// --- Benchmark: streaming without debounce (old behavior) ---

func BenchmarkP60_StreamingNoDebounce(b *testing.B) {
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	for i := 0; i < 10; i++ {
		fmt.Printf("Line %d\n", i)
	}
}
`
	// Split into tokens for streaming
	lines := strings.Split(code, "\n")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cb := NewCodeBlock("go", "")
		cb.SetStreaming(true)
		cb.SetStreamDebounce(0) // no debounce = old behavior
		cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
		for _, line := range lines {
			cb.AppendSource(line + "\n")
		}
	}
}

// --- Benchmark: streaming with debounce (new behavior) ---

func BenchmarkP60_StreamingWithDebounce(b *testing.B) {
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	for i := 0; i < 10; i++ {
		fmt.Printf("Line %d\n", i)
	}
}
`
	lines := strings.Split(code, "\n")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cb := NewCodeBlock("go", "")
		cb.SetStreaming(true)
		cb.SetStreamDebounce(10) // debounce every 10 appends
		cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
		for _, line := range lines {
			cb.AppendSource(line + "\n")
		}
	}
}

// --- Benchmark: streaming 100 lines ---

func BenchmarkP60_Streaming100Lines_NoDebounce(b *testing.B) {
	code := strings.Repeat("x := make([]int, 100)\nfor i := range x {\n    x[i] = i * i\n}\n", 25)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cb := NewCodeBlock("go", "")
		cb.SetStreaming(true)
		cb.SetStreamDebounce(0)
		cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
		for _, ch := range code {
			cb.AppendSource(string(ch))
		}
	}
}

func BenchmarkP60_Streaming100Lines_WithDebounce(b *testing.B) {
	code := strings.Repeat("x := make([]int, 100)\nfor i := range x {\n    x[i] = i * i\n}\n", 25)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cb := NewCodeBlock("go", "")
		cb.SetStreaming(true)
		cb.SetStreamDebounce(10)
		cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
		for _, ch := range code {
			cb.AppendSource(string(ch))
		}
	}
}
