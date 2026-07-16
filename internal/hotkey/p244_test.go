package hotkey

import "testing"

// P244: Scope.String default, Enable/Disable not-found, PendingKeys empty

func TestScopeString_Default_P244(t *testing.T) {
	s := Scope(999)
	if str := s.String(); str == "" {
		t.Error("unknown scope should produce non-empty string")
	}
}

func TestEnable_NotFound_P244(t *testing.T) {
	m := NewManager()
	err := m.Enable("nonexistent")
	if err == nil {
		t.Error("Enable on nonexistent should error")
	}
}

func TestDisable_NotFound_P244(t *testing.T) {
	m := NewManager()
	err := m.Disable("nonexistent")
	if err == nil {
		t.Error("Disable on nonexistent should error")
	}
}

func TestPendingKeys_Empty_P244(t *testing.T) {
	m := NewManager()
	if keys := m.PendingKeys(); keys != nil {
		t.Error("PendingKeys with no pending should return nil")
	}
}
