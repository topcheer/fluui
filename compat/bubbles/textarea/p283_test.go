package textarea

import "testing"

func TestModel_SetStyles_P283(t *testing.T) {
	m := New()
	m.SetStyles(DefaultStyles(false)) // no-op but shouldn't panic
}
