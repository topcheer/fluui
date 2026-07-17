package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P271: form HandleKey + diffviewer + dialog + filepicker

func TestTextField_HandleKey_LeftRight_P271(t *testing.T) {
	f := NewTextField("name", "name", "hello")
	f.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	f.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	f.HandleKey(&term.KeyEvent{Key: term.KeyRight}) // at end, no-op
}

func TestCheckboxField_HandleKey_NilKey_P271(t *testing.T) {
	f := NewCheckboxField("check", "check", false)
	if f.HandleKey(nil) {
		t.Error("nil key should return false")
	}
}

func TestCheckboxField_HandleKey_Space_P271(t *testing.T) {
	f := NewCheckboxField("check", "check", false)
	f.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if f.Value() != "true" {
		t.Error("space should toggle checkbox to true")
	}
}

func TestSelectField_SetSelectedIndex_Empty_P271(t *testing.T) {
	f := NewSelectField("pick", "pick", []string{})
	f.SetSelectedIndex(5)
}

func TestSelectField_SetSelectedIndex_Clamp_P271(t *testing.T) {
	f := NewSelectField("pick", "pick", []string{"a", "b", "c"})
	f.SetSelectedIndex(-5)
	if f.Value() != "a" {
		t.Errorf("expected 'a' at index 0, got %q", f.Value())
	}
	f.SetSelectedIndex(100)
	if f.Value() != "c" {
		t.Errorf("expected 'c' at last index, got %q", f.Value())
	}
}

func TestSelectField_HandleKey_Empty_P271(t *testing.T) {
	f := NewSelectField("pick", "pick", nil)
	if f.HandleKey(&term.KeyEvent{Key: term.KeyUp}) {
		t.Error("empty options should return false")
	}
}

func TestSelectField_HandleKey_Nil_P271(t *testing.T) {
	f := NewSelectField("pick", "pick", []string{"a"})
	if f.HandleKey(nil) {
		t.Error("nil key should return false")
	}
}

func TestSelectField_HandleKey_Wrap_P271(t *testing.T) {
	f := NewSelectField("pick", "pick", []string{"a", "b", "c"})
	f.SetSelectedIndex(1)
	f.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if f.Value() != "a" {
		t.Errorf("expected 'a' after up, got %q", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if f.Value() != "c" {
		t.Errorf("expected 'c' after wrap, got %q", f.Value())
	}
	f.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if f.Value() != "a" {
		t.Errorf("expected 'a' after down, got %q", f.Value())
	}
}

func TestDiffViewer_MaxScrollOffset_ZeroHeight_P271(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("@@ -1,2 +1,2 @@\n-a\n+b\n")
	dv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 0})
}

func TestDiffViewer_VisibleHeight_WithTitle_P271(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetTitle("diff")
	dv.SetContent("@@ -1,1 +1,1 @@\n-a\n+b\n")
	dv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
}

func TestDialog_SetButtons_CursorReset_P271(t *testing.T) {
	d := NewDialog(DialogCustom, "Test", "Body")
	d.SetButtons([]DialogButton{{Label: "OK"}, {Label: "Cancel"}, {Label: "Apply"}, {Label: "Close"}})
	d.cursor = 3
	d.SetButtons([]DialogButton{{Label: "Yes"}, {Label: "No"}})
	if d.cursor != 0 {
		t.Error("cursor should reset when buttons shrink")
	}
}

func TestFilePicker_EmptyDir_P271(t *testing.T) {
	fp := NewFilePicker("")
	if fp == nil {
		t.Error("should not return nil")
	}
}

func TestDiffViewer_Measure_WithLineNumbers_P271(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetShowLineNumbers(true)
	dv.SetContent("@@ -1,3 +1,3 @@\n ctx\n-old\n+new\n")
	s := dv.Measure(Constraints{MaxWidth: 200, MaxHeight: 200})
	if s.W < 20 {
		t.Errorf("line numbers should add width, got %d", s.W)
	}
}
