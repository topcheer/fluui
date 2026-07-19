package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === ConversationView core tests ===

func TestP299_NewConversationView(t *testing.T) {
	cv := NewConversationView()
	if cv.MessageCount() != 0 {
		t.Errorf("count = %d, want 0", cv.MessageCount())
	}
	if cv.ID() == "" {
		t.Error("ID should not be empty")
	}
	if !cv.autoScroll {
		t.Error("should start with autoScroll")
	}
}

func TestP299_AddUserMessage(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("Hello")
	if cv.MessageCount() != 1 {
		t.Errorf("count = %d, want 1", cv.MessageCount())
	}
	msgs := cv.Messages()
	if msgs[0].Role != RoleUser {
		t.Errorf("role = %v, want RoleUser", msgs[0].Role)
	}
	if msgs[0].Content != "Hello" {
		t.Errorf("content = %q", msgs[0].Content)
	}
}

func TestP299_AddAssistantMessage(t *testing.T) {
	cv := NewConversationView()
	cv.AddAssistantMessage("Hi there!", "gpt-4")
	if cv.MessageCount() != 1 {
		t.Errorf("count = %d, want 1", cv.MessageCount())
	}
	msgs := cv.Messages()
	if msgs[0].Role != RoleAssistant {
		t.Errorf("role = %v", msgs[0].Role)
	}
	if msgs[0].ModelName != "gpt-4" {
		t.Errorf("model = %q", msgs[0].ModelName)
	}
}

func TestP299_AddSystemMessage(t *testing.T) {
	cv := NewConversationView()
	cv.AddSystemMessage("System notice")
	msgs := cv.Messages()
	if msgs[0].Role != RoleSystem {
		t.Errorf("role = %v", msgs[0].Role)
	}
}

func TestP299_AddMessage(t *testing.T) {
	cv := NewConversationView()
	ts := time.Now().Add(-1 * time.Hour)
	cv.AddMessage(ConversationMessage{
		Role:      RoleAssistant,
		Content:   "test",
		Timestamp: ts,
	})
	msgs := cv.Messages()
	if !msgs[0].Timestamp.Equal(ts) {
		t.Error("timestamp should be preserved")
	}
}

func TestP299_AddMessage_AutoTimestamp(t *testing.T) {
	cv := NewConversationView()
	cv.AddMessage(ConversationMessage{
		Role:    RoleUser,
		Content: "test",
	})
	msgs := cv.Messages()
	if msgs[0].Timestamp.IsZero() {
		t.Error("timestamp should be auto-set")
	}
}

func TestP299_AddToolCall(t *testing.T) {
	cv := NewConversationView()
	tc := NewToolCallView("read_file", `{"path":"/tmp"}`)
	cv.AddToolCall(tc)
	if cv.MessageCount() != 1 {
		t.Errorf("count = %d, want 1", cv.MessageCount())
	}
	msgs := cv.Messages()
	if msgs[0].ToolCall == nil {
		t.Error("ToolCall should not be nil")
	}
}

func TestP299_AddCitations(t *testing.T) {
	cv := NewConversationView()
	cb := NewCitationsBlock([]Citation{{Index: 1, Title: "Test", URL: "https://x.com"}})
	cv.AddCitations(cb)
	if cv.MessageCount() != 1 {
		t.Errorf("count = %d", cv.MessageCount())
	}
	if cv.Messages()[0].Citations == nil {
		t.Error("Citations should not be nil")
	}
}

func TestP299_Clear(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("a")
	cv.AddUserMessage("b")
	cv.Clear()
	if cv.MessageCount() != 0 {
		t.Errorf("count = %d after clear, want 0", cv.MessageCount())
	}
}

func TestP299_Messages_ReturnsCopy(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("original")
	msgs := cv.Messages()
	msgs[0].Content = "modified"
	if cv.Messages()[0].Content == "modified" {
		t.Error("Messages() should return a copy")
	}
}

func TestP299_SetStreaming(t *testing.T) {
	cv := NewConversationView()
	cv.SetStreaming(true)
	if !cv.IsStreaming() {
		t.Error("should be streaming")
	}
	if !cv.autoScroll {
		t.Error("streaming should enable autoScroll")
	}
	cv.SetStreaming(false)
	if cv.IsStreaming() {
		t.Error("should not be streaming")
	}
}

func TestP299_ScrollUp(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 20; i++ {
		cv.AddUserMessage("line " + string(rune('a'+i)))
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	cv.Paint(buffer.NewBuffer(60, 5))
	cv.ScrollUp(3)
	cv.mu.Lock()
	if cv.scrollOff < 0 {
		t.Error("scrollOff should not go negative")
	}
	cv.mu.Unlock()
}

func TestP299_ScrollUp_DisablesAutoScroll(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("test")
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	cv.Paint(buffer.NewBuffer(60, 5))
	cv.ScrollUp(1)
	cv.mu.Lock()
	if cv.autoScroll {
		t.Error("scrolling up should disable autoScroll")
	}
	cv.mu.Unlock()
}

func TestP299_ScrollDown(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 10; i++ {
		cv.AddUserMessage("msg " + string(rune('a'+i)))
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	cv.Paint(buffer.NewBuffer(60, 3))
	cv.ScrollUp(5)
	cv.ScrollDown(2)
	cv.mu.Lock()
	off := cv.scrollOff
	cv.mu.Unlock()
	if off < 0 {
		t.Error("scrollOff should not be negative")
	}
}

func TestP299_ScrollDown_AtBottom_ReenablesAutoScroll(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 10; i++ {
		cv.AddUserMessage("msg")
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	cv.Paint(buffer.NewBuffer(60, 3))
	cv.ScrollUp(5) // disables autoScroll
	cv.ScrollToBottom() // re-enables
	cv.mu.Lock()
	if !cv.autoScroll {
		t.Error("ScrollToBottom should re-enable autoScroll")
	}
	cv.mu.Unlock()
}

func TestP299_ScrollToTop(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 10; i++ {
		cv.AddUserMessage("msg")
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	cv.Paint(buffer.NewBuffer(60, 3))
	cv.ScrollToTop()
	cv.mu.Lock()
	if cv.scrollOff != 0 {
		t.Errorf("scrollOff = %d, want 0", cv.scrollOff)
	}
	if cv.autoScroll {
		t.Error("ScrollToTop should disable autoScroll")
	}
	cv.mu.Unlock()
}

func TestP299_ScrollToBottom(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 10; i++ {
		cv.AddUserMessage("msg")
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	cv.ScrollToBottom()
	cv.mu.Lock()
	if !cv.autoScroll {
		t.Error("should be autoScroll after ScrollToBottom")
	}
	cv.mu.Unlock()
}

func TestP299_HandleKey(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 10; i++ {
		cv.AddUserMessage("msg")
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	cv.Paint(buffer.NewBuffer(60, 3))

	tests := []struct {
		name string
		key  term.KeyCode
	}{
		{"up", term.KeyUp},
		{"down", term.KeyDown},
		{"pageup", term.KeyPageUp},
		{"pagedown", term.KeyPageDown},
		{"home", term.KeyHome},
		{"end", term.KeyEnd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &term.KeyEvent{Key: tt.key}
			if !cv.HandleKey(ev) {
				t.Errorf("HandleKey(%v) returned false", tt.key)
			}
		})
	}
}

func TestP299_HandleKey_Nil(t *testing.T) {
	cv := NewConversationView()
	if cv.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestP299_HandleKey_Unknown(t *testing.T) {
	cv := NewConversationView()
	ev := &term.KeyEvent{Key: term.KeyEnter}
	if cv.HandleKey(ev) {
		t.Error("HandleKey(Enter) should return false (not handled)")
	}
}

func TestP299_Paint_Empty(t *testing.T) {
	cv := NewConversationView()
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cv.Paint(buf) // should show "No messages" placeholder, not panic
}

func TestP299_Paint_WithMessages(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("Hello")
	cv.AddAssistantMessage("Hi!", "gpt-4")
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cv.Paint(buf)
}

func TestP299_Paint_WithToolCall(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("List files")
	tc := NewToolCallView("list_files", `{"dir":"/tmp"}`)
	cv.AddToolCall(tc)
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cv.Paint(buf)
}

func TestP299_Paint_WithCitations(t *testing.T) {
	cv := NewConversationView()
	cv.AddAssistantMessage("According to sources...", "perplexity")
	cb := NewCitationsBlock([]Citation{
		{Index: 1, Title: "Source A", URL: "https://a.com", Snippet: "Snippet A"},
	})
	cv.AddCitations(cb)
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	cv.Paint(buf)
}

func TestP299_Paint_Streaming(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("Tell me a story")
	cv.AddMessage(ConversationMessage{
		Role:     RoleAssistant,
		Content:  "Once upon a time, there was a",
		Streaming: true,
	})
	cv.SetStreaming(true)
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cv.Paint(buf)
}

func TestP299_Paint_Scrolled(t *testing.T) {
	cv := NewConversationView()
	for i := 0; i < 30; i++ {
		cv.AddUserMessage("Message " + string(rune('A'+i%26)))
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	cv.Paint(buf)
	cv.ScrollUp(10)
	cv.Paint(buf)
}

func TestP299_Paint_ZeroBounds(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("test")
	cv.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cv.Paint(buf) // should not panic
}

func TestP299_Paint_NarrowWidth(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("very long text that overflows")
	cv.AddAssistantMessage("response", "model")
	cv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	cv.Paint(buf)
}

func TestP299_Measure(t *testing.T) {
	cv := NewConversationView()
	s := cv.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 80 || s.H != 24 {
		t.Errorf("size = %dx%d, want 80x24", s.W, s.H)
	}
	// defaults
	s2 := cv.Measure(Constraints{})
	if s2.W != 80 || s2.H != 24 {
		t.Errorf("default size = %dx%d, want 80x24", s2.W, s2.H)
	}
}

func TestP299_Concurrent(t *testing.T) {
	cv := NewConversationView()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 200; i++ {
			cv.AddUserMessage("concurrent msg")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = cv.MessageCount()
		_ = cv.Messages()
	}
	<-done
}

func TestP299_SatisfiesComponent(t *testing.T) {
	var _ Component = (*ConversationView)(nil)
}
