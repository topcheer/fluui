package buffer

import (
	"testing"
)

// FuzzBufferSetCell tests SetCell/GetCell with random coordinates.
// The buffer must never panic on out-of-bounds or negative coordinates.
func FuzzBufferSetCell(f *testing.F) {
	seeds := [][]int{
		{10, 5, 0, 0},
		{10, 5, 5, 2},
		{1, 1, 0, 0},
		{10, 5, -1, 0},
		{10, 5, 0, -1},
		{10, 5, 100, 100},
		{10, 5, 10, 5},
		{10, 5, -100, -100},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}

	f.Fuzz(func(t *testing.T, w, h, x, y int) {
		// Skip absurd dimensions
		if w < 1 || h < 1 || w > 1000 || h > 1000 {
			t.Skip()
		}

		b := NewBuffer(w, h)

		// SetCell must not panic on any coordinates
		c := Cell{
			Rune:  'X',
			Width: 1,
		}
		b.SetCell(x, y, c)

		// GetCell must not panic on any coordinates
		_ = b.GetCell(x, y)

		// Fill must not panic
		b.Fill(BlankCell)
	})
}

// FuzzBufferDrawText tests DrawText with random text content and positions.
func FuzzBufferDrawText(f *testing.F) {
	seeds := []struct {
		w, h int
		x, y int
		text string
	}{
		{20, 10, 0, 0, "hello"},
		{20, 10, 0, 0, ""},
		{20, 10, 5, 3, "世界"},
		{5, 1, 0, 0, "this is a very long line"},
		{20, 10, -1, -1, "negative"},
		{20, 10, 0, 0, "\x00\x01\x02"},
		{1, 1, 0, 0, "a"},
	}
	for _, s := range seeds {
		f.Add(s.w, s.h, s.x, s.y, s.text)
	}

	f.Fuzz(func(t *testing.T, w, h, x, y int, text string) {
		if w < 1 || h < 1 || w > 1000 || h > 1000 {
			t.Skip()
		}
		if len(text) > 1000 {
			t.Skip()
		}

		b := NewBuffer(w, h)
		style := Style{}

		// None of these should panic
		b.DrawText(x, y, text, style)
		b.DrawTextClamped(x, y, text, style)
		b.Fill(BlankCell)
	})
}

// FuzzBufferBlit tests Blit with random source/dest coordinates.
func FuzzBufferBlit(f *testing.F) {
	seeds := [][]int{
		{10, 10, 10, 10, 0, 0, 0, 0, 5, 5},
		{10, 10, 10, 10, 5, 5, 5, 5, 10, 10},
		{10, 10, 10, 10, -1, 0, 0, 0, 5, 5},
		{10, 10, 10, 10, 0, 0, 0, 0, -1, -1},
	}
	for _, s := range seeds {
		f.Add(s[0], s[1], s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9])
	}

	f.Fuzz(func(t *testing.T, sw, sh, dw, dh, sx, sy, dx, dy, w, h int) {
		if sw < 1 || sh < 1 || dw < 1 || dh < 1 ||
			sw > 100 || sh > 100 || dw > 100 || dh > 100 {
			t.Skip()
		}

		src := NewBuffer(sw, sh)
		dst := NewBuffer(dw, dh)

		// Blit must not panic on any coordinates
		dst.Blit(src, sx, sy, dx, dy, w, h)
	})
}
