# Fluui Developer Guide

A comprehensive guide to building terminal applications with Fluui — from core concepts to advanced patterns.

## Table of Contents

1. [Project Structure](#project-structure)
2. [Core Concepts](#core-concepts)
3. [The Component System](#the-component-system)
4. [Block System for AI Content](#block-system-for-ai-content)
5. [Event Handling](#event-handling)
6. [Creating Custom Components](#creating-custom-components)
7. [Layout System](#layout-system)
8. [Theming](#theming)
9. [ChatApp Integration](#chatapp-integration)
10. [Testing Strategies](#testing-strategies)

---

## Project Structure

```
fluui/
├── fluui.go              # App entry point (terminal, renderer, event loop)
├── app/                  # ChatApp + InputLine + AI bridge + Clipboard + Search
├── buffer/               # Cell grid, styles, diff renderer
├── render/               # Painter, dirty tracking, efficient redraws
├── event/                # Event types, dispatcher, key shortcuts
├── component/            # 30+ UI components
│   └── layout/           # Flex layout (Row/Column/Stack/Center/Padding)
├── block/                # Semantic AI content blocks
├── overlay/              # Overlay manager, modal dialogs, popups
├── focus/                # Focus manager (Tab traversal)
├── hit/                  # Hit testing (regions, mouse clicks)
├── animation/            # Spinner, FadeIn, animation manager
├── markdown/             # goldmark AST renderer + chroma highlighter
├── theme/                # 10+ built-in themes
├── ai/                   # LLM client (OpenAI-compatible)
├── internal/
│   ├── term/             # Terminal raw mode, input/output, resize
│   ├── fuzzy/            # Fuzzy matching for CommandPalette
│   └── termcompat/       # Terminal capability detection
├── plugin/               # Plugin system
├── cmd/                  # 13 interactive demos
└── examples/             # 7 code examples
```

## Core Concepts

### The Six-Layer Architecture

Fluui is built bottom-up in six layers. Each layer depends only on layers below it:

```
Layer 6: Your App          (main.go)
Layer 5: ChatApp API       (app/)
Layer 4: Component System  (component/)
Layer 3: Content & Overlay (block/, overlay/, markdown/)
Layer 2: Rendering         (buffer/, render/)
Layer 1: Terminal          (internal/term/, event/)
```

### The Render Cycle

Fluui uses a double-buffer diff renderer for zero-flicker updates:

```
1. Event arrives (key, mouse, resize, AI delta)
2. State changes → MarkDirty()
3. On next frame:
   a. Paint to back buffer (off-screen)
   b. Diff back buffer vs front buffer
   c. Write only changed cells to terminal
   d. Swap buffers
```

This means you can call `MarkDirty()` freely from any event handler. The renderer batches updates and only writes what changed.

### The Buffer

The `buffer.Buffer` is a 2D grid of cells:

```go
buf := buffer.NewBuffer(80, 24)
buf.SetText(0, 0, "Hello, World!")
buf.SetStyle(0, 0, 13, buffer.Style{Fg: buffer.ColorBlue, Flags: buffer.Bold})
buf.FillRect(0, 0, 80, 1, ' ')  // clear first row
```

Key buffer operations:
- `SetText(x, y, text)` — write a string
- `SetStyle(x, y, width, style)` — apply styling to a range
- `FillRect(x, y, w, h, char)` — fill a rectangle
- `Clear()` — clear entire buffer

## The Component System

### Component Interface

Every widget implements the `Component` interface:

```go
type Component interface {
    Measure(Constraints) Size    // How big do I want to be?
    SetBounds(Rect)               // Here's your actual position
    Paint(*buffer.Buffer)         // Draw yourself
    Children() []Component        // Return children (or nil)
}
```

### Measure/Paint Layout Model

Fluui uses a two-pass layout model (like Flutter):

1. **Measure pass**: Parent asks each child "how big do you want to be?" given constraints (min/max width/height)
2. **Layout pass**: Parent assigns bounds (X, Y, W, H) to each child
3. **Paint pass**: Each component paints itself within its bounds

```go
func (c *MyComponent) Measure(cs component.Constraints) component.Size {
    // Given max width/height, calculate desired size
    return component.Size{W: cs.MaxWidth, H: 1}
}

func (c *MyComponent) SetBounds(r component.Rect) {
    c.bounds = r
}

func (c *MyComponent) Paint(buf *buffer.Buffer) {
    buf.SetText(c.bounds.X, c.bounds.Y, "Hello")
}
```

### BaseComponent

Most components embed `BaseComponent` which provides:
- `SetBounds` / `Bounds` / `Size`
- `MarkDirty` / `IsDirty` / `ClearDirty`
- `SetVisible` / `Visible`
- `SetEnabled` / `Enabled`
- `SetFocusable` / `Focusable`

### Built-in Components (30+)

| Category | Components |
|---|---|
| **Basic** | Text, Border, ScrollView, Gauge, Sparkline, Badge, ProgressBar |
| **Layout** | Flex (Row/Column), Stack, Center, Padding, Gap |
| **Data** | Table, Tree, Form, Pagination, VirtualScroller |
| **File & Nav** | FilePicker, TabBar, StatusBar, DiffPreview |
| **Overlay** | Dialog, AutoComplete, Wizard, ContextMenu, Tooltip |
| **Form** | Checkbox, RadioGroup, Slider, TextArea |
| **Feedback** | CommandPalette, Spinner, Notification, HelpOverlay |
| **Interactive** | SplitPane, Selection, LinkManager |

## Block System for AI Content

Fluui uses semantic blocks for AI-generated content instead of raw text:

```go
// Thinking block (collapsible)
thinking := chat.AddThinking()
thinking.AppendDelta("Let me analyze the request...")

// Tool call block (shows command + result)
toolCall := chat.AddToolCall("grep", `{"pattern": "TODO"}`)
toolCall.SetResult("main.go:42:TODO: fix this\napp.go:15:TODO: add tests")

// Assistant text (streaming)
assistant := chat.AddAssistantText()
assistant.AppendDelta("I found 2 TODO items.\n")

// User message
chat.AddUserMessage("What TODOs exist?")

// Error block
chat.AddError(errors.New("API timeout"))
```

### Streaming Pattern

```go
// AI streams tokens → append to block
block := chat.AddAssistantText()
for token := range tokenStream {
    block.AppendDelta(token)
    chat.MarkDirty() // trigger redraw
}
block.Complete()
```

### Serialization

Blocks can be saved and restored:

```go
// Save conversation
data := block.SerializeContainer(container)
os.WriteFile("chat.json", data, 0644)

// Restore conversation
data, _ := os.ReadFile("chat.json")
container := block.DeserializeContainer(data)
```

## Event Handling

### Event Types

```go
TypeKey     // Keyboard input
TypeMouse   // Mouse click/scroll/move
TypePaste   // Clipboard paste
TypeResize  // Terminal resize
TypeQuit    // Shutdown signal
```

### Key Event Handling

```go
// In ChatApp or custom component
func (c *MyComponent) HandleKey(key *term.KeyEvent) bool {
    // Printable character
    if key.Rune != 0 && key.Key == term.KeyUnknown {
        c.insertChar(key.Rune)
        return true
    }

    // Ctrl combinations
    if key.Modifiers&term.ModCtrl != 0 {
        switch key.Key {
        case term.KeyCtrlC: c.quit(); return true
        case term.KeyCtrlZ: c.undo(); return true
        }
    }

    // Special keys
    switch key.Key {
    case term.KeyEnter:  c.submit(); return true
    case term.KeyTab:    c.nextField(); return true
    case term.KeyBacktab: c.prevField(); return true
    }

    return false // event not handled
}
```

### Mouse Event Handling

```go
func (c *MyComponent) HandleMouse(m *term.MouseEvent) bool {
    switch m.Type {
    case term.MouseClick:
        if c.contains(m.X, m.Y) {
            c.handleClick(m.X, m.Y)
            return true
        }
    case term.MouseScrollUp:
        c.scrollUp()
        return true
    case term.MouseScrollDown:
        c.scrollDown()
        return true
    }
    return false
}
```

### Dispatcher and Key Bindings

```go
// Bind specific keys to handlers
disp := event.NewDispatcher()
disp.BindKey(
    event.NewKeyShortcut(term.KeyCtrlP, 0), // Ctrl+P
    func(e event.Event) bool {
        palette.Show()
        return true
    },
)
disp.OnKey(func(e event.Event) bool {
    // Default key handler
    return false
})
```

## Creating Custom Components

### Simple Component

```go
type Label struct {
    component.BaseComponent
    mu    sync.RWMutex
    text  string
    style buffer.Style
}

func NewLabel(text string) *Label {
    l := &Label{text: text}
    l.SetFocusable(false)
    return l
}

func (l *Label) Measure(cs component.Constraints) component.Size {
    l.mu.RLock()
    defer l.mu.RUnlock()
    w := len([]rune(l.text))
    if cs.MaxWidth > 0 && w > cs.MaxWidth {
        w = cs.MaxWidth
    }
    return component.Size{W: w, H: 1}
}

func (l *Label) Paint(buf *buffer.Buffer) {
    l.mu.RLock()
    defer l.mu.RUnlock()
    b := l.Bounds()
    buf.SetText(b.X, b.Y, l.text)
    buf.SetStyle(b.X, b.Y, len([]rune(l.text)), l.style)
}

func (l *Label) SetText(text string) {
    l.mu.Lock()
    l.text = text
    l.mu.Unlock()
    l.MarkDirty()
}

func (l *Label) Children() []component.Component { return nil }
```

### Interactive Component with Keyboard

```go
type Button struct {
    component.BaseComponent
    mu       sync.RWMutex
    label    string
    focused  bool
    OnClick  func()
}

func (b *Button) HandleKey(key *term.KeyEvent) bool {
    if key.Key == term.KeyEnter || key.Key == term.KeySpace {
        b.mu.RLock()
        fn := b.OnClick
        b.mu.RUnlock()
        if fn != nil {
            fn()
        }
        return true
    }
    return false
}

func (b *Button) Paint(buf *buffer.Buffer) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    bounds := b.Bounds()
    style := buffer.Style{}
    if b.focused {
        style.Flags |= buffer.Reverse
    }
    buf.SetText(bounds.X, bounds.Y, "["+b.label+"]")
    buf.SetStyle(bounds.X, bounds.Y, bounds.W, style)
}
```

### Composing Components with Flex

```go
func buildUI() *component.Flex {
    // Vertical layout
    col := component.NewColumn()

    // Header row
    header := component.NewRow()
    header.AddChild(component.NewLabel("Fluui Chat"))
    header.AddChild(component.NewSpacer())
    header.AddChild(component.NewLabel("[Ctrl+Q] Quit"))

    // Content area
    content := component.NewScrollView()

    // Input row
    input := component.NewLabel("> Type a message...")

    col.AddChild(header)
    col.AddChild(content)
    col.AddChild(input)

    return col
}
```

## Layout System

### Flex Layout

```go
// Horizontal (row) layout
row := component.NewRow()
row.SetGap(1)              // 1-space gap between children
row.AddChild(child1)
row.AddChild(child2)

// Vertical (column) layout
col := component.NewColumn()
col.SetGap(0)
col.AddChild(child1)
col.AddChild(child2)

// Stack (overlapping layers)
stack := component.NewStack()
stack.AddChild(background)
stack.AddChild(overlay)

// Center a child
center := component.NewCenter()
center.AddChild(myWidget)

// Padding
padded := component.NewPadding(2, 1) // horizontal=2, vertical=1
padded.AddChild(content)
```

### Constraints

```go
// Unbounded (no limits)
cs := component.Unbounded()

// Fixed width/height
cs := component.Bounded(80, 24)

// Use in Measure()
func (c *MyComp) Measure(cs component.Constraints) component.Size {
    return component.Size{
        W: min(c.desiredWidth, cs.MaxWidth),
        H: min(c.desiredHeight, cs.MaxHeight),
    }
}
```

## Theming

### Built-in Themes

Fluui ships with 10+ built-in themes accessible via keyboard shortcuts:

```
Ctrl+T     Cycle theme forward
Ctrl+Shift+T  Cycle theme backward
Ctrl+]     Cycle theme forward (P19 addition)
Ctrl+\     Cycle theme backward (P19 addition)
```

### Programmatic Theme Control

```go
// Get current theme
t := chat.Theme()
fmt.Println("Current:", chat.ThemeName())

// List all themes
names := chat.ThemeList()  // []string

// Switch theme
chat.SetThemeByName("Dracula")
chat.SetThemeByIndex(3)

// Custom theme
myTheme := theme.New("Custom")
myTheme.SetColor(theme.ColorBg, 0x1a1b26)
myTheme.SetColor(theme.ColorFg, 0xa9b1d6)
chat.SetTheme(myTheme)
```

### Theme Structure

Themes define colors for: Background, Foreground, Accent, Border, Selection, Error, Warning, Success, Muted, Link, plus Markdown syntax colors (headings, code, links, etc.).

## ChatApp Integration

### Basic ChatApp

```go
func main() {
    app := fluui.New()
    chat := app.NewChatApp(80, 24)

    app.OnKey(func(k *term.KeyEvent) {
        chat.HandleKey(k)
        app.MarkDirty()
    })

    app.OnPaint(func(buf *buffer.Buffer) {
        w, h := app.Size()
        chat.SetSize(w, h)
        chat.Render(buf)
    })

    app.Run()
}
```

### P19-P20 Features

```go
// Undo/Redo (P19)
// Ctrl+Z = undo, Ctrl+Shift+Z / Ctrl+Y = redo
// Max 100 undo states, redo cleared on new edit

// Theme cycling (P19)
// Ctrl+] = forward, Ctrl+\ = backward

// CommandPalette (P20)
// Ctrl+P = toggle command palette
chat.AddCommand(component.Command{
    ID:       "settings.theme",
    Label:    "Change Theme",
    Category: "Settings",
    Action:   func() { /* ... */ },
})

// Spinner (P20)
chat.StartSpinner("Loading models...")
// ... async work ...
chat.StopSpinner()
```

### AI Integration

```go
cfg, _ := ai.LoadConfig()
client := ai.NewClient(cfg)
chat.SetAIClient(client)
chat.SetSystemPrompt("You are a helpful assistant.")

chat.OnSubmit(func(text string) {
    chat.SendUserMessage(text) // streams AI response
})
```

## Testing Strategies

### Component Tests

```go
func TestMyComponent_Paint(t *testing.T) {
    c := NewMyComponent("test")
    c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

    buf := buffer.NewBuffer(80, 1)
    c.Paint(buf)

    // Verify output
    cell := buf.Get(0, 0)
    assert.Equal(t, 't', cell.Rune)
}
```

### Measure Tests

```go
func TestLabel_Measure(t *testing.T) {
    l := NewLabel("Hello")
    size := l.Measure(component.Unbounded())
    assert.Equal(t, 5, size.W)
    assert.Equal(t, 1, size.H)
}
```

### HandleKey Tests

```go
func TestButton_Enter(t *testing.T) {
    var clicked bool
    btn := NewButton("OK")
    btn.OnClick = func() { clicked = true }

    handled := btn.HandleKey(&term.KeyEvent{
        Key: term.KeyEnter,
    })

    assert.True(t, handled)
    assert.True(t, clicked)
}
```

### Concurrent Access Tests

```go
func TestCheckbox_Concurrent(t *testing.T) {
    cb := NewCheckbox([]string{"a", "b", "c"})

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(2)
        go func() { defer wg.Done(); cb.Toggle() }()
        go func() { defer wg.Done(); cb.Items() }()
    }
    wg.Wait()
}
```

### ChatApp Integration Tests

```go
func TestChatApp_Theme(t *testing.T) {
    chat := app.NewChatApp(80, 24)
    assert.True(t, chat.ThemeCount() > 0)

    chat.SetThemeByIndex(0)
    assert.Equal(t, 0, chat.ThemeIndex())

    chat.CycleTheme()
    assert.Equal(t, 1, chat.ThemeIndex())
}
```

### Running Tests

```bash
# All tests with race detector
go test -race -short ./...

# Specific package
go test -race ./component/

# Verbose
go test -race -v -run TestCheckbox ./component/

# Coverage
go test -cover -short ./component/

# Benchmarks
go test -bench=. -benchmem ./render/
```

---

## Next Steps

- [Getting Started](getting-started.md) — Quick 5-minute setup
- [Tutorial](tutorial.md) — Build an AI agent step by step
- [Components](components.md) — Full component reference
- [Widgets Guide](widgets-guide.md) — Phase 15-19 widget tutorials
- [API Reference](api-reference.md) — Complete API tables
- [Best Practices](best-practices.md) — Patterns and anti-patterns
- [Architecture](architecture.md) — Layer overview and design principles
- [CHANGELOG](CHANGELOG.md) — Phase 1-20 change history
