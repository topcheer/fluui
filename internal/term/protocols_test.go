package term

import (
	"strings"
	"testing"
)

// Tests for the terminal protocol helpers in protocols.go.
// Covers OSC 8 hyperlinks, Synchronized Output, Focus Tracking,
// Window Title, Alt Screen, Bracketed Paste, Mouse SGR, and Kitty Keyboard.

func TestOSC8Start_BasicURL(t *testing.T) {
	got := OSC8Start(HyperlinkOptions{URL: "https://example.com"})
	want := "\x1b]8;;https://example.com\x1b\\"
	if got != want {
		t.Errorf("OSC8Start basic: got %q want %q", got, want)
	}
}

func TestOSC8Start_WithID(t *testing.T) {
	got := OSC8Start(HyperlinkOptions{URL: "https://example.com", ID: "link1"})
	want := "\x1b]8;id=link1;https://example.com\x1b\\"
	if got != want {
		t.Errorf("OSC8Start with id: got %q want %q", got, want)
	}
}

func TestOSC8Start_WithParams(t *testing.T) {
	got := OSC8Start(HyperlinkOptions{URL: "https://example.com", Params: "icon=foo"})
	want := "\x1b]8;icon=foo;https://example.com\x1b\\"
	if got != want {
		t.Errorf("OSC8Start with params: got %q want %q", got, want)
	}
}

func TestOSC8Start_WithIDAndParams(t *testing.T) {
	got := OSC8Start(HyperlinkOptions{URL: "https://example.com", ID: "l1", Params: "icon=foo"})
	want := "\x1b]8;id=l1:icon=foo;https://example.com\x1b\\"
	if got != want {
		t.Errorf("OSC8Start id+params: got %q want %q", got, want)
	}
}

func TestOSC8Start_EmptyURL(t *testing.T) {
	got := OSC8Start(HyperlinkOptions{})
	want := "\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("OSC8Start empty: got %q want %q", got, want)
	}
}

func TestOSC8End_ClosesLink(t *testing.T) {
	got := OSC8End()
	want := "\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("OSC8End: got %q want %q", got, want)
	}
}

func TestOSC8Link_Complete(t *testing.T) {
	got := OSC8Link(HyperlinkOptions{URL: "https://example.com"}, "click here")
	want := "\x1b]8;;https://example.com\x1b\\click here\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("OSC8Link: got %q want %q", got, want)
	}
}

func TestOSC8Link_ContainsText(t *testing.T) {
	got := OSC8Link(HyperlinkOptions{URL: "https://example.com"}, "Hello")
	if !strings.Contains(got, "Hello") {
		t.Error("OSC8Link should contain visible text")
	}
	if !strings.HasPrefix(got, "\x1b]8") {
		t.Error("OSC8Link should start with OSC 8 begin")
	}
	if !strings.HasSuffix(got, "\x1b]8;;\x1b\\") {
		t.Error("OSC8Link should end with OSC 8 terminator")
	}
}

func TestOSC8Link_MultipleURLs(t *testing.T) {
	// Each link must be independently addressable
	link1 := OSC8Link(HyperlinkOptions{URL: "https://a.com"}, "A")
	link2 := OSC8Link(HyperlinkOptions{URL: "https://b.com"}, "B")
	if !strings.Contains(link1, "https://a.com") {
		t.Error("link1 missing a.com")
	}
	if !strings.Contains(link2, "https://b.com") {
		t.Error("link2 missing b.com")
	}
}

// --- Synchronized Output ---

func TestSyncBegin_Sequence(t *testing.T) {
	got := SyncBegin
	want := "\x1bP=1s\x1b\\"
	if got != want {
		t.Errorf("SyncBegin: got %q want %q", got, want)
	}
}

func TestSyncEnd_Sequence(t *testing.T) {
	got := SyncEnd
	want := "\x1bP=2s\x1b\\"
	if got != want {
		t.Errorf("SyncEnd: got %q want %q", got, want)
	}
}

func TestSync_WrapsOutput(t *testing.T) {
	got := Sync("hello")
	want := SyncBegin + "hello" + SyncEnd
	if got != want {
		t.Errorf("Sync: got %q want %q", got, want)
	}
}

func TestSync_EmptyInput(t *testing.T) {
	got := Sync("")
	if got != "" {
		t.Errorf("Sync empty should return empty, got %q", got)
	}
}

func TestSync_PreservesContent(t *testing.T) {
	input := "complex\x1b[31mred\x1b[0mtext"
	got := Sync(input)
	if !strings.Contains(got, input) {
		t.Error("Sync should preserve the inner content verbatim")
	}
	if !strings.HasPrefix(got, SyncBegin) {
		t.Error("Sync output should start with SyncBegin")
	}
	if !strings.HasSuffix(got, SyncEnd) {
		t.Error("Sync output should end with SyncEnd")
	}
}

// --- Focus Tracking ---

func TestEnableFocus_Sequence(t *testing.T) {
	if EnableFocus != "\x1b[?1004h" {
		t.Errorf("EnableFocus: got %q", EnableFocus)
	}
}

func TestDisableFocus_Sequence(t *testing.T) {
	if DisableFocus != "\x1b[?1004l" {
		t.Errorf("DisableFocus: got %q", DisableFocus)
	}
}

// --- Window Title ---

func TestSetWindowTitle_BELTerminator(t *testing.T) {
	got := SetWindowTitle("My App")
	want := "\x1b]2;My App\x07"
	if got != want {
		t.Errorf("SetWindowTitle: got %q want %q", got, want)
	}
}

func TestSetIconName(t *testing.T) {
	got := SetIconName("icon")
	want := "\x1b]1;icon\x07"
	if got != want {
		t.Errorf("SetIconName: got %q want %q", got, want)
	}
}

func TestSetWindowTitleAndIcon(t *testing.T) {
	got := SetWindowTitleAndIcon("title")
	want := "\x1b]0;title\x07"
	if got != want {
		t.Errorf("SetWindowTitleAndIcon: got %q want %q", got, want)
	}
}

func TestSetWindowTitle_StripsBELByte(t *testing.T) {
	// BEL inside title would prematurely terminate the OSC
	got := SetWindowTitle("hello\x07world")
	if strings.Contains(got, "\x07world") {
		t.Error("setTitleOSC should strip interior BEL bytes")
	}
	if !strings.HasSuffix(got, "\x07") {
		t.Error("should still end with single BEL terminator")
	}
}

func TestSetWindowTitle_Empty(t *testing.T) {
	got := SetWindowTitle("")
	want := "\x1b]2;\x07"
	if got != want {
		t.Errorf("empty title: got %q want %q", got, want)
	}
}

func TestQueryWindowTitle(t *testing.T) {
	got := QueryWindowTitle()
	if got != "\x1b]21\x1b\\" {
		t.Errorf("QueryWindowTitle: got %q", got)
	}
}

// --- Cursor Visibility ---

func TestHideCursor(t *testing.T) {
	if HideCursor != "\x1b[?25l" {
		t.Errorf("HideCursor: got %q", HideCursor)
	}
}

func TestShowCursor(t *testing.T) {
	if ShowCursor != "\x1b[?25h" {
		t.Errorf("ShowCursor: got %q", ShowCursor)
	}
}

// --- Alt Screen ---

func TestEnterAltScreen(t *testing.T) {
	if EnterAltScreen != "\x1b[?1049h" {
		t.Errorf("EnterAltScreen: got %q", EnterAltScreen)
	}
}

func TestLeaveAltScreen(t *testing.T) {
	if LeaveAltScreen != "\x1b[?1049l" {
		t.Errorf("LeaveAltScreen: got %q", LeaveAltScreen)
	}
}

// --- Bracketed Paste ---

func TestEnableBracketedPaste(t *testing.T) {
	if EnableBracketedPaste != "\x1b[?2004h" {
		t.Errorf("EnableBracketedPaste: got %q", EnableBracketedPaste)
	}
}

func TestDisableBracketedPaste(t *testing.T) {
	if DisableBracketedPaste != "\x1b[?2004l" {
		t.Errorf("DisableBracketedPaste: got %q", DisableBracketedPaste)
	}
}

// --- Mouse SGR ---

func TestEnableMouseSGR(t *testing.T) {
	if EnableMouseSGR != "\x1b[?1006h" {
		t.Errorf("EnableMouseSGR: got %q", EnableMouseSGR)
	}
}

func TestDisableMouseSGR(t *testing.T) {
	if DisableMouseSGR != "\x1b[?1006l" {
		t.Errorf("DisableMouseSGR: got %q", DisableMouseSGR)
	}
}

// --- Kitty Keyboard ---

func TestEnableKittyKeyboard(t *testing.T) {
	if EnableKittyKeyboard != "\x1b[>1u" {
		t.Errorf("EnableKittyKeyboard: got %q", EnableKittyKeyboard)
	}
}

func TestDisableKittyKeyboard(t *testing.T) {
	if DisableKittyKeyboard != "\x1b[<u" {
		t.Errorf("DisableKittyKeyboard: got %q", DisableKittyKeyboard)
	}
}

// --- Bell ---

func TestBell(t *testing.T) {
	if Bell != "\x07" {
		t.Errorf("Bell: got %q want \\x07", Bell)
	}
}

// --- Clipboard convenience ---

func TestCopyClipboard_DelegatesToOSC52(t *testing.T) {
	got := CopyClipboard("hello")
	expected := CopyOSC52("hello")
	if got != expected {
		t.Errorf("CopyClipboard: got %q want %q", got, expected)
	}
}

func TestCopyPrimary_DelegatesToOSC52(t *testing.T) {
	got := CopyPrimary("hello")
	expected := CopyOSC52Source("hello", ClipboardPrimary)
	if got != expected {
		t.Errorf("CopyPrimary: got %q want %q", got, expected)
	}
}

// --- HyperlinkOptions field defaults ---

func TestHyperlinkOptions_ZeroValue(t *testing.T) {
	var opts HyperlinkOptions
	start := OSC8Start(opts)
	// All empty -> produces bare OSC 8 with empty URL
	if start != "\x1b]8;;\x1b\\" {
		t.Errorf("zero value HyperlinkOptions: got %q", start)
	}
}

// --- Integration: writing protocols through a TestTerminal ---

func TestProtocols_WriteThroughTerminal(t *testing.T) {
	var buf strings.Builder
	tt := NewTestTerminal(strings.NewReader(""), &buf, 80, 24)
	if tt == nil {
		t.Skip("NewTestTerminal not available")
	}
	link := OSC8Link(HyperlinkOptions{URL: "https://example.com"}, "Click")
	tt.WriteRaw(link)

	title := SetWindowTitle("Test")
	tt.WriteRaw(title)

	synced := Sync("payload")
	tt.WriteRaw(synced)
	// If we got here without panic, the protocols are terminal-write safe.
}

// --- OSC8 with unicode URL and text ---

func TestOSC8Link_UnicodeText(t *testing.T) {
	got := OSC8Link(HyperlinkOptions{URL: "https://例え.jp"}, "クリック")
	if !strings.Contains(got, "クリック") {
		t.Error("OSC8Link should preserve unicode text")
	}
	if !strings.Contains(got, "https://例え.jp") {
		t.Error("OSC8Link should preserve unicode URL")
	}
}

// --- Title with unicode ---

func TestSetWindowTitle_Unicode(t *testing.T) {
	got := SetWindowTitle("日本語タイトル")
	want := "\x1b]2;日本語タイトル\x07"
	if got != want {
		t.Errorf("unicode title: got %q want %q", got, want)
	}
}

// --- Sync with multiline content ---

func TestSync_MultilineContent(t *testing.T) {
	multiline := "line1\nline2\nline3"
	got := Sync(multiline)
	// The content between begin/end should equal the input exactly
	inner := strings.TrimSuffix(strings.TrimPrefix(got, SyncBegin), SyncEnd)
	if inner != multiline {
		t.Errorf("Sync inner content changed: got %q want %q", inner, multiline)
	}
}

// --- Verify all protocol constants use ESC prefix ---

func TestProtocolConstants_StartWithESC(t *testing.T) {
	consts := []string{
		SyncBegin, SyncEnd,
		EnableFocus, DisableFocus,
		HideCursor, ShowCursor,
		EnterAltScreen, LeaveAltScreen,
		EnableBracketedPaste, DisableBracketedPaste,
		EnableMouseSGR, DisableMouseSGR,
		EnableKittyKeyboard, DisableKittyKeyboard,
	}
	for _, c := range consts {
		if !strings.HasPrefix(c, "\x1b") {
			t.Errorf("protocol constant %q should start with ESC (\\x1b)", c)
		}
	}
}

func TestBell_NotESC(t *testing.T) {
	// Bell is a single byte, not an ESC sequence — sanity check
	if Bell == "\x1b" {
		t.Error("Bell should be BEL (0x07), not ESC")
	}
}
