package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// TogglePill renders a compact on/off pill switch.
// Visually distinct from Switch: uses [ON ]/[OFF] pill format with
// color-coded background. Useful for compact settings panels.
//
// Thread-safe.
type TogglePill struct {
	BaseComponent
	mu      sync.RWMutex
	on      bool
	label   string
	onText  string
	offText string
}

// NewTogglePill creates a toggle pill with initial state.
func NewTogglePill(on bool) *TogglePill {
	return &TogglePill{
		BaseComponent: BaseComponent{id: GenerateID("toggle")},
		on:            on,
		onText:        "ON",
		offText:       "OFF",
	}
}

func (t *TogglePill) IsOn() bool { t.mu.RLock(); defer t.mu.RUnlock(); return t.on }
func (t *TogglePill) SetOn(b bool) { t.mu.Lock(); defer t.mu.Unlock(); t.on = b }
func (t *TogglePill) Toggle() { t.mu.Lock(); defer t.mu.Unlock(); t.on = !t.on }

func (t *TogglePill) Label() string { t.mu.RLock(); defer t.mu.RUnlock(); return t.label }
func (t *TogglePill) SetLabel(s string) { t.mu.Lock(); defer t.mu.Unlock(); t.label = s }

func (t *TogglePill) OnText() string { t.mu.RLock(); defer t.mu.RUnlock(); return t.onText }
func (t *TogglePill) SetOnText(s string) { t.mu.Lock(); defer t.mu.Unlock(); t.onText = s }

func (t *TogglePill) OffText() string { t.mu.RLock(); defer t.mu.RUnlock(); return t.offText }
func (t *TogglePill) SetOffText(s string) { t.mu.Lock(); defer t.mu.Unlock(); t.offText = s }

// Measure returns the preferred size.
func (t *TogglePill) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()
	w := 5 // "[ON ]" or "[OFF]"
	if t.label != "" { w += len(t.label) + 1 }
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

// Paint draws the toggle pill. Zero allocations.
func (t *TogglePill) Paint(buf *buffer.Buffer) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	bounds := t.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	x := bounds.X
	y := bounds.Y

	// Label
	if t.label != "" {
		muted := buffer.Style{Fg: tt.Fg}
		x = buf.DrawText(x, y, t.label+" ", muted)
	}

	// Pill: [ON ] or [OFF]
	var text string
	var fg buffer.Color
	if t.on {
		text = "[" + t.onText + "]"
		fg = tt.Success
	} else {
		text = "[" + t.offText + "]"
		fg = tt.Muted
	}
	style := buffer.Style{Fg: fg, Flags: buffer.Bold}
	buf.DrawText(x, y, text, style)
}
