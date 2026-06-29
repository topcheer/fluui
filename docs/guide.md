# Fluui Developer Guide

A comprehensive guide to building terminal applications with Fluui — from core concepts to advanced patterns.

## Table of Contents

1. [Quick Start (5 Minutes)](#quick-start-5-minutes)
2. [Core Concepts](#core-concepts)
3. [The Component System](#the-component-system)
4. [Component Cookbook](#component-cookbook)
5. [Layout System](#layout-system)
6. [Event Handling](#event-handling)
7. [AI Integration](#ai-integration)
8. [Block System](#block-system)
9. [Theming](#theming)
10. [Testing](#testing)
11. [Debugging Guide](#debugging-guide)
12. [Performance Tips](#performance-tips)

---

## Quick Start (5 Minutes)

### Install

```bash
go get github.com/topcheer/fluui
```

Requires Go 1.26+.

### Minimal App

```go
package main

import (
    fluui "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/internal/buffer"
)

func main() {
    app, _ := fluui.New()
    defer app.Close()

    app.OnPaint(func(buf *buffer.Buffer) {
        buf.DrawText(2, 1, "Hello, Fluui!", buffer.Style{})
    })

    app.Run()
}
```

Press `Ctrl+C` to exit.

### Core Pattern

Every Fluui app follows the same pattern:

```
App → Loop → Events → Components → Buffer → Renderer → Terminal
```

1. Create an `App` with `fluui.New()`
2. Register handlers: `OnPaint`, `OnKey`, `OnMouse`
3. Components implement `Measure()`, `SetBounds()`, `Paint()`
4. The framework handles rendering and terminal I/O

---

## Core Concepts

### The Six-Layer Architecture

| Layer | Package | Responsibility |
|---|---|---|
| Application | `app`, `.` (root) | App lifecycle, event loop |
| Components | `component` | Reusable UI widgets |
| Blocks | `block` | AI content containers (streaming) |
| Events | `event`, `internal/term` | Key/mouse parsing and dispatch |
| Rendering | `render`, `internal/buffer` | Buffer diffing and terminal output |
| Markdown | `markdown` | Streaming markdown renderer |

### Key Types

```go
// Component — the core interface
type Component interface {
    ID() string
    Measure(Constraints) Size
    SetBounds(Rect)
    Paint(*buffer.Buffer)
    HandleKey(*term.KeyEvent) bool
    HandleMouse(*term.MouseEvent) bool
    Children() []Component
    IsDirty() bool
    ClearDirty()
}

// Rect — position and size
type Rect struct { X, Y, W, H int }

// Constraints — layout bounds
type Constraints struct {
    MaxWidth, MaxHeight int
    MinWidth, MinHeight int
}

// Style — text appearance
type Style struct {
    Fg    Color
    Bg    Color
    Flags StyleFlags  // Bold, Italic, Underline, Reverse, Dim
}
```

---

## The Component System

### Creating a Custom Component

```go
type MyWidget struct {
    component.BaseComponent
    text string
}

func NewMyWidget(text string) *MyWidget {
    w := &MyWidget{text: text}
    w.SetID(component.GenerateID("mywidget"))
    return w
}

// Measure returns the desired size
func (w *MyWidget) Measure(cs component.Constraints) component.Size {
    return component.Size{
        W: min(len(w.text), cs.MaxWidth),
        H: 1,
    }
}

// Paint renders into the buffer
func (w *MyWidget) Paint(buf *buffer.Buffer) {
    b := w.Bounds()
    buf.DrawText(b.X, b.Y, w.text, buffer.Style{})
    w.ClearDirty()
}
```

### BaseComponent

`BaseComponent` provides default implementations for `ID()`, `Bounds()`, `SetBounds()`, `IsDirty()`, `MarkDirty()`, `ClearDirty()`, and `SetID()`. Embed it to satisfy the Component interface with minimal boilerplate.

---

## Component Cookbook

### Text and Labels

```go
// Simple text — draw directly to buffer
buf.DrawText(x, y, "Hello", buffer.Style{Fg: buffer.RGB(255, 255, 255)})

// Styled text with bold
style := buffer.Style{
    Fg:    buffer.RGB(100, 200, 255),
    Flags: buffer.Bold,
}
buf.DrawText(x, y, "Bold Blue", style)
```

### Table

```go
table := component.NewTable(
    []string{"Name", "Age", "City"},
    []string{"Alice", "30", "NYC"},
    []string{"Bob", "25", "LA"},
)
table.SetSize(80, 10)
table.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
table.Paint(buf)

// Update rows dynamically
table.SetRows([][]string{
    {"Charlie", "35", "SF"},
    {"Dave", "40", "Chicago"},
})

// Handle keyboard navigation
table.HandleKey(keyEvent) // Up/Down arrows scroll
```

### Checkbox List

```go
cb := component.NewCheckbox([]string{"Option A", "Option B", "Option C"})
cb.SetSize(40, 5)
cb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
cb.Paint(buf)

// Navigation
cb.HandleKey(keyEvent) // j/k or Up/Down to move, Space to toggle

// Get checked items
checked := cb.CheckedItems()
```

### Radio Group

```go
rg := component.NewRadioGroup([]string{"Red", "Green", "Blue"})
rg.SetSize(30, 4)
rg.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})
rg.Paint(buf)

// Selection
rg.HandleKey(keyEvent) // j/k or Up/Down
selectedIndex := rg.SelectedIndex()  // -1 if none selected
```

### Slider

```go
slider := component.NewSliderWithRange(0, 100, 50, 1)
// or: slider := component.NewSlider() — defaults to 0-100, step 1
slider.SetSize(30, 1)
slider.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 1})
slider.Paint(buf)

// Keyboard
slider.HandleKey(keyEvent) // h/l or Left/Right to adjust
// Home → min, End → max
```

### Gauge

```go
gauge := component.NewGauge()
gauge.SetValue(75.0)  // 75% full
gauge.SetSize(40, 3)
gauge.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
gauge.Paint(buf)
```

### Sparkline

```go
sl := component.NewSparkline()
sl.SetData([]float64{1, 3, 2, 5, 4, 6, 8, 7})
sl.SetColorMode(component.ColorGradient)
sl.SetAutoScale(true)
sl.SetSize(40, 5)
sl.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
sl.Paint(buf)
```

### ProgressBar

```go
pb := component.NewProgressBar()
pb.SetProgress(0.65)  // 65% — NOTE: uses SetProgress, not SetValue
pb.SetSize(40, 1)
pb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})
pb.Paint(buf)
```

### Command Palette

```go
cp := component.NewCommandPalette()
cp.AddCommand(component.Command{
    ID:       "save",
    Label:    "Save File",
    Category: "File",
    Action:   func() { /* save logic */ },
})
cp.SetSize(60, 10)
cp.SetBounds(component.Rect{X: 10, Y: 5, W: 60, H: 10})
cp.Show(10, 5)  // Must call Show() before SetQuery/Paint
cp.SetQuery("sa")  // Filter commands
cp.Paint(buf)

// Keyboard navigation
cp.HandleKey(keyEvent)  // Up/Down to navigate, Enter to execute
```

### Spinner

```go
sp := component.NewSpinner("Loading...")
sp.SetSize(20, 1)
sp.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 1})
sp.Start()
sp.Paint(buf)
// ... later
sp.Stop()
```

### TabBar

```go
tb := component.NewTabBar("Overview", "Details", "Settings")
tb.SetSize(80, 1)
tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
tb.Paint(buf)

// Navigation: use NextTab() / PrevTab() — NO HandleKey method
tb.NextTab()
activeIdx := tb.ActiveIndex()
```

### FilePicker

```go
fp := component.NewFilePicker(".")
fp.SetSize(60, 20)
fp.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
fp.Paint(buf)

// Handles all keyboard navigation internally
fp.HandleKey(keyEvent)  // Up/Down navigate, Enter open, Backspace go up
```

### Tree

```go
root := component.NewTreeNode("root", "Project")
src := component.NewTreeNode("src", "src/")
root.AddChild(src)
src.AddChild(component.NewTreeNode("main", "main.go"))

tree := component.NewTree()
tree.SetRoot(root)
tree.SetSize(40, 20)
tree.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
tree.Paint(buf)

tree.HandleKey(keyEvent)  // Up/Down navigate, Enter expand/collapse
```

### Dialog

```go
// Dialog type constants: DialogInfo, DialogConfirm, DialogPrompt
dlg := component.NewDialog(component.DialogConfirm, "Confirm", "Delete this file?")
dlg.AddButton(component.NewDialogButton("OK", component.DialogResultOK))
dlg.AddButton(component.NewDialogButton("Cancel", component.DialogResultCancel))
dlg.SetSize(50, 7)
dlg.SetBounds(component.Rect{X: 15, Y: 10, W: 50, H: 7})
dlg.Paint(buf)

// Convenience constructors
confirmDlg := component.NewConfirmDialog("Confirm", "Delete?")
infoDlg := component.NewInfoDialog("Info", "Done")
promptDlg := component.NewPromptDialog("Input", "Name:", "default")
```

### StatusBar

```go
sb := component.NewStatusBar()
sb.AddLeft("status", "Ready")
sb.AddRight("pos", "Ln 42, Col 8")
sb.SetSize(80, 1)
sb.SetBounds(component.Rect{X: 0, Y: 23, W: 80, H: 1})
sb.Paint(buf)

// Update text dynamically
sb.SetItemText("status", "Saving...")
```

### Pagination

```go
pg := component.NewPagination()
pg.SetTotalItems(1000)
pg.SetItemsPerPage(20)
pg.SetSize(40, 1)
pg.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})
pg.Paint(buf)

// Navigation — NOTE: uses CurrentPage(), HasNext(), HasPrev()
pg.Next()
pg.Prev()
page := pg.CurrentPage()       // 0-indexed!
hasNext := pg.HasNext()
hasPrev := pg.HasPrev()
```

### VirtualScroller

```go
vs := component.NewVirtualScroller()
vs.SetVisible(0, 24)  // visible rows 0-24
vs.SetSize(80, 24)
vs.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

// Filtering
indices := vs.Filter(func(item component.VirtualItem) bool {
    return item.Text != ""
})
```

### Notification

```go
// Use ToastManager for notification display
nm := component.NewToastManager(3)  // max 3 visible
nm.SetSize(40, 3)
nm.SetBounds(component.Rect{X: 40, Y: 0, W: 40, H: 3})
nm.Paint(buf)
```

### InputLine (via ChatApp)

```go
chat := app.NewChatApp(80, 24)
il := chat.InputLine()
il.SetText("hello")
text := il.Text()
// Ctrl+Z = undo, Ctrl+Shift+Z = redo, Ctrl+Y = redo
```

---

## Layout System

Layouts are in `component/layout` — a separate sub-package.

### Flex (Row/Column)

```go
import "github.com/topcheer/fluui/component/layout"

// Horizontal layout
flex := layout.NewFlex(layout.FlexRow)
flex.AddChild(table)
flex.AddChild(sidebar)
flex.Measure(cs)
flex.SetBounds(rect)
flex.Paint(buf)

// Vertical layout
flex := layout.NewFlex(layout.FlexColumn)
```

### Stack (Overlay)

```go
stack := layout.NewStack()
stack.AddChild(background)
stack.AddChild(overlay)
// All children get the same bounds — useful for overlapping
```

### Center

```go
// NewCenter takes the child as an argument
center := layout.NewCenter(dialog)
// Centers the child within the bounds
```

### Padding

```go
pad := layout.NewPadding(2, 1, 2, 1)  // left, top, right, bottom
pad.SetChild(table)
```

---

## Event Handling

### Key Events

```go
app.OnKey(func(k *term.KeyEvent) {
    switch {
    case k.Key == term.KeyCtrlC:
        app.Quit()
    case k.Key == term.KeyEnter:
        submitForm()
    case k.Rune == 'q':
        app.Quit()
    }
    app.MarkDirty()
})
```

### Mouse Events

```go
app.OnMouse(func(m *term.MouseEvent) {
    component.HandleMouse(m)  // Delegate to component
    app.MarkDirty()
})
```

### Dispatcher and Key Shortcuts

```go
d := event.NewDispatcher()
d.BindKey(event.CtrlRune('s'), func() { save() })
d.BindKey(event.Alt(term.KeyEnter), func() { fullscreen() })
d.OnKey(func(k *term.KeyEvent) { /* default handler */ })
```

### Resize Events

```go
app.OnResize(func(w, h int) {
    chat.SetSize(w, h)
    app.MarkDirty()
})
```

---

## AI Integration

### ChatApp

```go
chat := app.NewChatApp(80, 24)
chat.SetAIClient(client)
chat.SetSystemPrompt("You are helpful.")

// Handle key routing
app.OnKey(func(k *term.KeyEvent) {
    chat.HandleKey(k)
    app.MarkDirty()
})

// Render
app.OnPaint(func(buf *buffer.Buffer) {
    chat.Render(buf)
})
```

### Streaming

```go
// ChatApp handles streaming internally via:
chat.AddUserMessage("Hello")
// AI response streams in automatically as AssistantTextBlock

// Manual streaming
block := chat.AddAssistantBlock()
chat.StreamDelta(block.StreamDelta{
    Type:    block.StreamDeltaContent,
    Content: "partial text",
})
```

### Themes

```go
// Built-in themes: Default, Dark, Light, Solarized, Gruvbox, Monokai
chat.SetThemeByName("gruvbox")

// Cycle themes: Ctrl+] forward, Ctrl+\ backward
chat.HandleKey(keyEvent)

// Programmatic
chat.SetThemeByIndex(2)
name := chat.ThemeName()
count := chat.ThemeCount()
```

---

## Block System

Blocks are containers for AI-generated content within the ChatApp.

### Block Types

| Type | Description |
|---|---|
| `AssistantTextBlock` | Streaming markdown from AI |
| `UserMessageBlock` | User input |
| `ToolCallBlock` | Tool/function calls |
| `WorkflowBlock` | Multi-step AI workflows |

### BlockContainer

```go
container := block.NewBlockContainer()
container.AddBlock(assistantBlock)
container.Paint(buf)  // Renders all blocks
container.ClearDirty()
```

---

## Theming

### Built-in Themes

```go
themes := theme.Builtins()
// Default, Dark, Light, Solarized Dark/Light, Gruvbox, Monokai, Dracula
```

### Custom Theme

```go
myTheme := theme.Theme{
    Name: "Custom",
    Bg:   buffer.RGB(30, 30, 40),
    Fg:   buffer.RGB(220, 220, 220),
    Accent:   buffer.RGB(100, 200, 255),
    Success:  buffer.RGB(100, 255, 100),
    Error:    buffer.RGB(255, 100, 100),
    Warning:  buffer.RGB(255, 200, 100),
}
```

### Markdown Theme

```go
mdTheme := markdown.DefaultTheme()
mdTheme.Heading1.Fg = buffer.RGB(255, 100, 100)
renderer := markdown.NewMarkdownRenderer(mdTheme, 80)
```

---

## Testing

### Component Tests

```go
func TestMyWidget(t *testing.T) {
    w := NewMyWidget("hello")
    w.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 1})

    buf := buffer.NewBuffer(20, 1)
    w.Paint(buf)

    cell := buf.GetCell(0, 0)
    assert.Equal(t, 'h', cell.Rune)
}
```

### Concurrent Tests

```go
func TestConcurrentPaint(t *testing.T) {
    c := NewCheckbox([]string{"A", "B", "C"})
    c.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 5})

    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            buf := buffer.NewBuffer(20, 5)
            c.Paint(buf)  // Must be safe under concurrent reads
        }()
    }
    wg.Wait()
}
```

### Integration Tests

```go
func TestFullPipeline(t *testing.T) {
    // Key event → component → buffer → render
    cp := component.NewCommandPalette()
    cp.AddCommand(component.Command{ID: "test", Label: "Test"})
    cp.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
    cp.Show(0, 0)  // Must Show before SetQuery/Paint
    cp.SetQuery("test")

    buf := buffer.NewBuffer(60, 10)
    cp.Paint(buf)
    assert.True(t, bufNonBlank(buf) > 0)
}
```

---

## Debugging Guide

### Common Issues

#### "My component renders blank"

Checklist:
1. Did you call `SetBounds()` before `Paint()`?
2. Is `bounds.W >= 3 && bounds.H >= 3`? (many components skip tiny bounds)
3. For `CommandPalette`: did you call `Show(x, y)` before `SetQuery()` / `Paint()`?
4. Did you call `MarkDirty()` after state changes?

```go
// Debug: check what's in the buffer
buf := buffer.NewBuffer(80, 24)
component.Paint(buf)
for y := 0; y < buf.Height; y++ {
    for x := 0; x < buf.Width; x++ {
        cell := buf.GetCell(x, y)
        if cell.Rune != ' ' {
            fmt.Printf("(%d,%d)=%c ", x, y, cell.Rune)
        }
    }
}
```

#### "Race detector fires"

All components use `sync.RWMutex` for thread safety. If you see race warnings:

1. Don't write to component fields without holding the write lock
2. `Paint()` uses `RLock()` — safe for concurrent reads
3. External mutations (from key handlers) must acquire the write lock
4. `MarkDirty()` uses atomic operations — safe to call without lock

#### "Terminal shows garbled output"

1. Fluui uses TrueColor (24-bit). Use a modern terminal (iTerm2, Alacritty, Kitty, Windows Terminal).
2. Over SSH, enable tmux passthrough for OSC52 clipboard.
3. Set `TERM=xterm-256color` or `xterm-direct`.

#### "Component doesn't respond to keys"

1. Are you routing keys via `component.HandleKey(k)`?
2. Did you call `app.MarkDirty()` after the key handler?
3. Check the key constant: `term.KeyUp` not `term.KeyArrowUp`.

#### "Layout doesn't size correctly"

1. Always call `Measure(constraints)` before `SetBounds(rect)`.
2. `Flex` distributes space among children — small children may get squeezed.
3. `Constraints.MaxWidth = 0` means "unbounded".

### Debug Output

```go
// Enable debug logging
app.SetDebug(true)

// Or manually log in OnPaint
app.OnPaint(func(buf *buffer.Buffer) {
    w, h := app.Size()
    log.Printf("paint: %dx%d, dirty=%v", w, h, app.IsDirty())
    // ... render
})
```

### Buffer Inspection

```go
// Dump buffer contents for debugging
func dumpBuffer(buf *buffer.Buffer) {
    for y := 0; y < buf.Height; y++ {
        var line strings.Builder
        for x := 0; x < buf.Width; x++ {
            cell := buf.GetCell(x, y)
            if cell.Rune == 0 || cell.Rune == ' ' {
                line.WriteByte('.')
            } else {
                line.WriteRune(cell.Rune)
            }
        }
        fmt.Printf("%2d: %s\n", y, line.String())
    }
}
```

---

## Performance Tips

### Do

- Cache rendered output when content hasn't changed (see `AssistantTextBlock.getCachedBlocks()`)
- Use `buffer.DiffInto()` for fast diffing (avoids allocation)
- Pre-allocate slices with known capacity
- Use ASCII rune cache for common characters

### Don't

- Call `renderer.Render()` inside `Paint()` on every frame
- Create new `Style{}` structs in hot paths — reuse them
- Use `fmt.Sprintf()` inside `Paint()` — pre-format strings
- Allocate in tight loops

### Benchmarking

```bash
# Quick check
go test -bench=. -benchmem -benchtime=1s ./render/

# Compare before/after
go test -bench=. -benchmem ./render/ > before.txt
# ... make changes ...
go test -bench=. -benchmem ./render/ > after.txt
benchstat before.txt after.txt

# CPU profiling
go test -bench=BenchmarkRender -cpuprofile=cpu.prof ./render/
go tool pprof -http=:8080 cpu.prof
```

See [docs/performance.md](performance.md) for full benchmark baselines and profiling guide.

---

## API Gotchas

These are common pitfalls discovered during development:

| Gotcha | Correct Usage |
|---|---|
| ProgressBar setter | `SetProgress(float64)` NOT `SetValue(int)` |
| VirtualItem field | `Text` NOT `Label` |
| Pagination navigation | `HasNext()` / `HasPrev()` NOT `HasNextPage()` |
| Pagination page | `CurrentPage()` NOT `Page()` (0-indexed!) |
| TabBar navigation | `NextTab()` / `PrevTab()` — NO `HandleKey()` |
| CommandPalette | Call `Show(x,y)` before `SetQuery()` / `Paint()` |
| Key constants | `term.KeyUp` NOT `term.KeyArrowUp` |
| Constraints | `component.Constraints` NOT `buffer.Constraints` |
| Layout constructors | `layout.NewFlex()` NOT `NewFlex()` (in `component/layout` package) |
| Buffer fill | `Fill(cell)` NOT `Clear()` — no Clear method exists |
| Markdown renderer | `NewMarkdownRenderer(theme, width)` takes 2 args |
| Render return | `([]*Block, error)` — must handle both values |

---

## Project Structure

```
fluui/
├── *.go                  # Root package (App, Loop)
├── ai/                   # AI client (OpenAI-compatible)
├── animation/            # Frame animations (Spinner)
├── app/                  # ChatApp, InputLine
├── block/                # Block system (streaming content)
├── component/            # 30+ UI components
│   └── layout/           # Flex, Stack, Center, Padding
├── event/                # Event types, Dispatcher
├── focus/                # Focus management
├── internal/buffer/      # Buffer, Cell, Style, Color
├── internal/fuzzy/       # Fuzzy matcher
├── internal/hotkey/      # Global hotkey support
├── internal/term/        # Terminal I/O, Parser, Writer
├── markdown/             # Markdown renderer (goldmark)
├── overlay/              # Overlay layer system
├── render/               # Renderer (buffer diff → terminal)
├── theme/                # Theme system
├── cmd/                  # 14 demos
├── examples/             # 9+ real-world examples
└── docs/                 # This documentation
```

---

## Next Steps

- [Architecture Overview](architecture.md)
- [API Reference](api-reference.md)
- [Best Practices](best-practices.md)
- [Performance Guide](performance.md)
- [Widget Reference](widgets-guide.md)
- [Block System](blocks.md)
- [Theming](themes.md)
- [Examples](../examples/)
