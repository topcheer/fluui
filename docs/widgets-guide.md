# Widgets Guide

This guide covers Fluui's interactive widgets from Phase 15 through Phase 19.

## Phase 15 Widgets

### FilePicker

The `FilePicker` component provides a file browser with fuzzy filtering, multi-select, and vim-style keybindings.

### Basic Usage

```go
fp := component.NewFilePicker(".")
fp.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
fp.SetOnConfirm(func(files []component.FileEntry) {
    fmt.Printf("Selected %d files\n", len(files))
})
buf := buffer.NewBuffer(60, 20)
fp.Paint(buf)
```

### Navigation

| Key | Action |
|---|---|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `h` / `Left` | Go to parent directory |
| `l` / `Right` / `Enter` | Enter directory or select file |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Space` | Toggle file selection |
| `/` | Enter filter mode |
| `Esc` | Exit filter mode / Clear filter |
| `Enter` | Confirm selection (fires OnConfirm) |

### API

```go
// Construction
fp := component.NewFilePicker(dir string)

// Navigation
fp.MoveDown()
fp.MoveUp()
fp.EnterDir()
fp.GoUp()
fp.Cwd() string

// Filtering
fp.SetFilter("go")
fp.Filter() string
fp.AppendFilter('x')
fp.BackspaceFilter()
fp.FilteredCount() int

// Selection
fp.ToggleSelect()
fp.IsSelected(path string) bool
fp.SelectedFiles() []FileEntry
fp.ClearSelection()

// State
fp.Entries() []FileEntry
fp.Cursor() int
fp.CurrentEntry() (FileEntry, bool)

// Callbacks
fp.SetOnSelect(func(FileEntry))
fp.SetOnConfirm(func([]FileEntry))
fp.SetOnError(func(error))

// Styling
fp.SetStyle(component.DefaultFilePickerStyle())
fp.Style() FilePickerStyle

// Component interface
fp.Measure(constraints) Size
fp.SetBounds(rect)
fp.Paint(buf)
fp.HandleKey(key *term.KeyEvent) bool
```

### FileEntry

```go
type FileEntry struct {
    Name    string
    Path    string
    IsDir   bool
    Size    int64
    Mode    os.FileMode
    ModTime int64
}
```

## TabBar

The `TabBar` component manages tabs with close buttons, keyboard navigation, and active/hover styles.

### Basic Usage

```go
tb := component.NewTabBar()
tb.AddTab("chat-1", "Chat 1")
tb.AddTab("chat-2", "Chat 2")
tb.SetActive(0)
tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
buf := buffer.NewBuffer(80, 1)
tb.Paint(buf)
```

### Navigation

| Key | Action |
|---|---|
| `Tab` / `Right` | Next tab |
| `Shift+Tab` / `Left` | Previous tab |
| `Enter` | Activate tab |
| `Ctrl+W` / `x` | Close tab |

### API

```go
// Construction
tb := component.NewTabBar()

// Tab management
tb.AddTab(id, title string)
tb.AddClosableTab(id, title string)
tb.RemoveTab(id string)
tb.Tabs() []Tab
tb.TabCount() int

// Selection
tb.SetActive(index int)
tb.Active() int
tb.NextTab()
tb.PrevTab()

// Styling
tb.SetStyle(component.DefaultTabBarStyle())
tb.Style() TabBarStyle

// Component interface
tb.Measure(constraints) Size
tb.SetBounds(rect)
tb.Paint(buf)
tb.HandleKey(key *term.KeyEvent) bool
tb.HitTest(mx, my int) (int, bool)
tb.ClickAt(mx, my int) bool
```

## StatusBar

The `StatusBar` component displays status items aligned left, center, or right. Includes AI agent convenience methods.

### Basic Usage

```go
sb := component.NewStatusBar()
sb.AddLeft(component.NewStatusItem("left", "MODE: NORMAL"))
sb.AddCenter(component.NewStatusItem("center", "Ready"))
sb.AddRight(component.NewStatusItem("right", "100%"))
sb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
buf := buffer.NewBuffer(80, 1)
sb.Paint(buf)
```

### AI Agent Convenience

```go
sb.SetModel("GPT-4")
sb.SetTokenRate(1500)        // "1.5k tok/s"
sb.SetContextWindow(8000, 128000)  // "8k/128k"
sb.SetClock("14:32")
```

### API

```go
// Construction
sb := component.NewStatusBar()

// Items
sb.AddLeft(item StatusItem)
sb.AddCenter(item StatusItem)
sb.AddRight(item StatusItem)
sb.RemoveItem(id string)
sb.Clear()
sb.Items() []StatusItem
sb.ItemCount() int

// Updates
sb.SetItemText(id, text string)
sb.SetItemStyle(id string, style buffer.Style)
sb.SetSeparator(" | ")
sb.Separator() string

// Styling
sb.SetStyle(component.DefaultStatusBarStyle())
sb.Style() StatusBarStyle
sb.SetHeight(h int)

// Component interface
sb.Measure(constraints) Size
sb.SetBounds(rect)
sb.Paint(buf)
```

### StatusItem

```go
type StatusItem struct {
    ID        string
    Text      string
    Alignment StatusItemAlignment  // StatusAlignLeft | StatusAlignCenter | StatusAlignRight
    Style     buffer.Style
}
```

## DiffPreview

The `DiffPreview` component renders git diffs with syntax highlighting, scrollable viewport, and stats.

### Basic Usage

```go
dp := component.NewDiffPreview()
dp.SetDiff(`diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
-old
+new
+added`)
dp.SetTitle("main.go")
dp.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 20})
buf := buffer.NewBuffer(80, 20)
dp.Paint(buf)
```

### Diff Line Types

| Type | Prefix | Description |
|---|---|---|
| `DiffContext` | ` ` (space) | Unchanged context line |
| `DiffAdd` | `+` | Added line (green) |
| `DiffDel` | `-` | Removed line (red) |
| `DiffHunk` | `@@` | Hunk header (cyan) |
| `DiffFile` | `diff --git` | File header (bold) |
| `DiffMeta` | `---`, `+++`, `index`, etc. | Meta lines (dim) |

### API

```go
// Construction
dp := component.NewDiffPreview()

// Content
dp.SetDiff(diffText string)
dp.Lines() []DiffLine
dp.LineCount() int
dp.Stats() DiffStats
dp.IsEmpty() bool
dp.HasChanges() bool
dp.DiffSummary() string

// Scrolling
dp.ScrollDown(n int)
dp.ScrollUp(n int)
dp.ScrollTo(row int)
dp.ScrollY() int
dp.ScrollPageDown(viewHeight int)
dp.ScrollPageUp(viewHeight int)
dp.VisibleRange() (start, end int)

// Configuration
dp.SetTitle(title string)
dp.Title() string
dp.SetStyle(component.DefaultDiffPreviewStyle())
dp.Style() DiffPreviewStyle
dp.SetShowLineNumbers(show bool)
dp.ShowLineNumbers() bool

// Component interface
dp.Measure(constraints) Size
dp.SetBounds(rect)
dp.Paint(buf)
dp.Children() []Component
dp.String() string
```

### DiffStats

```go
type DiffStats struct {
    Additions  int
    Deletions  int
    Files      int
    Hunks      int
    TotalLines int
}
// String: "+2 -1 (1 files, 1 hunks)"
```

### Parsing

```go
// Parse a diff string into classified lines
lines := component.ParseDiff(diffText)
for _, line := range lines {
    fmt.Printf("%v: %s\n", line.Type, line.Content)
}
```

---

## ChatApp Integration (Phase 16)

All P15 widgets integrate into ChatApp via the `app/chat_p16.go` layer.

### Quick Start

```go
app := app.NewChatApp(120, 40)

// Attach StatusBar
sb := component.NewStatusBar()
app.SetStatusBar(sb)
app.SetModel("GPT-4")
app.SetTokenRate(1500)           // "1.5k tok/s"
app.SetContextWindow(8000, 128000)

// Attach TabBar for multi-session
tb := component.NewTabBar()
app.SetTabBar(tb)
app.AddSession("Research")       // returns index 0
app.AddSession("Debugging")      // returns index 1
app.AddSession("Refactoring")    // returns index 2

// Attach SelectionManager
sm := app.NewSelectionManager()
app.SetSelectionManager(sm)
```

### Multi-Session Navigation

```go
app.NextSession()                // wrap-around next tab
app.PrevSession()                // wrap-around prev tab
app.SwitchSession(1)             // jump to tab 1
app.CloseSession()               // close active tab

// Query sessions
sessions := app.Sessions()       // []SessionInfo
name := app.ActiveSessionName()  // "Debugging"
count := app.SessionCount()      // 3
```

### Keyboard Shortcuts

| Key | Action |
|---|---|
| Alt+] | Next session |
| Alt+[ | Previous session |
| Alt+W | Close active session |
| Alt+1 to Alt+9 | Switch to session N |

### Mouse Routing Priority

Mouse events are routed in priority order via `HandleMouseP16()`:

1. **Overlays** — modal dialogs, popups
2. **Tab bar** — tab clicks + close buttons
3. **Selection** — drag-to-select text
4. **Scroll** — wheel up/down
5. **Custom handler** — user-defined `onMouse` callback

### Rendering

The `renderP16()` method handles layout:

- **Tab bar** painted at the top (y=0, height=1)
- **Status bar** painted at the bottom (y=h-1, height=1)
- **Selection highlights** applied to the buffer

---

## Phase 18-19 Components

### Dialog (Phase 18)

```go
dialog := component.NewConfirmDialog("Delete File?", "Are you sure?")
dialog.OnConfirm = func() { os.Remove(path) }
dialog.OnCancel = func() { /* dismissed */ }
dialog.Show()
// HandleKey: Enter=confirm, Esc=cancel, Tab=navigate buttons
```

Four dialog types: `DialogConfirm` (Yes/No), `DialogInfo` (OK), `DialogPrompt` (text input), and `DialogCustom` (arbitrary buttons).

### AutoComplete (Phase 18)

```go
ac := component.NewAutoComplete()
ac.SetItems([]component.CompletionItem{
    {Label: "gpt-4", Detail: "OpenAI"},
    {Label: "claude-3", Detail: "Anthropic"},
})
ac.SetQuery("gp")
ac.Show(x, y)
// HandleKey: Up/Down=navigate, Tab/Enter=select, Esc=dismiss
```

Fuzzy-filtered popup with `MaxVisible` control, `OnSelect(item)` and `OnDismiss()` callbacks.

### Wizard (Phase 18)

```go
wizard := component.NewWizard([]*component.WizardStep{
    component.NewWizardStep("welcome", "Welcome").
        SetDescription("Let's get started"),
    component.NewWizardStep("config", "Configuration"),
    component.NewWizardStep("done", "Complete"),
})
wizard.SetOnFinish(func(w *component.Wizard) { start() })
// HandleKey: Tab/Right=Next, Left=Back, Enter=activate, Esc=cancel
```

Multi-step wizard with per-step `OnEnter`/`OnLeave` lifecycle hooks, dynamic button ordering, and `OnFinish`/`OnCancel` callbacks.

### Checkbox (Phase 19)

```go
cb := component.NewCheckbox([]string{
    "Enable streaming",
    "Show thinking",
    "Verbose mode",
})
cb.OnChange = func(idx int, checked bool) {
    fmt.Printf("Item %d: %v\n", idx, checked)
}
// HandleKey: Space/Enter=toggle, j/Down=next, k/Up=prev
// Ctrl+A=check all, Ctrl+D=uncheck all
```

Multi-item checkbox list with wrap-around cursor, disabled item skipping, and `CheckedItems()` / `CheckedLabels()` query methods.

### RadioGroup (Phase 19)

```go
rg := component.NewRadioGroup([]string{"Light", "Dark", "Auto"})
rg.OnChange = func(value string) {
    fmt.Println("Selected:", value)
}
// HandleKey: Up/Down/j/k=navigate, Enter/Space=select
```

Mutually exclusive single-selection group. `SetDisabled(idx)` disables an item and clears it if active.

### Slider (Phase 19)

```go
slider := component.NewSliderWithRange(0, 100, 50, 5)
slider.SetOrientation(component.SliderHorizontal)
slider.OnChange = func(value float64) {
    fmt.Printf("%.0f%%\n", value)
}
// HandleKey: Left/Right or h/l=step, Up/Down=large step
//             Home=min, End=max
```

Configurable range slider with horizontal/vertical orientation, `SetFromRatio(ratio)` for percentage control, and `Ratio()` for current ratio.

### CommandPalette (Phase 19)

```go
palette := component.NewCommandPalette()
palette.AddCommand(component.Command{
    ID:       "settings.theme",
    Label:    "Change Theme",
    Category: "Settings",
    Action:   func() { openThemePicker() },
})
palette.Show(x, y)
// HandleKey: printable=filter, Up/Down/Tab=navigate (wrap)
//             Enter=execute, Esc=dismiss, Backspace=edit query
```

VS Code-style fuzzy command palette with highlighted match segments (via `internal/fuzzy`), scroll management (`MaxVisible`), and `OnExecute(cmd)` / `OnDismiss()` callbacks.

### Spinner (Phase 19)

```go
spinner := component.NewSpinner("Loading...")
spinner.SetFrameStyle(animation.SpinnerDots)
spinner.Start()
// In render loop:
spinner.Update(dt) // returns true if frame changed
spinner.Paint(buf)
spinner.Stop()
```

Animated spinner with configurable frame styles (dots, arc, line, bounce, bars), label and prefix support, and `SetFrameIndex(idx)` for manual frame control with wrap-around.

---

## Phase 18-19 Components

### Dialog (Phase 18)

Modal dialog component with four types: Confirm, Info, Prompt, and Custom buttons.

```go
dialog := component.NewConfirmDialog("Delete File", "Are you sure?")
dialog.OnConfirm = func() { os.Remove(path) }
dialog.OnCancel = func() { /* cleanup */ }
dialog.Show()
```

Key bindings:
- `Tab` / `Left` / `Right` — navigate between buttons
- `Enter` / `Space` — activate focused button
- `Esc` — cancel

Prompt dialogs support full text editing:
```go
prompt := component.NewPromptDialog("Name", "Enter your name:", "")
prompt.SetInputValue("default")
// HandleKey processes text input (InsertRune, Backspace, etc.)
```

### AutoComplete (Phase 18)

Popup autocomplete with fuzzy filtering for suggestions.

```go
ac := component.NewAutoComplete()
ac.SetItems([]component.CompletionItem{
    {Label: "@alice", Detail: "User"},
    {Label: "@bob", Detail: "User"},
    {Label: "/help", Detail: "Command"},
})
ac.OnSelect = func(item component.CompletionItem) {
    fmt.Println("Selected:", item.Label)
}
ac.Show(x, y)
```

Key bindings:
- `Up` / `Down` — navigate filtered results
- `Tab` / `Enter` — select current item
- `Esc` — dismiss

### Wizard (Phase 18)

Multi-step wizard with lifecycle hooks and dynamic button ordering.

```go
wizard := component.NewWizard([]*component.WizardStep{
    component.NewWizardStep("welcome", "Welcome").
        SetDescription("Let's configure your project"),
    component.NewWizardStep("config", "Configuration").
        SetDescription("Choose settings"),
    component.NewWizardStep("done", "Complete").
        SetDescription("All done!"),
})
wizard.SetOnFinish(func(w *component.Wizard) { startApp() })
wizard.SetOnCancel(func(w *component.Wizard) { os.Exit(0) })
```

Key bindings:
- `Tab` / `Right` — Next button
- `Shift+Tab` / `Left` — Back button
- `Enter` — activate focused button
- `Esc` — cancel wizard
- `Ctrl+N` — skip to next step
- `Ctrl+B` — go back

### Checkbox (Phase 19)

Multi-select checkbox list with bulk operations and vim navigation.

```go
cb := component.NewCheckbox([]component.CheckboxItem{
    {Label: "Enable streaming", Checked: true},
    {Label: "Show thinking", Checked: false},
    {Label: "Verbose mode", Checked: false},
})
cb.OnChange = func(items []component.CheckboxItem) {
    for _, item := range items {
        if item.Checked {
            fmt.Println("Checked:", item.Label)
        }
    }
}
```

Key bindings:
- `Space` / `Enter` — toggle current item
- `j` / `Down` — move cursor down (wraps to top)
- `k` / `Up` — move cursor up (wraps to bottom)
- `Ctrl+A` — check all items
- `Ctrl+D` — uncheck all items

### RadioGroup (Phase 19)

Mutually exclusive selection group — selecting one clears the previous.

```go
rg := component.NewRadioGroup([]component.RadioItem{
    {Label: "GPT-4"},
    {Label: "Claude-3"},
    {Label: "Gemini"},
})
rg.SetSelected(0) // Select GPT-4
rg.OnChange = func(item component.RadioItem) {
    fmt.Println("Selected model:", item.Label)
}
```

Key bindings:
- `j` / `Down` — next item (wraps, skips disabled)
- `k` / `Up` — previous item (wraps, skips disabled)
- `Space` / `Enter` — select current item

### Slider (Phase 19)

Configurable range slider with horizontal/vertical orientation.

```go
slider := component.NewSlider()
slider.SetRange(0, 100)
slider.SetValue(50)
slider.SetStep(5)
slider.SetOrientation(component.SliderHorizontal)
slider.OnChange = func(val int) {
    fmt.Printf("Value: %d%%\n", val)
}
// Set by ratio (0.0 = min, 1.0 = max)
slider.SetFromRatio(0.75) // sets to 75
```

Key bindings:
- `Left` / `h` — decrement by step
- `Right` / `l` — increment by step
- `Up` / `Down` — large step (10x)
- `Home` — jump to minimum
- `End` — jump to maximum

### CommandPalette (Phase 19)

VS Code-style fuzzy command search palette.

```go
palette := component.NewCommandPalette()
palette.SetCommands([]component.Command{
    {ID: "file.new", Label: "New File", Category: "File"},
    {ID: "file.open", Label: "Open File", Category: "File"},
    {ID: "theme.cycle", Label: "Switch Theme", Category: "Settings"},
    {ID: "settings.toggle", Label: "Toggle Settings", Category: "Settings"},
})
palette.SetMaxVisible(8)
palette.OnExecute = func(cmd component.Command) {
    fmt.Println("Executing:", cmd.ID)
}
palette.OnDismiss = func() {
    fmt.Println("Palette dismissed")
}
palette.Show(x, y)
```

Key bindings:
- Printable chars — filter commands (fuzzy search)
- `Backspace` — edit query
- `Up` / `Down` / `Tab` — navigate results (wrap-around)
- `Enter` — execute selected command
- `Esc` — dismiss palette

### Spinner (Phase 19)

Animated spinner with configurable frame styles.

```go
spinner := component.NewSpinner()
spinner.SetLabel("Loading models...")
spinner.SetPrefix("> ")
spinner.SetFrameStyle(component.SpinnerFramesDots) // dots, arc, line, bounce, bars
spinner.Start()

// In render loop:
changed := spinner.Update(16 * time.Millisecond)
if changed {
    // Redraw only if frame changed
    spinner.Paint(buf)
}

spinner.Stop()
```

Frame styles available:
- `SpinnerFramesDots` — Braille dots animation
- `SpinnerFramesArc` — Arc rotation
- `SpinnerFramesLine` — Line spinner
- `SpinnerFramesBounce` — Bouncing dots
- `SpinnerFramesBars` — Vertical bars

### InputLine Undo/Redo (Phase 19)

The InputLine now supports undo/redo for all text mutations.

Key bindings:
- `Ctrl+Z` — undo last edit
- `Ctrl+Shift+Z` / `Ctrl+Y` — redo

Undo state is saved before each mutation:
- Printable character insertion
- Backspace
- `Ctrl+U` (clear line)
- `Ctrl+W` (delete word)

Maximum 100 undo states (oldest dropped). Redo stack is cleared on new edit (standard semantics).

```go
// Public API
il := app.NewInputLine("> ")
il.CanUndo()           // bool
il.CanRedo()           // bool
il.Undo()              // bool — restore previous state
il.Redo()              // bool — restore next state
il.UndoCount()         // int
il.RedoCount()         // int
il.ClearUndoHistory()  // reset stacks
```

### ChatApp Theme Management (Phase 19)

ChatApp supports theme cycling via keyboard shortcuts.

Key bindings:
- `Ctrl+]` — cycle theme forward
- `Ctrl+\` — cycle theme backward
- `Ctrl+T` — cycle forward (existing, still works)
- `Ctrl+Shift+T` — cycle backward (existing, still works)

```go
// Public API
chat := app.NewChatApp(80, 24)
chat.ThemeCount()           // int — number of builtin themes
chat.ThemeList()            // []string — theme names
chat.ThemeIndex()           // int — current theme index
chat.ThemeName()            // string — current theme name
chat.SetThemeByIndex(idx)   // switch to index
chat.SetThemeByName(name)   // switch by name (returns bool)
chat.CycleTheme()           // cycle forward
chat.CycleThemeBack()       // cycle backward
```
