package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// serializableTheme is the JSON representation of a Theme.
// Colors are stored as hex strings for readability.
type serializableTheme struct {
	Name string `json:"name"`

	// Base
	Bg     string `json:"bg"`
	Fg     string `json:"fg"`
	Accent string `json:"accent"`

	// Borders
	Border       string `json:"border"`
	BorderActive string `json:"border_active"`
	BorderMuted  string `json:"border_muted"`

	// Status
	Success string `json:"success"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
	Muted   string `json:"muted"`

	// Code
	CodeBg string `json:"code_bg"`
	CodeFg string `json:"code_fg"`

	// Diff
	DiffAdd  string `json:"diff_add"`
	DiffDel  string `json:"diff_del"`
	DiffMeta string `json:"diff_meta"`
	DiffHunk string `json:"diff_hunk"`
	DiffFile string `json:"diff_file"`

	// Blocks
	UserMsgBg    string `json:"user_msg_bg"`
	UserMsgFg    string `json:"user_msg_fg"`
	ThinkingBg   string `json:"thinking_bg"`
	ThinkingFg   string `json:"thinking_fg"`
	ToolCallBg   string `json:"tool_call_bg"`
	ToolResultBg string `json:"tool_result_bg"`
	ToolResultFg string `json:"tool_result_fg"`
	AssistantFg  string `json:"assistant_fg"`

	// Input
	PromptFg  string `json:"prompt_fg"`
	Separator string `json:"separator"`

	// Overlay
	MaskBg string `json:"mask_bg"`

	// Search
	SearchBarBg   string `json:"search_bar_bg"`
	SearchBarFg   string `json:"search_bar_fg"`
	SearchMatch   string `json:"search_match"`
	SearchNoMatch string `json:"search_no_match"`
}

// themeToSerializable converts a Theme to its JSON representation.
func themeToSerializable(t *Theme) serializableTheme {
	return serializableTheme{
		Name: t.Name,
		Bg:   colorToHexStr(t.Bg), Fg: colorToHexStr(t.Fg), Accent: colorToHexStr(t.Accent),
		Border: colorToHexStr(t.Border), BorderActive: colorToHexStr(t.BorderActive), BorderMuted: colorToHexStr(t.BorderMuted),
		Success: colorToHexStr(t.Success), Error: colorToHexStr(t.Error), Warning: colorToHexStr(t.Warning), Muted: colorToHexStr(t.Muted),
		CodeBg: colorToHexStr(t.CodeBg), CodeFg: colorToHexStr(t.CodeFg),
		DiffAdd: colorToHexStr(t.DiffAdd), DiffDel: colorToHexStr(t.DiffDel), DiffMeta: colorToHexStr(t.DiffMeta),
		DiffHunk: colorToHexStr(t.DiffHunk), DiffFile: colorToHexStr(t.DiffFile),
		UserMsgBg:    colorToHexStr(t.UserMsgBg), UserMsgFg: colorToHexStr(t.UserMsgFg),
		ThinkingBg:   colorToHexStr(t.ThinkingBg), ThinkingFg: colorToHexStr(t.ThinkingFg),
		ToolCallBg:   colorToHexStr(t.ToolCallBg), ToolResultBg: colorToHexStr(t.ToolResultBg), ToolResultFg: colorToHexStr(t.ToolResultFg),
		AssistantFg:  colorToHexStr(t.AssistantFg),
		PromptFg:     colorToHexStr(t.PromptFg), Separator: colorToHexStr(t.Separator),
		MaskBg:       colorToHexStr(t.MaskBg),
		SearchBarBg:  colorToHexStr(t.SearchBarBg), SearchBarFg: colorToHexStr(t.SearchBarFg),
		SearchMatch:  colorToHexStr(t.SearchMatch), SearchNoMatch: colorToHexStr(t.SearchNoMatch),
	}
}

// serializableToTheme converts JSON representation back to a Theme.
func serializableToTheme(s serializableTheme) *Theme {
	return &Theme{
		Name: s.Name,
		Bg:   hexStrToColor(s.Bg), Fg: hexStrToColor(s.Fg), Accent: hexStrToColor(s.Accent),
		Border: hexStrToColor(s.Border), BorderActive: hexStrToColor(s.BorderActive), BorderMuted: hexStrToColor(s.BorderMuted),
		Success: hexStrToColor(s.Success), Error: hexStrToColor(s.Error), Warning: hexStrToColor(s.Warning), Muted: hexStrToColor(s.Muted),
		CodeBg: hexStrToColor(s.CodeBg), CodeFg: hexStrToColor(s.CodeFg),
		DiffAdd: hexStrToColor(s.DiffAdd), DiffDel: hexStrToColor(s.DiffDel), DiffMeta: hexStrToColor(s.DiffMeta),
		DiffHunk: hexStrToColor(s.DiffHunk), DiffFile: hexStrToColor(s.DiffFile),
		UserMsgBg:    hexStrToColor(s.UserMsgBg), UserMsgFg: hexStrToColor(s.UserMsgFg),
		ThinkingBg:   hexStrToColor(s.ThinkingBg), ThinkingFg: hexStrToColor(s.ThinkingFg),
		ToolCallBg:   hexStrToColor(s.ToolCallBg), ToolResultBg: hexStrToColor(s.ToolResultBg), ToolResultFg: hexStrToColor(s.ToolResultFg),
		AssistantFg:  hexStrToColor(s.AssistantFg),
		PromptFg:     hexStrToColor(s.PromptFg), Separator: hexStrToColor(s.Separator),
		MaskBg:       hexStrToColor(s.MaskBg),
		SearchBarBg:  hexStrToColor(s.SearchBarBg), SearchBarFg: hexStrToColor(s.SearchBarFg),
		SearchMatch:  hexStrToColor(s.SearchMatch), SearchNoMatch: hexStrToColor(s.SearchNoMatch),
	}
}

// SaveToFile serializes the theme to a JSON file at the given path.
// The file format is human-readable JSON with hex color values.
func SaveToFile(t *Theme, path string) error {
	s := themeToSerializable(t)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("theme: marshal error: %w", err)
	}
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("theme: create dir: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("theme: write file: %w", err)
	}
	return nil
}

// LoadFromFile loads a theme from a JSON file at the given path.
func LoadFromFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("theme: read file: %w", err)
	}
	var s serializableTheme
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("theme: unmarshal error: %w", err)
	}
	return serializableToTheme(s), nil
}

// SaveActive saves the current active theme to a JSON file.
func SaveActive(path string) error {
	return SaveToFile(Active, path)
}

// LoadAndActivate loads a theme from file and sets it as active.
func LoadAndActivate(path string) error {
	t, err := LoadFromFile(path)
	if err != nil {
		return err
	}
	SetActive(t)
	return nil
}

// colorToHexStr converts a Color to a hex string.
func colorToHexStr(c Color) string {
	if c.Type == 3 { // buffer.ColorTrue
		return fmt.Sprintf("#%02X%02X%02X", c.R(), c.G(), c.B())
	}
	if c.Type == 2 { // buffer.Color256
		return fmt.Sprintf("256:%d", c.Val)
	}
	return ""
}

// hexStrToColor converts a hex string to a Color.
func hexStrToColor(s string) Color {
	if s == "" {
		return NoColor()
	}
	if len(s) > 4 && s[:4] == "256:" {
		var n int
		fmt.Sscanf(s[4:], "%d", &n)
		return Color{Type: 2, Val: uint32(n)}
	}
	return Hex(s)
}
