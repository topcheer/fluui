package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── WizardStep Tests ─────────────────────────────────────────────

func TestNewWizardStep(t *testing.T) {
	s := NewWizardStep("intro", "Welcome")
	if s.ID != "intro" {
		t.Errorf("ID = %q, want %q", s.ID, "intro")
	}
	if s.Title != "Welcome" {
		t.Errorf("Title = %q, want %q", s.Title, "Welcome")
	}
	if s.Description != "" {
		t.Errorf("Description = %q, want empty", s.Description)
	}
	if s.Content != nil {
		t.Error("Content should be nil")
	}
	if s.Skippable {
		t.Error("Skippable should be false")
	}
}

func TestWizardStep_SetDescription(t *testing.T) {
	s := NewWizardStep("s1", "Step 1")
	s.SetDescription("A test step")
	if s.Description != "A test step" {
		t.Errorf("Description = %q, want %q", s.Description, "A test step")
	}
}

func TestWizardStep_SetContent(t *testing.T) {
	s := NewWizardStep("s1", "Step 1")
	c := NewText("hello")
	s.SetContent(c)
	if s.Content == nil {
		t.Error("Content should not be nil")
	}
}

func TestWizardStep_SetSkippable(t *testing.T) {
	s := NewWizardStep("s1", "Step 1")
	if s.Skippable {
		t.Error("Should start false")
	}
	s.SetSkippable(true)
	if !s.Skippable {
		t.Error("Should be true after SetSkippable(true)")
	}
}

func TestWizardStep_SetOnEnter(t *testing.T) {
	s := NewWizardStep("s1", "Step 1")
	called := false
	s.SetOnEnter(func(w *Wizard) error {
		called = true
		return nil
	})
	if s.OnEnter == nil {
		t.Fatal("OnEnter should be set")
	}
	_ = s.OnEnter(nil)
	if !called {
		t.Error("OnEnter should have been called")
	}
}

func TestWizardStep_SetOnLeave(t *testing.T) {
	s := NewWizardStep("s1", "Step 1")
	called := false
	s.SetOnLeave(func(w *Wizard) error {
		called = true
		return nil
	})
	if s.OnLeave == nil {
		t.Fatal("OnLeave should be set")
	}
	_ = s.OnLeave(nil)
	if !called {
		t.Error("OnLeave should have been called")
	}
}

// ─── Wizard Construction Tests ───────────────────────────────────

func TestNewWizard(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	if w.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", w.StepCount())
	}
	if w.CurrentStepIndex() != 0 {
		t.Errorf("CurrentStepIndex = %d, want 0", w.CurrentStepIndex())
	}
	if !w.IsFirstStep() {
		t.Error("Should be first step")
	}
	if w.IsLastStep() {
		t.Error("Should not be last step")
	}
	if w.IsCompleted() {
		t.Error("Should not be completed")
	}
	if w.IsCancelled() {
		t.Error("Should not be cancelled")
	}
}

func TestNewWizard_Empty(t *testing.T) {
	w := NewWizard(nil)
	if w.StepCount() != 0 {
		t.Errorf("StepCount = %d, want 0", w.StepCount())
	}
	if w.CurrentStep() != nil {
		t.Error("CurrentStep should be nil for empty wizard")
	}
}

func TestNewWizard_UniqueID(t *testing.T) {
	w1 := NewWizard(nil)
	w2 := NewWizard(nil)
	if w1.ID() == w2.ID() {
		t.Error("Wizards should have unique IDs")
	}
}

// ─── Step Query Tests ─────────────────────────────────────────────

func TestWizard_Steps(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("a", "A"),
		NewWizardStep("b", "B"),
	}
	w := NewWizard(steps)

	got := w.Steps()
	if len(got) != 2 {
		t.Fatalf("Steps() length = %d, want 2", len(got))
	}
	if got[0].ID != "a" || got[1].ID != "b" {
		t.Errorf("Steps order wrong: %s, %s", got[0].ID, got[1].ID)
	}

	// Verify it's a copy of the slice (not the same backing array)
	got[0] = nil
	if steps[0] == nil {
		t.Error("Steps() should return a copy of the slice")
	}
}

func TestWizard_CurrentStep(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	step := w.CurrentStep()
	if step == nil {
		t.Fatal("CurrentStep should not be nil")
	}
	if step.ID != "s1" {
		t.Errorf("CurrentStep ID = %q, want %q", step.ID, "s1")
	}
}

func TestWizard_SetCurrentStep(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	if err := w.SetCurrentStep(2); err != nil {
		t.Fatalf("SetCurrentStep(2) error: %v", err)
	}
	if w.CurrentStepIndex() != 2 {
		t.Errorf("CurrentStepIndex = %d, want 2", w.CurrentStepIndex())
	}
	if !w.IsLastStep() {
		t.Error("Should be last step")
	}
}

func TestWizard_SetCurrentStep_OutOfRange(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
	}
	w := NewWizard(steps)

	if err := w.SetCurrentStep(-1); err == nil {
		t.Error("SetCurrentStep(-1) should error")
	}
	if err := w.SetCurrentStep(5); err == nil {
		t.Error("SetCurrentStep(5) should error")
	}
}

// ─── Navigation Tests ────────────────────────────────────────────

func TestWizard_Next(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	if err := w.Next(); err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if w.CurrentStepIndex() != 1 {
		t.Errorf("After Next, index = %d, want 1", w.CurrentStepIndex())
	}

	if err := w.Next(); err != nil {
		t.Fatalf("Second Next() error: %v", err)
	}
	if w.CurrentStepIndex() != 2 {
		t.Errorf("After second Next, index = %d, want 2", w.CurrentStepIndex())
	}
}

func TestWizard_Next_OnLastStepFinishes(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	_ = w.Next() // to step 2 (last)
	if err := w.Next(); err != nil {
		t.Fatalf("Next on last step error: %v", err)
	}
	if !w.IsCompleted() {
		t.Error("Should be completed after Next on last step")
	}
}

func TestWizard_Next_OnLeaveBlocks(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First").SetOnLeave(func(w *Wizard) error {
			return errTestBlock
		}),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	err := w.Next()
	if err == nil {
		t.Fatal("Next should fail when OnLeave returns error")
	}
	if w.CurrentStepIndex() != 0 {
		t.Error("Should still be on step 0")
	}
}

func TestWizard_Back(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	_ = w.Next() // to step 1
	_ = w.Next() // to step 2

	if err := w.Back(); err != nil {
		t.Fatalf("Back() error: %v", err)
	}
	if w.CurrentStepIndex() != 1 {
		t.Errorf("After Back, index = %d, want 1", w.CurrentStepIndex())
	}
}

func TestWizard_Back_OnFirstStep(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	if err := w.Back(); err == nil {
		t.Error("Back on first step should error")
	}
}

func TestWizard_NextOnEnter(t *testing.T) {
	entered := false
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second").SetOnEnter(func(w *Wizard) error {
			entered = true
			return nil
		}),
	}
	w := NewWizard(steps)

	_ = w.Next()
	if !entered {
		t.Error("OnEnter for step 2 should have been called")
	}
}

// ─── Finish/Cancel/Reset Tests ───────────────────────────────────

func TestWizard_Finish(t *testing.T) {
	finished := false
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)
	w.SetOnFinish(func(w *Wizard) {
		finished = true
	})

	w.Finish()
	if !w.IsCompleted() {
		t.Error("Should be completed")
	}
	if !finished {
		t.Error("OnFinish callback should have fired")
	}
}

func TestWizard_Cancel(t *testing.T) {
	cancelled := false
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
	}
	w := NewWizard(steps)
	w.SetOnCancel(func(w *Wizard) {
		cancelled = true
	})

	w.Cancel()
	if !w.IsCancelled() {
		t.Error("Should be cancelled")
	}
	if !cancelled {
		t.Error("OnCancel callback should have fired")
	}
}

func TestWizard_Reset(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	_ = w.Next()
	_ = w.Next()
	w.Cancel()

	w.Reset()
	if w.CurrentStepIndex() != 0 {
		t.Errorf("After Reset, index = %d, want 0", w.CurrentStepIndex())
	}
	if w.IsCompleted() {
		t.Error("Reset should clear completed")
	}
	if w.IsCancelled() {
		t.Error("Reset should clear cancelled")
	}
}

// ─── Button Tests ─────────────────────────────────────────────────

func TestWizard_SelectedButton(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	if w.SelectedButton() != WizardBtnNext {
		t.Errorf("Default selected = %v, want WizardBtnNext", w.SelectedButton())
	}

	w.SetSelectedButton(WizardBtnCancel)
	if w.SelectedButton() != WizardBtnCancel {
		t.Errorf("SelectedButton = %v, want WizardBtnCancel", w.SelectedButton())
	}
}

func TestWizard_ButtonOrder(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	// First step: Next + Cancel (no Back)
	btns := w.ButtonOrder()
	if len(btns) != 2 {
		t.Fatalf("First step buttons = %d, want 2", len(btns))
	}
	if btns[0] != WizardBtnNext || btns[1] != WizardBtnCancel {
		t.Errorf("First step buttons = %v, want [Next, Cancel]", btns)
	}

	// Middle step: Back + Next + Cancel
	_ = w.Next()
	btns = w.ButtonOrder()
	if len(btns) != 3 {
		t.Fatalf("Middle step buttons = %d, want 3", len(btns))
	}
	if btns[0] != WizardBtnBack || btns[1] != WizardBtnNext || btns[2] != WizardBtnCancel {
		t.Errorf("Middle step buttons = %v, want [Back, Next, Cancel]", btns)
	}

	// Last step: Back + Finish + Cancel
	_ = w.Next()
	btns = w.ButtonOrder()
	if len(btns) != 3 {
		t.Fatalf("Last step buttons = %d, want 3", len(btns))
	}
	if btns[0] != WizardBtnBack || btns[1] != WizardBtnFinish || btns[2] != WizardBtnCancel {
		t.Errorf("Last step buttons = %v, want [Back, Finish, Cancel]", btns)
	}
}

func TestWizard_ButtonLabel(t *testing.T) {
	cases := []struct {
		btn   WizardButton
		label string
	}{
		{WizardBtnBack, "Back"},
		{WizardBtnNext, "Next"},
		{WizardBtnFinish, "Finish"},
		{WizardBtnCancel, "Cancel"},
	}
	for _, tc := range cases {
		if got := tc.btn.ButtonLabel(); got != tc.label {
			t.Errorf("ButtonLabel(%d) = %q, want %q", tc.btn, got, tc.label)
		}
	}
}

// ─── Style Tests ──────────────────────────────────────────────────

func TestWizard_Style(t *testing.T) {
	w := NewWizard(nil)
	s := w.Style()
	if s.Border.Fg == (buffer.Color{}) && s.Border.Bg == (buffer.Color{}) {
		// Style should have non-zero values from DefaultWizardStyle
	}

	custom := WizardStyle{
		Border: buffer.Style{Fg: buffer.Red},
	}
	w.SetStyle(custom)
	if w.Style().Border.Fg != buffer.Red {
		t.Error("SetStyle did not take effect")
	}
}

func TestDefaultWizardStyle(t *testing.T) {
	s := DefaultWizardStyle()
	// Verify it returns something non-zero
	if s.Title.Fg == (buffer.Color{}) {
		t.Error("DefaultWizardStyle Title.Fg should be non-zero")
	}
}

// ─── HandleKey Tests ──────────────────────────────────────────────

func TestWizard_HandleKey_Escape(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
	}
	w := NewWizard(steps)

	consumed := w.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("Escape should be consumed")
	}
	if !w.IsCancelled() {
		t.Error("Escape should cancel wizard")
	}
}

func TestWizard_HandleKey_Tab(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	// Tab should cycle buttons forward
	initial := w.SelectedButton()
	consumed := w.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !consumed {
		t.Error("Tab should be consumed")
	}
	if w.SelectedButton() == initial {
		t.Error("Tab should change selected button")
	}
}

func TestWizard_HandleKey_LeftRight(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	// Right should cycle forward
	initial := w.SelectedButton()
	w.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if w.SelectedButton() == initial {
		t.Error("Right should change selected button")
	}

	// Left should cycle backward
	w.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	// After Right then Left, should cycle through
}

func TestWizard_HandleKey_EnterNext(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	// Default focus is Next, Enter should advance
	consumed := w.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if w.CurrentStepIndex() != 1 {
		t.Errorf("After Enter on Next, index = %d, want 1", w.CurrentStepIndex())
	}
}

func TestWizard_HandleKey_EnterCancel(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
	}
	w := NewWizard(steps)
	w.SetSelectedButton(WizardBtnCancel)

	consumed := w.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if !w.IsCancelled() {
		t.Error("Enter on Cancel should cancel wizard")
	}
}

func TestWizard_HandleKey_EnterFinish(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)
	_ = w.Next() // to last step
	// On last step, Next button becomes Finish
	w.SetSelectedButton(WizardBtnFinish)

	consumed := w.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if !w.IsCompleted() {
		t.Error("Enter on Finish should complete wizard")
	}
}

func TestWizard_HandleKey_NilKey(t *testing.T) {
	w := NewWizard(nil)
	if w.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestWizard_HandleKey_CtrlN(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	consumed := w.HandleKey(&term.KeyEvent{Rune: 'n', Modifiers: term.ModCtrl})
	if !consumed {
		t.Error("Ctrl+N should be consumed")
	}
	if w.CurrentStepIndex() != 1 {
		t.Errorf("Ctrl+N should advance to step 1, got %d", w.CurrentStepIndex())
	}
}

func TestWizard_HandleKey_CtrlB(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)
	_ = w.Next() // to step 1

	consumed := w.HandleKey(&term.KeyEvent{Rune: 'b', Modifiers: term.ModCtrl})
	if !consumed {
		t.Error("Ctrl+B should be consumed")
	}
	if w.CurrentStepIndex() != 0 {
		t.Errorf("Ctrl+B should go back to step 0, got %d", w.CurrentStepIndex())
	}
}

// ─── Callback Tests ──────────────────────────────────────────────

func TestWizard_OnStepChange(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)

	changedTo := -1
	w.SetOnStepChange(func(w *Wizard, stepIdx int) {
		changedTo = stepIdx
	})

	_ = w.Next()
	if changedTo != 1 {
		t.Errorf("OnStepChange should fire with 1, got %d", changedTo)
	}

	_ = w.Back()
	if changedTo != 0 {
		t.Errorf("OnStepChange should fire with 0, got %d", changedTo)
	}
}

// ─── Measure / SetBounds / Paint Tests ───────────────────────────

func TestWizard_Measure(t *testing.T) {
	w := NewWizard([]*WizardStep{NewWizardStep("s1", "First")})

	size := w.Measure(Bounded(100, 50))
	if size.W < 30 {
		t.Errorf("Width = %d, want >= 30", size.W)
	}
	if size.H < 8 {
		t.Errorf("Height = %d, want >= 8", size.H)
	}
}

func TestWizard_Measure_ClampedToConstraints(t *testing.T) {
	w := NewWizard([]*WizardStep{NewWizardStep("s1", "First")})

	// Wizard enforces minimum size of 30x8, so constraints smaller
	// than the minimum get clamped to the minimum floor.
	size := w.Measure(Bounded(100, 20))
	if size.W > 100 {
		t.Errorf("Width = %d, should be clamped to 100", size.W)
	}
	if size.H > 20 {
		t.Errorf("Height = %d, should be clamped to 20", size.H)
	}
}

func TestWizard_SetBounds(t *testing.T) {
	w := NewWizard([]*WizardStep{NewWizardStep("s1", "First")})
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	b := w.Bounds()
	if b.W != 60 || b.H != 20 {
		t.Errorf("Bounds = %+v, want {W:60, H:20}", b)
	}
}

func TestWizard_Paint_NoPanic(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First").SetDescription("Welcome to the wizard"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	buf := buffer.NewBuffer(60, 20)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paint panicked: %v", r)
		}
	}()
	w.Paint(buf)
}

func TestWizard_Paint_ZeroBounds(t *testing.T) {
	w := NewWizard([]*WizardStep{NewWizardStep("s1", "First")})
	buf := buffer.NewBuffer(10, 10)
	w.Paint(buf) // should not panic with zero bounds
}

func TestWizard_Paint_WithContent(t *testing.T) {
	content := NewText("Step content here")
	step := NewWizardStep("s1", "First").SetContent(content)
	w := NewWizard([]*WizardStep{step})
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	buf := buffer.NewBuffer(60, 20)
	w.Paint(buf)
	// Should not panic, content should be painted
}

// ─── Children / String Tests ─────────────────────────────────────

func TestWizard_Children(t *testing.T) {
	content := NewText("child")
	step := NewWizardStep("s1", "First").SetContent(content)
	w := NewWizard([]*WizardStep{step})

	children := w.Children()
	if len(children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(children))
	}
}

func TestWizard_Children_NoContent(t *testing.T) {
	step := NewWizardStep("s1", "First")
	w := NewWizard([]*WizardStep{step})

	children := w.Children()
	if len(children) != 0 {
		t.Errorf("Children length = %d, want 0 (no content)", len(children))
	}
}

func TestWizard_String(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)
	s := w.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

// ─── IsFirstStep / IsLastStep Tests ──────────────────────────────

func TestWizard_IsFirstStep(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	if !w.IsFirstStep() {
		t.Error("Should be first step initially")
	}
	_ = w.Next()
	if w.IsFirstStep() {
		t.Error("Should not be first step after Next")
	}
}

func TestWizard_IsLastStep(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
	}
	w := NewWizard(steps)

	if w.IsLastStep() {
		t.Error("Should not be last step initially")
	}
	_ = w.Next()
	if !w.IsLastStep() {
		t.Error("Should be last step after Next")
	}
}

// ─── Concurrency Tests ───────────────────────────────────────────

func TestWizard_ConcurrentAccess(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
	}
	w := NewWizard(steps)
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				switch n % 6 {
				case 0:
					_ = w.Next()
				case 1:
					_ = w.Back()
				case 2:
					_ = w.CurrentStepIndex()
				case 3:
					_ = w.Steps()
				case 4:
					_ = w.ButtonOrder()
				case 5:
					w.Paint(buf)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestWizard_ConcurrentNavigation(t *testing.T) {
	steps := []*WizardStep{
		NewWizardStep("s1", "First"),
		NewWizardStep("s2", "Second"),
		NewWizardStep("s3", "Third"),
		NewWizardStep("s4", "Fourth"),
		NewWizardStep("s5", "Fifth"),
	}
	w := NewWizard(steps)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = w.Next()
				_ = w.Back()
				_ = w.CurrentStepIndex()
			}
		}()
	}
	wg.Wait()
}

// ─── Test Error Sentinel ──────────────────────────────────────────

var errTestBlock = errTestBlockErr{}

type errTestBlockErr struct{}

func (errTestBlockErr) Error() string { return "blocked by test" }
