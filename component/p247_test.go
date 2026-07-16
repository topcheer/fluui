package component
import (
	"testing"
	"github.com/topcheer/fluui/internal/buffer"
)
func TestQuestion_Answer_QSingle_P247(t *testing.T) {
	q := Question{Kind: QSingle, Options: []string{"yes", "no"}, SingleIndex: 0}
	if q.Answer() != "yes" {
		t.Errorf("QSingle answer = %q", q.Answer())
	}
}
func TestQuestion_Answer_QSingleOutOfRange_P247(t *testing.T) {
	q := Question{Kind: QSingle, Options: []string{"yes"}, SingleIndex: 5}
	if q.Answer() != "" {
		t.Error("out of range should return empty")
	}
}
func TestQuestion_Answer_QMulti_P247(t *testing.T) {
	q := Question{Kind: QMulti, Options: []string{"a", "b", "c"}, Selected: []bool{true, false, true}}
	if q.Answer() != "a, c" {
		t.Errorf("QMulti answer = %q, want 'a, c'", q.Answer())
	}
}
func TestQuestion_IsAnswered_QMulti_P247(t *testing.T) {
	q := Question{Kind: QMulti, Options: []string{"a"}, Selected: []bool{true}}
	if !q.IsAnswered() {
		t.Error("QMulti with selection should be answered")
	}
}
func TestQuestion_IsAnswered_QMultiNone_P247(t *testing.T) {
	q := Question{Kind: QMulti, Options: []string{"a"}, Selected: []bool{false}}
	if q.IsAnswered() {
		t.Error("QMulti with no selection should not be answered")
	}
}
func TestAutoComplete_Paint_P247(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "alpha", Value: "alpha"},
		{Label: "beta", Value: "beta"},
		{Label: "gamma", Value: "gamma"},
	})
	
	ac.SetPosition(0, 0)
	ac.SetMaxVisible(5)
	// Start to set visible=true
	ac.Show(0, 0)
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf)
}
func TestAutoComplete_ClampScroll_P247(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "a1", Value: "a1"},
		{Label: "a2", Value: "a2"},
		{Label: "a3", Value: "a3"},
	})
	ac.Show(0, 0)
	ac.SetMaxVisible(2)
	ac.mu.Lock()
	ac.cursor = 100
	ac.clampScrollLocked()
	ac.mu.Unlock()
}
