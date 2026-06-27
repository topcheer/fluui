package component

import (
	"errors"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === TextField Tests ===

func TestTextFieldDefaults(t *testing.T) {
	f := NewTextField("Name", "name", "default")
	if f.Label() != "Name" {
		t.Errorf("Label = %q, want 'Name'", f.Label())
	}
	if f.Key() != "name" {
		t.Errorf("Key = %q, want 'name'", f.Key())
	}
	if f.Value() != "default" {
		t.Errorf("Value = %q, want 'default'", f.Value())
	}
}

func TestTextFieldEmpty(t *testing.T) {
	f := NewTextField("Name", "name", "")
	if f.Value() != "" {
		t.Errorf("Value = %q, want empty", f.Value())
	}
}

func TestTextFieldTypeChars(t *testing.T) {
	f := NewTextField("Name", "name", "")
	f.HandleKey(&term.KeyEvent{Rune: 'h'})
	f.HandleKey(&term.KeyEvent{Rune: 'i'})
	if f.Value() != "hi" {
		t.Errorf("Value = %q, want 'hi'", f.Value())
	}
}

func TestTextFieldBackspace(t *testing.T) {
	f := NewTextField("Name", "name", "hello")
	f.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	f.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	f.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if f.Value() != "hel" {
		t.Errorf("Value = %q, want 'hel'", f.Value())
	}
}

func TestTextFieldArrowNavigation(t *testing.T) {
	f := NewTextField("Name", "name", "abc")
	f.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	f.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	f.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	f.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) // clamped
	f.HandleKey(&term.KeyEvent{Rune: 'X'})
	if f.Value() != "Xabc" {
		t.Errorf("Value = %q, want 'Xabc'", f.Value())
	}
}

func TestTextFieldHomeEnd(t *testing.T) {
	f := NewTextField("Name", "name", "hello")
	f.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	f.HandleKey(&term.KeyEvent{Rune: 'X'})
	if f.Value() != "Xhello" {
		t.Errorf("Value = %q, want 'Xhello'", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	f.HandleKey(&term.KeyEvent{Rune: '!'})
	if f.Value() != "Xhello!" {
		t.Errorf("Value = %q, want 'Xhello!'", f.Value())
	}
}

func TestTextFieldValidate(t *testing.T) {
	f := NewTextField("Name", "name", "")
	f.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})
	if err := f.Validate(); err == nil {
		t.Error("Validate should fail on empty")
	}
}

func TestTextFieldValidatePass(t *testing.T) {
	f := NewTextField("Name", "name", "ok")
	f.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})
	if err := f.Validate(); err != nil {
		t.Errorf("Validate should pass: %v", err)
	}
}

func TestTextFieldPaint(t *testing.T) {
	f := NewTextField("Name", "name", "hello")
	buf := buffer.NewBuffer(40, 3)
	f.Paint(buf, 0, 0, 20, true)
	f.Paint(buf, 0, 1, 20, false)
}

func TestTextFieldRequired(t *testing.T) {
	f := NewTextField("Name", "name", "")
	f.SetRequired()
	if err := f.Validate(); err == nil {
		t.Error("Required field should fail validation when empty")
	}
}

func TestTextFieldRequiredFilled(t *testing.T) {
	f := NewTextField("Name", "name", "Alice")
	f.SetRequired()
	if err := f.Validate(); err != nil {
		t.Errorf("Required field with value should pass: %v", err)
	}
}

func TestTextFieldReset(t *testing.T) {
	f := NewTextField("Name", "name", "default")
	f.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	f.HandleKey(&term.KeyEvent{Rune: 'X'})
	f.Reset()
	if f.Value() != "default" {
		t.Errorf("Value = %q, want 'default' after reset", f.Value())
	}
}

// === CheckboxField Tests ===

func TestCheckboxDefaults(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", false)
	if f.Label() != "Agree" {
		t.Errorf("Label = %q", f.Label())
	}
	if f.IsChecked() {
		t.Error("Should be unchecked")
	}
	if f.Value() != "false" {
		t.Errorf("Value = %q, want 'false'", f.Value())
	}
}

func TestCheckboxToggle(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", false)
	f.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if !f.IsChecked() {
		t.Error("Should be checked after Space")
	}
	if f.Value() != "true" {
		t.Errorf("Value = %q, want 'true'", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if f.IsChecked() {
		t.Error("Should be unchecked after second Space")
	}
}

func TestCheckboxYN(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", false)
	f.HandleKey(&term.KeyEvent{Rune: 'y'})
	if !f.IsChecked() {
		t.Error("Should be checked after 'y'")
	}
	f.HandleKey(&term.KeyEvent{Rune: 'n'})
	if f.IsChecked() {
		t.Error("Should be unchecked after 'n'")
	}
}

func TestCheckboxArrowToggle(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", false)
	f.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !f.IsChecked() {
		t.Error("Should toggle on arrow")
	}
}

func TestCheckboxPaint(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", true)
	buf := buffer.NewBuffer(20, 2)
	f.Paint(buf, 0, 0, 3, true)
	f.Paint(buf, 0, 1, 3, false)
}

func TestCheckboxReset(t *testing.T) {
	f := NewCheckboxField("Agree", "agree", false)
	f.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	f.Reset()
	if f.IsChecked() {
		t.Error("Should be unchecked after reset")
	}
}

// === SelectField Tests ===

func TestSelectDefaults(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green", "blue"})
	if f.Value() != "red" {
		t.Errorf("Value = %q, want 'red'", f.Value())
	}
	if f.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", f.SelectedIndex())
	}
}

func TestSelectNavigation(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green", "blue"})
	f.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if f.Value() != "green" {
		t.Errorf("Value = %q, want 'green'", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if f.Value() != "blue" {
		t.Errorf("Value = %q, want 'blue'", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if f.Value() != "red" {
		t.Errorf("Value = %q, want 'red' (wrap)", f.Value())
	}
}

func TestSelectUpWrap(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green", "blue"})
	f.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if f.Value() != "blue" {
		t.Errorf("Value = %q, want 'blue' (wrap up)", f.Value())
	}
}

func TestSelectHomeEnd(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green", "blue"})
	f.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if f.SelectedIndex() != 2 {
		t.Errorf("End: SelectedIndex = %d, want 2", f.SelectedIndex())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if f.SelectedIndex() != 0 {
		t.Errorf("Home: SelectedIndex = %d, want 0", f.SelectedIndex())
	}
}

func TestSelectSetIndex(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"a", "b", "c"})
	f.SetSelectedIndex(2)
	if f.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2", f.SelectedIndex())
	}
	f.SetSelectedIndex(99)
	if f.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2 (clamped)", f.SelectedIndex())
	}
	f.SetSelectedIndex(-1)
	if f.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", f.SelectedIndex())
	}
}

func TestSelectPaint(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green"})
	buf := buffer.NewBuffer(20, 2)
	f.Paint(buf, 0, 0, 15, true)
	f.Paint(buf, 0, 1, 15, false)
}

func TestSelectReset(t *testing.T) {
	f := NewSelectField("Color", "color", []string{"red", "green", "blue"})
	f.SetDefault(2)
	f.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	f.Reset()
	if f.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2 after reset", f.SelectedIndex())
	}
}

// === Form Tests ===

func TestFormEmpty(t *testing.T) {
	f := NewForm()
	if f.FieldCount() != 0 {
		t.Errorf("FieldCount = %d, want 0", f.FieldCount())
	}
	if f.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0", f.ActiveIndex())
	}
}

func TestFormAddTextField(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "")
	if f.FieldCount() != 1 {
		t.Errorf("FieldCount = %d, want 1", f.FieldCount())
	}
	if tf == nil {
		t.Error("AddTextField returned nil")
	}
}

func TestFormAddCheckboxField(t *testing.T) {
	f := NewForm()
	cf := f.AddCheckboxField("Agree", "agree", false)
	if f.FieldCount() != 1 {
		t.Errorf("FieldCount = %d, want 1", f.FieldCount())
	}
	if cf == nil {
		t.Error("AddCheckboxField returned nil")
	}
}

func TestFormAddSelectField(t *testing.T) {
	f := NewForm()
	sf := f.AddSelectField("Color", "color", []string{"red", "green"})
	if f.FieldCount() != 1 {
		t.Errorf("FieldCount = %d, want 1", f.FieldCount())
	}
	if sf == nil {
		t.Error("AddSelectField returned nil")
	}
}

func TestFormAddMixedFields(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "test")
	f.AddCheckboxField("Active", "active", true)
	f.AddSelectField("Role", "role", []string{"admin", "user"})
	if f.FieldCount() != 3 {
		t.Errorf("FieldCount = %d, want 3", f.FieldCount())
	}
}

func TestFormValues(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "Alice")
	f.AddCheckboxField("Active", "active", true)
	f.AddSelectField("Role", "role", []string{"admin", "user"})

	vals := f.Values()
	if vals["name"] != "Alice" {
		t.Errorf("name = %q, want 'Alice'", vals["name"])
	}
	if vals["active"] != "true" {
		t.Errorf("active = %q, want 'true'", vals["active"])
	}
	if vals["role"] != "admin" {
		t.Errorf("role = %q, want 'admin'", vals["role"])
	}
}

func TestFormTabNavigation(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.AddTextField("C", "c", "")

	if f.ActiveIndex() != 0 {
		t.Fatalf("Initial ActiveIndex = %d, want 0", f.ActiveIndex())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if f.ActiveIndex() != 1 {
		t.Errorf("After Tab: ActiveIndex = %d, want 1", f.ActiveIndex())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if f.ActiveIndex() != 2 {
		t.Errorf("After Tab: ActiveIndex = %d, want 2", f.ActiveIndex())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if f.ActiveIndex() != 0 {
		t.Errorf("After wrap Tab: ActiveIndex = %d, want 0", f.ActiveIndex())
	}
}

func TestFormShiftTabNavigation(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.AddTextField("C", "c", "")

	f.HandleKey(&term.KeyEvent{Key: term.KeyBacktab})
	if f.ActiveIndex() != 2 {
		t.Errorf("After Backtab: ActiveIndex = %d, want 2", f.ActiveIndex())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyBacktab})
	if f.ActiveIndex() != 1 {
		t.Errorf("After Backtab: ActiveIndex = %d, want 1", f.ActiveIndex())
	}
}

func TestFormEscapeCancel(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	cancelled := false
	f.OnCancel(func() { cancelled = true })

	f.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !cancelled {
		t.Error("Cancel should have been called on Escape")
	}
}

func TestFormEnterSubmit(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "Alice")

	submitted := false
	f.OnSubmit(func(values map[string]string) error {
		submitted = true
		if values["name"] != "Alice" {
			t.Errorf("values['name'] = %q, want 'Alice'", values["name"])
		}
		return nil
	})

	f.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !submitted {
		t.Error("Submit should have been called on Enter")
	}
	if !f.IsSubmitted() {
		t.Error("Form should be marked as submitted")
	}
}

func TestFormSubmitBlockedByValidation(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "")
	tf.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	submitted := false
	f.OnSubmit(func(values map[string]string) error {
		submitted = true
		return nil
	})

	f.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if submitted {
		t.Error("Should NOT submit when validation fails")
	}
	if f.IsSubmitted() {
		t.Error("Should NOT be marked submitted when validation fails")
	}
}

func TestFormValidate(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "")
	tf.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	errs := f.Validate()
	if errs == nil {
		t.Fatal("Validate should return errors for invalid form")
	}
	if len(errs) != 1 {
		t.Fatalf("errs len = %d, want 1", len(errs))
	}
	if errs["name"] == nil {
		t.Error("Should have error for 'name' key")
	}
}

func TestFormValidateAllPass(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "Alice")
	tf.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	errs := f.Validate()
	if errs != nil {
		t.Errorf("Validate should return nil for valid form, got %v", errs)
	}
}

func TestFormValidateMultiple(t *testing.T) {
	f := NewForm()
	nameField := f.AddTextField("Name", "name", "")
	nameField.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})
	emailField := f.AddTextField("Email", "email", "")
	emailField.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	errs := f.Validate()
	if errs == nil {
		t.Fatal("Should have errors")
	}
	if len(errs) != 2 {
		t.Errorf("errs len = %d, want 2", len(errs))
	}
}

func TestFormFieldInput(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "")
	f.HandleKey(&term.KeyEvent{Rune: 'h'})
	f.HandleKey(&term.KeyEvent{Rune: 'i'})

	vals := f.Values()
	if vals["name"] != "hi" {
		t.Errorf("name = %q, want 'hi'", vals["name"])
	}
}

func TestFormFieldInputAfterTab(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.HandleKey(&term.KeyEvent{Rune: 'x'})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	f.HandleKey(&term.KeyEvent{Rune: 'y'})

	vals := f.Values()
	if vals["a"] != "x" {
		t.Errorf("a = %q, want 'x'", vals["a"])
	}
	if vals["b"] != "y" {
		t.Errorf("b = %q, want 'y'", vals["b"])
	}
}

func TestFormReset(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	f.Reset()

	if f.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0 after Reset", f.ActiveIndex())
	}
	if f.IsSubmitted() {
		t.Error("Should not be submitted after Reset")
	}
}

func TestFormFocusNextPrev(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.AddTextField("C", "c", "")

	f.FocusNext()
	if f.ActiveIndex() != 1 {
		t.Errorf("After FocusNext: ActiveIndex = %d, want 1", f.ActiveIndex())
	}
	f.FocusNext()
	if f.ActiveIndex() != 2 {
		t.Errorf("After FocusNext: ActiveIndex = %d, want 2", f.ActiveIndex())
	}
	f.FocusPrev()
	if f.ActiveIndex() != 1 {
		t.Errorf("After FocusPrev: ActiveIndex = %d, want 1", f.ActiveIndex())
	}
}

func TestFormFocusWrap(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")

	f.FocusNext()
	f.FocusNext()
	if f.ActiveIndex() != 0 {
		t.Errorf("After wrap FocusNext: ActiveIndex = %d, want 0", f.ActiveIndex())
	}
	f.FocusPrev()
	if f.ActiveIndex() != 1 {
		t.Errorf("After wrap FocusPrev: ActiveIndex = %d, want 1", f.ActiveIndex())
	}
}

func TestFormMeasure(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "")
	f.AddTextField("Email", "email", "")

	size := f.Measure(Constraints{})
	if size.W < 10 {
		t.Errorf("Width = %d, too small", size.W)
	}
	if size.H < 2 {
		t.Errorf("Height = %d, want >= 2", size.H)
	}
}

func TestFormMeasureClamped(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "")
	f.AddTextField("Email", "email", "")

	size := f.Measure(Constraints{MaxWidth: 15, MaxHeight: 1})
	if size.W > 15 {
		t.Errorf("Width = %d, should be <= 15", size.W)
	}
	if size.H > 1 {
		t.Errorf("Height = %d, should be <= 1", size.H)
	}
}

func TestFormMeasureEmpty(t *testing.T) {
	f := NewForm()
	size := f.Measure(Constraints{})
	if size.H < 1 {
		t.Errorf("Empty form Height = %d, want >= 1", size.H)
	}
}

func TestFormPaint(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "Alice")
	f.AddCheckboxField("Active", "active", true)
	f.AddSelectField("Role", "role", []string{"admin", "user"})

	f.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	f.Paint(buf)
}

func TestFormPaintWithErrors(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "")
	tf.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	f.Validate()
	f.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	f.Paint(buf)
}

func TestFormPaintTooSmall(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "Alice")
	f.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	f.Paint(buf)
}

func TestFormHandleKeyNil(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	if f.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestFormSetActiveIndex(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.AddTextField("C", "c", "")

	f.SetActiveIndex(2)
	if f.ActiveIndex() != 2 {
		t.Errorf("ActiveIndex = %d, want 2", f.ActiveIndex())
	}
	f.SetActiveIndex(-5)
	if f.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0 (clamped)", f.ActiveIndex())
	}
	f.SetActiveIndex(99)
	if f.ActiveIndex() != 2 {
		t.Errorf("ActiveIndex = %d, want 2 (clamped)", f.ActiveIndex())
	}
}

func TestFormString(t *testing.T) {
	f := NewForm()
	f.AddTextField("Name", "name", "Alice")
	f.AddCheckboxField("Active", "active", true)

	str := f.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestFormChildren(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	if children := f.Children(); children != nil {
		t.Errorf("Children() = %v, want nil", children)
	}
}

func TestFormAddCustomField(t *testing.T) {
	f := NewForm()
	custom := NewTextField("Custom", "custom", "val")
	f.AddField(custom)
	if f.FieldCount() != 1 {
		t.Errorf("FieldCount = %d, want 1", f.FieldCount())
	}
	if f.Values()["custom"] != "val" {
		t.Errorf("custom = %q, want 'val'", f.Values()["custom"])
	}
}

// === Concurrency Tests ===

func TestFormConcurrentAccess(t *testing.T) {
	f := NewForm()
	f.AddTextField("A", "a", "")
	f.AddTextField("B", "b", "")
	f.AddTextField("C", "c", "")

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			_ = f.Values()
			_ = f.ActiveIndex()
			_ = f.FieldCount()
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			f.FocusNext()
		}
		done <- true
	}()
	<-done
	<-done
}

func TestFormErrors(t *testing.T) {
	f := NewForm()
	tf := f.AddTextField("Name", "name", "")
	tf.SetValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	})

	f.Validate()
	errs := f.Errors()
	if len(errs) != 1 {
		t.Errorf("Errors len = %d, want 1", len(errs))
	}
}
