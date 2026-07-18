# GGCODE.md

> Durable project guidance for AI coding agents working on Fluui.
> Verify and update when conventions change.

## Project Overview

**Fluui** (流畅 + UI) is an AI-native TUI library for Go, built 100% from scratch with zero TUI framework dependencies. It is designed specifically for AI chat interfaces with streaming-first architecture, semantic content blocks, and zero-flicker double-buffer rendering.

- **Module:** `github.com/topcheer/fluui`
- **Go version:** 1.25.0
- **Tests:** 9,238 test functions, 75 benchmarks, all pass with `-race`
- **Size:** ~204K LOC, 69 packages, 232 source files + 513 test files

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

CI runs on push/PR via `.github/workflows/ci.yml` — build, vet, test with race detector.

## External Dependencies (4 only)

| Dependency | Purpose |
|---|---|
| `github.com/yuin/goldmark` | Markdown parsing (AST) |
| `github.com/alecthomas/chroma` | Syntax highlighting for code blocks |
| `github.com/skip2/go-qrcode` | QR code rendering |
| `golang.org/x/sys` | System calls (termios, ioctl, signals) |

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
| `component/` | 35+ components (Table, Checkbox, RadioGroup, Slider, CommandPalette, Spinner, Dialog, CodeBlock, RichLog, Viewport, Help, Badge, ThemeStudio, etc.) |
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

### Compat Packages (charm.land / charmbracelet drop-in)

The `compat/` tree provides drop-in replacements for [ggcode](https://github.com/topcheer/ggcode) and other projects migrating from charm.land TUI libraries to fluui. Source files only need import path changes — no code modifications.

| Package | Replaces |
|---|---|
| `compat/bubbletea/` | `charm.land/bubbletea/v2` |
| `compat/lipgloss/` | `charm.land/lipgloss/v2` |
| `compat/lipgloss/compat/` | `charm.land/lipgloss/v2/compat` |
| `compat/lipgloss/tree/` | `charm.land/lipgloss/v2/tree` |
| `compat/bubbles/textarea/` | `charm.land/bubbles/v2/textarea` |
| `compat/bubbles/textinput/` | `charm.land/bubbles/v2/textinput` |
| `compat/bubbles/viewport/` | `charm.land/bubbles/v2/viewport` |
| `compat/glamour/` | `charm.land/glamour/v2` |
| `compat/glamour/ansi/` | glamour ANSI render context |
| `compat/glamour/styles/` | glamour style configs |
| `compat/xterm/` | `github.com/charmbracelet/x/term` |

**textinput.Model** supports bubbles v2 field-write semantics (`m.EchoMode = textinput.EchoPassword`, `m.Placeholder = "..."`, `m.CharLimit = 512`). The `sync()` method pushes field values to the underlying `*component.TextInput` before View/Update.

### Internal Packages

| Package | Description |
|---|---|
| `internal/buffer/` | `Buffer` (2D Cell grid), `Cell`, `Color`, `Style`, `StyleFlags`, diff, wcwidth, highlight |
| `internal/term/` | Terminal abstraction: raw mode, alt screen, mouse, paste, OSC52, resize, input parser state machine |
| `internal/termcompat/` | Terminal capability detection (12+ terminals, image protocols) |
| `internal/fuzzy/` | Subsequence fuzzy string matching with scoring |
| `internal/hotkey/` | Global hotkey/shortcut registry |
| `internal/hotreload/` | File-watching live reload for development |
| `internal/integration/` | Cross-package integration test helpers |
| `internal/mock/` | Mock terminal for testing (no /dev/tty needed) |
| `internal/recorder/` | Session recording and replay |
| `internal/snapshot/` | UI snapshot testing and diffing |

### Demos and Examples

- `cmd/demo` through `cmd/demo22` — 22 progressive demos
- `examples/` — 11 real-world examples (minimal, chat, dashboard, file-manager, todo-app, calculator, ai-agent, etc.)

## Key Types and APIs

### Component Interface

```go
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
- Phase-numbered tests: `Test<Name>_P<phase>` (currently P1–P289)
- Coverage targets: compat packages at 95%+, component/ at 90%+

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
- **Struct literals:** Always use keyed fields (e.g., `Size{W: 10, H: 5}` not `Size{10, 5}`)
- **No `buf.Clear()`** — use `buf.Fill(buffer.BlankCell)` instead
- **No `term.KeyCtrlU`** — use `k.Rune == 'u' && k.Modifiers&term.ModCtrl != 0`
- **`buffer.Cell`** has `Fg`, `Bg`, `Flags` fields (no `.Style` field)
- **`buffer.Buffer`** has `.Width`, `.Height` fields (no `.Size()` method)

### Git Conventions

- Commit messages: `feat(P<phase>): description` or `test(P<phase>): description`
- All commits include `Co-Authored-By: ggcode <noreply@ggcode.dev>`
- Phase numbering: work is organized in sequential phases (currently P1–P289)
- Push to `main` branch after each phase

### Linting (golangci-lint v2)

Config in `.golangci.yml`:
- Default: standard linters
- Enabled: misspell, unconvert, gocritic, gocyclo (max 50), unparam
- Test files excluded from strict linters
- Demo/example files excluded from strict linters

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
| `docs/guide.md` | Developer guide with component cookbook |
| `docs/best-practices.md` | Concurrency, performance, testing patterns |
| `docs/api-reference.md` | Full API reference |
| `docs/performance.md` | Benchmarking, profiling, optimization strategies |
| `docs/getting-started.md` | 5-minute quickstart |
| `docs/tutorial.md` | Step-by-step tutorial |
| `docs/components.md` | Component catalog |
| `docs/widgets-guide.md` | Widget usage guide |
| `docs/CHANGELOG.md` | Phase changelog |
| `DESIGN.md` | Original design document |
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
| Struct literal `Size{10,5}` | Always use keyed fields `Size{W:10, H:5}` |
| Textinput `m.Prompt()` method | Use `m.Prompt` field (bubbles v2 compat) |

## Performance Notes

- Render pipeline uses ASCII rune cache + utf8.EncodeRune for zero-alloc hot path
- Double-buffer diff only writes changed cells to terminal
- ContainerPaint caches markdown render results (invalidated on content/width change)
- Virtual scrolling for 100+ blocks (binary search for visible range)
- AnnotateBuffer shares single `*Link` per link range (was 1 alloc/cell)
