package termcompat

import (
	"encoding/base64"
	"strings"
	"testing"
)

// === Detection Tests ===

func TestDetectImageProtocol_Iterm2(t *testing.T) {
	caps := DetectImageProtocolFromEnv("iTerm.app", "xterm-256color", "truecolor")
	if caps.Protocol != ImageIterm2 {
		t.Errorf("Protocol: got %s, want iTerm2", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay: expected true")
	}
	if !caps.CanAnimate {
		t.Error("CanAnimate: expected true for iTerm2")
	}
}

func TestDetectImageProtocol_Iterm2_ApplesTerminal(t *testing.T) {
	// Apple Terminal does NOT support inline images.
	// TERM_PROGRAM=Apple_Terminal should not match iTerm
	caps := DetectImageProtocolFromEnv("Apple_Terminal", "xterm-256color", "")
	if caps.Protocol != ImageNone {
		t.Errorf("Protocol: got %s, want None (Apple Terminal)", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_Kitty(t *testing.T) {
	caps := DetectImageProtocolFromEnv("kitty", "xterm-kitty", "truecolor")
	if caps.Protocol != ImageKitty {
		t.Errorf("Protocol: got %s, want Kitty", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay: expected true")
	}
	if !caps.CanAnimate {
		t.Error("CanAnimate: expected true for Kitty")
	}
}

func TestDetectImageProtocol_KittyViaTerm(t *testing.T) {
	// TERM=xterm-kitty without TERM_PROGRAM
	caps := DetectImageProtocolFromEnv("", "xterm-kitty", "")
	if caps.Protocol != ImageKitty {
		t.Errorf("Protocol: got %s, want Kitty", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_WezTerm(t *testing.T) {
	// WezTerm supports iTerm2 inline image protocol
	caps := DetectImageProtocolFromEnv("WezTerm", "wezterm", "truecolor")
	if caps.Protocol != ImageIterm2 {
		t.Errorf("Protocol: got %s, want iTerm2 (WezTerm)", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay: expected true for WezTerm")
	}
}

func TestDetectImageProtocol_WezTermViaTerm(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "wezterm-256color", "")
	if caps.Protocol != ImageIterm2 {
		t.Errorf("Protocol: got %s, want iTerm2 (WezTerm)", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_Sixel_mlterm(t *testing.T) {
	caps := DetectImageProtocolFromEnv("mlterm", "mlterm", "")
	if caps.Protocol != ImageSixel {
		t.Errorf("Protocol: got %s, want Sixel", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay: expected true")
	}
}

func TestDetectImageProtocol_Sixel_foot(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "foot", "")
	if caps.Protocol != ImageSixel {
		t.Errorf("Protocol: got %s, want Sixel", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_Sixel_mintty(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "mintty", "")
	if caps.Protocol != ImageSixel {
		t.Errorf("Protocol: got %s, want Sixel", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_SixelViaMarker(t *testing.T) {
	// Some terminals report sixel in TERM
	caps := DetectImageProtocolFromEnv("", "xterm-sixel", "")
	if caps.Protocol != ImageSixel {
		t.Errorf("Protocol: got %s, want Sixel", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_None(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "xterm-256color", "")
	if caps.Protocol != ImageNone {
		t.Errorf("Protocol: got %s, want None", ImageProtocolName(caps.Protocol))
	}
	if caps.CanDisplay {
		t.Error("CanDisplay: expected false")
	}
}

func TestDetectImageProtocol_NoneGnomeTerminal(t *testing.T) {
	// GNOME Terminal (stock) doesn't support any image protocol
	caps := DetectImageProtocolFromEnv("GNOME Terminal", "xterm-256color", "")
	if caps.Protocol != ImageNone {
		t.Errorf("Protocol: got %s, want None", ImageProtocolName(caps.Protocol))
	}
}

// === Priority Tests ===

func TestDetectImageProtocol_PriorityKittyOverIterm2(t *testing.T) {
	// If both Kitty and iTerm2 could match, Kitty should win.
	// This can't normally happen (TERM can't be both), but we test
	// the priority by checking that kitty detection comes first.
	caps := DetectImageProtocolFromEnv("kitty", "xterm-kitty", "")
	if caps.Protocol != ImageKitty {
		t.Errorf("Priority: expected Kitty, got %s", ImageProtocolName(caps.Protocol))
	}
}

func TestDetectImageProtocol_PriorityIterm2OverSixel(t *testing.T) {
	// WezTerm supports both iTerm2 and Sixel, but iTerm2 should be preferred.
	// WezTerm doesn't appear in the sixel list with that TERM_PROGRAM value,
	// but the logic should still prefer iTerm2.
	caps := DetectImageProtocolFromEnv("WezTerm", "wezterm", "")
	if caps.Protocol != ImageIterm2 {
		t.Errorf("Priority: expected iTerm2 for WezTerm, got %s", ImageProtocolName(caps.Protocol))
	}
}

// === ImageProtocolName Tests ===

func TestImageProtocolName(t *testing.T) {
	tests := []struct {
		proto   ImageProtocol
		want    string
	}{
		{ImageNone, "None"},
		{ImageSixel, "Sixel"},
		{ImageIterm2, "iTerm2"},
		{ImageKitty, "Kitty"},
	}
	for _, tt := range tests {
		got := ImageProtocolName(tt.proto)
		if got != tt.want {
			t.Errorf("ImageProtocolName(%d): got %q, want %q", tt.proto, got, tt.want)
		}
	}
}

func TestImageProtocolName_Unknown(t *testing.T) {
	got := ImageProtocolName(ImageProtocol(99))
	if got != "None" {
		t.Errorf("ImageProtocolName(99): got %q, want None", got)
	}
}

// === FormatIterm2Image Tests ===

func TestFormatIterm2Image_Basic(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("fake png data"))
	seq := FormatIterm2Image(data, "test.png", 100, 200)

	// Must start with OSC 1337
	if !strings.HasPrefix(seq, "\x1b]1337;File=") {
		t.Error("expected OSC 1337 prefix")
	}
	// Must end with ST
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Error("expected ST suffix")
	}
	// Must contain inline=1
	if !strings.Contains(seq, "inline=1") {
		t.Error("expected inline=1")
	}
	// Must contain the base64 data
	if !strings.Contains(seq, data) {
		t.Error("expected base64 data in sequence")
	}
}

func TestFormatIterm2Image_WithName(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("x"))
	name := "screenshot.png"
	encodedName := base64.StdEncoding.EncodeToString([]byte(name))
	seq := FormatIterm2Image(data, name, 0, 0)

	if !strings.Contains(seq, "name="+encodedName) {
		t.Error("expected encoded filename in sequence")
	}
}

func TestFormatIterm2Image_WithDimensions(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("x"))
	seq := FormatIterm2Image(data, "img.png", 320, 240)

	if !strings.Contains(seq, "width=320") {
		t.Error("expected width=320")
	}
	if !strings.Contains(seq, "height=240") {
		t.Error("expected height=240")
	}
}

func TestFormatIterm2Image_ZeroDimensions(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("x"))
	seq := FormatIterm2Image(data, "img.png", 0, 0)

	if strings.Contains(seq, "width=") {
		t.Error("should not contain width= when width=0")
	}
	if strings.Contains(seq, "height=") {
		t.Error("should not contain height= when height=0")
	}
}

// === FormatKittyImage Tests ===

func TestFormatKittyImage_Basic(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("fake png data"))
	seq := FormatKittyImage(data, 100, 200)

	// Must start with APC: ESC _ G
	if !strings.HasPrefix(seq, "\x1b_G") {
		t.Error("expected ESC _ G prefix")
	}
	// Must end with ST
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Error("expected ST suffix")
	}
	// Must contain a=T (transmit action)
	if !strings.Contains(seq, "a=T") {
		t.Error("expected a=T action")
	}
	// Must contain f=100 (PNG format)
	if !strings.Contains(seq, "f=100") {
		t.Error("expected f=100 format")
	}
	// Must contain t=d (direct transmission)
	if !strings.Contains(seq, "t=d") {
		t.Error("expected t=d transmission")
	}
}

func TestFormatKittyImage_Dimensions(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("x"))
	seq := FormatKittyImage(data, 640, 480)

	if !strings.Contains(seq, "s=640") {
		t.Error("expected s=640")
	}
	if !strings.Contains(seq, "v=480") {
		t.Error("expected v=480")
	}
}

func TestFormatKittyImage_ZeroDimensions(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("x"))
	seq := FormatKittyImage(data, 0, 0)

	if strings.Contains(seq, "s=0") {
		t.Error("should not contain s=0 when width=0")
	}
	if strings.Contains(seq, "v=0") {
		t.Error("should not contain v=0 when height=0")
	}
}

func TestFormatKittyImage_SingleChunk(t *testing.T) {
	// Small data — should be single chunk with m=0
	data := base64.StdEncoding.EncodeToString([]byte("small"))
	seq := FormatKittyImage(data, 0, 0)

	if !strings.Contains(seq, "m=0") {
		t.Error("expected m=0 for single chunk")
	}
}

func TestFormatKittyImage_MultiChunk(t *testing.T) {
	// Large data — should be split into multiple chunks
	largeData := base64.StdEncoding.EncodeToString(make([]byte, 20000)) // ~27KB base64
	seq := FormatKittyImage(largeData, 0, 0)

	// Should contain m=1 for first chunk and m=0 for last chunk
	count := strings.Count(seq, "\x1b_G")
	if count < 2 {
		t.Errorf("expected at least 2 chunks for large data, got %d", count)
	}
	if !strings.Contains(seq, "m=1") {
		t.Error("expected m=1 for multi-chunk data")
	}
	if !strings.Contains(seq, "m=0") {
		t.Error("expected m=0 for last chunk")
	}
}

// === FormatSixelPlaceholder Tests ===

func TestFormatSixelPlaceholder_Basic(t *testing.T) {
	// Raw sixel data (simplified)
	data := []byte("8;1;q#0;2;0;0;0#0!10~-")
	seq := FormatSixelPlaceholder(data)

	// Must start with DCS: ESC P
	if !strings.HasPrefix(seq, "\x1bP") {
		t.Error("expected ESC P (DCS) prefix")
	}
	// Must end with ST: ESC \
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Error("expected ST suffix")
	}
	// Must contain the raw data
	if !strings.Contains(seq, string(data)) {
		t.Error("expected raw sixel data in sequence")
	}
}

func TestFormatSixelPlaceholder_Empty(t *testing.T) {
	seq := FormatSixelPlaceholder(nil)
	if seq != "\x1bP\x1b\\" {
		t.Errorf("empty sixel: got %q, want %q", seq, "\x1bP\x1b\\")
	}
}

// === ImageCapabilities Struct Tests ===

func TestImageCapabilities_KittyFullyPopulated(t *testing.T) {
	caps := DetectImageProtocolFromEnv("kitty", "xterm-kitty", "truecolor")
	if caps.Protocol != ImageKitty {
		t.Fatalf("Protocol: got %s", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay should be true")
	}
	if !caps.CanAnimate {
		t.Error("CanAnimate should be true for Kitty")
	}
}

func TestImageCapabilities_SixelNoAnimation(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "foot", "")
	if caps.Protocol != ImageSixel {
		t.Fatalf("Protocol: got %s", ImageProtocolName(caps.Protocol))
	}
	if !caps.CanDisplay {
		t.Error("CanDisplay should be true")
	}
	if caps.CanAnimate {
		t.Error("CanAnimate should be false for Sixel")
	}
}

func TestImageCapabilities_NoneFullyPopulated(t *testing.T) {
	caps := DetectImageProtocolFromEnv("", "dumb", "")
	if caps.Protocol != ImageNone {
		t.Fatalf("Protocol: got %s", ImageProtocolName(caps.Protocol))
	}
	if caps.CanDisplay {
		t.Error("CanDisplay should be false")
	}
	if caps.CanAnimate {
		t.Error("CanAnimate should be false")
	}
}
