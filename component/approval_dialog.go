package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ApprovalDialogType specifies the kind of approval dialog.
type ApprovalDialogType int

const (
	ApprovalDialogApproval ApprovalDialogType = iota
	ApprovalDialogConfirm
	ApprovalDialogQuestionnaire
)

// QuestionKind specifies how a question should be answered.
type QuestionKind int

const (
	QSingle QuestionKind = iota // single-select (radio)
	QMulti                      // multi-select (checkbox)
	QText                       // free text
)

// Question represents a single question in a questionnaire dialog.
type Question struct {
	ID       string
	Text     string
	Kind     QuestionKind
	Options  []string
	Required bool

	// State for QSingle: index into Options; -1 = none
	SingleIndex int
	// State for QMulti: selected[i] = true if option i is selected
	Selected []bool
	// State for QText: typed text
	TextAnswer  string
	textCursor  int
}

// IsAnswered returns true if the question has a valid answer.
func (q *Question) IsAnswered() bool {
	switch q.Kind {
	case QSingle:
		return q.SingleIndex >= 0 && q.SingleIndex < len(q.Options)
	case QMulti:
		for _, s := range q.Selected {
			if s {
				return true
			}
		}
		return false
	case QText:
		return len(q.TextAnswer) > 0
	}
	return false
}

// Answer returns the answer as a string.
func (q *Question) Answer() string {
	switch q.Kind {
	case QSingle:
		if q.SingleIndex >= 0 && q.SingleIndex < len(q.Options) {
			return q.Options[q.SingleIndex]
		}
		return ""
	case QMulti:
		var parts []string
		for i, s := range q.Selected {
			if s && i < len(q.Options) {
				parts = append(parts, q.Options[i])
			}
		}
		return strings.Join(parts, ", ")
	case QText:
		return q.TextAnswer
	}
	return ""
}

// DialogAction represents a button action.
type DialogAction struct {
	ID    string
	Label string
}

// Common actions
var (
	ActionApprove = DialogAction{ID: "approve", Label: "Approve"}
	ActionDeny    = DialogAction{ID: "deny", Label: "Deny"}
	ActionOK      = DialogAction{ID: "ok", Label: "OK"}
	ActionCancel  = DialogAction{ID: "cancel", Label: "Cancel"}
	ActionSubmit  = DialogAction{ID: "submit", Label: "Submit"}
	ActionBack    = DialogAction{ID: "back", Label: "Back"}
	ActionNext    = DialogAction{ID: "next", Label: "Next"}
)

// ApprovalDialog is a modal dialog for user approval/confirmation/questionnaire.
type ApprovalDialog struct {
	BaseComponent

	dialogType ApprovalDialogType
	title      string
	body       string

	// For approval/confirm: list of actions
	actions    []DialogAction
	actionIdx  int

	// For questionnaire: list of questions
	questions      []Question
	currentQ       int
	completed      bool

	// Callbacks
	OnResult  func(actionID string, answers map[string]string)
	OnClose   func()

	// Layout
	width  int
	height int

	mu sync.RWMutex
}

// NewApprovalDialog creates an approval dialog with a title and body.
func NewApprovalDialog(title, body string) *ApprovalDialog {
	return &ApprovalDialog{
		dialogType: ApprovalDialogApproval,
		title:      title,
		body:       body,
		actions:    []DialogAction{ActionApprove, ActionDeny},
		actionIdx:  0,
		width:      60,
		height:     15,
	}
}

// NewConfirmDialog creates a simple OK/Cancel confirmation dialog.
func NewApprovalConfirmDialog(title, body string) *ApprovalDialog {
	return &ApprovalDialog{
		dialogType: ApprovalDialogApproval,
		title:      title,
		body:       body,
		actions:    []DialogAction{ActionOK, ActionCancel},
		actionIdx:  0,
		width:      50,
		height:     10,
	}
}

// NewQuestionnaireDialog creates a multi-step questionnaire dialog.
func NewQuestionnaireDialog(title string, questions []Question) *ApprovalDialog {
	// Initialize question state
	for i := range questions {
		q := &questions[i]
		if q.Kind == QSingle {
			q.SingleIndex = -1
		}
		if q.Kind == QMulti {
			if q.Selected == nil {
				q.Selected = make([]bool, len(q.Options))
			}
		}
	}
	return &ApprovalDialog{
		dialogType: ApprovalDialogQuestionnaire,
		title:      title,
		questions:  questions,
		currentQ:   0,
		actions:    []DialogAction{ActionNext, ActionCancel},
		actionIdx:  0,
		width:      60,
		height:     20,
	}
}

// SetDialogType sets the dialog type.
func (d *ApprovalDialog) SetDialogType(dt ApprovalDialogType) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dialogType = dt
}

// DialogType returns the current dialog type.
func (d *ApprovalDialog) DialogType() ApprovalDialogType {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dialogType
}

// SetTitle sets the dialog title.
func (d *ApprovalDialog) SetTitle(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.title = s
}

// Title returns the dialog title.
func (d *ApprovalDialog) Title() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.title
}

// SetBody sets the dialog body text.
func (d *ApprovalDialog) SetBody(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.body = s
}

// Body returns the dialog body text.
func (d *ApprovalDialog) Body() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.body
}

// SetActions sets the available actions for approval/confirm dialogs.
func (d *ApprovalDialog) SetActions(actions []DialogAction) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.actions = actions
	if d.actionIdx >= len(d.actions) {
		d.actionIdx = 0
	}
}

// SetWidth sets the dialog width.
func (d *ApprovalDialog) SetWidth(w int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if w < 20 {
		w = 20
	}
	d.width = w
}

// SetHeight sets the dialog height.
func (d *ApprovalDialog) SetHeight(h int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if h < 5 {
		h = 5
	}
	d.height = h
}

// SetOnResult sets the callback for when the user selects an action.
func (d *ApprovalDialog) SetOnResult(fn func(actionID string, answers map[string]string)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.OnResult = fn
}

// SetOnClose sets the callback for when the dialog is closed.
func (d *ApprovalDialog) SetOnClose(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.OnClose = fn
}

// CurrentQuestionIndex returns the current question index (questionnaire mode).
func (d *ApprovalDialog) CurrentQuestionIndex() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentQ
}

// IsCompleted returns whether the questionnaire is completed.
func (d *ApprovalDialog) IsCompleted() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.completed
}

// ActionIndex returns the currently focused action.
func (d *ApprovalDialog) ActionIndex() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.actionIdx
}

// CollectAnswers returns all answers as a map of question ID -> answer string.
func (d *ApprovalDialog) CollectAnswers() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.collectAnswersLocked()
}

func (d *ApprovalDialog) collectAnswersLocked() map[string]string {
	answers := make(map[string]string)
	for _, q := range d.questions {
		answers[q.ID] = q.Answer()
	}
	return answers
}

// Measure returns the dialog dimensions.
func (d *ApprovalDialog) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w, h := d.width, d.height
	if cs.MaxWidth > 0 && cs.MaxWidth < w {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && cs.MaxHeight < h {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the dialog into the buffer.
func (d *ApprovalDialog) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := d.Bounds()
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H
	if w <= 0 || h <= 0 {
		return
	}

	// Draw border
	borderStyle := buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan)}
	d.drawBoxLocked(buf, x, y, w, h, borderStyle)

	// Draw title
	if d.title != "" {
		titleStr := " " + adTruncateStr(d.title, w-4) + " "
		tx := x + (w-len(titleStr))/2
		if tx < x+1 {
			tx = x + 1
		}
		for i, r := range titleStr {
			if tx+i >= x+w-1 {
				break
			}
			buf.SetCell(tx+i, y, buffer.Cell{
				Rune: r, Width: 1,
				Fg: buffer.NamedColor(buffer.NamedCyan), Flags: buffer.Bold,
			})
		}
	}

	row := y + 2

	// Draw body or current question
	if d.dialogType == ApprovalDialogQuestionnaire && len(d.questions) > 0 {
		row = d.paintQuestionLocked(buf, x, row, w)
	} else {
		// Draw body text (word-wrapped)
		for _, line := range adWrapText(d.body, w-4) {
			if row >= y+h-2 {
				break
			}
			for i, r := range line {
				if x+1+i >= x+w-1 {
					break
				}
				buf.SetCell(x+1+i, row, buffer.Cell{
					Rune: r, Width: 1,
					Fg: buffer.NamedColor(buffer.NamedWhite),
				})
			}
			row++
		}
		row++
	}

	// Draw action buttons at the bottom
	d.paintActionsLocked(buf, x, y+h-2, w)
}

func (d *ApprovalDialog) paintQuestionLocked(buf *buffer.Buffer, x, y, w int) int {
	if d.currentQ < 0 || d.currentQ >= len(d.questions) {
		return y
	}
	q := d.questions[d.currentQ]

	// Progress indicator
	progress := fmt.Sprintf("Question %d/%d", d.currentQ+1, len(d.questions))
	for i, r := range progress {
		if x+1+i >= x+w-1 {
			break
		}
		buf.SetCell(x+1+i, y, buffer.Cell{
			Rune: r, Width: 1,
			Fg: buffer.NamedColor(buffer.NamedMagenta), Flags: buffer.Dim,
		})
	}
	y += 2

	// Question text
	for _, line := range adWrapText(q.Text, w-4) {
		if y >= y+10 {
			break
		}
		for i, r := range line {
			if x+1+i >= x+w-1 {
				break
			}
			buf.SetCell(x+1+i, y, buffer.Cell{
				Rune: r, Width: 1,
				Fg: buffer.NamedColor(buffer.NamedWhite), Flags: buffer.Bold,
			})
		}
		y++
	}
	y++

	// Options or text input
	switch q.Kind {
	case QSingle:
		for i, opt := range q.Options {
			if y >= d.Bounds().Y+d.Bounds().H-3 {
				break
			}
			marker := "○"
			fg := buffer.NamedColor(buffer.NamedWhite)
			if i == q.SingleIndex {
				marker = "●"
				fg = buffer.NamedColor(buffer.NamedGreen)
			}
			// Selected option highlighting
			optStyle := buffer.StyleFlags(0)
			if i == d.actionIdx && d.dialogType != ApprovalDialogQuestionnaire {
				optStyle = buffer.Reverse
			}
			for j, r := range marker {
				buf.SetCell(x+2+j, y, buffer.Cell{Rune: r, Width: 1, Fg: fg})
			}
			// In questionnaire, we don't use actionIdx for options — use a separate cursor
			// For simplicity, let's track option cursor via actionIdx
			isCurrent := false
			if d.dialogType == ApprovalDialogQuestionnaire && d.currentQ < len(d.questions) {
				// Use SingleIndex as cursor position display
			}
			_ = isCurrent

			for j, r := range opt {
				if x+5+j >= x+w-1 {
					break
				}
				buf.SetCell(x+5+j, y, buffer.Cell{Rune: r, Width: 1, Fg: fg, Flags: optStyle})
			}
			y++
		}

	case QMulti:
		for i, opt := range q.Options {
			if y >= d.Bounds().Y+d.Bounds().H-3 {
				break
			}
			marker := "☐"
			fg := buffer.NamedColor(buffer.NamedWhite)
			if i < len(q.Selected) && q.Selected[i] {
				marker = "☑"
				fg = buffer.NamedColor(buffer.NamedGreen)
			}
			for j, r := range marker {
				buf.SetCell(x+2+j, y, buffer.Cell{Rune: r, Width: 1, Fg: fg})
			}
			for j, r := range opt {
				if x+5+j >= x+w-1 {
					break
				}
				buf.SetCell(x+5+j, y, buffer.Cell{Rune: r, Width: 1, Fg: fg})
			}
			y++
		}

	case QText:
		// Text input line
		buf.SetCell(x+2, y, buffer.Cell{Rune: '>', Width: 1, Fg: buffer.NamedColor(buffer.NamedCyan)})
		col := 4
		for _, r := range q.TextAnswer {
			if x+col >= x+w-1 {
				break
			}
			buf.SetCell(x+col, y, buffer.Cell{Rune: r, Width: 1, Fg: buffer.NamedColor(buffer.NamedWhite)})
			col++
		}
		// Cursor
		if x+col < x+w-1 {
			buf.SetCell(x+col, y, buffer.Cell{Rune: '_', Width: 1, Fg: buffer.NamedColor(buffer.NamedWhite), Flags: buffer.Blink})
		}
		y++
	}

	return y
}

func (d *ApprovalDialog) paintActionsLocked(buf *buffer.Buffer, x, y, w int) {
	if len(d.actions) == 0 {
		return
	}

	// Calculate total button width
	totalBtnW := 0
	for _, a := range d.actions {
		totalBtnW += len(a.Label) + 4 // [<label>]
	}
	totalBtnW += len(d.actions) - 1 // spaces between

	startX := x + (w-totalBtnW)/2
	if startX < x+1 {
		startX = x + 1
	}

	curX := startX
	for i, a := range d.actions {
		label := fmt.Sprintf("[ %s ]", a.Label)
		fg := buffer.NamedColor(buffer.NamedWhite)
		if i == d.actionIdx {
			fg = buffer.NamedColor(buffer.NamedYellow)
		}
		for j, r := range label {
			if curX+j >= x+w-1 {
				break
			}
			flags := buffer.StyleFlags(0)
			if i == d.actionIdx {
				flags = buffer.Bold
			}
			buf.SetCell(curX+j, y, buffer.Cell{Rune: r, Width: 1, Fg: fg, Flags: flags})
		}
		curX += len(label) + 1
	}
}

func (d *ApprovalDialog) drawBoxLocked(buf *buffer.Buffer, x, y, w, h int, style buffer.Style) {
	if w < 2 || h < 2 {
		return
	}

	// Corners
	buf.SetCell(x, y, buffer.Cell{Rune: '╭', Width: 1, Fg: style.Fg})
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '╮', Width: 1, Fg: style.Fg})
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '╰', Width: 1, Fg: style.Fg})
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '╯', Width: 1, Fg: style.Fg})

	// Top and bottom
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '─', Width: 1, Fg: style.Fg})
		buf.SetCell(x+i, y+h-1, buffer.Cell{Rune: '─', Width: 1, Fg: style.Fg})
	}

	// Left and right
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.Cell{Rune: '│', Width: 1, Fg: style.Fg})
		buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: '│', Width: 1, Fg: style.Fg})
	}

	// Fill interior with space
	for row := y + 1; row < y+h-1; row++ {
		for col := x + 1; col < x+w-1; col++ {
			if col < buf.Width && row < buf.Height {
				buf.SetCell(col, row, buffer.Cell{Rune: ' ', Width: 1})
			}
		}
	}
}

// pendingAction stores callbacks to fire after lock release.
type pendingAction struct {
	resultID string
	answers  map[string]string
	close    bool
}

// HandleKey processes keyboard input for the dialog.
func (d *ApprovalDialog) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}

	d.mu.Lock()
	var pa *pendingAction
	var result bool

	if d.dialogType == ApprovalDialogQuestionnaire {
		result, pa = d.handleQuestionnaireKeyLocked(k)
	} else {
		result, pa = d.handleApprovalKeyLocked(k)
	}
	d.mu.Unlock()

	// Fire callbacks after lock release
	if pa != nil {
		cb := d.OnResult
		closeCB := d.OnClose
		if cb != nil {
			cb(pa.resultID, pa.answers)
		}
		if pa.close && closeCB != nil {
			closeCB()
		}
	}

	return result
}

func (d *ApprovalDialog) handleApprovalKeyLocked(k *term.KeyEvent) (bool, *pendingAction) {
	switch k.Key {
	case term.KeyLeft:
		if d.actionIdx > 0 {
			d.actionIdx--
		} else {
			d.actionIdx = len(d.actions) - 1
		}
		return true, nil

	case term.KeyRight, term.KeyTab:
		d.actionIdx++
		if d.actionIdx >= len(d.actions) {
			d.actionIdx = 0
		}
		return true, nil

	case term.KeyEnter:
		return d.executeActionLocked()

	case term.KeyEscape:
		return true, &pendingAction{resultID: "cancel", close: true}
	}

	// Keyboard shortcuts
	if k.Rune != 0 && (k.Modifiers&term.ModCtrl != 0) {
		switch k.Rune {
		case 'y', 'Y':
			d.actionIdx = 0
			return d.executeActionLocked()
		case 'n', 'N':
			if len(d.actions) > 1 {
				d.actionIdx = 1
				return d.executeActionLocked()
			}
		}
	}

	return false, nil
}

func (d *ApprovalDialog) handleQuestionnaireKeyLocked(k *term.KeyEvent) (bool, *pendingAction) {
	if d.currentQ < 0 || d.currentQ >= len(d.questions) {
		return false, nil
	}
	q := &d.questions[d.currentQ]

	switch q.Kind {
	case QSingle:
		switch k.Key {
		case term.KeyUp:
			if q.SingleIndex > 0 {
				q.SingleIndex--
			} else {
				q.SingleIndex = len(q.Options) - 1
			}
			return true, nil
		case term.KeyDown:
			q.SingleIndex++
			if q.SingleIndex >= len(q.Options) {
				q.SingleIndex = 0
			}
			return true, nil
		case term.KeyEnter:
			return d.handleQuestionnaireActionLocked(k)
		case term.KeyEscape:
			return true, &pendingAction{resultID: "cancel", close: true}
		case term.KeyLeft:
			if d.actionIdx > 0 {
				d.actionIdx--
			}
			return true, nil
		case term.KeyRight:
			if d.actionIdx < len(d.actions)-1 {
				d.actionIdx++
			}
			return true, nil
		}

	case QMulti:
		switch k.Key {
		case term.KeyUp:
			if d.actionIdx > 0 {
				d.actionIdx--
			}
			return true, nil
		case term.KeyDown:
			if d.actionIdx < len(q.Options)-1 {
				d.actionIdx++
			}
			return true, nil
		case term.KeyEnter, term.KeySpace:
			if d.actionIdx < len(q.Options) {
				if d.actionIdx < len(q.Selected) {
					q.Selected[d.actionIdx] = !q.Selected[d.actionIdx]
				}
				return true, nil
			}
			return d.handleQuestionnaireActionLocked(k)
		case term.KeyEscape:
			return true, &pendingAction{resultID: "cancel", close: true}
		}

	case QText:
		switch k.Key {
		case term.KeyTab:
			// Tab cycles between text input and action buttons
			if len(d.actions) > 1 {
				if d.actionIdx == 0 {
					d.actionIdx = 1
				} else {
					d.actionIdx = 0
				}
			}
			return true, nil
		case term.KeyEnter:
			return d.handleQuestionnaireActionLocked(k)
		case term.KeyBackspace:
			if q.textCursor > 0 && q.textCursor <= len(q.TextAnswer) {
				q.TextAnswer = q.TextAnswer[:q.textCursor-1] + q.TextAnswer[q.textCursor:]
				q.textCursor--
			}
			return true, nil
		case term.KeyEscape:
			return true, &pendingAction{resultID: "cancel", close: true}
		case term.KeyLeft:
			if q.textCursor > 0 {
				q.textCursor--
			}
			return true, nil
		case term.KeyRight:
			if q.textCursor < len(q.TextAnswer) {
				q.textCursor++
			}
			return true, nil
		default:
			if k.Rune >= 0x20 {
				q.TextAnswer = q.TextAnswer[:q.textCursor] + string(k.Rune) + q.TextAnswer[q.textCursor:]
				q.textCursor++
				return true, nil
			}
		}
	}

	return false, nil
}

func (d *ApprovalDialog) handleQuestionnaireActionLocked(k *term.KeyEvent) (bool, *pendingAction) {
	// Next/Submit (actionIdx == 0)
	if d.actionIdx == 0 {
		q := &d.questions[d.currentQ]
		if q.Required && !q.IsAnswered() {
			return true, nil
		}

		if d.currentQ < len(d.questions)-1 {
			d.currentQ++
			if d.currentQ == len(d.questions)-1 {
				d.actions = []DialogAction{ActionSubmit, ActionBack}
			} else {
				d.actions = []DialogAction{ActionNext, ActionCancel}
			}
			d.actionIdx = 0
			return true, nil
		}

		// Submit
		d.completed = true
		return true, &pendingAction{resultID: "submit", answers: d.collectAnswersLocked(), close: true}
	}

	// Back
	if d.actionIdx < len(d.actions) && d.actions[d.actionIdx].ID == "back" && d.currentQ > 0 {
		d.currentQ--
		if d.currentQ == 0 {
			d.actions = []DialogAction{ActionNext, ActionCancel}
		} else {
			d.actions = []DialogAction{ActionNext, ActionBack}
		}
		d.actionIdx = 0
		return true, nil
	}

	// Cancel
	return true, &pendingAction{resultID: "cancel", close: true}
}

func (d *ApprovalDialog) executeActionLocked() (bool, *pendingAction) {
	if d.actionIdx < 0 || d.actionIdx >= len(d.actions) {
		return false, nil
	}
	action := d.actions[d.actionIdx]
	answers := d.collectAnswersLocked()
	return true, &pendingAction{resultID: action.ID, answers: answers, close: true}
}

// Children returns nil.
func (d *ApprovalDialog) Children() []Component {
	return nil
}

// Helper functions

func adTruncateStr(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func adWrapText(text string, maxW int) []string {
	if maxW <= 0 {
		return []string{text}
	}
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		curLine := words[0]
		for _, w := range words[1:] {
			if len(curLine)+1+len(w) > maxW {
				lines = append(lines, curLine)
				curLine = w
			} else {
				curLine += " " + w
			}
		}
		lines = append(lines, curLine)
	}
	return lines
}
