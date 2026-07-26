# API Stability Levels

Fluui uses three stability levels to help users understand which APIs
are safe to depend on in production.

## Stability Tiers

### Stable (v1.0.0 frozen)

These APIs are frozen and will not change in a breaking way within the v1.x series.

**Core Framework:**
- `App`, `AppOptions`, `Run()` — application lifecycle
- `Component` interface (`Paint`, `Measure`, `SetBounds`, `Bounds`, `Children`)
- `BaseComponent` — embedded struct for all components
- `Rect`, `Size`, `Constraints` — layout types
- `buffer.Buffer`, `buffer.Cell`, `buffer.Style`, `buffer.Color` — rendering primitives
- `theme.Theme` — color system
- `event.Event`, `event.Handler` — event system
- `focus.Manager` — focus management
- `Version`, `VersionInfo` — versioning

**Stable Components (core, widely used):**
- `Text`, `Button`, `Label`, `Rule`, `Fill`, `Placeholder`
- `Table`, `TreeView`, `ListView`, `ListViewItem`
- `TabbedContent`, `TabBar`
- `Dialog`, `Form`, `Checkbox`, `RadioGroup`, `Slider`
- `ProgressBar`, `Spinner`, `Gauge`
- `CodeBlock`, `MarkdownViewer`
- `Viewport`, `SplitPane`
- `Input`, `TextArea`, `TextField`
- `BarChart`, `LineChart`, `Sparkline`, `PieChart`

### Beta (stable API, may have minor changes)

These components are functional and well-tested but may receive minor API
refinements before being promoted to Stable.

- `MessageBubble`, `ConversationView`, `ChatComposer`, `TokenUsageWidget`
- `ToolCallView`, `CitationsBlock`, `ThinkingIndicator`
- `SessionSidebar`, `ApprovalDialog`
- `DataGrid`, `Heatmap`, `FunnelChart`, `RadarChart`
- `Accordion`, `SegmentedControl`, `SkeletonLoader`
- `Stepper`, `Timeline`, `Pagination`
- `Notification`, `Tooltip`, `Popover`
- `CommandPalette`, `ContextMenu`, `AutoComplete`
- `DiffPreview`, `DiffViewer`
- `Avatar`, `KBD`, `DiffStatBar`, `ConfidenceMeter`
- `Chip`, `ColorSwatch`, `Toast`, `HintLabel`, `StatCard`
- `TerminalProfile`
- All chart components (Barchart, Gauge, etc.)

### Experimental

These APIs are under active development and may change significantly:
- `compat/bubbles/*` — Bubble Tea compatibility shims
- `compat/glamour/*` — Glamour compatibility shims
- `compat/lipgloss/*` — Lipgloss compatibility shims
- `compat/xterm/*` — XTerm compatibility shims
- `hit` — HIT testing utilities
- `internal/hotreload` — hot reload (internal)
- `overlay` — overlay system (API may change)
- `block` — block composition (API may change)

## Versioning Policy

- **v1.0.0-beta.1**: Feature-complete beta. Beta-tier APIs may change.
- **v1.0.0-rc.1**: Release candidate. No new features, only bug fixes.
- **v1.0.0**: First stable release. All Stable-tier APIs frozen.
- **v1.x**: Backward-compatible additions only. Breaking changes require v2.

## Deprecation Policy

Deprecated APIs will be marked with `// Deprecated:` comments and remain
functional for at least one minor version (e.g., v1.1 deprecation → v1.2 removal).
