package viewport

import (
	"strings"
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
)

func TestNewWithOptions_P200(t *testing.T) {
	m := New(WithWidth(60), WithHeight(10))
	if m.Width() != 60 {
		t.Errorf("expected width 60, got %d", m.Width())
	}
	if m.Height() != 10 {
		t.Errorf("expected height 10, got %d", m.Height())
	}
}

func TestSetContentString_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	lines := []string{}
	for i := 0; i < 10; i++ {
		lines = append(lines, "line "+string(rune('0'+i)))
	}
	content := strings.Join(lines, "\n")
	m.SetContent(content)
	if m.TotalLineCount() != 10 {
		t.Errorf("expected 10 lines, got %d", m.TotalLineCount())
	}
}

func TestView_RendersVisible_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("a\nb\nc\nd\ne\nf")
	view := m.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 visible lines, got %d", len(lines))
	}
}

func TestScrollUpDown_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetAutoFollow(false)
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.GotoTop()
	// Now at top (yOffset=0)
	m.ScrollUp(1) // at top already, should stay
	if m.YOffset() != 0 {
		t.Errorf("expected yOffset 0, got %d", m.YOffset())
	}
	m.ScrollDown(2)
	if m.YOffset() != 2 {
		t.Errorf("expected yOffset 2, got %d", m.YOffset())
	}
	m.ScrollUp(1)
	if m.YOffset() != 1 {
		t.Errorf("expected yOffset 1, got %d", m.YOffset())
	}
}

func TestGotoBottom_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.GotoBottom()
	if !m.AtBottom() {
		t.Error("should be at bottom after GotoBottom")
	}
	if m.YOffset() != 7 {
		t.Errorf("expected yOffset 7, got %d", m.YOffset())
	}
}

func TestGotoTop_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.GotoBottom()
	m.GotoTop()
	if m.YOffset() != 0 {
		t.Errorf("expected yOffset 0, got %d", m.YOffset())
	}
}

func TestAtBottom_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3")
	if !m.AtBottom() {
		t.Error("should be at bottom when content fits")
	}
}

func TestSetSize_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.SetSize(60, 5)
	if m.Width() != 60 || m.Height() != 5 {
		t.Error("SetSize should update dimensions")
	}
}

func TestSetWidthHeight_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetWidth(80)
	m.SetHeight(10)
	if m.Width() != 80 || m.Height() != 10 {
		t.Error("SetWidth/SetHeight failed")
	}
}

func TestScrollIndicatorStyle_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	// At bottom with autoFollow → empty
	if m.ScrollIndicatorStyle() != "" {
		t.Error("should be empty at bottom with autoFollow")
	}
	// Scroll up → non-empty
	m.ScrollUp(1)
	if m.ScrollIndicatorStyle() == "" {
		t.Error("should show indicator when not at bottom")
	}
}

func TestUpdateKeyPress_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.SetAutoFollow(false)
	m.GotoTop() // start from top
	// Test key down
	m2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m2.YOffset() != 1 {
		t.Errorf("expected yOffset 1 after KeyDown, got %d", m2.YOffset())
	}
	// Test key up
	m3, _ := m2.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m3.YOffset() != 0 {
		t.Errorf("expected yOffset 0 after KeyUp, got %d", m3.YOffset())
	}
}

func TestUpdateMouseWheel_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m2, _ := m.Update(tea.MouseWheelMsg{Down: true})
	if m2.YOffset() < 1 {
		t.Error("Update with MouseWheel down should scroll down")
	}
}

func TestUpdatePageKeys_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	// PageDown should scroll by viewport height
	m2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyPgDn})
	if m2.YOffset() < 3 {
		t.Errorf("expected yOffset >= 3 after PgDn, got %d", m2.YOffset())
	}
	// Home should go to top
	m3, _ := m2.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if m3.YOffset() != 0 {
		t.Errorf("expected yOffset 0 after Home, got %d", m3.YOffset())
	}
	// End should go to bottom
	m4, _ := m3.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if !m4.AtBottom() {
		t.Error("should be at bottom after End")
	}
}

func TestAutoFollow_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	if !m.AutoFollow() {
		t.Error("should default to autoFollow")
	}
	m.ScrollUp(1)
	if m.AutoFollow() {
		t.Error("ScrollUp should disable autoFollow")
	}
	m.GotoBottom()
	if !m.AutoFollow() {
		t.Error("GotoBottom should enable autoFollow")
	}
	m.SetAutoFollow(false)
	if m.AutoFollow() {
		t.Error("SetAutoFollow(false) should disable")
	}
}

func TestContent_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("hello\nworld")
	if m.Content() != "hello\nworld" {
		t.Error("Content() mismatch")
	}
}

func TestSetYOffset_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	m.SetYOffset(5)
	if m.YOffset() != 5 {
		t.Errorf("expected 5, got %d", m.YOffset())
	}
}

func TestVpFieldAccess_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("test content")
	// Verify vp field is accessible (ggcode uses m.vp.SetYOffset)
	if m.vp == nil {
		t.Error("vp field should not be nil")
	}
	m.vp.SetYOffset(0)
}

func TestEmptyContent_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(3))
	m.SetContent("")
	if m.TotalLineCount() != 0 {
		t.Errorf("expected 0 lines, got %d", m.TotalLineCount())
	}
	if m.View() != "" {
		t.Error("View should be empty for empty content")
	}
}

func TestVisibleLineCount_P200(t *testing.T) {
	m := New(WithWidth(40), WithHeight(5))
	if m.VisibleLineCount() != 5 {
		t.Errorf("expected 5, got %d", m.VisibleLineCount())
	}
}

func TestInit_P200(t *testing.T) {
	m := New()
	if m.Init() != nil {
		t.Error("Init should return nil")
	}
}