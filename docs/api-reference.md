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
