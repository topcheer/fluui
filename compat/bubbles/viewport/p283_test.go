package viewport

import "testing"

func TestVpInner_SetYOffset_P283(t *testing.T) {
	v := vpInner{}
	v.SetYOffset(10) // no-op on inner stub
}

func TestModel_NewWithOptions_P283(t *testing.T) {
	m := New(WithWidth(80), WithHeight(24))
	if m.Width() != 80 {
		t.Errorf("expected width 80, got %d", m.Width())
	}
	if m.Height() != 24 {
		t.Errorf("expected height 24, got %d", m.Height())
	}
}
