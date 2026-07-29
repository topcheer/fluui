package component

import (
	"strconv"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ResponseInspector: AI Response Metadata Inspector ───
//
// ResponseInspector renders a compact panel showing AI response metadata:
// model name, latency, token counts, finish reason, and temperature.
// Useful in debugging tools and AI observability dashboards.
//
// Usage:
//
//	ri := NewResponseInspector()
//	ri.SetModel("gpt-4o")
//	ri.SetLatency(450 * time.Millisecond)
//	ri.SetTokens(120, 350)
//	ri.SetFinishReason("stop")
//	ri.SetTemperature(0.7)
//	ri.Paint(buf)

// ResponseInspectorStyle holds styling for ResponseInspector.
type ResponseInspectorStyle struct {
	Label  buffer.Style
	Value  buffer.Style
	Header buffer.Style
	Border buffer.Style
}

// DefaultResponseInspectorStyle returns a sensible default style.
func DefaultResponseInspectorStyle() ResponseInspectorStyle {
	label := buffer.Style{Fg: buffer.RGB(100, 116, 139)} // slate-500
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240)} // slate-200
	header := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)} // slate-600
	return ResponseInspectorStyle{
		Label:  label,
		Value:  value,
		Header: header,
		Border: border,
	}
}

// FinishReason describes why an AI response completed.
type ResponseFinishReason string

const (
	FinishStop          ResponseFinishReason = "stop"
	FinishLength        ResponseFinishReason = "length"
	FinishToolCalls     ResponseFinishReason = "tool_calls"
	FinishContentFilter ResponseFinishReason = "content_filter"
)

// ResponseInspector displays AI response metadata in a compact panel.
type ResponseInspector struct {
	BaseComponent
	mu sync.Mutex

	model        string
	latency      time.Duration
	inputTokens  int
	outputTokens int
	finishReason ResponseFinishReason
	temperature  float64
	tempStr      string // cached formatted temperature
	latencyStr   string // cached formatted latency
	tokenStr     string // cached formatted tokens

	style ResponseInspectorStyle
}

// NewResponseInspector creates a ResponseInspector with default styling.
func NewResponseInspector() *ResponseInspector {
	ri := &ResponseInspector{
		style: DefaultResponseInspectorStyle(),
	}
	ri.SetID(GenerateID("inspector"))
	return ri
}

// SetModel sets the model name displayed.
func (ri *ResponseInspector) SetModel(m string) *ResponseInspector {
	ri.mu.Lock()
	ri.model = m
	ri.mu.Unlock()
	return ri
}

// Model returns the current model name.
func (ri *ResponseInspector) Model() string {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.model
}

// SetLatency sets the response latency.
func (ri *ResponseInspector) SetLatency(d time.Duration) *ResponseInspector {
	ri.mu.Lock()
	ri.latency = d
	ri.latencyStr = formatInspectorDuration(d)
	ri.mu.Unlock()
	return ri
}

// Latency returns the response latency.
func (ri *ResponseInspector) Latency() time.Duration {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.latency
}

// SetTokens sets input and output token counts.
func (ri *ResponseInspector) SetTokens(input, output int) *ResponseInspector {
	ri.mu.Lock()
	ri.inputTokens = input
	ri.outputTokens = output
	ri.tokenStr = " " + itoa(input) + " in / " + itoa(output) + " out"
	ri.mu.Unlock()
	return ri
}

// InputTokens returns input token count.
func (ri *ResponseInspector) InputTokens() int {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.inputTokens
}

// OutputTokens returns output token count.
func (ri *ResponseInspector) OutputTokens() int {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.outputTokens
}

// TotalTokens returns combined token count.
func (ri *ResponseInspector) TotalTokens() int {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.inputTokens + ri.outputTokens
}

// SetFinishReason sets why the response completed.
func (ri *ResponseInspector) SetFinishReason(r ResponseFinishReason) *ResponseInspector {
	ri.mu.Lock()
	ri.finishReason = r
	ri.mu.Unlock()
	return ri
}

// FinishReason returns the finish reason.
func (ri *ResponseInspector) FinishReason() ResponseFinishReason {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.finishReason
}

// SetTemperature sets the sampling temperature.
func (ri *ResponseInspector) SetTemperature(t float64) *ResponseInspector {
	ri.mu.Lock()
	ri.temperature = t
	ri.tempStr = strconv.FormatFloat(t, 'f', 1, 64)
	ri.mu.Unlock()
	return ri
}

// Temperature returns the sampling temperature.
func (ri *ResponseInspector) Temperature() float64 {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.temperature
}

// SetStyle sets the custom style.
func (ri *ResponseInspector) SetStyle(s ResponseInspectorStyle) *ResponseInspector {
	ri.mu.Lock()
	ri.style = s
	ri.mu.Unlock()
	return ri
}

// formatInspectorDuration converts latency to a human-readable string without fmt.Sprintf.
func formatInspectorDuration(d time.Duration) string {
	if d < time.Microsecond {
		return itoa(int(d.Nanoseconds())) + "ns"
	}
	if d < time.Millisecond {
		return itoa(int(d.Microseconds())) + "us"
	}
	if d < time.Second {
		return itoa(int(d.Milliseconds())) + "ms"
	}
	return itoa(int(d.Seconds())) + "s"
}

// Measure returns the preferred size for the inspector panel.
func (ri *ResponseInspector) Measure(cs Constraints) Size {
	w := 34
	h := 8
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the inspector panel into the buffer.
func (ri *ResponseInspector) Paint(buf *buffer.Buffer) {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	b := ri.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 {
		w = 34
	}
	if h < 8 {
		h = 8
	}

	// Draw border
	bs := ri.style.Border
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

	// Header
	headerText := " Response Inspector"
	for i, r := range headerText {
		if x+1+i < buf.Width && x+1+i < x+w-1 {
			buf.SetCell(x+1+i, y+1, buffer.Cell{Rune: r, Fg: ri.style.Header.Fg, Bg: ri.style.Header.Bg, Flags: ri.style.Header.Flags, Width: 1})
		}
	}

	// Metadata rows
	type metaRow struct {
		label string
		value string
	}
	rows := [...]metaRow{
		{"Model:", " " + ri.model},
		{"Latency:", " " + ri.latencyStr},
		{"Tokens:", ri.tokenStr},
		{"Finish:", " " + string(ri.finishReason)},
		{"Temp:", " " + ri.tempStr},
	}

	for idx, row := range rows {
		rowY := y + 2 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}
		col := x + 2
		for _, r := range row.label {
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: ri.style.Label.Fg, Bg: ri.style.Label.Bg, Flags: ri.style.Label.Flags, Width: 1})
			}
			col++
		}
		for _, r := range row.value {
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: ri.style.Value.Fg, Bg: ri.style.Value.Bg, Flags: ri.style.Value.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil — ResponseInspector has no children.
func (ri *ResponseInspector) Children() []Component { return nil }
