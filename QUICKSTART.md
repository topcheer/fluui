# Quickstart

## Installation

```bash
go get github.com/topcheer/fluui
```

## Minimal App

```go
package main

import (
    "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/component"
)

func main() {
    app := fluui.New()

    // Create a button
    btn := component.NewButton("Click me")
    btn.SetOnClick(func() {
        btn.SetText("Clicked!")
    })

    app.SetRoot(btn)
    app.Run()
}
```

## Hello World with Text

```go
package main

import (
    "fmt"
    "github.com/topcheer/fluui"
    "github.com/topcheer/fluui/component"
)

func main() {
    app := fluui.New()
    txt := component.NewText("Hello, Fluui!")
    txt.SetStyle(component.Style{Fg: component.ColorCyan, Bold: true})
    app.SetRoot(txt)
    app.Run()
}
```

## Components

Fluui provides 135+ components. Here are the most common:

| Category | Components |
|----------|-----------|
| **Input** | Button, Checkbox, RadioGroup, Switch, Slider, Input, TextArea, SearchBar |
| **Display** | Text, Table, TreeView, ListView, Badge, Banner, Notification, Toast |
| **Layout** | SplitPane, Tabs, Drawer, Accordion, Collapsible, Divider, Pages |
| **AI** | MessageBubble, ConversationView, ChatComposer, ToolCallView, CodeBlock, MarkdownStream |
| **Data** | BarChart, LineChart, PieChart, Sparkline, Gauge, MetricBar, Rating |
| **Feedback** | ProgressBar, Spinner, Skeleton, ConfidenceMeter, CircularProgress |

## Explore Examples

Run the Example functions to see each component in action:

```bash
go test ./component/ -run 'Example' -v
```

## Next Steps

- Read the [README](README.md) for full feature list
- Check [STABILITY.md](STABILITY.md) for API stability tiers
- Browse `examples/` directory for complete demo apps
