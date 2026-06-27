# Fluui Phase 2 Demo

Interactive demo showcasing the Phase 2 component system: Border, Text, Flex layout, and ScrollView.

## Run

```bash
cd /Volumes/new/ggai/fluui
go run ./cmd/demo2/
```

## Features

| Feature | Key | Description |
|---------|-----|-------------|
| **Border** | — | Root border with centered title; inner border around component showcase |
| **Text** | — | Styled text in Dracula palette colors (pink, cyan, green, yellow, purple, orange) |
| **Flex (Column)** | — | Vertical stacking of 40 text lines inside ScrollView |
| **Flex (Row)** | — | Horizontal arrangement of component labels with gaps |
| **ScrollView** | Up/Down | Scroll one line at a time |
| **ScrollView** | Left/Right | Scroll 5 lines at a time (page scroll) |
| **Resize** | — | Auto-adapts layout on terminal resize |
| **Quit** | q or Esc | Exit the demo |

## Layout

```
┌─── Fluui Phase 2 Demo ───────────────────┐
│ Welcome to Fluui Phase 2!                  │
│ A Go TUI framework with components...      │
│                                           │
│ ── Components ──                           │
│ [Text] [Border] [Flex] [ScrollView]       │
│                                           │
│ Line  1: The quick brown fox...           │
│ Line  2: The quick brown fox...           │
│ Line  3: The quick brown fox...           │
│ ...                          ░            │
│                             █             │
│  Up/Down: scroll   q/Esc: quit            │
└───────────────────────────────────────────┘
```

## Color Palette (Dracula)

| Element | Color | Hex |
|---------|-------|-----|
| Pink | H1, labels | #ff79c6 |
| Purple | Border, Bullet | #bd93f9 |
| Green | Text label | #50fa7b |
| Cyan | Labels | #8be9fd |
| Yellow | Border label | #f1fa8c |
| Orange | Labels | #ffb86c |
| Background | Terminal bg | #282a36 |
| Body text | Foreground | #f8f8f2 |
