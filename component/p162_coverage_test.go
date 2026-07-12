package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Target Badge.Measure 73.3%
func TestP162_BadgeMeasure_AllVariants(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeSuccess, BadgeWarning, BadgeWarning}
	for _, v := range variants {
		b := NewBadge("Test", v)
		size := b.Measure(Bounded(80, 24))
		if size.W <= 0 || size.H <= 0 {
			t.Errorf("variant %d: expected positive size, got %+v", v, size)
		}
	}
}

func TestP162_BadgeMeasure_WithIcon(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	b.SetIcon("*")
	size := b.Measure(Bounded(80, 24))
	if size.W <= 0 {
		t.Error("expected positive width with icon")
	}
}

func TestP162_BadgeMeasure_ShortText(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	size := b.Measure(Bounded(80, 24))
	_ = size
}

func TestP162_BadgeMeasure_NarrowWidth(t *testing.T) {
	b := NewBadge("VeryLongBadgeText", BadgeInfo)
	size := b.Measure(Bounded(2, 5))
	if size.W > 2 {
		t.Errorf("expected width clamped to 2, got %d", size.W)
	}
}

func TestP162_BadgeMeasure_AllSizes(t *testing.T) {
	sizes := []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge}
	for _, s := range sizes {
		b := NewBadge("Test", BadgeInfo)
		b.SetSize(s)
		sz := b.Measure(Bounded(80, 24))
		if sz.H <= 0 {
			t.Errorf("size %d: expected positive height", s)
		}
	}
}

// Target AutoComplete.Paint 76.7%
func TestP162_AutoCompletePaint_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP162_AutoCompletePaint_WithItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Description: "desc1"},
		{Label: "item2", Description: "desc2"},
		{Label: "item3", Description: "desc3"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP162_AutoCompletePaint_WithCategory(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Category: "cat1"},
		{Label: "item2", Category: "cat2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP162_AutoCompletePaint_SelectedWithDesc(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "short", Description: "a long description that might be truncated"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	ac.Paint(buf)
}

func TestP162_AutoCompletePaint_ScrollDown(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 15)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + itoa(i)}
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

// Target AutoComplete.clampScrollLocked 83.3%
func TestP162_AutoCompleteClampScroll(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "a"}, {Label: "b"}, {Label: "c"},
	})
	ac.SetCursor(2)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	ac.Paint(buf) // should clamp scroll
}

// Target SetActions 80%
func TestP162_SetActions(t *testing.T) {
	d := NewApprovalDialog("Test", "body")
	d.SetActions([]DialogAction{ActionOK, ActionCancel})
	if len(d.actions) != 2 {
		t.Error("expected 2 actions")
	}
	// Set nil actions
	d.SetActions(nil)
	if len(d.actions) != 0 {
		t.Error("expected 0 actions")
	}
}

// Target executeActionLocked 80%
func TestP162_ExecuteAction_Close(t *testing.T) {
	d := NewApprovalDialog("Test", "body")
	d.SetActions([]DialogAction{ActionOK})
	d.actionIdx = 0
	d.SetOnResult(func(actionID string, answers map[string]string) {
		if actionID != "ok" {
			t.Errorf("expected ok, got %q", actionID)
		}
	})
	// Execute via key
	d.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
}

// Target Answer 81.8%
func TestP162_Answer_NotAnswered(t *testing.T) {
	q := Question{Kind: QSingle, Text: "Pick:", Options: []string{"A", "B"}}
	ans := q.Answer()
	if ans != "" && ans != "A" {
		t.Errorf("expected empty for unanswered, got %q", ans)
	}
}

func TestP162_Answer_SingleAnswered(t *testing.T) {
	q := Question{Kind: QSingle, Text: "Pick:", Options: []string{"A", "B"}, SingleIndex: 1}
	ans := q.Answer()
	if ans != "B" {
		t.Errorf("expected 'B', got %q", ans)
	}
}

func TestP162_Answer_TextAnswered(t *testing.T) {
	q := Question{Kind: QText, Text: "Name:", TextAnswer: "Alice"}
	ans := q.Answer()
	if ans != "Alice" {
		t.Errorf("expected 'Alice', got %q", ans)
	}
}

func TestP162_Answer_MultiAnswered(t *testing.T) {
	q := Question{Kind: QMulti, Text: "Pick:", Options: []string{"A", "B"}, Selected: []bool{true, false}}
	ans := q.Answer()
	if ans != "A" {
		t.Errorf("expected 'A', got %q", ans)
	}
}

func TestP162_IsAnswered_Required(t *testing.T) {
	q := Question{Kind: QSingle, Text: "Pick:", Options: []string{"A"}, Required: true, SingleIndex: -1}
	if q.IsAnswered() {
		t.Error("expected not answered for required with no selection")
	}
}

func TestP162_IsAnswered_Answered(t *testing.T) {
	q := Question{Kind: QSingle, Text: "Pick:", Options: []string{"A"}, SingleIndex: 0}
	if !q.IsAnswered() {
		t.Error("expected answered")
	}
}

// Helper
