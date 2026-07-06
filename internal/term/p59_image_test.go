package term

import (
	"encoding/base64"
	"strings"
	"testing"
)

// --- ImageSize and ImageOptions ---

func TestAutoSize(t *testing.T) {
	if AutoSize.Unit != "auto" {
		t.Errorf("expected AutoSize.Unit='auto', got %q", AutoSize.Unit)
	}
}

func TestDefaultImageOptions(t *testing.T) {
	opts := DefaultImageOptions()
	if !opts.PreserveAspectRatio {
		t.Error("expected PreserveAspectRatio=true by default")
	}
	if !opts.Inline {
		t.Error("expected Inline=true by default")
	}
	if opts.Width.Unit != "auto" {
		t.Errorf("expected Width auto, got %q", opts.Width.Unit)
	}
	if opts.Height.Unit != "auto" {
		t.Errorf("expected Height auto, got %q", opts.Height.Unit)
	}
}

// --- InlineImage ---

func TestInlineImage_Prefix(t *testing.T) {
	data := []byte("fake image data")
	seq := InlineImage(data, DefaultImageOptions())
	if !strings.HasPrefix(seq, "\x1b]1337;File=") {
		t.Error("expected OSC 1337 prefix")
	}
	if !strings.HasSuffix(seq, "\x07") {
		t.Error("expected BEL terminator")
	}
}

func TestInlineImage_ContainsBase64(t *testing.T) {
	data := []byte("test data")
	seq := InlineImage(data, DefaultImageOptions())
	expectedB64 := base64.StdEncoding.EncodeToString(data)
	if !strings.Contains(seq, expectedB64) {
		t.Error("expected base64 encoded data in sequence")
	}
}

func TestInlineImage_ContainsSize(t *testing.T) {
	data := []byte("hello world")
	seq := InlineImage(data, DefaultImageOptions())
	if !strings.Contains(seq, "size=11;") {
		t.Errorf("expected size=11; in sequence, got partial: %s", seq[:50])
	}
}

func TestInlineImage_WithName(t *testing.T) {
	data := []byte("data")
	opts := DefaultImageOptions()
	opts.Name = "photo.png"
	seq := InlineImage(data, opts)
	// name should be base64 encoded
	expectedNameB64 := base64.StdEncoding.EncodeToString([]byte("photo.png"))
	if !strings.Contains(seq, "name="+expectedNameB64) {
		t.Error("expected base64-encoded name in sequence")
	}
}

func TestInlineImage_WidthCells(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.Width = ImageSize{Value: 10, Unit: ""}
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "width=10;") {
		t.Error("expected width=10; in sequence")
	}
}

func TestInlineImage_WidthPixels(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.Width = ImageSize{Value: 200, Unit: "px"}
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "width=200px;") {
		t.Error("expected width=200px; in sequence")
	}
}

func TestInlineImage_WidthPercent(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.Width = ImageSize{Value: 50, Unit: "%"}
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "width=50%;") {
		t.Error("expected width=50%; in sequence")
	}
}

func TestInlineImage_HeightCells(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.Height = ImageSize{Value: 5, Unit: ""}
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "height=5;") {
		t.Error("expected height=5; in sequence")
	}
}

func TestInlineImage_PreserveAspectRatio(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.PreserveAspectRatio = true
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "preserveAspectRatio=1;") {
		t.Error("expected preserveAspectRatio=1;")
	}

	opts.PreserveAspectRatio = false
	seq = InlineImage(data, opts)
	if !strings.Contains(seq, "preserveAspectRatio=0;") {
		t.Error("expected preserveAspectRatio=0;")
	}
}

func TestInlineImage_InlineFlag(t *testing.T) {
	data := []byte("x")
	opts := DefaultImageOptions()
	opts.Inline = true
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "inline=1;") {
		t.Error("expected inline=1;")
	}

	opts.Inline = false
	seq = InlineImage(data, opts)
	if !strings.Contains(seq, "inline=0;") {
		t.Error("expected inline=0;")
	}
}

func TestInlineImage_EmptyData(t *testing.T) {
	seq := InlineImage([]byte{}, DefaultImageOptions())
	if !strings.Contains(seq, "size=0;") {
		t.Error("expected size=0; for empty data")
	}
}

func TestInlineImage_LargeData(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	seq := InlineImage(data, DefaultImageOptions())
	if !strings.Contains(seq, "size=1024;") {
		t.Error("expected size=1024;")
	}
}

// --- InlineImageBase64 ---

func TestInlineImageBase64_Basic(t *testing.T) {
	data := []byte("test")
	b64 := base64.StdEncoding.EncodeToString(data)
	seq := InlineImageBase64(b64, DefaultImageOptions())
	if !strings.Contains(seq, b64) {
		t.Error("expected pre-encoded base64 data in sequence")
	}
}

func TestInlineImageBase64_NoSize(t *testing.T) {
	b64 := base64.StdEncoding.EncodeToString([]byte("x"))
	seq := InlineImageBase64(b64, DefaultImageOptions())
	// InlineImageBase64 does not add size= since it doesn't know raw size
	if strings.Contains(seq, "size=") {
		t.Error("expected no size= in pre-encoded variant")
	}
}

// --- Kitty Graphics Protocol ---

func TestKittyImageBase64_Prefix(t *testing.T) {
	seq := KittyImageBase64("aGVsbG8=", 100, 50)
	if !strings.HasPrefix(seq, "\x1b_G") {
		t.Error("expected Kitty graphics prefix ESC _ G")
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Error("expected ST terminator ESC \\")
	}
}

func TestKittyImageBase64_Params(t *testing.T) {
	seq := KittyImageBase64("aGVsbG8=", 200, 100)
	if !strings.Contains(seq, "a=t,") {
		t.Error("expected action=transmit (a=t)")
	}
	if !strings.Contains(seq, "f=100,") {
		t.Error("expected format=PNG (f=100)")
	}
	if !strings.Contains(seq, "s=200,") {
		t.Error("expected width s=200")
	}
	if !strings.Contains(seq, "v=100") {
		t.Error("expected height v=100")
	}
}

func TestKittyImageBase64_ContainsData(t *testing.T) {
	seq := KittyImageBase64("aGVsbG8=", 10, 10)
	if !strings.Contains(seq, "aGVsbG8=") {
		t.Error("expected base64 data in sequence")
	}
}

func TestKittyDeleteAllImages(t *testing.T) {
	seq := KittyDeleteAllImages()
	if seq != "\x1b_Ga=d,d=a\x1b\\" {
		t.Errorf("unexpected delete sequence: %q", seq)
	}
}

func TestKittyQueryCell(t *testing.T) {
	seq := KittyQueryCell(5, 10)
	if !strings.Contains(seq, "a=q") {
		t.Error("expected action=query (a=q)")
	}
	if !strings.Contains(seq, "s=5") {
		t.Error("expected col s=5")
	}
	if !strings.Contains(seq, "v=10") {
		t.Error("expected row v=10")
	}
}

func TestKittyQueryCell_Origin(t *testing.T) {
	seq := KittyQueryCell(0, 0)
	if !strings.Contains(seq, "s=0") {
		t.Error("expected col s=0")
	}
	if !strings.Contains(seq, "v=0") {
		t.Error("expected row v=0")
	}
}

// --- Round-trip / integration ---

func TestInlineImage_FullFormat(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	opts := ImageOptions{
		Name:                "test.png",
		Width:               ImageSize{Value: 20, Unit: ""},
		Height:              ImageSize{Value: 10, Unit: ""},
		PreserveAspectRatio: true,
		Inline:              true,
	}
	seq := InlineImage(data, opts)

	// Verify all expected parts are present
	parts := []string{
		"\x1b]1337;File=",
		"name=",
		"size=4;",
		"width=20;",
		"height=10;",
		"preserveAspectRatio=1;",
		"inline=1;",
		"\x07",
	}
	for _, p := range parts {
		if !strings.Contains(seq, p) {
			t.Errorf("expected %q in sequence", p)
		}
	}
}

func TestInlineImage_DownloadMode(t *testing.T) {
	data := []byte("file content")
	opts := DefaultImageOptions()
	opts.Inline = false
	seq := InlineImage(data, opts)
	if !strings.Contains(seq, "inline=0;") {
		t.Error("expected inline=0 for download mode")
	}
}

// --- Benchmark ---

func BenchmarkInlineImage_Small(b *testing.B) {
	data := []byte("small image data for benchmarking")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InlineImage(data, DefaultImageOptions())
	}
}

func BenchmarkInlineImage_Large(b *testing.B) {
	data := make([]byte, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InlineImage(data, DefaultImageOptions())
	}
}

func BenchmarkKittyImageBase64(b *testing.B) {
	b64Data := base64.StdEncoding.EncodeToString(make([]byte, 1024))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KittyImageBase64(b64Data, 200, 100)
	}
}
