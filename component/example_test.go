package component_test

import (
	"fmt"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
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

// ExampleAccordion demonstrates expandable sections.
func ExampleAccordion() {
	a := component.NewAccordion([]component.AccordionItem{
		{Title: "General", Content: "Settings"},
		{Title: "Advanced", Content: "Advanced options"},
	})
	a.Expand(0)
	fmt.Printf("Item 0 expanded: %v\n", a.IsExpanded(0))
	// Output: Item 0 expanded: true
}

// ExampleHeatmap demonstrates a GitHub-style activity grid.
func ExampleHeatmap() {
	h := component.NewHeatmap(7, 24) // 7 days × 24 hours
	h.SetCell(0, 9, 50)              // Monday 9am
	h.SetCell(0, 14, 80)             // Monday 2pm
	r, c := h.Dimensions()
	fmt.Printf("Grid: %dx%d, Max: %d\n", r, c, h.MaxValue())
	// Output: Grid: 7x24, Max: 80
}

// ExampleBreadcrumb demonstrates navigation path.
func ExampleBreadcrumb() {
	b := component.NewBreadcrumb([]string{"Home", "Settings", "AI"})
	fmt.Println(b.String())
	// Output: Home › Settings › AI
}

// ExamplePieChart demonstrates AI token usage visualization.
func ExamplePieChart() {
	p := component.NewPieChart([]component.PieSlice{
		{Label: "Input", Value: 100},
		{Label: "Output", Value: 50},
	})
	fmt.Printf("Total: %.0f\n", p.TotalValue())
	// Output: Total: 150
}

// ExampleAvatar demonstrates creating avatars for AI chat participants.
func ExampleAvatar() {
	// User avatar with auto-extracted initials
	user := component.NewAvatar("Alice Brown")
	fmt.Println(user.Initials())

	// AI assistant with emoji icon
	ai := component.NewAvatar("Assistant")
	ai.SetIcon("🤖")
	fmt.Println(ai.Icon())

	// Manual initials override
	bot := component.NewAvatar("fluui-bot")
	bot.SetInitials("FB")
	fmt.Println(bot.Initials())
	// Output:
	// AB
	// 🤖
	// FB
}

// ExampleKBD demonstrates rendering keyboard shortcuts.
func ExampleKBD() {
	// Default inverse-video style
	k1 := component.NewKBD("Ctrl+C")
	fmt.Println(k1.Text())

	// Bracket style
	k2 := component.NewKBD("Enter")
	k2.SetVariant(component.KBDBracket)
	fmt.Println(k2.Variant())

	// Bordered keycap style
	k3 := component.NewKBD("⌘K")
	k3.SetVariant(component.KBDBordered)
	fmt.Println(k3.Variant())
	// Output:
	// Ctrl+C
	// 1
	// 2
}

// ExampleDiffStatBar demonstrates rendering diff statistics for code review.
func ExampleDiffStatBar() {
	// Create a diff stat bar with additions and deletions
	d := component.NewDiffStatBar(120, 35)
	d.SetStats(120, 35, 3)

	// Text-only style: compact "+120 -35"
	d.SetStyle(component.DiffStatStyleText)
	fmt.Printf("Additions: %d, Deletions: %d\n", d.Additions(), d.Deletions())

	// Full style with file count
	d.SetStyle(component.DiffStatStyleFull)
	fmt.Printf("Files: %d\n", d.Files())
	// Output:
	// Additions: 120, Deletions: 35
	// Files: 3
}

// ExampleConfidenceMeter demonstrates displaying AI model confidence.
func ExampleConfidenceMeter() {
	c := component.NewConfidenceMeter(0.92)
	c.SetLabel("Confidence")
	fmt.Printf("Value: %.2f\n", c.Value())
	// Output: Value: 0.92
}

// ExampleToast demonstrates creating notification toasts.
func ExampleToast() {
	t := component.NewToast("Build complete", component.ToastSuccess)
	t.SetDuration(5 * 1000000000) // 5s
	fmt.Printf("Level: %v, Message: %s\n", t.Level(), t.Message())
	// Output: Level: 1, Message: Build complete
}

// ExampleColorSwatch demonstrates displaying colors.
func ExampleColorSwatch() {
	s := component.NewColorSwatch(buffer.RGB(0xFF, 0x80, 0x00))
	s.SetLabel("Accent Orange")
	fmt.Printf("ShowHex: %v\n", s.ShowHex())
	// Output: ShowHex: true
}

// ExampleChip demonstrates entity tags.
func ExampleChip() {
	c := component.NewChip("gpt-4")
	c.SetIcon("🤖")
	c.SetVariant(component.ChipFilled)
	fmt.Printf("Text: %s\n", c.Text())
	// Output: Text: gpt-4
}

// ExampleStatCard demonstrates metric display.
func ExampleStatCard() {
	sc := component.StatCardFromInt("Tokens", 42000)
	sc.SetDelta("+15%", true)
	delta, pos := sc.Delta()
	fmt.Printf("Value: %s, Delta: %s, Positive: %v\n", sc.Value(), delta, pos)
	// Output: Value: 42000, Delta: +15%, Positive: true
}

// ExampleMarkdownStream demonstrates streaming markdown rendering.
func ExampleMarkdownStream() {
	m := component.NewMarkdownStream()
	m.Append("# Hello")
	m.Append("\n\nWorld")
	fmt.Printf("Source length: %d\n", len(m.Source()))
	// Output: Source length: 14
}

// ExampleMetricBar demonstrates displaying a metric with range.
func ExampleMetricBar() {
	mb := component.NewMetricBar("CPU", 75.5, 0, 100)
	mb.SetUnit("%")
	fmt.Printf("Value: %.1f, Max: %.0f\n", mb.Value(), mb.Max())
	// Output: Value: 75.5, Max: 100
}

// ExampleHintLabel demonstrates creating help text.
func ExampleHintLabel() {
	h := component.NewHintLabel("Press ? for help")
	fmt.Printf("Text: %s\n", h.Text())
	// Output: Text: Press ? for help
}

// ExampleSearchBar demonstrates creating a search input.
func ExampleSearchBar() {
	s := component.NewSearchBar("Search files...")
	s.SetQuery("main.go")
	fmt.Printf("Query: %s\n", s.Query())
	// Output: Query: main.go
}

// ExampleFileTree demonstrates building a file tree.
func ExampleFileTree() {
	ft := component.NewFileTree("myapp", []component.FileNode{
		{Name: "main.go"},
		{Name: "src", IsDir: true, Expanded: true, Children: []component.FileNode{
			{Name: "handler.go"},
		}},
	})
	fmt.Printf("Root: %s\n", ft.Root())
	// Output: Root: myapp
}

// ExampleDivider demonstrates a section separator.
func ExampleDivider() {
	d := component.NewDivider("Settings")
	d.SetChar('═')
	fmt.Printf("Label: %s\n", d.Label())
	// Output: Label: Settings
}

// ExampleRating demonstrates a star rating display.
func ExampleRating() {
	r := component.NewRating(4.5, 5)
	r.SetShowNumber(true)
	fmt.Printf("Value: %.1f/%d\n", r.Value(), r.Max())
	// Output: Value: 4.5/5
}

// ExampleCircularProgress demonstrates a ring progress indicator.
func ExampleCircularProgress() {
	c := component.NewCircularProgress(0.75)
	c.SetLabel("Loading")
	fmt.Printf("Value: %.0f%%\n", c.Value()*100)
	// Output: Value: 75%
}

// ExampleTogglePill demonstrates an on/off toggle.
func ExampleTogglePill() {
	tp := component.NewTogglePill(true)
	tp.SetLabel("Auto-save")
	fmt.Printf("On: %v\n", tp.IsOn())
	// Output: On: true
}

// ExampleSkeleton demonstrates a loading placeholder.
func ExampleSkeleton() {
	s := component.NewSkeleton(20, 1)
	fmt.Printf("Width: %d, Animate: %v\n", s.Width(), s.Animate())
	// Output: Width: 20, Animate: true
}

func ExampleButton() {
	b := component.NewButton("Save")
	fmt.Printf("Label: %s\n", b.Label())
	// Output: Label: Save
}

func ExampleBadge() {
	b := component.NewBadge("NEW", component.BadgeInfo)
	fmt.Printf("Text: %s\n", b.Text())
	// Output: Text: NEW
}

func ExampleBanner() {
	b := component.NewBanner(component.BannerInfo, "Update available")
	fmt.Printf("Message: %s\n", b.Message())
	// Output: Message: Update available
}

func ExampleCheckbox() {
	c := component.NewCheckbox([]string{"Enable notifications"})
	fmt.Printf("Items: %d\n", len(c.Items()))
	// Output: Items: 1
}

func ExampleCodeBlock() {
	cb := component.NewCodeBlock("go", `fmt.Println("Hello")`)
	fmt.Printf("Language: %s\n", cb.Language())
	// Output: Language: go
}

func ExampleSpinner() {
	s := component.NewSpinner("Loading...")
	fmt.Printf("Label: %s\n", s.Label())
	// Output: Label: Loading...
}

func ExampleSlider() {
	s := component.NewSlider()
	s.SetRange(0, 100)
	s.SetValue(42)
	fmt.Printf("Value: %.0f\n", s.Value())
	// Output: Value: 42
}

func ExampleSwitch() {
	sw := component.NewSwitch("Toggle")
	sw.SetOn(true)
	fmt.Printf("On: %v\n", sw.IsOn())
	// Output: On: true
}

func ExampleProgressBar() {
	p := component.NewProgressBar()
	p.SetProgress(0.65)
	fmt.Printf("Progress: %.0f%%\n", p.Progress()*100)
	// Output: Progress: 65%
}

func ExampleCollapsible() {
	c := component.NewCollapsible("Details", component.NewText("content"))
	c.SetExpanded(true)
	fmt.Printf("Expanded: %v\n", c.Expanded())
	// Output: Expanded: true
}

func ExampleRadioGroup() {
	rg := component.NewRadioGroup([]string{"Option A", "Option B"})
	rg.SetSelected(1)
	fmt.Printf("Selected: %s\n", rg.SelectedLabel())
	// Output: Selected: Option B
}

func ExampleAutoComplete() {
	ac := component.NewAutoComplete()
	ac.SetItems([]component.CompletionItem{
		{Label: "fmt", Description: "format package"},
	})
	ac.SetQuery("f")
	fmt.Printf("Query: %s\n", ac.Query())
	// Output: Query: f
}

func ExampleTable() {
	tbl := component.NewTable([]string{"Name", "Score"}, []string{"Alice", "95"})
	fmt.Printf("Columns: %d\n", len(tbl.Headers()))
	// Output: Columns: 2
}

func ExampleColorPicker() {
	cp := component.NewColorPicker()
	fmt.Printf("Mode: %d\n", cp.Mode())
	// Output: Mode: 0
}

func ExampleCommandPalette() {
	cp := component.NewCommandPalette()
	cp.SetCommands([]component.Command{
		{Label: "Save file"},
	})
	fmt.Printf("Query: %q\n", cp.Query())
	// Output: Query: ""
}
