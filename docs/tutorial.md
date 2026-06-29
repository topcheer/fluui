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

## Step 8: Add P19-P26 Features

### Theme Cycling (P19)

```go
// Ctrl+T or Ctrl+] = cycle forward
// Ctrl+Shift+T or Ctrl+\ = cycle backward
// Programmatic:
chat.CycleTheme()
chat.SetThemeByName("Dracula")
```

### Undo/Redo (P19-P20)

```go
// InputLine supports undo/redo automatically
// Ctrl+Z = undo, Ctrl+Y / Ctrl+Shift+Z = redo
// Max 100 undo states, redo cleared on new edit
```

### Command Palette (P20)

```go
// Ctrl+P opens the command palette
chat.AddCommand(component.Command{
    ID:       "file.open",
    Label:    "Open File",
    Category: "File",
    Action:   func() { /* ... */ },
})
```

### Spinner (P20)

```go
chat.StartSpinner("Thinking...")
// ... async work ...
chat.StopSpinner()
```

### Dialog (P18)

```go
dialog := component.NewConfirmDialog(
    "Delete File",
    "Are you sure you want to delete main.go?",
)
dialog.OnConfirm = func(text string) bool {
    os.Remove("main.go")
    return true // close dialog
}
overlay.Show(dialog)
```

### Wizard (P18)

```go
wizard := component.NewWizard([]*component.WizardStep{
    {
        ID:    "welcome",
        Title: "Welcome",
        Description: "Let's set up your project",
    },
    {
        ID:    "config",
        Title: "Configuration",
        OnLeave: func(w *component.Wizard) error {
            return validateConfig() // block navigation if invalid
        },
    },
})
wizard.SetOnFinish(func() { /* done */ })
```

## Complete Example

See [`examples/chat/main.go`](../examples/chat/main.go) for the full working example.
Also see [`examples/ai-agent/`](../examples/ai-agent/) for a production AI agent demo.

## Key Takeaways

1. `fluui.New()` creates the base terminal app
2. `app.NewChatApp()` wraps it with chat-specific features
3. `chat.HandleKey()` / `chat.HandleMouse()` route events
4. `chat.Render()` paints to buffer
5. `chat.SetAIClient()` enables streaming AI responses
6. `block.SaveContainer()` / `LoadContainer()` persist conversations
7. **Ctrl+P** opens the command palette for discoverability
8. **Ctrl+T** cycles themes; all components auto-adapt
9. Use **Dialog/Wizard** for modal interactions
10. Test with `-race` flag for concurrent safety
