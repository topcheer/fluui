package event

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// mockFocusable implements FocusableComponent for testing.
type mockFocusable struct {
	consumed bool
	gotEvent bool
}

func (m *mockFocusable) HandleEvent(e Event) bool {
	m.gotEvent = true
	return m.consumed
}

func TestFocus_SetFocused(t *testing.T) {
	d := NewDispatcher()
	fc := &mockFocusable{consumed: false}
	d.SetFocused(fc)
	if d.Focused() != fc {
		t.Fatal("Focused() should return the set component")
	}
	d.SetFocused(nil)
	if d.Focused() != nil {
		t.Fatal("Focused() should be nil after clear")
	}
}

func TestFocus_KeyEvent_GoesToFocusedFirst(t *testing.T) {
	d := NewDispatcher()
	fc := &mockFocusable{consumed: true}
	d.SetFocused(fc)

	d.BindKey(PlainRune('q'), func(e Event) bool {
		t.Fatal("binding should not be reached when focused component consumes")
		return false
	})

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'q'}}
	result := d.Dispatch(ev)
	if !result {
		t.Fatal("should return true when focused consumes")
	}
	if !fc.gotEvent {
		t.Fatal("focused component should have received event")
	}
}

func TestFocus_KeyEvent_BubblesToBindings(t *testing.T) {
	d := NewDispatcher()
	fc := &mockFocusable{consumed: false} // does not consume
	d.SetFocused(fc)

	bindingCalled := false
	d.BindKey(PlainRune('q'), func(e Event) bool {
		bindingCalled = true
		return true
	})

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'q'}}
	result := d.Dispatch(ev)
	if !result {
		t.Fatal("should return true when binding consumes")
	}
	if !bindingCalled {
		t.Fatal("binding should have been called after focus bubbling")
	}
	if !fc.gotEvent {
		t.Fatal("focused component should still have received event first")
	}
}

func TestFocus_MouseEvent_GoesToFocusedFirst(t *testing.T) {
	d := NewDispatcher()
	fc := &mockFocusable{consumed: true}
	d.SetFocused(fc)

	mouseCalled := false
	d.OnMouse(func(e Event) bool {
		mouseCalled = true
		return true
	})

	ev := Event{Type: TypeMouse, Mouse: &term.MouseEvent{}}
	d.Dispatch(ev)
	if mouseCalled {
		t.Fatal("global mouse handler should not be called when focused consumes")
	}
}

func TestFocus_PasteEvent_GoesToFocusedFirst(t *testing.T) {
	d := NewDispatcher()
	fc := &mockFocusable{consumed: true}
	d.SetFocused(fc)

	pasteCalled := false
	d.OnPaste(func(e Event) bool {
		pasteCalled = true
		return true
	})

	ev := Event{Type: TypePaste, Paste: "hello"}
	d.Dispatch(ev)
	if pasteCalled {
		t.Fatal("global paste handler should not be called when focused consumes")
	}
}

func TestFocus_PushPopChain(t *testing.T) {
	d := NewDispatcher()

	fc1 := &mockFocusable{consumed: false}
	fc2 := &mockFocusable{consumed: false}

	d.PushFocus(fc1)
	d.PushFocus(fc2)

	chain := d.FocusChain()
	if len(chain) != 2 {
		t.Fatalf("expected chain length 2, got %d", len(chain))
	}

	popped := d.PopFocus()
	if popped != fc2 {
		t.Fatal("PopFocus should return the last pushed component")
	}

	chain = d.FocusChain()
	if len(chain) != 1 {
		t.Fatalf("expected chain length 1 after pop, got %d", len(chain))
	}
}

func TestFocus_PopEmptyChain(t *testing.T) {
	d := NewDispatcher()
	result := d.PopFocus()
	if result != nil {
		t.Fatal("PopFocus on empty chain should return nil")
	}
}

func TestFocus_ClearFocus(t *testing.T) {
	d := NewDispatcher()
	d.SetFocused(&mockFocusable{})
	d.PushFocus(&mockFocusable{})

	d.ClearFocus()
	if d.Focused() != nil {
		t.Fatal("Focused should be nil after ClearFocus")
	}
	if len(d.FocusChain()) != 0 {
		t.Fatal("chain should be empty after ClearFocus")
	}
}

func TestFocus_ChainPriority(t *testing.T) {
	d := NewDispatcher()

	// fc1 is at bottom of chain, fc2 at top
	fc1 := &mockFocusable{consumed: false}
	fc2 := &mockFocusable{consumed: true}
	d.PushFocus(fc1)
	d.PushFocus(fc2)

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}}
	d.Dispatch(ev)

	// fc2 is at end of chain, gets event first, consumes it
	if !fc2.gotEvent {
		t.Fatal("top of chain should receive event")
	}
	// fc1 should also have received it (chain iterates all, but fc2 consumes)
	// Actually: chain is iterated front to back. fc1 is [0], fc2 is [1].
	// fc1 (doesn't consume) → fc2 (consumes). So both get events.
	if !fc1.gotEvent {
		t.Fatal("fc1 should also have received event before fc2 consumed")
	}
}

func TestFocus_FocusedAfterChain(t *testing.T) {
	d := NewDispatcher()

	// Set both a chain and a focused component
	chainFC := &mockFocusable{consumed: false}
	d.PushFocus(chainFC)

	focusedFC := &mockFocusable{consumed: true}
	d.SetFocused(focusedFC)

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}}
	d.Dispatch(ev)

	// Chain is checked first, then focused
	if !chainFC.gotEvent {
		t.Fatal("chain component should receive event first")
	}
	if !focusedFC.gotEvent {
		t.Fatal("focused component should receive event after chain")
	}
}

func TestFocus_NoFocus_FallsThroughToDefault(t *testing.T) {
	d := NewDispatcher()
	defaultCalled := false
	d.OnKey(func(e Event) bool {
		defaultCalled = true
		return false
	})

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'x'}}
	d.Dispatch(ev)
	if !defaultCalled {
		t.Fatal("default key handler should be called when no focus and no bindings")
	}
}
