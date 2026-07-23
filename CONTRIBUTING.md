# Contributing to Fluui

Thank you for your interest in contributing to Fluui! This guide covers the basics.

## Development Setup

```bash
git clone https://github.com/topcheer/fluui
cd fluui
go build ./...
go test -race -short -count=1 ./... -timeout 300s
```

## Code Style

- **Tab indentation** (Go standard)
- **Mutex-protected** components (use `sync.Mutex` or `sync.RWMutex`)
- **Zero external deps** beyond goldmark (markdown) and chroma (syntax highlighting)
- Follow existing patterns in `component/` for new components

## Running Tests

```bash
# Full suite with race detector
GOCACHE=/tmp/go-cache-fluui go test -race -short -count=1 ./... -timeout 300s

# Single package
go test -race -short -count=1 ./component/ -timeout 60s

# Benchmarks
GOCACHE=/tmp/go-cache-fluui go test -bench=. -benchmem -benchtime=1s -run=^$ ./component/
```

## Adding a New Component

1. Create `component/your_component.go` implementing the `Component` interface:
   - `ID() string`
   - `Measure(cs Constraints) Size`
   - `SetBounds(r Rect)` / `Bounds() Rect`
   - `Paint(buf *buffer.Buffer)`
   - `Children() []Component`

2. Embed `BaseComponent` for default implementations.

3. Write tests in `component/your_component_test.go`.

4. Add benchmarks if the component has complex rendering.

## Coverage

- Target: 80%+ for all new functions
- Run: `go test -coverprofile=cov.out ./component/ && go tool cover -func=cov.out`
- No-op stub functions (`func Foo() {}`) report as 0% — this is a known Go limitation

## Commit Convention

Format: `type(Pxxx): description`

Types: `feat`, `fix`, `perf`, `test`, `docs`, `refactor`

## Release Checklist

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] Full test suite passes with `-race`
- [ ] No TODO/FIXME in core code
- [ ] README and docs are up to date
