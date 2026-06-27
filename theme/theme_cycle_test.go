package theme

import "testing"

func TestCycle(t *testing.T) {
	// Reset to known state
	SetByIndex(0)
	if Get().Name != "Dracula" {
		t.Fatalf("expected Dracula at index 0, got %s", Get().Name)
	}

	// Cycle forward through all themes
	expected := []string{"Nord", "Gruvbox", "SolarizedDark", "TokyoNight", "Dracula"}
	for _, name := range expected {
		next := Cycle()
		if next.Name != name {
			t.Errorf("expected %s, got %s", name, next.Name)
		}
	}
}

func TestCycleBack(t *testing.T) {
	SetByIndex(0)
	// Backward from Dracula should wrap to TokyoNight
	prev := CycleBack()
	if prev.Name != "TokyoNight" {
		t.Errorf("expected TokyoNight, got %s", prev.Name)
	}
}

func TestSetByIndex(t *testing.T) {
	SetByIndex(2)
	if Get().Name != "Gruvbox" {
		t.Errorf("expected Gruvbox at index 2, got %s", Get().Name)
	}
	if CurrentIndex() != 2 {
		t.Errorf("expected index 2, got %d", CurrentIndex())
	}

	// Out of range should be no-op
	SetByIndex(99)
	if CurrentIndex() != 2 {
		t.Errorf("index should not change for out-of-range, got %d", CurrentIndex())
	}
}

func TestSetActiveUpdatesIndex(t *testing.T) {
	SetByIndex(0)
	nord := Nord()
	SetActive(nord)
	if CurrentIndex() != 1 {
		t.Errorf("expected index 1 for Nord, got %d", CurrentIndex())
	}
}
