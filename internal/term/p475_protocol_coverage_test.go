package term

import (
	"testing"
)

func TestSixelCapability(t *testing.T) {
	got := QuerySixelCapability()
	if got != "\x1b[c" {
		t.Errorf("QuerySixelCapability = %q, want \\x1b[c", got)
	}
}

func TestSixelStartEnd(t *testing.T) {
	// Without palette
	got := SixelStart(0)
	if got != "\x1bPq" {
		t.Errorf("SixelStart(0) = %q, want \\x1bPq", got)
	}
	// With palette
	got = SixelStart(1)
	if got != "\x1bP1q" {
		t.Errorf("SixelStart(1) = %q, want \\x1bP1q", got)
	}
	// End
	got = SixelEnd()
	if got != "\x1b\\" {
		t.Errorf("SixelEnd = %q, want \\x1b\\\\", got)
	}
}

func TestRequestDECMode(t *testing.T) {
	got := RequestDECMode(2026)
	if got != "\x1b[?2026$p" {
		t.Errorf("RequestDECMode(2026) = %q", got)
	}
}

func TestRequestANSIMode(t *testing.T) {
	got := RequestANSIMode(4)
	if got != "\x1b[4$p" {
		t.Errorf("RequestANSIMode(4) = %q", got)
	}
}

func TestOSC52Copy(t *testing.T) {
	got := OSC52Copy("aGVsbG8=")
	want := "\x1b]52;c;aGVsbG8=\x07"
	if got != want {
		t.Errorf("OSC52Copy = %q, want %q", got, want)
	}
}

func TestOSC52CopySelection(t *testing.T) {
	got := OSC52CopySelection("dGVzdA==")
	want := "\x1b]52;p;dGVzdA==\x07"
	if got != want {
		t.Errorf("OSC52CopySelection = %q, want %q", got, want)
	}
}

func TestOSC52Query(t *testing.T) {
	got := OSC52Query()
	want := "\x1b]52;c;?\x07"
	if got != want {
		t.Errorf("OSC52Query = %q, want %q", got, want)
	}
}

func TestSelectiveEraseConstants(t *testing.T) {
	if SelectiveEraseDisplay != "\x1b[?2J" {
		t.Errorf("SelectiveEraseDisplay = %q", SelectiveEraseDisplay)
	}
	if SelectiveEraseToEnd != "\x1b[?0J" {
		t.Errorf("SelectiveEraseToEnd = %q", SelectiveEraseToEnd)
	}
	if SelectiveEraseToStart != "\x1b[?1J" {
		t.Errorf("SelectiveEraseToStart = %q", SelectiveEraseToStart)
	}
}

func TestSixelStartLargePalette(t *testing.T) {
	got := SixelStart(256)
	want := "\x1bP256q"
	if got != want {
		t.Errorf("SixelStart(256) = %q, want %q", got, want)
	}
}
