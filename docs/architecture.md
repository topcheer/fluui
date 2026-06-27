# Architecture

Fluui is built in 6 layers, each with clear responsibilities and zero external TUI dependencies.

## Layer Overview

```
┌─────────────────────────────────────────────┐
│                  app.go                      │  Layer 6: Entry Point
│  App (terminal + renderer + event loop)     │
├─────────────────────────────────────────────┤
│              app/chat.go                     │  Layer 5: ChatApp API
│  ChatApp + AIBridge + InputLine + Mouse     │
├─────────────────────────────────────────────┤
│     block/          overlay/      focus/    │  Layer 4.5: Content + Interaction
│  ThinkingBlock      Modal         Manager   │
│  ToolCallBlock      Popup                   │
│  AssistantText      OverlayManager          │
├─────────────────────────────────────────────┤
│          component/     component/layout/   │  Layer 4: Component System
│  Text  Border  ScrollView   Flex            │
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

## Layer Details

### Layer 1: Terminal (`internal/term/`)

Raw terminal abstraction. No external libraries beyond `golang.org/x/term` for termios.

- Raw mode entry/exit with automatic cleanup
- Alternate screen buffer (restored on exit)
- SGR mouse support (click, drag, scroll)
- Bracketed paste detection
- OSC52 clipboard (copy + paste)
- Input parser state machine (handles split sequences, CJK, multibyte)
- Window resize detection (SIGWINCH)

### Layer 2: Events + Hit Testing (`event/`, `hit/`)

**Event Loop** (`event/`):
- Channel-driven: single goroutine owns all state
- 60fps render tick with dirty-flag optimization
- Dispatcher routes keyboard/mouse/resize/custom events
- Thread-safe via channels (no shared state mutation from handlers)

**Hit Testing** (`hit/`):
- Region tree for mouse hit testing
- O(log n) lookup for click targets

### Layer 3: Buffer + Render (`internal/buffer/`, `render/`)

**Buffer** (`internal/buffer/`):
- `Cell` — single character with Color, Style, width
- `Color` — TrueColor (24-bit), 256-color, named, or default
- `Style` — Bold, Italic, Underline, Strikethrough, Reverse
- `Buffer` — 2D grid of Cells with DrawText, FillRect, etc.
- `Diff` — Row-skip diff algorithm (only redraw changed lines)
- `wcwidth` — Unicode display width (CJK = 2, combining = 0)

**Renderer** (`render/`):
- Double-buffer: front (displayed) + back (drawing)
- Per-frame diff: only changed cells written to terminal
- ANSI batching: consecutive cells with same style grouped

### Layer 3.5: Markdown + Animation

**Markdown** (`markdown/`):
- goldmark parser with table extension
- AST renderer: headings, lists, code blocks, links, tables, quotes
- chroma syntax highlighter (200+ languages)
- CJK-aware line wrapping
- OSC8 clickable hyperlinks
- Table column alignment (left/center/right)

**Animation** (`animation/`):
- Spinner (braille/dots)
- FadeIn
- Manager (frame-synchronized)

### Layer 4: Component System (`component/`)

Interface-based UI primitives:

```go
type Component interface {
    Measure(Constraints) Size
    SetBounds(Rect)
    Paint(*buffer.Buffer)
    Children() []Component
}
```

Built-in: Text, Border, ScrollView (with virtual scrolling + scrollbar drag)

Layout: Flex (Row/Column), Gap, Stack, Center, Padding

### Layer 4.5: Content + Interaction

**Blocks** (`block/`):
- Semantic content blocks: ThinkingBlock, ToolCallBlock, ToolResultBlock, AssistantTextBlock, UserMessageBlock, ErrorBlock
- BlockContainer: ordered collection with spacing, virtual paint
- StreamDispatcher: routes streaming deltas to correct block
- Serialization: SaveContainer/LoadContainer (JSON round-trip)

**Overlay** (`overlay/`):
- OverlayManager: z-index stacking
- Modal: centered dialog
- Popup: fullscreen viewer

**Focus** (`focus/`):
- Tab traversal
- Focus ring

### Layer 5: ChatApp API (`app/`)

High-level interface combining all layers:
- Block management (add/stream/serialize)
- Input line with cursor, history, Ctrl+A/E/U/W
- AI bridge (streaming, cancel, error handling)
- Mouse handler (scroll, click blocks, scrollbar drag)
- Clipboard (OSC52 copy/paste, tmux passthrough)
- Search (Ctrl+F, regex, highlight)
- Theme cycling (Ctrl+T)

### Layer 6: App Entry (`app.go`)

`fluui.New()` wires terminal + renderer + event loop. Provides:
- OnKey/OnMouse/OnResize/OnPaint callbacks
- SIGINT/SIGTERM graceful exit
- Copy/Paste clipboard helpers

## Design Principles

1. **Streaming-first**: Every layer optimized for incremental content updates
2. **Zero-flicker**: Double-buffer diff renderer only writes changed cells
3. **Concurrent-safe**: Single goroutine event loop, mutex-protected state
4. **Semantic blocks**: AI content (Thinking, ToolCall) are first-class types
5. **Zero dependencies**: No TUI framework — everything from scratch
6. **CJK support**: Display width, wrapping, and rendering for Asian languages
7. **Terminal compat**: Capability detection for OSC52, TrueColor, etc.
