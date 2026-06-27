# Component System

Fluui's component system provides composable UI primitives with a measure/paint layout model.

## Component Interface

```go
type Component interface {
    Measure(Constraints) Size
    SetBounds(Rect)
    Paint(*buffer.Buffer)
    Children() []Component
}
```

- **Measure**: Compute desired size given constraints (max width/height)
- **SetBounds**: Receive the allocated rectangle
- **Paint**: Render into the buffer at the allocated position
- **Children**: Return child components (for tree traversal)

## Constraints

```go
type Constraints struct {
    MaxWidth  int  // 0 = unbounded
    MaxHeight int  // 0 = unbounded
}
```

Helpers:
- `Unbounded()` — no limits
- `Fixed(w, h)` — exact size
- `Bounded(maxW, maxH)` — bounded

## Built-in Components

### Text

```go
text := component.NewText("Hello, World!")
text.Style.Fg = buffer.RGB(0xFF, 0x79, 0xC6)
text.Style.Flags = buffer.Bold
```

Renders styled text. Width = display width of content. Height = 1.

### Border

```go
text := component.NewText("Content")
border := component.NewBorder(text)
border.Title = " My Box "
border.Style.Fg = buffer.RGB(0x62, 0x72, 0xA4)
```

Renders a box-drawing border around a child component. Optional centered title.

Box characters: `┌─┐│└─┘├┤┬┴┼` (Unicode box drawing).

### ScrollView

```go
doc := component.NewText(longContent)
scroll := component.NewScrollView(doc)
scroll.ScrollDown(5)
```

A scrollable viewport. Features:
- Virtual scrolling (only paints visible blocks)
- Scrollbar with drag support
- `ScrollUp(n)` / `ScrollDown(n)` / `ScrollTo(offset)`
- `Offset()` / `MaxOffset()`
- `HandleScrollbarDown/Drag/Up()` for mouse interaction
- `IsDragging()` — true during scrollbar drag
- `PaintVisible` optimization for large content (O(log n) binary search)

## Layout Components

### Flex (Row/Column)

```go
row := layout.NewFlex(layout.Row)
row.AddChild(text1)
row.AddChild(text2)
```

Arranges children horizontally (Row) or vertically (Column). Optional gap.

```go
col := layout.NewFlexGap(layout.Column, 2) // 2-cell gap between children
```

### Center

```go
center := layout.NewCenter(text)
```

Centers child within available space.

### Padding

```go
padded := layout.NewPadding(text, layout.Insets{Top: 1, Bottom: 1, Left: 2, Right: 2})
```

Adds insets around child.

### Stack

```go
stack := layout.NewStack(background, overlay)
```

Overlays children in z-order (last = topmost).

## Composing Components

```go
// A bordered, scrollable text panel
scroll := component.NewScrollView(text)
border := component.NewBorder(scroll)
border.Title = " Document "

// Layout: sidebar + main panel
sidebar := component.NewBorder(component.NewText("Menu"))
row := layout.NewFlex(layout.Row)
row.AddChild(sidebar)
row.AddChild(border)
```

## Rendering

Components are painted via `Paint(buf *buffer.Buffer)`. The event loop calls this every frame:

```go
app.OnPaint(func(buf *buffer.Buffer) {
    root.Measure(component.Constraints{MaxWidth: w, MaxHeight: h})
    root.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
    root.Paint(buf)
})
```
