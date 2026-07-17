package component

import (
	"testing"
)

func TestWindowManager_Focused_OutOfRange_P273(t *testing.T) {
	wm := NewWindowManager(&BaseComponent{})
	wm.focused = -1 // set to invalid
	if wm.Focused() != nil {
		t.Error("negative focused index should return nil")
	}
	wm.focused = 999 // set beyond range
	if wm.Focused() != nil {
		t.Error("focused beyond panes should return nil")
	}
}
