package block

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// StepStatus represents the execution status of a workflow step.
type StepStatus uint8

const (
	StepPending StepStatus = iota // Waiting to run
	StepRunning                   // Currently executing
	StepDone                      // Completed successfully
	StepFailed                    // Errored
	StepSkipped                   // Skipped (conditional)
)

// String returns a human-readable status name.
func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepRunning:
		return "running"
	case StepDone:
		return "done"
	case StepFailed:
		return "failed"
	case StepSkipped:
		return "skipped"
	}
	return "unknown"
}

// statusIcon returns the icon rune for a status.
func statusIcon(s StepStatus) rune {
	switch s {
	case StepPending:
		return '○'
	case StepRunning:
		return '⠋' // default spinner frame; caller can override via SpinnerFrame()
	case StepDone:
		return '✓'
	case StepFailed:
		return '✗'
	case StepSkipped:
		return '⊘'
	}
	return '?'
}

// statusColor returns the display color for a status.
func statusColor(s StepStatus) buffer.Color {
	switch s {
	case StepPending:
		return buffer.RGB(0x62, 0x72, 0xA4) // dim blue-gray
	case StepRunning:
		return buffer.RGB(0x8B, 0xE9, 0xFD) // cyan
	case StepDone:
		return buffer.RGB(0x50, 0xFA, 0x7B) // green
	case StepFailed:
		return buffer.RGB(0xFF, 0x55, 0x55) // red
	case StepSkipped:
		return buffer.RGB(0xF1, 0xFA, 0x8C) // yellow
	}
	return buffer.NoColor()
}

// WorkflowStep represents a single step in an agent workflow.
type WorkflowStep struct {
	Name        string     // short step name
	Description string     // longer description
	Status      StepStatus // current status
	Duration    time.Duration // execution time (set when done/failed)
	Icon        rune       // override icon (0 = use default for status)
	startedAt   time.Time  // when the step entered Running state
}

// WorkflowBlock renders a multi-step agent workflow with live status updates.
// It is designed for streaming: steps are added dynamically and their statuses
// are updated in real time as the agent progresses.
type WorkflowBlock struct {
	BaseBlock
	mu       sync.RWMutex
	steps    []WorkflowStep
	title    string
	spinIdx  int       // spinner frame index for running steps
	lastSpin time.Time // last spinner update
}

// NewWorkflowBlock creates a new workflow block with the given title.
func NewWorkflowBlock(title string) *WorkflowBlock {
	return &WorkflowBlock{
		BaseBlock: NewBaseBlock(component.GenerateID("wf"), TypeWorkflow),
		title:     title,
		lastSpin:  time.Now(),
	}
}

// AddStep appends a new pending step to the workflow.
func (w *WorkflowBlock) AddStep(name, description string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.steps = append(w.steps, WorkflowStep{
		Name:        name,
		Description: description,
		Status:      StepPending,
	})
	w.markDirtyLocked()
}

// SetStepStatus updates the status of the step at the given index.
// When a step enters StepRunning, its timer starts. When it transitions
// to StepDone or StepFailed, the duration is computed.
// Out-of-range indices are silently ignored.
func (w *WorkflowBlock) SetStepStatus(index int, status StepStatus) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if index < 0 || index >= len(w.steps) {
		return
	}
	step := &w.steps[index]
	oldStatus := step.Status
	step.Status = status

	if status == StepRunning && oldStatus != StepRunning {
		step.startedAt = time.Now()
	}
	if (status == StepDone || status == StepFailed) && oldStatus == StepRunning {
		if !step.startedAt.IsZero() {
			step.Duration = time.Since(step.startedAt)
		}
	}

	// If all steps are terminal, auto-complete the block.
	allDone := true
	for i := range w.steps {
		s := w.steps[i].Status
		if s != StepDone && s != StepFailed && s != StepSkipped {
			allDone = false
			break
		}
	}
	if allDone && len(w.steps) > 0 {
		w.state = BlockComplete
	}
	w.markDirtyLocked()
}

// Steps returns a copy of the current step list.
func (w *WorkflowBlock) Steps() []WorkflowStep {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]WorkflowStep, len(w.steps))
	copy(result, w.steps)
	return result
}

// StepCount returns the number of steps.
func (w *WorkflowBlock) StepCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.steps)
}

// Progress returns the completion fraction (0.0 to 1.0).
// Done and skipped steps count as completed.
func (w *WorkflowBlock) Progress() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if len(w.steps) == 0 {
		return 0
	}
	completed := 0
	for _, s := range w.steps {
		if s.Status == StepDone || s.Status == StepSkipped {
			completed++
		}
	}
	return float64(completed) / float64(len(w.steps))
}

// ProgressText returns a "3/5 (60%)" style string.
func (w *WorkflowBlock) ProgressText() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	total := len(w.steps)
	if total == 0 {
		return "0/0 (0%)"
	}
	completed := 0
	for _, s := range w.steps {
		if s.Status == StepDone || s.Status == StepSkipped {
			completed++
		}
	}
	pct := int(float64(completed) / float64(total) * 100)
	return fmt.Sprintf("%d/%d (%d%%)", completed, total, pct)
}

// Title returns the workflow title.
func (w *WorkflowBlock) Title() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.title
}

// SetTitle updates the workflow title.
func (w *WorkflowBlock) SetTitle(title string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.title = title
	w.markDirtyLocked()
}

// SpinnerFrame returns the current spinner character for running steps,
// advancing the frame based on elapsed time.
func (w *WorkflowBlock) SpinnerFrame() rune {
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	w.mu.Lock()
	defer w.mu.Unlock()
	if time.Since(w.lastSpin) > 100*time.Millisecond {
		w.spinIdx = (w.spinIdx + 1) % len(frames)
		w.lastSpin = time.Now()
	}
	return frames[w.spinIdx]
}

// AdvanceSpinner manually advances the spinner frame (for testing).
func (w *WorkflowBlock) AdvanceSpinner() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.spinIdx = (w.spinIdx + 1) % 10
}

// HasRunning returns true if any step is currently running.
func (w *WorkflowBlock) HasRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, s := range w.steps {
		if s.Status == StepRunning {
			return true
		}
	}
	return false
}

// HasFailed returns true if any step has failed.
func (w *WorkflowBlock) HasFailed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, s := range w.steps {
		if s.Status == StepFailed {
			return true
		}
	}
	return false
}

// Measure computes the desired size for the workflow block.
// Height = title(1) + steps(n) + progress bar(1).
func (w *WorkflowBlock) Measure(cs component.Constraints) component.Size {
	w.mu.RLock()
	defer w.mu.RUnlock()

	h := 1 + len(w.steps) + 1 // title + steps + progress
	if len(w.steps) == 0 {
		h = 2 // title + "no steps"
	}
	width := cs.MaxWidth
	if width <= 0 {
		width = 40
	}
	return component.Size{W: width, H: h}
}

// Paint renders the workflow block into the buffer.
func (w *WorkflowBlock) Paint(buf *buffer.Buffer) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	bounds := w.Bounds()
	x, y := bounds.X, bounds.Y

	// --- Title line ---
	titleStyle := buffer.Style{
		Fg:    buffer.RGB(0xBD, 0x93, 0xF9), // purple
		Flags: buffer.Bold,
	}
	progressStr := fmt.Sprintf(" %s", w.progressTextLocked())
	titleText := w.title + progressStr
	buf.DrawText(x, y, titleText, titleStyle)
	y++

	// --- Steps ---
	for i, step := range w.steps {
		icon := step.Icon
		if icon == 0 {
			if step.Status == StepRunning {
				icon = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}[w.spinIdx]
			} else {
				icon = statusIcon(step.Status)
			}
		}

		fg := statusColor(step.Status)

		// Draw icon
		buf.SetCell(x, y, buffer.Cell{Rune: icon, Width: 1, Fg: fg})

		// Draw step name
		nameStyle := buffer.Style{Fg: fg}
		if step.Status == StepRunning {
			nameStyle.Flags = buffer.Bold
		}
		buf.DrawText(x+2, y, step.Name, nameStyle)

		// Draw duration if available
		if step.Duration > 0 {
			durStr := formatDurationWf(step.Duration)
			durStyle := buffer.Style{Fg: buffer.RGB(0x62, 0x72, 0xA4)} // dim
			durX := bounds.X + bounds.W - len([]rune(durStr)) - 1
			if durX > x+2+len([]rune(step.Name)) {
				buf.DrawText(durX, y, durStr, durStyle)
			}
		}

		// Draw description on the same line if there's room
		if step.Description != "" {
			descText := " — " + step.Description
			descStyle := buffer.Style{Fg: buffer.RGB(0x62, 0x72, 0xA4)}
			maxDescW := bounds.W - len([]rune(step.Name)) - 5
			if maxDescW > 10 {
				descRunes := []rune(descText)
				if len(descRunes) > maxDescW {
					descRunes = descRunes[:maxDescW-1]
					descRunes = append(descRunes, '…')
				}
				buf.DrawText(x+4+len([]rune(step.Name)), y, string(descRunes), descStyle)
			}
		}

		// Draw dependency arrow to next step
		if i < len(w.steps)-1 {
			arrowStyle := buffer.Style{Fg: buffer.RGB(0x62, 0x72, 0xA4)}
			_ = arrowStyle // arrows would be on the left margin in a full impl
		}

		y++
	}

	// --- Progress bar ---
	if bounds.W > 4 {
		barY := y
		pct := w.progressFractionLocked()
		filled := int(float64(bounds.W-2) * pct)

		// Bar border
		barStyle := buffer.Style{Fg: buffer.RGB(0x62, 0x72, 0xA4)}
		buf.SetCell(x, barY, buffer.Cell{Rune: '[', Width: 1, Fg: barStyle.Fg})

		for i := 0; i < bounds.W-2; i++ {
			var c buffer.Cell
			if i < filled {
				c = buffer.Cell{Rune: '█', Width: 1, Fg: buffer.RGB(0x50, 0xFA, 0x7B)}
			} else {
				c = buffer.Cell{Rune: '░', Width: 1, Fg: buffer.RGB(0x44, 0x47, 0x5A)}
			}
			buf.SetCell(x+1+i, barY, c)
		}
		buf.SetCell(x+bounds.W-1, barY, buffer.Cell{Rune: ']', Width: 1, Fg: barStyle.Fg})
	}
}

// SerializeState implements the Serializer interface.
func (w *WorkflowBlock) SerializeState() (json.RawMessage, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	type stepJSON struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		DurationMs  int64  `json:"duration_ms"`
		Icon        string `json:"icon,omitempty"`
	}

	steps := make([]stepJSON, len(w.steps))
	for i, s := range w.steps {
		steps[i] = stepJSON{
			Name:        s.Name,
			Description: s.Description,
			Status:      s.Status.String(),
			DurationMs:  s.Duration.Milliseconds(),
		}
		if s.Icon != 0 {
			steps[i].Icon = string(s.Icon)
		}
	}

	return json.Marshal(map[string]any{
		"title": w.title,
		"steps": steps,
	})
}

// DeserializeState implements the Deserializer interface.
func (w *WorkflowBlock) DeserializeState(data json.RawMessage) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var raw struct {
		Title string `json:"title"`
		Steps []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
			DurationMs  int64  `json:"duration_ms"`
			Icon        string `json:"icon"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	w.title = raw.Title
	w.steps = w.steps[:0]
	for _, s := range raw.Steps {
		step := WorkflowStep{
			Name:        s.Name,
			Description: s.Description,
			Status:      parseStepStatus(s.Status),
			Duration:    time.Duration(s.DurationMs) * time.Millisecond,
		}
		if s.Icon != "" {
			step.Icon = []rune(s.Icon)[0]
		}
		w.steps = append(w.steps, step)
	}
	w.markDirtyLocked()
	return nil
}

// TypeName returns the registry type name for serialization.
func (w *WorkflowBlock) TypeName() string { return "workflow" }

// --- Internal helpers (called under lock) ---

func (w *WorkflowBlock) progressTextLocked() string {
	total := len(w.steps)
	if total == 0 {
		return "0/0 (0%)"
	}
	completed := 0
	for _, s := range w.steps {
		if s.Status == StepDone || s.Status == StepSkipped {
			completed++
		}
	}
	pct := int(float64(completed) / float64(total) * 100)
	return fmt.Sprintf("%d/%d (%d%%)", completed, total, pct)
}

func (w *WorkflowBlock) progressFractionLocked() float64 {
	total := len(w.steps)
	if total == 0 {
		return 0
	}
	completed := 0
	for _, s := range w.steps {
		if s.Status == StepDone || s.Status == StepSkipped {
			completed++
		}
	}
	return float64(completed) / float64(total)
}

// parseStepStatus parses a status string back to StepStatus.
func parseStepStatus(s string) StepStatus {
	switch strings.ToLower(s) {
	case "pending":
		return StepPending
	case "running":
		return StepRunning
	case "done":
		return StepDone
	case "failed":
		return StepFailed
	case "skipped":
		return StepSkipped
	}
	return StepPending
}

// formatDurationWf renders a duration in human-friendly form.
func formatDurationWf(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}
