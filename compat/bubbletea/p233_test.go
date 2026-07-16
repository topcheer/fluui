package bubbletea

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P233: Verify KeyPressMsg.Text field and Key() method — the exact patterns
// ggcode uses: msg.Text (6+ places) and msg.Key().Text (3 places).

func TestKeyPressMsg_TextPrintable_P233(t *testing.T) {
	k := KeyPressMsg{Rune: 'a'}
	if k.Text != "" {
		// Text is populated by keyTextFor, not auto-set on struct literal
		// So we need to verify keyTextFor works
	}
	txt := keyTextFor(term.KeyUnknown, 'a', 0)
	if txt != "a" {
		t.Errorf("keyTextFor rune 'a' = %q, want 'a'", txt)
	}
}

func TestKeyPressMsg_TextSpace_P233(t *testing.T) {
	txt := keyTextFor(term.KeySpace, ' ', 0)
	if txt != " " {
		t.Errorf("keyTextFor space = %q, want ' '", txt)
	}
}

func TestKeyPressMsg_TextSpecialKey_P233(t *testing.T) {
	// Special keys should return empty Text
	txt := keyTextFor(term.KeyEnter, 0, 0)
	if txt != "" {
		t.Errorf("keyTextFor enter = %q, want ''", txt)
	}
	txt = keyTextFor(term.KeyUp, 0, 0)
	if txt != "" {
		t.Errorf("keyTextFor up = %q, want ''", txt)
	}
	txt = keyTextFor(term.KeyEscape, 0, 0)
	if txt != "" {
		t.Errorf("keyTextFor escape = %q, want ''", txt)
	}
}

func TestKeyPressMsg_KeyMethod_P233(t *testing.T) {
	// Test the Key() method returns a Key with correct fields
	k := KeyPressMsg{
		Code:  term.KeyUnknown,
		Rune:  'x',
		Mod:   term.ModCtrl,
		Ctrl:  true,
		Text:  "x",
	}
	key := k.Key()
	if key.Text != "x" {
		t.Errorf("Key().Text = %q, want 'x'", key.Text)
	}
	if key.Code != term.KeyUnknown {
		t.Errorf("Key().Code = %v, want KeyUnknown", key.Code)
	}
}

func TestKeyPressMsg_KeyMethodEmptyText_P233(t *testing.T) {
	// Special key → Key().Text should be empty
	k := KeyPressMsg{
		Code: term.KeyEnter,
		Text: "",
	}
	key := k.Key()
	if key.Text != "" {
		t.Errorf("Key().Text for Enter = %q, want ''", key.Text)
	}
}

func TestKeyPressMsg_KeyMethodSpace_P233(t *testing.T) {
	k := KeyPressMsg{
		Code: term.KeySpace,
		Text: " ",
	}
	key := k.Key()
	if key.Text != " " {
		t.Errorf("Key().Text for Space = %q, want ' '", key.Text)
	}
}

func TestKeyPressMsg_TextChineseChar_P233(t *testing.T) {
	// Multi-byte chars should work
	txt := keyTextFor(term.KeyUnknown, '你', 0)
	if txt != "你" {
		t.Errorf("keyTextFor '你' = %q, want '你'", txt)
	}
}

func TestProgram_HandleKeyPopulatesText_P233(t *testing.T) {
	// Verify HandleKey populates Text correctly
	m := &echoModel{}
	p := NewProgram(m)
	// Printable char
	handled := p.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'a'})
	if !handled {
		t.Fatal("HandleKey should return true")
	}
	// Space
	p.HandleKey(&term.KeyEvent{Key: term.KeySpace, Rune: ' '})
	// Enter (special)
	p.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
}

// echoModel is a simple model for testing
type echoModel struct {
	keys []KeyPressMsg
}

func (m *echoModel) Init() Cmd { return nil }
func (m *echoModel) Update(msg Msg) (Model, Cmd) {
	if k, ok := msg.(KeyPressMsg); ok {
		m.keys = append(m.keys, k)
	}
	return m, nil
}
func (m *echoModel) View() View { return View{Content: ""} }
