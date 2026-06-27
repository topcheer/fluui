package hit

import "testing"

func TestRegionTreeAdd(t *testing.T) {
	tree := NewRegionTree()

	if tree.Len() != 0 {
		t.Fatalf("Len() = %d, want 0 on new tree", tree.Len())
	}

	tree.Add(Region{
		ID:     "btn-ok",
		Bounds: Rect{X: 0, Y: 0, W: 10, H: 3},
		Action: Action{Type: ActionToggle},
	})
	tree.Add(Region{
		ID:     "link-help",
		Bounds: Rect{X: 5, Y: 0, W: 4, H: 3},
		Action: Action{Type: ActionOpenURL, URL: "https://example.com"},
	})

	if tree.Len() != 2 {
		t.Errorf("Len() = %d, want 2 after two Add calls", tree.Len())
	}
}

func TestRegionTreeHit(t *testing.T) {
	tree := NewRegionTree()
	tree.Add(Region{
		ID:     "bottom",
		Bounds: Rect{X: 0, Y: 0, W: 10, H: 5},
	})
	tree.Add(Region{
		ID:     "top",
		Bounds: Rect{X: 3, Y: 1, W: 4, H: 2}, // overlaps bottom
	})

	t.Run("hit inside single region", func(t *testing.T) {
		r := tree.Hit(1, 1)
		if r == nil {
			t.Fatal("Hit(1,1) = nil, want bottom region")
		}
		if r.ID != "bottom" {
			t.Errorf("Hit(1,1).ID = %q, want %q", r.ID, "bottom")
		}
	})

	t.Run("hit overlapping returns top-most", func(t *testing.T) {
		// (4,2) is inside both bottom and top; top was added later.
		r := tree.Hit(4, 2)
		if r == nil {
			t.Fatal("Hit(4,2) = nil, want top region")
		}
		if r.ID != "top" {
			t.Errorf("Hit(4,2).ID = %q, want %q (last-added wins)", r.ID, "top")
		}
	})

	t.Run("miss returns nil", func(t *testing.T) {
		r := tree.Hit(100, 100)
		if r != nil {
			t.Errorf("Hit(100,100) = %v, want nil", r)
		}
	})
}

func TestRegionTreeHitEmpty(t *testing.T) {
	tree := NewRegionTree()
	if r := tree.Hit(0, 0); r != nil {
		t.Errorf("Hit on empty tree = %v, want nil", r)
	}
}

func TestRegionTreeClear(t *testing.T) {
	tree := NewRegionTree()
	tree.Add(Region{ID: "a", Bounds: Rect{X: 0, Y: 0, W: 1, H: 1}})
	tree.Add(Region{ID: "b", Bounds: Rect{X: 0, Y: 0, W: 1, H: 1}})

	if tree.Len() != 2 {
		t.Fatalf("Len() = %d, want 2 before Clear", tree.Len())
	}

	tree.Clear()

	if tree.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after Clear", tree.Len())
	}
	// Hit must return nil after clearing.
	if r := tree.Hit(0, 0); r != nil {
		t.Errorf("Hit after Clear = %v, want nil", r)
	}
}

func TestRegionTreeClearReuseBackingArray(t *testing.T) {
	// Clear should reset length but keep the backing array for reuse,
	// avoiding per-frame allocation churn.
	tree := NewRegionTree()
	tree.Add(Region{ID: "a", Bounds: Rect{X: 0, Y: 0, W: 1, H: 1}})
	capBefore := cap(tree.regions)

	tree.Clear()

	if cap(tree.regions) != capBefore {
		t.Errorf("backing array capacity changed after Clear: got %d, want %d", cap(tree.regions), capBefore)
	}

	// Re-add should reuse the capacity.
	tree.Add(Region{ID: "b", Bounds: Rect{X: 0, Y: 0, W: 1, H: 1}})
	if cap(tree.regions) != capBefore {
		t.Errorf("backing array capacity changed after re-Add: got %d, want %d", cap(tree.regions), capBefore)
	}
}

func TestRegionTreeQuery(t *testing.T) {
	tree := NewRegionTree()
	tree.Add(Region{ID: "a", Bounds: Rect{X: 0, Y: 0, W: 3, H: 3}})   // [0,3)x[0,3)
	tree.Add(Region{ID: "b", Bounds: Rect{X: 5, Y: 5, W: 2, H: 2}})   // [5,7)x[5,7)
	tree.Add(Region{ID: "c", Bounds: Rect{X: 10, Y: 10, W: 1, H: 1}}) // [10,11)x[10,11)

	t.Run("query overlaps two regions", func(t *testing.T) {
		// Query rect [2,8)x[2,8) overlaps a and b but not c.
		q := Rect{X: 2, Y: 2, W: 6, H: 6}
		got := tree.Query(q)
		if len(got) != 2 {
			t.Fatalf("Query returned %d regions, want 2", len(got))
		}
		// Results are in insertion order.
		if got[0].ID != "a" || got[1].ID != "b" {
			t.Errorf("Query IDs = [%s, %s], want [a, b]", got[0].ID, got[1].ID)
		}
	})

	t.Run("query overlaps all three", func(t *testing.T) {
		q := Rect{X: 0, Y: 0, W: 20, H: 20}
		got := tree.Query(q)
		if len(got) != 3 {
			t.Fatalf("Query returned %d regions, want 3", len(got))
		}
	})

	t.Run("query overlaps none", func(t *testing.T) {
		q := Rect{X: 100, Y: 100, W: 5, H: 5}
		got := tree.Query(q)
		if len(got) != 0 {
			t.Errorf("Query returned %d regions, want 0", len(got))
		}
	})

	t.Run("query edge-touching", func(t *testing.T) {
		// Edge-touching (not overlapping) should not be included.
		q := Rect{X: 3, Y: 0, W: 1, H: 3} // starts at a's right edge (exclusive)
		got := tree.Query(q)
		if len(got) != 0 {
			t.Errorf("Query edge-touching returned %d regions, want 0", len(got))
		}
	})
}

func TestRegionTreeQueryEmpty(t *testing.T) {
	tree := NewRegionTree()
	got := tree.Query(Rect{X: 0, Y: 0, W: 100, H: 100})
	if len(got) != 0 {
		t.Errorf("Query on empty tree returned %d regions, want 0", len(got))
	}
}
