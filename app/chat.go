// Package app implements the high-level ChatApp API for building
// AI-native terminal chat interfaces.
//
// ChatApp wraps the block system, component tree, overlay manager,
// and event loop into a single cohesive interface. Users create blocks,
// stream deltas, and the app handles layout, rendering, and input.
package app

import (
	"fmt"
	"sync"
	"time"

	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/component/layout"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
	"github.com/topcheer/fluui/theme"
)

// ChatApp is a high-level chat interface that manages a scrolling
// list of AI content blocks with an optional overlay layer.
type ChatApp struct {
	mu sync.Mutex

	// Core components
	container   *block.BlockContainer
	scrollView  *component.ScrollView
	root        component.Component // scroll view or column(scroll + input)
	overlays    *overlay.OverlayManager

	// State
	width       int
	height      int
	inputHeight int // reserved for input line (0 = no input line)
	paddingX    int // horizontal padding (applied left + right)
	paddingTop  int // top padding

	// Theme
	theme     *theme.Theme
	bgColor  buffer.Color
	fgColor  buffer.Color

	// Theme cycling
	themeToast    string // temporary theme name display
	themeToastAt  time.Time

	// Callbacks
	onStreamDelta func(delta string)
	onKey         func(*term.KeyEvent)
	onMouse       func(*term.MouseEvent)
	onClipboard   func(text string)
	onQuit        func()

	// Input line (nil if not configured)
	inputLine *InputLine

	// AI bridge (nil if not configured)
	aiBridge *AIBridge

	// Clipboard config (nil = raw OSC52 fallback)
	clipboardConfig *ClipboardConfig

	// Search (nil = not created yet, created on first Ctrl+F)
	search *SearchMode

	// P16: StatusBar integration (nil = not attached)
	statusBar       *component.StatusBar
	statusBarHeight int

	// P16: TabBar integration (nil = not attached)
	tabBar       *component.TabBar
	tabBarHeight int

	// P16: SelectionManager integration (nil = not attached)
	selectionMgr *SelectionManager

	// P20: CommandPalette + Spinner integration (nil = not attached)
	commandPalette *component.CommandPalette
	spinner        *component.Spinner
}

// SetTheme updates the active theme. All components that reference
// theme.Active will pick up the new colors on the next render.
func (a *ChatApp) SetTheme(t *theme.Theme) {
	if t == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.theme = t
	a.bgColor = t.Bg
	a.fgColor = t.Fg
	theme.SetActive(t)
}

// CycleTheme advances to the next built-in theme and shows a toast.
func (a *ChatApp) CycleTheme() *theme.Theme {
	next := theme.Cycle()
	a.mu.Lock()
	a.theme = next
	a.bgColor = next.Bg
	a.fgColor = next.Fg
	a.themeToast = next.Name
	a.themeToastAt = time.Now()
	a.mu.Unlock()
	return next
}

// CycleThemeBack goes to the previous built-in theme and shows a toast.
func (a *ChatApp) CycleThemeBack() *theme.Theme {
	prev := theme.CycleBack()
	a.mu.Lock()
	a.theme = prev
	a.bgColor = prev.Bg
	a.fgColor = prev.Fg
	a.themeToast = prev.Name
	a.themeToastAt = time.Now()
	a.mu.Unlock()
	return prev
}

// ThemeToast returns the current theme toast text and whether it's still visible.
// The toast auto-expires after 3 seconds.
func (a *ChatApp) ThemeToast() (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.themeToast == "" {
		return "", false
	}
	if time.Since(a.themeToastAt) > 3*time.Second {
		return "", false
	}
	return a.themeToast, true
}

// Theme returns the current theme.
func (a *ChatApp) Theme() *theme.Theme {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.theme == nil {
		return theme.Default()
	}
	return a.theme
}

// NewChatApp creates a ChatApp with sensible defaults.
func NewChatApp(width, height int) *ChatApp {
	container := block.NewBlockContainer()
	container.SetSpacing(1)

	scrollView := component.NewScrollView(container)

	return &ChatApp{
		container:   container,
		scrollView:  scrollView,
		overlays:    overlay.NewOverlayManager(),
		width:       width,
		height:      height,
		inputHeight: 0,
		paddingX:    2,
		paddingTop:  1,
		theme:       theme.Dracula(),
		bgColor:     theme.Dracula().Bg,
		fgColor:     theme.Dracula().Fg,
	}
}

// Container returns the block container for direct manipulation.
// The pointer is stable (set once at construction); callers must not
// modify the container concurrently with Render/HandleKey/StreamDelta.
func (a *ChatApp) Container() *block.BlockContainer { return a.container }

// ScrollView returns the scroll view.
// The pointer is stable; callers must not modify concurrently.
func (a *ChatApp) ScrollView() *component.ScrollView { return a.scrollView }

// Overlays returns the overlay manager.
// The pointer is stable; callers must not modify concurrently.
func (a *ChatApp) Overlays() *overlay.OverlayManager { return a.overlays }

// SetInputHeight reserves space at the bottom for an input line.
func (a *ChatApp) SetInputHeight(h int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inputHeight = h
}

// SetInputLine attaches an InputLine component to this ChatApp.
// When set, key events are routed to the InputLine first, and the
// Render method paints it in the reserved input area.
func (a *ChatApp) SetInputLine(line *InputLine) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inputLine = line
	if a.inputHeight == 0 {
		a.inputHeight = 2 // separator + input row
	}
}

// InputLine returns the attached InputLine, or nil if none.
func (a *ChatApp) InputLine() *InputLine {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.inputLine
}

// OnSubmit sets a handler invoked when the user presses Enter in the InputLine.
// This is a convenience wrapper around SetInputLine + InputLine.onSubmit.
// If no InputLine is attached, a default one with prompt "> " is created.
func (a *ChatApp) OnSubmit(fn func(text string)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.inputLine == nil {
		a.inputLine = NewInputLine("> ")
		if a.inputHeight == 0 {
			a.inputHeight = 2
		}
	}
	a.inputLine.onSubmit = fn
}

// SetSize updates the terminal dimensions.
func (a *ChatApp) SetSize(w, h int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.width = w
	a.height = h
}

// Size returns the current terminal dimensions.
func (a *ChatApp) Size() (int, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.width, a.height
}

// AddUserMessage adds a user message block and returns it.
func (a *ChatApp) AddUserMessage(text string) *block.UserMessageBlock {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := fmt.Sprintf("user-%d", a.container.Len())
	msg := block.NewUserMessageBlock(id, text)
	a.container.AddBlock(msg)
	return msg
}

// AddThinking adds a new thinking block and returns it.
func (a *ChatApp) AddThinking() *block.ThinkingBlock {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := fmt.Sprintf("think-%d", a.container.Len())
	tb := block.NewThinkingBlock(id)
	a.container.AddBlock(tb)
	return tb
}

// AddAssistantText adds a new assistant text block and returns it.
func (a *ChatApp) AddAssistantText() *block.AssistantTextBlock {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := fmt.Sprintf("asst-%d", a.container.Len())
	at := block.NewAssistantTextBlock(id)
	a.container.AddBlock(at)
	return at
}

// AddToolCall adds a new tool call block and returns it.
func (a *ChatApp) AddToolCall(toolName, rawArgs string) *block.ToolCallBlock {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := fmt.Sprintf("tc-%d", a.container.Len())
	tc := block.NewToolCallBlock(id, toolName, rawArgs)
	a.container.AddBlock(tc)
	return tc
}

// AddToolResult adds a new tool result block and returns it.
func (a *ChatApp) AddToolResult() *block.ToolResultBlock {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := fmt.Sprintf("tr-%d", a.container.Len())
	tr := block.NewToolResultBlock(id)
	a.container.AddBlock(tr)
	return tr
}

// StreamDelta dispatches a streaming delta to the last block of the
// matching type (creating one if needed). This is the main streaming API.
func (a *ChatApp) StreamDelta(delta block.StreamDelta) {
	a.mu.Lock()
	defer a.mu.Unlock()

	dispatcher := block.NewStreamDispatcher(a.container)
	_ = dispatcher.Dispatch(delta)
}

// ScrollUp scrolls the content up by one line.
func (a *ChatApp) ScrollUp() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scrollView.ScrollUp(1)
}

// ScrollDown scrolls the content down by one line.
func (a *ChatApp) ScrollDown() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scrollView.ScrollDown(1)
}

// ScrollToBottom scrolls to the most recent content.
// This re-measures content height to get the correct maxOffset,
// since blocks may have grown since the last render.
func (a *ChatApp) ScrollToBottom() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scrollToBottomLocked()
}

// scrollToBottomLocked is the unlocked internal version.
// Caller must hold a.mu.
func (a *ChatApp) scrollToBottomLocked() {
	if a.width > 0 {
		contentH := a.height - a.inputHeight
		if contentH < 1 {
			contentH = 1
		}
		contentW := a.width - a.paddingX*2
		if contentW < 1 {
			contentW = 1
		}
		a.scrollView.Measure(component.Constraints{MaxWidth: contentW, MaxHeight: contentH})
	}
	a.scrollView.ScrollTo(a.scrollView.MaxOffset())
}

// OnKey sets a keyboard handler.
func (a *ChatApp) OnKey(fn func(*term.KeyEvent)) {
	a.mu.Lock()
	a.onKey = fn
	a.mu.Unlock()
}

// OnMouse sets a mouse handler.
func (a *ChatApp) OnMouse(fn func(*term.MouseEvent)) {
	a.mu.Lock()
	a.onMouse = fn
	a.mu.Unlock()
}

// OnQuit sets a quit handler.
func (a *ChatApp) OnQuit(fn func()) {
	a.mu.Lock()
	a.onQuit = fn
	a.mu.Unlock()
}

// OnClipboard sets a handler for OSC52 clipboard paste responses.
// When the terminal responds to a paste query, the decoded text is passed
// to this callback. The handler is invoked from the event loop goroutine.
func (a *ChatApp) OnClipboard(fn func(text string)) {
	a.mu.Lock()
	a.onClipboard = fn
	a.mu.Unlock()
}

// HandleClipboard processes a clipboard event (OSC52 paste response).
// Calls the onClipboard callback if registered. Returns true if consumed.
func (a *ChatApp) HandleClipboard(text string) bool {
	a.mu.Lock()
	cb := a.onClipboard
	a.mu.Unlock()

	if cb != nil {
		cb(text)
		return true
	}

	// Default: if InputLine is attached, insert the text at cursor.
	if a.inputLine != nil {
		a.mu.Lock()
		a.inputLine.InsertText(text)
		a.mu.Unlock()
		return true
	}

	return false
}

// HandleKey processes a key event. Returns true if consumed.
// Uses lock-then-copy: extracts shared state under lock, releases lock,
// then processes. This avoids deadlock when callbacks call AddXxx.
func (a *ChatApp) HandleKey(key *term.KeyEvent) bool {
	a.mu.Lock()
	overlays := a.overlays
	inputLine := a.inputLine
	height := a.height
	onQuit := a.onQuit
	onKey := a.onKey
	a.mu.Unlock()

	// Route to overlays first (has its own thread safety)
	if overlays.HandleKey(key) {
		return true
	}

	// Route to InputLine before scroll keys.
	if inputLine != nil {
		if inputLine.HandleKey(key) {
			return true
		}
	}

	// App-level keys
	if key.Key == term.KeyEscape || (key.Rune == 'q' && key.Modifiers == 0) {
		if overlays.HasModal() {
			return false // let modal handle it
		}
		if onQuit != nil {
			onQuit()
		}
		return true
	}

	// Ctrl+T / Ctrl+Shift+T: cycle theme
	// Ctrl+] / Ctrl+\: alternate theme cycling keys
	if a.handleThemeKey(key) {
		return true
	}
	if key.Modifiers&term.ModCtrl != 0 && (key.Rune == 't' || key.Rune == 'T') {
		if key.Rune == 'T' {
			a.CycleThemeBack()
		} else {
			a.CycleTheme()
		}
		return true
	}

	// Ctrl+F: toggle search mode
	if key.Modifiers&term.ModCtrl != 0 && (key.Rune == 'f' || key.Rune == 'F') {
		a.mu.Lock()
		if a.search == nil {
			a.search = NewSearchMode()
		}
		if a.search.IsActive() {
			// Ctrl+F while active = next match
			a.search.NextMatch()
		} else {
			a.search.StartSearch()
		}
		a.mu.Unlock()
		return true
	}

	// Route keys to search mode when active
	a.mu.Lock()
	searchMode := a.search
	a.mu.Unlock()
	if searchMode != nil && searchMode.IsActive() {
		if searchMode.HandleKey(key) {
			// Query changed — recompute matches
			a.mu.Lock()
			blocks := a.container.Blocks()
			a.search.UpdateQuery(a.search.Query(), blocks)
			a.mu.Unlock()
			return true
		}
	}

	// Scroll keys — each locks independently
	switch key.Key {
	case term.KeyUp:
		a.ScrollUp()
		return true
	case term.KeyDown:
		a.ScrollDown()
		return true
	case term.KeyHome:
		a.mu.Lock()
		a.scrollView.ScrollTo(0)
		a.mu.Unlock()
		return true
	case term.KeyEnd:
		a.ScrollToBottom()
		return true
	case term.KeyPageUp:
		a.mu.Lock()
		a.scrollView.ScrollUp(height)
		a.mu.Unlock()
		return true
	case term.KeyPageDown:
		a.mu.Lock()
		a.scrollView.ScrollDown(height)
		a.mu.Unlock()
		return true
	}

	// Custom handler (called without holding lock)
	if onKey != nil {
		onKey(key)
	}
	return true
}

// HandleMouse processes a mouse event. Returns true if consumed.
// Uses lock-then-copy pattern for thread safety.
func (a *ChatApp) HandleMouse(mouse *term.MouseEvent) bool {
	a.mu.Lock()
	overlays := a.overlays
	onMouse := a.onMouse
	a.mu.Unlock()

	// Route to overlays first
	if overlays.HandleMouse(mouse.X, mouse.Y) {
		return true
	}

	// Scroll wheel
	if mouse.Action == term.MouseWheel {
		switch mouse.Button {
		case term.MouseWheelUp:
			a.ScrollUp()
			return true
		case term.MouseWheelDown:
			a.ScrollDown()
			return true
		}
	}

	// Custom handler (called without holding lock)
	if onMouse != nil {
		onMouse(mouse)
	}
	return true
}

// Render draws the chat interface into the given buffer.
// This is called from the app's OnPaint handler.
func (a *ChatApp) Render(buf *buffer.Buffer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	w, h := a.width, a.height
	if w <= 0 || h <= 0 {
		w, h = buf.Width, buf.Height
	}

	// Fill background
	bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: a.bgColor}
	buf.Fill(bgCell)

	// Layout: scroll view takes most of the space, input line at bottom
	contentH := h - a.inputHeight
	if contentH < 1 {
		contentH = 1
	}

	// Apply padding: inset content from edges.
	padX := a.paddingX
	padTop := a.paddingTop
	contentW := w - padX*2
	if contentW < 1 {
		contentW = 1
		padX = 0
	}

	// Measure content with padded width so ScrollView knows true content height.
	a.scrollView.Measure(component.Constraints{MaxWidth: contentW, MaxHeight: contentH})

	a.scrollView.SetBounds(component.Rect{
		X: padX, Y: padTop,
		W: contentW,
		H: contentH - padTop,
	})
	a.scrollView.Paint(buf)

	// Input line area (if reserved)
	if a.inputHeight > 0 {
		a.renderInputLine(buf, w, h)
	}

	// Overlays on top
	a.overlays.Paint(buf)
}

// renderInputLine draws the input prompt at the bottom.
// If an InputLine component is attached, it is laid out and painted here,
// replacing the static prompt.
func (a *ChatApp) renderInputLine(buf *buffer.Buffer, w, h int) {
	// Separator line
	sepY := h - a.inputHeight
	sepStyle := buffer.Style{Fg: theme.Get().Separator}
	for x := 0; x < w; x++ {
		buf.SetCell(x, sepY, buffer.Cell{Rune: '─', Width: 1, Fg: sepStyle.Fg, Bg: a.bgColor})
	}

	// If an InputLine component is attached, lay it out and paint it.
	if a.inputLine != nil {
		inputW := w - 2 // leave 1px padding left + right
		if inputW < 1 {
			inputW = 1
		}
		a.inputLine.SetBounds(component.Rect{
			X: 1, Y: h - 1,
			W: inputW,
			H: 1,
		})
		a.inputLine.Paint(buf)
		return
	}

	// Fallback: static prompt when no InputLine component is set.
	promptStyle := buffer.Style{
		Fg:    theme.Get().PromptFg,
		Bg:    a.bgColor,
		Flags: buffer.Bold,
	}
	buf.DrawText(1, h-1, "▶ ", promptStyle)
}

// LastBlockText returns the text content of the last meaningful block
// (user message, assistant text, thinking, tool call, or tool result).
// Returns "" and false if no suitable block is found.
func (a *ChatApp) LastBlockText() (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	blocks := a.container.Blocks()
	if len(blocks) == 0 {
		return "", false
	}

	// Walk backwards to find the last block with extractable content.
	for i := len(blocks) - 1; i >= 0; i-- {
		text, ok := extractBlockText(blocks[i])
		if ok && text != "" {
			return text, true
		}
	}
	return "", false
}

// IsDirty returns true if the app needs re-rendering.
func (a *ChatApp) IsDirty() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.container.IsDirty()
}

// ClearDirty clears all dirty flags.
func (a *ChatApp) ClearDirty() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.container.ClearDirty()
}

// Clear removes all blocks.
func (a *ChatApp) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Remove all blocks by creating a new container
	for a.container.Len() > 0 {
		blocks := a.container.Blocks()
		a.container.RemoveBlock(blocks[0].ID())
	}
}

// Root returns the root component (for integration with event loop).
func (a *ChatApp) Root() component.Component {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.scrollView
}

// SetRootFlex wraps the scroll view in a flex layout (for custom layouts).
func (a *ChatApp) SetRootFlex(column *layout.Flex) {
	a.mu.Lock()
	a.root = column
	a.mu.Unlock()
}

// --- AI Bridge integration ---

// bridge is the AI streaming bridge (nil if AI is not configured).
// Use SetAIClient to initialize it.
func (a *ChatApp) bridge() *AIBridge {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aiBridge
}

// SetAIClient configures the ChatApp to use an AI client for streaming chat.
// This enables SendUserMessage and StopStreaming.
func (a *ChatApp) SetAIClient(client *ai.Client) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aiBridge = NewAIBridge(a, client)
}

// SetSystemPrompt sets the system prompt for AI conversations.
func (a *ChatApp) SetSystemPrompt(prompt string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aiBridge != nil {
		a.aiBridge.SetSystemPrompt(prompt)
	}
}

// SetOnAIError sets a callback for AI streaming errors.
func (a *ChatApp) SetOnAIError(fn func(err error)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aiBridge != nil {
		a.aiBridge.SetOnError(fn)
	}
}

// SendUserMessage sends a user message to the AI and streams the response
// into content blocks. Requires SetAIClient to have been called.
// This is non-blocking; the response is streamed via goroutine.
func (a *ChatApp) SendUserMessage(text string) {
	a.mu.Lock()
	bridge := a.aiBridge
	a.mu.Unlock()

	if bridge == nil {
		return
	}

	// Add user message block immediately
	_ = a.AddUserMessage(text)

	// Send to AI
	bridge.SendUserMessage(text)
}

// StopStreaming cancels any in-flight AI streaming request.
func (a *ChatApp) StopStreaming() {
	a.mu.Lock()
	bridge := a.aiBridge
	a.mu.Unlock()
	if bridge != nil {
		bridge.StopStreaming()
	}
}

// IsStreaming reports whether an AI streaming request is in progress.
func (a *ChatApp) IsStreaming() bool {
	a.mu.Lock()
	bridge := a.aiBridge
	a.mu.Unlock()
	if bridge != nil {
		return bridge.IsStreaming()
	}
	return false
}
