package snapshot

import (
	"strings"
	"testing"
)

func TestStrDiff_ActualLonger_P268(t *testing.T) {
	diff := strDiff("a\nb", "a\nb\nc")
	if !strings.Contains(diff, "+ c") {
		t.Errorf("expected '+ c' in diff, got: %s", diff)
	}
}

func TestStrDiff_ExpectedLonger_P268(t *testing.T) {
	diff := strDiff("a\nb\nc", "a\nb")
	if !strings.Contains(diff, "- c") {
		t.Errorf("expected '- c' in diff, got: %s", diff)
	}
}

func TestStrDiff_Mixed_P268(t *testing.T) {
	diff := strDiff("a\nb\nc\nd", "a\nX\nc\nd\ne")
	if !strings.Contains(diff, "  a") {
		t.Error("should contain unchanged 'a'")
	}
	if !strings.Contains(diff, "! exp: b") {
		t.Error("should contain changed 'b' as ! exp: b")
	}
	if !strings.Contains(diff, "! got: X") {
		t.Error("should contain changed 'X' as ! got: X")
	}
	if !strings.Contains(diff, "+ e") {
		t.Error("should contain added 'e'")
	}
}
