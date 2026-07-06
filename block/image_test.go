package block

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/termcompat"
)

// ═══════════════════════════════════════════════════════════════════════════
// ImageBlock Tests
// ═══════════════════════════════════════════════════════════════════════════

// ─── Construction ───

func TestImageBlock_New(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A}
	b := NewImageBlock("img1", "photo.png", data)
	if b.ID() != "img1" {
		t.Errorf("ID = %q, want 'img1'", b.ID())
	}
	if b.Type() != TypeImage {
		t.Errorf("Type = %v, want TypeImage", b.Type())
	}
	if b.State() != BlockComplete {
		t.Errorf("State = %v, want BlockComplete", b.State())
	}
	if b.Filename() != "photo.png" {
		t.Errorf("Filename = %q, want 'photo.png'", b.Filename())
	}
	if b.Format() != "png" {
		t.Errorf("Format = %q, want 'png'", b.Format())
	}
	if len(b.Data()) != len(data) {
		t.Errorf("Data length = %d, want %d", len(b.Data()), len(data))
	}
}

func TestImageBlock_New_WithDims(t *testing.T) {
	data := []byte("fake image data")
	b := NewImageBlockWithDims("img2", "test.jpg", data, 1920, 1080)
	if b.Width() != 1920 {
		t.Errorf("Width = %d, want 1920", b.Width())
	}
	if b.Height() != 1080 {
		t.Errorf("Height = %d, want 1080", b.Height())
	}
}

func TestImageBlock_NewRGBA(t *testing.T) {
	// 2x2 RGBA image = 16 bytes
	rgba := make([]byte, 4*2*2)
	for i := range rgba {
		rgba[i] = byte(i)
	}
	b := NewRGBAImageBlock("img3", "pixels.rgba", rgba, 2, 2)
	if b.Format() != "rgba" {
		t.Errorf("Format = %q, want 'rgba'", b.Format())
	}
	if b.Width() != 2 || b.Height() != 2 {
		t.Errorf("Dims = %dx%d, want 2x2", b.Width(), b.Height())
	}
}

func TestImageBlock_FormatDetection(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     []byte
		want     string
	}{
		{"png ext", "a.png", nil, "png"},
		{"jpg ext", "a.jpg", nil, "jpeg"},
		{"jpeg ext", "a.jpeg", nil, "jpeg"},
		{"gif ext", "a.gif", nil, "gif"},
		{"bmp ext", "a.bmp", nil, "bmp"},
		{"webp ext", "a.webp", nil, "webp"},
		{"png magic", "unknown", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A}, "png"},
		{"jpeg magic", "unknown", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "jpeg"},
		{"gif magic", "unknown", []byte{'G', 'I', 'F', '8', '9', 'a'}, "gif"},
		{"bmp magic", "unknown", []byte{'B', 'M', 0x00, 0x00}, "bmp"},
		{"unknown", "noext", []byte{1, 2, 3}, "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := NewImageBlock("t", tc.filename, tc.data)
			if b.Format() != tc.want {
				t.Errorf("Format = %q, want %q", b.Format(), tc.want)
			}
		})
	}
}

// ─── Display Size ───

func TestImageBlock_DisplaySize(t *testing.T) {
	b := NewImageBlock("img", "test.png", []byte("data"))
	if b.DisplayWidth() != 0 || b.DisplayHeight() != 0 {
		t.Error("default display size should be 0,0 (auto)")
	}
	b.SetDisplaySize(40, 20)
	if b.DisplayWidth() != 40 || b.DisplayHeight() != 20 {
		t.Errorf("Display = %dx%d, want 40x20", b.DisplayWidth(), b.DisplayHeight())
	}
}

// ─── File Size ───

func TestImageBlock_FileSize(t *testing.T) {
	tests := []struct {
		bytes int
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tc := range tests {
		b := NewImageBlock("img", "t.png", make([]byte, tc.bytes))
		got := b.FileSize()
		if got != tc.want {
			t.Errorf("FileSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

// ─── Protocol Detection ───

func TestImageBlock_Protocol(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	// Protocol depends on environment — just verify it doesn't panic
	_ = b.Protocol()
	_ = b.ProtocolName()
}

func TestImageBlock_HasImageSupport(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	// Just verify it returns a bool without panicking
	_ = b.HasImageSupport()
}

// ─── Sequence ───

func TestImageBlock_Sequence_Cached(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	seq1 := b.Sequence()
	seq2 := b.Sequence()
	// Both calls should return the same value (cached)
	if seq1 != seq2 {
		t.Error("Sequence should be cached and return same value")
	}
}

func TestImageBlock_Sequence_SequenceInvalidatedOnSetDisplaySize(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	_ = b.Sequence()
	b.SetDisplaySize(40, 20)
	// After SetDisplaySize, the cache should be invalidated
	seq := b.Sequence()
	_ = seq // just verify no panic
}

// ─── Measure ───

func TestImageBlock_Measure(t *testing.T) {
	b := NewImageBlock("img", "photo.png", make([]byte, 10240))
	b.mu.Lock()
	b.imgW = 256
	b.imgH = 256
	b.mu.Unlock()
	s := b.Measure(component.Bounded(80, 20))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("Measure = %v, expected positive dimensions", s)
	}
	if s.W > 80 {
		t.Errorf("Measure W = %d, should be <= 80", s.W)
	}
}

func TestImageBlock_Measure_MinWidth(t *testing.T) {
	b := NewImageBlock("img", "x.png", make([]byte, 100))
	s := b.Measure(component.Bounded(5, 5))
	if s.W < 10 {
		t.Errorf("Measure W = %d, should be >= 10", s.W)
	}
}

func TestImageBlock_Measure_NarrowConstraints(t *testing.T) {
	b := NewImageBlock("img", "very_long_filename_that_exceeds_width.png", make([]byte, 100))
	s := b.Measure(component.Bounded(20, 10))
	if s.W > 20 {
		t.Errorf("Measure W = %d, should be <= 20", s.W)
	}
}

// ─── Paint ───

func TestImageBlock_Paint(t *testing.T) {
	b := NewImageBlock("img1", "photo.png", make([]byte, 10240))
	b.mu.Lock()
	b.imgW = 256
	b.imgH = 256
	b.mu.Unlock()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)

	// Check top border
	cell := buf.GetCell(0, 0)
	if cell.Rune != '┌' {
		t.Errorf("top-left = %q, want '┌'", string(cell.Rune))
	}
	// Check bottom border
	cell = buf.GetCell(0, 4)
	if cell.Rune != '└' {
		t.Errorf("bottom-left = %q, want '└'", string(cell.Rune))
	}
	// Check content exists — find "photo.png" somewhere
	found := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			if buf.GetCell(x, y).Rune == 'p' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'p' from filename in paint output")
	}
}

func TestImageBlock_Paint_ZeroBounds(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should not panic
}

func TestImageBlock_Paint_NarrowWidth(t *testing.T) {
	b := NewImageBlock("img", "very_long_filename_here.png", make([]byte, 2048))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should not panic, text should be truncated
}

func TestImageBlock_Paint_NoFilename(t *testing.T) {
	b := NewImageBlock("img", "", []byte{0x89, 'P', 'N', 'G'})
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	b.Paint(buf)
	// Should show "image" as fallback filename
	found := false
	for y := 0; y < 4; y++ {
		for x := 0; x < 30; x++ {
			if buf.GetCell(x, y).Rune == 'i' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'i' from 'image' fallback in paint output")
	}
}

func TestImageBlock_Paint_WithOffset(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	b.SetBounds(component.Rect{X: 5, Y: 2, W: 30, H: 4})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	// Check border at offset
	cell := buf.GetCell(5, 2)
	if cell.Rune != '┌' {
		t.Errorf("top-left at offset = %q, want '┌'", string(cell.Rune))
	}
}

func TestImageBlock_Paint_WithProtocol(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	b.mu.Lock()
	b.protocol = termcompat.ImageIterm2
	b.mu.Unlock()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	b.Paint(buf)
	// Protocol line should be present
	found := false
	for y := 0; y < 6; y++ {
		for x := 0; x < 40; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == 'I' { // iTerm2 starts with 'I'
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'I' from iTerm2 protocol name in paint output")
	}
}

// ─── Serialize / Deserialize ───

func TestImageBlock_SerializeDeserialize(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x00, 0x00}
	b := NewImageBlockWithDims("img1", "photo.png", data, 256, 256)
	b.SetDisplaySize(40, 20)

	raw, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	b2 := NewImageBlock("img1", "", nil)
	if err := b2.DeserializeState(raw); err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}

	if b2.Filename() != "photo.png" {
		t.Errorf("Filename = %q, want 'photo.png'", b2.Filename())
	}
	if b2.Format() != "png" {
		t.Errorf("Format = %q, want 'png'", b2.Format())
	}
	if b2.Width() != 256 || b2.Height() != 256 {
		t.Errorf("Dims = %dx%d, want 256x256", b2.Width(), b2.Height())
	}
	if b2.DisplayWidth() != 40 || b2.DisplayHeight() != 20 {
		t.Errorf("Display = %dx%d, want 40x20", b2.DisplayWidth(), b2.DisplayHeight())
	}
	if len(b2.Data()) != len(data) {
		t.Errorf("Data length = %d, want %d", len(b2.Data()), len(data))
	}
}

func TestImageBlock_Deserialize_InvalidJSON(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	err := b.DeserializeState(json.RawMessage("not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImageBlock_Deserialize_EmptyData(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	emptyState, _ := json.Marshal(struct{}{})
	err := b.DeserializeState(emptyState)
	if err != nil {
		t.Errorf("unexpected error for empty state: %v", err)
	}
}

// ─── IsDirty ───

func TestImageBlock_IsDirty(t *testing.T) {
	b := NewImageBlock("img", "t.png", []byte("data"))
	if !b.IsDirty() {
		t.Error("new block should be dirty")
	}
	b.ClearDirty()
	if b.IsDirty() {
		t.Error("block should not be dirty after ClearDirty")
	}
	b.SetDisplaySize(40, 20)
	if !b.IsDirty() {
		t.Error("block should be dirty after SetDisplaySize")
	}
}

// ─── Concurrent Access ───

func TestImageBlock_ConcurrentPaintAndSerialize(t *testing.T) {
	b := NewImageBlock("img", "test.png", make([]byte, 2048))
	b.mu.Lock()
	b.imgW = 100
	b.imgH = 100
	b.mu.Unlock()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			buf := buffer.NewBuffer(40, 5)
			b.Paint(buf)
			_, _ = b.SerializeState()
			_ = b.Sequence()
		}(i)
	}
	wg.Wait()
}

func TestImageBlock_ConcurrentGetters(t *testing.T) {
	b := NewImageBlock("img", "test.png", make([]byte, 1024))
	b.mu.Lock()
	b.imgW = 200
	b.imgH = 200
	b.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Filename()
			_ = b.Format()
			_ = b.Data()
			_ = b.Width()
			_ = b.Height()
			_ = b.FileSize()
			_ = b.Protocol()
			_ = b.HasImageSupport()
			_ = b.Sequence()
		}()
	}
	wg.Wait()
}

// ─── Edge Cases ───

func TestImageBlock_EmptyData(t *testing.T) {
	b := NewImageBlock("img", "empty.png", []byte{})
	if len(b.Data()) != 0 {
		t.Error("empty data should remain empty")
	}
	if b.FileSize() != "0 B" {
		t.Errorf("FileSize = %q, want '0 B'", b.FileSize())
	}
}

func TestImageBlock_NilData(t *testing.T) {
	// NewImageBlock with nil data — registry uses this
	b := NewImageBlock("img", "", nil)
	_ = b.FileSize()
	_ = b.Format()
	// Should not panic
}

func TestImageBlock_TypeString(t *testing.T) {
	if TypeImage.String() != "image" {
		t.Errorf("TypeImage.String() = %q, want 'image'", TypeImage.String())
	}
}

// ─── Registry ───

func TestImageBlock_Registered(t *testing.T) {
	r := NewDefaultRegistry()
	if !r.Has("image") {
		t.Error("image type should be registered in default registry")
	}
	b, err := r.Create("image", "test-id")
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if b.Type() != TypeImage {
		t.Errorf("Type = %v, want TypeImage", b.Type())
	}
}

// ─── Info/Meta Lines (internal) ───

func TestImageBlock_InfoLine(t *testing.T) {
	b := NewImageBlock("img", "photo.png", make([]byte, 1000))
	b.mu.Lock()
	b.imgW = 256
	b.imgH = 256
	info := b.infoLineLocked()
	b.mu.Unlock()

	if !strings.Contains(info, "photo.png") {
		t.Errorf("info line should contain filename: %q", info)
	}
	if !strings.Contains(info, "256x256") {
		t.Errorf("info line should contain dimensions: %q", info)
	}
}

func TestImageBlock_MetaLine(t *testing.T) {
	b := NewImageBlock("img", "t.png", make([]byte, 2048))
	b.mu.Lock()
	meta := b.metaLineLocked()
	b.mu.Unlock()

	if !strings.Contains(meta, "KB") {
		t.Errorf("meta line should contain file size: %q", meta)
	}
	if !strings.Contains(meta, "PNG") {
		t.Errorf("meta line should contain format: %q", meta)
	}
}
