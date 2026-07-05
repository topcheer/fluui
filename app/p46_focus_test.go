package app

import (
	"testing"
)

func TestP46_ChatApp_OnFocus(t *testing.T) {
	app := NewChatApp(80, 24)
	var receivedFocus *bool
	app.OnFocus(func(focused bool) {
		receivedFocus = &focused
	})

	if receivedFocus != nil {
		t.Error("handler should not be called yet")
	}
}

func TestP46_ChatApp_OnFocus_Setter(t *testing.T) {
	app := NewChatApp(80, 24)
	called := false
	app.OnFocus(func(focused bool) {
		called = true
	})
	// Just verify the handler was set without panic
	if called {
		t.Error("handler should not be called yet")
	}
}
