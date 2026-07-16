package snapshot

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P239: cover SerializeDetailed multi-row + strDiff mismatch paths

func TestSerializeDetailed_MultiRow_P239(t *testing.T) {
	buf := buffer.NewBuffer(3, 3)
	buf.SetCell(0, 0, buffer.Cell{Rune: 'A', Width: 1})
	buf.SetCell(0, 1, buffer.Cell{Rune: 'B', Width: 1})
	buf.SetCell(0, 2, buffer.Cell{Rune: 'C', Width: 1})
	result := SerializeDetailed(buf)
	// Should contain newlines for multi-row
	if result == "" {
		t.Error("SerializeDetailed should produce non-empty output")
	}
}

func TestStrDiff_ExtraActualLines_P239(t *testing.T) {
	// When actual has more lines than expected → "+ " prefix
	buf1 := buffer.NewBuffer(2, 1) // expected: 1 row
	buf2 := buffer.NewBuffer(2, 2) // actual: 2 rows
	buf2.SetCell(0, 1, buffer.Cell{Rune: 'X', Width: 1})
	diff := strDiff(Serialize(buf1), Serialize(buf2))
	if diff == "" {
		t.Error("strDiff should show difference")
	}
}

func TestStrDiff_ExtraExpectedLines_P239(t *testing.T) {
	// When expected has more lines than actual → "- " prefix
	buf1 := buffer.NewBuffer(2, 2) // expected: 2 rows
	buf1.SetCell(0, 1, buffer.Cell{Rune: 'X', Width: 1})
	buf2 := buffer.NewBuffer(2, 1) // actual: 1 row
	diff := strDiff(Serialize(buf1), Serialize(buf2))
	if diff == "" {
		t.Error("strDiff should show difference")
	}
}
