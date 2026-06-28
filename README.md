# Fluui

> **流畅 (fluent) + UI** — An AI-native TUI library for Go, built from scratch.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev)
[![Tests](https://img.shields.io/badge/tests-2244-brightgreen)](#testing)
[![License](https://img.shields.io/badge/license-MIT-blue)](#license)

Fluui is a terminal UI framework designed specifically for AI chat interfaces. Every layer — from the input parser to the render engine — is optimized for streaming content, semantic content blocks, and zero-flicker updates.

## Why Fluui?

| Feature | Fluui | Bubble Tea | tview |
|---|---|---|---|
| Streaming-first architecture | Yes | Partial | No |
| Semantic content blocks (Thinking, ToolCall, ToolResult) | Yes | No | No |
| Zero-flicker double-buffer diff | Yes | Yes | Partial |
| Mouse-native (clickable, collapsible blocks) | Yes | No | Partial |
| Built-in AI client (OpenAI-compatible) | Yes | No | No |
| Markdown rendering with code highlighting | Yes | No | No |
| OSC8 clickable hyperlinks | Yes | No | No |
| Plugin system (custom block types) | Yes | No | No |
| Session recording/playback | Yes | No | No |
| Image protocol detection (Sixel/iTerm2/Kitty) | Yes | No | No |
| Terminal compatibility matrix (12+ terminals) | Yes | No | No |
| Multi-line text editor (TextArea) | Yes | No | Basic |
| Command palette (Ctrl+P fuzzy search) | Yes | No | No |
| Tab completion (slash commands, @mentions) | Yes | No | No |
| Search in conversation (Ctrl+F) | Yes | No | No |
| Conversation save/load (serialization) | Yes | No | No |
| Terminal capability detection (termcompat) | Yes | No | No |
| File browser (FilePicker) | Yes | No | No |
| Tab management (TabBar) | Yes | No | No |
| Status bar (StatusBar) | Yes | No | No |
| Diff viewer (DiffPreview) | Yes | No | No |
| Text selection + OSC52 copy | Yes | No | No |
| No TUI framework dependency | Yes (100% from scratch) | N/A | N/A |

## Quick Start

```bash
go get github.com/topcheer/fluui
```

### Minimal Example

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
        buf.DrawText(0, 0, "Hello, Fluui!", buffer.Style{})
    })

    app.Run()
}
```

### AI Chat Example

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

    // Configure AI client
    cfg, _ := ai.LoadConfig()
    client := ai.NewClient(cfg)

    // Create ChatApp with AI
    w, h := base.Size()
    chat := app.NewChatApp(w, h)
    chat.SetInputHeight(2)
    chat.SetAIClient(client)
    chat.SetSystemPrompt("You are a helpful assistant.")
    chat.SetOnAIError(func(err error) {
        chat.AddUserMessage("Error: " + err.Error())
    })

    base.OnKey(func(k *term.KeyEvent) {
        chat.HandleKey(k)
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

## Architecture

```
┌─────────────────────────────────────────────┐
│                  app.go                      │  Layer 6: Entry Point
│  App (terminal + renderer + event loop)     │
├─────────────────────────────────────────────┤
│              app/chat.go                     │  Layer 5: ChatApp API
│  ChatApp + AIBridge + InputLine + StatusBar │
│  TabBar + Selection + MouseHandler         │
├─────────────────────────────────────────────┤
│     block/          overlay/      focus/    │  Layer 4.5: Content + Interaction
│  ThinkingBlock      Modal         Manager   │
│  ToolCallBlock      Popup                   │
│  AssistantText      OverlayManager          │
├─────────────────────────────────────────────┤
│          component/     component/layout/   │  Layer 4: Component System
│  Text  Border  ScrollView  Flex  Table      │
│  Tree  Form  FilePicker  TabBar  StatusBar  │
│  DiffPreview  Links  Selection  Gauge        │
├─────────────────────────────────────────────┤
│          markdown/    animation/            │  Layer 3.5: Rendering
│  goldmark AST    Spinner  FadeIn            │
│  chroma highlight                        │
├─────────────────────────────────────────────┤
│            render/    internal/buffer/      │  Layer 3: Buffer + Render
│  Double-buffer diff   Cell  Color  Style    │
├─────────────────────────────────────────────┤
│            event/    hit/                   │  Layer 2: Events + Hit Testing
│  Channel loop    Region tree                │
├─────────────────────────────────────────────┤
│            internal/term/                    │  Layer 1: Terminal
│  Raw mode  Alt screen  Mouse  Paste  Resize │
│  Input parser state machine                 │
└─────────────────────────────────────────────┘
```

## Packages

| Package | Description | Tests |
|---|---|---|
| `internal/term/` | Terminal abstraction (raw mode, alt screen, mouse, paste) | 35 |
| `internal/buffer/` | Cell, Color, Style, Buffer, Diff, wcwidth (CJK support) | 55 |
| `render/` | Double-buffer diff renderer | 8 |
| `event/` | Channel-driven event loop + dispatcher | 5 |
| `component/` | Component interface, 20+ widgets (Table, Tree, Form, FilePicker, TabBar, StatusBar, DiffPreview, Links, Gauge, Sparkline, Badge, ProgressBar, ContextMenu, Tooltip, SplitPane, HelpOverlay, Notification, TextArea, Selection) | 850+ |
| `component/layout/` | Flex layout (Row/Column/Stack/Center/Padding) | 9 |
| `markdown/` | goldmark AST renderer, chroma highlighter, CJK wrap, OSC8 links, table alignment | 73 |
| `block/` | AI content blocks + container + stream dispatcher + serializer | 189 |
| `overlay/` | Overlay manager, Modal dialog, Popup viewer | 29 |
| `focus/` | Focus manager (Tab traversal, focus ring) | 9 |
| `hit/` | Hit testing (Region, RegionTree) | 12 |
| `animation/` | Spinner, FadeIn, Manager | 16 |
| `app/` | ChatApp API, InputLine, MouseHandler, AIBridge, Clipboard, Search, Selection | 170+ |
| `ai/` | OpenAI-compatible streaming client, config loader | 12 |
| `internal/hotkey/` | Configurable hotkey manager with key sequences | 54 |
| `internal/fuzzy/` | Fuzzy subsequence matcher with scoring | 44 |
| `theme/` | 5 built-in themes, theme cycling, hot-swap, search colors | 18 |
| `internal/termcompat/` | Terminal capability detection (OSC52, true color, tmux) | 25 |
| `internal/term/` | Terminal abstraction + ANSI writer | 78 |

## Configuration

Create a `.env` file (see `.env.example`):

```env
FLUUI_LLM_API_KEY=your-api-key
FLUUI_LLM_BASE_URL=https://open.bigmodel.cn/api/paas/v4
FLUUI_LLM_MODEL=glm-4-flash
# Optional: custom system prompt
FLUUI_LLM_SYSTEM_PROMPT=You are a helpful assistant.
```

Compatible with any OpenAI-compatible API: OpenAI, DeepSeek, ZAI (GLM), Moonshot, etc.

## Demos

```bash
# Phase 1: Terminal + buffer + mouse demo
go run ./cmd/demo/

# Phase 2: Component system (Border/Text/Flex/ScrollView)
go run ./cmd/demo2/

# Phase 3: AI chat simulation (streaming blocks)
go run ./cmd/demo3/

# Phase 10: TextArea + Command Palette + Tab Completion
go run ./cmd/demo7/

# Phase 12: Table/Tree/Form/ProgressBar widgets
go run ./cmd/demo8/

# Phase 13: Gauge/Sparkline/Badge/Notification widgets
go run ./cmd/demo9/

# Phase 14-15: ContextMenu/Tooltip/SplitPane/Help + FilePicker/TabBar/StatusBar
go run ./cmd/demo10/

# Phase 4: ChatApp + overlay + mouse interaction
go run ./cmd/demo4/

# Phase 5: Real AI chat (connects via .env)
go run ./cmd/demo5/

# Phase 8: Full interactive showcase (all features)
go run ./cmd/demo6/

# Phase 10: Production AI Agent demo
 go run ./cmd/demo7/
```

## Examples

```bash
go run ./examples/minimal/    # Hello World
go run ./examples/chat/       # Basic AI chat
go run ./examples/markdown/   # Markdown rendering
go run ./examples/search/     # Search feature
go run ./examples/custom-block/ # Custom block type
```

## Documentation

Full documentation is available in [`docs/`](docs/):

- [Getting Started](docs/getting-started.md) — Install, first program, connect AI
- [Architecture](docs/architecture.md) — 6-layer design overview
- [API Reference](docs/api-reference.md) — Complete public API
- [Tutorial](docs/tutorial.md) — Step-by-step AI Agent tutorial
- [Components](docs/components.md) — Widget system guide (20+ components)
- [Widgets Guide](docs/widgets-guide.md) — FilePicker/TabBar/StatusBar/DiffPreview tutorials
- [Blocks](docs/blocks.md) — Content block types and lifecycle
- [Themes](docs/themes.md) — Theme system and customization
- [Best Practices](docs/best-practices.md) — Concurrency, performance, tips

## Testing

```bash
# Run all tests with race detector
go test ./... -race -count=1

# Run specific package
go test ./internal/term/ -v -race

# Run benchmarks
go test ./... -bench=. -benchmem
```

**1917 tests** across 37 packages, all passing with `-race`. Plus 20 benchmarks across render, buffer, block, and term packages.

## Design Decisions

1. **No TUI framework dependency** — Everything from termios to render engine is built from scratch
2. **Streaming-first** — Every layer optimized for streaming data (< 16ms token-to-pixel)
3. **Block-centric** — Content organized as semantic blocks, not flat text lines
4. **Zero-flicker** — Double-buffer diff rendering ensures smooth streaming
5. **Mouse-native** — Clickable, collapsible blocks with hit region tree
6. **Channel-driven event loop** — Single goroutine owns all state

## Stats

- 258 Go source files
- ~76,535 lines of code
- 2244 tests (race-clean)
- 54 benchmarks
- 40 packages (incl. docs + examples)
- 12 interactive demos
- 10 documentation files
- 5 code examples
- CI/CD: GitHub Actions + golangci-lint

## License

MIT
