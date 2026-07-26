package term

import (
	"strings"
	"testing"
)

func TestP422_VSCodeShellIntegration(t *testing.T) {
	if VSCodePromptStart != "\x1b]633;A\x07" {
		t.Errorf("PromptStart = %q", VSCodePromptStart)
	}
	if VSCodeCommandStart != "\x1b]633;C\x07" {
		t.Errorf("CommandStart = %q", VSCodeCommandStart)
	}
	if VSCodeCommandEnd != "\x1b]633;D\x07" {
		t.Errorf("CommandEnd = %q", VSCodeCommandEnd)
	}
	if VSCodePromptEnd != "\x1b]633;B\x07" {
		t.Errorf("PromptEnd = %q", VSCodePromptEnd)
	}
}

func TestP422_VSCodeCwd(t *testing.T) {
	got := VSCodeCwd("/home/user")
	if got[:8] != "\x1b]633;P=" {
		t.Errorf("Cwd prefix = %q", got[:8])
	}
	if got[len(got)-1:] != "\x07" {
		t.Error("Cwd should end with BEL")
	}
}

func TestP422_KittyNotification(t *testing.T) {
	got := KittyNotification("Title", "Body")
	// Should contain OSC 99 sequences
	if got[:4] != "\x1b]99" {
		t.Errorf("Notification prefix = %q", got[:4])
	}
}

func TestP422_OSC8LinkWithID(t *testing.T) {
	got := OSC8LinkWithID("link1", "click me", HyperlinkOptions{URL: "https://example.com"})
	if !strings.Contains(got, "\x1b]") {
		t.Error("Should contain OSC sequence")
	}
}
