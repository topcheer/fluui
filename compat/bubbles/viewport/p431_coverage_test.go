package viewport

import "testing"

func TestVpInner_SetYOffset_P431(t *testing.T) {
	v := &vpInner{}
	v.SetYOffset(10) // no-op, should not panic
}
