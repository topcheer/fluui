# Tutorial: Build an AI Agent with Fluui

This tutorial walks through building a fully functional AI chat agent step by step.

## Prerequisites

- Go 1.26+
- An LLM API key (any OpenAI-compatible provider)

## Step 1: Create the Project

```bash
mkdir my-agent && cd my-agent
go mod init my-agent
go get github.com/topcheer/fluui
```

## Step 2: Basic Terminal Setup

Create `main.go`:

```go
package main

import (
    fluui "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/internal/buffer"
    "github.com/topcheer/fluui/internal/term"
)

func main() {
    base, err := fluui.New()
    if err != nil {
        panic(err)
    }
    defer base.Close()

    base.OnPaint(func(buf *buffer.Buffer) {
        buf.DrawText(2, 1, "My AI Agent", buffer.Style{
            Fg: buffer.RGB(0xFF, 0x79, 0xC6),
            Flags: buffer.Bold,
        })
    })

    base.Run()
}
```

Run: `go run main.go` — you should see a pink "My AI Agent" title.

## Step 3: Add ChatApp

Replace `main.go` with:

```go
package main

import (
    fluui "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/app"
    "github.com/topcheer/fluui/internal/buffer"
    "github.com/topcheer/fluui/internal/term"
)

func main() {
    base, _ := fluui.New()
    defer base.Close()

    w, h := base.Size()
    chat := app.NewChatApp(w, h)
    chat.SetInputHeight(2)

    // Add a welcome message
    chat.AddAssistantText().AppendDelta("Welcome! Type a message and press Enter.")

    // Handle input
    chat.OnSubmit(func(text string) {
        chat.AddUserMessage(text)
    })

    // Route events
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
    base.Run()
}
```

Now you can type messages. They appear as user message blocks.

## Step 4: Connect AI

Create `.env`:

```env
FLUUI_LLM_API_KEY=your-key
FLUUI_LLM_BASE_URL=https://open.bigmodel.cn/api/paas/v4
FLUUI_LLM_MODEL=glm-4-flash
```

Add AI integration:

```go
import (
    "github.com/topcheer/fluui/ai"
)

// After creating ChatApp:
cfg, _ := ai.LoadConfig()
client := ai.NewClient(cfg)
chat.SetAIClient(client)
chat.SetSystemPrompt("You are a helpful coding assistant.")
chat.SetOnAIError(func(err error) {
    chat.AddUserMessage("Error: " + err.Error())
})

// Replace OnSubmit with AI-powered version:
chat.OnSubmit(func(text string) {
    chat.SendUserMessage(text)  // sends to AI + streams response
})
```

Now typing a question streams the AI response into AssistantText blocks.

## Step 5: Add Keyboard Shortcuts

```go
base.OnKey(func(k *term.KeyEvent) {
    if chat.HandleKey(k) {
        base.MarkDirty()
        return
    }
    // Custom shortcuts
    switch {
    case k.Rune == 'q' && k.Modifiers&term.ModCtrl != 0:
        base.Quit()
    case k.Rune == 't' && k.Modifiers&term.ModCtrl != 0:
        chat.CycleTheme()
    case k.Rune == 'f' && k.Modifiers&term.ModCtrl != 0:
        // Ctrl+F triggers search (handled by ChatApp)
    }
    base.MarkDirty()
})
```

## Step 6: Save/Load Conversations

```go
import (
    "github.com/topcheer/fluui/block"
    "os"
)

func saveChat(chat *app.ChatApp) error {
    data, err := block.SaveContainer(chat.Container(), block.NewDefaultRegistry())
    if err != nil {
        return err
    }
    return os.WriteFile("conversation.json", data, 0644)
}

func loadChat(chat *app.ChatApp) error {
    data, err := os.ReadFile("conversation.json")
    if err != nil {
        return err
    }
    container, err := block.LoadContainer(data, block.NewDefaultRegistry())
    if err != nil {
        return err
    }
    chat.Clear()
    for _, b := range container.Blocks() {
        chat.Container().AddBlock(b)
    }
    return nil
}
```

## Step 7: Simulate Tool Calls

```go
// Simulate an AI tool call sequence
chat.AddToolCall("read_file", `{"path":"main.go"}`)
result := chat.AddToolResult()
result.AppendDelta("package main\n\nfunc main() { ... }\n")
result.Complete()
```

## Complete Example

See [`examples/chat/main.go`](../examples/chat/main.go) for the full working example.

## Key Takeaways

1. `fluui.New()` creates the base terminal app
2. `app.NewChatApp()` wraps it with chat-specific features
3. `chat.HandleKey()` / `chat.HandleMouse()` route events
4. `chat.Render()` paints to buffer
5. `chat.SetAIClient()` enables streaming AI responses
6. `block.SaveContainer()` / `LoadContainer()` persist conversations
