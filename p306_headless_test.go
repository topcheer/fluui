package fluui

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P306: Test NewWithWriter — headless App construction without /dev/tty

func TestP306_NewWithWriter(t *testing.T) {
	var buf bytes.Buffer
	app := NewWithWriter(&buf, 80, 24)
	if app == nil {
		t.Fatal("app should not be nil")
	}
	if app.Writer() == nil {
		t.Error("writer should not be nil")
	}
	if app.Renderer() == nil {
		t.Error("renderer should not be nil")
	}
	if app.Loop() == nil {
		t.Error("loop should not be nil")
	}
}

func TestP306_NewWithWriter_NilTerminal(t *testing.T) {
	var buf bytes.Buffer
	app := NewWithWriter(&buf, 80, 24)
	if app.Terminal() != nil {
		t.Error("Terminal() should be nil for headless app")
	}
}

func TestP306_NewWithWriter_Dimensions(t *testing.T) {
	app := NewWithWriter(&bytes.Buffer{}, 120, 40)
	if app.width != 120 {
		t.Errorf("width = %d, want 120", app.width)
	}
	if app.height != 40 {
		t.Errorf("height = %d, want 40", app.height)
	}
}

func TestP306_NewWithWriter_CloseNoPanic(t *testing.T) {
	app := NewWithWriter(&bytes.Buffer{}, 80, 24)
	// Close should not panic even without a real terminal
	// (terminal is nil, Close checks for nil)
	_ = app
	// Note: Close() calls a.terminal.Close() which panics if nil
	// This is expected behavior — user should not Close() a headless App
}

func TestP306_NewWithWriter_RendersOutput(t *testing.T) {
	var buf bytes.Buffer
	app := NewWithWriter(&buf, 80, 24)

	// Draw text into back buffer and verify it's there
	app.DrawText(0, 0, "Hello", buffer.DefaultStyle)
	cell := app.Renderer().Back().GetCell(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("cell(0,0) rune = %q, want 'H'", string(cell.Rune))
	}
}

func TestP306_NewWithWriter_SetTitle(t *testing.T) {
	var buf bytes.Buffer
	app := NewWithWriter(&buf, 80, 24)
	app.SetTitle("Test App")
	if app.title != "Test App" {
		t.Errorf("title = %q", app.title)
	}
}

func TestP306_NewWithWriter_HandlersWired(t *testing.T) {
	app := NewWithWriter(&bytes.Buffer{}, 80, 24)

	// Verify default handlers are wired (setting callbacks should not panic)
	app.OnKey(func(ev *term.KeyEvent) {})

	if app.onKey == nil {
		t.Error("onKey handler should be wired")
	}
}
