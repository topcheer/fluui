# Fluui — AI-Native TUI Library for Go

> 流畅 (fluent) + UI = Fluui
> 一个从零构建的、为 AI 对话场景深度优化的终端 UI 框架。

## 设计哲学

1. **Streaming-first** — 每一层都为流式数据优化，从 token delta 到屏幕像素的延迟 < 16ms
2. **Block-centric** — 内容以语义化的 Block 为单位，不是扁平的文本行
3. **Zero-flicker** — 双缓冲 diff 渲染，流式更新时屏幕绝对不闪
4. **Mouse-native** — 可折叠、可点击、可复制，鼠标是一等公民
5. **No TUI dependency** — 从 termios 到渲染引擎全部自研

---

## 模块结构

```
fluui/
├── go.mod                          # module github.com/topcheer/fluui
├── DESIGN.md                       # 本文档
│
├── internal/
│   ├── term/                       # Layer 1-2: 终端抽象 (不导出, 内部使用)
│   │   ├── term.go                 # Terminal 接口 + 实现 (alt screen, raw mode)
│   │   ├── input.go                # 输入解析器 (键盘 + 鼠标 + 粘贴 + resize)
│   │   ├── output.go               # ANSI escape sequence writer
│   │   ├── terminfo.go             # 终端能力检测 (颜色、鼠标、剪贴板)
│   │   └── rawmode_unix.go         # termios 封装 (golang.org/x/term)
│   │
│   ├── buffer/                     # Layer 3: 渲染基础
│   │   ├── cell.go                 # Cell (最小渲染单元)
│   │   ├── style.go                # Style (前景、背景、修饰)
│   │   ├── color.go                # Color (Named / 256 / TrueColor)
│   │   ├── buffer.go               # Buffer (二维 Cell 网格)
│   │   └── diff.go                 # Buffer diff 算法
│   │
│   └── wcwidth/                    # 宽字符宽度计算 (CJK/emoji)
│       └── wcwidth.go              # Unicode East Asian Width
│
├── render/                         # Layer 3: 渲染引擎
│   ├── renderer.go                 # 双缓冲 diff renderer
│   ├── painter.go                  # 将 Buffer 转为 ANSI 输出
│   └── cursor.go                   # 光标状态管理
│
├── markdown/                       # Layer 3.5: Markdown 渲染器
│   ├── renderer.go                 # goldmark AST → []CellRow
│   ├── theme.go                    # Markdown 主题
│   ├── codeblock.go                # 代码块 (chroma 高亮 + 边框 + 复制按钮)
│   ├── table.go                    # 表格渲染 (列宽计算)
│   ├── incremental.go              # 流式增量解析 (处理未闭合的代码块等)
│   └── width.go                    # 文本折行 (CJK 安全)
│
├── component/                      # Layer 4: 组件系统
│   ├── component.go                # Component 接口
│   ├── node.go                     # 组件树节点
│   ├── layout/
│   │   ├── flex.go                 # Flex 布局 (Row/Column)
│   │   ├── stack.go                # Stack 布局 (重叠)
│   │   ├── padding.go              # 内边距
│   │   └── center.go               # 居中
│   ├── primitive/
│   │   ├── text.go                 # 静态文本
│   │   ├── border.go               # 边框 (圆角/直角 + 标题)
│   │   ├── spacer.go               # 弹性空白
│   │   └── box.go                  # 带背景色的矩形
│   ├── scroll/
│   │   └── scroll_view.go          # 滚动视图 (auto-follow)
│   ├── input/
│   │   └── text_input.go           # 文本输入框 (带光标)
│   └── focus/
│       └── focus.go                # 焦点管理
│
├── event/                          # 事件系统
│   ├── event.go                    # 事件类型定义
│   ├── loop.go                     # 主事件循环 (channel 驱动)
│   └── dispatch.go                 # 事件分发 (capture → target → bubble)
│
├── block/                          # Layer 5: AI Content Blocks ⭐
│   ├── block.go                    # Block 接口 + 生命周期
│   ├── registry.go                 # Block 注册表 (自定义 block 插件)
│   ├── container.go                # BlockContainer (管理 block 列表)
│   ├── thinking.go                 # ThinkingBlock (可折叠)
│   ├── tool_call.go                # ToolCallBlock (工具调用)
│   ├── tool_result.go              # ToolResultBlock (工具结果)
│   ├── assistant_text.go           # AssistantTextBlock (流式 markdown)
│   ├── user_message.go             # UserMessageBlock
│   ├── code.go                     # CodeBlock (独立代码块)
│   ├── error.go                    # ErrorBlock
│   └── stream.go                   # 流式更新分发器
│
├── animation/                      # 动画系统
│   ├── animation.go                # Animation 接口 + 管理器
│   ├── spinner.go                  # 加载动画 (多种样式)
│   ├── fade.go                     # 淡入淡出
│   └── transition.go               # 展开/折叠过渡
│
├── theme/                          # 主题系统
│   ├── theme.go                    # Theme 定义
│   ├── builtin.go                  # 内置主题 (dark/light/monokai/...)
│   └── color.go                    # 颜色调色板
│
├── hit/                            # 鼠标命中测试
│   ├── region.go                   # HitRegion 定义
│   └── tree.go                     # 区域树 (高效命中查询)
│
├── overlay/                        # Overlay / Modal 层 (Layer 4.5)
│   ├── overlay.go                  # Overlay 管理器 (z-index 层叠)
│   ├── modal.go                    # Modal 对话框 (居中 + 遮罩)
│   ├── popup.go                    # Popup 弹出层 (代码全屏/链接预览)
│   └── mask.go                     # 半透明遮罩
│
└── app.go                          # Layer 6: Application 入口
```

---

## 核心数据类型

### Cell — 最小渲染单元

```go
// internal/buffer/cell.go

// Cell 是屏幕上的一个字符位置。一个 Cell 持有:
// - 一个字符 (可能是 CJK 双宽字符)
// - 完整的样式信息
// - 链接信息 (用于可点击区域)
type Cell struct {
    Rune  rune
    Width int    // 显示宽度: 0(组合字符) / 1(ASCII) / 2(CJK)

    // 样式
    Fg    Color
    Bg    Color
    Flags StyleFlags

    // 可选: 链接信息 (鼠标点击)
    Link  *Link
}

type StyleFlags uint8

const (
    Bold          StyleFlags = 1 << iota
    Italic
    Underline
    Strikethrough
    Reverse
    Dim
    Blink
)

type Link struct {
    URL  string
    Text string
}

// 预定义的空 Cell (用于清空)
var BlankCell = Cell{Rune: ' ', Width: 1}
var TransparentCell = Cell{Rune: ' ', Width: 0} // 不覆盖下层
```

### Color — 颜色系统

```go
// internal/buffer/color.go

type ColorType uint8

const (
    ColorNone    ColorType = iota // 使用默认终端色
    ColorNamed                     // 16 色命名色
    Color256                       // xterm 256 色
    ColorTrue                      // 24-bit TrueColor
)

type Color struct {
    Type ColorType
    Val  uint32 // TrueColor: 0xRRGGBB | Color256: 0-255 | ColorNamed: 见下
}

// 命名色 (终端标准 16 色)
const (
    Default Color = ...
    Black   Color = ...
    Red     Color = ...
    Green   Color = ...
    ...
    BrightWhite Color = ...
)

// 快捷构造
func RGB(r, g, b uint8) Color  // TrueColor
func X256(n uint8) Color       // 256 色
func Hex(hex string) Color     // "#ff6600" → TrueColor
```

### Buffer — 二维网格

```go
// internal/buffer/buffer.go

type Buffer struct {
    Width  int
    Height int
    cells  []Cell // 长度 = Width * Height
}

func NewBuffer(w, h int) *Buffer

// 基本操作
func (b *Buffer) SetCell(x, y int, cell Cell)
func (b *Buffer) GetCell(x, y int) Cell
func (b *Buffer) Fill(cell Cell)                          // 全填充
func (b *Buffer) FillRect(rect Rect, cell Cell)           // 区域填充
func (b *Buffer) DrawText(x, y int, text string, style Style) int // 返回结束 x

// 从其他 Buffer 拷贝区域
func (b *Buffer) Blit(src *Buffer, srcX, srcY, dstX, dstY, w, h int)

// 区域操作
func (b *Buffer) Sub(rect Rect) *Buffer  // 返回子区域视图 (共享底层数组)

// 按行操作 (用于 markdown 渲染输出)
func (b *Buffer) SetRow(y int, row CellRow, xOffset int)
```

### CellRow — 一行带样式的 cells

```go
// 用于 markdown 渲染器与渲染引擎之间的数据传递
type CellRow struct {
    Cells []Cell
}
```

---

## Layer 1-2: 终端抽象

### Terminal 接口

```go
// internal/term/term.go

type Terminal struct {
    w       io.Writer   // 通常是 *os.File (/dev/tty)
    r       io.Reader
    width   int
    height  int
    profile ColorProfile
    closed  bool
}

// 生命周期
func Open() (*Terminal, error)    // 进入 raw mode + alt screen + 启用鼠标
func (t *Terminal) Close() error  // 恢复终端

// 输出
func (t *Terminal) Write(b []byte) (int, error)
func (t *Terminal) WriteStr(s string)

// 尺寸
func (t *Terminal) Size() (w, h int)

// 能力
func (t *Terminal) ColorProfile() ColorProfile  // None / ANSI16 / X256 / TrueColor
func (t *Terminal) SupportsMouse() bool
```

### 终端初始化序列

Open() 时发送的 escape sequences：

```
\e[?1049h   — 进入 Alt Screen (保存主屏幕状态)
\e[?25l     — 隐藏光标
\e[?2004h   — 启用 Bracketed Paste
\e[?1000h   — 启用鼠标 (基本)
\e[?1006h   — 启用 SGR 鼠标模式 (精确坐标)
\e[?1003h   — 启用全鼠标追踪 (可选, 用于 hover)
```

Close() 时逆序恢复。

### 输入解析器

这是最关键也最复杂的部分。终端输入是一串字节流，需要解析出结构化事件：

```go
// internal/term/input.go

type EventType uint8

const (
    EventKey EventType = iota
    EventMouse
    EventPaste
    EventResize
)

type InputEvent struct {
    Type EventType
    Key  *KeyEvent    // nil if not key
    Mouse *MouseEvent  // nil if not mouse
    Paste string       // paste content
    Size  *Size        // resize
}

type KeyEvent struct {
    Key       KeyCode    // 见下
    Modifiers ModMask    // Ctrl/Alt/Shift
    Rune      rune       // 文本字符 (仅可打印字符)
}

type KeyCode uint16

const (
    KeyUnknown KeyCode = iota
    KeyEnter
    KeyTab
    KeyBacktab
    KeyBackspace
    KeyDelete
    KeyInsert
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyEscape
    KeySpace
    // F1-F12
    KeyF1  KeyCode = 1001 + iota
    // ...
)

type ModMask uint8

const (
    ModCtrl  ModMask = 1 << iota
    ModAlt
    ModShift
)

type MouseEvent struct {
    X, Y       int
    Button     MouseButton    // Left/Right/Middle/WheelUp/WheelDown
    Modifiers  ModMask
    Type       MouseAction    // Down/Up/Move/Drag/Wheel
}

type MouseAction uint8

const (
    MouseDown MouseAction = iota
    MouseUp
    MouseMove
    MouseDrag
    MouseWheel
)

type MouseButton uint8

const (
    MouseLeft MouseButton = iota
    MouseRight
    MouseMiddle
    MouseWheelUp
    MouseWheelDown
)

// 输入解析器 (状态机)
type Parser struct {
    state parseState
    buf   []byte
}

// 喂入原始字节, 返回解析出的事件
func (p *Parser) Feed(data []byte) []InputEvent
```

#### 解析逻辑要点

```
输入字节流示例:
  "abc"           → 3 个 KeyPress 事件
  "\x1b[A"        → KeyUp
  "\x1b[1;5A"     → Ctrl+Up
  "\x1b[<0;42;10M"→ MouseLeft 按下 at (42,10)
  "\x1b[<64;42;10M" → MouseWheelUp at (42,10)
  "\x1b[200~hello\x1b[201~" → Paste "hello"
```

解析器是一个状态机:
- `StateNormal`: 读普通字符, `\x1b` 进入 `StateEscape`
- `StateEscape`: 读到 `[` 进入 `StateCSI`; 读到 `O` 进入 `StateSS3`
- `StateCSI`: 收集参数直到终结符 (`A-Z~`)
- `StatePaste`: 检测 `\x1b[200~` 开始, `\x1b[201~` 结束

### 输出: ANSI Writer

```go
// internal/term/output.go

type Writer struct {
    w       io.Writer
    profile ColorProfile
    buf     bytes.Buffer  // 批量写入减少 syscall
}

// 光标控制
func (w *Writer) MoveTo(x, y int)       // 1-based 坐标
func (w *Writer) MoveToRow(y int)
func (w *Writer) SaveCursor()
func (w *Writer) RestoreCursor()
func (w *Writer) HideCursor()
func (w *Writer) ShowCursor()

// 样式 (批量设置)
func (w *Writer) SetStyle(s Style)
func (w *Writer) ResetStyle()

// 写入
func (w *Writer) WriteRune(r rune)
func (w *Writer) WriteString(s string)
func (w *Writer) Flush() error  // 一次性 flush 到底层 writer

// 滚动 (区域滚动)
func (w *Writer) ScrollUp(n int)
func (w *Writer) SetScrollRegion(top, bottom int)
func (w *Writer) ClearScrollRegion()
```

---

## Layer 3: 渲染引擎

### 双缓冲 Diff Renderer

```go
// render/renderer.go

type Renderer struct {
    term    *term.Terminal
    front   *buffer.Buffer   // 上一次渲染的状态
    back    *buffer.Buffer   // 当前构建的状态
    writer  *term.Writer
    dirty   *DirtyTracker    // 脏区域追踪
}

// 渲染流程:
// 1. 组件树渲染到 back buffer
// 2. 与 front buffer diff
// 3. 输出 diff 的 ANSI 序列
// 4. swap front/back

func (r *Renderer) BeginFrame(w, h int) *buffer.Buffer {
    // 返回 back buffer 供组件树渲染
    // 如果尺寸变化, 重新分配
    return r.back
}

func (r *Renderer) EndFrame() error {
    // diff front vs back
    // 输出变化部分
    // swap
    diffs := buffer.Diff(r.front, r.back)
    r.flushDiffs(diffs)
    r.front, r.back = r.back, r.front
}
```

### Diff 算法

```go
// internal/buffer/diff.go

type DiffOp struct {
    X, Y   int
    Cell   Cell
}

// 逐 cell 比较, 返回需要更新的操作列表
// 优化: 按行批量处理, 跳过完全相同的行
func Diff(front, back *Buffer) []DiffOp

// 输出优化策略:
// 1. 将连续的变化 cell 合并为一行写入
// 2. 相邻行间用 \r\n 或 cursor movement 连接
// 3. 只在样式变化时发送颜色切换序列
```

### Dirty Region 追踪

```go
// render/renderer.go

type DirtyTracker struct {
    regions []Rect  // 脏区域列表
}

func (d *DirtyTracker) MarkDirty(rect Rect)
func (d *DirtyTracker) Dirty() bool
func (d *DirtyTracker) Regions() []Rect
func (d *DirtyTracker) Clear()
```

渲染时只需 diff 脏区域内的 cells，不是全屏。

---

## Layer 3.5: Markdown 渲染器

### 流式 Markdown 处理

这是 AI-native TUI 最独特的部分。传统 markdown 渲染器假设输入是完整的，但 AI 输出是流式的。

```go
// markdown/incremental.go

// IncrementalMarkdown 处理流式 markdown
type IncrementalMarkdown struct {
    raw      strings.Builder  // 累积的原始文本
    parser   goldmark.Markdown
    lastLen  int              // 上次解析的长度

    // 状态跟踪
    inCodeFence  bool
    codeLang     string
    codeFenceSym string        // ``` 或 ~~~
    inTable      bool

    // 缓存
    cachedRows   []CellRow
    cacheValid   bool
}

// AppendDelta 追加流式数据, 返回是否需要重新渲染
func (im *IncrementalMarkdown) AppendDelta(delta string) bool {
    im.raw.WriteString(delta)
    im.cacheValid = false

    // 判断是否需要立即渲染:
    // - 如果在代码块内, 等到更多数据或代码块关闭
    // - 否则, debounce 后渲染
    return im.shouldRender()
}

// Render 将当前累积的 markdown 渲染为 CellRow 列表
func (im *IncrementalMarkdown) Render(width int, theme *MarkdownTheme) []CellRow {
    source := im.raw.String()
    doc := goldmark.Parse(source)

    renderer := &astRenderer{
        width: width,
        theme: theme,
        codeHL: chromaLexer,
    }
    return renderer.Render(doc, source)
}
```

### Markdown AST 渲染器

```go
// markdown/renderer.go

type astRenderer struct {
    width   int
    theme   *MarkdownTheme
    codeHL  *CodeHighlighter
    rows    []CellRow
    links   []LinkInfo  // 收集所有链接用于点击
}

type LinkInfo struct {
    URL    string
    Bounds Rect        // 屏幕坐标
}

func (r *astRenderer) Render(doc ast.Node, source []byte) []CellRow {
    // 遍历 AST, 为每种节点类型生成 CellRow
    ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        switch node := n.(type) {
        case *ast.Heading:
            r.renderHeading(node, source, entering)
        case *ast.Paragraph:
            r.renderParagraph(node, source, entering)
        case *ast.CodeBlock:
            r.renderCodeBlock(node, source, entering)
        case *ast.FencedCodeBlock:
            r.renderFencedCode(node, source, entering)
        case *ast.List:
            r.renderList(node, source, entering)
        case *ast.Table:
            r.renderTable(node, source, entering)
        case *ast.Link:
            r.renderLink(node, source, entering)
        case *ast.Blockquote:
            r.renderBlockquote(node, source, entering)
        case *ast.Emphasis:
            r.renderEmphasis(node, source, entering)
        case *ast.Strikethrough:
            r.renderStrikethrough(node, source, entering)
        }
        return ast.WalkContinue, nil
    })
    return r.rows
}
```

### 渲染样式示例

```
# Heading 1     →  Bold, 前景色=theme.Heading1, 底部空行
## Heading 2    →  Bold, 前景色=theme.Heading2

**bold**        →  Bold flag
*italic*        →  Italic flag
~~strike~~      →  Strikethrough flag
`inline code`   →  Bg=theme.CodeBg, Fg=theme.CodeFg

> quote         →  左侧 │ 竖线, Fg=dim
> quote

- item          →  • 前缀, 悬挂缩进
  - nested      →  ◦ 前缀

1. ordered      →  1. 前缀

| A | B |       →  ┌───┬───┐
|---|---|       →  │ A │ B │
| 1 | 2 |       →  ├───┼───┤
                →  │ 1 │ 2 │
                →  └───┴───┘

```go           →  ╭─ go ─────────── [📋] ─╮
func main(){}   →  │ func main() {}      │
```             →  ╰──────────────────────╯

[link](url)     →  Underline, Fg=blue, 可点击

---             →  ───────────────────── (全宽分隔线)
```

### 代码块高亮

```go
// markdown/codeblock.go

type CodeHighlighter struct {
    lexer   chroma.Lexer
    formatter *cellFormatter  // chroma token → Cell style
    theme   *chroma.Style
}

func NewCodeHighlighter(themeName string) *CodeHighlighter {
    // chroma 内置主题: monokai, dracula, github-dark, etc.
    style := styles.Get(themeName)
    return &CodeHighlighter{
        lexer:    lexers.Get,  // 按语言自动选择
        theme:    style,
    }
}

// 将代码高亮为 CellRow (带边框和复制按钮区域)
func (h *CodeHighlighter) Highlight(code, lang string, width int, theme *MarkdownTheme) CodeBlockRender {
    // 1. 用 chroma lexer 分词
    // 2. 每个 token 映射到 (rune, Style)
    // 3. 构建带边框的 CellRow
    // 4. 注册复制按钮的 hit region
    return CodeBlockRender{
        Rows:        rows,
        CopyRegion:  Rect{...},  // 复制按钮区域
        Language:    lang,
        LineCount:   len(lines),
    }
}
```

### 文本折行 (CJK 安全)

```go
// markdown/width.go

// 将文本按指定宽度折行, 正确处理 CJK 双宽字符
func WrapText(text string, width int) []string {
    // 1. 计算每个 rune 的显示宽度 (wcwidth)
    // 2. 累加宽度, 超过 width 时折行
    // 3. 优先在空格/标点处折行
    // 4. emoji 和 CJK 字符不拆分
}
```

---

## Layer 4: 组件系统

### Component 接口

```go
// component/component.go

type Component interface {
    // 标识
    ID() string

    // 布局: 给定约束, 返回期望尺寸
    Measure(constraints Constraints) Size

    // 布局: 设定最终位置和尺寸
    SetBounds(bounds Rect)

    // 渲染: 将自身绘制到 buffer 的指定区域
    Paint(buf *buffer.Buffer)

    // 事件处理: 返回是否消费了事件
    HandleEvent(event event.Event) bool

    // 交互区域: 返回此组件的可点击区域列表
    HitRegions() []hit.Region

    // 脏标记
    IsDirty() bool
    ClearDirty()
}

type Rect struct {
    X, Y, W, H int
}

type Size struct {
    W, H int
}

type Constraints struct {
    MinWidth, MaxWidth  int
    MinHeight, MaxHeight int
}

func Loose(maxW, maxH int) Constraints  // 0 ~ max
func Tight(w, h int) Constraints         // 固定
func Expand(maxW, maxH int) Constraints  // max ~ max (尽可能大)
```

### 组件树

```go
// component/node.go

// Parent/Child 关系管理
type Node struct {
    component  Component
    parent     *Node
    children   []*Node
    bounds     Rect
}

// 树操作
func (n *Node) AddChild(c Component) *Node
func (n *Node) RemoveChild(id string)
func (n *Node) Find(id string) *Node
func (n *Node) Walk(fn func(*Node))  // 深度优先遍历

// 布局传递
func (n *Node) Layout(constraints Constraints)
```

### 布局: Flex

```go
// component/layout/flex.go

type Direction uint8

const (
    FlexRow    Direction = iota  // 水平排列
    FlexColumn                    // 垂直排列
)

type MainAxisAlignment uint8

const (
    MainStart    MainAxisAlignment = iota
    MainCenter
    MainEnd
    MainSpaceBetween
    MainSpaceEvenly
)

type Flex struct {
    Direction       Direction
    MainAxisAlignment MainAxisAlignment
    Gap             int
    children        []Component
}

func (f *Flex) Measure(c Constraints) Size {
    // 根据方向和子组件计算总尺寸
    // Row: 高度 = max(子高度), 宽度 = sum(子宽度) + gaps
    // Column: 宽度 = max(子宽度), 高度 = sum(子高度) + gaps
}

func (f *Flex) SetBounds(bounds Rect) {
    // 将 bounds 分配给子组件
    // 处理 Spacer (弹性填充)
}
```

### ScrollView

```go
// component/scroll/scroll_view.go

type ScrollView struct {
    content    Component     // 单个子组件 (通常是 BlockContainer)
    offset     int           // 垂直滚动偏移
    autoFollow bool          // 流式更新时自动跟随底部
    height     int           // 可视区高度

    // 滚动条
    showScrollbar bool
}

func (sv *ScrollView) Paint(buf *buffer.Buffer) {
    // 1. 创建一个虚拟 buffer, 让 content 渲染
    // 2. 从 offset 开始截取可视部分
    // 3. Blit 到目标 buffer

    // 或者更高效: 直接设置 content 的 bounds 带偏移
}

func (sv *ScrollView) HandleEvent(e event.Event) bool {
    switch ev := e.(type) {
    case *event.MouseEvent:
        switch ev.Button {
        case event.MouseWheelUp:
            sv.offset = max(0, sv.offset - 3)
            return true
        case event.MouseWheelDown:
            sv.offset += 3
            sv.autoFollow = false  // 用户手动滚动, 取消自动跟随
            return true
        }
    case *event.KeyEvent:
        switch ev.Key {
        case event.KeyPageDown:
            sv.offset += sv.height
            return true
        case event.KeyPageUp:
            sv.offset = max(0, sv.offset - sv.height)
            return true
        }
    }
    return false
}

// 流式更新时调用: 如果 autoFollow, 滚动到底部
func (sv *ScrollView) NotifyContentChanged(contentHeight int) {
    if sv.autoFollow {
        sv.offset = max(0, contentHeight - sv.height)
    }
}
```

---

## Layer 5: Content Block 系统 ⭐

### Block 生命周期

```go
// block/block.go

type BlockState uint8

const (
    BlockPending   BlockState = iota  // 等待数据 (转圈)
    BlockStreaming                     // 正在接收数据
    BlockComplete                      // 数据完成
    BlockError                         // 发生错误
)

func (s BlockState) String() string {
    switch s {
    case BlockPending:   return "pending"
    case BlockStreaming: return "streaming"
    case BlockComplete:  return "complete"
    case BlockError:     return "error"
    }
    return "unknown"
}

// Block 是所有 AI 内容块的接口
type Block interface {
    component.Component  // Block 首先是一个 Component

    // 生命周期
    State() BlockState
    SetState(BlockState)

    // 流式更新 (在事件循环 goroutine 中调用)
    AppendDelta(delta string)   // 追加流式文本
    Complete()                  // 标记完成
    Fail(err error)             // 标记错误

    // 元信息
    CreatedAt() time.Time
    Duration() time.Duration    // 从创建到完成的时间
}
```

### BlockContainer — 管理所有 blocks

```go
// block/container.go

type BlockContainer struct {
    blocks    []Block
    maxWidth  int
    theme     *theme.Theme

    // 布局
    totalHeight int            // 所有 block 的总高度 (含间距)
    blockHeights map[string]int // 每个 block 的高度缓存

    // 回调
    onChange func()             // 内容变化通知 (触发 ScrollView 检查)
}

func (c *BlockContainer) AddBlock(b Block)
func (c *BlockContainer) RemoveBlock(id string)
func (c *BlockContainer) LastBlock() Block
func (c *BlockContainer) Blocks() []Block

// 查找指定坐标下的 block (用于鼠标点击)
func (c *BlockContainer) BlockAt(y int) Block
```

### ThinkingBlock — 可折叠的思考过程

```go
// block/thinking.go

type ThinkingBlock struct {
    id         string
    state      BlockState
    content    strings.Builder
    collapsed  bool          // 默认折叠
    startedAt  time.Time
    endedAt    time.Time
    theme      *theme.Theme
    spinner    *animation.Spinner

    // 渲染缓存
    cachedRows []CellRow
    dirty      bool

    // hit regions
    toggleRegion hit.Region  // 点击折叠/展开的区域
}

func NewThinkingBlock(id string) *ThinkingBlock {
    return &ThinkingBlock{
        id:        id,
        state:     BlockStreaming,
        collapsed: true,          // 默认折叠, 用户可展开查看
        startedAt: time.Now(),
        spinner:   animation.NewSpinner(animation.Dots),
    }
}

func (b *ThinkingBlock) AppendDelta(delta string) {
    b.content.WriteString(delta)
    b.dirty = true
}

func (b *ThinkingBlock) Complete() {
    b.state = BlockComplete
    b.endedAt = time.Now()
    b.dirty = true
}

func (b *ThinkingBlock) Measure(c Constraints) Size {
    if b.collapsed {
        return Size{W: c.MaxWidth, H: 1}  // 折叠时只占 1 行
    }
    // 展开时: 标题行 + content 行数
    rows := b.renderContent(c.MaxWidth)
    return Size{W: c.MaxWidth, H: 1 + len(rows)}
}

func (b *ThinkingBlock) Paint(buf *buffer.Buffer) {
    // 折叠状态:
    //   ▸ 💭 Thinking... (2.3s)          [dim]
    //   或
    //   ▸ 💭 Thought for 2.3s            [dim, complete]

    // 展开状态:
    //   ▼ 💭 Thinking... (2.3s)          [标题行]
    //   ┌─────────────────────────────┐
    //   │ I need to analyze the user's│  [dim italic]
    //   │ request and determine...    │
    //   └─────────────────────────────┘
}

func (b *ThinkingBlock) HandleEvent(e event.Event) bool {
    // 点击 toggleRegion → collapsed = !collapsed
    if me, ok := e.(*event.MouseEvent); ok {
        if me.Type == event.MouseDown && b.toggleRegion.Contains(me.X, me.Y) {
            b.collapsed = !b.collapsed
            b.dirty = true
            return true
        }
    }
    return false
}
```

### ToolCallBlock — 工具调用

```go
// block/tool_call.go

type ToolCallBlock struct {
    id        string
    state     BlockState
    toolName  string
    args      map[string]any
    rawArgs   string            // 原始参数 (用于显示)
    startedAt time.Time
    endedAt   time.Time
    spinner   *animation.Spinner
    result    *ToolResultBlock  // 关联的结果 block
    theme     *theme.Theme
    dirty     bool
}

func (b *ToolCallBlock) Paint(buf *buffer.Buffer) {
    // Pending/Streaming:
    //   ⏺ ReadFile(src/main.go) ⠋          [spinner 转圈]
    //
    // Complete:
    //   ⏺ ReadFile(src/main.go) ✓ 2.1KB     [green ✓]
    //
    // Error:
    //   ⏺ ReadFile(src/main.go) ✗ not found  [red ✗]

    // 如果 args 复杂, 可以点击展开查看完整参数:
    //   ⏺ ▸ SearchFiles(pattern="*.go", maxResults=10) ⠋
}

func (b *ToolCallBlock) HandleEvent(e event.Event) bool {
    // 点击可以展开查看完整 args
    return false
}
```

### ToolResultBlock — 工具结果

```go
// block/tool_result.go

type ToolResultBlock struct {
    id         string
    state      BlockState
    toolCallID string         // 关联的 ToolCall
    output     strings.Builder
    truncated  bool
    collapsed  bool           // 长结果默认折叠
    maxPreview int            // 折叠时预览行数 (默认 5)
    theme      *theme.Theme
    dirty      bool
}

func (b *ToolResultBlock) Paint(buf *buffer.Buffer) {
    // 折叠状态 (短结果或默认预览):
    //   ╭─ Result ─────────────── 5 lines ─╮
    //   │ 42  func main() {                │
    //   │ 43      fmt.Println("hello")     │
    //   │ 44  }                            │
    //   ╰──────────────────────────────────╯
    //
    // 展开状态 (完整结果, 带滚动):
    //   ╭─ Result ───────────── 142 lines ─╮
    //   │ ... (全部内容)                    │
    //   ╰──────────────────────────────────╯
    //
    // 点击右上角可展开/折叠

    // 如果是 diff 输出, 渲染为带颜色的 diff:
    //   ╭─ Diff ─────────────────────────────╮
    //   │ @@ -10,3 +10,4 @@                    │
    //   │   unchanged line                     │
    //   │ - removed line              [red bg] │
    //   │ + added line                [green]  │
    //   ╰─────────────────────────────────────╯
}

func (b *ToolResultBlock) AppendDelta(delta string) {
    b.output.WriteString(delta)
    b.dirty = true
}

func (b *ToolResultBlock) Complete() {
    b.state = BlockComplete
    // 如果输出超过 maxPreview 行, 自动折叠
    if lineCount(b.output.String()) > b.maxPreview {
        b.collapsed = true
    }
    b.dirty = true
}
```

### AssistantTextBlock — 流式 Markdown 文本

```go
// block/assistant_text.go

type AssistantTextBlock struct {
    id       string
    state    BlockState
    md       *markdown.IncrementalMarkdown  // 流式 markdown 处理器
    theme    *theme.Theme
    width    int
    dirty    bool

    // 渲染缓存
    cachedRows []markdown.CellRow
    cachedWidth int
    links      []markdown.LinkInfo
}

func NewAssistantTextBlock(id string, theme *theme.Theme) *AssistantTextBlock {
    return &AssistantTextBlock{
        id:    id,
        state: BlockStreaming,
        md:    markdown.NewIncrementalMarkdown(),
        theme: theme,
    }
}

func (b *AssistantTextBlock) AppendDelta(delta string) {
    needsRender := b.md.AppendDelta(delta)
    if needsRender {
        b.dirty = true
    }
}

func (b *AssistantTextBlock) Paint(buf *buffer.Buffer) {
    if b.dirty || b.cachedWidth != b.width {
        b.cachedRows = b.md.Render(b.width, &b.theme.Markdown)
        b.cachedWidth = b.width
        b.dirty = false
    }
    // 将 cachedRows 写入 buffer
    for y, row := range b.cachedRows {
        buf.SetRow(b.bounds.Y+y, row, b.bounds.X)
    }
}

func (b *AssistantTextBlock) HitRegions() []hit.Region {
    // 返回所有链接的可点击区域
    var regions []hit.Region
    for _, link := range b.links {
        regions = append(regions, hit.Region{
            Bounds: link.Bounds,
            Action: hit.ActionOpenURL(link.URL),
        })
    }
    return regions
}
```

### UserMessageBlock — 用户消息

```go
// block/user_message.go

type UserMessageBlock struct {
    id      string
    content string
    theme   *theme.Theme
    dirty   bool
}

func (b *UserMessageBlock) Paint(buf *buffer.Buffer) {
    // 用户消息样式: 左侧带彩色竖线 + 背景色
    //
    // ┃ What is the capital of France?    [theme.UserMsgFg on theme.UserMsgBg]
    // ┃
    //
    // 或更优雅的设计: 带圆角的卡片
    // ╭─ You ──────────────────────╮
    // │ What is the capital of...? │
    // ╰────────────────────────────╯
}
```

### CodeBlock — 独立代码块

```go
// block/code.go

type CodeBlock struct {
    id        string
    language  string
    code      strings.Builder
    state     BlockState
    highlight *markdown.CodeHighlighter
    theme     *theme.Theme
    copied    bool           // 点击复制后短暂显示
    copiedAt  time.Time
    dirty     bool
}

func (b *CodeBlock) Paint(buf *buffer.Buffer) {
    // ╭─ go ──────────────────── [📋 Copy] ─╮
    // │ package main                          │
    // │                                       │
    // │ func main() {                         │
    // │     fmt.Println("Hello")              │
    // │ }                                     │
    // ╰───────────────────────────────────────╯
    //
    // 点击 [📋 Copy] 后变为 [✓ Copied!] 持续 2 秒
}

func (b *CodeBlock) HitRegions() []hit.Region {
    // 返回复制按钮的区域
    return []hit.Region{{
        Bounds: b.copyButtonBounds(),
        Action: hit.ActionCopy(b.code.String()),
        OnClick: func() {
            b.copied = true
            b.copiedAt = time.Now()
            b.dirty = true
        },
    }}
}
```

### ErrorBlock

```go
// block/error.go

type ErrorBlock struct {
    id      string
    err     error
    theme   *theme.Theme
}

func (b *ErrorBlock) Paint(buf *buffer.Buffer) {
    // ⚠ Error: connection refused
    //   Retry? [Y/n]
}
```

---

## 流式分发器

```go
// block/stream.go

// StreamDelta 描述一次流式更新的内容
type StreamDelta struct {
    Type      StreamDeltaType
    BlockID   string           // 目标 block ID
    BlockType BlockType        // block 类型 (仅在创建新 block 时有效)
    Content   string           // delta 内容
    Metadata  map[string]any   // 工具名、参数等
}

type StreamDeltaType uint8

const (
    DeltaStart    StreamDeltaType = iota  // 新 block 开始
    DeltaAppend                            // 追加内容
    DeltaComplete                          // block 完成
    DeltaError                             // block 出错
)

type BlockType uint8

const (
    BlockThinking     BlockType = iota
    BlockAssistantText
    BlockToolCall
    BlockToolResult
    BlockCode
    BlockError
    BlockUserMessage
)

// StreamDispatcher 将 AI delta 路由到正确的 block
type StreamDispatcher struct {
    container *BlockContainer
    factory   *BlockFactory
    current   map[BlockType]Block  // 当前活跃的每种类型的 block
}

func (d *StreamDispatcher) HandleDelta(delta StreamDelta) {
    switch delta.Type {
    case DeltaStart:
        // 创建新 block
        block := d.factory.Create(delta.BlockType, delta.BlockID)
        d.container.AddBlock(block)
        d.current[delta.BlockType] = block

    case DeltaAppend:
        // 路由到现有 block
        block := d.findBlock(delta.BlockID, delta.BlockType)
        block.AppendDelta(delta.Content)

    case DeltaComplete:
        block := d.findBlock(delta.BlockID, delta.BlockType)
        block.Complete()
        delete(d.current, delta.BlockType)

    case DeltaError:
        block := d.findBlock(delta.BlockID, delta.BlockType)
        block.Fail(fmt.Errorf("%v", delta.Metadata["error"]))
   }
}

// 智能路由: 如果没有指定 BlockID, 根据 BlockType 路由到当前活跃的 block
func (d *StreamDispatcher) findBlock(id string, bt BlockType) Block {
    if id != "" {
        // 直接查找
        for _, b := range d.container.Blocks() {
            if b.ID() == id {
                return b
            }
        }
    }
    // fallback: 返回当前活跃的该类型 block
    return d.current[bt]
}
```

---

## Layer 4.5: 鼠标命中测试

```go
// hit/region.go

type Region struct {
    ID     string
    Bounds Rect
    Action Action
    Cursor CursorStyle  // 鼠标悬停时的光标样式
    BlockID string       // 所属 block (用于调试)
}

type CursorStyle uint8

const (
    CursorDefault CursorStyle = iota
    CursorPointer
    CursorText
)

// Action 描述点击后的行为
type Action struct {
    Type   ActionType
    URL    string         // OpenURL
    Text   string         // Copy
    Fn     func()         // Custom
}

type ActionType uint8

const (
    ActionToggle    ActionType = iota  // 折叠/展开
    ActionOpenURL                       // 打开链接
    ActionCopy                          // 复制到剪贴板
    ActionCustom                        // 自定义回调
)
```

```go
// hit/tree.go

// RegionTree 用于高效命中测试
// 每帧渲染后, 从组件树收集所有 HitRegion, 构建 RegionTree
type RegionTree struct {
    regions []Region
}

func NewRegionTree() *RegionTree

func (t *RegionTree) Add(r Region)
func (t *RegionTree) Clear()

// 查找指定坐标下最上层的 region
func (t *RegionTree) Hit(x, y int) *Region

// 查找指定区域内的所有 region
func (t *RegionTree) Query(rect Rect) []Region
```

---

## 事件系统

### 事件类型

```go
// event/event.go

type Event interface {
    Type() EventType
    Timestamp() time.Time
}

// 具体事件类型
type KeyEvent    struct { ... }   // 键盘
type MouseEvent  struct { ... }   // 鼠标
type PasteEvent  struct { ... }   // 粘贴
type ResizeEvent struct { ... }   // 窗口尺寸变化
type StreamEvent struct { ... }   // AI 流式数据
type CustomEvent struct { ... }   // 自定义事件
```

### 事件分发

```go
// event/dispatch.go

// 事件分发采用 capture → target → bubble 模式:
// 1. Capture: 从根组件向目标组件传递 (拦截机会)
// 2. Target:  目标组件处理
// 3. Bubble:  从目标向根传递 (冒泡)

type Dispatcher struct {
    root      *component.Node
    hitTree   *hit.RegionTree
    focused   Component        // 当前焦点组件
    mouseOver *hit.Region      // 当前鼠标悬停的区域
}

func (d *Dispatcher) Dispatch(e Event) {
    // 1. 鼠标事件 → hit test → 找到目标 → dispatch
    // 2. 键盘事件 → 发送到 focused 组件
    // 3. 其他事件 → 广播
}
```

### 主事件循环

```go
// event/loop.go

type Loop struct {
    terminal   *term.Terminal
    parser     *term.Parser
    renderer   *render.Renderer
    root       *component.Node
    dispatcher *Dispatcher
    hitTree    *hit.RegionTree

    // 事件 channels
    eventCh    chan Event       // 终端事件 (键盘/鼠标/粘贴/resize)
    streamCh   chan StreamDelta // AI 流式数据
    renderTick <-chan time.Time // 渲染节拍 (60fps)

    // 状态
    running    bool
    dirty      bool             // 是否需要重新渲染
}

func (l *Loop) Run() error {
    // 启动输入读取 goroutine
    go l.readInput()

    // 主循环
    for l.running {
        select {
        case ev := <-l.eventCh:
            l.handleEvent(ev)

        case delta := <-l.streamCh:
            l.handleStream(delta)

        case <-l.renderTick:
            if l.dirty {
                l.render()
            }

        case <-l.quitCh:
            l.running = false
        }
    }
    return nil
}

func (l *Loop) handleEvent(ev Event) {
    // 分发事件到组件树
    l.dispatcher.Dispatch(ev)
    // 如果有组件处理了事件并标记 dirty
    if l.dispatcher.AnyDirty() {
        l.dirty = true
    }
}

func (l *Loop) handleStream(delta StreamDelta) {
    // 路由到 block
    l.streamDispatcher.HandleDelta(delta)
    l.dirty = true
}

func (l *Loop) render() {
    buf := l.renderer.BeginFrame(l.width, l.height)

    // 1. 布局: 如果尺寸变化或有新组件
    l.root.Layout(Expand(l.width, l.height))

    // 2. 渲染组件树到 buffer
    l.root.Paint(buf)

    // 3. 收集 hit regions
    l.hitTree.Clear()
    l.collectHitRegions(l.root)

    // 4. Diff + flush
    l.renderer.EndFrame()

    l.dirty = false
}
```

### 输入读取 (独立 goroutine)

```go
func (l *Loop) readInput() {
    buf := make([]byte, 4096)
    for l.running {
        n, err := l.terminal.Read(buf)
        if err != nil {
            l.eventCh <- &ErrorEvent{Err: err}
            return
        }

        // 解析原始字节为事件
        events := l.parser.Feed(buf[:n])
        for _, ev := range events {
            l.eventCh <- ev
        }
    }
}
```

### 渲染节拍

```go
// 渲染 tick 不是固定 60fps, 而是智能调度:
// - 有 dirty 标记时, 最多等待 16ms 就渲染
// - 无 dirty 时, 阻塞在 select (不消耗 CPU)
// - 用户输入/流式数据到达时, 标记 dirty, 下一个 tick 渲染

func (l *Loop) Run() error {
    for l.running {
        // 重置 renderTick
        l.renderTick = time.After(16 * time.Millisecond)

        select {
        case ev := <-l.eventCh:
            l.handleEvent(ev)
            if l.dirty {
                // 输入事件后可以立即渲染 (不等 tick), 保证低延迟
                l.render()
            }

        case delta := <-l.streamCh:
            l.handleStream(delta)
            // 流式数据通过 tick 渲染 (debounce 16ms, 避免频繁重绘)

        case <-l.renderTick:
            if l.dirty {
                l.render()
            }
        }
    }
}
```

---

## 动画系统

```go
// animation/animation.go

type Animation interface {
    Update(delta time.Duration) bool  // 返回是否完成
    Done() <-chan struct{}
}

type Manager struct {
    animations []Animation
    ticker     *time.Ticker
    onDirty    func()  // 动画更新时触发重绘
}

// --- Spinner ---

type Spinner struct {
    frames  []string
    current int
    interval time.Duration
    lastTick time.Time
}

var SpinnerFrames = map[string][]string{
    "dots":    {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
    "arc":     {"◜", "◠", "◝", "◞", "◡", "◟"},
    "arrow":   {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
    "bouncing":{"⠁", "⠂", "⠄", "⠂"},
}

func (s *Spinner) Current() string {
    return s.frames[s.current]
}

// --- Fade In (新 block 出现) ---

type FadeIn struct {
    progress  float64   // 0 → 1
    duration  time.Duration
}

func (f *FadeIn) Style(base Style) Style {
    // 根据 progress 调整透明度 (dim → normal)
    if f.progress < 1.0 {
        base.Flags |= Dim
    }
    return base
}
```

---

## 主题系统

```go
// theme/theme.go

type Theme struct {
    Name string

    // 基础色
    Background Color
    Foreground Color

    // Block 专用色
    Thinking     ThinkingTheme
    ToolCall     ToolCallTheme
    ToolResult   ToolResultTheme
    Assistant    AssistantTheme
    UserMessage  UserMessageTheme
    CodeBlock    CodeBlockTheme
    Error        ErrorTheme

    // Markdown 专用色
    Markdown     MarkdownTheme

    // 边框样式
    Border       BorderTheme
}

type ThinkingTheme struct {
    Icon     string  // "💭" 或 "🧠"
    Fg       Color   // 内容前景色 (通常 dim)
    CollapsedIcon string  // "▸"
    ExpandedIcon  string  // "▾"
}

type ToolCallTheme struct {
    PendingIcon  string  // "⏺"
    SuccessIcon  string  // "✓"
    ErrorIcon    string  // "✗"
    PendingFg    Color   // yellow
    SuccessFg    Color   // green
    ErrorFg      Color   // red
    ArgFg        Color   // dim
}

type MarkdownTheme struct {
    H1          Color
    H2          Color
    H3          Color
    H4          Color
    H5          Color
    H6          Color
    Bold        Color
    Italic      Color
    Strike      Color
    CodeFg      Color
    CodeBg      Color
    LinkFg      Color
    LinkUrlFg   Color
    QuoteFg     Color
    QuoteBar    Color
    ListBullet  Color
    TableBorder Color
    TableHeader Color
    Hr          Color
}
```

### 内置主题

```go
// theme/builtin.go

var DarkTheme = &Theme{
    Name:       "dark",
    Background: Hex("#1a1b26"),
    Foreground: Hex("#a9b1d6"),
    // ...
}

var TokyoNight = &Theme{
    // 基于 tokyonight 色板
}

var Catppuccin = &Theme{
    // 基于 catppuccin mocha
}

var Gruvbox = &Theme{
    // 基于 gruvbox dark
}

var LightTheme = &Theme{
    // 浅色主题
}
```

---

## Layer 6: Application API

### App 入口

```go
// app.go

type App struct {
    terminal   *term.Terminal
    renderer   *render.Renderer
    root       *component.Node
    loop       *event.Loop
    theme      *theme.Theme
    dispatcher *StreamDispatcher

    // 布局
    scrollView *scroll.ScrollView
    container  *block.BlockContainer
    inputBar   *input.TextInput
}

func New(opts ...Option) (*App, error) {
    app := &App{
        theme: theme.DarkTheme,
    }
    for _, opt := range opts {
        opt(app)
    }

    // 初始化终端
    app.terminal, _ = term.Open()

    // 构建组件树:
    //
    // Root (Flex Column)
    //   ├── ScrollView (填充剩余空间)
    //   │   └── BlockContainer
    //   └── InputBar (固定高度)

    app.container = block.NewBlockContainer(app.theme)
    app.scrollView = scroll.NewScrollView(app.container)
    app.scrollView.AutoFollow = true

    app.inputBar = input.NewTextInput()

    root := layout.NewFlex(layout.FlexColumn, 0)
    root.AddChild(app.scrollView)  // flex-grow
    root.AddChild(layout.NewSpacer(1))  // 分隔线
    root.AddChild(app.inputBar)    // 固定高度

    app.root = component.NewNode(root)
    return app, nil
}

type Option func(*App)

func WithTheme(t *theme.Theme) Option {
    return func(a *App) { a.theme = t }
}
```

### 运行 + 输入处理

```go
func (a *App) Run() error {
    return a.loop.Run()
}

// 用户输入提交时的回调
func (a *App) OnSubmit(fn func(text string)) {
    a.inputBar.OnSubmit(func(text string) {
        // 自动添加 UserMessageBlock
        a.container.AddBlock(block.NewUserMessageBlock(text, a.theme))
        // 触发用户回调
        fn(text)
    })
}
```

### 便捷的 AI 集成 API

```go
// 用户侧典型用法

func main() {
    app, _ := fluui.New(
        fluui.WithTheme(theme.TokyoNight),
    )

    app.OnSubmit(func(text string) {
        // 用户发送消息, 开始 AI 响应
        go func() {
            resp := aiClient.StreamChat(ctx, text)

            for {
                delta, err := resp.Recv()
                if err == io.EOF {
                    break
                }

                app.SendDelta(fluui.StreamDelta{
                    Type:    fluui.DeltaAppend,
                    Content: delta.Content,
                })
            }
        }()
    })

    app.Run()
}
```

### 更完整的 AI 集成示例

```go
func main() {
    app, _ := fluui.New()

    app.OnSubmit(func(text string) {
        go handleAIResponse(app, text)
    })

    app.Run()
}

func handleAIResponse(app *fluui.App, userText string) {
    ctx := context.Background()
    stream, _ := aiClient.ChatStream(ctx, userText)

    for {
        event, err := stream.Recv()
        if err != nil {
            app.SendDelta(block.StreamDelta{
                Type:  block.DeltaError,
                Metadata: map[string]any{"error": err.Error()},
            })
            return
        }

        switch event.Type {

        case ai.EventThinking:
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaStart,
                BlockType: block.BlockThinking,
                BlockID:   event.ID,
            })
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaAppend,
                BlockType: block.BlockThinking,
                BlockID:   event.ID,
                Content:   event.Content,
            })

        case ai.EventText:
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaStart,
                BlockType: block.BlockAssistantText,
                BlockID:   event.ID,
            })
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaAppend,
                BlockType: block.BlockAssistantText,
                BlockID:   event.ID,
                Content:   event.Content,
            })

        case ai.EventToolCall:
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaStart,
                BlockType: block.BlockToolCall,
                BlockID:   event.ID,
                Metadata: map[string]any{
                    "tool": event.ToolName,
                    "args": event.ToolArgs,
                },
            })

        case ai.EventToolResult:
            app.SendDelta(block.StreamDelta{
                Type:      block.DeltaStart,
                BlockType: block.BlockToolResult,
                BlockID:   event.ID,
                Content:   event.Output,
            })

        case ai.EventDone:
            // 完成所有活跃的 block
            app.SendDelta(block.StreamDelta{
                Type: block.DeltaComplete,
            })
            return
        }
    }
}
```

---

## Overlay / Modal 层 (Layer 4.5)

Overlay 层允许在主内容之上叠加组件, 支持 Modal 对话框、全屏代码查看器、链接预览弹出窗等。

### 架构设计

```go
// overlay/overlay.go

// Overlay 是一个可层叠的覆盖层
type Overlay interface {
    component.Component
    
    Z() int                 // 层级 (越大越在上)
    Modal() bool            // 是否模态 (阻止下层事件)
    Animation() animation.Animation  // 出现/消失动画
    
    Show()
    Hide()
    IsVisible() bool
}

// OverlayManager 管理所有活跃的 overlay
type OverlayManager struct {
    layers   []Overlay       // 按 z-index 排序
    dirty    bool
    theme    *theme.Theme
}

func (m *OverlayManager) Add(o Overlay)
func (m *OverlayManager) Remove(id string)
func (m *OverlayManager) Top() Overlay           // 最上层
func (m *OverlayManager) HitTest(x, y int) *Overlay  // 命中测试

// 渲染: 按层级从低到高依次绘制
func (m *OverlayManager) Paint(buf *buffer.Buffer) {
    for _, o := range m.layers {
        if !o.IsVisible() { continue }
        o.Paint(buf)
    }
}

// 事件分发: 从最上层开始
func (m *OverlayManager) HandleEvent(e event.Event) bool {
    for i := len(m.layers) - 1; i >= 0; i-- {
        o := m.layers[i]
        if !o.IsVisible() { continue }
        if o.HandleEvent(e) { return true }
        if o.Modal() { return true }  // 模态层吃掉事件
    }
    return false
}
```

### Modal 对话框

```go
// overlay/modal.go

type Modal struct {
    id       string
    title    string
    content  component.Component
    actions  []ModalAction    // 按钮列表
    bounds   Rect             // 居中计算
    z        int              // 默认 100
    visible  bool
    fade     *animation.FadeIn
}

type ModalAction struct {
    Label   string
    Style   Style          // Primary / Secondary / Danger
    OnClick func() bool     // 返回 true 关闭 modal
}

// 渲染效果:
//
//   ████████████████████████████████  ← 半透明遮罩 (dim 背景)
//   ██                            ██
//   ██  ╭─ Confirm ─────────────╮ ██
//   ██  │                       │ ██
//   ██  │  Delete this file?    │ ██
//   ██  │                       │ ██
//   ██  │  [Cancel]  [Delete]   │ ██
//   ██  ╰───────────────────────╯ ██
//   ██                            ██
//   ████████████████████████████████

func NewModal(title string, content component.Component, actions []ModalAction) *Modal
```

### Popup (全屏代码查看器)

```go
// overlay/popup.go

// Popup 用于点击代码块时展开为全屏查看
type Popup struct {
    id       string
    title    string
    content  component.Component  // 通常是 ScrollView + CodeBlock
    bounds   Rect                 // 全屏或近全屏
    z        int                  // 默认 50
    visible  bool
    closeKey KeyCode              // 按 Esc 或 q 关闭
}

// 渲染效果:
//
//   ╭─ main.go ─────────────────────────── [Esc close] ─╮
//   │  1  package main                                   │
//   │  2                                                 │
//   │  3  import "fmt"                                   │
//   │  4                                                 │
//   │  5  func main() {                                  │
//   │  6      fmt.Println("Hello, World!")               │
//   │  7  }                                              │
//   ╰────────────────────────────────────────────────────╯

func NewCodePopup(title, code, lang string, theme *theme.Theme) *Popup
func NewLinkPopup(url, content string, theme *theme.Theme) *Popup
```

### 与主渲染管线的关系

```
渲染管线更新:

  root.Paint(buffer)           ← 主组件树渲染
  overlayManager.Paint(buffer)  ← overlay 层叠加渲染
  collectHitRegions()           ← 收集所有层 (含 overlay) 的 hit regions
  renderer.EndFrame()           ← diff + flush

事件分发更新:

  if overlayManager.HandleEvent(e) {
      // overlay 层处理了事件, 不传递到主组件树
  } else {
      dispatcher.Dispatch(e)    // 主组件树处理
  }
```

### 使用示例

```go
// 点击代码块时弹出全屏查看器
codeBlock.OnClick(func(code, lang string) {
    popup := overlay.NewCodePopup("main.go", code, lang, app.Theme())
    app.Overlay().Add(popup)
})

// 显示确认对话框
modal := overlay.NewModal(
    "Clear History",
    component.NewText("This will remove all messages. Continue?"),
    []overlay.ModalAction{
        {Label: "Cancel", Style: overlay.Secondary, OnClick: func() bool { return true }},
        {Label: "Clear", Style: overlay.Danger, OnClick: func() bool {
            app.Container().Clear()
            return true
        }},
    },
)
app.Overlay().Add(modal)
```

---

## ToolResultBlock 的 Diff 自动检测

ToolResultBlock 在渲染前会自动检测输出是否为 diff 格式, 如果是则用红绿高亮渲染。

```go
// block/diffdetect.go

// DetectDiffFormat 检测文本是否为 unified diff 格式
type DiffFormat int

const (
    DiffNone    DiffFormat = iota  // 不是 diff
    DiffUnified                      // unified diff (git diff)
    DiffGit                          // git diff with file headers
)

func DetectDiff(text string) DiffFormat {
    // 检测规则:
    // 1. 以 "diff --git" 开头 → DiffGit
    // 2. 包含 "@@ ... @@" hunk 头 → DiffUnified
    // 3. 行首高频出现 + / - 前缀 → DiffUnified
    // 防误判: 要求至少有 3 行匹配 diff 模式
}

// RenderDiff 将 diff 文本渲染为带颜色的 CellRow
type DiffRenderer struct {
    theme *theme.Theme
}

func (r *DiffRenderer) Render(diff string, width int) []CellRow {
    for _, line := range strings.Split(diff, "\n") {
        switch {
        case strings.HasPrefix(line, "diff --git"):
            // 文件头: 蓝色粗体
        case strings.HasPrefix(line, "@@"):
            // hunk 头: 青色
        case strings.HasPrefix(line, "+"):
            // 添加行: 绿色背景
        case strings.HasPrefix(line, "-"):
            // 删除行: 红色背景
        case strings.HasPrefix(line, "index "), strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
            // 元信息行: dim
        default:
            // 上下文行: 正常
        }
    }
}
```

ToolResultBlock 在 Paint 时检查:

```go
func (b *ToolResultBlock) Paint(buf *buffer.Buffer) {
    if b.diffFormat == DiffNone {
        b.diffFormat = DetectDiff(b.output.String())
    }
    
    if b.diffFormat != DiffNone {
        // 用 DiffRenderer 渲染
        rows := b.diffRenderer.Render(b.output.String(), b.width)
        // ... 写入 buffer
    } else {
        // 用普通文本渲染
        // ... 写入 buffer
    }
}
```

---

## 输入解析器测试策略

### 策略 A: 录制 + 回放

使用真实终端录制原始字节流作为测试 fixture:

```go
// internal/term/input_test.go

// 测试数据存放在 testdata/ 目录
//   testdata/
//     input_arrow_keys.txt          → \x1b[A\x1b[B\x1b[C\x1b[D
//     input_ctrl_combo.txt          → \x01\x02\x03 ... (Ctrl+A/B/C)
//     input_mouse_click.txt         → \x1b[<0;42;10M
//     input_mouse_drag.txt          → \x1b[<32;42;10M...
//     input_bracketed_paste.txt     → \x1b[200~hello\x1b[201~
//     input_shift_tab.txt           → \x1b[Z
//     input_fn_keys.txt             → \x1bOP (F1) \x1b[15~ (F5) ...
//     input_tmux_wrapped.txt        → nested escape sequences
//     input_rapid_typing.txt        → 多键合并到一个 TCP 包

func TestParseArrowKeys(t *testing.T) {
    raw := readTestdata(t, "input_arrow_keys.txt")
    p := NewParser()
    events := p.Feed(raw)
    
    require.Len(t, events, 4)
    assert.Equal(t, KeyUp, events[0].Key.Key)
    assert.Equal(t, KeyDown, events[1].Key.Key)
    assert.Equal(t, KeyRight, events[2].Key.Key)
    assert.Equal(t, KeyLeft, events[3].Key.Key)
}

func TestParseMouseClick(t *testing.T) {
    raw := readTestdata(t, "input_mouse_click.txt")
    p := NewParser()
    events := p.Feed(raw)
    
    require.Len(t, events, 1)
    assert.Equal(t, MouseLeft, events[0].Mouse.Button)
    assert.Equal(t, 42, events[0].Mouse.X)
    assert.Equal(t, 10, events[0].Mouse.Y)
}

// 录制工具: 提供一个 CLI 命令录制终端输出
// go run ./cmd/record input_arrow_keys.txt
// 然后用户在终端按相应键, 录制为原始字节
```

### 策略 B: 规范覆盖

基于 ECMA-48 / xterm 规范构建系统化的覆盖矩阵:

```go
func TestParseCoverageMatrix(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     InputEvent
    }{
        // 单键
        {"a", "a", KeyEvent{Rune: 'a'}},
        {"Enter", "\r", KeyEvent{Key: KeyEnter}},
        {"Tab", "\t", KeyEvent{Key: KeyTab}},
        {"Backspace", "\x7f", KeyEvent{Key: KeyBackspace}},
        {"Esc", "\x1b", KeyEvent{Key: KeyEscape}},
        
        // Ctrl 组合 (Ctrl+A = 0x01 ... Ctrl+Z = 0x1A)
        {"Ctrl+A", "\x01", KeyEvent{Key: KeyA, Mod: ModCtrl}},
        {"Ctrl+C", "\x03", KeyEvent{Key: KeyC, Mod: ModCtrl}},
        {"Ctrl+Z", "\x1a", KeyEvent{Key: KeyZ, Mod: ModCtrl}},
        
        // Alt 组合 (Esc + char)
        {"Alt+a", "\x1ba", KeyEvent{Rune: 'a', Mod: ModAlt}},
        
        // 方向键
        {"Up", "\x1b[A", KeyEvent{Key: KeyUp}},
        {"Down", "\x1b[B", KeyEvent{Key: KeyDown}},
        {"Right", "\x1b[C", KeyEvent{Key: KeyRight}},
        {"Left", "\x1b[D", KeyEvent{Key: KeyLeft}},
        
        // 修饰键 + 方向键
        {"Shift+Up", "\x1b[1;2A", KeyEvent{Key: KeyUp, Mod: ModShift}},
        {"Ctrl+Up", "\x1b[1;5A", KeyEvent{Key: KeyUp, Mod: ModCtrl}},
        {"Alt+Up", "\x1b[1;3A", KeyEvent{Key: KeyUp, Mod: ModAlt}},
        
        // 功能键
        {"Home", "\x1b[H", KeyEvent{Key: KeyHome}},
        {"End", "\x1b[F", KeyEvent{Key: KeyEnd}},
        {"PageUp", "\x1b[5~", KeyEvent{Key: KeyPageUp}},
        {"PageDown", "\x1b[6~", KeyEvent{Key: KeyPageDown}},
        {"Delete", "\x1b[3~", KeyEvent{Key: KeyDelete}},
        
        // F 键
        {"F1", "\x1bOP", KeyEvent{Key: KeyF1}},
        {"F5", "\x1b[15~", KeyEvent{Key: KeyF5}},
        
        // Shift+Tab
        {"BackTab", "\x1b[Z", KeyEvent{Key: KeyBacktab}},
        
        // 鼠标 SGR
        {"MouseLeftDown", "\x1b[<0;10;5M", MouseEvent{X:10, Y:5, Button: MouseLeft, Type: MouseDown}},
        {"MouseLeftUp", "\x1b[<0;10;5m", MouseEvent{X:10, Y:5, Button: MouseLeft, Type: MouseUp}},
        {"MouseWheelUp", "\x1b[<64;10;5M", MouseEvent{X:10, Y:5, Button: MouseWheelUp}},
        
        // 粘贴
        {"Paste", "\x1b[200~hello world\x1b[201~", PasteEvent{Text: "hello world"}},
        
        // CJK
        {"Chinese", "你好", KeyEvent{Rune: '你'}},
        {"Emoji", "🎉", KeyEvent{Rune: '🎉'}},
        
        // 分片到达 (一个序列分两次到达)
        {"SplitEscape", feedInParts("\x1b[A"), KeyEvent{Key: KeyUp}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser()
            events := p.Feed([]byte(tt.input))
            require.Len(t, events, 1)
            assert.Equal(t, tt.want, events[0])
        })
    }
}
```

---

## 渲染管线路径 (完整流程)

```
用户/AI 产生数据
  │
  ▼
App.SendDelta(StreamDelta)
  │  (通过 channel 发送到事件循环)
  ▼
Loop.handleStream(delta)
  │  (在事件循环 goroutine 中)
  ▼
StreamDispatcher.HandleDelta(delta)
  │  (路由到对应的 Block)
  ▼
Block.AppendDelta(content)
  │  (Block 标记 dirty = true)
  ▼
  ─── 等待下一个 render tick (≤16ms) ───
  │
  ▼
Loop.render()
  │
  ├── 1. root.Layout(constraints)
  │     └── Flex 分配空间 → ScrollView → BlockContainer → 各 Block.Measure/SetBounds
  │
  ├── 2. root.Paint(buffer)
  │     └── 各 Block.Paint(buf)
  │         ├── ThinkingBlock: 渲染标题行 + (展开时) 内容
  │         ├── AssistantTextBlock: IncrementalMarkdown.Render() → CellRow → buf
  │         ├── ToolCallBlock: 渲染图标 + 工具名 + spinner
  │         └── ...
  │
  ├── 3. overlayManager.Paint(buffer)
  │     └── 按层级从低到高绘制所有可见 overlay
  │         ├── Modal: 遮罩 + 对话框
  │         ├── Popup: 全屏代码查看器
  │         └── ...
  │
  ├── 4. 收集 HitRegions
  │     └── 遍历组件树 + overlay 层, 收集可点击区域
  │
  ├── 5. Renderer.EndFrame()
  │     ├── Diff: front buffer vs back buffer
  │     │   └── 只比较脏区域内的 cells
  │     ├── 生成 ANSI 输出
  │     │   ├── MoveTo(x, y)
  │     │   ├── SetStyle(fg, bg, flags)
  │     │   └── WriteString(text)
  │     └── Flush 到终端
  │
  └── 6. dirty = false
```

---

## 性能优化要点

### 1. 增量 Markdown 渲染

```go
// 不要每次 delta 都重新解析整个 markdown
// 策略:
// - debounce: 16ms 内的多个 delta 合并为一次渲染
// - 行级缓存: 只重新渲染变化的行
// - 代码块检测: 在代码块内时不触发 markdown 重新解析
```

### 2. Buffer Diff 优化

```go
// 不比较全屏, 只比较脏区域
// 脏区域追踪: 组件标记 dirty 时, 记录其 bounds
// diff 时只遍历脏区域内的 cells

// 进一步优化: 按行比较
// 如果一整行完全相同, 跳过 (memcmp 级别快速比较)
```

### 3. 写入批处理

```go
// ANSI 输出使用 bytes.Buffer 批量构建
// 一次 Flush 到底层 writer, 减少 syscall
// 相邻的相同样式 cells 合并为一个 WriteString 调用
```

### 4. 流式更新限流

```go
// 流式 delta 频率可能很高 (每 token 一次)
// 但渲染被 60fps tick 限制
// 所以即使 delta 到达 1000次/秒, 实际渲染 60次/秒
// 用户感知不到延迟 (16ms < 人眼阈值)
```

---

## 线程模型

```
┌─────────────────────────────────────────────────┐
│ Goroutine: Input Reader                         │
│   for { read(/dev/tty) → parser.Feed → eventCh }│
└──────────────────────┬──────────────────────────┘
                       │ eventCh
                       ▼
┌──────────────────────────────────────────────────┐
│ Goroutine: Main Event Loop (唯一的状态修改者)     │
│                                                  │
│   select {                                       │
│     case ev  := <-eventCh:  dispatch to tree     │
│     case app := <-streamCh: dispatch to blocks   │
│     case     := <-tick:     render if dirty      │
│   }                                              │
│                                                  │
│   所有状态修改在此 goroutine (无数据竞争)          │
└──────────────────────┬───────────────────────────┘
                       │ render output
                       ▼
┌──────────────────────────────────────────────────┐
│ Goroutine: AI Stream Reader (用户代码)            │
│   for { aiAPI.Recv() → app.SendDelta(streamCh) } │
└──────────────────────────────────────────────────┘
```

**核心原则**: 只有一个 goroutine (事件循环) 修改状态。输入读取和 AI 流式读取只通过 channel 发送数据。**零数据竞争**。

---

## 依赖清单

```
# 底层 (终端控制)
golang.org/x/term       # termios 封装 (仅 MakeRaw/Restore)
golang.org/x/sys        # syscall (SIGWINCH, 文件描述符)

# Markdown 解析
github.com/yuin/goldmark  # Markdown → AST (只用解析器, 自行渲染)

# 代码高亮
github.com/alecthomas/chroma  # 200+ 语言词法分析 (只用 lexer + token, 自行映射颜色)

# 无任何 TUI 框架依赖
# 不使用: bubbletea, lipgloss, tview, termui, glamour
```

---

## 开发路线图

### Phase 1: 终端地基 ✅

- [x] `internal/term`: raw mode + alt screen + SGR mouse + bracketed paste + resize
- [x] `internal/term/input.go`: 输入解析状态机 (键盘/鼠标/粘贴/分割序列)
- [x] `internal/buffer`: Cell/Color/Style/Buffer/Diff + wcwidth + CJK 宽字符
- [x] `render`: 双缓冲 diff 渲染器 + ANSI 批量输出
- [x] `event`: 通道驱动事件循环 + 60fps 渲染 tick
- [x] 目标: 能在终端绘制彩色文字, 响应键盘鼠标

### Phase 2: 组件 + Markdown ✅

- [x] `component`: Component 接口 + Flex + ScrollView + Text + Border
- [x] `markdown`: goldmark 集成 + AST 渲染 (heading/list/code/link/table)
- [x] `markdown/codeblock`: chroma 代码高亮
- [x] `markdown`: 流式 markdown 渲染
- [x] 目标: 能渲染完整的 markdown 文档, 代码带高亮

### Phase 3: AI Content Blocks ✅

- [x] `block`: Block 接口 + 生命周期 + Container
- [x] `block/thinking`: 可折叠 thinking block
- [x] `block/tool_call` + `tool_result`: 工具调用链
- [x] `block/assistant_text`: 流式 markdown block
- [x] `block/stream`: StreamDispatcher
- [x] `block/error`: ErrorBlock
- [x] 目标: 完整的 AI 对话流渲染

### Phase 4: 交互 + 动画 ✅

- [x] `hit`: 命中测试系统 (Region + RegionTree)
- [x] `animation`: spinner + fade + transition
- [x] 鼠标点击: 折叠/展开/复制/链接
- [x] `app/input`: InputLine 文本输入框
- [x] `theme`: 完整主题系统 (5 内置主题)
- [x] `overlay`: OverlayManager + Modal + Popup
- [x] `focus`: FocusManager (Tab 遍历)
- [x] `app`: ChatApp 高级 API
- [x] 目标: 完整可交互的 AI TUI

### Phase 5: AI 集成 ✅

- [x] `ai/client`: OpenAI 兼容流式聊天客户端
- [x] `ai/config`: .env 配置 + 环境变量
- [x] `app/ai_bridge`: AIBridge 对话桥接
- [x] Context.Context 支持 + 取消
- [x] 目标: 连接真实 LLM API

### Phase 6: 打磨 ✅

- [x] `theme`: 5 内置主题 + 31 颜色迁移 + 热切换
- [x] `block/registry`: Block 注册表 + 工厂模式
- [x] `block/error`: ErrorBlock + Registry
- [x] `component/layout`: Stack + Center + Padding
- [x] Windows 构建标签 (term_unix.go + term_windows.go)
- [x] README.md 完整项目文档

### Phase 7: 并发安全 + 生产基础 ✅

- [x] ChatApp mutex 修复 (6 race tests)
- [x] E2E 集成测试 (8 tests)
- [x] 输入历史 Up/Down (12 tests)
- [x] OSC52 剪贴板 (20 subtests)
- [x] Ctrl+C 优雅退出 (5 tests)
- [x] 虚拟滚动 PaintVisible (6 tests + 2 benchmarks)

### Phase 8: 深度优化 ✅

- [x] PaintVisible O(log n) 二分搜索 (32x 加速)
- [x] Block 序列化 SaveContainer/LoadContainer (16 tests)
- [x] OSC8 URL 超链接渲染 (12 tests)
- [x] Clipboard + termcompat 集成 (15 tests)
- [x] 终端兼容性检测 (25 tests, 12 终端)
- [x] Demo6 完整交互展示

### Phase 9: 质量增强 ✅

- [x] Ctrl+F 搜索功能 (19 tests)
- [x] Terminal output/term 层测试补全 (22 tests)
- [x] Component 测试补全 (21 tests)
- [x] Markdown theme 测试 + table 对齐 (21 tests)
- [x] 性能 benchmark 套件 (13 benchmarks)

### Phase 10: 生产级功能 ✅

- [x] TextArea 多行编辑器 (38 tests, emacs 快捷键 + Alt+Up/Down)
- [x] Command Palette Ctrl+P (21 tests, 模糊搜索)
- [x] Tab 补全 slash commands + @mentions (25 tests)
- [x] 生产级文档 docs/ (8 files) + examples/ (5 files)
- [x] Demo7 生产级 AI Agent (812 行)

### Phase 11: 愿景差异化 (进行中)

- [ ] `block/plugin`: 插件系统 (自定义 Block 类型)
- [ ] `app/recorder`: 会话录制/回放
- [ ] `internal/termcompat/image`: 图片协议检测 (Sixel/iTerm2/Kitty)
- [ ] `internal/termcompat/matrix`: 跨终端兼容性测试矩阵
- [ ] 核心模块测试补全 (buffer/theme/event)
- [ ] 目标: 超越所有竞品 TUI 库的差异化能力
