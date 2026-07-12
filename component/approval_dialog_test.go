package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Construction ===

func TestNewApprovalDialog(t *testing.T) {
	d := NewApprovalDialog("Approve?", "Allow access to files?")
	if d == nil {
		t.Fatal("NewApprovalDialog returned nil")
	}
	if d.Title() != "Approve?" {
		t.Errorf("Title = %q, want 'Approve?'", d.Title())
	}
	if d.Body() != "Allow access to files?" {
		t.Errorf("Body = %q", d.Body())
	}
	if d.DialogType() != ApprovalDialogApproval {
		t.Errorf("ApprovalDialogType = %v, want ApprovalDialogApproval", d.DialogType())
	}
}

func TestNewConfirmDialogP2(t *testing.T) {
	d := NewApprovalConfirmDialog("Confirm", "Are you sure?")
	if d.DialogType() != ApprovalDialogApproval {
		t.Errorf("ApprovalDialogType = %v, want ApprovalDialogApproval", d.DialogType())
	}
	if len(d.actions) != 2 {
		t.Errorf("actions = %d, want 2 (OK/Cancel)", len(d.actions))
	}
}

func TestNewQuestionnaireDialog(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Pick one", Kind: QSingle, Options: []string{"A", "B", "C"}},
		{ID: "q2", Text: "Type something", Kind: QText},
	}
	d := NewQuestionnaireDialog("Survey", questions)
	if d.DialogType() != ApprovalDialogQuestionnaire {
		t.Errorf("ApprovalDialogType = %v", d.DialogType())
	}
	if d.CurrentQuestionIndex() != 0 {
		t.Errorf("currentQ = %d, want 0", d.CurrentQuestionIndex())
	}
	if d.IsCompleted() {
		t.Error("should not be completed")
	}

	// Verify SingleIndex initialized to -1
	if d.questions[0].SingleIndex != -1 {
		t.Errorf("SingleIndex = %d, want -1", d.questions[0].SingleIndex)
	}
}

// === Setters ===

func TestApprovalDialog_SetTitle(t *testing.T) {
	d := NewApprovalDialog("old", "body")
	d.SetTitle("new")
	if d.Title() != "new" {
		t.Errorf("Title = %q, want 'new'", d.Title())
	}
}

func TestApprovalDialog_SetBody(t *testing.T) {
	d := NewApprovalDialog("title", "old")
	d.SetBody("new body")
	if d.Body() != "new body" {
		t.Errorf("Body = %q", d.Body())
	}
}

func TestApprovalDialog_SetActions(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetActions([]DialogAction{ActionApprove, ActionDeny, ActionCancel})
	if len(d.actions) != 3 {
		t.Errorf("actions = %d, want 3", len(d.actions))
	}
}

func TestApprovalDialog_SetWidthHeight(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetWidth(80)
	d.SetHeight(30)

	s := d.Measure(Constraints{MaxWidth: 200, MaxHeight: 100})
	if s.W != 80 {
		t.Errorf("W = %d, want 80", s.W)
	}
	if s.H != 30 {
		t.Errorf("H = %d, want 30", s.H)
	}
}

func TestApprovalDialog_SetWidthMin(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetWidth(5) // should clamp to 20
	if d.width != 20 {
		t.Errorf("width = %d, want 20 (min)", d.width)
	}
}

func TestApprovalDialog_SetHeightMin(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetHeight(1) // should clamp to 5
	if d.height != 5 {
		t.Errorf("height = %d, want 5 (min)", d.height)
	}
}

func TestApprovalDialog_MeasureClamp(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetWidth(100)
	d.SetHeight(50)

	s := d.Measure(Constraints{MaxWidth: 50, MaxHeight: 20})
	if s.W != 50 {
		t.Errorf("W = %d, want 50 (clamped)", s.W)
	}
	if s.H != 20 {
		t.Errorf("H = %d, want 20 (clamped)", s.H)
	}
}

func TestApprovalDialog_SetDialogType(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetDialogType(ApprovalDialogQuestionnaire)
	if d.DialogType() != ApprovalDialogQuestionnaire {
		t.Error("ApprovalDialogType not set")
	}
}

// === HandleKey: Approval ===

func TestApprovalDialog_HandleKeyLeftRight(t *testing.T) {
	d := NewApprovalDialog("t", "b")

	// Start at index 0
	if d.ActionIndex() != 0 {
		t.Fatalf("ActionIndex = %d, want 0", d.ActionIndex())
	}

	d.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if d.ActionIndex() != 1 {
		t.Errorf("ActionIndex = %d after Right, want 1", d.ActionIndex())
	}

	// Right at end should wrap
	d.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if d.ActionIndex() != 0 {
		t.Errorf("ActionIndex = %d after Right wrap, want 0", d.ActionIndex())
	}

	// Left at start should wrap
	d.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if d.ActionIndex() != 1 {
		t.Errorf("ActionIndex = %d after Left wrap, want 1", d.ActionIndex())
	}
}

func TestApprovalDialog_HandleKeyTab(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if d.ActionIndex() != 1 {
		t.Errorf("ActionIndex = %d after Tab, want 1", d.ActionIndex())
	}
}

func TestApprovalDialog_HandleKeyEnter(t *testing.T) {
	d := NewApprovalDialog("t", "b")

	var resultID string
	var called bool
	d.SetOnResult(func(id string, answers map[string]string) {
		resultID = id
		called = true
	})

	// Index 0 = approve
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !called {
		t.Error("OnResult not called")
	}
	if resultID != "approve" {
		t.Errorf("resultID = %q, want 'approve'", resultID)
	}
}

func TestApprovalDialog_HandleKeyEscape(t *testing.T) {
	d := NewApprovalDialog("t", "b")

	var resultID string
	var closed bool
	d.SetOnResult(func(id string, _ map[string]string) {
		resultID = id
	})
	d.SetOnClose(func() {
		closed = true
	})

	d.HandleKey(&term.KeyEvent{Key: term.KeyEscape})

	if resultID != "cancel" {
		t.Errorf("resultID = %q, want 'cancel'", resultID)
	}
	if !closed {
		t.Error("OnClose not called")
	}
}

func TestApprovalDialog_HandleKeyCtrlY(t *testing.T) {
	d := NewApprovalDialog("t", "b")

	var resultID string
	d.SetOnResult(func(id string, _ map[string]string) {
		resultID = id
	})

	d.HandleKey(&term.KeyEvent{Rune: 'y', Modifiers: term.ModCtrl, Key: 0})

	if resultID != "approve" {
		t.Errorf("resultID = %q, want 'approve'", resultID)
	}
}

func TestApprovalDialog_HandleKeyCtrlN(t *testing.T) {
	d := NewApprovalDialog("t", "b")

	var resultID string
	d.SetOnResult(func(id string, _ map[string]string) {
		resultID = id
	})

	d.HandleKey(&term.KeyEvent{Rune: 'n', Modifiers: term.ModCtrl, Key: 0})

	if resultID != "deny" {
		t.Errorf("resultID = %q, want 'deny'", resultID)
	}
}

func TestApprovalDialog_HandleKeyNil(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	if d.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

// === HandleKey: Questionnaire ===

func TestQuestionnaire_QSingleNavigation(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Pick", Kind: QSingle, Options: []string{"A", "B", "C"}},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Down should move selection
	d.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if d.questions[0].SingleIndex != 0 {
		t.Errorf("SingleIndex = %d after Down, want 0", d.questions[0].SingleIndex)
	}

	d.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if d.questions[0].SingleIndex != 1 {
		t.Errorf("SingleIndex = %d after 2nd Down, want 1", d.questions[0].SingleIndex)
	}

	// Wrap around
	d.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	d.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // past end, wraps to 0
	if d.questions[0].SingleIndex != 0 {
		t.Errorf("SingleIndex = %d after wrap, want 0", d.questions[0].SingleIndex)
	}

	// Up should wrap backward
	d.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if d.questions[0].SingleIndex != 2 {
		t.Errorf("SingleIndex = %d after Up wrap, want 2", d.questions[0].SingleIndex)
	}
}

func TestQuestionnaire_QSingleSubmit(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Pick", Kind: QSingle, Options: []string{"A", "B"}},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	var resultID string
	var answers map[string]string
	var called bool
	d.SetOnResult(func(id string, ans map[string]string) {
		resultID = id
		answers = ans
		called = true
	})

	// Select option B
	d.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	d.HandleKey(&term.KeyEvent{Key: term.KeyDown})

	// Submit (actionIdx=0 = next/submit)
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !called {
		t.Error("OnResult not called")
	}
	if resultID != "submit" {
		t.Errorf("resultID = %q, want 'submit'", resultID)
	}
	if answers["q1"] != "B" {
		t.Errorf("answers['q1'] = %q, want 'B'", answers["q1"])
	}
}

func TestQuestionnaire_QMultiToggle(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Pick many", Kind: QMulti, Options: []string{"X", "Y", "Z"}},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Enter should toggle first option
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !d.questions[0].Selected[0] {
		t.Error("option 0 should be selected after Enter")
	}

	// Toggle again to deselect
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if d.questions[0].Selected[0] {
		t.Error("option 0 should be deselected after 2nd Enter")
	}
}

func TestQuestionnaire_QTextInput(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Type", Kind: QText},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	d.HandleKey(&term.KeyEvent{Rune: 'h'})
	d.HandleKey(&term.KeyEvent{Rune: 'i'})

	if d.questions[0].TextAnswer != "hi" {
		t.Errorf("TextAnswer = %q, want 'hi'", d.questions[0].TextAnswer)
	}

	// Backspace
	d.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if d.questions[0].TextAnswer != "h" {
		t.Errorf("TextAnswer = %q after backspace, want 'h'", d.questions[0].TextAnswer)
	}
}

func TestQuestionnaire_MultipleSteps(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "First", Kind: QText},
		{ID: "q2", Text: "Second", Kind: QText},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Answer first question
	d.HandleKey(&term.KeyEvent{Rune: 'a'})

	// Next
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if d.CurrentQuestionIndex() != 1 {
		t.Errorf("currentQ = %d, want 1", d.CurrentQuestionIndex())
	}

	// Answer second question
	d.HandleKey(&term.KeyEvent{Rune: 'b'})

	// Submit
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !d.IsCompleted() {
		t.Error("should be completed")
	}

	answers := d.CollectAnswers()
	if answers["q1"] != "a" {
		t.Errorf("answers['q1'] = %q, want 'a'", answers["q1"])
	}
	if answers["q2"] != "b" {
		t.Errorf("answers['q2'] = %q, want 'b'", answers["q2"])
	}
}

func TestQuestionnaire_Back(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "First", Kind: QText},
		{ID: "q2", Text: "Second", Kind: QText},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Answer and advance to q2
	d.HandleKey(&term.KeyEvent{Rune: 'a'})
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if d.CurrentQuestionIndex() != 1 {
		t.Fatalf("should be on q2, got %d", d.CurrentQuestionIndex())
	}

	// Go back: Tab to switch from Submit to Back button, then Enter
	d.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if d.CurrentQuestionIndex() != 0 {
		t.Errorf("currentQ = %d after back, want 0", d.CurrentQuestionIndex())
	}
}

func TestQuestionnaire_Escape(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "First", Kind: QText},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	var resultID string
	d.SetOnResult(func(id string, _ map[string]string) {
		resultID = id
	})

	d.HandleKey(&term.KeyEvent{Key: term.KeyEscape})

	if resultID != "cancel" {
		t.Errorf("resultID = %q, want 'cancel'", resultID)
	}
}

func TestQuestionnaire_RequiredNotAnswered(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Required", Kind: QText, Required: true},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Try to submit without answering
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if d.IsCompleted() {
		t.Error("should not be completed when required question is unanswered")
	}
}

// === Question helpers ===

func TestQuestion_IsAnswered(t *testing.T) {
	q1 := Question{Kind: QSingle, Options: []string{"A"}, SingleIndex: -1}
	if q1.IsAnswered() {
		t.Error("QSingle with -1 should not be answered")
	}
	q1.SingleIndex = 0
	if !q1.IsAnswered() {
		t.Error("QSingle with 0 should be answered")
	}

	q2 := Question{Kind: QMulti, Options: []string{"A"}, Selected: []bool{false}}
	if q2.IsAnswered() {
		t.Error("QMulti with none selected should not be answered")
	}
	q2.Selected[0] = true
	if !q2.IsAnswered() {
		t.Error("QMulti with one selected should be answered")
	}

	q3 := Question{Kind: QText, TextAnswer: ""}
	if q3.IsAnswered() {
		t.Error("QText empty should not be answered")
	}
	q3.TextAnswer = "hello"
	if !q3.IsAnswered() {
		t.Error("QText non-empty should be answered")
	}
}

func TestQuestion_Answer(t *testing.T) {
	q1 := Question{Kind: QSingle, Options: []string{"Yes", "No"}, SingleIndex: 1}
	if q1.Answer() != "No" {
		t.Errorf("Answer = %q, want 'No'", q1.Answer())
	}

	q2 := Question{Kind: QMulti, Options: []string{"A", "B", "C"}, Selected: []bool{true, false, true}}
	if q2.Answer() != "A, C" {
		t.Errorf("Answer = %q, want 'A, C'", q2.Answer())
	}

	q3 := Question{Kind: QText, TextAnswer: "hello"}
	if q3.Answer() != "hello" {
		t.Errorf("Answer = %q, want 'hello'", q3.Answer())
	}
}

// === Paint ===

func TestApprovalDialog_Paint(t *testing.T) {
	d := NewApprovalDialog("Test Title", "This is the body text.")
	d.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})

	buf := buffer.NewBuffer(60, 20)
	buf.Fill(buffer.BlankCell)
	d.Paint(buf)

	// Should have border chars
	c := buf.GetCell(0, 0)
	if c.Rune != '╭' {
		t.Errorf("corner(0,0) = %q, want '╭'", c.Rune)
	}

	// Should have title
	found := false
	for x := 0; x < 50; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == 'T' {
			found = true
			break
		}
	}
	if !found {
		t.Error("title text not found in render")
	}
}

func TestApprovalDialog_PaintNilBuffer(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.Paint(nil) // should not panic
}

func TestApprovalDialog_PaintZeroBounds(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	d.SetBounds(Rect{0, 0, 0, 0})
	buf := buffer.NewBuffer(50, 15)
	d.Paint(buf) // should not crash
}

func TestApprovalDialog_PaintQuestionnaire(t *testing.T) {
	questions := []Question{
		{ID: "q1", Text: "Pick option", Kind: QSingle, Options: []string{"A", "B"}},
	}
	d := NewQuestionnaireDialog("Survey", questions)
	d.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 20})

	buf := buffer.NewBuffer(60, 25)
	buf.Fill(buffer.BlankCell)
	d.Paint(buf)

	// Should have progress text
	found := false
	for y := 2; y < 5; y++ {
		for x := 1; x < 50; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == 'Q' { // "Question 1/1"
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("progress text not found")
	}
}

func TestApprovalDialog_Children(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	if d.Children() != nil {
		t.Error("Children should return nil")
	}
}

// === CollectAnswers ===

func TestApprovalDialog_CollectAnswers(t *testing.T) {
	questions := []Question{
		{ID: "name", Text: "Name", Kind: QText, TextAnswer: "Alice"},
		{ID: "color", Text: "Color", Kind: QSingle, Options: []string{"Red", "Blue"}},
	}
	d := NewQuestionnaireDialog("Survey", questions)

	// Set answer after construction (constructor resets SingleIndex to -1)
	d.mu.Lock()
	d.questions[1].SingleIndex = 1
	d.mu.Unlock()

	answers := d.CollectAnswers()
	if answers["name"] != "Alice" {
		t.Errorf("answers['name'] = %q, want 'Alice'", answers["name"])
	}
	if answers["color"] != "Blue" {
		t.Errorf("answers['color'] = %q, want 'Blue'", answers["color"])
	}
}

// === Concurrent ===

func TestApprovalDialog_Concurrent(t *testing.T) {
	d := NewApprovalDialog("t", "b")
	done := make(chan struct{})

	go func() {
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(60, 20)
			buf.Fill(buffer.BlankCell)
			d.SetBounds(Rect{0, 0, 60, 20})
			d.Paint(buf)
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 50; i++ {
			d.HandleKey(&term.KeyEvent{Key: term.KeyRight})
			d.Title()
			d.Body()
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

// === Helper functions ===

func TestWrapText(t *testing.T) {
	lines := adWrapText("hello world this is a test", 10)
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
	for _, line := range lines {
		if len(line) > 10 && !containsSpace(line) {
			// Long words can exceed maxW
		}
	}
}

func TestWrapTextEmpty(t *testing.T) {
	lines := adWrapText("", 10)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("adWrapText('') = %v, want ['']", lines)
	}
}

func TestWrapTextNewlines(t *testing.T) {
	lines := adWrapText("line1\nline2", 100)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestTruncateStr(t *testing.T) {
	if adTruncateStr("hello", 10) != "hello" {
		t.Error("should not truncate short string")
	}
	result := adTruncateStr("hello world", 5)
	if len([]rune(result)) > 5 {
		t.Errorf("truncated rune count = %d, should be <= 5", len([]rune(result)))
	}
}

func TestTruncateStrZero(t *testing.T) {
	if adTruncateStr("hello", 0) != "" {
		t.Error("adTruncateStr with 0 should return empty")
	}
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' {
			return true
		}
	}
	return false
}
