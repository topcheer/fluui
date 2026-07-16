package viewport

import "testing"

// P234: cover SetYOffset(), ScrollIndicatorStyle() option — at 0%

func TestViewport_SetYOffset_P234(t *testing.T) {
	m := New(WithWidth(40), WithHeight(10))
	m.SetContent("line1\nline2\nline3\nline4\nline5")
	m.SetYOffset(2)
	// no-op in fluui — should not panic
}

func TestViewport_ScrollIndicatorStyleOption_P234(t *testing.T) {
	m := New(ScrollIndicatorStyle("│"))
	// Verify the option was applied
	if m.scrollIndicatorChar != "│" {
		t.Errorf("scrollIndicatorChar = %q, want '│'", m.scrollIndicatorChar)
	}
}

func TestViewport_ScrollIndicatorStyleMethod_P234(t *testing.T) {
	// Need content > height and not at bottom for scroll indicator
	m := New(WithWidth(40), WithHeight(3))
	content := "l1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nl10"
	m.SetContent(content)
	m.SetYOffset(0) // not at bottom
	// Disable autoFollow so AtBottom() check doesn't short-circuit
	m.autoFollow = false
	s := m.ScrollIndicatorStyle()
	// Should show a scroll indicator since content > height
	if s == "" {
		t.Error("ScrollIndicatorStyle() should return non-empty when not at bottom")
	}
}
