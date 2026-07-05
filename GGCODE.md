# GGCODE.md

> Durable project guidance for AI coding agents working on Fluui.
> Verify and update when conventions change.

## Project Overview

**Fluui** (流畅 + UI) is an AI-native TUI library for Go, built 100% from scratch with zero TUI framework dependencies. It is designed specifically for AI chat interfaces with streaming-first architecture, semantic content blocks, and zero-flicker double-buffer rendering.

- **Module:** `github.com/topcheer/fluui`
- **Go version:** 1.26+
- **Tests:** 3,328 tests, 58 benchmarks, all pass with `-race`
- **Size:** ~328 Go files, ~100K LOC, 46 packages

## Validation Commands

```bash
# Build (must be clean)
go build ./...

# Vet (must be clean)
go vet ./...

# Full test suite with race detector
go test -race -short ./... -timeout 300s

# Run single package tests
go test -race -count=1 ./component/

# Run benchmarks
go test -bench=. -benchmem ./...

# Lint (golangci-lint v2)
golangci-lint run
```

## External Dependencies (3 only)

| Dependency | Purpose |
|---|---|
| `github.com/yuin/goldmark` | Markdown parsing (AST) |
| `github.com/alecthomas/chroma` | Syntax highlighting for code blocks |
| `golang.org/x/term` | Termios (raw mode) — the only system-level dependency |

No TUI framework, no cgo, no protobuf. Pure Go.

## Architecture (6 Layers)

```
Layer 6: Entry Point       app.go — App (terminal + renderer + event loop)
Layer 5: ChatApp API       app/ — ChatApp + AIBridge + InputLine + Mouse + Selection
Layer 4.5: Content         block/ overlay/ focus/ — Blocks, Modals, FocusManager
Layer 4: Components        component/ component/layout/ — 35+ UI components
Layer 3.5: Rendering       markdown/ animation/ — goldmark AST, chroma, spinner, fade
Layer 3: Buffer + Render   internal/buffer/ render/ — Cell grid, double-buffer diff
Layer 2: Events + Hit      event/ hit/ — Channel loop, O(log n) hit testing
Layer 1: Terminal          internal/term/ — Raw mode, mouse, paste, resize, parser
```

### Key Principle

The event loop owns all state. Mutate ChatApp state only from event callbacks (`OnKey`, `OnMouse`, `OnPaint`). AI streaming runs in background goroutines via channels or AIBridge.

## Package Map

### Core Packages

| Package | Description |
|---|---|
| `.` (root) | `App` struct — wires terminal, renderer, event loop |
| `app/` | `ChatApp`, `AIBridge`, `InputLine`, `SelectionManager`, `SearchMode`, `ReplaceMode`, `WindowManager` |
| `block/` | Semantic content blocks: `ThinkingBlock`, `AssistantTextBlock`, `ToolCallBlock`, `ToolResultBlock`, `UserMessageBlock`, `BlockContainer` |
| `component/` | 35+ components (Table, Checkbox, RadioGroup, Slider, CommandPalette, Spinner, Dialog, etc.) |
| `component/layout/` | Flex, Stack, Center, Padding layout containers |
| `event/` | Channel-driven event loop, Dispatcher, KeyShortcut bindings |
| `render/` | Double-buffer diff renderer with ANSI batching |
| `markdown/` | Goldmark-based markdown renderer with chroma syntax highlighting |
| `animation/` | Spinner frames, easing functions (linear, ease-in/out/back), fade transitions |
| `theme/` | 8 built-in themes (Dracula, Gruvbox, Solarized, etc.), programmatic theme API |
| `overlay/` | Modal dialogs, popups, OverlayManager |
| `focus/` | FocusManager for widget focus cycling |
| `hit/` | Region tree for O(log n) mouse hit testing |
| `ai/` | OpenAI-compatible LLM client with tool-calling support |

### Internal Packages

| Package | Description |
|---|---|
| `internal/buffer/` | `Buffer` (2D Cell grid), `Cell`, `Color`, `Style`, `StyleFlags`, diff, wcwidth, highlight |
| `internal/term/` | Terminal abstraction: raw mode, alt screen, mouse, paste, OSC52, resize, input parser state machine |
| `internal/termcompat/` | Terminal capability detection (12+ terminals, image protocols) |
| `internal/fuzzy/` | Subsequence fuzzy string matching with scoring |
| `internal/hotkey/` | Global hotkey/shortcut registry |
| `internal/integration/` | Cross-package integration test helpers |
| `internal/mock/` | Mock terminal for testing (no /dev/tty needed) |

### Demos and Examples

- `cmd/demo` through `cmd/demo14` — 15 progressive demos
- `examples/` — 11 real-world examples (minimal, chat, dashboard, file-manager, todo-app, calculator, ai-agent, etc.)

## Key Types and APIs

### Core Types

```go
// Component interface (component/component.go)
type Component interface {
    Measure(cs Constraints) Size      // Compute desired size
    SetBounds(r Rect)                 // Apply layout bounds
    Paint(buf *buffer.Buffer)         // Render into buffer
    HandleKey(k *term.KeyEvent) bool  // Handle keyboard input
    Children() []Component            // Return child components
}

// BaseComponent provides defaults; embed in every component
type BaseComponent struct { ... }
```

### Buffer API

```go
buf := buffer.NewBuffer(width, height)
buf.DrawText(x, y, "text", buffer.Style{Fg: buffer.RGB(255,0,0)})
buf.SetCell(x, y, buffer.Cell{Rune: 'A', Width: 1, Fg: color})
buf.Fill(buffer.BlankCell)        // Clear (NOT buf.Clear() — doesn't exist)
cell := buf.GetCell(x, y)         // Returns copy
// buf.Width and buf.Height are FIELDS (int), not methods
```

### Block System

```go
container := block.NewBlockContainer()
container.AddBlock(block.NewUserMessageBlock("id", "text"))
container.AddBlock(block.NewThinkingBlock("id"))
container.AddBlock(block.NewAssistantTextBlock("id"))
container.AddBlock(block.NewToolCallBlock("id", "name", "args"))
```

Block states: `BlockPending` -> `BlockStreaming` -> `BlockComplete` | `BlockError`

### ChatApp API

```go
chat := app.NewChatApp(width, height)
chat.SetAIClient(client)
chat.SendUserMessage("Hello")
chat.OnKey(func(k *term.KeyEvent) { ... })
chat.OnPaint(func(buf *buffer.Buffer) { ... })
// Ctrl+P = command palette, Ctrl+F = search, Ctrl+Z = undo
```

## Conventions

### Testing

- **All tests must pass with `-race`** — thread safety is mandatory
- Use `internal/mock.NewTestTerminal()` for terminal-free testing
- Test helpers: `newTestAppWithTerminal(w, h)`, `newBlockingLoop()` (io.Pipe, never EOF)
- External test packages use `*_test` suffix (e.g., `component_test`)
- Fuzz tests in `internal/term/`, `internal/buffer/`, `markdown/`
- Test name conventions: `Test<ComponentName>_<Scenario>`, e.g., `TestCheckbox_Toggle`

### Concurrency

- Event loop owns all state — mutate only from callbacks
- `ChatApp`, `BlockContainer`, all blocks use `sync.Mutex` or `sync.RWMutex`
- All 35+ components use `sync.RWMutex` internally
- Components return defensive copies (e.g., `Items()`, `Labels()`)
- Never share block pointers across goroutines without locking

### Code Style

- **Indentation:** Tabs (not spaces)
- **Imports:** Standard library first, then external, then internal
- **Style flags:** Use `buffer.Bold`, `buffer.Italic` (NOT `buffer.Style{Bold: true}`)
- **Colors:** Use `buffer.RGB(r, g, b)` for true color, named constants for theme colors
- **Component naming:** `New<Name>()` constructor, embed `BaseComponent`
- **Key handling:** `HandleKey(k *term.KeyEvent) bool` — return true if consumed
- **No `buf.Clear()`** — use `buf.Fill(buffer.BlankCell)` instead
- **No `term.KeyCtrlU`** — use `k.Rune == 'u' && k.Modifiers&term.ModCtrl != 0`
- **`buffer.Cell`** has `Fg`, `Bg`, `Flags` fields (no `.Style` field)
- **`buffer.Buffer`** has `.Width`, `.Height` fields (no `.Size()` method)

### Git Conventions

- Commit messages: `feat(P<phase>-<task>): description` or `test: description`
- All commits include `Co-Authored-By: ggcode <noreply@ggcode.dev>`
- Phase numbering: work is organized in sequential phases (currently P1-P29)

### Linting (golangci-lint v2)

Config in `.golangci.yml`:
- Enabled: gocritic, misspell, unconvert, gocyclo (max 50), unparam
- Test files excluded from strict linters
- Demo/example files excluded from strict linters
- Theme definitions excluded from dupl check

## LLM Configuration

Fluui includes an OpenAI-compatible LLM client (`ai/` package). Configure via `.env`:

```env
FLUUI_LLM_API_KEY=your-key
FLUUI_LLM_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4  # or OpenAI/DeepSeek
FLUUI_LLM_MODEL=glm-5.2
```

Never commit `.env` — it is gitignored.

## Documentation

| File | Content |
|---|---|
| `docs/architecture.md` | 6-layer architecture overview |
| `docs/guide.md` | Developer guide (872 lines) with component cookbook |
| `docs/best-practices.md` | Concurrency, performance, testing patterns |
| `docs/api-reference.md` | Full API reference (780 lines) |
| `docs/performance.md` | Benchmarking, profiling, optimization strategies |
| `docs/getting-started.md` | 5-minute quickstart |
| `docs/tutorial.md` | Step-by-step tutorial |
| `docs/components.md` | Component catalog |
| `docs/widgets-guide.md` | Widget usage guide |
| `docs/CHANGELOG.md` | Phase 1-20 changelog |
| `DESIGN.md` | Original design document (71KB) |
| `README.md` | Project overview and feature comparison |

## API Gotchas

| Pitfall | Correct Way |
|---|---|
| `buf.Clear()` doesn't exist | `buf.Fill(buffer.BlankCell)` |
| `term.KeyArrowUp` doesn't exist | `term.KeyUp`, `term.KeyDown` |
| `buffer.Style{Bold: true}` doesn't work | `buffer.Style{Flags: buffer.Bold}` |
| `Table.SetCell()` doesn't exist | `Table.SetRows([][]string)` |
| `CommandPalette.IsVisible()` | `CommandPalette.Visible()` |
| `Spinner.IsActive()` | `Spinner.Running()` |
| `buffer.Cell.Style` field | Use `Fg`, `Bg`, `Flags` directly |
| `buffer.Buffer.Size()` method | Use `.Width`, `.Height` fields |
| `term.KeyCtrlU` doesn't exist | `k.Rune == 'u' && k.Modifiers&term.ModCtrl != 0` |
| `CommandPalette.SetFilter()` | `CommandPalette.SetQuery()` |
| `InputLine.SetInput()` | `InputLine.SetText()` |

## Performance Notes

- Render pipeline uses ASCII rune cache + utf8.EncodeRune for zero-alloc hot path
- Double-buffer diff only writes changed cells to terminal
- ContainerPaint caches markdown render results (invalidated on content/width change)
- Virtual scrolling for 100+ blocks (binary search for visible range)
- AnnotateBuffer shares single `*Link` per link range (was 1 alloc/cell)
