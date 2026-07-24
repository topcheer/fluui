package component

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ThinkingIndicator displays an animated "AI is thinking" indicator with
// three dots that fill progressively (○○○ → ●○○ → ●●○ → ●●● → repeat).
// This is the classic animation used by ChatGPT, Claude, and other AI
// assistants to signal that a response is being generated.
//
// The indicator is thread-safe and supports manual frame advance for
// testing without real timers.
type ThinkingIndicator struct {
	BaseComponent
	mu sync.Mutex

	label    string
	frames   uint64 // atomic, incremented each AdvanceFrame
	running  atomic.Bool
	stopCh   chan struct{}
	style    ThinkingStyle
}

// ThinkingStyle controls the visual appearance of the indicator.
type ThinkingStyle struct {
	// DotChar is the filled dot character.
	DotChar string
	// EmptyChar is the unfilled dot character.
	EmptyChar string
	// Spacing between dots.
	Spacing string
	// Label color from theme; defaults to Muted.
	UseAccent bool // if true, use accent color instead of muted
}

// DefaultThinkingStyle returns the standard thinking indicator style.
func DefaultThinkingStyle() ThinkingStyle {
	return ThinkingStyle{
		DotChar:   "●",
		EmptyChar: "○",
		Spacing:   " ",
	}
}

// NewThinkingIndicator creates a thinking indicator with an optional label
// (e.g., "Thinking" or "Generating response"). The indicator starts in a
// stopped state.
func NewThinkingIndicator(label string) *ThinkingIndicator {
	return &ThinkingIndicator{
		BaseComponent: BaseComponent{id: GenerateID("thinking")},
		label:         label,
		style:         DefaultThinkingStyle(),
		stopCh:        make(chan struct{}),
	}
}

// SetLabel changes the text shown next to the dots.
func (t *ThinkingIndicator) SetLabel(s string) {
	t.mu.Lock()
	t.label = s
	t.mu.Unlock()
}

// Label returns the current label.
func (t *ThinkingIndicator) Label() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.label
}

// SetStyle changes the visual style.
func (t *ThinkingIndicator) SetStyle(s ThinkingStyle) {
	t.mu.Lock()
	t.style = s
	t.mu.Unlock()
}

// Start begins the animation cycle at the given interval (e.g., 300ms).
// Calling Start when already running is a no-op.
func (t *ThinkingIndicator) Start(interval time.Duration) {
	if t.running.Load() {
		return
	}
	t.running.Store(true)
	t.mu.Lock()
	t.stopCh = make(chan struct{})
	stopCh := t.stopCh
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				atomic.AddUint64(&t.frames, 1)
			}
		}
	}()
}

// Stop halts the animation.
func (t *ThinkingIndicator) Stop() {
	if !t.running.Load() {
		return
	}
	t.running.Store(false)
	t.mu.Lock()
	if t.stopCh != nil {
		select {
		case <-t.stopCh:
			// already closed
		default:
			close(t.stopCh)
		}
	}
	t.mu.Unlock()
}

// IsRunning returns whether the animation is active.
func (t *ThinkingIndicator) IsRunning() bool {
	return t.running.Load()
}

// AdvanceFrame manually increments the animation frame (for testing).
func (t *ThinkingIndicator) AdvanceFrame() {
	atomic.AddUint64(&t.frames, 1)
}

// FrameIndex returns the current animation frame index (0-3).
func (t *ThinkingIndicator) FrameIndex() int {
	f := atomic.LoadUint64(&t.frames) % 4
	return int(f)
}

// Measure computes the desired size.
func (t *ThinkingIndicator) Measure(cs Constraints) Size {
	t.mu.Lock()
	label := t.label
	style := t.style
	t.mu.Unlock()

	// Width = len(label) + spacing + 3 dots + 2 spacings
	w := 0
	if label != "" {
		w += len([]rune(label)) + 1 // +1 for space after label
	}
	dotW := len([]rune(style.DotChar))*3 + 2 // 3 dots + 2 spaces
	w += dotW
	if w < 1 {
		w = 1
	}
	return Size{W: w, H: 1}
}

// Paint renders the thinking indicator.
func (t *ThinkingIndicator) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	label := t.label
	style := t.style
	t.mu.Unlock()

	th := theme.Get()
	fg := th.Muted
	if style.UseAccent {
		fg = th.Accent
	}
	cellStyle := buffer.Style{Fg: fg}

	frame := t.FrameIndex()
	b := t.Bounds()
	x := b.X
	y := b.Y

	// Draw label
	if label != "" {
		x += buf.DrawText(x, y, label+" ", cellStyle)
	}

	// Draw three dots: first N are filled, rest are empty
	for i := 0; i < 3; i++ {
		ch := style.EmptyChar
		if i < frame+1 && frame < 3 {
			ch = style.DotChar
		}
		// On frame 3 (all filled), show all filled
		if frame == 3 {
			ch = style.DotChar
		}
		buf.DrawText(x, y, ch, cellStyle)
		x += len([]rune(ch))
		if i < 2 {
			buf.DrawText(x, y, style.Spacing, cellStyle)
			x += len([]rune(style.Spacing))
		}
	}
}
