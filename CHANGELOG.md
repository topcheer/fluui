# Changelog — Fluui TUI Library

## v1.0.0 (2026-03-01)

### Summary

Fluui is an AI-native Go TUI library providing the most comprehensive terminal UI
component ecosystem. This release includes **196 component files** with **194
runnable examples**, **80+ terminal protocol functions**, and **54 components**
verified as **zero-allocation** in their Paint methods.

### Key Metrics

| Metric | Value |
|--------|-------|
| Component files | 196 |
| Example functions | 194 |
| Terminal protocol functions | 80+ |
| Zero-alloc verified components | 54 (P465-P530) |
| Test packages | 37/37 passing |
| Coverage | 93.7%+ across all packages |
| Vet | Clean |
| Race | Clean |

### Component Categories

#### Direction A — Terminal Protocols
- OSC 8/9/777/99/133/1337 hyperlink, notification, shell integration
- Synchronized output (DECSET 2026)
- SGR mouse (1006), focus events (1004)
- Sixel graphics (DCS q), DECRQM mode requests
- OSC 52 clipboard, DECSED selective erase
- Kitty keyboard protocol, DA1/DA2 device attributes
- Color queries (OSC 10/11/12), iTerm2/Kitty image support

#### Direction B — General UI Components
- **Layout**: Divider, BreadcrumbTrail, Carousel, FileTree, MergeView
- **Input**: MultiSelect, SearchBar, CommandPalette, ChipBadge, PasswordStrength
- **Display**: Calendar, MarkdownTable, TagCloud, Legend, GaugeCluster
- **Navigation**: StepProgress, NotificationStack, ActivityFeed, StatusBarSegment
- **Data**: DataLabel, MiniGauge, Sparkline, ScatterPlot, SankeyChart
- **Indicators**: SpinnerDots, KeyHintBar, ImagePreview, BlockQuote

#### Direction C — AI-Native Components
- **Streaming**: AIStreamRenderer, CodeBlockStream, StreamProgressIndicator
- **Monitoring**: ResponseInspector, ContextWindowBar, RateLimitIndicator
- **Pipeline**: AITokenFlow, FunctionCallVisualizer, AIPanelHeader
- **Confidence**: AIConfidenceBar, CostTracker, TokenMeter
- **UX**: AIThinkingIndicator, StopReasonBadge, ThinkingTrace
- **Markdown**: Full GFM ecosystem (Table, List, Heading, Blockquote, TaskList,
  HorizontalRule, InlineCode, Strikethrough, Emphasis, Link, Footnote,
  DefinitionList, Emoji, Superscript, Subscript, Image)

#### Direction D — Performance
- All 54 components P465-P530 verified zero-allocation via benchmarks
- Pre-allocated byte buffers for formatted strings
- Cached computed values in setters, not in Paint
- `formatDurationBytes` / `writeItoaBytes` zero-alloc formatting helpers

#### Direction E — Developer Experience
- 194 runnable `Example*` functions
- Consistent API patterns across all components
- Thread-safe (mutex-protected) all public methods

### Test Results
- **37/37 packages pass** (100%)
- **0 non-demo TODO/FIXME**
- **Vet clean, -race clean**

### Breaking Changes
None — this is the initial release.
