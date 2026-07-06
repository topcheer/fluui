package term

import (
	"strings"
	"testing"
)

// ─── EncodeSixel basic tests ───

func TestEncodeSixel_EmptyImage(t *testing.T) {
	result := EncodeSixel(nil, 0, 0)
	if result != "" {
		t.Errorf("EncodeSixel(nil, 0, 0) = %q, want empty", result)
	}
}

func TestEncodeSixel_ZeroWidth(t *testing.T) {
	result := EncodeSixel([]byte{0, 0, 0, 255}, 0, 1)
	if result != "" {
		t.Errorf("EncodeSixel with width=0 = %q, want empty", result)
	}
}

func TestEncodeSixel_SinglePixel(t *testing.T) {
	// 1x1 red pixel
	rgba := []byte{255, 0, 0, 255}
	result := EncodeSixel(rgba, 1, 1)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Must start with DCS introducer
	if !strings.HasPrefix(result, "\x1bP") {
		t.Errorf("result must start with DCS (ESC P), got %q", result[:5])
	}
	// Must end with string terminator
	if !strings.HasSuffix(result, "\x1b\\") {
		t.Errorf("result must end with ST (ESC \\)")
	}
	// Must contain raster attributes (")
	if !strings.Contains(result, "\"") {
		t.Error("result must contain raster attributes")
	}
	// Must contain color register (#0;2;...)
	if !strings.Contains(result, "#0;2;") {
		t.Error("result must contain RGB color register")
	}
}

func TestEncodeSixel_SingleColor(t *testing.T) {
	// 4x6 all-red image
	width, height := 4, 6
	rgba := make([]byte, width*height*4)
	for i := 0; i < width*height; i++ {
		rgba[i*4] = 255
		rgba[i*4+3] = 255
	}
	result := EncodeSixel(rgba, width, height)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Should have exactly 1 color in palette
	// Color register definition appears once: #0;2;100;0;0
	if strings.Count(result, ";2;") != 1 {
		t.Errorf("expected 1 color register, got %d occurrences of ';2;'", strings.Count(result, ";2;"))
	}
	// Sixel data should be all-on for red: 0x3F + 0x3F = 0x7E = '~'
	// Each column has 6 vertical pixels all on → sixel = 0x3F (all on = 0x3F? no...)
	// Actually: all 6 bits on = 0x3F, character = 0x3F + 0x3F = 0x7E = '~'
	if !strings.Contains(result, "~") {
		t.Error("expected '~' (all pixels on) in sixel data")
	}
}

func TestEncodeSixel_TwoColors(t *testing.T) {
	// 2x1: red and blue
	rgba := []byte{
		255, 0, 0, 255, // red
		0, 0, 255, 255, // blue
	}
	result := EncodeSixel(rgba, 2, 1)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Color register format: #N;2;R;G;B. Count occurrences of "#0;2;" and "#1;2;"
	// (note: raster attributes may contain ;2; for width=2, so we check specifically)
	if !strings.Contains(result, "#0;2;") {
		t.Error("expected color register #0")
	}
	if !strings.Contains(result, "#1;2;") {
		t.Error("expected color register #1")
	}
}

func TestEncodeSixel_MultiBand(t *testing.T) {
	// 1x12 image (2 bands of 6)
	width, height := 1, 12
	rgba := make([]byte, width*height*4)
	for i := 0; i < width*height; i++ {
		rgba[i*4] = uint8(i * 10)
		rgba[i*4+1] = uint8(i * 5)
		rgba[i*4+2] = uint8(255 - i*10)
		rgba[i*4+3] = 255
	}
	result := EncodeSixel(rgba, width, height)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Should contain a newline character (-) between bands
	if !strings.Contains(result, "-") {
		t.Error("expected band separator '-' in multi-band image")
	}
}

func TestEncodeSixel_DCSFormat(t *testing.T) {
	rgba := []byte{0, 255, 0, 255} // 1x1 green
	result := EncodeSixel(rgba, 1, 1)

	// Verify DCS introducer format: ESC P 8 ; ; ; q
	if !strings.HasPrefix(result, "\x1bP8;;;q") {
		t.Errorf("expected ESC P 8 ; ; ; q prefix, got %q", result[:min(10, len(result))])
	}
}

func TestEncodeSixel_ColorPercentages(t *testing.T) {
	// Pure green: G=255, R=0, B=0 → percentage 0;100;0
	rgba := []byte{0, 255, 0, 255}
	result := EncodeSixel(rgba, 1, 1)

	// Green should be 100%, R and B should be 0%
	// Format: #0;2;0;100;0
	if !strings.Contains(result, "#0;2;0;100;0") {
		t.Errorf("expected '#0;2;0;100;0' for pure green, got: %s", result)
	}
}

func TestEncodeSixel_PaletteLimit(t *testing.T) {
	// Create image with more than 256 unique colors
	width, height := 300, 1
	rgba := make([]byte, width*height*4)
	for i := 0; i < width; i++ {
		rgba[i*4] = uint8(i)       // R: 0-255
		rgba[i*4+1] = uint8(i / 2) // G
		rgba[i*4+2] = uint8(i / 3) // B
		rgba[i*4+3] = 255
	}
	// Should not crash, should produce output
	result := EncodeSixel(rgba, width, height)
	if result == "" {
		t.Fatal("expected non-empty result even with >256 colors")
	}
}

// ─── EncodeSixelSimple tests ───

func TestEncodeSixelSimple_EmptyImage(t *testing.T) {
	result := EncodeSixelSimple(nil, 0, 0)
	if result != "" {
		t.Errorf("EncodeSixelSimple(nil, 0, 0) = %q, want empty", result)
	}
}

func TestEncodeSixelSimple_SinglePixel(t *testing.T) {
	gray := []byte{128} // mid gray
	result := EncodeSixelSimple(gray, 1, 1)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if !strings.HasPrefix(result, "\x1bP") {
		t.Error("result must start with DCS")
	}
	if !strings.HasSuffix(result, "\x1b\\") {
		t.Error("result must end with ST")
	}
}

func TestEncodeSixelSimple_GrayscalePalette(t *testing.T) {
	// 4x1: black, dark, light, white
	gray := []byte{0, 85, 170, 255}
	result := EncodeSixelSimple(gray, 4, 1)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Should have 4 grayscale color registers
	if strings.Count(result, ";2;") != 4 {
		t.Errorf("expected 4 color registers, got %d", strings.Count(result, ";2;"))
	}
}

func TestEncodeSixelSimple_MultiBand(t *testing.T) {
	gray := make([]byte, 12) // 1x12
	for i := range gray {
		gray[i] = byte(i * 20)
	}
	result := EncodeSixelSimple(gray, 1, 12)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if !strings.Contains(result, "-") {
		t.Error("expected band separator in multi-band image")
	}
}

// ─── quantizeColors tests ───

func TestQuantizeColors_SingleColor(t *testing.T) {
	rgba := []byte{
		100, 200, 50, 255,
		100, 200, 50, 255,
		100, 200, 50, 255,
	}
	palette, indices := quantizeColors(rgba, 3)
	if len(palette) != 1 {
		t.Errorf("expected 1 color, got %d", len(palette))
	}
	if palette[0].r != 100 || palette[0].g != 200 || palette[0].b != 50 {
		t.Errorf("palette color = %+v, want {100, 200, 50}", palette[0])
	}
	for i, idx := range indices {
		if idx != 0 {
			t.Errorf("indices[%d] = %d, want 0", i, idx)
		}
	}
}

func TestQuantizeColors_MultipleColors(t *testing.T) {
	rgba := []byte{
		255, 0, 0, 255,
		0, 255, 0, 255,
		0, 0, 255, 255,
	}
	palette, _ := quantizeColors(rgba, 3)
	if len(palette) != 3 {
		t.Errorf("expected 3 colors, got %d", len(palette))
	}
}

func TestQuantizeColors_OverLimit(t *testing.T) {
	// 300 unique colors
	rgba := make([]byte, 300*4)
	for i := 0; i < 300; i++ {
		rgba[i*4] = uint8(i % 256)
		rgba[i*4+1] = uint8((i + 50) % 256)
		rgba[i*4+2] = uint8((i + 100) % 256)
		rgba[i*4+3] = 255
	}
	palette, _ := quantizeColors(rgba, 300)
	if len(palette) > 256 {
		t.Errorf("palette should be capped at 256, got %d", len(palette))
	}
}

// ─── Sixel character encoding verification ───

func TestEncodeSixel_AllPixelsOff(t *testing.T) {
	// 1x6 black image — all pixels in color 0 (black)
	rgba := make([]byte, 1*6*4) // all zeros = black
	result := EncodeSixel(rgba, 1, 6)

	// Black is color 0. The sixel data for black should be all 6 bits on
	// (because all pixels match color 0). So the sixel char = 0x3F + 0x3F = '~'
	if !strings.Contains(result, "~") {
		t.Error("expected '~' (all pixels of color 0 on) in sixel data")
	}
}

func TestEncodeSixel_AllPixelsOn(t *testing.T) {
	// 1x6 white image — all pixels on
	rgba := make([]byte, 1*6*4)
	for i := 0; i < 6; i++ {
		rgba[i*4] = 255
		rgba[i*4+1] = 255
		rgba[i*4+2] = 255
		rgba[i*4+3] = 255
	}
	result := EncodeSixel(rgba, 1, 6)

	// White pixel: all bits on for white color → sixel char = '?' + 0x3F = '~'
	if !strings.Contains(result, "~") {
		t.Error("expected '~' (all pixels on) in sixel data")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
