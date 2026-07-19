package term

import (
	"testing"
)

// P294: Terminal capability detection protocol tests

func TestP294_QueryDA1(t *testing.T) {
	if QueryDA1 != "\x1b[c" {
		t.Errorf("QueryDA1 = %q, want %q", QueryDA1, "\x1b[c")
	}
}

func TestP294_QueryDA2(t *testing.T) {
	if QueryDA2 != "\x1b[>c" {
		t.Errorf("QueryDA2 = %q, want %q", QueryDA2, "\x1b[>c")
	}
}

func TestP294_QueryXTVersion(t *testing.T) {
	if QueryXTVersion != "\x1b[>q" {
		t.Errorf("QueryXTVersion = %q, want %q", QueryXTVersion, "\x1b[>q")
	}
}

func TestP294_QueryXTGetTCAP(t *testing.T) {
	got := QueryXTGetTCAP("544e")
	want := "\x1b[+q544e"
	if got != want {
		t.Errorf("QueryXTGetTCAP = %q, want %q", got, want)
	}
}

func TestP294_EncodeTCapName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TN", "544e"},
		{"Co", "436f"},
		{"setrgbf", "73657472676266"},
		{"", ""},
	}
	for _, tt := range tests {
		got := EncodeTCapName(tt.input)
		if got != tt.want {
			t.Errorf("EncodeTCapName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestP294_ParseDA1Response(t *testing.T) {
	tests := []struct {
		input string
		attrs []int
		ok    bool
	}{
		{"\x1b[?62;1;2;4;6;9;15;16;29c", []int{62, 1, 2, 4, 6, 9, 15, 16, 29}, true},
		{"\x1b[?62c", []int{62}, true},
		{"\x1b[?1;2c", []int{1, 2}, true},
		// failures
		{"", nil, false},
		{"short", nil, false},
		{"\x1b[?c", nil, false},           // empty body
		{"\x1b[?62;1x", nil, false},       // wrong terminator
		{"\x1b[62c", nil, false},          // missing '?'
		{"\x1b[?62;1c", []int{62, 1}, true},
	}
	for _, tt := range tests {
		attrs, ok := ParseDA1Response(tt.input)
		if ok != tt.ok {
			t.Errorf("ParseDA1Response(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if ok && !intSliceEqual(attrs, tt.attrs) {
			t.Errorf("ParseDA1Response(%q) attrs = %v, want %v", tt.input, attrs, tt.attrs)
		}
	}
}

func TestP294_ParseDA2Response(t *testing.T) {
	tests := []struct {
		input string
		resp  DA2Response
		ok    bool
	}{
		{"\x1b[>1;276;0c", DA2Response{TerminalType: 1, Version: 276, ROMCartridges: 0}, true},
		{"\x1b[>0;100;0c", DA2Response{TerminalType: 0, Version: 100, ROMCartridges: 0}, true},
		{"\x1b[>24;32801;0c", DA2Response{TerminalType: 24, Version: 32801, ROMCartridges: 0}, true},
		{"\x1b[>62c", DA2Response{TerminalType: 62, Version: 0, ROMCartridges: 0}, true},
		// failures
		{"", DA2Response{}, false},
		{"short", DA2Response{}, false},
		{"\x1b[>x", DA2Response{}, false},       // wrong terminator
		{"\x1b[?62c", DA2Response{}, false},     // wrong prefix
	}
	for _, tt := range tests {
		resp, ok := ParseDA2Response(tt.input)
		if ok != tt.ok {
			t.Errorf("ParseDA2Response(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if ok && resp != tt.resp {
			t.Errorf("ParseDA2Response(%q) = %+v, want %+v", tt.input, resp, tt.resp)
		}
	}
}

func TestP294_ParseXTVersionResponse(t *testing.T) {
	tests := []struct {
		input string
		resp  XTVersionResponse
		ok    bool
	}{
		// name(version) with ST terminator
		{"\x1b[>xterm(372)\x1b\\", XTVersionResponse{Name: "xterm", Version: "372"}, true},
		// name(version) with BEL terminator
		{"\x1b[>tmux(3.3a)\x07", XTVersionResponse{Name: "tmux", Version: "3.3a"}, true},
		// legacy CSI > name;version q format
		{"\x1b[>contour;0.3\x07", XTVersionResponse{Name: "contour", Version: "0.3"}, true},
		// name only
		{"\x1b[>wezterm\x07", XTVersionResponse{Name: "wezterm", Version: ""}, true},
		// failures
		{"", XTVersionResponse{}, false},
		{"short", XTVersionResponse{}, false},
		// no terminator at all
		{"\x1b[>xterm(372)", XTVersionResponse{}, false},
	}
	for _, tt := range tests {
		resp, ok := ParseXTVersionResponse(tt.input)
		if ok != tt.ok {
			t.Errorf("ParseXTVersionResponse(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if ok && resp != tt.resp {
			t.Errorf("ParseXTVersionResponse(%q) = %+v, want %+v", tt.input, resp, tt.resp)
		}
	}
}

func TestP294_ParseXTGetTCapResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hexName  string
		hexValue string
		ok       bool
	}{
		{"success_ST", "\x1bP1+r544e=787465726d\x1b\\", "544e", "787465726d", true},
		{"success_BEL", "\x1bP1+r436f=3136\x07", "436f", "3136", true},
		{"success_no_value", "\x1bP1+r544e\x1b\\", "544e", "", true},
		{"failure", "\x1bP0+r\x1b\\", "", "", false},
		// bad inputs
		{"empty", "", "", "", false},
		{"no_esc", "garbage", "", "", false},
		{"no_terminator", "\x1bP1+r544e=787465726d", "", "", false},
		{"bad_prefix", "\x1bP2+r544e\x1b\\", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexName, hexValue, ok := ParseXTGetTCapResponse(tt.input)
			if ok != tt.ok {
				t.Errorf("ParseXTGetTCapResponse(%q) ok = %v, want %v", tt.input, ok, tt.ok)
				return
			}
			if ok && (hexName != tt.hexName || hexValue != tt.hexValue) {
				t.Errorf("ParseXTGetTCapResponse(%q) = (%q,%q), want (%q,%q)",
					tt.input, hexName, hexValue, tt.hexName, tt.hexValue)
			}
		})
	}
}

func TestP294_ProtocolFormatsAreValidCSI(t *testing.T) {
	// All new query constants should be valid escape sequences
	seqs := map[string]string{
		"QueryDA1":        QueryDA1,
		"QueryDA2":        QueryDA2,
		"QueryXTVersion":  QueryXTVersion,
	}
	for name, seq := range seqs {
		if len(seq) < 3 {
			t.Errorf("%s too short: %q", name, seq)
		}
		if seq[0] != 0x1b || seq[1] != '[' {
			t.Errorf("%s doesn't start with CSI: %q", name, seq)
		}
	}
}

func TestP294_EncodeTCapNameRoundTrip(t *testing.T) {
	// Encoding "TN" should produce valid hex that can be used in queries
	encoded := EncodeTCapName("TN")
	q := QueryXTGetTCAP(encoded)
	// Should contain the hex after CSI + q
	if len(q) < 6 {
		t.Errorf("query too short: %q", q)
	}
	prefix := q[:4] // ESC [ + q
	if prefix != "\x1b[+q" {
		t.Errorf("query prefix wrong: %q", prefix)
	}
}

// helpers
func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
