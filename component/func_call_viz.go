package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── FunctionCallVisualizer: AI Tool/Function Call Chain Display ───
//
// FunctionCallVisualizer renders a sequence of AI function/tool calls with
// indentation for nested calls, showing name, args summary, duration, and
// status (running, success, error). Useful in AI agent debugging and
// observability dashboards.
//
// Usage:
//
//	fcv := NewFunctionCallVisualizer()
//	fcv.AddCall("search_web", `{"q":"go tui"}`, 120*time.Millisecond, "success")
//	fcv.AddCall("read_file", `{"path":"main.go"}`, 5*time.Millisecond, "success")
//	fcv.Paint(buf)

// CallStatus describes the outcome of a function call.
type CallStatus string

const (
	CallRunning CallStatus = "running"
	CallSuccess CallStatus = "success"
	CallError   CallStatus = "error"
)

// FuncCall represents a single function/tool call entry.
type FuncCall struct {
	Name     string
	Args     string
	Duration time.Duration
	DurStr   string // cached formatted duration (pre-allocated)
	Status   CallStatus
	Indent   int // nesting level
}

// FuncCallVizStyle holds styling for FunctionCallVisualizer.
type FuncCallVizStyle struct {
	Name       buffer.Style
	Args       buffer.Style
	Duration   buffer.Style
	Status     [3]buffer.Style // [running, success, error]
	Indent     buffer.Style
	Border     buffer.Style
}

// DefaultFuncCallVizStyle returns sensible defaults.
func DefaultFuncCallVizStyle() FuncCallVizStyle {
	name := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold} // blue-400 bold
	args := buffer.Style{Fg: buffer.RGB(148, 163, 184)}                     // slate-400
	dur := buffer.Style{Fg: buffer.RGB(100, 116, 139)}                     // slate-500
	running := buffer.Style{Fg: buffer.RGB(234, 179, 8)}                   // yellow-500
	success := buffer.Style{Fg: buffer.RGB(34, 197, 94)}                   // green-500
	errCol := buffer.Style{Fg: buffer.RGB(239, 68, 68)}                    // red-500
	indent := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                    // slate-600
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                    // slate-600
	return FuncCallVizStyle{
		Name:     name,
		Args:     args,
		Duration: dur,
		Status:   [3]buffer.Style{running, success, errCol},
		Indent:   indent,
		Border:   border,
	}
}

// FunctionCallVisualizer displays a chain of AI function/tool calls.
type FunctionCallVisualizer struct {
	BaseComponent
	mu sync.Mutex

	calls []FuncCall
	style FuncCallVizStyle
}

// NewFunctionCallVisualizer creates a FunctionCallVisualizer with defaults.
func NewFunctionCallVisualizer() *FunctionCallVisualizer {
	fcv := &FunctionCallVisualizer{
		style: DefaultFuncCallVizStyle(),
	}
	fcv.SetID(GenerateID("funccall"))
	return fcv
}

// AddCall appends a function call entry. indent controls nesting depth.
func (fcv *FunctionCallVisualizer) AddCall(name, args string, duration time.Duration, status CallStatus) *FunctionCallVisualizer {
	fcv.mu.Lock()
	fcv.calls = append(fcv.calls, FuncCall{
		Name:     name,
		Args:     args,
		Duration: duration,
		DurStr:   formatInspectorDuration(duration),
		Status:   status,
		Indent:   0,
	})
	fcv.mu.Unlock()
	return fcv
}

// AddNestedCall appends a function call with a specific indentation level.
func (fcv *FunctionCallVisualizer) AddNestedCall(name, args string, duration time.Duration, status CallStatus, indent int) *FunctionCallVisualizer {
	fcv.mu.Lock()
	fcv.calls = append(fcv.calls, FuncCall{
		Name:     name,
		Args:     args,
		Duration: duration,
		DurStr:   formatInspectorDuration(duration),
		Status:   status,
		Indent:   indent,
	})
	fcv.mu.Unlock()
	return fcv
}

// Clear removes all calls.
func (fcv *FunctionCallVisualizer) Clear() *FunctionCallVisualizer {
	fcv.mu.Lock()
	fcv.calls = fcv.calls[:0]
	fcv.mu.Unlock()
	return fcv
}

// CallCount returns the number of calls.
func (fcv *FunctionCallVisualizer) CallCount() int {
	fcv.mu.Lock()
	defer fcv.mu.Unlock()
	return len(fcv.calls)
}

// SetStyle sets the custom style.
func (fcv *FunctionCallVisualizer) SetStyle(s FuncCallVizStyle) *FunctionCallVisualizer {
	fcv.mu.Lock()
	fcv.style = s
	fcv.mu.Unlock()
	return fcv
}

// statusIndexLocked returns 0=running, 1=success, 2=error.
func callStatusIndex(s CallStatus) int {
	switch s {
	case CallRunning:
		return 0
	case CallSuccess:
		return 1
	case CallError:
		return 2
	default:
		return 1
	}
}

// statusIconLocked returns the icon rune for a status.
func callStatusIcon(s CallStatus) rune {
	switch s {
	case CallRunning:
		return '⟳'
	case CallSuccess:
		return '✓'
	case CallError:
		return '✗'
	default:
		return '·'
	}
}

// Measure returns the preferred size.
func (fcv *FunctionCallVisualizer) Measure(cs Constraints) Size {
	fcv.mu.Lock()
	callCount := len(fcv.calls)
	fcv.mu.Unlock()

	w := 60
	h := callCount + 2 // calls + borders
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the function call chain into the buffer.
func (fcv *FunctionCallVisualizer) Paint(buf *buffer.Buffer) {
	fcv.mu.Lock()
	defer fcv.mu.Unlock()

	b := fcv.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 {
		w = 60
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := fcv.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw each call
	for idx, call := range fcv.calls {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		col := x + 1

		// Indentation
		indentStyle := fcv.style.Indent
		for i := 0; i < call.Indent; i++ {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: indentStyle.Fg, Bg: indentStyle.Bg, Flags: indentStyle.Flags, Width: 1})
			col++
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: '│', Fg: indentStyle.Fg, Bg: indentStyle.Bg, Flags: indentStyle.Flags, Width: 1})
			col++
		}

		// Status icon
		statusIdx := callStatusIndex(call.Status)
		statusStyle := fcv.style.Status[statusIdx]
		icon := callStatusIcon(call.Status)
		if col < x+w-1 && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: icon, Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
			col++
		}

		// Name
		nameStyle := fcv.style.Name
		for _, r := range call.Name {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}

		// Args summary (truncated) — iterate string directly, no []rune alloc
		argsStyle := fcv.style.Args
		if col < x+w-1 && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: argsStyle.Fg, Bg: argsStyle.Bg, Flags: argsStyle.Flags, Width: 1})
			col++
		}
		// Truncate args to fit remaining space
		remaining := x + w - 2 - col
		maxArgLen := remaining - 8
		if maxArgLen > 0 {
			argCount := 0
			for _, r := range call.Args {
				if argCount >= maxArgLen {
					break
				}
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: argsStyle.Fg, Bg: argsStyle.Bg, Flags: argsStyle.Flags, Width: 1})
				col++
				argCount++
			}
		}

		// Duration at the right edge — use cached string from AddCall (zero-alloc)
		durStr := call.DurStr
		durLen := len(durStr)
		durStart := x + w - 2 - durLen
		if durStart < col {
			durStart = col
		}
		durStyle := fcv.style.Duration
		durCol := durStart
		for _, r := range durStr {
			if durCol < x+w-1 && durCol < buf.Width {
				buf.SetCell(durCol, rowY, buffer.Cell{Rune: r, Fg: durStyle.Fg, Bg: durStyle.Bg, Flags: durStyle.Flags, Width: 1})
			}
			durCol++
		}
	}
}

// Children returns nil.
func (fcv *FunctionCallVisualizer) Children() []Component { return nil }
