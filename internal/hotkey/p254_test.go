package hotkey

import "testing"

func TestImportConfig_ParseSequenceError_P254(t *testing.T) {
	m := NewManager()
	cfg := Config{
		Bindings: []ConfigBinding{
			{Action: "test", Sequence: "!!!invalid", Scope: "global"},
		},
	}
	err := m.ImportConfig(cfg)
	if err == nil {
		t.Error("invalid sequence should return error")
	}
}

func TestImportConfig_ParseScopeError_P254(t *testing.T) {
	m := NewManager()
	cfg := Config{
		Bindings: []ConfigBinding{
			{Action: "test", Sequence: "ctrl+a", Scope: "!!!invalid"},
		},
	}
	err := m.ImportConfig(cfg)
	if err == nil {
		t.Error("invalid scope should return error")
	}
}

func TestImportConfig_Success_P254(t *testing.T) {
	m := NewManager()
	cfg := Config{
		Bindings: []ConfigBinding{
			{Action: "save", Sequence: "ctrl+s", Scope: "global", Description: "Save file", Enabled: true},
		},
	}
	if err := m.ImportConfig(cfg); err != nil {
		t.Errorf("valid config should not error: %v", err)
	}
}
