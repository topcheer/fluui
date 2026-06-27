package app

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TestChatAppConcurrentRender tests that Render and AddXxx can run concurrently
// without data races. The streaming goroutine calls AddAssistantText + AppendDelta
// while the event loop calls Render.
func TestChatAppConcurrentRender(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: simulate streaming AddXxx + AppendDelta
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			at := app.AddAssistantText()
			for j := 0; j < 5; j++ {
				at.AppendDelta(fmt.Sprintf("delta %d-%d ", i, j))
			}
		}
	}()

	// Goroutine 2: simulate Render loop
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.Render(buf)
		}
	}()

	wg.Wait()
}

// TestChatAppConcurrentKeys tests HandleKey + Render + AddXxx concurrently.
func TestChatAppConcurrentKeys(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)

	keys := []term.KeyEvent{
		{Key: term.KeyUp},
		{Key: term.KeyDown},
		{Key: term.KeyPageUp},
		{Key: term.KeyPageDown},
		{Key: term.KeyHome},
		{Key: term.KeyEnd},
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: keys
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.HandleKey(&keys[i%len(keys)])
		}
	}()

	// Goroutine 2: render
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.Render(buf)
		}
	}()

	// Goroutine 3: add blocks
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			app.AddAssistantText().AppendDelta("hello")
		}
	}()

	wg.Wait()
}

// TestChatAppConcurrentScroll tests ScrollUp/Down/ToBottom + Render concurrently.
func TestChatAppConcurrentScroll(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)

	// Add some content so scroll has work to do
	for i := 0; i < 30; i++ {
		app.AddAssistantText().AppendDelta(fmt.Sprintf("line %d\n", i))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: scroll operations
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.ScrollDown()
			app.ScrollUp()
			app.ScrollToBottom()
		}
	}()

	// Goroutine 2: render
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.Render(buf)
		}
	}()

	wg.Wait()
}

// TestChatAppConcurrentStreamDelta tests StreamDelta + Render concurrently.
func TestChatAppConcurrentStreamDelta(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: stream deltas
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.StreamDelta(block.StreamDelta{
				Type:    "text",
				Content: fmt.Sprintf("delta %d ", i),
			})
		}
	}()

	// Goroutine 2: render
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.Render(buf)
		}
	}()

	wg.Wait()
}

// TestChatAppConcurrentDirty tests IsDirty/ClearDirty + Render concurrently.
func TestChatAppConcurrentDirty(t *testing.T) {
	app := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: mark dirty via AddXxx
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			app.AddThinking().AppendDelta("thinking...")
		}
	}()

	// Goroutine 2: IsDirty + ClearDirty
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.IsDirty()
			app.ClearDirty()
		}
	}()

	// Goroutine 3: Render
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.Render(buf)
		}
	}()

	wg.Wait()
}

// TestChatAppConcurrentSize tests SetSize + Render + Size concurrently.
func TestChatAppConcurrentSize(t *testing.T) {
	app := NewChatApp(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: resize
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			app.SetSize(80+i%10, 24+i%5)
		}
	}()

	// Goroutine 2: Size + Render
	go func() {
		defer wg.Done()
		buf := buffer.NewBuffer(90, 30)
		for i := 0; i < 100; i++ {
			app.Size()
			app.Render(buf)
		}
	}()

	wg.Wait()
}
