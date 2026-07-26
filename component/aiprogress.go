package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// AIPhase represents a stage in the LLM processing pipeline.
type AIPhase int

const (
	AIPhaseIdle AIPhase = iota
	AIPhaseThinking
	AIPhaseAnalyzing
	AIPhaseGenerating
	AIPhaseReviewing
	AIPhaseComplete
	AIPhaseError
)

var phaseLabels = [...]string{
	"Idle", "Thinking...", "Analyzing...", "Generating...", "Reviewing...", "Complete", "Error",
}

var phaseIcons = [...]string{
	"\u25cb", // ○ idle
	"\U0001F914", // 🤔 thinking
	"\U0001F50D", // 🔍 analyzing
	"\u270F", // ✏️ generating
	"\U0001F50E", // 🔎 reviewing
	"\u2714", // ✔️ complete
	"\u2716", // ✖️ error
}

var phaseSpinFrames = [...][]rune{
	{'|'},                                  // idle (static)
	{'\u280b', '\u2819', '\u2839', '\u2838', '\u283c', '\u2834', '\u2826', '\u2827', '\u2807', '\u280f'}, // thinking (braille)
	{'\u280b', '\u2819', '\u2839', '\u2838'}, // analyzing
	{46, 46, 46},                       // generating
	{'\u280b', '\u2827', '\u2807'},           // reviewing
	{},                                       // complete (no spin)
	{},                                       // error (no spin)
}

// AIProgress renders an animated LLM processing status indicator.
// Shows the current phase (thinking, analyzing, generating), an animated
// spinner, elapsed label, and optional progress percentage.
//
// Thread-safe.
type AIProgress struct {
	BaseComponent
	mu        sync.RWMutex
	phase     AIPhase
	frame     int   // spinner frame counter
	showLabel bool
	showSpin  bool
	pct       float64 // optional progress 0-1
}

// NewAIProgress creates an idle AI progress indicator.
func NewAIProgress() *AIProgress {
	return &AIProgress{
		BaseComponent: BaseComponent{id: GenerateID("aiprogress")},
		phase:         AIPhaseIdle,
		showLabel:     true,
		showSpin:      true,
	}
}

// Phase returns the current phase.
func (a *AIProgress) Phase() AIPhase { a.mu.RLock(); defer a.mu.RUnlock(); return a.phase }

// SetPhase updates the processing phase.
func (a *AIProgress) SetPhase(p AIPhase) { a.mu.Lock(); defer a.mu.Unlock(); a.phase = p; a.frame = 0 }

// PhaseLabel returns the human-readable label for the current phase.
func (a *AIProgress) PhaseLabel() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return phaseLabels[a.phase]
}

// ShowLabel returns whether the label is displayed.
func (a *AIProgress) ShowLabel() bool { a.mu.RLock(); defer a.mu.RUnlock(); return a.showLabel }

// SetShowLabel toggles label display.
func (a *AIProgress) SetShowLabel(b bool) { a.mu.Lock(); defer a.mu.Unlock(); a.showLabel = b }

// ShowSpin returns whether the spinner is displayed.
func (a *AIProgress) ShowSpin() bool { a.mu.RLock(); defer a.mu.RUnlock(); return a.showSpin }

// SetShowSpin toggles spinner display.
func (a *AIProgress) SetShowSpin(b bool) { a.mu.Lock(); defer a.mu.Unlock(); a.showSpin = b }

// Progress returns the optional progress fraction (0.0-1.0). -1 = not shown.
func (a *AIProgress) Progress() float64 { a.mu.RLock(); defer a.mu.RUnlock(); return a.pct }

// SetProgress sets an optional progress fraction. Pass -1 to hide.
func (a *AIProgress) SetProgress(p float64) { a.mu.Lock(); defer a.mu.Unlock(); a.pct = p }

// Tick advances the spinner animation. Call on a timer.
func (a *AIProgress) Tick() {
	a.mu.Lock()
	defer a.mu.Unlock()
	frames := phaseSpinFrames[a.phase]
	if len(frames) > 0 {
		a.frame = (a.frame + 1) % len(frames)
	}
}

// IsBusy returns true when the phase indicates active processing.
func (a *AIProgress) IsBusy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.phase >= AIPhaseThinking && a.phase <= AIPhaseReviewing
}

func (a *AIProgress) resolveColorLocked() buffer.Color {
	t := theme.Get()
	switch a.phase {
	case AIPhaseComplete:
		return t.Success
	case AIPhaseError:
		return t.Error
	case AIPhaseIdle:
		return t.Muted
	default:
		return t.Accent
	}
}

// Measure returns the preferred size.
func (a *AIProgress) Measure(cs Constraints) Size {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w := 2 // spinner + space
	if a.showLabel {
		w += len(phaseLabels[a.phase])
	}
	if a.pct >= 0 {
		w += 5 // " NN%"
	}
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

// Paint draws the AI progress indicator. Zero allocations.
func (a *AIProgress) Paint(buf *buffer.Buffer) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	bounds := a.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	t := theme.Get()
	color := a.resolveColorLocked()
	style := buffer.Style{Fg: color, Flags: buffer.Bold}
	muted := buffer.Style{Fg: t.Muted}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Spinner
	if a.showSpin {
		frames := phaseSpinFrames[a.phase]
		if len(frames) > 0 {
			r := frames[a.frame%len(frames)]
			// Handle multi-rune frames (like "...")
			if r < 0x80 && r == '.' {
				// Generating: show "..." variant
				dots := int(a.frame%3) + 1
				for i := 0; i < dots && x < maxX; i++ {
					buf.SetCell(x, y, buffer.Cell{Rune: '.', Width: 1, Fg: color, Flags: buffer.Bold})
					x++
				}
			} else {
				buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: color, Flags: buffer.Bold})
				x++
			}
		} else {
			// Static icon for complete/error
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2714', Width: 1, Fg: color, Flags: buffer.Bold})
			x++
		}
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: muted.Fg})
			x++
		}
	}

	// Label
	if a.showLabel && x < maxX {
		x = buf.DrawText(x, y, phaseLabels[a.phase], style)
	}

	// Progress percentage
	if a.pct >= 0 && x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: muted.Fg})
		x++
		if x < maxX {
			// Write percentage via stack buffer
			pctVal := int(a.pct * 100)
			if pctVal < 0 { pctVal = 0 }
			if pctVal > 100 { pctVal = 100 }
			var pb [8]byte
			pbs := pb[:0]
			// Simple itoa
			if pctVal == 0 {
				pbs = append(pbs, '0')
			} else {
				tmp := pctVal
				var digits [4]byte
				n := 0
				for tmp > 0 {
					digits[n] = byte('0' + tmp%10)
					tmp /= 10
					n++
				}
				for i := n - 1; i >= 0; i-- {
					pbs = append(pbs, digits[i])
				}
			}
			pbs = append(pbs, '%')
			buf.DrawBytes(x, y, pbs, style)
		}
	}
}
