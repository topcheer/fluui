package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
)

func TestModal_MeasureDefaults_P248(t *testing.T) {
	m := NewModal("m1", "Test", component.NewText("Body"), []string{"OK"})
	s := m.Measure(component.Constraints{})
	if s.W < 20 || s.W > 80 {
		t.Errorf("width=%d", s.W)
	}
}

func TestModal_MeasureSmallMax_P248(t *testing.T) {
	m := NewModal("m2", "Test", component.NewText("Body"), []string{"OK"})
	s := m.Measure(component.Constraints{MaxWidth: 10, MaxHeight: 5})
	if s.W < 20 {
		t.Errorf("width=%d", s.W)
	}
}

func TestModal_SetBoundsClamp_P248(t *testing.T) {
	m := NewModal("m3", "Test", component.NewText("Body"), []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 3})
}

func TestModal_SetBoundsLarge_P248(t *testing.T) {
	m := NewModal("m4", "Test", component.NewText("Body"), []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 200, H: 100})
}
