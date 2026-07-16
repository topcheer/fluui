package app

import "testing"

type stubProvider struct{ items []CompletionItem }

func (s *stubProvider) Candidates(prefix string) []CompletionItem { return s.items }

func TestHandleTab_NilCompletion_P253(t *testing.T) {
	i := NewInputLine(">")
	if i.handleTab() {
		t.Error("nil completion should return false")
	}
}

func TestHandleTab_EmptyPrefix_P253(t *testing.T) {
	i := NewInputLine(">")
	i.SetText(" ")
	i.SetCompletionManager(NewCompletionManager(&stubProvider{}))
	if i.handleTab() {
		t.Error("empty prefix should return false")
	}
}

func TestHandleTab_CycleNextFail_P253(t *testing.T) {
	i := NewInputLine(">")
	i.SetText("ab")
	// provider returns no candidates → Start fails
	i.SetCompletionManager(NewCompletionManager(&stubProvider{items: nil}))
	if i.handleTab() {
		t.Error("no candidates should return false")
	}
}

func TestHandleShiftTab_NilCompletion_P253(t *testing.T) {
	i := NewInputLine(">")
	if i.handleShiftTab() {
		t.Error("nil completion should return false")
	}
}

func TestHandleShiftTab_Inactive_P253(t *testing.T) {
	i := NewInputLine(">")
	i.SetCompletionManager(NewCompletionManager(&stubProvider{}))
	if i.handleShiftTab() {
		t.Error("inactive completion should return false")
	}
}
