package component

import "testing"

func TestBaseComponent(t *testing.T) {
	var c Component = &BaseComponent{id: "test"}

	if c.ID() != "test" {
		t.Errorf(`ID: got %q, want "test"`, c.ID())
	}

	r := Rect{X: 1, Y: 2, W: 10, H: 5}
	c.SetBounds(r)
	if c.Bounds() != r {
		t.Errorf("Bounds: got %v, want %v", c.Bounds(), r)
	}

	if c.Children() != nil {
		t.Error("BaseComponent should have no children")
	}
}

func TestConstraints(t *testing.T) {
	u := Unbounded()
	if u.MaxWidth != 0 || u.MaxHeight != 0 {
		t.Error("Unbounded should have 0 max")
	}

	f := Fixed(10, 5)
	if f.MinWidth != 10 || f.MaxWidth != 10 {
		t.Errorf("Fixed width: min=%d max=%d, want 10/10", f.MinWidth, f.MaxWidth)
	}

	b := Bounded(80, 24)
	if b.MaxWidth != 80 || b.MaxHeight != 24 {
		t.Errorf("Bounded: got %v, want 80x24", b)
	}
}

func TestSizeAndRect(t *testing.T) {
	s := Size{W: 10, H: 5}
	if s.W != 10 || s.H != 5 {
		t.Error("Size fields incorrect")
	}

	r := Rect{X: 0, Y: 0, W: 80, H: 24}
	if r.W != 80 || r.H != 24 {
		t.Error("Rect fields incorrect")
	}
}
