# Best Practices

## Concurrency

### Single Goroutine for State

The event loop owns all state. Mutate ChatApp state only from event callbacks (`OnKey`, `OnMouse`, `OnPaint`).

### Streaming from Goroutines

AI streaming runs in a background goroutine. Use channels or the built-in AIBridge:

```go
// Good: AIBridge handles this automatically
chat.SetAIClient(client)
chat.SendUserMessage("Hello")

// Manual streaming: use chat.AddXxx() which is mutex-protected
chat.AddAssistantText().AppendDelta("...")
```

### Mutex Protection

ChatApp, BlockContainer, and all blocks use `sync.Mutex` or `sync.RWMutex`. Never share block pointers across goroutines without locking.

All 30+ components (Checkbox, RadioGroup, Slider, CommandPalette, Spinner, etc.) use `sync.RWMutex` for thread-safe access:

```go
// Safe: component methods handle locking internally
checkbox.Toggle()
items := checkbox.Items()  // returns defensive copy

// Safe: concurrent access is tested with -race
go func() { checkbox.MoveDown() }()
go func() { _ = checkbox.Items() }()
```

## Performance

### Use Virtual Scrolling

For 100+ blocks, virtual scrolling only paints visible blocks:

```go
container := block.NewBlockContainer()
// PaintVisible is automatically used by ScrollView
```

### Batch Delta Appends

For large content, batch deltas to reduce render cycles:

```go
// Good: append in chunks
for _, line := range lines {
    text.AppendDelta(line + "\n")
}

// Better: accumulate then append once
var sb strings.Builder
for _, line := range lines {
    sb.WriteString(line + "\n")
}
text.AppendDelta(sb.String())
```

### Let the Diff Renderer Work

Don't call `buf.Clear()` every frame. The double-buffer diff renderer only writes changed cells. Drawing the same content twice is a no-op.

### Spinner Frame Efficiency

Use `Update(dt)` return value to skip unnecessary redraws:

```go
// Good: only paint when frame actually changed
if spinner.Update(dt) {
    spinner.Paint(buf)
    markDirty()
}
```

### Avoid Redundant Measure Calls

`Measure()` can be expensive for complex component trees. Cache the result and only re-measure when constraints change:

```go
// Bad: measuring every frame
func (c *MyComp) Paint(buf *buffer.Buffer) {
    size := c.Measure(component.Unbounded()) // expensive!
    // ...
}

// Good: measure once, use cached size
func (c *MyComp) Layout() {
    c.cachedSize = c.Measure(component.Unbounded())
}
```

## Layout

### Use ChatApp Defaults

ChatApp handles padding, scroll view sizing, and input line layout automatically. Use `chat.Render(buf)` instead of manual layout.

### SetSize on Resize

Always update ChatApp dimensions in OnPaint:

```go
app.OnPaint(func(buf *buffer.Buffer) {
    w, h := app.Size()
    chat.SetSize(w, h)
    chat.Render(buf)
})
```

### Flex Layout Best Practices

Prefer Flex layout for arranging children — it handles measurement, distribution, and bounds automatically:

```go
layout := component.NewFlex()
layout.Direction = component.FlexColumn
layout.AddChild(header)    // fixed height
layout.AddChild(scrollView) // flexible
layout.AddChild(statusBar)  // fixed height
```

## Component Design

### Building Custom Components

Always embed `BaseComponent` and implement all four interface methods:

```go
type Widget struct {
    component.BaseComponent
    mu sync.RWMutex
    // ... fields
}

func (w *Widget) Measure(cs component.Constraints) component.Size {
    // Calculate desired size from content, bounded by cs.MaxWidth/MaxHeight
}

func (w *Widget) SetBounds(r component.Rect) {
    w.BaseComponent.SetBounds(r)
}

func (w *Widget) Paint(buf *buffer.Buffer) {
    // Draw at w.Bounds().X, w.Bounds().Y
}

func (w *Widget) Children() []component.Component { return nil }
```

### Thread Safety in Components

Use `sync.RWMutex` — read lock for queries, write lock for mutations:

```go
func (w *Widget) SetValue(v int) {
    w.mu.Lock()
    defer w.mu.Unlock()
    w.value = v
    w.MarkDirty()
}

func (w *Widget) Value() int {
    w.mu.RLock()
    defer w.mu.RUnlock()
    return w.value
}
```

### Callback Patterns

Components should expose callbacks, not control flow:

```go
// Good: component exposes callback, caller decides what to do
type Slider struct {
    // ...
    OnChange func(value int)
}

// Caller wires it up
slider.OnChange = func(v int) {
    statusbar.SetText(fmt.Sprintf("Volume: %d%%", v))
    markDirty()
}
```

### Defensive Copies

Return copies of internal slices to prevent external mutation:

```go
func (c *Checkbox) Items() []CheckboxItem {
    c.mu.RLock()
    defer c.mu.RUnlock()
    items := make([]CheckboxItem, len(c.items))
    copy(items, c.items)
    return items
}
```

### Navigation with Wrap-Around

Follow the established vim-style navigation pattern:

```go
func (c *Component) MoveDown() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cursor++
    if c.cursor >= c.itemCount {
        c.cursor = 0 // wrap to top
    }
    // skip disabled items
    for c.isDisabled(c.cursor) {
        c.cursor = (c.cursor + 1) % c.itemCount
    }
}
```

## Input Handling

### Route Through ChatApp

```go
base.OnKey(func(k *term.KeyEvent) {
    chat.HandleKey(k)  // handles input, scroll, search, theme, undo/redo
    base.MarkDirty()
})
```

ChatApp.HandleKey handles:
- Enter (submit input)
- Ctrl+A/E/U/W (line editing)
- Ctrl+Z (undo), Ctrl+Y/Ctrl+Shift+Z (redo)
- Up/Down (scroll / input history)
- Ctrl+F (search)
- Ctrl+T / Ctrl+] (theme forward), Ctrl+Shift+T / Ctrl+\ (theme backward)
- Ctrl+P (toggle command palette)
- Esc (close search/overlay)

### Always MarkDirty

Call `base.MarkDirty()` after handling events to trigger a redraw.

### Component Key Handling

Components return `true` when they consume a key event:

```go
func (c *Slider) HandleKey(k *term.KeyEvent) bool {
    switch {
    case k.Key == term.KeyLeft || (k.Rune == 'h' && k.Key == term.KeyUnknown):
        c.Decrement()
        return true
    case k.Key == term.KeyRight || (k.Rune == 'l' && k.Key == term.KeyUnknown):
        c.Increment()
        return true
    case k.Key == term.KeyHome:
        c.SetValue(c.Min())
        return true
    }
    return false
}
```

## Error Handling

### AI Errors

```go
chat.SetOnAIError(func(err error) {
    chat.AddUserMessage("AI Error: " + err.Error())
})
```

### Graceful Shutdown

```go
base.OnQuit(func() {
    chat.StopStreaming()  // cancel in-flight requests
})
```

### Ctrl+C Handling

```go
base.OnInterrupt(func() bool {
    if chat.IsStreaming() {
        chat.StopStreaming()
        return false  // don't quit, just stop streaming
    }
    return true  // quit
})
```

## Serialization

### Save on Quit

```go
base.OnQuit(func() {
    data, _ := block.SaveContainer(chat.Container(), block.NewDefaultRegistry())
    os.WriteFile("conversation.json", data, 0644)
})
```

### Restore on Start

```go
data, _ := os.ReadFile("conversation.json")
container, _ := block.LoadContainer(data, block.NewDefaultRegistry())
for _, b := range container.Blocks() {
    chat.Container().AddBlock(b)
}
```

## Terminal Compatibility

### OSC52 Clipboard

Fluui auto-detects terminal capabilities. Clipboard works in:
- iTerm2, Alacritty, Kitty, Ghostty (native)
- tmux/screen (passthrough wrapping)
- Unknown terminals (disabled to avoid garbage)

### TrueColor

All themes use 24-bit TrueColor. Falls back gracefully on older terminals.

## Testing

### Run with Race Detector

```bash
go test ./... -race -count=1
```

### Test Components in Isolation

```go
func TestSlider(t *testing.T) {
    s := component.NewSlider()
    s.SetRange(0, 100)
    s.SetValue(50)

    s.Measure(component.Unbounded())
    s.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

    buf := buffer.New(80, 1)
    s.Paint(buf)
    // assert buffer contents
}
```

### Concurrent Component Tests

All components should have concurrent test coverage:

```go
func TestCheckbox_Concurrent(t *testing.T) {
    cb := component.NewCheckbox([]string{"A", "B", "C"})
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(3)
        go func() { defer wg.Done(); cb.Toggle() }()
        go func() { defer wg.Done(); _ = cb.Items() }()
        go func() { defer wg.Done(); cb.MoveDown() }()
    }
    wg.Wait()
}
```

## Common Patterns

### Chat Loop Pattern

```go
// Standard event routing
base.OnKey(func(k *term.KeyEvent) {
    chat.HandleKey(k)
    base.MarkDirty()
})
base.OnMouse(func(m *term.MouseEvent) {
    chat.HandleMouse(m)
    base.MarkDirty()
})
base.OnPaint(func(buf *buffer.Buffer) {
    w, h := base.Size()
    chat.SetSize(w, h)
    chat.Render(buf)
})
chat.OnQuit(func() { base.Quit() })
```

### Block-First Content

Always use semantic blocks instead of raw text:
```go
// Good: semantic blocks
chat.AddThinking().AppendDelta("Let me analyze...")
chat.AddToolCall("grep", `{"pattern":"TODO"}`)

// Avoid: raw text
chat.AddUserMessage("Thinking: Let me analyze...")
```

### CommandPalette Integration

Register commands and let users discover them via Ctrl+P:

```go
palette := chat.CommandPalette()
palette.AddCommand(component.Command{
    ID: "settings.theme",
    Label: "Change Theme",
    Category: "Settings",
    Action: func() { openThemePicker() },
})
```

### Spinner for Async Operations

Show loading state during network calls:

```go
chat.StartSpinner("Fetching models...")
go func() {
    models := fetchModels()
    chat.StopSpinner()
    // update UI with models
}()
```

### Undo/Redo Safety

InputLine undo/redo is automatic. Don't interfere with the undo stack:

```go
// Good: saveUndo() is called automatically before mutations
// by HandleKey. Just use the public API.
il.HandleKey(keyEvent) // undo state saved internally

// Don't: manually clear undo history unless user requests it
// il.ClearUndoHistory()  // only on explicit user action
```
