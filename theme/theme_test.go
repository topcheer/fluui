package theme

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestDracula(t *testing.T) {
	d := Dracula()
	if d.Name != "Dracula" {
		t.Errorf("Name: got %q", d.Name)
	}
	// Bg should be #282a36
	expected := buffer.RGB(0x28, 0x2A, 0x36)
	if !d.Bg.Equal(expected) {
		t.Errorf("Bg: got %v, want %v", d.Bg, expected)
	}
	// Fg should be #f8f8f2
	expected = buffer.RGB(0xF8, 0xF8, 0xF2)
	if !d.Fg.Equal(expected) {
		t.Errorf("Fg: got %v, want %v", d.Fg, expected)
	}
}

func TestThemeColorsNonZero(t *testing.T) {
	themes := []*Theme{Dracula(), Nord(), Gruvbox(), SolarizedDark(), TokyoNight()}
	for _, th := range themes {
		t.Run(th.Name, func(t *testing.T) {
			checks := []struct {
				name string
				c    Color
			}{
				{"Bg", th.Bg}, {"Fg", th.Fg}, {"Accent", th.Accent},
				{"Border", th.Border}, {"BorderActive", th.BorderActive},
				{"Success", th.Success}, {"Error", th.Error},
				{"Warning", th.Warning}, {"Muted", th.Muted},
				{"CodeBg", th.CodeBg}, {"CodeFg", th.CodeFg},
				{"DiffAdd", th.DiffAdd}, {"DiffDel", th.DiffDel},
				{"DiffMeta", th.DiffMeta}, {"DiffHunk", th.DiffHunk},
				{"DiffFile", th.DiffFile},
				{"UserMsgFg", th.UserMsgFg}, {"ThinkingFg", th.ThinkingFg},
				{"ToolResultFg", th.ToolResultFg}, {"AssistantFg", th.AssistantFg},
				{"PromptFg", th.PromptFg}, {"Separator", th.Separator},
				{"MaskBg", th.MaskBg},
			}
			for _, ch := range checks {
				if ch.c.IsDefault() {
					t.Errorf("%s: %s should not be default color", th.Name, ch.name)
				}
			}
		})
	}
}

func TestThemesAreDistinct(t *testing.T) {
	themes := Builtin()
	if len(themes) != 5 {
		t.Fatalf("expected 5 builtin themes, got %d", len(themes))
	}
	// Each theme should have a unique Bg color
	seen := make(map[string]bool)
	for _, th := range themes {
		key := th.Name
		if seen[key] {
			t.Errorf("duplicate theme name: %s", key)
		}
		seen[key] = true
	}
	// Dracula and Nord should have different backgrounds
	d := Dracula()
	n := Nord()
	if d.Bg.Equal(n.Bg) {
		t.Error("Dracula and Nord should have different Bg")
	}
	if d.Accent.Equal(n.Accent) {
		t.Error("Dracula and Nord should have different Accent")
	}
}

func TestSetActive(t *testing.T) {
	original := Active
	defer func() { Active = original }()

	nord := Nord()
	SetActive(nord)
	if Active != nord {
		t.Error("SetActive should update Active")
	}

	// nil should be ignored
	SetActive(nil)
	if Active != nord {
		t.Error("SetActive(nil) should not change Active")
	}
}

func TestGet(t *testing.T) {
	original := Active
	defer func() { Active = original }()

	g := Get()
	if g != Active {
		t.Error("Get should return Active")
	}
}

func TestDefault(t *testing.T) {
	d := Default()
	if d.Name != "Dracula" {
		t.Errorf("Default: expected Dracula, got %s", d.Name)
	}
}

func TestBuiltin(t *testing.T) {
	all := Builtin()
	names := []string{"Dracula", "Nord", "Gruvbox", "SolarizedDark", "TokyoNight"}
	if len(all) != len(names) {
		t.Fatalf("expected %d themes, got %d", len(names), len(all))
	}
	for i, name := range names {
		if all[i].Name != name {
			t.Errorf("theme %d: expected %s, got %s", i, name, all[i].Name)
		}
	}
}

func TestC(t *testing.T) {
	c := C(0xFF, 0x00, 0x00)
	expected := buffer.RGB(0xFF, 0x00, 0x00)
	if !c.Equal(expected) {
		t.Errorf("C: got %v, want %v", c, expected)
	}
}

func TestNoColor(t *testing.T) {
	nc := NoColor()
	if !nc.IsDefault() {
		t.Error("NoColor should be default")
	}
}

func TestDraculaDiffColors(t *testing.T) {
	d := Dracula()
	// DiffAdd should be green
	if !d.DiffAdd.Equal(buffer.RGB(0x50, 0xFA, 0x7B)) {
		t.Errorf("DiffAdd: got %v", d.DiffAdd)
	}
	// DiffDel should be red
	if !d.DiffDel.Equal(buffer.RGB(0xFF, 0x55, 0x55)) {
		t.Errorf("DiffDel: got %v", d.DiffDel)
	}
	// DiffHunk should be cyan
	if !d.DiffHunk.Equal(buffer.RGB(0x8B, 0xE9, 0xFD)) {
		t.Errorf("DiffHunk: got %v", d.DiffHunk)
	}
	// DiffFile should be purple
	if !d.DiffFile.Equal(buffer.RGB(0xBD, 0x93, 0xF9)) {
		t.Errorf("DiffFile: got %v", d.DiffFile)
	}
	// DiffMeta should be gray-blue
	if !d.DiffMeta.Equal(buffer.RGB(0x62, 0x72, 0xA4)) {
		t.Errorf("DiffMeta: got %v", d.DiffMeta)
	}
}
