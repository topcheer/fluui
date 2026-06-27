// Package main implements demo9 — Phase 13 Productivity Widgets Showcase.
//
// A print-based demo that renders sample frames of every Phase 13 widget:
// Gauge, Sparkline, Badge, Fuzzy Matcher, and Notification/Toast.
//
// Usage: go run ./cmd/demo9/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/fuzzy"
	"github.com/topcheer/fluui/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 9 — Phase 13 Productivity Widgets Showcase")
		fmt.Println("Usage: go run ./cmd/demo9/")
		return
	}

	theme.SetActive(theme.Dracula())
	width := 70

	// ─── Gauge Demo ───────────────────────────────────
	printBanner(width, "Gauge Widget")

	// Horizontal gauges at different values
	fmt.Println("  Horizontal (gradient green→red):")
	for _, val := range []float64{0, 25, 50, 75, 100} {
		g := component.NewGauge()
		g.SetValue(val)
		g.SetLabel(fmt.Sprintf("CPU Load %3.0f%%", val))
		g.SetBounds(component.Rect{X: 0, Y: 0, W: width - 4, H: 2})
		buf := buffer.NewBuffer(width, 2)
		g.Paint(buf)
		printBufferLines(buf, width, 2)
	}

	fmt.Println()
	fmt.Println("  Horizontal with custom thresholds (green < 60, yellow < 85, red ≥ 85):")
	for _, val := range []float64{45, 72, 90} {
		g := component.NewGauge()
		g.SetValue(val)
		g.SetThresholds(component.DefaultThresholds())
		g.SetBounds(component.Rect{X: 0, Y: 0, W: width - 4, H: 1})
		buf := buffer.NewBuffer(width, 1)
		g.Paint(buf)
		fmt.Println("  " + renderLine(buf, width))
	}

	fmt.Println()
	fmt.Println("  Vertical gauge (bottom-to-top):")
	g := component.NewGauge()
	g.SetValue(60)
	g.SetOrientation(component.GaugeVertical)
	g.SetBounds(component.Rect{X: 0, Y: 0, W: 1, H: 6})
	buf := buffer.NewBuffer(1, 6)
	g.Paint(buf)
	for y := 0; y < 6; y++ {
		fmt.Println("  " + renderLineAt(buf, 1, y))
	}

	fmt.Println()
	fmt.Println("  Radial gauge (40% and 80%):")
	for _, val := range []float64{40, 80} {
		g := component.NewGauge()
		g.SetValue(val)
		g.SetRadial(true)
		g.SetBounds(component.Rect{X: 0, Y: 0, W: 7, H: 5})
		buf := buffer.NewBuffer(7, 5)
		g.Paint(buf)
		printBufferLines(buf, 7, 5)
		fmt.Println()
	}

	fmt.Println()

	// ─── Sparkline Demo ───────────────────────────────
	printBanner(width, "Sparkline Widget")

	// Basic sparkline with varying data
	fmt.Println("  Single color (green):")
	sl := component.NewSparkline()
	sl.SetData([]float64{3, 5, 2, 8, 5, 7, 4, 9, 6, 10, 8, 12})
	sl.SetLabel("requests/s")
	sl.SetColor(buffer.NamedColor(buffer.NamedGreen))
	sl.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf = buffer.NewBuffer(width, 1)
	sl.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()
	fmt.Println("  Gradient mode (green→red by bar height):")
	sl2 := component.NewSparkline()
	sl2.SetData([]float64{1, 3, 5, 7, 9, 7, 5, 3, 1, 5, 10, 5})
	sl2.SetColorMode(component.ColorGradient)
	sl2.SetShowMinMax(true)
	sl2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf = buffer.NewBuffer(width, 1)
	sl2.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()
	fmt.Println("  Value-threshold mode (based on actual values):")
	sl3 := component.NewSparkline()
	sl3.SetData([]float64{10, 25, 50, 75, 90, 60, 40, 20, 55, 80, 35, 65})
	sl3.SetColorMode(component.ColorValue)
	sl3.SetLabel("latency ms")
	sl3.SetShowMinMax(true)
	sl3.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf = buffer.NewBuffer(width, 1)
	sl3.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()
	fmt.Println("  Negative values (range -10 to +10):")
	sl4 := component.NewSparkline()
	sl4.SetData([]float64{-5, -2, 3, 8, 5, -3, -7, 0, 4, 9, 6, -1})
	sl4.SetColorMode(component.ColorGradient)
	sl4.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf = buffer.NewBuffer(width, 1)
	sl4.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()

	// ─── Badge Demo ───────────────────────────────────
	printBanner(width, "Badge Component")

	// Single badges of each variant
	fmt.Println("  Variants:")
	badges := []struct {
		text    string
		variant component.BadgeVariant
	}{
		{"INFO", component.BadgeInfo},
		{"PASS", component.BadgeSuccess},
		{"WARN", component.BadgeWarning},
		{"FAIL", component.BadgeError},
		{"CRITICAL", component.BadgeCritical},
		{"NEUTRAL", component.BadgeNeutral},
	}

	// Create a BadgeGroup and render each badge individually
	for _, b := range badges {
		bd := component.NewBadge(b.text, b.variant)
		bd.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 1})
		buf = buffer.NewBuffer(20, 1)
		bd.Paint(buf)
		fmt.Printf("  %-10s → %s\n", b.text, renderLine(buf, 20))
	}

	fmt.Println()
	fmt.Println("  BadgeGroup (horizontal layout):")
	group := component.NewBadgeGroup()
	group.SetSpacing(1)
	group.Add(component.NewBadge("v2.0", component.BadgeInfo))
	group.Add(component.NewBadge("stable", component.BadgeSuccess))
	group.Add(component.NewBadge("beta", component.BadgeWarning))
	group.Add(component.NewBadge("deprecated", component.BadgeError))
	group.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf = buffer.NewBuffer(width, 1)
	group.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()
	fmt.Println("  Sizes (small / normal / large):")
	for _, size := range []component.BadgeSize{component.BadgeSizeSmall, component.BadgeSizeNormal, component.BadgeSizeLarge} {
		bd := component.NewBadgeWithSize("BUILD", component.BadgeSuccess, size)
		sz := bd.Measure(component.Unbounded())
		bd.SetBounds(component.Rect{X: 0, Y: 0, W: sz.W, H: sz.H})
		buf = buffer.NewBuffer(sz.W, sz.H)
		bd.Paint(buf)
		fmt.Printf("  %-6s %s\n", component.SizeName(size), renderLine(buf, sz.W))
	}

	fmt.Println()

	// ─── Fuzzy Matcher Demo ───────────────────────────
	printBanner(width, "Fuzzy Matcher")

	candidates := []string{
		"component/gauge.go",
		"component/sparkline.go",
		"component/badge.go",
		"component/notification.go",
		"internal/fuzzy/matcher.go",
		"internal/buffer/buffer.go",
		"internal/term/input.go",
		"block/thinking.go",
		"block/tool_call.go",
		"block/workflow.go",
		"app/chat.go",
		"app/search.go",
	}

	queries := []string{"gauge", "comp", "go", "t", "bl"}
	matcher := fuzzy.NewMatcher()

	for _, q := range queries {
		fmt.Printf("  Query: \"%s\"\n", q)
		results := matcher.RankTopN(q, candidates, 4)
		if len(results) == 0 {
			fmt.Println("    (no matches)")
		}
		for _, r := range results {
			// Render with highlight markers
			highlighted := renderHighlight(r.Item, r.Positions)
			fmt.Printf("    %6.1f  %s\n", r.Score, highlighted)
		}
		fmt.Println()
	}

	// ─── Notification/Toast Demo ──────────────────────
	printBanner(width, "Notification / Toast System")

	// Create toast manager and push various notifications
	tm := component.NewToastManager(5)
	tm.PushInfo("Welcome", "Phase 13 widgets loaded successfully.")
	tm.PushSuccess("Build Complete", "All 20 packages compiled with zero errors.")
	tm.PushWarning("Deprecated API", "OldStyleSheet will be removed in v3.0.")
	tm.PushError("Connection Lost", "Failed to reach LLM endpoint (timeout).")

	fmt.Println("  Toast notifications (stacked):")
	tm.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 12})
	buf = buffer.NewBuffer(width, 12)
	tm.Paint(buf)
	printBufferLines(buf, width, 12)

	fmt.Println()
	fmt.Println("  Level icons and durations:")
	for _, level := range []component.NotificationLevel{
		component.LevelInfo,
		component.LevelSuccess,
		component.LevelWarning,
		component.LevelError,
	} {
		fmt.Printf("    %s  %-8s  duration: %s  accent: %s\n",
			level.Icon(),
			level.String(),
			component.DefaultDurationFor(level),
			colorName(level.AccentColor()),
		)
	}

	fmt.Println()
	fmt.Println("  Dismiss + auto-expiry:")
	tm2 := component.NewToastManager(3)
	id := tm2.PushInfo("Test", "This will be dismissed manually")
	fmt.Printf("    Pushed: %s, count=%d\n", id, tm2.Count())
	tm2.Dismiss(id)
	fmt.Printf("    After dismiss: count=%d\n", tm2.Count())

	tm2.PushSuccess("Auto", "This will auto-expire")
	tm2.PushError("Persistent", "Error notifications persist")
	fmt.Printf("    Before tick: count=%d\n", tm2.Count())
	expired := tm2.Tick()
	fmt.Printf("    After tick: expired=%d, remaining=%d\n", len(expired), tm2.Count())

	fmt.Println()

	// ─── Summary ──────────────────────────────────────
	printBanner(width, "Phase 13 Complete")
	fmt.Println("  Gauge         44 tests  |  linear/vertical/radial + thresholds")
	fmt.Println("  Sparkline     36 tests  |  3 color modes + autoscale + scroll")
	fmt.Println("  Badge         52 tests  |  6 variants + 3 sizes + BadgeGroup")
	fmt.Println("  Fuzzy Matcher 44 tests  |  subsequence + scoring + highlight")
	fmt.Println("  Notification  32 tests  |  4 levels + auto-expiry + stacking")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Println("  Total:      +208 tests  |  Phase 13 production-ready")
	fmt.Println()
}

// printBanner prints a section banner.
func printBanner(width int, title string) {
	pad := width - 4 - len(title)
	if pad < 0 {
		pad = 0
	}
	fmt.Println()
	fmt.Printf("  ┌─ %s%s──┐\n", title, strings.Repeat("─", pad))
	fmt.Println()
}

// renderLine renders a single buffer row as a string (trimming trailing spaces).
func renderLine(buf *buffer.Buffer, width int) string {
	return renderLineAt(buf, width, 0)
}

// renderLineAt renders a specific row from the buffer.
func renderLineAt(buf *buffer.Buffer, width, y int) string {
	var sb strings.Builder
	w := buf.Width
	if width < w {
		w = width
	}
	for x := 0; x < w; x++ {
		cell := buf.GetCell(x, y)
		if cell.Width == 0 {
			continue
		}
		sb.WriteRune(cell.Rune)
	}
	return strings.TrimRight(sb.String(), " ")
}

// printBufferLines renders all rows from a buffer.
func printBufferLines(buf *buffer.Buffer, width, height int) {
	for y := 0; y < height; y++ {
		line := renderLineAt(buf, width, y)
		if strings.TrimSpace(line) != "" {
			fmt.Println("  " + line)
		}
	}
}

// renderHighlight wraps matched positions in [] brackets.
func renderHighlight(text string, positions []int) string {
	if len(positions) == 0 {
		return text
	}
	posSet := make(map[int]bool)
	for _, p := range positions {
		posSet[p] = true
	}
	runes := []rune(text)
	var sb strings.Builder
	for i, r := range runes {
		if posSet[i] {
			sb.WriteString("[")
			sb.WriteRune(r)
			sb.WriteString("]")
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// colorName returns a human-readable name for a Color.
func colorName(c buffer.Color) string {
	switch c.Type {
	case buffer.ColorNamed:
		switch c.Val {
		case buffer.NamedRed:
			return "red"
		case buffer.NamedGreen:
			return "green"
		case buffer.NamedYellow:
			return "yellow"
		case buffer.NamedBlue:
			return "blue"
		case buffer.NamedMagenta:
			return "magenta"
		case buffer.NamedCyan:
			return "cyan"
		case buffer.NamedWhite:
			return "white"
		}
		return fmt.Sprintf("named(%d)", c.Val)
	case buffer.Color256:
		return fmt.Sprintf("256(%d)", c.Val)
	case buffer.ColorTrue:
		return fmt.Sprintf("#%06X", c.Val)
	default:
		return "default"
	}
}
