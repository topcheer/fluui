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

// ExampleBorder demonstrates wrapping a component with a border.
func ExampleBorder() {
	b := component.NewBorder(component.NewText("content"))
	fmt.Printf("Has child: %v\n", b != nil)
	// Output: Has child: true
}

// ExampleCalendar demonstrates a date picker calendar.
func ExampleCalendar() {
	c := component.NewCalendar()
	fmt.Printf("Has calendar: %v\n", c != nil)
	// Output: Has calendar: true
}

// ExampleContextMenu demonstrates a right-click context menu.
func ExampleContextMenu() {
	cm := component.NewContextMenu()
	fmt.Printf("Has menu: %v\n", cm != nil)
	// Output: Has menu: true
}

// ExampleDrawer demonstrates a slide-out drawer panel.
func ExampleDrawer() {
	d := component.NewDrawer(component.DrawerLeft, "Filters")
	fmt.Printf("Title: %s\n", d.Title())
	// Output: Title: Filters
}

// ExampleFooter demonstrates a bottom status footer.
func ExampleFooter() {
	f := component.NewFooter()
	fmt.Printf("Has footer: %v\n", f != nil)
	// Output: Has footer: true
}

// ExampleForm demonstrates creating a form.
func ExampleForm() {
	f := component.NewForm()
	fmt.Printf("Has form: %v\n", f != nil)
	// Output: Has form: true
}

// ExampleGauge demonstrates a gauge indicator.
func ExampleGauge() {
	g := component.NewGauge()
	g.SetRange(0, 100)
	g.SetValue(72)
	g.SetLabel("CPU")
	fmt.Printf("Value: %.0f\n", g.Value())
	// Output: Value: 72
}

// ExampleHeader demonstrates a top header bar.
func ExampleHeader() {
	h := component.NewHeader("MyApp")
	fmt.Printf("Has header: %v\n", h != nil)
	// Output: Has header: true
}

// ExamplePopover demonstrates a floating popover.
func ExamplePopover() {
	p := component.NewPopover(component.Rect{X: 0, Y: 0, W: 10, H: 5}, "Title", "Body text")
	fmt.Printf("Has popover: %v\n", p != nil)
	// Output: Has popover: true
}

// ExampleStatusBar demonstrates creating a status bar.
func ExampleStatusBar() {
	sb := component.NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	fmt.Printf("Has statusbar: %v\n", sb != nil)
	// Output: Has statusbar: true
}

// ExampleTooltip demonstrates a help tooltip.
func ExampleTooltip() {
	tt := component.NewTooltip("Press Ctrl+S to save")
	fmt.Printf("Has tooltip: %v\n", tt != nil)
	// Output: Has tooltip: true
}

// ExampleDiffViewer demonstrates a diff display.
func ExampleDiffViewer() {
	dv := component.NewDiffViewer()
	fmt.Printf("Has diff: %v\n", dv != nil)
	// Output: Has diff: true
}

// ExamplePagination demonstrates page navigation.
func ExamplePagination() {
	p := component.NewPagination()
	fmt.Printf("Has pagination: %v\n", p != nil)
	// Output: Has pagination: true
}

// ExampleStepper demonstrates a multi-step wizard flow.
func ExampleStepper() {
	s := component.NewStepper([]component.StepperStep{{Title: "Step 1"}})
	fmt.Printf("Steps: %d\n", s.StepCount())
	// Output: Steps: 1
}

// ExampleTextArea demonstrates a multi-line text input.
func ExampleTextArea() {
	ta := component.NewTextArea()
	fmt.Printf("Has textarea: %v\n", ta != nil)
	// Output: Has textarea: true
}

// ExampleTimeline demonstrates an event timeline.
func ExampleTimeline() {
	tl := component.NewTimeline([]component.TimelineEvent{{Title: "Created"}})
	fmt.Printf("Has timeline: %v\n", tl != nil)
	// Output: Has timeline: true
}

// ExamplePlaceholder demonstrates a loading placeholder.
func ExamplePlaceholder() {
	p := component.NewPlaceholder("Loading...")
	fmt.Printf("Label: %s\n", p.Label())
	// Output: Label: Loading...
}

// ExampleFill demonstrates a fill character component.
func ExampleFill() {
	f := component.NewFill('-', buffer.Style{})
	fmt.Printf("Has fill: %v\n", f != nil)
	// Output: Has fill: true
}

// ExampleDigits demonstrates a numeric display.
func ExampleDigits() {
	d := component.NewDigits("42")
	fmt.Printf("Has digits: %v\n", d != nil)
	// Output: Has digits: true
}

// ExampleWizard demonstrates a step-by-step wizard.
func ExampleWizard() {
	w := component.NewWizard(nil)
	fmt.Printf("Has wizard: %v\n", w != nil)
	// Output: Has wizard: true
}

func ExampleAIProgress() {
	a := component.NewAIProgress()
	a.SetPhase(component.AIPhaseGenerating)
	fmt.Printf("Phase: %s\n", a.PhaseLabel())
	// Output: Phase: Generating...
}

func ExampleApprovalDialog() {
	d := component.NewApprovalDialog("Deploy?", "Push to production?")
	fmt.Printf("Title: %s\n", d.Title())
	// Output: Title: Deploy?
}

func ExampleBarChart() {
	bc := component.NewBarChart()
	fmt.Printf("Has chart: %v\n", bc != nil)
	// Output: Has chart: true
}

func ExampleCallout() {
	c := component.NewCallout(component.CalloutInfo, "System update available")
	fmt.Printf("Has callout: %v\n", c != nil)
	// Output: Has callout: true
}

func ExampleCanvas() {
	c := component.NewCanvas()
	fmt.Printf("Has canvas: %v\n", c != nil)
	// Output: Has canvas: true
}

func ExampleContentSwitcher() {
	cs := component.NewContentSwitcher()
	fmt.Printf("Has switcher: %v\n", cs != nil)
	// Output: Has switcher: true
}

func ExampleDialog() {
	d := component.NewDialog(component.DialogInfo, "Notice", "File saved")
	fmt.Printf("Has dialog: %v\n", d != nil)
	// Output: Has dialog: true
}

func ExampleDiffPreview() {
	dp := component.NewDiffPreview()
	fmt.Printf("Has diff: %v\n", dp != nil)
	// Output: Has diff: true
}

func ExampleEmptyState() {
	es := component.NewEmptyState("No results", "Try a different search")
	fmt.Printf("Has empty state: %v\n", es != nil)
	// Output: Has empty state: true
}

func ExampleFilteredList() {
	fl := component.NewFilteredList([]string{"apple", "banana"})
	fmt.Printf("Items: %d\n", len(fl.Items()))
	// Output: Items: 2
}

func ExampleGrid() {
	g := component.NewGrid()
	fmt.Printf("Has grid: %v\n", g != nil)
	// Output: Has grid: true
}

func ExampleInfoCard() {
	c := component.NewInfoCard("ℹ", "Status", "All good")
	fmt.Printf("Title: %s\n", c.Title())
	// Output: Title: Status
}

func ExampleLineChart() {
	lc := component.NewLineChart()
	fmt.Printf("Has chart: %v\n", lc != nil)
	// Output: Has chart: true
}

func ExampleListView() {
	lv := component.NewListView([]string{"Item A", "Item B"})
	fmt.Printf("Has list: %v\n", lv != nil)
	// Output: Has list: true
}

func ExampleMaskedInput() {
	mi := component.NewMaskedInput("###-####")
	fmt.Printf("Has input: %v\n", mi != nil)
	// Output: Has input: true
}

func ExampleDropdown() {
	d := component.NewDropdown([]component.DropdownItem{{Label: "A"}})
	fmt.Printf("Items: %d\n", len(d.Items()))
	// Output: Items: 1
}

func ExampleComboBox() {
	cb := component.NewComboBox([]string{"a", "b"})
	fmt.Printf("Items: %d\n", len(cb.Items()))
	// Output: Items: 2
}

func ExampleFunnelChart() {
	fc := component.NewFunnelChart()
	fc.AddStage(component.FunnelStage{Label: "Visit", Value: 1000})
	fmt.Printf("Has funnel: %v\n", fc != nil)
	// Output: Has funnel: true
}

func ExampleDataGrid() {
	dg := component.NewDataGrid(nil)
	fmt.Printf("Has grid: %v\n", dg != nil)
	// Output: Has grid: true
}

func ExampleSliderRange() {
	sr := component.NewSliderRange().
		SetLow(25).
		SetHigh(75).
		SetStep(5).
		SetLabel("Price Range")
	fmt.Printf("Range: %.0f - %.0f\n", sr.Low(), sr.High())
	// Output: Range: 25 - 75
}

// ─── P425: 15 new Examples (Direction E) ───

func ExampleTree() {
	root := component.NewTreeNode("root", "Project")
	src := component.NewTreeNode("src", "src")
	src.AddChild(component.NewTreeNode("main", "main.go"))
	src.AddChild(component.NewTreeNode("util", "util.go"))
	root.AddChild(src)
	root.AddChild(component.NewTreeNode("test", "test"))

	tree := component.NewTree()
	tree.SetRoot(root)
	fmt.Printf("Has root: %v\n", tree.Root() != nil)
	// Output: Has root: true
}

func ExampleTextInput() {
	ti := component.NewTextInput()
	ti.SetValue("hello world")
	ti.SetPlaceholder("Type here...")
	fmt.Println(ti.Value())
	// Output: hello world
}

func ExampleTabbedContent() {
	tc := component.NewTabbedContent()
	tc.AddTab("home", "Home", component.NewText("Welcome"))
	tc.AddTab("settings", "Settings", component.NewText("Settings page"))
	tc.AddTab("about", "About", component.NewText("About page"))
	fmt.Printf("Tabs: %d, active: %s\n", tc.TabCount(), tc.ActiveTab())
	// Output: Tabs: 3, active: home
}

func ExampleTagInput() {
	ti := component.NewTagInput("Add tag...")
	fmt.Printf("Placeholder set: %v\n", ti != nil)
	// Output: Placeholder set: true
}

func ExampleSelect() {
	sel := component.NewSelect([]component.SelectOption{
		{Label: "Red", Value: "red"},
		{Label: "Green", Value: "green"},
		{Label: "Blue", Value: "blue"},
	})
	fmt.Printf("Options: %d\n", len(sel.Options()))
	// Output: Options: 3
}

func ExampleQRCode() {
	qr := component.NewQRCode("https://github.com/topcheer/fluui")
	fmt.Printf("Has data: %v\n", qr != nil)
	// Output: Has data: true
}

func ExampleSparkline() {
	sp := component.NewSparkline()
	sp.SetData([]float64{1, 3, 2, 5, 4, 6, 3, 7, 5, 8})
	sp.SetLabel("CPU %")
	fmt.Printf("Points: %d\n", sp.Count())
	// Output: Points: 10
}

func ExampleScrollView() {
	child := component.NewText("A very long text that needs scrolling...")
	sv := component.NewScrollView(child)
	fmt.Printf("Has child: %v\n", sv != nil)
	// Output: Has child: true
}

func ExampleSplitPane() {
	left := component.NewText("Left pane")
	right := component.NewText("Right pane")
	sp := component.NewSplitPane(left, right)
	fmt.Printf("Has panes: %v\n", sp != nil)
	// Output: Has panes: true
}

func ExampleViewport() {
	content := component.NewText("Scrollable content")
	vp := component.NewViewport(content)
	fmt.Printf("Has content: %v\n", vp != nil)
	// Output: Has content: true
}

func ExampleSelectionList() {
	sl := component.NewSelectionList([]string{"Apple", "Banana", "Cherry"})
	sl.Toggle(1) // select Banana
	selected := sl.SelectedItems()
	fmt.Println(selected)
	// Output: [1]
}

func ExampleNumberInput() {
	ni := component.NewNumberInput(5, 0, 100)
	fmt.Println(ni.Value())
	// Output: 5
}

func ExampleOTPInput() {
	otp := component.NewOTPInput(6)
	otp.SetValue("123456")
	fmt.Printf("Length: %d, Value: %s\n", otp.Length(), otp.Value())
	// Output: Length: 6, Value: 123456
}

func ExampleLineGauge() {
	lg := component.NewLineGauge()
	lg.SetPercent(0.75)
	lg.SetLabel("Progress")
	fmt.Printf("%.0f%%\n", lg.Percent()*100)
	// Output: 75%
}

func ExampleRichLog() {
	rl := component.NewRichLog()
	rl.SetMaxSize(100)
	rl.Info("Server started")
	rl.Warn("High memory usage")
	rl.Error("Connection refused")
	fmt.Printf("Entries: %d\n", rl.EntryCount())
	// Output: Entries: 3
}

// ─── P427: 10 more Examples (Direction E) ───

func ExampleTabBar() {
	tb := component.NewTabBar()
	tb.AddTab("tab1", "Files")
	tb.AddTab("tab2", "Edit")
	tb.AddTab("tab3", "View")
	fmt.Printf("Tabs: %d\n", tb.TabCount())
	// Output: Tabs: 3
}

func ExampleVirtualScroller() {
	vs := component.NewVirtualScroller()
	vs.AddItem(component.VirtualItem{ID: "1", Text: "Item 1"})
	vs.AddItem(component.VirtualItem{ID: "2", Text: "Item 2"})
	vs.AddItem(component.VirtualItem{ID: "3", Text: "Item 3"})
	fmt.Printf("Items: %d\n", vs.ItemCount())
	// Output: Items: 3
}

func ExampleStatusIndicator() {
	si := component.NewStatusIndicator()
	si.SetMessage("Connecting...")
	si.Start()
	fmt.Printf("Running: %v\n", si.IsRunning())
	// Output: Running: true
}

func ExampleBadgeGroup() {
	bg := component.NewBadgeGroup()
	bg.Add(component.NewNeutralBadge("v1.0"))
	bg.Add(component.NewSuccessBadge("stable"))
	bg.Add(component.NewWarningBadge("beta"))
	fmt.Printf("Badges: %d\n", bg.Count())
	// Output: Badges: 3
}

func ExampleStyleSheet() {
	ss := component.NewStyleSheet()
	bold := true
	ss.Add("primary", component.StyleDecl{Bold: &bold})
	fmt.Printf("Classes: %d\n", ss.Count())
	// Output: Classes: 1
}

func ExamplePages() {
	p := component.NewPages()
	p.AddPage("home", component.NewText("Home"))
	p.AddPage("settings", component.NewText("Settings"))
	p.SwitchTo("settings")
	fmt.Printf("Current: %s\n", p.CurrentPage())
	// Output: Current: settings
}

func ExampleStreamingText() {
	st := component.NewStreamingText()
	st.SetText("Hello, world!")
	st.Skip() // reveal all text instantly
	fmt.Printf("Completed: %v\n", st.Completed())
	// Output: Completed: true
}

func ExampleTerminalProfile() {
	tp := component.NewTerminalProfile()
	fmt.Printf("Has profile: %v\n", tp != nil)
	// Output: Has profile: true
}

func ExampleDebugInspector() {
	di := component.NewDebugInspector()
	fmt.Printf("Has inspector: %v\n", di != nil)
	// Output: Has inspector: true
}

func ExampleNewSeparator() {
	sep := component.NewSeparator()
	fmt.Printf("Is separator: %v\n", sep != nil)
	// Output: Is separator: true
}

func ExampleModelBadge() {
	mb := component.NewModelBadge("claude-sonnet-4-20250514")
	mb.SetContextWindow(200000)
	fmt.Printf("%s: %s (%s)\n", mb.ProviderName(), mb.DisplayName(), mb.ModelID())
	// Output: Anthropic: Claude (claude-sonnet-4-20250514)
}

// ─── P434: 10 more Examples (Direction E) ───

func ExampleNewConfirmDialog() {
	d := component.NewConfirmDialog("Delete File", "Are you sure you want to delete this file?")
	fmt.Println(d.Title())
	// Output: Delete File
}

func ExampleNewInfoDialog() {
	d := component.NewInfoDialog("About", "Fluui v1.0.0-beta.1")
	fmt.Println(d.Title())
	// Output: About
}

func ExampleNewPromptDialog() {
	d := component.NewPromptDialog("Rename", "Enter new name:", "untitled.txt")
	fmt.Println(d.Message())
	// Output: Enter new name:
}

func ExampleLoadingIndicator() {
	li := component.NewLoadingIndicator("Loading data...")
	li.Start()
	fmt.Println(li.Text())
	// Output: Loading data...
}

func ExampleParagraph() {
	p := component.NewParagraph("This is a paragraph of text that wraps automatically.")
	fmt.Printf("Has content: %v\n", p != nil)
	// Output: Has content: true
}

func ExampleReactiveInt() {
	ri := component.NewReactiveInt(42)
	fmt.Println(ri.Get())
	// Output: 42
}

func ExampleReactiveString() {
	rs := component.NewReactiveString("hello")
	fmt.Println(rs.Get())
	// Output: hello
}

func ExampleReactiveBool() {
	rb := component.NewReactiveBool(true)
	fmt.Println(rb.Get())
	// Output: true
}

func ExampleMenuBar() {
	mb := component.NewMenuBar([]component.Menu{
		{ID: "file", Title: "File"},
		{ID: "edit", Title: "Edit"},
		{ID: "view", Title: "View"},
	})
	fmt.Printf("Menus: %d\n", len(mb.Menus()))
	// Output: Menus: 3
}

func ExampleLinkManager() {
	lm := component.NewLinkManager()
	fmt.Printf("Has links: %v\n", lm != nil)
	// Output: Has links: true
}

// ─── P437: 10 more Examples (Direction E) ───

func ExamplePopup() {
	p := component.NewPopup(component.NewText("Popup content"))
	fmt.Printf("Has popup: %v\n", p != nil)
	// Output: Has popup: true
}

func ExampleKeybindingManager() {
	km := component.NewKeybindingManager()
	fmt.Printf("Has keybindings: %v\n", km != nil)
	// Output: Has keybindings: true
}

func ExampleHelpOverlay() {
	groups := []component.HelpGroup{
		{Name: "Navigation", Entries: []component.HelpEntry{{Keys: "j/k", Description: "Move up/down"}}},
	}
	ho := component.NewHelpOverlay(groups)
	fmt.Printf("Has help: %v\n", ho != nil)
	// Output: Has help: true
}

func ExampleSessionSidebar() {
	sb := component.NewSessionSidebar()
	fmt.Printf("Has sidebar: %v\n", sb != nil)
	// Output: Has sidebar: true
}

func ExampleWindowManager() {
	wm := component.NewWindowManager(component.NewText("Main panel"))
	fmt.Printf("Panes: %d\n", wm.PaneCount())
	// Output: Panes: 1
}

func ExampleRadarChart() {
	rc := component.NewRadarChart([]component.RadarAxis{
		{Label: "Speed", Max: 100},
		{Label: "Power", Max: 100},
		{Label: "Range", Max: 100},
		{Label: "Quality", Max: 100},
	})
	fmt.Printf("Has chart: %v\n", rc != nil)
	// Output: Has chart: true
}

func ExampleNewQuestionnaireDialog() {
	d := component.NewQuestionnaireDialog("Survey", []component.Question{
		{ID: "q1", Text: "How satisfied are you?", Required: true},
	})
	fmt.Println(d.Title())
	// Output: Survey
}

func ExampleNewPrettyString() {
	ps := component.NewPrettyString("hello world")
	fmt.Printf("Has string: %v\n", ps != nil)
	// Output: Has string: true
}

func ExampleSentimentBar() {
	sb := component.NewSentimentBar(0.72)
	sb.SetConfidence(0.85)
	fmt.Println(sb.Label())
	// Output: positive
}

// ─── P447: Chart component examples ───

func ExampleGanttChart() {
	gc := component.NewGanttChart()
	gc.AddTask(component.GanttTask{Label: "Design", Start: 0, End: 14})
	gc.AddTask(component.GanttTask{Label: "Build", Start: 10, End: 40})
	fmt.Printf("Tasks: %d\n", gc.TaskCount())
	// Output: Tasks: 2
}

func ExampleWaterfallChart() {
	wc := component.NewWaterfallChart()
	wc.AddBar(component.WaterfallBar{Label: "Start", Value: 100, Type: component.WaterfallStart})
	wc.AddBar(component.WaterfallBar{Label: "Rev", Value: 40, Type: component.WaterfallPositive})
	fmt.Printf("Bars: %d\n", wc.BarCount())
	// Output: Bars: 2
}

func ExampleSunburstChart() {
	sc := component.NewSunburstChart()
	sc.AddSegment(component.SunburstSegment{Label: "A", Value: 40})
	sc.AddSegment(component.SunburstSegment{Label: "B", Value: 60})
	fmt.Printf("Total: %.0f\n", sc.TotalValue())
	// Output: Total: 100
}

func ExampleCandlestickChart() {
	cc := component.NewCandlestickChart()
	cc.AddCandle(component.Candle{Open: 100, High: 105, Low: 98, Close: 103})
	fmt.Printf("Candles: %d\n", cc.CandleCount())
	// Output: Candles: 1
}

func ExampleNetworkGraph() {
	ng := component.NewNetworkGraph()
	ng.AddNode(component.GraphNode{ID: "srv", Label: "Server"})
	ng.AddNode(component.GraphNode{ID: "db", Label: "DB"})
	ng.AddEdge(component.GraphEdge{From: "srv", To: "db"})
	fmt.Printf("Nodes: %d Edges: %d\n", ng.NodeCount(), ng.EdgeCount())
	// Output: Nodes: 2 Edges: 1
}

func ExampleBubbleChart() {
	bc := component.NewBubbleChart()
	bc.AddBubble(component.BubbleData{X: 10, Y: 20, Size: 5, Label: "A"})
	bc.AddBubble(component.BubbleData{X: 30, Y: 50, Size: 10, Label: "B"})
	fmt.Printf("Bubbles: %d\n", bc.BubbleCount())
	// Output: Bubbles: 2
}

func ExampleStreamingMarkdownDiff() {
	d := component.NewStreamingMarkdownDiff()
	d.SetOld("hello\nworld")
	d.SetNew("hello\nGo")
	added, removed, _ := d.Stats()
	fmt.Printf("+%d -%d\n", added, removed)
	// Output: +1 -1
}

func ExampleProgressTimeline() {
	pt := component.NewProgressTimeline()
	pt.AddMilestone(component.TimelineMilestone{Label: "Design", Status: component.MilestoneDone})
	pt.AddMilestone(component.TimelineMilestone{Label: "Build", Status: component.MilestoneActive})
	pt.AddMilestone(component.TimelineMilestone{Label: "Ship", Status: component.MilestonePending})
	fmt.Printf("Progress: %.0f%%\n", pt.Progress()*100)
	// Output: Progress: 33%
}

func ExampleOrgChart() {
	oc := component.NewOrgChart()
	oc.SetRoot(component.OrgNode{ID: "ceo", Label: "CEO"})
	oc.AddChild("ceo", component.OrgNode{ID: "cto", Label: "CTO"})
	fmt.Printf("Nodes: %d\n", oc.NodeCount())
	// Output: Nodes: 2
}

func ExampleStockTicker() {
	st := component.NewStockTicker("AAPL", 189.50, 1.25)
	fmt.Printf("%s: %.2f (%+.2f)\n", st.Symbol(), st.Price(), st.Change())
	// Output: AAPL: 189.50 (+1.25)
}

func ExampleAIStreamRenderer() {
	r := component.NewAIStreamRenderer()
	r.StartWithModel("gpt-4o")
	r.Append("Hello world")
	r.SetTokens(10, 25.0)
	fmt.Printf("Status: %d Tokens: %d\n", r.Status(), r.TokenCount())
	// Output: Status: 2 Tokens: 10
}

func ExampleHeatmapGrid() {
	hg := component.NewHeatmapGrid(7, 20)
	hg.Set(0, 0, 5)
	hg.Set(0, 1, 12)
	fmt.Printf("Cells: %d Filled: %d\n", hg.CellCount(), hg.FilledCount())
	// Output: Cells: 140 Filled: 2
}

func ExampleTreemapChart() {
	tc := component.NewTreemapChart()
	tc.AddNode(component.TreemapNode{Label: "Docs", Value: 40})
	tc.AddNode(component.TreemapNode{Label: "Photos", Value: 30})
	fmt.Printf("Total: %.0f\n", tc.TotalValue())
	// Output: Total: 70
}

func ExampleCodeBlockStream() {
	cbs := component.NewCodeBlockStream("go")
	cbs.SetCode("x := 42")
	fmt.Printf("Lines: %d Lang: %s\n", cbs.LineCount(), cbs.Language())
	// Output: Lines: 1 Lang: go
}

func ExampleTokenMeter() {
	tm := component.NewTokenMeter(128000)
	tm.SetUsed(45000)
	fmt.Printf("Used: %d%% Remaining: %d\n", int(tm.Percent()), tm.Remaining())
	// Output: Used: 35% Remaining: 83000
}

func ExampleStopReasonBadge() {
	sr := component.NewStopReasonBadge(component.StopReasonStop)
	fmt.Printf("%s: %c\n", sr.ReasonText(), sr.ReasonIcon())
	// Output: stop: ✓
}

func ExampleThinkingTrace() {
	tt := component.NewThinkingTrace()
	tt.Start()
	tt.Append("Analyzing input...")
	tt.Complete()
	fmt.Printf("State: %d Collapsed: %v\n", tt.State(), tt.IsCollapsed())
	// Output: State: 2 Collapsed: true
}


// ExampleCostTracker demonstrates cost tracking for AI API usage.
func ExampleCostTracker() {
	ct := component.NewCostTracker()
	ct.SetPricing(15.0, 60.0) // $15/M input, $60/M output
	ct.AddTokens(50000, 12000)
	ct.SetBudget(5.0)
	fmt.Printf("In:%d Out:%d Total:%d Cost:$%.4f Budget:$%.2f OverBudget:%v\n",
		ct.InputTokens(), ct.OutputTokens(), ct.TotalTokens(),
		ct.Cost(), ct.Budget(), ct.IsOverBudget())
	// Output: In:50000 Out:12000 Total:62000 Cost:$1.4700 Budget:$5.00 OverBudget:false
}

// ExampleResponseInspector demonstrates AI response metadata inspection.
func ExampleResponseInspector() {
	ri := component.NewResponseInspector()
	ri.SetModel("gpt-4o")
	ri.SetTokens(120, 350)
	ri.SetFinishReason(component.FinishStop)
	fmt.Printf("Model:%s Tokens:%d Finish:%s\n",
		ri.Model(), ri.TotalTokens(), ri.FinishReason())
	// Output: Model:gpt-4o Tokens:470 Finish:stop
}

// ExampleContextWindowBar demonstrates context window usage visualization.
func ExampleContextWindowBar() {
	cwb := component.NewContextWindowBar()
	cwb.SetContextLimit(128000)
	cwb.SetUsed(95000)
	fmt.Printf("Used:%d Limit:%d Pct:%.1f%%\n",
		cwb.Used(), cwb.ContextLimit(), cwb.UsagePercent())
	// Output: Used:95000 Limit:128000 Pct:74.2%
}

// ExampleRateLimitIndicator demonstrates API rate limit status display.
func ExampleRateLimitIndicator() {
	rl := component.NewRateLimitIndicator()
	rl.SetLimit(5000)
	rl.SetRemaining(3200)
	fmt.Printf("Remaining:%d Limit:%d Used:%.0f%% Limited:%v\n",
		rl.Remaining(), rl.Limit(), rl.UsagePercent(), rl.IsRateLimited())
	// Output: Remaining:3200 Limit:5000 Used:36% Limited:false
}

// ExampleSankeyChart demonstrates flow diagram visualization.
func ExampleSankeyChart() {
	sc := component.NewSankeyChart()
	sc.AddFlow("Revenue", "Marketing", 500)
	sc.AddFlow("Revenue", "Engineering", 800)
	fmt.Printf("Sources:%v Targets:%v Flows:%d\n",
		sc.Sources(), sc.Targets(), len(sc.Flows()))
	// Output: Sources:[Revenue] Targets:[Marketing Engineering] Flows:2
}

// ExampleScatterPlot demonstrates 2D scatter plot visualization.
func ExampleScatterPlot() {
	sp := component.NewScatterPlot()
	sp.SetXRange(0, 100)
	sp.SetYRange(0, 100)
	sp.AddPoint(10, 20)
	sp.AddPoint(50, 80)
	xMin, xMax := sp.XRange()
	yMin, yMax := sp.YRange()
	fmt.Printf("Points:%d X:(%.0f,%.0f) Y:(%.0f,%.0f)\n",
		sp.PointCount(), xMin, xMax, yMin, yMax)
	// Output: Points:2 X:(0,100) Y:(0,100)
}

// ExampleMergeView demonstrates side-by-side diff/merge visualization.
func ExampleMergeView() {
	mv := component.NewMergeView()
	mv.SetLeft("ours", "same\nold\nsame")
	mv.SetRight("theirs", "same\nnew\nsame")
	fmt.Printf("Left:%s Right:%s Lines:%d Conflicts:%v\n",
		mv.LeftLabel(), mv.RightLabel(), mv.LineCount(), mv.HasConflicts())
	// Output: Left:ours Right:theirs Lines:4 Conflicts:false
}

// ExampleFunctionCallVisualizer demonstrates AI tool call chain display.
func ExampleFunctionCallVisualizer() {
	fcv := component.NewFunctionCallVisualizer()
	fcv.AddCall("search_web", `{"q":"go tui"}`, 120000000, component.CallSuccess)
	fcv.AddCall("read_file", `{"path":"main.go"}`, 5000000, component.CallError)
	fmt.Printf("Calls:%d\n", fcv.CallCount())
	// Output: Calls:2
}

// ExampleCodeEditor demonstrates syntax-highlighted code display.
func ExampleCodeEditor() {
	ce := component.NewCodeEditor()
	ce.SetLanguage("go")
	ce.SetCode("package main\nfunc main() {}")
	fmt.Printf("Lang:%s Lines:%d Cursor:%d\n",
		ce.Language(), ce.LineCount(), ce.CursorLine())
	// Output: Lang:go Lines:2 Cursor:-1
}

// ExampleStreamProgressIndicator demonstrates AI streaming progress display.
func ExampleStreamProgressIndicator() {
	sp := component.NewStreamProgressIndicator()
	sp.SetExpected(500)
	sp.Start()
	sp.AddTokens(250)
	fmt.Printf("Tokens:%d Pct:%.0f%% State:%d\n",
		sp.TokensReceived(), sp.Percent(), sp.State())
	// Output: Tokens:250 Pct:50% State:1
}

// ExampleAIConfidenceBar demonstrates AI confidence score visualization.
func ExampleAIConfidenceBar() {
	cb := component.NewAIConfidenceBar()
	cb.SetLabel("Prediction")
	cb.SetConfidence(85.5)
	fmt.Printf("Label:%s Confidence:%.1f\n", cb.Label(), cb.Confidence())
	// Output: Label:Prediction Confidence:85.5
}

// ExampleMarkdownTable demonstrates markdown pipe table rendering.
func ExampleMarkdownTable() {
	mt := component.NewMarkdownTable()
	mt.SetMarkdown(`| Name | Age |
|------|-----|
| Alice | 30 |
| Bob | 25 |`)
	fmt.Printf("Rows:%d Cols:%d\n", mt.RowCount(), mt.ColumnCount())
	// Output: Rows:2 Cols:2
}

// ExampleMarkdownBlockquote demonstrates blockquote rendering.
func ExampleMarkdownBlockquote() {
	bq := component.NewMarkdownBlockquote()
	bq.SetMarkdown("> To be or not to be.\n>> — Shakespeare")
	fmt.Printf("Lines:%d\n", bq.LineCount())
	// Output: Lines:2
}

// ExampleCarousel demonstrates paginated content navigation.
func ExampleCarousel() {
	c := component.NewCarousel()
	c.AddSlide("Welcome", "Get started!")
	c.AddSlide("Features", "160+ components")
	c.Next()
	fmt.Printf("Slides:%d Current:%d\n", c.SlideCount(), c.CurrentIndex())
	// Output: Slides:2 Current:1
}

// ExampleMarkdownTaskList demonstrates task list rendering.
func ExampleMarkdownTaskList() {
	tl := component.NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Pending\n- [x] Done")
	tl.ToggleItem(0)
	fmt.Printf("Completed:%d Total:%d\n", tl.CompletedCount(), tl.TaskCount())
	// Output: Completed:2 Total:2
}

// ExampleMarkdownHorizontalRule demonstrates horizontal rule rendering.
func ExampleMarkdownHorizontalRule() {
	hr := component.NewMarkdownHorizontalRule()
	hr.SetMarkdown("Above\n---\nBelow")
	fmt.Printf("Lines:%d Rules:%d\n", hr.LineCount(), hr.RuleCount())
	// Output: Lines:3 Rules:1
}

// ExampleMarkdownInlineCode demonstrates inline code and fenced block rendering.
func ExampleMarkdownInlineCode() {
	mic := component.NewMarkdownInlineCode()
	mic.SetMarkdown("Use `fmt.Println` to print.\n```go\nfunc main() {}\n```")
	fmt.Printf("Inline:%d Blocks:%d\n", mic.InlineCodeCount(), mic.CodeBlockCount())
	// Output: Inline:1 Blocks:2
}

// ExampleStepProgress demonstrates multi-step progress indicator.
func ExampleStepProgress() {
	sp := component.NewStepProgress()
	sp.AddStep("Account")
	sp.AddStep("Profile")
	sp.AddStep("Confirm")
	sp.SetCurrentStep(1)
	fmt.Printf("Steps:%d Current:%d\n", sp.StepCount(), sp.CurrentStep())
	// Output: Steps:3 Current:1
}

// ExampleMarkdownStrikethrough demonstrates strikethrough rendering.
func ExampleMarkdownStrikethrough() {
	ms := component.NewMarkdownStrikethrough()
	ms.SetMarkdown("~~deprecated~~ new API")
	fmt.Printf("Struck:%d\n", ms.StrikethroughCount())
	// Output: Struck:1
}

// ExampleMarkdownEmphasis demonstrates bold/italic rendering.
func ExampleMarkdownEmphasis() {
	me := component.NewMarkdownEmphasis()
	me.SetMarkdown("**bold** and *italic*")
	fmt.Printf("Bold:%d Italic:%d\n", me.BoldCount(), me.ItalicCount())
	// Output: Bold:1 Italic:1
}

// ExampleMarkdownList demonstrates ordered/unordered list rendering.
func ExampleMarkdownList() {
	ml := component.NewMarkdownList()
	ml.SetMarkdown("- Apple\n- Banana")
	fmt.Printf("Items:%d Type:%s\n", ml.ItemCount(), ml.ListType())
	// Output: Items:2 Type:unordered
}

// ExampleBreadcrumbTrail demonstrates path breadcrumb navigation.
func ExampleBreadcrumbTrail() {
	bt := component.NewBreadcrumbTrail()
	bt.AddCrumb("Home")
	bt.AddCrumb("Docs")
	bt.AddCrumb("API")
	fmt.Printf("Crumbs:%d\n", bt.CrumbCount())
	// Output: Crumbs:3
}

// ExampleNotificationStack demonstrates notification stack rendering.
func ExampleNotificationStack() {
	ns := component.NewNotificationStack()
	ns.AddNotification("Build", "OK", component.NotifSuccess)
	ns.AddNotification("Warn", "Deprecated", component.NotifWarning)
	fmt.Printf("Count:%d\n", ns.Count())
	// Output: Count:2
}

// ExampleImagePreview demonstrates image preview placeholder.
func ExampleImagePreview() {
	ip := component.NewImagePreview()
	ip.SetFormat("PNG")
	ip.SetDimensions(800, 600)
	ip.SetLabel("photo.png")
	w, h := ip.Dimensions()
	fmt.Printf("Format:%s %dx%d\n", ip.Format(), w, h)
	// Output: Format:PNG 800x600
}

// ExampleMarkdownLink demonstrates hyperlink rendering.
func ExampleMarkdownLink() {
	ml := component.NewMarkdownLink()
	ml.SetMarkdown("Visit [Fluui](https://fluui.dev) now.")
	fmt.Printf("Links:%d\n", ml.LinkCount())
	// Output: Links:1
}

// ExampleTagCloud demonstrates weighted tag cloud.
func ExampleTagCloud() {
	tc := component.NewTagCloud()
	tc.AddTag("go", 80)
	tc.AddTag("tui", 50)
	fmt.Printf("Tags:%d\n", tc.TagCount())
	// Output: Tags:2
}

// ExampleMarkdownFootnote demonstrates footnote rendering.
func ExampleMarkdownFootnote() {
	mf := component.NewMarkdownFootnote()
	mf.SetMarkdown("See[^1] for details.\n\n[^1]: https://example.com")
	fmt.Printf("Footnotes:%d\n", mf.FootnoteCount())
	// Output: Footnotes:1
}

// ExampleLegend demonstrates chart legend display.
func ExampleLegend() {
	l := component.NewLegend()
	l.AddEntry("Revenue", buffer.RGB(34, 197, 94))
	l.AddEntry("Costs", buffer.RGB(239, 68, 68))
	fmt.Printf("Entries:%d\n", l.EntryCount())
	// Output: Entries:2
}
