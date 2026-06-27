package hit

// RegionTree is a flat collection of Regions used for hit testing.
//
// Regions are stored in insertion order. When multiple regions overlap at the
// same coordinate, Hit returns the one that was added last, matching the
// visual stacking order in which components paint (later paints render on top).
//
// The tree is not safe for concurrent use. All access — Add, Clear, Hit, Query
// — must happen from the same goroutine that owns the render/mouse loop.
type RegionTree struct {
	regions []Region
}

// NewRegionTree returns an empty RegionTree ready to accept regions.
func NewRegionTree() *RegionTree {
	return &RegionTree{
		regions: make([]Region, 0),
	}
}

// Add registers a region. Regions added later are considered "on top" for hit
// testing purposes.
func (t *RegionTree) Add(r Region) {
	t.regions = append(t.regions, r)
}

// Clear removes all regions. Called at the start of each render frame before
// components re-register their interactive areas.
func (t *RegionTree) Clear() {
	// Reset to zero length but keep the backing array for reuse, avoiding
	// repeated allocations across frames.
	t.regions = t.regions[:0]
}

// Hit returns the top-most region whose bounds contain (x, y), or nil if no
// region matches. Regions are searched in reverse insertion order so that the
// last-added (visually top-most) region wins.
func (t *RegionTree) Hit(x, y int) *Region {
	for i := len(t.regions) - 1; i >= 0; i-- {
		if t.regions[i].Bounds.Contains(x, y) {
			return &t.regions[i]
		}
	}
	return nil
}

// Query returns all regions whose bounds intersect r, in insertion order. This
// is useful for drag-selection or range-based interactions.
func (t *RegionTree) Query(r Rect) []Region {
	var result []Region
	for i := range t.regions {
		if intersects(t.regions[i].Bounds, r) {
			result = append(result, t.regions[i])
		}
	}
	return result
}

// Len returns the number of registered regions.
func (t *RegionTree) Len() int {
	return len(t.regions)
}

// intersects reports whether two rectangles share at least one cell.
func intersects(a, b Rect) bool {
	if a.W <= 0 || a.H <= 0 || b.W <= 0 || b.H <= 0 {
		return false
	}
	return a.X < b.X+b.W && a.X+a.W > b.X &&
		a.Y < b.Y+b.H && a.Y+a.H > b.Y
}
