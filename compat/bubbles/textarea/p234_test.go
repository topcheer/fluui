package textarea

import "testing"

// P234: cover Width(), Height(), SetStyles() — all at 0%

func TestTextArea_WidthHeight_P234(t *testing.T) {
	m := New()
	m.SetWidth(40)
	m.SetHeight(10)
	if m.Width() != 40 {
		t.Errorf("Width = %d, want 40", m.Width())
	}
	if m.Height() != 10 {
		t.Errorf("Height = %d, want 10", m.Height())
	}
}

func TestTextArea_SetStyles_P234(t *testing.T) {
	m := New()
	m.SetStyles(DefaultStyles(true)) // no-op in fluui, but should not panic
}
