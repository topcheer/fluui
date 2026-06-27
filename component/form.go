package component

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ErrValidation is the sentinel validation error.
var ErrValidation = errors.New("validation failed")

// --- FormField interface ---

// FormField is a single editable field within a form.
type FormField interface {
	// Label returns the display label.
	Label() string
	// Key returns the unique identifier for this field.
	Key() string
	// Value returns the current value as a string.
	Value() string
	// Validate returns an error if the current value is invalid, nil otherwise.
	Validate() error
	// HandleKey processes a key event. Returns true if consumed.
	HandleKey(key *term.KeyEvent) bool
	// Paint draws the field value at (x, y) within width w.
	// focused indicates whether this field has focus (for highlighting).
	Paint(buf *buffer.Buffer, x, y, w int, focused bool)
	// Reset restores the field to its default value.
	Reset()
}

// --- TextField ---

// TextField is a single-line text input field.
type TextField struct {
	label     string
	key       string
	value     []rune
	cursor    int
	defValue  string
	maxLen    int // max length (0 = unlimited)
	required  bool
	validator func(string) error
}

// NewTextField creates a text input field.
func NewTextField(label, key, defValue string) *TextField {
	return &TextField{
		label:    label,
		key:      key,
		value:    []rune(defValue),
		cursor:   len([]rune(defValue)),
		defValue: defValue,
	}
}

// SetMaxLength limits the number of characters that can be typed.
func (f *TextField) SetMaxLength(n int) { f.maxLen = n }

// SetRequired marks this field as required.
func (f *TextField) SetRequired() { f.required = true }

// SetValidator sets a custom validation function.
func (f *TextField) SetValidator(fn func(string) error) { f.validator = fn }

// Label returns the display label.
func (f *TextField) Label() string { return f.label }

// Key returns the unique field key.
func (f *TextField) Key() string { return f.key }

// Value returns the current text value.
func (f *TextField) Value() string { return string(f.value) }

// Reset restores the default value.
func (f *TextField) Reset() {
	f.value = []rune(f.defValue)
	f.cursor = len(f.value)
}

// HandleKey processes keyboard input for the text field.
func (f *TextField) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	switch {
	case key.Key == term.KeyBackspace:
		if f.cursor > 0 {
			f.value = append(f.value[:f.cursor-1], f.value[f.cursor:]...)
			f.cursor--
		}
		return true

	case key.Key == term.KeyLeft:
		if f.cursor > 0 {
			f.cursor--
		}
		return true

	case key.Key == term.KeyRight:
		if f.cursor < len(f.value) {
			f.cursor++
		}
		return true

	case key.Key == term.KeyHome:
		f.cursor = 0
		return true

	case key.Key == term.KeyEnd:
		f.cursor = len(f.value)
		return true

	case key.Rune != 0 && key.Rune >= 0x20 && key.Modifiers&term.ModCtrl == 0:
		if f.maxLen > 0 && len(f.value) >= f.maxLen {
			return true // at max length
		}
		f.value = append(f.value[:f.cursor], append([]rune{key.Rune}, f.value[f.cursor:]...)...)
		f.cursor++
		return true
	}

	return false
}

// Validate checks if the field has a value (if required) or passes custom validation.
func (f *TextField) Validate() error {
	if f.validator != nil {
		return f.validator(string(f.value))
	}
	if f.required && len(f.value) == 0 {
		return fmt.Errorf("%s is required", f.label)
	}
	return nil
}

// Paint draws the text field.
func (f *TextField) Paint(buf *buffer.Buffer, x, y, w int, focused bool) {
	t := theme.Get()
	style := buffer.Style{Fg: t.Fg, Bg: t.Bg}
	if focused {
		style.Fg = t.Accent
		style.Flags = buffer.Bold
	}

	// Draw value text (clamped to width)
	text := string(f.value)
	if buffer.StringWidth(text) > w {
		start := 0
		for buffer.StringWidth(text[start:]) > w-1 && start < f.cursor {
			start++
		}
		text = text[start:]
	}
	buf.DrawTextClamped(x, y, text, style)

	// Draw cursor
	if focused && f.cursor >= 0 {
		cursorX := x + buffer.StringWidth(string(f.value[:clampInt(f.cursor, len(f.value))]))
		if cursorX < x+w && cursorX < buf.Width {
			buf.SetCell(cursorX, y, buffer.Cell{
				Rune:  '_',
				Width: 1,
				Fg:    t.Accent,
				Bg:    t.Bg,
				Flags: buffer.Bold,
			})
		}
	}
}

// --- CheckboxField ---

// CheckboxField is a toggle field showing [x] or [ ].
type CheckboxField struct {
	label    string
	key      string
	checked  bool
	defValue bool
}

// NewCheckboxField creates a checkbox field.
func NewCheckboxField(label, key string, checked bool) *CheckboxField {
	return &CheckboxField{label: label, key: key, checked: checked, defValue: checked}
}

// Label returns the display label.
func (f *CheckboxField) Label() string { return f.label }

// Key returns the unique field key.
func (f *CheckboxField) Key() string { return f.key }

// Value returns "true" or "false".
func (f *CheckboxField) Value() string {
	if f.checked {
		return "true"
	}
	return "false"
}

// IsChecked returns the boolean state.
func (f *CheckboxField) IsChecked() bool { return f.checked }

// Reset restores the default value.
func (f *CheckboxField) Reset() { f.checked = f.defValue }

// Validate always passes for checkboxes.
func (f *CheckboxField) Validate() error { return nil }

// HandleKey processes keyboard input for the checkbox.
func (f *CheckboxField) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	if key.Key == term.KeySpace || key.Rune == ' ' {
		f.checked = !f.checked
		return true
	}
	if key.Rune == 'y' || key.Rune == 'Y' {
		f.checked = true
		return true
	}
	if key.Rune == 'n' || key.Rune == 'N' {
		f.checked = false
		return true
	}
	if key.Key == term.KeyLeft || key.Key == term.KeyRight {
		f.checked = !f.checked
		return true
	}
	return false
}

// Paint draws the checkbox field.
func (f *CheckboxField) Paint(buf *buffer.Buffer, x, y, w int, focused bool) {
	t := theme.Get()
	style := buffer.Style{Fg: t.Fg, Bg: t.Bg}
	if focused {
		style.Fg = t.Accent
		style.Flags = buffer.Bold
	}

	mark := "[ ] "
	if f.checked {
		mark = "[x] "
	}
	buf.DrawTextClamped(x, y, mark, style)

	if focused {
		hintStyle := buffer.Style{Fg: t.Muted, Bg: t.Bg}
		buf.DrawTextClamped(x+4, y, "(space to toggle)", hintStyle)
	}
}

// --- SelectField ---

// SelectField is a dropdown-style option selector.
type SelectField struct {
	label    string
	key      string
	options  []string
	selected int
	defIndex int
}

// NewSelectField creates a select field.
func NewSelectField(label, key string, options []string) *SelectField {
	return &SelectField{
		label:   label,
		key:     key,
		options: options,
	}
}

// Label returns the display label.
func (f *SelectField) Label() string { return f.label }

// Key returns the unique field key.
func (f *SelectField) Key() string { return f.key }

// Value returns the currently selected option.
func (f *SelectField) Value() string {
	if len(f.options) == 0 || f.selected < 0 || f.selected >= len(f.options) {
		return ""
	}
	return f.options[f.selected]
}

// SelectedIndex returns the index of the selected option.
func (f *SelectField) SelectedIndex() int { return f.selected }

// SetSelectedIndex sets the selection. Clamped to valid range.
func (f *SelectField) SetSelectedIndex(i int) {
	if len(f.options) == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(f.options) {
		i = len(f.options) - 1
	}
	f.selected = i
}

// SetDefault sets the default selected index.
func (f *SelectField) SetDefault(index int) {
	if index >= 0 && index < len(f.options) {
		f.defIndex = index
		f.selected = index
	}
}

// Reset restores the default selection.
func (f *SelectField) Reset() { f.selected = f.defIndex }

// Validate passes if there are options.
func (f *SelectField) Validate() error {
	if len(f.options) == 0 {
		return fmt.Errorf("%s has no options", f.label)
	}
	return nil
}

// HandleKey processes keyboard input for the select field.
func (f *SelectField) HandleKey(key *term.KeyEvent) bool {
	if key == nil || len(f.options) == 0 {
		return false
	}

	switch key.Key {
	case term.KeyUp, term.KeyLeft:
		if f.selected > 0 {
			f.selected--
		} else {
			f.selected = len(f.options) - 1
		}
		return true

	case term.KeyDown, term.KeyRight:
		if f.selected < len(f.options)-1 {
			f.selected++
		} else {
			f.selected = 0
		}
		return true

	case term.KeyHome:
		f.selected = 0
		return true

	case term.KeyEnd:
		f.selected = len(f.options) - 1
		return true
	}

	return false
}

// Paint draws the select field.
func (f *SelectField) Paint(buf *buffer.Buffer, x, y, w int, focused bool) {
	t := theme.Get()
	style := buffer.Style{Fg: t.Fg, Bg: t.Bg}
	if focused {
		style.Fg = t.Accent
		style.Flags = buffer.Bold
	}

	current := f.Value()
	if current == "" {
		current = "(none)"
	}
	display := fmt.Sprintf("< %s >", current)
	buf.DrawTextClamped(x, y, display, style)

	if focused && len(f.options) > 0 {
		hint := fmt.Sprintf("(%d/%d)", f.selected+1, len(f.options))
		hintStyle := buffer.Style{Fg: t.Muted, Bg: t.Bg}
		buf.DrawTextClamped(x+buffer.StringWidth(display)+1, y, hint, hintStyle)
	}
}

// --- Form ---

// Form is a component that manages a collection of form fields with
// keyboard navigation (Tab/Shift+Tab), validation, and submit/cancel callbacks.
type Form struct {
	BaseComponent
	mu               sync.RWMutex
	fields           []FormField
	focusIdx         int
	submitted        bool
	cancelled        bool
	onSubmit         func(values map[string]string) error
	onCancel         func()
	validationErrors map[string]error
}

// NewForm creates an empty form with no fields.
func NewForm() *Form {
	return &Form{
		focusIdx:         0,
		validationErrors: make(map[string]error),
	}
}

// AddTextField adds a text input field and returns it for configuration.
func (f *Form) AddTextField(label, key, defValue string) *TextField {
	f.mu.Lock()
	defer f.mu.Unlock()
	tf := NewTextField(label, key, defValue)
	f.fields = append(f.fields, tf)
	return tf
}

// AddCheckboxField adds a checkbox field and returns it for configuration.
func (f *Form) AddCheckboxField(label, key string, checked bool) *CheckboxField {
	f.mu.Lock()
	defer f.mu.Unlock()
	cf := NewCheckboxField(label, key, checked)
	f.fields = append(f.fields, cf)
	return cf
}

// AddSelectField adds a select field and returns it for configuration.
func (f *Form) AddSelectField(label, key string, options []string) *SelectField {
	f.mu.Lock()
	defer f.mu.Unlock()
	sf := NewSelectField(label, key, options)
	f.fields = append(f.fields, sf)
	return sf
}

// AddField adds a custom FormField implementation.
func (f *Form) AddField(field FormField) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fields = append(f.fields, field)
}

// FieldCount returns the number of fields.
func (f *Form) FieldCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.fields)
}

// ActiveIndex returns the currently focused field index.
func (f *Form) ActiveIndex() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.focusIdx
}

// SetActiveIndex sets focus to the given index (clamped).
func (f *Form) SetActiveIndex(idx int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(f.fields) {
		idx = len(f.fields) - 1
	}
	f.focusIdx = idx
}

// IsSubmitted reports whether the form was submitted.
func (f *Form) IsSubmitted() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.submitted
}

// IsCancelled reports whether the form was cancelled.
func (f *Form) IsCancelled() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cancelled
}

// OnSubmit sets the submit callback.
func (f *Form) OnSubmit(fn func(values map[string]string) error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onSubmit = fn
}

// OnCancel sets the cancel callback.
func (f *Form) OnCancel(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onCancel = fn
}

// Values returns all field values as key→value.
func (f *Form) Values() map[string]string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	values := make(map[string]string, len(f.fields))
	for _, field := range f.fields {
		values[field.Key()] = field.Value()
	}
	return values
}

// Validate validates all fields and returns a map of key→error (nil if all valid).
func (f *Form) Validate() map[string]error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validationErrors = make(map[string]error)
	for _, field := range f.fields {
		if err := field.Validate(); err != nil {
			f.validationErrors[field.Key()] = err
		}
	}
	if len(f.validationErrors) == 0 {
		return nil
	}
	return f.validationErrors
}

// Errors returns the current validation errors.
func (f *Form) Errors() map[string]error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.validationErrors) == 0 {
		return nil
	}
	result := make(map[string]error, len(f.validationErrors))
	for k, v := range f.validationErrors {
		result[k] = v
	}
	return result
}

// Reset resets all fields to defaults and clears state.
func (f *Form) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, field := range f.fields {
		field.Reset()
	}
	f.focusIdx = 0
	f.submitted = false
	f.cancelled = false
	f.validationErrors = make(map[string]error)
}

// FocusNext moves focus to the next field (wraps around).
func (f *Form) FocusNext() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.fields) == 0 {
		return
	}
	f.focusIdx = (f.focusIdx + 1) % len(f.fields)
}

// FocusPrev moves focus to the previous field (wraps around).
func (f *Form) FocusPrev() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.fields) == 0 {
		return
	}
	f.focusIdx = (f.focusIdx - 1 + len(f.fields)) % len(f.fields)
}

// HandleKey processes keyboard input for the form.
func (f *Form) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	f.mu.Lock()

	// Tab → next field
	if key.Key == term.KeyTab {
		if len(f.fields) > 0 {
			f.focusIdx = (f.focusIdx + 1) % len(f.fields)
		}
		f.mu.Unlock()
		return true
	}

	// Backtab / Shift+Tab → prev field
	if key.Key == term.KeyBacktab || (key.Key == term.KeyTab && key.Modifiers&term.ModShift != 0) {
		if len(f.fields) > 0 {
			f.focusIdx = (f.focusIdx - 1 + len(f.fields)) % len(f.fields)
		}
		f.mu.Unlock()
		return true
	}

	// Escape → cancel
	if key.Key == term.KeyEscape {
		f.cancelled = true
		cb := f.onCancel
		f.mu.Unlock()
		if cb != nil {
			cb()
		}
		return true
	}

	// Enter → submit (if validation passes)
	if key.Key == term.KeyEnter {
		// Let focused field try to handle Enter first
		if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
			if f.fields[f.focusIdx].HandleKey(key) {
				f.mu.Unlock()
				return true
			}
		}

		// Validate all fields
		f.validationErrors = make(map[string]error)
		for _, field := range f.fields {
			if err := field.Validate(); err != nil {
				f.validationErrors[field.Key()] = err
			}
		}
		if len(f.validationErrors) > 0 {
			f.mu.Unlock()
			return true // validation failed, stay open
		}

		f.submitted = true
		cb := f.onSubmit
		values := make(map[string]string, len(f.fields))
		for _, field := range f.fields {
			values[field.Key()] = field.Value()
		}
		f.mu.Unlock()

		if cb != nil {
			_ = cb(values)
		}
		return true
	}

	// Route to focused field
	var consumed bool
	if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
		consumed = f.fields[f.focusIdx].HandleKey(key)
	}
	f.mu.Unlock()
	return consumed
}

// Measure returns the preferred size for the form.
func (f *Form) Measure(cs Constraints) Size {
	f.mu.RLock()
	defer f.mu.RUnlock()

	maxW := 0
	h := len(f.fields)
	if h == 0 {
		h = 1
	}

	for _, field := range f.fields {
		lw := buffer.StringWidth(field.Label()) + 2 + 20
		if lw > maxW {
			maxW = lw
		}
	}
	if maxW < 10 {
		maxW = 10
	}

	if cs.HasWidth() && maxW > cs.MaxWidth {
		maxW = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}

	return Size{W: maxW, H: h}
}

// SetBounds sets the position and size of the form.
func (f *Form) SetBounds(r Rect) {
	f.mu.Lock()
	f.bounds = r
	f.mu.Unlock()
}

// Paint renders the form into the buffer.
func (f *Form) Paint(buf *buffer.Buffer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	t := theme.Get()
	bounds := f.bounds

	// Calculate label column width
	labelW := 0
	for _, field := range f.fields {
		lw := buffer.StringWidth(field.Label())
		if lw > labelW {
			labelW = lw
		}
	}
	labelW += 2 // for ": "

	for i, field := range f.fields {
		y := bounds.Y + i
		if y >= bounds.Y+bounds.H || y >= buf.Height {
			break
		}

		focused := i == f.focusIdx

		// Draw label
		labelText := field.Label() + ":"
		labelStyle := buffer.Style{Fg: t.Fg, Bg: t.Bg}
		if focused {
			labelStyle.Fg = t.Accent
			labelStyle.Flags = buffer.Bold
		}
		buf.DrawTextClamped(bounds.X, y, labelText, labelStyle)

		// Draw field
		fieldX := bounds.X + labelW
		fieldW := bounds.W - labelW
		if fieldW < 1 {
			fieldW = 1
		}
		field.Paint(buf, fieldX, y, fieldW, focused)

		// Draw error indicator
		if err, hasErr := f.validationErrors[field.Key()]; hasErr && err != nil {
			errX := fieldX + 20
			if errX < bounds.X+bounds.W-2 {
				errStyle := buffer.Style{Fg: t.Error, Bg: t.Bg}
				buf.DrawTextClamped(errX, y, " !", errStyle)
			}
		}
	}

	// Draw error summary at bottom
	if len(f.validationErrors) > 0 {
		errY := bounds.Y + len(f.fields)
		if errY < bounds.Y+bounds.H && errY < buf.Height {
			errStyle := buffer.Style{Fg: t.Error, Bg: t.Bg}
			lines := make([]string, 0, len(f.validationErrors))
			for k, e := range f.validationErrors {
				lines = append(lines, fmt.Sprintf("  ! %s: %s", k, e.Error()))
			}
			msg := strings.Join(lines, "  |  ")
			buf.DrawTextClamped(bounds.X, errY, msg, errStyle)
		}
	}
}

// Children returns nil (form manages FormFields, not child Components).
func (f *Form) Children() []Component { return nil }

// String returns a debug representation.
func (f *Form) String() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	parts := make([]string, len(f.fields))
	for i, field := range f.fields {
		parts[i] = fmt.Sprintf("%s=%s", field.Key(), field.Value())
	}
	return fmt.Sprintf("Form{%s}", strings.Join(parts, ", "))
}

// --- helpers ---

func clampInt(v, max int) int {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}
