package term

import (
	"strings"
	"testing"
)

// P293: Test new protocol constants and functions

func TestP293_SaveRestoreCursor(t *testing.T) {
	if SaveCursor != "\x1b7" {
		t.Errorf("SaveCursor = %q, want %q", SaveCursor, "\x1b7")
	}
	if RestoreCursor != "\x1b8" {
		t.Errorf("RestoreCursor = %q, want %q", RestoreCursor, "\x1b8")
	}
	if SaveCursorANSI != "\x1b[s" {
		t.Errorf("SaveCursorANSI = %q, want %q", SaveCursorANSI, "\x1b[s")
	}
	if RestoreCursorANSI != "\x1b[u" {
		t.Errorf("RestoreCursorANSI = %q, want %q", RestoreCursorANSI, "\x1b[u")
	}
}

func TestP293_SetScrollRegion(t *testing.T) {
	tests := []struct {
		top, bottom int
		want        string
	}{
		{0, 24, "\x1b[0;24r"},
		{1, 24, "\x1b[1;24r"},
		{5, 20, "\x1b[5;20r"},
		{-1, -1, "\x1b[0;0r"},
		{0, 0, "\x1b[0;0r"},
	}
	for _, tt := range tests {
		got := SetScrollRegion(tt.top, tt.bottom)
		if got != tt.want {
			t.Errorf("SetScrollRegion(%d,%d) = %q, want %q", tt.top, tt.bottom, got, tt.want)
		}
	}
}

func TestP293_ResetScrollRegion(t *testing.T) {
	if ResetScrollRegion != "\x1b[r" {
		t.Errorf("ResetScrollRegion = %q, want %q", ResetScrollRegion, "\x1b[r")
	}
}

func TestP293_QueryCursorPosition(t *testing.T) {
	if QueryCursorPosition != "\x1b[6n" {
		t.Errorf("QueryCursorPosition = %q, want %q", QueryCursorPosition, "\x1b[6n")
	}
}

func TestP293_QueryTerminalSize(t *testing.T) {
	if QueryTerminalSize != "\x1b[14t" {
		t.Errorf("QueryTerminalSize = %q, want %q", QueryTerminalSize, "\x1b[14t")
	}
}

func TestP293_QueryCellSize(t *testing.T) {
	if QueryCellSize != "\x1b[16t" {
		t.Errorf("QueryCellSize = %q, want %q", QueryCellSize, "\x1b[16t")
	}
}

func TestP293_ParseCursorPositionResponse(t *testing.T) {
	tests := []struct {
		input   string
		row     int
		col     int
		ok      bool
	}{
		{"\x1b[10;20R", 10, 20, true},
		{"\x1b[1;1R", 1, 1, true},
		{"\x1b[0;0R", 1, 1, true}, // 0 defaults to 1
		{"\x1b[100;200R", 100, 200, true},
		{"", 0, 0, false},
		{"short", 0, 0, false},
		{"\x1b[abcR", 0, 0, false},       // no semicolon
		{"\x1b[10;abcR", 10, 1, true},    // non-numeric col defaults to 1
		{"no esc", 0, 0, false},
		{"\x1b[5;10X", 0, 0, false},      // wrong terminator
	}
	for _, tt := range tests {
		row, col, ok := ParseCursorPositionResponse(tt.input)
		if row != tt.row || col != tt.col || ok != tt.ok {
			t.Errorf("ParseCursorPositionResponse(%q) = (%d,%d,%v), want (%d,%d,%v)",
				tt.input, row, col, ok, tt.row, tt.col, tt.ok)
		}
	}
}

func TestP293_intToStr(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{255, "255"},
		{1000, "1000"},
	}
	for _, tt := range tests {
		got := intToStr(tt.input)
		if got != tt.want {
			t.Errorf("intToStr(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestP293_Bell(t *testing.T) {
	if Bell != "\x07" {
		t.Errorf("Bell = %q, want %q", Bell, "\x07")
	}
}

// Verify protocol string formats are valid CSI sequences
func TestP293_ProtocolFormatsValid(t *testing.T) {
	// All new constants and functions should produce valid escape sequences
	seqs := []string{
		SaveCursor, RestoreCursor, SaveCursorANSI, RestoreCursorANSI,
		SetScrollRegion(1, 24), ResetScrollRegion,
		QueryCursorPosition, QueryTerminalSize, QueryCellSize,
		Bell,
	}
	for i, s := range seqs {
		if len(s) == 0 {
			t.Errorf("protocol %d is empty", i)
		}
		// Bell is a single char, others start with ESC
		if i != len(seqs)-1 && !strings.HasPrefix(s, "\x1b") {
			t.Errorf("protocol %d (%q) doesn't start with ESC", i, s)
		}
	}
}
