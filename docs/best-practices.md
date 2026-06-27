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

## Input Handling

### Route Through ChatApp

```go
base.OnKey(func(k *term.KeyEvent) {
    chat.HandleKey(k)  // handles input, scroll, search, theme cycling
    base.MarkDirty()
})
```

ChatApp.HandleKey handles:
- Enter (submit input)
- Ctrl+A/E/U/W (line editing)
- Up/Down (scroll / input history)
- Ctrl+F (search)
- Ctrl+T (theme cycling)
- Esc (close search/overlay)

### Always MarkDirty

Call `base.MarkDirty()` after handling events to trigger a redraw.

## Error Handling

### AI Errors

```go
chat.SetOnAIError(func(err error) {
    // Log, display, or retry
    chat.AddUserMessage("AI Error: " + err.Error())
})
```

### Graceful Shutdown

```go
base.OnQuit(func() {
    chat.StopStreaming()  // cancel in-flight requests
    // Save state if needed
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

### Test Blocks in Isolation

```go
func TestMyBlock(t *testing.T) {
    b := block.NewThinkingBlock("test")
    b.AppendDelta("reasoning")
    b.Complete()

    buf := buffer.New(80, 10)
    b.Paint(buf)
    // assert buffer contents
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
