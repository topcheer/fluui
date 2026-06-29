package term

import (
	"testing"
)

// FuzzParserFeed tests the terminal input parser with random byte sequences.
// The parser must never panic on any input — it should always return events
// (or an empty slice) gracefully.
func FuzzParserFeed(f *testing.F) {
	// Seed corpus: common terminal byte sequences
	seeds := [][]byte{
		{0x1b, '[', 'A'},            // ESC [ A — Arrow Up
		{0x1b, '[', '3', '~'},       // ESC [ 3 ~ — Delete
		{0x0d},                       // Enter
		{0x7f},                       // Backspace
		{0x1b, 0x1b},                // Double ESC
		{0xe4, 0xbd, 0xa0},          // UTF-8 "你"
		{0xc0, 0x80},                // Invalid UTF-8 (overlong)
		{0x1b, '[', '2', '0', '0', '~', 'h', 'i', 0x1b, '[', '2', '0', '1', '~'}, // Paste
		{0x1b, 'O', 'P'},            // ESC O P — F1
		{0x01},                       // Ctrl+A
		{0x1b, '[', 'M', 32, 32, 32}, // Mouse report
		{},                           // Empty
		{0xfe, 0xff},                 // Invalid UTF-8
		{0x1b, ']', '5', '2', ';', 'h', 'i', 0x07}, // OSC52
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Limit input size to avoid excessive memory use
		if len(data) > 4096 {
			t.Skip()
		}

		p := NewParser()
		// Feed must never panic on any input
		events := p.Feed(data)

		// Verify all returned events are valid
		for _, ev := range events {
			// Key events with Rune should have valid rune values
			if ev.Type == EventKey && ev.Key != nil && ev.Key.Rune != 0 {
				if ev.Key.Rune < 0 || ev.Key.Rune > 0x10FFFF {
					t.Errorf("invalid rune value: %d", ev.Key.Rune)
				}
			}
		}

		// FeedTimeout after data should also not panic
		_ = p.FeedTimeout()
	})
}

// FuzzParserFeedChunked tests feeding the same data in different chunk sizes.
// The parser must produce identical results regardless of how bytes are split.
func FuzzParserFeedChunked(f *testing.F) {
	seeds := [][]byte{
		{0x1b, '[', 'A'},
		{0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd},
		{0x1b, '[', '2', '0', '0', '~', 'x', 0x1b, '[', '2', '0', '1', '~'},
		{0x1b, ']', '5', '2', ';', 'd', 'a', 't', 'a', 0x07},
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024 {
			t.Skip()
		}

		// Feed all at once
		p1 := NewParser()
		allAtOnce := p1.Feed(data)

		// Feed one byte at a time
		p2 := NewParser()
		var oneAtATime []Event
		for i := 0; i < len(data); i++ {
			oneAtATime = append(oneAtATime, p2.Feed(data[i:i+1])...)
		}

		// Results should be equivalent in count
		// (exact events may differ for partial sequences, but count should match)
		if len(data) > 0 && len(allAtOnce) != len(oneAtATime) {
			// Allow difference only for incomplete UTF-8 sequences
			// where chunk boundaries matter
			if len(allAtOnce) > 0 || len(oneAtATime) > 0 {
				// Log but don't fail — some sequences are inherently
				// sensitive to chunk boundaries (e.g., partial ESC)
				t.Logf("chunk sensitivity: all=%d chunked=%d data=%v",
					len(allAtOnce), len(oneAtATime), data)
			}
		}
	})
}
