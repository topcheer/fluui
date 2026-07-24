# Changelog

All notable changes to Fluui are documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Performance — All Zero-Allocation Rendering (P347)
- **Table: 1→0 allocs** via visibleColumnRangeLocked (index range, not slice)
- **22/25 benchmarked Paint operations now zero-alloc**

### Coverage (P346)
- Heatmap Paint coverage 64→98%

### Components — SegmentedControl + SkeletonLoader (P339)
- **SegmentedControl** — iOS-style segmented mode switcher (zero-alloc Paint)
- **SkeletonLoader** — animated loading placeholder blocks for AI loading states
- Component count: **89**

### Performance — Continued Zero-Alloc Optimization (P340-P341)
- **ChatComposer: 2→0 allocs** via stack-buffer token formatting
- **ToolCallView expanded: 12→3 allocs** via eliminating []rune/string(r) conversions
- **16/19 benchmarked components now zero-alloc Paint**

### Coverage (P338)
- appendTokenCount 66→100%, appendProgressBar 75→92%
- **ConversationView: 402→0 allocs** (10 messages) — fully zero-allocation chat rendering
- **MessageBubble: 3→0 allocs** (short) via `time.AppendFormat` on stack buffer
- **CitationsBlock: 7→0 allocs** (collapsed) via single stack buffer formatting
- **ToolCallView: 2→0 allocs** (collapsed) via direct-to-buffer piecewise drawing
- **TokenUsageWidget: 5→1 alloc** via single [256]byte stack buffer
- **Zero-copy `countWrappedLines`** — byte-slice word-wrap scanning replaces strings.Split/Fields

### Added — Protocols & Components (P331-P332)
- **DECSCUSR cursor shape** — 7 cursor styles (blinking/steady block/underline/bar)
- **OSC 9 desktop notification** — system notifications for iTerm2/WezTerm
- **ThinkingIndicator** — animated "AI thinking" three-dot indicator (component 87)
- Protocol count: **23 classes**

### Coverage (P329)
- Root package coverage 85.9% → 93.8% (newFromTerminal + Run lifecycle tests)

### Added — AI-Native Chat Framework (P295-P300)
- **ConversationView** — scrollable chat history with auto-scroll
- **MessageBubble** — role-based message rendering (User/Assistant/System/Tool)
- **ChatComposer** — input box with Enter-to-send, token display, slash commands
- **ToolCallView** — AI tool/function call visualization with streaming results
- **CitationsBlock** — source citations with collapsible detail view
- **TokenUsageWidget** — token consumption, cost estimation, context window bar
- **MarkdownViewer streaming** — incremental AI markdown rendering with debounce
- **StreamingRenderer** — markdown incremental parser with cached blocks

### Added — Terminal Protocols (P293-P297)
- OSC 133 shell integration / prompt marking (5 functions + parser)
- DA1/DA2/XTVERSION/XTGETTCAP capability detection (4 query types + parsers)
- Cursor Save/Restore (DECSC/DECRC + ANSI.SYS variants)
- Scroll Region (DECSTBM set/reset)
- DSR cursor position, terminal size, cell size queries
- **ProtocolCapabilities** — unified detection from DA1/DA2/XTVERSION/env
- **OSC8 hyperlinks** in CitationsBlock
- Protocol count: **21 classes (27 constants)**

### Added — Developer Experience (P306-P309)
- `NewWithWriter(w, width, height)` — headless/CI App construction
- `MockTerminal` — testable terminal I/O without /dev/tty
- `Makefile` — 13 dev targets (build, test, bench, coverage, lint)
- `CONTRIBUTING.md` — contribution guide
- `doc.go` — root package documentation for godoc/pkg.go.dev
- PTY integration test scaffolding

### Added — Component Enhancements (P302)
- TabBar close button (`SetOnCloseTab`, `HandleCloseClick`, `SetTabClosable`)
- Close button character upgraded to `✕`
- ProtocolCapabilities wired into AppShell (`DetectCapabilities`, `Capabilities`)

### Performance (P303-P317)
- MessageBubble long: 127→15 allocs (88% reduction)
- ConversationView 10msg: 402→148 allocs (63% reduction)
- TokenUsageWidget: 17→5 allocs (71% reduction)
- CitationsBlock expanded: 11→**0 allocs** (zero-allocation rendering)
- ToolCallView collapsed: 25→2 allocs (92% reduction)
- ChatComposer: 7→2 allocs (71% reduction)
- Techniques: `utf8.RuneCountInString`, `strconv.AppendFloat`, `renderBubble` reuse

### Coverage
- All packages: 92.8-99.4% (weighted average ~96.3%)
- 85 benchmarks across component, markdown, render, buffer packages
- Zero TODO/FIXME in core code
- Zero `go vet` warnings

### Dependencies
- Unchanged: goldmark (markdown) + chroma (syntax highlighting)

## [0.1.0] - 2026-01

### Initial Release
- 85+ components matching Textual/Bubble Tea/tview/Ratatui
- Zero-allocation double-buffered renderer
- 13 terminal protocols (OSC52, Sixel, Kitty Graphics, iTerm2 images, OSC8, etc.)
- Full markdown rendering (headers, code, tables, math, mermaid, diff, task lists)
- Compat layers for bubbletea, lipgloss, glamour, x/term
- Theme system with JSON serialization
- Event system with hotkeys, focus tracking
- Hot reload, snapshot testing, recorder
- 12 example applications
