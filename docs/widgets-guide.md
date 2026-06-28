# Widgets Guide

This guide covers Fluui's Phase 15 interactive widgets: FilePicker, TabBar, StatusBar, and DiffPreview.

## FilePicker

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
