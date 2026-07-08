package component

import (
	"strconv"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ProgressBarMode determines how the ProgressBar renders.
type ProgressBarMode int

const (
	// ProgressDeterminate shows a percentage-filled bar (0–100%).
	ProgressDeterminate ProgressBarMode = iota
	// ProgressIndeterminate shows an animated scanning bar.
	ProgressIndeterminate
)

// ProgressBar is a component that renders a horizontal progress bar.
// In determinate mode, it fills proportionally to the progress value (0–100).
// In indeterminate mode, it shows a scanning segment that moves back and forth.
// The bar color transitions from red → yellow → green as progress increases.
type ProgressBar struct {
	BaseComponent

	mu sync.RWMutex

	progress float64 // 0.0 – 100.0
	label    string
	mode     ProgressBarMode

	// Indeterminate animation state.
	animPos    int // current scan position (0-based column)
	animDir    int // +1 = rightward, -1 = leftward
	animWidth  int // width of the scanning segment

	showPercentage bool
	style          buffer.Style
}

// NewProgressBar creates a determinate ProgressBar at 0%.
func NewProgressBar() *ProgressBar {
	p := &ProgressBar{
		progress:       0,
		mode:           ProgressDeterminate,
		showPercentage: true,
		animDir:        1,
		animWidth:      5,
		style:          buffer.DefaultStyle,
	}
	p.SetID(GenerateID("progress"))
	return p
}

// SetProgress sets the progress percentage (clamped to 0–100).
// Only meaningful in determinate mode.
func (p *ProgressBar) SetProgress(percent float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	p.progress = percent
}

// Progress returns the current progress value (0–100).
func (p *ProgressBar) Progress() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.progress
}

// SetLabel sets an optional text label displayed before the bar.
func (p *ProgressBar) SetLabel(label string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.label = label
}

// Label returns the current label.
func (p *ProgressBar) Label() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.label
}

// SetMode sets the bar mode (determinate or indeterminate).
func (p *ProgressBar) SetMode(mode ProgressBarMode) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mode = mode
	if mode == ProgressIndeterminate {
		p.animPos = 0
		p.animDir = 1
	}
}

// Mode returns the current mode.
func (p *ProgressBar) Mode() ProgressBarMode {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.mode
}

// SetShowPercentage toggles whether the numeric percentage is shown.
func (p *ProgressBar) SetShowPercentage(show bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.showPercentage = show
}

// SetStyle sets the base style for the bar's unfilled portion.
func (p *ProgressBar) SetStyle(s buffer.Style) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.style = s
}

// SetIndeterminateWidth sets the scanning segment width for indeterminate mode.
func (p *ProgressBar) SetIndeterminateWidth(w int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if w < 1 {
		w = 1
	}
	p.animWidth = w
}

// Tick advances the indeterminate animation by one step.
// Call this on a timer (e.g. every 100ms) when in indeterminate mode.
func (p *ProgressBar) Tick() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mode != ProgressIndeterminate {
		return
	}
	w := p.bounds.W
	if w <= 0 {
		return
	}
	p.animPos += p.animDir
	if p.animPos >= w-p.animWidth {
		p.animPos = w - p.animWidth
		p.animDir = -1
	}
	if p.animPos <= 0 {
		p.animPos = 0
		p.animDir = 1
	}
}

// progressColor returns a color that transitions from red → yellow → green
// based on the progress value (0–100).
func progressColor(progress float64) buffer.Color {
	// 0% = red, 50% = yellow, 100% = green
	ratio := progress / 100.0
	var r, g uint8
	if ratio < 0.5 {
		// Red → Yellow: red=255, green increases
		r = 255
		g = uint8(ratio * 2 * 255)
	} else {
		// Yellow → Green: green=255, red decreases
		g = 255
		r = uint8((1.0 - (ratio-0.5)*2) * 255)
	}
	return buffer.RGB(r, g, 0)
}

// Measure returns the desired size: full width, height = 1 (bar) + 1 (label line if present).
func (p *ProgressBar) Measure(cs Constraints) Size {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	h := 1
	if p.label != "" {
		h = 2 // label line + bar line
	}
	return Size{W: w, H: h}
}

// Paint renders the progress bar into the buffer.
func (p *ProgressBar) Paint(buf *buffer.Buffer) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	b := p.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	barY := b.Y
	if p.label != "" {
		// Draw label on the first line.
		labelStyle := p.style
		buf.DrawText(b.X, b.Y, p.label, labelStyle)
		barY = b.Y + 1
	}

	barW := b.W
	if barW <= 0 {
		return
	}

	// Calculate label/percentage width to reserve space.
	availableW := barW
	var percentStr string
	if p.showPercentage && p.mode == ProgressDeterminate {
		percentStr = strings.TrimSpace(formatPercent(p.progress))
		availableW = barW - buffer.StringWidth(percentStr) - 1
		if availableW < 1 {
			availableW = 1
		}
	}

	// Draw the bar.
	if p.mode == ProgressDeterminate {
		p.paintDeterminate(buf, b.X, barY, availableW, percentStr)
	} else {
		p.paintIndeterminate(buf, b.X, barY, availableW)
	}
}

// paintDeterminate draws a filled bar proportional to progress.
func (p *ProgressBar) paintDeterminate(buf *buffer.Buffer, x, y, w int, percentStr string) {
	filledCount := int(p.progress / 100.0 * float64(w))
	if filledCount > w {
		filledCount = w
	}

	fillColor := progressColor(p.progress)
	fillStyle := buffer.Style{Fg: fillColor, Bg: fillColor}
	emptyStyle := buffer.Style{
		Fg: buffer.RGB(60, 60, 60),
		Bg: buffer.RGB(40, 40, 40),
	}

	for i := 0; i < w; i++ {
		if i < filledCount {
			buf.SetCell(x+i, y, buffer.Cell{Rune: '█', Width: 1, Fg: fillStyle.Fg, Bg: fillStyle.Bg})
		} else {
			buf.SetCell(x+i, y, buffer.Cell{Rune: '░', Width: 1, Fg: emptyStyle.Fg, Bg: emptyStyle.Bg})
		}
	}

	// Draw percentage string after the bar.
	if percentStr != "" {
		buf.DrawText(x+w+1, y, percentStr, p.style)
	}
}

// paintIndeterminate draws a scanning segment.
func (p *ProgressBar) paintIndeterminate(buf *buffer.Buffer, x, y, w int) {
	trackStyle := buffer.Style{
		Fg: buffer.RGB(60, 60, 60),
		Bg: buffer.RGB(40, 40, 40),
	}
	scanColor := buffer.RGB(100, 200, 255)

	for i := 0; i < w; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '░', Width: 1, Fg: trackStyle.Fg, Bg: trackStyle.Bg})
	}

	end := p.animPos + p.animWidth
	if end > w {
		end = w
	}
	for i := p.animPos; i < end; i++ {
		if i >= 0 && i < w {
			buf.SetCell(x+i, y, buffer.Cell{Rune: '█', Width: 1, Fg: scanColor, Bg: scanColor})
		}
	}
}

// formatPercent returns a right-justified percentage string like " 45%".
// percentStrings pre-computes "0%" through "100%" to avoid heap allocation
// in the Paint hot path. fmt.Sprintf/strconv.AppendInt + string() all allocate.
var percentStrings [101]string

func init() {
	for i := 0; i <= 100; i++ {
		percentStrings[i] = strconv.Itoa(i) + "%"
	}
}

func formatPercent(progress float64) string {
	pct := int(progress)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return percentStrings[pct]
}
