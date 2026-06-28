package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ─── Wizard Step ──────────────────────────────────────────────────────────

// WizardStep represents a single step in a multi-step wizard.
type WizardStep struct {
	ID          string   // unique identifier
	Title       string   // heading displayed at top
	Description string   // optional sub-heading
	Content     Component // optional body component
	Skippable   bool     // can this step be skipped
	OnEnter     func(w *Wizard) error // called when step becomes active
	OnLeave     func(w *Wizard) error // called when leaving (Back/Next), error blocks
}

// NewWizardStep creates a step with the given ID and title.
func NewWizardStep(id, title string) *WizardStep {
	return &WizardStep{ID: id, Title: title}
}

// SetDescription sets the step description.
func (s *WizardStep) SetDescription(desc string) *WizardStep {
	s.Description = desc
	return s
}

// SetContent sets the step body content component.
func (s *WizardStep) SetContent(c Component) *WizardStep {
	s.Content = c
	return s
}

// SetSkippable marks the step as skippable.
func (s *WizardStep) SetSkippable(v bool) *WizardStep {
	s.Skippable = v
	return s
}

// SetOnEnter sets the enter callback.
func (s *WizardStep) SetOnEnter(fn func(w *Wizard) error) *WizardStep {
	s.OnEnter = fn
	return s
}

// SetOnLeave sets the leave callback.
func (s *WizardStep) SetOnLeave(fn func(w *Wizard) error) *WizardStep {
	s.OnLeave = fn
	return s
}

// ─── Wizard Style ─────────────────────────────────────────────────────────

// WizardStyle holds the visual styling for the wizard.
type WizardStyle struct {
	Border         buffer.Style
	Title          buffer.Style
	Description    buffer.Style
	StepActive     buffer.Style
	StepDone       buffer.Style
	StepPending    buffer.Style
	ButtonNormal   buffer.Style
	ButtonSelected buffer.Style
	Help           buffer.Style
}

// DefaultWizardStyle returns a default style based on the active theme.
func DefaultWizardStyle() WizardStyle {
	t := theme.Get()
	return WizardStyle{
		Border:         buffer.Style{Fg: t.Border, Bg: t.Bg},
		Title:          buffer.Style{Fg: t.Accent, Bg: t.Bg, Flags: buffer.Bold},
		Description:    buffer.Style{Fg: t.Muted, Bg: t.Bg},
		StepActive:     buffer.Style{Fg: t.Accent, Bg: t.Bg, Flags: buffer.Bold},
		StepDone:       buffer.Style{Fg: t.Success, Bg: t.Bg},
		StepPending:    buffer.Style{Fg: t.Muted, Bg: t.Bg},
		ButtonNormal:   buffer.Style{Fg: t.Fg, Bg: t.Bg},
		ButtonSelected: buffer.Style{Fg: t.Bg, Bg: t.Accent, Flags: buffer.Bold},
		Help:           buffer.Style{Fg: t.Muted, Bg: t.Bg},
	}
}

// ─── Wizard Button ────────────────────────────────────────────────────────

// WizardButton identifies a button in the wizard navigation bar.
type WizardButton int

const (
	WizardBtnBack WizardButton = iota
	WizardBtnNext
	WizardBtnFinish
	WizardBtnCancel
)

// ButtonLabel returns a human-readable label for the button.
func (b WizardButton) ButtonLabel() string {
	switch b {
	case WizardBtnBack:
		return "Back"
	case WizardBtnNext:
		return "Next"
	case WizardBtnFinish:
		return "Finish"
	case WizardBtnCancel:
		return "Cancel"
	default:
		return "?"
	}
}

// ─── Wizard ───────────────────────────────────────────────────────────────

// Wizard is a multi-step navigation component with progress indicators.
type Wizard struct {
	BaseComponent
	mu sync.RWMutex

	steps       []*WizardStep
	current     int
	style       WizardStyle
	selected    WizardButton
	width       int
	height      int
	cancelled   bool
	completed   bool
	buttonOrder []WizardButton

	OnFinish      func(w *Wizard)
	OnCancel      func(w *Wizard)
	OnStepChange  func(w *Wizard, stepIdx int)
}

// NewWizard creates a wizard with the given steps.
func NewWizard(steps []*WizardStep) *Wizard {
	w := &Wizard{
		steps:    steps,
		current:  0,
		style:    DefaultWizardStyle(),
		selected: WizardBtnNext,
		width:    60,
		height:   20,
	}
	w.SetID(GenerateID("wizard"))
	w.recomputeButtonsLocked()
	return w
}

// recomputeButtonsLocked updates the button order based on the current step.
// Must be called with write lock held.
func (w *Wizard) recomputeButtonsLocked() {
	w.buttonOrder = w.buttonOrder[:0]
	if w.current > 0 {
		w.buttonOrder = append(w.buttonOrder, WizardBtnBack)
	}
	if w.current < len(w.steps)-1 {
		w.buttonOrder = append(w.buttonOrder, WizardBtnNext)
	} else {
		w.buttonOrder = append(w.buttonOrder, WizardBtnFinish)
	}
	w.buttonOrder = append(w.buttonOrder, WizardBtnCancel)
}

// ─── Step Management ──────────────────────────────────────────────────────

// Steps returns a copy of the steps slice.
func (w *Wizard) Steps() []*WizardStep {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]*WizardStep, len(w.steps))
	copy(out, w.steps)
	return out
}

// StepCount returns the number of steps.
func (w *Wizard) StepCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.steps)
}

// CurrentStepIndex returns the zero-based index of the active step.
func (w *Wizard) CurrentStepIndex() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.current
}

// CurrentStep returns the active step, or nil if there are no steps.
func (w *Wizard) CurrentStep() *WizardStep {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.current < 0 || w.current >= len(w.steps) {
		return nil
	}
	return w.steps[w.current]
}

// SetCurrentStep moves to the given step index. Returns an error if out of range.
func (w *Wizard) SetCurrentStep(idx int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if idx < 0 || idx >= len(w.steps) {
		return fmt.Errorf("step index %d out of range [0, %d)", idx, len(w.steps))
	}
	w.current = idx
	w.recomputeButtonsLocked()
	return nil
}

// IsFirstStep returns true if the current step is the first one.
func (w *Wizard) IsFirstStep() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.current == 0
}

// IsLastStep returns true if the current step is the last one.
func (w *Wizard) IsLastStep() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.current == len(w.steps)-1
}

// ─── Navigation ───────────────────────────────────────────────────────────

// Next advances to the next step. On the last step it finishes the wizard.
func (w *Wizard) Next() error {
	w.mu.Lock()

	// Run OnLeave for current step
	if w.current < len(w.steps) {
		step := w.steps[w.current]
		if step != nil && step.OnLeave != nil {
			w.mu.Unlock()
			if err := step.OnLeave(w); err != nil {
				return err
			}
			w.mu.Lock()
		}
	}

	if w.current >= len(w.steps)-1 {
		w.completed = true
		cb := w.OnFinish
		w.mu.Unlock()
		if cb != nil {
			cb(w)
		}
		return nil
	}

	w.current++
	w.recomputeButtonsLocked()
	stepIdx := w.current
	newStep := w.steps[w.current]
	w.mu.Unlock()

	// Run OnEnter for new step
	if newStep != nil && newStep.OnEnter != nil {
		if err := newStep.OnEnter(w); err != nil {
			return err
		}
	}

	w.fireStepChange(stepIdx)
	return nil
}

// Back moves to the previous step. Returns an error if already on the first step.
func (w *Wizard) Back() error {
	w.mu.Lock()

	if w.current <= 0 {
		w.mu.Unlock()
		return fmt.Errorf("already on first step")
	}

	// Run OnLeave for current step
	step := w.steps[w.current]
	if step != nil && step.OnLeave != nil {
		w.mu.Unlock()
		if err := step.OnLeave(w); err != nil {
			return err
		}
		w.mu.Lock()
	}

	w.current--
	w.recomputeButtonsLocked()
	stepIdx := w.current
	newStep := w.steps[w.current]
	w.mu.Unlock()

	// Run OnEnter
	if newStep != nil && newStep.OnEnter != nil {
		if err := newStep.OnEnter(w); err != nil {
			return err
		}
	}

	w.fireStepChange(stepIdx)
	return nil
}

// Finish completes the wizard from any step.
func (w *Wizard) Finish() {
	w.mu.Lock()
	w.completed = true
	cb := w.OnFinish
	w.mu.Unlock()
	if cb != nil {
		cb(w)
	}
}

// Cancel aborts the wizard.
func (w *Wizard) Cancel() {
	w.mu.Lock()
	w.cancelled = true
	cb := w.OnCancel
	w.mu.Unlock()
	if cb != nil {
		cb(w)
	}
}

// Reset returns the wizard to the first step.
func (w *Wizard) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.current = 0
	w.cancelled = false
	w.completed = false
	w.selected = WizardBtnNext
	w.recomputeButtonsLocked()
}

// ─── State ────────────────────────────────────────────────────────────────

// IsCompleted returns true if the wizard has been finished.
func (w *Wizard) IsCompleted() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.completed
}

// IsCancelled returns true if the wizard was cancelled.
func (w *Wizard) IsCancelled() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.cancelled
}

// SelectedButton returns the currently focused button.
func (w *Wizard) SelectedButton() WizardButton {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.selected
}

// SetSelectedButton sets the focused button.
func (w *Wizard) SetSelectedButton(b WizardButton) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.selected = b
}

// ButtonOrder returns the visible buttons for the current step.
func (w *Wizard) ButtonOrder() []WizardButton {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]WizardButton, len(w.buttonOrder))
	copy(out, w.buttonOrder)
	return out
}

// ─── Style ────────────────────────────────────────────────────────────────

// Style returns the current wizard style.
func (w *Wizard) Style() WizardStyle {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.style
}

// SetStyle sets the wizard style.
func (w *Wizard) SetStyle(s WizardStyle) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.style = s
}

// ─── Callbacks ────────────────────────────────────────────────────────────

// SetOnFinish sets the finish callback.
func (w *Wizard) SetOnFinish(fn func(w *Wizard)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.OnFinish = fn
}

// SetOnCancel sets the cancel callback.
func (w *Wizard) SetOnCancel(fn func(w *Wizard)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.OnCancel = fn
}

// SetOnStepChange sets the step-change callback.
func (w *Wizard) SetOnStepChange(fn func(w *Wizard, stepIdx int)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.OnStepChange = fn
}

func (w *Wizard) fireStepChange(idx int) {
	w.mu.RLock()
	fn := w.OnStepChange
	w.mu.RUnlock()
	if fn != nil {
		fn(w, idx)
	}
}

// ─── Keyboard Input ───────────────────────────────────────────────────────

// HandleKey processes keyboard input. Returns true if consumed.
func (w *Wizard) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	// First try routing to step content
	w.mu.RLock()
	var content Component
	if w.current >= 0 && w.current < len(w.steps) {
		step := w.steps[w.current]
		if step != nil {
			content = step.Content
		}
	}
	w.mu.RUnlock()

	if content != nil {
		if kp, ok := content.(interface{ HandleKey(*term.KeyEvent) bool }); ok {
			if kp.HandleKey(key) {
				return true
			}
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	switch key.Key {
	case term.KeyTab:
		if key.Modifiers&term.ModShift != 0 {
			w.moveButtonBackward()
		} else {
			w.moveButtonForward()
		}
		return true

	case term.KeyLeft:
		w.moveButtonBackward()
		return true

	case term.KeyRight:
		w.moveButtonForward()
		return true

	case term.KeyEnter:
		return w.activateButtonLocked()

	case term.KeyEscape:
		w.cancelled = true
		cb := w.OnCancel
		w.mu.Unlock()
		if cb != nil {
			cb(w)
		}
		w.mu.Lock()
		return true
	}

	// Ctrl+N = next, Ctrl+B = back
	if key.Modifiers&term.ModCtrl != 0 {
		switch key.Rune {
		case 'n':
			w.mu.Unlock()
			_ = w.Next()
			w.mu.Lock()
			return true
		case 'b':
			w.mu.Unlock()
			_ = w.Back()
			w.mu.Lock()
			return true
		}
	}

	return false
}

func (w *Wizard) moveButtonForward() {
	if len(w.buttonOrder) == 0 {
		return
	}
	for i, b := range w.buttonOrder {
		if b == w.selected {
			w.selected = w.buttonOrder[(i+1)%len(w.buttonOrder)]
			return
		}
	}
	w.selected = w.buttonOrder[0]
}

func (w *Wizard) moveButtonBackward() {
	if len(w.buttonOrder) == 0 {
		return
	}
	for i, b := range w.buttonOrder {
		if b == w.selected {
			w.selected = w.buttonOrder[(i-1+len(w.buttonOrder))%len(w.buttonOrder)]
			return
		}
	}
	w.selected = w.buttonOrder[0]
}

func (w *Wizard) activateButtonLocked() bool {
	switch w.selected {
	case WizardBtnBack:
		w.mu.Unlock()
		_ = w.Back()
		w.mu.Lock()
		return true
	case WizardBtnNext:
		w.mu.Unlock()
		_ = w.Next()
		w.mu.Lock()
		return true
	case WizardBtnFinish:
		w.completed = true
		cb := w.OnFinish
		w.mu.Unlock()
		if cb != nil {
			cb(w)
		}
		w.mu.Lock()
		return true
	case WizardBtnCancel:
		w.cancelled = true
		cb := w.OnCancel
		w.mu.Unlock()
		if cb != nil {
			cb(w)
		}
		w.mu.Lock()
		return true
	}
	return false
}

// ─── Component Interface ──────────────────────────────────────────────────

// Measure returns the preferred size for the wizard.
func (w *Wizard) Measure(cs Constraints) Size {
	w.mu.RLock()
	defer w.mu.RUnlock()

	width := w.width
	height := w.height

	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && height > cs.MaxHeight {
		height = cs.MaxHeight
	}
	if width < 30 {
		width = 30
	}
	if height < 8 {
		height = 8
	}

	return Size{W: width, H: height}
}

// SetBounds sets the position and size of the wizard.
func (w *Wizard) SetBounds(r Rect) {
	w.mu.Lock()
	w.BaseComponent.SetBounds(r)
	width := r.W
	height := r.H
	// Update step content bounds
	if w.current >= 0 && w.current < len(w.steps) {
		step := w.steps[w.current]
		if step != nil && step.Content != nil {
			bodyY := r.Y + 5
			bodyH := height - 8
			if bodyH < 1 {
				bodyH = 1
			}
			bodyW := width - 2
			if bodyW < 1 {
				bodyW = 1
			}
			w.mu.Unlock()
			step.Content.SetBounds(Rect{X: r.X + 1, Y: bodyY, W: bodyW, H: bodyH})
			w.mu.Lock()
		}
	}
	w.mu.Unlock()
}

// Paint renders the wizard into the buffer.
func (w *Wizard) Paint(buf *buffer.Buffer) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	bounds := w.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	s := w.style

	// Draw border
	w.drawBorderLocked(buf, bounds, s)

	// Title and description
	y := bounds.Y + 1
	if w.current >= 0 && w.current < len(w.steps) && w.steps[w.current] != nil {
		step := w.steps[w.current]
		titleText := fmt.Sprintf("Step %d/%d: %s", w.current+1, len(w.steps), step.Title)
		buf.DrawText(bounds.X+1, y, titleText, s.Title)
		y++

		if step.Description != "" {
			buf.DrawTextClamped(bounds.X+1, y, step.Description, s.Description)
		}
	}

	// Progress indicator
	progressY := bounds.Y + bounds.H - 3
	w.drawProgressLocked(buf, bounds.X+1, progressY, bounds.W-2, s)

	// Buttons
	buttonY := bounds.Y + bounds.H - 2
	w.drawButtonsLocked(buf, bounds.X+1, buttonY, bounds.W-2, s)

	// Paint step content
	if w.current >= 0 && w.current < len(w.steps) {
		step := w.steps[w.current]
		if step != nil && step.Content != nil {
			step.Content.Paint(buf)
		}
	}
}

func (w *Wizard) drawBorderLocked(buf *buffer.Buffer, r Rect, s WizardStyle) {
	buf.SetCell(r.X, r.Y, buffer.NewCell('┌', s.Border))
	buf.SetCell(r.X+r.W-1, r.Y, buffer.NewCell('┐', s.Border))
	for x := r.X + 1; x < r.X+r.W-1; x++ {
		buf.SetCell(x, r.Y, buffer.NewCell('─', s.Border))
	}
	for y := r.Y + 1; y < r.Y+r.H-1; y++ {
		buf.SetCell(r.X, y, buffer.NewCell('│', s.Border))
		buf.SetCell(r.X+r.W-1, y, buffer.NewCell('│', s.Border))
	}
	buf.SetCell(r.X, r.Y+r.H-1, buffer.NewCell('└', s.Border))
	buf.SetCell(r.X+r.W-1, r.Y+r.H-1, buffer.NewCell('┘', s.Border))
	for x := r.X + 1; x < r.X+r.W-1; x++ {
		buf.SetCell(x, r.Y+r.H-1, buffer.NewCell('─', s.Border))
	}
}

func (w *Wizard) drawProgressLocked(buf *buffer.Buffer, x, y, width int, s WizardStyle) {
	if len(w.steps) == 0 {
		return
	}

	colX := x
	for i, step := range w.steps {
		if colX >= x+width {
			break
		}
		var icon rune
		var iconStyle buffer.Style
		if i < w.current {
			icon = '✓'
			iconStyle = s.StepDone
		} else if i == w.current {
			icon = '●'
			iconStyle = s.StepActive
		} else {
			icon = '○'
			iconStyle = s.StepPending
		}
		buf.SetCell(colX, y, buffer.NewCell(icon, iconStyle))
		colX++

		label := step.Title
		maxLabelW := 15
		if buffer.StringWidth(label) > maxLabelW {
			runes := []rune(label)
			label = string(runes[:maxLabelW-1]) + "…"
		}

		labelStyle := s.StepPending
		if i == w.current {
			labelStyle = s.StepActive
		} else if i < w.current {
			labelStyle = s.StepDone
		}
		buf.DrawText(colX, y, " "+label, labelStyle)
		colX += 1 + buffer.StringWidth(label)

		if i < len(w.steps)-1 && colX < x+width-1 {
			buf.SetCell(colX, y, buffer.NewCell('→', s.StepPending))
			colX++
		}
	}
}

func (w *Wizard) drawButtonsLocked(buf *buffer.Buffer, x, y, width int, s WizardStyle) {
	if len(w.buttonOrder) == 0 {
		return
	}

	totalW := 0
	buttonLabels := make([]string, len(w.buttonOrder))
	for i, b := range w.buttonOrder {
		buttonLabels[i] = fmt.Sprintf(" [%s] ", b.ButtonLabel())
		totalW += len(buttonLabels[i])
	}

	curX := x + width - totalW
	if curX < x {
		curX = x
	}
	for i, b := range w.buttonOrder {
		style := s.ButtonNormal
		if b == w.selected {
			style = s.ButtonSelected
		}
		buf.DrawText(curX, y, buttonLabels[i], style)
		curX += len(buttonLabels[i])
	}

	helpText := "Tab=switch  Enter=select  Esc=cancel"
	if width > totalW+len(helpText)+2 {
		buf.DrawText(x, y, helpText, s.Help)
	}
}

// Children returns the step content components.
func (w *Wizard) Children() []Component {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var children []Component
	if w.current >= 0 && w.current < len(w.steps) {
		step := w.steps[w.current]
		if step != nil && step.Content != nil {
			children = append(children, step.Content)
		}
	}
	return children
}

// String returns a string representation.
func (w *Wizard) String() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return fmt.Sprintf("Wizard{steps=%d current=%d completed=%v cancelled=%v}",
		len(w.steps), w.current, w.completed, w.cancelled)
}

// Ensure strings import is used
var _ = strings.TrimSpace
