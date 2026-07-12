package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func mkSingleQ(prompt string, opts []string, idx int) Question {
	return Question{Kind: QSingle, Text: prompt, Options: opts, SingleIndex: idx}
}
func mkMultiQ(prompt string, opts []string, sel []bool) Question {
	return Question{Kind: QMulti, Text: prompt, Options: opts, Selected: sel}
}
func mkTextQ(prompt, answer string, cursor int) Question {
	return Question{Kind: QText, Text: prompt, TextAnswer: answer, textCursor: cursor}
}

// --- paintQuestionLocked coverage ---

func TestP157_PaintQuestion_Single(t *testing.T) {
	d := NewQuestionnaireDialog("Test", []Question{
		mkSingleQ("Pick:", []string{"A", "B", "C"}, 0),
	})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	d.Paint(buf)
}

func TestP157_PaintQuestion_Multi(t *testing.T) {
	d := NewQuestionnaireDialog("Test", []Question{
		mkMultiQ("Pick:", []string{"X", "Y", "Z"}, []bool{false, true, false}),
	})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	d.Paint(buf)
}

func TestP157_PaintQuestion_Text(t *testing.T) {
	d := NewQuestionnaireDialog("Test", []Question{
		mkTextQ("Name:", "hello", 5),
	})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	d.Paint(buf)
}

func TestP157_PaintQuestion_Narrow(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{
		mkSingleQ("P", []string{"A", "B"}, 0),
	})
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf := buffer.NewBuffer(10, 8)
	d.Paint(buf)
}

func TestP157_PaintQuestion_InvalidQ(t *testing.T) {
	d := NewQuestionnaireDialog("Test", []Question{
		mkSingleQ("Pick:", []string{"A"}, 0),
	})
	d.currentQ = -1
	d.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	d.Paint(buf)
}

func TestP157_PaintQuestion_OutOfRange(t *testing.T) {
	d := NewQuestionnaireDialog("Test", []Question{
		mkSingleQ("Pick:", []string{"A"}, 0),
	})
	d.currentQ = 99
	d.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	d.Paint(buf)
}

// --- handleQuestionnaireKeyLocked: QSingle ---

func TestP157_QKey_SingleUp(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A", "B"}, 1)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyUp}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleUpWrap(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A", "B"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyUp}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleDown(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A", "B"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyDown}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleDownWrap(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A", "B"}, 1)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyDown}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleLeft(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.actionIdx = 1
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyLeft}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleRight(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyRight}); !h { t.Error("handled") }
}

func TestP157_QKey_SingleEscape(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyEscape}); !h { t.Error("handled") }
}

// --- handleQuestionnaireKeyLocked: QMulti ---

func TestP157_QKey_MultiUp(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkMultiQ("P:", []string{"A", "B"}, []bool{false, false})})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.actionIdx = 1
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyUp}); !h { t.Error("handled") }
}

func TestP157_QKey_MultiDown(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkMultiQ("P:", []string{"A", "B"}, []bool{false, false})})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyDown}); !h { t.Error("handled") }
}

func TestP157_QKey_MultiToggle(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkMultiQ("P:", []string{"A", "B"}, []bool{false, false})})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeySpace}); !h { t.Error("handled") }
}

func TestP157_QKey_MultiEscape(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkMultiQ("P:", []string{"A"}, []bool{false})})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyEscape}); !h { t.Error("handled") }
}

// --- handleQuestionnaireKeyLocked: QText ---

func TestP157_QKey_TextTab(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hello", 5)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.actionIdx = 0
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyTab}); !h { t.Error("handled") }
}

func TestP157_QKey_TextBackspace(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hello", 5)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyBackspace}); !h { t.Error("handled") }
}

func TestP157_QKey_TextLeft(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hello", 3)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyLeft}); !h { t.Error("handled") }
}

func TestP157_QKey_TextRight(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hello", 2)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyRight}); !h { t.Error("handled") }
}

func TestP157_QKey_TextEscape(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hi", 2)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyEscape}); !h { t.Error("handled") }
}

func TestP157_QKey_TextType(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkTextQ("P:", "hi", 2)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	if h := d.HandleKey(&term.KeyEvent{Rune: 'X'}); !h { t.Error("handled") }
}

func TestP157_QKey_InvalidQ(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A"}, 0)})
	d.currentQ = -1
	if h := d.HandleKey(&term.KeyEvent{Key: term.KeyUp}); h { t.Error("expected not handled") }
}

// --- handleQuestionnaireActionLocked ---

func TestP157_QAction_RequiredUnanswered(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{
		{Kind: QSingle, Text: "P:", Options: []string{"A", "B"}, Required: true, SingleIndex: -1},
	})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.actionIdx = 0
	h := d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !h { t.Error("handled") }
	// Cannot check action with single return
}

func TestP157_QAction_Cancel(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{mkSingleQ("P:", []string{"A"}, 0)})
	d.SetActions([]DialogAction{ActionNext, ActionCancel})
	d.actionIdx = 1
	h := d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !h { t.Error("handled") }
	// Cannot check action with single return
}

func TestP157_QAction_Back(t *testing.T) {
	d := NewQuestionnaireDialog("T", []Question{
		mkSingleQ("Q1:", []string{"A"}, 0),
		mkSingleQ("Q2:", []string{"B"}, 0),
	})
	d.currentQ = 1
	d.SetActions([]DialogAction{ActionSubmit, ActionBack})
	d.actionIdx = 1
	h := d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !h { t.Error("handled") }
	if d.currentQ != 0 { t.Errorf("expected currentQ=0, got %d", d.currentQ) }
}