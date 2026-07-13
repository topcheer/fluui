package app

import "testing"

func TestP166_NewView(t *testing.T) {
	v := NewView("hello")
	if v.String() != "hello" {
		t.Errorf("expected 'hello', got %q", v.String())
	}
}

func TestP166_NewView_Empty(t *testing.T) {
	v := NewView("")
	if !v.IsEmpty() {
		t.Error("expected empty")
	}
	if v.Len() != 0 {
		t.Error("expected len 0")
	}
}

func TestP166_NewView_MultiLine(t *testing.T) {
	v := NewView("line1\nline2\nline3")
	if v.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", v.LineCount())
	}
	lines := v.Lines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestP166_NewView_Stringer(t *testing.T) {
	v := NewView("content")
	s := v.String()
	if s != "content" {
		t.Errorf("expected 'content', got %q", s)
	}
}