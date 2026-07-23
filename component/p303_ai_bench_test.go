package component

import (
	"fmt"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P303: Benchmark all new AI-native components

// === ToolCallView ===

func BenchmarkToolCallView_Paint_Collapsed(b *testing.B) {
	tc := NewToolCallView("read_file", `{"path":"/usr/local/go/src/runtime/proc.go","mode":"rw"}`)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Paint(buf)
	}
}

func BenchmarkToolCallView_Paint_Expanded(b *testing.B) {
	tc := NewToolCallView("exec_command", `{"cmd":"go","args":["test","-race","-count=1","./..."],"timeout":"300s"}`)
	tc.SetExpanded(true)
	tc.SetResult(strings.Repeat("ok\tgithub.com/topcheer/fluui\t2.1s\n", 5))
	tc.Complete()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Paint(buf)
	}
}

// === MessageBubble ===

func BenchmarkMessageBubble_Paint_Short(b *testing.B) {
	mb := NewMessageBubble(RoleAssistant, "Hello! How can I help you today?")
	mb.SetModel("gpt-4")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Paint(buf)
	}
}

func BenchmarkMessageBubble_Paint_Long(b *testing.B) {
	long := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 12)
	mb := NewMessageBubble(RoleAssistant, long)
	mb.SetModel("claude-3")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Paint(buf)
	}
}

// === ConversationView ===

func BenchmarkConversationView_Paint_10Messages(b *testing.B) {
	cv := NewConversationView()
	for i := 0; i < 5; i++ {
		cv.AddUserMessage(fmt.Sprintf("User question number %d about Go programming", i))
		cv.AddAssistantMessage(fmt.Sprintf("Here is the answer to question %d with detailed explanation.", i), "gpt-4")
	}
	cv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.Paint(buf)
	}
}

func BenchmarkConversationView_AddMessage(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv := NewConversationView()
		for j := 0; j < 100; j++ {
			cv.AddUserMessage(fmt.Sprintf("message %d", j))
		}
	}
}

// === ChatComposer ===

func BenchmarkChatComposer_Paint(b *testing.B) {
	c := NewChatComposer()
	c.SetText("line one\nline two\nline three")
	c.SetTokenCount(1500, 800)
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

// === CitationsBlock ===

func BenchmarkCitationsBlock_Paint_Collapsed(b *testing.B) {
	cits := make([]Citation, 5)
	for i := range cits {
		cits[i] = Citation{Index: i + 1, Title: fmt.Sprintf("Source %d", i+1), URL: "https://example.com", Snippet: "Some snippet"}
	}
	cb := NewCitationsBlock(cits)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}

func BenchmarkCitationsBlock_Paint_Expanded(b *testing.B) {
	cits := make([]Citation, 5)
	for i := range cits {
		cits[i] = Citation{Index: i + 1, Title: fmt.Sprintf("Source %d", i+1), URL: "https://example.com", Snippet: "Some snippet text here for testing"}
	}
	cb := NewCitationsBlock(cits)
	cb.SetExpanded(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}

// === TokenUsageWidget ===

func BenchmarkTokenUsageWidget_Paint(b *testing.B) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(15420, 8230)
	w.SetContextUsage(65000, 128000)
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Paint(buf)
	}
}
