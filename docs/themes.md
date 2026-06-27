# Theme System

Fluui includes 5 built-in themes and a converter from global themes to markdown themes.

## Built-in Themes

| Theme | Vibe | Accent |
|---|---|---|
| **Dracula** (default) | Dark purple/pink | #bd93f9 purple |
| **Nord** | Arctic blue | #88c0d0 ice blue |
| **Gruvbox** | Warm earth tones | #fe8019 orange |
| **SolarizedDark** | Precision colors | #268bd2 blue |
| **TokyoNight** | Neon city night | #7aa2f7 neon blue |

## Using Themes

### Set Active Theme

```go
import "github.com/topcheer/fluui/theme"

// Set directly
theme.SetActive(theme.Nord())

// Or via ChatApp
chat.SetTheme(theme.Gruvbox())
```

### Cycle Themes (Ctrl+T)

```go
// In your key handler:
case k.Rune == 't' && k.Modifiers&term.ModCtrl != 0:
    chat.CycleTheme()

// Shift+Ctrl+T for reverse
chat.CycleThemeBack()
```

A toast notification shows the theme name for 3 seconds.

### Programmatic Access

```go
theme.Get()              // current active theme
theme.Default()          // Dracula
theme.Builtin()          // all 5 themes
theme.CurrentIndex()     // index in Builtin() list
```

## Theme Structure

```go
type Theme struct {
    Name string

    // Base
    Bg     Color   // terminal background
    Fg     Color   // terminal foreground
    Accent Color   // highlight color

    // Borders
    Border       Color
    BorderActive Color
    BorderMuted  Color

    // Status
    Success Color
    Error   Color
    Warning Color
    Muted   Color

    // Code
    CodeBg Color
    CodeFg Color

    // Diff
    DiffAdd  Color
    DiffDel  Color
    DiffMeta Color
    DiffHunk Color
    DiffFile Color

    // Block-specific
    UserMsgBg, UserMsgFg     Color
    ThinkingBg, ThinkingFg   Color
    ToolCallBg, ToolResultBg Color
    ToolResultFg, AssistantFg Color

    // Input
    PromptFg  Color
    Separator Color

    // Overlay
    MaskBg Color
}
```

## Markdown Theme

The markdown renderer uses a separate `MarkdownTheme` with 20 color fields.

```go
import "github.com/topcheer/fluui/markdown"

// Default Dracula-based markdown theme
mdTheme := markdown.DefaultTheme()

// Convert from global theme
mdTheme := markdown.MarkdownThemeFromTheme(theme.Get())

// Use with renderer
renderer := markdown.NewMarkdownRenderer(mdTheme, 80)
```

### Color Mapping

| MarkdownTheme Field | Global Theme Source |
|---|---|
| H1 | Accent |
| H2 | DiffHunk (cyan) |
| H3 | Success (green) |
| H4 | DiffFile (purple) |
| H5, H6 | Warning (yellow) |
| CodeFg | CodeFg |
| CodeBg | CodeBg |
| LinkFg | Accent |
| QuoteFg, QuoteBar | Muted / Border |
| TableBorder | Border |
| TableHeader | Accent |
| Hr | Separator |

## Custom Themes

```go
custom := &theme.Theme{
    Name:    "Custom",
    Bg:      theme.C(0x1a, 0x1a, 0x2e),
    Fg:      theme.C(0xee, 0xee, 0xee),
    Accent:  theme.C(0xe9, 0x45, 0x60),
    Border:  theme.C(0x0f, 0x34, 0x60),
    Success: theme.C(0x16, 0x21, 0x3e),
    Error:   theme.C(0xe9, 0x45, 0x60),
    Warning: theme.C(0xf3, 0xa9, 0x53),
    // ... fill all fields
}
theme.SetActive(custom)
```

## Color Types

```go
buffer.NoColor()              // terminal default
buffer.NamedColor(1)          // ANSI named (0-15)
buffer.Color256Val(196)        // 256-color palette
buffer.RGB(0xFF, 0x79, 0xC6)  // TrueColor (24-bit)
buffer.Hex("#ff79c6")          // parse hex string
```
