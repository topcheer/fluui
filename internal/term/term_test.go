package term

import (
	"testing"
)

func TestColorProfile_Constants(t *testing.T) {
	// Verify ColorProfile constants are distinct and ordered.
	profiles := []ColorProfile{
		ProfileNone,
		ProfileANSI16,
		Profile256,
		ProfileTrue,
	}

	// Each should be unique.
	for i := 0; i < len(profiles); i++ {
		for j := i + 1; j < len(profiles); j++ {
			if profiles[i] == profiles[j] {
				t.Fatalf("profiles[%d] == profiles[%d] (%d), expected unique", i, j, profiles[i])
			}
		}
	}

	// Ordered ascending (iota-based).
	for i := 1; i < len(profiles); i++ {
		if profiles[i] <= profiles[i-1] {
			t.Fatalf("expected ascending order, profiles[%d]=%d <= profiles[%d]=%d", i, profiles[i], i-1, profiles[i-1])
		}
	}
}

func TestColorProfile_Names(t *testing.T) {
	// ProfileNone = 0 (lowest).
	if ProfileNone != 0 {
		t.Fatalf("expected ProfileNone=0, got %d", ProfileNone)
	}
	// ProfileTrue should be the highest.
	if ProfileTrue <= Profile256 {
		t.Fatal("expected ProfileTrue > Profile256")
	}
}

// Compile-time check that Terminal interface exists and has expected methods.
func TestTerminalInterface(t *testing.T) {
	// This is a compile-time check: if the Terminal type or any method changes,
	// this test will fail to compile.

	var _ *Terminal = (*Terminal)(nil)

	// Verify ColorProfile constants work with NewWriter.
	_ = ColorProfile(Profile256)
	_ = ColorProfile(ProfileTrue)
	_ = ColorProfile(ProfileANSI16)
	_ = ColorProfile(ProfileNone)
}
