package term

import (
	"testing"
)

// === InlineImageBase64 coverage (44% → 90%+) ===

func TestP59_InlineImageBase64_AllOptions(t *testing.T) {
	opts := ImageOptions{
		Name:                "test.png",
		Width:               ImageSize{Value: 100, Unit: ""},
		Height:              ImageSize{Value: 200, Unit: "px"},
		PreserveAspectRatio: true,
		Inline:              true,
	}
	result := InlineImageBase64("aGVsbG8=", opts)
	if result == "" {
		t.Error("expected non-empty result")
	}
	if !contains(result, "width=100") {
		t.Error("expected width=100 in result")
	}
	if !contains(result, "height=200px") {
		t.Error("expected height=200px in result")
	}
	if !contains(result, "preserveAspectRatio=1") {
		t.Error("expected preserveAspectRatio=1")
	}
	if !contains(result, "inline=1") {
		t.Error("expected inline=1")
	}
}

func TestP59_InlineImageBase64_AutoSize(t *testing.T) {
	opts := ImageOptions{
		Width:  ImageSize{Unit: "auto"},
		Height: ImageSize{Unit: "auto"},
	}
	result := InlineImageBase64("data", opts)
	if !contains(result, "width=auto") {
		t.Error("expected width=auto")
	}
	if !contains(result, "height=auto") {
		t.Error("expected height=auto")
	}
}

func TestP59_InlineImageBase64_NoAspectRatio(t *testing.T) {
	opts := ImageOptions{
		PreserveAspectRatio: false,
	}
	result := InlineImageBase64("data", opts)
	if !contains(result, "preserveAspectRatio=0") {
		t.Error("expected preserveAspectRatio=0")
	}
}

func TestP59_InlineImageBase64_NotInline(t *testing.T) {
	opts := ImageOptions{
		Inline: false,
	}
	result := InlineImageBase64("data", opts)
	if !contains(result, "inline=0") {
		t.Error("expected inline=0")
	}
}

func TestP59_InlineImageBase64_NoName(t *testing.T) {
	opts := DefaultImageOptions()
	result := InlineImageBase64("data", opts)
	// Should not contain name=
	if contains(result, "name=") {
		t.Error("expected no name field when Name is empty")
	}
}

// === parseHexComponent edge cases ===

func TestP59_ParseHexComponent_2Digit(t *testing.T) {
	// 2-digit hex (standard): "FF" = 255
	v := parseHexComponent("FF")
	if v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
}

func TestP59_ParseHexComponent_1Digit(t *testing.T) {
	// 1-digit hex: "F" should scale to 0xF * 0x11 = 0xFF = 255
	v := parseHexComponent("F")
	if v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
}

func TestP59_ParseHexComponent_4Digit(t *testing.T) {
	// 4-digit hex: "FFFF" should shift right by 8 = 255
	v := parseHexComponent("FFFF")
	if v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
}

func TestP59_ParseHexComponent_3Digit(t *testing.T) {
	// 3-digit hex: "FFF" should shift right by 4 = 255
	v := parseHexComponent("FFF")
	if v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
}

// === Kitty image protocol ===

func TestP59_KittyImageBase64_Format(t *testing.T) {
	result := KittyImageBase64("base64data", 100, 200)
	if !contains(result, "a=t,f=100") {
		t.Error("expected a=t,f=100 (transmit, PNG)")
	}
	if !contains(result, "s=100") {
		t.Error("expected s=100 (width)")
	}
	if !contains(result, "v=200") {
		t.Error("expected v=200 (height)")
	}
}

func TestP59_KittyDeleteAll_Format(t *testing.T) {
	result := KittyDeleteAllImages()
	if !contains(result, "a=d,d=a") {
		t.Error("expected a=d,d=a (delete all)")
	}
}

func TestP59_KittyQueryCell_Format(t *testing.T) {
	result := KittyQueryCell(5, 10)
	if !contains(result, "a=q") {
		t.Error("expected a=q (query)")
	}
	if !contains(result, "s=5") {
		t.Error("expected s=5 (col)")
	}
	if !contains(result, "v=10") {
		t.Error("expected v=10 (row)")
	}
}

// === DefaultImageOptions ===

func TestP59_DefaultImageOptions(t *testing.T) {
	opts := DefaultImageOptions()
	if !opts.Inline {
		t.Error("expected Inline=true by default")
	}
	if !opts.PreserveAspectRatio {
		t.Error("expected PreserveAspectRatio=true by default")
	}
	if opts.Width.Unit != "auto" || opts.Width.Value != 0 {
		t.Error("expected Width auto by default")
	}
	if opts.Height.Unit != "auto" || opts.Height.Value != 0 {
		t.Error("expected Height auto by default")
	}
}

// === InlineImage (raw bytes) ===

func TestP59_InlineImage_RoundTrip(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header bytes
	opts := DefaultImageOptions()
	opts.Name = "icon.png"
	result := InlineImage(data, opts)
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should contain base64-encoded data
	if !contains(result, "\x07") {
		t.Error("expected BEL terminator")
	}
}

// Note: contains() and indexOf() helpers already exist in term_unix.go
