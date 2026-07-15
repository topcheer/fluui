package hit

import "testing"

// P208: hit.intersects edge cases

func TestIntersects_ZeroSize_P208(t *testing.T) {
	if intersects(Rect{X: 0, Y: 0, W: 0, H: 5}, Rect{X: 0, Y: 0, W: 5, H: 5}) {
		t.Error("zero width should not intersect")
	}
	if intersects(Rect{X: 0, Y: 0, W: 5, H: 0}, Rect{X: 0, Y: 0, W: 5, H: 5}) {
		t.Error("zero height should not intersect")
	}
}

func TestIntersects_Negative_P208(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, W: -1, H: 5}
	r2 := Rect{X: 0, Y: 0, W: 5, H: 5}
	if intersects(r1, r2) {
		t.Error("negative width should not intersect")
	}
}

func TestIntersects_AdjacentNotOverlapping_P208(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, W: 5, H: 5}
	r2 := Rect{X: 5, Y: 0, W: 5, H: 5} // starts where r1 ends
	if intersects(r1, r2) {
		t.Error("adjacent rects should not intersect")
	}
}

func TestIntersects_PartialOverlap_P208(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, W: 5, H: 5}
	r2 := Rect{X: 3, Y: 3, W: 5, H: 5}
	if !intersects(r1, r2) {
		t.Error("partially overlapping rects should intersect")
	}
}

func TestIntersects_Contained_P208(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, W: 10, H: 10}
	r2 := Rect{X: 2, Y: 2, W: 3, H: 3}
	if !intersects(r1, r2) {
		t.Error("contained rect should intersect")
	}
}