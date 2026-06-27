# Getting Started with Fluui

Fluui is an AI-native terminal UI library for Go. This guide gets you building in 5 minutes.

## Installation

```bash
go get github.com/topcheer/fluui
```

Requires Go 1.26+.

## Minimal Example

Create `main.go`:

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

Run it:

```bash
go run main.go
```

Press `Ctrl+C` to exit.

## Connect an AI Backend

Fluui includes an OpenAI-compatible streaming client. Create a `.env` file:

```env
FLUUI_LLM_API_KEY=your-api-key
FLUUI_LLM_BASE_URL=https://open.bigmodel.cn/api/paas/v4
FLUUI_LLM_MODEL=glm-4-flash
FLUUI_LLM_SYSTEM_PROMPT=You are a helpful assistant.
```

Then build a full chat app:

```go
package main

import (
    fluui "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/ai"
    "github.com/topcheer/fluui/app"
    "github.com/topcheer/fluui/internal/buffer"
    "github.com/topcheer/fluui/internal/term"
)

func main() {
    base, _ := fluui.New()
    defer base.Close()

    cfg, _ := ai.LoadConfig()
    client := ai.NewClient(cfg)

    w, h := base.Size()
    chat := app.NewChatApp(w, h)
    chat.SetInputHeight(2)
    chat.SetAIClient(client)
    chat.SetSystemPrompt("You are a helpful assistant.")
    chat.SetOnAIError(func(err error) {
        chat.AddUserMessage("Error: " + err.Error())
    })

    // Route keys to ChatApp
    base.OnKey(func(k *term.KeyEvent) {
        chat.HandleKey(k)
        base.MarkDirty()
    })

    // Route mouse to ChatApp
    base.OnMouse(func(m *term.MouseEvent) {
        chat.HandleMouse(m)
        base.MarkDirty()
    })

    // Render ChatApp
    base.OnPaint(func(buf *buffer.Buffer) {
        w, h := base.Size()
        chat.SetSize(w, h)
        chat.Render(buf)
    })

    chat.OnQuit(func() { base.Quit() })
    base.Run()
}
```

Run with `go run main.go`. Type a message and press Enter to chat with the AI.

## Compatible AI Providers

Any OpenAI-compatible API works:

| Provider | Base URL | Models |
|---|---|---|
| ZAI (Zhipu) | `https://open.bigmodel.cn/api/paas/v4` | glm-4-flash, glm-4 |
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat |
| Moonshot | `https://api.moonshot.cn/v1` | moonshot-v1-8k |

## FAQ

**Q: The terminal shows garbled output.**
A: Fluui uses TrueColor (24-bit). Ensure your terminal supports it (iTerm2, Alacritty, Kitty, Windows Terminal, etc.).

**Q: Ctrl+C doesn't exit.**
A: Fluui handles SIGINT/SIGTERM. If you set `OnInterrupt`, return `true` to quit.

**Q: How do I copy text?**
A: Fluui uses OSC52 clipboard sequences. Works in iTerm2, Alacritty, Kitty, and over SSH (with tmux passthrough).

**Q: Mouse scrolling doesn't work.**
A: Ensure your terminal supports SGR mouse mode. Most modern terminals do.

## Next Steps

- [Architecture Overview](architecture.md)
- [API Reference](api-reference.md)
- [Tutorial: Build an AI Agent](tutorial.md)
- [Examples](../examples/)
