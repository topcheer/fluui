package styles

import (
	"testing"
)

func TestDarkStyleConfig_P200(t *testing.T) {
	if DarkStyleConfig.Document.Color == nil {
		t.Error("DarkStyleConfig should have Document.Color")
	}
	if DarkStyleConfig.Code.Color == nil {
		t.Error("DarkStyleConfig should have Code.Color")
	}
}

func TestLightStyleConfig_P200(t *testing.T) {
	if LightStyleConfig.Document.Color == nil {
		t.Error("LightStyleConfig should have Document.Color")
	}
}

func TestNoTTYStyleConfig_P200(t *testing.T) {
	if NoTTYStyleConfig.Document.Margin == nil {
		t.Error("NoTTYStyleConfig should have Document.Margin")
	}
}

func TestDefaultStyleConfig_P200(t *testing.T) {
	cfg := DefaultStyleConfig()
	_ = cfg
}