# Gap Analysis: ggcode TUI → fluui Migration

> Date: 2026-07-16
> Goal: Identify all capabilities ggcode's TUI depends on from bubbletea/bubbles/lipgloss,
> and map them against fluui's current capabilities.

## 1. Current State

### ggcode TUI Scale
- **169 non-test Go files**, 67,593 lines of non-test code
- **67 test files**
- **108 files** import `charm.land/bubbletea/v2`
- Model struct: 1,509 lines, 120+ fields (single god struct)

### Dependency Chain
```
ggcode TUI depends on:
├── charm.land/bubbletea/v2     (tea)       — 108 files: event loop, Cmd/Msg, Program
├── charm.land/bubbles/v2/
│   ├── textarea                            — main multi-line input (cursor, selection, IME)
│   ├── textinput                           — single-line input (15+ panels)
│   └── viewport                            — scrolling chat list
├── charm.land/lipgloss/v2                  — 609 refs: colors, borders, layout, width
└── golang.org/x/term                       — terminal width detection
```

### Paradigm Difference (CRITICAL)

| Aspect | bubbletea (ggcode) | fluui |
|--------|-------------------|-------|
| Architecture | Elm (Init/Update/View) | Retained Component (Measure/SetBounds/Paint) |
| Rendering | String concatenation per frame | Buffer-based pixel-level painting |
| State | Single god Model struct | Distributed component state |
| Async | `tea.Cmd` → `tea.Msg` (declarative) | `Loop.Send(Event)` (imperative) |
| Update | Central `Update(msg) (Model, Cmd)` | Per-component event handlers via Dispatcher |
| View | Central `View() string` | Per-component `Paint(buf)` composed by layout |
| Styling | lipgloss declarative style strings | theme system + buffer cell attributes |

**This is not a 1:1 API mapping — it's a fundamental architecture rewrite.**

## 2. Gap Analysis by Category

### 2.1 Event Loop & Program Lifecycle

| Capability | bubbletea | fluui | Gap |
|-----------|-----------|-------|-----|
| Event loop | `tea.Program.Run()` | `event.Loop.Run()` | ✅ Equivalent |
| Custom messages | `tea.Msg` (any type) | `event.Event` (typed) | ⚠️ fluui Event is struct-based, need to verify custom data attachment |
| Async commands | `tea.Cmd` (func returning Msg) | `Loop.Send(Event)` | ❌ No Cmd composition (Batch, Sequence, Tick) |
| Quit | `tea.Quit` | `Loop.Quit()` | ✅ Equivalent |
| AltScreen | `tea.WithAltScreen()` | Terminal raw mode | ⚠️ Need to verify fluui enters alt screen |
| Mouse capture | `tea.WithMouseCellMotion()` | `term.MouseEvent` | ⚠️ fluui has mouse events, untested in ggcode context |
| Window title | `tea.SetWindowTitle` | Not found | ❌ Missing |
| Program.Send | `program.Send(msg)` | `Loop.Send(Event)` | ✅ Equivalent (from goroutines) |
| Window size | `tea.WindowSizeMsg` | `Event{Type: TypeResize}` | ✅ Equivalent |

**Gap Severity: MEDIUM** — Cmd composition (tea.Batch, tea.Tick) is the main missing piece.
ggcode uses tea.Batch in 30+ places for parallel async operations.

### 2.2 Input Components

| Component | bubbles | fluui | Gap |
|-----------|---------|-------|-----|
| Multi-line textarea | `textarea.Model` | `component.TextArea` | ⚠️ Need API compat check (cursor, selection, IME, line wrapping) |
| Single-line textinput | `textinput.Model` | fluui has input in chat app | ⚠️ Need to extract as standalone component |
| Autocomplete | Custom (ggcode built on top) | `component.Autocomplete` | ✅ fluui has this |
| Input history | Custom (ggcode ArrowUp/Down) | Not found as component | ⚠️ Easy to implement |

**Gap Severity: HIGH** — textarea is the most critical component. ggcode's entire chat input
relies on textarea's cursor movement, selection, multi-line editing, paste handling, and IME support.
Any incompatibility here blocks migration.

### 2.3 Rendering & Styling

| Capability | lipgloss | fluui | Gap |
|-----------|----------|-------|-----|
| Color (foreground) | `lipgloss.Color("12")` | `theme.Color` / buffer cell fg | ⚠️ Different API, mechanical translation |
| Bold/italic/underline | `style.Bold(true)` | buffer cell style flags | ⚠️ Different API |
| Border | `style.Border(...)` | `component.Border` | ✅ fluui has border component |
| Width measurement | `lipgloss.Width(s)` | buffer measurement | ⚠️ Need equivalent |
| String rendering | `style.Render(s)` returns string | Direct buffer cell writing | ❌ Fundamentally different |
| Padding/Margin | `style.Padding(1,2)` | `layout.Padding` | ✅ fluui has layout |
| Table | Custom lipgloss tables | `component.Table` | ✅ fluui has table |
| Progress bar | Custom | `component.Progress` / `component.Gauge` | ✅ fluui has both |

**Gap Severity: HIGH** — 609 lipgloss references need translation. Not just API renaming,
but paradigm shift from "style a string" to "paint cells with attributes".
This is the bulk of the migration work.

### 2.4 Layout System

| Capability | ggcode (manual) | fluui | Gap |
|-----------|----------------|-------|-----|
| Flex layout | Manual height calculation | `layout.Flex` | ✅ fluui is better |
| Stack layout | Manual | `layout.Stack` | ✅ fluui is better |
| Center layout | Manual | `layout.Center` | ✅ fluui is better |
| Padding | Manual | `layout.Padding` | ✅ fluui is better |
| Split pane | Manual | `component.SplitPane` | ✅ fluui has it |

**Gap Severity: LOW** — fluui's layout system is superior. This is a net win.

### 2.5 Scrolling & Virtual Lists

| Capability | ggcode | fluui | Gap |
|-----------|--------|-------|-----|
| Chat message list | Custom `chat.List` with virtual scroll | `component.VirtualScroller` / `component.ListView` | ⚠️ Need to verify virtual scroll performance |
| Viewport scrolling | `bubbles/viewport` | `component.Viewport` | ✅ fluui has viewport |
| Scrollbar | None (custom indicator) | `component.Scrollbar` | ✅ fluui is better |

**Gap Severity: MEDIUM** — Virtual scrolling for 10K+ messages is performance-critical.
Need to benchmark fluui's VirtualScroller against ggcode's chat.List.

### 2.6 Component Library

fluui has 40+ components that ggcode could use directly:

| fluui Component | ggcode Equivalent | Status |
|----------------|-------------------|--------|
| Table | Custom lipgloss table | ✅ fluui better |
| Tree | File browser (custom) | ✅ fluui better |
| Checkbox | None | ✅ new capability |
| RadioGroup | None | ✅ new capability |
| Slider | None | ✅ new capability |
| ProgressBar | Custom | ✅ equivalent |
| Spinner | `ToolSpinner` | ✅ equivalent |
| Badge | Custom inline | ✅ new capability |
| TabBar | Custom tab rendering | ✅ equivalent |
| StatusBar | Custom status line | ✅ equivalent |
| Dialog | Custom overlay | ✅ equivalent |
| CommandPalette | None (Cmd+K is desktop only) | ✅ new capability |
| ContextMenu | None | ✅ new capability |
| Tooltip | None | ✅ new capability |
| Notification | Custom inline | ✅ equivalent |
| CodeBlock | Custom markdown rendering | ✅ relevant for code display |
| DiffViewer | None | ✅ new capability |
| FilePicker | Custom file browser | ✅ equivalent |
| Form | None | ✅ new capability |
| Wizard | Custom onboard flow | ✅ could replace onboard |
| Searchable | None | ✅ new capability |
| Sparkline | None | ✅ new capability |
| LineChart | None | ✅ new capability |
| BarChart | None | ✅ new capability |
| QRCode | Custom QR overlay | ✅ equivalent |

### 2.7 Markdown Rendering

| Capability | ggcode | fluui | Gap |
|-----------|--------|-------|-----|
| Markdown parsing | goldmark (shared dep) | goldmark | ✅ same |
| Markdown → terminal | Custom lipgloss rendering | `markdown/` package | ⚠️ Need to verify fluui's output quality |
| Syntax highlighting | chroma (shared dep) | chroma | ✅ same |
| Mermaid diagrams | Custom SVG renderer | Not found | ❌ Missing (desktop only anyway) |

**Gap Severity: LOW** — Both use the same underlying libraries.

## 3. Critical Blockers (Must Fix Before Migration)

### 3.1 tea.Cmd Composition (BLOCKER)
ggcode uses these patterns extensively:
```go
tea.Batch(cmd1, cmd2, cmd3)        // parallel execution
tea.Tick(duration, func)            // periodic timer
tea.Sequence(cmd1, cmd2)            // sequential execution
```

**fluui equivalent needed:** A command system that can batch/sequence async operations.
Currently fluui only has `Loop.Send(Event)` which is fire-and-forget.

**Recommendation:** Add a `Command` type to fluui:
```go
type Command func() Event
func Batch(cmds ...Command) Command
func Tick(d time.Duration, f func(time.Time) Event) Command
func Sequence(cmds ...Command) Command
```

### 3.2 TextArea Feature Parity (BLOCKER)
ggcode's textarea must support:
- Multi-line editing with cursor movement (arrows, Home/End, Ctrl+arrow)
- Text selection (Shift+arrow, Ctrl+A)
- Cut/Copy/Paste (Ctrl+X/C/V)
- IME composition (CJK input)
- Auto-resize (grows with content up to max height)
- Placeholder text
- Character counter / token estimate
- Slash command autocomplete overlay
- Input history (ArrowUp/Down recall)
- Draft persistence (survives session switch)

**Status:** fluui has `TextArea` component but API parity is unverified.

### 3.3 lipgloss → Buffer Translation Layer (HIGH EFFORT)
609 lipgloss references need translation. Options:
1. **Compatibility shim:** Create a `lipgloss`-compatible API that writes to fluui buffers
2. **Full rewrite:** Rewrite all rendering in fluui's buffer API
3. **Hybrid:** Keep lipgloss for string generation, convert strings to buffer cells

**Recommendation:** Option 1 (compatibility shim) for fastest migration.

## 4. Migration Strategy

### Phase 1: Framework Core Validation (1-2 weeks)
- [ ] Verify TextArea API parity with bubbles textarea
- [ ] Implement Cmd/Msg system (Batch, Tick, Sequence)
- [ ] Create lipgloss compatibility shim
- [ ] Benchmark VirtualScroller with 10K messages

### Phase 2: Minimal Integration POC (1 week)
- [ ] Create `internal/fluui-repl/` in ggcode
- [ ] Port: input area + message list + status bar
- [ ] Verify: keyboard input, streaming output, session switching

### Phase 3: Feature Parity Migration (4-8 weeks)
- [ ] Port all 15 IM panels
- [ ] Port file browser
- [ ] Port session sidebar
- [ ] Port slash command system
- [ ] Port approval/ask_user dialogs
- [ ] Port harness panel
- [ ] Port LAN chat panel
- [ ] Port extpane (terminal panel integration)

### Phase 4: Polish & Cleanup (1-2 weeks)
- [ ] Remove bubbletea/bubbles/lipgloss dependencies
- [ ] Update tests
- [ ] Performance optimization

## 5. Recommendations for fluui

### Architecture Cleanup
1. **Separate framework from application** — `block/`, `app/`, `ai/` should be a separate
   `fluui-chat` module or examples, not part of the framework core
2. **Add Cmd/Msg system** — This is essential for any Elm-architecture migration
3. **Add lipgloss compat layer** — `fluui/compat/lipgloss` package with Style API

### Missing Components
1. **TextArea** — needs verification/enhancement for multi-line editing parity
2. **TextInput** — extract from chat app as standalone single-line input component
3. **Mermaid/SVG renderer** — needed for diagram rendering (currently desktop-only)

### Testing
1. **Integration test suite** — Port ggcode's keyboard interaction tests to fluui
2. **Performance benchmarks** — Virtual scroll with 10K+ items, 60 FPS rendering

## 6. Conclusion

**Migration is feasible but represents a 6-12 week effort** for full feature parity.

The biggest risk is the paradigm shift from string-based rendering (lipgloss) to
buffer-based rendering (fluui). A compatibility shim can reduce this risk significantly.

fluui's component library and layout system are superior to ggcode's hand-rolled equivalents,
so the migration would be a net improvement in code quality and maintainability.

**Recommended next step:** Implement Phase 1 (framework core validation) to de-risk
the migration before committing to full rewrite.
