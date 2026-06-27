# API Reference

## `fluui` (package: `github.com/topcheer/fluui`)

### App

The main entry point. Wires terminal, renderer, and event loop.

| Method | Description |
|---|---|
| `New() (*App, error)` | Create app, enter raw mode + alt screen |
| `Close() error` | Restore terminal |
| `Run() error` | Start event loop (blocks until Quit) |
| `Quit()` | Stop event loop |
| `Size() (int, int)` | Terminal width, height |
| `MarkDirty()` | Request immediate redraw |
| `Send(e Event)` | Inject custom event |

**Callbacks:**

| Method | Description |
|---|---|
| `OnKey(fn func(*KeyEvent))` | Keyboard handler |
| `OnMouse(fn func(*MouseEvent))` | Mouse handler |
| `OnResize(fn func(w, h int))` | Terminal resize handler |
| `OnPaint(fn func(*Buffer))` | Render callback (draw into back buffer) |
| `OnQuit(fn func())` | Called before terminal cleanup |
| `OnInterrupt(fn func() bool)` | Ctrl+C handler (true=quit, false=ignore) |

**Drawing helpers:**

| Method | Description |
|---|---|
| `DrawText(x, y int, text string, style Style) int` | Draw text at position |
| `DrawTextClamped(x, y int, text string, style Style) int` | Draw text (clamped to width) |
| `FillRect(rect Rect, cell Cell)` | Fill rectangular area |
| `Copy(text string)` | OSC52 clipboard copy |
| `CopySelection(text string, sel ClipboardSource)` | OSC52 to specific selection |
| `PasteFromClipboard()` | OSC52 paste query |

---

## `app` (package: `github.com/topcheer/fluui/app`)

### ChatApp

High-level AI chat interface.

**Construction:**

| Method | Description |
|---|---|
| `NewChatApp(width, height int) *ChatApp` | Create with dimensions |

**AI Integration:**

| Method | Description |
|---|---|
| `SetAIClient(client *ai.Client)` | Attach AI client for streaming |
| `SetSystemPrompt(prompt string)` | Set system prompt |
| `SetOnAIError(fn func(error))` | Error callback |
| `SendUserMessage(text string)` | Send message + stream response (non-blocking) |
| `StopStreaming()` | Cancel in-flight streaming |
| `IsStreaming() bool` | Check if streaming |

**Block Management:**

| Method | Description |
|---|---|
| `AddUserMessage(text string) *UserMessageBlock` | Add user message block |
| `AddThinking() *ThinkingBlock` | Add thinking block |
| `AddAssistantText() *AssistantTextBlock` | Add assistant text block |
| `AddToolCall(name, args string) *ToolCallBlock` | Add tool call block |
| `AddToolResult() *ToolResultBlock` | Add tool result block |
| `StreamDelta(delta StreamDelta)` | Dispatch streaming delta |
| `Container() *BlockContainer` | Direct container access |
| `Clear()` | Remove all blocks |

**Input Line:**

| Method | Description |
|---|---|
| `SetInputHeight(h int)` | Reserve bottom space for input |
| `SetInputLine(line *InputLine)` | Attach input component |
| `InputLine() *InputLine` | Get attached input line |
| `OnSubmit(fn func(string))` | Set Enter handler (auto-creates InputLine) |

**Scrolling:**

| Method | Description |
|---|---|
| `ScrollUp()` | Scroll up one line |
| `ScrollDown()` | Scroll down one line |
| `ScrollToBottom()` | Jump to latest content |
| `ScrollView() *ScrollView` | Direct scroll view access |

**Rendering:**

| Method | Description |
|---|---|
| `Render(buf *Buffer)` | Paint to buffer |
| `SetSize(w, h int)` | Update dimensions |
| `Size() (int, int)` | Current dimensions |
| `IsDirty() bool` | Check if redraw needed |
| `ClearDirty()` | Reset dirty flag |

**Interaction:**

| Method | Description |
|---|---|
| `HandleKey(key *KeyEvent) bool` | Process key event |
| `HandleMouse(mouse *MouseEvent) bool` | Process mouse event |
| `OnKey(fn func(*KeyEvent))` | Custom key handler |
| `OnMouse(fn func(*MouseEvent))` | Custom mouse handler |
| `OnClipboard(fn func(string))` | Clipboard paste handler |
| `OnQuit(fn func())` | Quit handler |

**Theme:**

| Method | Description |
|---|---|
| `SetTheme(t *theme.Theme)` | Set active theme |
| `CycleTheme() *theme.Theme` | Next built-in theme |
| `CycleThemeBack() *theme.Theme` | Previous theme |
| `Theme() *theme.Theme` | Current theme |
| `ThemeToast() (string, bool)` | Theme switch notification text |

**Overlay:**

| Method | Description |
|---|---|
| `Overlays() *OverlayManager` | Direct overlay access |

---

## `block` (package: `github.com/topcheer/fluui/block`)

### Block Types

| Constructor | Type | Key Methods |
|---|---|---|
| `NewThinkingBlock(id) *ThinkingBlock` | Thinking | AppendDelta, Toggle, Content, Collapsed |
| `NewAssistantTextBlock(id) *AssistantTextBlock` | Assistant | AppendDelta, Content |
| `NewToolCallBlock(id, name, args) *ToolCallBlock` | Tool Call | ToolName, RawArgs |
| `NewToolResultBlock(id) *ToolResultBlock` | Tool Result | AppendDelta, Output, Toggle, Collapsed |
| `NewUserMessageBlock(id, content) *UserMessageBlock` | User Msg | Content |
| `NewErrorBlock(id) *ErrorBlock` | Error | AppendDelta, Message, Timestamp |
| `NewErrorBlockWithMessage(id, msg) *ErrorBlock` | Error | Message, Timestamp |

### Block Lifecycle

All blocks implement:
- `ID() string` — unique identifier
- `Type() BlockType` — block type enum
- `State() BlockState` — `Streaming`, `Complete`, `Error`
- `Complete()` — mark as finished
- `Fail(err error)` — mark as errored
- `IsDirty() bool` / `ClearDirty()` — dirty tracking
- `Measure(Constraints) Size` — compute required space
- `Paint(*Buffer)` — render to buffer

### BlockContainer

| Method | Description |
|---|---|
| `NewBlockContainer() *BlockContainer` | Create empty container |
| `AddBlock(b Block)` | Append block |
| `Blocks() []Block` | Get all blocks (copy) |
| `Len() int` | Block count |
| `Clear()` | Remove all blocks |
| `SetSpacing(s int)` | Set gap between blocks |

### Serialization

| Method | Description |
|---|---|
| `SaveContainer(c, r) ([]byte, error)` | Serialize to JSON |
| `LoadContainer(data, r) (*BlockContainer, error)` | Deserialize from JSON |

### StreamDelta

```go
type StreamDelta struct {
    Type      DeltaType  // DeltaThinking, DeltaContent, DeltaToolCall, ...
    Content   string
    BlockID   string
}
```

---

## `component` (package: `github.com/topcheer/fluui/component`)

### Component Interface

```go
type Component interface {
    Measure(Constraints) Size
    SetBounds(Rect)
    Paint(*buffer.Buffer)
    Children() []Component
}
```

### Built-in Components

| Constructor | Description |
|---|---|
| `NewText(content string) *Text` | Styled text leaf |
| `NewBorder(child Component) *Border` | Box-drawing border around child |
| `NewScrollView(child Component) *ScrollView` | Scrollable viewport with scrollbar |

### ScrollView Methods

| Method | Description |
|---|---|
| `ScrollUp(n int)` / `ScrollDown(n int)` | Scroll by lines |
| `ScrollTo(offset int)` | Jump to position |
| `Offset() int` / `MaxOffset() int` | Current/max scroll position |
| `IsScrollbarVisible() bool` | Check scrollbar state |
| `HandleScrollbarDown(relY int)` | Begin drag or jump |
| `HandleScrollbarDrag(relY int)` | Drag scrollbar thumb |
| `HandleScrollbarUp()` | End drag |

### Layout (`component/layout/`)

| Constructor | Description |
|---|---|
| `NewFlex(dir Direction) *Flex` | Horizontal or vertical flex |
| `NewFlexGap(dir Direction, gap int) *Flex` | Flex with spacing |
| `NewCenter(child) *Center` | Center child in available space |
| `NewPadding(child, inset) *Padding` | Pad child with insets |
| `NewStack(children...) *Stack` | Overlay children (z-order) |

---

## `ai` (package: `github.com/topcheer/fluui/ai`)

### Config

| Method | Description |
|---|---|
| `LoadConfig(envPath ...string) (*Config, error)` | Load from .env or env vars |
| `MaskedKey(key string) string` | Show first/last 4 chars only |

**Environment variables:**
- `FLUUI_LLM_API_KEY` (required)
- `FLUUI_LLM_BASE_URL` (default: ZAI)
- `FLUUI_LLM_MODEL` (default: glm-5.2)
- `FLUUI_LLM_SYSTEM_PROMPT`

### Client

| Method | Description |
|---|---|
| `NewClient(cfg *Config) *Client` | Create streaming client |
| `ChatStream(ctx, messages) (<-chan StreamChunk, error)` | Start SSE stream |
| `ChatStreamWithSystem(ctx, system, messages) (<-chan StreamChunk, error)` | With system prompt |

---

## `theme` (package: `github.com/topcheer/fluui/theme`)

### Functions

| Method | Description |
|---|---|
| `Get() *Theme` | Current active theme |
| `SetActive(t *Theme)` | Set active theme |
| `Default() *Theme` | Dracula theme |
| `Cycle() *Theme` | Next built-in theme |
| `CycleBack() *Theme` | Previous theme |
| `Builtin() []*Theme` | All 5 built-in themes |

### Built-in Themes

- `Dracula()` — Dark purple/pink
- `Nord()` — Arctic blue
- `Gruvbox()` — Warm earth tones
- `SolarizedDark()` — Precision colors
- `TokyoNight()` — Neon city night

### Markdown Theme

| Method | Description |
|---|---|
| `markdown.DefaultTheme() *MarkdownTheme` | Dracula-based markdown colors |
| `markdown.MarkdownThemeFromTheme(t *Theme) *MarkdownTheme` | Convert global theme |

---

## `overlay` (package: `github.com/topcheer/fluui/overlay`)

| Constructor | Description |
|---|---|
| `NewOverlayManager() *OverlayManager` | Create manager |
| `NewModal(content Component, w, h int) *Modal` | Centered dialog |
| `NewPopup(content Component) *Popup` | Fullscreen overlay |

### OverlayManager Methods

| Method | Description |
|---|---|
| `Add(o Overlay)` | Add overlay (higher z-index) |
| `Remove(id string)` | Remove overlay |
| `Top() Overlay` | Topmost overlay |
| `Paint(buf *Buffer)` | Paint all overlays bottom-to-top |
