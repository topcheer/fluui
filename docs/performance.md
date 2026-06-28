# Performance Guide

Fluui is designed for responsive terminal UIs with zero framework dependencies. This document covers benchmark methodology, performance baselines, profiling techniques, and optimization strategies.

## Table of Contents

- [Running Benchmarks](#running-benchmarks)
- [Performance Baseline](#performance-baseline)
- [Profiling with pprof](#profiling-with-pprof)
- [Optimization Strategies](#optimization-strategies)
- [Component Performance Tips](#component-performance-tips)
- [Memory Management](#memory-management)

## Running Benchmarks

Fluui includes **54 benchmarks** across 8 benchmark files covering rendering, buffer operations, block containers, components, and link detection.

### Quick Benchmark Run

```bash
# Run all benchmarks (1 iteration each, quick smoke test)
go test -bench=. -benchtime=1x -run=^$ ./...

# Run with memory allocation reporting
go test -bench=. -benchmem -benchtime=100x -run=^$ ./...

# Run benchmarks for a specific package
go test -bench=. -benchmem ./render/...
go test -bench=. -benchmem ./block/...
go test -bench=. -benchmem ./component/...
go test -bench=. -benchmem ./internal/buffer/...

# Full benchmark with statistical significance
go test -bench=. -benchmem -benchtime=3s -count=5 -run=^$ ./...
```

### Comparing Benchmarks

```bash
# Save baseline
go test -bench=. -benchmem -count=3 -run=^$ ./... > bench-before.txt

# Make optimization changes, then compare
go test -bench=. -benchmem -count=3 -run=^$ ./... > bench-after.txt

# View comparison
benchstat bench-before.txt bench-after.txt
```

### Benchmark Files

| File | Package | Focus |
|---|---|---|
| `render/benchmark_test.go` | render | Full render, diff render, no-change, large screen |
| `block/benchmark_test.go` | block | Container add blocks, paint blocks |
| `block/virtual_test.go` | block | Virtual scroll paint (binary search vs linear) |
| `block/virtual_binary_test.go` | block | Paint visible/all blocks |
| `internal/buffer/benchmark_test.go` | buffer | Buffer diff (identical, small, large), draw text |
| `internal/term/benchmark_test.go` | term | Writer batch style changes, flush |
| `component/p16_bench_test.go` | component | Table, Tree, ProgressBar paint |
| `component/p17_link_bench_test.go` | component | Link detection, scan text, annotate buffer |

## Performance Baseline

Baseline measured on Apple M2 (darwin/arm64), Go 1.23+, benchtime=100x.

### Render Pipeline

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| RenderFull | 55,724 | 10,208 | 161 | Full 80x24 screen render |
| RenderDiff | 45,165 | 11 | 0 | Incremental diff render |
| RenderNoChange | 43,441 | 0 | 0 | No-change render (optimization target) |
| RenderLargeScreen | 279,996 | 58,575 | 840 | 200x50 screen render |

### Block Container

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| ContainerAddBlocks100 | 32,778 | 33,499 | 509 | Add 100 blocks |
| ContainerAddBlocks1000 | 329,874 | 319,639 | 6,500 | Add 1000 blocks |
| ContainerPaint100 | 1,059,349 | 3,224,531 | 6,135 | Paint 100 blocks (optimization target) |
| ContainerBlocks | 235 | 1,792 | 1 | Single block access |
| PaintVisible1000Blocks | 21 | 0 | 0 | Virtual scroll visible paint |
| PaintVisible10000Blocks | 29 | 0 | 0 | Virtual scroll 10K blocks |
| PaintVisibleBinarySearch | 22 | 0 | 0 | Binary search for visible range |
| PaintVisibleLinearScan | 640 | 0 | 0 | Linear scan comparison |
| PaintAll1000Blocks | 964 | 0 | 0 | Paint all blocks |

### Buffer Operations

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| BufferDrawText | 619 | 0 | 0 | Draw ASCII text |
| BufferDrawTextCJK | 235 | 0 | 0 | Draw CJK text |
| DiffIdentical | 14,160 | 0 | 0 | Diff identical buffers |
| DiffSmall | 14,527 | 64 | 1 | Diff with 1 cell changed |
| DiffLarge | 36,533 | 146,037 | 11 | Diff fully different buffers |

### Terminal Writer

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| WriterBatchStyleChanges | 42,236 | 32,955 | 799 | Batch style changes (optimization target) |
| WriterFlush | 345 | 5,073 | 0 | Flush to writer |

### Components

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| TablePaint | 11,650 | 24 | 1 | Table render |
| TreePaint | 3,695 | 560 | 47 | Tree render |
| ProgressBarPaint | 291 | 16 | 2 | Progress bar render |
| StatusBar_Paint | 783 | 40 | 2 | Status bar render |
| StatusBar_Measure | 106 | 0 | 0 | Status bar measure |
| StatusBar_UpdateItems | 180 | 32 | 3 | Update status items |
| TabBar_Paint | 1,191 | 0 | 0 | Tab bar render (zero-alloc) |
| TabBar_Measure | 186 | 0 | 0 | Tab bar measure |
| TabBar_Navigation | 32 | 0 | 0 | Tab switching |
| TabBar_AddTabs | 1,522 | 1,760 | 26 | Add tabs |
| DiffPreview_Paint | 4,711 | 0 | 0 | Diff preview render (zero-alloc) |
| DiffPreview_Scroll | 30 | 0 | 0 | Scroll diff preview |
| DiffPreview_SetDiff | 3,857 | 5,492 | 4 | Set diff content |
| FilePicker_Measure | 87 | 0 | 0 | File picker measure (zero-alloc) |
| FilePicker_Paint | 4,835 | 0 | 0 | File picker render (zero-alloc) |
| FilePicker_Navigation | 27 | 0 | 0 | File navigation (zero-alloc) |
| FilePicker_Filter | 6,778 | 1,792 | 1 | Filter files |
| FilePicker_LoadDir_100 | 5,227 | 1,680 | 11 | Load 100 entries |
| FilePicker_LoadDir_1000 | 38,872 | 8,976 | 11 | Load 1000 entries |

### Link Detection

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| DetectLinks_10_URLs | 1,362 | 4,224 | 4 | Detect 10 URLs |
| DetectLinks_100_URLs | 12,575 | 39,296 | 7 | Detect 100 URLs |
| DetectLinks_SingleURL | 359 | 256 | 1 | Single URL detection |
| DetectLinks_NoURLs_500 | 2,075 | 0 | 0 | No URLs (fast exit) |
| DetectLinks_MixedSchemes | 6,084 | 19,328 | 22 | Mixed URL schemes |
| DetectLinks_WWW_100 | 11,602 | 22,016 | 106 | WWW-style URLs |
| ScanText_100Lines | 14,444 | 22,080 | 57 | Scan 100 lines |
| ScanText_500Lines | 67,602 | 127,144 | 343 | Scan 500 lines (optimization target) |
| AnnotateBuffer_20Links | 4,173 | 6,400 | 200 | Annotate 20 links (optimization target) |
| LinkManager_ScanText | 23,872 | 65,024 | 108 | Full scan |
| LinkManager_AnnotateBuffer | 7,981 | 13,792 | 431 | Full annotate (optimization target) |

### ParseDiff

| Benchmark | ns/op | B/op | allocs/op | Notes |
|---|---|---|---|---|
| ParseDiff_100 | 4,082 | 5,120 | 2 | Parse 100-line diff |
| ParseDiff_500 | 16,878 | 23,040 | 2 | Parse 500-line diff |

## Profiling with pprof

### CPU Profiling

```bash
# CPU profile for a specific benchmark
go test -bench=BenchmarkRenderFull -cpuprofile=cpu.prof -benchtime=5s -run=^$ ./render/

# Analyze
go tool pprof cpu.prof
# Interactive commands: top, list <func>, web, png

# Generate SVG call graph
go tool pprof -svg cpu.prof > cpu.svg

# Flame graph (requires go-torch or pprof built-in)
go tool pprof -flame cpu.prof > flame.svg
```

### Memory Profiling

```bash
# Memory profile
go test -bench=BenchmarkContainerPaint100 -memprofile=mem.prof -benchtime=5s -run=^$ ./block/

# Analyze allocations
go tool pprof -alloc_objects mem.prof

# Analyze allocation bytes
go tool pprof -alloc_space mem.prof
```

### Trace

```bash
# Execution trace
go test -bench=BenchmarkRenderFull -trace=trace.out -benchtime=2s -run=^$ ./render/
go tool trace trace.out
```

### Live Profiling

For running applications, add pprof HTTP endpoints:

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then capture:
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// go tool pprof http://localhost:6060/debug/pprof/heap
```

## Optimization Strategies

### 1. Reduce Allocations

The single most impactful optimization in Go TUI rendering is reducing heap allocations.

**Use sync.Pool for reusable objects:**

```go
var cellPool = sync.Pool{
    New: func() interface{} {
        return &buffer.Cell{Width: 1}
    },
}

// Acquire
cell := cellPool.Get().(*buffer.Cell)
cell.Rune = 'x'
cell.Fg = color

// ... use cell ...

// Release (reset before returning)
cell.Rune = 0
cell.Fg = 0
cellPool.Put(cell)
```

**Pre-allocate slices:**

```go
// Bad: grows slice incrementally
var cells []Cell
for _, item := range items {
    cells = append(cells, makeCell(item))
}

// Good: pre-allocate exact capacity
cells := make([]Cell, 0, len(items))
for _, item := range items {
    cells = append(cells, makeCell(item))
}
```

### 2. Avoid String Concatenation in Hot Paths

```go
// Bad: allocates a new string each iteration
for _, r := range text {
    s += string(r)
}

// Good: use strings.Builder
var sb strings.Builder
sb.Grow(len(text))
for _, r := range text {
    sb.WriteRune(r)
}
result := sb.String()
```

### 3. Batch Buffer Operations

```go
// Bad: SetCell one at a time with style computation
for x := 0; x < w; x++ {
    buf.SetCell(x, y, computeCell(x, y))
}

// Good: pre-compute style, reuse cell
style := computeStyle()
cell := buffer.NewCell(' ', style)
for x := 0; x < w; x++ {
    cell.Rune = getRune(x)
    buf.SetCell(x, y, cell)
}
```

### 4. Short-Circuit No-Change Paths

```go
// Fast path: if nothing changed, skip rendering entirely
if !dirty {
    return
}
```

### 5. Use Binary Search for Virtual Scrolling

Fluui's block container uses binary search to find visible blocks in O(log n) instead of O(n) linear scan:

```go
// Binary search for first visible block
lo, hi := 0, len(blocks)
for lo < hi {
    mid := (lo + hi) / 2
    if blocks[mid].Bounds.Y+blocks[mid].Bounds.H <= scrollY {
        lo = mid + 1
    } else {
        hi = mid
    }
}
```

## Component Performance Tips

### Measure Before Paint

Always call `Measure()` before `SetBounds()` to let the component compute its preferred size:

```go
comp.Measure(component.Bounded(w, h))
comp.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
comp.Paint(buf)
```

### Minimize Work in OnPaint

The `OnPaint` callback runs on every render cycle. Keep it lean:

```go
// Good: compute layout once, cache results
app.OnPaint(func(buf *buffer.Buffer) {
    w, h := app.Size()
    if w != cachedW || h != cachedH {
        recomputeLayout(w, h)
        cachedW, cachedH = w, h
    }
    paintCachedLayout(buf)
})
```

### Use HandleKey for Built-in Navigation

Components like `Table`, `FilePicker`, and `Tree` have built-in `HandleKey()` methods that handle navigation without extra allocations:

```go
app.OnKey(func(k *term.KeyEvent) {
    if table.HandleKey(k) {
        app.MarkDirty()
        return
    }
    // custom key handling
})
```

### Batch Data Updates

When updating component data, use batch methods instead of incremental updates:

```go
// Good: single SetRows call
table.SetRows(allRows)

// Avoid: incremental AddRow in a loop
for _, row := range allRows {
    table.AddRow(row) // triggers internal recalculation each time
}
```

## Memory Management

### Buffer Reuse

The renderer maintains front and back buffers. The `Diff()` function compares them to produce minimal output:

- `DiffIdentical` (0 allocs) — row-skip optimization skips identical rows
- `DiffSmall` (1 alloc) — only changed rows produce output
- `DiffLarge` (11 allocs) — worst case, all rows differ

### Writer Batching

The terminal writer batches consecutive cells with the same style to minimize ANSI escape sequences:

```go
// The writer groups cells by style automatically
writer.SetStyle(style1)
writer.WriteText("Hello ")
writer.SetStyle(style2)
writer.WriteText("World")
writer.Flush() // single write syscall
```

### Goroutine Safety

All Fluui components use `sync.RWMutex` for thread safety. To avoid lock contention:

- Use `RLock()` for read-heavy paths (Measure, Paint)
- Use `Lock()` only for mutations (SetValue, SetData)
- Keep critical sections short
- Call `MarkDirty()` outside the lock

## Performance Checklist

When adding new components or optimizing existing ones:

- [ ] Benchmark with `-benchmem` to measure allocations
- [ ] Target zero allocations for hot paths (Measure, Paint, HandleKey)
- [ ] Use `sync.Pool` for frequently allocated objects
- [ ] Pre-allocate slices with known capacity
- [ ] Profile with pprof before and after optimization
- [ ] Run `benchstat` to verify improvement
- [ ] Ensure all tests still pass with `-race`

---

## Phase 24 Optimization Results

### Render Pipeline (P24-A: gg_arch, commit 6fa6a2d)
| Benchmark | Before ns/op | After ns/op | Improvement | Allocs |
|---|---|---|---|---|
| RenderFull | 43,690 | 24,262 | **44% faster** | 161 (unchanged) |
| RenderDiff | 44,201 | 23,712 | **46% faster** | 0 (unchanged) |
| RenderNoChange | 43,779 | 24,297 | **44% faster** | 0 (unchanged) |
| RenderLargeScreen | 228,865, 16 allocs | 119,664, 8 allocs | **48% faster** | **50% fewer allocs** |

Key optimizations:
1. EndFrame fast path: `len(ops)==0` skips ALL terminal I/O and buffer copy
2. ASCII rune cache (`asciiChars[128]`): pre-computed single-char strings for runes 0-127
3. `utf8.EncodeRune` + `WriteRaw` for non-ASCII: stack buffer, no heap allocation
4. `DiffInto` + `cellFastEqual`: fast `==` path for cells without links

### Link Detection (P24-B: gg_dev2, commit 1f062d8)
| Benchmark | Before | After | Improvement |
|---|---|---|---|
| AnnotateBuffer_20Links | 4,387 ns, 6.4KB, 200 allocs | 582 ns, 640B, 20 allocs | **87% faster, 90% less mem, 90% fewer allocs** |
| LinkManager_AnnotateBuffer | 9,169 ns, 13.8KB, 431 allocs | 875 ns, 640B, 20 allocs | **90% faster, 95% less mem, 95% fewer allocs** |
| ScanText_500Lines | 66,486 ns, 127KB, 343 allocs | 57,697 ns, 46.8KB, 334 allocs | **13% faster, 63% less mem** |
| ScanText_100Lines | 16,263 ns, 22KB, 57 allocs | 15,131 ns, 12.8KB, 50 allocs | **7% faster, 42% less mem, 12% fewer allocs** |
| DetectLinks_SingleURL | 409 ns, 536B, 2 allocs | 297 ns, 256B, 1 alloc | **27% faster, 52% less mem, 50% fewer allocs** |

Key optimizations:
1. AnnotateBuffer: Shared `*buffer.Link` pointer per link range (was 1 alloc/cell → 1 alloc/link). Also direct `buf.Cells[idx]` access instead of GetCell/SetCell (zero-copy)
2. ScanText: Reuse existing slice capacity via `lm.links[:0:cap(lm.links)]` with fallback cap=8. Skip empty DetectLinks results
3. DetectLinks: Documentation only (www string concat is unavoidable — each URL is different)

### Container/Buffer (P24-D)
| Benchmark | Before allocs | After allocs | Status |
|---|---|---|---|
| ContainerPaint100 | 6,135 | TBD | in progress |

Key finding: goldmark markdown parser is the primary allocation source (~3.3% each).

---

*Phase 24 optimization in progress. Baseline measured on Apple M2 Ultra (darwin/arm64), Go 1.23+.*
