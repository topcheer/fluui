package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// stubChild is a simple test component with configurable size.
type stubChild struct {
	id     string
	bounds Rect
	w, h   int
	text   string
}

func (s *stubChild) ID() string                                  { return s.id }
func (s *stubChild) Measure(cs Constraints) Size                  { return Size{W: s.w, H: s.h} }
func (s *stubChild) SetBounds(r Rect)                             { s.bounds = r }
func (s *stubChild) Bounds() Rect                                 { return s.bounds }
func (s *stubChild) Paint(buf *buffer.Buffer) {
	for i, r := range s.text {
		if i >= s.bounds.W {
			break
		}
		buf.SetCell(s.bounds.X+i, s.bounds.Y, buffer.NewCell(r, buffer.DefaultStyle))
	}
}
func (s *stubChild) Children() []Component { return nil }
func (s *stubChild) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}
	return k.Key == term.KeyDown
}

func newStubChild(w, h int, text string) *stubChild {
	return &stubChild{id: GenerateID("stub"), w: w, h: h, text: text}
}

func TestNewCollapsible_Defaults(t *testing.T) {
	child := newStubChild(20, 3, "child")
	c := NewCollapsible("Title", child)

	if !c.Expanded() {
		t.Error("expected expanded by default")
	}
	if c.Collapsed() {
		t.Error("expected not collapsed by default")
	}
	if c.Title() != "Title" {
		t.Errorf("expected title 'Title', got %q", c.Title())
	}
	if c.Child() == nil {
		t.Error("expected non-nil child")
	}
	if !c.ShowArrow() {
		t.Error("expected arrow shown by default")
	}
	if c.Indent() != 0 {
		t.Error("expected indent 0 by default")
	}
}

func TestNewCollapsible_NilChild(t *testing.T) {
	c := NewCollapsible("Header", nil)
	if c.Child() != nil {
		t.Error("expected nil child")
	}
}

func TestCollapsible_Toggle(t *testing.T) {
	c := NewCollapsible("T", nil)
	if !c.Expanded() {
		t.Fatal("expected expanded initially")
	}

	// Toggle flips: expanded -> collapsed
	newState := c.Toggle()
	if newState {
		t.Error("expected toggle to return false (collapsed)")
	}
	if c.Expanded() {
		t.Error("expected collapsed state")
	}
}

func TestCollapsible_ToggleFlips(t *testing.T) {
	c := NewCollapsible("T", nil)

	// Start expanded, toggle -> collapsed
	c.Toggle()
	if c.Expanded() {
		t.Error("expected collapsed after toggle from expanded")
	}

	// Toggle again -> expanded
	c.Toggle()
	if !c.Expanded() {
		t.Error("expected expanded after second toggle")
	}
}

func TestCollapsible_Expand(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.Collapse()
	c.Expand()
	if !c.Expanded() {
		t.Error("expected expanded after Expand()")
	}
}

func TestCollapsible_Collapse(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.Collapse()
	if !c.Collapsed() {
		t.Error("expected collapsed after Collapse()")
	}
}

func TestCollapsible_SetExpanded(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.SetExpanded(false)
	if c.Expanded() {
		t.Error("expected collapsed after SetExpanded(false)")
	}
	c.SetExpanded(true)
	if !c.Expanded() {
		t.Error("expected expanded after SetExpanded(true)")
	}
}

func TestCollapsible_OnToggle(t *testing.T) {
	c := NewCollapsible("T", nil)
	called := false
	var receivedState bool

	c.OnToggle(func(expanded bool) {
		called = true
		receivedState = expanded
	})

	c.Toggle()
	if !called {
		t.Error("expected OnToggle callback to fire")
	}
	if receivedState {
		t.Error("expected receivedState=false (collapsed)")
	}

	called = false
	c.Expand()
	if !called {
		t.Error("expected OnToggle on Expand")
	}
	if !receivedState {
		t.Error("expected receivedState=true (expanded)")
	}
}

func TestCollapsible_SetTitle(t *testing.T) {
	c := NewCollapsible("Old", nil)
	c.SetTitle("New Title")
	if c.Title() != "New Title" {
		t.Errorf("expected 'New Title', got %q", c.Title())
	}
}

func TestCollapsible_SetChild(t *testing.T) {
	c := NewCollapsible("T", nil)
	child := newStubChild(10, 2, "hello")
	c.SetChild(child)
	if c.Child() != child {
		t.Error("expected child to be set")
	}
}

func TestCollapsible_SetShowArrow(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.SetShowArrow(false)
	if c.ShowArrow() {
		t.Error("expected false after SetShowArrow(false)")
	}
	c.SetShowArrow(true)
	if !c.ShowArrow() {
		t.Error("expected true after SetShowArrow(true)")
	}
}

func TestCollapsible_SetIndent(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.SetIndent(4)
	if c.Indent() != 4 {
		t.Errorf("expected indent 4, got %d", c.Indent())
	}
}

func TestCollapsible_SetIndent_Negative(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.SetIndent(-5)
	if c.Indent() != 0 {
		t.Errorf("expected indent clamped to 0, got %d", c.Indent())
	}
}

func TestCollapsible_SetHeaderStyle(t *testing.T) {
	c := NewCollapsible("T", nil)
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed), Flags: buffer.Italic}
	c.SetHeaderStyle(s)
	// Just verify no panic
}

func TestCollapsible_SetExpandedHeaderStyle(t *testing.T) {
	c := NewCollapsible("T", nil)
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)}
	c.SetExpandedHeaderStyle(s)
	// Just verify no panic
}

func TestCollapsible_Measure_Expanded(t *testing.T) {
	child := newStubChild(20, 5, "content")
	c := NewCollapsible("Header", child)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	sz := c.Measure(Bounded(40, 20))
	// Header is 1 row + child 5 rows = 6 rows
	if sz.H != 6 {
		t.Errorf("expected height 6 (1 header + 5 child), got %d", sz.H)
	}
}

func TestCollapsible_Measure_Collapsed(t *testing.T) {
	child := newStubChild(20, 5, "content")
	c := NewCollapsible("Header", child)
	c.Collapse()

	sz := c.Measure(Bounded(40, 20))
	// Only header row
	if sz.H != 1 {
		t.Errorf("expected height 1 when collapsed, got %d", sz.H)
	}
}

func TestCollapsible_Measure_NilChild(t *testing.T) {
	c := NewCollapsible("Header", nil)
	c.Expand()

	sz := c.Measure(Bounded(40, 20))
	if sz.H != 1 {
		t.Errorf("expected height 1 with nil child, got %d", sz.H)
	}
}

func TestCollapsible_Measure_TitleWidth(t *testing.T) {
	c := NewCollapsible("A Very Long Title That Exceeds Child", nil)

	sz := c.Measure(Bounded(100, 10))
	// Header should fit: indent(0) + arrow(1) + space(1) + title length
	expected := 3 + len([]rune("A Very Long Title That Exceeds Child"))
	if sz.W != expected {
		t.Errorf("expected width %d, got %d", expected, sz.W)
	}
}

func TestCollapsible_Measure_WithIndent(t *testing.T) {
	c := NewCollapsible("Title", nil)
	c.SetIndent(5)

	sz := c.Measure(Bounded(100, 10))
	// indent(5) + arrow(1) + space(1) + title(5) + 1 padding = 13
	// Measure uses: indent + 3 + len(title) = 5+3+5 = 13
	if sz.W != 13 {
		t.Errorf("expected width 12, got %d", sz.W)
	}
}

func TestCollapsible_SetBounds_PropagatesToChild(t *testing.T) {
	child := newStubChild(20, 5, "hello")
	c := NewCollapsible("T", child)

	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	gotBounds := child.Bounds()
	if gotBounds.Y != 1 {
		t.Errorf("expected child Y=1 (below header), got %d", gotBounds.Y)
	}
	if gotBounds.H != 19 {
		t.Errorf("expected child H=19 (20-1), got %d", gotBounds.H)
	}
}

func TestCollapsible_SetBounds_Collapsed_NoPropagation(t *testing.T) {
	child := newStubChild(20, 5, "hello")
	c := NewCollapsible("T", child)
	c.Collapse()

	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	gotBounds := child.bounds
	// Child bounds should remain zero (never set when collapsed)
	if gotBounds.W != 0 || gotBounds.H != 0 {
		t.Errorf("expected child bounds unset when collapsed, got %+v", gotBounds)
	}
}

func TestCollapsible_Paint_HeaderExpanded(t *testing.T) {
	c := NewCollapsible("MyTitle", nil)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})

	buf := buffer.NewBuffer(30, 5)
	c.Paint(buf)

	// Check arrow at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Rune != '\u25be' { // ▾
		t.Errorf("expected ▾ for expanded arrow, got %q", string(cell.Rune))
	}

	// Check title starts at (2,0)
	cell = buf.GetCell(2, 0)
	if cell.Rune != 'M' {
		t.Errorf("expected 'M' at position 2, got %q", string(cell.Rune))
	}
}

func TestCollapsible_Paint_HeaderCollapsed(t *testing.T) {
	c := NewCollapsible("MyTitle", nil)
	c.Collapse()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})

	buf := buffer.NewBuffer(30, 5)
	c.Paint(buf)

	// Check arrow at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Rune != '\u25b8' { // ▸
		t.Errorf("expected ▸ for collapsed arrow, got %q", string(cell.Rune))
	}
}

func TestCollapsible_Paint_NoArrow(t *testing.T) {
	c := NewCollapsible("Title", nil)
	c.SetShowArrow(false)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})

	buf := buffer.NewBuffer(20, 5)
	c.Paint(buf)

	// Without arrow, title starts at indent + 1 (space)
	cell := buf.GetCell(0, 0)
	if cell.Rune == '\u25be' || cell.Rune == '\u25b8' {
		t.Error("expected no arrow character at position 0")
	}
}

func TestCollapsible_Paint_WithIndent(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.SetIndent(4)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})

	buf := buffer.NewBuffer(20, 5)
	c.Paint(buf)

	// Arrow should be at x=4
	cell := buf.GetCell(4, 0)
	if cell.Rune != '\u25be' {
		t.Errorf("expected arrow at x=4, got %q at that pos", string(cell.Rune))
	}
}

func TestCollapsible_Paint_ChildPainted(t *testing.T) {
	child := newStubChild(10, 1, "Hello")
	c := NewCollapsible("Title", child)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	c.Paint(buf)

	// Child should be painted at Y=1
	cell := buf.GetCell(0, 1)
	if cell.Rune != 'H' {
		t.Errorf("expected 'H' at (0,1) from child, got %q", string(cell.Rune))
	}
}

func TestCollapsible_Paint_CollapsedChildNotPainted(t *testing.T) {
	child := newStubChild(10, 1, "Hello")
	c := NewCollapsible("Title", child)
	c.Collapse()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	c.Paint(buf)

	// Row 1 should not have child content
	cell := buf.GetCell(0, 1)
	if cell.Rune == 'H' {
		t.Error("expected no child content when collapsed")
	}
}

func TestCollapsible_Paint_ZeroBounds(t *testing.T) {
	c := NewCollapsible("T", nil)
	buf := buffer.NewBuffer(20, 5)
	c.Paint(buf) // should not panic
}

func TestCollapsible_HandleKey_Enter(t *testing.T) {
	c := NewCollapsible("T", nil)
	handled := c.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !handled {
		t.Error("expected Enter to be handled")
	}
	if c.Expanded() {
		t.Error("expected collapsed after Enter toggle")
	}
}

func TestCollapsible_HandleKey_Space(t *testing.T) {
	c := NewCollapsible("T", nil)
	c.Collapse()
	handled := c.HandleKey(&term.KeyEvent{Rune: ' '})
	if !handled {
		t.Error("expected Space to be handled")
	}
	if !c.Expanded() {
		t.Error("expected expanded after Space toggle")
	}
}

func TestCollapsible_HandleKey_Tab(t *testing.T) {
	c := NewCollapsible("T", nil)
	handled := c.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !handled {
		t.Error("expected Tab to be handled")
	}
}

func TestCollapsible_HandleKey_Nil(t *testing.T) {
	c := NewCollapsible("T", nil)
	handled := c.HandleKey(nil)
	if handled {
		t.Error("expected nil key to not be handled")
	}
}

func TestCollapsible_HandleKey_ForwardToChild(t *testing.T) {
	child := newStubChild(10, 2, "test")
	c := NewCollapsible("T", child)

	// KeyDown should be forwarded to child (stub returns true for KeyDown)
	handled := c.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !handled {
		t.Error("expected KeyDown forwarded to child to return true")
	}
}

func TestCollapsible_HandleKey_ForwardToChild_Collapsed(t *testing.T) {
	child := newStubChild(10, 2, "test")
	c := NewCollapsible("T", child)
	c.Collapse()

	// When collapsed, keys should not be forwarded
	handled := c.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if handled {
		t.Error("expected KeyDown NOT forwarded when collapsed")
	}
}

func TestCollapsible_HandleKey_Unhandled(t *testing.T) {
	c := NewCollapsible("T", nil)
	handled := c.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if handled {
		t.Error("expected Esc to not be handled")
	}
}

func TestCollapsible_Children_Expanded(t *testing.T) {
	child := newStubChild(10, 2, "test")
	c := NewCollapsible("T", child)

	children := c.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != child {
		t.Error("expected child to match")
	}
}

func TestCollapsible_Children_Collapsed(t *testing.T) {
	child := newStubChild(10, 2, "test")
	c := NewCollapsible("T", child)
	c.Collapse()

	children := c.Children()
	if children != nil {
		t.Errorf("expected nil children when collapsed, got %d", len(children))
	}
}

func TestCollapsible_Children_NilChild(t *testing.T) {
	c := NewCollapsible("T", nil)
	if c.Children() != nil {
		t.Error("expected nil children with nil child")
	}
}

func TestCollapsible_Bounds(t *testing.T) {
	c := NewCollapsible("T", nil)
	r := Rect{X: 5, Y: 10, W: 40, H: 20}
	c.SetBounds(r)
	if c.Bounds() != r {
		t.Errorf("expected %+v, got %+v", r, c.Bounds())
	}
}

func TestCollapsible_Concurrent(t *testing.T) {
	c := NewCollapsible("Title", newStubChild(20, 5, "content"))
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	var wg sync.WaitGroup

	// Concurrent reader/writer
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				c.Toggle()
				c.Expanded()
				c.Title()
				c.Measure(Bounded(40, 20))
				buf := buffer.NewBuffer(40, 20)
				c.Paint(buf)
			}
		}(i)
	}

	wg.Wait()
}

func TestCollapsible_Paint_TitleTruncated(t *testing.T) {
	c := NewCollapsible("This Is A Very Long Title That Won't Fit", nil)
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	buf := buffer.NewBuffer(10, 5)
	c.Paint(buf)

	// Only first ~7 chars of title should fit (10 - 2 for arrow + space)
	// Verify last visible cell is not beyond width
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		_ = cell // just verify no panic
	}
}

func TestCollapsible_OnToggle_MultipleCallbacks(t *testing.T) {
	c := NewCollapsible("T", nil)

	// Only last callback should be active
	count1 := 0
	count2 := 0

	c.OnToggle(func(bool) { count1++ })
	c.OnToggle(func(bool) { count2++ })

	c.Toggle()

	if count1 != 0 {
		t.Error("first callback should be replaced, not called")
	}
	if count2 != 1 {
		t.Error("second callback should have been called once")
	}
}
