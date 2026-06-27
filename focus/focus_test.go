package focus

import (
	"testing"
)

// mockFocusable is a minimal Focusable implementation for testing.
type mockFocusable struct {
	id      string
	focused bool
}

func (m *mockFocusable) Focus()       { m.focused = true }
func (m *mockFocusable) Blur()        { m.focused = false }
func (m *mockFocusable) Focused() bool { return m.focused }

func TestFocusAdd(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	if m.Len() != 3 {
		t.Fatalf("expected len 3, got %d", m.Len())
	}
	// No item should be focused after Add alone.
	if m.Current() != nil {
		t.Fatalf("expected nil current after Add, got %v", m.Current())
	}
}

func TestFocusNext(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	// Tab through the full ring: a → b → c → a → b → c
	m.Next()
	if !a.Focused() {
		t.Fatal("expected a focused after first Next")
	}
	if m.Current() != a {
		t.Fatal("expected current = a")
	}

	m.Next()
	if !b.Focused() {
		t.Fatal("expected b focused")
	}
	if a.Focused() {
		t.Fatal("a should be blurred")
	}

	m.Next()
	if !c.Focused() {
		t.Fatal("expected c focused")
	}

	m.Next()
	if !a.Focused() {
		t.Fatal("expected a focused (wrap around)")
	}

	m.Next()
	if !b.Focused() {
		t.Fatal("expected b focused")
	}

	m.Next()
	if !c.Focused() {
		t.Fatal("expected c focused")
	}
}

func TestFocusPrev(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	// Reverse traversal: a → c → b → a
	m.Prev()
	if !a.Focused() {
		t.Fatal("expected a focused after first Prev")
	}

	m.Prev()
	if !c.Focused() {
		t.Fatal("expected c focused after Prev (wrap)")
	}
	if a.Focused() {
		t.Fatal("a should be blurred")
	}

	m.Prev()
	if !b.Focused() {
		t.Fatal("expected b focused")
	}

	m.Prev()
	if !a.Focused() {
		t.Fatal("expected a focused")
	}
}

func TestFocusSet(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	m.Set(b)
	if !b.Focused() {
		t.Fatal("expected b focused after Set")
	}
	if m.Current() != b {
		t.Fatal("expected current = b")
	}

	// Set c — b should be blurred.
	m.Set(c)
	if !c.Focused() {
		t.Fatal("expected c focused")
	}
	if b.Focused() {
		t.Fatal("b should be blurred")
	}

	// Set a non-member — focus should be unchanged.
	external := &mockFocusable{id: "external"}
	m.Set(external)
	if !c.Focused() {
		t.Fatal("focus should remain on c when setting non-member")
	}
	if external.Focused() {
		t.Fatal("external should not be focused")
	}
}

func TestFocusClear(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}

	m.Add(a)
	m.Add(b)

	m.Next() // a focused
	m.Clear()

	if m.Current() != nil {
		t.Fatal("expected nil current after Clear")
	}
	if a.Focused() {
		t.Fatal("a should be blurred after Clear")
	}
	if b.Focused() {
		t.Fatal("b should be blurred after Clear")
	}

	// Next after Clear should still work.
	m.Next()
	if !a.Focused() {
		t.Fatal("expected a focused after Next following Clear")
	}
}

func TestFocusRemove(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	// Focus b, then remove it — focus should transfer.
	m.Set(b)
	m.Remove(b)

	if m.Len() != 2 {
		t.Fatalf("expected len 2, got %d", m.Len())
	}
	if b.Focused() {
		t.Fatal("b should be blurred after removal")
	}
	// b was at index 1; after removal c shifts to index 1, so c should be focused.
	if !c.Focused() {
		t.Fatal("expected c focused after removing b")
	}
	if m.Current() != c {
		t.Fatal("expected current = c")
	}

	// Remove a non-focused item before current — current index should adjust.
	// Current is c at index 1. Remove a (index 0).
	m.Remove(a)
	if m.Len() != 1 {
		t.Fatalf("expected len 1, got %d", m.Len())
	}
	if !c.Focused() {
		t.Fatal("c should still be focused")
	}
	if m.Current() != c {
		t.Fatal("expected current = c")
	}

	// Remove the last item.
	m.Remove(c)
	if m.Len() != 0 {
		t.Fatalf("expected len 0, got %d", m.Len())
	}
	if m.Current() != nil {
		t.Fatal("expected nil current after removing all")
	}
}

func TestFocusEmpty(t *testing.T) {
	m := NewFocusManager()

	// None of these should panic.
	m.Next()
	m.Prev()
	m.Clear()

	if m.Current() != nil {
		t.Fatal("expected nil current on empty manager")
	}
	if m.Len() != 0 {
		t.Fatalf("expected len 0, got %d", m.Len())
	}

	// Remove on empty should be a no-op.
	m.Remove(&mockFocusable{id: "nonexistent"})
	if m.Len() != 0 {
		t.Fatal("len should still be 0")
	}
}

func TestFocusSingle(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}

	m.Add(a)

	// Next on a single-element ring should focus a.
	m.Next()
	if !a.Focused() {
		t.Fatal("expected a focused")
	}

	// Next again should still be a (cyclic to self).
	m.Next()
	if !a.Focused() {
		t.Fatal("a should still be focused")
	}

	// Prev should also be a.
	m.Prev()
	if !a.Focused() {
		t.Fatal("a should still be focused after Prev")
	}
}

func TestFocusWrap(t *testing.T) {
	m := NewFocusManager()
	a := &mockFocusable{id: "a"}
	b := &mockFocusable{id: "b"}
	c := &mockFocusable{id: "c"}

	m.Add(a)
	m.Add(b)
	m.Add(c)

	// Move focus to the last element.
	m.Set(c)

	// Next should wrap to a.
	m.Next()
	if !a.Focused() {
		t.Fatal("expected a focused after wrap from c")
	}
	if c.Focused() {
		t.Fatal("c should be blurred")
	}
	if m.Current() != a {
		t.Fatal("expected current = a")
	}

	// Prev from a should wrap to c.
	m.Prev()
	if !c.Focused() {
		t.Fatal("expected c focused after Prev wrap from a")
	}
	if m.Current() != c {
		t.Fatal("expected current = c")
	}
}
