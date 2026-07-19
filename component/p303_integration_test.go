package component

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P303: Integration test simulating a full AI conversation flow

func TestP303_FullConversationFlow(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)

	// 1. System message
	cv.AddSystemMessage("You are a helpful coding assistant.")
	cv.Paint(buf)
	if cv.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", cv.MessageCount())
	}

	// 2. User asks a question
	cv.AddUserMessage("How do I read a file in Go?")
	cv.Paint(buf)
	if cv.MessageCount() != 2 {
		t.Fatalf("expected 2 messages, got %d", cv.MessageCount())
	}

	// 3. Assistant starts streaming
	cv.AddMessage(ConversationMessage{
		Role:      RoleAssistant,
		Content:   "Let me check the docs...",
		ModelName: "gpt-4",
		Streaming:  true,
	})
	cv.SetStreaming(true)
	cv.Paint(buf)
	if cv.MessageCount() != 3 {
		t.Fatalf("expected 3 messages, got %d", cv.MessageCount())
	}

	// 4. Tool call appears (running → complete)
	tc := NewToolCallView("search_docs", `{"query":"go read file os.ReadFile","limit":5}`)
	cv.AddToolCall(tc)
	cv.Paint(buf)

	// Tool completes
	tc.SetResult("Found 5 relevant docs about os.ReadFile and bufio.Scanner")
	tc.Complete()
	cv.Paint(buf)
	if cv.MessageCount() != 4 {
		t.Fatalf("expected 4 messages, got %d", cv.MessageCount())
	}

	// 5. Assistant responds with the answer
	cv.AddAssistantMessage(
		"You can use `os.ReadFile` to read an entire file:\n\n```go\ndata, err := os.ReadFile(\"file.txt\")\n```\n\nOr use `bufio.Scanner` for line-by-line reading.",
		"gpt-4",
	)
	cv.Paint(buf)

	// 6. Citations
	cv.AddCitations(NewCitationsBlock([]Citation{
		{Index: 1, Title: "Go os package", URL: "https://pkg.go.dev/os", Snippet: "Package os provides platform-independent interface to OS"},
		{Index: 2, Title: "Go bufio package", URL: "https://pkg.go.dev/bufio", Snippet: "Package bufio implements buffered I/O"},
	}))
	cv.Paint(buf)

	if cv.MessageCount() != 6 {
		t.Fatalf("expected 6 messages, got %d", cv.MessageCount())
	}

	// 7. Clear conversation
	cv.Clear()
	if cv.MessageCount() != 0 {
		t.Errorf("expected 0 messages after clear, got %d", cv.MessageCount())
	}
	cv.Paint(buf) // should show empty placeholder
}

func TestP303_ConcurrentPaintAndAdd(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine: add messages
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cv.AddUserMessage(fmt.Sprintf("message %d", i))
		}
	}()

	// Reader goroutine: paint
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cv.Paint(buf)
		}
	}()

	wg.Wait()

	if cv.MessageCount() != 50 {
		t.Errorf("expected 50 messages, got %d", cv.MessageCount())
	}
}

func TestP303_ConcurrentScrollAndAdd(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 30; i++ {
		cv.AddUserMessage(fmt.Sprintf("msg %d", i))
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	cv.Paint(buf) // initialize scroll

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			cv.ScrollUp(1)
			cv.ScrollDown(1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			cv.Paint(buf)
		}
	}()

	wg.Wait()
}

func TestP303_ClearMidConversation(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)

	// Add messages
	cv.AddUserMessage("msg1")
	cv.AddAssistantMessage("resp1", "gpt-4")
	cv.AddToolCall(NewToolCallView("tool", "{}"))
	cv.Paint(buf)

	// Clear while scrolling
	cv.ScrollUp(5)
	cv.Clear()
	cv.Paint(buf)

	if cv.MessageCount() != 0 {
		t.Errorf("expected 0 after clear, got %d", cv.MessageCount())
	}
}

func TestP303_StreamingThenComplete(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)

	// Start streaming
	cv.SetStreaming(true)
	cv.AddUserMessage("Tell me a story")
	cv.AddMessage(ConversationMessage{
		Role:      RoleAssistant,
		Content:   "Once upon",
		ModelName: "gpt-4",
		Streaming:  true,
	})
	cv.Paint(buf)

	// Stop streaming
	cv.SetStreaming(false)
	cv.Paint(buf)

	// Verify no deadlock or panic
	msgs := cv.Messages()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestP303_AllComponentTypesMixed(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(60, 30)

	// Mix all component types
	cv.AddSystemMessage("System initialized")
	cv.AddUserMessage("Complex query")

	tc := NewToolCallView("analyze", `{"data":"large"}`)
	tc.SetExpanded(true)
	tc.SetResult("Analysis complete with 42 results")
	tc.Complete()
	cv.AddToolCall(tc)

	cv.AddAssistantMessage("Based on the analysis, here are the key findings.", "claude-3")

	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Paper A", URL: "https://a.com", Snippet: "Finding A"},
		{Index: 2, Title: "Paper B", URL: "https://b.com", Snippet: "Finding B"},
	})
	cb.SetExpanded(true)
	cv.AddCitations(cb)

	cv.AddMessage(ConversationMessage{
		Role:     RoleAssistant,
		Content:  "Error: could not complete request",
		Error:    true,
	})

	cv.Paint(buf) // should render everything without panic

	if cv.MessageCount() != 6 {
		t.Errorf("expected 6 messages, got %d", cv.MessageCount())
	}
}

func TestP303_ComposerSubmitFlow(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)

	composer := NewChatComposer()
	submitted := []string{}
	composer.SetOnSubmit(func(text string) {
		submitted = append(submitted, text)
		cv.AddUserMessage(text)
	})

	// Type and send
	composer.SetText("What is 2+2?")
	composer.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if len(submitted) != 1 || submitted[0] != "What is 2+2?" {
		t.Fatalf("submitted = %v", submitted)
	}
	if cv.MessageCount() != 1 {
		t.Errorf("expected 1 msg in conversation, got %d", cv.MessageCount())
	}

	// Simulate AI response
	cv.AddAssistantMessage("2+2 = 4", "gpt-4")
	cv.Paint(buf)

	// Send another
	composer.SetText("Thanks!")
	composer.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if cv.MessageCount() != 3 {
		t.Errorf("expected 3 messages, got %d", cv.MessageCount())
	}
}
