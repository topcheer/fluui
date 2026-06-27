package theme

import (
	"testing"
)

// TestThemeColorFields_NonZero verifies that all color fields in the Theme
// struct are populated (non-default) for each built-in theme — with the
// exception of NoColor fields that are intentionally left as terminal default.
func TestThemeColorFields_NonZero(t *testing.T) {
	// Fields that are intentionally NoColor (terminal default)
	noColorFields := map[string]bool{
		"UserMsgBg":    true,
		"ThinkingBg":   true,
		"ToolCallBg":   true,
		"ToolResultBg": true,
	}

	// All color field names in the Theme struct (excluding Name).
	colorFields := []string{
		"Bg", "Fg", "Accent",
		"Border", "BorderActive", "BorderMuted",
		"Success", "Error", "Warning", "Muted",
		"CodeBg", "CodeFg",
		"DiffAdd", "DiffDel", "DiffMeta", "DiffHunk", "DiffFile",
		"UserMsgFg", "ThinkingFg", "ToolResultFg", "AssistantFg",
		"PromptFg", "Separator",
		"MaskBg",
		"SearchBarBg", "SearchBarFg", "SearchMatch", "SearchNoMatch",
	}

	for _, themeFn := range []func() *Theme{Dracula, Nord, Gruvbox, SolarizedDark, TokyoNight} {
		t.Run(themeName(themeFn), func(t *testing.T) {
			th := themeFn()
			for _, field := range colorFields {
				c := getThemeColor(th, field)
				if noColorFields[field] {
					continue // skip intentionally NoColor fields
				}
				if c.IsDefault() {
					t.Errorf("%s.%s is default/NoColor, expected a real color", th.Name, field)
				}
				// All non-NoColor fields should be TrueColor
					if c.Type != bufferColorTrue {
						t.Errorf("%s.%s has Type=%d, want TrueColor", th.Name, field, c.Type)
					}
				}
			})
		}
	}

// bufferColorTrue is the ColorType value for TrueColor.
const bufferColorTrue = 3 // ColorTrue from buffer package

func themeName(fn func() *Theme) string {
	return fn().Name
}

// getThemeColor retrieves a color field from a Theme by name.
// This uses a switch to access each named field.
func getThemeColor(th *Theme, field string) Color {
	switch field {
	case "Bg":
		return th.Bg
	case "Fg":
		return th.Fg
	case "Accent":
		return th.Accent
	case "Border":
		return th.Border
	case "BorderActive":
		return th.BorderActive
	case "BorderMuted":
		return th.BorderMuted
	case "Success":
		return th.Success
	case "Error":
		return th.Error
	case "Warning":
		return th.Warning
	case "Muted":
		return th.Muted
	case "CodeBg":
		return th.CodeBg
	case "CodeFg":
		return th.CodeFg
	case "DiffAdd":
		return th.DiffAdd
	case "DiffDel":
		return th.DiffDel
	case "DiffMeta":
		return th.DiffMeta
	case "DiffHunk":
		return th.DiffHunk
	case "DiffFile":
		return th.DiffFile
	case "UserMsgBg":
		return th.UserMsgBg
	case "UserMsgFg":
		return th.UserMsgFg
	case "ThinkingBg":
		return th.ThinkingBg
	case "ThinkingFg":
		return th.ThinkingFg
	case "ToolCallBg":
		return th.ToolCallBg
	case "ToolResultBg":
		return th.ToolResultBg
	case "ToolResultFg":
		return th.ToolResultFg
	case "AssistantFg":
		return th.AssistantFg
	case "PromptFg":
		return th.PromptFg
	case "Separator":
		return th.Separator
	case "MaskBg":
		return th.MaskBg
	case "SearchBarBg":
		return th.SearchBarBg
	case "SearchBarFg":
		return th.SearchBarFg
	case "SearchMatch":
		return th.SearchMatch
	case "SearchNoMatch":
		return th.SearchNoMatch
	}
	return NoColor()
}

// TestColorHelpers verifies C(), Hex(), and NoColor() helper functions.
func TestColorHelpers(t *testing.T) {
	// C() should create TrueColor
	c := C(255, 128, 0)
	if c.Type != bufferColorTrue {
		t.Errorf("C().Type = %d, want %d (TrueColor)", c.Type, bufferColorTrue)
	}
	if c.R() != 255 || c.G() != 128 || c.B() != 0 {
		t.Errorf("C(255,128,0) = R:%d G:%d B:%d", c.R(), c.G(), c.B())
	}

	// Hex() valid
	h := Hex("#ff0000")
	if h.Type != bufferColorTrue {
		t.Errorf("Hex valid should be TrueColor")
	}
	if h.R() != 255 || h.G() != 0 || h.B() != 0 {
		t.Errorf("Hex(#ff0000) = R:%d G:%d B:%d", h.R(), h.G(), h.B())
	}

	// Hex() invalid
	hi := Hex("invalid")
	if !hi.IsDefault() {
		t.Error("Hex(invalid) should return default")
	}

	// NoColor() should be default
	nc := NoColor()
	if !nc.IsDefault() {
		t.Error("NoColor() should be default")
	}
}

// TestThemeColorConsistency verifies that Error is always red-ish and
// Success is always green-ish across all themes.
func TestThemeColorConsistency(t *testing.T) {
	for _, themeFn := range []func() *Theme{Dracula, Nord, Gruvbox, SolarizedDark, TokyoNight} {
		th := themeFn()
		t.Run(th.Name, func(t *testing.T) {
			// Error should have a high red component.
			if th.Error.R() < 128 {
				t.Errorf("%s.Error.R() = %d, expected high red (>=128)", th.Name, th.Error.R())
			}
			// Success should have a high green component.
			if th.Success.G() < 128 {
				t.Errorf("%s.Success.G() = %d, expected high green (>=128)", th.Name, th.Success.G())
			}
			// MaskBg should be black.
			if th.MaskBg.R() != 0 || th.MaskBg.G() != 0 || th.MaskBg.B() != 0 {
				t.Errorf("%s.MaskBg should be black", th.Name)
			}
		})
	}
}
