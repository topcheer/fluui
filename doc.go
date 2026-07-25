// Package fluui is an AI-native TUI library for Go, built from scratch.
//
// Fluui provides a complete terminal UI framework with 107+ components, 23 terminal
// protocols, 92% zero-allocation rendering, and first-class AI chat support.
//
// # Quick Start
//
//	app, err := fluui.New()
//	if err != nil { panic(err) }
//	defer app.Close()
//
//	app.OnPaint(func(b *buffer.Buffer) {
//		b.DrawText(0, 0, "Hello, Fluui!", buffer.DefaultStyle)
//	})
//	app.Run()
//
// For headless/CI usage without /dev/tty:
//
//	app := fluui.NewWithWriter(os.Stdout, 80, 24)
//
// # AI-Native Components
//
// Fluui includes a complete AI chat framework:
//
//	- component.ConversationView  — scrollable chat history (0 allocs)
//	- component.MessageBubble     — role-based message rendering (0 allocs)
//	- component.ChatComposer      — input box with token display (0 allocs)
//	- component.ToolCallView      — tool/function call visualization (0 allocs)
//	- component.CitationsBlock    — source citations (0 allocs)
//	- component.TokenUsageWidget  — token/cost/context display
//	- component.ThinkingIndicator — animated "AI thinking" dots (0 allocs)
//	- component.SkeletonLoader    — animated loading placeholders (0 allocs)
//	- component.PieChart          — token distribution donut chart (0 allocs)
//	- component.Timeline          — event log with typed markers
//
// # Charts (8 types)
//
// BarChart, LineChart, Sparkline, Gauge (radial), Heatmap,
// PieChart (donut), FunnelChart, RadarChart — all zero-alloc.
//
// # Protocols
//
// 23 terminal protocols supported: OSC52 clipboard, Sixel images, Kitty Graphics,
// iTerm2 inline images, OSC8 hyperlinks, OSC133 shell integration, OSC 9 desktop
// notification, DECSCUSR cursor shape, TrueColor, Kitty Keyboard, SGR Mouse,
// Bracketed Paste, Focus Tracking, Sync Output, Window Titles, Color Query,
// Cursor Save/Restore, Scroll Region, DSR, DA1/DA2/XTVERSION/XTGETTCAP.
//
// # Release
//
// Version info available via fluui.Version and fluui.Info.
// Release process documented in RELEASE.md. CI via GitHub Actions.
//
// # Compatibility
//
// Drop-in compatibility layers for bubbletea, lipgloss, glamour, and x/term.
package fluui
