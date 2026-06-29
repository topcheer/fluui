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

**Input Line Undo/Redo (Phase 19):**

| Method | Description |
|---|---|
| `Undo() bool` | Undo last edit (Ctrl+Z) |
| `Redo() bool` | Redo last undo (Ctrl+Shift+Z, Ctrl+Y) |
| `CanUndo() bool` | Check if undo is available |
| `CanRedo() bool` | Check if redo is available |
| `UndoCount() int` | Number of undo states |
| `RedoCount() int` | Number of redo states |
| `ClearUndoHistory()` | Clear all undo/redo state |

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

**CommandPalette (Phase 20):**

| Method | Description |
|---|---|
| `ToggleCommandPalette() bool` | Show/hide command palette (Ctrl+P) |
| `CommandPalette() *component.CommandPalette` | Direct palette access |
| `SetCommandPalette(cp *component.CommandPalette)` | Set custom palette |
| `IsCommandPaletteVisible() bool` | Check if palette is shown |

**Spinner (Phase 20):**

| Method | Description |
|---|---|
| `StartSpinner(label string)` | Show loading spinner with label |
| `StopSpinner()` | Hide loading spinner |
| `Spinner() *component.Spinner` | Direct spinner access |
| `SetSpinner(s *component.Spinner)` | Set custom spinner |
| `IsSpinnerActive() bool` | Check if spinner is running |

**Theme:**

| Method | Description |
|---|---|
| `SetTheme(t *theme.Theme)` | Set active theme |
| `CycleTheme() *theme.Theme` | Next built-in theme (Ctrl+T, Ctrl+]) |
| `CycleThemeBack() *theme.Theme` | Previous theme (Ctrl+Shift+T, Ctrl+\) |
| `Theme() *theme.Theme` | Current theme |
| `ThemeToast() (string, bool)` | Theme switch notification text |
| `ThemeCount() int` | Number of builtin themes |
| `ThemeList() []string` | All builtin theme names |
| `ThemeIndex() int` | Current theme index |
| `ThemeName() string` | Current theme name |
| `SetThemeByIndex(idx int)` | Switch to theme at index |
| `SetThemeByName(name string) bool` | Switch by name (returns success) |

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

---

## `component` (package: `github.com/topcheer/fluui/component`)

### Phase 15 Components

#### FilePicker

File browser with fuzzy filtering, multi-select, and vim-style keybindings.

| Method | Description |
|---|---|
| `NewFilePicker(dir string) *FilePicker` | Create from directory |
| `MoveDown()` / `MoveUp()` | Navigate cursor |
| `EnterDir()` / `GoUp()` | Directory navigation |
| `Cwd() string` | Current working directory |
| `SetFilter(s string)` / `Filter() string` | Set/get filter pattern |
| `AppendFilter(r rune)` / `BackspaceFilter()` | Incremental filtering |
| `FilteredCount() int` | Number of filtered entries |
| `ToggleSelect()` / `IsSelected(path string) bool` | Multi-select |
| `SelectedFiles() []FileEntry` | Get all selected files |
| `ClearSelection()` | Clear all selections |
| `Entries() []FileEntry` | All visible entries |
| `Cursor() int` / `CurrentEntry() (FileEntry, bool)` | Cursor state |
| `SetOnSelect(fn)` / `SetOnConfirm(fn)` / `SetOnError(fn)` | Callbacks |
| `SetStyle(FilePickerStyle)` / `Style() FilePickerStyle` | Styling |
| `HandleKey(*term.KeyEvent) bool` | Keyboard input |
| `Measure(Constraints) Size` | Measure size |
| `SetBounds(Rect)` / `Bounds() Rect` | Set/get bounds |
| `Paint(*buffer.Buffer)` | Render to buffer |
| `Children() []Component` | Child components |

#### StatusBar

Status bar with left/center/right item alignment.

| Method | Description |
|---|---|
| `NewStatusBar() *StatusBar` | Create status bar |
| `AddItem(item StatusItem)` | Add item with alignment |
| `AddLeft(item)` / `AddCenter(item)` / `AddRight(item)` | Add to position |
| `RemoveItem(id string)` | Remove item by ID |
| `Clear()` | Remove all items |
| `Items() []StatusItem` | Get all items (copy) |
| `ItemCount() int` | Number of items |
| `SetItemText(id, text string)` | Update item text |
| `SetItemStyle(id string, style buffer.Style)` | Update item style |
| `SetSeparator(sep string)` / `Separator() string` | Separator between items |
| `SetModel(model string)` | Set AI model display |
| `SetTokenRate(rate int)` | Set token rate ("1.5k tok/s") |
| `SetContextWindow(used, total int)` | Set context window display |
| `SetClock(time string)` | Set clock display |
| `SetHeight(h int)` | Set bar height (min 1) |
| `SetStyle(StatusBarStyle)` / `Style() StatusBarStyle` | Styling |
| `Measure(Constraints) Size` | Measure size |
| `SetBounds(Rect)` / `Bounds() Rect` | Set/get bounds |
| `Paint(*buffer.Buffer)` | Render to buffer |

#### TabBar

Tab management with close buttons and keyboard navigation.

| Method | Description |
|---|---|
| `NewTabBar() *TabBar` | Create tab bar |
| `AddTab(id, title string)` | Add a tab |
| `AddClosableTab(id, title string)` | Add a closable tab |
| `RemoveTab(id string)` | Remove tab by ID |
| `Tabs() []Tab` | Get all tabs |
| `TabCount() int` | Number of tabs |
| `SetActive(index int)` / `Active() int` | Set/get active tab |
| `NextTab()` / `PrevTab()` | Cycle tabs |
| `SetStyle(TabBarStyle)` / `Style() TabBarStyle` | Styling |
| `HandleKey(*term.KeyEvent) bool` | Keyboard input |
| `HitTest(mx, my int) (int, bool)` | Mouse hit test |
| `ClickAt(mx, my int) bool` | Click handling |
| `IsCloseButton(mx, my int) (int, bool)` | Check close button hit |
| `Measure(Constraints) Size` | Measure size |
| `SetBounds(Rect)` / `Bounds() Rect` | Set/get bounds |
| `Paint(*buffer.Buffer)` | Render to buffer |

#### DiffPreview

Scrollable diff viewer with syntax highlighting.

| Method | Description |
|---|---|
| `NewDiffPreview() *DiffPreview` | Create diff viewer |
| `SetDiff(text string)` | Set diff content |
| `Lines() []DiffLine` | Get classified lines |
| `LineCount() int` | Number of lines |
| `Stats() DiffStats` | Diff statistics |
| `IsEmpty() bool` / `HasChanges() bool` | Content state |
| `DiffSummary() string` | Human-readable summary |
| `ScrollDown(n int)` / `ScrollUp(n int)` | Scroll by lines |
| `ScrollTo(row int)` / `ScrollY() int` | Set/get scroll position |
| `ScrollPageDown(vh int)` / `ScrollPageUp(vh int)` | Page scroll |
| `VisibleRange() (int, int)` | Visible line range |
| `SetTitle(title string)` / `Title() string` | Set/get title |
| `SetStyle(DiffPreviewStyle)` | Styling |
| `SetShowLineNumbers(bool)` / `ShowLineNumbers() bool` | Line number display |
| `Measure(Constraints) Size` | Measure size |
| `SetBounds(Rect)` / `Bounds() Rect` | Set/get bounds |
| `Paint(*buffer.Buffer)` | Render to buffer |
| `Children() []Component` | Child components |

#### LinkManager

URL detection, OSC8 hyperlinks, and click handling.

| Method | Description |
|---|---|
| `NewLinkManager() *LinkManager` | Create link manager |
| `DetectLinks(text string, row int) []LinkRange` | Detect URLs in text |
| `AddLink(link LinkRange)` | Add a link manually |
| `Links() []LinkRange` | Get all links |
| `LinkAt(x, y int) (LinkRange, bool)` | Find link at position |
| `ClickLink(x, y int) bool` | Click link at position |
| `AnnotateBuffer(buf *buffer.Buffer)` | Apply OSC8 to buffer |
| `ScanText(text string, row int) []LinkRange` | Scan for URLs |
| `SetStyle(LinkStyle)` | Styling |
| `SetOnClick(fn func(string))` | Set click callback |

### Constants and Types

```go
// DiffType
type DiffType uint8
const (
    DiffContext DiffType = iota
    DiffAdd
    DiffDel
    DiffHunk
    DiffFile
    DiffMeta
)

// StatusItemAlignment
type StatusItemAlignment int
const (
    StatusAlignLeft   StatusItemAlignment = 0
    StatusAlignCenter StatusItemAlignment = 1
    StatusAlignRight  StatusItemAlignment = 2
)

// FileEntry
type FileEntry struct {
    Name    string
    Path    string
    IsDir   bool
    Size    int64
    Mode    os.FileMode
    ModTime int64
}

// DiffStats
type DiffStats struct {
    Additions  int
    Deletions  int
    Files      int
    Hunks      int
    TotalLines int
}

// Tab
type Tab struct {
    ID       string
    Title    string
    Closable bool
    Active   bool
}
```

---

## `app/selection` (package: `github.com/topcheer/fluui/app`)

### SelectionManager

Text selection with mouse and keyboard support, plus OSC52 clipboard copy.

| Method | Description |
|---|---|
| `NewSelectionManager() *SelectionManager` | Create manager |
| `StartSelection(x, y int)` | Begin mouse selection |
| `ExtendSelection(x, y int)` | Extend to position |
| `EndSelection()` | End selection |
| `StartKeyboardSelection(x, y int)` | Begin keyboard selection |
| `ExtendKeyboardSelection(dx, dy, bw, bh int)` | Extend by delta |
| `HasSelection() bool` | Check if selection active |
| `Selection() Selection` | Get normalized selection |
| `Cursor() SelectionPoint` | Current cursor position |
| `Anchor() SelectionPoint` | Selection anchor point |
| `Clear()` | Clear selection |
| `SelectedText(lines []string) string` | Extract selected text |

---

## ChatApp P16 Integration (package: `github.com/topcheer/fluui/app`)

The ChatApp integrates P15 StatusBar, TabBar, and SelectionManager components with a unified mouse/key handler.

### StatusBar Integration

| Method | Description |
|---|---|
| `SetStatusBar(sb *component.StatusBar)` | Attach/detach status bar (nil to detach) |
| `StatusBar() *component.StatusBar` | Get attached status bar |
| `SetModel(name string)` | Set AI model name in status bar |
| `SetTokenRate(rate int)` | Set token generation rate (auto-formats k notation) |
| `SetContextWindow(used, total int)` | Set context window usage display |
| `UpdateClock()` | Refresh clock to current time |

### TabBar Integration (Multi-Session)

| Method | Description |
|---|---|
| `SetTabBar(tb *component.TabBar)` | Attach/detach tab bar |
| `TabBar() *component.TabBar` | Get attached tab bar |
| `AddSession(name string) int` | Create new session tab (returns index, -1 if no tab bar) |
| `SwitchSession(idx int)` | Switch to session at index |
| `NextSession() / PrevSession()` | Switch to next/prev session (wraps around) |
| `CloseSession()` | Close active session tab |
| `SessionCount() int` | Number of open sessions |
| `ActiveSession() int` | Active session index |
| `ActiveSessionName() string` | Active session title |
| `Sessions() []SessionInfo` | All session metadata (ID, Name, Index) |

### SelectionManager Integration

| Method | Description |
|---|---|
| `SetSelectionManager(sm *SelectionManager)` | Attach/detach selection manager |
| `SelectionManager() *SelectionManager` | Get attached selection manager |
| `HasSelection() bool` | Check if text selection is active |
| `ClearSelection()` | Clear current selection |

### Enhanced Event Handling

| Method | Description |
|---|---|
| `HandleMouseP16(mouse *term.MouseEvent) bool` | Unified mouse routing: overlays -> tabs -> selection -> scroll -> custom |
| `handleP16Keys(key *term.KeyEvent) bool` | Alt+[/] next/prev session, Alt+W close, Alt+1-9 switch |
| `renderP16(buf *buffer.Buffer, w, h int)` | Paint tab bar (top) + status bar (bottom) + selection highlights |

### Keyboard Shortcuts

| Key | Action |
|---|---|
| `Alt+]` | Next session |
| `Alt+[` | Previous session |
| `Alt+W` | Close active session |
| `Alt+1` to `Alt+9` | Switch to session N |

## Phase 17: VirtualScroller + Pagination

### VirtualScroller

| Method | Description |
|---|---|
| `NewVirtualScroller()` | Create with defaults |
| `SetItems(items []VirtualItem)` | Set data source |
| `SetCursor(n int)` | Move cursor to item |
| `Cursor() int` | Current cursor position |
| `ScrollTo(n int)` | Scroll viewport to position |
| `VisibleRange() (start, end int)` | Get visible window |
| `Filter(q string) []VirtualItem` | Filter items by text |
| `SetHeader(text string)` | Set header label |
| `SetBounds(Rect) / Paint(*buffer.Buffer)` | Component interface |

### Pagination

| Method | Description |
|---|---|
| `NewPagination()` | Create with defaults |
| `SetTotalItems(n int)` | Total item count |
| `SetItemsPerPage(n int)` | Items per page |
| `SetPage(n int)` | Jump to page |
| `CurrentPage() int` | Active page (0-based) |
| `TotalPages() int` | Total page count |
| `PageStartIndex() / PageEndIndex() int` | Item range on current page |

## Phase 18: Dialog + AutoComplete + Wizard

### Dialog

| Method | Description |
|---|---|
| `NewDialog(dt DialogType, title, msg string)` | Create dialog |
| `NewConfirmDialog(title, msg string)` | Yes/No dialog |
| `NewInfoDialog(title, msg string)` | Single OK button |
| `NewPromptDialog(title, msg, default string)` | Text input + OK/Cancel |
| `SetTitle(string) / SetMessage(string)` | Update content |
| `AddButton(DialogButton) / SetButtons([]DialogButton)` | Manage buttons |
| `Cursor() / SetCursor(int)` | Button cursor |
| `MoveLeft() / MoveRight()` | Navigate buttons (wraps) |
| `InputValue() / SetInputValue(string)` | Prompt input text |
| `InsertRune(rune) / Backspace() / Delete()` | Edit input |
| `Confirm() bool` | Confirm (false if OnConfirm rejects) |
| `Cancel() / PressButton()` | Other actions |
| `Show() / Hide() / Visible() bool` | Visibility |
| `Result() DialogResult` | OK/Cancel/Custom |
| `HandleKey(*term.KeyEvent) bool` | Esc/Enter/Tab/Left/Right/printable |

### AutoComplete

| Method | Description |
|---|---|
| `NewAutoComplete()` | Create autocomplete |
| `SetItems([]CompletionItem)` | Set candidate items |
| `AddItem(item CompletionItem)` | Add single item |
| `Items() / ItemCount() / Clear()` | Item queries |
| `SetQuery(string) / Query()` | Set/get filter text |
| `FilteredCount() / HasResults() / FilteredItems()` | Filter results |
| `Cursor() / SetCursor(int)` | Result cursor |
| `MoveUp() / MoveDown() / CurrentItem()` | Navigate results |
| `Show(x, y int) / Hide() / Visible()` | Popup control |
| `Select()` | Select current item (fires OnSelect) |
| `SetOnSelect(func(CompletionItem)) / SetOnDismiss(func())` | Callbacks |
| `SetMaxVisible(int) / SetCaseSensitive(bool)` | Configuration |
| `HandleKey(*term.KeyEvent) bool` | Up/Down/Tab/Enter/Esc |

### Wizard

| Method | Description |
|---|---|
| `NewWizard(steps []*WizardStep)` | Create wizard |
| `NewWizardStep(id, title string)` | Create step (chainable) |
| `SetDescription / SetContent / SetSkippable` | Step configuration |
| `SetOnEnter(func(*Wizard) error) / SetOnLeave(func(*Wizard) error)` | Lifecycle hooks |
| `Next() error / Back() error` | Navigate (error blocks) |
| `SetCurrentStep(idx int) error` | Jump to step |
| `Reset()` | Return to step 0 |
| `Finish() / Cancel()` | Complete/cancel wizard |
| `StepCount() / CurrentStepIndex()` | Step queries |
| `CurrentStep() / Steps() []*WizardStep` | Get steps |
| `IsFirstStep() / IsLastStep() / IsCompleted() / IsCancelled()` | State queries |
| `SelectedButton() / SetSelectedButton(WizardButton)` | Button focus |
| `ButtonOrder() []WizardButton` | Dynamic buttons |
| `SetOnFinish(func(*Wizard)) / SetOnCancel(func(*Wizard))` | Completion callbacks |
| `SetOnStepChange(func(*Wizard, int))` | Step change callback |
| `HandleKey(*term.KeyEvent) bool` | Tab/Left/Right/Enter/Esc/Ctrl+N/Ctrl+B |

### Checkbox

| Method | Description |
|---|---|
| `NewCheckbox()` | Create checkbox list |
| `AddItem(label string)` | Add a checkbox item |
| `SetItems([]CheckboxItem)` | Set all items |
| `Items() []CheckboxItem` | Get all items (defensive copy) |
| `ItemCount() int` | Number of items |
| `CheckedItems() []CheckboxItem` | Get checked items only |
| `CheckedLabels() []string` | Get checked item labels |
| `Toggle()` | Toggle current cursor item |
| `SetChecked(idx int, checked bool)` | Set specific item |
| `IsChecked(idx int) bool` | Check if item is checked |
| `CheckAll() / UncheckAll()` | Bulk operations |
| `Cursor() / SetCursor(int)` | Cursor position |
| `MoveUp() / MoveDown()` | Navigate (wrap-around, skip disabled) |
| `SetStyle(CheckboxStyle) / Style()` | Style configuration |
| `OnChange` | Callback: `func(idx int, checked bool)` |
| `HandleKey(*term.KeyEvent) bool` | Space/Enter=toggle, j/k=nav, Ctrl+A/D |

### RadioGroup

| Method | Description |
|---|---|
| `NewRadioGroup()` | Create radio group |
| `AddItem(label, value string)` | Add option |
| `Labels() []string` | All labels |
| `SelectedIndex() int` | Currently selected (-1 if none) |
| `SelectedLabel() / SelectedValue()` | Selected item info |
| `SetSelected(idx int)` | Select an item |
| `Select()` | Select current cursor item |
| `IsDisabled(idx int) / SetDisabled(idx int, bool)` | Disable management |
| `Cursor() / SetCursor(int)` | Cursor position |
| `MoveUp() / MoveDown()` | Navigate (wrap-around, skip disabled) |
| `SetStyle(RadioGroupStyle) / Style()` | Style configuration |
| `OnChange` | Callback: `func(value string)` |
| `HandleKey(*term.KeyEvent) bool` | Up/Down/j/k=nav, Enter/Space=select |

### Slider

| Method | Description |
|---|---|
| `NewSlider()` | Create slider (default: min=0, max=100, step=1) |
| `Value() / SetValue(float64)` | Current value |
| `Min() / Max() / SetRange(min, max float64)` | Range |
| `Step() / SetStep(float64)` | Step size |
| `Orientation() / SetOrientation(SliderOrientation)` | Horizontal/Vertical |
| `Ratio() float64` | Current ratio (0.0-1.0) |
| `SetFromRatio(ratio float64)` | Set by percentage |
| `Increment() / Decrement()` | Step by one |
| `IncrementBy(n int)` | Step by n |
| `SetLabel(string) / Label()` | Display label |
| `SetShowValue(bool) / ShowValue()` | Show/hide value display |
| `SetStyle(SliderStyle) / Style()` | Style configuration |
| `SetOnChange(func(float64))` | Value change callback |
| `HandleKey(*term.KeyEvent) bool` | Arrows/h/l=step, Home/End=min/max |

### CommandPalette

| Method | Description |
|---|---|
| `NewCommandPalette()` | Create palette |
| `AddCommand(Command) / SetCommands([]Command)` | Add/set commands |
| `Commands() []Command` | All commands (defensive copy) |
| `CommandCount() int` | Total commands |
| `SetQuery(string) / Query()` | Search text |
| `FilteredCount() / HasResults() / FilteredCommands()` | Filtered results |
| `Cursor() / SetCursor(int)` | Selection cursor (wraps) |
| `MoveUp() / MoveDown()` | Navigate filtered results |
| `CurrentCommand() *Command` | Command at cursor |
| `ScrollY() int` | Current scroll offset |
| `SetMaxVisible(int) / MaxVisible() int` | Scroll window size |
| `Show(x, y int) / Hide() / Visible()` | Visibility control |
| `Reset()` | Clear query, reset cursor |
| `SetStyle(CommandPaletteStyle)` | Normal/Matched/Selected/Prompt styles |
| `OnExecute / OnDismiss` | Callbacks |
| `HandleKey(*term.KeyEvent) bool` | Print/BS=filter, Up/Down/Tab=nav, Enter/Esc |

### Spinner

| Method | Description |
|---|---|
| `NewSpinner()` | Create spinner |
| `SetLabel(string) / Label()` | Display label |
| `SetPrefix(string) / Prefix()` | Text before frame |
| `SetStyle(SpinnerStyle) / Style()` | Frame/Label/Prefix style |
| `SetFrameStyle(string) / FrameStyle()` | Animation style (dots/arc/line/bounce/bars) |
| `Start() / Stop() / Running()` | Animation control |
| `Update(delta time.Duration) bool` | Advance frame, returns if changed |
| `CurrentFrame() string` | Current frame string |
| `SetFrameIndex(int) / FrameIndex()` | Direct frame control (wraps) |

---

## Phase 12-25 Component API Reference

### Table (Phase 12)

| Method | Description |
|---|---|
| `NewTable(headers []string, rows ...[]string)` | Create table with headers and data |
| `SetRows([][]string) / Rows() [][]string` | Set/get table data |
| `SetHeaders([]string) / Headers() []string` | Set/get column headers |
| `SortBy(col int, ascending bool)` | Sort by column |
| `SetAlignment(col int, align Alignment)` | Set column alignment (AlignLeft/Center/Right) |
| `SetZebraStriping(bool)` | Enable/disable alternating row colors |
| `MoveUp() / MoveDown() / SetCursor(int) / Cursor() int` | Row navigation |
| `CurrentRow() []string` | Get selected row data |
| `OnSelect(func([]string))` | Selection callback |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: j/k, Up/Down, Ctrl+1-9 sort |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Tree (Phase 12)

| Method | Description |
|---|---|
| `NewTreeNode(id, label string)` | Create tree node |
| `node.AddChild(child) / Children() []*TreeNode` | Child management |
| `node.SetExpanded(bool) / IsExpanded() bool` | Expand/collapse |
| `node.SetIcon(string) / Icon() string` | Custom icon |
| `SetRoot(node *TreeNode)` | Set root node |
| `ExpandAll() / CollapseAll()` | Bulk expand/collapse |
| `MoveUp() / MoveDown() / SetCursor(int)` | Node navigation |
| `CurrentNode() *TreeNode` | Get selected node |
| `OnSelect(func(*TreeNode))` | Selection callback |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: j/k, Enter toggle, arrows |

### Form (Phase 12)

| Method | Description |
|---|---|
| `NewForm() *Form` | Create empty form |
| `AddField(field FormField)` | Add form field |
| `Fields() []FormField` | Get all fields |
| `Validate() error` | Validate all fields |
| `Submit() bool` | Submit form (returns true if valid) |
| `OnSubmit(func()) / OnCancel(func())` | Callbacks |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Tab next, Shift+Tab prev, Enter submit, Esc cancel |
| **TextField**: `NewTextField(id, label string)` | Text input field |
| **CheckboxField**: `NewCheckboxField(id, label string)` | Checkbox field |
| **SelectField**: `NewSelectField(id, label string, options []string)` | Dropdown field |

### ProgressBar (Phase 12)

| Method | Description |
|---|---|
| `NewProgressBar()` | Create progress bar |
| `SetProgress(float64) / Progress() float64` | Set/get progress (0-100) |
| `SetIndeterminate(bool)` | Toggle indeterminate animation |
| `SetLabel(string) / Label() string` | Display label |
| `SetWidth(int)` | Bar width in characters |

### StatusIndicator (Phase 12)

| Method | Description |
|---|---|
| `NewStatusIndicator()` | Create status indicator |
| `SetStatus(status string)` | Set status text |
| `SetSpinnerStyle(style int)` | Spinner style: 0=Dots, 1=Arc, 2=Line, 3=Bounce, 4=Bars |

### Gauge (Phase 13)

| Method | Description |
|---|---|
| `NewGauge() *Gauge` | Create gauge |
| `SetValue(float64) / Value() float64` | Set/get value (0-1 or custom range) |
| `SetRange(min, max float64)` | Set value range |
| `SetOrientation(GaugeOrientation)` | Horizontal, Vertical, Radial |
| `SetThresholds([]GaugeThreshold)` | Color threshold bands |
| `SetShowValue(bool)` | Display numeric value |

### Sparkline (Phase 13)

| Method | Description |
|---|---|
| `NewSparkline() *Sparkline` | Create sparkline |
| `Push(value float64)` | Add data point |
| `SetData([]float64) / Data() []float64` | Set/get data series |
| `SetColorMode(SparkColorMode)` | Solid, Gradient, Value-based |
| `SetMaxPoints(int)` | Maximum visible points |

### Badge (Phase 13)

| Method | Description |
|---|---|
| `NewBadge(text string) *Badge` | Create badge |
| `SetText(string) / Text() string` | Set/get text |
| `SetVariant(BadgeVariant)` | Primary, Success, Warning, Danger, Info, Neutral |
| `SetSize(BadgeSize)` | Small, Medium, Large |
| `NewBadgeGroup() *BadgeGroup` | Horizontal badge container |
| `group.Add(badge) / Remove(id string)` | BadgeGroup management |

### Notification (Phase 13)

| Method | Description |
|---|---|
| `NewToastManager() *ToastManager` | Create toast notification manager |
| `Push(level NotificationLevel, title, msg string)` | Show notification |
| `Dismiss(id string) / Clear()` | Remove notifications |
| `Tick()` | Update auto-expiry timers |
| `Notifications() []Notification` | Get active notifications |
| **Levels**: `NotifyInfo, NotifySuccess, NotifyWarning, NotifyError` | Notification severity |

### ContextMenu (Phase 14)

| Method | Description |
|---|---|
| `NewContextMenu() *ContextMenu` | Create context menu |
| `AddItem(item *MenuItem) / AddLabel(id, label string) / AddSeparator()` | Add items |
| `Remove(id string) / Clear() / Items() []*MenuItem` | Item management |
| `Show(x, y int) / Hide() / Visible() bool` | Visibility control |
| `MoveUp() / MoveDown() / SetCursor(int) / Cursor() int` | Navigation |
| `Activate()` | Fire current item's action / open submenu |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Up/Down/Enter/Esc/Right(submenu)/Left(close) |
| `HitTest(mx, my int) bool / ClickAt(mx, my int)` | Mouse interaction |
| **MenuItem**: `NewMenuItem(id, label string)` | Menu item builder |
| `item.SetShortcut(s) / SetIcon(s) / SetEnabled(bool) / SetSubmenu(cm) / SetAction(fn)` | Item config |

### Tooltip (Phase 14)

| Method | Description |
|---|---|
| `NewTooltip() *Tooltip` | Create tooltip |
| `SetText(string) / Text() string` | Set/get tooltip text |
| `Show(x, y int) / Hide() / Visible() bool` | Visibility |
| `SetDelay(duration)` | Hover delay before showing |
| `SetMaxWidth(int)` | Wrap width |

### SplitPane (Phase 14)

| Method | Description |
|---|---|
| `NewSplitPane(left, right Component) *SplitPane` | Create split pane |
| `SetOrientation(SplitOrientation)` | Horizontal, Vertical |
| `SetRatio(float64) / Ratio() float64` | Split position (0-1) |
| `SetResizable(bool)` | Enable drag-to-resize |
| `SetMinSize(int)` | Minimum pane size |
| `HandleKey(*term.KeyEvent) bool` | Keyboard resize (Ctrl+arrow) |
| `HandleMouse(*term.MouseEvent) bool` | Mouse drag resize |

### HelpOverlay (Phase 14)

| Method | Description |
|---|---|
| `NewHelpOverlay(groups []HelpGroup) *HelpOverlay` | Create help overlay |
| `SetGroups([]HelpGroup) / Groups() []HelpGroup` | Set/get help groups |
| `SetQuery(string) / Query() string` | Search filter |
| `AppendQuery(rune) / BackspaceQuery() / ClearQuery()` | Search input |
| `FilteredGroups() []HelpGroup / TotalRows() int` | Filtered results |
| `SelectNext() / SelectPrev() / SetSelected(int)` | Selection nav |
| `ScrollUp() / ScrollDown() / ScrollY() int` | Scroll control |
| `Show() / Hide() / Visible() bool` | Visibility |

### HotkeyManager (Phase 14)

| Method | Description |
|---|---|
| `NewManager() *Manager` | Create hotkey manager |
| `Register(action, sequence string, opts ...Option) error` | Register binding |
| `Unregister(action string)` | Remove binding |
| `Enable(action) / Disable(action)` | Toggle bindings |
| `Match(*term.KeyEvent) (action string, result MatchResult)` | Match key event |
| `HasPending() bool / ResetPending()` | Multi-key sequence state |
| `SetSequenceTimeout(duration)` | Partial match timeout |
| `Groups() / BindingsByGroup(name) / BindingsByScope(s)` | Query bindings |
| `ExportConfig() / ImportConfig(data)` | Serialize/restore |
| `HasConflict(seq KeySequence) bool` | Conflict detection |
| `ParseCombo("Ctrl+F") / ParseSequence("g g")` | Parse key strings |

### FilePicker (Phase 15)

| Method | Description |
|---|---|
| `NewFilePicker(dir string) *FilePicker` | Create file browser |
| `SetDirReader(fn DirReader)` | Custom directory reader (for testing) |
| `EnterDir() / GoUp() / Cwd() string` | Directory navigation |
| `SetFilter(string) / Filter() string` | Name filter |
| `MoveDown() / MoveUp() / Cursor() int` | File navigation |
| `ToggleSelect() / IsSelected(path) bool / SelectedFiles() []string` | Multi-select |
| `ClearSelection()` | Clear all selections |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: j/k/h/l/g/G, arrows, Enter/Space/Esc |

### TabBar (Phase 15)

| Method | Description |
|---|---|
| `NewTabBar() *TabBar` | Create tab bar |
| `AddTab(title string) int / RemoveTab(idx)` | Tab management |
| `SetTabs([]TabInfo)` | Bulk set tabs |
| `NextTab() / PrevTab() / SwitchTab(idx)` | Navigation |
| `ActiveIndex() int / SetActiveIndex(idx)` | Get/set active tab |
| `SetClosable(bool) / Closable() bool` | Close button toggle |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Left/Right, Ctrl+W close |

### StatusBar (Phase 15)

| Method | Description |
|---|---|
| `NewStatusBar() *StatusBar` | Create status bar |
| `AddLeft(id, text string) / AddCenter(id, text string) / AddRight(id, text string)` | Add items |
| `RemoveItem(id string) / Clear()` | Remove items |
| `SetItemText(id, text string) / SetItemStyle(id, style)` | Update items |
| `SetSeparator(string)` | Item separator (default " │ ") |
| `SetModel(string) / SetTokenRate(int) / SetContextWindow(used, total int)` | AI agent convenience |
| `SetHeight(int)` | Status bar height (min 1) |

### DiffPreview (Phase 15)

| Method | Description |
|---|---|
| `NewDiffPreview() *DiffPreview` | Create diff viewer |
| `SetDiff(text string) / Lines() []DiffLine` | Set/get diff |
| `Stats() DiffStats` | Get {Additions, Deletions, Files, Hunks, TotalLines} |
| `ParseDiff(text string) []DiffLine` | Parse diff text |
| `ScrollDown(n) / ScrollUp(n) / ScrollTo(row) / ScrollY() int` | Scrolling |
| `ScrollPageDown(vh) / ScrollPageUp(vh)` | Page scroll |
| `VisibleRange(viewportH) (start, end int)` | Visible line range |

### Link (Phase 15)

| Method | Description |
|---|---|
| `DetectLinks(text string, row int) []LinkRange` | Find URLs in text |
| `NewLinkManager() *LinkManager` | Create link manager |
| `mgr.AddLink(link) / Links() []LinkRange` | Link management |
| `mgr.LinkAt(x, y int) (LinkRange, bool)` | Hit test for click |
| `mgr.ClickLink(link)` | Open URL |
| `mgr.AnnotateBuffer(buf, style)` | Apply OSC8 annotations to buffer |
| `mgr.ScanText(text, row)` | Scan and detect all links |

### Dialog (Phase 18)

| Method | Description |
|---|---|
| `NewDialog(dt DialogType, title, msg string) *Dialog` | Create dialog |
| `NewConfirmDialog(title, msg) / NewInfoDialog(title, msg)` | Convenience constructors |
| `NewPromptDialog(title, msg, default string)` | Prompt dialog |
| `SetTitle(string) / SetMessage(string)` | Update content |
| `AddButton(DialogButton) / SetButtons([]DialogButton)` | Custom buttons |
| `MoveLeft() / MoveRight() / Cursor() int` | Button navigation |
| `InputValue() / SetInputValue(string) / InputCursor() int` | Prompt input |
| `InsertRune(r) / Backspace() / Delete()` | Input editing |
| `Confirm() bool / Cancel() / PressButton()` | Dialog actions |
| `OnConfirm func(string) bool` | Validation callback (false = keep open) |
| `OnCancel func() / OnClose func(DialogResult, string)` | Dismiss callbacks |
| `Show() / Hide() / Visible() bool / Result() DialogResult` | Visibility |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Esc/Enter/Tab/Left/Right + printable for prompt |

### AutoComplete (Phase 18)

| Method | Description |
|---|---|
| `NewAutoComplete() *AutoComplete` | Create autocomplete |
| `SetItems([]CompletionItem) / AddItem(item)` | Set candidates |
| `Items() / ItemCount() / Clear()` | Query/manage items |
| `SetQuery(string) / Query() string` | Filter text |
| `FilteredCount() int / HasResults() bool / FilteredItems() []CompletionItem` | Filter results |
| `MoveUp() / MoveDown() / Cursor() int / SetCursor(int)` | Navigation (wrap) |
| `CurrentItem() CompletionItem` | Selected item |
| `Show(x, y int) / Hide() / Visible() bool / Position() (int, int)` | Popup control |
| `Select() / SetOnSelect(fn) / SetOnDismiss(fn)` | Selection callbacks |
| `SetMaxVisible(int) / SetCaseSensitive(bool)` | Configuration |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Up/Down/Tab/Enter/Esc |

### Wizard (Phase 18)

| Method | Description |
|---|---|
| `NewWizard(steps []*WizardStep) *Wizard` | Create wizard |
| `NewWizardStep(id, title string) *WizardStep` | Create step |
| `step.SetDescription(s) / SetContent(c) / SetSkippable(bool)` | Step config |
| `step.SetOnEnter(fn) / SetOnLeave(fn)` | Lifecycle hooks (OnLeave can block) |
| `Next() / Back() / SetCurrentStep(idx) / Reset()` | Navigation |
| `Finish() / Cancel()` | End wizard |
| `CurrentStepIndex() / StepCount() / CurrentStep()` | State queries |
| `IsFirstStep() / IsLastStep() / IsCompleted() / IsCancelled()` | State checks |
| `SelectedButton() / SetSelectedButton(btn)` | Button focus |
| `SetOnFinish(fn) / SetOnCancel(fn) / SetOnStepChange(fn)` | Callbacks |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Tab/Left/Right cycle, Enter, Esc, Ctrl+N/B |

### Checkbox (Phase 19)

| Method | Description |
|---|---|
| `NewCheckbox(items []CheckboxItem) *Checkbox` | Create checkbox list |
| `SetItems([]CheckboxItem) / Items() []CheckboxItem` | Set/get items |
| `MoveDown() / MoveUp() / Cursor() int / SetCursor(int)` | Navigation |
| `Toggle()` | Toggle current item |
| `CheckAll() / UncheckAll()` | Bulk operations |
| `OnChange func([]CheckboxItem)` | Change callback |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Space/Enter toggle, j/k, Ctrl+A/D |

### RadioGroup (Phase 19)

| Method | Description |
|---|---|
| `NewRadioGroup(items []RadioItem) *RadioGroup` | Create radio group |
| `SetItems([]RadioItem) / Items() []RadioItem` | Set/get items |
| `MoveDown() / MoveUp() / Cursor() int / SetCursor(int)` | Navigation |
| `Select()` | Select current item (clears previous) |
| `Selected() RadioItem` | Get selected item |
| `OnChange func(RadioItem)` | Change callback |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Space/Enter select, j/k nav |

### Slider (Phase 19)

| Method | Description |
|---|---|
| `NewSlider() *Slider` | Create slider |
| `SetRange(min, max int) / SetValue(int) / Value() int` | Range/value |
| `SetStep(int) / SetOrientation(SliderOrientation)` | Config |
| `SetFromRatio(float64) / AsRatio() float64` | Ratio-based positioning |
| `OnChange func(int)` | Change callback |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: arrows/h/l, Home/End |

### CommandPalette (Phase 19)

| Method | Description |
|---|---|
| `NewCommandPalette() *CommandPalette` | Create palette |
| `SetCommands([]Command) / Commands() []Command` | Set/get commands |
| `SetQuery(string) / Query() string` | Search filter |
| `FilteredCommands() []Command` | Filtered results |
| `MoveUp() / MoveDown() / Cursor() int` | Navigation |
| `CurrentCommand() Command` | Selected command |
| `Show() / Hide() / Visible() bool` | Visibility |
| `SetMaxVisible(int)` | Max visible results |
| `OnExecute func(Command) / OnDismiss func()` | Callbacks |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Up/Down/Tab, Enter, Esc |

### Spinner (Phase 19)

| Method | Description |
|---|---|
| `NewSpinner() *Spinner` | Create spinner |
| `SetLabel(string) / Label() string` | Display label |
| `SetPrefix(string) / Prefix() string` | Text before spinner |
| `SetFrameStyle(string)` | dots, arc, line, bounce, bars |
| `Start() / Stop() / Running() bool` | Animation control |
| `Update(delta time.Duration) bool` | Advance frame (returns if changed) |
| `CurrentFrame() string` | Current frame string |
| `SetFrameIndex(int) / FrameIndex() int` | Direct frame control (wraps) |

### SelectionManager (Phase 15/16)

| Method | Description |
|---|---|
| `NewSelectionManager() *SelectionManager` | Create selection manager |
| `StartSelection(row, col int)` | Begin text selection |
| `UpdateSelection(row, col int)` | Extend selection |
| `ClearSelection() / HasSelection() bool` | Selection state |
| `SelectedText(buf *buffer.Buffer) string` | Extract selected text |
| `HandleMouse(*term.MouseEvent) bool` | Mouse: drag to select |
| `CopyToClipboard() string` | OSC52 clipboard copy |

### ChatApp P16 Integration (Phase 16)

| Method | Description |
|---|---|
| `app.SetStatusBar(sb) / app.StatusBar()` | Attach status bar |
| `app.SetModel(string) / app.SetTokenRate(int)` | Status bar convenience |
| `app.SetContextWindow(used, total int) / app.UpdateClock()` | Status bar updates |
| `app.SetTabBar(tb) / app.TabBar()` | Attach tab bar |
| `app.AddSession(name string) int` | Add tab/session |
| `app.SwitchSession(idx) / NextSession() / PrevSession()` | Tab navigation |
| `app.CloseSession() / SessionCount() / ActiveSession()` | Tab management |
| `app.SetSelectionManager(sm) / app.SelectionManager()` | Attach selection |
| `app.HasSelection() / ClearSelection() / SelectedText(buf)` | Selection access |
| `app.HandleMouseP16(mouse)` | Enhanced mouse routing |
| `app.handleP16Keys(key)` | Alt+[/] tabs, Alt+W close, Alt+1-9 switch |

### Undo/Redo (Phase 20)

| Method | Description |
|---|---|
| `inputLine.SetUndoEnabled(true)` | Enable undo system |
| `inputLine.Undo() / Redo()` | Undo/redo last edit |
| `inputLine.CanUndo() / CanRedo() bool` | History state |
| `inputLine.ClearHistory()` | Reset undo stack |

### Theme Cycling (Phase 21)

| Method | Description |
|---|---|
| `app.CycleTheme() / SetTheme(name string)` | Switch themes at runtime |
| `theme.Get() / SetActive(name) / ListThemes() []string` | Theme management |
| 5 built-in: Dark, Light, Solarized, Dracula, Gruvbox | Pre-defined themes |

### Performance (Phase 24)

| Optimization | Impact |
|---|---|
| PaintVisible (binary search) | O(log n) — 31.5x faster for 1000 blocks |
| LinkManager scanner (no regex) | 13.2x faster URL detection |
| Zero-alloc Paint paths | StatusBar/TabBar/DiffPreview: 0 allocs |
| Buffer row-skip diff | Only dirty rows re-rendered |
| VirtualScroller | Only visible components painted |

## Phase 12: Data Widgets

### Table

Data table with auto column width, alignment, zebra striping, and sorting.

| Method | Description |
|---|---|
| `NewTable(headers []string, rows ...[]string)` | Create table with headers and optional rows |
| `SetRows([][]string) / AddRow([]string) / Rows()` | Row management |
| `SetHeaders([]string) / Headers()` | Header management |
| `SortBy(col int, asc bool)` | Sort rows by column |
| `SetCursor(int) / Cursor() / MoveDown() / MoveUp()` | Row cursor navigation |
| `SetSelectedStyle(Style) / SetZebraStyle(Style)` | Visual styling |
| `SetOnSelect(func(row int))` | Selection callback |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Tree

Hierarchical tree widget with expand/collapse, DFS flatten, and keyboard navigation.

| Method | Description |
|---|---|
| `NewTreeNode(id, label string)` | Create tree node |
| `SetRoot(node *TreeNode)` | Set root node |
| `AddChild(child *TreeNode) / Children()` | Node hierarchy |
| `SetExpanded(bool) / IsExpanded()` | Expand/collapse state |
| `ExpandAll() / CollapseAll()` | Bulk expand/collapse |
| `SetCursor(int) / Cursor() / MoveDown() / MoveUp()` | Cursor navigation |
| `Flatten() []FlatNode` | DFS flatten visible nodes |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Form

Interactive form with fields (TextField, CheckboxField, SelectField), Tab navigation, and validation.

| Method | Description |
|---|---|
| `NewForm()` | Create empty form |
| `AddField(field FormField) / Fields()` | Add/retrieve fields |
| `SetValues() / Values()` | Get/set all field values |
| `Validate() error` | Run all field validators |
| `SetOnSubmit(func()) / SetOnCancel(func())` | Form callbacks |
| `Next() / Prev() / SetCursor(int) / Cursor()` | Field navigation (Tab/Shift+Tab) |
| `HandleKey(*term.KeyEvent) bool` | Keyboard handler |
| `Measure / SetBounds / Paint / Children` | Component interface |

### ProgressBar

Determinate/indeterminate progress bar with color thresholds.

| Method | Description |
|---|---|
| `NewProgressBar()` | Create progress bar |
| `SetProgress(float64) / Progress()` | Set/get progress (0-100) |
| `SetWidth(int) / SetHeight(int)` | Size control |
| `SetStyle(ProgressBarStyle)` | Visual styling |
| `SetIndeterminate(bool)` | Indeterminate animation mode |
| `Measure / SetBounds / Paint / Children` | Component interface |

### StatusIndicator

Animated status indicator with 5 spinner styles (dots, arc, line, bounce, bars).

| Method | Description |
|---|---|
| `NewStatusIndicator()` | Create indicator |
| `SetStatus(string) / Status()` | Status text |
| `SetStyle(StatusStyle) / Style()` | Visual styling |
| `SetSpinnerStyle(string)` | Spinner animation style |
| `Start() / Stop() / Running()` | Animation control |
| `Update(dt time.Duration) bool` | Advance frame |
| `Measure / SetBounds / Paint / Children` | Component interface |

## Phase 13: Status & Notification Widgets

### Gauge

Linear, vertical, and radial gauges with color thresholds and gradient support.

| Method | Description |
|---|---|
| `NewGauge()` | Create gauge |
| `SetValue(float64) / Value()` | Set/get value (0.0-1.0) |
| `SetMax(float64) / Max()` | Maximum value |
| `SetType(GaugeType)` | GaugeLinear / GaugeVertical / GaugeRadial |
| `SetThresholds([]Threshold)` | Color thresholds (green/yellow/red) |
| `SetGradient(bool)` | Enable smooth color gradient |
| `SetLabel(string) / Label()` | Display label |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Sparkline

Unicode bar chart sparkline with autoscale and streaming push.

| Method | Description |
|---|---|
| `NewSparkline() *Sparkline` | Create sparkline |
| `Push(value float64)` | Add data point (streaming) |
| `SetData([]float64) / Data()` | Set/get all data |
| `SetMaxPoints(int)` | Max visible points |
| `SetColorMode(SparkColorMode)` | Single / Gradient / Value-based |
| `SetAutoscale(bool)` | Auto-adjust y-axis |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Badge

Status badge with 6 variants (info/success/warning/error/primary/secondary) and 3 sizes.

| Method | Description |
|---|---|
| `NewBadge(text string)` | Create badge |
| `SetText(string) / Text()` | Badge text |
| `SetVariant(BadgeVariant)` | Info / Success / Warning / Error / Primary / Secondary |
| `SetSize(BadgeSize)` | Small / Medium / Large |
| `SetStyle(BadgeStyle)` | Visual styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

**BadgeGroup:**

| Method | Description |
|---|---|
| `NewBadgeGroup()` | Create group |
| `Add(badge *Badge)` | Add badge |
| `SetSpacing(int)` | Gap between badges |

### Notification/Toast

Toast notification system with 4 levels (info/success/warning/error) and auto-expiry.

| Method | Description |
|---|---|
| `NewToastManager()` | Create manager (implements Component) |
| `Push(n Notification)` | Show notification |
| `Dismiss(id string)` | Dismiss specific notification |
| `Clear()` | Dismiss all |
| `Notifications() []Notification` | Get active notifications (deep copy) |
| `Tick(dt time.Duration)` | Advance auto-expiry timers |
| `SetPosition(NotificationPosition)` | TopLeft / TopRight / BottomLeft / BottomRight |
| `Measure / SetBounds / Paint / Children` | Component interface |

## Phase 14: Overlay & Interaction Widgets

### ContextMenu

Nested context menu with submenus, separators, disabled items, shortcuts, and icons.

| Method | Description |
|---|---|
| `NewContextMenu()` | Create menu |
| `AddItem(item *MenuItem) / AddLabel(id, label string) / AddSeparator()` | Add items |
| `Remove(id string) / Clear() / Items()` | Item management |
| `Show(x, y int) / Hide() / Visible() / Position()` | Visibility control |
| `MoveUp() / MoveDown() / SetCursor(int) / Cursor()` | Navigation (skips separators + disabled) |
| `Activate()` | Open submenu or fire Action callback |
| `HitTest(mx, my int) bool / ClickAt(mx, my int)` | Mouse interaction |
| `HandleKey(*term.KeyEvent) bool` | Keyboard: Up/Down/Enter/Esc/Right(submenu)/Left(close) |
| `Measure / SetBounds / Paint / Children` | Component interface |

**MenuItem:**

| Method | Description |
|---|---|
| `NewMenuItem(id, label string)` | Create item |
| `SetShortcut(string) / SetIcon(string) / SetEnabled(bool)` | Item config |
| `SetSubmenu(*ContextMenu) / HasSubmenu()` | Submenu support |
| `SetAction(func())` | Click action |

### Tooltip

Hover tooltip with smart positioning and configurable display delay.

| Method | Description |
|---|---|
| `NewTooltip()` | Create tooltip |
| `SetText(string) / Text()` | Tooltip text |
| `Show(x, y int) / Hide() / Visible()` | Visibility |
| `SetDelay(time.Duration) / Delay()` | Display delay |
| `SetStyle(TooltipStyle)` | Visual styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

### SplitPane

Draggable split pane with horizontal/vertical orientation and keyboard resize.

| Method | Description |
|---|---|
| `NewSplitPane(orientation SplitOrientation)` | Create pane (SplitHorizontal / SplitVertical) |
| `SetLeft(child Component) / SetRight(child Component)` | Set panes |
| `SetSplit(float64) / Split()` | Split ratio (0.0-1.0) |
| `SetMinSize(int)` | Minimum pane size |
| `SetDraggable(bool)` | Enable/disable mouse drag |
| `Measure / SetBounds / Paint / Children` | Component interface |

### HelpOverlay

Keyboard shortcut cheatsheet with search filter and scroll.

| Method | Description |
|---|---|
| `NewHelpOverlay(groups []HelpGroup)` | Create overlay |
| `SetGroups([]HelpGroup) / Groups()` | Set/get shortcut groups |
| `SetQuery(string) / Query() / AppendQuery(rune) / BackspaceQuery() / ClearQuery()` | Search filter (case-insensitive) |
| `FilteredGroups() / HasResults() / TotalRows()` | Filtered results |
| `SelectNext() / SelectPrev() / SetSelected(int) / SelectedIndex()` | Selection |
| `ScrollUp() / ScrollDown() / ScrollY()` | Scroll control |
| `SetTitle(string) / Title() / SetMaxWidth(int) / SetMaxHeight(int)` | Config |
| `SetStyle(HelpStyle)` | Visual styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

## `internal/hotkey` (package: `github.com/topcheer/fluui/internal/hotkey`)

### HotkeyManager

Configurable hotkey manager with key sequences, conflict detection, and scope support.

| Method | Description |
|---|---|
| `NewManager() / SetDefaultGroup(string)` | Create manager |
| `Register(action string, seq KeySequence, opts ...RegisterOpt) error` | Register binding |
| `Unregister(action string)` | Remove binding |
| `Enable(action string) / Disable(action string)` | Toggle bindings |
| `Match(*term.KeyEvent) (action string, result MatchResult)` | Match key event |
| `ResetPending() / HasPending() / PendingKeys()` | Multi-key sequence state |
| `SetSequenceTimeout(time.Duration)` | Pending key timeout (default 1s) |
| `SetAllowOverride(bool)` | Allow conflicting key reassignment |
| `Groups() / BindingsByGroup(string) / BindingsByScope(Scope)` | Query bindings |
| `HasConflict(seq KeySequence) bool` | Conflict detection |
| `ExportConfig() / ImportConfig()` | Serialization |

**Key Parsing:**

| Function | Description |
|---|---|
| `ParseCombo(s string) (KeyCombo, error)` | Parse "Ctrl+F", "Alt+X", "Shift+G" |
| `ParseSequence(s string) (KeySequence, error)` | Parse "g g", "Ctrl+K Ctrl+D" |
| `MustParseSequence(s string) KeySequence` | Panics on error |

## `internal/fuzzy` (package: `github.com/topcheer/fluui/internal/fuzzy`)

### Fuzzy Matcher

Subsequence fuzzy matcher with scoring and highlight positions.

| Function | Description |
|---|---|
| `Rank(query string, candidates []string) []Result` | Rank candidates by fuzzy match |
| `Filter(query string, candidates []string) []string` | Return only matching candidates |

**Result struct:**

| Field | Description |
|---|---|
| `OriginalIndex int` | Index in original candidates |
| `Score int` | Match score (higher = better) |
| `Highlight() []int` | Positions of matched characters |

## Performance (Phase 24)

### Render Pipeline Optimizations

| Optimization | Impact |
|---|---|
| LinkManager regex to scanner | 158us to 12us (13x faster for 100 URLs) |
| Markdown cache invalidation | 40% faster re-render of unchanged content |
| Row-skip diff renderer | O(changed rows) not O(all cells) |
| PaintVisible (virtual scroll) | Paints only ~8 visible blocks of 1000 |
| Binary search PaintVisible | 18ns vs 580ns linear (31.5x faster) |
| Zero-alloc Paint paths | StatusBar/TabBar/DiffPreview: 0 allocations |

### Benchmark Baselines (M2 Ultra)

| Component | Operation | Time |
|---|---|---|
| Table | Paint (80x24) | 10.3us |
| FilePicker | Paint | 4.7us |
| StatusBar | Paint | 783ns |
| TabBar | Paint | 1.2us |
| DiffPreview | Paint | 4.4us |
| LinkManager | DetectLinks (100 URLs) | 12us |
| PaintVisible | 1000 blocks | 664ns |
| Fuzzy Rank | 100 candidates | ~50us |

---

## Phase 12-13 Components

### Table (Phase 12)

| Method | Description |
|---|---|
| `NewTable(headers []string, rows ...[]string)` | Create table with headers and data |
| `SetHeaders([]string) / Headers()` | Column headers |
| `SetRows(...[]string) / Rows() / RowCount()` | Row data |
| `AddRow([]string) / RemoveRow(int)` | Modify rows |
| `SetCursor(int) / Cursor() / MoveDown() / MoveUp()` | Row navigation (wrap-around) |
| `Sort(col int) / SortAsc(int) / SortDesc(int)` | Column sorting (Ctrl+1-9) |
| `SetSelected(*int) / Selected()` | Selection tracking |
| `OnSelect(func(int))` | Row selection callback |
| `SetStyle(TableStyle) / Style()` | Styling (zebra, header, selected, border) |
| `SetAlignment([]Alignment)` | Per-column alignment (AlignLeft/Center/Right) |
| `SetZebraStripes(bool)` | Alternating row colors |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Tree (Phase 12)

| Method | Description |
|---|---|
| `NewTreeNode(id, label string)` | Create tree node |
| `SetRoot(*TreeNode)` | Set root node |
| `Root()` | Get root node |
| `ExpandAll() / CollapseAll()` | Expand/collapse entire tree |
| `Expand(string) / Collapse(string) / Toggle(string)` | Per-node expand/collapse |
| `IsExpanded(string)` | Check node expansion |
| `SetCursor(string) / Cursor()` | Current node ID |
| `MoveDown() / MoveUp()` | DFS-order navigation |
| `OnSelect(func(string))` | Node selection callback |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Form (Phase 12)

| Method | Description |
|---|---|
| `NewForm()` | Create empty form |
| `AddField(FormField)` | Add field (TextField/CheckboxField/SelectField) |
| `RemoveField(string)` | Remove field by ID |
| `Fields() / FieldCount()` | Query fields |
| `FieldValue(string) (string, error)` | Get field value |
| `SetFieldValue(string, string) error` | Set field value |
| `Validate() error` | Run all validators |
| `OnSubmit(func()) / OnCancel(func())` | Submit/cancel callbacks |
| `Next() / Prev() / SetCursor(int) / Cursor()` | Tab navigation |
| `SetStyle(FormStyle) / Style()` | Styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

**FormField interface:** `ID() / Label() / Value() / SetValue(string) / Validate() error`

### ProgressBar (Phase 12)

| Method | Description |
|---|---|
| `NewProgressBar()` | Create progress bar (0-100) |
| `SetProgress(float64) / Progress()` | Set/get progress (0-100) |
| `SetWidth(int)` | Bar width in cells |
| `SetStyle(ProgressBarStyle) / Style()` | Styling (fill/empty/border) |
| `SetIndeterminate(bool)` | Indeterminate animation mode |
| `Measure / SetBounds / Paint / Children` | Component interface |

### StatusIndicator (Phase 12)

| Method | Description |
|---|---|
| `NewStatusIndicator()` | Create status indicator |
| `SetStatus(string) / Status()` | Status text |
| `SetState(StatusState)` | State: StatusInfo/Success/Warning/Error/Loading |
| `Start() / Stop() / Running()` | Spinner animation control |
| `SetStyle(StatusStyle)` | Per-state styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Gauge (Phase 13)

| Method | Description |
|---|---|
| `NewGauge()` | Create gauge (0.0-1.0) |
| `SetValue(float64) / Value()` | Set/get value (0-1) |
| `SetOrientation(Orientation)` | GaugeHorizontal / GaugeVertical / GaugeRadial |
| `SetThresholds([]Threshold)` | Color thresholds (green/yellow/red) |
| `SetGradient(bool)` | Enable gradient fill (green→red) |
| `SetWidth / SetHeight` | Dimensions |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Sparkline (Phase 13)

| Method | Description |
|---|---|
| `NewSparkline()` | Create sparkline |
| `Push(float64)` | Add data point (scrolls window) |
| `SetData([]float64) / Data()` | Set/get data series |
| `SetWidth(int)` | Number of bars to display |
| `SetColorMode(SparkColorMode)` | Solid / Gradient / Threshold |
| `SetStyle(SparklineStyle)` | Styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

### Badge (Phase 13)

| Method | Description |
|---|---|
| `NewBadge(text string)` | Create badge |
| `SetText(string) / Text()` | Badge text |
| `SetVariant(BadgeVariant)` | Info/Success/Warning/Error/Primary/Secondary |
| `SetSize(BadgeSize)` | Small/Medium/Large |
| `SetStyle(BadgeStyle)` | Custom styling |
| `Measure / SetBounds / Paint / Children` | Component interface |
| `BadgeGroup: NewBadgeGroup()` | Horizontal layout container with spacing |

### Notification/Toast (Phase 13)

| Method | Description |
|---|---|
| `NewToastManager()` | Create toast manager (implements Component) |
| `Push(Notification)` | Add notification (auto-expires) |
| `Dismiss(string)` | Dismiss by ID |
| `Clear()` | Dismiss all |
| `Notifications()` | Get all notifications (deep copy) |
| `Tick(delta time.Duration)` | Advance expiry timers |
| `SetPosition(ToastPosition)` | TopLeft/TopRight/BottomLeft/BottomRight |
| `SetMaxVisible(int)` | Stack limit |
| `Measure / SetBounds / Paint / Children` | Component interface |

**Notification struct:** `ID, Level (Info/Success/Warning/Error), Title, Message string, Duration time.Duration`

---

## Phase 14 Components

### ContextMenu (Phase 14)

| Method | Description |
|---|---|
| `NewContextMenu()` | Create context menu |
| `AddItem(*MenuItem) / AddLabel(id, label) / AddSeparator()` | Add items |
| `Remove(id) / Clear()` | Remove items |
| `Items() / ItemCount() / ItemAt(int) / Find(id)` | Query items |
| `Show(x, y int) / Hide() / Visible() / Position()` | Visibility control |
| `MoveUp() / MoveDown() / SetCursor(int) / Cursor()` | Navigation (skips separators/disabled) |
| `Activate()` | Open submenu or fire Action + OnSelect |
| `HandleKey(*term.KeyEvent)` | Up/Down/Enter/Esc/Right(left)/Left(close) |
| `HitTest(mx, my) / ClickAt(mx, my)` | Mouse support |
| `SetStyle(ContextMenuStyle) / Style()` | Styling |
| `Measure / SetBounds / Paint / Children / String` | Component interface |

**MenuItem:** `ID, Label, Shortcut, Icon string; Enabled, Separator bool; Submenu *ContextMenu; Action func()`

### Tooltip (Phase 14)

| Method | Description |
|---|---|
| `NewTooltip(text string)` | Create tooltip |
| `SetText(string) / Text()` | Tooltip text |
| `Show(x, y int) / Hide() / Visible() / Position()` | Visibility |
| `SetDelay(time.Duration)` | Hover delay before showing |
| `SetStyle(TooltipStyle)` | Styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

### SplitPane (Phase 14)

| Method | Description |
|---|---|
| `NewSplitPane(top, bottom Component)` | Vertical split (top/bottom) |
| `NewHorizontalSplitPane(left, right Component)` | Horizontal split (left/right) |
| `SetSplit(float64) / Split()` | Split ratio (0.0-1.0) |
| `SetMinSize(int)` | Minimum pane size |
| `SetDraggable(bool) / Draggable()` | Drag handle to resize |
| `SetKeyboardResize(bool)` | Enable keyboard resize (Ctrl+arrows) |
| `Top() / Bottom()` | Access child panes |
| `Measure / SetBounds / Paint / Children` | Component interface |

### HelpOverlay (Phase 14)

| Method | Description |
|---|---|
| `NewHelpOverlay(groups []HelpGroup)` | Create help overlay |
| `SetGroups([]HelpGroup) / Groups()` | Set/get help groups |
| `SetQuery(string) / Query() / AppendQuery(rune) / BackspaceQuery() / ClearQuery()` | Search filter (case-insensitive) |
| `FilteredGroups() / HasResults() / TotalRows()` | Filtered results |
| `SelectNext() / SelectPrev() / SetSelected(int) / SelectedIndex()` | Row selection |
| `ScrollUp() / ScrollDown() / ScrollY()` | Scrolling |
| `SetTitle(string) / Title()` | Overlay title |
| `SetMaxWidth(int) / SetMaxHeight(int)` | Size limits |
| `SetStyle(HelpStyle) / Style() / DefaultHelpStyle()` | Styling |
| `Measure / SetBounds / Paint / Children` | Component interface |

**HelpGroup:** `Name string; Entries []HelpEntry` where `HelpEntry: Keys, Description string`

### HotkeyManager (Phase 14, `internal/hotkey`)

| Method | Description |
|---|---|
| `NewManager()` | Create hotkey manager |
| `Register(Binding) error` | Register key binding |
| `Unregister(string) / Get(string) / Bindings() / Count()` | Query/cleanup |
| `Match(*term.KeyEvent) (action string, result MatchResult)` | Match key event → action |
| `ResetPending() / HasPending() / PendingKeys()` | Multi-key sequence state |
| `Enable(string) / Disable(string) / SetDefaultGroup(string)` | Enable/disable by action name |
| `Groups() / BindingsByGroup(string) / BindingsByScope(Scope)` | Grouping |
| `HasConflict(Binding) bool` | Prefix conflict detection |
| `SetAllowOverride(bool) / SetSequenceTimeout(time.Duration)` | Configuration |
| `ExportConfig() / ImportConfig([]Binding)` | Serialization |

**ParseCombo("Ctrl+F") → KeyCombo, ParseSequence("g g") → KeySequence, MustParseSequence(...)**

---

## Phase 24: Performance Optimizations

### Render Pipeline Optimization
- LinkManager: regex-based URL detection replaced with O(n) scanner (158μs → 12μs, 13x faster)
- PaintVisible: binary search O(log n) for visible block range (31x faster than linear)
- Render diff: row-skip optimization reduces redundant paint calls

### Benchmark Results (Apple M2 Ultra)
| Component | Paint | Measure | Navigation |
|---|---|---|---|
| FilePicker | 4.7μs | 83ns | 24ns |
| Table | 10.3μs | — | — |
| StatusBar | 783ns | 101ns | — |
| TabBar | 1.2μs | 179ns | 24ns |
| DiffPreview | 4.4μs | — | 24ns |
| LinkManager (100 URLs) | 12μs | — | — |

---

## Phase 12: Data + Status Widgets

### Table

Sortable data grid with auto-sized columns and alignment.

| Method | Description |
|---|---|
| `NewTable(headers []string, rows ...[]string) *Table` | Create with headers and initial rows |
| `AddRow([]string)` | Append a row |
| `SetRows([][]string)` | Replace all rows |
| `Rows() [][]string` | Get all rows |
| `Columns() []Column` | Get column definitions |
| `RowCount() int` | Number of rows |
| `Cursor() / SetCursor(int)` | Cursor position |
| `MoveUp() / MoveDown()` | Navigate (wrap-around) |
| `ScrollUp(n) / ScrollDown(n)` | Scroll viewport |
| `SortBy(col int)` | Sort by column (toggles asc/desc) |
| `SortColumn() int` | Current sort column (-1 if none) |
| `SortAscending() bool` | Sort direction |
| `ZebraStriping(bool)` | Alternating row colors |
| `OnSelect func(int)` | Row selection callback |
| `SetStyle(TableStyle) / Style()` | Styling |
| `HandleKey(*term.KeyEvent) bool` | j/k/Ctrl+1-9/arrows |

### Tree

Expandable tree view with DFS flatten.

| Method | Description |
|---|---|
| `NewTree() *Tree` | Create empty tree |
| `NewTreeNode(id, label string) *TreeNode` | Create a node |
| `SetRoot(*TreeNode)` | Set root node |
| `Root() *TreeNode` | Get root |
| `ExpandAll() / CollapseAll()` | Bulk expand/collapse |
| `Cursor() / SetCursor(int)` | Cursor position |
| `MoveUp() / MoveDown()` | Navigate visible nodes |
| `ToggleExpand()` | Expand/collapse current |
| `ExpandNode() / CollapseNode()` | Expand/collapse explicitly |
| `SelectedNode() *TreeNode` | Node at cursor |
| `OnSelect func(*TreeNode)` | Selection callback |
| `SetStyle(TreeStyle) / Style()` | Styling |
| `HandleKey(*term.KeyEvent) bool` | j/k/Enter/Space/arrows |

### Form

Form fields with validation and Tab navigation.

| Method | Description |
|---|---|
| `NewForm() *Form` | Create empty form |
| `AddField(FormField)` | Add a field |
| `Fields() []FormField` | Get all fields |
| `NewTextField(label, key, defValue string) *TextField` | Text input field |
| `NewCheckboxField(label, key string, checked bool) *CheckboxField` | Checkbox field |
| `NewSelectField(label, key string, options []string) *SelectField` | Dropdown field |
| `Values() map[string]string` | Get all field values |
| `Validate() error` | Run field validations |
| `OnSubmit func(map[string]string)` | Submit callback |
| `OnCancel func()` | Cancel callback |
| `Cursor() / SetCursor(int)` | Field focus |
| `NextField() / PrevField()` | Tab / Shift+Tab |
| `HandleKey(*term.KeyEvent) bool` | Tab/Enter/Esc/arrows |

### ProgressBar

Determinate or indeterminate progress display.

| Method | Description |
|---|---|
| `NewProgressBar() *ProgressBar` | Create progress bar |
| `SetProgress(float64)` | Set progress (0.0-1.0) |
| `Progress() float64` | Current progress |
| `SetIndeterminate(bool)` | Spinner mode |
| `IsIndeterminate() bool` | Check mode |
| `SetLabel(string) / Label()` | Display label |
| `SetStyle(ProgressBarStyle)` | Styling |

### StatusIndicator

Animated spinner with status states.

| Method | Description |
|---|---|
| `NewStatusIndicator() *StatusIndicator` | Create indicator |
| `SetStatus(StatusState)` | Idle/Running/Success/Error/Warning |
| `Status() StatusState` | Current status |
| `SetSpinnerStyle(SpinnerStyle)` | dots/line/bounce/arc/bars |
| `Update(time.Duration) bool` | Advance frame |

---

## Phase 13: Gauge + Sparkline + Badge + Notification

### Gauge

Linear, vertical, or radial gauge with color thresholds.

| Method | Description |
|---|---|
| `NewGauge() *Gauge` | Create gauge |
| `SetValue(float64)` | Set value (0.0-1.0) |
| `Value() float64` | Current value |
| `SetType(GaugeType)` | GaugeLinear/GaugeVertical/GaugeRadial |
| `SetColorThresholds([]Threshold)` | Green/yellow/red bands |
| `DefaultThresholds() []Threshold` | Built-in 3-band thresholds |
| `SetGradient(bool)` | Enable gradient coloring |
| `SetLabel(string) / Label()` | Display label |

### Sparkline

Unicode bar chart with autoscale and streaming.

| Method | Description |
|---|---|
| `NewSparkline() *Sparkline` | Create sparkline |
| `Push(value float64)` | Append data point |
| `Values() []float64` | All data points |
| `SetMaxPoints(int)` | Maximum visible points |
| `SetColorMode(SparkColorMode)` | SparkSolid/SparkGradient/SparkThreshold |
| `Autoscale()` | Recalculate scale |

### Badge

Status badges with 6 variants and 3 sizes.

| Method | Description |
|---|---|
| `NewBadge(text string, variant BadgeVariant) *Badge` | Create badge |
| `NewBadgeWithSize(text, variant, size)` | Create with explicit size |
| `NewBadgeGroup() *BadgeGroup` | Create horizontal badge container |
| `SetText(string) / Text()` | Badge text |
| `SetVariant(BadgeVariant)` | Default/Success/Warning/Error/Info/Critical |
| `SetSize(BadgeSize)` | Small/Medium/Large |
| `Convenience: NewInfoBadge / NewSuccessBadge / NewWarningBadge / NewErrorBadge / NewCriticalBadge / NewNeutralBadge` |

### Notification (ToastManager)

Toast notifications with auto-expiry and stacking.

| Method | Description |
|---|---|
| `NewToastManager(maxVisible int) *ToastManager` | Create manager |
| `Push(message string, level NotificationLevel) string` | Show notification (returns ID) |
| `Dismiss(id string)` | Dismiss by ID |
| `Clear()` | Dismiss all |
| `Notifications() []Notification` | Get all active (deep copy) |
| `Tick(time.Duration)` | Advance expiry timers |
| `Levels: NotifyInfo / NotifySuccess / NotifyWarning / NotifyError` |

---

## Phase 14: Overlay + Interaction Widgets

### ContextMenu

Nested submenu context menu with keyboard/mouse navigation.

| Method | Description |
|---|---|
| `NewContextMenu() *ContextMenu` | Create menu |
| `NewMenuItem(id, label string) *MenuItem` | Create item |
| `NewSeparator() *MenuItem` | Create separator |
| `AddItem(*MenuItem) / AddLabel(id, label) / AddSeparator()` | Add items |
| `Remove(id string) / Clear()` | Remove items |
| `Items() []*MenuItem / ItemCount() / ItemAt(int)` | Query items |
| `Show(x, y int) / Hide()` | Visibility |
| `MoveUp() / MoveDown()` | Navigate (skips separators/disabled) |
| `SetCursor(int) / Cursor() / CurrentItem()` | Cursor state |
| `Activate()` | Open submenu or fire Action |
| `HandleKey(*term.KeyEvent) bool` | Up/Down/Enter/Esc/Right/Left |
| `HitTest(mx, my int) (int, bool)` | Mouse hit test |
| `ClickAt(mx, my int) bool` | Mouse click |
| `SetStyle(ContextMenuStyle) / Style()` | Styling |

**MenuItem chainable:** `.SetShortcut(s) / .SetIcon(s) / .SetEnabled(bool) / .SetSubmenu(*ContextMenu) / .SetAction(func())`

### Tooltip

Hover tooltip with smart positioning and configurable delay.

| Method | Description |
|---|---|
| `NewTooltip(text string) *Tooltip` | Create tooltip |
| `SetText(string) / Text()` | Tooltip content |
| `SetDelay(time.Duration) / Delay()` | Display delay |
| `Show(x, y int) / Hide()` | Visibility |
| `SetStyle(TooltipStyle)` | Styling |

### SplitPane

Draggable split pane with keyboard resize.

| Method | Description |
|---|---|
| `NewSplitPane(first, second Component) *SplitPane` | Create split |
| `SetDirection(SplitDirection)` | SplitHorizontal / SplitVertical |
| `SetRatio(float64)` | Split position (0.0-1.0) |
| `Ratio() float64` | Current ratio |
| `SetResizable(bool)` | Enable drag resize |
| `SetMinRatio(float64) / SetMaxRatio(float64)` | Resize bounds |
| `HandleKey(*term.KeyEvent) bool` | Ctrl+arrows to resize |

### HelpOverlay

Searchable keyboard shortcut cheatsheet.

| Method | Description |
|---|---|
| `NewHelpOverlay(groups []HelpGroup) *HelpOverlay` | Create overlay |
| `SetGroups([]HelpGroup) / Groups()` | Help groups |
| `SetQuery(string) / Query()` | Search text |
| `AppendQuery(rune) / BackspaceQuery() / ClearQuery()` | Edit search |
| `FilteredGroups() []HelpGroup` | Filtered results |
| `TotalRows() int / HasResults() bool` | Result queries |
| `SelectNext() / SelectPrev()` | Navigate results |
| `ScrollUp() / ScrollDown()` | Scroll results |
| `SetTitle(string) / SetMaxWidth(int) / SetMaxHeight(int)` | Configuration |
| `SetStyle(HelpStyle) / DefaultHelpStyle()` | Styling |

---

## Phase 24: Performance Optimization

### Render Pipeline Optimization

The render pipeline was optimized in Phase 24 for 44-48% faster rendering with 50% fewer allocations:

- **Row-skip diff**: Unchanged rows are skipped entirely (O(1) per row check)
- **Cell-level diff**: Only changed cells generate ANSI output
- **Batch ANSI sequences**: Consecutive cells with same style share one SGR escape

### Link Detection Optimization

`LinkManager.AnnotateBuffer` was rewritten from regex to hand-rolled scanner:

| Operation | Before (regex) | After (scanner) | Speedup |
|---|---|---|---|
| DetectLinks 100 URLs | 158 us | 12 us | 13.2x |
| ScanText 100 lines | 182 us | 21 us | 8.5x |

### Markdown Render Caching

`AssistantTextBlock` caches rendered markdown cells:

- First paint: full goldmark + chroma rendering
- Subsequent paints: cached cell copy (99.85% fewer allocations)
- Cache invalidated on content change

---

## Phase 25: Fuzz Testing

Fluui includes 6 native Go fuzz tests across 3 packages:

| Fuzz Target | Package | Description |
|---|---|---|
| `FuzzParserFeed` | `internal/term/` | Random bytes through terminal input parser |
| `FuzzParserFeedChunked` | `internal/term/` | Chunk-invariant parsing verification |
| `FuzzBufferSetCell` | `internal/buffer/` | Random coords on SetCell/GetCell/Fill |
| `FuzzBufferDrawText` | `internal/buffer/` | Random text + positions on DrawText |
| `FuzzBufferBlit` | `internal/buffer/` | Random source/dest coords on Blit |
| `FuzzRendererRender` | `markdown/` | Malformed/extreme markdown through goldmark |

Millions of executions, zero panics, zero deadlocks.
