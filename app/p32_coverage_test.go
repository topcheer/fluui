package app

import (
	"testing"

	"github.com/topcheer/fluui/component/layout"
)

// mockProvider is a simple CompletionProvider for testing.
type mockProvider struct{}

func (m *mockProvider) Candidates(prefix string) []CompletionItem {
	return nil
}

// Coverage tests for 0% functions in app package.

func TestP32_ChatApp_ScrollView(t *testing.T) {
	app := NewChatApp(80, 24)
	sv := app.ScrollView()
	if sv == nil {
		t.Error("ScrollView should not be nil")
	}
}

func TestP32_ChatApp_Root(t *testing.T) {
	app := NewChatApp(80, 24)
	_ = app.Root() // exercising the getter
}

func TestP32_ChatApp_SetRootFlex(t *testing.T) {
	app := NewChatApp(80, 24)
	flex := layout.NewFlex(layout.FlexColumn)
	app.SetRootFlex(flex)
	if app.Root() == nil {
		t.Error("Root should not be nil after SetRootFlex")
	}
}

func TestP32_InputLine_CompletionManager(t *testing.T) {
	il := NewInputLine("> ")
	cm := NewCompletionManager(&mockProvider{})
	il.SetCompletionManager(cm)
	got := il.CompletionManager()
	if got == nil {
		t.Error("CompletionManager should not be nil after SetCompletionManager")
	}
}

func TestP32_CompletionManager_SelectedIndex(t *testing.T) {
	cm := NewCompletionManager(&mockProvider{})
	if cm.SelectedIndex() != 0 {
		t.Errorf("New SelectedIndex: got %d want 0", cm.SelectedIndex())
	}
}

func TestP32_CompletionManager_Items(t *testing.T) {
	cm := NewCompletionManager(&mockProvider{})
	if cm.Items() != nil {
		t.Error("New Items should be nil")
	}
}
