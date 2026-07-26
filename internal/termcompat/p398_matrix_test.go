package termcompat

import "testing"

func TestP398_Matrix(t *testing.T) {
	if len(Matrix) != 7 {
		t.Errorf("Matrix has %d terminals, want 7", len(Matrix))
	}
	for _, term := range CompatibilityMatrix {
		if term.Name == "" {
			t.Error("terminal name should not be empty")
		}
		if len(term.Protocols) == 0 {
			t.Errorf("terminal %q has no protocols", term.Name)
		}
	}
}

func TestP398_SupportString(t *testing.T) {
	tests := []struct {
		s    Support
		want string
	}{
		{SupportYes, "YES"},
		{SupportPartial, "PARTIAL"},
		{SupportNo, "NO"},
		{SupportUnknown, "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Support(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestP398_CountSupport(t *testing.T) {
	yes, partial, no, unknown := CountSupport("OSC8")
	if yes+partial+no+unknown != 7 {
		t.Errorf("total = %d, want 7", yes+partial+no+unknown)
	}
	if yes < 5 {
		t.Errorf("OSC8 yes = %d, want >= 5", yes)
	}

	// All terminals should have entries for common protocols
	for _, proto := range []string{"OSC8", "OSC52", "Sync", "AltScreen", "DECSCUSR"} {
		y, p, n, u := CountSupport(proto)
		if y+p+n+u != 7 {
			t.Errorf("%s: total = %d", proto, y+p+n+u)
		}
	}
}

func TestP398_ITerm2HasFullSupport(t *testing.T) {
	term := CompatibilityMatrix[0]
	if term.Name != "iTerm2 (3.5+)" {
		t.Errorf("first terminal = %q, want iTerm2", term.Name)
	}
	for _, proto := range []string{"OSC8", "OSC52", "Sync", "Images", "FocusReport"} {
		if term.Protocols[proto] != SupportYes {
			t.Errorf("iTerm2 %s = %v, want YES", proto, term.Protocols[proto])
		}
	}
}
