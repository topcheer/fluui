# Component System

Fluui's component system provides composable UI primitives with a measure/paint layout model.

## Component Interface

```go
type Component interface {
    Measure(Constraints) Size
    SetBounds(Rect)
    Bounds() Rect
    Paint(*buffer.Buffer)
    Children() []Component
}
```

## Base Component

`BaseComponent` provides default `ID()`, `SetID()`, `Bounds()`, and `SetBounds()` implementations. Embed it in custom components.

```go
type MyWidget struct {
    component.BaseComponent
    // ...
}
```

## Built-in Components

### Layout Components

| Component | Package | Description |
|---|---|---|
| Flex (Row/Column) | `component/layout/` | Flexible horizontal/vertical layout |
| Stack | `component/layout/` | Z-axis stacking |
| Center | `component/layout/` | Center child within bounds |
| Padding | `component/layout/` | Add padding around child |

### Primitive Components

| Component | Description |
|---|---|
| Text | Single-line text rendering with style |
| Border | Unicode box-drawing border (┌─┐│└─┘) |
| ScrollView | Scrollable viewport with virtual scrolling |
| TextArea | Multi-line text editor with cursor support |

### Data Display Components

| Component | Description |
|---|---|
| Table | Sortable data grid with column alignment |
| Tree | Expandable tree view with DFS flatten |
| Form | Form fields (Text/Checkbox/Select) with validation |
| ProgressBar | Determinate/indeterminate progress display |
| StatusIndicator | Animated spinner with 5 styles |
| Gauge | Linear/vertical/radial gauge with thresholds |
| Sparkline | Unicode bar chart with gradient colors |
| Badge | Status badges with 6 variants and 3 sizes |

### Interactive Components (Phase 14+)

| Component | Description |
|---|---|
| ContextMenu | Nested submenu context menu with keyboard/mouse nav |
| Tooltip | Hover tooltip with smart positioning |
| SplitPane | Draggable split pane with keyboard resize |
| HelpOverlay | Searchable keyboard shortcut cheatsheet |
| HotkeyManager | Configurable hotkeys with key sequences |

### File & Navigation Components (Phase 15)

| Component | Description |
|---|---|
| **FilePicker** | File browser with fuzzy filter, multi-select, vim keys |
| **TabBar** | Tab management with close buttons, keyboard nav |
| **StatusBar** | Status bar with left/center/right item alignment |
| **DiffPreview** | Scrollable diff viewer with syntax highlighting |

### Overlay & Utility Components (Phase 15)

| Component | Description |
|---|---|
| **LinkManager** | URL detection, OSC8 hyperlinks, click handling |
| **SelectionManager** | Text selection with OSC52 clipboard copy |
| **Notification** | Toast notifications with auto-expiry (4 levels) |

## Size & Constraints

```go
type Constraints struct {
    MaxWidth  int
    MaxHeight int
}

type Size struct {
    W, H int
}

type Rect struct {
    X, Y, W, H int
}
```

## Creating a Custom Component

```go
type Label struct {
    component.BaseComponent
    text  string
    style buffer.Style
}

func NewLabel(text string) *Label {
    l := &Label{text: text, style: buffer.DefaultStyle}
    l.SetID(component.GenerateID("label"))
    return l
}

func (l *Label) Measure(cs Constraints) Size {
    w := buffer.StringWidth(l.text)
    if cs.MaxWidth > 0 && w > cs.MaxWidth {
        w = cs.MaxWidth
    }
    return Size{W: w, H: 1}
}

func (l *Label) Paint(buf *buffer.Buffer) {
    b := l.Bounds()
    buf.DrawText(b.X, b.Y, l.text, l.style)
}
```

## Data Widgets (Phase 12+)

### Table

```go
table := component.NewTable()
table.SetColumns([]component.Column{
    {Title: "Name", Width: 20},
    {Title: "Status", Width: 10},
    {Title: "Score", Width: 8, Align: component.AlignRight},
})
table.AddRow([]string{"Alice", "Active", "95"})
table.AddRow([]string{"Bob", "Idle", "72"})
table.OnSelect = func(row int) { fmt.Println("Selected:", row) }
```

Auto-sized columns with alignment (left/center/right), zebra striping, sort by `Ctrl+1-9`, keyboard navigation, and `OnSelect` callback.

### Tree

```go
tree := component.NewTree()
root := tree.AddNode(nil, "Project", true)
tree.AddNode(root, "src/", true)
tree.AddNode(root, "README.md", false)
tree.AddNode(root, "go.mod", false)
```

DFS-flattened tree with expand/collapse icons, leaf icons, cursor highlight, and full keyboard navigation. Supports infinite nesting.

### Form

```go
form := component.NewForm()
form.AddField(component.NewTextField("username", "Username"))
form.AddField(component.NewCheckboxField("remember", "Remember me"))
form.AddField(component.NewSelectField("role", "Role", []string{"Admin", "User"}))
form.OnSubmit = func(values map[string]string) { /* process */ }
```

FormField interface with TextField, CheckboxField, SelectField. Tab navigation, validation, Enter submit, Esc cancel.

## Status & Progress Widgets (Phase 12-13)

### ProgressBar

```go
bar := component.NewProgressBar()
bar.SetProgress(0.75) // 75%
bar.SetIndeterminate(true) // Spinner mode
```

Determinate (fills proportionally) and indeterminate (animated) modes.

### StatusIndicator

```go
status := component.NewStatusIndicator()
status.SetStatus(component.StatusRunning)
status.SetSpinnerStyle(component.SpinnerDots)
```

5 spinner styles (dots, line, bounce, arc, bars) with status states (idle/running/success/error/warning).

### Gauge

```go
gauge := component.NewGauge()
gauge.SetValue(0.65) // 65%
gauge.SetType(component.GaugeLinear) // or GaugeVertical, GaugeRadial
gauge.SetColorThresholds(component.DefaultThresholds())
```

Linear, vertical, and radial gauges with color thresholds (green/yellow/red bands) and gradient support for load monitoring.

### Sparkline

```go
spark := component.NewSparkline()
spark.Push(42.0)
spark.Push(38.5)
spark.Push(51.2)
spark.SetColorMode(component.SparkGradient) // or SparkSolid, SparkThreshold
```

Unicode bar chart with autoscale and streaming `Push()` for real-time data. Three color modes: solid, gradient, threshold.

### Badge

```go
badge := component.NewBadge("NEW", component.BadgeSuccess)
badge.SetSize(component.BadgeMedium)

group := component.NewBadgeGroup()
group.Add(badge1)
group.Add(badge2)
```

6 variants (default/success/warning/error/info/critical) and 3 sizes. BadgeGroup for horizontal layout.

## Overlay & Interaction Widgets (Phase 14)

### ContextMenu

```go
menu := component.NewContextMenu()
menu.AddItem(component.NewMenuItem("save", "Save"))
menu.AddSeparator()
menu.AddItem(component.NewMenuItem("exit", "Exit").SetShortcut("Ctrl+Q"))
menu.Show(x, y)
```

Nested submenus, keyboard/mouse navigation, separators, disabled items, shortcuts, and icons.

### Tooltip

```go
tip := component.NewTooltip("Click to expand")
tip.SetDelay(500 * time.Millisecond)
```

Hover tooltip with smart positioning and configurable display delay.

### SplitPane

```go
split := component.NewSplitPane(left, right)
split.SetDirection(component.SplitHorizontal)
split.SetRatio(0.3) // 30% left
split.SetResizable(true)
```

Draggable split pane with keyboard resize. Horizontal or vertical direction.

### HelpOverlay

```go
help := component.NewHelpOverlay([]component.HelpGroup{
    {Name: "Navigation", Entries: []component.HelpEntry{
        {Keys: "j/k", Description: "Move down/up"},
        {Keys: "g/G", Description: "Top/bottom"},
    }},
})
help.SetTitle(" Shortcuts ")
```

Searchable help shortcut cheatsheet with fuzzy filtering and keyboard navigation.

### Notification (Toast)

```go
toast := component.NewToastManager()
toast.Push("File saved", component.NotifySuccess)
toast.Push("Connection lost", component.NotifyError)
```

4 levels (info/success/warning/error) with auto-expiry and stacking.

## File & Content Widgets (Phase 15)

### FilePicker

```go
picker := component.NewFilePicker(".")
picker.SetFilter("*.go")
picker.OnSelect = func(e component.FileEntry) { fmt.Println("Selected:", e.Path) }
picker.OnConfirm = func(files []string) { fmt.Println("Confirmed:", files) }
```

File browser with fuzzy filter, multi-select (Space to toggle), vim-style navigation (j/k/h/l/g/G/), and arrow keys + Enter/Space/Esc. Directories listed first, alphabetical sorting.

Key bindings:
- `j`/`Down` — move cursor down
- `k`/`Up` — move cursor up
- `h`/`Left` — go to parent directory
- `l`/`Right`/`Enter` — enter directory or confirm file
- `Space` — toggle file selection
- `/` — enter filter mode
- `Esc` — exit filter mode / close
- `g` — jump to top
- `G` — jump to bottom

### StatusBar

```go
status := component.NewStatusBar()
status.AddLeft(component.StatusItem{ID: "mode", Text: "NORMAL"})
status.AddRight(component.StatusItem{ID: "pos", Text: "LN 42"})
status.AddCenter(component.StatusItem{ID: "file", Text: "main.go"})
```

Three-zone status bar (left/center/right) with separator. AI agent convenience methods:

```go
status.SetModel("gpt-4")
status.SetTokenRate(12500) // "12.5k tok/s"
status.SetContextWindow(8000, 128000) // used/total
status.SetClock()
```

### TabBar

```go
tabs := component.NewTabBar()
tabs.AddTab(component.Tab{Title: "Chat 1", Closable: true})
tabs.AddTab(component.Tab{Title: "Chat 2", Closable: true})
tabs.SetActive(0)
tabs.OnClose = func(index int) { fmt.Println("Closed:", index) }
```

Tab management with close buttons, keyboard navigation (Left/Right to switch, Ctrl+W to close), and hover highlighting.

### DiffPreview

```go
dp := component.NewDiffPreview()
dp.SetDiff(`--- a/main.go\n+++ b/main.go\n@@ -10,3 +10,4 @@\n func main() {\n-    fmt.Println("old")\n+    fmt.Println("new")\n+    // added\n }`)
fmt.Println(dp.Stats()) // {Additions:2, Deletions:1, ...}
dp.ScrollDown(3)
```

Unified diff viewer with syntax-aware highlighting (added/deleted/context/hunk/file/meta lines), scrolling, and statistics. Self-contained — no external dependencies.

Diff line types: `DiffContext`, `DiffAdd` (`+`), `DiffDel` (`-`), `DiffHunk` (`@@`), `DiffFile` (`+++`/`---`), `DiffMeta` (headers).

### Link

```go
text := "Visit https://fluui.dev for docs"
links := component.DetectLinks(text, 0) // row 0

mgr := component.NewLinkManager()
mgr.ScanText(text, 0)
mgr.AnnotateBuffer(buf, style)
// Click detection:
if link, ok := mgr.LinkAt(mx, my); ok {
    mgr.ClickLink(link)
}
```

URL detection with OSC8 hyperlink annotation. `DetectLinks()` finds URLs in text, `LinkManager` manages link ranges, annotation, and click handling.

## Composing Components

Use Flex to arrange children horizontally or vertically:

```go
root := layout.NewColumn()
root.Add(layout.NewRow().Add(border1, border2))
root.Add(border3)
root.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
root.Paint(buf)
```

