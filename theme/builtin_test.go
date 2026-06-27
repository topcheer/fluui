package theme

import (
	"testing"
)

// Tests in this file supplement the existing theme_test.go and theme_cycle_test.go
// with additional coverage. Conflicting names (TestGet, TestSetActive, TestCycle,
// TestCycleBack, TestSetByIndex, TestDefault) are already covered by those files.

func TestBuiltinThemes(t *testing.T) {
	all := Builtin()
	if len(all) != 5 {
		t.Fatalf("Builtin() returned %d themes, want 5", len(all))
	}

	expectedNames := []string{"Dracula", "Nord", "Gruvbox", "SolarizedDark", "TokyoNight"}
	for i, want := range expectedNames {
		if all[i].Name != want {
			t.Errorf("Builtin()[%d].Name = %q, want %q", i, all[i].Name, want)
		}
	}
}

func TestBuiltinThemesAreUnique(t *testing.T) {
	all := Builtin()
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[i].Bg.Equal(all[j].Bg) {
				t.Errorf("themes %s and %s have identical Bg", all[i].Name, all[j].Name)
			}
			if all[i].Accent.Equal(all[j].Accent) {
				t.Errorf("themes %s and %s have identical Accent", all[i].Name, all[j].Name)
			}
			if all[i].Name == all[j].Name {
				t.Errorf("duplicate theme name: %s", all[i].Name)
			}
		}
	}
}

func TestSetByIndex_OutOfRange(t *testing.T) {
	original := Active
	defer func() { Active = original }()

	// Out of range should be a no-op
	SetActive(Dracula())
	SetByIndex(-1)
	if Active.Name != "Dracula" {
		t.Error("SetByIndex(-1) should not change theme")
	}
	SetByIndex(99)
	if Active.Name != "Dracula" {
		t.Error("SetByIndex(99) should not change theme")
	}
}

func TestEachBuiltinTheme(t *testing.T) {
	// Test each theme has all required fields populated
	for _, themeFn := range []func() *Theme{Dracula, Nord, Gruvbox, SolarizedDark, TokyoNight} {
		th := themeFn()
		t.Run(th.Name, func(t *testing.T) {
			if th.Name == "" {
				t.Error("Name is empty")
			}
			if th.Bg.IsDefault() {
				t.Error("Bg is default")
			}
			if th.Fg.IsDefault() {
				t.Error("Fg is default")
			}
			if th.Accent.IsDefault() {
				t.Error("Accent is default")
			}
			if th.Error.IsDefault() {
				t.Error("Error is default")
			}
			if th.Success.IsDefault() {
				t.Error("Success is default")
			}
		})
	}
}
