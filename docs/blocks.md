# Block System

Blocks are semantic content units designed for AI chat interfaces. Each block type represents a distinct piece of conversation content.

## Block Types

| Type | Purpose | Color Theme |
|---|---|---|
| **ThinkingBlock** | AI reasoning/chain-of-thought | Purple/gray, collapsible |
| **AssistantTextBlock** | AI response text (markdown) | Body text, rendered with goldmark |
| **ToolCallBlock** | Tool/function invocation | Pink accent |
| **ToolResultBlock** | Tool execution output | Purple border, collapsible |
| **UserMessageBlock** | User input | Cyan text |
| **ErrorBlock** | Error messages | Red border |

## Block Lifecycle

```
Created → Streaming → Complete
              ↘ Error
```

- `BlockStreaming` — actively receiving content
- `BlockComplete` — finished successfully
- `BlockError` — failed with error

## Creating Blocks

```go
// Direct construction
thinking := block.NewThinkingBlock("think-1")
text := block.NewAssistantTextBlock("asst-1")
call := block.NewToolCallBlock("tc-1", "read_file", `{"path":"/etc/hosts"}`)
result := block.NewToolResultBlock("tr-1")
userMsg := block.NewUserMessageBlock("user-1", "Hello!")
errBlock := block.NewErrorBlockWithMessage("err-1", "Permission denied")
```

Or via ChatApp helpers:

```go
thinking := chat.AddThinking()
text := chat.AddAssistantText()
call := chat.AddToolCall("read_file", `{"path":"/etc/hosts"}`)
result := chat.AddToolResult()
userMsg := chat.AddUserMessage("Hello!")
```

## Streaming Content

### Manual Streaming

```go
text := chat.AddAssistantText()
text.AppendDelta("Hello ")
text.AppendDelta("world!")
text.Complete()
```

### StreamDelta API

```go
delta := block.StreamDelta{
    Type:    block.DeltaContent,
    Content: "Hello world!",
}
chat.StreamDelta(delta)
```

Delta types:
- `DeltaThinking` — route to last ThinkingBlock
- `DeltaContent` — route to last AssistantTextBlock
- `DeltaToolCall` — route to last ToolCallBlock
- `DeltaToolResult` — route to last ToolResultBlock

### AI Bridge (automatic)

```go
chat.SetAIClient(client)
chat.SendUserMessage("Explain Go interfaces")
// Response auto-streams into appropriate blocks
```

The AIBridge routes:
- `reasoning_content` → ThinkingBlock
- `content` → AssistantTextBlock
- `tool_calls` → ToolCallBlock

## BlockContainer

Ordered collection managing layout, spacing, and virtual scrolling.

```go
container := block.NewBlockContainer()
container.SetSpacing(1) // 1-line gap between blocks
container.AddBlock(thinking)
container.AddBlock(text)

blocks := container.Blocks() // copy of all blocks
container.Len()              // block count
container.Clear()            // remove all
```

### Virtual Scrolling

For conversations with 1000+ blocks, the container supports `PaintVisible`:
- Binary search (O(log n)) to find visible blocks
- Only paints ~8 visible blocks out of 1000
- 31.5x faster than linear scan

## Collapsible Blocks

ThinkingBlock and ToolResultBlock support collapse/expand:

```go
thinking.Toggle()             // toggle collapse state
thinking.Collapsed()          // check current state
```

Click on a thinking or tool result block header to toggle via mouse.

## Diff Detection

ToolResultBlock automatically detects unified diff content and applies green/red syntax highlighting:

```go
result := chat.AddToolResult()
result.AppendDelta("--- old.txt\n+++ new.txt\n@@ -1,3 +1,3 @@\n-old line\n+new line\n context\n")
result.Complete()
// Lines with '+' render green, '-' render red
```

## Serialization

Save and restore conversations:

```go
registry := block.NewDefaultRegistry()

// Save
data, _ := block.SaveContainer(chat.Container(), registry)
os.WriteFile("chat.json", data, 0644)

// Load
container, _ := block.LoadContainer(data, registry)
for _, b := range container.Blocks() {
    chat.Container().AddBlock(b)
}
```

Format: Versioned JSON (`{"version":1,"blocks":[...]}`). Each block implements `Serializer`/`Deserializer` for type-specific state.

## Registry

The block registry maps type names to factory functions:

```go
registry := block.NewDefaultRegistry()

// Custom block registration
registry.Register("my-block", func(id string) block.Block {
    return NewMyBlock(id)
})
```

Used by `LoadContainer` to reconstruct blocks from JSON.
