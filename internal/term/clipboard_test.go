package term

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCopyOSC52(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"ascii", "Hello World"},
		{"unicode", "Hello 世界"},
		{"multiline", "line1\nline2\nline3"},
		{"special", ";|\\"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seq := CopyOSC52(tc.input)

			// Must start with ESC ] 52; c;
			prefix := "\x1b]52;c;"
			if !strings.HasPrefix(seq, prefix) {
				t.Errorf("prefix: got %q, want prefix %q", seq[:min(7, len(seq))], prefix)
			}

			// Must end with ST (ESC \)
			if !strings.HasSuffix(seq, "\x1b\\") {
				t.Error("should end with ST (ESC \\)")
			}

			// Round-trip: parse should recover original text
			got, ok := ParseOSC52Response(seq)
			if !ok {
				t.Fatal("ParseOSC52Response returned false")
			}
			if got != tc.input {
				t.Errorf("round-trip: got %q, want %q", got, tc.input)
			}
		})
	}
}

func TestCopyOSC52Source(t *testing.T) {
	seq := CopyOSC52Source("text", ClipboardPrimary)
	if !strings.HasPrefix(seq, "\x1b]52;p;") {
		t.Errorf("primary prefix: got %q", seq[:min(7, len(seq))])
	}
}

func TestPasteQuery(t *testing.T) {
	seq := PasteQuery()

	if !strings.HasPrefix(seq, "\x1b]52;c;") {
		t.Errorf("prefix: got %q", seq[:min(7, len(seq))])
	}
	if !strings.Contains(seq, "?") {
		t.Error("should contain '?' query marker")
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Error("should end with ST (ESC \\)")
	}
}

func TestParseOSC52Response(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOk bool
	}{
		{
			name:   "simple ST terminated",
			input:  "\x1b]52;c;" + b64("Hello") + "\x1b\\",
			want:   "Hello",
			wantOk: true,
		},
		{
			name:   "BEL terminated",
			input:  "\x1b]52;c;" + b64("Hello") + "\x07",
			want:   "Hello",
			wantOk: true,
		},
		{
			name:   "unicode",
			input:  "\x1b]52;c;" + b64("你好") + "\x1b\\",
			want:   "你好",
			wantOk: true,
		},
		{
			name:   "empty clipboard",
			input:  "\x1b]52;c;\x1b\\",
			want:   "",
			wantOk: true,
		},
		{
			name:   "query echo",
			input:  "\x1b]52;c;?\x1b\\",
			want:   "",
			wantOk: true,
		},
		{
			name:   "not OSC52",
			input:  "garbage",
			wantOk: false,
		},
		{
			name:   "missing separator",
			input:  "\x1b]52;garbage\x1b\\",
			wantOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseOSC52Response(tc.input)
			if ok != tc.wantOk {
				t.Fatalf("ok: got %v, want %v", ok, tc.wantOk)
			}
			if ok && got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsOSC52Response(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"\x1b]52;c;" + b64("Hi") + "\x1b\\", true},
		{"\x1b]52;c;?\x1b\\", true},
		{"\x1b]52;p;" + b64("Hi") + "\x1b\\", true},
		{"garbage", false},
		{"", false},
		{"\x1b]52", false},
	}

	for _, tc := range tests {
		got := IsOSC52Response(tc.input)
		if got != tc.want {
			t.Errorf("IsOSC52Response(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestOSC52RoundTripLong(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 4096; i++ {
		sb.WriteByte(byte('A' + (i % 26)))
	}
	long := sb.String()

	seq := CopyOSC52(long)
	got, ok := ParseOSC52Response(seq)
	if !ok {
		t.Fatal("round-trip: ParseOSC52Response returned false")
	}
	if got != long {
		t.Error("round-trip mismatch for long string")
	}
}

// b64 is a test helper that base64-encodes a string.
func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
