package glamour

import (
	"strings"
	"testing"
)

func TestNewTermRenderer(t *testing.T) {
	r, err := NewTermRenderer(WithWordWrap(80))
	if err != nil {
		t.Fatalf("NewTermRenderer error: %v", err)
	}
	if r == nil {
		t.Fatal("renderer should not be nil")
	}
}

func TestNewTermRendererWithStyles(t *testing.T) {
	r, err := NewTermRenderer(WithStyles(DefaultStyleConfig()), WithWordWrap(40))
	if err != nil {
		t.Fatalf("NewTermRenderer error: %v", err)
	}
	if r.width != 40 {
		t.Errorf("expected width 40, got %d", r.width)
	}
}

func TestRender(t *testing.T) {
	r, err := NewTermRenderer(WithWordWrap(80))
	if err != nil {
		t.Fatalf("NewTermRenderer error: %v", err)
	}
	output, err := r.Render("# Hello\n\nWorld")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(output) == 0 {
		t.Error("output should not be empty")
	}
}

func TestRenderEmpty(t *testing.T) {
	r, _ := NewTermRenderer()
	output, err := r.Render("")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	_ = strings.TrimSpace(output)
}

func TestDefaultStyleConfig(t *testing.T) {
	cfg := DefaultStyleConfig()
	_ = cfg // just verify it doesn't panic
}