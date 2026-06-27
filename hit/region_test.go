package hit

import "testing"

func TestRectContains(t *testing.T) {
	r := Rect{X: 2, Y: 3, W: 4, H: 5} // spans x=[2,6) y=[3,8)

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		// Interior points
		{"interior", 3, 4, true},
		{"another interior", 5, 7, true},

		// Top-left corner (inclusive)
		{"top-left corner", 2, 3, true},

		// Right edge is exclusive (X+W == 6)
		{"right edge exclusive", 6, 4, false},
		// Bottom edge is exclusive (Y+H == 8)
		{"bottom edge exclusive", 3, 8, false},

		// Outside points
		{"left of", 1, 4, false},
		{"above", 3, 2, false},
		{"right of", 7, 4, false},
		{"below", 3, 9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Contains(tt.x, tt.y); got != tt.want {
				t.Errorf("Rect%v.Contains(%d, %d) = %v, want %v", r, tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestRectContainsZeroArea(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
	}{
		{"zero width", Rect{X: 0, Y: 0, W: 0, H: 5}},
		{"zero height", Rect{X: 0, Y: 0, W: 5, H: 0}},
		{"both zero", Rect{X: 0, Y: 0, W: 0, H: 0}},
		{"negative width", Rect{X: 0, Y: 0, W: -3, H: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r.Contains(0, 0) {
				t.Errorf("Rect%v.Contains(0, 0) = true, want false (zero-area rect)", tt.r)
			}
		})
	}
}

func TestCursorStyleValues(t *testing.T) {
	// Ensure iota values are distinct and ordered as documented.
	if CursorDefault != 0 {
		t.Errorf("CursorDefault = %d, want 0", CursorDefault)
	}
	if CursorPointer <= CursorDefault {
		t.Errorf("CursorPointer (%d) should be > CursorDefault (%d)", CursorPointer, CursorDefault)
	}
	if CursorText <= CursorPointer {
		t.Errorf("CursorText (%d) should be > CursorPointer (%d)", CursorText, CursorPointer)
	}
}

func TestActionTypeValues(t *testing.T) {
	// Ensure iota values are distinct and ordered as documented.
	if ActionToggle != 0 {
		t.Errorf("ActionToggle = %d, want 0", ActionToggle)
	}
	if ActionOpenURL <= ActionToggle {
		t.Errorf("ActionOpenURL (%d) should be > ActionToggle (%d)", ActionOpenURL, ActionToggle)
	}
	if ActionCopy <= ActionOpenURL {
		t.Errorf("ActionCopy (%d) should be > ActionOpenURL (%d)", ActionCopy, ActionOpenURL)
	}
	if ActionCustom <= ActionCopy {
		t.Errorf("ActionCustom (%d) should be > ActionCopy (%d)", ActionCustom, ActionCopy)
	}
}

func TestActionCustomCallback(t *testing.T) {
	called := false
	a := Action{Type: ActionCustom, Fn: func() { called = true }}
	a.Fn()
	if !called {
		t.Error("Action.Fn was not invoked")
	}
}
