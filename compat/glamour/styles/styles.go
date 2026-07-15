// Package styles provides predefined glamour style configurations.
//
// This mirrors charm.land/glamour/v2/styles, providing DarkStyleConfig,
// LightStyleConfig, and NoTTYStyleConfig.
package styles

import (
	"github.com/topcheer/fluui/compat/glamour/ansi"
)

// strPtr returns a pointer to a string.
func strPtr(s string) *string { return &s }

// DarkStyleConfig is the default dark mode style configuration.
var DarkStyleConfig = ansi.StyleConfig{
	Document: ansi.Document{
		Margin:          ansi.PtrUint(2),
		Color:           strPtr("#a9b1d6"),
		BackgroundColor: nil,
	},
	Code: ansi.StyleBlock{
		Color: strPtr("#7aa2f7"),
	},
	CodeBlock: ansi.StyleBlock{
		Color: strPtr("#a9b1d6"),
	},
}

// LightStyleConfig is the default light mode style configuration.
var LightStyleConfig = ansi.StyleConfig{
	Document: ansi.Document{
		Margin:          ansi.PtrUint(2),
		Color:           strPtr("#334155"),
		BackgroundColor: nil,
	},
	Code: ansi.StyleBlock{
		Color: strPtr("#005f87"),
	},
	CodeBlock: ansi.StyleBlock{
		Color: strPtr("#334155"),
	},
}

// NoTTYStyleConfig is the style configuration for non-terminal output.
var NoTTYStyleConfig = ansi.StyleConfig{
	Document: ansi.Document{
		Margin: ansi.PtrUint(0),
	},
}

// DefaultStyleConfig returns DarkStyleConfig (glamour default).
func DefaultStyleConfig() ansi.StyleConfig {
	return DarkStyleConfig
}