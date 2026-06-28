# Changelog

All notable changes to Fluui across 20 development phases.

## Phase 20 — Documentation Sync + Integration (Latest)

### P20-A: Documentation Update
- Updated `docs/components.md` with Dialog, AutoComplete, Wizard, Checkbox, RadioGroup, Slider, CommandPalette, Spinner sections
- Updated `docs/widgets-guide.md` with Phase 18-19 component tutorials
- Updated `docs/api-reference.md` with P19 component API tables, Undo/Redo, Theme Management
- Updated `README.md` with current stats (2509+ tests, 30+ components)

### P20-B: Demo13
- `cmd/demo13/main.go` — showcases all 5 Phase 19 components (print-based)

### P20-C: Root Package Tests + CI
- 34 new root package tests (`app_test.go`) — 0% → 59.3% coverage
- App struct: added `sync.Mutex` for concurrent resize safety
- Fixed terminal test hang: `RUN_TERMINAL_TESTS` env guard

### P20-D: ChatApp Integration
- CommandPalette integration (Ctrl+P) for command execution
- Spinner integration for loading states
- 18 new tests

**Stats: 2,555 tests | 277 files | ~83,638 LOC | 41 packages**

---

## Phase 19 — Advanced UI Widgets + ChatApp Features

### P19-A: Checkbox + RadioGroup + Slider
- `component/checkbox.go` (404 lines) — toggle, CheckAll (Ctrl+A), UncheckAll (Ctrl+D), j/k nav
- `component/radiogroup.go` (350 lines) — mutual exclusion, disabled skipping, wrap-around
- `component/slider.go` (510 lines) — H/V orientation, Increment/Decrement, Home/End, h/l vim keys
- 127 tests, 2,585 lines (commit `c000d1e`)

### P19-B: CommandPalette + Spinner
- `component/commandpalette.go` (706 lines) — fuzzy search via `internal/fuzzy`, cursor nav with wrap, scroll management, OnExecute/OnDismiss
- `component/spinner.go` (286 lines) — animation integration (dots/arc/line/bounce/bars), Start/Stop, Update(dt)
- 77 tests, ~1,950 lines (commit `f103f4a`)

### P19-C: InputLine Undo/Redo + ChatApp Theme
- `app/inputline_undo.go` (207 lines) — undoStack with sync.Mutex, max 100 states, Ctrl+Z/Y
- `app/chat_theme.go` (82 lines) — Ctrl+]/Ctrl+\ theme cycling, SetThemeByIndex/Name
- 39 tests, 891 lines (commit `1e7b20f`)

### P19-D: Root Package + Event Coverage
- `event/p19_coverage_test.go` — 30+ tests for Dispatcher, Loop, FromTermEvent
- Fixed: `d.OnCustom` → `d.OnPaste`, `Event{Data}` → `Event{Paste}`, terminal test hang guards
- (commit `901b5bc`)

**Stats: ~273 new tests | 5 new components + 2 new features**

---

## Phase 18 — Dialog, AutoComplete, Wizard Components

- `component/dialog.go` — Modal dialogs: Confirm, Info, Prompt, Custom types
- `component/autocomplete.go` — Popup fuzzy autocomplete with OnSelect/OnDismiss
- `component/wizard.go` — Multi-step wizard with lifecycle hooks (OnEnter/OnLeave), dynamic buttons
- `cmd/demo12/main.go` — Full production demo
- (commit `5d44641`)

---

## Phase 17 — Production Hardening + Performance

- CI/CD pipeline: GitHub Actions + golangci-lint
- Performance optimization: render path, buffer operations
- Linter fixes: errcheck, gocritic style cleanup
- (commits `c647336`, `b39eed0`)

---

## Phase 16 — ChatApp Integration + Benchmarks + Documentation

- ChatApp integration: TabBar, StatusBar, DiffPreview wired into chat UI
- 54 benchmarks across render, buffer, block, term packages
- Documentation: getting-started, tutorial, best-practices, architecture, themes
- `examples/ai-agent/` — full AI agent example
- (commit `14a7a46`)

---

## Phase 15 — FilePicker, StatusBar, Selection, TabBar, Links, DiffPreview

- `component/filepicker.go` — File browser with fuzzy filter, multi-select, vim keys
- `component/tabbar.go` — Tab management with close buttons, keyboard nav
- `component/statusbar.go` — Status bar with left/center/right alignment
- `component/diffpreview.go` — Scrollable diff viewer with syntax highlighting
- LinkManager — URL detection, OSC8 hyperlinks, click handling
- SelectionManager — Text selection with OSC52 clipboard copy
- `cmd/demo10/` — Phase 15 components demo
- (commit `104a05e`)

---

## Phase 14 — Notification, HelpOverlay, HotkeyManager

- `component/notification.go` — Toast notifications with auto-expiry (4 levels)
- `component/helpoverlay.go` — Searchable keyboard shortcut cheatsheet
- HotkeyManager — Configurable hotkeys with key sequences

---

## Phases 1-13 — Foundation

### Phase 1-4: Core Infrastructure
- Terminal layer (`internal/term/`): raw mode, key/mouse/resize/paste events
- Buffer layer (`buffer/`): cell grid, styles, diff renderer, double-buffering
- Render layer (`render/`): painter, dirty tracking, efficient redraws

### Phase 5-7: Block System
- `block/` package: semantic AI content blocks (Thinking, ToolCall, ToolResult, AssistantText, UserMessage, Error)
- BlockContainer: ordered collection with virtual paint
- StreamDispatcher: routes streaming deltas to correct block
- Serialization: SaveContainer/LoadContainer (JSON round-trip)

### Phase 8-10: Component System
- `component/` package: Component interface (Measure/SetBounds/Paint/Children)
- Base components: Text, Border, ScrollView, Flex (Row/Column/Stack/Center/Padding)
- Data components: Table, Tree, Form, Gauge, Sparkline, Badge, ProgressBar
- Interactive: ContextMenu, Tooltip, SplitPane, TextArea

### Phase 11-13: Integration & Polish
- Plugin system (`plugin/`)
- Focus manager (`focus/`)
- Hit testing (`hit/`)
- Animation (`animation/`): Spinner, FadeIn
- Overlay manager (`overlay/`)
- Markdown renderer (`markdown/`): goldmark AST, chroma highlighter, CJK wrap, OSC8 links
- ChatApp API (`app/`): InputLine, MouseHandler, AIBridge, Clipboard, Search

**Initial release: 25 tests, ~50 files, ~5,000 LOC** (commit `be676c7`)

---

## Summary

| Phase | Focus | Key Deliverables |
|---|---|---|
| 1-4 | Core infrastructure | Terminal, Buffer, Renderer |
| 5-7 | Block system | Semantic AI blocks, streaming, serialization |
| 8-10 | Component system | 15+ base components, layout |
| 11-13 | Integration | Plugins, focus, hit testing, markdown, ChatApp |
| 14 | Polish | Notifications, help overlay, hotkeys |
| 15 | File & nav | FilePicker, TabBar, StatusBar, DiffPreview |
| 16 | Integration | ChatApp wiring, benchmarks, docs |
| 17 | Hardening | CI/CD, performance, linter fixes |
| 18 | Dialogs | Dialog, AutoComplete, Wizard |
| 19 | Advanced widgets | Checkbox, RadioGroup, Slider, CommandPalette, Spinner, Undo/Redo |
| 20 | Sync | Docs, Demo13, root tests, ChatApp integration |

**Growth: 25 → 2,555 tests (102x), 20 phases, zero dependencies.**
