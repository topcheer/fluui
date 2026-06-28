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

### Form & Dialog Components (Phase 18)

| Component | Description |
|---|---|
| **Dialog** | Modal dialog: Confirm, Info, Prompt, Custom types |
| **AutoComplete** | Popup fuzzy-filtered completion suggestions |
| **Wizard** | Multi-step wizard with lifecycle hooks |

### Advanced Widgets (Phase 19)

| Component | Description |
|---|---|
| **Checkbox** | Multi-item checkbox list with check/uncheck all |
| **RadioGroup** | Mutually exclusive single-selection group |
| **Slider** | Range slider with H/V orientation and vim keys |
| **CommandPalette** | Fuzzy-search command palette (Ctrl+P style) |
| **Spinner** | Animated loading spinner with frame styles |

### Dialog & Flow Components (Phase 18)

| Component | Description |
|---|---|
| **Dialog** | Modal dialog (Confirm/Info/Prompt/Custom) with callbacks |
| **AutoComplete** | Popup fuzzy autocomplete with OnSelect/OnDismiss |
| **Wizard** | Multi-step wizard with lifecycle hooks and dynamic buttons |

### Form & Input Components (Phase 19)

| Component | Description |
|---|---|
| **Checkbox** | Multi-item checkbox list with CheckAll/UncheckAll, j/k nav |
| **RadioGroup** | Mutually exclusive selection with vim-style navigation |
| **Slider** | Range slider (horizontal/vertical) with step, Home/End, h/l keys |
| **CommandPalette** | Fuzzy-search palette with highlighted matches, scroll, callbacks |
| **Spinner** | Animated spinner (dots/arc/line/bounce/bars) with label/prefix |

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

## Dialog (Phase 18)

```go
dialog := component.NewConfirmDialog("Delete File?", "Are you sure?")
dialog.OnConfirm = func() { os.Remove(path) }
dialog.Show()
// HandleKey: Enter=confirm, Esc=cancel, Tab/Left/Right=navigate buttons
```

Modal dialog with 4 types: `DialogConfirm` (Yes/No), `DialogInfo` (OK only), `DialogPrompt` (text input + OK/Cancel), and `DialogCustom` (arbitrary buttons). Supports `OnConfirm`/`OnCancel`/`OnCustom` callbacks. Prompt dialogs have full text editing (InsertRune, Backspace, Delete).

## AutoComplete (Phase 18)

```go
ac := component.NewAutoComplete()
ac.SetItems([]component.CompletionItem{
    {Label: "gpt-4", Detail: "OpenAI"},
    {Label: "claude-3", Detail: "Anthropic"},
})
ac.SetQuery("gp")
ac.Show(x, y)
// HandleKey: Up/Down navigate, Tab/Enter select, Esc dismiss
```

Popup autocomplete with fuzzy filtering, configurable `MaxVisible`, case-insensitive matching by default, and `OnSelect`/`OnDismiss` callbacks.

## Wizard (Phase 18)

```go
wizard := component.NewWizard([]*component.WizardStep{
    component.NewWizardStep("welcome", "Welcome").
        SetDescription("Let's set up your project"),
    component.NewWizardStep("config", "Configuration").
        SetDescription("Choose your settings"),
    component.NewWizardStep("done", "Complete").
        SetDescription("All done!"),
})
wizard.OnFinish = func(w *component.Wizard) { startApp() }
// HandleKey: Tab/Right=Next, Left=Back, Enter=select, Esc=cancel
```

Multi-step wizard with `OnEnter`/`OnLeave` lifecycle hooks per step (can block navigation by returning error), `OnFinish`/`OnCancel` callbacks, and dynamic button ordering based on step position.

## Checkbox (Phase 19)

```go
checkbox := component.NewCheckbox()
checkbox.AddItem("Enable streaming", false)
checkbox.AddItem("Show thinking", false)
checkbox.AddItem("Verbose mode", true)
// HandleKey: Space/Enter=toggle, j/Down=next, k/Up=prev
//             Ctrl+A=check all, Ctrl+D=uncheck all
checkbox.OnChange = func(idx int, checked bool) {}
```

Multi-item checkbox list with toggle (Space/Enter), check all (Ctrl+A), uncheck all (Ctrl+D), j/k navigation with wrap-around, disabled item skipping, and `OnChange` callback per item.

## RadioGroup (Phase 19)

```go
radio := component.NewRadioGroup()
radio.AddItem("GPT-4", "gpt-4")
radio.AddItem("Claude-3", "claude-3")
radio.AddItem("Gemini", "gemini")
radio.SetSelected("gpt-4")
// HandleKey: Up/Down/j/k navigate, Enter/Space=select
radio.OnChange = func(value string) { fmt.Println("Selected:", value) }
```

Mutually exclusive single-selection group. `SetDisabled(value)` disables an item and clears it if active. Vim-style j/k navigation with wrap-around, disabled item skipping, and `OnChange` callback.

## Slider (Phase 19)

```go
slider := component.NewSlider()
slider.SetMin(0).SetMax(100).SetValue(50)
slider.SetStep(5)
slider.Horizontal()
// HandleKey: Left/Right or h/l=step, Up/Down=large step
//             Home=min, End=max
slider.OnChange = func(value float64) { fmt.Printf("%.0f%%\n", value) }
```

Range slider with horizontal/vertical orientation, configurable min/max/step, `SetFromRatio(ratio)` for percentage-based control, Home/End shortcuts, h/j/k/l vim keys, and `OnChange` callback.

## CommandPalette (Phase 19)

```go
palette := component.NewCommandPalette()
palette.AddCommand(component.Command{
    ID: "settings.theme",
    Label: "Change Theme",
    Category: "Settings",
    Action: func() { openThemePicker() },
})
palette.AddCommand(component.Command{
    ID: "file.new",
    Label: "New File",
    Category: "File",
})
palette.Show(x, y)
// HandleKey: printable=filter, Up/Down/Tab=navigate (wrap)
//             Enter=execute, Esc=dismiss, Backspace=edit query
palette.OnExecute = func(cmd component.Command) {}
```

Fuzzy-search command palette with highlighted match segments (via `internal/fuzzy`), cursor navigation with wrap-around, scroll management with configurable `MaxVisible`, `OnExecute`/`OnDismiss` callbacks, and configurable styles per state (Normal/Matched/Selected/Prompt).

## Spinner (Phase 19)

```go
spinner := component.NewSpinner()
spinner.SetLabel("Loading...")
spinner.SetFrameStyle(animation.SpinnerDots)
spinner.Start()
// In render loop:
spinner.Update(dt) // advance frame
spinner.Paint(buf)
// When done:
spinner.Stop()
```

Animated spinner integrating `animation.Spinner` with frame styles (dots, arc, line, bounce, bars). `Start()`/`Stop()` control, `Update(dt)` advances frames and returns whether frame changed (for efficient redraws), configurable label and prefix, `SetFrameIndex(idx)` with wrap, and `SetFrameStyle(style)` for dynamic switching.

## Dialog (Phase 18)

```go
dialog := component.NewConfirmDialog("Delete File", "Are you sure?")
dialog.OnConfirm = func() { fmt.Println("Confirmed") }
dialog.Show()
// HandleKey: Esc=cancel, Enter=confirm, Tab/Left/Right=navigate buttons
```

Modal dialog with 4 types: Confirm (Yes/No), Info (OK only), Prompt (text input + OK/Cancel), and custom. Features button navigation with wrap, cursor highlighting, and optional text input field.

```go
prompt := component.NewPromptDialog("Username", "Enter your name:", "")
prompt.SetInputValue("default")
if prompt.Confirm() {
    name := prompt.InputValue()
}
```

## AutoComplete (Phase 18)

```go
ac := component.NewAutoComplete()
ac.SetItems([]component.CompletionItem{
    {Label: "golang", Detail: "Programming Language"},
    {Label: "python", Detail: "Programming Language"},
})
ac.SetQuery("go")
ac.Show(x, y)
// HandleKey: Up/Down=navigate, Tab/Enter=select, Esc=dismiss
```

Popup autocomplete with fuzzy-filtered candidates. Supports keyboard navigation (Up/Down/Tab/Enter/Esc), configurable max visible items, case sensitivity toggle, and `OnSelect`/`OnDismiss` callbacks.

## Wizard (Phase 18)

```go
wizard := component.NewWizard([]*component.WizardStep{
    component.NewWizardStep("welcome", "Welcome").
        SetDescription("Let's get started!"),
    component.NewWizardStep("config", "Configuration").
        SetDescription("Set up your preferences."),
    component.NewWizardStep("done", "Complete").
        SetDescription("All done!"),
})
wizard.SetOnFinish(func(w *component.Wizard) { fmt.Println("Finished!") })
```

Multi-step wizard with step navigation (Next/Back), skip support, lifecycle hooks (`OnEnter`/`OnLeave`), progress tracking, and dynamic button rendering. Keyboard: Tab/Left/Right=navigate buttons, Enter=activate, Esc=cancel, Ctrl+N=next, Ctrl+B=back.

## Checkbox (Phase 19)

```go
cb := component.NewCheckbox([]component.CheckboxItem{
    {Label: "Enable notifications", Checked: true},
    {Label: "Dark mode", Checked: false},
    {Label: "Auto-save", Checked: true},
})
cb.OnChange = func(items []component.CheckboxItem) { fmt.Println("Changed!") }
// HandleKey: Space/Enter=toggle, j/Down=next, k/Up=prev
// Ctrl+A=check all, Ctrl+D=uncheck all
```

Multi-select checkbox list with toggle (Space/Enter), bulk operations (Ctrl+A check all, Ctrl+D uncheck all), vim-style j/k navigation, disabled item skipping, and wrap-around cursor.

## RadioGroup (Phase 19)

```go
rg := component.NewRadioGroup([]component.RadioItem{
    {Label: "Light"},
    {Label: "Dark"},
    {Label: "Auto"},
})
rg.OnChange = func(item component.RadioItem) { fmt.Println("Selected:", item.Label) }
// HandleKey: Space/Enter=select, j/Down=next, k/Up=prev
```

Mutually exclusive radio button group. One selection at a time — selecting an item clears the previous. Supports j/k navigation, disabled item skipping, wrap-around cursor, and `OnChange` callback.

## Slider (Phase 19)

```go
slider := component.NewSlider()
slider.SetRange(0, 100)
slider.SetValue(50)
slider.SetStep(5)
slider.OnChange = func(val int) { fmt.Println("Value:", val) }
// HandleKey: Left/Right or h/l=step, Up/Down=fine step
// Home=min, End=max
```

Horizontal/vertical slider with configurable range, step size, and orientation. Keyboard: arrow keys for stepping, h/j/k/l vim keys, Home/End for min/max, ratio-based positioning via `SetFromRatio(float64)`.

## CommandPalette (Phase 19)

```go
palette := component.NewCommandPalette()
palette.SetCommands([]component.Command{
    {ID: "save", Label: "Save File", Category: "File"},
    {ID: "open", Label: "Open File", Category: "File"},
    {ID: "theme", Label: "Switch Theme", Category: "Settings"},
})
palette.OnExecute = func(cmd component.Command) { fmt.Println("Exec:", cmd.ID) }
palette.Show()
// HandleKey: Up/Down=navigate, Tab=next, Enter=execute, Esc=dismiss
```

Fuzzy-search command palette (VS Code style). Type to filter commands, navigate results with Up/Down/Tab, execute with Enter, dismiss with Esc. Features highlighted match segments, scroll management with configurable `MaxVisible`, multi-category commands, and `OnExecute`/`OnDismiss` callbacks.

## Spinner (Phase 19)

```go
spinner := component.NewSpinner()
spinner.SetLabel("Loading...")
spinner.SetFrameStyle(component.SpinnerFramesDots) // dots, arc, line, etc.
spinner.Start()
// Animation frames auto-advance via Update(dt)
spinner.Stop()
```

Animated spinner component integrating `animation.Spinner` with configurable frame styles (dots, arc, line, bounce, bars). Supports label and prefix rendering, start/stop control, manual frame index setting with wrap-around, and `Update(dt)` for efficient redraws.

## Composing Components

Use Flex to arrange children horizontally or vertically:

```go
root := layout.NewColumn()
root.Add(layout.NewRow().Add(border1, border2))
root.Add(border3)
root.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
root.Paint(buf)
```

