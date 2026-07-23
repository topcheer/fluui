// Package fluui is an AI-native TUI library for Go, built from scratch.
//
// Fluui provides a complete terminal UI framework with 86+ components, 21 terminal
// protocols, zero-allocation rendering, and first-class AI chat support.
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
//	- component.ConversationView  — scrollable chat history
//	- component.MessageBubble     — role-based message rendering
//	- component.ChatComposer      — input box with token display
//	- component.ToolCallView      — tool/function call visualization
//	- component.CitationsBlock    — source citations
//	- component.TokenUsageWidget  — token/cost/context display
//
// # Protocols
//
// 21 terminal protocols supported: OSC52 clipboard, Sixel images, Kitty Graphics,
// iTerm2 inline images, OSC8 hyperlinks, OSC133 shell integration, TrueColor,
// Kitty Keyboard, SGR Mouse, Bracketed Paste, Focus Tracking, Sync Output,
// Window Titles, Color Query, Cursor Save/Restore, Scroll Region, DSR,
// DA1/DA2/XTVERSION/XTGETTCAP capability detection.
//
// # Compatibility
//
// Drop-in compatibility layers for bubbletea, lipgloss, glamour, and x/term.
package fluui
