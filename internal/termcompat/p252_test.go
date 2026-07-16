package termcompat

import "testing"

func TestColorDepth_TrueColor_P252(t *testing.T) {
	c := Capabilities{HasTrueColor: true, Has256Color: true}
	if c.ColorDepth() != 24 {
		t.Error("truecolor should return 24")
	}
}

func TestColorDepth_256Color_P252(t *testing.T) {
	c := Capabilities{Has256Color: true}
	if c.ColorDepth() != 8 {
		t.Error("256color should return 8")
	}
}

func TestColorDepth_Basic_P252(t *testing.T) {
	c := Capabilities{}
	if c.ColorDepth() != 4 {
		t.Error("no color should return 4")
	}
}
