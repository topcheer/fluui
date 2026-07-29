package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelSwitcher: AI Model Selection Display ───
//
// ModelSwitcher renders a compact model selector showing the currently
// active model with an indicator arrow. Supports cycling through a list
// of models with SetActive to change the selection.
//
// Usage:
//
//	ms := NewModelSwitcher()
//	ms.SetModels("gpt-4", "claude-3", "gemini-pro")
//	ms.SetActive(1) // claude-3
//	ms.Paint(buf)

// ModelSwitcherStyle holds styling.
type ModelSwitcherStyle struct {
	Active   buffer.Style
	Inactive buffer.Style
	Arrow    buffer.Style
	Bracket  buffer.Style
}

// DefaultModelSwitcherStyle returns defaults.
func DefaultModelSwitcherStyle() ModelSwitcherStyle {
	return ModelSwitcherStyle{
		Active:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Inactive: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Arrow:    buffer.Style{Fg: buffer.RGB(251, 191, 36)},
		Bracket:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const modelSwitcherMax = 8

// ModelSwitcher renders a model selector.
type ModelSwitcher struct {
	BaseComponent
	mu sync.Mutex

	models  [modelSwitcherMax]string
	count   int
	active  int
	style   ModelSwitcherStyle
	// cached
	counterStr string
}

// NewModelSwitcher creates a ModelSwitcher.
func NewModelSwitcher() *ModelSwitcher {
	ms := &ModelSwitcher{style: DefaultModelSwitcherStyle()}
	ms.SetID(GenerateID("modelsw"))
	ms.recomputeCounterLocked()
	return ms
}

func (ms *ModelSwitcher) recomputeCounterLocked() {
	ms.counterStr = "(" + itoa(ms.active+1) + "/" + itoa(ms.count) + ")"
}

// SetModels sets the list of available models.
func (ms *ModelSwitcher) SetModels(models ...string) *ModelSwitcher {
	ms.mu.Lock()
	ms.count = 0
	for _, m := range models {
		if ms.count >= modelSwitcherMax { break }
		ms.models[ms.count] = m
		ms.count++
	}
	if ms.active >= ms.count { ms.active = 0 }
	ms.recomputeCounterLocked()
	ms.mu.Unlock()
	return ms
}

// SetActive sets the active model index.
func (ms *ModelSwitcher) SetActive(idx int) *ModelSwitcher {
	ms.mu.Lock()
	if ms.count == 0 {
		ms.active = 0
	} else if idx < 0 {
		ms.active = 0
	} else if idx >= ms.count {
		ms.active = ms.count - 1
	} else {
		ms.active = idx
	}
	ms.recomputeCounterLocked()
	ms.mu.Unlock()
	return ms
}

// ActiveIndex returns the active model index.
func (ms *ModelSwitcher) ActiveIndex() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.active
}

// CycleNext advances to the next model (wraps around).
func (ms *ModelSwitcher) CycleNext() *ModelSwitcher {
	ms.mu.Lock()
	if ms.count > 0 {
		ms.active = (ms.active + 1) % ms.count
	}
	ms.recomputeCounterLocked()
	ms.mu.Unlock()
	return ms
}

// SetStyle sets custom style.
func (ms *ModelSwitcher) SetStyle(s ModelSwitcherStyle) *ModelSwitcher {
	ms.mu.Lock()
	ms.style = s
	ms.mu.Unlock()
	return ms
}

// Measure returns preferred size.
func (ms *ModelSwitcher) Measure(cs Constraints) Size {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	maxLen := 8
	for i := 0; i < ms.count; i++ {
		l := len(ms.models[i])
		if l > maxLen { maxLen = l }
	}
	w := maxLen + 4 // arrow + space + model + space
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the model switcher.
func (ms *ModelSwitcher) Paint(buf *buffer.Buffer) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	b := ms.Bounds()
	x, y := b.X, b.Y

	if ms.count == 0 { return }

	activeStyle := ms.style.Active
	arrowStyle := ms.style.Arrow
	bracketStyle := ms.style.Bracket

	col := x
	// Arrow indicator
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '▶', Fg: arrowStyle.Fg, Bg: arrowStyle.Bg, Flags: arrowStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
		col++
	}

	// Active model name
	activeModel := ms.models[ms.active]
	for _, r := range activeModel {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
		col++
	}

	// Counter: (n/total)
	counter := ms.counterStr
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
	for _, r := range counter {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ms *ModelSwitcher) Children() []Component { return nil }
