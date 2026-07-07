package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/termcompat"
)

// ─── generateSequenceLocked coverage (33.3% → 90%+) ───

// TestP101_GenerateSequence_Iterm2 covers the iTerm2 protocol branch.
// Since generateSequenceLocked reads from env, we set TERM_PROGRAM.
func TestP101_GenerateSequence_Iterm2(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	b := NewImageBlock("img", "test.png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	b.SetDisplaySize(10, 5)

	seq := b.Sequence()
	if seq == "" {
		t.Fatal("expected non-empty sequence for iTerm2 protocol")
	}
	if b.Protocol() != termcompat.ImageIterm2 {
		t.Errorf("protocol: got %v, want ImageIterm2", b.Protocol())
	}
	// Verify it contains OSC 1337 prefix
	if !containsStr(seq, "1337") {
		t.Error("expected OSC 1337 in iTerm2 sequence")
	}
}

// TestP101_GenerateSequence_Kitty covers the Kitty protocol branch.
func TestP101_GenerateSequence_Kitty(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("TERM", "xterm-kitty")

	b := NewImageBlock("img", "test.png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A})
	b.SetDisplaySize(10, 5)

	seq := b.Sequence()
	if seq == "" {
		t.Fatal("expected non-empty sequence for Kitty protocol")
	}
	if b.Protocol() != termcompat.ImageKitty {
		t.Errorf("protocol: got %v, want ImageKitty", b.Protocol())
	}
	// Verify it contains Kitty APC sequence
	if !containsStr(seq, "G") {
		t.Error("expected 'G' (graphics) in Kitty sequence")
	}
}

// TestP101_GenerateSequence_SixelRGBA covers the Sixel protocol with RGBA data.
func TestP101_GenerateSequence_SixelRGBA(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "foot")

	// Create RGBA image block
	rgba := make([]byte, 4*4*4) // 4x4 RGBA
	for i := range rgba {
		rgba[i] = byte(i % 256)
	}
	b := NewRGBAImageBlock("img", "pixels.rgba", rgba, 4, 4)

	seq := b.Sequence()
	if seq == "" {
		t.Fatal("expected non-empty sequence for Sixel protocol with RGBA data")
	}
	if b.Protocol() != termcompat.ImageSixel {
		t.Errorf("protocol: got %v, want ImageSixel", b.Protocol())
	}
	// Verify it contains DCS (ESC P)
	if !containsStr(seq, "\x1bP") {
		t.Error("expected DCS prefix in Sixel sequence")
	}
}

// TestP101_GenerateSequence_SixelNonRGBA covers Sixel with non-RGBA data (should return empty).
func TestP101_GenerateSequence_SixelNonRGBA(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "foot")

	// PNG data, not RGBA — Sixel can't encode without pixel data
	b := NewImageBlock("img", "test.png", []byte{0x89, 'P', 'N', 'G'})

	seq := b.Sequence()
	if seq != "" {
		t.Error("expected empty sequence for Sixel with non-RGBA data")
	}
}

// TestP101_GenerateSequence_None covers ImageNone (default environment).
func TestP101_GenerateSequence_None(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "xterm-256color")

	b := NewImageBlock("img", "test.png", []byte("data"))

	seq := b.Sequence()
	if seq != "" {
		t.Error("expected empty sequence for ImageNone")
	}
	if b.Protocol() != termcompat.ImageNone {
		t.Errorf("protocol: got %v, want ImageNone", b.Protocol())
	}
}

// TestP101_GenerateSequence_SequenceCached verifies that Sequence() caches result.
func TestP101_GenerateSequence_SequenceCached(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	b := NewImageBlock("img", "test.png", []byte("testdata"))
	b.SetDisplaySize(5, 3)

	seq1 := b.Sequence()
	seq2 := b.Sequence()
	if seq1 != seq2 {
		t.Error("Sequence() should return cached value on second call")
	}
}

// TestP101_GenerateSequence_InvalidatedOnSetDisplaySize verifies cache invalidation.
func TestP101_GenerateSequence_InvalidatedOnSetDisplaySize(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	b := NewImageBlock("img", "test.png", []byte("testdata"))
	b.SetDisplaySize(5, 3)

	seq1 := b.Sequence()
	b.SetDisplaySize(10, 6)
	seq2 := b.Sequence()

	if seq1 == seq2 {
		t.Error("Sequence() should differ after SetDisplaySize")
	}
}

// ─── measureHeightLocked coverage (71.4% → 90%+) ───

func TestP101_MeasureHeight_WidthZero(t *testing.T) {
	b := NewImageBlock("img", "test.png", make([]byte, 1024))
	b.SetDisplaySize(0, 0)
	// measureHeightLocked with w=0 should return a minimal height
	height := b.measureHeightLocked(0)
	if height < 1 {
		t.Errorf("expected at least 1 line height for w=0, got %d", height)
	}
}

func TestP101_MeasureHeight_NarrowWidth(t *testing.T) {
	b := NewImageBlock("img", "very_long_filename_here_test.png", make([]byte, 100))
	height := b.measureHeightLocked(20)
	if height < 1 {
		t.Errorf("expected at least 1 line height for narrow width, got %d", height)
	}
}

func TestP101_MeasureHeight_WithDisplayDims(t *testing.T) {
	b := NewImageBlock("img", "test.png", make([]byte, 1024))
	b.SetDisplaySize(10, 5)
	height := b.measureHeightLocked(80)
	// Should be at least displayH (5) plus info lines (2)
	if height < 5 {
		t.Errorf("expected at least display height, got %d", height)
	}
}

func TestP101_MeasureHeight_NoFilename(t *testing.T) {
	b := NewImageBlock("img", "", make([]byte, 100))
	height := b.measureHeightLocked(80)
	if height < 1 {
		t.Errorf("expected at least 1 line height, got %d", height)
	}
}

// ─── AssistantTextBlock.Paint additional coverage (72.2% → 80%+) ───

func TestP101_AssistantTextBlock_Paint_HorizontalRule(t *testing.T) {
	// This is tested in existing tests, just verify it doesn't panic
	blk := NewAssistantTextBlock("test")
	blk.AppendDelta("---\n\nAfter horizontal rule\n")
	blk.Complete()
	blk.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	blk.Paint(buffer.NewBuffer(80, 10))
}

func TestP101_AssistantTextBlock_Paint_Blockquote(t *testing.T) {
	blk := NewAssistantTextBlock("test")
	blk.AppendDelta("> This is a quote\n> Second line\n")
	blk.Complete()
	blk.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	blk.Paint(buffer.NewBuffer(80, 10))
}

func TestP101_AssistantTextBlock_Paint_NestedFormatting(t *testing.T) {
	blk := NewAssistantTextBlock("test")
	blk.AppendDelta("**Bold _and italic_ text**\n")
	blk.Complete()
	blk.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	blk.Paint(buffer.NewBuffer(80, 10))
}

// ─── helpers ───

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
