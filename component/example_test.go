package component_test

import (
	"fmt"

	"github.com/topcheer/fluui/component"
)

// ExampleConversationView demonstrates creating a chat conversation
// with user and assistant messages, tool calls, and citations.
func ExampleConversationView() {
	conv := component.NewConversationView()

	// Add a user message
	conv.AddUserMessage("What files are in /tmp?")

	// Add a tool call (e.g., file listing)
	tc := component.NewToolCallView("list_files", `{"dir":"/tmp"}`)
	conv.AddToolCall(tc)
	tc.SetResult("file1.go\nfile2.go")
	tc.Complete()

	// Add an assistant response
	conv.AddAssistantMessage("Found 2 files in /tmp.", "gpt-4")

	fmt.Printf("Messages: %d\n", conv.MessageCount())
	// Output: Messages: 3
}

// ExampleMessageBubble demonstrates role-based message rendering.
func ExampleMessageBubble() {
	mb := component.NewMessageBubble(component.RoleUser, "Hello!")
	fmt.Printf("Role: %s\n", mb.Role())
	// Output: Role: You
}

// ExampleChatComposer demonstrates input composition with submit callback.
func ExampleChatComposer() {
	composer := component.NewChatComposer()
	composer.SetText("Type a message")
	composer.SetOnSubmit(func(text string) {
		fmt.Printf("Submitted: %s\n", text)
	})
	fmt.Printf("Text: %s\n", composer.Text())
	// Output: Text: Type a message
}

// ExampleTokenUsageWidget demonstrates token tracking display.
func ExampleTokenUsageWidget() {
	w := component.NewTokenUsageWidget("gpt-4")
	w.AddTokens(1500, 800)
	fmt.Printf("Total: %d\n", w.TotalTokens())
	// Output: Total: 2300
}

// ExampleCitationsBlock demonstrates source citation rendering.
func ExampleCitationsBlock() {
	cb := component.NewCitationsBlock([]component.Citation{
		{Index: 1, Title: "Go Docs", URL: "https://go.dev"},
		{Index: 2, Title: "Wikipedia", URL: "https://wikipedia.org"},
	})
	fmt.Printf("Citations: %d\n", cb.Count())
	// Output: Citations: 2
}

// ExampleThinkingIndicator demonstrates the AI "thinking" animation.
func ExampleThinkingIndicator() {
	ti := component.NewThinkingIndicator("Thinking")
	ti.SetLabel("Generating response")

	// Manually advance frames for deterministic output
	for i := 0; i < 4; i++ {
		fmt.Printf("Frame %d\n", ti.FrameIndex())
		ti.AdvanceFrame()
	}
	// Output:
	// Frame 0
	// Frame 1
	// Frame 2
	// Frame 3
}

// ExampleToolCallView demonstrates AI tool call visualization.
func ExampleToolCallView() {
	tc := component.NewToolCallView("read_file", `{"path":"/tmp/data.json"}`)
	tc.Complete()
	fmt.Printf("Tool: %s\n", tc.ToolName())
	// Output: Tool: read_file
}

// ExampleSegmentedControl demonstrates mode switching.
func ExampleSegmentedControl() {
	sc := component.NewSegmentedControl([]string{"Chat", "Code", "Settings"})
	fmt.Printf("Active: %s\n", sc.ActiveLabel())
	sc.SelectNext()
	fmt.Printf("After next: %s\n", sc.ActiveLabel())
	// Output:
	// Active: Chat
	// After next: Code
}

// ExampleSkeletonLoader demonstrates loading placeholder creation.
func ExampleSkeletonLoader() {
	sk := component.NewSkeletonText(3, 40)
	fmt.Printf("Blocks: %d\n", len(sk.Blocks()))
	// Output: Blocks: 3
}
