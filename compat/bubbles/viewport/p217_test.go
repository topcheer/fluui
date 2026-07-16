package viewport

import "testing"

// P217: viewport.View edge cases

func TestView_EmptyContent_P217(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetContent("")
	if m.View() != "" {
		t.Error("empty content should return empty")
	}
}

func TestView_ZeroHeight_P217(t *testing.T) {
	m := New(WithWidth(10), WithHeight(0))
	m.SetContent("test")
	if m.View() != "" {
		t.Error("zero height should return empty")
	}
}

func TestView_OffsetBeyondContent_P217(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent("line1\nline2")
	m.SetAutoFollow(false)
	m.SetYOffset(100) // way beyond content
	_ = m.View()
}

func TestView_Padding_P217(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetContent("short")
	m.SetAutoFollow(false)
	m.GotoTop()
	result := m.View()
	// Should pad to 5 lines
	lines := 0
	for _, c := range result {
		if c == '\n' {
			lines++
		}
	}
	if lines < 4 {
		t.Errorf("expected at least 4 newlines for padding, got %d", lines)
	}
}