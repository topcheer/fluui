package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// LoadingIndicatorStyle holds visual styling.
type LoadingIndicatorStyle struct {
	Fg       buffer.Color
	BarBg    buffer.Color
	BarFg    buffer.Color
	TextFg   buffer.Color
}

// DefaultLoadingIndicatorStyle returns a blue-themed style.
func DefaultLoadingIndicatorStyle() LoadingIndicatorStyle {
	return LoadingIndicatorStyle{
		Fg:     buffer.NamedColor(buffer.NamedCyan),
		BarBg:  buffer.NamedColor(buffer.NamedBrightBlack),
		BarFg:  buffer.NamedColor(buffer.NamedCyan),
		TextFg: buffer.NamedColor(buffer.NamedWhite),
	}
}

// LoadingIndicator displays an animated loading indicator with optional
// indeterminate progress bar. Unlike Spinner (which shows a small icon),
// LoadingIndicator shows a full animated bar with text.
type LoadingIndicator struct {
	BaseComponent

	text     string
	style    LoadingIndicatorStyle
	frame    int
	maxFrame int
	running  bool
	stopCh   chan struct{}
	ticker   *time.Ticker

	mu sync.RWMutex
}

// NewLoadingIndicator creates a loading indicator with optional label.
func NewLoadingIndicator(text string) *LoadingIndicator {
	return &LoadingIndicator{
		text:     text,
		style:    DefaultLoadingIndicatorStyle(),
		maxFrame: 20,
	}
}

// SetText sets the loading message.
func (l *LoadingIndicator) SetText(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.text = s
}

// Text returns the current loading message.
func (l *LoadingIndicator) Text() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.text
}

// SetStyle sets the visual style.
func (l *LoadingIndicator) SetStyle(s LoadingIndicatorStyle) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.style = s
}

// Start begins the animation.
func (l *LoadingIndicator) Start() {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return
	}
	l.running = true
	l.stopCh = make(chan struct{})
	l.ticker = time.NewTicker(80 * time.Millisecond)
	stopCh := l.stopCh
	ticker := l.ticker
	l.mu.Unlock()

	go func() {
		for {
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				l.mu.Lock()
				l.frame++
				if l.frame >= l.maxFrame {
					l.frame = 0
				}
				l.mu.Unlock()
			}
		}
	}()
}

// Stop halts the animation.
func (l *LoadingIndicator) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.running {
		return
	}
	l.running = false
	close(l.stopCh)
	l.ticker.Stop()
}

// IsRunning returns whether the indicator is animating.
func (l *LoadingIndicator) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.running
}

// AdvanceFrame manually advances the animation frame (for testing).
func (l *LoadingIndicator) AdvanceFrame() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.frame++
	if l.frame >= l.maxFrame {
		l.frame = 0
	}
}

// Frame returns the current frame index.
func (l *LoadingIndicator) Frame() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.frame
}

// Measure returns the desired size.
func (l *LoadingIndicator) Measure(cs Constraints) Size {
	w := 30
	if cs.MaxWidth > 0 && cs.MaxWidth < w {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 3}
}

// Paint renders the loading indicator.
func (l *LoadingIndicator) Paint(buf *buffer.Buffer) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := l.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w < 10 {
		w = 10
	}

	// Text line
	text := l.text
	if text == "" {
		text = "Loading..."
	}
	if len(text) > w {
		text = text[:w]
	}
	for i, r := range text {
		if x+i < bounds.X+bounds.W {
			buf.SetCell(x+i, y, buffer.Cell{Rune: r, Width: 1, Fg: l.style.TextFg})
		}
	}

	// Animated progress bar (indeterminate)
	barY := y + 1
	barW := w
	if barW > 3 {
		// Draw background
		for i := 0; i < barW; i++ {
			buf.SetCell(x+i, barY, buffer.Cell{Rune: '░', Width: 1, Fg: l.style.BarBg})
		}

		// Draw animated segment
		segLen := barW / 4
		if segLen < 2 {
			segLen = 2
		}
		pos := l.frame % (barW - segLen + 1)
		for i := 0; i < segLen; i++ {
			if pos+i < barW {
				buf.SetCell(x+pos+i, barY, buffer.Cell{Rune: '█', Width: 1, Fg: l.style.BarFg})
			}
		}
	}

	// Animated dots line
	dotsY := y + 2
	numDots := (l.frame / 2) % 4
	dotStr := "Loading"
	for i, r := range dotStr {
		if x+i < bounds.X+bounds.W {
			buf.SetCell(x+i, dotsY, buffer.Cell{Rune: r, Width: 1, Fg: l.style.Fg})
		}
	}
	dx := x + len(dotStr)
	for i := 0; i < 3; i++ {
		var r rune = ' '
		if i < numDots {
			r = '.'
		}
		if dx+i < bounds.X+bounds.W {
			buf.SetCell(dx+i, dotsY, buffer.Cell{Rune: r, Width: 1, Fg: l.style.Fg})
		}
	}
}

// HandleKey is a no-op.
func (l *LoadingIndicator) HandleKey(_ *term.KeyEvent) bool { return false }

// Children returns nil.
func (l *LoadingIndicator) Children() []Component { return nil }
