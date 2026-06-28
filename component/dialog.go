package component

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Types ─────────────────────────────────────────────────────

// DialogType controls which elements the dialog shows.
type DialogType uint8

const (
	// DialogInfo shows a message with a single OK button.
	DialogInfo DialogType = iota
	// DialogConfirm shows a message with OK and Cancel buttons.
	DialogConfirm
	// DialogPrompt shows a message, a text input, and OK/Cancel buttons.
	DialogPrompt
	// DialogCustom allows fully custom button sets.
	DialogCustom
)

// DialogResult identifies which button was pressed.
type DialogResult int

const (
	// DialogResultNone means the dialog hasn't been closed yet.
	DialogResultNone DialogResult = iota
	// DialogResultOK means the primary/confirm button was pressed.
	DialogResultOK
	// DialogResultCancel means the cancel button was pressed.
	DialogResultCancel
	// DialogResultCustom means a custom button (index ≥ 2) was pressed.
	DialogResultCustom
)

// DialogButton represents a single button in the dialog.
type DialogButton struct {
	Label  string
	Result DialogResult
	Action func()
}

// NewDialogButton creates a button with the given label and result.
func NewDialogButton(label string, result DialogResult) DialogButton {
	return DialogButton{Label: label, Result: result}
}

// ─── Style ─────────────────────────────────────────────────────

// DialogStyle holds the visual style for every part of the dialog.
type DialogStyle struct {
	Border          buffer.Style
	Title           buffer.Style
	TitleBar        buffer.Style
	Message         buffer.Style
	Input           buffer.Style
	InputCursor     buffer.Style
	InputBorder     buffer.Style
	Button          buffer.Style
	ButtonActive    buffer.Style
	ButtonSeparator buffer.Style
	Shadow          buffer.Style
}

// DefaultDialogStyle returns a readable dark-theme dialog style.
func DefaultDialogStyle() DialogStyle {
	borderFg := buffer.NamedColor(buffer.NamedCyan)
	return DialogStyle{
		Border:          buffer.Style{Fg: borderFg},
		Title:           buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightYellow), Flags: buffer.Bold},
		TitleBar:        buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan)},
		Message:         buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightWhite)},
		Input:           buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightWhite)},
		InputCursor:     buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlack), Bg: buffer.NamedColor(buffer.NamedBrightWhite), Flags: buffer.Reverse},
		InputBorder:     buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)},
		Button:          buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightWhite)},
		ButtonActive:    buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlack), Bg: buffer.NamedColor(buffer.NamedCyan), Flags: buffer.Bold},
		ButtonSeparator: buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)},
		Shadow:          buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack), Flags: buffer.Dim},
	}
}

// ─── Dialog ────────────────────────────────────────────────────

// Dialog is a modal dialog box with an optional title, message,
// text input field, and a row of buttons.
//
// It implements the Component interface and is designed to be
// rendered as an overlay on top of the main content.
type Dialog struct {
	BaseComponent
	mu sync.RWMutex

	style      DialogStyle
	dialogType DialogType
	title      string
	message    string
	buttons    []DialogButton
	cursor     int // selected button index

	// Input field state (DialogPrompt only)
	input       []rune
	inputCursor int

	// Visibility and result
	visible bool
	result  DialogResult
	width   int
	height  int

	// Callbacks
	OnConfirm func(text string) bool // return false to keep dialog open
	OnCancel  func()
	OnClose   func(result DialogResult, text string)
}

// NewDialog creates a dialog of the given type with default OK/Cancel buttons.
func NewDialog(dt DialogType, title, message string) *Dialog {
	d := &Dialog{
		dialogType: dt,
		title:      title,
		message:    message,
		style:      DefaultDialogStyle(),
		input:      []rune{},
		visible:    true,
	}
	d.SetID(GenerateID("dialog"))

	switch dt {
	case DialogInfo:
		d.buttons = []DialogButton{
			NewDialogButton("OK", DialogResultOK),
		}
	case DialogConfirm:
		d.buttons = []DialogButton{
			NewDialogButton("  OK  ", DialogResultOK),
			NewDialogButton("Cancel", DialogResultCancel),
		}
	case DialogPrompt:
		d.buttons = []DialogButton{
			NewDialogButton("  OK  ", DialogResultOK),
			NewDialogButton("Cancel", DialogResultCancel),
		}
	case DialogCustom:
		// caller must call SetButtons
	}
	return d
}

// NewConfirmDialog is a convenience constructor for a confirmation dialog.
func NewConfirmDialog(title, message string) *Dialog {
	return NewDialog(DialogConfirm, title, message)
}

// NewInfoDialog is a convenience constructor for an informational dialog.
func NewInfoDialog(title, message string) *Dialog {
	return NewDialog(DialogInfo, title, message)
}

// NewPromptDialog is a convenience constructor for a text input dialog.
// The defaultValue is placed in the input field.
func NewPromptDialog(title, message, defaultValue string) *Dialog {
	d := NewDialog(DialogPrompt, title, message)
	d.input = []rune(defaultValue)
	d.inputCursor = len(d.input)
	return d
}

// ─── Accessors ─────────────────────────────────────────────────

// Type returns the dialog type.
func (d *Dialog) Type() DialogType {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dialogType
}

// Title returns the dialog title.
func (d *Dialog) Title() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.title
}

// SetTitle sets the dialog title.
func (d *Dialog) SetTitle(title string) {
	d.mu.Lock()
	d.title = title
	d.mu.Unlock()
}

// Message returns the dialog message text.
func (d *Dialog) Message() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.message
}

// SetMessage sets the dialog message text.
func (d *Dialog) SetMessage(msg string) {
	d.mu.Lock()
	d.message = msg
	d.mu.Unlock()
}

// InputValue returns the current text in the input field.
func (d *Dialog) InputValue() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return string(d.input)
}

// SetInputValue replaces the input field text and moves cursor to end.
func (d *Dialog) SetInputValue(text string) {
	d.mu.Lock()
	d.input = []rune(text)
	d.inputCursor = len(d.input)
	d.mu.Unlock()
}

// InputCursor returns the cursor position within the input field.
func (d *Dialog) InputCursor() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.inputCursor
}

// SetInputCursor sets the cursor position, clamped to [0, len(input)].
func (d *Dialog) SetInputCursor(pos int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setInputCursorLocked(pos)
}

func (d *Dialog) setInputCursorLocked(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(d.input) {
		pos = len(d.input)
	}
	d.inputCursor = pos
}

// Result returns the result of the dialog (valid after it's been closed).
func (d *Dialog) Result() DialogResult {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.result
}

// Visible returns whether the dialog is currently shown.
func (d *Dialog) Visible() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.visible
}

// Show makes the dialog visible.
func (d *Dialog) Show() {
	d.mu.Lock()
	d.visible = true
	d.mu.Unlock()
}

// Hide hides the dialog without setting a result.
func (d *Dialog) Hide() {
	d.mu.Lock()
	d.visible = false
	d.mu.Unlock()
}

// Closed reports whether the dialog has been dismissed by the user.
func (d *Dialog) Closed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return !d.visible
}

// ─── Buttons ───────────────────────────────────────────────────

// Buttons returns a copy of the current buttons.
func (d *Dialog) Buttons() []DialogButton {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]DialogButton, len(d.buttons))
	copy(out, d.buttons)
	return out
}

// SetButtons replaces the button set (use with DialogCustom).
func (d *Dialog) SetButtons(btns []DialogButton) {
	d.mu.Lock()
	d.buttons = make([]DialogButton, len(btns))
	copy(d.buttons, btns)
	if d.cursor >= len(d.buttons) {
		d.cursor = 0
	}
	d.mu.Unlock()
}

// AddButton appends a button.
func (d *Dialog) AddButton(btn DialogButton) {
	d.mu.Lock()
	d.buttons = append(d.buttons, btn)
	d.mu.Unlock()
}

// Cursor returns the index of the highlighted button.
func (d *Dialog) Cursor() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.cursor
}

// SetCursor sets the highlighted button index (wraps around).
func (d *Dialog) SetCursor(idx int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setCursorLocked(idx)
}

func (d *Dialog) setCursorLocked(idx int) {
	if len(d.buttons) == 0 {
		d.cursor = 0
		return
	}
	// wrap around
	for idx < 0 {
		idx += len(d.buttons)
	}
	d.cursor = idx % len(d.buttons)
}

// MoveLeft moves the button cursor left (wraps).
func (d *Dialog) MoveLeft() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setCursorLocked(d.cursor - 1)
}

// MoveRight moves the button cursor right (wraps).
func (d *Dialog) MoveRight() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setCursorLocked(d.cursor + 1)
}

// CurrentButton returns the highlighted button, or nil if no buttons.
func (d *Dialog) CurrentButton() *DialogButton {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.cursor < 0 || d.cursor >= len(d.buttons) {
		return nil
	}
	return &d.buttons[d.cursor]
}

// ─── Input editing ─────────────────────────────────────────────

// InsertRune inserts a rune at the cursor position.
func (d *Dialog) InsertRune(r rune) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.insertRuneLocked(r)
}

func (d *Dialog) insertRuneLocked(r rune) {
	pos := d.inputCursor
	d.input = append(d.input, 0)       // grow
	copy(d.input[pos+1:], d.input[pos:]) // shift right
	d.input[pos] = r
	d.inputCursor++
}

// Backspace deletes the rune before the cursor.
func (d *Dialog) Backspace() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inputCursor <= 0 || len(d.input) == 0 {
		return
	}
	pos := d.inputCursor - 1
	d.input = append(d.input[:pos], d.input[pos+1:]...)
	d.inputCursor--
}

// Delete deletes the rune at the cursor.
func (d *Dialog) Delete() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inputCursor >= len(d.input) {
		return
	}
	pos := d.inputCursor
	d.input = append(d.input[:pos], d.input[pos+1:]...)
}

// CursorLeft moves the input cursor left by one rune.
func (d *Dialog) CursorLeft() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setInputCursorLocked(d.inputCursor - 1)
}

// CursorRight moves the input cursor right by one rune.
func (d *Dialog) CursorRight() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setInputCursorLocked(d.inputCursor + 1)
}

// CursorStart moves the input cursor to the beginning.
func (d *Dialog) CursorStart() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setInputCursorLocked(0)
}

// CursorEnd moves the input cursor to the end.
func (d *Dialog) CursorEnd() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.setInputCursorLocked(len(d.input))
}

// ─── Style ─────────────────────────────────────────────────────

// Style returns the current dialog style.
func (d *Dialog) Style() DialogStyle {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.style
}

// SetStyle replaces the dialog style.
func (d *Dialog) SetStyle(s DialogStyle) {
	d.mu.Lock()
	d.style = s
	d.mu.Unlock()
}

// ─── Actions ───────────────────────────────────────────────────

// Confirm closes the dialog with DialogResultOK.
// If OnConfirm returns false, the dialog stays open.
func (d *Dialog) Confirm() bool {
	d.mu.Lock()
	text := string(d.input)
	d.mu.Unlock()

	if d.OnConfirm != nil {
		if !d.OnConfirm(text) {
			return false
		}
	}

	d.mu.Lock()
	d.result = DialogResultOK
	d.visible = false
	callbacks := d.extractCallbacksLocked()
	d.mu.Unlock()
	d.fireCallbacks(callbacks, DialogResultOK, text)
	return true
}

// Cancel closes the dialog with DialogResultCancel.
func (d *Dialog) Cancel() {
	d.mu.Lock()
	text := string(d.input)
	d.result = DialogResultCancel
	d.visible = false
	callbacks := d.extractCallbacksLocked()
	d.mu.Unlock()

	if callbacks.onCancel != nil {
		callbacks.onCancel()
	}
	if callbacks.onClose != nil {
		callbacks.onClose(DialogResultCancel, text)
	}
}

// PressButton activates the button at the current cursor position.
func (d *Dialog) PressButton() {
	d.mu.Lock()
	if len(d.buttons) == 0 {
		d.mu.Unlock()
		return
	}
	idx := d.cursor
	btn := d.buttons[idx]
	text := string(d.input)
	d.result = btn.Result
	d.visible = false
	callbacks := d.extractCallbacksLocked()
	action := btn.Action
	d.mu.Unlock()

	if action != nil {
		action()
	}
	if callbacks.onClose != nil {
		callbacks.onClose(btn.Result, text)
	}
}

// callbacksBundle captures callbacks outside the lock.
type callbacksBundle struct {
	onConfirm func(string) bool
	onCancel  func()
	onClose   func(DialogResult, string)
}

func (d *Dialog) extractCallbacksLocked() callbacksBundle {
	return callbacksBundle{
		onConfirm: d.OnConfirm,
		onCancel:  d.OnCancel,
		onClose:   d.OnClose,
	}
}

func (d *Dialog) fireCallbacks(cb callbacksBundle, result DialogResult, text string) {
	if cb.onClose != nil {
		cb.onClose(result, text)
	}
}

// ─── HandleKey ─────────────────────────────────────────────────

// HandleKey processes keyboard input.
// Returns true if the key was consumed.
func (d *Dialog) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	// In prompt mode, printable runes go to the input field
	if d.dialogType == DialogPrompt && key.Rune != 0 && key.Rune != 0xFFFF && key.Key == term.KeyUnknown {
		d.InsertRune(key.Rune)
		return true
	}

	switch key.Key {
	case term.KeyLeft:
		if d.dialogType == DialogPrompt {
			d.CursorLeft()
		} else {
			d.MoveLeft()
		}
		return true
	case term.KeyRight:
		if d.dialogType == DialogPrompt {
			d.CursorRight()
		} else {
			d.MoveRight()
		}
		return true
	case term.KeyUp:
		// In prompt mode, Up/Down could navigate buttons
		d.MoveLeft()
		return true
	case term.KeyDown:
		d.MoveRight()
		return true
	case term.KeyTab:
		d.MoveRight()
		return true
	case term.KeyBackspace:
		if d.dialogType == DialogPrompt {
			d.Backspace()
			return true
		}
	case term.KeyEnter:
		// Enter activates the current button
		if d.dialogType == DialogPrompt && d.cursor == 0 {
			// In prompt mode, Enter on OK button confirms
			if d.Confirm() {
				return true
			}
			return true
		}
		d.PressButton()
		return true
	case term.KeyEscape:
		d.Cancel()
		return true
	}

	// Ctrl+A / Ctrl+E for cursor movement in prompt mode
	if d.dialogType == DialogPrompt && key.Modifiers != 0 {
		if key.Modifiers&term.ModCtrl != 0 {
			switch key.Key {
			case 0x01: // Ctrl+A — Home
				d.CursorStart()
				return true
			case 0x05: // Ctrl+E — End
				d.CursorEnd()
				return true
			}
		}
	}

	return false
}

// ─── Measure / Paint ───────────────────────────────────────────

// Measure computes the desired size of the dialog.
func (d *Dialog) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := d.measureWidthLocked()
	h := d.measureHeightLocked()

	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}

	d.width = w
	d.height = h
	return Size{W: w, H: h}
}

func (d *Dialog) measureWidthLocked() int {
	w := 0
	// Title
	if len(d.title) > 0 {
		w = runeLen(d.title) + 4 // padding + borders
	}
	// Message lines
	for _, line := range splitLines(d.message) {
		lw := runeLen(line) + 4
		if lw > w {
			w = lw
		}
	}
	// Input field width
	if d.dialogType == DialogPrompt {
		inputW := 20 + 4 // min 20 chars + borders
		if inputW > w {
			w = inputW
		}
	}
	// Buttons
	btnW := d.measureButtonsWidthLocked() + 4
	if btnW > w {
		w = btnW
	}
	if w < 20 {
		w = 20
	}
	return w
}

func (d *Dialog) measureHeightLocked() int {
	h := 2 // top + bottom border
	if len(d.title) > 0 {
		h++ // title bar
	}
	h += len(splitLines(d.message)) // message lines
	if d.dialogType == DialogPrompt {
		h += 2 // input box (border + content)
	}
	h += 2 // button row + padding
	return h
}

func (d *Dialog) measureButtonsWidthLocked() int {
	total := 0
	for i, btn := range d.buttons {
		total += runeLen(btn.Label)
		if i < len(d.buttons)-1 {
			total += 3 // " │ " separator
		}
	}
	if total == 0 {
		total = 10
	}
	return total
}

// Paint renders the dialog into the buffer at its Bounds position.
func (d *Dialog) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.visible {
		return
	}

	b := d.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w <= 0 {
		w = d.width
	}
	if w < 20 {
		w = 20
	}

	put := func(px, py int, r rune, s buffer.Style) {
		buf.SetCell(px, py, buffer.NewCell(r, s))
	}

	bs := d.style.Border
	py := y

	// ─── Top border ───
	put(x, py, '┌', bs)
	for i := x + 1; i < x+w-1; i++ {
		put(i, py, '─', bs)
	}
	put(x+w-1, py, '┐', bs)
	py++

	// ─── Title bar ───
	if len(d.title) > 0 {
		put(x, py, '│', d.style.TitleBar)
		titleRunes := []rune(d.title)
		tx := x + 2
		for _, r := range titleRunes {
			if tx < x+w-1 {
				put(tx, py, r, d.style.Title)
				tx++
			}
		}
		// Fill rest of title bar
		for i := tx; i < x+w-1; i++ {
			put(i, py, ' ', d.style.TitleBar)
		}
		put(x+w-1, py, '│', bs)

		// Separator line under title
		py++
		put(x, py, '├', bs)
		for i := x + 1; i < x+w-1; i++ {
			put(i, py, '─', bs)
		}
		put(x+w-1, py, '┤', bs)
		py++
	}

	// ─── Message lines ───
	for _, line := range splitLines(d.message) {
		put(x, py, '│', bs)
		mx := x + 2
		for _, r := range line {
			if mx < x+w-1 {
				put(mx, py, r, d.style.Message)
				mx++
			}
		}
		put(x+w-1, py, '│', bs)
		py++
	}

	// ─── Input field ───
	if d.dialogType == DialogPrompt {
		py++ // blank line before input
		put(x, py, '│', bs)
		// Input border (simple bracket style)
		ix := x + 2
		inputW := w - 4 // content width
		if inputW < 1 {
			inputW = 1
		}

		// Render input text
		for i := 0; i < inputW; i++ {
			ixp := ix + i
			if ixp >= x+w-1 {
				break
			}
			if i < len(d.input) {
				put(ixp, py, d.input[i], d.style.Input)
			} else if i == d.inputCursor {
				put(ixp, py, ' ', d.style.InputCursor)
			} else {
				put(ixp, py, ' ', d.style.Input)
			}
		}
		// Draw cursor if at end
		if d.inputCursor >= inputW && d.inputCursor <= len(d.input) {
			// Cursor is off-screen; just draw last visible char
		} else if d.inputCursor == len(d.input) && d.inputCursor < inputW {
			// Already drawn above
		} else if d.inputCursor < len(d.input) && d.inputCursor < inputW {
			// Overwrite cursor position with cursor style
			cx := ix + d.inputCursor
			if cx < x+w-1 {
				put(cx, py, d.input[d.inputCursor], d.style.InputCursor)
			}
		}
		put(x+w-1, py, '│', bs)
		py++
	}

	// ─── Button row ───
	py++ // blank line before buttons
	put(x, py, '│', bs)
	put(x+w-1, py, '│', bs)

	// Center buttons
	btnTotalW := d.measureButtonsWidthLocked()
	bx := x + (w-btnTotalW)/2
	if bx < x+1 {
		bx = x + 1
	}

	for i, btn := range d.buttons {
		st := d.style.Button
		if i == d.cursor {
			st = d.style.ButtonActive
		}
		// Button brackets
		if bx < x+w-1 {
			put(bx, py, '[', st)
			bx++
		}
		for _, r := range btn.Label {
			if bx < x+w-1 {
				put(bx, py, r, st)
				bx++
			}
		}
		if bx < x+w-1 {
			put(bx, py, ']', st)
			bx++
		}
		// Separator between buttons
		if i < len(d.buttons)-1 {
			if bx < x+w-1 {
				put(bx, py, ' ', d.style.ButtonSeparator)
				bx++
			}
		}
	}

	py++

	// ─── Bottom border ───
	put(x, py, '└', bs)
	for i := x + 1; i < x+w-1; i++ {
		put(i, py, '─', bs)
	}
	put(x+w-1, py, '┘', bs)
}

// Children returns nil — Dialog is a leaf component.
func (d *Dialog) Children() []Component { return nil }

// ─── String ────────────────────────────────────────────────────

// String returns a human-readable description.
func (d *Dialog) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return "Dialog{type=" + dialogTypeString(d.dialogType) + ", title=\"" + d.title + "\"}"
}

// ─── Helpers ───────────────────────────────────────────────────

func dialogTypeString(dt DialogType) string {
	switch dt {
	case DialogInfo:
		return "Info"
	case DialogConfirm:
		return "Confirm"
	case DialogPrompt:
		return "Prompt"
	case DialogCustom:
		return "Custom"
	}
	return "Unknown"
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}
