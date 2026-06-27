package buffer

import (
	"testing"
)

// BenchmarkDiffIdentical benchmarks diffing two identical buffers.
// The row-skip optimization should make this very fast.
func BenchmarkDiffIdentical(b *testing.B) {
	front := NewBuffer(80, 24)
	back := NewBuffer(80, 24)

	// Fill both with the same content.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			c := Cell{Rune: rune('A' + (x+y)%26), Width: 1, Fg: RGB(100, 200, 50)}
			front.SetCell(x, y, c)
			back.SetCell(x, y, c)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Diff(front, back)
	}
}

// BenchmarkDiffSmall benchmarks diffing buffers with 1 cell different.
// Tests the row-skip optimization when most rows are identical.
func BenchmarkDiffSmall(b *testing.B) {
	front := NewBuffer(80, 24)
	back := NewBuffer(80, 24)

	// Fill both with the same content.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			c := Cell{Rune: rune('A' + (x+y)%26), Width: 1, Fg: RGB(100, 200, 50)}
			front.SetCell(x, y, c)
			back.SetCell(x, y, c)
		}
	}
	// Change 1 cell in back.
	back.SetCell(10, 5, Cell{Rune: 'X', Width: 1, Fg: RGB(255, 0, 0)})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Diff(front, back)
	}
}

// BenchmarkDiffLarge benchmarks diffing buffers where 50% of cells differ.
// This stresses the diff algorithm with many changes.
func BenchmarkDiffLarge(b *testing.B) {
	front := NewBuffer(80, 24)
	back := NewBuffer(80, 24)

	// Fill front with pattern A.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			front.SetCell(x, y, Cell{Rune: rune('A' + (x+y)%26), Width: 1, Fg: RGB(100, 200, 50)})
		}
	}
	// Fill back with pattern B — every other row is different.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			if y%2 == 0 {
				back.SetCell(x, y, Cell{Rune: rune('a' + (x+y)%26), Width: 1, Fg: RGB(200, 100, 50)})
			} else {
				back.SetCell(x, y, front.GetCell(x, y))
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Diff(front, back)
	}
}

// BenchmarkBufferDrawText benchmarks drawing 100 ASCII characters.
func BenchmarkBufferDrawText(b *testing.B) {
	buf := NewBuffer(120, 1)
	text := "The quick brown fox jumps over the lazy dog 0123456789!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/~`"
	style := Style{Fg: RGB(255, 121, 198), Bg: RGB(40, 42, 54), Flags: Bold}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.DrawText(0, 0, text, style)
	}
}

// BenchmarkBufferDrawTextCJK benchmarks drawing CJK (wide) text.
// CJK characters are 2 cells wide, testing the wide-char handling path.
func BenchmarkBufferDrawTextCJK(b *testing.B) {
	buf := NewBuffer(120, 1)
	text := "你好世界！这是一个测试中文字符渲染性能的句子。"
	style := Style{Fg: RGB(255, 121, 198), Bg: RGB(40, 42, 54)}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.DrawText(0, 0, text, style)
	}
}
